import { useCallback, useRef } from 'react';
import { useRealtimeSubscription, type RealtimeEvent } from './useRealtimeSubscription';

export function useRealtime(channel = 'admin_dashboard') {
  const callbacksRef = useRef<Set<(event: RealtimeEvent) => void>>(new Set());

  const { isConnected, error, sendMessage } = useRealtimeSubscription({
    channel,
    eventType: 'update',
    onEvent: (event) => {
      callbacksRef.current.forEach((callback) => {
        try {
          callback(event);
        } catch {
          // Ignore callback errors to keep broadcast loop alive
        }
      });
    },
  });

  const subscribe = useCallback((callback: (event: RealtimeEvent) => void) => {
    callbacksRef.current.add(callback);
    return () => {
      callbacksRef.current.delete(callback);
    };
  }, []);

  return {
    isConnected,
    error,
    subscribe,
    sendMessage,
  };
}
