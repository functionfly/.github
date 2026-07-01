import { useState, useMemo } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { usePageTitle } from '@/hooks';
import { useAgents, useDeleteAgent } from '@/hooks/useAgent';
import { agentApi, type AgentIdentity } from '@/api/agent';
import {
  Bot, Plus, Puzzle, Search, Settings, Trash2, MoreVertical,
  Loader2, Copy, Check, LayoutGrid, List, Edit3, Eye, RefreshCw,
} from 'lucide-react';
import { ROUTES } from '@/lib/constants';
import { canCreateAgent, getAgentsLimit, hasFeature } from '@/lib/plan-utils';
import { usePlan } from '@/hooks/usePlan';
import { toast } from 'sonner';
import {
  PageGrid, Chamber, CornerBrace, TrustSeal,
  SealedButton, FrameButton, StatusPill, AnnotationTag, Card,
} from '@/components/containment';
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem,
  DropdownMenuSeparator, DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import './AgentsPage/styles.css';

const statusToPill = (status: string): 'live' | 'pending' | 'revoked' => {
  if (status === 'active') return 'live';
  if (status === 'suspended') return 'revoked';
  return 'pending';
};

export function AgentsPage() {
  usePageTitle('Agents');
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { username } = useParams();
  const { plan } = usePlan();

  const { data, isLoading, error, refetch } = useAgents({ limit: 100 });
  const deleteAgent = useDeleteAgent();
  const agents = data?.agents ?? [];

  const [searchQuery, setSearchQuery] = useState('');
  const [createOpen, setCreateOpen] = useState(false);
  const [createSubmitting, setCreateSubmitting] = useState(false);
  const [createForm, setCreateForm] = useState({ agentId: '', name: '', description: '' });
  const [createdApiKey, setCreatedApiKey] = useState<string | null>(null);
  const [apiKeyCopied, setApiKeyCopied] = useState(false);
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');

  const agentCount = agents.length;
  const canCreate = canCreateAgent(plan, agentCount);
  const agentsUnlocked = hasFeature(plan, 'AGENTS');
  const agentsLimit = getAgentsLimit(plan);

  const slugFrom = (s: string) => (s ?? '').trim().toLowerCase().replace(/\s+/g, '-').replace(/[^a-z0-9-]/g, '');
  const agentIdRaw = createForm.agentId ?? '';
  const agentIdInvalid = agentIdRaw.trim() !== '' && /[^a-z0-9-]/.test(agentIdRaw.trim().toLowerCase());
  const normalizedAgentId = slugFrom(agentIdRaw);
  const existingAgentIds = agents.map((a) => (a.agentId ?? '').toLowerCase());
  const agentIdTaken = normalizedAgentId.length > 0 && existingAgentIds.includes(normalizedAgentId);
  const [agentIdTakenFromSubmit, setAgentIdTakenFromSubmit] = useState(false);
  const showAgentIdTaken = agentIdTaken || agentIdTakenFromSubmit;

  const handleCreateSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const agentId = (createForm.agentId ?? '').trim().toLowerCase().replace(/\s+/g, '-');
    const name = (createForm.name ?? '').trim();
    if (!agentId || !name) { toast.error(t('agents.agentIdNameRequired')); return; }
    setCreateSubmitting(true);
    try {
      const res = await agentApi.registerAgent({ agentId, name, description: createForm.description.trim() || undefined });
      setCreatedApiKey(res.api_key);
      toast.success(t('agents.agentCreatedSuccess'));
      setCreateForm({ agentId: '', name: '', description: '' });
    } catch (err: unknown) {
      const res = err && typeof err === 'object' && 'response' in err ? (err as { response?: { status?: number; data?: { error?: { code?: string; message?: string }; message?: string } } }).response : undefined;
      const isTaken = res?.status === 409 || res?.data?.error?.code === 'AGENT_ID_TAKEN' || (typeof res?.data?.message === 'string' && /already|duplicate|in use/i.test(res.data.message));
      if (isTaken) { setAgentIdTakenFromSubmit(true); toast.error(t('agents.agentIdInUse')); }
      else { toast.error(typeof res?.data?.error === 'object' ? res.data.error.message : (res?.data?.message ?? t('agents.failedToCreate'))); }
    } finally { setCreateSubmitting(false); }
  };

  const handleCreateClose = (open: boolean) => {
    if (!open) { setCreatedApiKey(null); setApiKeyCopied(false); setAgentIdTakenFromSubmit(false); setCreateForm({ agentId: '', name: '', description: '' }); refetch(); }
    setCreateOpen(open);
  };

  const copyApiKey = async () => {
    if (!createdApiKey) return;
    try { await navigator.clipboard.writeText(createdApiKey); setApiKeyCopied(true); toast.success(t('agents.apiKeyCopied')); setTimeout(() => setApiKeyCopied(false), 2000); }
    catch { toast.error(t('agents.failedToCopy')); }
  };

  const filteredAgents = agents.filter((agent) =>
    (agent.name ?? '').toLowerCase().includes(searchQuery.toLowerCase()) ||
    (agent.agentId ?? '').toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleDelete = async (agent: AgentIdentity) => {
    if (!confirm(t('agents.confirmDelete', { name: agent.name ?? agent.agentId }))) return;
    try { await deleteAgent.mutateAsync(agent.id); } catch { /* toast handled by mutation */ }
  };

  if (isLoading) {
    return (
      <div className="ag-page">
        <PageGrid />
        <div className="ag-loading"><Loader2 className="ag-loading__spinner" /></div>
      </div>
    );
  }

  return (
    <div className="ag-page">
      <PageGrid />

      {/* Hero */}
      <Chamber className="ag-hero" ribs>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="MODULE AG-01" secondary="Agents" position="top-right" />
        <div className="ag-hero__header">
          <div className="ag-hero__title-row">
            <TrustSeal size="lg" />
            <h1 className="ag-hero__title">{username ? `${username}'s agents` : t('agents.title')}</h1>
          </div>
          <p className="ag-hero__subtitle">{t('agents.manageDescription')}</p>
          <div className="ag-hero__actions">
            <FrameButton size="sm" onClick={() => navigate(ROUTES.SDK_INTEGRATIONS)} iconLeft={<Puzzle className="ag-icon-sm" />}>{t('agents.sdkSetup')}</FrameButton>
            <SealedButton size="sm" onClick={() => navigate(ROUTES.AGENT_NEW)} disabled={!canCreate}
              title={!agentsUnlocked ? t('agents.agentsOnStarter') : !canCreate ? t('agents.planLimitReached', { count: agentCount, limit: agentsLimit >= 10000 ? '∞' : agentsLimit }) : undefined}
              iconLeft={<Plus className="ag-icon-sm" />}>{t('agents.createAgent')}</SealedButton>
          </div>
        </div>
      </Chamber>

      {!canCreate && (
        <p className="ag-limit-hint">
          {!agentsUnlocked ? <>{t('agents.upgradeToRegister')} <Link to={ROUTES.PRICING} className="ag-link">{t('agents.viewPlans')}</Link></>
            : <>{t('agents.agentLimitReached')} <Link to={ROUTES.PRICING} className="ag-link">{t('agents.upgrade')}</Link></>}
        </p>
      )}

      {/* Create Modal */}
      {createOpen && (
        <div className="ag-modal-overlay" onClick={() => handleCreateClose(false)}>
          <div className="ag-modal" onClick={(e) => e.stopPropagation()}>
            <div className="ag-modal__header">
              <h2 className="ag-modal__title"><Bot className="ag-icon-sm" /> {t('agents.createAgent')}</h2>
              <p className="ag-modal__desc">{t('agents.registerNewAgent')}</p>
            </div>
            {createdApiKey ? (
              <div className="ag-modal__body">
                <div className="ag-alert ag-alert--ok"><Check className="ag-icon-sm" /> {t('agents.agentCreatedCopyKey')}</div>
                <div className="ag-key-display">
                  <code className="ag-key-display__code">{createdApiKey}</code>
                  <button className="ag-key-display__copy" onClick={copyApiKey}>
                    {apiKeyCopied ? <><Check className="ag-icon-xs ag-icon-ok" /> Copied</> : <><Copy className="ag-icon-xs" /> Copy</>}
                  </button>
                </div>
                <div className="ag-modal__footer"><SealedButton onClick={() => handleCreateClose(false)}>{t('agents.done')}</SealedButton></div>
              </div>
            ) : (
              <form onSubmit={handleCreateSubmit} className="ag-modal__body">
                <div className="ag-field">
                  <label className="ag-label">{t('agents.name')}</label>
                  <input className="ag-input" placeholder="e.g. My Assistant" value={createForm.name} autoFocus
                    onChange={(e) => { const name = e.target.value; const nextSlug = slugFrom(name); setAgentIdTakenFromSubmit(false); setCreateForm((f) => ({ ...f, name, agentId: f.agentId === slugFrom(f.name) || !f.agentId.trim() ? nextSlug : f.agentId })); }} />
                </div>
                <div className="ag-field">
                  <label className="ag-label">{t('agents.agentId')}</label>
                  <input className={`ag-input ${showAgentIdTaken ? 'ag-input--error' : ''}`} placeholder="e.g. my-assistant" value={createForm.agentId}
                    onChange={(e) => { setAgentIdTakenFromSubmit(false); setCreateForm((f) => ({ ...f, agentId: e.target.value.toLowerCase().replace(/\s+/g, '-') })); }} />
                  {showAgentIdTaken ? <p className="ag-field-error">{t('agents.agentIdInUse')}</p>
                    : agentIdInvalid ? <p className="ag-field-warn">{t('agents.agentIdChars')}</p>
                    : <p className="ag-field-hint">{t('agents.agentIdHint')}</p>}
                </div>
                <div className="ag-field">
                  <label className="ag-label">{t('agents.description')}</label>
                  <textarea className="ag-textarea" rows={3} placeholder="e.g. Handles support queries and triages tickets" value={createForm.description}
                    onChange={(e) => setCreateForm((f) => ({ ...f, description: e.target.value }))} />
                </div>
                <div className="ag-modal__footer">
                  <FrameButton type="button" onClick={() => handleCreateClose(false)}>{t('agents.cancel')}</FrameButton>
                  <SealedButton type="submit" disabled={createSubmitting || agentIdInvalid || showAgentIdTaken || !(createForm.name ?? '').trim() || !(createForm.agentId ?? '').trim()}>
                    {createSubmitting ? <><Loader2 className="ag-icon-xs ag-spin" /> {t('agents.creating')}</> : t('agents.create')}
                  </SealedButton>
                </div>
              </form>
            )}
          </div>
        </div>
      )}

      {/* Search + View Toggle */}
      <div className="ag-controls">
        <div className="ag-search">
          <Search className="ag-search__icon" />
          <input className="ag-input ag-search__input" placeholder={t('agents.searchPlaceholder')} value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)} />
        </div>
        <div className="ag-view-toggle">
          <button className={`ag-view-btn ${viewMode === 'grid' ? 'ag-view-btn--active' : ''}`} onClick={() => setViewMode('grid')}><LayoutGrid className="ag-icon-xs" /></button>
          <button className={`ag-view-btn ${viewMode === 'list' ? 'ag-view-btn--active' : ''}`} onClick={() => setViewMode('list')}><List className="ag-icon-xs" /></button>
        </div>
      </div>

      {error && (
        <Chamber className="ag-error-chamber">
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: '1rem' }}>
            <p className="ag-error-text">{error instanceof Error ? error.message : t('agents.failedToLoad')}</p>
            <FrameButton size="sm" onClick={() => refetch()} iconLeft={<RefreshCw className="ag-icon-xs" />}>
              {t('agents.retry') ?? 'Retry'}
            </FrameButton>
          </div>
        </Chamber>
      )}

      {/* Empty State */}
      {filteredAgents.length === 0 ? (
        <Chamber className="ag-empty">
          <Bot className="ag-empty__icon" />
          <h3 className="ag-empty__title">{t('agents.noAgentsFound')}</h3>
          <p className="ag-empty__desc">{searchQuery ? t('agents.adjustSearch') : t('agents.createFirstAgent')}</p>
        </Chamber>
      ) : viewMode === 'grid' ? (
        <div className="ag-grid">
          {filteredAgents.map((agent) => (
            <Card key={agent.id} className="ag-agent-card">
              <div className="ag-agent-card__header">
                <div>
                  <h3 className="ag-agent-card__name">{agent.name ?? '—'}</h3>
                  <p className="ag-agent-card__id">{agent.agentId ?? '—'}</p>
                </div>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <button className="ag-icon-btn" aria-label="Agent options"><MoreVertical className="ag-icon-sm" /></button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="w-44">
                    <DropdownMenuItem onClick={() => navigate(ROUTES.agentPath(agent.agentId))} className="gap-2">
                      <Eye className="h-4 w-4" /> {t('agents.view')}
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={() => navigate(ROUTES.agentEditPath(agent.agentId))} className="gap-2">
                      <Edit3 className="h-4 w-4" /> {t('agents.edit')}
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={() => { navigator.clipboard.writeText(agent.agentId ?? ''); toast.success(t('agents.agentIdCopied')); }} className="gap-2">
                      <Copy className="h-4 w-4" /> {t('agents.copyId')}
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem onClick={() => handleDelete(agent)} className="gap-2 text-destructive focus:text-destructive">
                      <Trash2 className="h-4 w-4" /> {t('agents.delete')}
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
              <div className="ag-agent-card__body">
                <div className="ag-agent-card__badges">
                  <StatusPill status={statusToPill(agent.status)} label={agent.status} />
                  {agent.swarmRole && <span className="ag-role-badge">{agent.swarmRole}</span>}
                </div>
                {agent.description && <p className="ag-agent-card__desc">{agent.description}</p>}
                <div className="ag-agent-card__actions">
                  <FrameButton size="sm" onClick={() => navigate(ROUTES.agentPath(agent.agentId))} iconLeft={<Settings className="ag-icon-xs" />}>{t('agents.manage')}</FrameButton>
                  <button className="ag-delete-btn" onClick={() => handleDelete(agent)} aria-label="Delete agent"><Trash2 className="ag-icon-xs" /></button>
                </div>
              </div>
            </Card>
          ))}
        </div>
      ) : (
        <Chamber className="ag-list-chamber">
          <div className="ag-list">
            {filteredAgents.map((agent) => (
              <div key={agent.id} className="ag-list-item">
                <div className="ag-list-info">
                  <div className="ag-list-name-row">
                    <h4 className="ag-list-name">{agent.name ?? '—'}</h4>
                    <StatusPill status={statusToPill(agent.status)} label={agent.status} />
                    {agent.swarmRole && <span className="ag-role-badge">{agent.swarmRole}</span>}
                  </div>
                  <p className="ag-list-id">{agent.agentId ?? '—'}</p>
                </div>
                <div className="ag-list-actions">
                  <button className="ag-icon-btn" onClick={() => navigate(ROUTES.agentPath(agent.agentId))}><Eye className="ag-icon-sm" /></button>
                  <button className="ag-icon-btn" onClick={() => navigate(ROUTES.agentEditPath(agent.agentId))}><Edit3 className="ag-icon-sm" /></button>
                  <button className="ag-icon-btn ag-icon-btn--danger" onClick={() => handleDelete(agent)}><Trash2 className="ag-icon-sm" /></button>
                </div>
              </div>
            ))}
          </div>
        </Chamber>
      )}
    </div>
  );
}

export default AgentsPage;
