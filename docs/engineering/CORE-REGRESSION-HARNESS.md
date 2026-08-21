# Core Regression Harness — Mission 1

**Date:** 2026-08-21 (Asia/Kolkata)

## Scope

Mission 1 adds deterministic, account-free regression coverage around the
reverse-engineered protocol core. The fixtures are synthetic protocol-shaped
fixtures built from the current parser contract; no Google account or live
provider was available in this workspace to capture authoritative upstream
traffic. They must therefore not be described as live compatibility proof.

The new coverage is distributed across:

| Area | Regression coverage |
|---|---|
| Sparse payload | Form decoding, exact 102-slot length, prompt, language, thinking, model, extra slots, UUID shape, and image references in `internal/gemini/golden_test.go` |
| Response parsing | Cumulative text, multiline Unicode/Indic output, citations, code/artifact sanitization, malformed nested JSON, and Bard errors |
| Streaming | Arbitrary byte boundaries, repeated cumulative frames, malformed lines, Bard errors, truncated final frames, and retry deduplication |
| Authentication | Deterministic SAPISIDHASH digest at an injected timestamp |
| Thinking | One-byte opening/closing fence fragmentation, Unicode thinking, immediate content, and missing closing fence behavior in `internal/format/golden_test.go` |
| Tool extraction | Nested JSON, arrays, nullable values, Unicode arguments, multiple calls, invalid JSON, accidental Markdown, and choice modes |
| Protocol adapters | Equivalent semantic text preservation through OpenAI, Anthropic, and Google formatting paths |
| Multimodal | Two-step resumable upload headers, endpoint handoff, auth-user routing, and exact uploaded bytes |
| Models | Existing deterministic catalog/resolution tests remain the source of truth for aliases, modes, thinking, and sparse extras |

## Protected-core changes justified by failing tests

The protected files were left unchanged until the Mission 1 tests existed.
The following surgical changes are now justified by those tests:

1. `internal/gemini/auth.go` adds `SAPISIDHashAt(sapisid, timestamp)` and
   makes the existing `SAPISIDHash` delegate to it. Production behavior is
   unchanged because the wrapper still supplies `time.Now().Unix()`; the
   timestamped primitive makes the exact SHA-1 input and output testable.
2. `internal/gemini/stream.go` adds `StreamParser.Flush`, which turns a
   buffered unterminated record into one final line. `internal/gemini/client.go`
   invokes it when `ReadBytes` reaches EOF. Before this change, a valid final
   record without a trailing newline was silently discarded; afterward it is
   parsed and emitted once.

The invariants protecting these changes are the sparse payload positions,
SAPISIDHASH digest format, cumulative-prefix deduplication, malformed-line
resilience, Bard-error propagation, and final-frame delivery. They are
asserted by `internal/gemini/golden_test.go` and the existing focused tests.

## Remaining evidence boundary

- The suite does not prove that Google still accepts every sparse-array field.
- The suite does not prove native tool calling; current tool extraction remains
  prompt/Markdown emulation and is classified separately in the verification
  matrix.
- Real authenticated Scotty upload, live Gemini streaming, provider rate
  limits, model identity, and browser-served behavior remain live gates.
- A future authorized capture should add immutable raw fixtures under a
  dedicated `testdata/` directory and compare them against these synthetic
  fixtures without replacing the deterministic tests.
