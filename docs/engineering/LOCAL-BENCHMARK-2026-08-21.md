# Local-Only Benchmark Baseline

**Run date:** 2026-08-21 (Asia/Kolkata)  
**Command:** `go run ./cmd/benchmark-local -requests 100`  
**Environment:** `go1.26.5 darwin/arm64`  
**Upstream:** deterministic in-process Gemini requester; no Google/network
request was used for the benchmark workload.

| Concurrency | Success | P50 | P90 | P95 | P99 | Avg | Throughput | Allocs/request | RSS reported | Goroutines | Max connections |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 1 | 100/100 | 0.124 ms | 0.181 ms | 0.223 ms | 0.457 ms | 0.151 ms | 6,381 req/s | 330.57 | 20.1 MiB | 6 | 1 |
| 10 | 100/100 | 0.785 ms | 1.509 ms | 1.885 ms | 2.093 ms | 0.849 ms | 10,954 req/s | 338.29 | 21.9 MiB | 11 | 10 |
| 50 | 100/100 | 3.149 ms | 6.455 ms | 6.666 ms | 7.004 ms | 3.263 ms | 13,392 req/s | 369.41 | 24.7 MiB | 52 | 65 |
| 100 | 100/100 | 7.872 ms | 9.871 ms | 10.120 ms | 10.147 ms | 7.853 ms | 9,662 req/s | 388.26 | 28.1 MiB | 70 | 100 |

RSS is the process `getrusage` maximum and is reported as an aggregate
process value while profiles run sequentially; it is not a per-request
allocation or a clean-room binary RSS measurement. Allocation counts are
runtime allocation deltas divided by the requested count and include the
local HTTP server/client fixture. Max connections come from a tracked local
dialer and are not Google connection limits.

This baseline must not be quoted as live-provider latency, model latency,
rate-limit capacity, or proof of the README's older performance claims. A
separate explicitly authorized live profile is still required for upstream
behavior and Google rate-limit observations.
