import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { sentryVitePlugin } from '@sentry/vite-plugin'
import path from 'path'

// When dashboard runs in Docker, set API_PROXY_TARGET=http://orchestrator-api:8080.
// On host, use localhost for WebSocket compatibility
const apiProxyTarget = process.env.VITE_PROXY_API_TARGET || process.env.API_PROXY_TARGET || 'http://127.0.0.1:8080'

function proxyConfigure(proxy: any) {
  proxy.on('error', (err: Error, _req: any, res: any) => {
    console.error('[Vite proxy] Cannot reach API at', apiProxyTarget, err.message)
    if (res && typeof res.writeHead === 'function' && !res.headersSent) {
      try {
        res.writeHead(500, { 'Content-Type': 'text/plain' })
        res.end('Proxy error: ' + err.message)
      } catch (e) {
        console.error('[Vite proxy] Failed to send error response:', e)
      }
    }
  })
  proxy.on('proxyReq', (_proxyReq: any, req: any) => {
    if (req.url?.includes('/auth/login')) {
      console.log('[Vite proxy] Proxying', req.method, req.url, '->', apiProxyTarget)
    }
  })
}

// Log proxy target at startup so we can confirm what the dashboard will use
console.log('[Vite] API proxy target:', apiProxyTarget)

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
].join('; ')

export default defineConfig({
  plugins: [
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
  ].filter(Boolean),
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          // React and React DOM
          if (id.includes('react') || id.includes('react-dom')) {
            return 'react-vendor';
          }
          // UI library components
          if (id.includes('@radix-ui/')) {
            return 'radix-ui';
          }
          // Data fetching and state management
          if (id.includes('@tanstack/react-query') || id.includes('axios') || id.includes('zustand')) {
            return 'data-vendor';
          }
          // Charts and tables
          if (id.includes('@tanstack/react-table') || id.includes('recharts')) {
            return 'charts-vendor';
          }
          // Utilities
          if (
            id.includes('clsx') ||
            id.includes('tailwind-merge') ||
            id.includes('class-variance-authority') ||
            id.includes('date-fns') ||
            id.includes('framer-motion') ||
            id.includes('lucide-react') ||
            id.includes('react-hook-form') ||
            id.includes('react-router-dom') ||
            id.includes('sonner') ||
            id.includes('zod')
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
      'Pragma': 'no-cache',
      'Expires': '0',
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
})
