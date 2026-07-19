const CACHE_NAME = 'sadqa-ledger-v1';
const ASSETS = [
  '/',
  '/static/vendor/basecoat/basecoat.min.css',
  '/static/css/output.css',
  '/static/vendor/basecoat/basecoat.min.js',
  '/static/vendor/htmx/htmx.min.js',
  '/manifest.json'
];

self.addEventListener('install', event => {
  event.waitUntil(
    caches.open(CACHE_NAME).then(cache => {
      return cache.addAll(ASSETS);
    })
  );
});

self.addEventListener('fetch', event => {
  event.respondWith(
    fetch(event.request).catch(() => {
      return caches.match(event.request);
    })
  );
});
