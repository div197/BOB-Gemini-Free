# Local-Only Benchmark — 20–30 Client Profiles

**Run date:** 2026-08-25 (Asia/Kolkata)
**Command:** `go run ./cmd/benchmark-local -profiles 1,10,20,30 -requests 100`
**Environment:** `go1.26.6 darwin/arm64`
**Upstream:** deterministic in-process requester; no Google or external
network request was used.

| Concurrency | Success | P50 | P90 | P95 | P99 | Avg | Throughput | Allocs/request | RSS reported | Goroutines | Max connections |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 1 | 100/100 | 0.131 ms | 0.192 ms | 0.211 ms | 0.549 ms | 0.156 ms | 6,221 req/s | 330.11 | 19.7 MiB | 5 | 1 |
| 10 | 100/100 | 0.662 ms | 1.080 ms | 1.302 ms | 1.645 ms | 0.689 ms | 13,562 req/s | 336.05 | 21.6 MiB | 16 | 10 |
| 20 | 100/100 | 1.282 ms | 2.352 ms | 2.557 ms | 3.294 ms | 1.422 ms | 12,724 req/s | 342.55 | 22.3 MiB | 17 | 20 |
| 30 | 100/100 | 1.553 ms | 3.643 ms | 4.233 ms | 4.788 ms | 1.862 ms | 14,182 req/s | 344.59 | 23.5 MiB | 28 | 30 |

All profiles completed with zero local failures. These values are a local
gateway baseline only; they are not Google latency, quota, model throughput,
or shared-IP capacity measurements. RSS is a process-level maximum while the
profiles run sequentially. Allocation counts include the local HTTP fixture.

The next validation gate is a small authorized provider pilot, not a 30-way
Google burst. See [`PREVIEW-ROLLOUT-VALIDATION.md`](PREVIEW-ROLLOUT-VALIDATION.md).
