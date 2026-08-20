/**
 * BOB Gemini Free - Cloudflare Pages Edge Function
 * GET /health — also handles OPTIONS for CORS preflight
 */

const DEPLOY_EPOCH = 1755648000; // Unix timestamp of approximate deploy base (Aug 2026)

function corsHeaders() {
  return {
    "Access-Control-Allow-Origin": "*",
    "Access-Control-Allow-Methods": "GET, OPTIONS",
    "Access-Control-Allow-Headers": "Content-Type, Authorization, x-api-key",
    "Content-Type": "application/json"
  };
}

export async function onRequestOptions() {
  return new Response(null, { status: 204, headers: corsHeaders() });
}

export async function onRequestGet() {
  const uptime = Math.max(0, Math.floor(Date.now() / 1000) - DEPLOY_EPOCH);
  return new Response(JSON.stringify({
    status: "ok",
    version: "v0.1.7",
    engine: "cloudflare-pages-edge",
    uptime_seconds: uptime,
    requests_served: 0,
    tokens_processed: 0,
    estimated_savings_usd: "$0.00",
    note: "Stateless serverless edge — stats reset each request"
  }), {
    headers: corsHeaders()
  });
}
