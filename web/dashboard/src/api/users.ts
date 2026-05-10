import type { PublicProfileStats, PublicUserProfile } from '@/types';
import { apiClient } from './client';

// Helper function to get upgrade description based on plan
function getUpgradeDescription(plan: string): string {
  const descriptions: Record<string, string> = {
    enterprise: 'Unlimited functions, dedicated support, and premium enterprise features',
    professional: 'Advanced features, priority support, and increased limits',
    pro: 'Advanced features, priority support, and increased limits',
    starter: 'Expanded features and higher execution limits',
    free: 'Basic features and community support',
  };
  return descriptions[plan.toLowerCase()] || 'Membership upgraded with new features and benefits';
}

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
  dateOfBirth?: string | null; // YYYY-MM-DD, null to clear
  language?: string; // User's preferred language code (e.g., 'en', 'es')
}

export interface MeResponse {
  id: string;
  tenantId?: string;
  name: string;
  username?: string;
  companyName?: string;
  bio?: string;
  email: string;
  avatar?: string;
  dateOfBirth?: string | null;
  plan?: string;
  updatedAt: string;
  /** Account creation time (ISO 8601) */
  createdAt?: string;
  stats?: PublicProfileStats;
  // Online status fields from API
  isOnline?: boolean;
  lastActive?: string;
  // Profile number for early adopter tracking (e.g., Member #123)
  profileNumber?: number;
  // Platform admin role (super_admin, admin, support, etc.)
  role?: string;
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
    dateOfBirth?: string | null;
    updatedAt: string;
    // Platform admin role for badge display
    role?: string;
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

export interface UserContributionsResponse {
  days: { date: string; count: number; level: 0 | 1 | 2 | 3 | 4 }[];
  currentStreak: number;
  longestStreak: number;
  lastContributionDate: string | null;
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
  level: 'beginner' | 'intermediate' | 'advanced' | 'expert';
  category?: string;
}

export interface UserSkillsResponse {
  skills: UserSkill[];
}

export interface AddSkillRequest {
  name: string;
  level?: 'beginner' | 'intermediate' | 'advanced' | 'expert';
  category?: string;
}

export interface UserLookupEntry {
  id: string;
  username: string;
  name: string;
}

export interface ReportProfileRequest {
  reason: string;
  details: string;
  acknowledged_accuracy: boolean;
}

// ============================================================================
// Username Change Types
// ============================================================================

export interface UsernameChangeEligibility {
  canChangeFreely: boolean;
  canChangeWithFee: boolean;
  nextFreeChangeDate?: string;
  changesUsedThisYear: number;
  changesRemaining: number;
  earlyChangeFeeCents: number;
  message: string;
}

export interface UsernameChangeRequest {
  new_username: string;
  pay_early_fee?: boolean;
  stripe_payment_id?: string;
}

export interface UsernameChangeResponse {
  success: boolean;
  old_username?: string;
  new_username?: string;
  fee_paid_cents?: number;
  message: string;
}

export interface UsernameChangeHistoryItem {
  id: string;
  user_id: string;
  old_username: string;
  new_username: string;
  changed_at: string;
  was_early_change: boolean;
  fee_paid_cents: number;
  fee_currency: string;
}

export const usersApi = {
  /**
   * Get the public profile for a user by username.
   * Returns only safe public fields — no email, tenantId, or role.
   */
  getPublicProfile: (username: string) =>
    apiClient.get<PublicUserProfile>(`/v1/users/${encodeURIComponent(username)}`),

  /**
   * Report a user profile for TOS / community violations (auth required).
   */
  reportProfile: (username: string, body: ReportProfileRequest) =>
    apiClient.post<{ message: string }>(
      `/v1/users/${encodeURIComponent(username)}/report`,
      body
    ),

  /**
   * Get the current authenticated user's full profile (includes plan from tenant).
   */
  getMe: () => apiClient.get<MeResponse>('/v1/users/me'),

  /**
   * Update the current authenticated user's profile.
   */
  updateMe: (data: UpdateProfileRequest) =>
    apiClient.patch<UpdateProfileResponse>('/v1/users/me', data),

  /**
   * Change the current user's password.
   */
  changePassword: (data: PasswordChangeRequest) =>
    apiClient.post<{ message: string }>('/v1/users/me/change-password', data),

  /**
   * Request a password reset email.
   */
  requestPasswordReset: (email: string) =>
    apiClient.post<{ message: string }>('/v1/auth/password-reset', { email }),

  /**
   * Confirm a password reset with token and new password.
   */
  confirmPasswordReset: (token: string, newPassword: string) =>
    apiClient.post<{ message: string }>('/v1/auth/password-reset/confirm', {
      token,
      newPassword,
    }),

  /**
   * List active sessions for the current user.
   */
  listSessions: () => apiClient.get<{ sessions: SessionItem[] }>('/v1/users/me/sessions'),

  /**
   * Revoke a single session by ID.
   */
  revokeSession: (sessionId: string) =>
    apiClient.delete<{ message: string }>(`/v1/users/me/sessions/${sessionId}`),

  /**
   * Revoke all other sessions (keep current).
   */
  revokeOtherSessions: () =>
    apiClient.post<{ message: string }>('/v1/users/me/sessions/revoke-others'),

  /**
   * Delete current authenticated user account.
   */
  deleteMe: () => apiClient.delete<{ message: string }>('/v1/users/me'),

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
    apiClient.get<{ settings: Record<string, unknown> }>(
      `/v1/users/${encodeURIComponent(username)}/settings`
    ),

  /**
   * Update current user notification settings.
   */
  updateMyNotificationSettings: (data: Record<string, boolean>) =>
    apiClient.patch<{ message: string }>(`/v1/users/me/settings/notifications`, data),

  /**
   * Update user notification settings by username.
   */
  updateNotificationSettings: (username: string, data: Record<string, boolean>) =>
    apiClient.patch<{ message: string }>(
      `/v1/users/${encodeURIComponent(username)}/settings/notifications`,
      data
    ),

  /**
   * Update current user privacy settings.
   */
  updateMyPrivacySettings: (data: Record<string, boolean>) =>
    apiClient.patch<{ message: string }>(`/v1/users/me/settings/privacy`, data),

  /**
   * Update current user security settings (session timeout, remember devices).
   */
  updateMySecuritySettings: (data: { sessionTimeout?: string; rememberDevices?: boolean }) =>
    apiClient.patch<{ message: string }>(`/v1/users/me/settings/security`, data),

  /**
   * Update user privacy settings by username.
   */
  updatePrivacySettings: (username: string, data: Record<string, boolean>) =>
    apiClient.patch<{ message: string }>(
      `/v1/users/${encodeURIComponent(username)}/settings/privacy`,
      data
    ),

  /**
   * Update current user visibility settings.
   */
  updateMyVisibilitySettings: (data: Record<string, unknown>) =>
    apiClient.patch<{ message: string }>(`/v1/users/me/settings/visibility`, data),

  /**
   * Update user visibility settings by username.
   */
  updateVisibilitySettings: (username: string, data: Record<string, unknown>) =>
    apiClient.patch<{ message: string }>(
      `/v1/users/${encodeURIComponent(username)}/settings/visibility`,
      data
    ),

  /**
   * Update current user platform settings (DNA, runtime defaults, canary, etc).
   */
  updateMyPlatformSettings: (data: Record<string, unknown>) =>
    apiClient.patch<{ message: string }>(`/v1/users/me/settings/platform`, data),

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
    apiClient.get<UserAchievementsResponse>(
      `/v1/users/${encodeURIComponent(username)}/achievements`
    ),

  // ============================================================================
  // User Activity Feed
  // ============================================================================

  /**
   * Get activity feed for a user profile.
   */
  getUserActivity: (username: string, params?: { limit?: number; offset?: number }) =>
    apiClient.get<UserActivityResponse>(`/v1/users/${encodeURIComponent(username)}/activity`, {
      params,
    }),

  /**
   * Contribution heatmap + streaks (user_activity + registry publishes per UTC day).
   */
  getUserContributions: (username: string) =>
    apiClient.get<UserContributionsResponse>(
      `/v1/users/${encodeURIComponent(username)}/contributions`
    ),

  /**
   * Create a new activity feed item (for authenticated user).
   */
  createActivity: (data: CreateActivityRequest) =>
    apiClient.post<{ id: string; message: string; createdAt: string }>(
      '/v1/users/me/activity',
      data
    ),

  /**
   * Create a membership upgrade activity (for authenticated user).
   * Call this after a successful plan upgrade to show the upgrade in the activity feed.
   */
  createMembershipUpgradeActivity: (plan: string, previousPlan?: string) =>
    apiClient.post<{ id: string; message: string; createdAt: string }>('/v1/users/me/activity', {
      activityType: 'membership_upgraded',
      title: `Upgraded to ${plan.charAt(0).toUpperCase() + plan.slice(1)}`,
      description: getUpgradeDescription(plan),
      metadata: {
        plan,
        previousPlan: previousPlan || 'free',
        upgradedAt: new Date().toISOString(),
      },
      isPublic: true,
    }),

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
    apiClient.post<{ id: string; name: string; level: string; message: string }>(
      '/v1/users/me/skills',
      data
    ),

  /**
   * Remove a skill for the authenticated user.
   */
  removeSkill: (skillId: string) =>
    apiClient.delete<{ message: string }>(`/v1/users/me/skills/${skillId}`),

  /**
   * Resolve user UUIDs to id, username, and display name (auth required; for UI labels e.g. conversations).
   */
  lookupUsersByIds: (userIds: string[]) =>
    apiClient.post<{ users: UserLookupEntry[] }>('/v1/users/lookup-by-ids', {
      user_ids: userIds,
    }),

  /** Username prefix search for autocomplete (auth required). */
  searchUsersByUsername: (q: string, limit = 8) =>
    apiClient.get<{ users: UserLookupEntry[] }>(
      `/v1/users/search?q=${encodeURIComponent(q)}&limit=${limit}`
    ),

  // ============================================================================
  // Username Change (2-per-year limit with early-change fee)
  // ============================================================================

  /**
   * Check username change eligibility for current user.
   * Returns information about free changes remaining, fees, and next available date.
   */
  getUsernameChangeEligibility: () =>
    apiClient.get<UsernameChangeEligibility>('/v1/users/me/username/eligibility'),

  /**
   * Change username with optional early-change fee.
   * Users get 2 free changes per year; additional changes require waiting 6 months or paying a fee.
   */
  changeUsername: (data: UsernameChangeRequest) =>
    apiClient.post<UsernameChangeResponse>('/v1/users/me/username/change', data),

  /**
   * Get username change history for current user.
   */
  getUsernameChangeHistory: () =>
    apiClient.get<{ history: UsernameChangeHistoryItem[] }>('/v1/users/me/username/history'),

  /**
   * Create a Stripe checkout session for paid username changes.
   * Used when user has used their 2 free changes and wants to pay the early-change fee.
   */
  createUsernameChangeCheckout: async (data: {
    new_username: string;
    success_url?: string;
    cancel_url?: string;
  }) => {
    const csrfToken = await apiClient.fetchCSRFToken();
    const headers: Record<string, string> = {};
    if (csrfToken) {
      headers['X-CSRF-Token'] = csrfToken;
    }
    return apiClient.post<{ session_id: string; url: string; pending_id: string; message: string }>(
      '/v1/users/me/username/checkout',
      data,
      { headers }
    );
  },

  getCustomStatus: () =>
    apiClient.get<{ customStatus: string; customStatusEmoji: string }>('/v1/users/me/settings/status'),

  updateCustomStatus: (data: { customStatus: string; customStatusEmoji?: string }) => {
    return apiClient.patch<{ message: string; customStatus: string; customStatusEmoji: string }>(
      '/v1/users/me/settings/status',
      data
    );
  },
};
