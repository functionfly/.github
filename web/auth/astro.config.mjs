import react from "@astrojs/react";
import { defineConfig } from "astro/config";

const site = process.env.PUBLIC_AUTH_URL || "https://auth.functionfly.com";

// Production CSP headers (from _headers) - also used in dev to prevent extension interference
const CSP_VALUE = [
  "default-src 'none'",
  "script-src 'self' 'unsafe-inline' 'wasm-unsafe-eval' blob: https://challenges.cloudflare.com",
  "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
  "img-src 'self' data: blob: https://challenges.cloudflare.com",
  "font-src 'self' https://fonts.gstatic.com",
  "connect-src 'self' https://api.functionfly.com https://api.staging.functionfly.com http://localhost:8080 http://127.0.0.1:8080",
  "frame-ancestors 'none'",
  "base-uri 'self'",
  "form-action 'self' https://api.functionfly.com https://api.staging.functionfly.com",
  "upgrade-insecure-requests",
  "frame-src https://challenges.cloudflare.com",
].join("; ");

export default defineConfig({
  site,
  integrations: [react()],
  output: "server",
  server: {
    host: true,
    port: 4324,
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
      // Apply CSP headers in dev to prevent browser extensions (MetaMask, etc.) from
      // injecting conflicting CSP that breaks Cloudflare Turnstile
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
  },
});
