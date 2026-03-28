import { apiClient } from './client';

export type ConversationType =
  | 'dm'
  | 'function_thread'
  | 'issue_thread'
  | 'fix_mode'
  | 'bounty_thread'
  | 'org_thread'
  | 'security_disclosure';

export interface MessageEmbeddings {
  function_ref?: { author: string; name: string; version?: string };
  execution_id?: string;
  execution_root_hash?: string;
  replay_link?: string;
  capability_declaration?: Record<string, unknown>;
  input_summary?: string;
  output_summary?: string;
}

export interface Conversation {
  id: string;
  type: ConversationType;
  participant_ids: string[];
  source_thread_id?: string | null;
  organization_id?: string | null;
  metadata: Record<string, unknown>;
  resolved_at?: string | null;
  resolved_by_user_id?: string | null;
  resolved_by_message_id?: string | null;
  created_at: string;
  updated_at: string;
  /** Present on list responses; messages from others since last read. */
  unread_count?: number;
}

export interface ConversationMessage {
  id: string;
  conversation_id: string;
  author_id: string;
  content: string;
  embeddings: MessageEmbeddings | Record<string, unknown>;
  created_at: string;
}

export interface ListConversationsParams {
  limit?: number;
  offset?: number;
}

export interface ListMessagesParams {
  limit?: number;
  offset?: number;
}

export interface CreateConversationRequest {
  type?: ConversationType;
  participant_ids: string[];
  source_thread_id?: string;
  organization_id?: string;
}

export interface CreateMessageRequest {
  content: string;
  embeddings?: MessageEmbeddings | Record<string, unknown>;
}

class ConversationsApi {
  async listConversations(
    params?: ListConversationsParams
  ): Promise<{ conversations: Conversation[] }> {
    const q = new URLSearchParams();
    if (params?.limit != null) q.set('limit', String(params.limit));
    if (params?.offset != null) q.set('offset', String(params.offset));
    const query = q.toString();
    return apiClient.get<{ conversations: Conversation[] }>(
      `/v1/conversations${query ? `?${query}` : ''}`
    );
  }

  async createConversation(body: CreateConversationRequest): Promise<Conversation> {
    return apiClient.post<Conversation>('/v1/conversations', body);
  }

  async getConversation(id: string): Promise<Conversation> {
    return apiClient.get<Conversation>(`/v1/conversations/${id}`);
  }

  async markConversationRead(conversationId: string): Promise<void> {
    await apiClient.post(`/v1/conversations/${conversationId}/read`, {});
  }

  async listMessages(
    conversationId: string,
    params?: ListMessagesParams
  ): Promise<{ messages: ConversationMessage[] }> {
    const q = new URLSearchParams();
    if (params?.limit != null) q.set('limit', String(params.limit));
    if (params?.offset != null) q.set('offset', String(params.offset));
    const query = q.toString();
    return apiClient.get<{ messages: ConversationMessage[] }>(
      `/v1/conversations/${conversationId}/messages${query ? `?${query}` : ''}`
    );
  }

  async createMessage(
    conversationId: string,
    body: CreateMessageRequest
  ): Promise<ConversationMessage> {
    return apiClient.post<ConversationMessage>(
      `/v1/conversations/${conversationId}/messages`,
      body
    );
  }

  async resolveConversation(conversationId: string, messageId?: string): Promise<Conversation> {
    return apiClient.post<Conversation>(`/v1/conversations/${conversationId}/resolve`, {
      message_id: messageId || '',
    });
  }

  async listBounties(conversationId: string): Promise<{ bounties: ConversationBounty[] }> {
    return apiClient.get<{ bounties: ConversationBounty[] }>(
      `/v1/conversations/${conversationId}/bounties`
    );
  }

  async createBounty(
    conversationId: string,
    body: { amount_reputation: number; amount_cents?: number; security_weight_multiplier?: number }
  ): Promise<ConversationBounty> {
    return apiClient.post<ConversationBounty>(`/v1/conversations/${conversationId}/bounties`, body);
  }

  async claimBounty(conversationId: string, bountyId: string): Promise<ConversationBounty> {
    return apiClient.post<ConversationBounty>(
      `/v1/conversations/${conversationId}/bounties/${bountyId}/claim`,
      {}
    );
  }
}

export interface ConversationBounty {
  id: string;
  conversation_id: string;
  offered_by: string;
  amount_reputation: number;
  amount_cents: number;
  security_weight_multiplier: number;
  claimed_by?: string | null;
  claimed_at?: string | null;
  created_at: string;
}

export const conversationsApi = new ConversationsApi();
