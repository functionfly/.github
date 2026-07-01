import { StatusPill } from '@/components/containment';
import { useAgentUsage } from '@/hooks/useAgent';
import { useAgentHealth } from '../hooks/useAgentHealth';
import { AlertTriangle, Heart, Shield } from 'lucide-react';

interface HealthViewProps {
  agentId: string;
}

export function HealthView({ agentId }: HealthViewProps) {
  const { health, healthLoading } = useAgentHealth(agentId);
  const { data: usageData } = useAgentUsage(agentId);
  const usage = usageData?.usage as any;

  if (healthLoading) {
    return <div className="aw-loading"><div className="aw-loading__spinner" /></div>;
  }

  const statusMap: Record<string, 'live' | 'pending' | 'revoked'> = {
    healthy: 'live',
    degraded: 'pending',
    critical: 'revoked',
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
      <div className="aw-center__header">
        <div>
          <h2 className="aw-center__title">Health</h2>
          <p className="aw-center__subtitle">Agent health monitoring and SLA tracking</p>
        </div>
      </div>

      {/* Health Score */}
      <div className="aw-stats">
        <div className="aw-stat" style={{ textAlign: 'center' }}>
          <p className="aw-stat__label">Health Score</p>
          <p className="aw-stat__value" style={{
            fontSize: '40px',
            color: (health?.health_score ?? 0) >= 80 ? 'var(--status-ok)' :
                   (health?.health_score ?? 0) >= 50 ? 'var(--status-pending)' :
                   'var(--status-revoked)',
          }}>
            {health?.health_score ?? '—'}
          </p>
          {health && <StatusPill status={statusMap[health.status] ?? 'pending'} label={health.status} />}
        </div>
        <div className="aw-stat">
          <p className="aw-stat__label">Calls / Min</p>
          <p className="aw-stat__value">{usage?.calls_per_minute?.toFixed(1) ?? '0'}</p>
        </div>
        <div className="aw-stat">
          <p className="aw-stat__label">Concurrent</p>
          <p className="aw-stat__value">{usage?.concurrent_executions ?? 0}</p>
        </div>
        <div className="aw-stat">
          <p className="aw-stat__label">Avg Exec Time</p>
          <p className="aw-stat__value">{usage?.avg_execution_time_ms ? `${(usage.avg_execution_time_ms / 1000).toFixed(1)}s` : '—'}</p>
        </div>
      </div>

      {/* Anomalies */}
      <div className="aw-card">
        <div className="aw-card__header">
          <span className="aw-card__title" style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
            <AlertTriangle size={14} />
            Anomalies
            {health?.anomalies && health.anomalies.length > 0 && (
              <span className="aw-nav-item__badge">{health.anomalies.length}</span>
            )}
          </span>
        </div>
        <div className="aw-card__body">
          {(!health?.anomalies || health.anomalies.length === 0) ? (
            <div className="aw-empty" style={{ padding: 'var(--space-5)' }}>
              <Shield size={32} className="aw-empty__icon" />
              <span className="aw-empty__title">No anomalies detected</span>
              <span className="aw-empty__desc">All systems operating normally</span>
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
              {health.anomalies.map((anomaly: any, i: number) => (
                <div key={i} className="aw-feed-item">
                  <span className={`aw-feed-item__dot ${anomaly.severity === 'critical' ? 'aw-feed-item__dot--error' : 'aw-feed-item__dot--action'}`} />
                  <div className="aw-feed-item__content">
                    <div className="aw-feed-item__kind">{anomaly.type ?? anomaly.severity ?? 'anomaly'}</div>
                    <div className="aw-feed-item__body">{anomaly.description ?? anomaly.message ?? JSON.stringify(anomaly)}</div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Children Health */}
      {health?.children && Array.isArray(health.children) && health.children.length > 0 && (
        <div className="aw-card">
          <div className="aw-card__header">
            <span className="aw-card__title" style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
              <Heart size={14} />
              Children Health
            </span>
          </div>
          <div className="aw-card__body">
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
              {health.children.map((child: any) => (
                <div key={child.agent_id} className="aw-swarm-node">
                  <span className="aw-swarm-node__avatar">{(child.agent_id ?? '?')[0].toUpperCase()}</span>
                  <div className="aw-swarm-node__info">
                    <span className="aw-swarm-node__name">{child.agent_id}</span>
                    <span className="aw-swarm-node__role">Score: {child.health_score ?? '—'}</span>
                  </div>
                  <StatusPill
                    status={child.status === 'healthy' ? 'live' : child.status === 'critical' ? 'revoked' : 'pending'}
                    label={child.status}
                  />
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
