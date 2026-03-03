/**
 * Notification Real-time Hook
 *
 * A specialized hook that wraps useRealtimeSubscription with notification-specific
 * logic for handling WebSocket events related to notifications, trust alerts,
 * and activity feed items.
 */

import { useEffect, useRef, useState, useCallback, useMemo } from 'react';
import { useRealtimeSubscription } from './useRealtimeSubscription';
import { RealtimeEvent } from './types';
import {
  Notification,
  TrustAlert,
  ActivityFeedItem,
  NotificationCategory,
} from '@/types/notifications';

// ============================================================================
// Type Definitions
// ============================================================================

export interface UseNotificationRealtimeOptions {
  /** Callback invoked when a new notification is received */
  onNewNotification?: (notification: Notification) => void;
  /** Callback invoked when a trust alert is received */
  onTrustAlert?: (alert: TrustAlert) => void;
  /** Callback invoked when an activity feed item is received */
  onActivity?: (activity: ActivityFeedItem) => void;
  /** Callback invoked when a notification is marked as read */
  onNotificationRead?: (notificationId: string) => void;
  /** Callback invoked when multiple notifications are updated */
  onBulkUpdate?: (notificationIds: string[]) => void;
  /** Whether the WebSocket subscription is enabled */
  enabled?: boolean;
  /** Categories to filter notifications by */
  categories?: NotificationCategory[];
}

export type ConnectionState = 'connecting' | 'connected' | 'disconnected' | 'reconnecting';

export interface UseNotificationRealtimeReturn {
  /** Whether the WebSocket is currently connected */
  isConnected: boolean;
  /** Current connection state with visual feedback states */
  connectionState: ConnectionState;
  /** Any error that occurred during connection or operation */
  error: Error | null;
  /** Function to manually trigger a reconnection */
  reconnect: () => void;
  /** The last event received */
  lastEvent: RealtimeEvent | null;
}

// Event types handled by this hook
type NotificationEventType =
  | 'new_notification'
  | 'notification_read'
  | 'trust_alert'
  | 'activity'
  | 'bulk_update';

interface NotificationEvent extends RealtimeEvent {
  type: NotificationEventType;
  payload?: unknown;
}

// ============================================================================
// Type Guards
// ============================================================================

/**
 * Type guard to check if payload is a Notification
 */
function isNotification(payload: unknown): payload is Notification {
  if (!payload || typeof payload !== 'object') return false;
  const p = payload as Record<string, unknown>;
  return (
    typeof p.id === 'string' &&
    typeof p.type === 'string' &&
    typeof p.category === 'string' &&
    typeof p.title === 'string' &&
    typeof p.message === 'string' &&
    typeof p.timestamp === 'string' &&
    typeof p.priority === 'string' &&
    typeof p.status === 'string' &&
    typeof p.userId === 'string' &&
    typeof p.tenantId === 'string'
  );
}

/**
 * Type guard to check if payload is a TrustAlert
 */
function isTrustAlert(payload: unknown): payload is TrustAlert {
  if (!payload || typeof payload !== 'object') return false;
  const p = payload as Record<string, unknown>;
  return (
    typeof p.id === 'string' &&
    (p.type === 'trust_drop' || p.type === 'replay_failed' || p.type === 'determinism_broken') &&
    (p.severity === 'warning' || p.severity === 'critical' || p.severity === 'emergency') &&
    typeof p.title === 'string' &&
    typeof p.description === 'string' &&
    Array.isArray(p.affectedFunctions) &&
    typeof p.recommendedAction === 'string' &&
    typeof p.triggeredAt === 'string' &&
    typeof p.acknowledged === 'boolean'
  );
}

/**
 * Type guard to check if payload is an ActivityFeedItem
 */
function isActivityFeedItem(payload: unknown): payload is ActivityFeedItem {
  if (!payload || typeof payload !== 'object') return false;
  const p = payload as Record<string, unknown>;
  return (
    typeof p.id === 'string' &&
    p.actor && typeof p.actor === 'object' &&
    typeof (p.actor as Record<string, unknown>).id === 'string' &&
    typeof (p.actor as Record<string, unknown>).name === 'string' &&
    typeof p.action === 'string' &&
    p.target && typeof p.target === 'object' &&
    typeof (p.target as Record<string, unknown>).id === 'string' &&
    typeof (p.target as Record<string, unknown>).type === 'string' &&
    typeof (p.target as Record<string, unknown>).name === 'string' &&
    typeof p.timestamp === 'string'
  );
}

/**
 * Type guard to check if payload contains notification IDs for bulk update
 */
function isBulkUpdatePayload(payload: unknown): payload is { notificationIds: string[] } {
  if (!payload || typeof payload !== 'object') return false;
  const p = payload as Record<string, unknown>;
  return (
    Array.isArray(p.notificationIds) &&
    p.notificationIds.every((id) => typeof id === 'string')
  );
}

/**
 * Type guard to check if payload contains a notification ID
 */
function isNotificationIdPayload(payload: unknown): payload is { notificationId: string } {
  if (!payload || typeof payload !== 'object') return false;
  const p = payload as Record<string, unknown>;
  return typeof p.notificationId === 'string';
}

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Simple debounce function for rapid events
 */
function debounce<T extends (...args: any[]) => void>(
  fn: T,
  delay: number
): (...args: Parameters<T>) => void {
  let timeoutId: NodeJS.Timeout | null = null;
  return (...args: Parameters<T>) => {
    if (timeoutId) {
      clearTimeout(timeoutId);
    }
    timeoutId = setTimeout(() => {
      fn(...args);
      timeoutId = null;
    }, delay);
  };
}

/**
 * Get channel name based on categories
 */
function getNotificationChannel(categories?: NotificationCategory[]): string {
  if (!categories || categories.length === 0 || categories.includes('all')) {
    return 'notifications:all';
  }
  return `notifications:${categories.join(',')}`;
}

// ============================================================================
// Hook Implementation
// ============================================================================

/**
 * Hook for notification-specific real-time WebSocket operations
 *
 * This hook wraps useRealtimeSubscription with notification-specific logic
 * including event type routing, connection state tracking, event buffering,
 * and automatic reconnection.
 *
 * @example
 * ```typescript
 * const { isConnected, connectionState, error, reconnect } = useNotificationRealtime({
 *   onNewNotification: (notification) => console.log('New:', notification),
 *   onTrustAlert: (alert) => console.log('Trust Alert:', alert),
 *   onActivity: (activity) => console.log('Activity:', activity),
 *   categories: ['trust', 'security'],
 *   enabled: true,
 * });
 * ```
 */
export function useNotificationRealtime(
  options: UseNotificationRealtimeOptions = {}
): UseNotificationRealtimeReturn {
  const {
    onNewNotification,
    onTrustAlert,
    onActivity,
    onNotificationRead,
    onBulkUpdate,
    enabled = true,
    categories,
  } = options;

  // Connection state tracking
  const [connectionState, setConnectionState] = useState<ConnectionState>('disconnected');
  const [error, setError] = useState<Error | null>(null);
  const [lastEvent, setLastEvent] = useState<RealtimeEvent | null>(null);

  // Event buffer for disconnection periods
  const eventBufferRef = useRef<NotificationEvent[]>([]);
  const isBufferingRef = useRef(false);

  // Reconnection state
  const reconnectAttemptRef = useRef(0);
  const maxReconnectAttempts = 5;
  const reconnectDelayRef = useRef(1000);

  // Debounced callbacks
  const debouncedOnNewNotification = useMemo(
    () => (onNewNotification ? debounce(onNewNotification, 100) : undefined),
    [onNewNotification]
  );

  const debouncedOnActivity = useMemo(
    () => (onActivity ? debounce(onActivity, 50) : undefined),
    [onActivity]
  );

  // Channel name based on categories
  const channelName = useMemo(() => getNotificationChannel(categories), [categories]);

  /**
   * Process buffered events when connection is restored
   */
  const flushEventBuffer = useCallback(() => {
    if (eventBufferRef.current.length === 0) return;

    const bufferedEvents = [...eventBufferRef.current];
    eventBufferRef.current = [];
    isBufferingRef.current = false;

    // Process buffered events in order
    bufferedEvents.forEach((event) => {
      handleEvent(event);
    });

    if (process.env.NODE_ENV === 'development') {
      console.log(`[useNotificationRealtime] Flushed ${bufferedEvents.length} buffered events`);
    }
  }, []);

  /**
   * Route events to appropriate handlers
   */
  const handleEvent = useCallback((event: NotificationEvent) => {
    setLastEvent(event);

    switch (event.type) {
      case 'new_notification':
        if (isNotification(event.payload)) {
          debouncedOnNewNotification?.(event.payload);
        } else if (process.env.NODE_ENV === 'development') {
          console.warn('[useNotificationRealtime] Invalid notification payload:', event.payload);
        }
        break;

      case 'notification_read':
        if (isNotificationIdPayload(event.payload)) {
          onNotificationRead?.(event.payload.notificationId);
        } else if (process.env.NODE_ENV === 'development') {
          console.warn('[useNotificationRealtime] Invalid notification_read payload:', event.payload);
        }
        break;

      case 'trust_alert':
        if (isTrustAlert(event.payload)) {
          onTrustAlert?.(event.payload);
        } else if (process.env.NODE_ENV === 'development') {
          console.warn('[useNotificationRealtime] Invalid trust_alert payload:', event.payload);
        }
        break;

      case 'activity':
        if (isActivityFeedItem(event.payload)) {
          debouncedOnActivity?.(event.payload);
        } else if (process.env.NODE_ENV === 'development') {
          console.warn('[useNotificationRealtime] Invalid activity payload:', event.payload);
        }
        break;

      case 'bulk_update':
        if (isBulkUpdatePayload(event.payload)) {
          onBulkUpdate?.(event.payload.notificationIds);
        } else if (process.env.NODE_ENV === 'development') {
          console.warn('[useNotificationRealtime] Invalid bulk_update payload:', event.payload);
        }
        break;

      default:
        if (process.env.NODE_ENV === 'development') {
          console.warn('[useNotificationRealtime] Unknown event type:', (event as any).type);
        }
    }
  }, [debouncedOnNewNotification, onNotificationRead, onTrustAlert, debouncedOnActivity, onBulkUpdate]);

  /**
   * Handle connection state changes.
   * When isConnected becomes false, do not set 'disconnected' if we're already
   * 'connecting' — that would fight with the effect that sets 'connecting'
   * and cause an infinite update loop.
   */
  const handleConnectionChange = useCallback((connected: boolean) => {
    if (connected) {
      if (reconnectAttemptRef.current > 0) {
        setConnectionState('connected');
        reconnectAttemptRef.current = 0;
        reconnectDelayRef.current = 1000;
        flushEventBuffer();
      } else {
        setConnectionState('connected');
      }
      setError(null);
    } else {
      setConnectionState((prev) => {
        if (prev === 'connected') {
          isBufferingRef.current = true;
          return 'reconnecting';
        }
        // Don't transition to 'disconnected' when we're 'connecting' — avoids
        // loop with the effect that sets 'connecting' when disconnected.
        if (prev === 'connecting' || prev === 'reconnecting') return prev;
        return 'disconnected';
      });
    }
  }, [flushEventBuffer]);

  /**
   * Main event handler that either processes immediately or buffers
   */
  const onEvent = useCallback((event: RealtimeEvent) => {
    const notificationEvent = event as NotificationEvent;

    if (isBufferingRef.current) {
      // Buffer events during disconnection
      eventBufferRef.current.push(notificationEvent);

      // Prevent buffer from growing too large
      if (eventBufferRef.current.length > 100) {
        eventBufferRef.current = eventBufferRef.current.slice(-50);
        if (process.env.NODE_ENV === 'development') {
          console.warn('[useNotificationRealtime] Event buffer truncated to prevent memory issues');
        }
      }
    } else {
      handleEvent(notificationEvent);
    }
  }, [handleEvent]);

  // Set up the WebSocket subscription
  const { isConnected, error: wsError } = useRealtimeSubscription<RealtimeEvent>(
    channelName,
    'notification_event',
    onEvent
  );

  // Manual reconnect function
  const reconnect = useCallback(() => {
    if (reconnectAttemptRef.current < maxReconnectAttempts) {
      reconnectAttemptRef.current++;
      setConnectionState('reconnecting');

      // Exponential backoff
      const delay = reconnectDelayRef.current * Math.pow(2, reconnectAttemptRef.current - 1);
      reconnectDelayRef.current = delay;

      setTimeout(() => {
        // Force a reconnection by toggling connection state
        setConnectionState('connecting');
      }, Math.min(delay, 30000)); // Cap at 30 seconds
    } else {
      setConnectionState('disconnected');
      setError(new Error('Max reconnection attempts reached'));
    }
  }, []);

  // Update connection state when WebSocket state changes
  useEffect(() => {
    handleConnectionChange(isConnected);
  }, [isConnected, handleConnectionChange]);

  // Handle WebSocket errors
  useEffect(() => {
    if (wsError) {
      setError(new Error(wsError));
      if (isConnected) {
        reconnect();
      }
    }
  }, [wsError, isConnected, reconnect]);

  // Set initial connecting state when enabled
  useEffect(() => {
    if (enabled && connectionState === 'disconnected') {
      setConnectionState('connecting');
    }
  }, [enabled, connectionState]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      // Clear any buffered events
      eventBufferRef.current = [];
      isBufferingRef.current = false;
    };
  }, []);

  return {
    isConnected,
    connectionState,
    error,
    reconnect,
    lastEvent,
  };
}

// ============================================================================
// Specialized Hooks for Common Use Cases
// ============================================================================

/**
 * Hook specifically for trust alerts
 */
export function useTrustAlerts(
  onAlert: (alert: TrustAlert) => void,
  enabled?: boolean
): Pick<UseNotificationRealtimeReturn, 'isConnected' | 'connectionState' | 'error'> {
  const { isConnected, connectionState, error } = useNotificationRealtime({
    onTrustAlert: onAlert,
    enabled,
    categories: ['trust', 'security'],
  });

  return { isConnected, connectionState, error };
}

/**
 * Hook specifically for activity feed
 */
export function useActivityRealtime(
  onActivity: (activity: ActivityFeedItem) => void,
  enabled?: boolean
): Pick<UseNotificationRealtimeReturn, 'isConnected' | 'connectionState' | 'error' | 'lastEvent'> {
  const { isConnected, connectionState, error, lastEvent } = useNotificationRealtime({
    onActivity,
    enabled,
  });

  return { isConnected, connectionState, error, lastEvent };
}

/**
 * Hook for notification inbox with full functionality
 */
export function useNotificationInbox(options: {
  onNewNotification?: (notification: Notification) => void;
  onNotificationRead?: (notificationId: string) => void;
  onBulkUpdate?: (notificationIds: string[]) => void;
  enabled?: boolean;
}): UseNotificationRealtimeReturn {
  const { onNewNotification, onNotificationRead, onBulkUpdate, enabled } = options;

  return useNotificationRealtime({
    onNewNotification,
    onNotificationRead,
    onBulkUpdate,
    enabled,
  });
}

export default useNotificationRealtime;
