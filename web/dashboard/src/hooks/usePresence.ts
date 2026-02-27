import { useEffect, useState } from 'react';
import { supabase } from '../lib/neon';
import { useAuthStore } from '../stores/authStore';
import { useRealtimeSubscription } from './useRealtimeSubscription.ts';
import { PresenceEvent } from './types';

// Hook for user presence/online status
export function useUserPresence() {
  const { user } = useAuthStore();
  const [onlineUsers, setOnlineUsers] = useState<string[]>([]);
  const [presenceEvents, setPresenceEvents] = useState<PresenceEvent[]>([]);

  const { isConnected } = useRealtimeSubscription<PresenceEvent>(
    `tenant_${user?.tenantId}_presence`,
    'presence_join',
    (event) => {
      if (event.type === 'presence_join') {
        setOnlineUsers(prev => [...new Set([...prev, ...event.new_presences?.map(p => p.user_id) || []])]);
      } else if (event.type === 'presence_leave') {
        setOnlineUsers(prev => prev.filter(id => !event.left_presences?.some(p => p.user_id === id)));
      }
      setPresenceEvents(prev => [event, ...prev]);
    }
  );

  // Track own presence via database update
  useEffect(() => {
    if (user?.id) {
      const updatePresence = async () => {
        try {
          const { error } = await supabase
            .from('user_profiles')
            .update({
              last_active_at: new Date().toISOString(),
              updated_at: new Date().toISOString()
            })
            .eq('user_id', user.id);

          if (error) {
            console.error('Error updating presence:', error);
          }
        } catch (error) {
          console.error('Error updating presence:', error);
        }
      };

      // Update presence immediately and then every 30 seconds
      updatePresence();
      const interval = setInterval(updatePresence, 30000);

      return () => {
        clearInterval(interval);
      };
    }
  }, [user?.id]);

  return {
    onlineUsers,
    presenceEvents,
    isConnected,
  };
}

// Hook to get user details for presence indicators
export function useUserPresenceDetails(userIds: string[]) {
  const [userDetails, setUserDetails] = useState<Array<{
    id: string;
    name: string;
    email: string;
    avatar_url?: string;
    last_active_at?: string;
  }>>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchUserDetails = async () => {
      if (userIds.length === 0) {
        setUserDetails([]);
        return;
      }

      setLoading(true);
      setError(null);

      try {
        const { getUsersByIds } = await import('../lib/neon');
        const details = await getUsersByIds(userIds);
        setUserDetails(details);
      } catch (err) {
        console.error('Error fetching user details:', err);
        setError(err instanceof Error ? err.message : 'Failed to fetch user details');
        // Provide fallback data
        const fallback = userIds.map(id => ({
          id,
          name: `User ${id.slice(0, 8)}`,
          email: `user${id.slice(0, 8)}@example.com`,
          avatar_url: undefined,
          last_active_at: undefined,
        }));
        setUserDetails(fallback);
      } finally {
        setLoading(false);
      }
    };

    fetchUserDetails();
  }, [userIds]);

  return {
    userDetails,
    loading,
    error,
  };
}

// Hook to get detailed user presence info including last active time
export function useUserPresenceInfo(userId: string) {
  const { isOnline, isRealtimeEnabled } = useUserOnlineStatus(userId);
  const [lastActiveAt, setLastActiveAt] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const fetchLastActive = async () => {
      if (!isOnline) {
        setLoading(true);
        try {
          const { getUserLastActive } = await import('../lib/neon');
          const lastActive = await getUserLastActive(userId);
          setLastActiveAt(lastActive);
        } catch (error) {
          console.error('Error fetching last active time:', error);
          setLastActiveAt(null);
        } finally {
          setLoading(false);
        }
      } else {
        setLastActiveAt(null);
      }
    };

    fetchLastActive();
  }, [userId, isOnline]);

  const formatLastSeen = (timestamp: string | null): string => {
    if (!timestamp) return 'Recently';

    const lastActive = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - lastActive.getTime();
    const diffMins = Math.floor(diffMs / (1000 * 60));
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return lastActive.toLocaleDateString();
  };

  return {
    isOnline,
    isRealtimeEnabled,
    lastActiveAt,
    lastSeen: formatLastSeen(lastActiveAt),
    loading,
  };
}

// Hook to check if a specific user is online
export function useUserOnlineStatus(userId: string) {
  const { onlineUsers, presenceEvents } = useUserPresence();

  const isOnline = onlineUsers.includes(userId);
  const isRealtimeEnabled = presenceEvents.length > 0 || onlineUsers.length > 0;

  return {
    isOnline,
    isRealtimeEnabled,
    lastSeen: isOnline ? null : 'offline'
  };
}
