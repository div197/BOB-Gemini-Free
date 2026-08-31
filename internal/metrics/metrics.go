package metrics

import (
	"sync/atomic"
	"time"
)

var latencyBounds = [...]time.Duration{
	time.Millisecond,
	10 * time.Millisecond,
	50 * time.Millisecond,
	100 * time.Millisecond,
	250 * time.Millisecond,
	500 * time.Millisecond,
	time.Second,
	5 * time.Second,
}

// Route identifies a bounded, user-visible protocol path. Keep this enum
// deliberately finite: metric labels must never be derived from URLs, model
// names, credentials, or other request-controlled values.
type Route uint8

const (
	RouteOpenAIChatWebRPC Route = iota
	RouteOpenAIResponsesWebRPC
	RouteAnthropicWebRPC
	RouteGoogleWebRPC
	RouteGeminiDeveloperAPI
	RouteImageGenerationWebRPC
	RouteRefineWebRPC
	routeCount
)

var routeNames = [...]string{
	"openai_chat_web_rpc",
	"openai_responses_web_rpc",
	"anthropic_web_rpc",
	"google_web_rpc",
	"gemini_developer_api",
	"image_generation_web_rpc",
	"refine_web_rpc",
}

type Histogram struct {
	count  atomic.Uint64
	sumNS  atomic.Uint64
	bucket [len(latencyBounds) + 1]atomic.Uint64
}

func (h *Histogram) Observe(duration time.Duration) {
	if duration < 0 {
		duration = 0
	}
	h.count.Add(1)
	h.sumNS.Add(uint64(duration))
	for i, bound := range latencyBounds {
		if duration <= bound {
			h.bucket[i].Add(1)
			return
		}
	}
	h.bucket[len(latencyBounds)].Add(1)
}

func (h *Histogram) Snapshot() map[string]any {
	buckets := make([]uint64, len(h.bucket))
	for i := range h.bucket {
		buckets[i] = h.bucket[i].Load()
	}
	return map[string]any{
		"count":     h.count.Load(),
		"sum_ms":    float64(h.sumNS.Load()) / float64(time.Millisecond),
		"bucket_le": []string{"1ms", "10ms", "50ms", "100ms", "250ms", "500ms", "1s", "5s", "+Inf"},
		"buckets":   buckets,
	}
}

// Registry contains only aggregate counters and bounded latency summaries.
// It intentionally has no fields for request content, cookies, auth headers,
// image bytes, or user identifiers.
type Registry struct {
	RequestsTotal      atomic.Uint64
	RequestsInFlight   atomic.Int64
	UpstreamRequests   atomic.Uint64
	UpstreamErrors     atomic.Uint64
	Upstream429        atomic.Uint64
	StreamRetries      atomic.Uint64
	SessionPoolTotal   atomic.Int64
	SessionPoolHealthy atomic.Int64
	SessionFailovers   atomic.Uint64
	ImageUploads       atomic.Uint64
	ImageCacheHits     atomic.Uint64
	ImageCacheMisses   atomic.Uint64
	TokensEstimated    atomic.Uint64
	RequestLatency     Histogram
	UpstreamLatency    Histogram
	routes             [routeCount]atomic.Uint64
}

func New() *Registry {
	return &Registry{}
}

// ObserveRoute records one request against a fixed route name. Invalid enum
// values are ignored so a future caller cannot panic or create an unbounded
// metric dimension.
func (r *Registry) ObserveRoute(route Route) {
	if r == nil || route >= routeCount {
		return
	}
	r.routes[route].Add(1)
}

func (r *Registry) Snapshot() map[string]any {
	if r == nil {
		return map[string]any{}
	}
	routes := make(map[string]uint64, routeCount)
	for index, name := range routeNames {
		routes[name] = r.routes[index].Load()
	}
	return map[string]any{
		"requests_total":          r.RequestsTotal.Load(),
		"requests_inflight":       r.RequestsInFlight.Load(),
		"upstream_requests_total": r.UpstreamRequests.Load(),
		"upstream_errors_total":   r.UpstreamErrors.Load(),
		"upstream_429_total":      r.Upstream429.Load(),
		"stream_retries_total":    r.StreamRetries.Load(),
		"session_pool_total":      r.SessionPoolTotal.Load(),
		"session_pool_healthy":    r.SessionPoolHealthy.Load(),
		"session_failovers_total": r.SessionFailovers.Load(),
		"image_upload_total":      r.ImageUploads.Load(),
		"image_cache_hits":        r.ImageCacheHits.Load(),
		"image_cache_misses":      r.ImageCacheMisses.Load(),
		"tokens_estimated":        r.TokensEstimated.Load(),
		"request_latency":         r.RequestLatency.Snapshot(),
		"upstream_latency":        r.UpstreamLatency.Snapshot(),
		"routes":                  routes,
	}
}
