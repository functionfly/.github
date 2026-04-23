/** @type {{ siteUrl: string; changefreq: string; priority: number; transform: Function; additionalPaths: Function; robotsTxtOptions: object; generateRobotsTxt: boolean }} */
export default {
  siteUrl: process.env.SITE_URL || 'https://functionfly.com',
  generateRobotsTxt: true,
  generateIndexSitemap: false,
  changefreq: 'weekly',
  priority: 0.7,
  sitemapSize: 5000,
  robotsTxtOptions: {
    policies: [
      {
        userAgent: '*',
        allow: '/',
        disallow: ['/dashboard/', '/admin/', '/api/'],
      },
    ],
  },
  transform: async (config, path) => {
    // Custom transform function
    const defaultTransform = {
      loc: path,
      changefreq: config.changefreq,
      priority: config.priority,
      lastmod: config.autoLastmod ? new Date().toISOString() : undefined,
    };

    // Customize priority and changefreq based on path
    if (path === '/') {
      return {
        ...defaultTransform,
        priority: 1.0,
        changefreq: 'daily',
      };
    }

    if (path.startsWith('/pricing') || path.startsWith('/features')) {
      return {
        ...defaultTransform,
        priority: 0.9,
        changefreq: 'weekly',
      };
    }

    if (path.startsWith('/team')) {
      return {
        ...defaultTransform,
        priority: 0.6,
        changefreq: 'monthly',
      };
    }

    return defaultTransform;
  },
  // Blog URLs are indexed on the marketing site (web/site) sitemap, not the app.
  additionalPaths: async () => [],
};
