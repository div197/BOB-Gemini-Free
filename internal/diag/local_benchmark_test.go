package diag

import "testing"

func TestNormalizeLocalBenchmarkSettings(t *testing.T) {
	if gotConcurrency, gotRequests := normalizeLocalBenchmarkSettings(0, 0); gotConcurrency != 1 || gotRequests != 100 {
		t.Fatalf("defaults = %d/%d, want 1/100", gotConcurrency, gotRequests)
	}
	if gotConcurrency, gotRequests := normalizeLocalBenchmarkSettings(maxLocalBenchmarkConcurrency+1, maxLocalBenchmarkRequests+1); gotConcurrency != maxLocalBenchmarkConcurrency || gotRequests != maxLocalBenchmarkRequests {
		t.Fatalf("caps = %d/%d, want %d/%d", gotConcurrency, gotRequests, maxLocalBenchmarkConcurrency, maxLocalBenchmarkRequests)
	}
}

func TestRunLocalBenchmarkUsesMockedUpstream(t *testing.T) {
	report := RunLocalBenchmark(2, 8)
	if report.Successful != 8 || report.Failed != 0 {
		t.Fatalf("local benchmark result = %+v", report)
	}
	if report.P50Latency <= 0 || report.P99Latency < report.P50Latency {
		t.Fatalf("invalid latency quantiles = %+v", report)
	}
	if report.AllocsPerRequest <= 0 || report.MaxConnections <= 0 {
		t.Fatalf("missing allocation/connection measurements = %+v", report)
	}
}
