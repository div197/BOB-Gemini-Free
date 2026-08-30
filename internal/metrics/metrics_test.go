package metrics

import (
	"testing"
	"time"
)

func TestRegistrySnapshotContainsOnlyAggregateMetrics(t *testing.T) {
	registry := New()
	registry.RequestsTotal.Add(2)
	registry.RequestsInFlight.Add(1)
	registry.TokensEstimated.Add(42)
	registry.RequestLatency.Observe(12 * time.Millisecond)
	registry.UpstreamLatency.Observe(2 * time.Second)
	snapshot := registry.Snapshot()
	if snapshot["requests_total"] != uint64(2) || snapshot["tokens_estimated"] != uint64(42) {
		t.Fatalf("unexpected counters: %#v", snapshot)
	}
	if _, ok := snapshot["prompt"]; ok {
		t.Fatal("metrics snapshot must not expose prompts")
	}
	if _, ok := snapshot["cookie"]; ok {
		t.Fatal("metrics snapshot must not expose cookies")
	}
	latency, ok := snapshot["request_latency"].(map[string]any)
	if !ok || latency["count"] != uint64(1) {
		t.Fatalf("unexpected request latency: %#v", snapshot["request_latency"])
	}
}

func TestRegistryTracksFixedCardinalityRoutesWithoutSensitiveData(t *testing.T) {
	registry := New()
	registry.ObserveRoute(RouteOpenAIChatWebRPC)
	registry.ObserveRoute(RouteOpenAIChatWebRPC)
	registry.ObserveRoute(RouteGeminiDeveloperAPI)
	registry.ObserveRoute(Route(255))

	snapshot := registry.Snapshot()
	routes, ok := snapshot["routes"].(map[string]uint64)
	if !ok {
		t.Fatalf("routes snapshot = %#v, want map[string]uint64", snapshot["routes"])
	}
	if routes["openai_chat_web_rpc"] != 2 {
		t.Fatalf("OpenAI web-RPC route count = %d, want 2", routes["openai_chat_web_rpc"])
	}
	if routes["gemini_developer_api"] != 1 {
		t.Fatalf("Developer API route count = %d, want 1", routes["gemini_developer_api"])
	}
	if len(routes) != int(routeCount) {
		t.Fatalf("route cardinality = %d, want %d", len(routes), routeCount)
	}
	if _, ok := snapshot["prompt"]; ok {
		t.Fatal("route metrics must not expose prompts")
	}
	if _, ok := snapshot["cookie"]; ok {
		t.Fatal("route metrics must not expose cookies")
	}
}
