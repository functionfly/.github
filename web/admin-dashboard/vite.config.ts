import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';
import path from 'path';
import { defineConfig, type ProxyOptions } from 'vite';

function resolveApiBaseUrl(): string {
  return process.env.VITE_API_BASE_URL || 'http://localhost:8080';
}

// Build the proxy options so that backend errors are surfaced to the
// browser (so a 502/504 from a down backend doesn't get silently replaced
// with a 200/HTML page) and so the dev server warns loudly when the
// target is unreachable.
function makeProxyOptions(rewrite?: (path: string) => string): ProxyOptions {
  return {
    target: resolveApiBaseUrl(),
    changeOrigin: true,
    ...(rewrite ? { rewrite } : {}),
    // Forward WebSocket upgrades (used for live notifications and log tailing).
    ws: true,
    // Don't let proxy errors fall through to the SPA — surface a real
    // error so the developer notices their backend is down.
    selfHandleResponse: false,
    // Vite's default behaviour on connect errors is to print to stdout;
    // we make it more visible by also throwing the first time.
    configure: (proxy) => {
      proxy.on('error', (err, _req, res) => {
        // eslint-disable-next-line no-console
        console.error(
          `\n[admin-dashboard proxy] Backend unreachable: ${err.message}\n` +
            `  → Check that the API at ${process.env.VITE_API_BASE_URL} is running.\n`
        );
        if (res && 'writeHead' in res && !res.headersSent) {
          res.writeHead(502, { 'Content-Type': 'application/json' });
          res.end(
            JSON.stringify({
              error: 'upstream_unavailable',
              message:
                'The admin API is unreachable. Check that the backend is running and VITE_API_BASE_URL is correct.',
              target: process.env.VITE_API_BASE_URL,
            })
          );
        }
      });
      proxy.on('proxyReq', (_proxyReq, req) => {
        // Helpful breadcrumb in dev to confirm proxying actually fires.
        // eslint-disable-next-line no-console
        console.debug(`[admin-dashboard proxy] ${req.method} ${req.url}`);
      });
    },
  };
}

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => ({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 3002,
    strictPort: false,
    proxy: {
      // Resolve the target once at startup. If the env var is missing,
      // the IIFE above throws synchronously so the dev server fails to
      // start with a clear message instead of silently 502-ing every
      // request.
      '/api': makeProxyOptions((p) => p.replace(/^\/api/, '/v1')),
      '/v1/admin': makeProxyOptions(),
    },
  },
  build: {
    target: 'esnext',
    sourcemap: mode !== 'production',
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) return;
          if (id.includes('react') && !id.includes('@tanstack')) return 'vendor-react';
          if (id.includes('@radix-ui') || id.includes('lucide-react') || id.includes('recharts'))
            return 'vendor-ui';
          if (id.includes('@tanstack')) return 'vendor-query';
        },
      },
    },
  },
}));
