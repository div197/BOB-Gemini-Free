package diag

import "testing"

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
