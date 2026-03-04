/**
 * Seed script: insert default blog posts into the blog API DB.
 * Run with: npx ts-node -r tsconfig-paths/register src/seed/seed-default-blog-posts.ts
 * Requires DATABASE_URL in the environment (same as the NestJS app).
 */
import { drizzle } from 'drizzle-orm/node-postgres';
import { eq } from 'drizzle-orm';
import * as pg from 'pg';
import * as schema from '../db/schema/index';
import { stateFabricPost, slug as stateFabricSlug } from '../data/default-posts/state-fabric';
import { welcomePost, slug as welcomeSlug } from '../data/default-posts/welcome-functionfly';
import { flywheelPost, slug as flywheelSlug } from '../data/default-posts/flywheel-network';
import { secretsVaultPost, slug as secretsVaultSlug } from '../data/default-posts/secrets-vault';
import { ccpPost, slug as ccpSlug } from '../data/default-posts/compute-capsules-protocol';
import { aiAgentPost, slug as aiAgentSlug } from '../data/default-posts/ai-agent-integration';
import { tutorialPost, slug as tutorialSlug } from '../data/default-posts/getting-started-tutorial';
import { successStoryPost, slug as successStorySlug } from '../data/default-posts/builder-success-story';

// List of all default posts to seed
const defaultPosts = [
  { post: stateFabricPost, slug: stateFabricSlug },
  { post: welcomePost, slug: welcomeSlug },
  { post: flywheelPost, slug: flywheelSlug },
  { post: secretsVaultPost, slug: secretsVaultSlug },
  { post: ccpPost, slug: ccpSlug },
  { post: aiAgentPost, slug: aiAgentSlug },
  { post: tutorialPost, slug: tutorialSlug },
  { post: successStoryPost, slug: successStorySlug },
];

async function seedPost(db: any, postData: any, slug: string, now: Date) {
  const existing = await db
    .select({ id: schema.blogPosts.id })
    .from(schema.blogPosts)
    .where(eq(schema.blogPosts.slug, slug))
    .limit(1);

  if (existing.length > 0) {
    console.log(`Post with slug "${slug}" already exists, skipping.`);
    return false;
  }

  await db.insert(schema.blogPosts).values({
    title: postData.title,
    slug: postData.slug,
    description: postData.description,
    body: postData.body,
    tags: [...postData.tags],
    heroImage: null,
    status: 'published',
    publishedAt: postData.publishedAt ? new Date(postData.publishedAt) : now,
    scheduledAt: null,
    updatedAt: now,
    seoTitle: postData.seoTitle ?? null,
    seoDescription: postData.seoDescription ?? null,
    keywords: postData.keywords ? [...postData.keywords] : null,
    canonicalUrl: postData.canonicalUrl ?? null,
    ogImage: null,
  });

  console.log(`✓ Inserted blog post: ${postData.title} (slug: ${slug})`);
  return true;
}

async function seed() {
  const databaseUrl = process.env.DATABASE_URL;
  if (!databaseUrl) {
    console.error('DATABASE_URL is required');
    process.exit(1);
  }

  const pool = new pg.Pool({ connectionString: databaseUrl });
  const db = drizzle(pool, { schema });

  const now = new Date();
  let insertedCount = 0;

  console.log('🌱 Starting blog post seeding...');

  for (const { post, slug } of defaultPosts) {
    const inserted = await seedPost(db, post, slug, now);
    if (inserted) insertedCount++;
  }

  console.log(`✅ Seeding complete! Inserted ${insertedCount} new blog posts.`);

  await pool.end();
}

seed().catch((err) => {
  console.error('Seed failed:', err);
  process.exit(1);
});
