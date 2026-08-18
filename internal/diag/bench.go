package diag

import (
	"bytes"
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
}

// RunBenchmark conducts a concurrent stress and performance benchmark against the gateway.
func RunBenchmark(baseURL, apiKey string, concurrency, totalRequests int) BenchmarkReport {
	baseURL = strings.TrimRight(baseURL, "/")
	client := &http.Client{Timeout: 60 * time.Second}

	if concurrency <= 0 {
		concurrency = 3
	}
	if totalRequests <= 0 {
		totalRequests = 6
	}

	var successful, failed, totalTokens atomic.Int64
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
				httpReq, err := http.NewRequest("POST", baseURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
				if err != nil {
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

				var chatRes struct {
					Usage struct {
						TotalTokens int `json:"total_tokens"`
					} `json:"usage"`
				}
				_ = json.NewDecoder(resp.Body).Decode(&chatRes)
				resp.Body.Close()

				successful.Add(1)
				if chatRes.Usage.TotalTokens > 0 {
					totalTokens.Add(int64(chatRes.Usage.TotalTokens))
				} else {
					totalTokens.Add(30) // estimated fallback
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
	tps := float64(totalTokens.Load()) / totalDur.Seconds()

	return BenchmarkReport{
		TotalRequests:   totalRequests,
		Concurrency:     concurrency,
		Successful:      successful.Load(),
		Failed:          failed.Load(),
		TotalDuration:   totalDur,
		AverageLatency:  avg,
		P50Latency:      p50,
		P90Latency:      p90,
		P99Latency:      p99,
		TotalTokens:     totalTokens.Load(),
		TokensPerSecond: tps,
		RequestsPerSec:  rps,
	}
}
