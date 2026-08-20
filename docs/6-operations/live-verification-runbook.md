# Live Verification Runbook

Use this runbook after publishing a new release, updating Cloudflare Pages, or preparing a classroom / workshop session.

It separates three proof levels:

- Deployment proof: the public site and static edge endpoints are reachable.
- Local gateway proof: the built binary starts and exposes expected local endpoints.
- Upstream proof: Gemini-backed generation succeeds through the selected route.

Do not treat one proof level as another. A passing public `/health` check does not prove Gemini generation. A Cloudflare `429` does not prove the local LAN gateway is blocked.

## 1. Public Deployment Proof

Set the public origin:

```bash
ORIGIN="https://bob-gemini-free.abcsteps.com"
```

Check the public HTML:

```bash
curl -L -sS -D /tmp/bob_headers.txt "$ORIGIN/" -o /tmp/bob_index.html
wc -c /tmp/bob_index.html
sed -n '1,20p' /tmp/bob_headers.txt
```

Expected:

- HTTP `200`
- Non-empty HTML body
- Cloudflare / Pages response headers

Check public health:

```bash
curl -sS "$ORIGIN/health"
```

Expected:

- `"status":"ok"`
- version matching the release being verified

Check public model listing:

```bash
curl -sS "$ORIGIN/v1/models"
```

Expected:

- OpenAI-compatible model list JSON

Check service worker cache version:

```bash
curl -sS "$ORIGIN/sw.js" | grep CACHE_NAME
```

Expected:

- cache name matching the release being verified

Check the deployed UI error copy:

```bash
grep -n "API Rate Limit / Upstream Error" /tmp/bob_index.html
```

Expected:

- the string exists after the rate-limit UI fix is deployed

## 2. Local Gateway Proof

Build and start a local gateway on a temporary port:

```bash
make build
./bob-gemini-free --host 127.0.0.1 --port 19610
```

In another terminal:

```bash
curl -sS http://127.0.0.1:19610/
curl -sS http://127.0.0.1:19610/v1/models
curl -sS http://127.0.0.1:19610/v1beta/models
```

Expected:

- health JSON
- OpenAI-compatible model list
- Google-compatible model list

These checks do not prove upstream Gemini generation. They prove the local gateway binary starts and serves local endpoints.

## 3. Local Upstream Generation Proof

Only run this when upstream/provider calls are intended:

```bash
curl http://127.0.0.1:19610/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-3.7-flash",
    "messages": [{"role": "user", "content": "Reply with exactly one word."}],
    "stream": false
  }'
```

Expected:

- HTTP `200`
- OpenAI-compatible completion JSON

If this fails with `429` or `403`, record the status and route. Do not collapse Cloudflare egress limits, local ISP limits, account limits, and session-cookie issues into one generic failure.

## 4. Automated Diagnostic Proof

Only run this when live upstream calls are intended:

```bash
./bob-gemini-free --test --test-url http://127.0.0.1:19610
```

This diagnostic suite sends real requests through multiple protocol surfaces. It is stronger than health checks, but it consumes upstream capacity.

## 5. Bounded Benchmark Proof

Start conservatively:

```bash
./bob-gemini-free --bench --bench-url http://127.0.0.1:19610 --bench-concurrency 3 --bench-requests 6
```

Increase only after the smaller run is clean. For classroom sessions, prefer proof from the actual LAN and account pool that will be used during class.

## 6. Release Build Proof

Verify source and release builds:

```bash
go test -count=1 ./...
go vet ./...
make build
make dist
git diff --check
```

Expected:

- all commands exit `0`
- release binaries appear in `dist/`

## 7. Current Known Failure Interpretation

If public Cloudflare generation returns upstream `429` or `403`, the public edge path may be rate-limited by Gemini from Cloudflare datacenter egress. For concentrated classroom traffic, switch to the local LAN gateway and verify that path separately.

If local health passes but local generation fails, investigate upstream Gemini/session/account/proxy state.

If public `/health` reports an old version while local source has a newer version, the live deployment is stale or has not finished rebuilding.
