/**
 * Seed script: insert default blog posts, categories, authors, and hero images.
 * Run with: bun run seed:blog  (from cmd/blog-api)
 * Requires DATABASE_URL in the environment (same as the NestJS app).
 */
import { drizzle } from 'drizzle-orm/node-postgres';
import { eq } from 'drizzle-orm';
import * as pg from 'pg';
import * as fs from 'node:fs';
import * as path from 'node:path';
import * as schema from '../db/schema/index';
import {
  defaultBlogPostEntries,
  defaultBlogCategories,
  defaultBlogAuthor,
  DEFAULT_HERO_IMAGE_URL,
} from './default-blog-posts-list';

function loadLocalEnvFileIfPresent() {
  // The seed script is run directly (ts-node/bun) and does not automatically load .env.
  // For local dev, we mirror what the Nest app does by reading cmd/blog-api/.env when needed.
  const required = ['DATABASE_URL'];
  const missing = required.some((k) => !process.env[k]);
  if (!missing) return;

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
    if (!process.env[key]) process.env[key] = value;
  }
}

async function ensureCategory(
  db: ReturnType<typeof drizzle<typeof schema>>,
  def: (typeof defaultBlogCategories)[number],
  now: Date,
): Promise<string> {
  const existing = await db
    .select({ id: schema.categories.id })
    .from(schema.categories)
    .where(eq(schema.categories.slug, def.slug))
    .limit(1);

  if (existing.length > 0) {
    return existing[0].id;
  }

  const [row] = await db
    .insert(schema.categories)
    .values({
      title: def.title,
      slug: def.slug,
      description: def.description,
      color: null,
      icon: null,
      order: def.order,
      createdAt: now,
      updatedAt: now,
    })
    .returning({ id: schema.categories.id });

  console.log(`✓ Category: ${def.title} (${def.slug})`);
  return row.id;
}

async function ensureAuthor(
  db: ReturnType<typeof drizzle<typeof schema>>,
  now: Date,
): Promise<string> {
  const { slug, name, role, bio } = defaultBlogAuthor;
  const existing = await db
    .select({ id: schema.authors.id })
    .from(schema.authors)
    .where(eq(schema.authors.slug, slug))
    .limit(1);

  if (existing.length > 0) {
    return existing[0].id;
  }

  const [row] = await db
    .insert(schema.authors)
    .values({
      name,
      slug,
      bio,
      email: null,
      website: null,
      photo: null,
      socialLinks: null,
      role,
      active: true,
      createdAt: now,
      updatedAt: now,
    })
    .returning({ id: schema.authors.id });

  console.log(`✓ Author: ${name} (${slug})`);
  return row.id;
}

async function seedOrSyncPost(
  db: ReturnType<typeof drizzle<typeof schema>>,
  postData: (typeof defaultBlogPostEntries)[number]['post'],
  slug: string,
  authorId: string,
  categoryId: string,
  now: Date,
): Promise<'inserted' | 'updated'> {
  const heroImage = {
    url: DEFAULT_HERO_IMAGE_URL,
    alt: postData.title,
  };

  const existing = await db
    .select({ id: schema.blogPosts.id })
    .from(schema.blogPosts)
    .where(eq(schema.blogPosts.slug, slug))
    .limit(1);

  if (existing.length === 0) {
    await db.insert(schema.blogPosts).values({
      title: postData.title,
      slug: postData.slug,
      description: postData.description,
      body: postData.body,
      tags: [...postData.tags],
      heroImage,
      status: 'published',
      publishedAt: postData.publishedAt ? new Date(postData.publishedAt) : now,
      scheduledAt: null,
      updatedAt: now,
      seoTitle: postData.seoTitle ?? null,
      seoDescription: postData.seoDescription ?? null,
      keywords: postData.keywords ? [...postData.keywords] : null,
      canonicalUrl: postData.canonicalUrl ?? null,
      ogImage: null,
      authorId,
      categoryId,
    });

    console.log(`✓ Inserted blog post: ${postData.title} (slug: ${slug})`);
    return 'inserted';
  }

  await db
    .update(schema.blogPosts)
    .set({
      authorId,
      categoryId,
      heroImage,
      updatedAt: now,
    })
    .where(eq(schema.blogPosts.slug, slug));

  console.log(`↻ Updated metadata for: ${postData.title} (slug: ${slug})`);
  return 'updated';
}

async function seed() {
  loadLocalEnvFileIfPresent();
  const databaseUrl = process.env.DATABASE_URL;
  if (!databaseUrl) {
    console.error('DATABASE_URL is required');
    process.exit(1);
  }

  const pool = new pg.Pool({ connectionString: databaseUrl });
  const db = drizzle(pool, { schema });

  const now = new Date();
  let inserted = 0;
  let updated = 0;

  console.log('🌱 Starting blog post seeding...');

  const categoryIds = new Map<string, string>();
  for (const cat of defaultBlogCategories) {
    const id = await ensureCategory(db, cat, now);
    categoryIds.set(cat.slug, id);
  }

  const authorId = await ensureAuthor(db, now);

  for (const entry of defaultBlogPostEntries) {
    const categoryId = categoryIds.get(entry.categorySlug);
    if (!categoryId) {
      console.error(`Unknown category slug: ${entry.categorySlug}`);
      process.exit(1);
    }
    const result = await seedOrSyncPost(db, entry.post, entry.slug, authorId, categoryId, now);
    if (result === 'inserted') inserted++;
    if (result === 'updated') updated++;
  }

  console.log(`✅ Seeding complete! Inserted ${inserted} new posts, synced metadata on ${updated} existing posts.`);

  await pool.end();
}

seed().catch((err) => {
  console.error('Seed failed:', err);
  process.exit(1);
});
