# Local Benchmark — 2026-08-29

**Purpose:** release-preparation baseline for local gateway parsing and
formatting only. This run did not contact Google, GitHub, or any student
session.

**Environment:** macOS 26.2, Apple Silicon (`darwin/arm64`), Go 1.26.6,
`go run ./cmd/benchmark-local -profiles 1,10,20,30 -requests 100`.

| Concurrency | Requests | Failed | P50 | P90 | P95 | P99 | RPS | Allocs/request | RSS | Goroutines | Max connections |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 1 | 100 | 0 | 0.139 ms | 0.238 ms | 0.306 ms | 0.757 ms | 5,435 | 363.78 | 20.5 MiB | 5 | 1 |
| 10 | 100 | 0 | 0.581 ms | 0.927 ms | 1.051 ms | 1.396 ms | 16,018 | 260.13 | 21.6 MiB | 11 | 10 |
| 20 | 100 | 0 | 1.230 ms | 2.357 ms | 2.926 ms | 3.035 ms | 14,751 | 263.86 | 21.9 MiB | 17 | 21 |
| 30 | 100 | 0 | 1.205 ms | 4.004 ms | 4.474 ms | 4.762 ms | 15,618 | 283.24 | 23.1 MiB | 35 | 41 |

These figures are a local synthetic baseline, not Google latency, quota,
throughput, RAM, or provider-capacity guarantees. Repeat the run after any
protocol, stream, concurrency, or release-build change and compare using the
same host profile where possible.
