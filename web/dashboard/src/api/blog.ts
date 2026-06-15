import axios, { AxiosInstance } from 'axios';
import { tokenVault } from '@/utils/token-vault';
import { apiClient } from './client';

// Types based on Go backend API
export enum ContentStatus {
  DRAFT = 'draft',
  IN_REVIEW = 'in_review',
  APPROVED = 'approved',
  SCHEDULED = 'scheduled',
  PUBLISHED = 'published',
}

export interface BlogPost {
  id: string;
  title: string;
  slug: string;
  description: string;
  body: any; // Portable Text or structured content
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
  author?: {
    id: string;
    username: string;
    name?: string;
    avatar?: string;
  };
  category?: {
    id: string;
    name: string;
    slug: string;
  };
  viewCount?: number;
  createdAt: string;
  updatedAt: string;
}

export interface BlogCategory {
  id: string;
  name: string;
  slug: string;
  description?: string;
  postCount?: number;
  createdAt: string;
  updatedAt: string;
}

export interface BlogAuthor {
  id: string;
  username: string;
  name?: string;
  bio?: string;
  avatar?: string;
  postCount?: number;
  createdAt: string;
  updatedAt: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  limit: number;
  hasMore: boolean;
}

export interface BlogPostQuery {
  page?: number;
  limit?: number;
  status?: string;
  category?: string;
  author?: string;
  tag?: string;
  search?: string;
}

// Token helper - prefers tokenVault, falls back to localStorage for migration
async function getAuthToken(): Promise<string> {
  await tokenVault.initialize();
  const token = await tokenVault.getAccessToken();
  if (token) return token;
  // Fallback for migration period
  return localStorage.getItem('ff-access-token') || '';
}

export class BlogApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: apiClient.getBaseUrl(),
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Add JWT token from secure storage
    this.client.interceptors.request.use(
      async (config) => {
        const token = await getAuthToken();
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );
  }

  // ============ Blog Posts (Public) ============

  async getPosts(query?: BlogPostQuery): Promise<PaginatedResponse<BlogPost>> {
    const params = new URLSearchParams();
    if (query?.page) params.set('page', query.page.toString());
    if (query?.limit) params.set('limit', query.limit.toString());
    if (query?.status) params.set('status', query.status);
    if (query?.category) params.set('category', query.category);
    if (query?.author) params.set('author', query.author);
    if (query?.tag) params.set('tag', query.tag);
    if (query?.search) params.set('search', query.search);

    const response = await this.client.get<PaginatedResponse<BlogPost>>('/v1/blog/posts', {
      params,
    });
    return response.data;
  }

  async getPost(slugOrId: string): Promise<BlogPost> {
    const response = await this.client.get<BlogPost>(`/v1/blog/posts/${slugOrId}`);
    return response.data;
  }

  async getPostBySlug(slug: string): Promise<BlogPost> {
    const response = await this.client.get<BlogPost>(`/v1/blog/posts/slug/${slug}`);
    return response.data;
  }

  // ============ Blog Categories ============

  async getCategories(): Promise<BlogCategory[]> {
    const response = await this.client.get<BlogCategory[]>('/v1/blog/categories');
    return response.data;
  }

  async getCategory(slugOrId: string): Promise<BlogCategory> {
    const response = await this.client.get<BlogCategory>(`/v1/blog/categories/${slugOrId}`);
    return response.data;
  }

  // ============ Blog Authors ============

  async getAuthors(): Promise<BlogAuthor[]> {
    const response = await this.client.get<BlogAuthor[]>('/v1/blog/authors');
    return response.data;
  }

  async getAuthor(usernameOrId: string): Promise<BlogAuthor> {
    const response = await this.client.get<BlogAuthor>(`/v1/blog/authors/${usernameOrId}`);
    return response.data;
  }

  // ============ Blog Search ============

  async searchPosts(searchQuery: string): Promise<BlogPost[]> {
    const response = await this.client.get<BlogPost[]>('/v1/blog/search', {
      params: { q: searchQuery },
    });
    return response.data;
  }
}

export const blogApi = new BlogApiClient();
