import { agentApi } from '@/api/agent';
import { formatDistanceToNow } from 'date-fns';
import { Loader2, MessageSquare, Plus, Trash2 } from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';
import './agent-chat.css';

interface ChatSession {
  id: string;
  title: string;
  agent_id: string;
  message_count: number;
  last_message_at?: string;
  created_at: string;
}

interface ConversationSidebarProps {
  agentId: string;
  activeSessionId: string | null;
  onSelectSession: (sessionId: string) => void;
  onNewSession: () => void;
}

export function ConversationSidebar({ agentId, activeSessionId, onSelectSession, onNewSession }: ConversationSidebarProps) {
  const [sessions, setSessions] = useState<ChatSession[]>([]);
  const [loading, setLoading] = useState(true);
  const [deleting, setDeleting] = useState<string | null>(null);

  const fetchSessions = useCallback(async () => {
    try {
      const res = await agentApi.listChatSessions(agentId, 100, 0);
      setSessions(res.sessions ?? []);
    } catch {
      setSessions([]);
    } finally {
      setLoading(false);
    }
  }, [agentId]);

  useEffect(() => {
    fetchSessions();
  }, [fetchSessions]);

  const handleDelete = async (e: React.MouseEvent, sessionId: string) => {
    e.stopPropagation();
    setDeleting(sessionId);
    try {
      await agentApi.deleteChatSession(agentId, sessionId);
      setSessions((prev) => prev.filter((s) => s.id !== sessionId));
    } catch {
      // silently ignore
    } finally {
      setDeleting(null);
    }
  };

  return (
    <div className="acs-wrap">
      <div className="acs-header">
        <span className="acs-header__title">History</span>
        <button className="acs-new-btn" onClick={onNewSession} title="New conversation">
          <Plus style={{ width: 14, height: 14 }} />
        </button>
      </div>

      <div className="acs-list">
        {loading && (
          <div className="acs-loading">
            <Loader2 className="acs-loading__spinner" />
          </div>
        )}

        {!loading && sessions.length === 0 && (
          <div className="acs-empty">
            <MessageSquare style={{ width: 20, height: 20, opacity: 0.3 }} />
            <span className="acs-empty__text">No conversations yet</span>
          </div>
        )}

        {sessions.map((session) => (
          <div
            key={session.id}
            className={`acs-item ${session.id === activeSessionId ? 'acs-item--active' : ''}`}
            onClick={() => onSelectSession(session.id)}
          >
            <div className="acs-item__content">
              <span className="acs-item__title">{session.title || 'New Chat'}</span>
              <span className="acs-item__meta">
                {session.message_count} message{session.message_count !== 1 ? 's' : ''}
                {session.last_message_at && (
                  <> &middot; {formatDistanceToNow(new Date(session.last_message_at), { addSuffix: true })}</>
                )}
              </span>
            </div>
            <button
              className="acs-item__delete"
              onClick={(e) => handleDelete(e, session.id)}
              disabled={deleting === session.id}
              title="Delete conversation"
            >
              {deleting === session.id ? (
                <Loader2 className="acs-loading__spinner" />
              ) : (
                <Trash2 style={{ width: 12, height: 12 }} />
              )}
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}
