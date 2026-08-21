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
