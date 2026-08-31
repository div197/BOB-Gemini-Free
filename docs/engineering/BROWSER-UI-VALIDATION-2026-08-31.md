# Browser UI Validation — 2026-08-31

> Historical fixture record. This browser run exercised the source before the
> Preview 5 package was published. The artifact-preview, responsive-header,
> and multiline-SSE behavior recorded below is represented in the immutable
> Preview 5 package. Later public-main security and credential-input follow-ups
> are not retroactively claimed in those package bytes.

## Scope

This is a local, deterministic browser check of the hosted Studio bundle. It
does not prove Google-provider availability, native Wails rendering, CDN
availability, or a clean-device release install.

- Historical page fixture: `http://127.0.0.1:19614/playground`
- Historical gateway fixture: local source build on `127.0.0.1:19614`
- Historical upstream fixture: deterministic local SSE server on `127.0.0.1:19615`
- Input: split HTML code fence and HTML source delivered across multiple SSE
  chunks
- Credentials: none; no Google account, cookie, GitHub credential, or provider
  key was used
- Bundle preparation: `make web`

## Reproduced defect and fix

Opening an interactive artifact assigned `iframe.srcdoc` while the artifact
modal was still hidden. In the tested WebView, the iframe document then
calculated a 0×0 layout and the preview appeared blank even though the editor
contained source. The preview path now opens the managed modal and selects the
preview pane before assigning `srcdoc`. The focused regression test is
`TestArtifactPreviewOpensBeforeIframeHydration`.

## Evidence

The patched bundle was served and exercised in the in-app browser:

- The streamed artifact registered as `Browser Smoke Artifact`.
- Launching it left the runtime-error alert hidden.
- The iframe contained `Artifact hydrated`; its body measured non-zero layout
  (`1242×79` in the 1440px test), and the `Run` action changed the heading to
  `Artifact ran`.
- The source editor contained 314 characters, including the title and script;
  its status read `HTML5 • 1 lines • 314 chars`.
- At a constrained 1440×500 viewport, moving to the top exposed `Jump to
  bottom`, which moved the message list to its bottom; moving back exposed
  `Jump to top`, which returned it to the top.

## Responsive acceptance snapshot

| Viewport | Document/body width | Header result | Key controls |
|---|---:|---|---|
| 1440×900 | 1440 / 1440 | Full controls; language and theme selectors visible | Artifact and chat surface available |
| 768×1024 | 768 / 768 | Tablet labels compacted; theme moved out of the header | New, Chat prompt, and UI language visible |
| 390×844 | 390 / 390 | Phone controls remain bounded; theme hidden | New, Chat prompt, and UI language visible |

The tablet regression was a measured 836px document/body width before the
responsive fix. After the fix, the 768px document, body, and header widths
match the viewport. The collapsed configuration/integration drawers may retain
off-screen descendants by design; they are inside the clipped workspace and do
not create page-level horizontal scroll.

## Release boundary

This validation covers the local hosted bundle after the artifact-preview,
responsive-header, and multiline-SSE framing fixes. The historical fixture
ran before the immutable `v0.2.0-preview.5` package was built, but those three
behaviors were included in that package after clean-source packaging and
signature verification. At that intermediate checkpoint, public `main` had
advanced to `ade691d` through PR #77 (browser security evidence) and PR #78
(credential-input probe hygiene). It then advanced through PRs #80–#83 to
`9f11eef`, documentation/test reconciliation PR #84 moved public `main` to
`0cc81b2`, and PR #86 added the browser gateway-key transport guard at
`49e0d3b`. These post-publication source changes are not claimed for Preview 5;
they are included in the published Preview 6 package from `f9b3410`. Native
macOS, Windows, Linux, CDN/offline, and clean-device acceptance remain open gates in
[`FAILURE-REGISTER-100.md`](FAILURE-REGISTER-100.md).

## Current-source responsive follow-up — 2026-08-31

This addendum records a later source-candidate run. It does not rewrite the
historical fixture above or upgrade the immutable public Preview 6 package.

- Candidate: current source build identifying as `v0.2.0-preview.7`, served at
  `http://127.0.0.1:18081/playground` from the isolated local candidate.
- Provider boundary: no generation request was made and no provider key or
  cookie was entered in the browser; this is shell/settings evidence only.
- Before the fix, a 390×844 browser viewport measured compact header controls
  between 18px and 33px wide, and the drawer measured 331.5px because
  `min-width:85vw` overrode its intended 320px maximum.
- After the fix, 1440×900, 1024×768, and 390×844 all reported matching
  document/body widths with no page-level horizontal overflow. At 390×844,
  compact controls measured at least 24px wide and 44px high; the segmented
  control row measured 48px high; and the configuration drawer measured
  320px wide.
- The mobile configuration drawer still exposed `aria-hidden`/`inert` state,
  moved focus to its close control, and returned focus to its trigger after
  closing. The Gateway dialog remained a scrollable `role="dialog"` with
  `aria-modal="true"`; its endpoint, BOB Gateway Access Key, and Google
  Gemini Developer API key fields remained distinct and the two credential
  fields were empty.

The 24px horizontal value is the repository's absolute compact-control
containment floor, not a claim that every secondary header action is a full
44×44 touch target. The header hierarchy remains an isolated L2 design
experiment; no navigation or feature removal was smuggled into this fix.
Artifact/provider/native-WebView, 200% browser zoom, clean-device, and fleet
acceptance remain separate gates.

## Gateway settings touch-target follow-up — 2026-08-31

The current source bundle was opened in the in-app browser with no provider
credential, cookie, or Google request. The Gateway dialog was inspected at
390×844, 1024×768, and 1440×900.

- Before the scoped fix, Gateway-modal controls measured 28–38px high.
- After the fix, every Gateway-modal button and every non-checkbox text or
  password input measured at least 44px high at all three widths.
- At 390×844, the dialog was 350px wide, its internal body remained
  `overflow-y:auto`, and the document/body widths both remained 390px.
- At 1024×768 and 1440×900, the dialog remained within the viewport and the
  page widths matched their viewports; no controls below 44px were observed.
- Closing the dialog returned focus to `btn-gateway-status` and the modal's
  two credential fields remained separately labelled and masked.

The focused source regression is
`TestGatewayModalControlsMeetTouchTargetContract`. This is local rendered
browser evidence; it is not proof of native-WebView, provider, clean-device,
or fleet behavior.
