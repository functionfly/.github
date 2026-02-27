import { useState, useCallback } from 'react';
import { supabase } from '../lib/neon';
import { useAuthStore } from '../stores/authStore';
import { useRealtimeSubscription } from './useRealtimeSubscription.ts';
import { NewNotificationEvent } from './types';

// Hook for user-specific notifications
export function useUserNotifications() {
  const { user } = useAuthStore();
  const [notifications, setNotifications] = useState<NewNotificationEvent[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);

  const { isConnected } = useRealtimeSubscription<NewNotificationEvent>(
    `user_${user?.id}_notifications`,
    'new_notification',
    (event) => {
      setNotifications(prev => [event, ...prev]);
      setUnreadCount(prev => prev + 1);
    }
  );

  const markAsRead = useCallback(async (notificationId: string) => {
    try {
      const { error } = await supabase
        .from('user_notifications')
        .update({ read_at: new Date().toISOString() })
        .eq('id', notificationId)
        .eq('user_id', user?.id);

      if (error) throw error;

      setNotifications(prev =>
        prev.map(n =>
          n.notification_id === notificationId
            ? { ...n, read_at: new Date().toISOString() }
            : n
        )
      );
      setUnreadCount(prev => Math.max(0, prev - 1));
    } catch (error) {
      console.error('Error marking notification as read:', error);
    }
  }, [user?.id]);

  const markAllAsRead = useCallback(async () => {
    try {
      const { error } = await supabase
        .from('user_notifications')
        .update({ read_at: new Date().toISOString() })
        .eq('user_id', user?.id)
        .is('read_at', null);

      if (error) throw error;

      setNotifications(prev =>
        prev.map(n => ({ ...n, read_at: new Date().toISOString() }))
      );
      setUnreadCount(0);
    } catch (error) {
      console.error('Error marking all notifications as read:', error);
    }
  }, [user?.id]);

  return {
    notifications,
    unreadCount,
    isConnected,
    markAsRead,
    markAllAsRead,
  };
}