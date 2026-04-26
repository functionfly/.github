import { useEffect, useState, useCallback, useRef } from 'react';
import { useAuthStore } from '../stores/authStore';
import { presenceApi, PresenceWebSocket, type UserPresence, type PresenceSocketEvent, type PresenceStatus } from '../api/presence';

export interface UsePresenceOptions {
  enableWebSocket?: boolean;
  heartbeatInterval?: number;
  reconnectAttempts?: number;
}

export interface UsePresenceReturn {
  onlineUsers: UserPresence[];
  myPresence: UserPresence | null;
  isConnected: boolean;
  isLoading: boolean;
  error: string | null;
  updatePresence: () => Promise<void>;
  getUserStatus: (userId: string) => PresenceStatus;
  formatLastSeen: (timestamp: string | null) => string;
}

const defaultOptions: UsePresenceOptions = {
  enableWebSocket: true,
  heartbeatInterval: 30000,
  reconnectAttempts: 5,
};

export function usePresence(options: UsePresenceOptions = defaultOptions): UsePresenceReturn {
  const opts = { ...defaultOptions, ...options };
  const { user } = useAuthStore();
  const [onlineUsers, setOnlineUsers] = useState<UserPresence[]>([]);
  const [myPresence, setMyPresence] = useState<UserPresence | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const wsRef = useRef<PresenceWebSocket | null>(null);
  const heartbeatRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const fetchPresence = useCallback(async () => {
    if (!user?.id) return;

    try {
      const [presenceRes, myRes] = await Promise.all([
        presenceApi.getPresence().catch(() => null),
        presenceApi.getMyPresence().catch(() => null),
      ]);

      if (presenceRes?.data?.users) {
        setOnlineUsers(presenceRes.data.users.filter(u => u.userId !== user.id));
      }

      if (myRes?.data) {
        setMyPresence({
          userId: myRes.data.userId,
          status: myRes.data.status,
          lastActive: myRes.data.lastActive,
          tenantId: myRes.data.tenantId,
          username: myRes.data.username,
          displayName: myRes.data.name,
        });
      }

      setError(null);
    } catch (err) {
      console.error('Failed to fetch presence:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch presence');
    } finally {
      setIsLoading(false);
    }
  }, [user?.id]);

  const updatePresence = useCallback(async () => {
    if (!user?.id) return;

    try {
      await presenceApi.updateMyPresence();
    } catch (err) {
      console.error('Failed to update presence:', err);
    }
  }, [user?.id]);

  const connectWebSocket = useCallback(() => {
    if (wsRef.current?.isConnected()) return;

    const ws = new PresenceWebSocket();
    wsRef.current = ws;

    ws.on('connected', () => {
      setIsConnected(true);
      setError(null);
    });

    ws.on('disconnected', () => {
      setIsConnected(false);
    });

    ws.on('reconnecting', () => {
      setIsConnected(false);
    });

    ws.on('failed', () => {
      setIsConnected(false);
      setError('Failed to connect to presence service');
    });

    ws.on('presence_join', (data: PresenceSocketEvent) => {
      if (data.type === 'presence_join') {
        setOnlineUsers(prev => {
          if (prev.some(u => u.userId === data.userId)) return prev;
          return [...prev, {
            userId: data.userId,
            status: data.status || 'online',
            lastActive: new Date().toISOString(),
          }];
        });
      }
    });

    ws.on('presence_leave', (data: PresenceSocketEvent) => {
      if (data.type === 'presence_leave') {
        setOnlineUsers(prev => prev.filter(u => u.userId !== data.userId));
      }
    });

    ws.on('presence_update', (data: PresenceSocketEvent) => {
      if (data.type === 'presence_update' && data.userId !== user?.id) {
        setOnlineUsers(prev => prev.map(u =>
          u.userId === data.userId
            ? { ...u, status: data.status, lastActive: new Date().toISOString() }
            : u
        ));
      }
    });

    ws.connect();

    heartbeatRef.current = setInterval(() => {
      ws.sendHeartbeat();
      void updatePresence();
    }, opts.heartbeatInterval);
  }, [user?.id, opts.heartbeatInterval, updatePresence]);

  const disconnectWebSocket = useCallback(() => {
    if (heartbeatRef.current) {
      clearInterval(heartbeatRef.current);
      heartbeatRef.current = null;
    }
    if (wsRef.current) {
      wsRef.current.disconnect();
      wsRef.current = null;
    }
    setIsConnected(false);
  }, []);

  useEffect(() => {
    void fetchPresence();

    if (opts.enableWebSocket && user?.id) {
      connectWebSocket();
    }

    return () => {
      disconnectWebSocket();
    };
  }, [user?.id, opts.enableWebSocket, fetchPresence, connectWebSocket, disconnectWebSocket]);

  const getUserStatus = useCallback((userId: string): PresenceStatus => {
    if (userId === user?.id) {
      return myPresence?.status || 'online';
    }
    const found = onlineUsers.find(u => u.userId === userId);
    return found?.status || 'offline';
  }, [user?.id, onlineUsers, myPresence]);

  const formatLastSeen = useCallback((timestamp: string | null): string => {
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
  }, []);

  return {
    onlineUsers,
    myPresence,
    isConnected,
    isLoading,
    error,
    updatePresence,
    getUserStatus,
    formatLastSeen,
  };
}

export function useUserOnlineStatus(userId: string) {
  const { getUserStatus, onlineUsers } = usePresence();

  const status = getUserStatus(userId);
  const isOnline = status === 'online';
  const isAway = status === 'away';

  return {
    isOnline,
    isAway,
    status,
    isOnlineUsersLoaded: onlineUsers.length > 0,
  };
}

export function useTenantOnlineUsers() {
  const { onlineUsers, isLoading, error, isConnected } = usePresence();

  return {
    onlineUsers,
    isLoading,
    error,
    isConnected,
    count: onlineUsers.length,
  };
}

export const useUserPresence = usePresence;
