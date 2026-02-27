import { useState } from 'react';
import { useAuthStore } from '../stores/authStore';
import { useRealtimeSubscription } from './useRealtimeSubscription.ts';
import { RealtimeEvent, NewNotificationEvent, ProfileUpdateEvent, UserStatusChangeEvent } from './types';

// Custom hook for real-time activity feed
export function useActivityFeed() {
  const { user } = useAuthStore();
  const [activities, setActivities] = useState<RealtimeEvent[]>([]);

  // Combine multiple real-time sources
  useRealtimeSubscription(
    `user_${user?.id}_notifications`,
    'new_notification',
    (event) => {
      setActivities(prev => [event, ...prev]);
    }
  );

  useRealtimeSubscription(
    `user_${user?.id}_profile`,
    'profile_update',
    (event) => {
      setActivities(prev => [event, ...prev]);
    }
  );

  useRealtimeSubscription(
    `tenant_${user?.tenantId}_users`,
    'user_status_change',
    (event) => {
      setActivities(prev => [event, ...prev]);
    }
  );

  return activities.sort((a, b) =>
    new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
  );
}