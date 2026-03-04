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

// Middleware to disable all caching in dev (avoids ERR_CACHE_READ_FAILURE in WSL/some browsers)
function noCacheMiddleware() {
  return {
    name: 'no-cache',
    configureServer(server: any) {
      server.middlewares.use((_req: any, res: any, next: () => void) => {
        res.setHeader('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
        res.setHeader('Pragma', 'no-cache')
        res.setHeader('Expires', '0')
        next()
      })
    },
  }
}

// Security headers middleware for auth pages and sensitive routes
function securityHeadersMiddleware() {
  return {
    name: 'security-headers',
    configureServer(server: any) {
      server.middlewares.use((req: any, res: any, next: () => void) => {
        // Apply strict security headers to all routes, especially auth pages
        res.setHeader('X-Frame-Options', 'DENY')
        res.setHeader('X-Content-Type-Options', 'nosniff')
        res.setHeader('X-XSS-Protection', '1; mode=block')
        res.setHeader('Referrer-Policy', 'strict-origin-when-cross-origin')

        // Strict CSP for auth pages and API routes
        const isAuthPage = req.url?.includes('/login') || req.url?.includes('/signup') || req.url?.includes('/auth')
        if (isAuthPage) {
          // Very restrictive CSP for auth pages
          res.setHeader('Content-Security-Policy',
            "default-src 'self'; " +
            "script-src 'self' 'unsafe-inline'; " +
            "style-src 'self' 'unsafe-inline'; " +
            "img-src 'self' data: https:; " +
            "font-src 'self'; " +
            "connect-src 'self'; " +
            "frame-ancestors 'none';"
          )
        } else {
          // Standard CSP for other pages
          res.setHeader('Content-Security-Policy',
            "default-src 'self'; " +
            "script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
            "style-src 'self' 'unsafe-inline'; " +
            "img-src 'self' data: https: blob:; " +
            "font-src 'self' https:; " +
            "connect-src 'self' https: wss:;"
          )
        }

        // HSTS for HTTPS
        if (req.headers['x-forwarded-proto'] === 'https' || req.protocol === 'https') {
          res.setHeader('Strict-Transport-Security', 'max-age=31536000; includeSubDomains')
        }

        next()
      })
    },
  }
}

export default defineConfig({
  plugins: [
    noCacheMiddleware(),
    securityHeadersMiddleware(),
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
    },
    proxy: {
      // /api/v1/... -> backend /v1/... (dashboard uses VITE_API_URL=/api)
      // In Docker set API_PROXY_TARGET=http://orchestrator-api:8080
      '/api': {
        target: apiProxyTarget,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''),
        configure: proxyConfigure,
        ws: true, // Enable WebSocket proxying for realtime connections
      },
      // Fallback: /v1/... -> backend /v1/...
      '/v1': {
        target: apiProxyTarget,
        changeOrigin: true,
        configure: proxyConfigure,
      },
    },
  },
})
