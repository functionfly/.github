import react from "@astrojs/react";
import sitemap from "@astrojs/sitemap";
import { defineConfig } from "astro/config";

const site = process.env.PUBLIC_SITE_URL || "https://functionfly.com";
const blogSiteUrl =
  process.env.PUBLIC_BLOG_URL || "https://blog.functionfly.com";

// https://astro.build/config
export default defineConfig({
  site,
  integrations: [
    react(),
    sitemap({
      changefreq: "weekly",
      priority: 0.7,
      lastmod: new Date(),
      filter: (page) => !page.includes("/blog/"),
    }),
  ],
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
