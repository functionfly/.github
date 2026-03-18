/** @type {import('next-sitemap').IConfig} */
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

    if (path.startsWith('/team') || path.startsWith('/blog')) {
      return {
        ...defaultTransform,
        priority: 0.6,
        changefreq: 'monthly',
      };
    }

    return defaultTransform;
  },
  additionalPaths: async (config) => {
    const result = [];

    // Add dynamic blog posts
    try {
      // Try to fetch from API first
      const apiUrl =
        process.env.API_URL || `${process.env.SITE_URL || 'https://functionfly.dev'}/api`;
      const response = await fetch(`${apiUrl}/v1/content/blog?limit=100`, {
        headers: {
          'Content-Type': 'application/json',
        },
        // Add timeout for build-time execution
        signal: AbortSignal.timeout(10000), // 10 second timeout
      });

      if (response.ok) {
        const data = await response.json();
        const posts = data.posts || [];

        posts.forEach((post) => {
          if (post.slug && post.is_published) {
            result.push({
              loc: `/blog/${post.slug}`,
              changefreq: 'monthly',
              priority: 0.6,
              lastmod: post.updated_at || post.created_at || new Date().toISOString(),
            });
          }
        });

        console.log(`Added ${result.length} blog posts to sitemap from API`);
      } else {
        throw new Error(`API returned ${response.status}: ${response.statusText}`);
      }
    } catch (error) {
      console.warn(
        'Failed to fetch blog posts from API, falling back to known posts:',
        error.message
      );

      // Fallback to known blog post slugs if API fails
      const blogPosts = [
        'welcome-functionfly',
        'getting-started-tutorial',
        'compute-capsules-protocol',
        'flywheel-network',
        'secrets-vault',
        'ai-agent-integration',
        'builder-success-story',
      ];

      blogPosts.forEach((slug) => {
        result.push({
          loc: `/blog/${slug}`,
          changefreq: 'monthly',
          priority: 0.6,
          lastmod: new Date().toISOString(),
        });
      });

      console.log(`Added ${result.length} blog posts to sitemap using fallback`);
    }

    return result;
  },
};
