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
  edited_at?: string | null;
  deleted_at?: string | null;
  created_at: string;
  attachments?: MessageAttachment[];
}

export interface MessageAttachment {
  id: string;
  message_id: string;
  conversation_id: string;
  uploaded_by: string;
  filename: string;
  content_type: string;
  size_bytes: number;
  storage_url: string;
  thumbnail_url?: string | null;
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

export interface EditMessageRequest {
  content: string;
}

export interface SearchMessagesParams {
  q: string;
  conversation_id?: string;
  limit?: number;
  offset?: number;
}

export interface UploadAttachmentRequest {
  filename: string;
  content_type: string;
  size_bytes: number;
  storage_url: string;
  thumbnail_url?: string;
}

class ConversationsApi {
  async listConversations(
    username: string,
    params?: ListConversationsParams
  ): Promise<{ conversations: Conversation[] }> {
    const q = new URLSearchParams();
    if (params?.limit != null) q.set('limit', String(params.limit));
    if (params?.offset != null) q.set('offset', String(params.offset));
    const query = q.toString();
    return apiClient.get<{ conversations: Conversation[] }>(
      `/v1/u/${username}/conversations${query ? `?${query}` : ''}`
    );
  }

  async createConversation(username: string, body: CreateConversationRequest): Promise<Conversation> {
    return apiClient.post<Conversation>(`/v1/u/${username}/conversations`, body);
  }

  async getConversation(username: string, id: string): Promise<Conversation> {
    return apiClient.get<Conversation>(`/v1/u/${username}/conversations/${id}`);
  }

  async markConversationRead(username: string, conversationId: string): Promise<void> {
    await apiClient.post(`/v1/u/${username}/conversations/${conversationId}/read`, {});
  }

  async listMessages(
    username: string,
    conversationId: string,
    params?: ListMessagesParams
  ): Promise<{ messages: ConversationMessage[] }> {
    const q = new URLSearchParams();
    if (params?.limit != null) q.set('limit', String(params.limit));
    if (params?.offset != null) q.set('offset', String(params.offset));
    const query = q.toString();
    return apiClient.get<{ messages: ConversationMessage[] }>(
      `/v1/u/${username}/conversations/${conversationId}/messages${query ? `?${query}` : ''}`
    );
  }

  async getMessage(
    username: string,
    conversationId: string,
    messageId: string
  ): Promise<ConversationMessage> {
    return apiClient.get<ConversationMessage>(
      `/v1/u/${username}/conversations/${conversationId}/messages/${messageId}`
    );
  }

  async createMessage(
    username: string,
    conversationId: string,
    body: CreateMessageRequest
  ): Promise<ConversationMessage> {
    return apiClient.post<ConversationMessage>(
      `/v1/u/${username}/conversations/${conversationId}/messages`,
      body
    );
  }

  async resolveConversation(username: string, conversationId: string, messageId?: string): Promise<Conversation> {
    return apiClient.post<Conversation>(`/v1/u/${username}/conversations/${conversationId}/resolve`, {
      message_id: messageId || '',
    });
  }

  async listBounties(username: string, conversationId: string): Promise<{ bounties: ConversationBounty[] }> {
    return apiClient.get<{ bounties: ConversationBounty[] }>(
      `/v1/u/${username}/conversations/${conversationId}/bounties`
    );
  }

  async createBounty(
    username: string,
    conversationId: string,
    body: { amount_reputation: number; amount_cents?: number; security_weight_multiplier?: number }
  ): Promise<ConversationBounty> {
    return apiClient.post<ConversationBounty>(`/v1/u/${username}/conversations/${conversationId}/bounties`, body);
  }

  async claimBounty(username: string, conversationId: string, bountyId: string): Promise<ConversationBounty> {
    return apiClient.post<ConversationBounty>(
      `/v1/u/${username}/conversations/${conversationId}/bounties/${bountyId}/claim`,
      {}
    );
  }

  async editMessage(
    username: string,
    conversationId: string,
    messageId: string,
    body: EditMessageRequest
  ): Promise<ConversationMessage> {
    return apiClient.patch<ConversationMessage>(
      `/v1/u/${username}/conversations/${conversationId}/messages/${messageId}`,
      body
    );
  }

  async deleteMessage(username: string, conversationId: string, messageId: string): Promise<void> {
    await apiClient.delete(`/v1/u/${username}/conversations/${conversationId}/messages/${messageId}`);
  }

  async searchMessages(
    username: string,
    params: SearchMessagesParams
  ): Promise<{ messages: ConversationMessage[] }> {
    const q = new URLSearchParams();
    q.set('q', params.q);
    if (params.conversation_id) q.set('conversation_id', params.conversation_id);
    if (params.limit != null) q.set('limit', String(params.limit));
    if (params.offset != null) q.set('offset', String(params.offset));
    return apiClient.get<{ messages: ConversationMessage[] }>(
      `/v1/u/${username}/conversations/search?${q.toString()}`
    );
  }

  async uploadAttachment(
    username: string,
    conversationId: string,
    messageId: string,
    body: UploadAttachmentRequest
  ): Promise<MessageAttachment> {
    return apiClient.post<MessageAttachment>(
      `/v1/u/${username}/conversations/${conversationId}/messages/${messageId}/attachments`,
      body
    );
  }

  async deleteAttachment(
    username: string,
    conversationId: string,
    messageId: string,
    attachmentId: string
  ): Promise<void> {
    await apiClient.delete(
      `/v1/u/${username}/conversations/${conversationId}/messages/${messageId}/attachments/${attachmentId}`
    );
  }

  async listAttachments(
    username: string,
    conversationId: string,
    messageId: string
  ): Promise<{ attachments: MessageAttachment[] }> {
    return apiClient.get<{ attachments: MessageAttachment[] }>(
      `/v1/u/${username}/conversations/${conversationId}/messages/${messageId}/attachments`
    );
  }

  async getAttachment(
    username: string,
    conversationId: string,
    messageId: string,
    attachmentId: string
  ): Promise<MessageAttachment> {
    return apiClient.get<MessageAttachment>(
      `/v1/u/${username}/conversations/${conversationId}/messages/${messageId}/attachments/${attachmentId}`
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
