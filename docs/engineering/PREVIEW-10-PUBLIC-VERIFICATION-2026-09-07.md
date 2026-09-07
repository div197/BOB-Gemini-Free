# v0.2.0-preview.10 — Public Release Verification

**Audit date:** 2026-09-07 (Asia/Kolkata)  
**Release:** [GitHub v0.2.0-preview.10](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.10)  
**Source target:** `1862d313804d4290411439d69779ba36d364c327`  
**Release mode:** manual publication; no GitHub Actions workflow is present

## Decision

Preview 10 is the current public native macOS beta. It is a real universal
Wails application with a signed project release manifest and a visible
`/Applications` drag target. It remains a controlled preview: the app is
ad-hoc signed, not Apple Developer ID signed or notarized, and the updater is
user-consented rather than a silent fleet controller.

## Release gates

| Gate | Result | Evidence | Boundary |
|---|---|---|---|
| Reviewed source identity | PASS | `origin/main` and local `main` point to merge commit `1862d31`; `scripts/verify-release-source.sh v0.2.0-preview.10` passed | The source gate does not prove Google availability or platform trust |
| Native build identity | PASS | The post-Wails guard verified `v0.2.0-preview.10`, channel `preview`, and the checked-in updater public key before packaging | The project key is not an Apple signing identity |
| Package shape | PASS | Universal macOS app, ZIP, DMG, ad-hoc bundle signature, and DMG `/Applications` alias passed local checks | Apple Gatekeeper may still require **Open Anyway** |
| Project authenticity | PASS | `SHA256SUMS` and `SHA256SUMS.sig` were created through the owner-controlled local Keychain path and verified with the checked-in public key | The private key remains local and is not in GitHub or the package |
| Public-byte reconciliation | PASS | All five public assets were downloaded again and matched the local signed inputs byte-for-byte; the public manifest/signature re-verified | GitHub reports `immutable: false`; write-once tag discipline remains operational |
| Signed discovery feed | PASS | After publication, the raw `main` feed and detached signature were downloaded and matched the checked-in bytes; the feed now selects `v0.2.0-preview.10` and validates through `2026-09-14T08:44:36Z` | The feed reduces metadata/API pressure; it does not replace the official release archive or its signed manifest |
| Candidate startup | PASS | Fresh Preview 10 launch returned `/healthz` `ok`, `X-BOB-Version: v0.2.0-preview.10`, native shell HTTP 200, and clean shutdown | One macOS host is not a cross-platform acceptance matrix |
| Existing-bundle discovery | PASS | The installed public Preview 9 on this Mac reported Preview 10, `has_update=true`, a native ZIP, and a signed manifest | Discovery alone does not replace an app |
| Existing-bundle migration | PASS on a preserved copy | A copy of the installed Preview 9 bundle accepted **Help → Check for Updates → Install Update**, replaced only the copy, restarted as Preview 10, reported healthy `/healthz`, and left no updater staging directory | One writable-host transaction is not clean-device, rollback-failure, or fleet evidence |

## Exact public asset set

The release contains exactly these five assets:

```text
bob-gemini-free-macos-universal.dmg
bob-gemini-free-macos-universal.zip
RELEASE-NOTICE.txt
SHA256SUMS
SHA256SUMS.sig
```

The downloaded public files were compared to the locally signed directory and
passed `go run ./cmd/release-verify` against the checked-in public key.

The checked-in `updates/desktop-feed.json` and detached signature were also
re-downloaded from the public raw `main` path after PR #128 merged. Both files
matched the local bytes, and the feed selected Preview 10 as the highest
published preview. This confirms public discovery-feed propagation, not a
silent installation.

## Installed-lineage consequence

- Current-key Preview 9 and current `v0.2.0` previews can discover a newer
  signed preview when the package exists and the app is writable.
- Public Preview 7 can discover a newer same-key preview, but its compiled
  updater predates the current stable-first policy.
- Preview 1–6 and other builds with a missing or obsolete trust key still
  require one manual installation of a current signed package.
- The stable CLI `v0.1.5` is not a native Wails app and cannot convert itself
  into this desktop package.

The updater does not silently download, replace, or restart an app. Students
must still approve the update, and each device needs a writable application
location and a recovery path. No cookie, Google API key, session, prompt, or
private release key was included in the package.

## Remaining gates before broad rollout

1. Pilot the real installed Preview 7 and any legacy-key cohort separately;
   use manual migration where the trust lineage requires it.
2. Record clean-device install, Gatekeeper approval, state preservation,
   interruption recovery, deliberate rollback, and uninstall evidence.
3. Repeat the migration on two or three ordinary student Macs before any
   20–30-device wave.
4. Keep the release labelled preview until Apple/Windows platform trust,
   provider behavior, and the pilot acceptance record are complete.
