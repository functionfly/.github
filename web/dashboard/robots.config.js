module.exports = {
  policy: [
    {
      userAgent: '*',
      allow: '/',
      disallow: [
        '/dashboard/',
        '/admin/',
        '/api/',
        '/_next/',
        '/auth/',
        '/onboarding/',
        '/settings/',
        '/functions/',
        '/providers/',
        '/analytics/',
        '/search'
      ]
    },
    {
      userAgent: 'Googlebot',
      allow: '/',
      disallow: ['/api/']
    }
  ],
  sitemap: 'https://functionfly.dev/sitemap.xml',
  host: 'https://functionfly.dev'
};