import apiClient from './client';

export interface AIChatSession {
  id: string;
  user_id: string;
  title: string;
  context_type: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface AIChatMessage {
  id: string;
  session_id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  tokens_used: number;
  model?: string;
  created_at: string;
}

export const aiChatApi = {
  listSessions: () => apiClient.get<{ sessions: AIChatSession[] }>('/v1/ai-chat/sessions'),
  getSession: (id: string) => apiClient.get<{ session: AIChatSession; messages: AIChatMessage[] }>(`/v1/ai-chat/sessions/${id}`),
  createSession: (data?: { title?: string; context_type?: string }) => apiClient.post<{ session: AIChatSession }>('/v1/ai-chat/sessions', data),
  sendMessage: (sessionId: string, message: string) => apiClient.post<{ message: AIChatMessage; reply: AIChatMessage }>(`/v1/ai-chat/sessions/${sessionId}/messages`, { message }),
  deleteSession: (id: string) => apiClient.delete(`/v1/ai-chat/sessions/${id}`),
};
