import { AgentChatHistory } from '@/components/agent-console/AgentChatHistory';
import { ConversationSidebar } from '@/components/agent-console/ConversationSidebar';
import { FrameButton } from '@/components/containment';
import { Terminal, Wifi, WifiOff, Trash2 } from 'lucide-react';
import { useCallback, useState } from 'react';
import { useAgentRealtime, type RealtimeEvent } from '../hooks/useAgentRealtime';
import { formatDistanceToNow } from 'date-fns';

interface ConsoleViewProps {
  agentId: string;
  agentName: string;
  model: string;
}

const kindColor: Record<string, string> = {
  INPUT: 'aw-feed-item__dot--input',
  DECISION: 'aw-feed-item__dot--decision',
  ACTION: 'aw-feed-item__dot--action',
  RESULT: 'aw-feed-item__dot--result',
  ERROR: 'aw-feed-item__dot--error',
  TOOL_CALL: 'aw-feed-item__dot--tool',
};

export function ConsoleView({ agentId, agentName, model }: ConsoleViewProps) {
  const { events, connected, clearEvents } = useAgentRealtime({
    agentId,
    enabled: true,
  });

  const [activeSessionId, setActiveSessionId] = useState<string | null>(null);
  const [sidebarKey, setSidebarKey] = useState(0);

  const handleSessionCreated = useCallback((sessionId: string) => {
    setActiveSessionId(sessionId);
    setSidebarKey((k) => k + 1);
  }, []);

  const handleSelectSession = useCallback((sessionId: string) => {
    setActiveSessionId(sessionId);
  }, []);

  const handleNewSession = useCallback(() => {
    setActiveSessionId(null);
  }, []);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)', flex: 1, minHeight: 0 }}>
      <div className="aw-center__header">
        <div>
          <h2 className="aw-center__title">Console</h2>
          <p className="aw-center__subtitle">Chat with your agent and monitor live executions</p>
        </div>
      </div>

      <div style={{ display: 'flex', gap: 'var(--space-4)', minHeight: 0, flex: 1 }}>
        {/* Conversation History Sidebar */}
        <div className="aw-card" style={{ width: 240, flexShrink: 0, display: 'flex', flexDirection: 'column', minHeight: 0 }}>
          <div className="aw-card__body aw-card__body--flush" style={{ flex: 1, display: 'flex', flexDirection: 'column', minHeight: 0 }}>
            <ConversationSidebar
              key={sidebarKey}
              agentId={agentId}
              activeSessionId={activeSessionId}
              onSelectSession={handleSelectSession}
              onNewSession={handleNewSession}
            />
          </div>
        </div>

        {/* Chat Panel */}
        <div style={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column' }}>
          <div className="aw-card" style={{ flex: 1, display: 'flex', flexDirection: 'column', minHeight: 0 }}>
            <div className="aw-card__body aw-card__body--flush" style={{ flex: 1, display: 'flex', flexDirection: 'column', minHeight: 0 }}>
              <AgentChatHistory
                agentId={agentId}
                agentName={agentName}
                model={model}
                sessionId={activeSessionId}
                onSessionCreated={handleSessionCreated}
              />
            </div>
          </div>
        </div>

        {/* Live Execution Feed */}
        <div style={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column' }}>
          <div className="aw-card" style={{ flex: 1, display: 'flex', flexDirection: 'column', minHeight: 0 }}>
            <div className="aw-card__header">
              <span className="aw-card__title" style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
                <Terminal size={14} />
                Live Feed
                <span style={{ display: 'flex', alignItems: 'center', gap: '4px', fontSize: '11px', color: connected ? 'var(--status-ok)' : 'var(--status-revoked)' }}>
                  {connected ? <Wifi size={12} /> : <WifiOff size={12} />}
                  {connected ? 'Connected' : 'Disconnected'}
                </span>
              </span>
              <div style={{ display: 'flex', gap: 'var(--space-2)', alignItems: 'center' }}>
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: '11px', color: 'var(--text-faint)' }}>
                  {events.length} events
                </span>
                <FrameButton size="sm" onClick={clearEvents} iconLeft={<Trash2 size={12} />}>
                  Clear
                </FrameButton>
              </div>
            </div>
            <div className="aw-card__body aw-card__body--flush" style={{ flex: 1, overflowY: 'auto', padding: 'var(--space-3)' }}>
              {events.length === 0 ? (
                <div className="aw-empty">
                  <Terminal size={32} className="aw-empty__icon" />
                  <span className="aw-empty__title">No events yet</span>
                  <span className="aw-empty__desc">Events will appear here when the agent executes</span>
                </div>
              ) : (
                <div className="aw-feed">
                  {events.map((event, i) => (
                    <FeedItem key={event.id || i} event={event} />
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function FeedItem({ event }: { event: RealtimeEvent }) {
  const kind = event.kind?.toUpperCase() ?? 'UNKNOWN';
  const dotClass = kindColor[kind] ?? 'aw-feed-item__dot--input';

  return (
    <div className="aw-feed-item">
      <span className={`aw-feed-item__dot ${dotClass}`} />
      <div className="aw-feed-item__content">
        <div className="aw-feed-item__kind">{kind}</div>
        <div className="aw-feed-item__body">
          {event.tool_name && (
            <span style={{ color: 'var(--accent)', fontWeight: 600, marginRight: 'var(--space-2)' }}>
              [{event.tool_name}]
            </span>
          )}
          {typeof event.payload === 'string'
            ? event.payload
            : JSON.stringify(event.payload).slice(0, 200)}
        </div>
        <div className="aw-feed-item__meta">
          {event.timestamp && (
            <span>{formatDistanceToNow(new Date(event.timestamp), { addSuffix: true })}</span>
          )}
          {event.cost_usd !== undefined && event.cost_usd > 0 && (
            <span>${event.cost_usd.toFixed(4)}</span>
          )}
          {event.tokens_in !== undefined && (
            <span>{event.tokens_in} in / {event.tokens_out ?? 0} out</span>
          )}
        </div>
      </div>
    </div>
  );
}
