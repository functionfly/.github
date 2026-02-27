/** @type {import('next-sitemap').IConfig} */
module.exports = {
  siteUrl: process.env.SITE_URL || 'https://functionfly.dev',
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
        disallow: ['/dashboard/', '/admin/', '/api/']
      }
    ]
  },
  transform: async (config, path) => {
    // Custom transform function
    const defaultTransform = {
      loc: path,
      changefreq: config.changefreq,
      priority: config.priority,
      lastmod: config.autoLastmod ? new Date().toISOString() : undefined
    };

    // Customize priority and changefreq based on path
    if (path === '/') {
      return {
        ...defaultTransform,
        priority: 1.0,
        changefreq: 'daily'
      };
    }

    if (path.startsWith('/pricing') || path.startsWith('/features')) {
      return {
        ...defaultTransform,
        priority: 0.9,
        changefreq: 'weekly'
      };
    }

    if (path.startsWith('/team') || path.startsWith('/blog')) {
      return {
        ...defaultTransform,
        priority: 0.6,
        changefreq: 'monthly'
      };
    }

    return defaultTransform;
  },
  additionalPaths: async (config) => {
    const result = [];

    // Add dynamic blog posts if you have them
    // const blogs = await fetchBlogPosts();
    // blogs.forEach((blog) => {
    //   result.push({
    //     loc: `/blog/${blog.slug}`,
    //     changefreq: 'monthly',
    //     priority: 0.6,
    //     lastmod: blog.updatedAt
    //   });
    // });

    return result;
  }
};