import { apiClient } from './client';

export interface CollabEvent {
  id: string;
  tenant_id: string;
  event_type: string;
  created_by: string;
  metadata: Record<string, any>;
  created_at: string;
  updated_at: string;
}

export interface CreateCollabEventRequest {
  event_type: string;
  metadata?: Record<string, any>;
}

export interface UpdateCollabEventRequest {
  metadata: Record<string, any>;
}

export interface CreateActivityRequest {
  action: string;
  target?: string;
  icon?: string;
  user_name?: string;
  user_color?: string;
  is_ai?: boolean;
}

export const studioCollabApi = {
  listEvents: (params?: { type?: string; limit?: number; offset?: number }) =>
    apiClient.get<{ events: CollabEvent[] }>(
      `/v1/studio/collab/events?${new URLSearchParams(
        Object.entries(params || {}).filter(([, v]) => v !== undefined).map(([k, v]) => [k, String(v)])
      ).toString()}`
    ),

  getEvent: (id: string) =>
    apiClient.get<CollabEvent>(`/v1/studio/collab/events/${id}`),

  createEvent: (data: CreateCollabEventRequest) =>
    apiClient.post<CollabEvent>('/v1/studio/collab/events', data),

  updateEvent: (id: string, data: UpdateCollabEventRequest) =>
    apiClient.patch<CollabEvent>(`/v1/studio/collab/events/${id}`, data),

  deleteEvent: (id: string) =>
    apiClient.delete<{ message: string }>(`/v1/studio/collab/events/${id}`),

  getActivityFeed: (limit?: number) =>
    apiClient.get<{ activities: CollabEvent[] }>(`/v1/studio/collab/activity?${limit ? `limit=${limit}` : ''}`),

  createActivity: (data: CreateActivityRequest) =>
    apiClient.post<CollabEvent>('/v1/studio/collab/activity', data),
};
