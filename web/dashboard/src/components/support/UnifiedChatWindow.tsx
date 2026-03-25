/**
 * Unified chat: Support API (persistence, AI, escalation) + Flywheel-style UI
 * (minimize, suggestions, quick actions).
 */

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { cn } from '@/lib/utils';
import {
  AlertTriangle,
  Bot,
  ChevronDown,
  Cpu,
  FileCode,
  HelpCircle,
  MessageSquare,
  Rocket,
  Send,
  Sparkles,
  User,
  Users,
  Wand2,
  X,
  Zap,
} from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { SupportMessage, useSupportChat } from './SupportChat';

const WELCOME_SUGGESTIONS = [
  { icon: Rocket, text: 'How do I deploy a function?' },
  { icon: AlertTriangle, text: 'My function is failing' },
  { icon: HelpCircle, text: 'Explain how billing works' },
];

const QUICK_ACTIONS = [
  { icon: Wand2, label: 'Optimize code', prompt: 'Help me optimize my function for performance and lower cold-start time.' },
  { icon: Sparkles, label: 'Explain', prompt: 'Explain how FunctionFly runs my code and how to read execution logs.' },
  { icon: FileCode, label: 'Debug', prompt: 'Help me debug an error in my function.' },
  { icon: Cpu, label: 'Best practices', prompt: 'What are the best practices for writing efficient functions?' },
];

interface UnifiedChatWindowProps {
  className?: string;
}

function messageRole(message: SupportMessage): 'user' | 'assistant' | 'system' {
  if (message.author_type === 'user') return 'user';
  if (message.author_type === 'system') return 'system';
  return 'assistant';
}

function formatMessageTime(timestamp: string): string {
  const date = new Date(timestamp);
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

export function UnifiedChatWindow({ className }: UnifiedChatWindowProps) {
  const {
    isOpen,
    conversation,
    messages,
    isLoading,
    isSending,
    staffOnline,
    openError,
    closeChat,
    sendMessage,
    escalateToHuman,
    requestEmergencyFix,
  } = useSupportChat();

  const [inputValue, setInputValue] = useState('');
  const [isMinimized, setIsMinimized] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages, isSending, isLoading]);

  // Focus input when chat opens
  useEffect(() => {
    if (isOpen && !isMinimized && inputRef.current) {
      setTimeout(() => inputRef.current?.focus(), 100);
    }
  }, [isOpen, isMinimized]);

  const handleSend = async () => {
    if (!inputValue.trim() || isSending || !conversation) return;
    const content = inputValue.trim();
    setInputValue('');
    await sendMessage(content);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleEmergencyFix = () => {
    const functionId = prompt('Enter the function ID that is failing:');
    if (!functionId) return;
    const reason = prompt('Describe the issue (what happened, expected behavior):');
    if (!reason) return;
    requestEmergencyFix(functionId, reason);
  };

  const insertPrompt = (text: string) => {
    setInputValue(text);
    inputRef.current?.focus();
  };

  if (!isOpen) return null;

  const title = conversation?.title ?? 'FunctionFly Support';
  const subtitle = staffOnline 
    ? '🟢 Staff available now' 
    : '✨ AI powered support';
  const canSend = !!conversation && inputValue.trim().length > 0 && !isSending;
  const setupFailed = !isLoading && !conversation;

  return (
    <div
      className={cn(
        'fixed bottom-6 right-6 z-[9999] flex flex-col overflow-hidden rounded-2xl border shadow-2xl transition-all duration-300 ease-out',
        'bg-[var(--bg-secondary)] border-[var(--border-default)]',
        isMinimized
          ? 'h-14 w-80 translate-y-0'
          : 'h-[min(680px,calc(100vh-6rem))] w-[min(440px,calc(100vw-2rem))]',
        className
      )}
      role="dialog"
      aria-label="FunctionFly support chat"
    >
      {/* Header - Enhanced with glass effect */}
      <div className="relative flex h-14 shrink-0 items-center justify-between overflow-hidden px-4">
        {/* Animated gradient background */}
        <div className="absolute inset-0 bg-gradient-to-r from-indigo-600 via-violet-600 to-purple-600" />
        <div className="absolute inset-0 bg-gradient-to-br from-white/10 via-transparent to-black/10" />
        
        {/* Animated shine effect */}
        <div className="absolute inset-0 -translate-x-full animate-shimmer bg-gradient-to-r from-transparent via-white/10 to-transparent" />
        
        <div className="relative flex min-w-0 items-center gap-3">
          <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-white/20 backdrop-blur-sm ring-1 ring-white/30">
            <Sparkles className="h-4 w-4 text-white" />
          </div>
          <div className="min-w-0">
            <p className="truncate text-sm font-semibold text-white">{title}</p>
            <p className="truncate text-xs text-white/80 font-medium">{subtitle}</p>
          </div>
        </div>
        <div className="relative flex shrink-0 items-center gap-1">
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-white/90 hover:bg-white/20 hover:text-white"
            onClick={() => setIsMinimized(!isMinimized)}
            aria-label={isMinimized ? 'Expand chat' : 'Minimize chat'}
            aria-expanded={!isMinimized}
          >
            <ChevronDown
              className={cn('h-4 w-4 transition-transform duration-200', isMinimized && 'rotate-180')}
            />
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-white/90 hover:bg-white/20 hover:text-white"
            onClick={closeChat}
            aria-label="Close chat"
          >
            <X className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {!isMinimized && (
        <>
          {/* Messages Area */}
          <div
            ref={scrollRef}
            className="min-h-0 flex-1 overflow-y-auto overscroll-contain px-4 py-4 scrollbar-thin scrollbar-thumb-[var(--border-default)] scrollbar-track-transparent"
          >
            <div className="space-y-5">
              {isLoading && (
                <div className="flex flex-col items-center justify-center gap-3 py-12 text-sm text-[var(--text-muted)]">
                  <div className="relative flex h-10 w-10 items-center justify-center">
                    <div className="absolute h-full w-full animate-pulse rounded-full bg-indigo-500/20" />
                    <div className="h-6 w-6 animate-spin rounded-full border-2 border-indigo-500 border-t-transparent" />
                  </div>
                  <span>Loading conversation…</span>
                </div>
              )}

              {setupFailed && (
                <div className="rounded-xl border border-amber-500/30 bg-amber-500/10 px-4 py-4 text-center">
                  <AlertTriangle className="mx-auto mb-2 h-5 w-5 text-amber-600 dark:text-amber-400" />
                  <p className="text-sm font-medium text-amber-900 dark:text-amber-100">
                    {openError || 'Could not start support chat'}
                  </p>
                  <p className="mt-1 text-xs text-amber-700 dark:text-amber-300">
                    Check your connection and try again
                  </p>
                </div>
              )}

              {!isLoading && !setupFailed && messages.length === 0 && (
                <div className="space-y-6">
                  {/* Welcome Message */}
                  <div className="flex gap-3">
                    <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-indigo-500 via-violet-500 to-purple-600 shadow-lg shadow-indigo-500/20">
                      <Bot className="h-5 w-5 text-white" />
                    </div>
                    <div className="space-y-3 flex-1">
                      <div className="rounded-2xl rounded-tl-sm bg-[var(--bg-tertiary)] border border-[var(--border-subtle)] px-4 py-3.5 shadow-sm">
                        <p className="font-semibold text-[var(--text-primary)]">Hi there! 👋</p>
                        <p className="mt-2 text-sm leading-relaxed text-[var(--text-secondary)]">
                          I&apos;m your AI support assistant. I can help with deployment, functions,
                          errors, and general questions. If needed, you can talk to a human or
                          request an emergency fix.
                        </p>
                      </div>
                      
                      {/* Welcome Suggestions - Pill style */}
                      <div className="flex flex-wrap gap-2">
                        {WELCOME_SUGGESTIONS.map(({ icon: Icon, text }) => (
                          <button
                            key={text}
                            type="button"
                            onClick={() => insertPrompt(text)}
                            className="group flex items-center gap-2 rounded-full border border-[var(--border-default)] bg-[var(--bg-primary)] px-3.5 py-2 text-xs font-medium text-[var(--text-secondary)] transition-all duration-200 hover:border-indigo-500/50 hover:bg-indigo-500/10 hover:text-indigo-600 dark:hover:text-indigo-400"
                          >
                            <Icon className="h-3.5 w-3.5 transition-colors" />
                            <span>{text}</span>
                          </button>
                        ))}
                      </div>
                    </div>
                  </div>

                  {/* Quick Actions Grid */}
                  <div className="grid grid-cols-2 gap-2 pt-2">
                    {QUICK_ACTIONS.map(({ icon: Icon, label, prompt }) => (
                      <button
                        key={label}
                        type="button"
                        onClick={() => insertPrompt(prompt)}
                        className="group flex items-center gap-2 rounded-xl border border-[var(--border-subtle)] bg-[var(--bg-primary)]/50 p-3 text-left transition-all duration-200 hover:border-indigo-500/30 hover:bg-indigo-500/5"
                      >
                        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br from-indigo-500/20 to-violet-500/20 transition-colors group-hover:from-indigo-500/30 group-hover:to-violet-500/30">
                          <Icon className="h-4 w-4 text-indigo-600 dark:text-indigo-400" />
                        </div>
                        <span className="text-xs font-medium text-[var(--text-secondary)] group-hover:text-[var(--text-primary)]">
                          {label}
                        </span>
                      </button>
                    ))}
                  </div>
                </div>
              )}

              {/* Message List */}
              {!isLoading &&
                messages.map((message, index) => {
                  const role = messageRole(message);
                  if (role === 'system') {
                    return (
                      <div
                        key={message.id}
                        className="flex items-center justify-center gap-2 py-2"
                      >
                        <div className="rounded-full bg-amber-500/10 px-4 py-1.5 text-center text-xs font-medium text-amber-700 dark:text-amber-300 border border-amber-500/20">
                          {message.content}
                        </div>
                      </div>
                    );
                  }
                  const isUser = role === 'user';
                  const isStaff = message.author_type === 'staff';
                  const showAvatar = index === 0 || messages[index - 1]?.author_type !== message.author_type;
                  
                  return (
                    <div
                      key={message.id}
                      className={cn('flex gap-3', isUser && 'flex-row-reverse')}
                    >
                      {/* Avatar */}
                      <div className={cn(
                        'flex h-8 w-8 shrink-0 items-center justify-center rounded-lg transition-all',
                        !showAvatar && 'opacity-0',
                        isUser
                          ? 'bg-[var(--bg-hover)]'
                          : isStaff
                            ? 'bg-gradient-to-br from-emerald-500 to-teal-600 shadow-lg shadow-emerald-500/20'
                            : 'bg-gradient-to-br from-indigo-500 via-violet-500 to-purple-600 shadow-lg shadow-indigo-500/20'
                      )}>
                        {isUser ? (
                          <User className="h-4 w-4 text-[var(--text-secondary)]" />
                        ) : (
                          <Bot className="h-4 w-4 text-white" />
                        )}
                      </div>
                      
                      {/* Message Content */}
                      <div className={cn('min-w-0 space-y-1', isUser && 'items-end text-right')}
                        style={{ maxWidth: 'calc(100% - 48px)' }}
                      >
                        {/* Sender Name */}
                        {!isUser && showAvatar && (
                          <p className="text-[11px] font-semibold text-[var(--text-muted)] uppercase tracking-wide">
                            {isStaff ? 'Support Agent' : 'AI Assistant'}
                          </p>
                        )}
                        
                        {/* Message Bubble */}
                        <div
                          className={cn(
                            'inline-block rounded-2xl px-4 py-2.5 text-sm leading-relaxed shadow-sm transition-all',
                            isUser
                              ? 'rounded-tr-sm bg-gradient-to-br from-indigo-600 to-violet-600 text-white'
                              : isStaff
                                ? 'rounded-tl-sm bg-gradient-to-br from-emerald-500 to-teal-600 text-white'
                                : 'rounded-tl-sm bg-[var(--bg-tertiary)] border border-[var(--border-subtle)] text-[var(--text-primary)]'
                          )}
                        >
                          {message.content.split('\n').map((line, i) => (
                            <p key={i} className={cn(i > 0 && 'mt-2')}>
                              {line}
                            </p>
                          ))}
                        </div>
                        
                        {/* Meta Info */}
                        <div className={cn('flex items-center gap-2 text-[10px] text-[var(--text-muted)]', isUser && 'justify-end')}>
                          <span>{formatMessageTime(message.created_at)}</span>
                          {message.ai_confidence != null && message.author_type === 'ai' && (
                            <span className="flex items-center gap-1">
                              <Zap className="h-3 w-3" />
                              {Math.round(message.ai_confidence * 100)}% confident
                            </span>
                          )}
                        </div>
                      </div>
                    </div>
                  );
                })}

              {/* Typing Indicator */}
              {isSending && (
                <div className="flex gap-3">
                  <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br from-indigo-500 via-violet-500 to-purple-600 shadow-lg shadow-indigo-500/20">
                    <Bot className="h-4 w-4 text-white" />
                  </div>
                  <div className="flex items-center gap-1 rounded-2xl rounded-tl-sm bg-[var(--bg-tertiary)] border border-[var(--border-subtle)] px-4 py-3 shadow-sm">
                    <span
                      className="h-2 w-2 animate-bounce rounded-full bg-[var(--text-muted)]"
                      style={{ animationDelay: '0ms' }}
                    />
                    <span
                      className="h-2 w-2 animate-bounce rounded-full bg-[var(--text-muted)]"
                      style={{ animationDelay: '150ms' }}
                    />
                    <span
                      className="h-2 w-2 animate-bounce rounded-full bg-[var(--text-muted)]"
                      style={{ animationDelay: '300ms' }}
                    />
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Action Bar - Escalation */}
          {conversation && conversation.status === 'active' && (
            <div className="shrink-0 border-t border-[var(--border-default)] bg-[var(--bg-primary)]/50 px-4 py-3">
              <div className="flex gap-2">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="h-9 flex-1 gap-2 text-xs font-medium border-[var(--border-default)] bg-[var(--bg-secondary)] hover:bg-[var(--bg-hover)] hover:border-[var(--border-default)]"
                  onClick={escalateToHuman}
                  title="Connect with a human support agent"
                >
                  <Users className="h-3.5 w-3.5 text-emerald-500" />
                  Talk to human
                </Button>
                <Button
                  type="button"
                  variant="destructive"
                  size="sm"
                  className="h-9 flex-1 gap-2 text-xs font-medium bg-gradient-to-r from-red-600 to-red-500 hover:from-red-500 hover:to-red-400 border-0"
                  onClick={handleEmergencyFix}
                  title="Request immediate help for production issues"
                >
                  <AlertTriangle className="h-3.5 w-3.5" />
                  Emergency fix
                </Button>
              </div>
            </div>
          )}

          {/* Input Area */}
          <div className="shrink-0 border-t border-[var(--border-default)] bg-[var(--bg-primary)] p-4">
            <div className="relative flex gap-2">
              <div className="relative flex-1">
                <Input
                  ref={inputRef}
                  value={inputValue}
                  onChange={(e) => setInputValue(e.target.value)}
                  onKeyDown={handleKeyDown}
                  placeholder="Ask anything..."
                  disabled={!conversation || isSending}
                  className="h-11 flex-1 rounded-xl border-[var(--border-default)] bg-[var(--bg-secondary)] pr-12 text-sm text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus-visible:ring-2 focus-visible:ring-indigo-500/50 focus-visible:border-indigo-500/50 transition-all"
                />
                {inputValue.length > 0 && (
                  <span className="absolute right-3 top-1/2 -translate-y-1/2 text-[10px] text-[var(--text-muted)]">
                    ↵
                  </span>
                )}
              </div>
              <Button
                type="button"
                onClick={handleSend}
                disabled={!canSend}
                size="icon"
                className="h-11 w-11 shrink-0 rounded-xl bg-gradient-to-br from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 disabled:opacity-40 disabled:cursor-not-allowed shadow-lg shadow-indigo-500/20 transition-all duration-200 hover:scale-105 active:scale-95"
              >
                <Send className="h-4 w-4 text-white" />
              </Button>
            </div>
            
            {/* Status Bar */}
            <div className="mt-2 flex items-center justify-between text-[10px] text-[var(--text-muted)]">
              <div className="flex items-center gap-1.5">
                <div className={cn(
                  'h-1.5 w-1.5 rounded-full',
                  conversation ? 'bg-emerald-500' : 'bg-amber-500'
                )} />
                <span>
                  {isLoading 
                    ? 'Connecting...' 
                    : conversation 
                      ? 'Connected' 
                      : 'Starting conversation...'}
                </span>
              </div>
              {conversation && (
                <span className="flex items-center gap-1">
                  <MessageSquare className="h-3 w-3" />
                  {messages.length} message{messages.length !== 1 ? 's' : ''}
                </span>
              )}
            </div>
          </div>
        </>
      )}
    </div>
  );
}

export default UnifiedChatWindow;
