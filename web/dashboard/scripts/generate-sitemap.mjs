#!/usr/bin/env node
/**
 * Vite/SPA sitemap generator. Run after build (e.g. postbuild) to emit sitemap.xml and robots.txt into dist/.
 * Reads config from next-sitemap.config.js for siteUrl, transform, additionalPaths, robots.
 */
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(__dirname, '..');
const outDir = path.resolve(root, process.argv[2] || process.env.BUILD_OUT_DIR || 'dist');

/** Public indexable routes (SPA paths that should appear in the sitemap) */
const STATIC_PATHS = [
  '/',
  '/launch',
  '/coming-soon',
  '/status',
  '/pricing',
  '/features',
  '/integrations',
  '/team',
  '/privacy',
  '/security',
  '/terms',
  '/changelog',
  '/feedback',
  '/faq',
  '/contact',
  '/docs',
  '/products/state-fabric',
  '/registry',
];

function escapeXml(s) {
  if (s == null) return '';
  return String(s)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;');
}

function urlEntry(siteUrl, { loc, lastmod, changefreq, priority }) {
  const fullLoc = loc.startsWith('http')
    ? loc
    : `${siteUrl.replace(/\/$/, '')}${loc.startsWith('/') ? loc : '/' + loc}`;
  let out = `  <url>\n    <loc>${escapeXml(fullLoc)}</loc>\n`;
  if (lastmod) out += `    <lastmod>${escapeXml(lastmod)}</lastmod>\n`;
  if (changefreq) out += `    <changefreq>${escapeXml(changefreq)}</changefreq>\n`;
  if (priority != null) out += `    <priority>${Number(priority)}</priority>\n`;
  out += '  </url>\n';
  return out;
}

async function main() {
  const configPath = path.join(root, 'next-sitemap.config.js');
  const config = (await import(configPath)).default;
  const siteUrl = (config.siteUrl || 'https://functionfly.com').replace(/\/$/, '');

  const transform =
    config.transform ||
    (async (c, loc) => ({
      loc,
      changefreq: c.changefreq,
      priority: c.priority,
      lastmod: c.autoLastmod ? new Date().toISOString() : undefined,
    }));
  const baseConfig = {
    changefreq: config.changefreq ?? 'weekly',
    priority: config.priority ?? 0.7,
    autoLastmod: config.autoLastmod,
  };

  const entries = [];

  for (const loc of STATIC_PATHS) {
    const t = await transform(baseConfig, loc);
    entries.push({
      loc: t.loc || loc,
      lastmod: t.lastmod,
      changefreq: t.changefreq,
      priority: t.priority,
    });
  }

  let additional = [];
  if (typeof config.additionalPaths === 'function') {
    try {
      additional = await config.additionalPaths(config);
    } catch (e) {
      console.warn('[generate-sitemap] additionalPaths failed:', e.message);
    }
  }
  for (const item of additional) {
    const loc = typeof item === 'string' ? item : item.loc;
    const lastmod = item.lastmod ?? (config.autoLastmod ? new Date().toISOString() : undefined);
    const changefreq = item.changefreq ?? baseConfig.changefreq;
    const priority = item.priority ?? baseConfig.priority;
    entries.push({ loc, lastmod, changefreq, priority });
  }

  const sitemapXml =
    '<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n' +
    entries.map((e) => urlEntry(siteUrl, e)).join('') +
    '</urlset>';

  if (!fs.existsSync(outDir)) {
    fs.mkdirSync(outDir, { recursive: true });
  }
  fs.writeFileSync(path.join(outDir, 'sitemap.xml'), sitemapXml, 'utf8');
  console.log('[sitemap] Wrote sitemap.xml with', entries.length, 'URLs');

  if (config.generateRobotsTxt !== false) {
    const opts = config.robotsTxtOptions || {};
    const policies = opts.policies || [
      { userAgent: '*', allow: '/', disallow: ['/dashboard/', '/admin/', '/api/'] },
    ];
    let robots = '';
    for (const p of policies) {
      if (p.userAgent) robots += `User-agent: ${p.userAgent}\n`;
      if (p.allow)
        robots +=
          (Array.isArray(p.allow) ? p.allow : [p.allow]).map((a) => `Allow: ${a}`).join('\n') +
          '\n';
      if (p.disallow)
        robots +=
          (Array.isArray(p.disallow) ? p.disallow : [p.disallow])
            .map((d) => `Disallow: ${d}`)
            .join('\n') + '\n';
    }
    robots += `Sitemap: ${siteUrl}/sitemap.xml\n`;
    fs.writeFileSync(path.join(outDir, 'robots.txt'), robots, 'utf8');
    console.log('[sitemap] Wrote robots.txt');
  }
}

main().catch((err) => {
  console.error('[generate-sitemap]', err);
  process.exit(1);
});
