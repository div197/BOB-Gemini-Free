# Free Native Desktop Beta

**Status:** corrected public experimental preview published as `v0.1.7-preview.2`.

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
- no embedded Google cookie, API key, release key, or teacher credential;
- a native Help menu with an explicit “Check for Updates” action.

The update action checks the official GitHub release metadata only when the
user selects it. It never silently downloads or replaces an application. It
opens the official stable release page only when a matching native package is
available. The preview is a prerelease and is intentionally downloaded
manually from its dedicated release page; the native updater does not silently
promote or install prereleases.

## Public preview release

The manually published [v0.1.7-preview.2 native desktop beta](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.1.7-preview.2)
contains:

- `bob-gemini-free-wails-macos-universal.dmg` (legacy Preview 2 name);
- `bob-gemini-free-wails-macos-universal.zip` (legacy Preview 2 name);
- `bob-gemini-free-wails-windows-amd64.exe` (legacy Preview 2 name);
- `RELEASE-NOTICE.txt` and `SHA256SUMS`.

The release is suitable for informed evaluation and a controlled pilot. It is
not a Developer ID/notarized Mac release, a Windows publisher-signed release,
a Linux release, or proof of provider availability or unlimited use.

Preview 2 also corrects the Web Studio generation lifecycle: `STOP` returns to
`SEND` on completion, cancellation, timeout, upstream failure, and truncated
stream, while incomplete streams are no longer silently treated as complete.

## Build the free macOS preview

From macOS:

```bash
make desktop-preview-mac
```

The default output is a new directory at:

```text
/tmp/bob-gemini-free-preview/
```

It contains:

```text
BOB Gemini Free.app
bob-gemini-free-macos-universal.zip
bob-gemini-free-macos-universal.dmg
RELEASE-NOTICE.txt
SHA256SUMS
```

To choose another output directory or build one architecture:

```bash
BOB_WAILS_PLATFORM=darwin/arm64 \
  scripts/package-wails-preview.sh /tmp/bob-gemini-free-arm64-preview
```

The bundle is ad-hoc signed so local macOS validation can inspect a coherent
bundle. It is not signed with a Developer ID certificate and is not notarized.
The release notice must remain beside any preview artifact shared with a
student or tester.

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
```

## Next trust upgrade

When the project is ready for a professional Mac release, add Apple Developer
ID signing, hardened runtime, notarization, stapling, and clean-device testing.
That is an external distribution gate; it is not required for the native
architecture itself.
