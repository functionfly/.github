import axios from 'axios';

// Types matching the NestJS API response
export interface BlogPost {
  id: string;
  title: string;
  slug: string;
  description: string;
  body: any;
  authorId?: string;
  categoryId?: string;
  tags?: string[];
  heroImage?: { url: string; alt: string; caption?: string };
  status: string;
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
    website?: string;
    socialLinks?: any;
    role?: string;
    active: boolean;
  };
  category?: {
    id: string;
    title: string;
    slug: string;
    description?: string;
    color?: string;
    icon?: string;
    order: number;
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

export interface PaginatedResponse<T> {
  data: T[];
  meta: {
    total: number;
    page: number;
    limit: number;
    totalPages: number;
  };
}

const blogApi = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080',
});

// Helper function to get published blog posts
export async function getBlogPosts(params?: {
  page?: number;
  limit?: number;
  category?: string;
  author?: string;
  search?: string;
}): Promise<PaginatedResponse<BlogPost>> {
  try {
    // Convert page to offset for backend compatibility
    const page = params?.page || 1;
    const limit = params?.limit || 10;
    const offset = (page - 1) * limit;

    // Build query params for backend
    const queryParams: any = {
      limit,
      offset,
    };

    // Add tag filter if category is provided (backend expects 'tags' parameter)
    if (params?.category) {
      queryParams.tags = params.category;
    }

    const response = await blogApi.get('/v1/content/blog', { params: queryParams });

    // Transform backend response to expected format
    const { posts, limit: responseLimit, offset: responseOffset } = response.data;
    const total = posts?.length || 0; // Backend doesn't provide total count, so we use current page count
    const totalPages = Math.ceil(total / limit);

    return {
      data: posts || [],
      meta: {
        total,
        page,
        limit: responseLimit || limit,
        totalPages,
      },
    };
  } catch (error) {
    console.error('Failed to fetch blog posts:', error);
    throw error;
  }
}

// Helper function to get a single blog post by slug
export async function getBlogPostBySlug(slug: string): Promise<BlogPost> {
  try {
    const response = await blogApi.get(`/v1/content/blog/${slug}`);
    return response.data;
  } catch (error) {
    console.error('Failed to fetch blog post:', error);
    throw error;
  }
}

// Helper function to get all categories
export async function getCategories(): Promise<Category[]> {
  try {
    const response = await blogApi.get('/v1/content/categories');
    // Transform the backend response to match frontend Category interface
    const { categories } = response.data;
    return categories.map((name: string, index: number) => ({
      id: `category-${index}`,
      title: name,
      slug: name.toLowerCase().replace(/\s+/g, '-'),
      order: index,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }));
  } catch (error) {
    console.error('Failed to fetch categories:', error);
    throw error;
  }
}

// Helper function to get all authors
export async function getAuthors(): Promise<Author[]> {
  try {
    const response = await blogApi.get('/v1/content/authors');
    // Transform the backend response to match frontend Author interface
    const { authors } = response.data;
    return authors.map((name: string, index: number) => ({
      id: `author-${index}`,
      name: name,
      slug: name.toLowerCase().replace(/\s+/g, '-'),
      bio: '', // Backend doesn't store bio
      avatar: '', // Backend doesn't store avatar
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }));
  } catch (error) {
    console.error('Failed to fetch authors:', error);
    throw error;
  }
}

// Helper function to get blog posts by category
export async function getBlogPostsByCategory(categorySlug: string, params?: {
  page?: number;
  limit?: number;
}): Promise<PaginatedResponse<BlogPost>> {
  return getBlogPosts({ ...params, category: categorySlug });
}

// Helper function to get blog posts by author
export async function getBlogPostsByAuthor(authorSlug: string, params?: {
  page?: number;
  limit?: number;
}): Promise<PaginatedResponse<BlogPost>> {
  return getBlogPosts({ ...params, author: authorSlug });
}

// Helper function to get blog posts by tag
export async function getBlogPostsByTag(tag: string, params?: {
  page?: number;
  limit?: number;
}): Promise<PaginatedResponse<BlogPost>> {
  // Note: The current API doesn't support tag filtering directly
  // This would need to be implemented on the backend or filtered client-side
  const allPosts = await getBlogPosts({ ...params, limit: 100 });
  const filteredPosts = allPosts.data.filter(post =>
    post.tags?.some(postTag => postTag.toLowerCase() === tag.toLowerCase())
  );

  return {
    data: filteredPosts,
    meta: {
      ...allPosts.meta,
      total: filteredPosts.length,
      totalPages: Math.ceil(filteredPosts.length / (params?.limit || 10))
    }
  };
}
