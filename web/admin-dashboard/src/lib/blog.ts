/**
 * Blog utilities
 *
 * Centralized helpers for blog post data. Eliminates the need for
 * scattered conversions between API responses, form state, and storage.
 */

import type { BlogPost } from '@/types/blog';

/**
 * Extract author name from a post object.
 * Handles both API responses (with author.summary) and form state.
 */
export function getAuthorName(post: BlogPost | null | undefined): string {
  if (!post) return '';
  if (typeof post.author === 'string') return post.author;
  if (post.author && typeof post.author === 'object' && 'name' in post.author) {
    return post.author.name || '';
  }
  return '';
}

/**
 * Get tags as comma-separated string for form input.
 */
export function tagsToString(tags: string[] | undefined | null): string {
  if (!Array.isArray(tags)) return '';
  return tags.join(', ');
}

/**
 * Parse comma-separated tags string into array.
 */
export function stringToTags(str: string): string[] {
  return str
    .split(',')
    .map((t) => t.trim())
    .filter(Boolean);
}

/**
 * Get keywords as comma-separated string.
 */
export function keywordsToString(keywords: string[] | undefined | null): string {
  if (!Array.isArray(keywords)) return '';
  return keywords.join(', ');
}

export function stringToKeywords(str: string): string[] {
  return str
    .split(',')
    .map((t) => t.trim())
    .filter(Boolean);
}

/**
 * Generate URL-safe slug from title.
 * Collapses multiple consecutive spaces into a single hyphen.
 */
export function slugify(text: string): string {
  return text
    .toLowerCase()
    .trim()
    .replace(/\s+/g, '-')        // spaces -> hyphens (collapses multiple)
    .replace(/[^a-z0-9-]/g, '')  // strip other punctuation
    .replace(/-+/g, '-')         // collapse multiple hyphens
    .replace(/^-+|-+$/g, '');    // trim leading/trailing hyphens
}

/**
 * Format date for display in tables.
 * Returns "—" for null/invalid dates.
 */
export function formatPostDate(date: string | null | undefined): string {
  if (!date) return '—';
  const d = new Date(date);
  if (isNaN(d.getTime())) return '—';
  return d.toLocaleDateString();
}

/**
 * Convert a BlogPost from API to a safe form-state shape.
 * Use this to initialize form data.
 */
export function postToFormData(post: BlogPost): {
  title: string;
  slug: string;
  body: string;
  excerpt: string;
  author: string;
  tags: string;
  isPublished: boolean;
  featuredImage: string;
  seoTitle: string;
  seoDescription: string;
  keywords: string;
  canonicalUrl: string;
  ogImageUrl: string;
  ogImageAlt: string;
} {
  return {
    title: post.title ?? '',
    slug: post.slug ?? '',
    body: post.body ?? '',
    excerpt: post.description ?? '',
    author: getAuthorName(post),
    tags: tagsToString(post.tags),
    isPublished: post.isPublished ?? post.status === 'published',
    featuredImage: post.heroImage?.url ?? '',
    seoTitle: post.seoTitle ?? '',
    seoDescription: post.seoDescription ?? '',
    keywords: keywordsToString(post.keywords),
    canonicalUrl: post.canonicalUrl ?? '',
    ogImageUrl: post.ogImage?.url ?? '',
    ogImageAlt: post.ogImage?.alt ?? '',
  };
}

/**
 * Build the payload for POST/PUT /blog/posts from form data.
 */
export function formDataToPayload(data: ReturnType<typeof postToFormData>): Record<string, unknown> {
  const keywords = stringToKeywords(data.keywords);
  return {
    title: data.title,
    slug: data.slug || slugify(data.title),
    body: data.body,
    description: data.excerpt,
    excerpt: data.excerpt,
    author: data.author,
    tags: stringToTags(data.tags),
    isPublished: data.isPublished,
    status: data.isPublished ? 'published' : 'draft',
    featuredImage: data.featuredImage || undefined,
    heroImage: data.featuredImage ? { url: data.featuredImage } : undefined,
    seoTitle: data.seoTitle || undefined,
    seoDescription: data.seoDescription || undefined,
    keywords: keywords.length > 0 ? keywords : undefined,
    canonicalUrl: data.canonicalUrl || undefined,
    ogImage: data.ogImageUrl ? { url: data.ogImageUrl, alt: data.ogImageAlt } : undefined,
  };
}
