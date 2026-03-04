/**
 * FlywheelChatAssistant - Bottom-right AI assistant chat
 */

import { useState, useRef, useEffect } from 'react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import {
  MessageSquare,
  X,
  Send,
  Bot,
  User,
  Sparkles,
  ChevronDown,
  Lightbulb,
  Wand2,
} from 'lucide-react';

interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
  suggestions?: string[];
}

interface FlywheelChatAssistantProps {
  className?: string;
}

export function FlywheelChatAssistant({ className }: FlywheelChatAssistantProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [isMinimized, setIsMinimized] = useState(false);
  const [inputValue, setInputValue] = useState('');
  const [messages, setMessages] = useState<Message[]>([
    {
      id: 'welcome',
      role: 'assistant',
      content: 'Hi! I\'m your Flywheel AI assistant. I can help you with:\n\n• Understanding problems\n• Suggesting approaches\n• Explaining solutions\n• Code optimization tips\n\nWhat would you like help with?',
      timestamp: new Date(),
      suggestions: ['How do I earn reputation?', 'Explain this thread', 'Suggest an approach'],
    },
  ]);
  const [isTyping, setIsTyping] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages, isTyping]);

  const handleSend = async () => {
    if (!inputValue.trim()) return;

    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: inputValue,
      timestamp: new Date(),
    };

    setMessages((prev) => [...prev, userMessage]);
    setInputValue('');
    setIsTyping(true);

    // Simulate AI response (replace with actual API call)
    setTimeout(() => {
      const assistantMessage: Message = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: 'I\'d be happy to help with that! This is a simulated response. In production, this would connect to your AI backend for intelligent responses based on the thread context and your question.',
        timestamp: new Date(),
        suggestions: ['Tell me more', 'Show me an example', 'Try another approach'],
      };
      setMessages((prev) => [...prev, assistantMessage]);
      setIsTyping(false);
    }, 1500);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleSuggestionClick = (suggestion: string) => {
    setInputValue(suggestion);
  };

  if (!isOpen) {
    return (
      <Button
        onClick={() => setIsOpen(true)}
        className={cn(
          'fixed bottom-6 right-6 h-14 w-14 rounded-full bg-gradient-to-r from-indigo-600 to-violet-600 p-0 shadow-lg shadow-indigo-500/25 hover:from-indigo-500 hover:to-violet-500',
          className
        )}
        aria-label="Open AI assistant"
      >
        <Sparkles className="h-6 w-6 text-white" />
      </Button>
    );
  }

  return (
    <div
      className={cn(
        'fixed bottom-6 right-6 z-50 flex flex-col overflow-hidden rounded-2xl border border-slate-800 bg-slate-900 shadow-2xl transition-all duration-300',
        isMinimized ? 'h-14 w-72' : 'h-[500px] w-96',
        className
      )}
    >
      {/* Header */}
      <div className="flex h-14 items-center justify-between border-b border-slate-800 bg-slate-950/50 px-4">
        <div className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-indigo-500 to-violet-600">
            <Bot className="h-4 w-4 text-white" />
          </div>
          <div>
            <p className="text-sm font-medium text-white">Flywheel AI</p>
            <p className="text-xs text-slate-500">Always here to help</p>
          </div>
        </div>
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-slate-400 hover:text-slate-200"
            onClick={() => setIsMinimized(!isMinimized)}
          >
            <ChevronDown
              className={cn('h-4 w-4 transition-transform', isMinimized && 'rotate-180')}
            />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-slate-400 hover:text-slate-200"
            onClick={() => setIsOpen(false)}
          >
            <X className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {!isMinimized && (
        <>
          {/* Messages */}
          <ScrollArea className="flex-1 p-4" ref={scrollRef}>
            <div className="space-y-4">
              {messages.map((message) => (
                <div
                  key={message.id}
                  className={cn(
                    'flex gap-3',
                    message.role === 'user' && 'flex-row-reverse'
                  )}
                >
                  <div
                    className={cn(
                      'flex h-8 w-8 shrink-0 items-center justify-center rounded-lg',
                      message.role === 'assistant'
                        ? 'bg-gradient-to-br from-indigo-500 to-violet-600'
                        : 'bg-slate-700'
                    )}
                  >
                    {message.role === 'assistant' ? (
                      <Bot className="h-4 w-4 text-white" />
                    ) : (
                      <User className="h-4 w-4 text-slate-300" />
                    )}
                  </div>
                  <div className={cn('space-y-2', message.role === 'user' && 'items-end')}>                    <div
                      className={cn(
                        'rounded-2xl px-4 py-2.5 text-sm',
                        message.role === 'assistant'
                          ? 'bg-slate-800 text-slate-200'
                          : 'bg-indigo-600 text-white'
                      )}
                    >
                      {message.content.split('\n').map((line, i) => (
                        <p key={i} className={i > 0 ? 'mt-2' : ''}>
                          {line}
                        </p>
                      ))}
                    </div>

                    {/* Suggestions */}
                    {message.suggestions && message.suggestions.length > 0 && (
                      <div className="flex flex-wrap gap-2">
                        {message.suggestions.map((suggestion) => (
                          <button
                            key={suggestion}
                            onClick={() => handleSuggestionClick(suggestion)}
                            className="flex items-center gap-1 rounded-full border border-slate-700 bg-slate-800/50 px-2.5 py-1 text-xs text-slate-400 transition-colors hover:border-slate-600 hover:text-slate-300"
                          >
                            <Lightbulb className="h-3 w-3" />
                            {suggestion}
                          </button>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              ))}

              {isTyping && (
                <div className="flex gap-3">
                  <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-indigo-500 to-violet-600">
                    <Bot className="h-4 w-4 text-white" />
                  </div>
                  <div className="flex items-center gap-1 rounded-2xl bg-slate-800 px-4 py-2.5">
                    <span className="h-2 w-2 animate-bounce rounded-full bg-slate-500" style={{ animationDelay: '0ms' }} />
                    <span className="h-2 w-2 animate-bounce rounded-full bg-slate-500" style={{ animationDelay: '150ms' }} />
                    <span className="h-2 w-2 animate-bounce rounded-full bg-slate-500" style={{ animationDelay: '300ms' }} />
                  </div>
                </div>
              )}
            </div>
          </ScrollArea>

          {/* Quick Actions */}
          <div className="border-t border-slate-800 bg-slate-950/30 px-4 py-2">
            <div className="flex gap-2">
              <button className="flex items-center gap-1 rounded-full bg-slate-800/50 px-2.5 py-1 text-xs text-slate-400 transition-colors hover:bg-slate-800 hover:text-slate-300">
                <Wand2 className="h-3 w-3" />
                Optimize code
              </button>
              <button className="flex items-center gap-1 rounded-full bg-slate-800/50 px-2.5 py-1 text-xs text-slate-400 transition-colors hover:bg-slate-800 hover:text-slate-300">
                <Sparkles className="h-3 w-3" />
                Explain
              </button>
            </div>
          </div>

          {/* Input */}
          <div className="border-t border-slate-800 p-4">
            <div className="flex gap-2">
              <Input
                value={inputValue}
                onChange={(e) => setInputValue(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Ask anything..."
                className="flex-1 border-slate-700 bg-slate-800 text-sm text-slate-200 placeholder:text-slate-500 focus-visible:ring-indigo-500"
              />
              <Button
                onClick={handleSend}
                disabled={!inputValue.trim() || isTyping}
                size="icon"
                className="shrink-0 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50"
              >
                <Send className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
