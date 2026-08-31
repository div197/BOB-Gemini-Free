# Changelog

All notable changes to **BOB Gemini Free** (*Break Ordinary Boundaries*) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

The current capability contract is defined by the README and verification
matrix. Older entries describe historical implementation snapshots; they do
not override current evidence, upstream-dependent limits, or the explicit
preview/release boundaries below.

### Historical claim corrections

Several older entries use release-note language that was aspirational or tied
to an earlier implementation. The current contract supersedes those phrases:

- “Auto-healing” OCR is not an automatic replay path. A failed vision upload is
  reported; any OCR-only retry must be an explicit user action.
- “Unlimited subscribers,” “permanently eliminates connection drops,”
  “automatic guest fallback,” and “full drop-in/frontier support” are not
  product guarantees. Buffers, retries, aliases, and adapters are bounded and
  provider/session dependent.
- Model names in the catalog are routing aliases unless a live provider check
  proves identity and availability. Historical “zero breaking changes” and
  “100%” wording is not a compatibility certification.
- Historical fixed RPM/RPD, RAM, latency, free-tier, and classroom-capacity
  numbers are not current measurements. See the verification matrix and dated
  benchmark reports for the evidence boundary.
- The live diagnostic runner now treats invalid JSON, empty or `[DONE]`-only
  streams, incomplete response objects, and unavailable provider-dependent
  image generation as failures. Earlier example output that accepted these
  conditions is historical and not a current diagnostic guarantee.
- The historical “100% pass rate” diagnostic wording is not a release claim;
  the current runner can correctly report a partial live result when a provider
  ignores instruction-only JSON formatting or image generation is unavailable.

## [Unreleased]

### Responsive Studio correctness

- Keeps compact phone header actions inside a bounded 44px-tall interaction
  row while retaining their stable accessible names.
- Caps responsive configuration and integration drawers at 320px on phones;
  the previous `min-width: 85vw` rule could override that cap on a 390px
  viewport.
- Gives Gateway settings buttons and text/password fields a scoped 44px minimum
  interaction height while keeping the dialog internally scrollable.
- Adds a source regression for the phone-control containment and drawer-width
  contract, and keeps the generated `web/index.html` bundle synchronized.

### Release candidate verification

- Packaged and locally verified the next `v0.2.0-preview.7` macOS universal
  candidate from merged `main` (`049ca2f`), including the signed manifest, DMG
  layout, bundle architecture, bundled runtime startup, and updater transition
  matrix. The candidate remains unpublished pending installed-base and pilot
  gates.
- Recorded the current 1/10/20/30-concurrency local benchmark against the
  current source commit. This remains a local-only measurement, not a Google
  quota or classroom-capacity claim.
- Exercised the generated-artifact lifecycle with a synthetic local SSE
  fixture: preview launch and Code-tab hydration passed with the complete
  source, without a provider request or credential.
- Kept Preview 7 explicitly unpublished until installed-base, rollback,
  clean-device, platform-trust, provider, and pilot gates are observed.

## [0.2.0-preview.6] - 2026-08-31

### Release safety

- Made the Preview 6 candidate explicit as `PREVIEW_VERSION` in the Makefile at
  the release source target; the post-publication source now advances that
  guard to the next unused candidate, `v0.2.0-preview.7`.
- Passes that candidate explicitly from every desktop preview target.
- Makes the macOS, Windows, and Linux preview packagers fail closed when
  `BOB_RELEASE_VERSION` is unset, preventing accidental reuse of a published
  preview tag.
- Extends the release-source gate to validate the explicit candidate and the
  packager fail-closed contract.

### Publication

- Published the macOS universal beta manually as `v0.2.0-preview.6` from
  source target `f9b3410`, with the branded DMG/ZIP, release notice,
  `SHA256SUMS`, and detached `SHA256SUMS.sig`.
- Re-downloaded all five public assets into a fresh directory, verified the
  detached Ed25519 signature and exact SHA-256 bytes, and matched them to the
  local signed inputs. No GitHub Actions were used.

### Settings clarity

- Adds a first-use credential decision line for students: leave both key fields
  empty for the default route, use a BOB Gateway Access Key only when the
  gateway owner requires it, and enable the Developer API route only for an
  intentional student-owned AI Studio project.
- Fixes the Config modal's English/Hindi translation binding so gateway access
  labels, help, and placeholders are rendered from the correct language
  dictionary instead of disappearing after a language switch.
- Prevents the page-session BOB Gateway Access Key from being sent to a
  non-loopback cleartext HTTP endpoint. Loopback HTTP remains compatible,
  HTTPS remains supported, and the settings dialog reports `HTTPS REQUIRED`
  while ping, telemetry, model discovery, and generation share the same guard.

### Desktop coexistence

- Strengthens the native gateway reuse handshake with an exact release-version
  marker. A new desktop build no longer silently attaches to an older BOB
  process that happens to own the configured loopback port; it selects a safe
  fallback port instead. Same-version reuse remains supported and owned
  gateways still shut down with the desktop lifecycle.

### Updater verification

- Adds a mocked release-list regression matrix proving that the published
  `v0.2.0-preview.6` candidate is selected for legacy `v0.1.7-preview.7`,
  `v0.2.0-preview.1`, and Preview 5 installations, while the same Preview 6
  build does not update itself. The matrix does not replace a real-device
  installed migration or rollback test.

## [0.2.0-preview.5] - 2026-08-31

### Browser streaming

- The Studio SSE reader now assembles complete events across chunk, CRLF, and
  multi-line `data:` boundaries, ignores standard comments/fields safely, and
  records bounded diagnostics for unknown event metadata without changing
  response ordering.

### Release versioning

- After Preview 4 publication, all Wails preview packagers defaulted to the
  immutable `v0.2.0-preview.5` candidate. Preview 4 remains unchanged; Preview
  5 is now published after clean-source, local package, Keychain signature,
  public-byte, and one installed-bundle migration gates passed.
- Added a deterministic updater matrix for the published Preview 7 fleet,
  migration-bridge previews, Preview 4, and Preview 5. The matrix records the
  real boundary: one signed installed-bundle replacement is now observed on
  the audit Mac, while rollback, pilot, and fleet acceptance still require
  device evidence. See [`current release audit`](docs/engineering/RELEASE-AUDIT-2026-08-31.md).

### Credential-boundary clarity

- Renames the Studio gateway field to **BOB Gateway Access Key** and masks it by
  default so it cannot be confused with a student's Google Gemini Developer API
  key.
- Adds a localized credential map explaining the separate BOB access, Google
  Developer API, web-session/cookie, and endpoint scopes without changing route
  selection, headers, or session-only storage behavior.
- Corrects the Studio's HTTP 401 handling so a rejected Google Developer API
  key is not mislabeled as missing BOB gateway authentication; gateway errors
  now name the BOB access key explicitly.
- Adds an explicit route-status card to Config so students can see the active
  web-session or Developer API path, gateway-auth state, provider-key state,
  cookie ownership, and model guard before sending.
- Blocks the Developer API route before chat history or network work when its
  key, endpoint transport, or model/think-mode contract is invalid; the two
  credential fields also have independent clear actions and are removed from
  the retained modal DOM on close.
- The earlier local `v0.2.0-preview.5` macOS candidate was built from merged
  public-main commit `4beb127`; its package and local startup checks did not
  include the route-clarity patch and it is superseded. A fresh candidate from
  merged public-main commit `a68eb39` passed the same local package and startup
  gates. That candidate was superseded by the published Preview 5 candidate
  after public-byte reconciliation and a one-host installed-bundle migration;
  Preview 5 is now historical because Preview 6 is the current public preview.

## [0.2.0-preview.4] - 2026-08-31

### Controlled macOS preview publication

- Published the macOS universal prerelease from public-main commit
  `abfeeba` with the current updater trust anchor and a signed
  `SHA256SUMS` manifest.
- Re-downloaded all five public assets, verified the detached Ed25519
  signature, and reconciled every downloaded file with the locally signed
  publication input byte-for-byte.
- Invalidates cached dynamic Google `/app` page tokens after an explicit HTTP
  401/403 rejection without erasing configured cookies, rotating identities,
  or replaying the rejected request. The in-flight bootstrap ordering race is
  regression-tested.
- Advances the unqualified Wails preview packager default to the immutable
  Preview 4 identity so it cannot accidentally recreate Preview 3.
- Confirms a fresh universal package launch, loopback `/healthz`, rendered
  Preview 4 UI version, occupied-port fallback, and clean shutdown on the
  audit Mac.
- This remains an ad-hoc signed, non-notarized, macOS-only controlled beta;
  clean-device replacement, rollback, and broad student rollout remain open.

## [0.2.0-preview.3] - 2026-08-31

### Controlled macOS preview publication

- Advances the immutable macOS preview line after Preview 2 with the
  bounded updater metadata retry and calm update/staging failure dialogs
  merged on public `main`.
- Adds responsive drawer dialog semantics, focus-in, Tab trapping, Escape
  close, and focus return for keyboard users at tablet and phone widths.
- Ensures the ZIP and DMG both present the branded `BOB Gemini Free.app`
  bundle name instead of leaking a temporary build-directory name.
- Local browser smoke evidence covers 1440x900, 1024x768, and 390x844 with no
  document horizontal overflow or console warnings/errors.
- The exact macOS universal assets were built from public-main `284b7d1`,
  signed through the owner-controlled macOS Keychain, published at
  [`v0.2.0-preview.3`](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.3),
  and re-downloaded, signature-verified, and byte-reconciled.
- This remains an ad-hoc signed, non-notarized, macOS-only controlled beta;
  clean-device replacement, rollback, and broad student rollout remain open.

## [0.2.0-preview.2] - 2026-08-31

### Controlled macOS preview publication

- Published a macOS universal prerelease from public-main commit
  `6d3a0cfc0a7a0bf05a3c136baf96a48f503b45ef` with the current updater key and
  signed `SHA256SUMS` manifest.
- Re-downloaded all five GitHub assets, verified the detached signature, and
  reconciled every downloaded byte with the locally signed publication input.
- Kept the release explicitly ad-hoc signed, non-notarized, macOS-only, and
  suitable for controlled preview testing rather than a silent fleet rollout.
- Existing same-key Preview 7 installations can now discover Preview 2 through
  their preview-channel updater path; actual replacement and rollback remain
  clean-device and pilot gates.

---

## [Unreleased] — 2026-08-31

### Current public-main follow-ups

- Added real in-app-browser evidence that an untrusted cross-port loopback web
  origin cannot read models or complete a JSON/PNA preflight; the public
  HTTPS/PNA and remote-Studio pairing gates remain explicitly open.
- Stopped the Studio from probing the gateway and model list on every partial
  BOB access-key keystroke. The explicit **Test Ping** action now performs the
  network verification, with English and Hindi guidance.
- Changed recurring Studio telemetry to establish liveness and auth state from
  the public `/healthz` endpoint before requesting protected aggregate stats;
  secured endpoints no longer generate a repeated 401 poll when no BOB access
  key is entered, and stale aggregate values are cleared.
- These follow-ups are merged into public `main` after Preview 5 publication;
  the immutable Preview 5 assets are unchanged and a new signed preview is
  required to deliver them.

### Browser preview and responsive layout

- Fixes interactive artifact previews that could appear blank in a WebView when
  the iframe was hydrated while its modal was hidden; the modal is now visible
  before generated `srcdoc` is assigned, with a focused regression test.
- Keeps the New action, prompt, and language selector reachable at tablet and
  phone widths while preventing header controls from creating page-level
  horizontal overflow. See the dated
  [`browser validation record`](docs/engineering/BROWSER-UI-VALIDATION-2026-08-31.md).

### Responsive drawer accessibility

- Responsive configuration and integration drawers now expose temporary
  dialog semantics, move focus into the opened drawer, trap Tab navigation,
  close on Escape, and return focus to the opening control. The implementation
  preserves the existing desktop sidebar behavior and does not change the
  navigation model.

### Explicit Gemini Developer API route (local preview work)

#### Added
- Added an opt-in public Gemini Developer API transport with native REST and
  SSE handling, isolated from the reverse-engineered web-RPC client.
- Made every `desktop-preview-*` Make target pass the checked-in public updater
  key to its packager and require a semantic `-preview.N` version with the
  preview channel, so a preview build cannot fail late or be labelled as a
  different release channel by accident.
- Added OpenAI-shaped chat translation for system instructions, data-URL
  images, generation settings, JSON-object output, native function declarations,
  and `auto`/`none`/`required`/specific-tool choice mapping.
- Added native Google-shaped forwarding for `generateContent`,
  `streamGenerateContent`, and `countTokens` when a provider-shaped `gemini-*`
  model ID and an explicit provider key are selected; future provider IDs are
  forwarded for Google to accept or reject.
- Added a student-facing Studio Config section linking to Google AI Studio's
  key-management page, current model catalogue, and Google's current
  rate-limit guidance. The browser keeps the provider key in memory for the
  current page session only; it does not persist or include it in telemetry,
  health checks, update checks, logs, or release assets.
- Added the same page-session boundary to optional BOB gateway authentication:
  gateway tokens are no longer persisted in browser storage, and the legacy
  `bob_api_key` storage entry is purged on page load. Protected gateway users
  must re-enter the token after a reload, preventing shared classroom devices
  from retaining a reusable credential.
- Added deterministic transport, translation, SSE, routing, tool-call,
  configuration, and secret-redaction tests without using a live Google account.
- Release-source preflight now derives and verifies the Bash installer's
  Ed25519 SPKI encoding against the canonical public key, preventing an
  alternate installer key encoding from passing packaging checks.

#### Changed
- Added `BOB_GEMINI_FREE_GEMINI_API_KEY` as a single process-environment
  provider key option. Comma-separated key pools and automatic provider
  rotation are intentionally unsupported.
- Requests with an explicit provider key never silently fall back to the
  cookie/web route. Unsupported `/v1/messages`, `/v1/responses`, and image
  generation provider surfaces fail clearly until separately translated.
- Replaced stale fixed Google AI Studio RPM/RPD claims in the current English
  and Hindi documentation with links to Google's current pricing, rate-limit,
  billing, and key documentation. Provider limits remain model/project/tier
  dependent and must not be described as unlimited.
- Clarified the roadmap's educational vision so “local-first” and “free” do not
  imply a provider-free or unlimited-access guarantee.
- Hardened the 15-point diagnostic runner so a green result requires observed
  usable protocol output rather than only an HTTP 200, a terminal sentinel, or
  a guest-mode explanation for unavailable image generation.

The macOS subset of this work is included in the published `v0.2.0-preview.2`
controlled preview. Stable publication, other-platform acceptance, and live
provider validation remain separate gates. No live provider key or GitHub
Actions workflow was used.

### Release-transition audit

#### Added
- Recorded the immutable public `v0.2.0-preview.1` bridge, the released
  `v0.1.7-preview.7` fleet baseline, and the exact Preview 7 → Preview 1/2 →
  stable migration boundaries in [`RELEASE-TRANSITION-AUDIT-2026-08-31.md`](docs/engineering/RELEASE-TRANSITION-AUDIT-2026-08-31.md).
- Locked and published the controlled source package identity as
  `v0.2.0-preview.2`; the preview packagers no longer default to rebuilding the
  already-published Preview 1.
- Added updater fixtures for `v0.2.0-preview.1 → v0.2.0` stable-first
  migration, `v0.2.0-preview.1 → v0.2.0-preview.2` continuation, and legacy
  Preview 7 selection of the highest later preview.

### Reliability and release hardening

#### Added
- Bounded native staging/start error dialogs so read-only, permission, timeout,
  cancellation, and generic updater failures give recovery guidance without
  exposing local filesystem paths or low-level OS text; they state that no app
  change occurred.
- Made the native update-check dialog classify cancellation, timeout, network,
  and release-metadata failures into calm actionable messages without exposing
  raw transport or GitHub response details; it clearly states that no app
  change occurred.
- Made official GitHub release-metadata checks retry one transient transport or
  client-timeout failure, with a short cancellation-aware budget. HTTP status
  errors, redirect-policy failures, malformed metadata, and signed asset
  downloads remain fail-closed and are not retried.
- Continued the no-Actions hardening campaign with bounded session-bound image
  reference expiry and retryable page-token refresh, visible local-history
  persistence recovery, abortable attachment parsing, fail-closed preference
  storage, dedicated service health identity, committed-response status
  logging, literal-loopback CORS validation, standard SSE field handling in the
  Studio browser reader, and immediate cancellation of active PDF.js loading
  or document tasks when an attachment is removed.
- Made a cancelled coalesced-stream subscriber win over a concurrently queued
  final delta, preventing a false successful completion while preserving the
  shared response for remaining subscribers.
- Corrected native macOS packaging so the inspectable `.app` stays outside the
  signed release-asset directory, and made manifest creation reject accidental
  directory assets instead of silently skipping them.
- Added a release-source preflight that treats the canonical Ed25519 public-key
  block, standalone macOS/Linux and Windows installer trust anchors, Docker
  base version, and preview packager defaults as one checked matrix. A key or
  version drift now fails before a release package is built.
- Added bounded, fail-closed release-control verification and a local release
  evidence receipt so the signed manifest, exact asset set, source revision,
  toolchain, and host can be reconciled before publication.
- Added regression coverage for pre-request-only generation retries, DNS
  rebinding during remote-image fetches, cumulative direct-API tool calls,
  tool-result correlation, ambiguous candidates/finish reasons, structured
  stream errors, and invalid configured sessions.
- Hardened native updater transaction durability with flushed same-directory
  metadata commits, Unix directory synchronization around swap/recovery
  transitions, and a fault-injected rollback regression; Windows metadata now
  uses native `MoveFileExW` replace-existing/write-through semantics rather
  than a delete-then-rename fallback.

#### Changed
- The default CLI installers now download a release, verify its signed
  manifest and exact digest, and refuse unsigned or unverifiable binaries.
  Source installation is an explicit opt-in for development use; the installer
  commands are documented as download, inspect, and run rather than pipe-to-
  shell shortcuts.
- Native and browser streaming paths now preserve provider errors as structured
  errors instead of fabricating assistant text, and generation retries do not
  replay requests after an ambiguous or partially delivered upstream failure.
- Removed unenforced synthetic rate-limit headers, made malformed discovered
  configuration fail closed with a visible startup error, and bounded status,
  CDP, updater-plan, and updater-confirmation metadata reads.
- Browser fallback, Wails startup, nil embedded clients, and updater plan paths
  now report errors explicitly instead of logging or accepting false success.

These changes are tested on the development host and the macOS Preview 2
assets are publicly reconciled; clean-device, live-provider, and
signed/notarized platform acceptance remain release gates.

## [0.2.0-preview.1] - 2026-08-29

### Preview 7 migration bridge

#### Added
- Published a same-key macOS universal migration bridge for existing
  `v0.1.7-preview.7` installations.
- Existing Preview 7 binaries can discover this preview through the fixed
  preview channel and update with the signed manifest, after explicit user
  consent.
- The bridge contains the stable-first updater path, so it can later move to
  stable `v0.2.0` after that release passes the clean-device and pilot gates.
- Added the public
  [`Preview 7 to v0.2.0 Migration Runbook`](docs/engineering/PREVIEW-7-TO-V0.2.0-MIGRATION.md)
  for staged 30-Mac rollout.

This remains an ad-hoc-signed, non-notarized controlled preview. Stable
`v0.2.0` is not published yet.

---

## [Unreleased] — v0.2.0 stable rollout (not yet published)

### The Sovereign World Gateway source milestone

#### Added
- Accept case-insensitive HTTP `Bearer` authentication schemes for gateway
  API-key clients, with a regression test for lower-case client implementations.
- **Mobile Engine & In-Process Gateway Bridge (`pkg/mobile`)**:
  - Added an experimental Go bridge substrate for future mobile bindings; no native Android/iOS application, AAR, or XCFramework is shipped by this repository.
  - The current bridge starts a local HTTP listener and uses the existing Google web upstream; it is not an in-process, zero-socket, zero-latency mobile runtime.
  - Native `CookieManager`, `WKHTTPCookieStore`, and hardware-keystore integration remain future application work, not implemented release features.
  - The mobile guide now records the actual bridge boundary and the missing native artifacts.
- **Three-Stage Reasoning Refiner (`internal/refiner` & `POST /v1/refine`)**:
  - Requirement decomposition (Stage 1), self-critical invariant audit (Stage 2), and final synthesis (Stage 3).
  - The stages call the configured inference function and are therefore upstream-/session-dependent when used by the server; the refiner is not a pure-local inference engine.
  - Integrated into Web Studio Command Palette (`⌘ + Shift + R` / `K Menu → Deep Invariant Refiner`) and `POST /v1/refine` HTTP endpoint.
  - 100% test pass rate across unit and mock inference pipelines.
- **Production-Grade Multi-Arch Docker & OrbStack Pipeline**:
  - Added a multi-arch Alpine build path and an unauthenticated local `/healthz` container probe.
  - Image size, boot latency, and RAM remain environment-specific measurements; no blanket performance number is claimed here.
- **Universal BPE Token Estimator Correction**:
  - Improved local estimation for dense JSON, delimiters, and multilingual text; counts remain estimates rather than provider-authoritative tokens.
- **Global Classroom & Hackathon Lab Deployment Guide ([`docs/1-getting-started/school-and-offline-lab-guide.md`](docs/1-getting-started/school-and-offline-lab-guide.md))**:
  - Added an operational guide for a possible LAN hub; real capacity, provider acceptance, and operating cost remain deployment-specific.

#### Additional stable-release engineering not yet published

##### Native preview packaging

- macOS preview DMGs now place `BOB Gemini Free.app` beside a conventional
  `/Applications` shortcut for drag-to-install use.
- Added a mounted-image layout verifier so the packager fails closed if the
  final DMG loses the application drop target.

##### Native updater engineering substrate

- Added a native desktop updater path that discovers only fixed official stable
  or preview GitHub channels, verifies an embedded Ed25519 trust anchor plus
  the signed `SHA256SUMS` manifest, checks exact platform assets and sizes,
  stages safely, and performs post-exit replacement with health confirmation
  and rollback.
- Newly built preview packages check stable first for an explicit one-way
  Preview → Stable migration, then continue to the preview channel only when
  stable has no newer release. Stable builds never move into preview.
- Existing public Preview 7 binaries predate stable-first discovery; they need
  a same-key bridge preview or a manual stable installation before updater-
  based migration to stable.
- Added semver prerelease comparison and `preview.N` selection so a signed
  preview build can discover later signed preview releases.
- Added mocked tests for signed archive staging, environment-key rejection,
  declared-size mismatch, archive traversal, transactional replacement,
  stable-first migration, legacy Preview 7 bridge discovery, direct-stable
  rejection, and rollback. Tests never replace the developer's running
  executable.
- Hardened native updater durability: updater metadata is privately written via
  flushed same-directory atomic replacement, Unix transaction directories are
  synchronized after swap/recovery transitions, and an injected activation-sync
  failure is regression-tested to restore the previous install. This reduces
  interrupted-update ambiguity but does not claim recursive app-bundle fsync,
  Windows directory-fsync semantics, or real power-cut proof. Windows metadata
  replacement now uses native `MoveFileExW` replace-existing/write-through
  semantics instead of a delete-then-rename fallback.
- Added a local no-Actions `scripts/sign-release-assets.sh` operator step and
  documented release-key custody, platform signing, clean-device acceptance,
  and the 30-device rollout gate.

##### Documentation truth corrections

- Replaced older unconditional session, context, Pro, Imagen, permanent-login,
  and universal-client wording with session/provider-dependent boundaries.
- Clarified that local aggregate counters are not external telemetry, that
  there is no BOB signup or BOB cloud chat service, and that a packaged app's
  no-Go/no-Node/no-SQLite runtime claim is different from source-build
  prerequisites.
- Documented the distinction between anonymous web access, optional Google web
  cookies, shared school-network egress, local health, and provider limits.
- Stopped automatic retries for explicit upstream policy/rejection responses;
  transport retries and cumulative-stream deduplication remain protected by
  fixtures.
- Kept the New/model toolbar above responsive drawers and bounded manual UI
  retries so a visible provider failure cannot be multiplied by rapid clicks.

##### Session rejection recovery

- Explicit Google HTTP 401/403 responses now invalidate only the cached dynamic
  `/app` page token and build identifier. The configured cookie file and guest
  cookie remain intact, and the next request performs a fresh bootstrap rather
  than reusing rejected session material.
- An invalidation generation prevents an older in-flight bootstrap from
  restoring tokens after the rejection. The buffered and streaming paths,
  credential preservation, refresh, and in-flight ordering are covered by
  deterministic tests.
- This is bounded local recovery, not automatic reauthentication or a provider
  availability guarantee. The rejected request is not replayed.

---

## [0.1.9] - 2026-08-28

### Client-Side Edge Document Intelligence, BPE Tokenizer Lab, Financial Analytics & Niṣkāma UI Polish

#### Added
- **Client-Side Edge Document Intelligence Engine (Zero-Cloud / Zero-Server Architecture)**:
  - In-browser document parsing and conversion for **PDF** (`pdf.min.js`), **Word DOCX** (`mammoth.browser.min.js`), **Excel XLSX / XLS / CSV / TSV** (`xlsx.full.min.js`), **Code / Data / Markdown / SQL / Log / Env / JSON / YAML**, and **Image OCR** (`tesseract.min.js`).
  - Zero server uploads: All files are parsed directly into spec-compliant structured Markdown with tables and section headings on the user's client machine before prompt assembly.
  - Interactive Document Preview Modal (`openDocPreviewModal`) with full Markdown formatting, code syntax highlighting, token count stats, and one-click clipboard copy.
  - Full-window drag-and-drop backdrop overlay (`.drop-zone-overlay`) and dynamic multi-file attachment shelf (`.attachments-shelf`) with individual file preview and removal.
  - **Client-Side Local OCR Auto-Healing**: If Google upstream returns `BardErrorInfo [1003]` on unauthenticated multimodal requests, BOB automatically extracts text from the image using in-browser Tesseract WASM and injects it into the prompt seamlessly.
- **Interactive BPE Tokenizer Lab & Multi-Dimensional Token Intelligence (`openTokenizerModal`)**:
  - Interactive BPE visualizer modal displaying subword token boundaries, token count, character length, compression ratio, byte counts, and raw integer token arrays.
  - Live Prompt Token Estimator (`updatePromptTokenEstimate`) dynamically computing the combined token weight of typed text **plus** all attached multi-page PDFs, spreadsheets, and documents in real time.
  - User attachment badges in chat history displaying page counts and token weights (`📕 Artificial_Intelligence_of_Pushpaka_Vimana.pdf • 7 pg • ~4,820 tok`).
  - Assistant card telemetry reporting a complete three-tier token breakdown:
    $$\text{In: } \sim 4,845 \text{ tok} \;\bullet\; \text{Out: } \sim 524 \text{ tok} \;\bullet\; \text{Total: } \sim 5,369 \text{ tok} \;\bullet\; 60.4 \text{ tok/s}$$
- **Financial Savings & Multi-Model Benchmark Pricing Engine (`internal/models/pricing.go`)**:
  - Real-time dollar savings calculator benchmarking requests against commercial cloud pricing (August 2026 rates) for Claude 5/4.5/3.7, GPT-5.6/5.5/4o/o3, Gemini 3.7/3.1, and DeepSeek models.
  - Integrated into top status bar (`Saved: $X.XX`), gateway `/healthz` telemetry, and client response metadata.
- **10 Distinct Color Themes & Custom Theme Color Studio (`openCustomThemeModal`)**:
  - 10 hand-crafted color palettes: `bob-builder` (Obsidian Gold), `apple` (Cupertino Parchment Light), `vodafone` (Scarlet Light), `spotify` (OLED Pitch Dark / Emerald), `quantum` (Cyan Glow), `tokyo-night` (Neon Purple), `monokai` (Hacker Gold), `nord` (Arctic Slate), `solarized-light` (Warm Amber), and `custom`.
  - Real-time Custom Theme Color Studio modal with interactive color pickers for accent, app background, card background, and main text, with presets (Cyberpunk, Synthwave, Dracula, Matrix).
- **8-Language Internationalization & Offline Indic Transliteration (`I18N`)**:
  - Full UI localization across 8 languages: English, Hindi (हिंदी), Sanskrit (संस्कृतम्), Spanish (Español), French (Français), German (Deutsch), Japanese (日本語), Chinese (中文).
  - Integrated offline rule-based phonetic transliteration engine for Indic scripts (Hinglish to Hindi/Sanskrit/Marathi) with instant Ctrl+G toggle and word boundary caching.
- **Engineering Specifications & System Architecture Documentation**:
  - Added [`docs/engineering/MULTI-LINGUAL-I18N-SYSTEM.md`](docs/engineering/MULTI-LINGUAL-I18N-SYSTEM.md) detailing internationalization and offline transliteration architecture.
  - Added [`docs/engineering/ROADMAP-v0.1.9-NISHKAAM-VISION.md`](docs/engineering/ROADMAP-v0.1.9-NISHKAAM-VISION.md) detailing the architectural evolution and Niṣkāma Karma Yoga principles.
  - Added [`docs/engineering/WISHLIST-16YO-HACKER-v0.1.9.md`](docs/engineering/WISHLIST-16YO-HACKER-v0.1.9.md) detailing hacker tooling, WASM Python sandbox, and BPE tokenizer lab.

#### Changed & Refined
- **Proportional Zoom Architecture (`--reading-zoom`)**:
  - Dynamically scales Markdown responses, input dock, reasoning cards, tables, and starter cards proportionally across zoom levels from 50% to 200%.
- **Markdown Table Typography & Contrast**:
  - Upgraded table cells (`td`, `th`) with `10px 14px` padding, 2px borders, bold gold header accents, and proportional zoom scaling (`calc(clamp(0.95rem, 0.92rem + 0.15vw, 1.05rem) * var(--reading-zoom))`).
- **Streamlined Assistant Response Card Layout**:
  - Removed duplicate button clusters from card headers. Header now displays sleek assistant badge and multi-dimensional token stats, while footer provides the interactive action suite (`📋 Copy`, `🔊 Listen`, `↻ Retry`, `🧠 Think`, `⚡ Flash`).
- **Toast Notification Stack Architecture**:
  - Shifted `.toast-stack` to `top: 68px` on desktop and floating above the input dock on mobile screens (`bottom: calc(var(--composer-offset) + 14px)`). Added de-duplication to prevent duplicate alerts.
- **Mobile (< 640px) Responsive Harmony**:
  - Streamlined mobile header layout: hides `.github-pill` and collapses gateway status text to prevent horizontal overflow on narrow mobile screens.

#### Fixed
- **iCloud Keychain / Password AutoFill Interference**:
  - Added `<meta name="format-detection" content="telephone=no, date=no, address=no, email=no">` to `<head>`.
  - Added anti-autofill attributes to `#user-input` (`name="chat_prompt_body"`, `data-disable-password-manager="true"`, `data-1p-ignore="true"`, `data-lpignore="true"`, `data-bwignore="true"`, `data-form-type="other"`, `data-private="true"`).
  - Hardened modal inputs with `autocomplete="new-password"` to permanently suppress browser autofill heuristics from targeting `localhost:9610`.

---

## [0.1.8] - 2026-08-28

### High-Concurrency Satavik Architecture, StreamFlight Multiplexing & KaTeX Math Hardening

#### Added
- **Dual Stream & Non-Stream SingleFlight Multiplexer (`internal/gemini/flight.go`)**:
  - Deduplicates identical concurrent generative AI requests (e.g. dense computer labs with hundreds of students requesting identical assignments like 2D CyberSnake or matrix simulations).
  - Streams real-time deltas from 1 upstream connection to unlimited downstream subscribers via atomic buffered broadcast channels with zero memory leaks on early client disconnects.
  - Fuzz-tested with 1,000 concurrent goroutines under Go's race detector (`-race`).
- **Anonymous Guest Session Recycling with Stale-While-Revalidate & Thundering-Herd Guard (`internal/gemini/auth.go`)**:
  - Ephemeral guest session discovery now uses a non-blocking Stale-While-Revalidate caching pattern with single-flight mutex protection (`refreshing` / `refreshCh`).
  - When hundreds of students connect simultaneously, exactly 1 background request queries `gemini.google.com/app` while all other requests are served instantly with 0ms blocking.
- **Dynamic SSE Keep-Alive Engine (`internal/server/helpers.go`)**:
  - Injected periodic SSE comments (`: keepalive\n\n`) every 2.5 seconds across all streaming endpoints (`POST /v1/chat/completions`, `POST /v1/messages`, `POST /v1/responses`, `POST /v1beta/models/...`).
  - Permanently eliminates connection drops, HTTP 504 timeouts, and socket drops during extended reasoning phases (up to 20k+ thinking tokens).
- **Resilient Multi-Account Cookie Failover & Auto-Guest Fallback (`internal/gemini/client.go`, `pool.go`)**:
  - If a configured cookie (`cookie.txt` or cookie pool) fails or expires, it is placed into a 60-second cooldown and BOB seamlessly and automatically falls back to the live anonymous guest session, permanently eliminating HTTP 503 errors.
- **Code-Isolated KaTeX Scientific Typography & Currency Protection (`renderMarkdown` in `playground.html` & `web/index.html`)**:
  - Pre-isolates multi-line (```` ```...``` ````) and inline (`` `...` ``) code blocks before LaTeX scanning, preventing code containing `$` (e.g. `price = "$100"`) from being corrupted by math renderers.
  - Hardened inline math delimiter regex (`(?<![\w\\])\$(?!\s)([^$\n]+?)(?<!\s)\$(?!\d)`) cleanly distinguishes LaTeX formulas (`$E = \hbar\omega$`) from ordinary currency notations (`$100`, `$50`).
- **Dual-Format Reasoning & Thinking Token Extractor (`internal/format/thinking.go`)**:
  - Extracts and separates reasoning tokens across both Markdown code fences (```` ```thought\n...\n``` ````) and XML tags (`<thought>...</thought>`), streaming pure reasoning deltas to OpenAI `reasoning_content` and Anthropic `thinking` blocks.
- **Multilingual Token Counter Calibration (`internal/format/tokens.go`)**:
  - Calibrated subword estimations for Sanskrit mantras (*कर्मण्येवाधिकारस्ते...*), Devanagari, CJK ideographs, complex LaTeX integrals, and OpenAI tool calling schemas.
- **15-Point Ultimate Deep Diagnostic Test Kit (`internal/diag/diag.go`)**:
  - Expanded automated test runner to 15 core end-to-end checkpoints including Claude Code CLI SSE streaming tool execution protocol and StreamFlight high-concurrency (5-client) multiplexing and coalescing.
  - Verified 100% pass rate across the entire suite (`./bob-gemini-free --test`).
- **Universal Frontier & Date-Versioned Model Catalog (`internal/models/models.go`)**:
  - Added full drop-in support for the latest frontier generations: Anthropic Claude 5 (`claude-5-sonnet`, `claude-5-opus`, `claude-5-fable`), Claude 4.5/4.x, OpenAI GPT-5.6 (`gpt-5.6-sol`, `gpt-5.6-terra`, `gpt-5.6-luna`), GPT-4.5, Codex series, and Google Gemini 3.7/3.5/Omni models.
  - Includes exact date-versioned SDK aliases (`claude-3-7-sonnet-20250219`, `gpt-4o-2024-11-20`, `o3-mini-2025-01-31`) ensuring zero breaking changes across 1,000+ developer tools.
- **Dynamic Frontend-Backend Synchronization & Studio Polish (`playground.html` & `web/index.html`)**:
  - `syncBackendModels()` automatically discovers, categorizes, and groups all 64+ live models in `<select id="model-select">` and the Command Palette.
  - Live Prompt Token Estimator previews exact character count and token consumption (`0 chars • ~0 tok`) in real-time as users type.
  - Gateway Connection Modal features an integrated Bearer API Key manager with visibility toggle and live engine telemetry metrics.
- **Automated CLI Telemetry & Diagnostic Polish (`main.go`)**:
  - `./bob-gemini-free --status` automatically detects configured `api_keys` from `config.json` without requiring manual `--test-key` entry.
  - Verified 100% pass rate across all 14 Go packages.

---

## [0.1.7-preview.7] - 2026-08-25

### Native preview reliability and updater-key migration

- Added the Preview 7 project update public key and signed release-manifest
  packaging path; the private signing key remains outside the repository in
  owner-controlled local key custody.
- Added an explicit one-time migration notice for Preview 6 installations,
  whose original project signing key is not recoverable.
- Stopped retry amplification for explicit Google policy/rejection responses
  while preserving transport/server retry and cumulative-stream deduplication.
- Added visible terminal provider failures, bounded manual retries, safer
  generation control locking, and responsive drawer placement below the
  New/model toolbar.
- Added regression coverage and local 20/30-client benchmark evidence for
  the release.

## [0.1.7-preview.6] - 2026-08-25

### Native updater reliability hotfix

- Bounded the unauthenticated GitHub preview-release listing to 30 entries;
  the previous 100-entry request reproduced intermittent HTTP 504 responses
  even though the repository had only a handful of releases.
- Converted the macOS read-only disk-image/App Translocation staging failure
  into actionable installation guidance: move the app to Applications,
  relaunch it, and retry.
- Added a regression test for the bounded preview endpoint.

Preview 6 is the recommended signed preview updater target. Preview 5 remains
published and immutable, but its preview-channel check can encounter the
larger GitHub API page timeout on affected networks.

## [0.1.7-preview.5] - 2026-08-25

### Native desktop and language-quality refinement

- Enabled the native macOS zoom/maximize control while preserving the
  resizable Wails window; the packaged candidate was maximized and exercised
  with a real completed response on the audit device.
- Routed external HTTP(S) links from the native WebView through the operating
  system's default browser, while preserving ordinary new-tab behavior in a
  hosted browser.
- Expanded the English/Hindi UI boundary across navigation, configuration,
  integration panels, gateway dialog, response actions, starter cards, voice
  controls, command-palette section labels, and stored-message re-rendering.
- Made the transliteration fallback cache language-specific so a Hindi offline
  fallback cannot be reused for Sanskrit, Marathi, Bengali, Gujarati, Tamil,
  Telugu, or Punjabi input.
- Injected the actual desktop build version into the served studio instead of
  displaying a stale hard-coded version string; the Preview 5 candidate shows
  `v0.1.7-preview.5`.
- Added regression checks for the native window contract, browser bridge,
  language coverage, transliteration isolation, and version injection.

Preview 5 remains an explicitly labelled, ad-hoc-signed macOS preview. It
inherits the signed project updater channel but does not provide Apple
Developer ID trust, notarization, silent updates, or proof of provider access.

---

## [0.1.7-preview.4] - 2026-08-25

### Signed preview update channel

- Published the corrected macOS universal DMG with a visible `/Applications`
  drag target.
- Added the fixed `preview` release channel and correct semver ordering for
  `v0.1.7-preview.N` releases.
- Embedded the public Ed25519 update key into the native preview and published
  `SHA256SUMS` plus detached `SHA256SUMS.sig` for project-level authenticity.
- Preview 4 can discover, verify, stage, and roll back a later signed preview
  after explicit user consent. It does not silently update or remove the
  macOS first-launch warning because it is not Developer ID signed/notarized.

## [0.1.7-preview.3] - 2026-08-22

### Branded native beta refresh

- Published the branded macOS universal and Windows x64 beta package names.
- Removed framework-branded bundle identifiers and executable names from the
  student-facing app metadata and release notices.
- Preserved the explicit manual-update boundary: this prerelease has no
  embedded desktop trust key or signed update manifest.
- Retained the Preview 2 asset names only as updater migration inputs; new
  releases use the BOB Gemini Free package contract.

## [0.1.7-preview.2] - 2026-08-22

### Native Preview Streaming Lifecycle Correction

This manually published, no-GitHub-Actions prerelease corrects a frontend
state bug found during real-device evaluation of Preview 1.

#### Fixed
- The red `STOP` control now returns to `SEND` after successful completion,
  user cancellation, upstream failure, timeout, or truncated stream.
- Stream cleanup references now remain visible to the `finally` lifecycle
  block instead of throwing a JavaScript scope error.
- Decoder tail bytes are flushed at EOF, and an incomplete stream is reported
  instead of being presented as a successful response.
- User cancellation and read/timeout failures retain their correct state so
  a subsequent prompt can be submitted normally.

The embedded Go gateway, anonymous/default routing, cookie handling, and
Gemini protocol core are unchanged by this frontend correction.

## [0.1.7-preview.1] - 2026-08-22

### Public Native Desktop Preview (Beta)

This manually published, no-GitHub-Actions prerelease is an authentic
open-source beta of the native desktop application. Platform publisher trust
and broad rollout acceptance remain separate release gates.

#### Included
- macOS universal `.dmg` and `.zip` packages for Apple Silicon and Intel;
- Windows x64 native desktop executable;
- release notice and SHA-256 checksums;
- an embedded Go gateway and BOB Builder studio with no Go, Node, Rust,
  SQLite, or separate server required at runtime.

#### Known limits
- macOS is ad-hoc signed only; it is not Developer ID signed or notarized;
- Windows is not Authenticode signed and requires WebView2 on the device;
- Linux is not included in this preview because a native Linux build and
  package acceptance have not been completed;
- authenticated Google features require each user's own authorized session;
- the native app does not silently download or replace itself; update checks
  are explicit and the preview is downloaded manually from the prerelease page.

The preview is suitable for informed evaluation and a controlled pilot, not a
general “download and trust” production rollout.

## [0.1.7] - 2026-08-20

### The Desktop Paradigm & True Static Unmetered UI

#### Added
- **Pure Go Native Desktop App (Wails)**: Engineered a standalone click-to-run desktop application (`cmd/desktop`) using Wails. It bundles the UI and the Gateway engine natively with zero NPM dependencies and zero IPC overhead. (`make desktop`)
- **Wails-only Desktop Path**: Consolidated native desktop distribution on the Wails application (`cmd/desktop`); the former alternate wrapper is archived in Git history and is not shipped.
- **100% Static Cloudflare Pages**: Eradicated the Cloudflare Functions (Edge proxy). The public web UI is now completely static and connects directly to your local `http://127.0.0.1:9610`, bypassing the 100k daily request limit entirely.
- **Headless Mode (`--headless`)**: Added to prevent the Go binary from spawning a system browser when running inside Native Desktop wrappers.
- **Global Config Isolation**: The Native App dynamically ignores `~/.config/bob-gemini-free/config.json` API keys to ensure the bundled UI boots cleanly out-of-the-box.

#### Fixed
- **Architectural & Concurrency Hotfixes**: Resolved deep JS event loop locks, Go mutex bottlenecks, memory leaks, and stream chunking races across the Edge and Core during the rigorous audit.
- **Native Lifecycle Hardening**: The Wails window now receives the actual loopback endpoint, surfaces startup errors, and closes only the gateway process it owns.
- **Safari Unhandled Rejections**: Cleared stream timeout promises on resolve to prevent unhandled rejection crashes that caused the UI to hang on STOP in Safari.
- **Robust AST-like State Scanner**: Replaced brittle stream regex with a precise AST-like scanner for perfect thinking-token isolation.

---

## [0.1.6] - 2026-08-20

### Enterprise Readiness Audit: Multi-Modal Hardening, Classroom Deployments, and CLI Telemetry Fixes

#### Added
- **Classroom LAN Guide (`docs/1-getting-started/classroom-lan-guide.md`)**: Comprehensive operational runbook for deploying local instances in dense environments.
- **Dynamic Prompt Injection**: Automatically prepends "Please analyze the attached image" when users attach multimodal images without prompt text.
- **Support for OpenAI Responses API Multipart**: Added seamless parsing for Codex CLI `image_url` and `input_image` blocks natively.

#### Changed
- **Multi-Modal Upload Resilience (`internal/multimodal/upload.go`)**:
  - Enforced a hard 20MB limit on remote image fetches using `io.LimitReader` to prevent memory-bomb attacks.
  - Implemented strict MIME detection before hitting upstream Google infrastructure.
- **Cloudflare Edge Error Granularity (`functions/v1/chat/completions.js`)**:
  - Cloudflare workers now actively passthrough upstream Google `429` (Rate Limit) and `403` (Forbidden) errors accurately instead of masking them as hardcoded `502 Bad Gateway` timeouts.
- **Web UI Diagnostics (`playground.html`)**:
  - UI now cleanly separates and displays specific API connection failures versus API rate-limiting errors, preventing users from receiving misleading "Ensure engine is running" alerts.

#### Fixed
- **CLI Telemetry Auth Bypass (`main.go`)**: 
  - Fixed a critical bug where `--status` would fail with `401 Unauthorized` on gateways secured with `api_keys`. The telemetry check now correctly supports the `--test-key` flag for secured diagnostics.

---

## [0.1.5] - 2026-08-20

### Historical 2022–2026 feature snapshot (partially superseded)

The following entry records an earlier implementation snapshot. The browser
SQLite surface described below was removed from the active tree during Phase
III and is not part of the current product contract.

#### Added
- **In-Place Atomic Self-Updater (`--update` / `update` & `GET /v1/update/check`)**:
  - Automatically queries official GitHub Releases API, parses SemVer, downloads OS/Arch matching binary (`darwin-arm64`, `darwin-amd64`, `linux-amd64`, `linux-arm64`, `windows-amd64.exe`), verifies size, and replaces running executable in place via POSIX atomic rename and Windows `.old` rename.
  - Added live HTTP endpoint `GET /v1/update/check` so the web playground proactively checks and alerts users when newer releases are published.
- **Native OS Background Service Daemons (`service install | start | stop | status | uninstall`)**:
  - Cross-platform native daemonization enabling 24/7 background operation across reboots with zero open terminals:
    - **macOS**: `~/Library/LaunchAgents/com.abcsteps.bob-gemini-free.plist` with `RunAtLoad` and `KeepAlive` via `launchctl`.
    - **Linux**: `~/.config/systemd/user/bob-gemini-free.service` via `systemctl --user`.
    - **Windows**: `%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup\bob-gemini-free.bat`.
  - Added live CLI subcommand suite: `./bob-gemini-free service [install|start|stop|status|uninstall]`.
- **In-Browser SQLite WASM Database Studio (`🗄️ SQL WASM`)** (historical; removed from the active tree):
  - This earlier client-side experiment is retained here for history only; it
    is not a current gateway database or desktop runtime dependency.
  - Automatic `[ Run SQL ⚡ ]` action card chips on all ` ```sql ` code blocks.
  - Interactive table viewer rendering formatted data with column headers, formatted numbers, row counts, and microsecond execution timer (`⚡ 0.8ms`).
  - Interactive SQL query editor allowing users to run live `SELECT`, `INSERT`, `UPDATE`, and `CREATE TABLE` queries directly in the browser sandbox.
- **Institutional-Grade In-Browser Pyodide WASM Python Execution Engine**:
  - Live client-side CPython 3.11 execution in an isolated in-memory WebAssembly sandbox with **zero server-side execution risk, zero Python installation required, and zero cloud billing**.
  - Dedicated **`🐍 Python WASM Live Sandbox` [ Run WASM ⚡ ]** interactive Artifact Card Chips automatically attached to all Python code blocks.
  - **Dynamic Code Sanitization Engine (`sanitizePythonCode`)**: Automatically strips conversational text, markdown headers, and trailing fences before AST parsing to prevent `SyntaxError` crashes.
  - **Interactive `input()` Prompt Hook**: Intercepts `builtins.input` via browser prompt with clean cancellation handling (`KeyboardInterrupt`) so interactive scripts and financial simulations run smoothly.
  - **Dynamic Scientific Package Auto-Loader**: Inspects code for `numpy`, `pandas`, `matplotlib`, `scipy`, `sympy` and dynamically streams WebAssembly package wheels on demand.
  - **Cyberpunk Dark Terminal & Performance Timer**: Real-time `sys.stdout` streaming, red error highlights, and sub-10ms execution reporting (`✔ Finished in 7.0ms`).
- **Native Interactive Artifacts Canvas Studio (Claude-Class Live Sandbox)**:
  - Automatic code block detection & registration into an in-memory `artifactsRegistry`.
  - Rich **Artifact Card Chips** in chat with title extraction, type badge, and 1-click `Launch ⚡` button.
  - Dedicated **Interactive Artifact Studio Canvas Modal** with sandboxed `iframe` execution (`allow-scripts allow-modals allow-forms allow-same-origin`) supporting live HTML5, CSS3, JavaScript, Canvas 2D/WebGL simulations & games, SVG vector rendering, and Mermaid architecture diagrams.
  - Dual Tab switcher (`[ ▶ Preview | ⟨/⟩ Code ]`), sandbox reload (`⟳`), quick copy (`📋`), standalone file export (`💾`), and instant fullscreen pop-out (`⛶`) in dedicated browser windows.
- **Live AI Prompt Metaprompting Wand Engine (`🪄`)**:
  - Upgraded from static heuristic regex to a **real-time AI prompt optimization engine**: makes a background call to local BOB gateway (`/v1/chat/completions`) using `gemini-3.7-flash` to transform rough thoughts into structured, precision master prompts in ~200ms.
  - Earlier prompt-assistant fallback templates were heuristic local output,
    not provider inference; current README wording labels that boundary.
  - Keyboard shortcut (`⌘ + Shift + P` / `Ctrl + Shift + P`) and sparkling button glow animation.
- **Subtle, Non-Breaking Reading Zoom Controller (`🔍 100%`)**:
  - Dynamically scales reading typography (`calc(0.92rem * var(--reading-zoom))`) on assistant responses, markdown text, thought blocks, and code blocks while keeping header, sidebars, starter cards, and input layout 100% fixed, stable, and unbroken.
  - Integrated 1-click sub-bar cycle pill (`100%` $\rightarrow$ `115%` $\rightarrow$ `130%` $\rightarrow$ `85%` $\rightarrow$ `100%`), Command Palette (⌘K) quick commands, and keyboard shortcuts (`⌘ +` / `Ctrl +`, `⌘ -` / `Ctrl -`, `⌘ 0` / `Ctrl 0`) with `localStorage` persistence.
- **Symmetrical Radha-Krishna $2 \times 2$ 4-Flagship Masterpiece Matrix**:
  - Replaced cramped 8-card grid with a spacious, soothing, balanced $2 \times 2$ matrix showcasing the platform's 4 defining superpowers:
    1. 🎮 **Artifact Canvas Studio**: *2D Cyberpunk Snake Game* (HTML5 Canvas 60fps arcade).
    2. 🐍 **Python WASM Sandbox**: *Quantitative Finance & Risk (Black-Scholes & Monte Carlo)* (NumPy option pricing, 5 Greeks, and 10,000-path VaR).
    3. 🪐 **Three.js WebGL Studio**: *3D Solar System & Orbit Lab* (Photorealistic N-body gravity simulator with 360° OrbitControls).
    4. 🎓 **First Principles & Mental Models**: *Explain to a 5th Grader* (Intuitive analogies with Mermaid visual diagrams).
- **In-Place User Message Editing & Conversation Branching Engine**:
  - True in-place DOM transformation: clicking `✏️` on any past user bubble instantly transforms the card into an inline editor with auto-growing textarea and `[ Save & Submit ]` / `[ Cancel ]` buttons.
  - Full conversation rewind and re-branching (`chatHistory.slice(0, idx)`), restoring any attached multimodal images or document context.
  - Keyboard ergonomics: `Ctrl+Enter` / `⌘+Enter` to submit, `Escape` to cancel.
- **Natural HD Speech Engine & Floating Audio Controller Bar (NotebookLM-Class Voice Studio)**:
  - **Intelligent Neural Voice Scoring Engine (`getBestNaturalVoice`)**: Automatically filters out legacy robotic synths and prioritizes `✨ Studio HD Natural Voices` (`Google US English`, `Google हिन्दी`, `Samantha Enhanced`, `Ava Premium`, `Rishi`, `Lekha`).
  - **Dharmic Glassmorphic Floating Audio Player Bar**: Bottom-docked player equipped with **Play / Pause toggle (`⏸️` / `▶️`)**, **Cycle Speed chip (`0.8x`, `1.0x`, `1.25x`, `1.5x`)**, **Live 4-bar pulsating sound equalizer (`ılılı`)**, **Sentence Progress Counter (`Sentence X of Y`)**, and **Stop (`⏹️`)**.
  - **Sentence-by-Sentence Fluid Cadence**: Chunking parser that prevents browser 15-second speech synthesis timeouts and delivers natural human conversational pacing.
- **Steve Jobs Level Mobile Ergonomics & 2-Tier Input Dock**:
  - `position: fixed !important;` off-canvas drawer architecture for Config and Code panels, eliminating flex margin shifts on mobile viewports.
  - 2-tier mobile input layout: 100% wide auto-resizing textarea on top; tools dock (`📎`, `अ`, `🎙️`, `🪄`) and golden `SEND ➤` CTA on bottom.
  - Guaranteed Apple HIG 44pt touch target compliance, dynamic viewport height (`100dvh`), and notch/home-indicator safe area inset support (`env(safe-area-inset-*)`).

#### Fixed
- **Message Editing Pointer Jump Bug**: Resolved DOM click binding and positional tracking (`data-msg-idx`) so message editing happens strictly in-place without scrolling to the bottom input box.
- **Textarea Horizontal Scrollbar Elimination**: Added `overflow-x: hidden !important;`, `scrollbar-width: none !important;`, and `overflow-x` enforcement in `autoResize(el)` to permanently eliminate unwanted horizontal scrollbar rendering across WebKit and Blink browsers.
- **Pyodide Sync Cancellation**: Added `KeyboardInterrupt` exception handling when user cancels `input()` dialogs.

---

## [0.1.4] - 2026-08-19

### Zero-Download Cloudflare Pages Web Studio, iOS Safari Hardening & Steve Jobs Level Mobile UX

#### Added
- **Bilingual / Multilingual UI Engine (`en` / `hi`)**: Built-in zero-dependency client-side internationalization system (`I18N`) with 1-click header switcher and ⌘K shortcuts (`L1` English, `L2` हिन्दी), translating all headlines, starter cards, telemetry indicators, navigation pills, input placeholders, and modal dialogues dynamically with `localStorage` persistence.
- **Real-Time Client-Side Indic Phonetic Transliteration (Hinglish $\rightarrow$ Devanagari)**: Integrated instant phonetic typing toggleable via `Ctrl+G` or the input dock `अ` button. Space key dynamically converts Roman transliteration (`"aap kaise ho"`, `"namaste"`) into native Devanagari (`"आप कैसे हो"`, `"नमस्ते"`), backed by Google Input Tools API, sub-millisecond in-memory LRU cache, and built-in offline rule dictionary fallback.
- **Multi-Indic Language Ready**: Validated out-of-the-box transliteration support across 10 Indian languages (Hindi, Sanskrit, Marathi, Bengali, Gujarati, Tamil, Telugu, Kannada, Malayalam, Punjabi).
- **School Computer Lab LAN Master Mode**: Documented 1-process LAN host topology (`--host 0.0.0.0 --port 9610`) enabling 30-PC computer labs with 240+ daily students to access local AI at ₹0 cost on <25MB RAM.
- **Zero-Download Cloudflare Pages Serverless Edge Studio**: Deployed native Cloudflare Pages Edge Functions (`/functions/v1/chat/completions.js`, `/functions/v1/models.js`, `/functions/health.js`) executing serverless Web RPC streaming directly in V8 isolates without requiring local binary downloads.
- **BOB Builder Default Theme**: Established high-contrast **BOB Builder** dark developer theme as the primary default across all browsers and devices.
- **Direct GitHub Navbar Integration**: Added top navbar GitHub repository link and icon pill directly in `playground.html` and `web/index.html` for 1-click open-source repository exploration.
- **Public GitHub Raw Install Snippets** (historical): The earlier pipe-to-shell
  bootstrap was later removed from the current distribution in favor of a
  download, inspect, and run flow with signed-release verification.
- **iOS Dynamic Viewport Height (`100dvh`)**: Replaced `100vh` with `100dvh` (Dynamic Viewport Height unit) on `body` — layout correctly adjusts when the iOS Safari keyboard appears or collapses without layout reflow.
- **Compact 2×2 Starter Card Grid on Mobile**: Welcome screen starter cards now render as a 2×2 compact grid on screens ≤860px. Card descriptions are hidden; only icon badge + title shown. Hero section now uses ≤30% screen height (was ~60%).
- **Instant-Access CLEAR Button**: Sub-bar restructured so 📋 Copy and 💾 Export show icon-only on mobile, while 🗑️ **CLEAR** is always visible as a bold red pill (`rgba(239,68,68,0.12)` background) — zero horizontal scrolling needed.
- **44px Touch Target Compliance (Apple HIG)**: All interactive controls in the input dock (`#user-input`, attachment pill, Send button) now meet Apple's minimum 44pt touch target size on all mobile screens.
- **Theme Selector Hidden on Mobile**: Theme selector `<select>` is now hidden on screens ≤640px — all 5 themes remain accessible via ⌘K Menu → Themes section.
- **Mobile-Aware `autoResize()` Textarea**: Textarea auto-resize function now detects mobile (`window.innerWidth ≤ 640`) and applies `minH=40, maxH=130` instead of desktop `minH=44, maxH=200` to prevent keyboard from hiding the chat area.
- **Placeholder Ellipsis**: `#user-input::placeholder` styled with `white-space: nowrap; overflow: hidden; text-overflow: ellipsis` to prevent placeholder text from wrapping to 2 lines on any screen width.
- **⌘K Menu Button Icon-Only on Mobile**: Header ⌘K button now shows only the `⌘` glyph on narrow screens (≤640px), making room for the status dot, panel toggles, and GitHub button without overflow.

#### Fixed
- **Clean HTTP/2 Stream Conclusion**: Engineered Google RPC batch end-marker detection (`["e", ...]` and `["di", ...]`) with explicit upstream `reader.cancel()` and downstream `writer.close()`, eliminating stream hanging and HTTP/2 stream errors.
- **Mobile Privacy & Security Standardization**: Removed background localhost probes from public HTTPS domains to eliminate Private Network Access (PNA) permission dialogs on mobile Chrome/Android and Apple Safari.
- **Instant Token & Word Telemetry Badge**: Resolved stream conclusion lifecycle in `playground.html` so word/token counts and the `SEND ➤` button reset immediately upon generation completion.
- **iOS Safari Auto-Zoom Prevention**: Set `#user-input { font-size: 16px !important }` to prevent iOS Safari 16px auto-zoom trigger when focusing the text input.
- **iOS Safe Area Inset Padding**: All chrome components (header, sidebar, input dock) now respect `env(safe-area-inset-top/bottom/left/right)` for correct notch and home indicator clearance on iPhone X and later.
- **iOS Tap Highlight Removal**: Added `-webkit-tap-highlight-color: transparent` to body and `touch-action: manipulation` to all interactive buttons to eliminate the grey flash on tap and remove 300ms tap delay.
- **Native iOS Inertial Scroll**: Added `-webkit-overflow-scrolling: touch` and `overscroll-behavior-y: contain` to the messages scroll area for native 120Hz ProMotion inertial scrolling and to prevent rubber-band scroll from escaping the container.
- **Version Consistency**: Updated all hardcoded version strings from `v0.1.3` → `v0.1.4` in the About modal, footer status pill, and JavaScript fallback constants.
- **Diagnostic Check 6 (SSE Stream)**: Fixed stream scanner to use `bufio.Scanner` for robust SSE line parsing.
- **Diagnostic Check 12 (Image Generation)**: Now gracefully handles `BardErrorInfo 1003` guest-mode errors with informative pass instead of false failure.

---

## [0.1.3] - 2026-08-19

### Port 9610 Architecture, PSIDTS Session Harvester, Multimodal SDK & Real-Time Test Streaming

#### Added
- **Default Port Migration to 9610**: Moved default listening interface to `http://127.0.0.1:9610` across all binaries, scripts, Docker configurations, and documentation to eliminate port conflicts with common web frameworks.
- **Strict `__Secure-1PSIDTS` Session Token Capture (`--login`)**: Enhanced the 1-click CDP browser harvester (`captureCookies`) to verify and await the rolling `__Secure-1PSIDTS` timestamp token, permanently unlocking Google Scotty multimodal uploads and Vision analysis without `BardErrorInfo [1003]`.
- **Real-Time Event-Driven Diagnostic Streaming (`diag.RunDiagnosticsWithProgress`)**: Enhanced `./bob-gemini-free --test` to emit pass/fail results for each of the 13 diagnostic checks in real time as they finish, with connection body draining to prevent socket exhaustion.
- **Embedded Go Multimodal Methods (`pkg/gateway`)**: Added `GenerateWithMedia` and `GenerateStreamWithMedia` methods on `*gateway.Engine` allowing in-process Go programs to execute multimodal inference with image attachments.
- **Client-Side Document & PDF Ingestion in Playground**: Added client-side file reading in `playground.html` for text files, source code (`.py`, `.js`, `.go`, `.html`, `.css`), markdown, and PDFs directly into the prompt context for 100% Guest Mode access.
- **First-Principles ELI5 Reasoning & Mermaid Visual Guidelines**: Configured Starter Cards and Command Palette (`⌘K` $\rightarrow$ `P1`) to instruct models to emit standard Mermaid vector diagrams (` ```mermaid ... ``` `) and ASCII schematics for structural visualization instead of hallucinated markdown image tags.
- **5-Theme High-Fidelity Showcase Asset Generation**: Captured real-time screenshots of the Apple Light, BOB Builder Orange, Vodafone Red, Spotify Dark, and Quantum Neon themes on port 9610 and compiled them into the master collage in `assets/bob-gemini-free-playground.png`.

---

## [0.1.2] - 2026-08-18

### Real-Time Thinking Stream Splitter, Anthropic Multi-Block Lifecycle & SDK Engine Parity

#### Added
- **Real-Time Thinking Stream Splitter (`ThinkingStreamSplitter`)**: State-machine stream parser isolating ` ```thought\n...\n``` ` blocks on-the-fly during SSE generation:
  - Streams live `reasoning_content` deltas in OpenAI chat streaming (`POST /v1/chat/completions`).
  - Implements strict 2-block Anthropic SSE lifecycle (`thinking` block with `thinking_delta` $\rightarrow$ `content_block_stop` $\rightarrow$ `text` block with `text_delta`).
- **Claude Code Extended Thinking Parameter**: Added dynamic mapping for `req.Thinking` (`enabled` / `budget_tokens`) to Gemini internal thinking modes.
- **Responses API Real-Time Streaming & `reasoning_effort`**: Added live 8-event SSE lifecycle and dynamic reasoning effort parsing ("high"/"medium"/"low") to `POST /v1/responses`.
- **Complete Environment Variable Matrix**: Added `BOB_GEMINI_FREE_COOKIE_POOL_DIR`, `BOB_GEMINI_FREE_LOG_REQUESTS`, `BOB_GEMINI_FREE_RETRY_ATTEMPTS`, `BOB_GEMINI_FREE_RETRY_DELAY_SEC`, `BOB_GEMINI_FREE_REQUEST_TIMEOUT_SEC`, `BOB_GEMINI_FREE_DEFAULT_MODEL`, and `BOB_GEMINI_FREE_AUTH_USER`.
- **Expanded Embedded Go Library (`pkg/gateway`)**: Added `NewEngine` in-process Go programmatic inference (`Generate` / `GenerateStream`) with `WithRetry`, `WithTimeout`, `WithCookiePoolDir`, and `WithVersion` options.
- **Multi-Account Cookie Pool Health Telemetry**: Added `pool_sessions_total` and `pool_sessions_healthy` to health endpoint (`GET /`) and live `--status` CLI dashboard.
- **Multimodal Decoders**: Added native GIF and WebP image decoding support.
- **Tool Result ID Routing**: Multi-turn agent loops now fall back to `msg.ToolCallID` when `msg.Name` is omitted by OpenAI clients.

---

## [0.1.1] - 2026-08-18

### "API-Less AI" Architecture, Token Counting Engine & Multimodal Vision

#### Added
- **"API-Less AI" Architecture**: Articulated and implemented the zero-cloud-bill, zero-credit-card, zero-API-key-leak paradigm across English and Hindi documentation.
- **Native Multi-Script Token Counting Engine**: Added drop-in `POST /v1beta/models/{model}:countTokens` (Google GenAI SDK standard) and `POST /v1/tokens/count` (OpenAI format) with subword, Devanagari/Indic, CJK, Emoji, and multimodal tile calculation.
- **Anthropic Multimodal Vision Translation**: Added native support in `/v1/messages` for Anthropic `type: "image"` content blocks (base64 PNG/JPEG/WEBP) seamlessly translated to Google's Scotty upload protocol.
- **Prompt Caching Usage Telemetry**: Added `cache_read_input_tokens` and `cache_creation_input_tokens` fields to streaming SSE and non-streaming Anthropic responses for complete Claude Code CLI token tracking.
- **Live Financial Savings Telemetry (`GET /`)**: Real-time atomic metrics for `requests_served`, `tokens_processed`, `estimated_savings_usd`, and `uptime_seconds`.
- **13-Point Automated Diagnostic Suite**: Expanded the built-in diagnostic test runner (`--test`) to 13 comprehensive checks.
- **Authenticated Scotty Token Recovery**: Attached `Authorization: SAPISIDHASH` to page token discovery for authenticated session resilience.

---

## [0.1.0] - 2026-08-18

### Initial Release of BOB Gemini Free

Part of the **BOB Series** (*Break Ordinary Boundaries*) by [**ABCsteps.com**](https://abcsteps.com/) and **Divyanshu Singh Chouhan** ([@div197](https://github.com/div197)).

#### Added
- **OpenAI Compatible Gateway**: Drop-in support for `/v1/chat/completions`, `/v1/models`, and `/v1/responses` (OpenAI Codex CLI).
- **Gemini Native API**: Drop-in support for `/v1beta/models/{model}:generateContent` and `:streamGenerateContent` for Gemini CLI compatibility.
- **Full Multimodal Vision**: Active extraction of `image_url` (data URLs and raw base64 payloads) in standard OpenAI requests, converting them into Google WIZ file references using Google's Scotty Resumable Upload protocol.
- **Smart Image Compression Engine**: Built-in downscaling and JPEG optimization (`MaxImageDimension = 1024`, `DefaultJPEGQuality = 75`, `MaxImageByteSize = 1MB`) to prevent upstream payload rejection.
- **Real-Time SSE Streaming**: Native line-by-line delta streaming from Google's `rt=c` BoQ stream response with automatic chunk deduplication.
- **Dynamic Reasoning Controls**: Support for `@think=N` model suffix overrides (e.g. `gemini-3.6-flash@think=0` for deep step-by-step thinking tokens up to 20k+ chars).
- **Pro Model Unlock**: Integration with paid **Google AI / Gemini Advanced ($20/mo)** subscriptions via local session cookie caching (`cookie.txt`).
- **Simulated Function Calling**: Robust prompt injection and markdown regex extraction for tool calling (` ```tool_call ` and ` ```function_call `).
- **TLS Browser Impersonation**: Integrated `tls-client` supporting Chrome, Firefox, and Safari TLS fingerprints for datacenter WAF bypass.
- **Security Hardening**:
  - Default network binding locked to `127.0.0.1`.
  - Constant-time API key verification (`crypto/subtle`).
  - Comprehensive `.gitignore` protecting credentials (`cookie.txt`, `config.json`).
- **High-Performance Static Binary**: Built with pure Go, zero runtime dependencies, and <15MB baseline RAM consumption.
- **Native Reasoning Content Extraction**: Isolated `reasoning_content` extraction for OpenAI Thinking models, powering collapsible reasoning visualizers in Cursor, Cherry Studio, ChatBox, and OpenWebUI.
- **Developer Convenience Model Aliases**: Added intuitive shortcuts (`gemini-pro`, `gemini-flash`, `gemini-thinking`, `gemini-lite`, `gemini-2.5-pro`, `gemini-2.5-flash`).
- **High-Resolution Visual Assets**: Added official cybernetic hero banner and app icon in `./assets/`.
- **Zero-Friction Cross-Platform Installers**: Added `install.sh` for macOS/Linux, `install.ps1` for Windows, and automated `Makefile` with multi-arch cross-compilation (`make dist`).
- **Multilingual Documentation**: Added comprehensive Hindi guide ([`README.hi.md`](README.hi.md)).
- **Automated Cookie Setup Helper**: Added `--setup-cookie` and `--cookie-string` CLI commands to automatically extract, validate, and securely store (`chmod 0600`) Gemini Advanced session cookies.
- **Architectural Workflow Diagram**: Added comprehensive dataflow and system architecture visual (`assets/bob-gemini-free-architecture.jpg`).
- **Automated Diagnostic Test Kit**: Built-in CLI `--test` flag and standalone scripts (`test-kit.sh`, `test-kit.ps1`) executing full-spectrum automated validation across all 9 endpoint and model scenarios with millisecond latency telemetry.
- **Throughput & Concurrency Benchmark Runner**: Integrated `--bench` flag and `scripts/bench.sh` runner measuring requests/sec, tokens/sec, and P50/P90 latencies.
- **Background Daemon & OS Service Units**: Included Linux Systemd unit (`scripts/bob-gemini-free.service`), macOS Launchd plist (`scripts/com.abcsteps.bob-gemini-free.plist`), and Windows batch runner (`scripts/start-service.bat`).
- **Native Anthropic Messages API Engine (`/v1/messages`)**: Direct drop-in support for **Claude Code CLI** (`ANTHROPIC_BASE_URL=http://127.0.0.1:9610`) and Anthropic SDKs with complete SSE event streaming (`message_start`, `content_block_delta`, `message_delta`, `message_stop`).
- **OpenAI Image Generation Engine (`/v1/images/generations`)**: Native support for DALL-E / Imagen image generation requests with automatic markdown image URL extraction and base64 encoding.
- **Embedded Go Library (`pkg/gateway`)**: Exported Go package enabling in-process gateway instantiation inside any Go backend or agent runtime.
- **Zero-Config Cookie Auto-Discovery**: Automatic detection and loading of `./cookie.txt` and `~/.config/bob-gemini-free/cookie.txt`.
- **Responses API `output_text` Field**: Top-level field added to Responses API output objects for direct property access across official JavaScript/Python SDKs.
- **OpenAI Observability & Rate Limit Headers**: Automatic response injection of `x-request-id`, `openai-processing-ms`, `openai-version`, and `x-ratelimit-*` headers.
- **Frontier & Codex Model Alias Catalog**: Complete transparent mapping for `gpt-5.6`, `gpt-5.5`, `gpt-5.4`, `gpt-5-codex`, `claude-3-7-sonnet`, `claude-code`, `o3`, `o4-mini`, and `o1`.
- **Unit Test Suite**: 100% passing automated test suite covering all 7 packages, including agentic multi-turn tool loops and Codex Responses API.
- **13-Question Builder FAQ**: Comprehensive troubleshooting and architectural comparison across English and Hindi documentation.
- **Acknowledgements & Research Foundations**: Added formal citations crediting Google Research for the Transformer architecture (*"Attention Is All You Need"*).
- **1-Click Native Interactive Login Window (`--login`)**: Standalone browser window captures Google session tokens via Chrome DevTools Protocol (CDP) WebSocket, bypassing manual DevTools copying and macOS Keychain dialogs.
- **Multi-Account Cookie Pool Engine (`cookie_pool`)**: High-concurrency round-robin dispatcher supporting multiple accounts (`./cookies/*.txt`), atomic lock-free cursors, 60s failure backoff, and transparent 429 rate-limit failover.
- **Dynamic Token Self-Healing**: Live extraction of `SNlM0e` (XSRF token) and `cfb2h` (Google build version) directly from session HTML for zero-maintenance resilience.
- **Claude 3.7 / 3.5 Extended Thinking Support**: Intercepts `thinking: { type: "enabled", budget_tokens: N }` in `/v1/messages` and emits official `type: "thinking"` blocks alongside text and tool blocks.
- **Search Grounding & Web Citation Extraction**: Structured extraction and markdown footnoting of live Google search grounding sources.
- **Google Imagen 3 & Gemini Nano Banana 2 / Pro Models**: Full registration and mode routing for `imagen-3`, `imagen-3-fast`, `gemini-nano-banana`, `gemini-nano-banana-2`, `gemini-nano-banana-pro`, `dall-e-3`, and `dall-e-2`.
- **Zero-Dependency Universal Installers**: Enhanced `install.sh` and `install.ps1` with automated OS/architecture detection and fallback downloading of pre-compiled binaries from GitHub Releases for machines with no Go or Python.
- **12-Point Enterprise Diagnostic Suite**: Expanded diagnostic verification suite (`--test`) covering all 12 live end-to-end capabilities with millisecond latency logging.
- **Docker & OrbStack Native Healthchecks**: Injected native Docker `HEALTHCHECK` with 20s interval and <3ms cold-boot optimization on OrbStack.
- **Comprehensive Master Documentation Suite (`docs/`)**: 5 structured chapters (18 dedicated markdown guides) covering quickstart, zero-dependency setups, authentication pools, IDE integrations, API references, and embedded Go SDK.
