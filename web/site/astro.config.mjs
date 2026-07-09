import react from "@astrojs/react";
import sitemap from "@astrojs/sitemap";
import vercel from "@astrojs/vercel";
import robotsTxt from "astro-robots-txt";
import { defineConfig } from "astro/config";

const site = process.env.PUBLIC_SITE_URL || "https://functionfly.com";
const blogSiteUrl =
  process.env.PUBLIC_BLOG_URL || "https://blog.functionfly.com";

// All supported locales (matches web/dashboard/src/lib/i18n/languages.ts)
const SUPPORTED_LOCALES = [
  "en",
  "es",
  "fr",
  "de",
  "zh",
  "ja",
  "ko",
  "pt",
  "ar",
  "ru",
  "hi",
  "nl",
  "pl",
  "tr",
  "vi",
];

const isDev = process.env.NODE_ENV === "development" || !process.env.NODE_ENV;

// https://astro.build/config
export default defineConfig({
  site,
  integrations: [
    react(),
    sitemap({
      changefreq: "weekly",
      lastmod: new Date(),
      i18n: {
        defaultLocale: "en",
        locales: Object.fromEntries(
          SUPPORTED_LOCALES.map((code) => [code, code]),
        ),
      },
      customPages: [
        { url: "/", priority: 1.0 },
        { url: "/pricing/", priority: 1.0 },
        { url: "/registry/", priority: 0.9 },
        { url: "/blog/", priority: 0.9 },
        { url: "/trust/", priority: 0.8 },
        { url: "/security/", priority: 0.8 },
        { url: "/changelog/", priority: 0.7 },
        { url: "/about/", priority: 0.7 },
        { url: "/contact/", priority: 0.6 },
        { url: "/careers/", priority: 0.6 },
      ],
      filter: (page) =>
        !page.url.includes("/studio/") &&
        !page.url.includes("/api/") &&
        !page.url.includes("/@"),
    }),
    robotsTxt({
      sitemap: [
        `${site}/sitemap-index.xml`,
      ],
      rules: [
        {
          userAgent: "*",
          allow: "/",
          disallow: ["/studio/", "/api/", "/docs/"],
        },
      ],
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
    strictPort: true,
    hmr: {
      overlay: true,
    },
  },
  build: {
    format: "directory",
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
      watch: {
        ignored: ["**/node_modules/**", "**/.astro/**"],
      },
    },
    build: {
      cssMinify: true,
      minify: "esbuild",
    },
    optimizeDeps: {
      include: ["react", "react-dom", "framer-motion", "typewriter-effect"],
    },
  },
});
