# Tool-Calling Fidelity Lab — Mission 5

## Measurement boundary

The current Google web-RPC payload builder does not place OpenAI, Anthropic,
or Google function declarations into a verified native upstream function-call
field. The adapters instead serialize declarations into prompts and parse
Markdown-like model output. That is useful compatibility behavior, but it is
not native tool calling and cannot be labeled “full native compatibility.”

The torture fixtures in `internal/format/golden_test.go` and the existing
format tests cover nested schemas, enums, arrays, nullable values, Unicode,
large arguments, multiple/parallel calls, malformed output, accidental
Markdown, tool choices, tool results, and semantic adapter conversion.

## Current classification

| Surface or behavior | Classification | Evidence and limitation |
|---|---|---|
| OpenAI tool declaration transport | EMULATED_PARTIALLY | `internal/format/openai.go:45-176` injects JSON definitions into the prompt; no native Google declaration field is proven |
| Anthropic `tool_use` / `tool_result` conversion | EMULATED_PARTIALLY | `internal/format/anthropic.go:11-170` maps request structures, but generation still flows through prompt extraction |
| Google function declarations | EMULATED_PARTIALLY | `internal/format/google.go:19-146` serializes declarations and function responses as text |
| Nested object/array/enum/nullable schemas | EMULATED_RELIABLY for serialization | Deterministic prompt and JSON-parser fixtures pass; upstream model compliance remains unknown. The explicit Gemini Developer API tool-call boundary is stricter: its documented `FunctionCall.args` field is a JSON object, so scalar/array/null arguments are rejected rather than silently coerced or omitted. See the [GenerateContent schema](https://ai.google.dev/api/generate-content). |
| Unicode and large JSON arguments | EMULATED_RELIABLY for parser fixtures | Parser preserves valid JSON in tested bounds; no provider maximum has been measured |
| Multiple or parallel tool requests | EMULATED_PARTIALLY | Multiple fenced blocks parse; execution/parallel scheduling is outside this gateway |
| `none` choice | EMULATED_RELIABLY as an instruction | The model receives a normalized prohibition; malformed values are rejected, but native enforcement is absent |
| `auto`, `required`, and specific tool choice | EMULATED_PARTIALLY | Choices are normalized and named tools must be declared; prompt constraints still do not enforce model obedience or exact selection through an upstream schema |
| Tool-result continuation | EMULATED_PARTIALLY | Results are represented in subsequent prompt text; no live multi-turn tool executor exists in the gateway |
| Tool-call streaming | EMULATED_PARTIALLY | `internal/server/chat.go:74-208` buffers tool requests and replays a response instead of streaming native tool deltas |
| Native function calling through Google web RPC | UNKNOWN / not evidenced | Requires a captured accepted native request/response pair and a live differential test |

## Decision

Request-side tool-choice validation is now shared by the prompt adapter and the
direct Developer API route. Invalid modes, malformed named choices, and names
not present in the request's declarations fail before an upstream call. The
Anthropic adapter normalizes `auto`, `none`, `any`, and named `tool` choices to
the shared representation; `disable_parallel_tool_use: true` remains an
explicitly unsupported semantic rather than being silently ignored. This
reduces false-success behavior. The Google-shaped adapter likewise normalizes
`AUTO`/`NONE`/`ANY` and rejects unknown modes or incompatible allowed names, but
these controls still do not make model selection native.

Assistant tool-call prompt blocks are JSON-encoded rather than assembled by
interpolating names. This keeps bounded quotes, backslashes, and newlines from
corrupting the adapter's own control document; it is a serialization safeguard,
not a trust boundary for model-generated tool requests.

The direct Gemini Developer API translator now fails closed when an incoming
OpenAI-shaped tool call uses scalar, array, or `null` JSON arguments. The
previous map-only unmarshal accepted `null` as an absent map and returned a
different error for other non-object values, making the boundary inconsistent.
This is an intentional target-schema constraint, not a claim that OpenAI's
prompt adapter or the web-RPC emulation has native non-object function-call
support.

Do not rewrite tool calling in Phase II. The lab now makes the actual behavior
executable and prevents accidental “native/full” wording. A future native
implementation requires source or live evidence of the upstream schema,
fixture parity across all three adapters, and an explicit execution/security
contract for returned tools.
