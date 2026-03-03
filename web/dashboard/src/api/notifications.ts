/**
 * Notifications API
 *
 * API module for notification management using Neon/Supabase.
 */

import { supabase } from '@/lib/neon';
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
    let query = supabase
      .from('user_notifications')
      .select(`
      id,
      status,
      read_at,
      archived_at,
      notification:notifications!inner(
        id,
        type,
        category,
        title,
        message,
        priority,
        metadata,
        action_url,
        icon,
        created_at
      )
    `)
    .order('created_at', { ascending: false, foreignTable: 'notifications' })
    .range(offset, offset + limit - 1);

  if (category && category !== 'all') {
    query = query.eq('notifications.category', category);
  }

  if (status) {
    if (status === 'unread') {
      query = query.is('read_at', null);
    } else if (status === 'read') {
      query = query.not('read_at', 'is', null);
    } else if (status === 'archived') {
      query = query.not('archived_at', 'is', null);
    }
  }

  if (priority) {
    query = query.eq('notifications.priority', priority);
  }

    const { data, error } = await query;

    if (error) {
      throw new Error(`Failed to fetch notifications: ${error.message}`);
    }

    // Transform the joined data into Notification objects
    return (data || []).map((item: any) => ({
    id: item.id,
    type: item.notification.type,
    category: item.notification.category,
    title: item.notification.title,
    message: item.notification.message,
    timestamp: item.notification.created_at,
    priority: item.notification.priority,
    status: item.status as NotificationStatus,
    metadata: item.notification.metadata || {},
    userId: item.user_id,
    tenantId: item.tenant_id,
    actionUrl: item.notification.action_url,
    icon: item.notification.icon,
    readAt: item.read_at,
    archivedAt: item.archived_at,
  }));
  } catch (err) {
    if (err instanceof TypeError && err.message?.includes('URL')) {
      console.warn(
        'Notifications: VITE_NEON_DATA_API_URL may be invalid or missing. Set it to a valid Data API URL (e.g. https://...neon.tech).'
      );
      return [];
    }
    throw err;
  }
}

/**
 * Fetch unread notification counts by category.
 * Returns zeros if Neon/Data API URL is misconfigured (e.g. Invalid URL) so the UI does not break.
 */
export async function fetchUnreadCounts(): Promise<NotificationCount> {
  try {
    const { data, error } = await supabase.rpc('get_notification_counts');

    if (error) {
      throw new Error(`Failed to fetch notification counts: ${error.message}`);
    }

    return {
      total: data?.total ?? 0,
      unread: data?.unread ?? 0,
      byCategory: data?.by_category ?? {},
    };
  } catch (err) {
    if (err instanceof TypeError && err.message?.includes('URL')) {
      console.warn(
        'Notifications: VITE_NEON_DATA_API_URL may be invalid or missing. Set it to a valid Data API URL (e.g. https://...neon.tech).'
      );
      return { total: 0, unread: 0, byCategory: {} };
    }
    throw err;
  }
}

/**
 * Mark a notification as read
 */
export async function markNotificationAsRead(notificationId: string): Promise<void> {
  const { error } = await supabase
    .from('user_notifications')
    .update({
      status: 'read' as NotificationStatus,
      read_at: new Date().toISOString(),
    })
    .eq('id', notificationId);

  if (error) {
    throw new Error(`Failed to mark notification as read: ${error.message}`);
  }
}

/**
 * Mark all notifications as read for the current user
 */
export async function markAllNotificationsAsRead(): Promise<number> {
  const { data, error } = await supabase.rpc('mark_all_notifications_read');

  if (error) {
    throw new Error(`Failed to mark all notifications as read: ${error.message}`);
  }

  return data?.count || 0;
}

/**
 * Archive a notification
 */
export async function archiveNotification(notificationId: string): Promise<void> {
  const { error } = await supabase
    .from('user_notifications')
    .update({
      status: 'archived' as NotificationStatus,
      archived_at: new Date().toISOString(),
    })
    .eq('id', notificationId);

  if (error) {
    throw new Error(`Failed to archive notification: ${error.message}`);
  }
}

/**
 * Delete a notification
 */
export async function deleteNotification(notificationId: string): Promise<void> {
  const { error } = await supabase
    .from('user_notifications')
    .delete()
    .eq('id', notificationId);

  if (error) {
    throw new Error(`Failed to delete notification: ${error.message}`);
  }
}

/**
 * Get notification preferences for the current user
 */
export async function getNotificationPreferences(): Promise<{
  emailEnabled: boolean;
  pushEnabled: boolean;
  categories: Record<string, boolean>;
}> {
  const { data, error } = await supabase
    .from('notification_preferences')
    .select('*')
    .single();

  if (error && error.code !== 'PGRST116') {
    // PGRST116 is "no rows returned" - return defaults
    throw new Error(`Failed to fetch preferences: ${error.message}`);
  }

  return {
    emailEnabled: data?.email_enabled ?? true,
    pushEnabled: data?.push_enabled ?? true,
    categories: data?.category_settings || {},
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
  const { error } = await supabase
    .from('notification_preferences')
    .upsert({
      email_enabled: preferences.emailEnabled,
      push_enabled: preferences.pushEnabled,
      category_settings: preferences.categories,
      updated_at: new Date().toISOString(),
    });

  if (error) {
    throw new Error(`Failed to update preferences: ${error.message}`);
  }
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
