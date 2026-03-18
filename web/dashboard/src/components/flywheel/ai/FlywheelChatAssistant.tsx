/**
 * FlywheelChatAssistant - Bottom-right AI assistant chat
 * In production (when VITE_AI_SERVICE_URL is set), uses the AI service completion API.
 */

import { complete, isAiServiceConfigured } from '@/api/ai';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { cn } from '@/lib/utils';
import { Bot, ChevronDown, Lightbulb, Send, Sparkles, User, Wand2, X } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';

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
      content:
        "Hi! I'm your Flywheel AI assistant. I can help you with:\n\n• Understanding problems\n• Suggesting approaches\n• Explaining solutions\n• Code optimization tips\n\nWhat would you like help with?",
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

    const addAssistantReply = (content: string, suggestions?: string[]) => {
      const assistantMessage: Message = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content,
        timestamp: new Date(),
        suggestions,
      };
      setMessages((prev) => [...prev, assistantMessage]);
      setIsTyping(false);
    };

    if (isAiServiceConfigured()) {
      try {
        const systemPrompt = `You are the Flywheel AI assistant. You help users with:
- Understanding problems and thread context
- Suggesting approaches and explaining solutions
- Code optimization tips and earning reputation on the Flywheel network
Keep replies concise and helpful. Use markdown for code or lists when useful.`;
        const apiMessages = [
          { role: 'system' as const, content: systemPrompt },
          ...messages
            .filter((m) => m.role === 'user' || m.role === 'assistant')
            .map((m) => ({ role: m.role, content: m.content })),
          { role: 'user' as const, content: userMessage.content },
        ];
        const response = await complete({
          messages: apiMessages,
          temperature: 0.7,
          max_tokens: 1024,
        });
        addAssistantReply(response.content.trim(), [
          'Tell me more',
          'Show me an example',
          'Try another approach',
        ]);
      } catch (err) {
        const message =
          err instanceof Error ? err.message : 'The assistant is temporarily unavailable.';
        addAssistantReply(`Sorry, I couldn’t complete that. ${message}`, ['Try again', 'Go Home']);
      }
    } else {
      // No AI service configured: simulated response for dev/demo
      setTimeout(() => {
        addAssistantReply(
          "I'd be happy to help with that! This is a simulated response. Set VITE_AI_SERVICE_URL to your FlyMind ai-service to get real AI replies.",
          ['Tell me more', 'Show me an example', 'Try another approach']
        );
      }, 800);
    }
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
          'flywheel-chat-fab fixed bottom-8 right-6 z-50 h-14 w-14 rounded-full bg-gradient-to-r from-indigo-600 to-violet-600 p-0 shadow-lg shadow-indigo-500/25 hover:from-indigo-500 hover:to-violet-500',
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
        'flywheel-ai-panel fixed bottom-8 right-6 z-50 flex flex-col overflow-hidden rounded-2xl border border-border-default bg-bg-elevated shadow-xl transition-all duration-300',
        isMinimized ? 'h-14 w-64' : 'h-[380px] min-h-[280px] max-h-[calc(100vh-10rem)] w-80',
        className
      )}
      role="dialog"
      aria-label="Flywheel AI assistant"
    >
      {/* Header */}
      <div className="flex h-14 items-center justify-between border-b border-border-default bg-bg-secondary/80 px-4">
        <div className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-indigo-500 to-violet-600">
            <Bot className="h-4 w-4 text-white" />
          </div>
          <div>
            <p className="text-sm font-medium text-text-primary">Flywheel AI</p>
            <p className="text-xs text-text-muted">Always here to help</p>
          </div>
        </div>
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-text-muted hover:text-text-primary"
            onClick={() => setIsMinimized(!isMinimized)}
            aria-label={isMinimized ? 'Expand chat' : 'Minimize chat'}
            aria-expanded={!isMinimized}
          >
            <ChevronDown
              className={cn('h-4 w-4 transition-transform', isMinimized && 'rotate-180')}
            />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-text-muted hover:text-text-primary"
            onClick={() => setIsOpen(false)}
            aria-label="Close chat"
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
                  className={cn('flex gap-3', message.role === 'user' && 'flex-row-reverse')}
                >
                  <div
                    className={cn(
                      'flex h-8 w-8 shrink-0 items-center justify-center rounded-lg',
                      message.role === 'assistant'
                        ? 'bg-gradient-to-br from-indigo-500 to-violet-600'
                        : 'bg-bg-hover'
                    )}
                  >
                    {message.role === 'assistant' ? (
                      <Bot className="h-4 w-4 text-white" />
                    ) : (
                      <User className="h-4 w-4 text-text-secondary" />
                    )}
                  </div>
                  <div className={cn('space-y-2', message.role === 'user' && 'items-end')}>
                    {' '}
                    <div
                      className={cn(
                        'rounded-2xl px-4 py-2.5 text-sm',
                        message.role === 'assistant'
                          ? 'bg-bg-hover text-text-primary'
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
                            className="flywheel-ai-suggestion flex items-center gap-1 rounded-full border border-border-default bg-bg-hover/60 px-2.5 py-1 text-xs text-text-secondary transition-colors hover:border-border-strong hover:bg-bg-hover hover:text-text-primary"
                          >
                            <Lightbulb className="h-3 w-3 shrink-0" />
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
                  <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br from-indigo-500 to-violet-600">
                    <Bot className="h-4 w-4 text-white" />
                  </div>
                  <div className="flex items-center gap-1 rounded-2xl bg-bg-hover px-4 py-2.5">
                    <span
                      className="h-2 w-2 animate-bounce rounded-full bg-text-muted"
                      style={{ animationDelay: '0ms' }}
                    />
                    <span
                      className="h-2 w-2 animate-bounce rounded-full bg-text-muted"
                      style={{ animationDelay: '150ms' }}
                    />
                    <span
                      className="h-2 w-2 animate-bounce rounded-full bg-text-muted"
                      style={{ animationDelay: '300ms' }}
                    />
                  </div>
                </div>
              )}
            </div>
          </ScrollArea>

          {/* Quick Actions */}
          <div className="border-t border-border-default bg-bg-primary/30 px-4 py-2">
            <div className="flex gap-2">
              <button className="flex items-center gap-1 rounded-full bg-bg-hover/50 px-2.5 py-1 text-xs text-text-muted transition-colors hover:bg-bg-hover hover:text-text-secondary">
                <Wand2 className="h-3 w-3" />
                Optimize code
              </button>
              <button className="flex items-center gap-1 rounded-full bg-bg-hover/50 px-2.5 py-1 text-xs text-text-muted transition-colors hover:bg-bg-hover hover:text-text-secondary">
                <Sparkles className="h-3 w-3" />
                Explain
              </button>
            </div>
          </div>

          {/* Input */}
          <div className="border-t border-border-default p-4">
            <div className="flex gap-2">
              <Input
                value={inputValue}
                onChange={(e) => setInputValue(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Ask anything..."
                className="flex-1 border-border-default bg-bg-hover text-sm text-text-primary placeholder:text-text-muted focus-visible:ring-indigo-500"
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
