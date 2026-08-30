# BOB Gemini Free — Product Design & Correctness Audit

**Date:** 2026-08-31 (Asia/Kolkata)

**Workspace:** `/Users/apple31/Documents/BOB-Gemini-Free`

**Audited surface:** local and hosted-capable Web Studio (`internal/server/playground.html`, generated `web/index.html`)

**Git baseline:** `523ceeb` (`origin/main`); design-audit source snapshot is `9ac2d87`; current reviewed source-hardening tip is `88d68dd` on `codex/release-readiness-v0.2.0` (the branch also contains subsequent audit-documentation commits)

**Audit status:** source and served-runtime checks complete; interactive browser/viewport evidence blocked in this session

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
| 3 | High | The viewport disabled browser text scaling with `maximum-scale=1.0` and `user-scalable=no`. This directly contradicted the reading-zoom and accessibility intent. | L1 | **Fixed.** The scale lock was removed while preserving `viewport-fit` and keyboard-resize behavior. |
| 4 | High | `--text-subdued` failed normal-text contrast in the default theme and several selectable themes; muted text also failed in the Tokyo Night and Solarized Light themes. | L1 | **Fixed.** Theme token values were corrected and covered by a contrast regression test. |
| 5 | High | The responsive shell has enough density to make overlap, drawer coverage, composer collision, and New Chat visibility a real risk at tablet and phone widths. Source CSS contains mitigation, but no live viewport evidence was available. | L2 evidence gate | **Open.** Verify before restructuring; keep the change isolated if a section-level fix is needed. |
| 6 | Medium | The top area gives similar visual weight to status, configuration, code, command menu, GitHub, glossary, theme, language, model, zoom, export, and clear. This makes the prompt workspace compete with the product chrome. | L2 | **Proposal only.** Test a section-level hierarchy change with before/after captures. |
| 7 | Medium | The Studio depends on multiple remote CDN scripts and styles for markdown, math, diagrams, document parsing, OCR, and runtime helpers. That is a reliability boundary for a product described as local-first and immediate. | L3 | **Proposal only.** Define explicit degraded states or package only the critical path; do not invent fake offline fallbacks. |
| 8 | Medium | `playground.html` is a monolithic 11,674-line document containing tokens, layout, markup, runtime state, rendering, and feature logic. It makes visual regressions and ownership boundaries difficult to see. | L3 | **Proposal only.** Split by component/build boundary only after behavior and generated-artifact parity are protected. |
| 9 | Low | The interface uses several metaphors simultaneously: “Sacred Geometry,” “Celestial Glassmorphic,” “Cyberpunk,” “Editorial,” “Quantum,” and branded theme references. This is memorable but not yet a coherent product language. | L3 | **Proposal only.** Keep selectable themes for now; establish one canonical BOB default first. |

## 4. What was improved and why

### L1 — completed in this audit

| Improvement | Why it matters | Evidence |
|---|---|---|
| Removed the mobile scale lock | Users can use browser accessibility zoom and platform text enlargement without the page refusing it. | `internal/server/playground.html:10`; `TestPlaygroundAccessibilityFloorIsExplicit` |
| Added global visible focus styling | Keyboard users can see the active control across native buttons, links, fields, and role buttons. | `internal/server/playground.html:839-855` |
| Added reduced-motion behavior | The product remains usable for people who request less motion, including dialog, artifact, and micro-interaction animations. | `internal/server/playground.html:849-855` |
| Normalized static and dynamic dialogs | Dialog surfaces now expose `role="dialog"`, `aria-modal`, `aria-hidden`, and an accessible label. | `internal/server/playground.html:3940-4483` |
| Added shared modal lifecycle | Opening stores the trigger, focuses the intended first control, traps Tab/Shift+Tab, handles Escape, and returns focus on close. | `internal/server/playground.html:7373-7433` |
| Repaired hosted onboarding surface | The onboarding popup now uses the existing dialog design system, wraps long commands, and remains bounded by the viewport. | `internal/server/playground.html:7737-7784` |
| Made core click-only controls keyboard-usable | Starter cards, model chip, token estimate, and reasoning headers can be activated with Enter or Space. | `internal/server/playground.html:7435-7444`, `5175-5206` |
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
| Primary prompt/composer | All | Accessible name, focus visibility, loading/stop/error recovery | High | **Partially source-fixed**; lifecycle tests pass, browser state walk pending |
| Chat scrolling | All | Top/bottom navigation, streaming bottom anchor, composer offset | High | **Source-supported** by existing regression tests; real scroll walk pending |
| Starter cards | Desktop/tablet/phone | Click-only activation excluded keyboard users | Medium | **Source-fixed** with role/tabindex/Enter/Space; browser keyboard walk pending |
| Model chip and token estimate | Desktop/tablet/phone | Click-only controls lacked keyboard semantics | Medium | **Source-fixed**; browser keyboard walk pending |
| Dialogs and command palette | All | Missing semantics, focus trap/return, and consistent Escape behavior | High | **Source-fixed**; browser focus-trap and return test pending |
| Hosted onboarding dialog | All | Undefined modal classes and long command content could render poorly | High | **Source-fixed**; browser resize/wrap test pending |
| Gateway/configuration panel | Tablet/phone | Dense controls and 28px actions may be difficult to operate by touch | Medium | **Open**; assess target size and panel overflow in browser |
| Artifact Preview/Code | All | Empty/error/loading/source-limit states and generated game launch | High | **Open for rendered acceptance**; source-level sandbox/recovery tests pass |
| Language selector and Hindi UI | All | Translation coverage and text expansion | Medium | **Open**; docs correctly classify coverage as English/Hindi plus partial targets |
| Theme surfaces | All built-in themes | Subdued/muted text contrast | High | **Source-fixed** with numeric contrast regression; browser visual sampling pending |
| External GitHub/release links and Copy actions | All | Default-browser handoff and clipboard success/failure feedback | Low | **Source-supported** by real anchors/`noopener` and guarded clipboard writes; click behavior not browser-verified here |
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

Completed after the source pass:

```text
make web                                      PASS
web/index.html source parity                  PASS
go test -count=1 ./...                        PASS
go vet ./...                                  PASS
git diff --check                              PASS
inline playground JavaScript syntax check    PASS
clipboard failure-safety regression          PASS
GET /playground on 127.0.0.1:19614           200 OK
GET /healthz on 127.0.0.1:19614              200 OK
```

The local HTTP checks verified that the served route contains the corrected
viewport, focus, dialog, and onboarding markers. They did not exercise layout,
keyboard focus, touch, rendering libraries, or provider generation.

No GitHub Actions workflow was added or invoked. No release was tagged or
published by this audit. No credential, API key, cookie, or signing key is
stored in this document or source change.

## 10. Files changed and deliberately untouched

Changed:

- `internal/server/playground.html` — L1 accessibility, modal, keyboard,
  reduced-motion, onboarding, contrast-token, attachment-dispatch, and
  duplicate-handler corrections.
- `internal/server/playground_test.go` — source-level accessibility, theme
  contrast, single-dispatch attachment, and clipboard failure-safety
  regression tests.
- `internal/updater/desktop_stage.go` — actionable permission-denied staging
  error classification.
- `internal/updater/desktop_stage_test.go` — read-only and permission-denied
  staging error regression tests.
- `internal/server/middleware.go` — optional request-logger nil guard for
  partial embedded apps.
- `internal/server/server_test.go` — partial-app request-logging regression.
- `internal/gemini/client.go` — nil-safe optional retry logger for buffered and
  streaming upstream retries.
- `internal/gemini/client_test.go` — retry regressions for an absent logger.
- `web/index.html` — generated static distribution synchronized by `make web`.
- `docs/engineering/FAILURE-REGISTER-100.md` — refreshed branch/main evidence
  and the attachment failure-path status.
- `docs/engineering/DESIGN-AUDIT-2026-08-31.md` — this audit.

Deliberately untouched:

- Gemini payload construction, SAPISID authentication, cookie/session routing,
  streaming deduplication, thinking extraction, Scotty upload, and model-mode
  mapping.
- OpenAI, Anthropic, and Google adapter wire formats.
- Gateway CORS/PNA policy, provider routing, API-key handling, updater
  selection/replacement behavior, Wails/Tauri history, release assets, and
  GitHub publication state.
- Navigation model, feature removal, theme catalog, and component/build
  restructuring; these are L2/L3 decisions requiring evidence or explicit
  approval.

## 11. Next decision

The next responsible step is not more visual polish. Recover the real browser
runtime, run section 8, capture the evidence, and then choose exactly one L2
experiment. If that run is clean, the product has a strong correctness base for
the next design phase. If it exposes drawer, artifact, or touch failures, fix
those as independent regression-locked changes before considering the L3
design-language or monolith split proposals.
