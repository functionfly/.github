import { apiClient } from './client';

export interface ChatSession {
  id: string;
  tenant_id: string;
  user_id: string;
  title: string;
  model: string;
  connector_ids: string[];
  metadata: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface ChatMessage {
  id: string;
  session_id: string;
  role: 'user' | 'assistant' | 'system' | 'function';
  content: string;
  attachments: string[];
  model?: string;
  tokens_used?: number;
  latency_ms?: number;
  metadata: Record<string, unknown>;
  created_at: string;
}

export interface ChatConnector {
  id: string;
  tenant_id: string;
  user_id: string;
  name: string;
  type: string;
  icon: string;
  config: Record<string, unknown>;
  is_active: boolean;
  last_tested_at?: string;
  last_test_success?: boolean;
  created_at: string;
}

export interface AIModel {
  id: string;
  name: string;
  provider: string;
}

class ChatApi {
  async createSession(title: string, model?: string, connectorIds?: string[]): Promise<ChatSession> {
    return apiClient.post<ChatSession>('/v1/chat/sessions', { title, model, connector_ids: connectorIds });
  }

  async listSessions(limit = 20, offset = 0): Promise<{ sessions: ChatSession[] }> {
    return apiClient.get<{ sessions: ChatSession[] }>(`/v1/chat/sessions?limit=${limit}&offset=${offset}`);
  }

  async getSession(id: string): Promise<{ session: ChatSession; messages: ChatMessage[] }> {
    return apiClient.get<{ session: ChatSession; messages: ChatMessage[] }>(`/v1/chat/sessions/${id}`);
  }

  async deleteSession(id: string): Promise<void> {
    await apiClient.delete(`/v1/chat/sessions/${id}`);
  }

  async updateSession(id: string, updates: { title?: string; model?: string }): Promise<void> {
    await apiClient.patch(`/v1/chat/sessions/${id}`, updates);
  }

  async sendMessage(sessionId: string, content: string, stream = false): Promise<{
    message: ChatMessage;
    token_usage: number;
    latency_ms: number;
  }> {
    return apiClient.post(`/v1/chat/messages`, { session_id: sessionId, content, stream });
  }

  async listModels(): Promise<{ models: AIModel[] }> {
    return apiClient.get<{ models: AIModel[] }>('/v1/chat/models');
  }

  async listConnectors(): Promise<{ connectors: ChatConnector[] }> {
    return apiClient.get<{ connectors: ChatConnector[] }>('/v1/chat/connectors');
  }

  async registerConnector(name: string, type: string, config: Record<string, unknown>): Promise<ChatConnector> {
    return apiClient.post<ChatConnector>('/v1/chat/connectors', { name, type, config });
  }

  async deleteConnector(id: string): Promise<void> {
    await apiClient.delete(`/v1/chat/connectors/${id}`);
  }

  async testConnector(id: string, credentials: Record<string, string>): Promise<{ success: boolean; error?: string }> {
    return apiClient.post(`/v1/chat/connectors/${id}/test`, { credentials });
  }
}

export const chatApi = new ChatApi();