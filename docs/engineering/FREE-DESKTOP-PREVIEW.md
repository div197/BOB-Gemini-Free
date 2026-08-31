# Free Native Desktop Beta

**Status:** branded public macOS preview `v0.2.0-preview.3` published as the
current controlled candidate; `v0.2.0-preview.1` remains the immutable
migration bridge and `v0.1.7-preview.7` remains the existing fleet baseline.

This is the no-Apple-membership path for evaluating the BOB Gemini Free
desktop product. It creates a real branded native application and packages it
for controlled testing, but it intentionally does not claim Apple Developer ID
trust, notarization, Windows publisher trust, or a production student release.
The desktop toolkit is an internal build detail; the application, bundle, and
student-facing release names are BOB Gemini Free.

## What the preview contains

- branded native window;
- embedded Go gateway;
- BOB Builder studio;
- loopback-only gateway boundary;
- safe occupied-port fallback;
- no Go, Node, Rust, SQLite, or separate server required at runtime;
- no embedded Google cookie, API key, private release key, or teacher
  credential; Preview 7 contains only the public updater trust key;
- native macOS maximize control, default-browser external links, and expanded
  English/Hindi UI language coverage;
- a native Help menu with an explicit “Check for Updates” action.

The update action checks fixed official GitHub channels only when the user
selects it. A newly built preview first checks for a newer stable release so it
can make an explicit Preview → Stable migration; if none exists, it checks the
preview channel. The already-published Preview 7 binary predates that
stable-first behavior and can reach stable through the updater only after a
same-key bridge preview, or through a manual stable install. It never silently
downloads or replaces an application. A newer signed candidate requires
consent, manifest/package verification, safe staging, and rollback protection.
Historical `v0.1.7-preview.3` and other builds without the current embedded
desktop trust key still require a manual migration. This project-level signature does not create Apple
Developer ID or notarization trust.

The standalone CLI installer is a separate path. It verifies the signed
release manifest before installing a CLI binary; it is not a native desktop
installer. Download the installer as a local file, inspect it, and run it
locally. Do not pipe a mutable repository branch directly into a shell, and do
not interpret the CLI's project signature as Apple or Windows publisher trust.

## Public preview releases

The current [v0.2.0-preview.3 controlled macOS preview](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.3)
contains the universal macOS package shape and current-key signed manifest.
Existing Preview 7 users can discover it directly through their preview-only
lookup. The immutable [v0.2.0-preview.1 migration bridge](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.1)
remains available if a device has already selected that intermediate step.

The manually published [v0.1.7-preview.7 native desktop beta](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.1.7-preview.7)
contains:

- `bob-gemini-free-macos-universal.dmg`;
- `bob-gemini-free-macos-universal.zip`;
- `RELEASE-NOTICE.txt`, `SHA256SUMS`, and `SHA256SUMS.sig`.

Historical `v0.1.7-preview.3` remains available separately with the Windows x64
preview asset. Preview 7 is intentionally a macOS-first signed-update pilot;
Preview 6 installations require a one-time manual migration because their
older project signing key cannot verify Preview 7. Windows and Linux require
their own native build and acceptance evidence.

The release is suitable for informed evaluation and a controlled pilot. It is
not a Developer ID/notarized Mac release, a Windows publisher-signed release,
a Linux release, or proof of provider availability or unlimited use.

The current Preview 3 package carries the branded package refresh, signed
preview updater, native window/browser refinements, language coverage, and the
Web Studio generation lifecycle correction: `STOP` returns to
`SEND` on completion, cancellation, timeout, upstream failure, and truncated
stream, while incomplete streams are no longer silently treated as complete.
It also bounds the GitHub preview listing request and explains the
read-only-App-Translocation case instead of exposing a raw staging error.

## Build the free macOS preview

From macOS:

```bash
make desktop-preview-mac
```

In the current source this command defaults to the published
`v0.2.0-preview.3` package identity. The signed `v0.2.0-preview.1` migration
bridge is immutable and already public. Set `BOB_RELEASE_VERSION` explicitly
for every publication;
the already-published `v0.1.7-preview.7` package remains the historical public
preview and is not rebuilt in place.

The updater-capable preview packager requires the non-secret public trust key
in `BOB_GEMINI_FREE_UPDATE_PUBLIC_KEY`. The matching private key is never
needed to build the app and must be used only by the local signing step. A
release package without the embedded public key is rejected by the packager;
this prevents a preview from claiming safe updates while lacking its trust
anchor.

The default output is a new directory at:

```text
/tmp/bob-gemini-free-preview/
```

The release-asset directory contains only the files intended for signing and
publication:

```text
bob-gemini-free-macos-universal.zip
bob-gemini-free-macos-universal.dmg
RELEASE-NOTICE.txt
SHA256SUMS
```

For local inspection, the signed app bundle is written beside the directory as
`<output-directory>.app`; it is deliberately outside the release-asset
directory because the signed manifest covers regular publication files only.

The no-Actions release operator signs the exact inspected artifact directory
with `scripts/sign-release-assets.sh`; the script regenerates any unsigned
`SHA256SUMS` and refuses to overwrite an existing detached signature.

To choose another output directory or build one architecture:

```bash
BOB_WAILS_PLATFORM=darwin/arm64 \
  scripts/package-wails-preview.sh /tmp/bob-gemini-free-arm64-preview
```

The bundle is ad-hoc signed so local macOS validation can inspect a coherent
bundle. It is not signed with a Developer ID certificate and is not notarized.
The DMG presents the app beside a conventional `/Applications` shortcut for
drag-to-install use. The release notice must remain beside any preview artifact
shared with a student or tester.

## Windows and Linux

The native build must be run and tested on a native host:

```text
Windows: make desktop-windows
Linux:   make desktop-linux
```

For a free controlled Windows preview from a host that can cross-compile the
native executable:

```bash
make desktop-preview-windows
```

This produces only an unsigned raw `.exe`; it is not an installer and cannot
bootstrap WebView2. The default NSIS target now fails closed when `makensis`
does not actually create the installer instead of reporting a warning as a
successful release build.

For Linux, use the native-host preview packager:

```bash
make desktop-preview-linux
```

It creates a documented `.tar.gz` package after checking GTK3 and WebKit2GTK
4.1 development libraries. It is not an AppImage.

The native Windows target can produce an NSIS installer when `makensis` is
installed, and may need the Microsoft WebView2 runtime. The free
cross-compiled preview produces only a raw `.exe`. Linux requires the
GTK3/WebKit2GTK libraries documented by the platform; the current free path is
therefore a technical preview with explicit runtime requirements, not a
universal zero-setup promise.

No Apple membership is involved in either build. Windows publisher signing and
Linux package verification are separate release decisions.

## Student pilot boundary

The free preview can be used with a small, informed pilot. It must not be
described as a production-ready download because:

1. macOS may show a first-launch developer-verification warning;
2. each student still needs an explicit, authorized Google session for
   authenticated features;
3. the native app does not yet provide a complete first-run Google sign-in
   wizard;
4. Windows and Linux artifacts still require native-host acceptance;
5. project checksums prove file integrity only when the release manifest is
   signed and its public key is trusted.

Do not disable Gatekeeper globally, embed a shared cookie, or call an ad-hoc
preview “notarized”, “Apple trusted”, “unlimited”, or “zero setup on every
platform”.

To verify the local package after building:

```bash
(cd /tmp/bob-gemini-free-preview && shasum -a 256 -c SHA256SUMS)
unzip -t /tmp/bob-gemini-free-preview/bob-gemini-free-macos-universal.zip
bash scripts/verify-macos-dmg-layout.sh \
  /tmp/bob-gemini-free-preview/bob-gemini-free-macos-universal.dmg
```

## Next trust upgrade

When the project is ready for a professional Mac release, add Apple Developer
ID signing, hardened runtime, notarization, stapling, and clean-device testing.
That is an external distribution gate; it is not required for the native
architecture itself.
