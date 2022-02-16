import { CacheableResponsePlugin } from "workbox-cacheable-response";
import { ExpirationPlugin } from "workbox-expiration";
import { registerRoute } from "workbox-routing";
import { CacheFirst, StaleWhileRevalidate } from "workbox-strategies";

registerRoute(
  ({ request, url }) => request.mode === "navigate" && !url.pathname.startsWith("/manage"),
  new StaleWhileRevalidate({
    cacheName: "pages",
    plugins: [new CacheableResponsePlugin({ statuses: [200] })]
  })
);

registerRoute(
  ({ url }) => url.origin === "https://fonts.googleapis.com",
  new StaleWhileRevalidate({
    cacheName: "google-fonts-stylesheets"
  })
);

registerRoute(
  ({ url }) => url.origin === "https://fonts.gstatic.com",
  new CacheFirst({
    cacheName: "google-fonts-webfonts",
    plugins: [
      new CacheableResponsePlugin({
        statuses: [0, 200]
      }),
      new ExpirationPlugin({
        maxAgeSeconds: 60 * 60 * 24 * 365,
        maxEntries: 30
      })
    ]
  })
);

registerRoute(
  ({ request }) =>
    request.destination === "style" || request.destination === "script" || request.destination === "worker",
  new StaleWhileRevalidate({
    cacheName: "assets",
    plugins: [new CacheableResponsePlugin({ statuses: [200] })]
  })
);

registerRoute(
  ({ url }) => url.pathname.startsWith("/covers/"),
  new CacheFirst({
    cacheName: "covers",
    plugins: [
      new CacheableResponsePlugin({ statuses: [200] }),
      new ExpirationPlugin({
        maxAgeSeconds: 60 * 60 * 24 * 30,
        purgeOnQuotaError: true
      })
    ]
  })
);

registerRoute(
  ({ url }) => url.pathname.startsWith("/pages/"),
  new CacheFirst({
    cacheName: "chapter pages",
    plugins: [
      new CacheableResponsePlugin({ statuses: [200] }),
      new ExpirationPlugin({
        maxEntries: 512,
        maxAgeSeconds: 60 * 60 * 24 * 7,
        purgeOnQuotaError: true
      })
    ]
  })
);
