import { ContentStatus } from '@/api/blog';

// Old blog post format (from Go backend)
export interface LegacyBlogPost {
  id: string;
  title: string;
  slug: string;
  content: string;
  excerpt: string;
  author: string;
  tags: string[];
  featured_image?: string;
  sanity_id?: string;
  is_published: boolean;
  published_at?: string;
  created_at: string;
  updated_at: string;
}

// New blog post format (for NestJS API)
export interface NewBlogPost {
  title: string;
  slug: string;
  description: string;
  body: any;
  authorId?: string;
  categoryId?: string;
  tags?: string[];
  heroImage?: { url: string; alt: string; caption?: string };
  status: ContentStatus;
  publishedAt?: string;
  scheduledAt?: string;
  seoTitle?: string;
  seoDescription?: string;
  keywords?: string[];
  canonicalUrl?: string;
  ogImage?: { url: string; alt: string };
  campaign?: string;
}

/**
 * Migrates legacy blog post data to the new NestJS format
 */
export function migrateBlogPost(legacyPost: LegacyBlogPost): NewBlogPost {
  // Map status
  let status = ContentStatus.DRAFT;
  if (legacyPost.is_published) {
    status = ContentStatus.PUBLISHED;
  }

  // Map content to body (try to parse as JSON, fallback to string)
  let body: any = legacyPost.content;
  try {
    // If content looks like JSON, parse it
    if (typeof legacyPost.content === 'string' &&
        (legacyPost.content.startsWith('{') || legacyPost.content.startsWith('['))) {
      body = JSON.parse(legacyPost.content);
    }
  } catch (e) {
    // Keep as string if parsing fails
    body = legacyPost.content;
  }

  // Map featured image to hero image
  const heroImage = legacyPost.featured_image ? {
    url: legacyPost.featured_image,
    alt: `${legacyPost.title} featured image`,
  } : undefined;

  // Use excerpt as description
  const description = legacyPost.excerpt || legacyPost.content.substring(0, 500);

  // Basic SEO mapping
  const seoDescription = legacyPost.excerpt || description.substring(0, 160);

  return {
    title: legacyPost.title,
    slug: legacyPost.slug,
    description,
    body,
    tags: legacyPost.tags.length > 0 ? legacyPost.tags : undefined,
    heroImage,
    status,
    publishedAt: legacyPost.published_at,
    seoDescription,
    keywords: legacyPost.tags.length > 0 ? legacyPost.tags.slice(0, 10) : undefined,
    // Note: authorId, categoryId need to be mapped separately based on author names/categories
    // This would require looking up authors and categories by name/slug
  };
}

/**
 * Batch migrate multiple blog posts
 */
export function migrateBlogPosts(legacyPosts: LegacyBlogPost[]): NewBlogPost[] {
  return legacyPosts.map(migrateBlogPost);
}

/**
 * Creates a simple author lookup map (name -> id)
 * This would be populated from the actual authors API
 */
export function createAuthorLookup(authors: { id: string; name: string }[]): Map<string, string> {
  const lookup = new Map<string, string>();
  authors.forEach(author => {
    lookup.set(author.name.toLowerCase(), author.id);
  });
  return lookup;
}

/**
 * Creates a simple category lookup map (title -> id)
 * This would be populated from the actual categories API
 */
export function createCategoryLookup(categories: { id: string; title: string }[]): Map<string, string> {
  const lookup = new Map<string, string>();
  categories.forEach(category => {
    lookup.set(category.title.toLowerCase(), category.id);
  });
  return lookup;
}

/**
 * Applies author and category IDs to migrated posts
 */
export function applyAuthorCategoryIds(
  migratedPosts: NewBlogPost[],
  legacyPosts: LegacyBlogPost[],
  authorLookup: Map<string, string>,
  categoryLookup?: Map<string, string>
): NewBlogPost[] {
  return migratedPosts.map((post, index) => {
    const legacyPost = legacyPosts[index];
    const authorId = authorLookup.get(legacyPost.author.toLowerCase());

    return {
      ...post,
      authorId: authorId || post.authorId,
      // Note: category mapping would require additional logic
      // as legacy posts don't have category info
    };
  });
}