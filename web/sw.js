/*
  BOB Gemini Free — Local-First PWA Service Worker
  Engineered by ABCsteps (abcsteps.com) & Divyanshu Singh Chouhan (@div197)
*/

const CACHE_NAME = 'bob-gemini-studio-v0.1.3';
const PRECACHE_ASSETS = [
  './',
  './index.html',
  './manifest.json',
  'https://cdn.jsdelivr.net/npm/marked@12.0.0/marked.min.js',
  'https://cdn.jsdelivr.net/npm/dompurify@3.0.9/dist/purify.min.js',
  'https://cdn.jsdelivr.net/npm/katex@0.16.9/dist/katex.min.css',
  'https://cdn.jsdelivr.net/npm/katex@0.16.9/dist/katex.min.js',
  'https://cdn.jsdelivr.net/npm/katex@0.16.9/dist/contrib/auto-render.min.js',
  'https://cdn.jsdelivr.net/npm/prismjs@1.29.0/themes/prism-tomorrow.min.css',
  'https://cdn.jsdelivr.net/npm/prismjs@1.29.0/prism.min.js',
  'https://cdn.jsdelivr.net/npm/prismjs@1.29.0/plugins/autoloader/prism-autoloader.min.js',
  'https://cdn.jsdelivr.net/npm/mermaid@10.9.0/dist/mermaid.min.js'
];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.addAll(PRECACHE_ASSETS).catch(() => {});
    }).then(() => self.skipWaiting())
  );
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((keys) => {
      return Promise.all(
        keys.filter((key) => key !== CACHE_NAME).map((key) => caches.delete(key))
      );
    }).then(() => self.clients.claim())
  );
});

self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url);

  // Always pass-through API completions, model endpoints and localhost requests to network
  if (
    url.pathname.startsWith('/v1/') ||
    url.pathname.startsWith('/v1beta/') ||
    url.hostname === 'localhost' ||
    url.hostname === '127.0.0.1' ||
    event.request.method !== 'GET'
  ) {
    return;
  }

  // Cache-first strategy for static scripts, fonts & stylesheets
  event.respondWith(
    caches.match(event.request).then((cachedResponse) => {
      if (cachedResponse) {
        return cachedResponse;
      }
      return fetch(event.request).then((networkResponse) => {
        if (!networkResponse || networkResponse.status !== 200 || networkResponse.type !== 'basic') {
          return networkResponse;
        }
        const responseToCache = networkResponse.clone();
        caches.open(CACHE_NAME).then((cache) => {
          cache.put(event.request, responseToCache);
        });
        return networkResponse;
      }).catch(() => {
        // Fallback to cached index.html for navigation requests
        if (event.request.mode === 'navigate') {
          return caches.match('./index.html');
        }
      });
    })
  );
});
