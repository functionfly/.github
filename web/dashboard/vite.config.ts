import { sentryVitePlugin } from '@sentry/vite-plugin';
import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';
import { visualizer } from 'rollup-plugin-visualizer';
import fs from 'fs';
import path from 'path';
import type { PluginOption } from 'vite';
import { defineConfig } from 'vite';

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

/** Generate sitemap.xml and robots.txt into dist after build */
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

// API proxy target
const apiProxyTarget =
  process.env.VITE_PROXY_API_TARGET ||
  process.env.API_PROXY_TARGET ||
  'http://localhost:8080';

function proxyConfigure(proxy: any) {
  proxy.on('error', (err: Error, _req: any, res: any) => {
    console.error('[Vite proxy] Cannot reach API at', apiProxyTarget, err.message);
    if (res && typeof res.writeHead === 'function' && !res.headersSent) {
      res.writeHead(500, { 'Content-Type': 'text/plain' });
      res.end('Proxy error: ' + err.message);
    }
  });
}

export default defineConfig({
  appType: 'spa',
  base: '/',
  plugins: [
    cloudflarePagesPlugin(),
    sitemapPlugin(),
    react(),
    tailwindcss(),

    process.env.SENTRY_AUTH_TOKEN
      ? sentryVitePlugin({
          org: process.env.SENTRY_ORG,
          project: process.env.SENTRY_PROJECT,
          sourcemaps: { assets: './dist/**' },
        })
      : undefined,
    visualizer({
      filename: 'dist/bundle-stats.html',
      open: false,
      gzipSize: true,
      brotliSize: true,
    }),
  ].filter(Boolean) as PluginOption[],

  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@functionfly/ui-core': path.resolve(__dirname, '../../packages/ui-core/src'),
      '@functionfly/ui-data-visualization': path.resolve(__dirname, '../../packages/ui-data-visualization/src'),
      '@functionfly/ui-adaptive-ux': path.resolve(__dirname, '../../packages/ui-adaptive-ux/src'),
      '@functionfly/ui-agent': path.resolve(__dirname, '../../packages/ui-agent/src'),
      '@functionfly/ui-ai': path.resolve(__dirname, '../../packages/ui-ai/src'),
      '@functionfly/ui-code-intelligence': path.resolve(__dirname, '../../packages/ui-code-intelligence/src'),
      '@functionfly/ui-collaboration': path.resolve(__dirname, '../../packages/ui-collaboration/src'),
      '@functionfly/ui-devops': path.resolve(__dirname, '../../packages/ui-devops/src'),
      '@functionfly/ui-editor': path.resolve(__dirname, '../../packages/ui-editor/src'),
      '@functionfly/ui-extensibility': path.resolve(__dirname, '../../packages/ui-extensibility/src'),
      '@functionfly/ui-futuristic': path.resolve(__dirname, '../../packages/ui-futuristic/src'),
      '@functionfly/ui-ghost': path.resolve(__dirname, '../../packages/ui-ghost/src'),
      '@functionfly/ui-graph': path.resolve(__dirname, '../../packages/ui-graph/src'),
      '@functionfly/ui-marketplace': path.resolve(__dirname, '../../packages/ui-marketplace/src'),
      '@functionfly/ui-marketplace-economy': path.resolve(__dirname, '../../packages/ui-marketplace-economy/src'),
      '@functionfly/ui-memory': path.resolve(__dirname, '../../packages/ui-memory/src'),
      '@functionfly/ui-observability': path.resolve(__dirname, '../../packages/ui-observability/src'),
      '@functionfly/ui-robotics': path.resolve(__dirname, '../../packages/ui-robotics/src'),
      '@functionfly/ui-runtime': path.resolve(__dirname, '../../packages/ui-runtime/src'),
      '@functionfly/ui-security': path.resolve(__dirname, '../../packages/ui-security/src'),
      '@functionfly/ui-simulation': path.resolve(__dirname, '../../packages/ui-simulation/src'),
      '@functionfly/ui-universal-runtime': path.resolve(__dirname, '../../packages/ui-universal-runtime/src'),
      '@functionfly/ui-visualization': path.resolve(__dirname, '../../packages/ui-visualization/src'),
      '@functionfly/shared': path.resolve(__dirname, '../../packages/shared/src'),
    },
    conditions: ['import', 'module', 'es2020', 'es2015', 'require'],
    mainFields: ['module', 'browser', 'main'],
    // Force all React imports to resolve to the same instance, preventing the
    // "multiple copies of React" bug caused by file:-protocol local packages.
    dedupe: ['react', 'react-dom'],
  },

  optimizeDeps: {
    force: true,
    include: [
      'react',
      'react-dom',
      'zustand',
      'zustand/middleware/immer',
      'immer',
      'vanilla-cookieconsent',
      'three',
      '@react-three/fiber',
      '@react-three/drei',
      '@react-three/postprocessing',
    ],
    exclude: [],
  },

  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules')) {
            if (id.includes('react') && (id.includes('react-dom') || id.includes('react-router'))) return 'vendor-react';
            if (id.includes('framer-motion')) return 'vendor-motion';
            if (id.includes('@radix-ui/')) return 'vendor-radix';
            if (id.includes('zustand') || id.includes('immer') || id.includes('date-fns') || id.includes('clsx') || id.includes('tailwind-merge') || id.includes('lodash') || id.includes('axios') || id.includes('zod')) return 'vendor-utils';
            if (id.includes('three') || id.includes('@react-three/')) return 'vendor-three';
            if (id.includes('recharts')) return 'vendor-charts';
            if (id.includes('monaco-editor') || id.includes('@monaco-editor/')) return 'vendor-monaco';
          }
        },
      },
    },
  },

  server: {
    port: Number(process.env.VITE_DEV_PORT) || 3000,
    strictPort: true,
    host: '0.0.0.0',
    proxy: {
      '/api': {
        target: apiProxyTarget,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''),
        configure: proxyConfigure,
        ws: true,
      },
      '/v1': {
        target: apiProxyTarget,
        changeOrigin: true,
        configure: proxyConfigure,
        ws: true,
      },
      '/v2': {
        target: apiProxyTarget,
        changeOrigin: true,
        configure: proxyConfigure,
        ws: true,
      },
    },
  },
});
