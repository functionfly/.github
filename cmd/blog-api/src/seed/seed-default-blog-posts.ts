/**
 * Seed script: insert default blog posts (e.g. State Fabric) into the blog API DB.
 * Run with: npx ts-node -r tsconfig-paths/register src/seed/seed-default-blog-posts.ts
 * Requires DATABASE_URL in the environment (same as the NestJS app).
 */
import { drizzle } from 'drizzle-orm/node-postgres';
import { eq } from 'drizzle-orm';
import * as pg from 'pg';
import * as schema from '../db/schema/index';
import { stateFabricPost, slug as stateFabricSlug } from '../data/default-posts/state-fabric';

async function seed() {
  const databaseUrl = process.env.DATABASE_URL;
  if (!databaseUrl) {
    console.error('DATABASE_URL is required');
    process.exit(1);
  }

  const pool = new pg.Pool({ connectionString: databaseUrl });
  const db = drizzle(pool, { schema });

  const now = new Date();

  // State Fabric default post
  const existing = await db
    .select({ id: schema.blogPosts.id })
    .from(schema.blogPosts)
    .where(eq(schema.blogPosts.slug, stateFabricSlug))
    .limit(1);

  if (existing.length > 0) {
    console.log(`Post with slug "${stateFabricSlug}" already exists, skipping.`);
    await pool.end();
    return;
  }

  await db.insert(schema.blogPosts).values({
    title: stateFabricPost.title,
    slug: stateFabricPost.slug,
    description: stateFabricPost.description,
    body: stateFabricPost.body,
    tags: [...stateFabricPost.tags],
    heroImage: null,
    status: 'published',
    publishedAt: stateFabricPost.publishedAt ? new Date(stateFabricPost.publishedAt) : now,
    scheduledAt: null,
    updatedAt: now,
    seoTitle: stateFabricPost.seoTitle ?? null,
    seoDescription: stateFabricPost.seoDescription ?? null,
    keywords: stateFabricPost.keywords ? [...stateFabricPost.keywords] : null,
    canonicalUrl: stateFabricPost.canonicalUrl ?? null,
    ogImage: null,
  });

  console.log(`Inserted default blog post: ${stateFabricPost.title} (slug: ${stateFabricSlug})`);
  await pool.end();
}

seed().catch((err) => {
  console.error('Seed failed:', err);
  process.exit(1);
});
