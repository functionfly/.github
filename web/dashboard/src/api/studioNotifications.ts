import { apiClient } from './client';

export type StudioNotificationType = 'info' | 'success' | 'warning' | 'error' | 'update';
export type StudioNotificationCategory = 'system' | 'workflow' | 'plugin' | 'billing' | 'community';

export interface StudioNotification {
  id: string;
  tenant_id: string;
  user_id?: string;
  environment: string;
  type: StudioNotificationType;
  category: StudioNotificationCategory;
  title: string;
  message: string;
  read: boolean;
  action_url?: string;
  created_at: string;
}

export interface StudioNotificationsResponse {
  notifications: StudioNotification[];
  unread_count: number;
}

export interface CreateStudioNotificationRequest {
  type?: StudioNotificationType;
  category?: StudioNotificationCategory;
  title: string;
  message: string;
  action_url?: string;
}

export const studioNotificationsApi = {
  list: (params?: { limit?: number; offset?: number; unread_only?: boolean }) => {
    const search = new URLSearchParams();
    if (params?.limit !== undefined) search.set('limit', String(params.limit));
    if (params?.offset !== undefined) search.set('offset', String(params.offset));
    if (params?.unread_only) search.set('unread_only', 'true');
    const qs = search.toString();
    return apiClient.get<StudioNotificationsResponse>(`/v1/studio/notifications${qs ? `?${qs}` : ''}`);
  },

  markRead: (id: string) =>
    apiClient.patch<StudioNotification>(`/v1/studio/notifications/${id}`, {}),

  markAllRead: () =>
    apiClient.post<{ message: string; count: number }>('/v1/studio/notifications/mark-all-read', {}),

  delete: (id: string) =>
    apiClient.delete<{ message: string }>(`/v1/studio/notifications/${id}`),

  clearAll: () =>
    apiClient.delete<{ message: string; count: number }>('/v1/studio/notifications'),

  create: (data: CreateStudioNotificationRequest) =>
    apiClient.post<StudioNotification>('/v1/studio/notifications', data),
};
