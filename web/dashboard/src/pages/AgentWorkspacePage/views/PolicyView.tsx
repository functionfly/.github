import { agentApi } from '@/api/agent';
import { FrameButton, SealedButton } from '@/components/containment';
import { useAgentPolicy, useUpdateAgentPolicy } from '@/hooks/useAgent';
import { useQuery } from '@tanstack/react-query';
import { AlertTriangle, CheckCircle, RefreshCw, Shield, ShieldAlert, ShieldCheck, XCircle } from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';
import { toast } from 'sonner';

interface PolicyViewProps {
  agentId: string;
}

export function PolicyView({ agentId }: PolicyViewProps) {
  const { data: policyData, isLoading, error: policyError } = useAgentPolicy(agentId);
  const updatePolicy = useUpdateAgentPolicy(agentId);
  const policy = policyData?.policy;

  const { data: analyticsData } = useQuery({
    queryKey: ['agent-analytics', agentId],
    queryFn: () => agentApi.getAnalytics(agentId),
    enabled: !!agentId,
  });

  const violationCount = analyticsData?.analytics?.policy_violation_count ?? 0;

  const [editDepth, setEditDepth] = useState('10');
  const [editRecursion, setEditRecursion] = useState('5');
  const [editWallTime, setEditWallTime] = useState('30000');
  const [editMemory, setEditMemory] = useState('256');

  useEffect(() => {
    if (policy) {
      setEditDepth(String(policy.maxExecutionDepth ?? 10));
      setEditRecursion(String(policy.maxRecursionDepth ?? 5));
      setEditWallTime(String(policy.maxWallTimeMs ?? 30000));
      setEditMemory(String(policy.maxMemoryGrowthMB ?? 256));
    }
  }, [policy]);

  const handleSave = useCallback(async () => {
    try {
      await updatePolicy.mutateAsync({
        maxExecutionDepth: parseInt(editDepth) || 10,
        maxRecursionDepth: parseInt(editRecursion) || 5,
        maxWallTimeMs: parseInt(editWallTime) || 30000,
        maxMemoryGrowthMB: parseInt(editMemory) || 256,
      });
      toast.success('Policy saved');
    } catch (err: any) {
      toast.error(`Failed to save policy: ${err?.message ?? 'Unknown error'}`);
    }
  }, [editDepth, editRecursion, editWallTime, editMemory, updatePolicy]);

  if (isLoading) {
    return <div className="aw-loading"><div className="aw-loading__spinner" /></div>;
  }

  if (policyError) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
        <div className="aw-center__header">
          <div>
            <h2 className="aw-center__title">Policy</h2>
            <p className="aw-center__subtitle">Behavioral guardrails and violation tracking</p>
          </div>
        </div>
        <div className="aw-empty">
          <AlertTriangle size={40} className="aw-empty__icon" style={{ color: 'var(--status-error)' }} />
          <span className="aw-empty__title">Failed to load policy</span>
          <span className="aw-empty__desc">{(policyError as any)?.message ?? 'An error occurred'}</span>
        </div>
      </div>
    );
  }

  const forbiddenFns = policy?.forbiddenFunctions ?? [];
  const allowedCaps = policy?.allowedCapabilities ?? [];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
      <div className="aw-center__header">
        <div>
          <h2 className="aw-center__title">Policy</h2>
          <p className="aw-center__subtitle">Behavioral guardrails and violation tracking</p>
        </div>
      </div>

      {/* Active Rules */}
      <div className="aw-card">
        <div className="aw-card__header">
          <span className="aw-card__title" style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
            <Shield size={14} />
            Active Rules
          </span>
        </div>
        <div className="aw-card__body">
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
            {/* Deny rules */}
            {forbiddenFns.map((fn: string, i: number) => (
              <div key={`deny-${i}`} className="aw-policy-row">
                <span className="aw-policy-row__type aw-policy-row__type--deny">DENY</span>
                <span className="aw-policy-row__text">Function: {fn}</span>
              </div>
            ))}

            {/* Allow rules */}
            {allowedCaps.map((cap: string, i: number) => (
              <div key={`allow-${i}`} className="aw-policy-row">
                <span className="aw-policy-row__type aw-policy-row__type--allow">ALLOW</span>
                <span className="aw-policy-row__text">Capability: {cap}</span>
              </div>
            ))}

            {/* Limit rules */}
            <div className="aw-policy-row">
              <span className="aw-policy-row__type aw-policy-row__type--limit">LIMIT</span>
              <span className="aw-policy-row__text">Max execution depth: {policy?.maxExecutionDepth ?? 10}</span>
            </div>
            <div className="aw-policy-row">
              <span className="aw-policy-row__type aw-policy-row__type--limit">LIMIT</span>
              <span className="aw-policy-row__text">Max recursion depth: {policy?.maxRecursionDepth ?? 5}</span>
            </div>
            <div className="aw-policy-row">
              <span className="aw-policy-row__type aw-policy-row__type--limit">LIMIT</span>
              <span className="aw-policy-row__text">Max wall time: {((policy?.maxWallTimeMs ?? 30000) / 1000).toFixed(0)}s</span>
            </div>
            <div className="aw-policy-row">
              <span className="aw-policy-row__type aw-policy-row__type--limit">LIMIT</span>
              <span className="aw-policy-row__text">Max memory growth: {policy?.maxMemoryGrowthMB ?? 256} MB</span>
            </div>

            {policy?.deterministicOnly && (
              <div className="aw-policy-row">
                <span className="aw-policy-row__type aw-policy-row__type--deny">DENY</span>
                <span className="aw-policy-row__text">Non-deterministic operations</span>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Edit Limits */}
      <div className="aw-card">
        <div className="aw-card__header">
          <span className="aw-card__title">Edit Limits</span>
        </div>
        <div className="aw-card__body">
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 'var(--space-4)' }}>
            {[
              { label: 'Max Execution Depth', value: editDepth, setter: setEditDepth },
              { label: 'Max Recursion Depth', value: editRecursion, setter: setEditRecursion },
              { label: 'Max Wall Time (ms)', value: editWallTime, setter: setEditWallTime },
              { label: 'Max Memory Growth (MB)', value: editMemory, setter: setEditMemory },
            ].map(field => (
              <div key={field.label} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-1)' }}>
                <label style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', fontWeight: 500, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--text-faint)' }}>
                  {field.label}
                </label>
                <input
                  style={{ padding: 'var(--space-2) var(--space-3)', fontFamily: 'var(--font-mono)', fontSize: '13px', color: 'var(--text)', background: 'var(--panel-raised)', border: '1px solid var(--steel)', borderRadius: 'var(--radius)' }}
                  type="number"
                  value={field.value}
                  onChange={e => field.setter(e.target.value)}
                />
              </div>
            ))}
          </div>
          <div style={{ marginTop: 'var(--space-4)' }}>
            <SealedButton size="sm" onClick={handleSave} loading={updatePolicy.isPending} iconLeft={<ShieldCheck size={12} />}>
              Save Policy
            </SealedButton>
          </div>
        </div>
      </div>

      {/* Violations */}
      <div className="aw-card">
        <div className="aw-card__header">
          <span className="aw-card__title" style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
            <ShieldAlert size={14} />
            Violations
            {violationCount > 0 && (
              <span className="aw-nav-item__badge">{violationCount}</span>
            )}
          </span>
        </div>
        <div className="aw-card__body">
          {violationCount === 0 ? (
            <div className="aw-empty" style={{ padding: 'var(--space-4)' }}>
              <CheckCircle size={32} className="aw-empty__icon" style={{ color: 'var(--status-ok)', opacity: 1 }} />
              <span className="aw-empty__title">No violations recorded</span>
              <span className="aw-empty__desc">All executions have complied with the active policy</span>
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', padding: 'var(--space-3)', background: 'rgba(255, 107, 107, 0.06)', border: '1px solid rgba(255, 107, 107, 0.2)', borderRadius: 'var(--radius)' }}>
                <AlertTriangle size={16} style={{ color: 'var(--status-revoked)', flexShrink: 0 }} />
                <span style={{ fontFamily: 'var(--font-body)', fontSize: '13px', color: 'var(--text)' }}>
                  {violationCount} policy violation{violationCount !== 1 ? 's' : ''} recorded. Review execution logs for details.
                </span>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Dry Run */}
      <div className="aw-card">
        <div className="aw-card__header">
          <span className="aw-card__title">Dry Run Mode</span>
        </div>
        <div className="aw-card__body">
          <p style={{ fontFamily: 'var(--font-body)', fontSize: '13px', color: 'var(--text-dim)', margin: '0 0 var(--space-3)' }}>
            Simulate policy changes against the last 24h of executions to see what would have been blocked.
          </p>
          <FrameButton size="sm" disabled iconLeft={<Shield size={12} />}>
            Simulate Policy Changes (Coming Soon)
          </FrameButton>
        </div>
      </div>
    </div>
  );
}
