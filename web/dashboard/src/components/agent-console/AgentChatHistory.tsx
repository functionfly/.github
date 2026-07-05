import { agentApi } from '@/api/agent';
import { groupMessagesByDate } from '@/components/conversations/conversation-ui';
import { SealedButton } from '@/components/containment';
import { format, formatDistanceToNow } from 'date-fns';
import { ArrowDown, Loader2, MessageSquare, Plus, Search, X } from 'lucide-react';
import { useCallback, useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { ChatThinking } from './ChatThinking';
import './agent-chat.css';

interface ThinkingData {
  content: string;
  tokens: number;
}

interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  model?: string;
  thinking?: ThinkingData;
  created_at: string;
}

interface AgentChatHistoryProps {
  agentId: string;
  agentName: string;
  model: string;
  sessionId?: string | null;
  onSessionCreated?: (sessionId: string) => void;
}

const PAGE_SIZE = 50;

export function AgentChatHistory({ agentId, agentName, model, sessionId, onSessionCreated }: AgentChatHistoryProps) {
  const { t } = useTranslation();
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState('');
  const [sending, setSending] = useState(false);
  const [loading, setLoading] = useState(true);
  const [hasMore, setHasMore] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [showJump, setShowJump] = useState(false);
  const [creating, setCreating] = useState(false);
  const messagesRef = useRef<HTMLDivElement>(null);
  const bottomRef = useRef<HTMLDivElement>(null);
  const prevLengthRef = useRef(0);

  const scrollToBottom = useCallback((behavior: ScrollBehavior = 'smooth') => {
    bottomRef.current?.scrollIntoView({ behavior });
  }, []);

  const handleNewChat = async () => {
    if (creating) return;
    setCreating(true);
    try {
      const res = await agentApi.createChatSession(agentId);
      if (res.ok && res.session_id) {
        onSessionCreated?.(res.session_id);
      }
    } catch {
      // silently ignore
    } finally {
      setCreating(false);
    }
  };

  useEffect(() => {
    if (!loading && messages.length > 0 && messages.length > prevLengthRef.current) {
      const el = messagesRef.current;
      if (el) {
        const nearBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 120;
        if (nearBottom || prevLengthRef.current === 0) {
          scrollToBottom(prevLengthRef.current === 0 ? 'auto' : 'smooth');
        } else {
          setShowJump(true);
        }
      }
    }
    prevLengthRef.current = messages.length;
  }, [messages.length, loading, scrollToBottom]);

  // Load messages when sessionId or agentId changes
  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setMessages([]);
    setHasMore(false);
    setSearchQuery('');

    agentApi.getChatHistory(agentId, PAGE_SIZE, 0, sessionId ?? undefined).then((res) => {
      if (cancelled) return;
      const msgs = (res.messages ?? []).map((m) => ({
        id: m.id || `msg-${m.created_at}-${Math.random().toString(36).slice(2, 9)}`,
        role: m.role as 'user' | 'assistant',
        content: m.content,
        model: m.model,
        thinking: m.metadata?.thinking,
        created_at: m.created_at,
      }));
      setMessages(msgs);
      setHasMore(msgs.length >= PAGE_SIZE);
    }).catch(() => {
      if (!cancelled) setMessages([]);
    }).finally(() => {
      if (!cancelled) setLoading(false);
    });
    return () => { cancelled = true; };
  }, [agentId, sessionId]);

  const handleLoadOlder = async () => {
    if (loadingMore || !hasMore) return;
    setLoadingMore(true);
    try {
      const res = await agentApi.getChatHistory(agentId, PAGE_SIZE, messages.length, sessionId ?? undefined);
      const older = (res.messages ?? []).map((m) => ({
        id: m.id || `msg-${m.created_at}-${Math.random().toString(36).slice(2, 9)}`,
        role: m.role as 'user' | 'assistant',
        content: m.content,
        model: m.model,
        thinking: m.metadata?.thinking,
        created_at: m.created_at,
      }));
      setMessages((prev) => [...older, ...prev]);
      setHasMore(older.length >= PAGE_SIZE);
    } catch {
      // silently ignore
    } finally {
      setLoadingMore(false);
    }
  };

  const handleSend = async () => {
    const text = input.trim();
    if (!text || sending) return;
    setInput('');
    const now = new Date().toISOString();
    const userId = `user-${now}-${Math.random().toString(36).slice(2, 9)}`;
    setMessages((prev) => [...prev, { id: userId, role: 'user', content: text, created_at: now }]);
    setSending(true);
    try {
      const res = await agentApi.agentChat(agentId, text, sessionId ?? undefined);
      // If this was the first message in a new session, notify parent
      if (res.session_id && !sessionId && onSessionCreated) {
        onSessionCreated(res.session_id);
      }
      setMessages((prev) => [
        ...prev,
        {
          id: `assistant-${new Date().toISOString()}-${Math.random().toString(36).slice(2, 9)}`,
          role: 'assistant',
          content: res.message || '(no response)',
          model: res.model,
          thinking: res.thinking,
          created_at: new Date().toISOString(),
        },
      ]);
    } catch (err) {
      setMessages((prev) => [
        ...prev,
        { id: `error-${new Date().toISOString()}-${Math.random().toString(36).slice(2, 9)}`, role: 'assistant', content: `Error: ${err instanceof Error ? err.message : 'Failed to reach agent'}`, created_at: new Date().toISOString() },
      ]);
    } finally {
      setSending(false);
    }
  };

  const handleScroll = () => {
    const el = messagesRef.current;
    if (!el) return;
    const nearBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 80;
    setShowJump((prev) => (prev === !nearBottom ? prev : !nearBottom));
  };

  const filtered = searchQuery.trim()
    ? messages.filter((m) => m.content.toLowerCase().includes(searchQuery.toLowerCase()))
    : messages;

  const groups = groupMessagesByDate(filtered);

  return (
    <div className="ach-wrap">
      {/* Toolbar */}
      <div className="ach-toolbar">
        <div className="ach-search">
          <Search className="ach-search__icon" />
          <input
            className="ach-search__input"
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder={t('agentDetail.chatSearchPlaceholder', 'Search messages…')}
          />
          {searchQuery && (
            <>
              <span className="ach-search__count">
                {filtered.length} / {messages.length}
              </span>
              <button className="ach-search__clear" onClick={() => setSearchQuery('')} aria-label="Clear search">
                <X style={{ width: 12, height: 12 }} />
              </button>
            </>
          )}
        </div>
        <button className="ach-new-chat" onClick={handleNewChat} disabled={creating} title="New chat">
          {creating ? <Loader2 className="ach-new-chat__spinner" /> : <Plus style={{ width: 14, height: 14 }} />}
          <span>New Chat</span>
        </button>
      </div>

      {/* Messages */}
      <div className="ach-messages" ref={messagesRef} onScroll={handleScroll}>
        {/* Load older */}
        {hasMore && !searchQuery && (
          <div className="ach-load-older">
            <button className="ach-load-older__btn" onClick={handleLoadOlder} disabled={loadingMore}>
              {loadingMore ? (
                <>
                  <Loader2 className="ach-load-older__spinner" />
                  Loading…
                </>
              ) : (
                'Load older messages'
              )}
            </button>
          </div>
        )}

        {/* Loading skeleton */}
        {loading && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', padding: '1rem' }}>
            {Array.from({ length: 4 }).map((_, i) => (
              <div
                key={i}
                style={{
                  height: 32,
                  width: i % 2 === 0 ? '60%' : '40%',
                  alignSelf: i % 2 === 0 ? 'flex-start' : 'flex-end',
                  borderRadius: 12,
                  background: 'var(--panel-raised)',
                  opacity: 0.5,
                }}
              />
            ))}
          </div>
        )}

        {/* Empty state */}
        {!loading && messages.length === 0 && (
          <div className="ach-empty">
            <MessageSquare className="ach-empty__icon" />
            <p className="ach-empty__title">Start a conversation with {agentName}</p>
            <p className="ach-empty__model">Model: {model}</p>
          </div>
        )}

        {/* No search results */}
        {!loading && messages.length > 0 && filtered.length === 0 && (
          <div className="ach-no-results">
            <p className="ach-no-results__text">No messages match "{searchQuery}"</p>
          </div>
        )}

        {/* Date groups */}
        {groups.map((group) => (
          <div key={group.label}>
            <div className="ach-date-sep">
              <span className="ach-date-sep__line" />
              <span className="ach-date-sep__label">{group.label}</span>
              <span className="ach-date-sep__line" />
            </div>
            {group.messages.map((msg, i) => (
              <div key={`${group.label}-${i}`} className={`ach-msg ach-msg--${msg.role}`}>
                {msg.role === 'assistant' && msg.thinking && msg.thinking.content && (
                  <ChatThinking content={msg.thinking.content} tokens={msg.thinking.tokens} />
                )}
                <div className="ach-msg__bubble">{msg.content}</div>
                <div className="ach-msg__meta">
                  <span className="ach-msg__time" title={format(new Date(msg.created_at), 'PPpp')}>
                    {formatDistanceToNow(new Date(msg.created_at), { addSuffix: true })}
                  </span>
                  {msg.role === 'assistant' && msg.model && (
                    <span className="ach-msg__model">{msg.model}</span>
                  )}
                </div>
              </div>
            ))}
          </div>
        ))}

        {/* Typing indicator */}
        {sending && (
          <div className="ach-typing">
            <Loader2 className="ach-typing__spinner" />
            <span style={{ fontSize: '0.8rem' }}>Thinking…</span>
          </div>
        )}

        <div ref={bottomRef} style={{ height: 1, flexShrink: 0 }} />
      </div>

      {/* Jump to latest */}
      {showJump && (
        <div className="ach-jump">
          <button
            className="ach-jump__btn"
            onClick={() => {
              scrollToBottom();
              setShowJump(false);
            }}
          >
            <ArrowDown className="ach-jump__icon" />
            Jump to latest
          </button>
        </div>
      )}

      {/* Composer */}
      <div className="ach-composer">
        <input
          className="adp-input ach-composer__input"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && !e.shiftKey && (e.preventDefault(), handleSend())}
          placeholder={t('agentDetail.consolePlaceholder', { name: agentName })}
          disabled={sending}
        />
        <SealedButton
          onClick={handleSend}
          disabled={sending || !input.trim()}
          loading={sending}
          iconLeft={<MessageSquare style={{ width: 14, height: 14 }} />}
        >
          {t('agentDetail.consoleSend')}
        </SealedButton>
      </div>
    </div>
  );
}
