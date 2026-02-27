import { useEffect, useRef, useState, useCallback } from 'react';
import { useAuthStore } from '../stores/authStore';
import { RealtimeEvent } from './types';

// WebSocket connection state
interface WebSocketConnection {
  ws: WebSocket | null;
  isConnected: boolean;
  reconnectAttempts: number;
  maxReconnectAttempts: number;
  reconnectDelay: number;
}

// Generic hook for real-time subscriptions using WebSocket
export function useRealtimeSubscription<T extends RealtimeEvent>(
  channelName: string,
  eventType: string,
  onEvent?: (event: T) => void
) {
  const { user } = useAuthStore();
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const connectionRef = useRef<WebSocketConnection>({
    ws: null,
    isConnected: false,
    reconnectAttempts: 0,
    maxReconnectAttempts: 5,
    reconnectDelay: 1000,
  });

  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Get WebSocket URL from environment with authentication (must be absolute)
  const getWebSocketUrl = useCallback(() => {
    const apiUrl = (import.meta.env.VITE_API_URL as string) ?? '';
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;

    let href: string;
    if (apiUrl.startsWith('http://') || apiUrl.startsWith('https://')) {
      const base = apiUrl.replace(/^http/, 'ws');
      href = `${base.endsWith('/') ? base.slice(0, -1) : base}/api/monitoring/realtime`;
    } else {
      const path = apiUrl.startsWith('/') ? apiUrl : `/${apiUrl || 'api'}`;
      href = `${protocol}//${host}${path}/monitoring/realtime`;
    }

    const token = localStorage.getItem('sb-access-token');
    const url = new URL(href);
    if (token) {
      url.searchParams.set('token', token);
    }

    return url.toString();
  }, []);

  // Handle incoming WebSocket messages
  const handleMessage = useCallback((event: MessageEvent) => {
    try {
      const data = JSON.parse(event.data);

      // Handle system messages
      if (data.type === 'connection_established') {
        setIsConnected(true);
        setError(null);
        return;
      }

      if (data.type === 'subscribed' || data.type === 'unsubscribed') {
        // Channel subscription status messages
        return;
      }

      // Handle broadcast messages
      if (data.type === 'broadcast' && data.event === eventType && data.payload) {
        if (onEvent) {
          onEvent(data.payload as T);
        }
      }
    } catch (err) {
      console.error('Failed to parse WebSocket message:', err);
    }
  }, [eventType, onEvent]);

  // Connect to WebSocket
  const connect = useCallback(() => {
    if (connectionRef.current.ws?.readyState === WebSocket.OPEN) {
      return; // Already connected
    }

    try {
      const wsUrl = getWebSocketUrl();
      const ws = new WebSocket(wsUrl);

      ws.onopen = () => {
        console.log('WebSocket connected to', wsUrl);
        connectionRef.current.isConnected = true;
        connectionRef.current.reconnectAttempts = 0;

        // Subscribe to the specified channel
        ws.send(JSON.stringify({
          type: 'subscribe',
          channel: channelName,
        }));

        setIsConnected(true);
        setError(null);
      };

      ws.onmessage = handleMessage;

      ws.onclose = (event) => {
        console.log('WebSocket disconnected:', event.code, event.reason);
        connectionRef.current.isConnected = false;
        setIsConnected(false);

        // Attempt to reconnect if not a clean close and within retry limits
        if (!event.wasClean && connectionRef.current.reconnectAttempts < connectionRef.current.maxReconnectAttempts) {
          connectionRef.current.reconnectAttempts++;
          const delay = connectionRef.current.reconnectDelay * Math.pow(2, connectionRef.current.reconnectAttempts - 1);

          console.log(`Attempting to reconnect in ${delay}ms (attempt ${connectionRef.current.reconnectAttempts}/${connectionRef.current.maxReconnectAttempts})`);

          reconnectTimeoutRef.current = setTimeout(() => {
            connect();
          }, delay);
        } else if (connectionRef.current.reconnectAttempts >= connectionRef.current.maxReconnectAttempts) {
          setError('Failed to reconnect after maximum attempts');
        }
      };

      ws.onerror = (event) => {
        console.error('WebSocket error:', event);
        setError('WebSocket connection error');
      };

      connectionRef.current.ws = ws;
    } catch (err) {
      console.error('Failed to create WebSocket connection:', err);
      setError('Failed to establish WebSocket connection');
    }
  }, [channelName, getWebSocketUrl, handleMessage]);

  // Disconnect from WebSocket
  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    if (connectionRef.current.ws) {
      connectionRef.current.ws.close(1000, 'Component unmounting');
      connectionRef.current.ws = null;
    }

    connectionRef.current.isConnected = false;
    setIsConnected(false);
  }, []);

  // Send message to channel
  const sendMessage = useCallback((event: string, payload: any) => {
    if (connectionRef.current.ws?.readyState === WebSocket.OPEN) {
      connectionRef.current.ws.send(JSON.stringify({
        type: 'broadcast',
        event,
        payload,
        channel: channelName,
      }));
    } else {
      console.warn('WebSocket not connected, cannot send message');
    }
  }, [channelName]);

  // Effect to manage WebSocket connection
  useEffect(() => {
    if (user?.id) {
      connect();
    }

    return () => {
      disconnect();
    };
  }, [user?.id, connect, disconnect]);

  // Effect to handle channel name changes
  useEffect(() => {
    if (connectionRef.current.ws?.readyState === WebSocket.OPEN) {
      // Unsubscribe from old channel and subscribe to new one
      connectionRef.current.ws.send(JSON.stringify({
        type: 'unsubscribe',
        channel: channelName,
      }));

      connectionRef.current.ws.send(JSON.stringify({
        type: 'subscribe',
        channel: channelName,
      }));
    }
  }, [channelName]);

  return {
    isConnected,
    error,
    sendMessage,
  };
}
