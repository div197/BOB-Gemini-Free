# CTO deep review — 2026-10-14

Scope: full repository review of `main` at `afe7881` (= `origin/main`), plus
five parallel read-only area reviews (updater and release trust, gateway HTTP
surface, Gemini upstream client, desktop/mobile/protocol, web studio). Nothing
was pushed, tagged, released, signed, or sent to a live provider.

Evidence labels used below:

- **V** — verified by the author in code, a test, or a command.
- **R** — reported by an area review with file:line evidence; not re-checked.
- **I** — inferred; needs a live or packaged-app check to confirm.

## 1. Repository state

- Local `main` was 330 commits behind `origin/main`; fast-forwarded to
  `afe7881` (V).
- No open PRs or issues. No remote branch has commits missing from `main` (V).
- The local-only `codex/public-readiness-audit` has no patch missing from
  `main` (`git cherry`) (V).
- The newest tag is `v0.2.0-preview.10`; six commits sit after it on `main`,
  all release-documentation or feed commits (V).
- `.github/` is empty: there is **no CI**. The local scripts and tests are the
  only gates (V).

## 2. Health at start

| Check | Result |
|---|---|
| `go test -count=1 ./...` | 1 failing test: `internal/updater` `TestSignedDesktopFeedSelectsLatestPreviewWithoutGitHubAPI` (V) |
| `go vet ./...` | clean (V) |
| `-race` on `gemini`, `server`, `updater` | clean once the updater fixture was fixed (V) |
| `gofmt -s -l .` | one whitespace-only drift in `internal/models/pricing.go` (V) |
| Tracked-file secret scan | only test fixtures; no cookie, config, or key files tracked (V) |
| `web/index.html` vs `playground.html` | byte-identical after version substitution (V) |

The updater failure was a time bomb: the fixture used fixed 2026-09 dates
while production checks expiry against the wall clock (V,
`internal/updater/update_feed.go:118,182`).

## 3. Fixed on `fix/updater-feed-test-clock` (local, not pushed)

| Commit | Change | Evidence |
|---|---|---|
| `53b164b` | Updater test derives its validity window from now | V — suite green |
| `5fe10aa` | gofmt alignment in `pricing.go` (whitespace only) | V — `git diff -w` empty |
| `b85ca2e` | **Stored XSS** in the studio: KaTeX fallback re-inserted captured model text unescaped; history replays run before the deferred KaTeX script (`playground.html:10200`, `loadHistory()` at `:12686`). Also bumps the hosted service-worker cache name (`web/sw.js:6`), which had been frozen since `b66a306`, and regenerates `web/index.html` | V — new test fails on old code |
| `bfd07c0` | **Anthropic image blocks always failed** (`format/anthropic.go` built `map[string]string`; extractor requires `map[string]any`) → "has no image URL". **Echoed `thinking`/`redacted_thinking` blocks failed the whole request** | V — both new tests fail on old code, pass on fix |

Full suite is green (18 packages) after all commits, and `git diff --check` is clean (V).

A follow-up commit `c81b93b` adds tests for a `[thinking, tool_use]` round trip with its `tool_result`, and for a thinking-only assistant turn. Both pass without a guard. (V)

Placeholder audit for the same bug class: `renderMarkdown` has two placeholder families. The `@@@BOB_CODE_n@@@` family is restored before `marked.parse` (`playground.html:10155`, `:10177`), so `marked` escapes the code. The `@@@BOB_MATH_n@@@` family is restored after `DOMPurify.sanitize` (`:10189`, `:10229`). That math restore was the only sibling, and `b85ca2e` fixes it. (V)

## 4. Open findings, by priority

### High

| ID | Finding | Evidence | Where |
|---|---|---|---|
| H1 | **Anthropic streaming with tools leaks the tool-call fence as text** and then appends a real `tool_use`. Claude Code always sends tools, so this is likely on every tool turn. Chat and Responses already buffer when tools are present. | V (structure) / I (client impact) | `internal/server/anthropic.go:130,237,290` |
| H2 | **Markdown-image exfiltration path.** CSP `default-src … https:` with `connect-src *`, so model-generated `<img>` can beacon to any HTTPS host, and injected script can send data anywhere. | V (CSP) / R (renderer) | `internal/server/playground.go:34`; image renderer in `playground.html` |
| H3 | **Docker exposes the gateway unauthenticated.** Dockerfile sets `HOST=0.0.0.0`; compose publishes `9610:9610`; nothing refuses a non-loopback bind when `api_keys` is empty; `config.example.json` ships the usable-looking key `sk-gemini`. | V | `Dockerfile:23`, `docker-compose.yml:8,11`, `config.example.json:12`, `main.go:568` |
| H4 | **Multi-account pool pairs the wrong token with the cookie.** The access token and guest cookies come from the cookie cache, which is seeded only from `cookie_file`; each request's `Cookie` header comes from a round-robin pool session. Pool-only configs get a guest token with a pool cookie. | V (both sides read: body token from `GetSessionInfo`, `client.go:457`; `Cookie`, `Authorization` and `X-Goog-AuthUser` from the pool session, `client.go:162-197`) / I (Google rejection) | `internal/gemini/client.go:155,457-470,614-638`; `auth.go` `GetSessionInfo` |
| H5 | **Update manifest does not bind the version.** `SHA256SUMS` is only `digest  name` lines. A validly signed old manifest reused under a newer tag installs the old build, and the staged app's own version is never checked. Needs release-write access, not the signing key. | V (format) / R (exploit path) | `internal/updater/release_manifest.go:83-86`, `desktop.go:231`, `desktop_stage.go` |
| H6 | **Production update feed has expired** (`expires_at` 2026-09-14). Desktop discovery now falls back silently to the GitHub API, which is the classroom burst the feed exists to prevent. Re-signing needs the owner's Keychain key. | V | `updates/desktop-feed.json`; `desktop.go:124-126`; `update_feed.go:182` |

### Medium

| ID | Finding | Evidence | Where |
|---|---|---|---|
| M1 | Developer API client has a hard-coded 30 s timeout that includes body reads; `request_timeout_sec` is ignored on this path. | V | `internal/server/server.go:57`; `internal/geminiapi` has no timeout use |
| M2 | HTTP server sets `ReadHeaderTimeout` and `WriteTimeout: 0` only; no `ReadTimeout` or `IdleTimeout`. | V | `main.go:568-571` |
| M3 | Empty `host` is accepted and becomes `:9610` (all interfaces). | V | `config.go:138-139`, `main.go:568` |
| M4 | Single-account setup: one failure puts the only session in a 60 s cooldown, so every request reports "unavailable". | V | `pool.go:316-321`; `client.go:767-779` |
| M5 | Image-fetch address filter lets through CGNAT (100.64/10), 198.18/15, Class E and NAT64 space. | V | `internal/multimodal/upload.go:327-329` |
| M6 | Process-level Developer key makes `/v1/messages`, `/v1/responses` and `/v1/images/generations` return 400 for clients that do not send the Studio-only route header. | V | `anthropic.go:15`, `responses.go:15`, `images.go:17`, `gemini_api.go:27-38` |
| M7 | Saving `./cookie.txt` runs `os.Chmod(".", 0700)` on the current directory. | R | `internal/gemini/auth.go:416`, `main.go:75` |
| M8 | Login window: the debugging port has no explicit bind or auth; the temporary profile can survive Ctrl-C because there is no signal handler during `--login`. | R / I | `internal/browser/browser.go:155-182`; `main.go:499` |
| M9 | Mobile package: default config has no API key, and its HTTP server has no timeouts; any co-resident app can use the session. | R | `pkg/mobile/mobile.go:108,164` |
| M10 | Diagnostics attaches the first configured API key to any `--test-url`, including plain HTTP. | R | `main.go:489-492` |
| M11 | Unknown explicit models silently reroute to the default model (`Resolve`), while `/v1/models/{id}` returns 404 for the same name. | R | `chat.go:48`, `anthropic.go:47`, `responses.go:36`, `google.go:81`; `models.go:322-325` |
| M12 | Error envelopes deviate from AGENTS.md §4: several have no `type`; an oversize body returns 400 instead of 413; a bad `@think` returns `not_found_error`. | R | `chat.go`, `responses.go`, `google.go`, `anthropic.go:49-52`, `helpers.go:68` |
| M13 | Responses API: flat forced `tool_choice` returns 400; `web_search` and `local_shell` tools return 400; `reasoning` input items return 400. Untested. | R | `format/openai.go:73-76`, `server/responses.go:70-82`, `format/responses.go:112-114` |
| M14 | Parsed tool calls are not checked against declared tools or forced `tool_choice`, so an echoed fence becomes a real call. | R | `format/openai.go:416`, `format/google.go:345` |
| M15 | Windows CLI and desktop builds share the asset name `bob-gemini-free-windows-amd64.exe`. | V (names) / I (impact) | `scripts/release-local.sh:43-46`; `desktop.go:524` |
| M16 | Studio: DOMPurify fails open if its script does not load; Mermaid runs with `securityLevel: 'loose'`; gateway JSON is inserted into `innerHTML` unescaped. | R / I | `playground.html:10187-10188`, `:8177`, `:7606` |
| M17 | Hosted service worker caches gateway GETs (`/`, `/health`, `/healthz`) cache-first on remote hosts. | R | `web/sw.js:40-78` |

### Low (selected)

- `thinking.go:61`: a leading newline before a thought fence makes the header leak into reasoning (R).
- `cmd/desktop/gateway.go:37-47`: port reuse adopts any 127.0.0.1 responder with matching headers, even one bound to 0.0.0.0 (R).
- `refiner.go:40` labels nanoseconds as `duration_ms` (R).
- Login cookie filter uses `Contains(domain, "google.com")`, which matches look-alike domains (R).
- `openai-processing-ms` is set after the body is written, so it is never sent (R).
- Studio CSP allows `unsafe-inline` and `unsafe-eval`; this matters because of H2 (R).
- GET endpoints under DNS rebinding: not re-checked (gap).

## 5. Reported sound by area reviews (not re-checked)

- Protected routes: every non-public route requires the key; all key comparisons are constant-time (V).
- `?key=` only with `allow_query_api_key` (R).
- Developer key is sent only in the `x-goog-api-key` header to the fixed Google host, never persisted, never logged (R).
- Cookie values and SAPISID never appear in logs or errors; cookie files are 0600 and directories 0700 (R).
- Retries never replay a generation POST; only pre-connect dial and lookup errors are retried (R).
- Stream parser caps: 16 MiB per line, 32 MiB per stream (R).
- Image fetch: DNS re-resolved and connected to the checked IP, no redirects, 20 MiB cap (R).
- Updater: signature checked over exact bytes before parsing; zip extraction rejects symlinks, traversal, and duplicates; channel rules and semver ordering correct (R).
- Studio: no `allow-same-origin` on either sandbox; every CDN tag pinned with SRI (R).
- Desktop bridge exposes only `Greet` and `GatewayURL`; listener fixed to loopback (R).

## 6. Decisions needed from the owner

1. **Push and PR.** The four commits are local on `fix/updater-feed-test-clock`. Say whether to push and open a PR.
2. **Re-sign the update feed (H6).** Run `scripts/sign-update-feed.sh` with your Keychain key. Decide whether to automate it, given the 14-day maximum lifetime.
3. **Docker default (H3).** Recommended: refuse to start on a non-loopback bind unless `api_keys` is set, and stop shipping `sk-gemini` in the example config.
4. **Unknown explicit models (M11).** Fail closed, or keep the current fallback for compatibility?
5. **CI.** `docs/engineering/VERIFICATION-MATRIX.md` lists "no GitHub Actions" as a boundary. Confirm whether a minimal gate (`go test -race`, `vet`, `gofmt`) is now acceptable.
6. **Multi-account pool (H4).** Confirm multi-account rotation is a supported feature. The fix needs a live test with an authorized account.

## 7. Development plan

**P1 — correctness and exposure (small, testable, next):**
H1 (buffer Anthropic streaming when tools are present), H2 (restrict `img-src`
and markdown image hosts), H3 (bind guard and example key), M1 and M2 (timeouts),
M3 (host validation), M4 (single-session cooldown), M5 (address ranges), M7
(no chmod on the current directory), M10 (key only to loopback or HTTPS).

**P2 — architecture (needs design and a live check):**
H4 (token derived per selected session), H5 (version binding and staged-app
version check, with a compatibility plan for published manifests), M15 (asset
naming).

**P3 — compatibility and hardening:**
M6, M11, M12, M13, M14, M16, M17, M8, M9.

**P4 — operations and docs:**
Feed re-sign automation and expiry alert (H6); CI gate (decision 5); slim
`CLAUDE.md` (about 52 KB, loaded every session, with stale §25–26); refresh
`AGENTS.md` (broken code fence, lists 8 of 13 internal packages, claims
"minimal dependencies" despite Wails and tls-client).

## 8. Proof boundary

- No live Google request, real cookie, or authenticated Pro/image route was exercised (V).
- No release, signing, tag, or push was performed (V).
- Not verified: Claude Code and Codex client behaviour (thinking, tools, `local_shell`); Windows and Linux desktop IPC; Mermaid and DOMPurify runtime behaviour; Docker runtime; Google's rejection of mismatched tokens (H4).
- Dependency currency was not checked (no network).
