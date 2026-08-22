# Student Desktop Distribution Contract

**Status:** native macOS smoke-tested; Windows/Linux desktop release artifacts
and public signed distribution remain pending.

This document separates the working Wails application shown in the local
screenshots from a product release that students can download and trust.

## What is already proven

- The Wails app embeds the Go gateway and the BOB Builder studio.
- A packaged macOS ARM64 app opened on this device, reached the actual
  loopback endpoint, rendered the studio and artifact canvas, and released its
  gateway cleanly on quit.
- The app can discover a user's existing local config/cookie files, but it
  forcibly remains loopback-only and does not accept server API keys or remote
  origins in desktop mode.
- The current public GitHub release is a CLI release. It is not a student-ready
  native desktop installer set.

## Student artifact matrix

| Platform | Student artifact | Native build boundary | Runtime truth | Current status |
|---|---|---|---|---|
| macOS Apple Silicon / Intel | Signed `.app` inside `.dmg` or `.zip` | Build on macOS; use `make desktop-mac` or `BOB_WAILS_PLATFORM=darwin/universal scripts/build-wails-local.sh ...` | No separate Go/Node/Rust/SQLite service | ARM64 ad-hoc app tested; Developer ID, notarization, universal clean-device test pending |
| Windows 10/11 x64 | Signed NSIS `.exe` installer | Build and test on Windows; use `make desktop-windows` | Wails uses Edge WebView2; installer can bootstrap the runtime | Templates and command exist; no Windows artifact has been tested in this audit |
| Linux x64 | AppImage or distro package | Build and test on Linux; use `make desktop-linux` for the Wails binary | GTK3 and WebKit2GTK runtime libraries remain required | No Linux GUI artifact has been built or distro-tested in this audit |
| Any platform | CLI binary | `make dist` / signed local release matrix | Terminal + browser; this is not the native desktop app | Existing CLI release path; separate from the student GUI product |

Wails documents Windows NSIS packaging and WebView2 handling, while its Linux
documentation explicitly requires GTK3/WebKit2GTK runtime libraries. Linux
therefore cannot honestly be marketed as “no setup on every distribution” for
the current Wails v2 product.

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
5. The existing `--login` flow belongs to the CLI. The Wails app currently
   honors a user's existing local cookie/config but does not yet provide a
   complete in-app first-run sign-in wizard. Broad student rollout should wait
   for that UX or provide an explicit, tested per-user setup path.

## Release contract without GitHub Actions

The project can keep its no-Actions policy. The maintainer release operator
builds on the native macOS, Windows, and Linux hosts, runs the acceptance
matrix below, then uploads the exact artifacts manually to a GitHub Release.

Every release must include:

```text
bob-gemini-free-wails-macos-*.zip or .dmg
bob-gemini-free-wails-windows-amd64.exe
bob-gemini-free-wails-linux-amd64.AppImage or a documented package
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
ad-hoc signature is useful for local testing only.

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
until the WebKit2GTK ABI/package matrix is tested. If a class needs immediate
access before those installers exist, use the already-tested local macOS app or
the CLI/browser path, but label it as a pilot—not a finished cross-platform
release.
