'use strict';
/* Service worker del panel: siempre pide todo a la PC y solo usa la copia guardada
   si la PC no responde, así la app abre igual y avisa que no hay conexión.
   La API nunca se guarda. */

const CACHE = 'panel-v3';
const BASE = ['./', 'index.html', 'estilo.css', 'app.js', 'musica.js', 'manifest.json', 'icono.svg', 'icono-192.png', 'icono-512.png'];

self.addEventListener('install', e => {
  e.waitUntil(caches.open(CACHE).then(c => c.addAll(BASE)).then(() => self.skipWaiting()));
});

self.addEventListener('activate', e => {
  e.waitUntil(
    caches.keys()
      .then(claves => Promise.all(claves.filter(k => k !== CACHE).map(k => caches.delete(k))))
      .then(() => self.clients.claim()),
  );
});

self.addEventListener('fetch', e => {
  const url = new URL(e.request.url);
  if (e.request.method !== 'GET' || url.origin !== location.origin || url.pathname.includes('/api/')) return;
  e.respondWith(
    fetch(e.request)
      .then(r => {
        if (r.ok) {
          const copia = r.clone();
          caches.open(CACHE).then(c => c.put(e.request, copia));
        }
        return r;
      })
      .catch(() => caches.match(e.request).then(r => r || caches.match('index.html'))),
  );
});
