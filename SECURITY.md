# Security Policy

BOB Gemini Free is a local gateway that may hold Google session cookies and
forward privileged requests. Treat security reports involving credentials,
localhost trust, upstream authentication, image fetching, release updates, or
protocol parsing as sensitive.

## Reporting a vulnerability

Please do not open a public issue for an undisclosed vulnerability. Use
[GitHub Private Vulnerability Reporting](https://github.com/div197/BOB-Gemini-Free/security/advisories/new)
or a private GitHub Security Advisory so the report can be reviewed before
public disclosure.

Include:

- affected commit, version, or platform;
- a minimal reproduction that contains no real cookies, tokens, prompts, or
  image data;
- impact and likely attack boundary; and
- any proposed mitigation or regression test.

Do not attach `cookie.txt`, SAPISID values, API keys, release private keys,
personal access tokens, or real user prompts. Redact logs and replace all
credentials with clearly fake fixtures.

## Supported security surface

Reports are especially valuable for:

- malicious-webpage access to the loopback gateway;
- authentication, cookie, or session-pool leakage;
- SSRF or unsafe remote-image handling;
- updater signature, manifest, rollback, or replacement failures;
- malformed Gemini Web RPC input causing crashes or resource exhaustion; and
- desktop lifecycle or endpoint-handoff bypasses.

The project does not promise upstream Google availability, model identity,
rate limits, or compatibility. Those are provider-dependent behavior rather
than security guarantees.

## Public disclosure

Please allow maintainers reasonable time to reproduce, fix, test, and publish a
patched release before public disclosure. The repository's release process
requires signed manifests and fail-closed updater verification; unsigned
artifacts must not be used to distribute a security fix.
