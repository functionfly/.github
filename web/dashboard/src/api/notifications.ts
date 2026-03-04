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

export interface FetchNotificationsParams {
  category?: NotificationCategory;
  status?: NotificationStatus;
  priority?: NotificationPriority;
  limit?: number;
  offset?: number;
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
): Promise<Notification[]> {
  const { category, status, priority, limit = 50, offset = 0 } = params;

  try {
    const queryParams = new URLSearchParams();
    if (limit) queryParams.set('limit', limit.toString());
    if (offset) queryParams.set('offset', offset.toString());
    if (category && category !== 'all') queryParams.set('category', category);
    if (status) queryParams.set('status', status);
    // Note: priority filtering might not be supported by backend yet

    const url = `/v1/notifications${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
    const response = await apiClient.get(url) as { notifications: any[] };

    // Transform the response into Notification objects
    return (response.notifications || []).map((item: any) => ({
      id: item.id,
      type: item.type,
      category: item.category,
      title: item.title,
      message: item.message,
      timestamp: item.created_at,
      priority: item.priority,
      status: item.status as NotificationStatus,
      metadata: item.metadata || {},
      userId: item.user_id,
      tenantId: item.tenant_id,
      actionUrl: item.action_url,
      icon: item.icon,
      readAt: item.read_at,
      archivedAt: item.archived_at,
    }));
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
  const response = await apiClient.post('/v1/notifications/read-all') as { count?: number };
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
  const response = await apiClient.get('/v1/users/me/notification-preferences') as { preferences?: any };

  return {
    emailEnabled: response.preferences?.email_enabled ?? true,
    pushEnabled: response.preferences?.push_enabled ?? true,
    categories: response.preferences?.category_settings || {},
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
    preferences: [{
      email_enabled: preferences.emailEnabled,
      push_enabled: preferences.pushEnabled,
      category_settings: preferences.categories,
    }]
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
