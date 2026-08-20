/**
 * BOB Gemini Free - Cloudflare Pages Edge Function
 * GET /health
 */

const DEPLOY_EPOCH = 1755648000; // Unix timestamp of approximate deploy base (Aug 2026)

export async function onRequestGet() {
  const uptime = Math.floor(Date.now() / 1000) - DEPLOY_EPOCH;
  return new Response(JSON.stringify({
    status: "ok",
    version: "v0.1.6",
    engine: "cloudflare-pages-edge",
    uptime_seconds: uptime,
    requests_served: 0,
    tokens_processed: 0,
    estimated_savings_usd: "$0.00",
    note: "Stateless serverless edge — stats reset each request"
  }), {
    headers: {
      "Access-Control-Allow-Origin": "*",
      "Content-Type": "application/json"
    }
  });
}
