# Browser UI Validation — 2026-08-31

## Scope

This is a local, deterministic browser check of the hosted Studio bundle. It
does not prove Google-provider availability, native Wails rendering, CDN
availability, or a clean-device release install.

- Page: `http://127.0.0.1:19614/playground`
- Gateway: local source build on `127.0.0.1:19614`
- Upstream fixture: deterministic local SSE server on `127.0.0.1:19615`
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

This validation covers the local hosted bundle after the working-tree fix. The
fix is now merged into public `main` at `5530edb`; the published
`v0.2.0-preview.4` assets are immutable and were built before this change. It
must be included in the next packaged preview, whose source default is
`v0.2.0-preview.5`, after that commit is rebuilt and signed. Native macOS,
Windows, Linux, CDN/offline, and
clean-device acceptance remain open gates in
[`FAILURE-REGISTER-100.md`](FAILURE-REGISTER-100.md).
