import {
  type Conversation,
  type ConversationMessage,
} from '@/api/conversations';
import { getApiBaseUrl } from '@/lib/constants';
import { useQueryClient } from '@tanstack/react-query';
import { useCallback, useEffect, useRef, useState } from 'react';
import { conversationKeys } from './useConversations';

// ---- Types ----------------------------------------------------------------

export interface ConversationWSMessage {
  type: string;
  payload: unknown;
}

export interface TypingEvent {
  user_id: string;
  conversation_id: string;
  typing: boolean;
}

export interface UseConversationWebSocketOptions {
  /** Conversation IDs to join immediately on connect. Can be updated reactively. */
  conversationIds?: string[];
  /** When false the socket will not connect. Defaults to true. */
  enabled?: boolean;
  /** Called when a new message arrives over the wire. */
  onNewMessage?: (message: ConversationMessage) => void;
  /** Called when a conversation is resolved. */
  onConversationResolved?: (conversation: Conversation) => void;
  /** Called when a typing indicator event arrives. */
  onTyping?: (event: TypingEvent) => void;
}

interface WSState {
  isConnected: boolean;
  isConnecting: boolean;
  error: Error | null;
  reconnectAttempt: number;
}

// ---- Reconnection config ---------------------------------------------------

const RECONNECT_BASE_DELAY = 1000;
const RECONNECT_MAX_DELAY = 30000;
const RECONNECT_MAX_ATTEMPTS = 10;

// ---- Hook -----------------------------------------------------------------

/**
 * Real-time WebSocket connection for the conversations feature.
 *
 * Replaces polling-based message fetching with instant push.
 * Manages join/leave for conversation rooms and keeps the
 * React-Query cache synchronised with incoming events.
 */
export function useConversationWebSocket(
  options: UseConversationWebSocketOptions = {},
) {
  const {
    conversationIds = [],
    enabled = true,
    onNewMessage,
    onConversationResolved,
    onTyping,
  } = options;

  const queryClient = useQueryClient();
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const isIntentionallyClosedRef = useRef(false);
  const joinedIdsRef = useRef<Set<string>>(new Set());

  const [state, setState] = useState<WSState>({
    isConnected: false,
    isConnecting: false,
    error: null,
    reconnectAttempt: 0,
  });

  // ---- Helpers ------------------------------------------------------------

  const getWebSocketUrl = useCallback(() => {
    const apiBaseUrl = getApiBaseUrl();

    // Vite dev proxy uses '/api' — fall back to window location.
    if (apiBaseUrl === '/api') {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      return `${protocol}//${window.location.host}/v1/conversations/ws`;
    }

    const wsUrl = apiBaseUrl
      .replace(/^http:/, 'ws:')
      .replace(/^https:/, 'wss:');
    return `${wsUrl}/v1/conversations/ws`;
  }, []);

  const getReconnectDelay = useCallback(() => {
    return Math.min(
      RECONNECT_BASE_DELAY * Math.pow(2, reconnectAttemptsRef.current),
      RECONNECT_MAX_DELAY,
    );
  }, []);

  const send = useCallback((msg: ConversationWSMessage) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(msg));
    }
  }, []);

  const joinConversation = useCallback(
    (id: string) => {
      joinedIdsRef.current.add(id);
      send({ type: 'join_conversation', payload: { conversation_id: id } });
    },
    [send],
  );

  const leaveConversation = useCallback(
    (id: string) => {
      joinedIdsRef.current.delete(id);
      send({ type: 'leave_conversation', payload: { conversation_id: id } });
    },
    [send],
  );

  // ---- Message handler ----------------------------------------------------

  const handleMessage = useCallback(
    (event: MessageEvent) => {
      try {
        const message: ConversationWSMessage = JSON.parse(event.data);

        switch (message.type) {
          case 'new_message': {
            const msg = message.payload as ConversationMessage;

            // Update the messages query cache for this conversation.
            queryClient.setQueryData(
              conversationKeys.messages(msg.conversation_id),
              (old: { messages: ConversationMessage[] } | undefined) => {
                if (!old) return { messages: [msg] };
                // Avoid duplicates.
                const exists = old.messages.some((m) => m.id === msg.id);
                if (exists) return old;
                return { messages: [...old.messages, msg] };
              },
            );

            // Invalidate conversation list so unread counts stay fresh.
            queryClient.invalidateQueries({
              queryKey: conversationKeys.lists(),
            });

            onNewMessage?.(msg);
            break;
          }

          case 'conversation_resolved': {
            const conv = message.payload as Conversation;

            queryClient.invalidateQueries({
              queryKey: conversationKeys.detail(conv.id),
            });
            queryClient.invalidateQueries({
              queryKey: conversationKeys.lists(),
            });

            onConversationResolved?.(conv);
            break;
          }

          case 'conversation_read': {
            // Someone read a conversation — refresh list for unread counts.
            queryClient.invalidateQueries({
              queryKey: conversationKeys.lists(),
            });
            break;
          }

          case 'user_typing': {
            const event = message.payload as TypingEvent;
            onTyping?.(event);
            break;
          }

          case 'message_updated': {
            const msg = message.payload as ConversationMessage;

            queryClient.setQueryData(
              conversationKeys.messages(msg.conversation_id),
              (old: { messages: ConversationMessage[] } | undefined) => {
                if (!old) return old;
                return {
                  messages: old.messages.map((m) => (m.id === msg.id ? msg : m)),
                };
              },
            );
            break;
          }

          case 'message_deleted': {
            const { conversation_id: convId, message_id: msgId } = message.payload as {
              conversation_id: string;
              message_id: string;
            };

            queryClient.setQueryData(
              conversationKeys.messages(convId),
              (old: { messages: ConversationMessage[] } | undefined) => {
                if (!old) return old;
                return {
                  messages: old.messages.map((m) =>
                    m.id === msgId ? { ...m, deleted_at: new Date().toISOString() } : m
                  ),
                };
              },
            );
            break;
          }

          case 'pong': {
            // Keep-alive response — no action needed.
            break;
          }
        }
      } catch (err) {
        console.error('Failed to parse conversations WS message:', err);
      }
    },
    [queryClient, onNewMessage, onConversationResolved, onTyping],
  );

  // ---- Connect / Disconnect -----------------------------------------------

  const connect = useCallback(() => {
    if (
      !enabled ||
      wsRef.current?.readyState === WebSocket.OPEN ||
      wsRef.current?.readyState === WebSocket.CONNECTING
    ) {
      return;
    }

    setState((prev) => ({ ...prev, isConnecting: true, error: null }));

    try {
      const ws = new WebSocket(getWebSocketUrl());

      ws.onopen = () => {
        reconnectAttemptsRef.current = 0;
        setState({
          isConnected: true,
          isConnecting: false,
          error: null,
          reconnectAttempt: 0,
        });

        // Re-join all tracked conversations.
        joinedIdsRef.current.forEach((id) => {
          ws.send(
            JSON.stringify({
              type: 'join_conversation',
              payload: { conversation_id: id },
            }),
          );
        });
      };

      ws.onmessage = handleMessage;

      ws.onerror = () => {
        setState((prev) => ({
          ...prev,
          isConnected: false,
          isConnecting: false,
          error: new Error('WebSocket connection error'),
        }));
      };

      ws.onclose = () => {
        setState((prev) => ({
          ...prev,
          isConnected: false,
          isConnecting: false,
        }));

        wsRef.current = null;

        if (!isIntentionallyClosedRef.current && enabled) {
          if (reconnectAttemptsRef.current < RECONNECT_MAX_ATTEMPTS) {
            const delay = getReconnectDelay();
            reconnectAttemptsRef.current += 1;

            setState((prev) => ({
              ...prev,
              reconnectAttempt: reconnectAttemptsRef.current,
            }));

            reconnectTimeoutRef.current = setTimeout(connect, delay);
          } else {
            setState((prev) => ({
              ...prev,
              error: new Error('Max reconnection attempts reached'),
            }));
          }
        }
      };

      wsRef.current = ws;
    } catch (error) {
      setState((prev) => ({
        ...prev,
        isConnecting: false,
        error: error instanceof Error ? error : new Error('Failed to connect'),
      }));
    }
  }, [enabled, getWebSocketUrl, handleMessage, getReconnectDelay]);

  const disconnect = useCallback(() => {
    isIntentionallyClosedRef.current = true;

    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    setState({
      isConnected: false,
      isConnecting: false,
      error: null,
      reconnectAttempt: 0,
    });
  }, []);

  const reconnect = useCallback(() => {
    disconnect();
    isIntentionallyClosedRef.current = false;
    reconnectAttemptsRef.current = 0;
    connect();
  }, [disconnect, connect]);

  // ---- Lifecycle ----------------------------------------------------------

  useEffect(() => {
    if (enabled) {
      isIntentionallyClosedRef.current = false;
      connect();
    } else {
      disconnect();
    }

    return () => {
      disconnect();
    };
  }, [enabled, connect, disconnect]);

  // Sync conversationIds changes: join new ones, leave removed ones.
  useEffect(() => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) return;

    const newSet = new Set(conversationIds);
    const currentSet = joinedIdsRef.current;

    // Join newly added.
    for (const id of newSet) {
      if (!currentSet.has(id)) {
        joinConversation(id);
      }
    }

    // Leave removed.
    for (const id of currentSet) {
      if (!newSet.has(id)) {
        leaveConversation(id);
      }
    }
  }, [conversationIds, joinConversation, leaveConversation]);

  // Cleanup on unmount.
  useEffect(() => {
    return () => {
      disconnect();
    };
  }, [disconnect]);

  // ---- Public API ---------------------------------------------------------

  return {
    ...state,
    connect,
    disconnect,
    reconnect,
    joinConversation,
    leaveConversation,
    send,
  };
}
