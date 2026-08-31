# Local-Only Benchmark — 2026-08-31

**Purpose:** current release-preparation baseline for local gateway parsing,
formatting, and HTTP concurrency only. This run did not contact Google,
GitHub, or any student session.

**Environment:** macOS 26.2 (build 25C56), Apple Silicon
(`darwin/arm64`), Go 1.26.6, source snapshot `2d42d44` (documentation-only
PR #92 later advanced `main` to `67e5337`).
**Command:** `go run ./cmd/benchmark-local -profiles 1,10,20,30 -requests 100`
**Upstream:** deterministic in-process requester.

| Concurrency | Requests | Failed | P50 | P90 | P95 | P99 | RPS | Allocs/request | RSS | Goroutines | Max connections |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 1 | 100 | 0 | 0.161 ms | 0.210 ms | 0.230 ms | 0.502 ms | 5,230 | 370.49 | 20.1 MiB | 5 | 1 |
| 10 | 100 | 0 | 0.579 ms | 1.032 ms | 1.138 ms | 1.364 ms | 15,157 | 269.25 | 21.2 MiB | 8 | 10 |
| 20 | 100 | 0 | 0.990 ms | 2.311 ms | 2.701 ms | 3.101 ms | 13,197 | 287.09 | 22.1 MiB | 16 | 21 |
| 30 | 100 | 0 | 1.160 ms | 2.822 ms | 3.027 ms | 3.467 ms | 18,171 | 260.91 | 23.3 MiB | 24 | 30 |

All 400 local requests completed successfully. RSS is the process-level
measurement reported after each sequential profile, and allocation counts
include the local HTTP fixture. These values are not Google latency, quota,
model throughput, rate-limit, shared-IP, or 20–30-device provider-capacity
claims. Repeat with the same host profile after protocol, stream, concurrency,
or release-build changes.

The next gate is an authorized, staggered provider pilot on one to three Macs,
not a simultaneous Google burst. See
[`PREVIEW-ROLLOUT-VALIDATION.md`](PREVIEW-ROLLOUT-VALIDATION.md).
