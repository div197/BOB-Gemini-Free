package diag

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// BenchmarkReport summarizes performance metrics across a concurrent batch of requests.
type BenchmarkReport struct {
	TotalRequests   int
	Concurrency     int
	Successful      int64
	Failed          int64
	TotalDuration   time.Duration
	AverageLatency  time.Duration
	P50Latency      time.Duration
	P90Latency      time.Duration
	P99Latency      time.Duration
	TotalTokens     int64
	TokensPerSecond float64
	RequestsPerSec  float64
	// TokenCountsMeasured is true only when every successful response supplied
	// a positive provider-reported total_tokens value. A benchmark must not
	// invent token counts when the gateway or provider omits usage metadata.
	TokenCountsMeasured bool
}

const (
	maxBenchmarkConcurrency = 128
	maxBenchmarkRequests    = 10000
)

func normalizeBenchmarkSettings(concurrency, totalRequests int) (int, int) {
	if concurrency <= 0 {
		concurrency = 3
	}
	if totalRequests <= 0 {
		totalRequests = 6
	}
	if concurrency > maxBenchmarkConcurrency {
		concurrency = maxBenchmarkConcurrency
	}
	if totalRequests > maxBenchmarkRequests {
		totalRequests = maxBenchmarkRequests
	}
	return concurrency, totalRequests
}

// RunBenchmark conducts a concurrent stress and performance benchmark against the gateway.
func RunBenchmark(baseURL, apiKey string, concurrency, totalRequests int) BenchmarkReport {
	baseURL = strings.TrimRight(baseURL, "/")

	concurrency, totalRequests = normalizeBenchmarkSettings(concurrency, totalRequests)

	transport := &http.Transport{
		MaxIdleConns:        concurrency * 4,
		MaxIdleConnsPerHost: concurrency * 4,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}

	var successful, failed, totalTokens, tokenCountSamples atomic.Int64
	latencies := make([]time.Duration, totalRequests)
	var latenciesMu sync.Mutex

	workChan := make(chan int, totalRequests)
	for i := 0; i < totalRequests; i++ {
		workChan <- i
	}
	close(workChan)

	startTotal := time.Now()
	var wg sync.WaitGroup

	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for reqIdx := range workChan {
				reqStart := time.Now()
				payload := map[string]any{
					"model": "gemini-3.7-flash",
					"messages": []map[string]string{
						{"role": "user", "content": fmt.Sprintf("Benchmark query #%d: reply with 5 random words.", reqIdx+1)},
					},
				}
				bodyBytes, _ := json.Marshal(payload)
				httpReq, err := newDiagnosticRequest(http.MethodPost, baseURL+"/v1/chat/completions", bodyBytes)
				if err != nil {
					latenciesMu.Lock()
					latencies[reqIdx] = time.Since(reqStart)
					latenciesMu.Unlock()
					failed.Add(1)
					continue
				}
				httpReq.Header.Set("Content-Type", "application/json")
				if apiKey != "" {
					httpReq.Header.Set("Authorization", "Bearer "+apiKey)
				}

				resp, err := client.Do(httpReq)
				dur := time.Since(reqStart)

				latenciesMu.Lock()
				latencies[reqIdx] = dur
				latenciesMu.Unlock()

				if err != nil || resp.StatusCode != http.StatusOK {
					failed.Add(1)
					if resp != nil {
						resp.Body.Close()
					}
					continue
				}

				responseBody, readErr := readDiagnosticBody(resp)
				resp.Body.Close()
				if readErr != nil {
					failed.Add(1)
					continue
				}

				var chatRes struct {
					Choices []struct {
						Message struct {
							Content   string            `json:"content"`
							Reasoning string            `json:"reasoning_content"`
							ToolCalls []json.RawMessage `json:"tool_calls"`
						} `json:"message"`
					} `json:"choices"`
					Usage struct {
						TotalTokens int `json:"total_tokens"`
					} `json:"usage"`
				}
				if !json.Valid(responseBody) || json.Unmarshal(responseBody, &chatRes) != nil {
					failed.Add(1)
					continue
				}
				if len(chatRes.Choices) == 0 {
					failed.Add(1)
					continue
				}
				message := chatRes.Choices[0].Message
				if strings.TrimSpace(message.Content) == "" && strings.TrimSpace(message.Reasoning) == "" && len(message.ToolCalls) == 0 {
					failed.Add(1)
					continue
				}
				successful.Add(1)
				if chatRes.Usage.TotalTokens > 0 {
					totalTokens.Add(int64(chatRes.Usage.TotalTokens))
					tokenCountSamples.Add(1)
				}
			}
		}()
	}

	wg.Wait()
	totalDur := time.Since(startTotal)

	var validLatencies []time.Duration
	var sumLatency time.Duration
	for _, l := range latencies {
		if l > 0 {
			validLatencies = append(validLatencies, l)
			sumLatency += l
		}
	}
	sort.Slice(validLatencies, func(i, j int) bool {
		return validLatencies[i] < validLatencies[j]
	})

	var p50, p90, p99, avg time.Duration
	if len(validLatencies) > 0 {
		avg = sumLatency / time.Duration(len(validLatencies))
		p50 = validLatencies[len(validLatencies)*50/100]
		p90 = validLatencies[len(validLatencies)*90/100]
		p99 = validLatencies[len(validLatencies)-1]
	}

	rps := float64(successful.Load()) / totalDur.Seconds()
	tokenCountsMeasured := successful.Load() > 0 && tokenCountSamples.Load() == successful.Load()
	tps := float64(0)
	if tokenCountsMeasured && totalDur > 0 {
		tps = float64(totalTokens.Load()) / totalDur.Seconds()
	}

	return BenchmarkReport{
		TotalRequests:       totalRequests,
		Concurrency:         concurrency,
		Successful:          successful.Load(),
		Failed:              failed.Load(),
		TotalDuration:       totalDur,
		AverageLatency:      avg,
		P50Latency:          p50,
		P90Latency:          p90,
		P99Latency:          p99,
		TotalTokens:         totalTokens.Load(),
		TokensPerSecond:     tps,
		RequestsPerSec:      rps,
		TokenCountsMeasured: tokenCountsMeasured,
	}
}
