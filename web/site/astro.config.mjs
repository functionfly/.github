import react from "@astrojs/react";
import sitemap from "@astrojs/sitemap";
import sentry from "@sentry/astro";
import vercel from "@astrojs/vercel";
import { defineConfig } from "astro/config";

const site = process.env.PUBLIC_SITE_URL || "https://functionfly.com";
const blogSiteUrl =
  process.env.PUBLIC_BLOG_URL || "https://blog.functionfly.com";

// All supported locales (matches web/dashboard/src/lib/i18n/languages.ts)
const SUPPORTED_LOCALES = [
  "en", "es", "fr", "de", "zh", "ja", "ko",
  "pt", "ar", "ru", "hi", "nl", "pl", "tr", "vi",
];

// https://astro.build/config
export default defineConfig({
  site,
  integrations: [
    react(),
    sentry({
      dsn: process.env.SENTRY_DSN,
      tracesSampleRate: 0.1,
    }),
    sitemap({
      changefreq: "weekly",
      priority: 0.7,
      lastmod: new Date(),
      filter: (page) => !page.includes("/blog/"),
      i18n: {
        defaultLocale: "en",
        locales: Object.fromEntries(
          SUPPORTED_LOCALES.map((code) => [code, code])
        ),
      },
    }),
  ],
  adapter: vercel(),
  i18n: {
    defaultLocale: "en",
    locales: SUPPORTED_LOCALES,
    routing: {
      prefixDefaultLocale: false,
      redirectToDefaultLocale: false,
    },
  },
  output: "static",
  server: {
    host: true,
    port: 4321,
    strictPort: false,
  },
  build: {
    format: "directory",
  },
  redirects: {
    "/blog": {
      destination: blogSiteUrl,
      status: 301,
    },
  },
  vite: {
    server: {
      proxy: {
        "/v1": {
          target: "http://localhost:8080",
          changeOrigin: true,
        },
        "/docs": {
          target: "http://localhost:4322",
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/docs/, ""),
        },
      },
    },
    build: {
      cssMinify: true,
      minify: "esbuild",
    },
  },
});
