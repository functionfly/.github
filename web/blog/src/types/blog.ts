/**
 * Blog TypeScript types and interfaces
 */

export interface BlogPost {
  id: string;
  title: string;
  slug: string;
  description: string;
  body: unknown;
  authorId?: string;
  categoryId?: string;
  tags?: string[];
  heroImage?: { url: string; alt?: string; caption?: string } | null;
  status: string;
  publishedAt?: string | null;
  scheduledAt?: string | null;
  seoTitle?: string | null;
  seoDescription?: string | null;
  keywords?: string[];
  canonicalUrl?: string | null;
  ogImage?: { url: string; alt?: string } | null;
  campaign?: string;
  ownerId?: string;
  createdAt: string;
  updatedAt: string;
  author?: {
    id: string;
    name: string;
    slug: string;
    bio?: string;
    photo?: unknown;
    email?: string;
    website?: string;
    socialLinks?: unknown;
    role?: string;
    active: boolean;
    createdAt?: string;
    updatedAt?: string;
  } | null;
  category?: {
    id: string;
    title: string;
    slug: string;
    description?: string;
    color?: string;
    icon?: string;
    order: number;
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
  socialLinks?: unknown;
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

export interface SEOProps {
  title: string;
  description?: string;
  canonical?: string;
  ogImage?: string;
  ogType?: 'article' | 'website';
  publishedAt?: string;
  modifiedAt?: string;
  author?: string;
  tags?: string[];
}
