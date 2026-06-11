import { useState, useEffect, useMemo } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { usePageTitle } from '@/hooks';
import './AgentsPage/styles.css';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { DataTable } from '@/components/ui/data-table';
import { ToggleButtonGroup } from '@/components/ui';
import type { ColumnDef } from '@tanstack/react-table';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { agentApi, type AgentIdentity } from '@/api/agent';
import {
  Bot,
  Plus,
  Puzzle,
  Search,
  Settings,
  Trash2,
  MoreVertical,
  Loader2,
  Copy,
  Check,
  LayoutGrid,
  List,
  Edit3,
  Eye,
} from 'lucide-react';
import { ROUTES } from '@/lib/constants';
import { canCreateAgent, getAgentsLimit, hasFeature } from '@/lib/plan-utils';
import { usePlan } from '@/hooks/usePlan';
import { toast } from 'sonner';

export function AgentsPage() {
  usePageTitle('Agents');
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { username } = useParams();
  const { plan } = usePlan();
  const [agents, setAgents] = useState<AgentIdentity[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
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

  const slugFrom = (s: string) =>
    (s ?? '')
      .trim()
      .toLowerCase()
      .replace(/\s+/g, '-')
      .replace(/[^a-z0-9-]/g, '');
  const agentIdRaw = createForm.agentId ?? '';
  const agentIdInvalid =
    agentIdRaw.trim() !== '' && /[^a-z0-9-]/.test(agentIdRaw.trim().toLowerCase());
  const normalizedAgentId = slugFrom(agentIdRaw);
  const existingAgentIds = agents.map((a) => (a.agentId ?? '').toLowerCase());
  const agentIdTaken = normalizedAgentId.length > 0 && existingAgentIds.includes(normalizedAgentId);
  const [agentIdTakenFromSubmit, setAgentIdTakenFromSubmit] = useState(false);
  const showAgentIdTaken = agentIdTaken || agentIdTakenFromSubmit;

  useEffect(() => {
    loadAgents();
  }, []);

  const loadAgents = async () => {
    try {
      setLoading(true);
      const response = await agentApi.listAgents({ limit: 100 });
      setAgents(response.agents);
    } catch (err) {
      console.error('Failed to load agents:', err);
      setError(t('agents.failedToLoad'));
    } finally {
      setLoading(false);
    }
  };

  const handleCreateSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const agentId = (createForm.agentId ?? '').trim().toLowerCase().replace(/\s+/g, '-');
    const name = (createForm.name ?? '').trim();
    if (!agentId || !name) {
      toast.error(t('agents.agentIdNameRequired'));
      return;
    }
    setCreateSubmitting(true);
    try {
      const res = await agentApi.registerAgent({
        agentId,
        name,
        description: createForm.description.trim() || undefined,
      });
      setCreatedApiKey(res.api_key);
      toast.success(t('agents.agentCreatedSuccess'));
      setCreateForm({ agentId: '', name: '', description: '' });
    } catch (err: unknown) {
      const res =
        err && typeof err === 'object' && 'response' in err
          ? (
              err as {
                response?: {
                  status?: number;
                  data?: { error?: { code?: string; message?: string }; message?: string };
                };
              }
            ).response
          : undefined;
      const status = res?.status;
      const data = res?.data;
      const code = typeof data?.error === 'object' ? data.error.code : undefined;
      const message =
        typeof data?.error === 'object'
          ? data.error.message
          : (data?.message ?? (typeof data?.error === 'string' ? data.error : null));
      const isTaken =
        status === 409 ||
        code === 'AGENT_ID_TAKEN' ||
        (typeof message === 'string' && /already|duplicate|in use/i.test(message));
      if (isTaken) {
        setAgentIdTakenFromSubmit(true);
        toast.error(t('agents.agentIdInUse'));
      } else {
        toast.error(message || t('agents.failedToCreate'));
      }
    } finally {
      setCreateSubmitting(false);
    }
  };

  const handleCreateClose = (open: boolean) => {
    if (!open) {
      setCreatedApiKey(null);
      setApiKeyCopied(false);
      setAgentIdTakenFromSubmit(false);
      setCreateForm({ agentId: '', name: '', description: '' });
      loadAgents();
    }
    setCreateOpen(open);
  };

  const copyApiKey = async () => {
    if (!createdApiKey) return;
    try {
      await navigator.clipboard.writeText(createdApiKey);
      setApiKeyCopied(true);
      toast.success(t('agents.apiKeyCopied'));
      setTimeout(() => setApiKeyCopied(false), 2000);
    } catch {
      toast.error(t('agents.failedToCopy'));
    }
  };

  const filteredAgents = agents.filter(
    (agent) =>
      (agent.name ?? '').toLowerCase().includes(searchQuery.toLowerCase()) ||
      (agent.agentId ?? '').toLowerCase().includes(searchQuery.toLowerCase())
  );

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
        return <Badge className="status-badge status-badge-active">{t('agents.active')}</Badge>;
      case 'suspended':
        return (
          <Badge className="status-badge status-badge-suspended">{t('agents.suspended')}</Badge>
        );
      case 'pending':
        return <Badge className="status-badge status-badge-pending">{t('agents.pending')}</Badge>;
      default:
        return <Badge className="status-badge">{status}</Badge>;
    }
  };

  const getSwarmRoleBadge = (role?: string) => {
    if (!role) return null;
    const roleClass: Record<string, string> = {
      worker: 'swarm-role-badge swarm-role-worker',
      manager: 'swarm-role-badge swarm-role-manager',
      infrastructure: 'swarm-role-badge swarm-role-infrastructure',
    };
    return <Badge className={roleClass[role] || 'swarm-role-badge'}>{role}</Badge>;
  };

  // Define table columns for list view
  const columns = useMemo<ColumnDef<AgentIdentity>[]>(
    () => [
      {
        accessorKey: 'name',
        header: t('agents.name'),
        size: 200,
        cell: ({ row }) => (
          <div className="flex flex-col">
            <span className="font-medium">{row.original.name}</span>
            <span className="text-xs text-muted-foreground font-mono">{row.original.agentId}</span>
          </div>
        ),
      },
      {
        accessorKey: 'status',
        header: t('agents.status'),
        size: 120,
        cell: ({ row }) => getStatusBadge(row.original.status),
      },
      {
        accessorKey: 'swarmRole',
        header: t('agents.swarmRole'),
        size: 140,
        cell: ({ row }) =>
          getSwarmRoleBadge(row.original.swarmRole) || (
            <span className="text-muted-foreground">-</span>
          ),
      },
      {
        accessorKey: 'createdAt',
        header: t('agents.created'),
        size: 150,
        cell: ({ row }) => {
          const date = new Date(row.original.createdAt);
          return (
            <span className="text-sm text-muted-foreground">
              {date.toLocaleDateString()}{' '}
              {date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
            </span>
          );
        },
      },
      {
        id: 'actions',
        header: t('agents.actions'),
        size: 150,
        enableSorting: false,
        cell: ({ row }) => (
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="icon"
              onClick={() => navigate(`/agents/${row.original.id}`)}
              className="h-8 w-8"
            >
              <Eye className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => navigate(`/agents/${row.original.id}/edit`)}
              className="h-8 w-8"
            >
              <Edit3 className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8 text-red-500 hover:text-red-600"
              onClick={() => handleDelete(row.original)}
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        ),
      },
    ],
    [navigate]
  );

  const handleDelete = async (agent: AgentIdentity) => {
    if (!confirm(t('agents.confirmDelete', { name: agent.name ?? agent.agentId }))) {
      return;
    }
    try {
      await agentApi.deleteAgent(agent.id);
      toast.success(t('agents.agentDeleted'));
      loadAgents();
    } catch {
      toast.error(t('agents.failedToDelete'));
    }
  };

  const handleBulkAction = async (action: string, selectedRows: AgentIdentity[]) => {
    if (action === 'delete' && selectedRows.length > 0) {
      if (!confirm(t('agents.confirmBulkDelete', { count: selectedRows.length }))) {
        return;
      }
      try {
        await Promise.all(selectedRows.map((row) => agentApi.deleteAgent(row.id)));
        toast.success(t('agents.agentsDeleted', { count: selectedRows.length }));
        loadAgents();
      } catch {
        toast.error(t('agents.failedToDelete'));
      }
    }
  };

  if (loading) {
    return (
      <div className="agent-loading-container">
        <Loader2 className="agent-loading-spinner" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="agent-header">
        <div>
          <h1 className="agent-title flex items-center gap-2">
            <Bot className="agent-avatar-icon h-8 w-8" />
            {username ? `${username}'s agents` : t('agents.title')}
          </h1>
          <p className="agent-subtitle">{t('agents.manageDescription')}</p>
        </div>
        <div className="agent-header-actions">
          <Button variant="outline" onClick={() => navigate(ROUTES.SDK_INTEGRATIONS)}>
            <Puzzle className="h-4 w-4 mr-2" />
            {t('agents.sdkSetup')}
          </Button>
          <Button
            onClick={() => navigate(ROUTES.AGENT_NEW)}
            disabled={!canCreate}
            title={
              !agentsUnlocked
                ? t('agents.agentsOnStarter')
                : !canCreate
                  ? t('agents.planLimitReached', {
                      count: agentCount,
                      limit: agentsLimit >= 10000 ? '∞' : agentsLimit,
                    })
                  : undefined
            }
          >
            <Plus className="h-4 w-4 mr-2" />
            {t('agents.createAgent')}
          </Button>
        </div>
      </div>
      {!canCreate && (
        <p className="text-sm text-muted-foreground text-right md:text-left">
          {!agentsUnlocked ? (
            <>
              {t('agents.upgradeToRegister')}{' '}
              <Link to={ROUTES.PRICING} className="text-brand-500 hover:underline">
                {t('agents.viewPlans')}
              </Link>
            </>
          ) : (
            <>
              {t('agents.agentLimitReached')}{' '}
              <Link to={ROUTES.PRICING} className="text-brand-500 hover:underline">
                {t('agents.upgrade')}
              </Link>
            </>
          )}
        </p>
      )}

      {/* Create Agent modal */}
      <Dialog open={createOpen} onOpenChange={handleCreateClose}>
        <DialogContent className="agent-dialog-content sm:max-w-xl max-h-[90vh] overflow-y-auto">
          <DialogHeader className="agent-dialog-header">
            <DialogTitle className="agent-dialog-title flex items-center gap-2">
              <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-brand-500/15 text-brand-500">
                <Bot className="h-4 w-4" />
              </span>
              {t('agents.createAgent')}
            </DialogTitle>
            <DialogDescription className="agent-dialog-description">
              {t('agents.registerNewAgent')}
            </DialogDescription>
          </DialogHeader>
          {createdApiKey ? (
            <div className="space-y-4 min-w-0">
              <div className="flex items-center gap-2 rounded-lg border border-green-500/30 bg-green-500/5 px-3 py-2 text-sm text-green-700 dark:text-green-400">
                <Check className="h-5 w-5 shrink-0" />
                <span>{t('agents.agentCreatedCopyKey')}</span>
              </div>
              <div className="flex items-center gap-2 min-w-0 rounded-lg border bg-muted/50 p-3 font-mono text-sm overflow-hidden">
                <code className="min-w-0 flex-1 truncate break-all">{createdApiKey}</code>
                <Button
                  type="button"
                  variant="secondary"
                  size="sm"
                  onClick={copyApiKey}
                  className="shrink-0"
                >
                  {apiKeyCopied ? (
                    <>
                      <Check className="h-4 w-4 mr-1.5 text-green-500" />
                      Copied
                    </>
                  ) : (
                    <>
                      <Copy className="h-4 w-4 mr-1.5" />
                      Copy
                    </>
                  )}
                </Button>
              </div>
              <DialogFooter className="mt-2">
                <Button onClick={() => handleCreateClose(false)}>{t('agents.done')}</Button>
              </DialogFooter>
            </div>
          ) : (
            <form onSubmit={handleCreateSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="create-name">{t('agents.name')}</Label>
                <Input
                  id="create-name"
                  placeholder="e.g. My Assistant"
                  value={createForm.name}
                  onChange={(e) => {
                    const name = e.target.value;
                    const nextSlug = slugFrom(name);
                    setAgentIdTakenFromSubmit(false);
                    setCreateForm((f) => ({
                      ...f,
                      name,
                      agentId:
                        f.agentId === slugFrom(f.name) || !f.agentId.trim() ? nextSlug : f.agentId,
                    }));
                  }}
                  className="agent-input"
                  autoFocus
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="create-agentId">{t('agents.agentId')}</Label>
                <Input
                  id="create-agentId"
                  placeholder="e.g. my-assistant"
                  value={createForm.agentId}
                  onChange={(e) => {
                    setAgentIdTakenFromSubmit(false);
                    setCreateForm((f) => ({
                      ...f,
                      agentId: e.target.value.toLowerCase().replace(/\s+/g, '-'),
                    }));
                  }}
                  className={
                    showAgentIdTaken
                      ? 'agent-input border-red-500 focus-visible:ring-red-500'
                      : 'agent-input'
                  }
                />
                {showAgentIdTaken ? (
                  <p className="text-xs text-red-600 dark:text-red-400">
                    {t('agents.agentIdInUse')}
                  </p>
                ) : agentIdInvalid ? (
                  <p className="text-xs text-amber-600 dark:text-amber-400">
                    {t('agents.agentIdChars')}
                  </p>
                ) : (
                  <p className="text-xs text-muted-foreground">{t('agents.agentIdHint')}</p>
                )}
              </div>
              <div className="space-y-2">
                <Label htmlFor="create-description">{t('agents.description')}</Label>
                <Textarea
                  id="create-description"
                  placeholder="e.g. Handles support queries and triages tickets"
                  value={createForm.description}
                  onChange={(e) => setCreateForm((f) => ({ ...f, description: e.target.value }))}
                  rows={3}
                  className="agent-textarea"
                />
              </div>
              <DialogFooter className="agent-dialog-footer gap-2 sm:gap-0">
                <Button type="button" variant="outline" onClick={() => handleCreateClose(false)}>
                  {t('agents.cancel')}
                </Button>
                <Button
                  type="submit"
                  disabled={
                    createSubmitting ||
                    agentIdInvalid ||
                    showAgentIdTaken ||
                    !(createForm.name ?? '').trim() ||
                    !(createForm.agentId ?? '').trim()
                  }
                >
                  {createSubmitting ? (
                    <>
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      {t('agents.creating')}
                    </>
                  ) : (
                    t('agents.create')
                  )}
                </Button>
              </DialogFooter>
            </form>
          )}
        </DialogContent>
      </Dialog>

      {/* Search and Controls */}
      <div className="flex flex-col sm:flex-row gap-4">
        <Card className="agent-search-card flex-1">
          <CardContent className="pt-6">
            <div className="relative">
              <Search className="agent-search-icon absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4" />
              <Input
                placeholder={t('agents.searchPlaceholder')}
                className="agent-search-input pl-10"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
            </div>
          </CardContent>
        </Card>
        <ToggleButtonGroup
          value={viewMode}
          onValueChange={(v) => setViewMode(v as 'grid' | 'list')}
          options={[
            { value: 'grid', label: t('agents.grid'), icon: <LayoutGrid className="h-4 w-4" /> },
            { value: 'list', label: t('agents.list'), icon: <List className="h-4 w-4" /> },
          ]}
          variant="outline"
          size="sm"
          className="agent-view-toggle h-fit"
        />
      </div>

      {error && (
        <Card className="border-red-500">
          <CardContent className="pt-6 text-red-500">{error}</CardContent>
        </Card>
      )}

      {/* Agents Display - Grid or List */}
      {filteredAgents.length === 0 ? (
        <Card>
          <CardContent className="agent-empty-state">
            <Bot className="agent-empty-icon h-12 w-12 mx-auto mb-4" />
            <h3 className="agent-empty-title">{t('agents.noAgentsFound')}</h3>
            <p className="agent-empty-description">
              {searchQuery ? t('agents.adjustSearch') : t('agents.createFirstAgent')}
            </p>
          </CardContent>
        </Card>
      ) : viewMode === 'grid' ? (
        <div className="agent-grid agent-grid-cols-1 md:agent-grid-md-2 lg:agent-grid-lg-3 gap-4">
          {filteredAgents.map((agent) => (
            <Card key={agent.id} className="agent-card agent-card-hoverable">
              <CardHeader className="flex flex-row items-start justify-between space-y-0 pb-2">
                <div className="space-y-1">
                  <CardTitle className="text-lg">{agent.name ?? '—'}</CardTitle>
                  <CardDescription className="text-xs font-mono">
                    {agent.agentId ?? '—'}
                  </CardDescription>
                </div>
                <Button variant="ghost" size="icon" aria-label="Agent options">
                  <MoreVertical className="h-4 w-4" />
                </Button>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="flex items-center gap-2">
                    {getStatusBadge(agent.status)}
                    {getSwarmRoleBadge(agent.swarmRole)}
                  </div>
                  {agent.description && (
                    <p className="text-sm text-muted-foreground line-clamp-2">
                      {agent.description}
                    </p>
                  )}
                  <div className="flex items-center gap-2 pt-2">
                    <Button
                      variant="outline"
                      size="sm"
                      className="flex-1"
                      onClick={() => navigate(`/agents/${agent.id}`)}
                    >
                      <Settings className="h-3 w-3 mr-1" />
                      {t('agents.manage')}
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      aria-label="Delete agent"
                      onClick={() => handleDelete(agent)}
                    >
                      <Trash2 className="h-3 w-3 text-red-500" />
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      ) : (
        <DataTable
          data={filteredAgents}
          columns={columns}
          enableRowSelection={true}
          enableColumnResize={true}
          enableColumnVisibility={true}
          enableExport={true}
          enableGlobalFilter={true}
          enableColumnFilters={true}
          onBulkAction={handleBulkAction}
          bulkActions={[
            { label: t('agents.deleteSelected'), value: 'delete', variant: 'destructive' },
          ]}
          exportFileName={`agents-${new Date().toISOString().split('T')[0]}`}
          emptyState={
            <Card>
              <CardContent className="agent-empty-state">
                <Bot className="agent-empty-icon h-12 w-12 mx-auto mb-4" />
                <h3 className="agent-empty-title">{t('agents.noAgentsFound')}</h3>
                <p className="agent-empty-description">
                  {searchQuery ? t('agents.adjustSearch') : t('agents.createFirstAgent')}
                </p>
              </CardContent>
            </Card>
          }
        />
      )}
    </div>
  );
}

export default AgentsPage;
