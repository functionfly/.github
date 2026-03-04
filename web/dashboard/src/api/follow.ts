import { apiClient } from "./client";

export interface FollowUserRequest {
  reason?: string;
  notify_on_new_function?: boolean;
  notify_on_function_update?: boolean;
  notify_on_new_version?: boolean;
}

export interface FollowFunctionRequest {
  reason?: string;
  notify_on_new_version?: boolean;
  notify_on_rating_change?: boolean;
  notify_on_trust_change?: boolean;
  notify_on_verification?: boolean;
}

export interface FollowResponse {
  ok: boolean;
  follow_id: string;
}

export interface UserFollowResponse {
  id: string;
  follower_id: string;
  follower_name?: string;
  follower_avatar?: string;
  followed_id?: string;
  followed_name?: string;
  followed_avatar?: string;
  reason?: string;
  created_at: string;
}

export interface FunctionFollowResponse {
  id: string;
  user_id: string;
  user_name?: string;
  function_id: string;
  function_name?: string;
  reason?: string;
  created_at: string;
}

export interface PaginatedUserFollowResponse {
  data: UserFollowResponse[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface PaginatedFunctionFollowResponse {
  data: FunctionFollowResponse[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface FollowStatusResponse {
  is_following: boolean;
  follower_count: number;
  following_count?: number;
}

export interface MyFollowStatsResponse {
  followers: number;
  following: number;
  functions_followed: number;
}

export const followApi = {
  // User follows
  followUser: (username: string, data?: FollowUserRequest) =>
    apiClient.post<FollowResponse>(`/v1/follow/users/${encodeURIComponent(username)}/follow`, data),

  unfollowUser: (username: string) =>
    apiClient.delete<{ ok: boolean }>(`/v1/follow/users/${encodeURIComponent(username)}/follow`),

  getUserFollowers: (username: string, page = 1, pageSize = 20) =>
    apiClient.get<PaginatedUserFollowResponse>(
      `/v1/follow/users/${encodeURIComponent(username)}/followers?page=${page}&page_size=${pageSize}`
    ),

  getUserFollowing: (username: string, page = 1, pageSize = 20) =>
    apiClient.get<PaginatedUserFollowResponse>(
      `/v1/follow/users/${encodeURIComponent(username)}/following?page=${page}&page_size=${pageSize}`
    ),

  getUserFollowStatus: (username: string) =>
    apiClient.get<FollowStatusResponse>(`/v1/follow/users/${encodeURIComponent(username)}/status`),

  // Function follows
  followFunction: (functionId: string, data?: FollowFunctionRequest) =>
    apiClient.post<FollowResponse>(`/v1/follow/functions/${functionId}/follow`, data),

  unfollowFunction: (functionId: string) =>
    apiClient.delete<{ ok: boolean }>(`/v1/follow/functions/${functionId}/follow`),

  getFunctionFollowers: (functionId: string, page = 1, pageSize = 20) =>
    apiClient.get<PaginatedFunctionFollowResponse>(
      `/v1/follow/functions/${functionId}/followers?page=${page}&page_size=${pageSize}`
    ),

  getFunctionFollowStatus: (functionId: string) =>
    apiClient.get<FollowStatusResponse>(`/v1/follow/functions/${functionId}/status`),

  // My follows
  getMyFollowedFunctions: (page = 1, pageSize = 20) =>
    apiClient.get<PaginatedFunctionFollowResponse>(
      `/v1/follow/me/functions?page=${page}&page_size=${pageSize}`
    ),

  getMyFollowStats: () =>
    apiClient.get<MyFollowStatsResponse>("/v1/follow/me/stats"),
};
