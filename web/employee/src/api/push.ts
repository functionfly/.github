import apiClient from './client';

export interface NotificationPreference {
  id: string;
  channel: string;
  event_type: string;
  is_enabled: boolean;
  quiet_hours_start?: string;
  quiet_hours_end?: string;
}

export const pushApi = {
  subscribe: (subscription: { endpoint: string; p256dh: string; auth: string }) =>
    apiClient.post('/v1/push/subscribe', subscription),
  getPreferences: () =>
    apiClient.get<{ preferences: NotificationPreference[] }>('/v1/notification-preferences'),
  updatePreference: (id: string, data: Partial<NotificationPreference>) =>
    apiClient.patch(`/v1/notification-preferences/${id}`, data),
};
