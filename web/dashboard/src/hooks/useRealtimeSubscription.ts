import axios from 'axios';
import { useCallback, useEffect, useRef, useState } from 'react';
import { API_BASE_URL } from '../lib/constants';
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
  const pingIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const errorSetForAttemptRef = useRef(false);

  // Refs so connect/handleMessage don't change when parent re-renders (avoids effect loop)
  const channelNameRef = useRef(channelName);
  const eventTypeRef = useRef(eventType);
  const onEventRef = useRef(onEvent);
  channelNameRef.current = channelName;
  eventTypeRef.current = eventType;
  onEventRef.current = onEvent;

  // Get WebSocket URL from same base as REST API (VITE_API_URL / API_BASE_URL)
  const getWebSocketUrl = useCallback(() => {
    const base =
      API_BASE_URL.startsWith('http://') || API_BASE_URL.startsWith('https://')
        ? API_BASE_URL
        : `${typeof window !== 'undefined' ? window.location.origin : ''}${API_BASE_URL}`;
    const wsBase = base.replace(/^http/, 'ws');
    const href = `${wsBase.replace(/\/$/, '')}/v1/monitoring/realtime`;

    const token = localStorage.getItem('ff-access-token');
    const url = new URL(href);
    if (token && token.trim()) {
      url.searchParams.set('token', token);
    }

    return url.toString();
  }, []);

  // Handle incoming WebSocket messages (stable callback; reads from refs)
  const handleMessage = useCallback((event: MessageEvent) => {
    try {
      const data = JSON.parse(event.data);
      const eventType = eventTypeRef.current;
      const onEvent = onEventRef.current;

      // Handle system messages
      if (data.type === 'connection_established') {
        setIsConnected(true);
        setError(null);
        return;
      }

      if (data.type === 'subscribed' || data.type === 'unsubscribed' || data.type === 'pong') {
        return;
      }

      if (data.type === 'broadcast' && data.event === eventType && data.payload) {
        if (onEvent) {
          onEvent(data.payload as T);
        }
      }
    } catch (err) {
      console.error('Failed to parse WebSocket message:', err);
    }
  }, []);

  // Connect to WebSocket (stable: only depends on getWebSocketUrl)
  const connect = useCallback(() => {
    if (connectionRef.current.ws?.readyState === WebSocket.OPEN) {
      return;
    }

    // Don't connect if no token
    const token = localStorage.getItem('ff-access-token');
    if (!token || !token.trim()) {
      return;
    }

    errorSetForAttemptRef.current = false;

    try {
      const wsUrl = getWebSocketUrl();
      const ws = new WebSocket(wsUrl);
      const channel = channelNameRef.current;

      ws.onopen = () => {
        if (import.meta.env.DEV) {
          console.log('WebSocket connected to', wsUrl);
        }
        connectionRef.current.isConnected = true;
        connectionRef.current.reconnectAttempts = 0;

        // Start ping interval to keep connection alive
        startPingInterval();

        ws.send(
          JSON.stringify({
            type: 'subscribe',
            channel,
          })
        );

        setIsConnected(true);
        setError(null);
      };

      ws.onmessage = handleMessage;

      ws.onclose = (event) => {
        if (import.meta.env.DEV) {
          console.log('WebSocket disconnected:', event.code, event.reason);
        }
        connectionRef.current.isConnected = false;
        connectionRef.current.ws = null;
        setIsConnected(false);

        // Don't retry on authentication failures (close codes that indicate auth issues)
        const isAuthFailure =
          event.code === 1008 || event.code === 1011 || event.reason?.includes('Authentication');
        const shouldRetry =
          !isAuthFailure &&
          !event.wasClean &&
          connectionRef.current.reconnectAttempts < connectionRef.current.maxReconnectAttempts;

        if (shouldRetry) {
          connectionRef.current.reconnectAttempts++;
          const delay =
            connectionRef.current.reconnectDelay *
            Math.pow(2, connectionRef.current.reconnectAttempts - 1);
          if (import.meta.env.DEV) {
            console.log(
              `Attempting to reconnect in ${delay}ms (attempt ${connectionRef.current.reconnectAttempts}/${connectionRef.current.maxReconnectAttempts})`
            );
          }
          reconnectTimeoutRef.current = setTimeout(() => {
            connect();
          }, delay);
        } else if (
          connectionRef.current.reconnectAttempts >= connectionRef.current.maxReconnectAttempts &&
          !errorSetForAttemptRef.current
        ) {
          errorSetForAttemptRef.current = true;
          setError('Failed to reconnect after maximum attempts');
        } else if (isAuthFailure) {
          // Clear invalid token on authentication failure
          localStorage.removeItem('ff-access-token');
          if (!errorSetForAttemptRef.current) {
            errorSetForAttemptRef.current = true;
            setError('Authentication failed - please log in again');
          }
        } else if (event.code === 1006 || !event.wasClean) {
          // Set error once per abnormal close to avoid update storm (onerror + onclose both fire)
          if (!errorSetForAttemptRef.current) {
            errorSetForAttemptRef.current = true;
            setError('WebSocket connection error');
          }
        }
      };

      ws.onerror = (event) => {
        // Check if this is an authentication failure by trying to detect it from the error
        // Since we can't get the HTTP status code from the WebSocket error, we'll let onclose handle it
        if (process.env.NODE_ENV === 'development') {
          console.warn('WebSocket error (connection may have closed abnormally)', event);
        }
      };

      connectionRef.current.ws = ws;
    } catch (err) {
      console.error('Failed to create WebSocket connection:', err);
      setError('Failed to establish WebSocket connection');
    }
  }, [getWebSocketUrl, handleMessage]);

  // Start ping interval to keep connection alive
  const startPingInterval = useCallback(() => {
    if (pingIntervalRef.current) {
      clearInterval(pingIntervalRef.current);
    }

    // Send ping every 30 seconds to keep connection alive (server timeout is 60 seconds)
    pingIntervalRef.current = setInterval(() => {
      if (connectionRef.current.ws?.readyState === WebSocket.OPEN) {
        try {
          connectionRef.current.ws.send(JSON.stringify({ type: 'ping' }));
        } catch (err) {
          console.warn('Failed to send ping:', err);
        }
      }
    }, 30000);
  }, []);

  // Stop ping interval
  const stopPingInterval = useCallback(() => {
    if (pingIntervalRef.current) {
      clearInterval(pingIntervalRef.current);
      pingIntervalRef.current = null;
    }
  }, []);

  // Disconnect from WebSocket
  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    stopPingInterval();

    if (connectionRef.current.ws) {
      connectionRef.current.ws.close(1000, 'Component unmounting');
      connectionRef.current.ws = null;
    }

    connectionRef.current.isConnected = false;
    setIsConnected(false);
  }, [stopPingInterval]);

  // Send message to channel
  const sendMessage = useCallback(
    (event: string, payload: any) => {
      if (connectionRef.current.ws?.readyState === WebSocket.OPEN) {
        connectionRef.current.ws.send(
          JSON.stringify({
            type: 'broadcast',
            event,
            payload,
            channel: channelName,
          })
        );
      } else {
        console.warn('WebSocket not connected, cannot send message');
      }
    },
    [channelName]
  );

  // Effect to manage WebSocket connection
  useEffect(() => {
    const checkAuthAndConnect = async () => {
      const token = localStorage.getItem('ff-access-token');
      if (import.meta.env.DEV) {
        console.log('WebSocket auth check:', {
          userId: user?.id,
          userEmail: user?.email,
          hasToken: !!token,
        });
      }

      // First check basic auth state
      if (!user?.id || !user?.email || !token) {
        if (import.meta.env.DEV) {
          console.log('Skipping WebSocket connection - missing auth data');
        }
        if (token) {
          if (import.meta.env.DEV) {
            console.log('Clearing invalid token');
          }
          localStorage.removeItem('ff-access-token');
        }
        disconnect();
        return;
      }

      // Validate session with server
      try {
        const response = await axios.get('/api/v1/auth/validate', {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (response.status === 200) {
          if (import.meta.env.DEV) {
            console.log('Session valid, attempting WebSocket connection');
          }
          connect();
        } else {
          throw new Error('Session invalid');
        }
      } catch (error) {
        if (import.meta.env.DEV) {
          console.log('Session validation failed, clearing auth state');
        }
        localStorage.removeItem('ff-access-token');
        disconnect();
      }
    };

    checkAuthAndConnect();

    return () => {
      disconnect();
    };
  }, [user?.id, user?.email, connect, disconnect]);

  // Effect to handle channel name changes
  useEffect(() => {
    if (connectionRef.current.ws?.readyState === WebSocket.OPEN) {
      // Unsubscribe from old channel and subscribe to new one
      connectionRef.current.ws.send(
        JSON.stringify({
          type: 'unsubscribe',
          channel: channelName,
        })
      );

      connectionRef.current.ws.send(
        JSON.stringify({
          type: 'subscribe',
          channel: channelName,
        })
      );
    }
  }, [channelName]);

  return {
    isConnected,
    error,
    sendMessage,
  };
}
