#!/usr/bin/env node
/**
 * Syncs markdown files from repo docs/ into Astro content collection.
 * Run from repo root: node web/docs/scripts/sync-docs.mjs
 * Or from web/docs: node scripts/sync-docs.mjs
 *
 * Only a small PUBLIC allowlist is copied. Internal runbooks, production/staging
 * playbooks, and operator guides stay in docs/ but must NOT be published here.
 */
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const REPO_DOCS = path.resolve(__dirname, "../../../docs");
const OUT_DIR = path.resolve(__dirname, "../src/content/docs");

// Authored only under web/docs (not from repo docs/). Never deleted by cleanup.
const HAND_WRITTEN_SLUGS = new Set([
  "cli",
  "deployment",
  "functions",
  "getting-started",
  "index",
  "playground",
  "trust-and-verification",
]);

// Public docs site only: integrator / trust / high-level security & onboarding.
// Do NOT add production runbooks, staging guides, disaster recovery, monitoring,
// vault ops, email/DNS internals, or repo-specific setup without legal/security review.
const ALLOWLIST = new Set([
  "DEPLOY_KEYS.md",
  "FUNCTION_WEBHOOKS.md",
  "OPEN_SOURCE_STRATEGY.md",
  "QUICK_START.md",
  "SECURITY.md",
  "TRUST_API.md",
  "TRUST_PROTOCOL_OPEN_SOURCE.md",
  "TRUST_PROTOCOL_SPEC.md",
]);

function toTitle(filename) {
  return filename
    .replace(/\.md$/i, "")
    .replace(/_/g, " ")
    .replace(/\b\w/g, (c) => c.toUpperCase());
}

function toSlug(filename) {
  return filename.replace(/\.md$/i, "").replace(/_/g, "-").toLowerCase();
}

if (!fs.existsSync(REPO_DOCS)) {
  console.error("Repo docs folder not found:", REPO_DOCS);
  process.exit(1);
}

if (!fs.existsSync(OUT_DIR)) {
  fs.mkdirSync(OUT_DIR, { recursive: true });
}

const allowedSlugs = new Set([...ALLOWLIST].map(toSlug));

const files = fs
  .readdirSync(REPO_DOCS)
  .filter((f) => f.endsWith(".md") && ALLOWLIST.has(f));
let count = 0;

for (const file of files) {
  const slug = toSlug(file);
  const title = toTitle(file);
  const srcPath = path.join(REPO_DOCS, file);
  const raw = fs.readFileSync(srcPath, "utf8");
  const hasFrontmatter = raw.startsWith("---");
  let body = raw;
  let frontmatter = `title: ${JSON.stringify(title)}\n`;

  if (hasFrontmatter) {
    const end = raw.indexOf("---", 3);
    if (end !== -1) {
      const fm = raw.slice(3, end).trim();
      if (fm.includes("title:")) {
        frontmatter = fm + "\n";
      } else {
        frontmatter = `title: ${JSON.stringify(title)}\n${fm}\n`;
      }
      body = raw.slice(end + 3).trimStart();
    }
  }

  const out = `---\n${frontmatter}---\n\n${body}`;
  const outFile = `${slug}.md`;
  const outPath = path.join(OUT_DIR, outFile);
  fs.writeFileSync(outPath, out, "utf8");
  count++;
}

// Remove synced files that are no longer public or stale; never touch hand-written pages.
for (const f of fs.readdirSync(OUT_DIR)) {
  if (!f.endsWith(".md")) continue;
  const slug = f.replace(/\.md$/i, "");
  if (HAND_WRITTEN_SLUGS.has(slug)) continue;
  if (allowedSlugs.has(slug)) continue;
  fs.unlinkSync(path.join(OUT_DIR, f));
  console.warn("Removed from public docs (not allowlisted):", f);
}

console.log(`Synced ${count} public doc(s) to ${OUT_DIR}`);
