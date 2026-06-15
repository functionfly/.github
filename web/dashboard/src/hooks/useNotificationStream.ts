import { normalizeNotificationCategory } from '@/api/notifications';
import { API_BASE_URL } from '@/lib/constants';
import { useAuthStore } from '@/stores/authStore';
import type { Notification, NotificationPriority } from '@/types/notifications';
import { useCallback, useEffect, useRef, useState } from 'react';

interface StreamPayload {
  id?: string;
  type?: string;
  category?: string;
  title?: string;
  body?: string;
  priority?: string;
  created_at?: string;
}

function wsBaseFromApiUrl(): string {
  const base =
    API_BASE_URL.startsWith('http://') || API_BASE_URL.startsWith('https://')
      ? API_BASE_URL
      : `${typeof window !== 'undefined' ? window.location.origin : ''}${API_BASE_URL}`;
  return base.replace(/^http/, 'ws').replace(/\/$/, '');
}

function normalizePriority(p?: string): NotificationPriority {
  if (p === 'urgent') return 'critical';
  if (p === 'high') return 'high';
  if (p === 'low') return 'low';
  return 'medium';
}

function payloadToNotification(payload: StreamPayload, userId: string): Notification | null {
  if (!payload.id || !payload.title) return null;
  return {
    id: payload.id,
    type: (payload.type === 'notification' ? 'info' : payload.type) as Notification['type'],
    category: normalizeNotificationCategory(String(payload.category ?? 'all')),
    title: payload.title,
    message: payload.body ?? '',
    timestamp: payload.created_at ?? new Date().toISOString(),
    priority: normalizePriority(payload.priority),
    status: 'unread',
    metadata: {},
    userId,
    tenantId: '',
  };
}

export interface UseNotificationStreamOptions {
  enabled?: boolean;
  onNotification?: (notification: Notification) => void;
}

export function useNotificationStream(options: UseNotificationStreamOptions = {}) {
  const { enabled = true, onNotification } = options;
  const userId = useAuthStore((s) => s.user?.id);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectAttempts = useRef(0);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const onNotificationRef = useRef(onNotification);
  onNotificationRef.current = onNotification;

  const connect = useCallback(async () => {
    if (!enabled || !isAuthenticated || !userId) return;

    const url = new URL(`${wsBaseFromApiUrl()}/v1/notifications/stream`);

    const ws = new WebSocket(url.toString());
    wsRef.current = ws;

    ws.onopen = () => {
      setIsConnected(true);
      setError(null);
      reconnectAttempts.current = 0;
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data as string) as {
          type?: string;
          payload?: StreamPayload | { message?: string };
        };
        if (data.type === 'connected') return;
        if (data.type !== 'notification') return;
        const payload = data.payload as StreamPayload | undefined;
        if (!payload) return;
        const notification = payloadToNotification(payload, userId);
        if (notification) {
          onNotificationRef.current?.(notification);
        }
      } catch {
        /* ignore malformed frames */
      }
    };

    ws.onclose = () => {
      setIsConnected(false);
      wsRef.current = null;
      if (enabled && isAuthenticated && reconnectAttempts.current < 5) {
        reconnectAttempts.current += 1;
        const delay = Math.min(1000 * 2 ** reconnectAttempts.current, 30000);
        reconnectTimeoutRef.current = setTimeout(() => void connect(), delay);
      }
    };

    ws.onerror = () => {
      setError('Notification stream connection failed');
    };
  }, [enabled, isAuthenticated, userId]);

  useEffect(() => {
    if (!enabled || !isAuthenticated) return;
    void connect();
    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
        reconnectTimeoutRef.current = null;
      }
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [enabled, isAuthenticated, connect]);

  return { isConnected, error };
}
