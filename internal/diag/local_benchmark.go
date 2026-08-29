package diag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/div197/bob-gemini-free/internal/config"
	"github.com/div197/bob-gemini-free/internal/server"
)

// LocalBenchmarkReport describes a gateway-only benchmark. The upstream
// requester is a deterministic in-process fixture, so these values exclude
// Google network latency and rate limiting by construction.
type LocalBenchmarkReport struct {
	RuntimeVersion    string        `json:"runtime_version"`
	GOOS              string        `json:"goos"`
	GOARCH            string        `json:"goarch"`
	GatewayVersion    string        `json:"gateway_version"`
	Concurrency       int           `json:"concurrency"`
	TotalRequests     int           `json:"total_requests"`
	Successful        int64         `json:"successful"`
	Failed            int64         `json:"failed"`
	TotalDuration     time.Duration `json:"total_duration_ns"`
	AverageLatency    time.Duration `json:"average_latency_ns"`
	P50Latency        time.Duration `json:"p50_latency_ns"`
	P90Latency        time.Duration `json:"p90_latency_ns"`
	P95Latency        time.Duration `json:"p95_latency_ns"`
	P99Latency        time.Duration `json:"p99_latency_ns"`
	RequestsPerSecond float64       `json:"requests_per_second"`
	AllocsPerRequest  float64       `json:"allocs_per_request"`
	AllocatedBytes    uint64        `json:"allocated_bytes"`
	RSSBytes          uint64        `json:"rss_bytes"`
	RSSAvailable      bool          `json:"rss_available"`
	Goroutines        int           `json:"goroutines"`
	MaxConnections    int64         `json:"max_connections"`
}

type localBenchmarkRequester struct{}

const (
	maxLocalBenchmarkConcurrency = 128
	maxLocalBenchmarkRequests    = 10000
)

func normalizeLocalBenchmarkSettings(concurrency, totalRequests int) (int, int) {
	if concurrency <= 0 {
		concurrency = 1
	}
	if totalRequests <= 0 {
		totalRequests = 100
	}
	if concurrency > maxLocalBenchmarkConcurrency {
		concurrency = maxLocalBenchmarkConcurrency
	}
	if totalRequests > maxLocalBenchmarkRequests {
		totalRequests = maxLocalBenchmarkRequests
	}
	return concurrency, totalRequests
}

func (localBenchmarkRequester) Do(req *http.Request) (*http.Response, error) {
	if req.Method == http.MethodGet {
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(bytes.NewReader([]byte(`<html></html>`)))}, nil
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(localBenchmarkGeminiBody())),
	}, nil
}

func localBenchmarkGeminiBody() []byte {
	text := "local benchmark response"
	inner := []any{nil, nil, nil, nil, []any{[]any{nil, []any{text}}}, nil, nil, "fixture-metadata-padding-" + string(make([]byte, 160))}
	innerBytes, _ := json.Marshal(inner)
	outer := []any{[]any{"wrb.fr", nil, string(innerBytes)}}
	outerBytes, _ := json.Marshal(outer)
	return append(outerBytes, '\n')
}

type trackedDialer struct {
	active atomic.Int64
	max    atomic.Int64
}

func (d *trackedDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	conn, err := (&net.Dialer{}).DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	active := d.active.Add(1)
	for {
		max := d.max.Load()
		if active <= max || d.max.CompareAndSwap(max, active) {
			break
		}
	}
	return &trackedConn{Conn: conn, active: &d.active}, nil
}

type trackedConn struct {
	net.Conn
	active *atomic.Int64
	once   sync.Once
}

func (c *trackedConn) Close() error {
	c.once.Do(func() { c.active.Add(-1) })
	return c.Conn.Close()
}

func RunLocalBenchmark(concurrency, totalRequests int) LocalBenchmarkReport {
	concurrency, totalRequests = normalizeLocalBenchmarkSettings(concurrency, totalRequests)

	cfg := config.Default()
	cfg.LogRequests = false
	cfg.RetryAttempts = 1
	app := server.New(cfg, "local-benchmark")
	defer app.Close()
	app.Gem.HTTP = localBenchmarkRequester{}
	serverInstance := httptest.NewServer(app.Handler())
	defer serverInstance.Close()

	dialer := &trackedDialer{}
	transport := &http.Transport{
		MaxIdleConns:        concurrency * 2,
		MaxIdleConnsPerHost: concurrency * 2,
		DialContext:         dialer.DialContext,
	}
	client := &http.Client{Transport: transport, Timeout: 30 * time.Second}
	payload := []byte(`{"model":"gemini-3.7-flash","messages":[{"role":"user","content":"local benchmark"}]}`)

	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	latencies := make([]time.Duration, totalRequests)
	var successful atomic.Int64
	var failed atomic.Int64
	work := make(chan int, totalRequests)
	for i := 0; i < totalRequests; i++ {
		work <- i
	}
	close(work)
	start := time.Now()
	var wg sync.WaitGroup
	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range work {
				req, err := http.NewRequest(http.MethodPost, serverInstance.URL+"/v1/chat/completions", bytes.NewReader(payload))
				if err != nil {
					failed.Add(1)
					continue
				}
				req.Header.Set("Content-Type", "application/json")
				reqStart := time.Now()
				resp, err := client.Do(req)
				latencies[index] = time.Since(reqStart)
				if err != nil || resp == nil || resp.StatusCode != http.StatusOK {
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
				var response struct {
					Choices []struct {
						Message struct {
							Content   string            `json:"content"`
							Reasoning string            `json:"reasoning_content"`
							ToolCalls []json.RawMessage `json:"tool_calls"`
						} `json:"message"`
					} `json:"choices"`
				}
				if !json.Valid(responseBody) || json.Unmarshal(responseBody, &response) != nil || len(response.Choices) == 0 {
					failed.Add(1)
					continue
				}
				message := response.Choices[0].Message
				if strings.TrimSpace(message.Content) == "" && strings.TrimSpace(message.Reasoning) == "" && len(message.ToolCalls) == 0 {
					failed.Add(1)
					continue
				}
				successful.Add(1)
			}
		}()
	}
	wg.Wait()
	totalDuration := time.Since(start)
	runtime.ReadMemStats(&after)
	transport.CloseIdleConnections()

	valid := make([]time.Duration, 0, totalRequests)
	var sum time.Duration
	for _, latency := range latencies {
		if latency > 0 {
			valid = append(valid, latency)
			sum += latency
		}
	}
	sort.Slice(valid, func(i, j int) bool { return valid[i] < valid[j] })
	quantile := func(percent int) time.Duration {
		if len(valid) == 0 {
			return 0
		}
		index := (len(valid) - 1) * percent / 100
		return valid[index]
	}
	var average time.Duration
	if len(valid) > 0 {
		average = sum / time.Duration(len(valid))
	}
	elapsedSeconds := totalDuration.Seconds()
	if elapsedSeconds <= 0 {
		elapsedSeconds = 1e-9
	}
	allocated := after.TotalAlloc - before.TotalAlloc
	allocs := after.Mallocs - before.Mallocs
	rss, rssAvailable := currentRSSBytes()
	return LocalBenchmarkReport{
		RuntimeVersion:    runtime.Version(),
		GOOS:              runtime.GOOS,
		GOARCH:            runtime.GOARCH,
		GatewayVersion:    "local-benchmark",
		Concurrency:       concurrency,
		TotalRequests:     totalRequests,
		Successful:        successful.Load(),
		Failed:            failed.Load(),
		TotalDuration:     totalDuration,
		AverageLatency:    average,
		P50Latency:        quantile(50),
		P90Latency:        quantile(90),
		P95Latency:        quantile(95),
		P99Latency:        quantile(99),
		RequestsPerSecond: float64(successful.Load()) / elapsedSeconds,
		AllocsPerRequest:  float64(allocs) / float64(totalRequests),
		AllocatedBytes:    allocated,
		RSSBytes:          rss,
		RSSAvailable:      rssAvailable,
		Goroutines:        runtime.NumGoroutine(),
		MaxConnections:    dialer.max.Load(),
	}
}

func formatLocalBenchmarkSummary(report LocalBenchmarkReport) string {
	return fmt.Sprintf("c=%d requests=%d success=%d p50=%s p95=%s p99=%s", report.Concurrency, report.TotalRequests, report.Successful, report.P50Latency, report.P95Latency, report.P99Latency)
}
