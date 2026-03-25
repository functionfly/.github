import react from "@astrojs/react";
import { defineConfig } from "astro/config";

// https://astro.build/config
export default defineConfig({
  site: "https://docs.functionfly.com",
  integrations: [react()],
  output: "static",
  server: {
    host: true,
    port: 4322,
    strictPort: false,
  },
  build: {
    format: "directory",
  },
  vite: {
    build: {
      cssMinify: true,
      minify: "esbuild",
    },
  },
});
