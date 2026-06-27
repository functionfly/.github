import axios from 'axios';
import { useCallback, useEffect, useRef, useState } from 'react';
import { getApiBaseUrl } from '../lib/constants';
import { tokenVault } from '@/utils/token-vault';
import { useAuthStore } from '../stores/authStore';
import { RealtimeEvent } from './types';

// Helper to get human-readable WebSocket close reason
function getWebSocketCloseReason(code: number, reason?: string): string {
  const codeReasons: Record<number, string> = {
    1000: 'Normal closure',
    1001: 'Going away (browser close/navigate)',
    1002: 'Protocol error',
    1003: 'Unsupported data',
    1004: 'Reserved',
    1005: 'No status received',
    1006: 'Abnormal closure (connection lost/server crash)',
    1007: 'Invalid frame payload',
    1008: 'Policy violation (auth failed)',
    1009: 'Message too big',
    1010: 'Mandatory extension missing',
    1011: 'Internal server error',
    1012: 'Service restart',
    1013: 'Try again later',
    1014: 'Bad gateway',
    1015: 'TLS handshake failure',
  };
  return codeReasons[code] || reason || 'Unknown reason';
}

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

  // Get WebSocket URL from same base as REST API (getApiBaseUrl())
  const getWebSocketUrl = useCallback(async () => {
    const base = getApiBaseUrl();
    const wsBase = base.replace(/^http/, 'ws');
    
    // Get token from TokenVault and append as query param
    await tokenVault.initialize();
    const token = await tokenVault.getAccessToken();
    
    const href = token
      ? `${wsBase.replace(/\/$/, '')}/v1/monitoring/realtime?token=${encodeURIComponent(token)}`
      : `${wsBase.replace(/\/$/, '')}/v1/monitoring/realtime`;
    
    return href;
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
  const connect = useCallback(async () => {
    if (connectionRef.current.ws?.readyState === WebSocket.OPEN) {
      return;
    }

    // Don't connect if no token
    const token = await tokenVault.getAccessToken();
    if (!token || !token.trim()) {
      return;
    }

    errorSetForAttemptRef.current = false;

    try {
      const wsUrl = await getWebSocketUrl();
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
        // Log detailed close reason in development
        if (import.meta.env.DEV) {
          const closeReason = getWebSocketCloseReason(event.code, event.reason);
          console.log(`WebSocket disconnected: ${event.code} (${closeReason})`, {
            wasClean: event.wasClean,
            reason: event.reason || 'No reason provided',
          });
        }

        connectionRef.current.isConnected = false;
        connectionRef.current.ws = null;
        setIsConnected(false);

        // Don't retry on authentication failures (close codes that indicate auth issues)
        const isAuthFailure =
          event.code === 1008 || event.code === 1011 || event.reason?.includes('Authentication');
        const isAbnormalClosure = event.code === 1006;

        const shouldRetry =
          !isAuthFailure &&
          !event.wasClean &&
          connectionRef.current.reconnectAttempts < connectionRef.current.maxReconnectAttempts;

        if (shouldRetry) {
          connectionRef.current.reconnectAttempts++;

          const delay =
            connectionRef.current.reconnectDelay *
            Math.pow(2, Math.min(connectionRef.current.reconnectAttempts - 1, 4)); // Cap at 16x delay
          
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
          // Delegate to auth store's logout — it handles token cleanup, state reset,
          // and cross-tab sync. Don't clear tokens directly as it races with initialize().
          if (!errorSetForAttemptRef.current) {
            errorSetForAttemptRef.current = true;
            setError('Authentication failed - please log in again');
            import('@/stores/authStore').then(({ useAuthStore }) => {
              useAuthStore.getState().logout(false);
            });
          }
        } else if (isAbnormalClosure) {
          // Set error once per abnormal close to avoid update storm (onerror + onclose both fire)
          if (!errorSetForAttemptRef.current) {
            errorSetForAttemptRef.current = true;
            setError('WebSocket connection interrupted - check if backend is running');
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

  // Track if a connection attempt is in progress to prevent duplicates
  const connectionAttemptRef = useRef<boolean>(false);
  const lastConnectTimeRef = useRef<number>(0);

  // Effect to manage WebSocket connection
  useEffect(() => {
    const checkAuthAndConnect = async () => {
      await tokenVault.initialize();
      const token = await tokenVault.getAccessToken();
      const hasAuth = !!(user?.id && user?.email && token);

      // Only log in dev when there's actual auth data to debug (reduces noise on public pages)
      if (import.meta.env.DEV && (user?.id || token)) {
        console.log('WebSocket auth check:', {
          userId: user?.id,
          userEmail: user?.email,
          hasToken: !!token,
          willConnect: hasAuth,
        });
      }

      // First check basic auth state
      if (!hasAuth) {
        disconnect();
        return;
      }

      // Prevent multiple simultaneous connection attempts
      if (connectionAttemptRef.current) {
        return;
      }

      // Debounce connection attempts - minimum 2 seconds between attempts
      const now = Date.now();
      if (now - lastConnectTimeRef.current < 2000) {
        return;
      }

      // Validate session with server
      connectionAttemptRef.current = true;
      try {
        const response = await axios.get(`${getApiBaseUrl()}/v1/auth/validate`, {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (response.status === 200) {
          if (import.meta.env.DEV) {
            console.log('Session valid, attempting WebSocket connection');
          }
          lastConnectTimeRef.current = Date.now();
          connect();
        } else {
          throw new Error('Session invalid');
        }
      } catch (error) {
        if (import.meta.env.DEV) {
          console.log('Session validation failed, deferring to auth store');
        }
        // Don't clear tokens directly — this races with initialize() on page refresh.
        // Let the auth store handle token lifecycle via its own validation/refresh flow.
        disconnect();
      } finally {
        connectionAttemptRef.current = false;
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
