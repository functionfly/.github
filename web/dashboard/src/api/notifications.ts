/**
 * Notifications API
 *
 * API module for notification management.
 */

import { apiClient } from '@/api/client';
import type {
  Notification,
  NotificationCategory,
  NotificationPriority,
  NotificationStatus,
} from '@/types/notifications';

/** API uses delivery statuses (pending, sent, …); UI uses inbox semantics (unread | read | archived). */
export function normalizeNotificationStatus(apiStatus: string): NotificationStatus {
  if (apiStatus === 'read') return 'read';
  if (apiStatus === 'archived') return 'archived';
  return 'unread';
}

/** Map orchestrator categories to dashboard tab ids (matches unread-count handler). */
export function normalizeNotificationCategory(apiCategory: string): NotificationCategory {
  switch (apiCategory) {
    case 'team':
    case 'messages':
      return 'messages';
    case 'billing':
      return 'revenue';
    case 'deployment':
      return 'issues';
    case 'system':
      return 'trust';
    case 'function':
    case 'registry':
      return 'trust';
    case 'trust':
    case 'revenue':
    case 'issues':
    case 'security':
    case 'all':
      return apiCategory;
    default:
      return 'all';
  }
}

export function isNotificationUnreadStatus(status: NotificationStatus): boolean {
  return status !== 'read' && status !== 'archived';
}

function normalizePriority(p: string): NotificationPriority {
  if (p === 'normal') return 'medium';
  if (p === 'low' || p === 'medium' || p === 'high' || p === 'critical') {
    return p;
  }
  return 'medium';
}

export interface FetchNotificationsParams {
  category?: NotificationCategory;
  status?: NotificationStatus;
  priority?: NotificationPriority;
  limit?: number;
  /** Cursor for cursor-based pagination (preferred over offset for large datasets) */
  cursor?: string;
  /** Offset-based pagination (deprecated - use cursor for large datasets) */
  offset?: number;
}

export interface PaginatedNotifications {
  notifications: Notification[];
  /** Cursor for next page (present if more data available) */
  nextCursor?: string;
  /** Whether more notifications exist after this page */
  hasMore: boolean;
  /** Total count (may be approximate with cursor pagination) */
  total?: number;
}

export interface NotificationCount {
  total: number;
  unread: number;
  byCategory: Record<string, number>;
}

/**
 * Fetch notifications for the current user
 */
export async function fetchNotifications(
  params: FetchNotificationsParams = {}
): Promise<PaginatedNotifications> {
  const { category, status, priority, limit = 50, cursor, offset } = params;

  try {
    const queryParams = new URLSearchParams();
    if (limit) queryParams.set('limit', limit.toString());
    if (cursor) {
      queryParams.set('cursor', cursor);
    } else if (offset !== undefined) {
      queryParams.set('offset', offset.toString());
    }
    if (category && category !== 'all') queryParams.set('category', category);
    if (status) queryParams.set('status', status);
    // Note: priority filtering might not be supported by backend yet

    const url = `/v1/notifications${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
    const response = (await apiClient.get(url)) as {
      notifications: any[];
      next_cursor?: string;
      has_more?: boolean;
      total?: number;
    };

    // Transform the response into Notification objects
    const notifications = (response.notifications || []).map((item: any) => ({
      id: item.id,
      type: item.type,
      category: normalizeNotificationCategory(String(item.category ?? 'all')),
      title: item.title,
      message: item.body ?? item.message ?? '',
      timestamp: item.created_at,
      priority: normalizePriority(String(item.priority ?? 'medium')),
      status: normalizeNotificationStatus(String(item.status ?? 'pending')),
      metadata: item.metadata || item.data || {},
      userId: item.user_id,
      tenantId: item.tenant_id,
      actionUrl: item.action_url ?? item.data?.action_url,
      icon: item.icon,
      readAt: item.read_at,
      archivedAt: item.archived_at,
    }));

    return {
      notifications,
      nextCursor: response.next_cursor,
      hasMore: response.has_more ?? (response.next_cursor !== undefined),
      total: response.total,
    };
  } catch (error: any) {
    console.error('Failed to fetch notifications:', error);
    throw new Error(`Failed to fetch notifications: ${error?.message || 'Unknown error'}`);
  }
}

/**
 * Fetch unread notification counts by category.
 */
export async function fetchUnreadCounts(): Promise<NotificationCount> {
  return await apiClient.get('/v1/notifications/unread-count');
}

/**
 * Mark a notification as read
 */
export async function markNotificationAsRead(notificationId: string): Promise<void> {
  await apiClient.patch(`/v1/notifications/${notificationId}/read`);
}

/**
 * Mark all notifications as read for the current user
 */
export async function markAllNotificationsAsRead(): Promise<number> {
  const response = (await apiClient.post('/v1/notifications/read-all')) as { count?: number };
  return response.count || 0;
}

/**
 * Archive a notification
 */
export async function archiveNotification(notificationId: string): Promise<void> {
  // Note: Archive functionality may need backend support
  await apiClient.patch(`/v1/notifications/${notificationId}`, {
    status: 'archived',
    archived_at: new Date().toISOString(),
  });
}

/**
 * Delete a notification
 */
export async function deleteNotification(notificationId: string): Promise<void> {
  await apiClient.delete(`/v1/notifications/${notificationId}`);
}

/**
 * Get notification preferences for the current user
 */
export async function getNotificationPreferences(): Promise<{
  emailEnabled: boolean;
  pushEnabled: boolean;
  categories: Record<string, boolean>;
}> {
  const response = (await apiClient.get('/v1/users/me/notification-preferences')) as {
    preferences?: any;
    email_enabled?: boolean;
    push_enabled?: boolean;
    category_settings?: Record<string, boolean>;
  };

  // Handle different response formats: object, array, or flat fields
  let prefs = response.preferences;
  if (Array.isArray(prefs) && prefs.length > 0) {
    prefs = prefs[0];
  }
  if (!prefs || typeof prefs !== 'object') {
    // Fallback to flat fields if preferences object is missing
    prefs = {
      email_enabled: response.email_enabled,
      push_enabled: response.push_enabled,
      category_settings: response.category_settings,
    };
  }

  return {
    emailEnabled: (prefs as any)?.email_enabled ?? true,
    pushEnabled: (prefs as any)?.push_enabled ?? true,
    categories: (prefs as any)?.category_settings || {},
  };
}

/**
 * Update notification preferences
 */
export async function updateNotificationPreferences(preferences: {
  emailEnabled?: boolean;
  pushEnabled?: boolean;
  categories?: Record<string, boolean>;
}): Promise<void> {
  await apiClient.patch('/v1/users/me/notification-preferences', {
    preferences: [
      {
        email_enabled: preferences.emailEnabled,
        push_enabled: preferences.pushEnabled,
        category_settings: preferences.categories,
      },
    ],
  });
}

// Export API object for convenience
export const notificationsApi = {
  fetchNotifications,
  fetchUnreadCounts,
  markNotificationAsRead,
  markAllNotificationsAsRead,
  archiveNotification,
  deleteNotification,
  getNotificationPreferences,
  updateNotificationPreferences,
};

export default notificationsApi;
