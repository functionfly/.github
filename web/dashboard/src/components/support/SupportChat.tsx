import React, { createContext, useCallback, useContext, useEffect, useRef, useState } from 'react';

import { API_BASE_URL } from '@/lib/constants';

// ============================================================================
// Types & Interfaces
// ============================================================================

export interface SupportMessage {
  id: string;
  conversation_id: string;
  author_id: string;
  author_type: 'user' | 'ai' | 'staff' | 'system';
  message_type: 'message' | 'context' | 'escalation' | 'resolution' | 'ai_response' | 'system';
  content: string;
  ai_confidence?: number;
  ai_model?: string;
  created_at: string;
}

export interface SupportConversation {
  id: string;
  user_id: string;
  type: 'support_ai' | 'support_human' | 'support_emergency';
  status: 'active' | 'pending' | 'resolved' | 'escalated';
  priority: 'low' | 'normal' | 'high' | 'critical';
  title: string;
  function_ref?: {
    author: string;
    name: string;
    version?: string;
  };
  ai_handled: boolean;
  ai_attempts: number;
  staff_id?: string;
  staff_joined_at?: string;
  is_emergency: boolean;
  created_at: string;
  updated_at: string;
}

export interface SupportContext {
  function_code?: string;
  function_logs?: string[];
  deployment_error?: string;
  environment_vars?: Record<string, string>;
}

// ============================================================================
// API Functions
// ============================================================================

// Helper to get auth token (dashboard uses ff-access-token; legacy auth_token fallback)
function getAuthToken(): string {
  return localStorage.getItem('ff-access-token') || localStorage.getItem('auth_token') || '';
}

async function createConversation(data: {
  type: string;
  title: string;
  function_author?: string;
  function_name?: string;
  function_version?: string;
  is_emergency?: boolean;
}): Promise<SupportConversation> {
  const response = await fetch(`${API_BASE_URL}/v1/support/conversations`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${getAuthToken()}`,
    },
    body: JSON.stringify(data),
  });
  if (!response.ok) {
    let msg = 'Failed to create conversation';
    try {
      const body = await response.json();
      if (body?.error) msg = body.error;
    } catch {
      // ignore parse error
    }
    throw new Error(msg);
  }
  return response.json();
}

async function getConversation(id: string): Promise<SupportConversation> {
  const response = await fetch(`${API_BASE_URL}/v1/support/conversations/${id}`, {
    headers: {
      Authorization: `Bearer ${getAuthToken()}`,
    },
  });
  if (!response.ok) throw new Error('Failed to get conversation');
  return response.json();
}

async function fetchMessages(
  conversationId: string,
  limit = 50,
  offset = 0
): Promise<SupportMessage[]> {
  const response = await fetch(
    `${API_BASE_URL}/v1/support/conversations/${conversationId}/messages?limit=${limit}&offset=${offset}`,
    {
      headers: {
        Authorization: `Bearer ${getAuthToken()}`,
      },
    }
  );
  if (!response.ok) throw new Error('Failed to get messages');
  const data = await response.json();
  return data.messages;
}

async function postMessage(conversationId: string, content: string): Promise<SupportMessage> {
  const response = await fetch(
    `${API_BASE_URL}/v1/support/conversations/${conversationId}/messages`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${getAuthToken()}`,
      },
      body: JSON.stringify({ content }),
    }
  );
  if (!response.ok) throw new Error('Failed to send message');
  return response.json();
}

async function escalateConversation(conversationId: string): Promise<void> {
  const response = await fetch(
    `${API_BASE_URL}/v1/support/conversations/${conversationId}/escalate`,
    {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${getAuthToken()}`,
      },
    }
  );
  if (!response.ok) throw new Error('Failed to escalate');
}

async function resolveConversationRequest(conversationId: string, note?: string): Promise<void> {
  const response = await fetch(
    `${API_BASE_URL}/v1/support/conversations/${conversationId}/resolve`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${getAuthToken()}`,
      },
      body: JSON.stringify({ note }),
    }
  );
  if (!response.ok) throw new Error('Failed to resolve');
}

async function createEmergencyFix(
  conversationId: string,
  functionId: string,
  reason: string
): Promise<void> {
  const response = await fetch(
    `${API_BASE_URL}/v1/support/conversations/${conversationId}/emergency`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${getAuthToken()}`,
      },
      body: JSON.stringify({ function_id: functionId, reason }),
    }
  );
  if (!response.ok) throw new Error('Failed to create emergency');
}

// ============================================================================
// WebSocket Hook
// ============================================================================

interface UseSupportWebSocketOptions {
  conversationId: string | null;
  onMessage: (message: SupportMessage) => void;
  onTyping: (userId: string, typing: boolean) => void;
  onStaffJoined?: (staffId: string) => void;
  onStaffLeft?: (staffId: string) => void;
}

function useSupportWebSocket({
  conversationId,
  onMessage,
  onTyping,
  onStaffJoined,
  onStaffLeft,
}: UseSupportWebSocketOptions) {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const connect = useCallback(() => {
    if (!conversationId) return;
    // Build WebSocket URL from API base (same pattern as useRealtimeSubscription)
    const base =
      API_BASE_URL.startsWith('http://') || API_BASE_URL.startsWith('https://')
        ? API_BASE_URL
        : `${typeof window !== 'undefined' ? window.location.origin : ''}${API_BASE_URL}`;
    const wsBase = base.replace(/^http/, 'ws');
    const wsUrl = `${wsBase.replace(/\/$/, '')}/v1/support/ws`;

    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log('Support WebSocket connected');
      // Join the conversation room
      ws.send(
        JSON.stringify({
          type: 'join_conversation',
          payload: { conversation_id: conversationId },
        })
      );
    };

    ws.onmessage = (event) => {
      const data = event.data;
      if (data == null || typeof data !== 'string') return;
      const messages = data.split('\n');
      for (const msgStr of messages) {
        if (!msgStr) continue;
        try {
          const msg = JSON.parse(msgStr);
          switch (msg.type) {
            case 'new_message':
              onMessage(msg.payload);
              break;
            case 'user_typing':
              onTyping(msg.payload.user_id, msg.payload.typing);
              break;
            case 'staff_joined':
              onStaffJoined?.(msg.payload.staff_id);
              break;
            case 'staff_left':
              onStaffLeft?.(msg.payload.staff_id);
              break;
          }
        } catch (e) {
          console.error('Failed to parse WebSocket message:', e);
        }
      }
    };

    ws.onclose = () => {
      console.log('Support WebSocket disconnected');
      // Attempt to reconnect
      reconnectTimeoutRef.current = setTimeout(() => {
        connect();
      }, 3000);
    };

    ws.onerror = (error) => {
      console.error('Support WebSocket error:', error);
    };
  }, [conversationId, onMessage, onTyping, onStaffJoined, onStaffLeft]);

  useEffect(() => {
    if (!conversationId) return;
    connect();
    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [connect]);

  const sendTypingIndicator = useCallback(
    (typing: boolean) => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(
          JSON.stringify({
            type: 'typing',
            payload: { conversation_id: conversationId, typing },
          })
        );
      }
    },
    [conversationId]
  );

  const sendWsMessage = useCallback(
    (content: string) => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(
          JSON.stringify({
            type: 'chat_message',
            payload: { conversation_id: conversationId, content },
          })
        );
      }
    },
    [conversationId]
  );

  return { sendTypingIndicator, sendWsMessage };
}

// ============================================================================
// Context & Provider
// ============================================================================

interface SupportChatContextValue {
  // State
  isOpen: boolean;
  conversation: SupportConversation | null;
  messages: SupportMessage[];
  isLoading: boolean;
  isSending: boolean;
  isConnected: boolean;
  staffOnline: boolean;
  openError: string | null;

  // Actions
  openChat: (options?: {
    functionRef?: { author: string; name: string; version?: string };
    isEmergency?: boolean;
  }) => Promise<void>;
  closeChat: () => void;
  sendMessage: (content: string) => Promise<void>;
  escalateToHuman: () => Promise<void>;
  resolveChat: (note?: string) => Promise<void>;
  requestEmergencyFix: (functionId: string, reason: string) => Promise<void>;
}

const SupportChatContext = createContext<SupportChatContextValue | null>(null);

export function useSupportChat(): SupportChatContextValue {
  const context = useContext(SupportChatContext);
  if (!context) {
    throw new Error('useSupportChat must be used within SupportChatProvider');
  }
  return context;
}

// ============================================================================
// Provider Component
// ============================================================================

interface SupportChatProviderProps {
  children: React.ReactNode;
}

export function SupportChatProvider({ children }: SupportChatProviderProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [conversation, setConversation] = useState<SupportConversation | null>(null);
  const [messages, setMessages] = useState<SupportMessage[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isSending, setIsSending] = useState(false);
  const [staffOnline, setStaffOnline] = useState(false);
  const [openError, setOpenError] = useState<string | null>(null);

  const handleNewMessage = useCallback((message: SupportMessage) => {
    setMessages((prev) => [...prev, message]);
  }, []);

  const handleTyping = useCallback((userId: string, typing: boolean) => {
    // Could show typing indicator UI
    console.log('User typing:', userId, typing);
  }, []);

  const { sendTypingIndicator, sendWsMessage } = useSupportWebSocket({
    conversationId: conversation?.id ?? null,
    onMessage: handleNewMessage,
    onTyping: handleTyping,
  });

  const openChat = useCallback(
    async (options?: {
      functionRef?: { author: string; name: string; version?: string };
      isEmergency?: boolean;
    }) => {
      setIsOpen(true);
      setIsLoading(true);
      setOpenError(null);
      try {
        const conv = await createConversation({
          type: options?.isEmergency ? 'support_emergency' : 'support_ai',
          title: options?.functionRef
            ? `Help with ${options.functionRef.author}/${options.functionRef.name}`
            : 'General support request',
          function_author: options?.functionRef?.author,
          function_name: options?.functionRef?.name,
          function_version: options?.functionRef?.version,
          is_emergency: options?.isEmergency,
        });
        setConversation(conv);
        setMessages([]);
      } catch (error) {
        console.error('Failed to open chat:', error);
        setConversation(null);
        setMessages([]);
        setOpenError(error instanceof Error ? error.message : 'Failed to open chat');
      } finally {
        setIsLoading(false);
      }
    },
    []
  );

  const closeChat = useCallback(() => {
    setIsOpen(false);
    setConversation(null);
    setMessages([]);
    setOpenError(null);
  }, []);

  const sendMessage = useCallback(
    async (content: string) => {
      if (!conversation || isSending) return;

      setIsSending(true);
      try {
        // Optimistically add user message
        const userMessage: SupportMessage = {
          id: `temp-${Date.now()}`,
          conversation_id: conversation.id,
          author_id: 'current-user',
          author_type: 'user',
          message_type: 'message',
          content,
          created_at: new Date().toISOString(),
        };
        setMessages((prev) => [...prev, userMessage]);

        // Send via WebSocket for real-time
        sendWsMessage(content);

        // Also send via REST for persistence
        await postMessage(conversation.id, content);
      } catch (error) {
        console.error('Failed to send message:', error);
      } finally {
        setIsSending(false);
      }
    },
    [conversation, isSending, sendWsMessage]
  );

  const escalateToHuman = useCallback(async () => {
    if (!conversation) return;
    try {
      await escalateConversation(conversation.id);
      // Refresh conversation
      const updated = await getConversation(conversation.id);
      setConversation(updated);
    } catch (error) {
      console.error('Failed to escalate:', error);
    }
  }, [conversation]);

  const resolveChat = useCallback(
    async (note?: string) => {
      if (!conversation) return;
      try {
        await resolveConversationRequest(conversation.id, note);
        closeChat();
      } catch (error) {
        console.error('Failed to resolve:', error);
      }
    },
    [conversation, closeChat]
  );

  const requestEmergencyFix = useCallback(
    async (functionId: string, reason: string) => {
      if (!conversation) return;
      try {
        await createEmergencyFix(conversation.id, functionId, reason);
        // Show confirmation to user
      } catch (error) {
        console.error('Failed to request emergency:', error);
      }
    },
    [conversation]
  );

  const value: SupportChatContextValue = {
    isOpen,
    conversation,
    messages,
    isLoading,
    isSending,
    isConnected: true,
    staffOnline,
    openError,
    openChat,
    closeChat,
    sendMessage,
    escalateToHuman,
    resolveChat,
    requestEmergencyFix,
  };

  return <SupportChatContext.Provider value={value}>{children}</SupportChatContext.Provider>;
}

export default SupportChatProvider;
