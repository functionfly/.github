import node from "@astrojs/node";
import react from "@astrojs/react";
import sitemap from "@astrojs/sitemap";
import { defineConfig } from "astro/config";

const site = process.env.PUBLIC_BLOG_SITE_URL || "https://blog.functionfly.com";

// https://astro.build/config
export default defineConfig({
  site,
  integrations: [react(), sitemap()],
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
    strictPort: false,
  },
  build: {
    format: "directory",
  },
  vite: {
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
