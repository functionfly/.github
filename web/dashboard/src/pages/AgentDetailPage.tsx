import { agentApi } from '@/api/agent';
import { useAgent, useAgentPolicy, useAgentUsage, useUpdateAgent, useUpdateAgentPolicy } from '@/hooks/useAgent';
import { useAgentChildren } from '@/hooks/useAgentSwarm';
import { useAgentMemories } from '@/hooks/useAgentMemory';
import { ROUTES } from '@/lib/constants';
import { usePageTitle } from '@/hooks';
import {
  ArrowLeft, Bot, Brain, Copy, Check, GitBranch, Loader2, MemoryStick, Save, Settings, Trash2, Wallet, BarChart3, Plus, X, Terminal, LayoutDashboard,
} from 'lucide-react';
import { AgentChatHistory } from '@/components/agent-console';
import { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';
import { AgentMemoryGraph } from './AgentDetailPage/components/AgentMemoryGraph';
import { SwarmTopologyView } from '@/components/topology';
import AgentAnalyticsComponent from '@/components/analytics';
import {
  PageGrid, Chamber, SealedButton, FrameButton, StatusPill, Card,
} from '@/components/containment';
import './AgentDetailPage.css';

const PRESET_CAPABILITIES = [
  {
    category: 'Data & Search',
    items: [
      { key: 'web_search', label: 'Web Search', description: 'Search the internet for real-time information' },
      { key: 'database_query', label: 'Database', description: 'Read and write to databases' },
      { key: 'file_read', label: 'File Read', description: 'Read files from storage' },
      { key: 'file_write', label: 'File Write', description: 'Write and modify files' },
    ],
  },
  {
    category: 'Communication',
    items: [
      { key: 'http_request', label: 'HTTP Requests', description: 'Make outbound HTTP calls to external APIs' },
      { key: 'email', label: 'Email', description: 'Send and receive emails' },
      { key: 'notification', label: 'Notifications', description: 'Send push or in-app notifications' },
    ],
  },
  {
    category: 'AI & Processing',
    items: [
      { key: 'code_execution', label: 'Code Execution', description: 'Run code in a sandboxed environment' },
      { key: 'image_generation', label: 'Image Generation', description: 'Generate images from text prompts' },
      { key: 'text_to_speech', label: 'Text to Speech', description: 'Convert text to spoken audio' },
      { key: 'speech_to_text', label: 'Speech to Text', description: 'Transcribe audio to text' },
    ],
  },
];

function sanitizeAgentIdParam(raw: string | undefined): string | null {
  const trimmed = raw?.trim();
  if (!trimmed || trimmed === 'undefined' || trimmed === 'null') return null;
  return trimmed;
}

const statusToPill = (status: string): 'live' | 'pending' | 'revoked' => {
  if (status === 'active') return 'live';
  if (status === 'suspended') return 'revoked';
  return 'pending';
};

export function AgentDetailPage() {
  usePageTitle('Agent Detail');
  const { t } = useTranslation();
  const { id: pathAgentId } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const agentId = sanitizeAgentIdParam(pathAgentId);

  const [searchParams, setSearchParams] = useSearchParams();
  const urlTab = searchParams.get('tab') || 'overview';

  const [activeTab, setActiveTab] = useState(urlTab);
  const [deleting, setDeleting] = useState(false);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [autonomousEnabled, setAutonomousEnabled] = useState(false);
  const [evolutionEnabled, setEvolutionEnabled] = useState(false);
  const [maxExecutionDepth, setMaxExecutionDepth] = useState(5);
  const [maxRecursionDepth, setMaxRecursionDepth] = useState(3);
  const [maxWallTimeMs, setMaxWallTimeMs] = useState(30000);
  const [maxMemoryGrowthMB, setMaxMemoryGrowthMB] = useState(512);
  const [saving, setSaving] = useState(false);
  const [capabilities, setCapabilities] = useState<Record<string, unknown>>({});
  const [newCapKey, setNewCapKey] = useState('');
  const [newCapValue, setNewCapValue] = useState('');
  const [copied, setCopied] = useState(false);
  const [model, setModel] = useState('gpt-4o-mini');
  const [thinkingMode, setThinkingMode] = useState<'off' | 'auto' | 'always'>('off');
  const [thinkingBudget, setThinkingBudget] = useState(10000);
  const [models, setModels] = useState<Array<{ id: string; name: string; provider: string; provider_label?: string; key_source?: string; tier: string; cost: string }>>([]);

  const { data: agentData, isLoading: loading, error } = useAgent(agentId ?? '');
  const { data: policyData } = useAgentPolicy(agentId ?? '');
  const { data: usageData } = useAgentUsage(agentId ?? '');
  const { data: childrenData } = useAgentChildren(agentId ?? '');
  const { data: memoriesData } = useAgentMemories({ agent_id: agentId ?? undefined }, !!agentId);
  const updateAgent = useUpdateAgent(agentId ?? '');
  const updatePolicy = useUpdateAgentPolicy(agentId ?? '');

  const agent = agentData?.agent ?? null;
  const policy = policyData?.policy;
  const usage = usageData?.usage;
  const children = childrenData?.children ?? [];
  const memories = memoriesData?.memories ?? [];

  useEffect(() => {
    if (agent) {
      setName(agent.name || '');
      setDescription(agent.description || '');
      setAutonomousEnabled(agent.autonomousEnabled ?? false);
      setEvolutionEnabled(agent.evolutionEnabled ?? false);
      setCapabilities(agent.capabilities ?? {});
      setModel(agent.model || 'gpt-4o-mini');
      setThinkingMode((agent.thinking_mode as 'off' | 'auto' | 'always') || 'off');
      setThinkingBudget(agent.thinking_budget || 10000);
    }
  }, [agent]);

  useEffect(() => {
    agentApi.getModels().then((res) => setModels(res.models)).catch(() => {});
  }, []);

  useEffect(() => {
    if (policy) {
      setMaxExecutionDepth(policy.maxExecutionDepth ?? 5);
      setMaxRecursionDepth(policy.maxRecursionDepth ?? 3);
      setMaxWallTimeMs(policy.maxWallTimeMs ?? 30000);
      setMaxMemoryGrowthMB(policy.maxMemoryGrowthMB ?? 512);
    }
  }, [policy]);

  useEffect(() => {
    if (urlTab && urlTab !== activeTab) {
      setActiveTab(urlTab);
    }
  }, [urlTab]);

  const isDirty = useMemo(() => {
    if (!agent) return false;
    if (name !== (agent.name || '')) return true;
    if (description !== (agent.description || '')) return true;
    if (autonomousEnabled !== (agent.autonomousEnabled ?? false)) return true;
    if (evolutionEnabled !== (agent.evolutionEnabled ?? false)) return true;
    if (JSON.stringify(capabilities) !== JSON.stringify(agent.capabilities ?? {})) return true;
    if (model !== (agent.model || 'gpt-4o-mini')) return true;
    if (policy) {
      if (maxExecutionDepth !== (policy.maxExecutionDepth ?? 5)) return true;
      if (maxRecursionDepth !== (policy.maxRecursionDepth ?? 3)) return true;
      if (maxWallTimeMs !== (policy.maxWallTimeMs ?? 30000)) return true;
      if (maxMemoryGrowthMB !== (policy.maxMemoryGrowthMB ?? 512)) return true;
    }
    return false;
  }, [agent, policy, name, description, autonomousEnabled, evolutionEnabled, capabilities, model, maxExecutionDepth, maxRecursionDepth, maxWallTimeMs, maxMemoryGrowthMB]);

  const blocker = null;

  useEffect(() => {
    const handler = (e: BeforeUnloadEvent) => {
      if (isDirty) { e.preventDefault(); }
    };
    window.addEventListener('beforeunload', handler);
    return () => window.removeEventListener('beforeunload', handler);
  }, [isDirty]);

  const handleDelete = async () => {
    if (!agentId || !confirm(t('agentDetail.confirmDelete', { name: agent?.name ?? agentId }))) return;
    setDeleting(true);
    try {
      await agentApi.deleteAgent(agentId);
      toast.success(t('agentDetail.agentDeleted'));
      navigate(ROUTES.AGENT_LIST);
    } catch (err) {
      console.error('Failed to delete agent:', err);
      toast.error(t('agentDetail.failedToDelete'));
    } finally {
      setDeleting(false);
    }
  };

  const handleSaveSettings = async () => {
    if (!agentId) return;
    setSaving(true);
    try {
      await updateAgent.mutateAsync({ name, description, capabilities, autonomous_enabled: autonomousEnabled, evolution_enabled: evolutionEnabled, model, thinking_mode: thinkingMode, thinking_budget: thinkingBudget });
      await updatePolicy.mutateAsync({ agentId, maxExecutionDepth, maxRecursionDepth, maxWallTimeMs, maxMemoryGrowthMB });
      toast.success(t('agentDetail.settingsSaved'));
    } catch (err) {
      console.error('Failed to save agent settings:', err);
      toast.error(t('agentDetail.failedToSave'));
    } finally {
      setSaving(false);
    }
  };

  const handleAddCapability = () => {
    const key = newCapKey.trim();
    if (!key) return;
    setCapabilities((prev) => ({ ...prev, [key]: newCapValue || true }));
    setNewCapKey('');
    setNewCapValue('');
  };

  const handleRemoveCapability = (key: string) => {
    setCapabilities((prev) => {
      const next = { ...prev };
      delete next[key];
      return next;
    });
  };

  const handleCopyAgentName = async () => {
    const value = agent?.name || agentId;
    if (!value) return;
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      toast.success(t('agentDetail.agentNameCopied'));
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error(t('agentDetail.copyFailed'));
    }
  };

  if (!agentId) {
    return (
      <div className="adp-page">
        <PageGrid />
        <FrameButton size="sm" onClick={() => navigate(ROUTES.AGENT_LIST)} iconLeft={<ArrowLeft className="adp-icon-sm" />}>
          {t('agentDetail.backToAgents')}
        </FrameButton>
        <p className="adp-hint">{t('agentDetail.invalidAgentLink')}</p>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="adp-page">
        <PageGrid />
        <div className="adp-loading"><Loader2 className="adp-loading__spinner" /><p className="adp-loading__text">{t('agentDetail.loadingAgent')}</p></div>
      </div>
    );
  }

  if (error || !agent) {
    return (
      <div className="adp-page">
        <PageGrid />
        <FrameButton size="sm" onClick={() => navigate(ROUTES.AGENT_LIST)} iconLeft={<ArrowLeft className="adp-icon-sm" />}>
          {t('agentDetail.backToAgents')}
        </FrameButton>
        <Chamber className="adp-error">
          <h2 className="adp-error__title">{t('agentDetail.agentNotFound')}</h2>
          <p className="adp-error__desc">{error instanceof Error ? error.message : t('agentDetail.agentNotFoundDescription')}</p>
        </Chamber>
      </div>
    );
  }

  const caps = agent.capabilities ? Object.keys(agent.capabilities) : [];

  const tabs = [
    { value: 'overview', label: t('agentDetail.overview'), icon: Brain },
    { value: 'settings', label: t('agentDetail.settings'), icon: Settings },
    { value: 'console', label: 'Console', icon: Terminal },
    { value: 'topology', label: t('agentDetail.topology'), icon: GitBranch },
    { value: 'analytics', label: t('agentDetail.analytics'), icon: BarChart3 },
    { value: 'memory', label: t('agentDetail.memoryGraph'), icon: MemoryStick },
  ];

  return (
    <div className="adp-page">
      <PageGrid />

      {/* Header */}
      <div className="adp-header">
        <div className="adp-header__left">
          <FrameButton size="sm" onClick={() => navigate(ROUTES.AGENT_LIST)} iconLeft={<ArrowLeft className="adp-icon-sm" />}>
            {t('agentDetail.backToAgents')}
          </FrameButton>
          <div className="adp-header__identity">
            <div className="adp-header__icon-wrap"><Bot className="adp-header__icon" /></div>
            <div>
              <h1 className="adp-header__title">{agent.name || agent.agentId}</h1>
              <button className="adp-header__meta-btn" onClick={handleCopyAgentName} title={t('agentDetail.copyAgentName')}>
                <span className="adp-header__meta">{agent.name || agent.agentId}</span>
                {copied ? <Check className="adp-icon-xs adp-copy-ok" /> : <Copy className="adp-icon-xs adp-copy-icon" />}
              </button>
            </div>
          </div>
        </div>
        <div className="adp-header__actions">
          <SealedButton size="sm" onClick={() => navigate(`/agents/${encodeURIComponent(agent?.agentId ?? agentId ?? '')}/workspace`)} iconLeft={<LayoutDashboard className="adp-icon-sm" />}>Workspace</SealedButton>
          <FrameButton size="sm" onClick={() => navigate(ROUTES.agentAnalyticsPath(agent?.agentId ?? agentId ?? ''))} iconLeft={<BarChart3 className="adp-icon-sm" />}>{t('agentDetail.analytics')}</FrameButton>
          <FrameButton size="sm" onClick={() => navigate(ROUTES.agentWalletPath(agent?.agentId ?? agentId ?? ''))} iconLeft={<Wallet className="adp-icon-sm" />}>{t('agentDetail.wallet')}</FrameButton>
        </div>
      </div>

      {/* Status Badges */}
      <div className="adp-badges">
        <StatusPill status={statusToPill(agent.status)} label={agent.status} />
        {agent.swarmRole && <span className="adp-role-badge">{agent.swarmRole}</span>}
        {agent.parentAgentId && (
          <span className="adp-parent-badge">
            {t('agentDetail.childOf')}{' '}
            <Link className="adp-parent-link" to={`/agents/${encodeURIComponent(agent.parentAgentId)}`}>{agent.parentAgentId}</Link>
          </span>
        )}
        {agent.autonomousEnabled && <span className="adp-feature-badge">{t('agentDetail.autonomous')}</span>}
        {agent.evolutionEnabled && <span className="adp-feature-badge">{t('agentDetail.evolutionBadge')}</span>}
      </div>

      {/* Tabs */}
      <div className="adp-tabs">
        {tabs.map((tab) => (
          <button key={tab.value} className={`adp-tab ${activeTab === tab.value ? 'adp-tab--active' : ''}`}
            onClick={() => { setActiveTab(tab.value); setSearchParams({ tab: tab.value }); }}>
            <tab.icon className="adp-icon-sm" />
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab Content */}
      {activeTab === 'overview' && (
        <div className="adp-tab-content">
          <div className="adp-grid-2">
            <Card className="adp-card">
              <div className="adp-card__header">
                <h3 className="adp-card__title">{t('agentDetail.overview')}</h3>
                <p className="adp-card__desc">{t('agentDetail.overviewDescription')}</p>
              </div>
              <div className="adp-card__body">
                {agent.description ? <p className="adp-text">{agent.description}</p> : <p className="adp-text adp-text--muted">{t('agentDetail.noDescription')}</p>}
                {caps.length > 0 && (
                  <div className="adp-caps">
                    <p className="adp-caps__label">{t('agentDetail.capabilities')}</p>
                    <div className="adp-caps__list">{caps.map((c) => <span key={c} className="adp-cap-tag">{c.replace(/_/g, ' ')}</span>)}</div>
                  </div>
                )}
              </div>
            </Card>

            <Card className="adp-card">
              <div className="adp-card__header">
                <h3 className="adp-card__title">{t('agentDetail.currentUsage')}</h3>
                <p className="adp-card__desc">{t('agentDetail.realtimeAgentMetrics')}</p>
              </div>
              <div className="adp-card__body">
                {usage ? (
                  <div className="adp-stats-grid">
                    <div className="adp-stat"><p className="adp-stat__label">{t('agentDetail.callsThisMinute')}</p><p className="adp-stat__value">{usage.callsThisMinute}</p></div>
                    <div className="adp-stat"><p className="adp-stat__label">{t('agentDetail.callsToday')}</p><p className="adp-stat__value">{usage.callsToday}</p></div>
                    <div className="adp-stat"><p className="adp-stat__label">{t('agentDetail.spendToday')}</p><p className="adp-stat__value">${usage.spendTodayUSD.toFixed(4)}</p></div>
                    <div className="adp-stat"><p className="adp-stat__label">{t('agentDetail.spendThisMonth')}</p><p className="adp-stat__value">${usage.spendThisMonthUSD.toFixed(4)}</p></div>
                  </div>
                ) : <p className="adp-text adp-text--muted">{t('agentDetail.noUsageData')}</p>}
              </div>
            </Card>
          </div>

          {policy && (
            <Card className="adp-card">
              <div className="adp-card__header">
                <h3 className="adp-card__title">{t('agentDetail.behavioralPolicy')}</h3>
                <p className="adp-card__desc">{t('agentDetail.safetyLimits')}</p>
              </div>
              <div className="adp-card__body">
                <div className="adp-stats-grid adp-stats-grid--4">
                  <div className="adp-stat"><p className="adp-stat__label">{t('agentDetail.maxExecutionDepth')}</p><p className="adp-stat__value">{policy.maxExecutionDepth}</p></div>
                  <div className="adp-stat"><p className="adp-stat__label">{t('agentDetail.maxRecursionDepth')}</p><p className="adp-stat__value">{policy.maxRecursionDepth}</p></div>
                  <div className="adp-stat"><p className="adp-stat__label">{t('agentDetail.maxWallTime')}</p><p className="adp-stat__value">{policy.maxWallTimeMs}ms</p></div>
                  <div className="adp-stat"><p className="adp-stat__label">{t('agentDetail.maxMemoryGrowth')}</p><p className="adp-stat__value">{policy.maxMemoryGrowthMB}MB</p></div>
                </div>
              </div>
            </Card>
          )}
        </div>
      )}

      {activeTab === 'settings' && (
        <Card className="adp-card">
          <div className="adp-card__header">
            <h3 className="adp-card__title">{t('agentDetail.agentSettings')}</h3>
            <p className="adp-card__desc">{t('agentDetail.configureIdentity')}</p>
          </div>
          <div className="adp-card__body adp-card__body--form">
            <div className="adp-field">
              <label className="adp-label">{t('agentDetail.name')}</label>
              <input className="adp-input" value={name} onChange={(e) => setName(e.target.value)} placeholder={t('agentDetail.namePlaceholder')} />
            </div>
            <div className="adp-field">
              <label className="adp-label">{t('agentDetail.description')}</label>
              <textarea className="adp-textarea" value={description} onChange={(e) => setDescription(e.target.value)} placeholder={t('agentDetail.descriptionPlaceholder')} rows={3} />
            </div>
            <div className="adp-field">
              <label className="adp-label">AI Model</label>
              <p className="adp-field__hint">Select the model this agent uses for reasoning and chat.</p>
              <select className="adp-input" value={model} onChange={(e) => setModel(e.target.value)}>
                {models.length === 0 && <option value="gpt-4o-mini">GPT-4o Mini (default)</option>}
                {(() => {
                  const grouped: Record<string, typeof models> = {};
                  models.forEach((m) => {
                    const key = m.tier;
                    if (!grouped[key]) grouped[key] = [];
                    grouped[key].push(m);
                  });
                  const tierLabels: Record<string, string> = { frontier: 'Frontier', fast: 'Fast', reasoning: 'Reasoning', code: 'Code', free: 'Free' };
                  return Object.entries(grouped).map(([tier, items]) => (
                    <optgroup key={tier} label={tierLabels[tier] || tier}>
                      {items.map((m) => (
                        <option key={m.id} value={m.id}>
                          {m.name} — {m.provider_label || m.provider} ({m.cost})
                        </option>
                      ))}
                    </optgroup>
                  ));
                })()}
              </select>
            </div>
            <div className="adp-field">
              <label className="adp-label">Thinking Mode</label>
              <p className="adp-field__hint">Enable provider-native reasoning for complex tasks. Auto activates thinking for long or analytical queries.</p>
              <div style={{ display: 'flex', gap: '8px', marginTop: '4px' }}>
                {(['off', 'auto', 'always'] as const).map((mode) => (
                  <button
                    key={mode}
                    type="button"
                    className={`adp-toggle-btn ${thinkingMode === mode ? 'adp-toggle-btn--active' : ''}`}
                    onClick={() => setThinkingMode(mode)}
                    style={{
                      padding: '6px 16px',
                      borderRadius: '6px',
                      border: `1px solid ${thinkingMode === mode ? 'var(--accent, #6366f1)' : 'var(--border-subtle, rgba(255,255,255,0.1))'}`,
                      background: thinkingMode === mode ? 'var(--accent-bg, rgba(99,102,241,0.1))' : 'transparent',
                      color: thinkingMode === mode ? 'var(--accent, #6366f1)' : 'var(--text-secondary, #a1a1aa)',
                      cursor: 'pointer',
                      fontSize: '13px',
                      fontFamily: 'var(--font-body)',
                      textTransform: 'capitalize',
                    }}
                  >
                    {mode}
                  </button>
                ))}
              </div>
            </div>
            {thinkingMode !== 'off' && (
              <div className="adp-field">
                <label className="adp-label">Thinking Budget</label>
                <p className="adp-field__hint">Maximum tokens allocated for reasoning per request. Higher values enable deeper analysis but increase cost.</p>
                <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginTop: '4px' }}>
                  <input
                    type="range"
                    min={1000}
                    max={50000}
                    step={1000}
                    value={thinkingBudget}
                    onChange={(e) => setThinkingBudget(Number(e.target.value))}
                    style={{ flex: 1 }}
                  />
                  <span style={{ fontSize: '13px', color: 'var(--text-secondary, #a1a1aa)', minWidth: '80px', textAlign: 'right', fontFamily: 'var(--font-mono, monospace)' }}>
                    {thinkingBudget.toLocaleString()}
                  </span>
                </div>
              </div>
            )}
            <div className="adp-toggles">
              <label className="adp-toggle-label">
                <input type="checkbox" checked={autonomousEnabled} onChange={(e) => setAutonomousEnabled(e.target.checked)} className="adp-checkbox" />
                {t('agentDetail.autonomousMode')}
              </label>
              <label className="adp-toggle-label">
                <input type="checkbox" checked={evolutionEnabled} onChange={(e) => setEvolutionEnabled(e.target.checked)} className="adp-checkbox" />
                {t('agentDetail.evolutionEnabled')}
              </label>
            </div>
            <div className="adp-fields-section">
              <h4 className="adp-fields-section__title">{t('agentDetail.capabilities')}</h4>
              <p className="adp-fields-section__desc">{t('agentDetail.capabilitiesDescription')}</p>

              <div className="adp-presets-grid">
                {PRESET_CAPABILITIES.map((group) => (
                  <div key={group.category} className="adp-presets-group">
                    <p className="adp-presets-group__label">{group.category}</p>
                    <div className="adp-presets-chips">
                      {group.items.map((cap) => {
                        const active = capabilities[cap.key] !== undefined;
                        return (
                          <button
                            key={cap.key}
                            type="button"
                            className={`adp-preset-chip ${active ? 'adp-preset-chip--active' : ''}`}
                            onClick={() => {
                              if (active) handleRemoveCapability(cap.key);
                              else setCapabilities((prev) => ({ ...prev, [cap.key]: true }));
                            }}
                            title={cap.description}
                          >
                            {active && <span className="adp-preset-chip__check">&#10003;</span>}
                            {cap.label}
                          </button>
                        );
                      })}
                    </div>
                  </div>
                ))}
              </div>

              {Object.keys(capabilities).some((k) => !PRESET_CAPABILITIES.flatMap((g) => g.items).some((p) => p.key === k)) && (
                <div className="adp-custom-caps">
                  <p className="adp-presets-group__label">{t('agentDetail.customCapabilities')}</p>
                  <div className="adp-custom-caps-list">
                    {Object.entries(capabilities)
                      .filter(([k]) => !PRESET_CAPABILITIES.flatMap((g) => g.items).some((p) => p.key === k))
                      .map(([key, val]) => (
                        <div key={key} className="adp-cap-entry">
                          <span className="adp-cap-key">{key}</span>
                          <span className="adp-cap-val">{String(val)}</span>
                          <button type="button" className="adp-cap-remove" onClick={() => handleRemoveCapability(key)} aria-label={t('agentDetail.removeCapability')}>
                            <X className="adp-icon-xs" />
                          </button>
                        </div>
                      ))}
                  </div>
                </div>
              )}

              <div className="adp-cap-add-row">
                <input className="adp-input adp-input--sm" value={newCapKey} onChange={(e) => setNewCapKey(e.target.value)} placeholder={t('agentDetail.capKeyPlaceholder')} onKeyDown={(e) => e.key === 'Enter' && (e.preventDefault(), handleAddCapability())} />
                <input className="adp-input adp-input--sm" value={newCapValue} onChange={(e) => setNewCapValue(e.target.value)} placeholder={t('agentDetail.capValuePlaceholder')} onKeyDown={(e) => e.key === 'Enter' && (e.preventDefault(), handleAddCapability())} />
                <button type="button" className="adp-cap-add-btn" onClick={handleAddCapability} disabled={!newCapKey.trim()}>
                  <Plus className="adp-icon-xs" /> {t('agentDetail.addCapability')}
                </button>
              </div>
            </div>
            <div className="adp-fields-section">
              <h4 className="adp-fields-section__title">{t('agentDetail.behavioralLimits')}</h4>
              <div className="adp-fields-grid">
                <div className="adp-field"><label className="adp-label">{t('agentDetail.maxExecutionDepth')}</label><input className="adp-input" type="number" value={maxExecutionDepth} onChange={(e) => setMaxExecutionDepth(parseInt(e.target.value) || 0)} min={1} max={20} /></div>
                <div className="adp-field"><label className="adp-label">{t('agentDetail.maxRecursionDepth')}</label><input className="adp-input" type="number" value={maxRecursionDepth} onChange={(e) => setMaxRecursionDepth(parseInt(e.target.value) || 0)} min={0} max={10} /></div>
                <div className="adp-field"><label className="adp-label">{t('agentDetail.maxWallTimeMs')}</label><input className="adp-input" type="number" value={maxWallTimeMs} onChange={(e) => setMaxWallTimeMs(parseInt(e.target.value) || 0)} min={1000} step={1000} /></div>
                <div className="adp-field"><label className="adp-label">{t('agentDetail.maxMemoryGrowthMB')}</label><input className="adp-input" type="number" value={maxMemoryGrowthMB} onChange={(e) => setMaxMemoryGrowthMB(parseInt(e.target.value) || 0)} min={64} step={64} /></div>
              </div>
            </div>
            <SealedButton onClick={handleSaveSettings} disabled={saving || !isDirty} loading={saving} iconLeft={<Save className="adp-icon-sm" />}>
              {saving ? t('agentDetail.saving') : isDirty ? t('agentDetail.saveSettings') : t('agentDetail.noChanges')}
            </SealedButton>
            <div className="adp-danger-zone">
              <h4 className="adp-danger-zone__title">{t('agentDetail.dangerZone')}</h4>
              <p className="adp-danger-zone__desc">{t('agentDetail.dangerZoneDescription')}</p>
              <button className="adp-delete-btn" onClick={() => void handleDelete()} disabled={deleting}>
                {deleting ? <Loader2 className="adp-icon-sm adp-spin" /> : <Trash2 className="adp-icon-sm" />}
                {t('agentDetail.delete')}
              </button>
            </div>
          </div>
        </Card>
      )}

      {activeTab === 'console' && (
        <Card className="adp-card">
          <div className="adp-card__header">
            <h3 className="adp-card__title">{t('agentDetail.consoleTitle')}</h3>
            <p className="adp-card__desc">{t('agentDetail.consoleDescription', { name: agent?.name || agentId, model })}</p>
          </div>
          <div className="adp-card__body">
            <AgentChatHistory agentId={agentId ?? ''} agentName={agent?.name || (agentId ?? '')} model={model} />
          </div>
        </Card>
      )}

      {activeTab === 'topology' && (
        <Card className="adp-card">
          <div className="adp-card__header">
            <h3 className="adp-card__title">{t('agentDetail.swarmTopology')}</h3>
            <p className="adp-card__desc">{t('agentDetail.swarmTopologyDescription')}</p>
          </div>
          <div className="adp-card__body">
            {children.length > 0 ? <SwarmTopologyView agentId={agentId ?? ''} /> : (
              <div className="adp-empty"><GitBranch className="adp-empty__icon" /><p className="adp-empty__title">{t('agentDetail.noChildAgents')}</p><p className="adp-empty__desc">{t('agentDetail.childAgentsWillAppear')}</p></div>
            )}
          </div>
        </Card>
      )}

      {activeTab === 'analytics' && (
        <Card className="adp-card">
          <div className="adp-card__header">
            <h3 className="adp-card__title">{t('agentDetail.agentAnalytics')}</h3>
            <p className="adp-card__desc">{t('agentDetail.analyticsDescription')}</p>
          </div>
          <div className="adp-card__body"><AgentAnalyticsComponent agentId={agentId ?? ''} /></div>
        </Card>
      )}

      {activeTab === 'memory' && (
        <Card className="adp-card">
          <div className="adp-card__header">
            <h3 className="adp-card__title">{t('agentDetail.memoryVisualization')}</h3>
            <p className="adp-card__desc">{t('agentDetail.memoryGraphDescription')}</p>
          </div>
          <div className="adp-card__body">
            {memories.length > 0 ? <AgentMemoryGraph memories={memories} agentId={agentId} /> : (
              <div className="adp-empty"><Brain className="adp-empty__icon" /><p className="adp-empty__title">{t('agentDetail.noMemoriesFound')}</p><p className="adp-empty__desc">{t('agentDetail.memoriesWillAppear')}</p></div>
            )}
          </div>
        </Card>
      )}
    </div>
  );
}

export default AgentDetailPage;
