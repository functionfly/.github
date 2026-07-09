import react from "@astrojs/react";
import sitemap from "@astrojs/sitemap";
import sentry from "@sentry/astro";
import node from "@astrojs/node";
import { defineConfig } from "astro/config";

const site = process.env.PUBLIC_AUTH_URL || "https://auth.functionfly.com";

// All supported locales (matches web/site and web/dashboard)
const SUPPORTED_LOCALES = [
  "en", "es", "fr", "de", "zh", "ja", "ko",
  "pt", "ar", "ru", "hi", "nl", "pl", "tr", "vi",
];

const isDev = process.env.NODE_ENV !== 'production';

const apiOrigins = isDev
  ? 'https://api.functionfly.com https://api.staging.functionfly.com http://localhost:8080 http://127.0.0.1:8080'
  : 'https://api.functionfly.com https://api.staging.functionfly.com';

const CSP_VALUE = [
  "default-src 'none'",
  "script-src 'self' 'unsafe-inline' 'wasm-unsafe-eval' blob: https://challenges.cloudflare.com https://www.googletagmanager.com https://cdn.mxpnl.com",
  "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
  "img-src 'self' data: blob: https://challenges.cloudflare.com https://www.google-analytics.com",
  "font-src 'self' https://fonts.gstatic.com",
  `connect-src 'self' ${apiOrigins} https://www.google-analytics.com https://www.googletagmanager.com https://api.mixpanel.com`,
  "frame-ancestors 'none'",
  "base-uri 'self'",
  "form-action 'self' https://api.functionfly.com https://api.staging.functionfly.com",
  "upgrade-insecure-requests",
  "frame-src https://challenges.cloudflare.com",
].join("; ");

export default defineConfig({
  site,
  integrations: [
    react(),
    sitemap({
      i18n: {
        defaultLocale: "en",
        locales: Object.fromEntries(
          SUPPORTED_LOCALES.map((code) => [code, code])
        ),
      },
    }),
    sentry({
      dsn: process.env.SENTRY_DSN,
      tracesSampleRate: 0.1,
    }),
  ],
  i18n: {
    defaultLocale: "en",
    locales: SUPPORTED_LOCALES,
    routing: {
      prefixDefaultLocale: false,
      redirectToDefaultLocale: false,
    },
  },
  adapter: node({
    mode: "standalone",
  }),
  output: "server",
  vite: {
    server: {
      host: true,
      port: 4324,
      strictPort: true,
      proxy: {
        "/api": {
          target: "http://localhost:8080",
          changeOrigin: true,
        },
      },
      headers: {
        "Content-Security-Policy": CSP_VALUE,
        "X-Frame-Options": "DENY",
        "X-Content-Type-Options": "nosniff",
        "Referrer-Policy": "strict-origin-when-cross-origin",
      },
    },
    build: {
      cssMinify: true,
      minify: "esbuild",
    },
    optimizeDeps: {
      include: [
        "react",
        "react-dom",
        "clsx",
        "tailwind-merge",
      ],
    },
  },
});
