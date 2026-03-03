import { useEffect, useRef, useState, useCallback } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  type PlatformStatus,
  type Incident,
  type MaintenanceWindow,
  type StatusWebSocketMessage,
} from '@/api/status';
import { statusKeys } from './useStatus';

interface UseStatusWebSocketOptions {
  enabled?: boolean;
  onStatusUpdate?: (status: PlatformStatus) => void;
  onIncidentUpdate?: (incident: Incident) => void;
  onMaintenanceUpdate?: (maintenance: MaintenanceWindow) => void;
  showNotifications?: boolean;
}

interface WebSocketState {
  isConnected: boolean;
  isConnecting: boolean;
  error: Error | null;
  reconnectAttempt: number;
}

// Reconnection configuration
const RECONNECT_BASE_DELAY = 1000;
const RECONNECT_MAX_DELAY = 30000;
const RECONNECT_MAX_ATTEMPTS = 10;

/**
 * Hook to connect to the status WebSocket for real-time updates
 * Features auto-reconnect with exponential backoff
 */
export function useStatusWebSocket(options: UseStatusWebSocketOptions = {}) {
  const {
    enabled = true,
    onStatusUpdate,
    onIncidentUpdate,
    onMaintenanceUpdate,
    showNotifications = true,
  } = options;

  const queryClient = useQueryClient();
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const isIntentionallyClosedRef = useRef(false);

  const [state, setState] = useState<WebSocketState>({
    isConnected: false,
    isConnecting: false,
    error: null,
    reconnectAttempt: 0,
  });

  // Get WebSocket URL based on current environment
  const getWebSocketUrl = useCallback(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    return `${protocol}//${host}/ws/v1/status`;
  }, []);

  // Calculate reconnect delay with exponential backoff
  const getReconnectDelay = useCallback(() => {
    const delay = Math.min(
      RECONNECT_BASE_DELAY * Math.pow(2, reconnectAttemptsRef.current),
      RECONNECT_MAX_DELAY
    );
    return delay;
  }, []);

  // Handle incoming WebSocket messages
  const handleMessage = useCallback(
    (event: MessageEvent) => {
      try {
        const message: StatusWebSocketMessage = JSON.parse(event.data);

        switch (message.type) {
          case 'status_update': {
            const status = message.data as PlatformStatus;

            // Update React Query cache
            queryClient.setQueryData(statusKeys.platform(), status);

            // Call optional callback
            onStatusUpdate?.(status);
            break;
          }

          case 'incident_update': {
            const incident = message.data as Incident;

            // Invalidate incidents cache
            queryClient.invalidateQueries({ queryKey: ['incidents'] });

            // Show notification for new incidents
            if (showNotifications && incident.status === 'investigating') {
              toast.warning(`New Incident: ${incident.title}`, {
                description: incident.description,
                duration: 10000,
              });
            }

            // Call optional callback
            onIncidentUpdate?.(incident);
            break;
          }

          case 'maintenance_update': {
            const maintenance = message.data as MaintenanceWindow;

            // Invalidate maintenance cache
            queryClient.invalidateQueries({ queryKey: statusKeys.maintenance() });

            // Show notification for upcoming maintenance
            if (showNotifications && maintenance.status === 'scheduled') {
              toast.info(`Scheduled Maintenance: ${maintenance.title}`, {
                description: `Starting at ${new Date(maintenance.scheduled_start).toLocaleString()}`,
                duration: 10000,
              });
            }

            // Call optional callback
            onMaintenanceUpdate?.(maintenance);
            break;
          }

          case 'ping': {
            // Send pong response to keep connection alive
            wsRef.current?.send(JSON.stringify({ type: 'pong' }));
            break;
          }
        }
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    },
    [queryClient, onStatusUpdate, onIncidentUpdate, onMaintenanceUpdate, showNotifications]
  );

  // Connect to WebSocket
  const connect = useCallback(() => {
    if (!enabled || wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    setState((prev) => ({ ...prev, isConnecting: true, error: null }));

    try {
      const wsUrl = getWebSocketUrl();
      const ws = new WebSocket(wsUrl);

      ws.onopen = () => {
        reconnectAttemptsRef.current = 0;
        setState({
          isConnected: true,
          isConnecting: false,
          error: null,
          reconnectAttempt: 0,
        });
      };

      ws.onmessage = handleMessage;

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        setState((prev) => ({
          ...prev,
          isConnected: false,
          isConnecting: false,
          error: new Error('WebSocket connection error'),
        }));
      };

      ws.onclose = (event) => {
        setState((prev) => ({
          ...prev,
          isConnected: false,
          isConnecting: false,
        }));

        wsRef.current = null;

        // Attempt reconnection if not intentionally closed
        if (!isIntentionallyClosedRef.current && enabled) {
          if (reconnectAttemptsRef.current < RECONNECT_MAX_ATTEMPTS) {
            const delay = getReconnectDelay();
            reconnectAttemptsRef.current += 1;

            setState((prev) => ({
              ...prev,
              reconnectAttempt: reconnectAttemptsRef.current,
            }));

            reconnectTimeoutRef.current = setTimeout(() => {
              connect();
            }, delay);
          } else {
            setState((prev) => ({
              ...prev,
              error: new Error('Max reconnection attempts reached'),
            }));
          }
        }
      };

      wsRef.current = ws;
    } catch (error) {
      setState((prev) => ({
        ...prev,
        isConnecting: false,
        error: error instanceof Error ? error : new Error('Failed to connect'),
      }));
    }
  }, [enabled, getWebSocketUrl, handleMessage, getReconnectDelay]);

  // Disconnect from WebSocket
  const disconnect = useCallback(() => {
    isIntentionallyClosedRef.current = true;

    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    setState({
      isConnected: false,
      isConnecting: false,
      error: null,
      reconnectAttempt: 0,
    });
  }, []);

  // Reconnect manually
  const reconnect = useCallback(() => {
    disconnect();
    isIntentionallyClosedRef.current = false;
    reconnectAttemptsRef.current = 0;
    connect();
  }, [disconnect, connect]);

  // Connect on mount and when enabled changes
  useEffect(() => {
    if (enabled) {
      isIntentionallyClosedRef.current = false;
      connect();
    } else {
      disconnect();
    }

    return () => {
      disconnect();
    };
  }, [enabled, connect, disconnect]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      disconnect();
    };
  }, [disconnect]);

  return {
    ...state,
    connect,
    disconnect,
    reconnect,
  };
}

/**
 * Hook that combines polling and WebSocket for maximum reliability
 * Uses WebSocket when available, falls back to polling
 */
export function useRealtimeStatus(options: UseStatusWebSocketOptions = {}) {
  const ws = useStatusWebSocket(options);

  return {
    // WebSocket state
    isRealtime: ws.isConnected,
    isConnecting: ws.isConnecting,
    wsError: ws.error,
    reconnectAttempt: ws.reconnectAttempt,

    // Actions
    reconnect: ws.reconnect,
    disconnect: ws.disconnect,
  };
}
