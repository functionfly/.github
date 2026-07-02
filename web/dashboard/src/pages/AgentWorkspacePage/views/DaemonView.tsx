import { agentApi } from '@/api/agent';
import { FrameButton, SealedButton, StatusPill } from '@/components/containment';
import { useAgent, useUpdateAgent } from '@/hooks/useAgent';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { AlertTriangle, Clock, DollarSign, Play, Square, Zap } from 'lucide-react';
import { useCallback, useState } from 'react';
import { toast } from 'sonner';

interface DaemonViewProps {
  agentId: string;
}

export function DaemonView({ agentId }: DaemonViewProps) {
  const { data: agentData, isLoading, error: agentError } = useAgent(agentId);
  const agent = agentData?.agent;
  const updateAgent = useUpdateAgent(agentId);
  const queryClient = useQueryClient();

  const [dailyLimit, setDailyLimit] = useState('10.00');
  const [alertThreshold, setAlertThreshold] = useState('5.00');
  const [autoRestart, setAutoRestart] = useState(true);
  const [autoPause, setAutoPause] = useState(true);

  const isRunning = agent?.is_daemon_running ?? false;
  const startedAt = agent?.daemon_started_at;
  const execCount = agent?.daemon_execution_count ?? 0;

  const uptime = startedAt
    ? Math.floor((Date.now() - new Date(startedAt).getTime()) / 60000)
    : 0;

  const uptimeStr = uptime > 1440
    ? `${Math.floor(uptime / 1440)}d ${Math.floor((uptime % 1440) / 60)}m`
    : uptime > 60
    ? `${Math.floor(uptime / 60)}h ${uptime % 60}m`
    : `${uptime}m`;

  const handleToggle = useCallback(async () => {
    if (!agentId) return;
    try {
      if (isRunning) {
        await agentApi.triggerKillSwitch(agentId, 'Daemon stopped from workspace');
        toast.success('Daemon stopped');
      } else {
        await agentApi.startSession(agentId);
        toast.success('Daemon started');
      }
      queryClient.invalidateQueries({ queryKey: ['agent', agentId] });
    } catch (err: any) {
      toast.error(`Failed to ${isRunning ? 'stop' : 'start'} daemon: ${err?.message ?? 'Unknown error'}`);
    }
  }, [agentId, isRunning, queryClient]);

  if (isLoading) {
    return <div className="aw-loading"><div className="aw-loading__spinner" /></div>;
  }

  if (agentError) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
        <div className="aw-center__header">
          <div>
            <h2 className="aw-center__title">Daemon</h2>
            <p className="aw-center__subtitle">Always-on agent management and controls</p>
          </div>
        </div>
        <div className="aw-empty">
          <AlertTriangle size={40} className="aw-empty__icon" style={{ color: 'var(--status-error)' }} />
          <span className="aw-empty__title">Failed to load daemon status</span>
          <span className="aw-empty__desc">{(agentError as any)?.message ?? 'An error occurred'}</span>
        </div>
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
      <div className="aw-center__header">
        <div>
          <h2 className="aw-center__title">Daemon</h2>
          <p className="aw-center__subtitle">Always-on agent management and controls</p>
        </div>
        {isRunning ? (
          <FrameButton size="sm" onClick={handleToggle} iconLeft={<Square size={12} />}>
            Stop Daemon
          </FrameButton>
        ) : (
          <SealedButton size="sm" onClick={handleToggle} iconLeft={<Play size={12} />}>
            Start Daemon
          </SealedButton>
        )}
      </div>

      {/* Status */}
      <div className="aw-stats">
        <div className="aw-stat">
          <p className="aw-stat__label">Status</p>
          <div style={{ marginTop: 'var(--space-1)' }}>
            <StatusPill status={isRunning ? 'live' : 'revoked'} label={isRunning ? 'Running' : 'Stopped'} />
          </div>
        </div>
        <div className="aw-stat">
          <p className="aw-stat__label">Uptime</p>
          <p className="aw-stat__value">{isRunning ? uptimeStr : '—'}</p>
        </div>
        <div className="aw-stat">
          <p className="aw-stat__label">Executions</p>
          <p className="aw-stat__value">{execCount}</p>
        </div>
        <div className="aw-stat">
          <p className="aw-stat__label">Started At</p>
          <p className="aw-stat__value" style={{ fontSize: '13px' }}>
            {startedAt ? new Date(startedAt).toLocaleString() : '—'}
          </p>
        </div>
      </div>

      {/* Daemon Config */}
      <div className="aw-card">
        <div className="aw-card__header">
          <span className="aw-card__title" style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
            <Zap size={14} />
            Configuration
          </span>
        </div>
        <div className="aw-card__body">
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <div>
                <span style={{ fontFamily: 'var(--font-body)', fontSize: '14px', fontWeight: 500, color: 'var(--text)' }}>Auto-restart</span>
                <p style={{ fontFamily: 'var(--font-body)', fontSize: '12px', color: 'var(--text-faint)', margin: '2px 0 0' }}>
                  Automatically restart on failure with exponential backoff
                </p>
              </div>
              <button className={`aw-switch ${autoRestart ? 'aw-switch--on' : ''}`} onClick={() => setAutoRestart(!autoRestart)}>
                <span className="aw-switch__thumb" />
              </button>
            </div>

            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <div>
                <span style={{ fontFamily: 'var(--font-body)', fontSize: '14px', fontWeight: 500, color: 'var(--text)' }}>Schedule</span>
                <p style={{ fontFamily: 'var(--font-body)', fontSize: '12px', color: 'var(--text-faint)', margin: '2px 0 0' }}>
                  Run on a cron schedule (e.g. */5 * * * *)
                </p>
              </div>
              <input
                style={{ width: '160px', padding: 'var(--space-1) var(--space-2)', fontFamily: 'var(--font-mono)', fontSize: '13px', color: 'var(--text)', background: 'var(--panel-raised)', border: '1px solid var(--steel)', borderRadius: 'var(--radius)' }}
                placeholder="* * * * *"
                defaultValue={String((agent as any)?.daemon_config?.cron ?? '')}
              />
            </div>

            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <div>
                <span style={{ fontFamily: 'var(--font-body)', fontSize: '14px', fontWeight: 500, color: 'var(--text)' }}>Max Concurrent</span>
                <p style={{ fontFamily: 'var(--font-body)', fontSize: '12px', color: 'var(--text-faint)', margin: '2px 0 0' }}>
                  Maximum concurrent daemon executions
                </p>
              </div>
              <input
                style={{ width: '80px', padding: 'var(--space-1) var(--space-2)', fontFamily: 'var(--font-mono)', fontSize: '13px', color: 'var(--text)', background: 'var(--panel-raised)', border: '1px solid var(--steel)', borderRadius: 'var(--radius)', textAlign: 'right' }}
                type="number"
                defaultValue={String((agent as any)?.daemon_config?.max_concurrent ?? 5)}
              />
            </div>
          </div>
        </div>
      </div>

      {/* Cost Guardrails */}
      <div className="aw-card">
        <div className="aw-card__header">
          <span className="aw-card__title" style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
            <DollarSign size={14} />
            Cost Guardrails
          </span>
        </div>
        <div className="aw-card__body">
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <span style={{ fontFamily: 'var(--font-body)', fontSize: '14px', color: 'var(--text)' }}>Daily Spend Limit</span>
              <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-1)' }}>
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: '13px', color: 'var(--text-dim)' }}>$</span>
                <input
                  style={{ width: '100px', padding: 'var(--space-1) var(--space-2)', fontFamily: 'var(--font-mono)', fontSize: '13px', color: 'var(--text)', background: 'var(--panel-raised)', border: '1px solid var(--steel)', borderRadius: 'var(--radius)', textAlign: 'right' }}
                  value={dailyLimit}
                  onChange={e => setDailyLimit(e.target.value)}
                />
              </div>
            </div>

            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <span style={{ fontFamily: 'var(--font-body)', fontSize: '14px', color: 'var(--text)' }}>Alert Threshold</span>
              <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-1)' }}>
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: '13px', color: 'var(--text-dim)' }}>$</span>
                <input
                  style={{ width: '100px', padding: 'var(--space-1) var(--space-2)', fontFamily: 'var(--font-mono)', fontSize: '13px', color: 'var(--text)', background: 'var(--panel-raised)', border: '1px solid var(--steel)', borderRadius: 'var(--radius)', textAlign: 'right' }}
                  value={alertThreshold}
                  onChange={e => setAlertThreshold(e.target.value)}
                />
              </div>
            </div>

            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <div>
                <span style={{ fontFamily: 'var(--font-body)', fontSize: '14px', color: 'var(--text)' }}>Auto-pause on limit</span>
                <p style={{ fontFamily: 'var(--font-body)', fontSize: '12px', color: 'var(--text-faint)', margin: '2px 0 0' }}>
                  Automatically pause daemon when daily limit is reached
                </p>
              </div>
              <button className={`aw-switch ${autoPause ? 'aw-switch--on' : ''}`} onClick={() => setAutoPause(!autoPause)}>
                <span className="aw-switch__thumb" />
              </button>
            </div>

            <SealedButton size="sm" iconLeft={<DollarSign size={12} />} onClick={async () => {
              try {
                await updateAgent.mutateAsync({
                  daemon_config: {
                    daily_limit_usd: parseFloat(dailyLimit) || 0,
                    alert_threshold_usd: parseFloat(alertThreshold) || 0,
                    auto_pause_on_limit: autoPause,
                  },
                } as any);
                toast.success('Guardrails saved');
              } catch (err: any) {
                toast.error(`Failed to save guardrails: ${err?.message ?? 'Unknown error'}`);
              }
            }}>
              Save Guardrails
            </SealedButton>
          </div>
        </div>
      </div>
    </div>
  );
}
