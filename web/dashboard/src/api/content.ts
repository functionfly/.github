import { apiClient } from './client';

// Types (Legacy - keeping for backward compatibility during migration)
export interface ChangelogEntry {
  id: string;
  version: string;
  date: string;
  type: 'major' | 'minor' | 'patch';
  title: string;
  description: string;
  changes: ChangelogChange[];
  release_url?: string;
  github_id?: string;
  is_published: boolean;
  created_at: string;
  updated_at: string;
}

export interface ChangelogChange {
  id: string;
  entry_id: string;
  category: string;
  icon: string;
  items: string[];
  created_at: string;
  updated_at: string;
}

// Legacy BlogPost interface - kept for backward compatibility
// New blog posts should use the types from blog.ts
export interface BlogPost {
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

// Frontend API functions (public)
export const contentApi = {
  // Get published changelog entries for frontend
  getPublishedChangelogEntries: async (limit = 10): Promise<ChangelogEntry[]> => {
    const response = await apiClient.get<{ entries: ChangelogEntry[] }>(`/v1/content/changelog?limit=${limit}`);
    return response.entries || [];
  },

  // Get published blog posts for frontend
  getPublishedBlogPosts: async (params?: {
    limit?: number;
    offset?: number;
    tags?: string[];
  }): Promise<{ posts: BlogPost[]; limit: number; offset: number }> => {
    const searchParams = new URLSearchParams();
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());
    if (params?.tags?.length) searchParams.set('tags', params.tags.join(','));

    const response = await apiClient.get<{ posts: BlogPost[]; limit: number; offset: number }>(`/v1/content/blog?${searchParams}`);
    // Ensure posts is always an array
    return {
      posts: response.posts || [],
      limit: response.limit || 10,
      offset: response.offset || 0,
    };
  },

  // Get published blog post by slug for frontend
  getPublishedBlogPostBySlug: async (slug: string): Promise<BlogPost> => {
    const response = await apiClient.get<BlogPost>(`/v1/content/blog/${slug}`);
    return response;
  },
};

// Admin API functions (protected)
// NOTE: These functions are being migrated to the new NestJS blog API.
// New blog management should use blogApi from './blog' instead.
export const contentAdminApi = {
  // Changelog management
  listChangelogEntries: async (params?: {
    limit?: number;
    offset?: number;
    published_only?: boolean;
  }): Promise<{ entries: ChangelogEntry[]; limit: number; offset: number }> => {
    const searchParams = new URLSearchParams();
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());
    if (params?.published_only !== undefined) searchParams.set('published_only', params.published_only.toString());

    const response = await apiClient.get<{ entries: ChangelogEntry[]; limit: number; offset: number }>(`/v1/admin/content/changelog?${searchParams}`);
    return response;
  },

  createChangelogEntry: async (entry: Omit<ChangelogEntry, 'id' | 'created_at' | 'updated_at'>): Promise<ChangelogEntry> => {
    const response = await apiClient.post<ChangelogEntry>('/v1/admin/content/changelog', entry);
    return response;
  },

  getChangelogEntry: async (id: string): Promise<ChangelogEntry> => {
    const response = await apiClient.get<ChangelogEntry>(`/v1/admin/content/changelog/${id}`);
    return response;
  },

  updateChangelogEntry: async (id: string, updates: Partial<ChangelogEntry>): Promise<ChangelogEntry> => {
    const response = await apiClient.patch<ChangelogEntry>(`/v1/admin/content/changelog/${id}`, updates);
    return response;
  },

  deleteChangelogEntry: async (id: string): Promise<void> => {
    await apiClient.delete(`/v1/admin/content/changelog/${id}`);
  },

  // Changelog changes management
  createChangelogChange: async (entryId: string, change: Omit<ChangelogChange, 'id' | 'entry_id' | 'created_at' | 'updated_at'>): Promise<ChangelogChange> => {
    const response = await apiClient.post<ChangelogChange>(`/v1/admin/content/changelog/${entryId}/changes`, change);
    return response;
  },

  updateChangelogChange: async (changeId: string, updates: Partial<ChangelogChange>): Promise<ChangelogChange> => {
    const response = await apiClient.patch<ChangelogChange>(`/v1/admin/content/changes/${changeId}`, updates);
    return response;
  },

  deleteChangelogChange: async (changeId: string): Promise<void> => {
    await apiClient.delete(`/v1/admin/content/changes/${changeId}`);
  },

  // Blog management
  listBlogPosts: async (params?: {
    limit?: number;
    offset?: number;
    published_only?: boolean;
    tags?: string[];
  }): Promise<{ posts: BlogPost[]; limit: number; offset: number }> => {
    const searchParams = new URLSearchParams();
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());
    if (params?.published_only !== undefined) searchParams.set('published_only', params.published_only.toString());
    if (params?.tags?.length) searchParams.set('tags', params.tags.join(','));

    const response = await apiClient.get<{ posts: BlogPost[]; limit: number; offset: number }>(`/v1/admin/content/blog?${searchParams}`);
    return response;
  },

  createBlogPost: async (post: Omit<BlogPost, 'id' | 'created_at' | 'updated_at'>): Promise<BlogPost> => {
    const response = await apiClient.post<BlogPost>('/v1/admin/content/blog', post);
    return response;
  },

  getBlogPost: async (id: string): Promise<BlogPost> => {
    const response = await apiClient.get<BlogPost>(`/v1/admin/content/blog/${id}`);
    return response;
  },

  updateBlogPost: async (id: string, updates: Partial<BlogPost>): Promise<BlogPost> => {
    const response = await apiClient.patch<BlogPost>(`/v1/admin/content/blog/${id}`, updates);
    return response;
  },

  deleteBlogPost: async (id: string): Promise<void> => {
    await apiClient.delete(`/v1/admin/content/blog/${id}`);
  },

  // Content sync
  syncGitHubReleases: async (): Promise<{ message: string }> => {
    const response = await apiClient.post<{ message: string }>('/v1/admin/content/sync/github-releases');
    return response;
  },

  syncSanityPosts: async (): Promise<{ message: string }> => {
    const response = await apiClient.post<{ message: string }>('/v1/admin/content/sync/sanity-posts');
    return response;
  },
};