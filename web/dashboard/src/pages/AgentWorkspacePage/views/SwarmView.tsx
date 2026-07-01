import { agentApi } from '@/api/agent';
import { FrameButton, SealedButton, StatusPill } from '@/components/containment';
import { useAgentChildren, useSwarmHealth, useSwarmStats, useAgentInbox, useSpawnChildAgent } from '@/hooks/useAgentSwarm';
import { useQuery } from '@tanstack/react-query';
import { GitBranch, MessageSquare, Plus, Users, Wallet } from 'lucide-react';
import { useState } from 'react';

interface SwarmViewProps {
  agentId: string;
  setRightContext: (ctx: { type: string; id: string } | null) => void;
}

export function SwarmView({ agentId, setRightContext }: SwarmViewProps) {
  const { data: childrenData } = useAgentChildren(agentId);
  const { data: healthData } = useSwarmHealth(agentId);
  const { data: statsData } = useSwarmStats(agentId);
  const { data: inboxData } = useAgentInbox(agentId);
  const spawnChild = useSpawnChildAgent();

  const { data: walletData } = useQuery({
    queryKey: ['agent-wallet', agentId],
    queryFn: () => agentApi.getWallet(agentId),
    enabled: !!agentId,
  });

  const children = childrenData?.children ?? [];
  const messages = inboxData?.messages ?? [];
  const stats = statsData?.stats;
  const health = healthData;
  const wallet = walletData?.wallet;

  const [showSpawn, setShowSpawn] = useState(false);
  const [spawnName, setSpawnName] = useState('');
  const [spawnRole, setSpawnRole] = useState('worker');

  const handleSpawn = async () => {
    if (!spawnName.trim()) return;
    await spawnChild.mutateAsync({ agentId, name: spawnName, swarmRole: spawnRole });
    setShowSpawn(false);
    setSpawnName('');
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
      <div className="aw-center__header">
        <div>
          <h2 className="aw-center__title">Swarm</h2>
          <p className="aw-center__subtitle">Multi-agent coordination and management</p>
        </div>
        <SealedButton size="sm" onClick={() => setShowSpawn(!showSpawn)} iconLeft={<Plus size={12} />}>
          Spawn Child
        </SealedButton>
      </div>

      {/* Spawn Form */}
      {showSpawn && (
        <div className="aw-card">
          <div className="aw-card__header">
            <span className="aw-card__title">Spawn Child Agent</span>
          </div>
          <div className="aw-card__body">
            <div style={{ display: 'flex', gap: 'var(--space-3)', alignItems: 'flex-end' }}>
              <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 'var(--space-1)' }}>
                <label style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', fontWeight: 500, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--text-faint)' }}>Name</label>
                <input
                  style={{ padding: 'var(--space-2) var(--space-3)', fontFamily: 'var(--font-body)', fontSize: '13px', color: 'var(--text)', background: 'var(--panel-raised)', border: '1px solid var(--steel)', borderRadius: 'var(--radius)' }}
                  placeholder="Child agent name"
                  value={spawnName}
                  onChange={e => setSpawnName(e.target.value)}
                />
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-1)' }}>
                <label style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', fontWeight: 500, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--text-faint)' }}>Role</label>
                <select
                  style={{ padding: 'var(--space-2) var(--space-3)', fontFamily: 'var(--font-body)', fontSize: '13px', color: 'var(--text)', background: 'var(--panel-raised)', border: '1px solid var(--steel)', borderRadius: 'var(--radius)' }}
                  value={spawnRole}
                  onChange={e => setSpawnRole(e.target.value)}
                >
                  <option value="worker">Worker</option>
                  <option value="manager">Manager</option>
                  <option value="infrastructure">Infrastructure</option>
                </select>
              </div>
              <SealedButton size="sm" onClick={handleSpawn} loading={spawnChild.isPending}>
                Spawn
              </SealedButton>
            </div>
          </div>
        </div>
      )}

      {/* Stats */}
      <div className="aw-stats">
        <div className="aw-stat">
          <p className="aw-stat__label">Total Agents</p>
          <p className="aw-stat__value">{children.length + 1}</p>
        </div>
        <div className="aw-stat">
          <p className="aw-stat__label">Messages</p>
          <p className="aw-stat__value">{messages.length}</p>
        </div>
        <div className="aw-stat">
          <p className="aw-stat__label">Health</p>
          <p className="aw-stat__value">{health?.health_score ?? '—'}</p>
        </div>
        <div className="aw-stat">
          <p className="aw-stat__label">Balance</p>
          <p className="aw-stat__value">${(wallet?.balance_usd ?? 0).toFixed(2)}</p>
        </div>
      </div>

      {/* Children List */}
      <div className="aw-card">
        <div className="aw-card__header">
          <span className="aw-card__title" style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
            <Users size={14} />
            Child Agents ({children.length})
          </span>
        </div>
        <div className="aw-card__body">
          {children.length === 0 ? (
            <div className="aw-empty" style={{ padding: 'var(--space-4)' }}>
              <Users size={32} className="aw-empty__icon" />
              <span className="aw-empty__title">No child agents</span>
              <span className="aw-empty__desc">Spawn a child agent to build a swarm</span>
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
              {children.map((child: any) => (
                <div
                  key={child.agent_id}
                  className="aw-swarm-node"
                  onClick={() => setRightContext({ type: 'swarm-node', id: child.agent_id })}
                >
                  <span className="aw-swarm-node__avatar">
                    {(child.name ?? child.agent_id ?? '?')[0].toUpperCase()}
                  </span>
                  <div className="aw-swarm-node__info">
                    <span className="aw-swarm-node__name">{child.name ?? child.agent_id}</span>
                    <span className="aw-swarm-node__role">{child.swarm_role ?? 'worker'}</span>
                  </div>
                  <StatusPill
                    status={child.status === 'active' ? 'live' : child.status === 'suspended' ? 'revoked' : 'pending'}
                    label={child.status}
                  />
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Topology */}
      <div className="aw-card">
        <div className="aw-card__header">
          <span className="aw-card__title" style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
            <GitBranch size={14} />
            Topology
          </span>
        </div>
        <div className="aw-card__body">
          <div style={{ fontFamily: 'var(--font-mono)', fontSize: '13px', color: 'var(--text)', lineHeight: 1.8 }}>
            <div style={{ color: 'var(--accent)' }}>You (parent)</div>
            {children.map((child: any, i: number) => (
              <div key={child.agent_id} style={{ paddingLeft: 'var(--space-5)' }}>
                {i === children.length - 1 ? '└─' : '├─'} <span style={{ color: 'var(--text-dim)' }}>{child.name ?? child.agent_id}</span>
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', color: 'var(--text-faint)', marginLeft: 'var(--space-2)' }}>
                  [{child.swarm_role ?? 'worker'}]
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Recent Messages */}
      <div className="aw-card">
        <div className="aw-card__header">
          <span className="aw-card__title" style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
            <MessageSquare size={14} />
            Recent Messages ({messages.length})
          </span>
        </div>
        <div className="aw-card__body">
          {messages.length === 0 ? (
            <div className="aw-empty" style={{ padding: 'var(--space-4)' }}>
              <MessageSquare size={32} className="aw-empty__icon" />
              <span className="aw-empty__title">No messages</span>
              <span className="aw-empty__desc">Inter-agent messages will appear here</span>
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
              {messages.slice(0, 10).map((msg: any) => (
                <div key={msg.id} className="aw-feed-item">
                  <span className="aw-feed-item__dot aw-feed-item__dot--decision" />
                  <div className="aw-feed-item__content">
                    <div className="aw-feed-item__kind">{msg.message_type ?? 'message'}</div>
                    <div className="aw-feed-item__body">
                      {typeof msg.payload === 'string' ? msg.payload : JSON.stringify(msg.payload).slice(0, 150)}
                    </div>
                    <div className="aw-feed-item__meta">
                      <span>From: {msg.from_agent_id?.slice(0, 8)}</span>
                      {msg.created_at && <span>{new Date(msg.created_at).toLocaleString()}</span>}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
