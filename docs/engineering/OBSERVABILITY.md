# Local Observability — Mission 7

The gateway now maintains an in-process aggregate registry in
`internal/metrics`. It never transmits telemetry and it does not store raw
prompts, tool arguments, image contents, cookies, SAPISID values, or auth
headers.

Safe aggregate data is available through authenticated `GET /v1/metrics` and
is also included in the human-facing `/` health response. `/healthz` remains a
small stable orchestration probe and does not include metrics.

Tracked counters include request totals/in-flight work, upstream attempts and
errors, HTTP 429s, stream retries, session-pool gauges/failovers, image upload
and cache hit/miss counts, and estimated tokens. Request and upstream latency
are bounded bucket summaries rather than per-request records.

The values are process-local and reset on restart. They are operational
signals, not a claim of provider telemetry, token-count accuracy, or live
performance measurement.
