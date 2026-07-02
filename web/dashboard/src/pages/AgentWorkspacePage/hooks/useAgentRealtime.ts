import { useCallback, useEffect, useRef, useState } from 'react';
import { apiClient } from '@/api/client';

export interface RealtimeEvent {
  id: string;
  kind: string;
  sequence: number;
  timestamp: string;
  system_id: string;
  payload: unknown;
  cost_usd?: number;
  tokens_in?: number;
  tokens_out?: number;
  tool_name?: string;
  reasoning?: string;
}

interface UseAgentRealtimeOptions {
  agentId: string;
  enabled?: boolean;
  onEvent?: (event: RealtimeEvent) => void;
}

export function useAgentRealtime({ agentId, enabled = true, onEvent }: UseAgentRealtimeOptions) {
  const [events, setEvents] = useState<RealtimeEvent[]>([]);
  const [connected, setConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const onEventRef = useRef(onEvent);
  onEventRef.current = onEvent;

  const connect = useCallback(() => {
    if (!enabled || !agentId) return;

    const token = apiClient.getToken();
    if (!token) {
      reconnectRef.current = setTimeout(connect, 3000);
      return;
    }

    try {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const host = window.location.host;
      const ws = new WebSocket(`${protocol}//${host}/api/v1/agent-observability/agents/${agentId}/stream?token=${token}`);

      ws.onopen = () => {
        setConnected(true);
      };

      ws.onmessage = (msg) => {
        try {
          const event = JSON.parse(msg.data) as RealtimeEvent;
          setEvents(prev => [...prev.slice(-499), event]);
          onEventRef.current?.(event);
        } catch {
          // ignore parse errors
        }
      };

      ws.onclose = () => {
        setConnected(false);
        reconnectRef.current = setTimeout(connect, 3000);
      };

      ws.onerror = () => {
        ws.close();
      };

      wsRef.current = ws;
    } catch {
      setConnected(false);
      reconnectRef.current = setTimeout(connect, 3000);
    }
  }, [agentId, enabled]);

  useEffect(() => {
    connect();
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
      if (reconnectRef.current) {
        clearTimeout(reconnectRef.current);
      }
    };
  }, [connect]);

  const reconnect = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.close();
    }
    connect();
  }, [connect]);

  const clearEvents = useCallback(() => {
    setEvents([]);
  }, []);

  return {
    events,
    connected,
    reconnect,
    clearEvents,
  };
}
