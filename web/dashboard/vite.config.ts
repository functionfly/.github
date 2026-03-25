import { sentryVitePlugin } from '@sentry/vite-plugin';
import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';
import fs from 'fs';
import path from 'path';
import type { PluginOption } from 'vite';
import { defineConfig } from 'vite';

/** SPA fallback: serve index.html for client routes so refresh on /api-keys, /dashboard, etc. works */
function spaFallbackPlugin() {
  return {
    name: 'spa-fallback',
    configureServer(server: any) {
      server.middlewares.use((req: any, res: any, next: () => void) => {
        const url = req.url?.split('?')[0] ?? '';
        // Skip SPA fallback for real API proxy paths (/api, /api/, /v1) but not client routes like /api-keys
        if (
          req.method !== 'GET' ||
          url === '/api' ||
          url.startsWith('/api/') ||
          url.startsWith('/v1') ||
          url.startsWith('/src') ||
          url.startsWith('/@') ||
          url.startsWith('/node_modules') ||
          url.includes('.')
        ) {
          return next();
        }
        const index = path.join(server.config.root, 'index.html');
        if (!fs.existsSync(index)) return next();
        req.url = '/index.html';
        next();
      });
    },
  };
}

/** Copy Cloudflare Pages _headers and _redirects to dist folder */
function cloudflarePagesPlugin() {
  return {
    name: 'cloudflare-pages',
    closeBundle() {
      const files = ['_headers', '_redirects'];
      const srcDir = path.resolve(__dirname);
      const outDir = path.resolve(__dirname, 'dist');

      for (const file of files) {
        const srcPath = path.join(srcDir, file);
        const outPath = path.join(outDir, file);
        if (fs.existsSync(srcPath)) {
          fs.copyFileSync(srcPath, outPath);
          console.log(`[cloudflare-pages] Copied ${file} to dist/`);
        }
      }
    },
  };
}

/** Generate sitemap.xml and robots.txt into dist after build (Vite/SPA approach). */
function sitemapPlugin() {
  return {
    name: 'vite-plugin-sitemap',
    apply: 'build',
    async closeBundle() {
      const outDir = path.resolve(__dirname, 'dist');
      const scriptPath = path.resolve(__dirname, 'scripts/generate-sitemap.mjs');
      const { spawnSync } = await import('child_process');
      const r = spawnSync(process.execPath, [scriptPath, outDir], {
        stdio: 'inherit',
        cwd: __dirname,
        shell: false,
      });
      if (r.status !== 0) {
        throw new Error(`sitemap script exited with ${r.status}`);
      }
    },
  };
}

// When dashboard runs in Docker, set API_PROXY_TARGET=http://orchestrator-api:8080.
// On host, use localhost for WebSocket compatibility
const apiProxyTarget =
  process.env.VITE_PROXY_API_TARGET || process.env.API_PROXY_TARGET || 'http://127.0.0.1:8080';

function proxyConfigure(proxy: any) {
  proxy.on('error', (err: Error, _req: any, res: any) => {
    console.error('[Vite proxy] Cannot reach API at', apiProxyTarget, err.message);
    if (res && typeof res.writeHead === 'function' && !res.headersSent) {
      try {
        res.writeHead(500, { 'Content-Type': 'text/plain' });
        res.end('Proxy error: ' + err.message);
      } catch (e) {
        console.error('[Vite proxy] Failed to send error response:', e);
      }
    }
  });
  proxy.on('proxyReq', (_proxyReq: any, req: any) => {
    if (req.url?.includes('/auth/login')) {
      console.log('[Vite proxy] Proxying', req.method, req.url, '->', apiProxyTarget);
    }
  });
}

// Log proxy target at startup so we can confirm what the dashboard will use
console.log('[Vite] API proxy target:', apiProxyTarget);

// Dev CSP: permissive enough for local tools (Vercel Analytics, Google Fonts, HMR).
// Production CSP is enforced via public/_headers (Vercel/CF deployments) and the Go backend middleware.
const DEV_CSP = [
  "default-src 'self'",
  "script-src 'self' 'unsafe-inline' 'unsafe-eval' https://va.vercel-scripts.com",
  "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
  "img-src 'self' data: https: blob:",
  "font-src 'self' https://fonts.gstatic.com",
  "connect-src 'self' https://fonts.googleapis.com https://fonts.gstatic.com https://va.vercel-scripts.com http://localhost:8080 https: ws: wss:",
  "worker-src 'self' blob:",
  "frame-ancestors 'none'",
].join('; ');

export default defineConfig({
  appType: 'spa',
  // Cloudflare Pages: output to dist folder (default)
  base: '/',
  plugins: [
    spaFallbackPlugin(),
    cloudflarePagesPlugin(),
    sitemapPlugin(),
    react(),
    tailwindcss(),
    // Upload source maps to Sentry when SENTRY_AUTH_TOKEN is set (e.g. in CI)
    process.env.SENTRY_AUTH_TOKEN
      ? sentryVitePlugin({
          org: process.env.SENTRY_ORG,
          project: process.env.SENTRY_PROJECT,
          sourcemaps: { assets: './dist/**' },
        })
      : undefined,
  ].filter(Boolean) as PluginOption[],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          // Only the React core runtime — must not include packages that merely
          // have "react" in their name (react-router-dom, react-hook-form, etc.)
          // to avoid circular chunk dependencies.
          if (
            /node_modules\/react\//.test(id) ||
            /node_modules\/react-dom\//.test(id) ||
            /node_modules\/scheduler\//.test(id)
          ) {
            return 'react-vendor';
          }
          // UI library components
          if (id.includes('node_modules/@radix-ui/')) {
            return 'radix-ui';
          }
          // Data fetching and state management
          if (
            id.includes('@tanstack/react-query') ||
            id.includes('node_modules/axios/') ||
            id.includes('node_modules/zustand/')
          ) {
            return 'data-vendor';
          }
          // Charts and tables
          if (id.includes('@tanstack/react-table') || id.includes('node_modules/recharts/')) {
            return 'charts-vendor';
          }
          // Utilities (includes react-router-dom, react-hook-form — kept here,
          // NOT in react-vendor, to break the circular chunk dependency)
          if (
            id.includes('node_modules/clsx/') ||
            id.includes('node_modules/tailwind-merge/') ||
            id.includes('node_modules/class-variance-authority/') ||
            id.includes('node_modules/date-fns/') ||
            id.includes('node_modules/framer-motion/') ||
            id.includes('node_modules/lucide-react/') ||
            id.includes('node_modules/react-hook-form/') ||
            id.includes('node_modules/react-router') ||
            id.includes('node_modules/@remix-run/') ||
            id.includes('node_modules/sonner/') ||
            id.includes('node_modules/zod/')
          ) {
            return 'utils-vendor';
          }
        },
      },
    },
    chunkSizeWarningLimit: 1000, // Increase warning limit to 1000KB
  },
  server: {
    port: Number(process.env.VITE_DEV_PORT) || 3000,
    strictPort: false, // try next port if 3000 is in use
    host: true, // listen on 0.0.0.0 so Docker can expose port 3000
    headers: {
      'Cache-Control': 'no-store, no-cache, must-revalidate, max-age=0',
      Pragma: 'no-cache',
      Expires: '0',
      'Content-Security-Policy': DEV_CSP,
      'X-Frame-Options': 'DENY',
      'X-Content-Type-Options': 'nosniff',
      'Referrer-Policy': 'strict-origin-when-cross-origin',
    },
    proxy: {
      // /api/health/... -> backend /health/... (health check endpoints)
      '/api/health': {
        target: apiProxyTarget,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/health/, '/health'),
        configure: proxyConfigure,
      },
      // /api/users/... -> backend /users/... (user endpoints without v1 prefix)
      '/api/users': {
        target: apiProxyTarget,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/users/, '/users'),
        configure: proxyConfigure,
      },
      // /api/auth/... -> backend /auth/... (auth endpoints without v1 prefix)
      '/api/auth': {
        target: apiProxyTarget,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/auth/, '/auth'),
        configure: proxyConfigure,
      },
      // /api/billing/... -> backend /billing/... (billing endpoints without v1 prefix)
      '/api/billing': {
        target: apiProxyTarget,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/billing/, '/billing'),
        configure: proxyConfigure,
      },
      // /api/v1/... -> backend /v1/... (API client calls with v1 prefix)
      '/api/v1': {
        target: apiProxyTarget,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/v1/, '/v1'),
        configure: proxyConfigure,
      },
      // /v1/... -> backend /v1/... (direct v1 calls from FunctionPage, PlaygroundPage, etc.)
      '/v1': {
        target: apiProxyTarget,
        changeOrigin: true,
        configure: proxyConfigure,
      },
      // /api/... -> backend /v1/... (fallback for other API calls)
      '/api': {
        target: apiProxyTarget,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '/v1'),
        configure: proxyConfigure,
        ws: true, // Enable WebSocket proxying for realtime connections
      },
    },
  },
});
