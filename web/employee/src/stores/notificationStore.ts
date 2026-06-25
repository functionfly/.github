import { create } from 'zustand';
import { notificationsApi, type Notification } from '@/api/notifications';

interface NotificationState {
  notifications: Notification[];
  unreadCount: number;
  isLoading: boolean;
  fetchNotifications: () => Promise<void>;
  fetchUnreadCount: () => Promise<void>;
  markRead: (id: string) => Promise<void>;
  markAllRead: () => Promise<void>;
}

export const useNotificationStore = create<NotificationState>((set, get) => ({
  notifications: [],
  unreadCount: 0,
  isLoading: false,

  fetchNotifications: async () => {
    set({ isLoading: true });
    try {
      const res = await notificationsApi.list({ limit: 20 });
      set({ notifications: res.data.notifications, isLoading: false });
    } catch {
      set({ isLoading: false });
    }
  },

  fetchUnreadCount: async () => {
    try {
      const res = await notificationsApi.unreadCount();
      set({ unreadCount: res.data.count });
    } catch {
      // ignore
    }
  },

  markRead: async (id: string) => {
    await notificationsApi.markRead(id);
    set({
      notifications: get().notifications.map((n) =>
        n.id === id ? { ...n, is_read: true, read_at: new Date().toISOString() } : n
      ),
      unreadCount: Math.max(0, get().unreadCount - 1),
    });
  },

  markAllRead: async () => {
    await notificationsApi.markAllRead();
    set({
      notifications: get().notifications.map((n) => ({
        ...n,
        is_read: true,
        read_at: new Date().toISOString(),
      })),
      unreadCount: 0,
    });
  },
}));
