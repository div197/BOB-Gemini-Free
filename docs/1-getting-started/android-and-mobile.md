# Android and Mobile Status

**Status: experimental Go binding substrate; no supported native Android or iOS
application is shipped.**

This page documents the boundary of `pkg/mobile` so it is not mistaken for a
finished mobile product. The current repository contains Go source and tests,
but it does not contain an Android Gradle project, Kotlin/Compose application,
iOS Xcode project, SwiftUI application, AAR, XCFramework, mobile updater, or
mobile release artifacts.

## What currently exists

`pkg/mobile` exposes a small Go API for exploratory embedding:

- start and stop a local HTTP gateway;
- report the gateway URL and health state;
- call the existing Go Gemini client for text generation and streaming;
- estimate tokens locally; and
- call the existing three-stage refiner orchestration.

The implementation currently creates a `net.Listener` and an HTTP server. It is
therefore not an in-process, zero-socket, or 0 ms mobile transport. It uses the
same Google web upstream as the desktop/CLI gateway, so generation remains
network-, session-, account-, and provider-dependent.

The current cookie argument is written to a temporary local file for the Go
client. It is not an Android `CookieManager`, iOS `WKHTTPCookieStore`, hardware
keystore, or Keychain integration. Do not pass a shared teacher cookie, and do
not bind the experimental gateway to `0.0.0.0` or another LAN interface.

## Current verification

The Go bridge has deterministic package tests:

```bash
go test ./pkg/mobile
go test -race ./pkg/mobile
```

These tests cover the Go lifecycle and local health endpoint only. They do not
prove an Android/iOS build, mobile UI, callback-thread safety, secure cookie
storage, cancellation behavior, app lifecycle integration, or a real device.
The audit host does not treat `gomobile` as an installed release tool.

## What is required for a real mobile release

Before describing BOB as a downloadable Android or iOS app, a separate mobile
project and acceptance track is required:

1. choose and document the mobile UI/runtime architecture;
2. implement app-sandboxed session storage using the platform security APIs;
3. define cancellation, lifecycle, background, offline, and callback-thread
   contracts;
4. build and sign an Android APK/AAB and an iOS archive on the native hosts;
5. test text, streaming, errors, expired sessions, large inputs, and app
   suspension on representative devices;
6. add a mobile-specific update/distribution policy; and
7. publish only artifacts that have actually passed those gates.

`gomobile bind` may be evaluated as an implementation experiment, but a bind
command alone is not a native application or a release acceptance proof. No
student should be asked to install an AAR/XCFramework generated from this
repository until the missing application and security work is complete.

## Desktop and CLI remain separate

The supported student-facing targets in this release cycle are the Go CLI and
the Wails desktop application. Their release, signing, updater, and 30-device
acceptance rules are documented in
[`docs/engineering/RELEASE-READINESS-v0.2.0.md`](../engineering/RELEASE-READINESS-v0.2.0.md).
The experimental mobile package is deliberately excluded from that desktop
release contract.
