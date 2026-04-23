import { useQuery, useMutation, useQueryClient, useInfiniteQuery } from '@tanstack/react-query';
import { useCallback, useEffect, useState } from 'react';
import { toast } from 'sonner';
import {
  conversationsApi,
  type Conversation,
  type ConversationMessage,
  type ConversationType,
} from '@/api/conversations';
import { useConversationWebSocket } from './useConversationWebSocket';

// Query keys
export const conversationKeys = {
  all: ['conversations'] as const,
  lists: (params?: { limit?: number; offset?: number }) =>
    [...conversationKeys.all, 'list', params] as const,
  detail: (id: string) => [...conversationKeys.all, 'detail', id] as const,
  messages: (id: string, params?: { limit?: number; offset?: number }) =>
    [...conversationKeys.all, 'messages', id, params] as const,
  bounties: (id: string) => [...conversationKeys.all, 'bounties', id] as const,
};

// List conversations
export function useConversations(params?: { limit?: number; offset?: number }) {
  return useQuery({
    queryKey: conversationKeys.lists(params),
    queryFn: () => conversationsApi.listConversations(params),
    staleTime: 1000 * 30,
  });
}

// Get conversation
export function useConversation(id: string) {
  return useQuery({
    queryKey: conversationKeys.detail(id),
    queryFn: () => conversationsApi.getConversation(id),
    enabled: !!id,
    staleTime: 1000 * 30,
  });
}

// Create conversation
export function useCreateConversation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: {
      type?: ConversationType;
      participant_ids: string[];
      source_thread_id?: string;
      organization_id?: string;
    }) => conversationsApi.createConversation(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: conversationKeys.all });
      toast.success('Conversation created successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create conversation: ${error.message}`);
    },
  });
}

// Mark conversation as read
export function useMarkConversationRead() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (conversationId: string) => conversationsApi.markConversationRead(conversationId),
    onSuccess: (_, conversationId) => {
      queryClient.invalidateQueries({ queryKey: conversationKeys.detail(conversationId) });
      queryClient.invalidateQueries({ queryKey: conversationKeys.lists() });
    },
    onError: (error: Error) => {
      toast.error(`Failed to mark as read: ${error.message}`);
    },
  });
}

// List messages (initial fetch — no polling; real-time updates come via WebSocket)
export function useMessages(
  conversationId: string,
  params?: { limit?: number; offset?: number }
) {
  return useQuery({
    queryKey: conversationKeys.messages(conversationId, params),
    queryFn: () => conversationsApi.listMessages(conversationId, params),
    enabled: !!conversationId,
    staleTime: 1000 * 60, // Cache aggressively; WS pushes keep it fresh
  });
}

// Real-time messages hook — combines initial fetch with WebSocket push.
// Returns the same data shape as useMessages plus connection state.
export function useRealtimeMessages(
  conversationId: string,
  params?: { limit?: number; offset?: number }
) {
  const messagesQuery = useMessages(conversationId, params);
  const [typingUsers, setTypingUsers] = useState<Record<string, boolean>>({});

  const handleTyping = useCallback(
    (data: { user_id: string; conversation_id: string; typing: boolean }) => {
      if (data.conversation_id !== conversationId) return;
      setTypingUsers((prev) => {
        const next = { ...prev };
        if (data.typing) {
          next[data.user_id] = true;
        } else {
          delete next[data.user_id];
        }
        return next;
      });
    },
    [conversationId],
  );

  const ws = useConversationWebSocket({
    conversationIds: conversationId ? [conversationId] : [],
    enabled: !!conversationId,
    onTyping: handleTyping,
  });

  const sendTyping = useCallback(
    (typing: boolean) => {
      if (!conversationId) return;
      ws.send({
        type: 'typing',
        payload: { conversation_id: conversationId, typing },
      });
    },
    [ws, conversationId],
  );

  return {
    ...messagesQuery,
    isConnected: ws.isConnected,
    isConnecting: ws.isConnecting,
    wsError: ws.error,
    typingUsers,
    sendTyping,
    reconnect: ws.reconnect,
  };
}

// Create message
export function useCreateMessage() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      conversationId,
      content,
      embeddings,
    }: {
      conversationId: string;
      content: string;
      embeddings?: Record<string, unknown>;
    }) =>
      conversationsApi.createMessage(conversationId, {
        content,
        embeddings,
      }),
    onSuccess: (_, { conversationId }) => {
      queryClient.invalidateQueries({ queryKey: conversationKeys.messages(conversationId) });
      queryClient.invalidateQueries({ queryKey: conversationKeys.detail(conversationId) });
    },
    onError: (error: Error) => {
      toast.error(`Failed to send message: ${error.message}`);
    },
  });
}

// Resolve conversation
export function useResolveConversation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ conversationId, messageId }: { conversationId: string; messageId?: string }) =>
      conversationsApi.resolveConversation(conversationId, messageId),
    onSuccess: (_, { conversationId }) => {
      queryClient.invalidateQueries({ queryKey: conversationKeys.detail(conversationId) });
      toast.success('Conversation resolved');
    },
    onError: (error: Error) => {
      toast.error(`Failed to resolve conversation: ${error.message}`);
    },
  });
}

// List bounties
export function useBounties(conversationId: string) {
  return useQuery({
    queryKey: conversationKeys.bounties(conversationId),
    queryFn: () => conversationsApi.listBounties(conversationId),
    enabled: !!conversationId,
    staleTime: 1000 * 30,
  });
}

// Create bounty
export function useCreateBounty() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      conversationId,
      amountReputation,
      amountCents,
      securityWeightMultiplier,
    }: {
      conversationId: string;
      amountReputation: number;
      amountCents?: number;
      securityWeightMultiplier?: number;
    }) =>
      conversationsApi.createBounty(conversationId, {
        amount_reputation: amountReputation,
        amount_cents: amountCents,
        security_weight_multiplier: securityWeightMultiplier,
      }),
    onSuccess: (_, { conversationId }) => {
      queryClient.invalidateQueries({ queryKey: conversationKeys.bounties(conversationId) });
      toast.success('Bounty created successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create bounty: ${error.message}`);
    },
  });
}

// Claim bounty
export function useClaimBounty() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ conversationId, bountyId }: { conversationId: string; bountyId: string }) =>
      conversationsApi.claimBounty(conversationId, bountyId),
    onSuccess: (_, { conversationId }) => {
      queryClient.invalidateQueries({ queryKey: conversationKeys.bounties(conversationId) });
      toast.success('Bounty claimed successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to claim bounty: ${error.message}`);
    },
  });
}
