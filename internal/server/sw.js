/*
  BOB Gemini Free — Local Gateway Service Worker

  This worker is served by the local gateway only. The hosted static bundle
  has a separate root-entry service worker in web/sw.js.
*/

const CACHE_NAME = "bob-gemini-local-studio-" + __BOB_CACHE_VERSION__;
const PRECACHE_ASSETS = ["./playground", "./manifest.json", "./favicon.ico"];

self.addEventListener("install", (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => cache.addAll(PRECACHE_ASSETS))
      .then(() => self.skipWaiting())
  );
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches.keys()
      .then((keys) => Promise.all(
        keys
          .filter((key) => key.startsWith("bob-gemini-local-studio-") && key !== CACHE_NAME)
          .map((key) => caches.delete(key))
      ))
      .then(() => self.clients.claim())
  );
});

self.addEventListener("fetch", (event) => {
  const request = event.request;
  const url = new URL(request.url);

  // Never cache gateway APIs, uploads, or non-GET requests.
  if (
    request.method !== "GET" ||
    url.pathname.startsWith("/v1/") ||
    url.pathname.startsWith("/v1beta/") ||
    url.origin !== self.location.origin
  ) {
    return;
  }

  event.respondWith(
    fetch(request)
      .then((response) => {
        if (response && response.ok) {
          const copy = response.clone();
          caches.open(CACHE_NAME).then((cache) => cache.put(request, copy)).catch(() => {});
        }
        return response;
      })
      .catch(() => caches.match(request).then((cached) => {
        if (cached) return cached;
        if (request.mode === "navigate") return caches.match("./playground");
        return Response.error();
      }))
  );
});
