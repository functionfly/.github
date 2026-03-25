import { useEffect, useRef, useState, useCallback } from 'react';
import { CACHE_KEYS } from '@/lib/constants';

export interface RealtimeEvent {
  type: string;
  timestamp?: string;
  [key: string]: unknown;
}

interface UseRealtimeSubscriptionResult {
  isConnected: boolean;
  error: string | null;
  sendMessage: (event: string, payload: unknown) => void;
}

interface UseRealtimeSubscriptionOptions {
  channel: string;
  eventType: string;
  onEvent?: (event: RealtimeEvent) => void;
}

export function useRealtimeSubscription({
  channel,
  eventType,
  onEvent,
}: UseRealtimeSubscriptionOptions): UseRealtimeSubscriptionResult {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectRef = useRef<number>(0);
  const reconnectTimeoutRef = useRef<number | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Keep a ref to onEvent so changing the callback never recreates the WebSocket.
  const onEventRef = useRef(onEvent);
  useEffect(() => {
    onEventRef.current = onEvent;
  }, [onEvent]);

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    const token = sessionStorage.getItem(CACHE_KEYS.ADMIN_ACCESS_TOKEN);
    if (!token) {
      setError('Missing admin access token for realtime connection');
      return;
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const baseUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';
    const wsBase = baseUrl.replace(/^http/, 'ws');
    const wsUrl = new URL(`${protocol}//${new URL(wsBase).host}/v1/monitoring/realtime`);
    wsUrl.searchParams.set('token', token);

    const ws = new WebSocket(wsUrl.toString());

    ws.onopen = () => {
      setIsConnected(true);
      setError(null);
      reconnectRef.current = 0;
      ws.send(JSON.stringify({ type: 'subscribe', channel }));
    };

    ws.onmessage = (message) => {
      try {
        const data = JSON.parse(message.data as string);
        if (data.type === 'broadcast' && data.event === eventType && onEventRef.current) {
          onEventRef.current(data.payload as RealtimeEvent);
        }
      } catch {
        setError('Failed to parse realtime message');
      }
    };

    ws.onclose = () => {
      setIsConnected(false);
      if (reconnectRef.current < 5) {
        reconnectRef.current += 1;
        const delay = 1000 * Math.pow(2, reconnectRef.current - 1);
        reconnectTimeoutRef.current = window.setTimeout(connect, delay);
      }
    };

    ws.onerror = () => {
      setError('Realtime WebSocket error');
    };

    wsRef.current = ws;
  }, [channel, eventType]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      window.clearTimeout(reconnectTimeoutRef.current);
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    setIsConnected(false);
  }, []);

  const sendMessage = useCallback((event: string, payload: unknown) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(
        JSON.stringify({ type: 'broadcast', channel, event, payload })
      );
    }
  }, [channel]);

  useEffect(() => {
    connect();
    return () => disconnect();
  }, [connect, disconnect]);

  return { isConnected, error, sendMessage };
}
