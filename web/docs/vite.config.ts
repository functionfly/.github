import { defineConfig } from 'vite'
import path from 'path'
import { fileURLToPath, URL } from 'url'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [],
  esbuild: {
    target: 'es2022',
    exclude: ['@astrojs/react', '@astrojs/starlight'],
  },
  build: {
    target: 'es2022',
  },
  optimizeDeps: {
    esbuildOptions: {
      target: 'esnext',
    },
    exclude: ['@astrojs/react'],
  },
  ssr: {
    noExternal: ['@astrojs/react'],
  },
  resolve: {
    alias: {
      '@': path.resolve(path.dirname(fileURLToPath(import.meta.url)), './src'),
      '@theme': path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../dashboard/src/styles'),
    },
  },
  server: {
    port: 3001,
    proxy: {
      '/docs': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/playground': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
