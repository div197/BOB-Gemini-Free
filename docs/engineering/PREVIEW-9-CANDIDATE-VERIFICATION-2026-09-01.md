# v0.2.0-preview.9 — Release Verification

**Date:** 2026-09-01 (Asia/Kolkata)
**Status:** public controlled macOS beta
**Source:** public `main` commit `4236f65b9e4972a581d140ce46b0c5126602df65`
**Release:** [GitHub v0.2.0-preview.9](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.9)

Preview 9 was rebuilt from the reviewed `main` source after the artifact-family
guard. It was signed through the owner-controlled macOS Keychain, published
manually without GitHub Actions, and verified again from a clean GitHub
download. The earlier local package built from `1410bc2` is superseded and is
not part of this release.

This receipt records package and updater evidence. It does not turn an ad-hoc
macOS package into a Developer ID/notarized release, and it does not claim
that every generated HTML artifact or Google upstream request will succeed.

## User-reported issues addressed

- Native Wails external links remain inside a small bootstrap shell while the
  loopback Studio runs in a same-window iframe. GitHub and other allow-listed
  external links are forwarded to Wails `BrowserOpenURL`, which opens the
  operating system's default browser.
- Artifact Full Screen is an in-place, reversible focus mode in the native
  shell. It does not depend on a second browser window or popup permission.
  Hosted-browser pop-outs retain their sandboxed path and expand in place when
  a browser blocks the popup.
- Generated HTML artifact diagnostics report a source location, failed
  external resource, or CSP-blocked resource when the sandbox can provide it;
  the artifact remains an opaque-origin sandbox without `allow-same-origin`.
- Ordinary Studio pages retain `X-Frame-Options: SAMEORIGIN`. Only the
  explicit `desktop_shell=1` Studio path allows the narrow Wails embedding
  origins required by the native bootstrap.

## Build and signing evidence

The inspected local release directory was:

```text
/tmp/bob-gemini-free-preview9-20260901-main-4236f65
/tmp/bob-gemini-free-preview9-20260901-main-4236f65.app
```

| Gate | Result |
|---|---|
| Source identity | PASS; clean `main` commit `4236f65b9e4972a581d140ce46b0c5126602df65` |
| Wails packaging | PASS; Wails `v2.15.0`, macOS universal |
| App bundle signature | PASS; ad-hoc signature validates on disk |
| Binary architecture | PASS; `x86_64` and `arm64` slices |
| Bundle metadata | PASS; numeric base version `0.2.0`, minimum macOS `10.13.0` |
| DMG layout | PASS; app plus conventional `/Applications` drag target |
| Detached manifest | PASS; `SHA256SUMS` and `SHA256SUMS.sig` verify with the checked-in public key |
| Asset checksums | PASS; notice, DMG, and ZIP match the signed manifest |
| Signing custody | PASS; private value streamed from the owner-controlled macOS Keychain and never displayed, exported, or committed |
| GitHub Actions | NOT USED; no repository workflow exists |

The release asset set is exactly:

```text
bob-gemini-free-macos-universal.dmg
bob-gemini-free-macos-universal.zip
RELEASE-NOTICE.txt
SHA256SUMS
SHA256SUMS.sig
```

Local SHA-256 values for the published bytes:

```text
RELEASE-NOTICE.txt                    52fbe5623355bcc5a6349ea0c94b351b98441c045549dd0f4bae0deaeaad60c9
SHA256SUMS                            1f8e77f840dd7ae99fea39af2909f98e3d6de87c48de4fda03085dbd128b5a7c
SHA256SUMS.sig                        2d0eb022a2df8723dfd757c79354bab8c2e9e95267e8fa59a3e6dee53a0f929b
bob-gemini-free-macos-universal.dmg  874e00e6b67984dd4c77aad3b720cbbbc33d4483f8d26ee68db14c8078251cdf
bob-gemini-free-macos-universal.zip  2cc75368b6b9256f99e78e1f10b5cb8a3510446cf3030b5fc7240f5dcad1db50
```

The public release has the same five file sizes as the local signed input:

```text
bob-gemini-free-macos-universal.dmg  20,567,178 bytes
bob-gemini-free-macos-universal.zip  19,017,019 bytes
RELEASE-NOTICE.txt                    1,257 bytes
SHA256SUMS                            289 bytes
SHA256SUMS.sig                        89 bytes
```

## Packaged runtime smoke

The fresh package was launched separately from the installed student
application and owned its own loopback endpoint at `127.0.0.1:8081`.

| Check | Result |
|---|---|
| `GET /healthz` | PASS; HTTP 200, stable `{"status":"ok"}`, no provider call |
| Package identity | PASS; `X-Bob-Version: v0.2.0-preview.9` |
| Ordinary `/playground` headers | PASS; `X-Frame-Options: SAMEORIGIN` |
| `/playground?desktop_shell=1` headers | PASS; no X-Frame-Options and Wails-only `frame-ancestors` CSP |
| Embedded bridge markers | PASS; native browser bridge, exact-origin message validation, and artifact diagnostics are embedded |
| Existing installed app | PRESERVED; it was not stopped, replaced, or modified |
| Current updater check | PASS; `/v1/update/check` reported channel `preview`, current/latest `v0.2.0-preview.9`, `has_update=false` |

The fresh temporary package was stopped after the smoke run. The installed
`/Applications/BOB Gemini Free.app` was left untouched.

## Public Preview 7 migration observation

The exact public `v0.1.7-preview.7` macOS app was downloaded and launched in
isolation, without installing over the student's app. The native Help menu
was opened and **Check for Updates** was selected. The app displayed:

```text
Update available
A newer signed desktop release is available: v0.2.0-preview.9
Download, verify, and install it after restarting BOB?
```

This is **VERIFIED_LIVE discovery** from the real public Preview 7 binary. The
test selected **Cancel** before the download/staging action. Therefore the
following remain unproven on an installed writable bundle:

- replacement and restart;
- preservation of config, cookies, and chat state;
- deliberate failed-candidate rollback;
- standard-user and managed-Mac permissions; and
- 30-device fleet acceptance.

The old Preview 7 web `/v1/update/check` endpoint can still report its legacy
stable metadata (`v0.1.5`). That endpoint is separate from the native Help
updater; it must not be used to conclude that the native Preview 7 app cannot
discover Preview 9.

## Browser and artifact boundary

The source and packaged browser checks covered desktop, tablet, and phone
layout constraints, external-link forwarding, native artifact focus mode,
resource diagnostics, and the existing opaque sandbox. A generated Solar
System artifact is not stored in this repository, so this release receipt
does not claim that a particular prior chat's generated source is correct.
If that artifact still fails, preserve its source as a fixture, use the
reported diagnostic to identify the failing line/resource, and repair the
artifact or generator contract rather than weakening the sandbox.

## Release gates and remaining work

| Gate | Status |
|---|---|
| Source, unit, race, vet, generated-bundle, package, signing, and public-byte checks | PASS |
| Public Preview 7 discovers Preview 9 through native Help | VERIFIED_LIVE |
| Preview 7 installed-bundle replacement/restart | OPEN; requires one approved writable pilot Mac |
| Deliberate rollback and state preservation | OPEN; must be exercised on an installed bundle |
| Clean-device and two/three-Mac pilot | OPEN |
| Apple Developer ID, hardened runtime, notarization, and stapling | OPEN; separate Apple account/service gate |
| Windows publisher-signed installer and Linux native acceptance | OPEN |
| Google anonymous/session/API-key availability, quotas, model identity | UPSTREAM-DEPENDENT |

Preview 9 is suitable for an explicitly labelled controlled macOS beta. It is
not yet a stable, notarized, unattended, or cross-platform student release.
No silent classroom-wide update is enabled: each updater-capable device must
be in a writable location and a user must approve the installation.
