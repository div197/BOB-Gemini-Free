# BOB Gemini Free — Phase IV Hardening and Release-Readiness Report

**Date:** 2026-08-31 (Asia/Kolkata)
**Workspace:** `/Users/apple31/Documents/BOB-Gemini-Free`
**Base:** `origin/main` `523ceeb`
**Reviewed continuation:** `codex/release-readiness-v0.2.0` through `4666dba`
**Operating rule:** local verification only; no GitHub Actions and no
provider, cookie, PAT, or private release-key material used.

## Executive decision

This continuation materially reduces three real failure classes: coalesced
stream cancellation and silent subscriber loss, memory exposure through highly
compressible remote images, and late updater failure when the app is running
from a read-only or translocated location. The source and deterministic test
gates are green.

BOB is not yet a universally verified student release. A real browser at the
required viewports, a clean `/Applications` update and rollback, signed public
release bytes, Apple/Windows platform trust, live Google behavior, and a
30-device pilot remain external gates. The correct product label remains
**controlled preview / release candidate**, not unattended production.

## Product intent and taste baseline

BOB is a local-first AI workbench for students and developers who want to ask,
learn, inspect reasoning, generate code, and run small artifacts through a
local gateway. Its character should be a calm, premium engineering instrument
with a teacher's bench: precise, trustworthy, educational, and quietly
playful when an artifact earns that moment.

The design audit therefore treats the prompt, response, and next action as the
focal point. Amber is the purposeful action color; dark neutral surfaces carry
the reading task; advanced controls should recede behind progressive
disclosure; neon, emoji, gradients, and motion are reserved for state or
artifact meaning. Source-level accessibility work is complete for the audited
surface, but the visual/browser ceiling is not certified until a real browser
walk is available.

## Baseline before this continuation

The prior reviewed tip was `700d343`. It already contained the native-control
and drawer-semantics pass, the 100-path failure register, PWA route protection,
signed updater transactions, provider-route guards, and source/static parity.
The continuation brief's earlier PWA-404 observation was stale against that
source: `/manifest.json`, `/sw.js`, and `/favicon.ico` were already embedded,
routed, and tested, so no redundant PWA rewrite was made.

The remaining high-value source risks were:

1. a participant cancellation could interact incorrectly with a shared stream;
2. a slow stream subscriber needed an explicit bounded failure rather than a
   silent transcript gap;
3. remote image bytes were MIME-checked but needed decode-dimension validation
   before they reached downstream image paths;
4. the updater could offer a valid update and only later reveal that a DMG or
   App Translocation location could not be staged;
5. the actual installed Preview 7 → bridge → stable device path remained
   unproven and could not be fabricated with a fixture.

## Work completed

### 1. Coalesced stream lifecycle — `7ee4a23`, `1f47a29`, `5eae3e2`

`internal/gemini/flight.go` now gives each subscriber an independent context
and bounded queue. A leader's cancellation or deadline detaches only the
leader. A follower can leave without terminating an independent participant.
The shared runner is cancelled when the last subscriber leaves, and it remains
bounded by the Gemini client's total request timeout. A full subscriber queue
returns `ErrStreamSubscriberTooSlow`; it never silently drops a delta.

The history path preserves the registration/replay boundary, caps retained
chunks and bytes, refuses a new subscriber after history truncation, and
removes completed or abandoned flights. The slow-subscriber regression was
also made deterministic in two layers: it waits for both leader and follower
before the burst, then paces each burst item on healthy-leader consumption.
This tests follower overflow rather than depending on the Go scheduler to
schedule a fast leader during an instantaneous burst.

Protected behavior includes:

- leader cancellation with a healthy follower;
- leader deadline with a healthy follower;
- follower cancellation with a healthy leader;
- last-subscriber cancellation of the shared runner;
- explicit slow-subscriber overflow;
- history-limit refusal;
- concurrent/race-detector cleanup.

### 2. Remote image decode boundary — `7f59bbc`

`internal/multimodal/upload.go` now validates fetched bytes with the same image
decode budget used by downstream image handling, not only the response MIME
header. This prevents a tiny, highly-compressible encoded image with excessive
dimensions from bypassing the source-dimension/pixel budget before upload or
Developer API base64 processing.

The change is intentionally narrow. It does not enable new image formats, make
OCR automatic, or claim live Google vision capability. Browser previews,
OCR CPU limits, and provider session requirements remain separately documented
boundaries. The local image-reference cache now also applies a conservative
15-minute age limit before reusing a Scotty reference; the provider's actual
expiry remains unknown, so this is a stale-reference guard rather than a
provider-lifetime claim.

### 3. Updater install-location preflight — `8651eba`

`internal/updater/desktop_stage.go` now exposes a no-network local preflight.
It validates the current install target, rejects macOS App Translocation, and
probes the same-filesystem parent for a temporary staging directory before an
artifact download begins. `StageDesktopUpdate` repeats the check immediately
before staging because permissions can change after discovery.

`cmd/desktop/updates.go` now defers automatic prompts when this local boundary
is not ready and gives an explicit manual-check error with Applications or
writable-location recovery guidance. The existing signed manifest, exact
asset, code-signature, helper, confirmation, rollback, and recovery gates are
unchanged.

The updater remains user-consented. A background check may discover and show a
candidate; it never installs without **Install Update**. No silent fleet push
was added.

### 4. Documentation and evidence — `1cc33d5`, `5eae3e2`

The design audit, release readiness document, desktop update operations, 100-
path register, and verification matrix now record the continuation truth:

- the stream and multimodal guards are source/test protected;
- updater preflight is protected locally but installed-bundle acceptance is
  external;
- older commit labels are not used as current-tip evidence;
- no browser claim is upgraded without a real browser runtime;
- no GitHub Actions workflow is required or invoked.

### 5. Session-bound image-reference age guard — `4a0f15d`

The bounded Scotty reference cache now records insertion time and expires a
local reference after a conservative 15-minute age. An expired entry is
removed before lookup and causes a fresh upload; it is never silently reused.
The cache remains scoped to a single authenticated cookie source and bounded
by entry count. The provider's actual reference lifetime is undocumented, so
the local age is explicitly a stale-reference guard rather than a claim about
Google's TTL. Clock-injected tests cover both expiry and refresh.

### 6. Local history persistence failure visibility — `838b066`

The Web Studio now routes conversation writes and new-chat deletion through
guarded storage helpers. Normal writes remain silent, while compacted writes,
corrupt/oversized history recovery, unavailable storage, and a complete write
failure produce an accessible status message near the composer. The message
does not expose the storage exception or conversation content; it tells the
student when attachment previews may be omitted and when to export work before
closing. This is a user-facing recovery boundary, not a claim that browser
`localStorage` has a uniform quota or that long-session CPU behavior is
proven across devices.

### 7. Developer API stream-error classification — `888a7fa`

The typed Developer API stream now recognizes a provider error envelope even
when Google returns it inside an HTTP-200 SSE event. Numeric and string status
codes are mapped to quota/auth classifications where known, the provider
message is sanitized, and the caller's API key is redacted. Unknown SSE fields,
named events, comments, and ordered multiline `data:` payloads remain
non-destructive and are covered by fixtures. This improves diagnosis without
claiming that every future provider event type is known or that the
reverse-engineered web-RPC stream has public-SSE semantics.

### 8. Session and quota failure classification — `e478874`

The web-RPC client now labels HTTP 401/403 responses as authentication/session
rejections and HTTP 429 as quota, while preserving the existing no-replay rule
for provider policy and quota responses. The Studio distinguishes these
provider failures from BOB's own optional gateway API-key protection, avoiding
the misleading instruction to enter a local key when Google's session has
expired or access has been denied. This is actionable error classification,
not automatic reauthentication or quota circumvention.

### 9. Anthropic SSE lifecycle write integrity — `4666dba`

The Anthropic adapter now treats a failed lifecycle write as terminal. Initial
stream setup, block flush/stop, tool block events, `message_delta`, and
`message_stop` are no longer allowed to be silently skipped while the handler
continues toward a success response. Deterministic fixtures cover ordered
success lifecycle and upstream-error termination; complete Claude Code event
parity and real client compatibility remain open.

## Verification completed

| Gate | Result | Evidence boundary |
|---|---|---|
| `go test -count=1 ./...` | PASS | All Go packages passed. |
| `go test -race -count=1 ./...` | PASS | Full race suite passed after the deterministic slow-subscriber test fix. |
| Repeated slow-subscriber race test | PASS | `-count=20` plus the full `internal/gemini` race package passed. |
| `go vet ./...` | PASS | No diagnostics. |
| `go build ./...` | PASS | Source builds on the audit host. |
| `go mod verify` | PASS | Module contents verified. |
| `make web` | PASS | Generated static bundle remained synchronized. |
| `git diff --check` | PASS | No whitespace errors. |
| `bash -n scripts/*.sh` | PASS | Local release scripts parse successfully. |
| `scripts/verify-release-source.sh v0.2.0` | PASS | Clean source, key/version coherence, installer trust-anchor, and web parity passed at the packaging source commit. |
| macOS Preview package | PASS locally | Universal Wails app, ZIP, DMG, visible `/Applications` shortcut, release notice, and ad-hoc `codesign --verify` passed. The package was not published and its local checksum manifest was not signed. |
| Browser desktop/tablet/phone walk | NOT AVAILABLE | The configured browser runtime reported no available browser; source tests are not rendered interaction proof. |
| Live Google/provider run | NOT RUN | No provider credential or cookie was used. |
| Clean installed-bundle update/rollback | NOT PROVEN | Requires the owner-controlled `/Applications` device and exact signed public assets. |

## Correctness and release table

| Surface | Current status | What is proven | What remains |
|---|---|---|---|
| PWA local assets | PROTECTED | Embedded routes, content types, versioned worker, and API exclusions are tested. | Worker activation and browser cache behavior on devices. |
| Stream coalescing | PROTECTED | Cancellation isolation, bounded queue/history, overflow error, cleanup, and race behavior are tested. | Live upstream/browser disconnect acceptance. |
| Remote image ingestion | PROTECTED for decode budget | Scheme/DNS/redirect/byte/dimension/pixel checks and no implicit OCR replay are covered. | OCR/device CPU and provider capability. |
| Desktop update discovery | PROTECTED | Official URLs, bounded metadata, channel ordering, same-key bridge, stable-first current preview, and failure visibility are covered. | Public GitHub state can change; re-download exact release metadata before publication. |
| Desktop staging | PROTECTED locally | Signed manifest, exact size/hash, bundle extraction, code signature, safe paths, preflight, helper lock, rollback, and recovery fixtures pass. | Real `/Applications` run, power interruption, Gatekeeper/quarantine, and rollback on a clean Mac. |
| Native visual interaction | UNKNOWN | Source semantics, focus markers, contrast tokens, modal bounds, and bundle parity are tested. | Desktop/tablet/phone browser walk, 200% zoom, touch, focus trap, artifact runtime, and CDN behavior. |
| Google web-RPC generation | UPSTREAM-DEPENDENT | Parser/adapter fixtures and failure boundaries are tested. | Cookie/session validity, IP/provider limits, model availability, and live route behavior. |
| Developer API route | PARTIAL/EXPLICIT | Separate user-key route, header-only handling, typed stream guards, and request validation are tested. | Student-owned key, Google quota, billing/data-use rules, and live model availability. |
| 30-device rollout | NOT READY | Migration runbook and recording schema exist. | One clean Mac, 2–3 pilots, then staggered waves with per-device evidence. |

## Updater truth for the existing Preview 7 fleet

The already-published `v0.1.7-preview.7` binary predates stable-first
discovery. Its updater can discover a newer preview signed by the same project
key, so the safe updater-mediated route is:

```text
Preview 7 → signed same-key v0.2.0-preview.1 bridge → signed stable v0.2.0
```

The bridge and stable steps are separate explicit user-consent actions. A
Preview 7 Mac does not silently become stable merely because a new release was
uploaded. Preview 6 or older installations with an unrecoverable old trust
anchor require one manual migration to a current-key package.

Before pressing **Install Update** on any Mac:

1. quit any mounted-DMG view and copy BOB Gemini Free to `/Applications` or a
   writable application directory;
2. open the copied app and approve the ordinary unnotarized macOS warning if
   shown; do not disable Gatekeeper globally;
3. use **Help → Check for Updates** and confirm the exact version/channel;
4. select **Install Update**, wait for restart, and verify the new version and
   local gateway health;
5. run one small ordinary prompt, then repeat the check for the next channel
   step only after the bridge has been accepted.

If the updater reports App Translocation or a read-only location, the current
app remains unchanged. Move it to Applications, relaunch, and retry. If a
check times out, do not click repeatedly across the fleet; stagger the next
attempt and record only device label, OS, architecture, version, health, and
error category.

## Deliberately not changed

The following boundaries were intentionally left untouched in this
continuation:

- Gemini payload construction and SAPISIDHASH authentication;
- cookie/session routing and provider policy;
- Google Web RPC wire shapes and adapter response formats;
- CORS/PNA trust design and remote Web Studio pairing;
- automatic quota evasion, proxy rotation, TLS spoofing, or shared cookies;
- silent installation or remote fleet control;
- Tauri history and the Wails-canonical decision;
- generated artifact sandbox permissions and the navigation model;
- release private-key custody and public-release publication.

## Remaining highest-value actions

1. Recover a real browser-capable test surface and execute the design matrix at
   desktop, tablet, phone, keyboard, touch, 200% zoom, artifact, and CDN
   failure states. Record screenshots for any L2 restructuring; do not make an
   L2 layout change without before/after evidence.
2. On one owner-controlled clean Mac, use the exact signed bridge asset from
   `/Applications`, verify Preview 7 → bridge, then bridge → stable, and run a
   deliberately invalid candidate to prove rollback without data loss.
3. Repeat on 2–3 pilot Macs before any 30-device wave. Keep installation
   success, gateway health, and Google generation success as separate fields.
4. Only after those gates pass, sign and manually publish immutable stable
   assets. Re-download the public bytes, verify the detached signature and
   hashes, and retain the non-secret evidence receipt.
5. Revoke all credentials pasted into the earlier conversation. Never store a
   replacement PAT, Google API key, cookie, or private release key in this
   public repository or in a student package.

The repository is now safer to evolve because its most fragile local
boundaries are executable and regression-locked. It is not honest to call the
remaining external gates complete until the corresponding device, browser,
provider, and public-release evidence exists.
