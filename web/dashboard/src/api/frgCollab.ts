import { useEffect, useState, useCallback, useRef } from 'react';
import { getApiBaseUrl } from '@/lib/constants';
import { tokenVault } from '@/utils/token-vault';

export interface CursorPosition {
  x: number;
  y: number;
  viewport_x: number;
  viewport_y: number;
  viewport_zoom: number;
  selected_node?: string;
}

export interface CollabParticipant {
  user_id: string;
  display_name: string;
  color: string;
  cursor?: CursorPosition;
  is_active: boolean;
  last_activity_at: string;
}

export interface CollabSessionState {
  type: 'session_state';
  graph_id: string;
  participants: CollabParticipant[];
  you: CollabParticipant;
  timestamp: number;
}

export interface CursorUpdateMessage {
  type: 'cursor_update';
  payload: {
    user_id: string;
    cursor: CursorPosition;
  };
  timestamp: number;
}

export interface ParticipantJoinedMessage {
  type: 'participant_joined';
  payload: CollabParticipant;
  timestamp: number;
}

export interface ParticipantLeftMessage {
  type: 'participant_left';
  payload: { user_id: string };
  timestamp: number;
}

export interface ViewportUpdateMessage {
  type: 'viewport_update';
  payload: {
    user_id: string;
    viewport_x: number;
    viewport_y: number;
    zoom: number;
  };
  timestamp: number;
}

export interface NodeSelectionMessage {
  type: 'node_selection';
  payload: {
    user_id: string;
    node_id: string;
    selected: boolean;
  };
  timestamp: number;
}

export type CollabMessage =
  | CollabSessionState
  | CursorUpdateMessage
  | ParticipantJoinedMessage
  | ParticipantLeftMessage
  | ViewportUpdateMessage
  | NodeSelectionMessage;

export type CollabEventType =
  | 'connected'
  | 'disconnected'
  | 'reconnecting'
  | 'failed'
  | 'cursor_update'
  | 'participant_joined'
  | 'participant_left'
  | 'viewport_update'
  | 'node_selection'
  | 'session_state';

export interface CollabEvent {
  type: CollabEventType;
  data?: CollabMessage;
  timestamp?: string;
}

export class FRGCollabWebSocket {
  private ws: WebSocket | null = null;
  private graphId: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private heartbeatInterval: ReturnType<typeof setInterval> | null = null;
  private listeners: Map<CollabEventType, Set<(event: CollabEvent) => void>> = new Map();
  private baseUrl: string;
  private intentionalClose = false;
  private cursorThrottle: ReturnType<typeof setTimeout> | null = null;
  private lastCursorSent = 0;
  private readonly CURSOR_THROTTLE_MS = 50;

  constructor(graphId: string) {
    this.graphId = graphId;
    this.baseUrl = getApiBaseUrl().replace(/^http/, 'ws').replace(/\/$/, '');
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  on(event: CollabEventType, callback: (event: CollabEvent) => void) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set());
    }
    this.listeners.get(event)!.add(callback);
  }

  off(event: CollabEventType, callback: (event: CollabEvent) => void) {
    this.listeners.get(event)?.delete(callback);
  }

  private emit(event: CollabEvent) {
    this.listeners.get(event.type)?.forEach((cb) => cb(event));
  }

  async connect() {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return;
    }

    try {
      await tokenVault.initialize();
      const token = await tokenVault.getAccessToken();
      const url = `${this.baseUrl}/v1/frg/${this.graphId}/collab/ws?token=${encodeURIComponent(token || '')}`;

      this.ws = new WebSocket(url);

      this.ws.onopen = () => {
        this.intentionalClose = false;
        this.reconnectAttempts = 0;
        this.startHeartbeat();
        this.emit({ type: 'connected', timestamp: new Date().toISOString() });
      };

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as CollabMessage;
          let eventType: CollabEventType;

          switch (data.type) {
            case 'session_state':
              eventType = 'session_state';
              break;
            case 'cursor_update':
              eventType = 'cursor_update';
              break;
            case 'participant_joined':
              eventType = 'participant_joined';
              break;
            case 'participant_left':
              eventType = 'participant_left';
              break;
            case 'viewport_update':
              eventType = 'viewport_update';
              break;
            case 'node_selection':
              eventType = 'node_selection';
              break;
            default:
              return;
          }

          this.emit({ type: eventType, data, timestamp: new Date().toISOString() });
        } catch (err) {
          console.error('Failed to parse FRG collab message:', err);
        }
      };

      this.ws.onclose = () => {
        this.stopHeartbeat();
        this.emit({ type: 'disconnected', timestamp: new Date().toISOString() });
        if (!this.intentionalClose) {
          this.attemptReconnect();
        }
      };

      this.ws.onerror = (error) => {
        console.error('FRG collab WebSocket error:', error);
        this.emit({ type: 'failed', timestamp: new Date().toISOString() });
      };
    } catch (error) {
      console.error('Failed to create FRG collab WebSocket:', error);
      this.attemptReconnect();
    }
  }

  disconnect() {
    this.intentionalClose = true;
    this.stopHeartbeat();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.reconnectAttempts = this.maxReconnectAttempts;
  }

  private attemptReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      this.emit({ type: 'failed', timestamp: new Date().toISOString() });
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
    this.emit({ type: 'reconnecting', timestamp: new Date().toISOString() });

    setTimeout(() => {
      void this.connect();
    }, delay);
  }

  private startHeartbeat() {
    this.stopHeartbeat();
    this.heartbeatInterval = setInterval(() => {
      this.sendHeartbeat();
    }, 30000);
  }

  private stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  private sendHeartbeat() {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type: 'heartbeat' }));
    }
  }

  sendCursorUpdate(cursor: CursorPosition) {
    const now = Date.now();
    if (now - this.lastCursorSent < this.CURSOR_THROTTLE_MS) {
      return;
    }
    this.lastCursorSent = now;

    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        type: 'cursor',
        ...cursor,
      }));
    }
  }

  sendViewportUpdate(viewportX: number, viewportY: number, zoom: number) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        type: 'viewport',
        viewport_x: viewportX,
        viewport_y: viewportY,
        viewport_zoom: zoom,
      }));
    }
  }

  sendNodeSelection(nodeId: string, selected: boolean) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        type: 'node_selection',
        node_id: nodeId,
        selected,
      }));
    }
  }
}

let _hmrDisposing = false;
if (import.meta.hot) {
  import.meta.hot.dispose(() => {
    _hmrDisposing = true;
  });
}

export interface UseFRGCollabOptions {
  graphId: string;
  enableWebSocket?: boolean;
}

export interface UseFRGCollabReturn {
  participants: Map<string, CollabParticipant>;
  isConnected: boolean;
  isLoading: boolean;
  error: string | null;
  updateCursor: (cursor: CursorPosition) => void;
  updateViewport: (x: number, y: number, zoom: number) => void;
  updateNodeSelection: (nodeId: string, selected: boolean) => void;
  connect: () => void;
  disconnect: () => void;
}

export function useFRGCollab(options: UseFRGCollabOptions): UseFRGCollabReturn {
  const { graphId, enableWebSocket = true } = options;

  const [participants, setParticipants] = useState<Map<string, CollabParticipant>>(new Map());
  const [isConnected, setIsConnected] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const wsRef = useRef<FRGCollabWebSocket | null>(null);

  const connect = useCallback(() => {
    if (wsRef.current?.isConnected() || !graphId) return;

    const ws = new FRGCollabWebSocket(graphId);
    wsRef.current = ws;

    ws.on('connected', () => {
      setIsConnected(true);
      setIsLoading(false);
      setError(null);
    });

    ws.on('disconnected', () => {
      setIsConnected(false);
    });

    ws.on('reconnecting', () => {
      setIsLoading(true);
    });

    ws.on('failed', () => {
      setIsConnected(false);
      setIsLoading(false);
      setError('Failed to connect to collaboration service');
    });

    ws.on('session_state', (event) => {
      if (event.data?.type === 'session_state') {
        const sessionData = event.data as CollabSessionState;
        const newParticipants = new Map<string, CollabParticipant>();
        for (const p of sessionData.participants) {
          newParticipants.set(p.user_id, p);
        }
        setParticipants(newParticipants);
      }
    });

    ws.on('participant_joined', (event) => {
      if (event.data?.type === 'participant_joined') {
        const msg = event.data as ParticipantJoinedMessage;
        setParticipants((prev) => {
          const next = new Map(prev);
          next.set(msg.payload.user_id, msg.payload);
          return next;
        });
      }
    });

    ws.on('participant_left', (event) => {
      if (event.data?.type === 'participant_left') {
        const msg = event.data as ParticipantLeftMessage;
        setParticipants((prev) => {
          const next = new Map(prev);
          next.delete(msg.payload.user_id);
          return next;
        });
      }
    });

    ws.on('cursor_update', (event) => {
      if (event.data?.type === 'cursor_update') {
        const msg = event.data as CursorUpdateMessage;
        setParticipants((prev) => {
          const next = new Map(prev);
          const existing = next.get(msg.payload.user_id);
          if (existing) {
            next.set(msg.payload.user_id, {
              ...existing,
              cursor: msg.payload.cursor,
              is_active: true,
            });
          }
          return next;
        });
      }
    });

    void ws.connect();
  }, [graphId]);

  const disconnect = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.disconnect();
      wsRef.current = null;
    }
    setParticipants(new Map());
    setIsConnected(false);
  }, []);

  const updateCursor = useCallback((cursor: CursorPosition) => {
    wsRef.current?.sendCursorUpdate(cursor);
  }, []);

  const updateViewport = useCallback((x: number, y: number, zoom: number) => {
    wsRef.current?.sendViewportUpdate(x, y, zoom);
  }, []);

  const updateNodeSelection = useCallback((nodeId: string, selected: boolean) => {
    wsRef.current?.sendNodeSelection(nodeId, selected);
  }, []);

  useEffect(() => {
    if (enableWebSocket && graphId) {
      connect();
    }

    return () => {
      if (_hmrDisposing) {
        return;
      }
      disconnect();
    };
  }, [graphId, enableWebSocket, connect, disconnect]);

  return {
    participants,
    isConnected,
    isLoading,
    error,
    updateCursor,
    updateViewport,
    updateNodeSelection,
    connect,
    disconnect,
  };
}
