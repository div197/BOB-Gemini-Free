/**
 * BOB Gemini Free - Cloudflare Pages Edge Function
 * GET /v1/models
 */

export async function onRequestOptions() {
  return new Response(null, {
    status: 204,
    headers: {
      "Access-Control-Allow-Origin": "*",
      "Access-Control-Allow-Methods": "GET, OPTIONS",
      "Access-Control-Allow-Headers": "Content-Type, Authorization"
    }
  });
}

export async function onRequestGet() {
  const models = [
    { id: "gemini-3.7-flash", object: "model", created: 1739836800, owned_by: "google" },
    { id: "gemini-3.7-flash-thinking", object: "model", created: 1739836800, owned_by: "google" },
    { id: "gemini-3.1-pro", object: "model", created: 1739836800, owned_by: "google" },
    { id: "gemini-3.6-flash", object: "model", created: 1739836800, owned_by: "google" },
    { id: "imagen-3", object: "model", created: 1739836800, owned_by: "google" }
  ];

  return new Response(JSON.stringify({ object: "list", data: models }), {
    headers: {
      "Access-Control-Allow-Origin": "*",
      "Content-Type": "application/json"
    }
  });
}
