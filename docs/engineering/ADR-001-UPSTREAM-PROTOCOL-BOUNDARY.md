# ADR-001: Defer a Pluggable Upstream Protocol Interface

**Status:** Accepted — defer implementation

**Date:** 2026-08-21

## Context

The Google Gemini web RPC format is undocumented and volatile. A proposed
boundary was:

```go
type UpstreamProtocol interface {
    BuildRequest(...)
    ParseResponse(...)
    ParseStream(...)
    Capabilities(...)
}
```

The current repository has one concrete upstream protocol and already keeps
payload construction, response parsing, stream retry state, authentication,
and formatting in separate packages. Mission 1 added deterministic fixtures
around those concrete seams.

## Evaluation

| Question | Finding |
|---|---|
| Does it isolate Google schema volatility? | Partly. Payload/parser files already isolate most volatility; an interface would not remove the sparse-array knowledge. |
| Does it simplify fixture testing? | No immediate evidence. Fixtures now call the concrete pure helpers directly, which makes wire invariants visible. |
| Does it reduce coupling? | Not yet. The current client, cookie pool, retry parser, and model resolver share Google-specific assumptions that the proposed four methods do not capture. |
| Would it create unnecessary abstraction? | Yes, with only one implementation and no alternate provider contract. It would add factories, error translation, and capability semantics before a second implementation exists. |
| Can it be introduced without changing wire behavior? | Technically possible, but the refactor would touch protected protocol code without a measured compatibility benefit. |

## Decision

Do not implement `UpstreamProtocol` in Phase II. Keep the concrete Google
seams and strengthen them with fixture tests, explicit invariants, and
documentation. Revisit the interface only when one of these conditions is
met:

1. a second upstream protocol is actually implemented;
2. the same client must support provider-specific capabilities without
   branching through Google-specific fields; or
3. a measured refactor demonstrates reduced test coupling or wire-risk.

This preserves behavior and avoids architecture astronautics while still
acknowledging the protocol's volatility.
