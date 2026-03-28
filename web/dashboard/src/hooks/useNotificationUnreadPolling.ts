import { notificationsApi } from '@/api/notifications';
import { unreadPartialFromServerCount } from '@/lib/notification-unread-sync';
import { useAuthStore } from '@/stores/authStore';
import { useNotificationStore } from '@/stores/notificationStore';
import { useEffect } from 'react';

const DEFAULT_MS = 15_000;

/**
 * Periodically refetches unread counts so the nav bell / messages badge update
 * when new in-app notifications are created (websocket payloads often do not match our client guard).
 */
export function useNotificationUnreadPolling(intervalMs: number = DEFAULT_MS) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const userId = useAuthStore((s) => s.user?.id);
  const updateUnreadCounts = useNotificationStore((s) => s.updateUnreadCounts);

  useEffect(() => {
    if (!isAuthenticated || !userId) return;

    const tick = async () => {
      const token =
        typeof localStorage !== 'undefined' ? localStorage.getItem('ff-access-token') : null;
      if (!token?.trim()) return;
      try {
        const counts = await notificationsApi.fetchUnreadCounts();
        updateUnreadCounts(unreadPartialFromServerCount(counts));
      } catch {
        // ignore — transient network errors
      }
    };

    void tick();
    const id = window.setInterval(tick, intervalMs);
    return () => window.clearInterval(id);
  }, [isAuthenticated, userId, intervalMs, updateUnreadCounts]);
}
