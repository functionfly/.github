import apiClient from './client';

export interface Notification {
  id: string;
  user_id: string;
  type: string;
  title: string;
  body?: string;
  action_url?: string;
  is_read: boolean;
  read_at?: string;
  created_at: string;
}

export const notificationsApi = {
  list: (params?: { unread_only?: boolean; limit?: number; offset?: number }) =>
    apiClient.get<{ notifications: Notification[]; total: number }>('/v1/notifications', { params }),

  unreadCount: () =>
    apiClient.get<{ count: number }>('/v1/notifications/unread-count'),

  markRead: (id: string) =>
    apiClient.post(`/v1/notifications/${id}/read`),

  markAllRead: () =>
    apiClient.post('/v1/notifications/read-all'),
};
