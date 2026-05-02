import React, { useEffect } from 'react';
import { useChatStore } from '@/stores/chatStore';
import { ChatInput, ChatMessage, ChatSidebar, ConnectorPicker, ModelSelector } from './ChatComponents';
import { Loader2 } from 'lucide-react';

export const ChatPage: React.FC = () => {
  const {
    sessions,
    currentSession,
    messages,
    connectors,
    models,
    isLoading,
    isSending,
    error,
    fetchSessions,
    createSession,
    selectSession,
    deleteSession,
    sendMessage,
    fetchConnectors,
    fetchModels,
    clearError,
  } = useChatStore();

  useEffect(() => {
    fetchSessions();
    fetchConnectors();
    fetchModels();
  }, [fetchSessions, fetchConnectors, fetchModels]);

  const handleSend = async (content: string) => {
    await sendMessage(content);
  };

  const handleNewSession = async () => {
    await createSession();
  };

  if (isLoading && sessions.length === 0) {
    return (
      <div className="flex items-center justify-center h-full">
        <Loader2 className="w-8 h-8 animate-spin text-indigo-600" />
      </div>
    );
  }

  return (
    <div className="flex h-full bg-white dark:bg-gray-900">
      <ChatSidebar
        sessions={sessions.map((s) => ({
          id: s.id,
          title: s.title,
          model: s.model,
          updated_at: s.updated_at,
        }))}
        currentSessionId={currentSession?.id ?? null}
        onSelect={(id) => selectSession(sessions.find((s) => s.id === id) ?? null)}
        onDelete={deleteSession}
        onNew={handleNewSession}
      />

      <div className="flex-1 flex flex-col">
        {error && (
          <div className="mx-4 mt-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg flex items-center justify-between">
            <span className="text-sm text-red-700 dark:text-red-400">{error}</span>
            <button onClick={clearError} className="text-red-500 hover:text-red-700">
              Dismiss
            </button>
          </div>
        )}

        {currentSession ? (
          <>
            <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
              <h2 className="text-lg font-semibold">{currentSession.title}</h2>
              <ModelSelector
                models={models}
                selected={currentSession.model}
                onSelect={(model) => console.log('Model change not implemented:', model)}
              />
            </div>

            <div className="flex-1 overflow-y-auto p-4 space-y-4">
              {messages.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-full text-gray-500">
                  <p className="text-lg mb-2">Start a conversation</p>
                  <p className="text-sm">Ask questions, analyze data, create functions</p>
                </div>
              ) : (
                messages.map((msg) => <ChatMessage key={msg.id} message={msg} />)
              )}
            </div>

            <ConnectorPicker connectors={connectors} onToggle={(id) => console.log('Toggle connector:', id)} />

            <ChatInput onSend={handleSend} disabled={isSending} />
          </>
        ) : (
          <div className="flex-1 flex flex-col items-center justify-center">
            <h2 className="text-2xl font-bold mb-4">FunctionFly Chat</h2>
            <p className="text-gray-500 mb-6">
              Connect your data sources and chat with AI to analyze, create, and automate.
            </p>
            <button
              onClick={handleNewSession}
              className="px-6 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors"
            >
              Start New Chat
            </button>
          </div>
        )}
      </div>
    </div>
  );
};