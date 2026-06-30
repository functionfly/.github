#!/usr/bin/env bun
/**
 * Migrate blog posts, authors, and categories from PostgreSQL to Sanity CMS.
 *
 * Usage:
 *   SANITY_API_TOKEN=skxxx bun run scripts/migrate-blog-to-sanity.ts
 *
 * Requires a Sanity API token with Editor or Admin access.
 */

import { createClient } from "@sanity/client";
import pg from "pg";

const SANITY_PROJECT_ID = "sg1k76uk";
const SANITY_DATASET = "production";
const token = process.env.SANITY_API_TOKEN;

if (!token) {
  console.error("Missing SANITY_API_TOKEN — generate one at https://sanity.io/manage/project/sg1k76uk/api#tokens");
  process.exit(1);
}

const sanity = createClient({
  projectId: SANITY_PROJECT_ID,
  dataset: SANITY_DATASET,
  apiVersion: "2024-01-01",
  token,
  useCdn: false,
});

const db = new pg.Pool({
  host: process.env.DB_HOST || "localhost",
  port: Number(process.env.DB_PORT || 5432),
  user: process.env.DB_USER || "postgres",
  password: process.env.DB_PASSWORD || "postgres",
  database: process.env.DB_NAME || "functionfly",
});

// ---- helpers ----

function sanitizeId(uuid: string): string {
  return uuid.replace(/[^a-zA-Z0-9._-]/g, "");
}

function authorSanityId(pgId: string): string {
  return `author-${sanitizeId(pgId)}`;
}

function categorySanityId(pgId: string): string {
  return `category-${sanitizeId(pgId)}`;
}

function postSanityId(pgId: string): string {
  return `blogPost-${sanitizeId(pgId)}`;
}

// ---- migration ----

async function migrateAuthors(): Promise<Map<string, string>> {
  const { rows } = await db.query(
    `SELECT id, name, slug, bio, role FROM blog_authors`
  );

  const idMap = new Map<string, string>();
  console.log(`\nMigrating ${rows.length} authors…`);

  for (const row of rows) {
    const sanityId = authorSanityId(row.id);
    idMap.set(row.id, sanityId);

    await sanity.createOrReplace({
      _type: "author",
      _id: sanityId,
      name: row.name,
      slug: { _type: "slug", current: row.slug },
      bio: row.bio || undefined,
      role: row.role || undefined,
    });
    console.log(`  ✓ ${row.name} (${row.slug})`);
  }

  return idMap;
}

async function migrateCategories(authorIdMap: Map<string, string>): Promise<Map<string, string>> {
  const { rows } = await db.query(
    `SELECT id, title, slug, description, color FROM blog_categories`
  );

  const idMap = new Map<string, string>();
  console.log(`\nMigrating ${rows.length} categories…`);

  for (const row of rows) {
    const sanityId = categorySanityId(row.id);
    idMap.set(row.id, sanityId);

    await sanity.createOrReplace({
      _type: "category",
      _id: sanityId,
      title: row.title,
      slug: { _type: "slug", current: row.slug },
      description: row.description || undefined,
      color: row.color || undefined,
    });
    console.log(`  ✓ ${row.title} (${row.slug})`);
  }

  return idMap;
}

async function migratePosts(
  authorIdMap: Map<string, string>,
  categoryIdMap: Map<string, string>,
) {
  const { rows } = await db.query(
    `SELECT id, title, slug, description, body, tags, hero_image, status,
            published_at, seo_title, seo_description, author_id, category_id
     FROM blog_posts
     ORDER BY published_at DESC NULLS LAST`
  );

  console.log(`\nMigrating ${rows.length} blog posts…`);

  for (const row of rows) {
    const sanityId = postSanityId(row.id);

    // Resolve references
    const authorRef = row.author_id ? authorIdMap.get(row.author_id) : null;
    const categoryRef = row.category_id ? categoryIdMap.get(row.category_id) : null;

    // Parse hero image from JSONB
    let heroImage: any = undefined;
    if (row.hero_image) {
      const img = typeof row.hero_image === "string"
        ? JSON.parse(row.hero_image)
        : row.hero_image;
      if (img?.url) {
        heroImage = {
          _type: "image",
          // We can't upload external URLs to Sanity, so skip asset ref.
          // The GROQ query handles asset->url for images uploaded via Studio.
        };
      }
    }

    // Parse tags from JSONB
    let tags: string[] = [];
    if (row.tags) {
      tags = typeof row.tags === "string"
        ? JSON.parse(row.tags)
        : row.tags;
    }

    // Body — store as-is (markdown text)
    let body: string | undefined = undefined;
    if (row.body) {
      body = typeof row.body === "string" ? row.body : JSON.stringify(row.body);
    }

    const doc: any = {
      _type: "blogPost",
      _id: sanityId,
      title: row.title,
      slug: { _type: "slug", current: row.slug },
      description: row.description || "",
      body,
      tags: tags.length > 0 ? tags : undefined,
      publishedAt: row.published_at
        ? new Date(row.published_at).toISOString()
        : undefined,
      seoTitle: row.seo_title || undefined,
      seoDescription: row.seo_description || undefined,
    };

    if (authorRef) {
      doc.author = { _type: "reference", _ref: authorRef };
    }
    if (categoryRef) {
      doc.category = { _type: "reference", _ref: categoryRef };
    }

    await sanity.createOrReplace(doc);
    console.log(`  ✓ ${row.title} (${row.slug})`);
  }
}

async function main() {
  console.log(`Migrating blog content from Postgres → Sanity`);
  console.log(`  Project: ${SANITY_PROJECT_ID} / ${SANITY_DATASET}`);
  console.log(`  DB: ${process.env.DB_HOST || "localhost"}:${process.env.DB_PORT || 5432}/${process.env.DB_NAME || "functionfly"}`);

  try {
    const authorIdMap = await migrateAuthors();
    const categoryIdMap = await migrateCategories(authorIdMap);
    await migratePosts(authorIdMap, categoryIdMap);

    console.log(`\n✅ Migration complete!`);
    console.log(`   Review at https://sanity.io/manage/project/${SANITY_PROJECT_ID}`);
    console.log(`   Studio at http://localhost:4321/studio`);
  } catch (err) {
    console.error(`\n❌ Migration failed:`, err);
    process.exit(1);
  } finally {
    await db.end();
  }
}

main();
