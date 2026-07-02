import { apiClient } from './client';

export type CommunityPostStatus = 'open' | 'solved' | 'locked';
export type CommunitySort = 'hot' | 'new' | 'top';
export type VoteTargetType = 'post' | 'comment';

export interface CommunityAuthor {
  id: string;
  username?: string;
  name?: string;
  avatar_url?: string;
}

export interface CommunityCategory {
  id: string;
  slug: string;
  name: string;
  description: string;
  icon: string;
  sort_order: number;
  post_count?: number;
}

export interface CommunityPost {
  id: string;
  category_id: string;
  author_id: string;
  title: string;
  slug: string;
  body: string;
  status: CommunityPostStatus;
  vote_score: number;
  reply_count: number;
  view_count: number;
  tags: string[];
  is_pinned: boolean;
  accepted_comment_id?: string | null;
  created_at: string;
  updated_at: string;
  last_activity_at: string;
  category_slug?: string;
  category_name?: string;
  author?: CommunityAuthor;
  user_vote?: number | null;
  is_bookmarked?: boolean;
}

export interface CommunityComment {
  id: string;
  post_id: string;
  parent_id?: string | null;
  author_id: string;
  body: string;
  vote_score: number;
  is_accepted: boolean;
  deleted_at?: string | null;
  created_at: string;
  updated_at: string;
  author?: CommunityAuthor;
  user_vote?: number | null;
}

export interface CommunityNotification {
  id: string;
  user_id: string;
  actor_id: string;
  type: string;
  post_id?: string;
  comment_id?: string;
  is_read: boolean;
  created_at: string;
  actor: CommunityAuthor;
  post_slug?: string;
  post_title?: string;
}

export interface CommunityRule {
  id: string;
  title: string;
  description: string;
  category: 'conduct' | 'content' | 'safety' | 'legal' | 'moderation';
  enforcement: 'info' | 'warning' | 'deletion' | 'suspension';
  sort_order: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface ListPostsParams {
  category?: string;
  sort?: CommunitySort;
  q?: string;
  tag?: string;
  limit?: number;
  offset?: number;
}

export interface CreatePostRequest {
  category_slug: string;
  title: string;
  body: string;
  tags?: string[];
}

export interface UpdatePostRequest {
  title: string;
  body: string;
  tags?: string[];
}

export interface CreateCommentRequest {
  body: string;
  parent_id?: string;
}

export interface VoteRequest {
  target_type: VoteTargetType;
  target_id: string;
  value: 1 | -1;
}

export const communityApi = {
  listCategories: () =>
    apiClient.get<{ categories: CommunityCategory[] }>('/v1/community/categories'),

  listPosts: (params: ListPostsParams = {}) =>
    apiClient.get<{ posts: CommunityPost[]; total: number; limit: number; offset: number }>(
      '/v1/community/posts',
      { params }
    ),

  getPost: (id: string) =>
    apiClient.get<{ post: CommunityPost & { category: CommunityCategory }; comments: CommunityComment[] }>(
      `/v1/community/posts/${id}`
    ),

  createPost: (data: CreatePostRequest) =>
    apiClient.post<CommunityPost>('/v1/community/posts', data),

  updatePost: (id: string, data: UpdatePostRequest) =>
    apiClient.put<{ status: string }>(`/v1/community/posts/${id}`, data),

  deletePost: (id: string) =>
    apiClient.delete<{ status: string }>(`/v1/community/posts/${id}`),

  createComment: (postId: string, data: CreateCommentRequest) =>
    apiClient.post<CommunityComment>(`/v1/community/posts/${postId}/comments`, data),

  updateComment: (commentId: string, body: string) =>
    apiClient.put<{ status: string }>(`/v1/community/comments/${commentId}`, { body }),

  deleteComment: (commentId: string) =>
    apiClient.delete<{ status: string }>(`/v1/community/comments/${commentId}`),

  vote: (data: VoteRequest) =>
    apiClient.post<{ vote_score: number; user_vote: number }>('/v1/community/votes', data),

  acceptComment: (postId: string, commentId: string) =>
    apiClient.post<{ status: string }>(`/v1/community/posts/${postId}/accept/${commentId}`),

  bookmarkPost: (postId: string) =>
    apiClient.post<{ status: string }>(`/v1/community/posts/${postId}/bookmark`),

  unbookmarkPost: (postId: string) =>
    apiClient.delete<{ status: string }>(`/v1/community/posts/${postId}/bookmark`),

  listBookmarks: (limit = 20, offset = 0) =>
    apiClient.get<{ posts: CommunityPost[]; total: number }>(`/v1/community/bookmarks`, { params: { limit, offset } }),

  listNotifications: (limit = 20, offset = 0) =>
    apiClient.get<{ notifications: CommunityNotification[]; total: number }>(`/v1/community/notifications`, { params: { limit, offset } }),

  unreadNotificationsCount: () =>
    apiClient.get<{ count: number }>('/v1/community/notifications/unread-count'),

  markNotificationsRead: () =>
    apiClient.post<{ status: string }>('/v1/community/notifications/read'),

  listPostsByAuthor: (userId: string, limit = 20, offset = 0) =>
    apiClient.get<{ posts: CommunityPost[]; total: number }>(`/v1/community/users/${userId}/posts`, { params: { limit, offset } }),

  listRules: () =>
    apiClient.get<{ rules: CommunityRule[] }>('/v1/community/rules'),
};

export function displayAuthor(author?: CommunityAuthor): string {
  if (!author) return 'Anonymous';
  if (author.username) return `@${author.username}`;
  if (author.name) return author.name;
  return 'Member';
}

export function formatRelativeTime(iso: string): string {
  const date = new Date(iso);
  const diffMs = Date.now() - date.getTime();
  const mins = Math.floor(diffMs / 60000);
  if (mins < 1) return 'just now';
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 30) return `${days}d ago`;
  return date.toLocaleDateString();
}

export function profileUrl(author?: CommunityAuthor): string {
  if (author?.username) return `/u/${author.username}`;
  return '#';
}
