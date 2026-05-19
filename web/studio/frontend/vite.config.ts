import { defineConfig } from 'vite';
import path from 'path';

export default defineConfig({
  root: __dirname,
  base: './',
  build: {
    outDir: '../src-tauri/dist',
    emptyOutDir: true,
    target: 'esnext',
    minify: 'esbuild',
  },
  server: {
    port: 3001,
    strictPort: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '/v1'),
      },
      '/studio': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
    },
  },
  resolve: {
    alias: {
      '@tauri-apps/api': path.resolve(__dirname, '../node_modules/@tauri-apps/api'),
    },
  },
});