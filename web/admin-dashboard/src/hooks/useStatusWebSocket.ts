import { useMemo } from 'react';
import { useRealtimeSubscription, type RealtimeEvent } from './useRealtimeSubscription';

interface UseStatusWebSocketOptions {
  enabled?: boolean;
  onStatusUpdate?: (event: RealtimeEvent) => void;
  onIncidentUpdate?: (event: RealtimeEvent) => void;
}

export function useStatusWebSocket(options: UseStatusWebSocketOptions = {}) {
  const { enabled = true, onStatusUpdate, onIncidentUpdate } = options;

  const statusSubscription = useRealtimeSubscription({
    channel: 'platform_status',
    eventType: 'status_update',
    onEvent: (event) => {
      if (enabled && onStatusUpdate) {
        onStatusUpdate(event);
      }
    },
  });

  const incidentSubscription = useRealtimeSubscription({
    channel: 'platform_incidents',
    eventType: 'incident_update',
    onEvent: (event) => {
      if (enabled && onIncidentUpdate) {
        onIncidentUpdate(event);
      }
    },
  });

  return useMemo(
    () => ({
      isConnected: statusSubscription.isConnected && incidentSubscription.isConnected,
      error: statusSubscription.error || incidentSubscription.error,
      sendStatusMessage: statusSubscription.sendMessage,
      sendIncidentMessage: incidentSubscription.sendMessage,
    }),
    [statusSubscription, incidentSubscription]
  );
}
