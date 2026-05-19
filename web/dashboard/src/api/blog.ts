import axios, { AxiosInstance } from 'axios';
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
  campaign?: string;
  ownerId?: string;
  createdAt: string;
  updatedAt: string;
  // Relations
  author?: {
    name: string;
    slug: string;
  };
  category?: {
    title: string;
    slug: string;
  };
}

export interface Author {
  id: string;
  name: string;
  slug: string;
  bio?: string;
  photo?: any;
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

export interface BlogPostQuery {
  page?: number;
  limit?: number;
  status?: ContentStatus;
  category?: string;
  author?: string;
  tag?: string;
  search?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  meta: {
    total: number;
    page: number;
    limit: number;
    totalPages: number;
    search?: string;
  };
}

const BLOG_ADMIN_BASE = '/v1/admin/blog';
const BLOG_PUBLIC_BASE = '/v1/blog';

class BlogApiClient {
  private client: AxiosInstance;

  constructor() {
    const mainApiUrl = import.meta.env.VITE_API_URL;
    if (!mainApiUrl) {
      throw new Error('VITE_API_URL environment variable is required');
    }
    this.client = axios.create({
      baseURL: mainApiUrl,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Add JWT token from localStorage
    this.client.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('ff-access-token');
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

    const response = await this.client.get(`${BLOG_PUBLIC_BASE}/posts?${params}`);
    return response.data;
  }

  async getPostBySlug(slug: string): Promise<BlogPost> {
    const response = await this.client.get(`${BLOG_PUBLIC_BASE}/posts/${slug}`);
    return response.data;
  }

  // ============ Blog Posts (Admin) ============

  async getPostById(id: string): Promise<BlogPost> {
    // Admin API uses the same endpoint with auth
    const response = await this.client.get(`${BLOG_ADMIN_BASE}/posts/${id}`);
    return response.data;
  }

  async createPost(post: Partial<BlogPost>): Promise<BlogPost> {
    const response = await this.client.post(`${BLOG_ADMIN_BASE}/posts`, post);
    return response.data;
  }

  async updatePost(id: string, updates: Partial<BlogPost>): Promise<BlogPost> {
    const response = await this.client.put(`${BLOG_ADMIN_BASE}/posts/${id}`, updates);
    return response.data;
  }

  async deletePost(id: string): Promise<void> {
    await this.client.delete(`${BLOG_ADMIN_BASE}/posts/${id}`);
  }

  // ============ Categories ============

  async getCategories(): Promise<Category[]> {
    // Public endpoint
    const response = await this.client.get(`${BLOG_PUBLIC_BASE}/categories`);
    return response.data;
  }

  async createCategory(
    category: Omit<Category, 'id' | 'createdAt' | 'updatedAt'>
  ): Promise<Category> {
    const response = await this.client.post(`${BLOG_ADMIN_BASE}/categories`, category);
    return response.data;
  }

  async updateCategory(id: string, updates: Partial<Category>): Promise<Category> {
    const response = await this.client.put(`${BLOG_ADMIN_BASE}/categories/${id}`, updates);
    return response.data;
  }

  async deleteCategory(id: string): Promise<void> {
    await this.client.delete(`${BLOG_ADMIN_BASE}/categories/${id}`);
  }

  // ============ Authors ============

  async getAuthors(): Promise<Author[]> {
    // Public endpoint
    const response = await this.client.get(`${BLOG_PUBLIC_BASE}/authors`);
    return response.data;
  }

  async createAuthor(author: Omit<Author, 'id' | 'createdAt' | 'updatedAt'>): Promise<Author> {
    const response = await this.client.post(`${BLOG_ADMIN_BASE}/authors`, author);
    return response.data;
  }

  async updateAuthor(id: string, updates: Partial<Author>): Promise<Author> {
    const response = await this.client.put(`${BLOG_ADMIN_BASE}/authors/${id}`, updates);
    return response.data;
  }

  async deleteAuthor(id: string): Promise<void> {
    await this.client.delete(`${BLOG_ADMIN_BASE}/authors/${id}`);
  }
}

export const blogApi = new BlogApiClient();
