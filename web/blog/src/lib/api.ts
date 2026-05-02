/**
 * Blog API Client for public blog consumption
 * Uses the Go backend at /v1/blog/*
 */

const BLOG_API_URL = import.meta.env.PUBLIC_MAIN_API_URL || 'http://localhost:8080/api/v1';

// Types matching the Go backend API response
export interface BlogPost {
  id: string;
  title: string;
  slug: string;
  description: string;
  body: unknown;
  authorId?: string;
  categoryId?: string;
  tags?: string[];
  heroImage?: { url: string; alt: string; caption?: string } | null;
  status: string;
  publishedAt?: string | null;
  scheduledAt?: string | null;
  seoTitle?: string | null;
  seoDescription?: string | null;
  keywords?: string[];
  canonicalUrl?: string | null;
  ogImage?: { url: string; alt: string } | null;
  campaign?: string;
  ownerId?: string;
  createdAt: string;
  updatedAt: string;
  // Relations (from Go backend)
  author?: {
    name: string;
    slug: string;
  } | null;
  category?: {
    title: string;
    slug: string;
  } | null;
}

export interface Author {
  id: string;
  name: string;
  slug: string;
  bio?: string;
  photo?: unknown;
  email?: string;
  website?: string;
  socialLinks?: { platform: string; url: string }[];
  role?: string;
  active: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface Category {
  id: string;
  title: string;
  slug: string;
  description?: string;
  color?: string;
  icon?: string;
  order: number;
  createdAt: string;
  updatedAt: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  meta: {
    total: number;
    page: number;
    limit: number;
    totalPages: number;
  };
}

export interface PostsQuery {
  page?: number;
  limit?: number;
  category?: string;
  author?: string;
  tag?: string;
  search?: string;
}

/**
 * Fetch blog posts with pagination and filtering
 */
export async function getPosts(query?: PostsQuery): Promise<PaginatedResponse<BlogPost>> {
  const params = new URLSearchParams();
  
  if (query?.page) params.set('page', query.page.toString());
  if (query?.limit) params.set('limit', query.limit.toString());
  if (query?.category) params.set('category', query.category);
  if (query?.author) params.set('author', query.author);
  if (query?.tag) params.set('tag', query.tag);
  if (query?.search) params.set('search', query.search);

  const url = `${BLOG_API_URL}/blog/posts?${params.toString()}`;
  
  try {
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(`Failed to fetch posts: ${response.status}`);
    }
    return await response.json();
  } catch (error) {
    console.error('Failed to fetch blog posts:', error);
    return {
      data: [],
      meta: { total: 0, page: 1, limit: 10, totalPages: 0 }
    };
  }
}

/**
 * Fetch a single blog post by slug
 */
export async function getPostBySlug(slug: string): Promise<BlogPost | null> {
  try {
    const response = await fetch(`${BLOG_API_URL}/blog/posts/${encodeURIComponent(slug)}`);
    if (!response.ok) {
      if (response.status === 404) return null;
      throw new Error(`Failed to fetch post: ${response.status}`);
    }
    return await response.json();
  } catch (error) {
    console.error('Failed to fetch blog post:', error);
    return null;
  }
}

/**
 * Fetch all categories from Go backend
 */
export async function getCategories(): Promise<Category[]> {
  try {
    const response = await fetch(`${BLOG_API_URL}/blog/categories`);
    if (!response.ok) {
      throw new Error(`Failed to fetch categories: ${response.status}`);
    }
    const data = await response.json();
    return Array.isArray(data) ? data : [];
  } catch (error) {
    console.error('Failed to fetch categories:', error);
    return [];
  }
}

/**
 * Fetch all authors from Go backend
 */
export async function getAuthors(): Promise<Author[]> {
  try {
    const response = await fetch(`${BLOG_API_URL}/blog/authors`);
    if (!response.ok) {
      throw new Error(`Failed to fetch authors: ${response.status}`);
    }
    const data = await response.json();
    return Array.isArray(data) ? data : [];
  } catch (error) {
    console.error('Failed to fetch authors:', error);
    return [];
  }
}

/**
 * Fetch posts by category slug
 */
export async function getPostsByCategory(
  categorySlug: string, 
  params?: Omit<PostsQuery, 'category'>
): Promise<PaginatedResponse<BlogPost>> {
  return getPosts({ ...params, category: categorySlug });
}

/**
 * Fetch posts by author slug
 */
export async function getPostsByAuthor(
  authorSlug: string,
  params?: Omit<PostsQuery, 'author'>
): Promise<PaginatedResponse<BlogPost>> {
  return getPosts({ ...params, author: authorSlug });
}

/**
 * Search posts by query string
 */
export async function searchPosts(
  query: string,
  params?: Omit<PostsQuery, 'search'>
): Promise<PaginatedResponse<BlogPost>> {
  return getPosts({ ...params, search: query });
}

/**
 * Fetch related posts (posts in the same category, excluding current post)
 */
export async function getRelatedPosts(
  currentPostId: string,
  categorySlug?: string | null,
  limit: number = 3
): Promise<BlogPost[]> {
  try {
    // First try to get posts from the same category
    if (categorySlug) {
      const response = await getPostsByCategory(categorySlug, { limit: limit + 1 });
      const filtered = response.data.filter(p => p.id !== currentPostId);
      if (filtered.length >= limit) {
        return filtered.slice(0, limit);
      }
    }
    
    // Fallback to latest posts
    const response = await getPosts({ limit: limit + 5 });
    return response.data
      .filter(p => p.id !== currentPostId)
      .slice(0, limit);
  } catch (error) {
    console.error('Failed to fetch related posts:', error);
    return [];
  }
}

/**
 * Fetch all posts for static path generation (RSS, sitemap)
 */
export async function getAllPosts(limit: number = 1000): Promise<BlogPost[]> {
  try {
    const response = await fetch(`${BLOG_API_URL}/blog/posts?limit=${limit}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch all posts: ${response.status}`);
    }
    const data: PaginatedResponse<BlogPost> = await response.json();
    return data.data || [];
  } catch (error) {
    console.error('Failed to fetch all posts:', error);
    return [];
  }
}
