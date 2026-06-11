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
import { useAuthStore } from '@/stores/authStore';

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
  const currentUser = useAuthStore((state) => state.user);
  return useQuery({
    queryKey: conversationKeys.lists(params),
    queryFn: () => conversationsApi.listConversations(currentUser?.username || '', params),
    staleTime: 1000 * 30,
  });
}

// Get conversation
export function useConversation(id: string) {
  const currentUser = useAuthStore((state) => state.user);
  return useQuery({
    queryKey: conversationKeys.detail(id),
    queryFn: () => conversationsApi.getConversation(currentUser?.username || '', id),
    enabled: !!id,
    staleTime: 1000 * 30,
  });
}

// Create conversation
export function useCreateConversation() {
  const queryClient = useQueryClient();
  const currentUser = useAuthStore((state) => state.user);

  return useMutation({
    mutationFn: (data: {
      type?: ConversationType;
      participant_ids: string[];
      source_thread_id?: string;
      organization_id?: string;
    }) => conversationsApi.createConversation(currentUser?.username || '', data),
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
  const currentUser = useAuthStore((state) => state.user);

  return useMutation({
    mutationFn: (conversationId: string) => conversationsApi.markConversationRead(currentUser?.username || '', conversationId),
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
  const currentUser = useAuthStore((state) => state.user);
  return useQuery({
    queryKey: conversationKeys.messages(conversationId, params),
    queryFn: () => conversationsApi.listMessages(currentUser?.username || '', conversationId, params),
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
  const currentUser = useAuthStore((state) => state.user);

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
      conversationsApi.createMessage(currentUser?.username || '', conversationId, {
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
  const currentUser = useAuthStore((state) => state.user);

  return useMutation({
    mutationFn: ({ conversationId, messageId }: { conversationId: string; messageId?: string }) =>
      conversationsApi.resolveConversation(currentUser?.username || '', conversationId, messageId),
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
  const currentUser = useAuthStore((state) => state.user);
  return useQuery({
    queryKey: conversationKeys.bounties(conversationId),
    queryFn: () => conversationsApi.listBounties(currentUser?.username || '', conversationId),
    enabled: !!conversationId,
    staleTime: 1000 * 30,
  });
}

// Create bounty
export function useCreateBounty() {
  const queryClient = useQueryClient();
  const currentUser = useAuthStore((state) => state.user);

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
      conversationsApi.createBounty(currentUser?.username || '', conversationId, {
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
  const currentUser = useAuthStore((state) => state.user);

  return useMutation({
    mutationFn: ({ conversationId, bountyId }: { conversationId: string; bountyId: string }) =>
      conversationsApi.claimBounty(currentUser?.username || '', conversationId, bountyId),
    onSuccess: (_, { conversationId }) => {
      queryClient.invalidateQueries({ queryKey: conversationKeys.bounties(conversationId) });
      toast.success('Bounty claimed successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to claim bounty: ${error.message}`);
    },
  });
}

// Add reaction
export function useAddReaction() {
  const queryClient = useQueryClient();
  const currentUser = useAuthStore((state) => state.user);

  return useMutation({
    mutationFn: ({ conversationId, messageId, reaction }: { conversationId: string; messageId: string; reaction: string }) =>
      conversationsApi.addReaction(currentUser?.username || '', conversationId, messageId, reaction),
    onSuccess: (_, { conversationId }) => {
      queryClient.invalidateQueries({ queryKey: conversationKeys.messages(conversationId) });
    },
    onError: (error: Error) => {
      toast.error(`Failed to add reaction: ${error.message}`);
    },
  });
}

// Remove reaction
export function useRemoveReaction() {
  const queryClient = useQueryClient();
  const currentUser = useAuthStore((state) => state.user);

  return useMutation({
    mutationFn: ({ conversationId, messageId, reaction }: { conversationId: string; messageId: string; reaction: string }) =>
      conversationsApi.removeReaction(currentUser?.username || '', conversationId, messageId, reaction),
    onSuccess: (_, { conversationId }) => {
      queryClient.invalidateQueries({ queryKey: conversationKeys.messages(conversationId) });
    },
    onError: (error: Error) => {
      toast.error(`Failed to remove reaction: ${error.message}`);
    },
  });
}

// Mark message as read
export function useMarkMessageRead() {
  const queryClient = useQueryClient();
  const currentUser = useAuthStore((state) => state.user);

  return useMutation({
    mutationFn: ({ conversationId, messageId }: { conversationId: string; messageId: string }) =>
      conversationsApi.markMessageRead(currentUser?.username || '', conversationId, messageId),
    onSuccess: (_, { messageId }) => {
      queryClient.invalidateQueries({ queryKey: conversationKeys.messages(messageId) });
    },
    onError: (error: Error) => {
      toast.error(`Failed to mark message as read: ${error.message}`);
    },
  });
}

// Get message read receipts
export function useMessageReadReceipts(conversationId: string, messageId: string) {
  const currentUser = useAuthStore((state) => state.user);
  return useQuery({
    queryKey: [...conversationKeys.messages(conversationId), 'read-receipts', messageId],
    queryFn: () => conversationsApi.getMessageReadReceipts(currentUser?.username || '', conversationId, messageId),
    enabled: !!conversationId && !!messageId,
  });
}
