import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';
import path from 'path';
import { defineConfig } from 'vite';

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
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@functionfly/ui-core': path.resolve(__dirname, '../../packages/ui-core/src'),
      '@functionfly/ui-data-visualization': path.resolve(__dirname, '../../packages/ui-data-visualization/src'),
      '@functionfly/shared': path.resolve(__dirname, '../../packages/shared/src'),
    },
    conditions: ['import', 'module', 'es2020', 'es2015', 'require'],
    mainFields: ['module', 'browser', 'main'],
    dedupe: ['react', 'react-dom'],
  },
  optimizeDeps: {
    force: true,
    include: ['react', 'react-dom', 'zustand'],
  },
  server: {
    port: Number(process.env.VITE_DEV_PORT) || 3003,
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
    },
    fs: { strict: false },
  },
});
