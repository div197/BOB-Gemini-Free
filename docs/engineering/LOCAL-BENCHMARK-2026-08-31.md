# Local-Only Benchmark — 2026-08-31

**Purpose:** current release-preparation baseline for local gateway parsing,
formatting, and HTTP concurrency only. This run did not contact Google,
GitHub, or any student session.

**Environment:** macOS 26.2 (build 25C56), Apple Silicon
(`darwin/arm64`), Go 1.26.6, source checkout `dffbd1b`.
**Command:** `go run ./cmd/benchmark-local -profiles 1,10,20,30 -requests 100`
**Upstream:** deterministic in-process requester.

| Concurrency | Requests | Failed | P50 | P90 | P95 | P99 | RPS | Allocs/request | RSS | Goroutines | Max connections |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 1 | 100 | 0 | 0.134 ms | 0.175 ms | 0.214 ms | 0.471 ms | 5,784 | 371.30 | 20.3 MiB | 5 | 1 |
| 10 | 100 | 0 | 0.529 ms | 1.089 ms | 1.784 ms | 1.837 ms | 15,298 | 269.93 | 22.1 MiB | 9 | 10 |
| 20 | 100 | 0 | 1.017 ms | 2.074 ms | 2.465 ms | 2.933 ms | 15,821 | 269.62 | 22.7 MiB | 17 | 20 |
| 30 | 100 | 0 | 1.137 ms | 2.353 ms | 3.181 ms | 3.366 ms | 20,337 | 259.06 | 23.8 MiB | 24 | 33 |

All 400 local requests completed successfully. RSS is the process-level
measurement reported after each sequential profile, and allocation counts
include the local HTTP fixture. These values are not Google latency, quota,
model throughput, rate-limit, shared-IP, or 20–30-device provider-capacity
claims. Repeat with the same host profile after protocol, stream, concurrency,
or release-build changes.

The next gate is an authorized, staggered provider pilot on one to three Macs,
not a simultaneous Google burst. See
[`PREVIEW-ROLLOUT-VALIDATION.md`](PREVIEW-ROLLOUT-VALIDATION.md).
