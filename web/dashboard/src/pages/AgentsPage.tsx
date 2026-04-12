import { useState, useEffect, useMemo } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { DataTable } from "@/components/ui/data-table";
import { ToggleButtonGroup } from "@/components/ui";
import type { ColumnDef } from "@tanstack/react-table";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { agentApi, type AgentIdentity } from "@/api/agent";
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
} from "lucide-react";
import { ROUTES } from "@/lib/constants";
import { canCreateAgent, getAgentsLimit, hasFeature } from "@/lib/plan-utils";
import { usePlan } from "@/hooks/usePlan";
import { toast } from "sonner";

export function AgentsPage() {
  const navigate = useNavigate();
  const { plan } = usePlan();
  const [agents, setAgents] = useState<AgentIdentity[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const [createSubmitting, setCreateSubmitting] = useState(false);
  const [createForm, setCreateForm] = useState({ agentId: "", name: "", description: "" });
  const [createdApiKey, setCreatedApiKey] = useState<string | null>(null);
  const [apiKeyCopied, setApiKeyCopied] = useState(false);
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');

  const agentCount = agents.length;
  const canCreate = canCreateAgent(plan, agentCount);
  const agentsUnlocked = hasFeature(plan, "AGENTS");
  const agentsLimit = getAgentsLimit(plan);

  const slugFrom = (s: string) =>
    (s ?? "")
      .trim()
      .toLowerCase()
      .replace(/\s+/g, "-")
      .replace(/[^a-z0-9-]/g, "");
  const agentIdRaw = createForm.agentId ?? "";
  const agentIdInvalid =
    agentIdRaw.trim() !== "" && /[^a-z0-9-]/.test(agentIdRaw.trim().toLowerCase());
  const normalizedAgentId = slugFrom(agentIdRaw);
  const existingAgentIds = agents.map((a) => (a.agentId ?? "").toLowerCase());
  const agentIdTaken =
    normalizedAgentId.length > 0 && existingAgentIds.includes(normalizedAgentId);
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
      console.error("Failed to load agents:", err);
      setError("Failed to load agents. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  const handleCreateSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const agentId = (createForm.agentId ?? "").trim().toLowerCase().replace(/\s+/g, "-");
    const name = (createForm.name ?? "").trim();
    if (!agentId || !name) {
      toast.error("Agent ID and name are required.");
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
      toast.success("Agent created successfully.");
      setCreateForm({ agentId: "", name: "", description: "" });
    } catch (err: unknown) {
      const res = err && typeof err === "object" && "response" in err ? (err as { response?: { status?: number; data?: { error?: { code?: string; message?: string }; message?: string } } }).response : undefined;
      const status = res?.status;
      const data = res?.data;
      const code = typeof data?.error === "object" ? data.error.code : undefined;
      const message = typeof data?.error === "object" ? data.error.message : data?.message ?? (typeof data?.error === "string" ? data.error : null);
      const isTaken = status === 409 || code === "AGENT_ID_TAKEN" || (typeof message === "string" && /already|duplicate|in use/i.test(message));
      if (isTaken) {
        setAgentIdTakenFromSubmit(true);
        toast.error("This agent ID is already in use. Choose a different one.");
      } else {
        toast.error(message || "Failed to create agent. Please try again.");
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
      setCreateForm({ agentId: "", name: "", description: "" });
      loadAgents();
    }
    setCreateOpen(open);
  };

  const copyApiKey = async () => {
    if (!createdApiKey) return;
    try {
      await navigator.clipboard.writeText(createdApiKey);
      setApiKeyCopied(true);
      toast.success("API key copied to clipboard.");
      setTimeout(() => setApiKeyCopied(false), 2000);
    } catch {
      toast.error("Failed to copy.");
    }
  };

  const filteredAgents = agents.filter(
    (agent) =>
      (agent.name ?? "").toLowerCase().includes(searchQuery.toLowerCase()) ||
      (agent.agentId ?? "").toLowerCase().includes(searchQuery.toLowerCase())
  );

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "active":
        return <Badge className="bg-green-500">Active</Badge>;
      case "suspended":
        return <Badge className="bg-red-500">Suspended</Badge>;
      case "pending":
        return <Badge className="bg-yellow-500">Pending</Badge>;
      default:
        return <Badge>{status}</Badge>;
    }
  };

  const getSwarmRoleBadge = (role?: string) => {
    if (!role) return null;
    const colors: Record<string, string> = {
      worker: "bg-blue-500",
      manager: "bg-purple-500",
      infrastructure: "bg-orange-500",
    };
    return <Badge className={colors[role] || "bg-gray-500"}>{role}</Badge>;
  };

  // Define table columns for list view
  const columns = useMemo<ColumnDef<AgentIdentity>[]>(() => [
    {
      accessorKey: 'name',
      header: 'Name',
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
      header: 'Status',
      size: 120,
      cell: ({ row }) => getStatusBadge(row.original.status),
    },
    {
      accessorKey: 'swarmRole',
      header: 'Swarm Role',
      size: 140,
      cell: ({ row }) => getSwarmRoleBadge(row.original.swarmRole) || <span className="text-muted-foreground">-</span>,
    },
    {
      accessorKey: 'createdAt',
      header: 'Created',
      size: 150,
      cell: ({ row }) => {
        const date = new Date(row.original.createdAt);
        return (
          <span className="text-sm text-muted-foreground">
            {date.toLocaleDateString()} {date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
          </span>
        );
      },
    },
    {
      id: 'actions',
      header: 'Actions',
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
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      ),
    },
  ], [navigate]);

  // Handle bulk actions
  const handleBulkAction = (action: string, selectedRows: AgentIdentity[]) => {
    if (action === 'delete') {
      // Implement bulk delete
      toast.info(`Would delete ${selectedRows.length} agents`);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <Loader2 className="h-8 w-8 animate-spin text-brand-500" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-2">
            <Bot className="h-8 w-8" />
            Agents
          </h1>
          <p className="text-muted-foreground mt-1">
            Manage your AI agents and their configurations
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" onClick={() => navigate(ROUTES.SDK_INTEGRATIONS)}>
            <Puzzle className="h-4 w-4 mr-2" />
            SDK Setup
          </Button>
          <Button
            onClick={() => setCreateOpen(true)}
            disabled={!canCreate}
            title={
              !agentsUnlocked
                ? "Agents are available on Starter and higher plans"
                : !canCreate
                  ? `Plan limit reached (${agentCount} of ${agentsLimit >= 10000 ? "∞" : agentsLimit})`
                  : undefined
            }
          >
            <Plus className="h-4 w-4 mr-2" />
            Create Agent
          </Button>
        </div>
      </div>
      {!canCreate && (
        <p className="text-sm text-muted-foreground text-right md:text-left">
          {!agentsUnlocked ? (
            <>
              Upgrade to register agents.{" "}
              <Link to={ROUTES.PRICING} className="text-brand-500 hover:underline">
                View plans
              </Link>
            </>
          ) : (
            <>
              Agent limit reached for your plan.{" "}
              <Link to={ROUTES.PRICING} className="text-brand-500 hover:underline">
                Upgrade
              </Link>
            </>
          )}
        </p>
      )}

      {/* Create Agent modal */}
      <Dialog open={createOpen} onOpenChange={handleCreateClose}>
        <DialogContent className="sm:max-w-xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-brand-500/15 text-brand-500">
                <Bot className="h-4 w-4" />
              </span>
              Create Agent
            </DialogTitle>
            <DialogDescription>
              Register a new AI agent. You’ll get an API key to authenticate requests — save it, it won’t be shown again.
            </DialogDescription>
          </DialogHeader>
          {createdApiKey ? (
            <div className="space-y-4 min-w-0">
              <div className="flex items-center gap-2 rounded-lg border border-green-500/30 bg-green-500/5 px-3 py-2 text-sm text-green-700 dark:text-green-400">
                <Check className="h-5 w-5 shrink-0" />
                <span>Agent created. Copy the API key below — it won’t be shown again.</span>
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
                <Button onClick={() => handleCreateClose(false)}>Done</Button>
              </DialogFooter>
            </div>
          ) : (
            <form onSubmit={handleCreateSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="create-name">Name</Label>
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
                      agentId: f.agentId === slugFrom(f.name) || !f.agentId.trim() ? nextSlug : f.agentId,
                    }));
                  }}
                  autoFocus
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="create-agentId">Agent ID</Label>
                <Input
                  id="create-agentId"
                  placeholder="e.g. my-assistant"
                  value={createForm.agentId}
                  onChange={(e) => {
                    setAgentIdTakenFromSubmit(false);
                    setCreateForm((f) => ({ ...f, agentId: e.target.value.toLowerCase().replace(/\s+/g, "-") }));
                  }}
                  className={showAgentIdTaken ? "font-mono border-red-500 focus-visible:ring-red-500" : "font-mono"}
                />
                {showAgentIdTaken ? (
                  <p className="text-xs text-red-600 dark:text-red-400">
                    This agent ID is already in use. Choose a different one.
                  </p>
                ) : agentIdInvalid ? (
                  <p className="text-xs text-amber-600 dark:text-amber-400">
                    Use only lowercase letters, numbers, and hyphens.
                  </p>
                ) : (
                  <p className="text-xs text-muted-foreground">Used in API calls and URLs. Edit if you want a different slug.</p>
                )}
              </div>
              <div className="space-y-2">
                <Label htmlFor="create-description">Description (optional)</Label>
                <Textarea
                  id="create-description"
                  placeholder="e.g. Handles support queries and triages tickets"
                  value={createForm.description}
                  onChange={(e) => setCreateForm((f) => ({ ...f, description: e.target.value }))}
                  rows={3}
                  className="resize-none"
                />
              </div>
              <DialogFooter className="gap-2 sm:gap-0">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => handleCreateClose(false)}
                >
                  Cancel
                </Button>
                <Button
                  type="submit"
                  disabled={createSubmitting || agentIdInvalid || showAgentIdTaken || !(createForm.name ?? "").trim() || !(createForm.agentId ?? "").trim()}
                >
                  {createSubmitting ? (
                    <>
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      Creating...
                    </>
                  ) : (
                    "Create Agent"
                  )}
                </Button>
              </DialogFooter>
            </form>
          )}
        </DialogContent>
      </Dialog>

      {/* Search and Controls */}
      <div className="flex flex-col sm:flex-row gap-4">
        <Card className="flex-1">
          <CardContent className="pt-6">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search agents..."
                className="pl-10"
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
            { value: 'grid', label: 'Grid', icon: <LayoutGrid className="h-4 w-4" /> },
            { value: 'list', label: 'List', icon: <List className="h-4 w-4" /> },
          ]}
          variant="outline"
          size="sm"
          className="h-fit"
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
          <CardContent className="pt-6 text-center py-12">
            <Bot className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
            <h3 className="text-lg font-medium">No agents found</h3>
            <p className="text-muted-foreground mt-1">
              {searchQuery
                ? "Try adjusting your search query"
                : "Create your first agent to get started"}
            </p>
          </CardContent>
        </Card>
      ) : viewMode === 'grid' ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredAgents.map((agent) => (
            <Card key={agent.id} className="hover:shadow-lg transition-shadow">
              <CardHeader className="flex flex-row items-start justify-between space-y-0 pb-2">
                <div className="space-y-1">
                  <CardTitle className="text-lg">{agent.name ?? "—"}</CardTitle>
                  <CardDescription className="text-xs font-mono">
                    {agent.agentId ?? "—"}
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
                    <Button variant="outline" size="sm" className="flex-1">
                      <Settings className="h-3 w-3 mr-1" />
                      Manage
                    </Button>
                    <Button variant="outline" size="sm" aria-label="Delete agent">
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
            { label: 'Delete Selected', value: 'delete', variant: 'destructive' },
          ]}
          exportFileName={`agents-${new Date().toISOString().split('T')[0]}`}
          emptyState={
            <Card>
              <CardContent className="pt-6 text-center py-12">
                <Bot className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
                <h3 className="text-lg font-medium">No agents found</h3>
                <p className="text-muted-foreground mt-1">
                  {searchQuery
                    ? "Try adjusting your search query"
                    : "Create your first agent to get started"}
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
