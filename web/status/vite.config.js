import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';
import fs from 'fs';
import path from 'path';
import { defineConfig } from 'vite';
// SPA fallback for client-side routing
function spaFallbackPlugin() {
    return {
        name: 'spa-fallback',
        configureServer(server) {
            server.middlewares.use((req, res, next) => {
                const url = req.url?.split('?')[0] ?? '';
                if (req.method !== 'GET' ||
                    url.startsWith('/api') ||
                    url.startsWith('/src') ||
                    url.startsWith('/@') ||
                    url.startsWith('/node_modules') ||
                    url.includes('.')) {
                    return next();
                }
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
export default defineConfig({
    appType: 'spa',
    base: '/',
    plugins: [
        // Order matters: Tailwind should process before React for better performance
        tailwindcss(),
        react(),
        spaFallbackPlugin(),
        cloudflarePagesPlugin(),
    ],
    resolve: {
        alias: {
            '@': path.resolve(__dirname, './src'),
        },
    },
    build: {
        // Speed up builds
        target: 'esnext',
        minify: 'esbuild',
        cssMinify: true,
        rollupOptions: {
            output: {
                manualChunks(id) {
                    if (id.includes('node_modules')) {
                        if (id.includes('react') || id.includes('react-dom')) {
                            return 'react';
                        }
                        if (id.includes('react-router-dom')) {
                            return 'router';
                        }
                        if (id.includes('@radix-ui')) {
                            return 'ui';
                        }
                        if (id.includes('recharts')) {
                            return 'charts';
                        }
                        if (id.includes('framer-motion')) {
                            return 'animation';
                        }
                        if (id.includes('clsx') || id.includes('tailwind-merge') || id.includes('date-fns')) {
                            return 'utils';
                        }
                    }
                },
            },
        },
        // Faster CSS processing
        cssCodeSplit: true,
    },
    server: {
        port: 3001, // Different from dashboard (3000)
        host: true,
        headers: {
            'Cache-Control': 'no-store, no-cache, must-revalidate, max-age=0',
            'X-Frame-Options': 'DENY',
            'X-Content-Type-Options': 'nosniff',
        },
        proxy: {
            '/api': {
                target: process.env.VITE_API_TARGET || 'http://localhost:8080',
                changeOrigin: true,
                rewrite: (path) => path.replace(/^\/api/, ''),
                ws: true,
            },
            '/ws': {
                target: process.env.VITE_API_TARGET || 'http://localhost:8080',
                changeOrigin: true,
                ws: true,
            },
        },
    },
});
