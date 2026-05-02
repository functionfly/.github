import { apiClient } from './client';
import { getApiBaseUrl } from '@/lib/constants';

export type PresenceStatus = 'online' | 'away' | 'offline';

export interface UserPresence {
  userId: string;
  status: PresenceStatus;
  lastActive: string;
  tenantId?: string;
  username?: string;
  displayName?: string;
  avatar?: string;
}

export interface PresenceResponse {
  users: UserPresence[];
  count: number;
  updated: string;
}

export interface MyPresenceResponse {
  userId: string;
  status: PresenceStatus;
  lastActive: string;
  tenantId?: string;
  username?: string;
  name?: string;
}

export interface PresenceUpdateResponse {
  status: 'ok';
  updated: string;
}

export const presenceApi = {
  getPresence: () =>
    apiClient.get<PresenceResponse>('/v1/users/presence'),

  getOnlineUsers: () =>
    apiClient.get<PresenceResponse>('/v1/users/presence/online'),

  getPresenceByIds: (userIds: string[]) =>
    apiClient.get<PresenceResponse>(`/v1/users/presence/users?user_ids=${userIds.join(',')}`),

  getMyPresence: () =>
    apiClient.get<MyPresenceResponse>('/v1/users/presence/me'),

  updateMyPresence: () =>
    apiClient.post<PresenceUpdateResponse>('/v1/users/presence/me'),
};

export class PresenceWebSocket {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private heartbeatInterval: ReturnType<typeof setInterval> | null = null;
  private listeners: Map<string, Set<(data: PresenceSocketEvent) => void>> = new Map();
  private url: string;
  private intentionalClose = false;

  constructor() {
    const baseUrl = getWebSocketUrl();
    const token = localStorage.getItem('ff-access-token');
    this.url = token
      ? `${baseUrl}/v1/users/presence/ws?token=${encodeURIComponent(token)}`
      : `${baseUrl}/v1/users/presence/ws`;
  }

  connect() {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return;
    }

    try {
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        this.intentionalClose = false;
        this.reconnectAttempts = 0;
        this.startHeartbeat();
        this.emit('connected', { type: 'connected', timestamp: new Date().toISOString() });
      };

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as PresenceSocketEvent;
          this.emit(data.type, data);
        } catch (err) {
          console.error('Failed to parse presence WebSocket message:', err);
        }
      };

      this.ws.onclose = () => {
        this.stopHeartbeat();
        this.emit('disconnected', { type: 'disconnected', timestamp: new Date().toISOString() });
        if (!this.intentionalClose) {
          this.attemptReconnect();
        }
      };

      this.ws.onerror = (error) => {
        console.error('Presence WebSocket error:', error);
        this.emit('error', { type: 'error', error: String(error) });
      };
    } catch (error) {
      console.error('Failed to create presence WebSocket:', error);
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

  sendHeartbeat() {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type: 'heartbeat' }));
    }
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

  private attemptReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      this.emit('failed', { type: 'failed', error: 'Max reconnect attempts reached' });
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

    setTimeout(() => {
      this.emit('reconnecting', { type: 'reconnecting', attempt: this.reconnectAttempts });
      this.connect();
    }, delay);
  }

  on(event: string, callback: (data: PresenceSocketEvent) => void) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set());
    }
    this.listeners.get(event)!.add(callback);
    return () => {
      this.listeners.get(event)?.delete(callback);
    };
  }

  private emit(event: string, data: PresenceSocketEvent) {
    this.listeners.get(event)?.forEach(cb => cb(data));
    this.listeners.get('*')?.forEach(cb => cb(data));
  }

  isConnected() {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

function getWebSocketUrl(): string {
  const base = getApiBaseUrl() || window.location.origin;
  if (base.startsWith('http://')) {
    return base.replace('http://', 'ws://');
  } else if (base.startsWith('https://')) {
    return base.replace('https://', 'wss://');
  }
  return `ws://${base}`;
}

export type PresenceSocketEvent =
  | { type: 'connected'; timestamp: string }
  | { type: 'disconnected'; timestamp: string }
  | { type: 'reconnecting'; attempt: number }
  | { type: 'failed'; error: string }
  | { type: 'error'; error: string }
  | { type: 'presence_join'; userId: string; status: PresenceStatus; heartbeat?: PresenceHeartbeatData }
  | { type: 'presence_leave'; userId: string; status: PresenceStatus }
  | { type: 'presence_update'; userId: string; status: PresenceStatus; heartbeat?: PresenceHeartbeatData }
  | { type: 'pong'; timestamp: string };

export interface PresenceHeartbeatData {
  userId: string;
  tenantId: string;
  username?: string;
  activeAt: string;
  lastActive: string;
}
