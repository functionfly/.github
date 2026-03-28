/**
 * HTTP check: every default launch post slug returns 200 from the public blog API.
 * Run with: BLOG_API_BASE_URL=http://localhost:3000 bun run verify:blog
 * Requires blog-api running and DB seeded.
 */
import { defaultBlogPostSlugs } from './default-blog-posts-list';
import * as fs from 'node:fs';
import * as path from 'node:path';

function loadLocalEnvFileIfPresent() {
  const envPath = path.resolve(__dirname, '../..', '.env');
  if (!fs.existsSync(envPath)) return;

  const raw = fs.readFileSync(envPath, 'utf8');
  const lines = raw.split(/\r?\n/);
  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith('#')) continue;
    const idx = trimmed.indexOf('=');
    if (idx === -1) continue;
    const key = trimmed.slice(0, idx).trim();
    const value = trimmed.slice(idx + 1).trim().replace(/^['"]|['"]$/g, '');
    // Do not override explicit CLI/shell env.
    if (!process.env[key]) process.env[key] = value;
  }
}

const base =
  process.env.BLOG_API_BASE_URL?.replace(/\/$/, '') ||
  process.env.PUBLIC_BLOG_API_URL?.replace(/\/$/, '') ||
  'http://localhost:3000';

async function main() {
  loadLocalEnvFileIfPresent();
  const failures: string[] = [];

  console.log(`Checking ${defaultBlogPostSlugs.length} posts at ${base}/api/v1/blog/posts/:slug\n`);

  for (const slug of defaultBlogPostSlugs) {
    const url = `${base}/api/v1/blog/posts/${encodeURIComponent(slug)}`;
    try {
      const res = await fetch(url, { redirect: 'manual' });
      if (res.status !== 200) {
        failures.push(`${slug} → ${res.status} ${res.statusText}`);
        console.log(`✗ ${slug} (${res.status})`);
      } else {
        console.log(`✓ ${slug}`);
      }
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      failures.push(`${slug} → ${msg}`);
      console.log(`✗ ${slug} (${msg})`);
    }
  }

  if (failures.length > 0) {
    console.error(`\n${failures.length} request(s) failed. Start blog-api, seed the DB, and retry.`);
    process.exit(1);
  }

  console.log('\nAll default blog post slugs returned 200.');
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
