import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { contentApi, contentAdminApi, type ChangelogEntry, type BlogPost } from '@/api/content';

// Query keys
export const contentKeys = {
  all: ['content'] as const,
  changelog: (params?: { limit?: number }) => [...contentKeys.all, 'changelog', params] as const,
  blogPosts: (params?: { limit?: number; offset?: number; tags?: string[] }) =>
    [...contentKeys.all, 'blog-posts', params] as const,
  blogPost: (slug: string) => [...contentKeys.all, 'blog-post', slug] as const,
  blogCategories: () => [...contentKeys.all, 'blog-categories'] as const,
  // Admin keys
  adminChangelog: (params?: { limit?: number; offset?: number }) =>
    [...contentKeys.all, 'admin', 'changelog', params] as const,
  adminBlogPosts: (params?: { limit?: number; offset?: number }) =>
    [...contentKeys.all, 'admin', 'blog-posts', params] as const,
};

// Public content hooks

export function useChangelogEntries(limit = 10) {
  return useQuery({
    queryKey: contentKeys.changelog({ limit }),
    queryFn: () => contentApi.getPublishedChangelogEntries(limit),
    staleTime: 1000 * 60 * 5,
  });
}

export function useBlogPosts(params?: { limit?: number; offset?: number; tags?: string[] }) {
  return useQuery({
    queryKey: contentKeys.blogPosts(params),
    queryFn: () => contentApi.getPublishedBlogPosts(params),
    staleTime: 1000 * 60 * 5,
  });
}

export function useBlogPost(slug: string) {
  return useQuery({
    queryKey: contentKeys.blogPost(slug),
    queryFn: () => contentApi.getPublishedBlogPostBySlug(slug),
    enabled: !!slug,
    staleTime: 1000 * 60 * 5,
  });
}

export function useBlogCategories() {
  return useQuery({
    queryKey: contentKeys.blogCategories(),
    queryFn: () => contentApi.getPublishedCategories(),
    staleTime: 1000 * 60 * 60,
  });
}

// Admin content hooks

export function useAdminChangelogEntries(params?: { limit?: number; offset?: number }) {
  return useQuery({
    queryKey: contentKeys.adminChangelog(params),
    queryFn: () => contentAdminApi.listChangelogEntries(params),
    staleTime: 1000 * 60,
  });
}

export function useAdminBlogPosts(params?: { limit?: number; offset?: number; published_only?: boolean; tags?: string[] }) {
  return useQuery({
    queryKey: contentKeys.adminBlogPosts(params),
    queryFn: () => contentAdminApi.listBlogPosts(params),
    staleTime: 1000 * 60,
  });
}

// Admin mutations

export function useCreateChangelogEntry() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (entry: Omit<ChangelogEntry, 'id' | 'created_at' | 'updated_at'>) =>
      contentAdminApi.createChangelogEntry(entry),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: contentKeys.all });
      toast.success('Changelog entry created');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create changelog entry: ${error.message}`);
    },
  });
}

export function useUpdateChangelogEntry() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, updates }: { id: string; updates: Partial<ChangelogEntry> }) =>
      contentAdminApi.updateChangelogEntry(id, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: contentKeys.all });
      toast.success('Changelog entry updated');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update changelog entry: ${error.message}`);
    },
  });
}

export function useDeleteChangelogEntry() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => contentAdminApi.deleteChangelogEntry(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: contentKeys.all });
      toast.success('Changelog entry deleted');
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete changelog entry: ${error.message}`);
    },
  });
}

export function useCreateBlogPost() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (post: Omit<BlogPost, 'id' | 'created_at' | 'updated_at'>) =>
      contentAdminApi.createBlogPost(post),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: contentKeys.all });
      toast.success('Blog post created');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create blog post: ${error.message}`);
    },
  });
}

export function useUpdateBlogPost() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, updates }: { id: string; updates: Partial<BlogPost> }) =>
      contentAdminApi.updateBlogPost(id, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: contentKeys.all });
      toast.success('Blog post updated');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update blog post: ${error.message}`);
    },
  });
}

export function useDeleteBlogPost() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => contentAdminApi.deleteBlogPost(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: contentKeys.all });
      toast.success('Blog post deleted');
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete blog post: ${error.message}`);
    },
  });
}

export function useSyncGitHubReleases() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => contentAdminApi.syncGitHubReleases(),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: contentKeys.all });
      toast.success(data.message || 'GitHub releases synced');
    },
    onError: (error: Error) => {
      toast.error(`Failed to sync GitHub releases: ${error.message}`);
    },
  });
}

// AI content generation

export function useGenerateChangelogContent() {
  return useMutation({
    mutationFn: (params: { version?: string; type?: string; topic?: string }) =>
      contentAdminApi.generateChangelogContent(params),
    onError: (error: Error) => {
      toast.error(`Failed to generate content: ${error.message}`);
    },
  });
}

export function useGenerateBlogContent() {
  return useMutation({
    mutationFn: (params: { topic?: string; title?: string }) =>
      contentAdminApi.generateBlogContent(params),
    onError: (error: Error) => {
      toast.error(`Failed to generate content: ${error.message}`);
    },
  });
}

export function useSyncSanityPosts() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => contentAdminApi.syncSanityPosts(),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: contentKeys.all });
      toast.success(data.message || 'Sanity posts synced');
    },
    onError: (error: Error) => {
      toast.error(`Failed to sync Sanity posts: ${error.message}`);
    },
  });
}
