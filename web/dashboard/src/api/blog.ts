import axios, { AxiosInstance } from "axios";

// Types based on NestJS DTOs
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
    id: string;
    name: string;
    slug: string;
    bio?: string;
    photo?: any;
    email?: string;
  };
  category?: {
    id: string;
    title: string;
    slug: string;
    description?: string;
    color?: string;
    icon?: string;
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
  socialLinks?: any;
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
  search?: string;
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

class BlogApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: import.meta.env.VITE_BLOG_API_URL || "http://localhost:3000",
      headers: {
        "Content-Type": "application/json",
      },
    });

    // Add JWT token from localStorage
    this.client.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem("sb-access-token");
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

  // ============ Blog Posts ============

  async getPosts(query?: BlogPostQuery): Promise<PaginatedResponse<BlogPost>> {
    const params = new URLSearchParams();
    if (query?.page) params.set('page', query.page.toString());
    if (query?.limit) params.set('limit', query.limit.toString());
    if (query?.status) params.set('status', query.status);
    if (query?.category) params.set('category', query.category);
    if (query?.author) params.set('author', query.author);
    if (query?.search) params.set('search', query.search);

    const response = await this.client.get(`/blog/posts?${params}`);
    return response.data;
  }

  async getPostBySlug(slug: string): Promise<BlogPost> {
    const response = await this.client.get(`/blog/posts/${slug}`);
    return response.data;
  }

  async getPostById(id: string): Promise<BlogPost> {
    const response = await this.client.get(`/blog/posts/${id}`);
    return response.data;
  }

  async createPost(post: Partial<BlogPost>): Promise<BlogPost> {
    const response = await this.client.post('/blog/posts', post);
    return response.data;
  }

  async updatePost(id: string, updates: Partial<BlogPost>): Promise<BlogPost> {
    const response = await this.client.put(`/blog/posts/${id}`, updates);
    return response.data;
  }

  async deletePost(id: string): Promise<void> {
    await this.client.delete(`/blog/posts/${id}`);
  }

  // ============ Categories ============

  async getCategories(): Promise<Category[]> {
    const response = await this.client.get('/blog/categories');
    return response.data;
  }

  async createCategory(category: Omit<Category, 'id' | 'createdAt' | 'updatedAt'>): Promise<Category> {
    const response = await this.client.post('/blog/categories', category);
    return response.data;
  }

  async updateCategory(id: string, updates: Partial<Category>): Promise<Category> {
    const response = await this.client.put(`/blog/categories/${id}`, updates);
    return response.data;
  }

  async deleteCategory(id: string): Promise<void> {
    await this.client.delete(`/blog/categories/${id}`);
  }

  // ============ Authors ============

  async getAuthors(): Promise<Author[]> {
    const response = await this.client.get('/blog/authors');
    return response.data;
  }

  async createAuthor(author: Omit<Author, 'id' | 'createdAt' | 'updatedAt'>): Promise<Author> {
    const response = await this.client.post('/blog/authors', author);
    return response.data;
  }

  async updateAuthor(id: string, updates: Partial<Author>): Promise<Author> {
    const response = await this.client.put(`/blog/authors/${id}`, updates);
    return response.data;
  }

  async deleteAuthor(id: string): Promise<void> {
    await this.client.delete(`/blog/authors/${id}`);
  }
}

export const blogApi = new BlogApiClient();