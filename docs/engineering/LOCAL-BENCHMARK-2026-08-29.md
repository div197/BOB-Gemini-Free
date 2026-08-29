# Local Benchmark — 2026-08-29

**Purpose:** release-preparation baseline for local gateway parsing and
formatting only. This run did not contact Google, GitHub, or any student
session.

**Environment:** macOS 26.2, Apple Silicon (`darwin/arm64`), Go 1.26.6,
`go run ./cmd/benchmark-local -profiles 1,10,20,30 -requests 100`.

| Concurrency | Requests | Failed | P50 | P90 | P95 | P99 | RPS | Allocs/request | RSS | Goroutines | Max connections |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 1 | 100 | 0 | 0.137 ms | 0.185 ms | 0.228 ms | 0.602 ms | 5,787 | 370.98 | 20.0 MiB | 4 | 1 |
| 10 | 100 | 0 | 0.537 ms | 1.008 ms | 1.274 ms | 1.615 ms | 15,011 | 266.04 | 21.6 MiB | 10 | 10 |
| 20 | 100 | 0 | 0.991 ms | 2.183 ms | 2.336 ms | 3.134 ms | 16,497 | 259.60 | 22.3 MiB | 15 | 20 |
| 30 | 100 | 0 | 1.312 ms | 4.405 ms | 4.790 ms | 5.718 ms | 14,641 | 268.69 | 23.1 MiB | 26 | 33 |

These figures are a local synthetic baseline, not Google latency, quota,
throughput, RAM, or provider-capacity guarantees. Repeat the run after any
protocol, stream, concurrency, or release-build change and compare using the
same host profile where possible. The benchmark validates usable JSON response
content and bounds both concurrency and request count at 128 workers and
10,000 requests.
