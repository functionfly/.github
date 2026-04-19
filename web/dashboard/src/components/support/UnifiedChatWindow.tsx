/**
 * Unified chat: Support API (persistence, AI, escalation) + Flywheel-style UI
 * (minimize, suggestions, quick actions).
 */

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
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
import { toast } from 'sonner';

const WELCOME_SUGGESTIONS = [
  { icon: Rocket, text: 'How do I deploy a function?' },
  { icon: HelpCircle, text: 'What can you help me with on FunctionFly?' },
  { icon: AlertTriangle, text: 'My function is failing — how do I debug it?' },
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
    wsError,
    isConnecting,
    isConnected,
    reconnectAttempt,
  } = useSupportChat();

  const [inputValue, setInputValue] = useState('');
  const [isMinimized, setIsMinimized] = useState(false);
  const [isEmergencyDialogOpen, setIsEmergencyDialogOpen] = useState(false);
  const [emergencyFunctionId, setEmergencyFunctionId] = useState('');
  const [emergencyReason, setEmergencyReason] = useState('');
  const [isRequestingEmergency, setIsRequestingEmergency] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const rootRef = useRef<HTMLDivElement>(null);

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

  useEffect(() => {
    if (!isOpen || isMinimized || isEmergencyDialogOpen) return;

    const selector =
      'button,[href],input,select,textarea,[tabindex]:not([tabindex="-1"])';

    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault();
        closeChat();
        return;
      }

      if (e.key !== 'Tab') return;
      const root = rootRef.current;
      if (!root) return;

      const focusable = Array.from(root.querySelectorAll<HTMLElement>(selector)).filter(
        (el) => !el.hasAttribute('disabled') && el.tabIndex !== -1
      );

      if (focusable.length === 0) return;

      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      const active = document.activeElement as HTMLElement | null;

      if (!e.shiftKey && active === last) {
        e.preventDefault();
        first.focus();
      } else if (e.shiftKey && active === first) {
        e.preventDefault();
        last.focus();
      }
    };

    document.addEventListener('keydown', onKeyDown);
    return () => document.removeEventListener('keydown', onKeyDown);
  }, [isOpen, isMinimized, isEmergencyDialogOpen, closeChat]);

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

  const openEmergencyFix = () => {
    setIsEmergencyDialogOpen(true);
  };

  const submitEmergencyFix = async () => {
    const functionId = emergencyFunctionId.trim();
    const reason = emergencyReason.trim();
    if (!functionId) return;
    if (!reason) return;

    setIsRequestingEmergency(true);
    try {
      await requestEmergencyFix(functionId, reason);
      toast.success('Emergency fix request sent');
      setIsEmergencyDialogOpen(false);
      setEmergencyFunctionId('');
      setEmergencyReason('');
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Failed to request emergency fix');
    } finally {
      setIsRequestingEmergency(false);
    }
  };

  const insertPrompt = (text: string) => {
    setInputValue(text);
    inputRef.current?.focus();
  };

  if (!isOpen) return null;

  const title = conversation?.title ?? 'FunctionFly Assistant';
  const subtitle = staffOnline
    ? '🟢 Staff available now'
    : '✨ FlyMind — your AI copilot (functions, billing, debugging, and more)';
  const canSend = !!conversation && inputValue.trim().length > 0 && !isSending;
  const setupFailed = !isLoading && !conversation;

  const connectionDotClassName = wsError
    ? 'bg-red-500'
    : isConnecting
      ? 'bg-amber-500'
      : isConnected
        ? 'bg-emerald-500'
        : 'bg-amber-500';

  const connectionLabel = wsError
    ? 'Offline'
    : isConnecting
      ? `Connecting${reconnectAttempt > 0 ? ` (${reconnectAttempt})` : ''}`
      : conversation
        ? 'Connected'
        : isLoading
          ? 'Starting conversation...'
          : 'Not connected';

  return (
    <div
      ref={rootRef}
      className={cn(
        'fixed bottom-6 right-6 z-[9999] flex flex-col overflow-hidden rounded-2xl border shadow-2xl transition-all duration-300 ease-out',
        'bg-[var(--bg-secondary)] border-[var(--border-default)]',
        isMinimized
          ? 'h-14 w-80 translate-y-0'
          : 'h-[min(680px,calc(100vh-6rem))] w-[min(440px,calc(100vw-2rem))]',
        className
      )}
      style={{
        // Use CSS variables - values automatically adapt based on data-theme
        backgroundColor: 'var(--bg-secondary)',
        borderColor: 'var(--border-default)',
        // Ensure solid background by resetting any potential transparency
        '--bg-secondary': 'var(--bg-secondary)',
      } as React.CSSProperties}
      role="dialog"
      aria-label="FunctionFly support chat"
      aria-modal="true"
      tabIndex={-1}
    >
      {/* Header - Velocity Brand: Flame Orange → Altitude Cyan */}
      <div className="relative flex h-14 shrink-0 items-center justify-between overflow-hidden px-4">
        {/* Animated gradient background - Velocity Brand */}
        <div className="absolute inset-0 bg-gradient-to-r from-[#FF6B35] via-[#FF4F5E] to-[#00D4FF]" />
        <div className="absolute inset-0 bg-[linear-gradient(to_bottom_right,rgba(255,255,255,0.1),transparent,rgba(0,0,0,0.1))]" />
        
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
                    <div className="absolute h-full w-full animate-pulse rounded-full bg-[#FF6B35]/20" />
                    <div className="h-6 w-6 animate-spin rounded-full border-2 border-[#FF6B35] border-t-transparent" />
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
                    <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-[linear-gradient(to_bottom_right,#FF6B35,#FF4F5E,#00D4FF)] shadow-lg shadow-[#FF6B35]/20">
                      <Bot className="h-5 w-5 text-white" />
                    </div>
                    <div className="space-y-3 flex-1">
                      <div className="rounded-2xl rounded-tl-sm bg-[var(--bg-tertiary)] border border-[var(--border-subtle)] px-4 py-3.5 shadow-sm">
                        <p className="font-semibold text-[var(--text-primary)]">Hi there! 👋</p>
                        <p className="mt-2 text-sm leading-relaxed text-[var(--text-secondary)]">
                          I&apos;m FlyMind — your AI copilot for everything on FunctionFly: deploying
                          functions, debugging runs, billing, the registry, agents, and how the platform
                          works. Ask me anything. You can also talk to a human or request an emergency
                          fix when something is on fire.
                        </p>
                      </div>
                      
                      {/* Welcome Suggestions - Pill style */}
                      <div className="flex flex-wrap gap-2">
                        {WELCOME_SUGGESTIONS.map(({ icon: Icon, text }) => (
                          <button
                            key={text}
                            type="button"
                            onClick={() => insertPrompt(text)}
                            className="group flex items-center gap-2 rounded-full border border-[var(--border-default)] bg-[var(--bg-primary)] px-3.5 py-2 text-xs font-medium text-[var(--text-secondary)] transition-all duration-200 hover:border-[#FF6B35]/50 hover:bg-[#FF6B35]/10 hover:text-[#FF6B35] dark:hover:text-[#FF8C42]"
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
                        className="group flex items-center gap-2 rounded-xl border border-[var(--border-subtle)] bg-[var(--bg-primary)]/50 p-3 text-left transition-all duration-200 hover:border-[#FF6B35]/30 hover:bg-[#FF6B35]/5"
                      >
                        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-[linear-gradient(to_bottom_right,rgba(255,107,53,0.2),rgba(0,212,255,0.2))] transition-colors group-hover:bg-[linear-gradient(to_bottom_right,rgba(255,107,53,0.3),rgba(0,212,255,0.3))]">
                          <Icon className="h-4 w-4 text-[#FF6B35] dark:text-[#FF8C42]" />
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
                            ? 'bg-[linear-gradient(to_bottom_right,#10b981,#0d9488)] shadow-lg shadow-emerald-500/20'
                            : 'bg-[linear-gradient(to_bottom_right,#FF6B35,#FF4F5E,#00D4FF)] shadow-lg shadow-[#FF6B35]/20'
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
                              ? 'rounded-tr-sm bg-[linear-gradient(to_bottom_right,#00D4FF,#5B7CF5)] text-white'
                              : isStaff
                                ? 'rounded-tl-sm bg-[linear-gradient(to_bottom_right,#059669,#0d9488)] text-white'
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
                  <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-[linear-gradient(to_bottom_right,#FF6B35,#FF4F5E,#00D4FF)] shadow-lg shadow-[#FF6B35]/20">
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
                  onClick={openEmergencyFix}
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
                  className="h-11 flex-1 rounded-xl border-[var(--border-default)] bg-[var(--bg-secondary)] pr-12 text-sm text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus-visible:ring-2 focus-visible:ring-[#FF6B35]/50 focus-visible:border-[#FF6B35]/50 transition-all"
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
                className="h-11 w-11 shrink-0 rounded-xl bg-[linear-gradient(to_bottom_right,#FF6B35,#E85A2A)] hover:brightness-110 disabled:opacity-40 disabled:cursor-not-allowed shadow-lg shadow-[#FF6B35]/30 transition-all duration-200 hover:scale-105 active:scale-95"
              >
                <Send className="h-4 w-4 text-white" />
              </Button>
            </div>
            
            {/* Status Bar */}
            <div className="mt-2 flex items-center justify-between text-[10px] text-[var(--text-muted)]">
              <div className="flex items-center gap-1.5">
                <div className={cn('h-1.5 w-1.5 rounded-full', connectionDotClassName)} />
                <span>{connectionLabel}</span>
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

      {/* Emergency Fix Modal */}
      <Dialog open={isEmergencyDialogOpen} onOpenChange={setIsEmergencyDialogOpen}>
        <DialogContent className="sm:max-w-md bg-bg-secondary border border-border-default">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <AlertTriangle className="h-5 w-5 text-red-400" />
              Emergency Fix Request
            </DialogTitle>
            <DialogDescription>
              Provide the function ID and what’s happening so support can prioritize quickly.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-3">
            <div className="space-y-1">
              <label className="text-sm text-text-secondary" htmlFor="emergency-function-id">
                Function ID
              </label>
              <Input
                id="emergency-function-id"
                value={emergencyFunctionId}
                onChange={(e) => setEmergencyFunctionId(e.target.value)}
                placeholder="e.g. fn_123..."
                disabled={isRequestingEmergency}
              />
            </div>
            <div className="space-y-1">
              <label className="text-sm text-text-secondary" htmlFor="emergency-reason">
                What’s the issue?
              </label>
              <textarea
                id="emergency-reason"
                value={emergencyReason}
                onChange={(e) => setEmergencyReason(e.target.value)}
                placeholder="What happened, expected behavior, and any relevant logs..."
                disabled={isRequestingEmergency}
                className="w-full min-h-[96px] resize-y rounded-xl border border-border-default bg-[var(--bg-secondary)] px-3 py-2 text-sm text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus-visible:ring-2 focus-visible:ring-[#FF6B35]/50 focus-visible:border-[#FF6B35]/50"
              />
            </div>
          </div>
          <DialogFooter className="gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => setIsEmergencyDialogOpen(false)}
              disabled={isRequestingEmergency}
            >
              Cancel
            </Button>
            <Button
              type="button"
              variant="destructive"
              onClick={() => void submitEmergencyFix()}
              disabled={isRequestingEmergency || !emergencyFunctionId.trim() || !emergencyReason.trim()}
            >
              {isRequestingEmergency ? 'Requesting...' : 'Request Emergency Fix'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export default UnifiedChatWindow;
