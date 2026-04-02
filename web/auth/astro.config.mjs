import react from "@astrojs/react";
import { defineConfig } from "astro/config";

const site = process.env.PUBLIC_AUTH_URL || "https://auth.functionfly.com";

export default defineConfig({
  site,
  integrations: [react()],
  output: "static",
  server: {
    host: true,
    port: 4323,
    strictPort: true,
  },
  build: {
    format: "file",
  },
  vite: {
    server: {
      proxy: {
        "/api": {
          target: "http://localhost:8080",
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
