import React, { useState, useRef, useEffect } from 'react';
import { useChatStore } from '@/stores/chatStore';
import { X, Send, Plus, Trash2, Settings, Zap } from 'lucide-react';

interface ChatInputProps {
  onSend: (content: string) => void;
  disabled?: boolean;
}

export const ChatInput: React.FC<ChatInputProps> = ({ onSend, disabled }) => {
  const [content, setContent] = useState('');
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (content.trim() && !disabled) {
      onSend(content.trim());
      setContent('');
    }
  };

  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
      textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 200)}px`;
    }
  }, [content]);

  return (
    <form onSubmit={handleSubmit} className="flex items-end gap-2 p-4 border-t border-gray-200 dark:border-gray-700">
      <textarea
        ref={textareaRef}
        value={content}
        onChange={(e) => setContent(e.target.value)}
        placeholder="Ask anything..."
        className="flex-1 resize-none rounded-lg border border-gray-300 dark:border-gray-600 px-4 py-2 dark:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-indigo-500"
        rows={1}
        disabled={disabled}
      />
      <button
        type="submit"
        disabled={disabled || !content.trim()}
        className="p-2 rounded-lg bg-indigo-600 text-white disabled:opacity-50 disabled:cursor-not-allowed hover:bg-indigo-700 transition-colors"
      >
        <Send className="w-5 h-5" />
      </button>
    </form>
  );
};

interface ChatMessageProps {
  message: {
    role: string;
    content: string;
    created_at: string;
    tokens_used?: number;
    latency_ms?: number;
  };
}

export const ChatMessage: React.FC<ChatMessageProps> = ({ message }) => {
  const isUser = message.role === 'user';

  return (
    <div className={`flex ${isUser ? 'justify-end' : 'justify-start'}`}>
      <div
        className={`max-w-[80%] rounded-lg px-4 py-2 ${
          isUser ? 'bg-indigo-600 text-white' : 'bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-gray-100'
        }`}
      >
        <div className="prose prose-sm dark:prose-invert max-w-none whitespace-pre-wrap">
          {message.content}
        </div>
        <div className={`text-xs mt-1 ${isUser ? 'text-indigo-200' : 'text-gray-500'}`}>
          {new Date(message.created_at).toLocaleTimeString()}
          {message.tokens_used != null && ` • ${message.tokens_used} tokens`}
          {message.latency_ms != null && ` • ${message.latency_ms}ms`}
        </div>
      </div>
    </div>
  );
};

interface ChatSidebarProps {
  sessions: Array<{
    id: string;
    title: string;
    model: string;
    updated_at: string;
  }>;
  currentSessionId: string | null;
  onSelect: (id: string) => void;
  onDelete: (id: string) => void;
  onNew: () => void;
}

export const ChatSidebar: React.FC<ChatSidebarProps> = ({
  sessions,
  currentSessionId,
  onSelect,
  onDelete,
  onNew,
}) => {
  return (
    <div className="w-64 border-r border-gray-200 dark:border-gray-700 flex flex-col h-full">
      <div className="p-4 border-b border-gray-200 dark:border-gray-700">
        <button
          onClick={onNew}
          className="w-full flex items-center justify-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors"
        >
          <Plus className="w-4 h-4" />
          New Chat
        </button>
      </div>
      <div className="flex-1 overflow-y-auto">
        {sessions.map((session) => (
          <div
            key={session.id}
            className={`group flex items-center justify-between p-3 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800 ${
              currentSessionId === session.id ? 'bg-gray-100 dark:bg-gray-800' : ''
            }`}
          >
            <div className="flex-1 min-w-0" onClick={() => onSelect(session.id)}>
              <div className="truncate text-sm font-medium">{session.title}</div>
              <div className="text-xs text-gray-500 truncate">{session.model}</div>
            </div>
            <button
              onClick={(e) => {
                e.stopPropagation();
                onDelete(session.id);
              }}
              className="opacity-0 group-hover:opacity-100 p-1 hover:text-red-500 transition-opacity"
            >
              <Trash2 className="w-4 h-4" />
            </button>
          </div>
        ))}
      </div>
      <div className="p-3 border-t border-gray-200 dark:border-gray-700">
        <button className="w-full flex items-center gap-2 px-3 py-2 text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 rounded-lg">
          <Settings className="w-4 h-4" />
          Settings
        </button>
      </div>
    </div>
  );
};

interface ConnectorPickerProps {
  connectors: Array<{
    id: string;
    name: string;
    type: string;
    icon: string;
    is_active: boolean;
  }>;
  onToggle: (id: string) => void;
}

export const ConnectorPicker: React.FC<ConnectorPickerProps> = ({ connectors, onToggle }) => {
  return (
    <div className="flex flex-wrap gap-2 p-3 border-t border-gray-200 dark:border-gray-700">
      <span className="text-sm text-gray-500 flex items-center">
        <Zap className="w-4 h-4 mr-1" />
        Integrations:
      </span>
      {connectors.map((connector) => (
        <button
          key={connector.id}
          onClick={() => onToggle(connector.id)}
          className={`px-3 py-1 text-sm rounded-full border transition-colors ${
            connector.is_active
              ? 'bg-indigo-100 border-indigo-300 text-indigo-700 dark:bg-indigo-900 dark:border-indigo-700 dark:text-indigo-300'
              : 'bg-gray-50 border-gray-200 text-gray-600 dark:bg-gray-800 dark:border-gray-700 dark:text-gray-400'
          }`}
        >
          {connector.name}
        </button>
      ))}
    </div>
  );
};

interface ModelSelectorProps {
  models: Array<{
    id: string;
    name: string;
    provider: string;
  }>;
  selected: string;
  onSelect: (id: string) => void;
}

export const ModelSelector: React.FC<ModelSelectorProps> = ({ models, selected, onSelect }) => {
  return (
    <select
      value={selected}
      onChange={(e) => onSelect(e.target.value)}
      className="text-sm rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-2 py-1"
    >
      {models.map((model) => (
        <option key={model.id} value={model.id}>
          {model.name} ({model.provider})
        </option>
      ))}
    </select>
  );
};