// @ts-check
import { defineConfig } from 'astro/config';
import vercel from '@astrojs/vercel/serverless';
import react from '@astrojs/react';
import mdx from '@astrojs/mdx';
import sitemap from '@astrojs/sitemap';
import tailwind from '@astrojs/tailwind';

// https://astro.build/config
export default defineConfig({
  site: 'https://functionfly.com',
  integrations: [
    react(),
    mdx(),
    sitemap({
      changefreq: 'weekly',
      priority: 0.7,
      lastmod: new Date(),
      // Quality gates: exclude thin content from sitemap
      filter: (page) => {
        const path = new URL(page).pathname;

        // Exclude search results and filtered pages
        if (path.includes('/search') || path.includes('?')) {
          return false;
        }

        // Exclude tag pages with low content (would be implemented with CMS)
        if (path.includes('/blog/tag/')) {
          // Logic would check post count per tag
          return false;
        }

        // Exclude empty programmatic pages
        if (path.includes('/category/')) {
          // Logic would check if category has content
          return false;
        }

        return true;
      },
      // Custom priority based on content type
      serialize: (item) => {
        const path = item.url.replace('https://functionfly.com', '');

        // High priority for core pages
        if (path === '/' || path === '/pricing' || path === '/docs') {
          item.priority = 1.0;
        }
        // Medium priority for content
        else if (path.startsWith('/docs/') || path.startsWith('/blog/')) {
          item.priority = 0.8;
        }
        // Lower priority for supporting pages
        else {
          item.priority = 0.6;
        }

        return item;
      },
    }),
    tailwind({
      applyBaseStyles: false,
    }),
  ],
  adapter: vercel(),
  output: 'server',
});
