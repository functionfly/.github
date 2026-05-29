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
  const onNotificationRef = useRef(onNotification);
  onNotificationRef.current = onNotification;

  const connect = useCallback(() => {
    if (!enabled || !isAuthenticated || !userId) return;

    const token = localStorage.getItem('ff-access-token');
    if (!token?.trim()) return;

    const url = new URL(`${wsBaseFromApiUrl()}/v1/notifications/stream`);
    url.searchParams.set('token', token);

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

    ws.onerror = () => {
      setError('Notification stream connection error');
    };

    ws.onclose = () => {
      setIsConnected(false);
      wsRef.current = null;
      if (!enabled || !isAuthenticated) return;
      if (reconnectAttempts.current >= 5) return;
      reconnectAttempts.current += 1;
      const delay = Math.min(1000 * 2 ** reconnectAttempts.current, 30000);
      window.setTimeout(connect, delay);
    };
  }, [enabled, isAuthenticated, userId]);

  useEffect(() => {
    connect();
    return () => {
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, [connect]);

  return { isConnected, error };
}

export default useNotificationStream;
