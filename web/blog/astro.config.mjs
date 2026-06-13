import node from "@astrojs/node";
import react from "@astrojs/react";
import sitemap from "@astrojs/sitemap";
import sentry from "@sentry/astro";
import astroMermaid from "astro-mermaid";
import { defineConfig } from "astro/config";

const site = process.env.PUBLIC_BLOG_SITE_URL || "https://blog.functionfly.com";

// https://astro.build/config
export default defineConfig({
  site,
  integrations: [
    react(),
    sitemap(),
    astroMermaid(),
    sentry({
      dsn: process.env.SENTRY_DSN,
      tracesSampleRate: 0.1,
    }),
  ],
  output: "server",
  adapter: node({ mode: "standalone" }),
  image: {
    service: {
      entrypoint: "astro/assets/services/sharp",
      config: {
        quality: 85,
        formats: ["webp", "jpeg"],
      },
    },
    domains: ["cdn.functionfly.com", "blog-api.functionfly.com"],
    remotePatterns: [
      {
        protocol: "https",
        hostname: "**.functionfly.com",
      },
    ],
  },
  server: {
    host: true,
    port: 4327,
    strictPort: true,
  },
  build: {
    format: "directory",
  },
  vite: {
    ssr: {
      noExternal: ["astro-gtm", "astro-useragent", "astro-robots-txt", "@astro-community/astro-embed-youtube", "@astro-community/astro-embed-twitter", "@astro-community/astro-embed-link-preview"],
    },
    server: {
      proxy: {
        "/api/v1": {
          target: "http://localhost:3001",
          changeOrigin: true,
        },
      },
    },
    build: {
      cssMinify: true,
      minify: "esbuild",
    },
  },
});
