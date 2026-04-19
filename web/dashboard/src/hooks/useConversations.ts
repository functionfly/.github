import { useQuery, useMutation, useQueryClient, useInfiniteQuery } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  conversationsApi,
  type Conversation,
  type ConversationMessage,
  type ConversationType,
} from '@/api/conversations';

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

// List messages
export function useMessages(
  conversationId: string,
  params?: { limit?: number; offset?: number }
) {
  return useQuery({
    queryKey: conversationKeys.messages(conversationId, params),
    queryFn: () => conversationsApi.listMessages(conversationId, params),
    enabled: !!conversationId,
    staleTime: 1000 * 10,
    refetchInterval: 10000, // Poll for new messages every 10s
  });
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
