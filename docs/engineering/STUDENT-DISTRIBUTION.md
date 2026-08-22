# Student Desktop Distribution Contract

**Status:** public macOS/Windows preview is published; production trust,
Linux acceptance, and broad student rollout remain pending.

This document separates the working native application shown in the local
screenshots from a product release that students can download and trust.

## What is already proven

- The native app embeds the Go gateway and the BOB Builder studio.
- A packaged macOS ARM64 app opened on this device, reached the actual
  loopback endpoint, rendered the studio and artifact canvas, and released its
  gateway cleanly on quit.
- The app can discover a user's existing local config/cookie files, but it
  forcibly remains loopback-only and does not accept server API keys or remote
  origins in desktop mode.
- The public [`v0.1.7-preview.2` release](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.1.7-preview.2)
  contains a macOS universal preview and a Windows x64 preview. It is a
  controlled beta, not a student-ready trusted installer set.

## Student artifact matrix

| Platform | Student artifact | Native build boundary | Runtime truth | Current status |
|---|---|---|---|---|
| macOS Apple Silicon / Intel | Free preview: ad-hoc `.app` inside `.dmg` or `.zip`; professional release: Developer ID package | Build on macOS; use `make desktop-preview-mac` for the free path or `make desktop-mac` for the raw bundle | No separate Go/Node/Rust/SQLite service | Public preview asset published; no Apple trust chain, clean-device acceptance, or production student release |
| Windows 10/11 x64 | Free preview: raw unsigned `.exe` with WebView2 already installed; professional release: publisher-signed NSIS installer | Build and test on Windows; use `make desktop-windows` | The native shell uses Edge WebView2; an NSIS installer can bootstrap the runtime | Public raw preview published; no Windows device or NSIS installer acceptance in this audit |
| Linux x64 | Free preview: documented native package/tarball; professional release: verified AppImage or distro package | Build and test on Linux; use `make desktop-linux` | GTK3 and WebKit2GTK runtime libraries remain required | Native Linux host build and distro acceptance remain pending |
| Any platform | CLI binary | `make dist` / signed local release matrix | Terminal + browser; this is not the native desktop app | Existing CLI release path; separate from the student GUI product |

The native shell's Windows packaging supports NSIS and WebView2 handling, while
its Linux build explicitly requires GTK3/WebKit2GTK runtime libraries. Linux
therefore cannot honestly be marketed as “no setup on every distribution” for
the current product.

## First-run and authentication contract

1. Students download only an official, signed artifact and verify its publisher
   and checksum.
2. The app starts a gateway on loopback and displays the actual endpoint; a
   busy port must not expose the service on the network.
3. Anonymous/guest upstream behavior is provider-dependent. A successful local
   UI launch does not guarantee a Google response, model identity, quota, or
   unlimited use.
4. Authenticated, Pro, vision, and image-generation paths require each
   student's own authorized Google session. A teacher/developer cookie must
   never be embedded in the app, shared in a class archive, or committed to
   Git.
5. The existing `--login` flow belongs to the CLI. The native app currently
   honors a user's existing local cookie/config but does not yet provide a
   complete in-app first-run sign-in wizard. Broad student rollout should wait
   for that UX or provide an explicit, tested per-user setup path.

## Release contract without GitHub Actions

The project can keep its no-Actions policy. The maintainer release operator
builds on the native macOS, Windows, and Linux hosts, runs the acceptance
matrix below, then uploads the exact artifacts manually to a GitHub Release.

The `v0.1.7-preview.1` release is an explicit exception to the production
artifact contract below: it contains only macOS and Windows preview packages,
an unsigned checksum manifest, and a warning notice. It must not be presented
as a signed production release.

Every release must include:

```text
bob-gemini-free-macos-*.zip or .dmg
bob-gemini-free-windows-amd64.exe
bob-gemini-free-linux-amd64.AppImage or a documented package
SHA256SUMS
SHA256SUMS.sig
release notes with supported OS versions and known provider limits
```

The updater's Ed25519 manifest is for the CLI update path. Native desktop
artifacts still need their platform trust chain:

- macOS: Developer ID application signing, hardened runtime/timestamp, Apple
  notarization, stapled ticket, and clean-device Gatekeeper test;
- Windows: Authenticode/MSIX/installer signing from a publicly trusted
  publisher, followed by SmartScreen and clean-device install testing;
- Linux: checksums/signature plus package metadata and explicit GTK/WebKit
  dependency declarations.

Do not publish an unsigned desktop artifact and call it production-ready. An
ad-hoc signature is useful for local testing and an explicitly labelled preview
only.

The native Help menu now offers a user-initiated metadata check. It does not
perform a silent replacement. Until signed production packages and manifests
are published, students must use the preview release page or the documented
pilot path.

## Automatic update plan and 30-device rollout gate

The public Preview 2 does **not** update itself. A new preview therefore
requires quitting BOB, downloading the replacement from the official release,
and installing/replacing the old app. Deleting the app is not normally
required, but the exact replacement step is still manual and every machine
must be handled separately.

The source now contains the tested substrate for a user-consented updater that
downloads a platform-matching package, verifies a release manifest with the
embedded Ed25519 public key, asks the user to restart, replaces the app via a
platform-specific helper, and retains rollback evidence. It is not enabled by
Preview 2 because that build has no embedded desktop key and the public
preview has no signed manifest. It must use a stable/beta channel policy and
never treat an unsigned preview archive as trusted. Apple Developer
ID/notarization and Windows publisher signing remain separate trust
requirements; an Ed25519 manifest authenticates release bytes, not the
platform publisher identity. See
[`DESKTOP-UPDATE-OPERATIONS.md`](DESKTOP-UPDATE-OPERATIONS.md) for the human
rollout risks and 30-device gate.

Do not install the current preview on all 30 student Macs yet. Use this gate:

1. one clean Mac outside the developer's environment;
2. two or three pilot Macs with the exact student account/session path;
3. a full 30-device rollout only after the signed/notarized package, updater,
   rollback path, per-user authentication instructions, and a manual fallback
   have each passed acceptance testing.

Until that gate is closed, a future bug still requires a manual replacement on
each device; the current app cannot silently repair itself.

## Acceptance gate per platform

- install/open on a clean device or VM;
- first-run window loads without a terminal or developer toolchain;
- `/healthz` is local-only and returns stable JSON;
- anonymous request behavior is recorded as upstream-dependent;
- per-user authentication, if enabled, never leaves the device;
- occupied port selects/reports a safe endpoint;
- no remote origin can invoke the privileged local gateway by default;
- app close releases the gateway it owns;
- uninstall removes only app-owned files and leaves user data explicit;
- checksum/signature and platform publisher verification succeed;
- release notes state exactly which platforms were tested.

## Immediate recommendation

For a first student pilot, ship macOS and Windows native installers only after
one clean-device test on each. Keep Linux as a documented technical preview
until the WebKit2GTK ABI/package matrix is tested. A controlled pilot can use
the published macOS/Windows preview with the release notice; it must remain
labelled beta and must not be presented as a finished cross-platform release.

## Same-day public path

The latest stable GitHub Release contains standalone CLI binaries, so a
student can start today without installing Go:

1. macOS/Linux: run the published `install.sh` command from the repository
   README; Windows: run the published `install.ps1` command.
2. Start the installed gateway with the exact path printed by the installer
   (`bob-gemini-free --port 9610` on macOS/Linux; the printed `.exe` path on
   Windows), then open its browser studio at the displayed loopback URL.
3. Complete authentication with that student's own authorized Google session
   if the selected capability requires it.

This path is a CLI plus browser experience. It is not a native desktop download,
and the hosted Cloudflare Studio alone does not silently access a student's
local machine. For the native beta, use the exact files listed on the
[`v0.1.7-preview.2` release page](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.1.7-preview.2):
macOS universal `.dmg`/`.zip` or Windows x64 `.exe`. Linux is not included.
