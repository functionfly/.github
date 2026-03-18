#!/usr/bin/env node
/**
 * Syncs markdown files from repo docs/ into Astro content collection.
 * Run from repo root: node web/docs/scripts/sync-docs.mjs
 * Or from web/docs: node scripts/sync-docs.mjs
 */
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const REPO_DOCS = path.resolve(__dirname, '../../../docs');
const OUT_DIR = path.resolve(__dirname, '../src/content/docs');

// Only sync docs we want to show to everyone (public docs site).
// Internal/reference-only docs (runbooks, audit reports, implementation status, etc.) are excluded.
const ALLOWLIST = new Set([
  'CDN_SETUP.md',
  'CLOUDFLARE.md',
  'COST_OPTIMIZED_DEPLOYMENT.md',
  'DISASTER_RECOVERY_RUNBOOK.md',
  'DOMAIN_AND_COMING_SOON_SETUP.md',
  'EMAIL_CONFIGURATION.md',
  'FLY_DEPLOYMENT.md',
  'FLYPY_COMPLEX_MODE.md',
  'GIT_WORKFLOW.md',
  'GITHUB_REPO_SETUP.md',
  'LOCAL_POSTGRES_17.md',
  'MIGRATION_GUIDE.md',
  'MONITORING.md',
  'NEON.md',
  'OBJECT_STORAGE.md',
  'OPEN_SOURCE_STRATEGY.md',
  'PGVECTOR_SETUP.md',
  'PERFORMANCE_TUNING_GUIDE.md',
  'PRODUCTION_DEPLOYMENT.md',
  'PRODUCTION_README.md',
  'QUICK_START.md',
  'REALTIME_FEATURES_README.md',
  'SECURITY.md',
  'STAGING.md',
  'VAULT_OPERATIONS.md',
  'VERSIONING_SYSTEM.md',
]);

function toTitle(filename) {
  return filename
    .replace(/\.md$/i, '')
    .replace(/_/g, ' ')
    .replace(/\b\w/g, (c) => c.toUpperCase());
}

function toSlug(filename) {
  return filename
    .replace(/\.md$/i, '')
    .replace(/_/g, '-')
    .toLowerCase();
}

if (!fs.existsSync(REPO_DOCS)) {
  console.error('Repo docs folder not found:', REPO_DOCS);
  process.exit(1);
}

if (!fs.existsSync(OUT_DIR)) {
  fs.mkdirSync(OUT_DIR, { recursive: true });
} else {
  // Remove existing synced files so dropped docs disappear from the site
  for (const f of fs.readdirSync(OUT_DIR)) {
    if (f.endsWith('.md')) fs.unlinkSync(path.join(OUT_DIR, f));
  }
}

const files = fs.readdirSync(REPO_DOCS).filter((f) => f.endsWith('.md') && ALLOWLIST.has(f));
let count = 0;

for (const file of files) {
  const slug = toSlug(file);
  const title = toTitle(file);
  const srcPath = path.join(REPO_DOCS, file);
  const raw = fs.readFileSync(srcPath, 'utf8');
  const hasFrontmatter = raw.startsWith('---');
  let body = raw;
  let frontmatter = `title: ${JSON.stringify(title)}\n`;

  if (hasFrontmatter) {
    const end = raw.indexOf('---', 3);
    if (end !== -1) {
      const fm = raw.slice(3, end).trim();
      if (fm.includes('title:')) {
        frontmatter = fm + '\n';
      } else {
        frontmatter = `title: ${JSON.stringify(title)}\n${fm}\n`;
      }
      body = raw.slice(end + 3).trimStart();
    }
  }

  const out = `---\n${frontmatter}---\n\n${body}`;
  const outFile = `${slug}.md`;
  const outPath = path.join(OUT_DIR, outFile);
  fs.writeFileSync(outPath, out, 'utf8');
  count++;
}

console.log(`Synced ${count} docs to ${OUT_DIR}`);
