# Contributing

Thank you for helping improve BOB Gemini Free. This repository contains
reverse-engineered Google Web RPC behavior, so correctness and reproducibility
matter more than broad refactoring.

## Before opening a pull request

Run the local validation gate from the repository root:

```bash
go test -count=1 ./...
go test -count=1 -race ./...
go vet ./...
go build ./...
git diff --check
```

For desktop changes on macOS, also run:

```bash
scripts/build-wails-local.sh /tmp/bob-gemini-free-wails-check
```

Remove the temporary output after inspection. Do not replace a running
developer executable during updater tests.

## Protocol and security boundaries

- Preserve wire behavior unless a fixture, reproducible bug, security flaw, or
  measured incompatibility justifies a change.
- Add a deterministic regression test before changing Gemini payload,
  authentication, streaming, thinking extraction, upload, or adapter behavior.
- Do not commit cookies, SAPISID values, API keys, personal access tokens,
  release private keys, user prompts, or image contents.
- Do not describe prompt-based tool extraction as native upstream tool calling.
- Keep the gateway loopback-first and preserve the explicit origin/API-key
  boundary.
- Live Google tests are optional and must use an explicitly authorized session;
  the default suite remains hermetic.

## Pull requests and releases

Use a focused branch and pull request with a concise description of the
behavioral invariant being protected. GitHub Actions are intentionally not
required for this repository; local validation and the Cloudflare Pages check
are complementary evidence, not a substitute for review.

Release binaries are built locally with `scripts/release-local.sh`. A release
requires the configured Ed25519 signing key pair, `SHA256SUMS`,
`SHA256SUMS.sig`, and clean-machine verification. Never publish unsigned
artifacts as an updater-compatible release.
