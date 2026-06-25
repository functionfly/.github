/**
 * Shared Zod schema for Blog Posts
 *
 * Used by both the admin dashboard and the API client to validate
 * blog post data structures. Keeping these in sync prevents type mismatches.
 */

import { z } from 'zod';

export const BlogPostStatusSchema = z.enum(['draft', 'published', 'scheduled', 'archived']);
export type BlogPostStatus = z.infer<typeof BlogPostStatusSchema>;

export const HeroImageSchema = z.object({
  url: z.string().url().optional(),
  alt: z.string().optional(),
  width: z.number().optional(),
  height: z.number().optional(),
});
export type HeroImage = z.infer<typeof HeroImageSchema>;

export const OGImageSchema = z.object({
  url: z.string().url().optional(),
  alt: z.string().optional(),
});
export type OGImage = z.infer<typeof OGImageSchema>;

export const AuthorSummarySchema = z.object({
  name: z.string(),
  slug: z.string(),
});
export type AuthorSummary = z.infer<typeof AuthorSummarySchema>;

export const CategorySummarySchema = z.object({
  title: z.string(),
  slug: z.string(),
});
export type CategorySummary = z.infer<typeof CategorySummarySchema>;

/**
 * Blog post schema. Body is Markdown text, not JSON.
 */
export const BlogPostSchema = z.object({
  id: z.string().uuid(),
  title: z.string().min(1, 'Title is required'),
  slug: z.string().min(1, 'Slug is required'),
  description: z.string().default(''),
  body: z.string().default(''),
  bodyHtml: z.string().optional(),
  authorId: z.string().uuid().optional().nullable(),
  categoryId: z.string().uuid().optional().nullable(),
  tags: z.array(z.string()).default([]),
  heroImage: HeroImageSchema.optional().nullable(),
  status: BlogPostStatusSchema.default('draft'),
  publishedAt: z.string().datetime().optional().nullable(),
  scheduledAt: z.string().datetime().optional().nullable(),
  updatedAt: z.string().datetime().optional().nullable(),
  createdAt: z.string().datetime(),
  seoTitle: z.string().optional().nullable(),
  seoDescription: z.string().optional().nullable(),
  keywords: z.array(z.string()).default([]),
  canonicalUrl: z.string().url().optional().nullable(),
  ogImage: OGImageSchema.optional().nullable(),
  campaign: z.string().optional().nullable(),
  ownerId: z.string().uuid().optional().nullable(),
  author: AuthorSummarySchema.optional().nullable(),
  category: CategorySummarySchema.optional().nullable(),
  isPublished: z.boolean().default(false),
});
export type BlogPost = z.infer<typeof BlogPostSchema>;

/**
 * Input schema for creating a new blog post.
 * Auto-generates ID, slug, and timestamps.
 */
export const BlogPostInputSchema = BlogPostSchema.partial({
  id: true,
  slug: true,
  createdAt: true,
  updatedAt: true,
  publishedAt: true,
  scheduledAt: true,
  isPublished: true,
  bodyHtml: true,
}).refine(
  (data) => data.title && data.title.trim().length > 0,
  { message: 'Title is required', path: ['title'] }
);
export type BlogPostInput = z.infer<typeof BlogPostInputSchema>;

/**
 * List response from GET /v1/admin/blog/posts
 */
export const BlogListResponseSchema = z.object({
  data: z.array(BlogPostSchema),
  meta: z.object({
    total: z.number(),
    page: z.number(),
    limit: z.number(),
    totalPages: z.number(),
    search: z.string().optional(),
  }),
});
export type BlogListResponse = z.infer<typeof BlogListResponseSchema>;

/**
 * Category schema
 */
export const BlogCategorySchema = z.object({
  id: z.string().uuid(),
  title: z.string(),
  slug: z.string(),
  description: z.string().optional().default(''),
  color: z.string().optional().default(''),
  icon: z.string().optional().default(''),
  order: z.number().default(0),
  postCount: z.number().default(0),
  createdAt: z.string().datetime(),
  updatedAt: z.string().datetime(),
});
export type BlogCategory = z.infer<typeof BlogCategorySchema>;

/**
 * Author schema
 */
export const BlogAuthorSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  slug: z.string(),
  bio: z.string().optional().default(''),
  email: z.string().optional().default(''),
  website: z.string().optional().default(''),
  role: z.string().optional().default(''),
  active: z.boolean().default(true),
  createdAt: z.string().datetime(),
  updatedAt: z.string().datetime(),
});
export type BlogAuthor = z.infer<typeof BlogAuthorSchema>;
