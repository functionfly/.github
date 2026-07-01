import { agentApi, normalizeAgentIdentity } from '@/api/agent';
import { FrameButton, SealedButton } from '@/components/containment';
import { useAgent, useUpdateAgent, useAgentPolicy, useUpdateAgentPolicy, useDeleteAgent } from '@/hooks/useAgent';
import { useQuery } from '@tanstack/react-query';
import { Save, Settings, Trash2 } from 'lucide-react';
import { useCallback, useEffect, useMemo, useState } from 'react';

interface ConfigViewProps {
  agentId: string;
}

const PRESET_CAPABILITIES = {
  'Data & Search': ['web_search', 'database_query', 'file_read', 'file_write'],
  'Communication': ['http_request', 'email', 'notification'],
  'AI & Processing': ['code_execution', 'image_generation', 'text_to_speech', 'speech_to_text'],
};

export function ConfigView({ agentId }: ConfigViewProps) {
  const { data: agentData, isLoading } = useAgent(agentId);
  const { data: policyData } = useAgentPolicy(agentId);
  const updateAgent = useUpdateAgent(agentId);
  const updatePolicy = useUpdateAgentPolicy(agentId);
  const deleteAgent = useDeleteAgent();

  const agent = agentData?.agent ? normalizeAgentIdentity(agentData.agent) : undefined;
  const policy = policyData?.policy;

  const { data: modelsData } = useQuery({
    queryKey: ['ai-models'],
    queryFn: () => agentApi.getModels(),
    staleTime: 300000,
  });

  const models = useMemo(() => {
    const raw = modelsData?.models;
    if (!raw) return [];
    if (Array.isArray(raw)) return raw;
    if (typeof raw === 'object') return Object.entries(raw).flatMap(([group, list]: [string, any]) =>
      Array.isArray(list) ? list.map((m: any) => ({ ...m, group })) : []
    );
    return [];
  }, [modelsData]);

  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [model, setModel] = useState('');
  const [thinkingMode, setThinkingMode] = useState('off');
  const [thinkingBudget, setThinkingBudget] = useState(1000);
  const [capabilities, setCapabilities] = useState<Record<string, string>>({});
  const [autonomousEnabled, setAutonomousEnabled] = useState(false);
  const [evolutionEnabled, setEvolutionEnabled] = useState(false);
  const [maxDepth, setMaxDepth] = useState('10');
  const [maxRecursion, setMaxRecursion] = useState('5');
  const [maxWallTime, setMaxWallTime] = useState('30000');
  const [maxMemory, setMaxMemory] = useState('256');
  const [initialized, setInitialized] = useState(false);
  const [deleteConfirm, setDeleteConfirm] = useState(false);

  useEffect(() => {
    if (agent && !initialized) {
      setName(agent.name ?? '');
      setDescription(agent.description ?? '');
      setModel(agent.model ?? 'gpt-4o-mini');
      setThinkingMode(agent.thinking_mode ?? 'off');
      setThinkingBudget(agent.thinking_budget ?? 1000);
      setCapabilities((agent.capabilities ?? {}) as Record<string, string>);
      setAutonomousEnabled(agent.autonomousEnabled ?? false);
      setEvolutionEnabled(agent.evolutionEnabled ?? false);
      setInitialized(true);
    }
  }, [agent, initialized]);

  useEffect(() => {
    if (policy) {
      setMaxDepth(String(policy.maxExecutionDepth ?? 10));
      setMaxRecursion(String(policy.maxRecursionDepth ?? 5));
      setMaxWallTime(String(policy.maxWallTimeMs ?? 30000));
      setMaxMemory(String(policy.maxMemoryGrowthMB ?? 256));
    }
  }, [policy]);

  const isCapEnabled = useCallback((key: string) => capabilities[key] === 'true' || capabilities[key] === 'enabled', [capabilities]);

  const toggleCap = useCallback((key: string) => {
    setCapabilities(prev => ({ ...prev, [key]: isCapEnabled(key) ? 'false' : 'true' }));
  }, [isCapEnabled]);

  const handleSave = useCallback(async () => {
    await Promise.all([
      updateAgent.mutateAsync({
        name,
        description,
        model,
        capabilities,
        autonomous_enabled: autonomousEnabled,
        evolution_enabled: evolutionEnabled,
      } as any),
      updatePolicy.mutateAsync({
        maxExecutionDepth: parseInt(maxDepth) || 10,
        maxRecursionDepth: parseInt(maxRecursion) || 5,
        maxWallTimeMs: parseInt(maxWallTime) || 30000,
        maxMemoryGrowthMB: parseInt(maxMemory) || 256,
      }),
    ]);
  }, [name, description, model, capabilities, autonomousEnabled, evolutionEnabled, maxDepth, maxRecursion, maxWallTime, maxMemory, updateAgent, updatePolicy]);

  const handleDelete = useCallback(async () => {
    if (!deleteConfirm) {
      setDeleteConfirm(true);
      return;
    }
    await deleteAgent.mutateAsync(agentId);
  }, [deleteConfirm, deleteAgent, agentId]);

  if (isLoading) {
    return <div className="aw-loading"><div className="aw-loading__spinner" /></div>;
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
      <div className="aw-center__header">
        <div>
          <h2 className="aw-center__title">Configuration</h2>
          <p className="aw-center__subtitle">Agent identity, model, and behavioral settings</p>
        </div>
        <SealedButton size="sm" onClick={handleSave} loading={updateAgent.isPending || updatePolicy.isPending} iconLeft={<Save size={12} />}>
          Save All
        </SealedButton>
      </div>

      {/* Identity */}
      <div className="aw-card">
        <div className="aw-card__header"><span className="aw-card__title">Identity</span></div>
        <div className="aw-card__body">
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)' }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-1)' }}>
              <label style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', fontWeight: 500, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--text-faint)' }}>Name</label>
              <input style={{ padding: 'var(--space-2) var(--space-3)', fontFamily: 'var(--font-body)', fontSize: '13px', color: 'var(--text)', background: 'var(--panel-raised)', border: '1px solid var(--steel)', borderRadius: 'var(--radius)' }} value={name} onChange={e => setName(e.target.value)} />
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-1)' }}>
              <label style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', fontWeight: 500, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--text-faint)' }}>Description</label>
              <textarea style={{ padding: 'var(--space-2) var(--space-3)', fontFamily: 'var(--font-body)', fontSize: '13px', color: 'var(--text)', background: 'var(--panel-raised)', border: '1px solid var(--steel)', borderRadius: 'var(--radius)', resize: 'vertical', minHeight: '60px' }} value={description} onChange={e => setDescription(e.target.value)} />
            </div>
          </div>
        </div>
      </div>

      {/* Model */}
      <div className="aw-card">
        <div className="aw-card__header"><span className="aw-card__title">Model</span></div>
        <div className="aw-card__body">
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)' }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-1)' }}>
              <label style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', fontWeight: 500, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--text-faint)' }}>Model</label>
              <select style={{ padding: 'var(--space-2) var(--space-3)', fontFamily: 'var(--font-mono)', fontSize: '13px', color: 'var(--text)', background: 'var(--panel-raised)', border: '1px solid var(--steel)', borderRadius: 'var(--radius)' }} value={model} onChange={e => setModel(e.target.value)}>
                {models.length > 0 ? models.map((m: any) => (
                  <option key={m.id ?? m.name} value={m.id ?? m.name}>{m.id ?? m.name}</option>
                )) : (
                  <>
                    <option value="gpt-4o-mini">gpt-4o-mini</option>
                    <option value="gpt-4o">gpt-4o</option>
                    <option value="claude-3-5-sonnet">claude-3-5-sonnet</option>
                  </>
                )}
              </select>
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
              <label style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', fontWeight: 500, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--text-faint)' }}>Thinking Mode</label>
              <div style={{ display: 'flex', gap: 'var(--space-1)' }}>
                {['off', 'auto', 'always'].map(mode => (
                  <button key={mode} className={`aw-nav-item ${thinkingMode === mode ? 'aw-nav-item--active' : ''}`} style={{ padding: 'var(--space-1) var(--space-2)', fontSize: '12px', textTransform: 'capitalize' }} onClick={() => setThinkingMode(mode)}>
                    {mode}
                  </button>
                ))}
              </div>
            </div>
            {thinkingMode !== 'off' && (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-1)' }}>
                <label style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', fontWeight: 500, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--text-faint)' }}>Thinking Budget: {thinkingBudget}</label>
                <input type="range" min="256" max="8192" step="256" value={thinkingBudget} onChange={e => setThinkingBudget(Number(e.target.value))} style={{ accentColor: 'var(--accent)' }} />
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Capabilities */}
      <div className="aw-card">
        <div className="aw-card__header"><span className="aw-card__title">Capabilities</span></div>
        <div className="aw-card__body">
          {Object.entries(PRESET_CAPABILITIES).map(([group, caps]) => (
            <div key={group} style={{ marginBottom: 'var(--space-3)' }}>
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', fontWeight: 600, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--text-faint)' }}>{group}</span>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 'var(--space-2)', marginTop: 'var(--space-2)' }}>
                {caps.map(cap => {
                  const enabled = isCapEnabled(cap);
                  return (
                    <button
                      key={cap}
                      className={`aw-nav-item ${enabled ? 'aw-nav-item--active' : ''}`}
                      style={{ padding: 'var(--space-1) var(--space-2)', fontSize: '11px', borderRadius: '100px' }}
                      onClick={() => toggleCap(cap)}
                    >
                      {enabled && '✓ '}{cap.replace(/_/g, ' ')}
                    </button>
                  );
                })}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Behavioral Limits */}
      <div className="aw-card">
        <div className="aw-card__header"><span className="aw-card__title">Behavioral Limits</span></div>
        <div className="aw-card__body">
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 'var(--space-3)' }}>
            {[
              { label: 'Max Execution Depth', value: maxDepth, setter: setMaxDepth },
              { label: 'Max Recursion Depth', value: maxRecursion, setter: setMaxRecursion },
              { label: 'Max Wall Time (ms)', value: maxWallTime, setter: setMaxWallTime },
              { label: 'Max Memory Growth (MB)', value: maxMemory, setter: setMaxMemory },
            ].map(f => (
              <div key={f.label} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-1)' }}>
                <label style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', fontWeight: 500, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--text-faint)' }}>{f.label}</label>
                <input type="number" style={{ padding: 'var(--space-2) var(--space-3)', fontFamily: 'var(--font-mono)', fontSize: '13px', color: 'var(--text)', background: 'var(--panel-raised)', border: '1px solid var(--steel)', borderRadius: 'var(--radius)' }} value={f.value} onChange={e => f.setter(e.target.value)} />
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Autonomous */}
      <div className="aw-card">
        <div className="aw-card__header"><span className="aw-card__title">Autonomous Mode</span></div>
        <div className="aw-card__body">
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <div>
                <span style={{ fontFamily: 'var(--font-body)', fontSize: '14px', fontWeight: 500, color: 'var(--text)' }}>Autonomous Execution</span>
                <p style={{ fontFamily: 'var(--font-body)', fontSize: '12px', color: 'var(--text-faint)', margin: '2px 0 0' }}>Allow self-directed execution without user prompts</p>
              </div>
              <button className={`aw-switch ${autonomousEnabled ? 'aw-switch--on' : ''}`} onClick={() => setAutonomousEnabled(!autonomousEnabled)}><span className="aw-switch__thumb" /></button>
            </div>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <div>
                <span style={{ fontFamily: 'var(--font-body)', fontSize: '14px', fontWeight: 500, color: 'var(--text)' }}>Self-Evolution</span>
                <p style={{ fontFamily: 'var(--font-body)', fontSize: '12px', color: 'var(--text-faint)', margin: '2px 0 0' }}>Allow the agent to propose and apply improvements</p>
              </div>
              <button className={`aw-switch ${evolutionEnabled ? 'aw-switch--on' : ''}`} onClick={() => setEvolutionEnabled(!evolutionEnabled)}><span className="aw-switch__thumb" /></button>
            </div>
          </div>
        </div>
      </div>

      {/* Danger Zone */}
      <div className="aw-card" style={{ borderColor: 'rgba(255, 107, 107, 0.3)' }}>
        <div className="aw-card__header"><span className="aw-card__title" style={{ color: 'var(--status-revoked)' }}>Danger Zone</span></div>
        <div className="aw-card__body">
          <p style={{ fontFamily: 'var(--font-body)', fontSize: '13px', color: 'var(--text-dim)', margin: '0 0 var(--space-3)' }}>
            Permanently delete this agent and all associated data. This action cannot be undone.
          </p>
          <button
            className="aw-nav-item"
            style={{ color: 'var(--status-revoked)', borderColor: 'rgba(255,107,107,0.3)', border: '1px solid rgba(255,107,107,0.3)', padding: 'var(--space-2) var(--space-3)' }}
            onClick={handleDelete}
            disabled={deleteAgent.isPending}
          >
            <Trash2 size={14} />
            {deleteConfirm ? 'Click again to confirm deletion' : 'Delete Agent'}
          </button>
        </div>
      </div>
    </div>
  );
}
