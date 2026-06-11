import { create } from 'zustand';
import { chatApi, type ChatSession, type ChatMessage, type ChatConnector, type AIModel } from '@/api/chat';

interface ChatState {
  sessions: ChatSession[];
  currentSession: ChatSession | null;
  messages: ChatMessage[];
  connectors: ChatConnector[];
  models: AIModel[];
  isLoading: boolean;
  isSending: boolean;
  error: string | null;

  fetchSessions: () => Promise<void>;
  createSession: (title?: string, model?: string) => Promise<ChatSession>;
  selectSession: (session: ChatSession | null) => void;
  deleteSession: (id: string) => Promise<void>;
  updateSession: (id: string, updates: { title?: string; model?: string }) => Promise<void>;

  fetchMessages: (sessionId: string) => Promise<void>;
  sendMessage: (content: string) => Promise<void>;

  fetchConnectors: () => Promise<void>;
  registerConnector: (name: string, type: string, config: Record<string, unknown>) => Promise<void>;
  deleteConnector: (id: string) => Promise<void>;

  fetchModels: () => Promise<void>;

  clearError: () => void;
}

export const useChatStore = create<ChatState>((set, get) => ({
  sessions: [],
  currentSession: null,
  messages: [],
  connectors: [],
  models: [],
  isLoading: false,
  isSending: false,
  error: null,

  fetchSessions: async () => {
    set({ isLoading: true, error: null });
    try {
      const { sessions } = await chatApi.listSessions();
      set({ sessions, isLoading: false });
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to fetch sessions', isLoading: false });
    }
  },

  createSession: async (title = 'New Chat', model = 'gpt-4o-mini') => {
    set({ isLoading: true, error: null });
    try {
      const session = await chatApi.createSession(title, model);
      set((state) => ({
        sessions: [session, ...state.sessions],
        currentSession: session,
        messages: [],
        isLoading: false,
      }));
      return session;
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to create session', isLoading: false });
      throw error;
    }
  },

  selectSession: (session) => {
    set({ currentSession: session, messages: [], error: null });
    if (session) {
      get().fetchMessages(session.id);
    }
  },

  deleteSession: async (id) => {
    try {
      await chatApi.deleteSession(id);
      set((state) => ({
        sessions: state.sessions.filter((s) => s.id !== id),
        currentSession: state.currentSession?.id === id ? null : state.currentSession,
      }));
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to delete session' });
    }
  },

  updateSession: async (id, updates) => {
    try {
      await chatApi.updateSession(id, updates);
      set((state) => ({
        sessions: state.sessions.map((s) =>
          s.id === id ? { ...s, ...updates } : s
        ),
        currentSession:
          state.currentSession?.id === id
            ? { ...state.currentSession, ...updates }
            : state.currentSession,
      }));
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to update session' });
    }
  },

  fetchMessages: async (sessionId) => {
    set({ isLoading: true, error: null });
    try {
      const { messages } = await chatApi.getSession(sessionId);
      set({ messages, isLoading: false });
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to fetch messages', isLoading: false });
    }
  },

  sendMessage: async (content) => {
    const { currentSession } = get();
    if (!currentSession) {
      set({ error: 'No session selected' });
      return;
    }

    set({ isSending: true, error: null });
    try {
      const { message } = await chatApi.sendMessage(currentSession.id, content);
      set((state) => ({
        messages: [...state.messages, message],
        isSending: false,
      }));
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to send message', isSending: false });
    }
  },

  fetchConnectors: async () => {
    try {
      const { connectors } = await chatApi.listConnectors();
      set({ connectors });
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to fetch connectors' });
    }
  },

  registerConnector: async (name, type, config) => {
    try {
      const connector = await chatApi.registerConnector(name, type, config);
      set((state) => ({ connectors: [...state.connectors, connector] }));
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to register connector' });
    }
  },

  deleteConnector: async (id) => {
    try {
      await chatApi.deleteConnector(id);
      set((state) => ({
        connectors: state.connectors.filter((c) => c.id !== id),
      }));
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to delete connector' });
    }
  },

  fetchModels: async () => {
    try {
      const { models } = await chatApi.listModels();
      set({ models });
    } catch (error) {
      set({ error: error instanceof Error ? error.message : 'Failed to fetch models' });
    }
  },

  clearError: () => set({ error: null }),
}));