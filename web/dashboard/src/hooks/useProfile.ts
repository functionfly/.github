import { useState } from 'react';
import { useAuthStore } from '../stores/authStore';
import { useRealtimeSubscription } from './useRealtimeSubscription.ts';
import { ProfileUpdateEvent } from './types';

// Hook for profile updates
export function useProfileUpdates() {
  const { user } = useAuthStore();
  const [profileUpdates, setProfileUpdates] = useState<ProfileUpdateEvent[]>([]);

  const { isConnected } = useRealtimeSubscription<ProfileUpdateEvent>(
    `user_${user?.id}_profile`,
    'profile_update',
    (event) => {
      setProfileUpdates(prev => [event, ...prev]);
    }
  );

  return {
    profileUpdates,
    isConnected,
  };
}