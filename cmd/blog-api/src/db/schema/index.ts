import { pgTable, uuid, varchar, text, timestamp, boolean, integer, jsonb, pgEnum } from 'drizzle-orm/pg-core';

// Enums
export const contentStatusEnum = pgEnum('content_status', ['draft', 'in_review', 'approved', 'scheduled', 'published']);
export const contentTypeEnum = pgEnum('content_type', ['blog_post', 'doc', 'case_study', 'tool', 'benchmark']);

// Authors table
export const authors = pgTable('authors', {
  id: uuid('id').primaryKey().defaultRandom(),
  name: varchar('name', { length: 255 }).notNull(),
  slug: varchar('slug', { length: 96 }).notNull().unique(),
  bio: text('bio'),
  email: varchar('email', { length: 255 }),
  website: varchar('website', { length: 255 }),
  photo: jsonb('photo'),
  socialLinks: jsonb('social_links').$type<{ platform: string; url: string }[]>(),
  role: varchar('role', { length: 100 }),
  active: boolean('active').default(true),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
});

// Categories table
export const categories = pgTable('categories', {
  id: uuid('id').primaryKey().defaultRandom(),
  title: varchar('title', { length: 255 }).notNull(),
  slug: varchar('slug', { length: 96 }).notNull().unique(),
  description: text('description'),
  color: varchar('color', { length: 7 }), // hex color like #FF5733
  icon: varchar('icon', { length: 50 }),
  order: integer('order').default(0),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
});

// Blog Posts table
export const blogPosts = pgTable('blog_posts', {
  id: uuid('id').primaryKey().defaultRandom(),
  title: varchar('title', { length: 255 }).notNull(),
  slug: varchar('slug', { length: 96 }).notNull().unique(),
  description: text('description').notNull(),
  body: jsonb('body').notNull(), // Rich text content (similar to Portable Text)
  authorId: uuid('author_id').references(() => authors.id),
  categoryId: uuid('category_id').references(() => categories.id),
  tags: jsonb('tags').$type<string[]>(),
  heroImage: jsonb('hero_image').$type<{ url: string; alt: string; caption?: string }>(),
  status: contentStatusEnum('status').notNull().default('draft'),
  publishedAt: timestamp('published_at'),
  scheduledAt: timestamp('scheduled_at'),
  updatedAt: timestamp('updated_at'),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  // SEO fields
  seoTitle: varchar('seo_title', { length: 70 }),
  seoDescription: text('seo_description'),
  keywords: jsonb('keywords').$type<string[]>(),
  canonicalUrl: varchar('canonical_url', { length: 500 }),
  ogImage: jsonb('og_image').$type<{ url: string; alt: string }>(),
  // Editorial fields
  campaign: varchar('campaign', { length: 100 }),
  ownerId: uuid('owner_id').references(() => authors.id),
});

// Documentation table
export const docs = pgTable('docs', {
  id: uuid('id').primaryKey().defaultRandom(),
  title: varchar('title', { length: 255 }).notNull(),
  slug: varchar('slug', { length: 96 }).notNull().unique(),
  description: text('description').notNull(),
  body: jsonb('body').notNull(),
  categoryId: uuid('category_id').references(() => categories.id),
  order: integer('order').default(0),
  tags: jsonb('tags').$type<string[]>(),
  heroImage: jsonb('hero_image').$type<{ url: string; alt: string }>(),
  status: contentStatusEnum('status').notNull().default('draft'),
  publishedAt: timestamp('published_at'),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at'),
  // SEO fields
  seoTitle: varchar('seo_title', { length: 70 }),
  seoDescription: text('seo_description'),
  keywords: jsonb('keywords').$type<string[]>(),
  canonicalUrl: varchar('canonical_url', { length: 500 }),
  ogImage: jsonb('og_image').$type<{ url: string; alt: string }>(),
});

// Case Studies table
export const caseStudies = pgTable('case_studies', {
  id: uuid('id').primaryKey().defaultRandom(),
  title: varchar('title', { length: 255 }).notNull(),
  slug: varchar('slug', { length: 96 }).notNull().unique(),
  description: text('description').notNull(),
  body: jsonb('body').notNull(),
  company: varchar('company', { length: 255 }),
  industry: varchar('industry', { length: 100 }),
  challenge: text('challenge'),
  solution: text('solution'),
  results: jsonb('results').$type<{ metric: string; value: string }[]>(),
  heroImage: jsonb('hero_image').$type<{ url: string; alt: string }>(),
  logo: jsonb('logo').$type<{ url: string; alt: string }>(),
  featured: boolean('featured').default(false),
  status: contentStatusEnum('status').notNull().default('draft'),
  publishedAt: timestamp('published_at'),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at'),
  // SEO fields
  seoTitle: varchar('seo_title', { length: 70 }),
  seoDescription: text('seo_description'),
  keywords: jsonb('keywords').$type<string[]>(),
  canonicalUrl: varchar('canonical_url', { length: 500 }),
  ogImage: jsonb('og_image').$type<{ url: string; alt: string }>(),
});

// Tools table
export const tools = pgTable('tools', {
  id: uuid('id').primaryKey().defaultRandom(),
  title: varchar('title', { length: 255 }).notNull(),
  slug: varchar('slug', { length: 96 }).notNull().unique(),
  description: text('description').notNull(),
  body: jsonb('body').notNull(),
  type: varchar('type', { length: 50 }).notNull(), // template, tool, checklist, guide, calculator
  categoryId: uuid('category_id').references(() => categories.id),
  tags: jsonb('tags').$type<string[]>(),
  heroImage: jsonb('hero_image').$type<{ url: string; alt: string }>(),
  downloadUrl: varchar('download_url', { length: 500 }),
  demoUrl: varchar('demo_url', { length: 500 }),
  featured: boolean('featured').default(false),
  status: contentStatusEnum('status').notNull().default('draft'),
  publishedAt: timestamp('published_at'),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at'),
  // SEO fields
  seoTitle: varchar('seo_title', { length: 70 }),
  seoDescription: text('seo_description'),
  keywords: jsonb('keywords').$type<string[]>(),
  canonicalUrl: varchar('canonical_url', { length: 500 }),
  ogImage: jsonb('og_image').$type<{ url: string; alt: string }>(),
});

// Benchmarks table
export const benchmarks = pgTable('benchmarks', {
  id: uuid('id').primaryKey().defaultRandom(),
  title: varchar('title', { length: 255 }).notNull(),
  slug: varchar('slug', { length: 96 }).notNull().unique(),
  description: text('description').notNull(),
  body: jsonb('body').notNull(),
  type: varchar('type', { length: 50 }).notNull(), // performance, load, comparison
  period: varchar('period', { length: 50 }), // Q1 2026, monthly, etc.
  methodology: text('methodology'),
  keyFindings: jsonb('key_findings').$type<string[]>(),
  data: jsonb('data'), // Raw benchmark data
  heroImage: jsonb('hero_image').$type<{ url: string; alt: string }>(),
  featured: boolean('featured').default(false),
  status: contentStatusEnum('status').notNull().default('draft'),
  publishedAt: timestamp('published_at'),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at'),
  // SEO fields
  seoTitle: varchar('seo_title', { length: 70 }),
  seoDescription: text('seo_description'),
  keywords: jsonb('keywords').$type<string[]>(),
  canonicalUrl: varchar('canonical_url', { length: 500 }),
  ogImage: jsonb('og_image').$type<{ url: string; alt: string }>(),
});

// Related posts junction table
export const relatedPosts = pgTable('related_posts', {
  id: uuid('id').primaryKey().defaultRandom(),
  postId: uuid('post_id').references(() => blogPosts.id, { onDelete: 'cascade' }).notNull(),
  relatedPostId: uuid('related_post_id').references(() => blogPosts.id, { onDelete: 'cascade' }).notNull(),
});

// CTA Blocks junction table
export const ctaBlocks = pgTable('cta_blocks', {
  id: uuid('id').primaryKey().defaultRandom(),
  postId: uuid('post_id').references(() => blogPosts.id, { onDelete: 'cascade' }).notNull(),
  title: varchar('title', { length: 255 }).notNull(),
  description: text('description'),
  buttonText: varchar('button_text', { length: 100 }).notNull(),
  buttonUrl: varchar('button_url', { length: 500 }).notNull(),
  style: varchar('style', { length: 20 }).default('primary'), // primary, secondary, outline
  order: integer('order').default(0),
});

// Type exports for TypeScript
export type Author = typeof authors.$inferSelect;
export type NewAuthor = typeof authors.$inferInsert;
export type Category = typeof categories.$inferSelect;
export type NewCategory = typeof categories.$inferInsert;
export type BlogPost = typeof blogPosts.$inferSelect;
export type NewBlogPost = typeof blogPosts.$inferInsert;
export type Doc = typeof docs.$inferSelect;
export type NewDoc = typeof docs.$inferInsert;
export type CaseStudy = typeof caseStudies.$inferSelect;
export type NewCaseStudy = typeof caseStudies.$inferInsert;
export type Tool = typeof tools.$inferSelect;
export type NewTool = typeof tools.$inferInsert;
export type Benchmark = typeof benchmarks.$inferSelect;
export type NewBenchmark = typeof benchmarks.$inferInsert;
