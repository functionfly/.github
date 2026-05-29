import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  studioNotificationsApi,
  type StudioNotification,
} from '@/api/studioNotifications';

const NOTIFICATIONS_KEY = 'studio-notifications';

export function useStudioNotifications(params?: { limit?: number; unread_only?: boolean }) {
  return useQuery({
    queryKey: [NOTIFICATIONS_KEY, params],
    queryFn: () => studioNotificationsApi.list(params),
    staleTime: 1000 * 15,
    refetchInterval: 30000,
  });
}

export function useStudioNotificationUnreadCount() {
  const { data } = useStudioNotifications({ limit: 1 });
  return data?.unread_count ?? 0;
}

export function useMarkStudioNotificationRead() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => studioNotificationsApi.markRead(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [NOTIFICATIONS_KEY] });
    },
  });
}

export function useMarkAllStudioNotificationsRead() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => studioNotificationsApi.markAllRead(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [NOTIFICATIONS_KEY] });
    },
  });
}

export function useDeleteStudioNotification() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => studioNotificationsApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [NOTIFICATIONS_KEY] });
    },
  });
}

export function useClearStudioNotifications() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => studioNotificationsApi.clearAll(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [NOTIFICATIONS_KEY] });
    },
  });
}

export function mapStudioNotification(n: StudioNotification) {
  return {
    id: n.id,
    type: n.type,
    title: n.title,
    message: n.message,
    timestamp: new Date(n.created_at),
    read: n.read,
    category: n.category,
    actionUrl: n.action_url,
  };
}
