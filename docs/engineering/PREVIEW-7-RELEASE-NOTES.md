# BOB Gemini Free v0.1.7-preview.7

BOB Gemini Free Preview 7 is a manually published, open-source macOS
universal beta package for controlled evaluation.

## Important migration note

Preview 7 uses the project update-signing key documented by the repository's
public-key file. The original Preview 6 private signing key was not
recoverable, so existing Preview 6 installations cannot automatically accept
Preview 7. Each Preview 6 device must install this release once manually,
then later releases signed with the Preview 7 key can be installed through
**Help -> Check for Updates** after the user gives consent.

The updater is explicit and user-consented. It does not silently check on
startup or push an update to other computers.

## Included macOS assets

- `bob-gemini-free-macos-universal.dmg`
- `bob-gemini-free-macos-universal.zip`
- `RELEASE-NOTICE.txt`
- `SHA256SUMS`
- `SHA256SUMS.sig`

The project manifest authenticates the exact release bytes. The application
is ad-hoc signed for bundle integrity but is not Apple Developer ID signed or
notarized; macOS may require first-launch approval in Privacy & Security.

## Preview 7 changes

- Explicit Google policy/rejection responses are no longer retried
  automatically; transport and server failures retain bounded retry behavior.
- Cumulative stream retry deduplication and cookie-pool recovery remain
  regression-tested.
- Provider failures are visible in the desktop studio instead of appearing as
  completed empty responses.
- Manual retries are bounded and generation controls are locked while a
  request is active.
- Responsive drawers remain below the New/model toolbar.
- The local-only `/healthz` endpoint and aggregate metrics remain available.

## Known boundaries

- Google availability, session validity, quotas, model identity, context
  limits, and shared-network behavior remain upstream-dependent.
- Do not distribute a Google cookie or API key with the application.
- This beta is not a notarized production release and is not a claim of
  unlimited or guaranteed Google access.

## Verification

The repository records the local test, race, vet, build, benchmark, and
release-candidate evidence in `docs/engineering/`. Verify the downloaded
asset against `SHA256SUMS` and the detached signature before installation.
