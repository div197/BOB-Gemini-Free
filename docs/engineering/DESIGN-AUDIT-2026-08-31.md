# BOB Gemini Free — Product Design & Correctness Audit

**Date:** 2026-08-31 (Asia/Kolkata)

**Workspace:** `/Users/apple31/Documents/BOB-Gemini-Free`

**Audited surface:** local and hosted-capable Web Studio (`internal/server/playground.html`, generated `web/index.html`)

**Git baseline:** `523ceeb` (`origin/main`); this design-audit continuation is reviewed through `5eae3e2` on `codex/release-readiness-v0.2.0`, with the generated `web/index.html` synchronized from the source studio.

**Audit status:** source and regression checks complete; interactive browser/viewport evidence blocked in this session

## Executive conclusion

BOB has a credible product core: a local AI workbench that lets students and
developers ask questions, inspect reasoning, generate code, and run small
artifacts through a local gateway. The current visual system is expressive and
memorable, but it sends too many signals at once: engineering console, sacred
geometry, glassmorphism, cyberpunk arcade, and a gallery of named theme
references. That is the main design-ceiling problem, not a lack of decoration.

This pass completed the unconditional correctness fixes that were deterministic
from source: browser text scaling is available again, dialogs have explicit
semantics and a shared focus lifecycle, click-only core controls are keyboard
reachable, form controls have accessible names, reduced-motion behavior is
defined, onboarding uses the shared dialog surface, and all built-in theme text
tokens now pass the normal-text contrast check against their declared surfaces.

No navigation model or design language was changed by the design audit. Later
release-readiness follow-ups separately added updater permission diagnostics,
Developer API tool-argument validation, desktop update-check jitter,
credential-safe public error handling, and nil-safe logging for partial server
and Gemini-client construction. Those changes are recorded in the failure
register and release docs. The result is not yet a visual release sign-off
because the real browser session required for desktop, tablet, phone,
keyboard, focus, and artifact walkthroughs was unavailable.

This continuation adds a narrow L1 correctness pass at `7efb6b7`: every static
and generated button declares `type="button"`, the brand/model/starter/token
and artifact controls are native buttons, collapsed drawers are removed from
the interaction tree with `inert`, drawer triggers expose expanded state,
modal surfaces are bounded and scrollable at short viewports, the primary
prompt has a skip link, and theme/language selectors have accessible names.
The changes preserve the navigation model and feature set; they do not turn
source-level proof into browser proof.

## 1. Product intent in three lines

1. **Who:** students and developers who want one approachable place to learn,
   ask, reason, write code, and experiment with generated artifacts.
2. **What:** a local-first AI workbench and protocol bridge, with a browser
   studio over the BOB gateway rather than a conventional account-based SaaS
   chat product.
3. **Character:** a premium engineering instrument with a teacher's bench;
   calm, trustworthy, educational, and precise. Playful artifact previews are
   secondary moments, not the identity of the whole application.

## 2. Taste baseline derived from the repository

At its best, BOB should feel like a quiet, capable workshop:

- one obvious path from connection status to prompt to **Send** to result;
- a dark neutral canvas with one purposeful BOB amber action color;
- strong reading typography and generous breathing room around generated work;
- advanced model, protocol, language, tokenizer, glossary, and gateway controls
  revealed progressively instead of competing with the prompt;
- status language that tells the user what to do next, not how impressive the
  implementation is;
- visual effects reserved for state or affordance; emoji and neon used when
  they carry meaning, not as generic labels;
- a stable, predictable interaction model across mouse, keyboard, touch, and
  a 200% reading zoom.

This baseline follows the product's actual intent and evidence boundaries. The
i18n document explicitly says that English and Hindi are present while the
other Indic entries are selector targets or provider-dependent, not a complete
offline eight-language UI (`docs/engineering/MULTI-LINGUAL-I18N-SYSTEM.md:3-6`,
`50-87`). The README likewise describes the Studio as local-first and
provider/session dependent (`README.md:369-401`). Those are stronger product
signals than the aspirational adjectives in the CSS comments.

## 3. Ranked findings by impact

| Rank | Impact | Finding | Authority | Decision / status |
|---|---|---|---|---|
| 1 | Release blocker | No controlled desktop/tablet/phone browser walk could be completed: the in-app browser runtime reported no available browser and an empty browser list. | Evidence gate | **Open.** Do not call the current visual state release-ready until the matrix in section 8 is run. |
| 2 | High | Several dialogs had no explicit dialog semantics or focus contract; the dynamically-created hosted onboarding surface used `.modal-overlay`/`.modal-content`, which were not defined by the current stylesheet. | L1 | **Fixed.** Shared dialog markup, labels, `aria-modal`, Escape behavior, focus trap, and focus return are now explicit. |
| 2a | High | Core click-only controls used generic `role="button"` elements and button templates omitted an explicit non-submit type. | L1 | **Fixed in source.** Brand, model, starter, token-estimate, artifact-rail, static, and generated controls now use native buttons; a source regression scans every button opening tag. |
| 2b | High | Collapsed configuration/code drawers were visually hidden but did not explicitly leave the keyboard and assistive-technology interaction tree. | L1 | **Fixed in source.** Drawer panels now synchronize `aria-hidden`/`inert` with `aria-expanded` trigger state. Live drawer focus and coverage still require the browser matrix. |
| 2c | Medium | Modal surfaces could clip long content at short viewports or large browser zoom. | L1 | **Fixed in source.** Shared dialog surfaces use a bounded dynamic viewport height and internal vertical scrolling. Rendered 200% zoom evidence remains pending. |
| 3 | High | The viewport disabled browser text scaling with `maximum-scale=1.0` and `user-scalable=no`. This directly contradicted the reading-zoom and accessibility intent. | L1 | **Fixed.** The scale lock was removed while preserving `viewport-fit` and keyboard-resize behavior. |
| 4 | High | `--text-subdued` failed normal-text contrast in the default theme and several selectable themes; muted text also failed in the Tokyo Night and Solarized Light themes. | L1 | **Fixed.** Theme token values were corrected and covered by a contrast regression test. |
| 5 | High | The responsive shell has enough density to make overlap, drawer coverage, composer collision, and New Chat visibility a real risk at tablet and phone widths. Source CSS contains mitigation, and drawer semantics are now explicit, but no live viewport evidence was available. | L2 evidence gate | **Open.** Verify before restructuring; keep the change isolated if a section-level fix is needed. |
| 6 | Medium | The top area gives similar visual weight to status, configuration, code, command menu, GitHub, glossary, theme, language, model, zoom, export, and clear. This makes the prompt workspace compete with the product chrome. | L2 | **Proposal only.** Test a section-level hierarchy change with before/after captures. |
| 7 | Medium | The Studio depends on multiple remote CDN scripts and styles for markdown, math, diagrams, document parsing, OCR, and runtime helpers. That is a reliability boundary for a product described as local-first and immediate. | L3 | **Proposal only.** Define explicit degraded states or package only the critical path; do not invent fake offline fallbacks. |
| 8 | Medium | `playground.html` is a monolithic 11,674-line document containing tokens, layout, markup, runtime state, rendering, and feature logic. It makes visual regressions and ownership boundaries difficult to see. | L3 | **Proposal only.** Split by component/build boundary only after behavior and generated-artifact parity are protected. |
| 9 | Low | The interface uses several metaphors simultaneously: “Sacred Geometry,” “Celestial Glassmorphic,” “Cyberpunk,” “Editorial,” “Quantum,” and branded theme references. This is memorable but not yet a coherent product language. | L3 | **Proposal only.** Keep selectable themes for now; establish one canonical BOB default first. |

## 4. What was improved and why

### L1 — completed in this audit

| Improvement | Why it matters | Evidence |
|---|---|---|
| Promoted core click-only surfaces to native controls | Native buttons provide correct keyboard activation, focus behavior, and assistive-technology semantics without a global keydown shim. Explicit `type="button"` also prevents accidental form submission if the shell is embedded later. | `internal/server/playground.html:5068-5354,11091-11110`; `TestPlaygroundUsesNativeControlAndDrawerSemantics` |
| Isolated collapsed drawers from the interaction tree | A visually translated panel must not leave hidden controls reachable by Tab or screen readers. The trigger and panel now expose one synchronized state. | `internal/server/playground.html:5097-5101,5136,5348,6716-6742`; `TestPlaygroundUsesNativeControlAndDrawerSemantics` |
| Added prompt skip link and selector names | Keyboard users can reach the primary task directly, and theme/language controls have names that do not depend on title tooltips. | `internal/server/playground.html:4020,5114-5127`; `TestPlaygroundUsesNativeControlAndDrawerSemantics` |
| Bounded shared dialog surfaces | Long onboarding, gateway, glossary, and instruction content can scroll inside the modal instead of clipping against a short viewport or 200% zoom. | `internal/server/playground.html:2776-2784,2867-2875`; `TestPlaygroundUsesNativeControlAndDrawerSemantics` |
| Restored missing semantic surface aliases | Theme-independent `--bg-hover` and `--bg-main` references now resolve to the active theme tokens instead of producing invalid computed declarations. | `internal/server/playground.html:74-76`; `TestPlaygroundUsesNativeControlAndDrawerSemantics` |
| Removed the mobile scale lock | Users can use browser accessibility zoom and platform text enlargement without the page refusing it. | `internal/server/playground.html:10`; `TestPlaygroundAccessibilityFloorIsExplicit` |
| Added global visible focus styling | Keyboard users can see the active control across native buttons, links, fields, and role buttons. | `internal/server/playground.html:839-855` |
| Added reduced-motion behavior | The product remains usable for people who request less motion, including dialog, artifact, and micro-interaction animations. | `internal/server/playground.html:849-855` |
| Normalized static and dynamic dialogs | Dialog surfaces now expose `role="dialog"`, `aria-modal`, `aria-hidden`, and an accessible label. | `internal/server/playground.html:3940-4483` |
| Added shared modal lifecycle | Opening stores the trigger, focuses the intended first control, traps Tab/Shift+Tab, handles Escape, and returns focus on close. | `internal/server/playground.html:7373-7433` |
| Repaired hosted onboarding surface | The onboarding popup now uses the existing dialog design system, wraps long commands, and remains bounded by the viewport. | `internal/server/playground.html:7737-7784` |
| Made secondary click-only controls keyboard-usable | Reasoning headers, glossary help affordances, and tokenizer pills can still be activated with Enter or Space; primary surfaces now use native buttons instead of relying on the global shim. | `internal/server/playground.html:7545-7557`, secondary role-button renderers; `TestPlaygroundAccessibilityFloorIsExplicit`, `TestPlaygroundUsesNativeControlAndDrawerSemantics` |
| Completed keyboard semantics for secondary controls | Command-palette options expose listbox/option state and active descendant tracking; the brand, glossary help affordances, and tokenizer pills are keyboard-operable. | `internal/server/playground.html:6782-6915`, `7435-7444`, `4969-5090`, tokenizer renderer |
| Added accessible names and label associations | Composer tools, gateway fields, provider key controls, model settings, tokenizer search, audio actions, and message actions no longer rely on emoji or title text alone. | `internal/server/playground.html:5043-5254`, dynamic message renderers |
| Removed accidental Enter-to-close behavior in About | Enter should activate the focused control, not dismiss an unrelated dialog. | `internal/server/playground.html:7436-7443` |
| Corrected all built-in subdued-text tokens | Small labels and metadata retain hierarchy while meeting a 4.5:1 normal-text contrast threshold on declared app/card/modal/input surfaces. | `internal/server/playground.html:46-300`; `TestPlaygroundThemeTextContrast` |
| Removed duplicate legacy attachment dispatch | A file drop no longer enters both the old unbounded `FileReader` path and the bounded universal extractor; one event now gets one bounded parse lifecycle. | `internal/server/playground.html:6221-6230`, `9477-9868`; `TestAttachmentParsingIsBoundedAndCancellable` |
| Removed duplicate code-copy declaration | The Studio now has one authoritative `copyCode` implementation, preventing silent function shadowing in the monolithic script. | `internal/server/playground.html:10295-10302`; `TestPlaygroundHasOneCodeCopyHandler` |
| Centralized clipboard writes with failure feedback | Restricted WebViews, `file://` contexts, denied permissions, and transient clipboard failures no longer produce silent exceptions or false success toasts; all Copy actions use one guarded helper. | `internal/server/playground.html:6524-6545`; `TestClipboardActionsHaveExplicitFailureState` |
| Explained permission-denied desktop staging failures | A managed Mac or protected install location now receives recovery guidance to move BOB to a writable location or grant the current user access, instead of an opaque `permission denied` error. | `internal/updater/desktop_stage.go:173-181`; `TestDesktopStagingDirectoryErrorExplainsPermissionDeniedInstall` |
| Regenerated the static distribution | The served/static artifact remains source-parity with the edited studio. | `web/index.html`; `make web`; parity check passed |
| Kept partially embedded logging safe | A host that enables request logging before attaching an optional logger no longer panics on a health request. | `internal/server/middleware.go:220-228`; `TestPartialAppWithRequestLoggingDoesNotPanic` |
| Kept optional Gemini retry logging safe | A partially constructed upstream client now returns the original retry failure instead of panicking when a retry logger is absent, for both buffered and streaming generation. | `internal/gemini/client.go:75-81,533,673`; `TestGenerateRetryWithNilLoggerDoesNotPanic`, `TestGenerateStreamRetryWithNilLoggerDoesNotPanic` |
| Enforced one Markdown-link protocol policy | Generated assistant links now pass through an explicit `http:`, `https:`, `mailto:`, or `tel:` allow-list; unsupported schemes become inert before rendering, and native/hosted external-link routing uses the same policy. | `internal/server/playground.html:5351-5368,5372-5406,8739-8748`; `TestMarkdownLinksUseStrictProtocolWhitelist` |
| Reduced artifact launch to one accessible action | The visual artifact card is now a named group; only its explicit `type="button"` launch control activates the preview, so the card no longer advertises a second click-only action or misleading pointer affordance. | `internal/server/playground.html:1966-1982,8684-8717`; `TestArtifactLaunchChipUsesOneKeyboardAction` |
| Made the optional Developer API route fail closed | Enabling the route without a key no longer leaves a misleading checked state, and clearing a key during an active provider session explicitly returns to the default route. | `internal/server/playground.html:6377-6412`; `TestDeveloperAPIRouteToggleFailsClosedWithoutKey` |
| Required safe transport for provider keys | Loopback HTTP remains available for the local gateway, while non-loopback endpoints require an explicit save and HTTPS; native context no longer bypasses that transport rule. | `internal/server/playground.html:6323-6348`; `TestDeveloperAPIRouteRequiresSafeGatewayTransport` |
| Preserved gateway status after upstream responses | The Studio now marks the gateway online as soon as it receives an HTTP response and only shows it offline when the request fails before a response; provider/HTTP/stream errors no longer erase accurate local connectivity state. | `internal/server/playground.html:11168,11373-11376,11643`; `TestReachableGatewayIsNotShownOfflineAfterHTTPOrStreamFailure` |
| Removed inline attachment-ID handlers | Attachment preview/remove actions now use escaped data attributes and delegated listeners, so restored local-history IDs cannot become JavaScript source during shelf rendering; both controls have explicit button types and names. | `internal/server/playground.html:7582-7603,9800-9804`; `TestAttachmentControlsDoNotEmbedUntrustedIDsInInlineJavaScript` |
| Escaped attachment metadata and hardened image previews | Persisted attachment icons are escaped before history HTML insertion. User image previews are now keyboard/touch buttons, accept only base64 raster data URLs, and present an explicit unavailable state for unsupported formats; preview navigation uses `noopener,noreferrer`. | `internal/server/playground.html:5448-5462,7634-7640,9911-9922,11054-11298`; `TestPersistedAttachmentIconsAreEscapedBeforeHistoryHTML`, `TestAttachmentImagePreviewsUseAccessibleRasterOnlyControls` |
| Made gateway recovery a real action control | Generation error copy no longer embeds a `javascript:` URL for Config. The recovery affordances are named buttons routed through the delegated action path, so they retain keyboard semantics and the strict external-link policy. | `internal/server/playground.html:7640-7643,11733-11751`; `TestErrorRecoveryConfigActionsAvoidJavaScriptURLs` |

The design audit deliberately did not change the Gemini wire protocol, provider
routing, authentication, gateway CORS policy, streaming behavior, artifact
sandbox permissions, updater selection or replacement semantics, or desktop
packaging. Subsequent hardening commits are intentionally tracked as separate
concerns: updater permission guidance, desktop check staggering, direct
Developer API argument validation, public-error redaction, and partial-app
logging safety. Real managed-device permissions and rendered behavior remain
external acceptance gates.

## 5. L2 work: isolated redesigns that need browser evidence

No L2 section restructure was implemented in this pass because the required
before/after browser captures could not be produced. The following are the
recommended isolated experiments, each in its own revertible change:

### L2-A — Header and command hierarchy

Keep the brand, gateway status, settings, and one compact command affordance in
the primary header. Move low-frequency GitHub, glossary, theme, language, and
diagnostic actions behind progressive disclosure. Preserve existing keyboard
shortcuts and URLs.

Acceptance evidence: desktop 1440x900, tablet 1024x768, phone 390x844; no
horizontal overflow; gateway status remains visible; New Chat and Send remain
reachable without opening a panel.

### L2-B — Chat workspace and sidebars

Treat the conversation and composer as the only primary workspace. Make the
left/right configuration drawers behave as true overlays on narrow widths with
an obvious close/return path, while preserving the current model and code
features. Keep the composer offset driven by measured height, not a guessed
constant.

Acceptance evidence: open each drawer while the conversation is scrolled,
start a new chat, stream a long response, use the top and bottom scroll
controls, rotate a phone viewport, and verify that the composer never covers
the final response.

### L2-C — Artifact studio

Keep Preview and Code as two clear modes with a single source of truth. Make
loading, empty source, runtime error, source-too-large, and unavailable-library
states visible and actionable. Preserve the existing sandbox boundary and
editor draft recovery behavior.

Acceptance evidence: generated HTML artifact, malformed artifact, empty editor,
large source, runtime exception, Preview/Code switching, reload, save, copy,
close, Escape, and focus return at desktop/tablet/phone widths.

## 6. L3 proposals — do not implement as incidental cleanup

1. **Adopt one canonical BOB design language.** Make “trusted local AI
   workbench / teacher's bench” the default. Keep alternate themes as an
   explicit experimental gallery, but stop letting “sacred,” “celestial,”
   “glassmorphic,” and “cyberpunk” describe the same core surface. The trigger
   for this proposal is the mismatch between the product intent above and the
   competing metaphors in `internal/server/playground.html:42-45`, `1300-1304`,
   and `3828-3846`, plus the theme catalog in `README.md:408-414`.
2. **Reduce capability-list copy in the product surface.** Rewrite labels to
   help a student act: “Connect gateway,” “Choose model,” “Add image,” “Run
   artifact,” and “Ask glossary.” Keep implementation detail in the glossary
   and engineering docs. The i18n document's explicit status language is the
   right model for claims (`docs/engineering/MULTI-LINGUAL-I18N-SYSTEM.md:3-21`).
3. **Create a token policy, not just a token collection.** Keep one neutral
   surface scale, one primary action color, one danger color, two radius sizes,
   and a documented contrast test. The current token block is a good starting
   point, but the previous subdued-text failures prove that names alone do not
   guarantee a usable system.
4. **Split the monolith behind a stable generated boundary.** Extract tokens,
   shell, chat, dialogs, artifact studio, and feature modules while preserving
   `go:embed`, `make web`, exact runtime behavior, and source/static parity.
   This is a maintainability project, not permission to redesign protocol or
   feature behavior.
5. **Define an honest offline/degraded contract.** The current CDN-backed
   renderers should expose a clear “feature unavailable until its library
   loads” state, with retry and source access. Do not silently fabricate model
   output or claim that a browser is fully offline when its critical libraries
   are network-fetched.

## 7. Correctness table

“Source-fixed” means a deterministic source invariant and regression test now
exist. It does not mean a real browser interaction has passed. “Browser
pending” is intentional and is the correct status while the browser runtime is
unavailable.

| Page / path | Viewport | Issue | Severity | Status |
|---|---|---|---|---|
| Studio shell `/playground` | Desktop 1440x900 | Full visual walk, hierarchy, clipping, and layout shift | High | **Browser pending**; no browser was available |
| Studio shell `/playground` | Tablet 1024x768 and 768x1024 | Drawer coverage, New Chat visibility, header wrapping, composer collision | High | **Browser pending**; responsive CSS exists but is not acceptance evidence |
| Studio shell `/playground` | Phone 390x844 and 320x844 | Narrow header, two-column starter cards, touch targets, safe-area behavior | High | **Browser pending**; source has phone rules, visual proof absent |
| Page typography | All | Browser text scaling was disabled | High | **Source-fixed**; scale lock removed and regression-tested |
| Primary prompt/composer | All | Accessible name, focus visibility, loading/stop/error recovery | High | **Partially source-fixed**; native control/selector tests pass, browser state walk pending |
| Chat scrolling | All | Top/bottom navigation, streaming bottom anchor, composer offset | High | **Source-supported** by existing regression tests; real scroll walk pending |
| Starter cards | Desktop/tablet/phone | Click-only activation excluded keyboard users and used invalid generic control semantics | Medium | **Source-fixed** with native buttons and phrasing-content markup; browser keyboard walk pending |
| Model chip and token estimate | Desktop/tablet/phone | Click-only controls lacked native button semantics | Medium | **Source-fixed**; browser keyboard walk pending |
| Dialogs and command palette | All | Missing semantics, focus trap/return, consistent Escape behavior, and short-viewport clipping | High | **Source-fixed in source**; browser focus-trap, return, and 200% zoom test pending |
| Hosted onboarding dialog | All | Undefined modal classes and long command content could render poorly | High | **Source-fixed**; browser resize/wrap test pending |
| Configuration/code drawers | Tablet/phone | Dense controls, panel coverage, hidden-panel focus, and compact actions may be difficult to operate by touch | High | **Partially source-fixed**; `aria-hidden`/`inert` state is explicit, but coverage and target-size evidence require browser |
| Gateway/configuration panel | Tablet/phone | Dense controls and compact actions may be difficult to operate by touch | Medium | **Open**; assess target size and panel overflow in browser |
| Artifact Preview/Code | All | Empty/error/loading/source-limit states and generated game launch | High | **Open for rendered acceptance**; source-level sandbox/recovery tests pass |
| Language selector and Hindi UI | All | Translation coverage and text expansion | Medium | **Open**; docs correctly classify coverage as English/Hindi plus partial targets |
| Theme surfaces | All built-in themes | Subdued/muted text contrast | High | **Source-fixed** with numeric contrast regression; browser visual sampling pending |
| External GitHub/release links, generated Markdown links, and Copy actions | All | Default-browser handoff, link protocol safety, and clipboard success/failure feedback | Low | **Source-supported** by explicit link allow-listing, real anchors/`noopener`, and guarded clipboard writes; click behavior not browser-verified here |
| Hosted Studio → localhost gateway | N/A | Origin/PNA trust is a security boundary, not a visual feature | High | **Separate release gate**; not changed by this design pass |

## 8. Required browser acceptance run

When a browser runtime is available, run the following exact matrix and attach
before/after screenshots for every L2 change:

| Path | Desktop | Tablet | Phone |
|---|---|---|---|
| Cold load, gateway offline, onboarding, close | 1440x900 | 1024x768 | 390x844 |
| New Chat → prompt → streaming response → stop/error recovery | 1440x900 | 1024x768 | 390x844 |
| Scroll top/bottom during a long response | 1440x900 | 768x1024 | 390x844 |
| Open left and right panels, change model, close/return | 1440x900 | 1024x768 | 390x844 |
| Open Gateway, About, Glossary, Tokenizer, Command Palette | 1440x900 | 1024x768 | 390x844 |
| Generate artifact, Preview/Code, empty/error/reload/save/copy | 1440x900 | 1024x768 | 390x844 |
| Hindi switch, 200% browser zoom, keyboard-only path | 1440x900 | 1024x768 | 390x844 |

For each path verify: no horizontal scroll, no clipping, visible focus, Tab
containment and focus return, usable Escape behavior, 24px minimum pointer
target at absolute minimum, comfortable touch targets for the primary actions,
stable layout while content streams, explicit loading/empty/error/disabled
states, and no lost New Chat or Send action.

## 9. Verification run

The initial design-correctness pass completed on clean source commit
`7efb6b7`; the continuation hardening was then completed through `8651eba`:

```text
make web                                      PASS
go test -count=1 ./...                        PASS
go test -race ./...                           PASS
go vet ./...                                  PASS
go build ./...                                PASS
go mod verify                                 PASS
git diff --check                              PASS
inline playground JavaScript syntax check    PASS
TestPlaygroundUsesNativeControlAndDrawerSemantics PASS
scripts/verify-release-source.sh              PASS
```

`make web` regenerated the checked-in static bundle from the source studio.
The source tests verify native button semantics, drawer state markers, modal
height bounds, theme aliases, and the existing focus/contrast contracts. A
real-browser connection was attempted for this audit, but the configured
browser runtime reported no available browser and an empty browser list. No
layout, keyboard focus, touch, rendering-library, artifact, or provider claim
is upgraded on the basis of the source checks.

No GitHub Actions workflow was added or invoked. No release was tagged or
published by this audit. No credential, API key, cookie, or signing key is
stored in this document or source change.

## 10. Files changed and deliberately untouched

Changed:

- `internal/server/playground.html` — L1 accessibility, modal, keyboard,
  reduced-motion, onboarding, contrast-token, attachment-dispatch, image
  preview, duplicate-handler, and native-control/drawer-semantics corrections
  (the current continuation is commit `7efb6b7`).
- `internal/server/playground_test.go` — source-level accessibility, theme
  contrast, single-dispatch attachment, clipboard failure-safety, Markdown
  link-protocol, artifact-action, and Developer API route-state regression
  tests, including persisted-attachment, image-preview, and native-control /
  drawer-state boundaries.
- `internal/updater/desktop_stage.go` — actionable permission-denied staging
  error classification plus a no-download install-location preflight.
- `internal/updater/desktop_stage_test.go` — read-only, permission-denied,
  writable-bundle, App Translocation, and unsupported-platform preflight tests.
- `cmd/desktop/updates.go` — defer automatic prompts and explain an
  unwriteable/translocated install before staging an artifact.
- `internal/gemini/flight.go` — isolate subscriber cancellation, bound
  subscriber queues/history, and cancel abandoned shared runners.
- `internal/gemini/flight_test.go` — leader/follower cancellation,
  deadline, overflow, history-limit, and deterministic race regressions.
- `internal/multimodal/upload.go` — validate fetched image decode dimensions
  before remote bytes reach upload or image-generation paths.
- `internal/multimodal/multimodal_test.go` — highly-compressible oversized
  image decode-budget regression.
- `internal/server/middleware.go` — optional request-logger nil guard for
  partial embedded apps.
- `internal/server/server_test.go` — partial-app request-logging regression.
- `internal/gemini/client.go` — nil-safe optional retry logger for buffered and
  streaming upstream retries.
- `internal/gemini/client_test.go` — retry regressions for an absent logger.
- `web/index.html` — generated static distribution synchronized by `make web`.
- `docs/engineering/FAILURE-REGISTER-100.md` — refreshed branch/main evidence
  and the stream/multimodal failure-path status.
- `docs/engineering/DESIGN-AUDIT-2026-08-31.md` — this audit.

Deliberately untouched:

- Gemini payload construction, SAPISID authentication, cookie/session routing,
  thinking extraction, Scotty wire upload, and model-mode mapping.
- OpenAI, Anthropic, and Google adapter wire formats.
- Gateway CORS/PNA policy, provider routing, API-key handling, updater release
  selection/replacement semantics beyond the new local staging preflight,
  Wails/Tauri history, release assets, and GitHub publication state.
- Navigation model, feature removal, theme catalog, and component/build
  restructuring; these are L2/L3 decisions requiring evidence or explicit
  approval.

## 11. Next decision

The next responsible step is not more visual polish. Recover the real browser
runtime, run section 8, and complete the clean `/Applications` updater proof.
If either run exposes drawer, artifact, touch, or transaction failures, fix
those as independent regression-locked changes before considering the L3
design-language or monolith split proposals.
