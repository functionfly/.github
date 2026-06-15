import { apiClient } from './client';
import { tokenVault } from '@/utils/token-vault';
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
  private baseUrl: string;
  private intentionalClose = false;

  constructor() {
    this.baseUrl = getWebSocketUrl();
  }

  private getWebSocketUrl(token?: string): string {
    return token
      ? `${this.baseUrl}/v1/users/presence/ws?token=${encodeURIComponent(token)}`
      : `${this.baseUrl}/v1/users/presence/ws`;
  }

  async connect() {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return;
    }

    try {
      await tokenVault.initialize();
      const token = await tokenVault.getAccessToken();
      const url = this.getWebSocketUrl(token || undefined);
      this.ws = new WebSocket(url);

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
      void this.connect();
    }, delay);
  }

  on(event: string, callback: (data: PresenceSocketEvent) => void) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set());
    }
    this.listeners.get(event)!.add(callback);
  }

  off(event: string, callback: (data: PresenceSocketEvent) => void) {
    this.listeners.get(event)?.delete(callback);
  }

  private emit(event: string, data: PresenceSocketEvent) {
    this.listeners.get(event)?.forEach((cb) => cb(data));
  }
}

function getWebSocketUrl(): string {
  const base =
    getApiBaseUrl().startsWith('http://') || getApiBaseUrl().startsWith('https://')
      ? getApiBaseUrl()
      : `${typeof window !== 'undefined' ? window.location.origin : ''}${getApiBaseUrl()}`;
  return base.replace(/^http/, 'ws').replace(/\/$/, '');
}

export type PresenceSocketEvent =
  | { type: 'connected'; timestamp: string }
  | { type: 'disconnected'; timestamp: string }
  | { type: 'reconnecting'; attempt: number }
  | { type: 'failed'; error: string }
  | { type: 'error'; error: string }
  | { type: 'presence_update'; userId: string; status: PresenceStatus };
