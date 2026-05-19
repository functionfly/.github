import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';
import path from 'path';
import { defineConfig } from 'vite';

// https://vitejs.dev/config/
export default defineConfig({
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
      '/api': {
        target: process.env.VITE_API_BASE_URL ||
          (() => { throw new Error('VITE_API_BASE_URL environment variable is required'); })(),
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '/v1'),
      },
      '/v1/admin': {
        target: process.env.VITE_API_BASE_URL ||
          (() => { throw new Error('VITE_API_BASE_URL environment variable is required'); })(),
        changeOrigin: true,
      },
    },
  },
  build: {
    target: 'esnext',
    sourcemap: false,
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
});
