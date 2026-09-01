# v0.2.0-preview.9 — Candidate Verification

**Date:** 2026-09-01 (Asia/Kolkata)
**Status:** superseded local package; do not publish
**Source candidate:** public `main` merge commit `1410bc2` (superseded by `6a5606d`)

The earlier `3518c24` commit only advanced the next candidate number. The
package receipt below is historical: it was built from `1410bc2` before the
artifact-family updater guard landed in `6a5606d`. The bytes are therefore not
eligible for publication or updater discovery. Rebuild Preview 9 from the
current reviewed `main` before using this receipt as a release checklist.

This receipt covers the focused reliability work after Preview 8. It does not
turn an ad-hoc macOS package into a Developer ID/notarized release, and it does
not claim that every generated HTML application is correct.

## User-reported issues addressed

- Native Wails external links now remain inside a small Wails bootstrap shell,
  while the loopback Studio runs in a same-window iframe. GitHub and other
  allow-listed external links are forwarded to Wails `BrowserOpenURL`, which
  opens the operating system's default browser.
- Artifact Full Screen is now an in-place, reversible focus mode in the native
  shell. It does not depend on a second browser window or a popup permission.
  Ordinary hosted-browser pop-outs retain the sandboxed popup path and expand
  in place when the browser blocks the popup.
- Generated HTML artifact diagnostics now report a source location, failed
  external resource, or CSP-blocked resource when the sandbox can provide it.
  The artifact remains an opaque-origin sandbox without
  `allow-same-origin`.
- Ordinary Studio pages retain `X-Frame-Options: SAMEORIGIN`. Only the
  explicit `desktop_shell=1` Studio path allows the narrow Wails embedding
  origins required by the native bootstrap.

## Local package gates

The superseded local candidate was staged at:

```text
/tmp/bob-gemini-free-preview9-20260901-main-1410bc2
/tmp/bob-gemini-free-preview9-20260901-main-1410bc2.app
```

| Gate | Result |
|---|---|
| Wails packaging | PASS; Wails `v2.15.0`, macOS universal |
| App bundle signature | PASS; ad-hoc signature is valid on disk and satisfies its designated requirement |
| Binary architecture | PASS; `x86_64` and `arm64` |
| Bundle metadata | PASS; numeric base version `0.2.0`, minimum macOS `10.13.0` |
| DMG layout | PASS; app plus conventional `/Applications` drag target |
| Detached manifest | PASS; `SHA256SUMS` and `SHA256SUMS.sig` verify against the checked-in public key |
| Asset checksums | PASS; notice, DMG, and ZIP match the signed manifest |
| GitHub Actions | NOT USED; no repository workflow exists |

The private release-signing value was read only through the owner-controlled
macOS Keychain signing path. It was not displayed, exported, copied, or put in
the package or repository.

## Packaged runtime smoke

The candidate was launched separately from the installed student application
and served its own loopback endpoint at `127.0.0.1:51024`.

| Check | Result |
|---|---|
| `GET /healthz` | PASS; HTTP 200, stable `{"status":"ok"}` JSON, no provider call |
| Candidate identity | PASS; `X-Bob-Version: v0.2.0-preview.9` |
| Ordinary `/playground` headers | PASS; `X-Frame-Options: SAMEORIGIN` |
| `/playground?desktop_shell=1` headers | PASS; no X-Frame-Options and Wails-only `frame-ancestors` CSP |
| Packaged bridge markers | PASS; native browser bridge, exact-origin message validation, and artifact diagnostics are embedded |
| Existing installed app | PRESERVED; it was not stopped, replaced, or modified |

The final merged candidate asset digests were recorded before any possible
publication:

```text
RELEASE-NOTICE.txt                    52fbe5623355bcc5a6349ea0c94b351b98441c045549dd0f4bae0deaeaad60c9
SHA256SUMS                            db8a0387452ea1962b3ea16bdd2cc90a2922326b4607af74447628adfdd3c9af
SHA256SUMS.sig                        f91fb898b3a573f4984b961e6686d327294ecb445155ae3fb5bb8f74b50f122a
bob-gemini-free-macos-universal.dmg  a7c02a22de210e0f85d3138c40a1ef3f16df7dc4b425647aa80d0ccda08e7d7f
bob-gemini-free-macos-universal.zip  f43ea08414f21499af187557b2facfa8cc14d0f1e5aaea32985553f3bc3dc278
```

These are superseded local-package hashes, not release inputs and not
public-release hashes. Do not upload them. Recompute all hashes from a clean
release directory after rebuilding from current `main`, then reconcile the
public bytes against GitHub before any student-facing announcement.

## Headless browser smoke

The packaged endpoint was exercised with a local Chromium binary at these
viewports:

| Viewport | Horizontal overflow | Result |
|---|---:|---|
| 1440 × 900 | none | PASS |
| 1024 × 900 | none | PASS |
| 390 × 844 | none | PASS |

Additional deterministic checks passed:

- an inline artifact exception was surfaced as `fixture boom` with a source
  location;
- a failed external script was surfaced with its exact resource URL;
- a representative Three.js plus OrbitControls fixture loaded in the same
  opaque sandbox and reached its ready state with no failed requests;
- the embedded artifact focus branch toggled `false → true → false → true`;
- the hosted top-level external-link fallback invoked the allow-listed URL.

These checks prove the BOB host boundary and failure reporting. They do not
prove the correctness of the particular Solar System HTML generated in the
user's earlier chat, because that generated source is not stored in this
repository and was not available as a fixture.

## Remaining acceptance gates

The following remain deliberately open before calling this a stable or
30-device release:

1. Rebuild a fresh Preview 9 package from current `main`, then open the exact
   generated Solar System artifact and record the
   new diagnostic if it still fails. If the diagnostic points into generated
   code, repair that source or add it as a regression fixture; do not weaken
   the sandbox to hide the defect.
2. On one writable Mac, use **Help → Check for Updates** from the installed
   Preview 7/Preview 8 lineage, approve the update, and verify restart,
   version, health, and rollback behavior. Local updater tests do not replace
   this installed-device observation.
3. After publication, re-download the public Preview 9 assets from a clean
   directory and compare all bytes with the signed local inputs.
4. Keep the Apple Gatekeeper warning boundary explicit. The candidate is
   ad-hoc signed and non-notarized; platform publisher trust is not established.
5. Treat Google availability, cookies, API keys, quotas, model identity, and
   generated-artifact CDN behavior as upstream- or source-dependent.

The superseded package is not suitable for distribution. A rebuilt Preview 9
is suitable only as an explicitly labelled controlled macOS beta until those
gates are closed. No silent classroom-wide update is enabled.

The installed-version transition rules, including the special case of local
or historical `v0.1.9` builds, are recorded in
[`RELEASE-TRANSITION-AUDIT-2026-09-01.md`](RELEASE-TRANSITION-AUDIT-2026-09-01.md).
