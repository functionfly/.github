import { useState, useRef, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { aiChatApi, type AIChatSession, type AIChatMessage } from '@/api/ai_chat';
import { Bot, Send, Plus, Trash2, MessageSquare, Loader2 } from 'lucide-react';
import { formatDate } from '@/lib/utils';

const contextTypes = [
  { value: 'general', label: 'General' },
  { value: 'career', label: 'Career' },
  { value: 'project', label: 'Project' },
  { value: 'learning', label: 'Learning' },
];

export function AIAssistantPage() {
  const queryClient = useQueryClient();
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null);
  const [message, setMessage] = useState('');
  const [contextType, setContextType] = useState('general');
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const { data: sessionsData } = useQuery({
    queryKey: ['ai-chat', 'sessions'],
    queryFn: () => aiChatApi.listSessions(),
  });

  const { data: sessionData, isLoading: sessionLoading } = useQuery({
    queryKey: ['ai-chat', 'session', activeSessionId],
    queryFn: () => aiChatApi.getSession(activeSessionId!),
    enabled: !!activeSessionId,
  });

  const createSessionMutation = useMutation({
    mutationFn: (data?: { title?: string; context_type?: string }) => aiChatApi.createSession(data),
    onSuccess: (res) => {
      queryClient.invalidateQueries({ queryKey: ['ai-chat', 'sessions'] });
      setActiveSessionId(res.data.session.id);
    },
  });

  const sendMessageMutation = useMutation({
    mutationFn: ({ sessionId, message }: { sessionId: string; message: string }) =>
      aiChatApi.sendMessage(sessionId, message),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ai-chat', 'session', activeSessionId] });
      queryClient.invalidateQueries({ queryKey: ['ai-chat', 'sessions'] });
    },
  });

  const deleteSessionMutation = useMutation({
    mutationFn: (id: string) => aiChatApi.deleteSession(id),
    onSuccess: (_, deletedId) => {
      queryClient.invalidateQueries({ queryKey: ['ai-chat', 'sessions'] });
      if (activeSessionId === deletedId) {
        setActiveSessionId(null);
      }
    },
  });

  const sessions = sessionsData?.data?.sessions || [];
  const messages = sessionData?.data?.messages || [];

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSend = async () => {
    if (!message.trim() || !activeSessionId) return;
    const msg = message;
    setMessage('');
    await sendMessageMutation.mutateAsync({ sessionId: activeSessionId, message: msg });
  };

  const handleNewChat = () => {
    createSessionMutation.mutate({ context_type: contextType });
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div className="flex h-[calc(100vh-7rem)] gap-4">
      {/* Sidebar */}
      <div className="flex w-72 flex-col rounded-xl border border-gray-800 bg-gray-900">
        <div className="border-b border-gray-800 p-4">
          <div className="mb-3 flex items-center justify-between">
            <h2 className="text-sm font-semibold text-gray-300">Chat Sessions</h2>
            <select
              value={contextType}
              onChange={(e) => setContextType(e.target.value)}
              className="rounded border border-gray-700 bg-gray-800 px-2 py-1 text-xs text-gray-300"
            >
              {contextTypes.map((ct) => (
                <option key={ct.value} value={ct.value}>{ct.label}</option>
              ))}
            </select>
          </div>
          <button
            onClick={handleNewChat}
            disabled={createSessionMutation.isPending}
            className="flex w-full items-center justify-center gap-2 rounded-lg bg-blue-600 px-3 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
          >
            <Plus className="h-4 w-4" />
            New Chat
          </button>
        </div>

        <div className="flex-1 overflow-y-auto p-2">
          {sessions.length === 0 ? (
            <p className="py-8 text-center text-sm text-gray-500">No conversations yet</p>
          ) : (
            <div className="space-y-1">
              {sessions.map((s) => (
                <div
                  key={s.id}
                  className={`group flex items-center gap-2 rounded-lg px-3 py-2 text-sm transition-colors ${
                    activeSessionId === s.id
                      ? 'bg-blue-600/20 text-blue-400'
                      : 'text-gray-400 hover:bg-gray-800 hover:text-gray-200'
                  }`}
                >
                  <button
                    onClick={() => setActiveSessionId(s.id)}
                    className="flex flex-1 items-center gap-2 truncate text-left"
                  >
                    <MessageSquare className="h-4 w-4 shrink-0" />
                    <span className="truncate">{s.title || 'Untitled'}</span>
                  </button>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      deleteSessionMutation.mutate(s.id);
                    }}
                    className="shrink-0 rounded p-1 text-gray-600 opacity-0 transition-opacity hover:bg-gray-700 hover:text-red-400 group-hover:opacity-100"
                  >
                    <Trash2 className="h-3 w-3" />
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Main chat area */}
      <div className="flex flex-1 flex-col rounded-xl border border-gray-800 bg-gray-900">
        {!activeSessionId ? (
          <div className="flex flex-1 flex-col items-center justify-center">
            <Bot className="mb-4 h-16 w-16 text-gray-600" />
            <h2 className="mb-2 text-xl font-semibold text-gray-300">FWOS AI Assistant</h2>
            <p className="mb-6 text-sm text-gray-500">Start a new conversation to get help with your work</p>
            <button
              onClick={handleNewChat}
              className="flex items-center gap-2 rounded-lg bg-blue-600 px-6 py-3 text-sm font-medium text-white hover:bg-blue-700"
            >
              <Plus className="h-4 w-4" />
              New Conversation
            </button>
          </div>
        ) : (
          <>
            {/* Messages */}
            <div className="flex-1 overflow-y-auto p-4">
              {sessionLoading ? (
                <div className="flex justify-center py-12">
                  <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
                </div>
              ) : (
                <div className="space-y-4">
                  {messages.map((msg) => (
                    <div
                      key={msg.id}
                      className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}
                    >
                      <div
                        className={`max-w-[70%] rounded-2xl px-4 py-3 ${
                          msg.role === 'user'
                            ? 'bg-blue-600 text-white'
                            : 'border border-gray-700 bg-gray-800 text-gray-200'
                        }`}
                      >
                        <p className="whitespace-pre-wrap text-sm">{msg.content}</p>
                        <p className={`mt-1 text-xs ${
                          msg.role === 'user' ? 'text-blue-200' : 'text-gray-500'
                        }`}>
                          {formatDate(msg.created_at)}
                        </p>
                      </div>
                    </div>
                  ))}
                  {sendMessageMutation.isPending && (
                    <div className="flex justify-start">
                      <div className="flex items-center gap-2 rounded-2xl border border-gray-700 bg-gray-800 px-4 py-3">
                        <Loader2 className="h-4 w-4 animate-spin text-blue-400" />
                        <span className="text-sm text-gray-400">Thinking...</span>
                      </div>
                    </div>
                  )}
                  <div ref={messagesEndRef} />
                </div>
              )}
            </div>

            {/* Input */}
            <div className="border-t border-gray-800 p-4">
              <div className="flex gap-3">
                <textarea
                  value={message}
                  onChange={(e) => setMessage(e.target.value)}
                  onKeyDown={handleKeyDown}
                  placeholder="Type your message..."
                  rows={1}
                  className="flex-1 resize-none rounded-lg border border-gray-700 bg-gray-800 px-4 py-3 text-sm text-gray-100 placeholder-gray-500 focus:border-blue-500 focus:outline-none"
                />
                <button
                  onClick={handleSend}
                  disabled={!message.trim() || sendMessageMutation.isPending}
                  className="flex items-center justify-center rounded-lg bg-blue-600 px-4 py-3 text-white hover:bg-blue-700 disabled:opacity-50"
                >
                  <Send className="h-4 w-4" />
                </button>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
