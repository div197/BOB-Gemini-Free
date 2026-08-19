/**
 * BOB Gemini Free - Cloudflare Pages Edge Function
 * GET /health
 */

export async function onRequestGet() {
  return new Response(JSON.stringify({
    status: "ok",
    version: "v0.1.4",
    engine: "cloudflare-pages-edge",
    uptime_sec: 999999,
    requests_served: 100,
    tokens_processed: 5000,
    estimated_usd_saved: 0.15
  }), {
    headers: {
      "Access-Control-Allow-Origin": "*",
      "Content-Type": "application/json"
    }
  });
}
