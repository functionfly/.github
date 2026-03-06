/**
 * Flywheel Network™ API Client
 *
 * React Query hooks for threads, replies, reputation, challenges, and WebSocket
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from './client';
import { getApiBaseUrl } from '@/lib/constants';
import type {
  Thread,
  ThreadFilters,
  Reply,
  ReputationProfile,
  Leaderboard,
  Challenge,
  ChallengeFilters,
  ChallengeSubmission,
  ExecutionResults,
  WebSocketMessage,
  ThreadUpdatePayload,
  ReplyAddedPayload,
  ExecutionCompletePayload,
  ReputationChangePayload,
  Pagination,
  Category,
} from '@/components/flywheel/types';

// ==================== API Endpoints ====================

const ENDPOINTS = {
  threads: '/v1/flywheel/threads',
  replies: '/v1/flywheel/replies',
  reputation: '/v1/flywheel/reputation',
  challenges: '/v1/flywheel/challenges',
  leaderboard: '/v1/flywheel/leaderboards',
  categories: '/v1/flywheel/categories',
  executions: '/v1/flywheel/executions',
  ws: '/v1/flywheel/ws',
} as const;

// ==================== Thread API ====================

interface ThreadsResponse {
  threads: Thread[];
  pagination: Pagination;
}

interface ThreadResponse {
  thread: Thread;
}

interface CreateThreadRequest {
  title: string;
  type: 'problem' | 'discussion' | 'challenge';
  categoryId: string;
  tags?: string[];
  problemData?: {
    description: string;
    constraints?: {
      timeComplexity?: string;
      spaceComplexity?: string;
    };
    testCases: Array<{
      name: string;
      description?: string;
      input: string;
      expectedOutput: string;
      isPublic: boolean;
    }>;
  };
  environmentSpecs?: {
    runtime: string;
    runtimeVersion: string;
    timeoutMs: number;
    memoryMb: number;
  };
}

export function useThreads(filters?: ThreadFilters, limit = 20, offset = 0) {
  return useQuery({
    queryKey: ['flywheel', 'threads', filters, limit, offset],
    queryFn: async (): Promise<ThreadsResponse> => {
      const params = new URLSearchParams();
      params.append('limit', String(limit));
      params.append('offset', String(offset));

      if (filters?.type) params.append('type', filters.type);
      if (filters?.status) params.append('status', filters.status);
      if (filters?.category) params.append('category', filters.category);
      if (filters?.search) params.append('search', filters.search);
      if (filters?.sortBy) params.append('sort_by', filters.sortBy);
      filters?.tags?.forEach(tag => params.append('tags', tag));

      return apiClient.get<ThreadsResponse>(`${ENDPOINTS.threads}?${params.toString()}`);
    },
    staleTime: 1000 * 60 * 2, // 2 minutes
  });
}

export function useThread(threadId: string | undefined) {
  return useQuery({
    queryKey: ['flywheel', 'thread', threadId],
    queryFn: async (): Promise<ThreadResponse> => {
      if (!threadId) throw new Error('Thread ID is required');
      return apiClient.get<ThreadResponse>(`${ENDPOINTS.threads}/${threadId}`);
    },
    enabled: !!threadId,
    staleTime: 1000 * 60, // 1 minute
  });
}

export function useCreateThread() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreateThreadRequest): Promise<ThreadResponse> => {
      return apiClient.post<ThreadResponse>(ENDPOINTS.threads, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['flywheel', 'threads'] });
    },
  });
}

export function useUpdateThread(threadId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: Partial<CreateThreadRequest>): Promise<ThreadResponse> => {
      return apiClient.patch<ThreadResponse>(`${ENDPOINTS.threads}/${threadId}`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['flywheel', 'thread', threadId] });
      queryClient.invalidateQueries({ queryKey: ['flywheel', 'threads'] });
    },
  });
}

export function useDeleteThread() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (threadId: string): Promise<void> => {
      return apiClient.delete<void>(`${ENDPOINTS.threads}/${threadId}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['flywheel', 'threads'] });
    },
  });
}

// ==================== Reply API ====================

interface RepliesResponse {
  replies: Reply[];
  pagination: Pagination;
}

interface CreateReplyRequest {
  threadId: string;
  content: string;
  codeBlocks?: Array<{
    language: string;
    code: string;
    filename?: string;
  }>;
  attachedCapsuleId?: string;
}

export function useReplies(threadId: string | undefined, limit = 50, offset = 0) {
  return useQuery({
    queryKey: ['flywheel', 'replies', threadId, limit, offset],
    queryFn: async (): Promise<RepliesResponse> => {
      if (!threadId) throw new Error('Thread ID is required');
      const params = new URLSearchParams();
      params.append('limit', String(limit));
      params.append('offset', String(offset));

      return apiClient.get<RepliesResponse>(
        `${ENDPOINTS.threads}/${threadId}/replies?${params.toString()}`
      );
    },
    enabled: !!threadId,
    staleTime: 1000 * 30, // 30 seconds
  });
}

export function useCreateReply() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreateReplyRequest): Promise<{ reply: Reply }> => {
      return apiClient.post<{ reply: Reply }>(
        `${ENDPOINTS.threads}/${data.threadId}/replies`,
        data
      );
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: ['flywheel', 'replies', variables.threadId]
      });
      queryClient.invalidateQueries({
        queryKey: ['flywheel', 'thread', variables.threadId]
      });
    },
  });
}

export function useMarkReplyHelpful() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ threadId, replyId }: { threadId: string; replyId: string }): Promise<void> => {
      return apiClient.post<void>(`${ENDPOINTS.replies}/${replyId}/helpful`);
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: ['flywheel', 'replies', variables.threadId]
      });
    },
  });
}

export function useAcceptSolution() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ threadId, replyId }: { threadId: string; replyId: string }): Promise<void> => {
      return apiClient.post<void>(`${ENDPOINTS.threads}/${threadId}/accept`, { replyId });
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: ['flywheel', 'thread', variables.threadId]
      });
      queryClient.invalidateQueries({
        queryKey: ['flywheel', 'replies', variables.threadId]
      });
    },
  });
}

// ==================== Execution API ====================

interface ExecuteRequest {
  replyId: string;
  code: string;
  language: string;
  testCases?: string[];
}

export function useExecuteReply() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: ExecuteRequest): Promise<{ executionId: string }> => {
      return apiClient.post<{ executionId: string }>(`${ENDPOINTS.executions}`, data);
    },
    onSuccess: () => {
      // Execution results will come via WebSocket
    },
  });
}

export function useExecutionResults(executionId: string | undefined) {
  return useQuery({
    queryKey: ['flywheel', 'execution', executionId],
    queryFn: async (): Promise<{ results: ExecutionResults }> => {
      if (!executionId) throw new Error('Execution ID is required');
      return apiClient.get<{ results: ExecutionResults }>(`${ENDPOINTS.executions}/${executionId}`);
    },
    enabled: !!executionId,
    refetchInterval: (query) => {
      const data = query.state.data;
      if (!data) return 1000;
      const status = data.results?.status;
      if (status === 'running' || status === 'queued' || status === 'pending') {
        return 1000; // Poll every second while running
      }
      return false;
    },
  });
}

export function useVerifySolution() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (replyId: string): Promise<{ verified: boolean }> => {
      return apiClient.post<{ verified: boolean }>(`${ENDPOINTS.replies}/${replyId}/verify`);
    },
    onSuccess: () => {
      // Invalidation handled by WebSocket
    },
  });
}

// ==================== Reputation API ====================

export function useReputation(userId: string | undefined) {
  return useQuery({
    queryKey: ['flywheel', 'reputation', userId],
    queryFn: async (): Promise<{ profile: ReputationProfile }> => {
      if (!userId) throw new Error('User ID is required');
      return apiClient.get<{ profile: ReputationProfile }>(`${ENDPOINTS.reputation}/${userId}`);
    },
    enabled: !!userId,
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}

export function useMyReputation() {
  return useQuery({
    queryKey: ['flywheel', 'reputation', 'me'],
    queryFn: async (): Promise<{ profile: ReputationProfile }> => {
      return apiClient.get<{ profile: ReputationProfile }>(`${ENDPOINTS.reputation}/me`);
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}

// ==================== Leaderboard API ====================

interface LeaderboardParams {
  type?: 'overall' | 'builder' | 'optimizer' | 'mentor' | 'agent_whisperer';
  timeframe?: 'daily' | 'weekly' | 'monthly' | 'all_time';
  limit?: number;
  offset?: number;
}

export function useLeaderboard(params: LeaderboardParams = {}) {
  const { type = 'overall', timeframe = 'all_time', limit = 100, offset = 0 } = params;

  return useQuery({
    queryKey: ['flywheel', 'leaderboard', type, timeframe, limit, offset],
    queryFn: async (): Promise<Leaderboard> => {
      const searchParams = new URLSearchParams();
      searchParams.append('type', type);
      searchParams.append('timeframe', timeframe);
      searchParams.append('limit', String(limit));
      searchParams.append('offset', String(offset));

      return apiClient.get<Leaderboard>(`${ENDPOINTS.leaderboard}?${searchParams.toString()}`);
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}

// ==================== Challenge API ====================

interface ChallengesResponse {
  challenges: Challenge[];
  pagination: Pagination;
}

export function useChallenges(filters?: ChallengeFilters, limit = 20, offset = 0) {
  return useQuery({
    queryKey: ['flywheel', 'challenges', filters, limit, offset],
    queryFn: async (): Promise<ChallengesResponse> => {
      const params = new URLSearchParams();
      params.append('limit', String(limit));
      params.append('offset', String(offset));

      if (filters?.type) params.append('type', filters.type);
      if (filters?.status) params.append('status', filters.status);
      if (filters?.timeframe) params.append('timeframe', filters.timeframe);

      return apiClient.get<ChallengesResponse>(`${ENDPOINTS.challenges}?${params.toString()}`);
    },
    staleTime: 1000 * 60 * 2, // 2 minutes
  });
}

export function useChallenge(challengeId: string | undefined) {
  return useQuery({
    queryKey: ['flywheel', 'challenge', challengeId],
    queryFn: async (): Promise<{ challenge: Challenge }> => {
      if (!challengeId) throw new Error('Challenge ID is required');
      return apiClient.get<{ challenge: Challenge }>(`${ENDPOINTS.challenges}/${challengeId}`);
    },
    enabled: !!challengeId,
    staleTime: 1000 * 60, // 1 minute
  });
}

interface ChallengeSubmissionRequest {
  challengeId: string;
  code: string;
  language: string;
  notes?: string;
}

export function useSubmitChallenge() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: ChallengeSubmissionRequest): Promise<{ submission: ChallengeSubmission }> => {
      return apiClient.post<{ submission: ChallengeSubmission }>(
        `${ENDPOINTS.challenges}/${data.challengeId}/submit`,
        data
      );
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: ['flywheel', 'challenge', variables.challengeId]
      });
      queryClient.invalidateQueries({ queryKey: ['flywheel', 'challenges'] });
    },
  });
}

export function useChallengeSubmissions(challengeId: string | undefined) {
  return useQuery({
    queryKey: ['flywheel', 'challenge', challengeId, 'submissions'],
    queryFn: async (): Promise<{ submissions: ChallengeSubmission[] }> => {
      if (!challengeId) throw new Error('Challenge ID is required');
      return apiClient.get<{ submissions: ChallengeSubmission[] }>(
        `${ENDPOINTS.challenges}/${challengeId}/my-submissions`
      );
    },
    enabled: !!challengeId,
  });
}

// ==================== Categories API ====================

interface CategoriesResponse {
  categories: Category[];
}

export function useCategories() {
  return useQuery({
    queryKey: ['flywheel', 'categories'],
    queryFn: async (): Promise<CategoriesResponse> => {
      return apiClient.get<CategoriesResponse>(ENDPOINTS.categories);
    },
    staleTime: 1000 * 60 * 60, // 1 hour
  });
}

// ==================== WebSocket Hook ====================

import { useEffect, useRef, useCallback, useState } from 'react';

interface UseWebSocketOptions {
  onThreadUpdate?: (payload: ThreadUpdatePayload) => void;
  onReplyAdded?: (payload: ReplyAddedPayload) => void;
  onExecutionComplete?: (payload: ExecutionCompletePayload) => void;
  onReputationChange?: (payload: ReputationChangePayload) => void;
  onChallengeUpdate?: (payload: unknown) => void;
  onConnect?: () => void;
  onDisconnect?: () => void;
  onError?: (error: Event) => void;
}

export function useWebSocket(options: UseWebSocketOptions = {}) {
  const wsRef = useRef<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const optionsRef = useRef(options);

  // Keep options ref up to date
  useEffect(() => {
    optionsRef.current = options;
  }, [options]);

  const connect = useCallback(() => {
    const apiBase = getApiBaseUrl();
    const wsUrl = apiBase.startsWith('http')
      ? apiBase.replace(/^http/, 'ws').replace(/\/$/, '') + '/v1/flywheel/ws'
      : `ws://${window.location.host}${apiBase}/v1/flywheel/ws`;

    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      setIsConnected(true);
      optionsRef.current.onConnect?.();

      // Send authentication token
      const token = localStorage.getItem('ff-access-token');
      if (token) {
        ws.send(JSON.stringify({ type: 'auth', token }));
      }
    };

    ws.onclose = () => {
      setIsConnected(false);
      optionsRef.current.onDisconnect?.();

      // Attempt to reconnect after 3 seconds
      reconnectTimeoutRef.current = setTimeout(() => {
        connect();
      }, 3000);
    };

    ws.onerror = (error) => {
      optionsRef.current.onError?.(error);
    };

    ws.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);

        switch (message.type) {
          case 'thread_update':
            optionsRef.current.onThreadUpdate?.(message.payload as ThreadUpdatePayload);
            break;
          case 'reply_added':
            optionsRef.current.onReplyAdded?.(message.payload as ReplyAddedPayload);
            break;
          case 'execution_complete':
            optionsRef.current.onExecutionComplete?.(message.payload as ExecutionCompletePayload);
            break;
          case 'reputation_change':
            optionsRef.current.onReputationChange?.(message.payload as ReputationChangePayload);
            break;
          case 'challenge_update':
            optionsRef.current.onChallengeUpdate?.(message.payload);
            break;
        }
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };

    wsRef.current = ws;
  }, []);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
    wsRef.current?.close();
    wsRef.current = null;
  }, []);

  const sendMessage = useCallback((message: Omit<WebSocketMessage, 'timestamp'>) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
    }
  }, []);

  useEffect(() => {
    connect();

    return () => {
      disconnect();
    };
  }, [connect, disconnect]);

  return {
    isConnected,
    sendMessage,
    connect,
    disconnect,
  };
}

// ==================== Real-time Thread Updates ====================

export function useThreadRealtime(threadId: string | undefined) {
  const queryClient = useQueryClient();

  const handleReplyAdded = useCallback((payload: ReplyAddedPayload) => {
    if (payload.threadId === threadId) {
      // Update the replies cache
      queryClient.setQueryData<{ replies: Reply[]; pagination: Pagination }>(
        ['flywheel', 'replies', threadId],
        (old) => {
          if (!old) return old;
          return {
            ...old,
            replies: [...old.replies, payload.reply],
          };
        }
      );

      // Invalidate thread to update reply count
      queryClient.invalidateQueries({ queryKey: ['flywheel', 'thread', threadId] });
    }
  }, [queryClient, threadId]);

  const handleExecutionComplete = useCallback((payload: ExecutionCompletePayload) => {
    // Update execution results in the cache
    queryClient.setQueryData<{ results: ExecutionResults }>(
      ['flywheel', 'execution', payload.executionId],
      { results: payload.results }
    );

    // Invalidate replies to show updated verification status
    if (threadId) {
      queryClient.invalidateQueries({ queryKey: ['flywheel', 'replies', threadId] });
    }
  }, [queryClient, threadId]);

  return useWebSocket({
    onReplyAdded: handleReplyAdded,
    onExecutionComplete: handleExecutionComplete,
  });
}

// ==================== Real-time Leaderboard Updates ====================

export function useLeaderboardRealtime() {
  const queryClient = useQueryClient();

  const handleReputationChange = useCallback((payload: ReputationChangePayload) => {
    // Invalidate all leaderboard queries
    queryClient.invalidateQueries({ queryKey: ['flywheel', 'leaderboard'] });
  }, [queryClient]);

  return useWebSocket({
    onReputationChange: handleReputationChange,
  });
}
