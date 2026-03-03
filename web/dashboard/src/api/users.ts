import { apiClient } from "./client";
import type { PublicUserProfile } from "@/types";

export interface UpdateProfileRequest {
  name?: string;
  username?: string;
  companyName?: string;
  bio?: string;
  avatar?: string; // Profile picture URL (or "" to clear)
  location?: string;
  website?: string;
  jobTitle?: string;
  socialLinks?: Record<string, string>;
  twitterUrl?: string;
  githubUrl?: string;
  linkedinUrl?: string;
}

export interface MeResponse {
  id: string;
  tenantId?: string;
  name: string;
  username?: string;
  companyName?: string;
  email: string;
  avatar?: string;
  plan?: string;
  updatedAt: string;
}

export interface UpdateProfileResponse {
  message: string;
  user: {
    id: string;
    name: string;
    username?: string;
    companyName?: string;
    email: string;
    avatar?: string;
    updatedAt: string;
  };
}

export interface PasswordChangeRequest {
  currentPassword: string;
  newPassword: string;
}

export interface SessionItem {
  id: string;
  device: string;
  ip: string;
  location: string;
  lastActive: string;
  currentSession: boolean;
}

// ============================================================================
// User Analytics Types
// ============================================================================

export interface ExecutionHistoryItem {
  date: string;
  executions: number;
  uniqueUsers?: number;
}

export interface PopularFunction {
  id: string;
  name: string;
  description: string;
  executionCount: number;
  rating: number;
  totalRatings: number;
}

export interface GeographicStat {
  region: string;
  executions: number;
}

export interface DeviceStat {
  device: string;
  executions: number;
}

export interface UserAnalyticsResponse {
  executionStats: {
    totalExecutions: number;
    totalUniqueUsers: number;
    functionCount: number;
    executionHistory: ExecutionHistoryItem[];
  };
  popularFunctions: PopularFunction[];
  geographicStats: {
    regions: GeographicStat[];
  };
  deviceStats: {
    devices: DeviceStat[];
  };
}

// ============================================================================
// User Achievements Types
// ============================================================================

export interface Achievement {
  id: string;
  slug: string;
  name: string;
  description: string;
  icon: string;
  color: string;
  category: string;
  points: number;
  earnedAt: string;
  progress: number;
  isCompleted: boolean;
  metadata: Record<string, unknown>;
}

export interface UserAchievementsResponse {
  achievements: Achievement[];
  totalPoints: number;
  available: number;
}

// ============================================================================
// User Activity Types
// ============================================================================

export interface UserActivityItem {
  id: string;
  type: string;
  title: string;
  description?: string;
  metadata: Record<string, unknown>;
  isPublic: boolean;
  createdAt: string;
}

export interface UserActivityResponse {
  activities: UserActivityItem[];
  limit: number;
  offset: number;
  total: number;
}

export interface CreateActivityRequest {
  activityType: string;
  title: string;
  description?: string;
  metadata?: Record<string, unknown>;
  isPublic?: boolean;
}

// ============================================================================
// User Skills Types
// ============================================================================

export interface UserSkill {
  id: string;
  name: string;
  level: "beginner" | "intermediate" | "advanced" | "expert";
  category?: string;
}

export interface UserSkillsResponse {
  skills: UserSkill[];
}

export interface AddSkillRequest {
  name: string;
  level?: "beginner" | "intermediate" | "advanced" | "expert";
  category?: string;
}

export const usersApi = {
  /**
   * Get the public profile for a user by username.
   * Returns only safe public fields — no email, tenantId, or role.
   */
  getPublicProfile: (username: string) =>
    apiClient.get<PublicUserProfile>(`/v1/users/${encodeURIComponent(username)}`),

  /**
   * Get the current authenticated user's full profile (includes plan from tenant).
   */
  getMe: () =>
    apiClient.get<MeResponse>("/v1/users/me"),

  /**
   * Update the current authenticated user's profile.
   */
  updateMe: (data: UpdateProfileRequest) =>
    apiClient.patch<UpdateProfileResponse>("/v1/users/me", data),

  /**
   * Change the current user's password.
   */
  changePassword: (data: PasswordChangeRequest) =>
    apiClient.post<{ message: string }>("/v1/users/me/change-password", data),

  /**
   * Request a password reset email.
   */
  requestPasswordReset: (email: string) =>
    apiClient.post<{ message: string }>("/v1/auth/password-reset", { email }),

  /**
   * Confirm a password reset with token and new password.
   */
  confirmPasswordReset: (token: string, newPassword: string) =>
    apiClient.post<{ message: string }>("/v1/auth/password-reset/confirm", {
      token,
      newPassword,
    }),

  /**
   * List active sessions for the current user.
   */
  listSessions: () =>
    apiClient.get<{ sessions: SessionItem[] }>("/v1/users/me/sessions"),

  /**
   * Revoke a single session by ID.
   */
  revokeSession: (sessionId: string) =>
    apiClient.delete<{ message: string }>(`/v1/users/me/sessions/${sessionId}`),

  /**
   * Revoke all other sessions (keep current).
   */
  revokeOtherSessions: () =>
    apiClient.post<{ message: string }>("/v1/users/me/sessions/revoke-others"),

  // ============================================================================
  // User Profile Settings
  // ============================================================================

  /**
   * Get current user profile settings.
   */
  getMySettings: () =>
    apiClient.get<{ settings: Record<string, unknown> }>(`/v1/users/me/settings`),

  /**
   * Get user profile settings by username.
   */
  getUserSettings: (username: string) =>
    apiClient.get<{ settings: Record<string, unknown> }>(`/v1/users/${encodeURIComponent(username)}/settings`),

  /**
   * Update current user notification settings.
   */
  updateMyNotificationSettings: (data: Record<string, boolean>) =>
    apiClient.patch<{ message: string }>(`/v1/users/me/settings/notifications`, data),

  /**
   * Update user notification settings by username.
   */
  updateNotificationSettings: (username: string, data: Record<string, boolean>) =>
    apiClient.patch<{ message: string }>(`/v1/users/${encodeURIComponent(username)}/settings/notifications`, data),

  /**
   * Update current user privacy settings.
   */
  updateMyPrivacySettings: (data: Record<string, boolean>) =>
    apiClient.patch<{ message: string }>(`/v1/users/me/settings/privacy`, data),

  /**
   * Update user privacy settings by username.
   */
  updatePrivacySettings: (username: string, data: Record<string, boolean>) =>
    apiClient.patch<{ message: string }>(`/v1/users/${encodeURIComponent(username)}/settings/privacy`, data),

  /**
   * Update current user visibility settings.
   */
  updateMyVisibilitySettings: (data: Record<string, unknown>) =>
    apiClient.patch<{ message: string }>(`/v1/users/me/settings/visibility`, data),

  /**
   * Update user visibility settings by username.
   */
  updateVisibilitySettings: (username: string, data: Record<string, unknown>) =>
    apiClient.patch<{ message: string }>(`/v1/users/${encodeURIComponent(username)}/settings/visibility`, data),

  // ============================================================================
  // User Profile Analytics
  // ============================================================================

  /**
   * Get analytics data for a user profile.
   * Includes execution history, popular functions, geographic stats, device stats.
   */
  getUserAnalytics: (username: string) =>
    apiClient.get<UserAnalyticsResponse>(`/v1/users/${encodeURIComponent(username)}/analytics`),

  // ============================================================================
  // User Achievements
  // ============================================================================

  /**
   * Get achievements for a user profile.
   */
  getUserAchievements: (username: string) =>
    apiClient.get<UserAchievementsResponse>(`/v1/users/${encodeURIComponent(username)}/achievements`),

  // ============================================================================
  // User Activity Feed
  // ============================================================================

  /**
   * Get activity feed for a user profile.
   */
  getUserActivity: (username: string, params?: { limit?: number; offset?: number }) =>
    apiClient.get<UserActivityResponse>(`/v1/users/${encodeURIComponent(username)}/activity`, { params }),

  /**
   * Create a new activity feed item (for authenticated user).
   */
  createActivity: (data: CreateActivityRequest) =>
    apiClient.post<{ id: string; message: string; createdAt: string }>("/v1/users/me/activity", data),

  // ============================================================================
  // User Skills
  // ============================================================================

  /**
   * Get skills for a user profile.
   */
  getUserSkills: (username: string) =>
    apiClient.get<UserSkillsResponse>(`/v1/users/${encodeURIComponent(username)}/skills`),

  /**
   * Add a skill for the authenticated user.
   */
  addSkill: (data: AddSkillRequest) =>
    apiClient.post<{ id: string; name: string; level: string; message: string }>("/v1/users/me/skills", data),

  /**
   * Remove a skill for the authenticated user.
   */
  removeSkill: (skillId: string) =>
    apiClient.delete<{ message: string }>(`/v1/users/me/skills/${skillId}`),
};
