import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
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
  Search,
  Settings,
  Trash2,
  MoreVertical,
  Loader2,
  Copy,
  Check,
} from "lucide-react";
import { toast } from "sonner";

export function AgentsPage() {
  const [agents, setAgents] = useState<AgentIdentity[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const [createSubmitting, setCreateSubmitting] = useState(false);
  const [createForm, setCreateForm] = useState({ agentId: "", name: "", description: "" });
  const [createdApiKey, setCreatedApiKey] = useState<string | null>(null);
  const [apiKeyCopied, setApiKeyCopied] = useState(false);

  const slugFrom = (s: string) =>
    s
      .trim()
      .toLowerCase()
      .replace(/\s+/g, "-")
      .replace(/[^a-z0-9-]/g, "");
  const agentIdInvalid =
    createForm.agentId.trim() !== "" && /[^a-z0-9-]/.test(createForm.agentId.trim().toLowerCase());
  const normalizedAgentId = slugFrom(createForm.agentId);
  const existingAgentIds = agents.map((a) => a.agentId.toLowerCase());
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
    const agentId = createForm.agentId.trim().toLowerCase().replace(/\s+/g, "-");
    const name = createForm.name.trim();
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
      agent.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      agent.agentId.toLowerCase().includes(searchQuery.toLowerCase())
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
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
          Create Agent
        </Button>
      </div>

      {/* Create Agent modal */}
      <Dialog open={createOpen} onOpenChange={handleCreateClose}>
        <DialogContent className="sm:max-w-md">
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
            <div className="space-y-4">
              <div className="flex items-center gap-2 rounded-lg border border-green-500/30 bg-green-500/5 px-3 py-2 text-sm text-green-700 dark:text-green-400">
                <Check className="h-5 w-5 shrink-0" />
                <span>Agent created. Copy the API key below — it won’t be shown again.</span>
              </div>
              <div className="flex items-center gap-2 rounded-lg border bg-muted/50 p-3 font-mono text-sm">
                <code className="min-w-0 flex-1 truncate">{createdApiKey}</code>
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
              <DialogFooter>
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
                  disabled={createSubmitting || agentIdInvalid || showAgentIdTaken || !createForm.name.trim() || !createForm.agentId.trim()}
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

      {/* Search */}
      <Card>
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

      {error && (
        <Card className="border-red-500">
          <CardContent className="pt-6 text-red-500">{error}</CardContent>
        </Card>
      )}

      {/* Agents Grid */}
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
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredAgents.map((agent) => (
            <Card key={agent.id} className="hover:shadow-lg transition-shadow">
              <CardHeader className="flex flex-row items-start justify-between space-y-0 pb-2">
                <div className="space-y-1">
                  <CardTitle className="text-lg">{agent.name}</CardTitle>
                  <CardDescription className="text-xs font-mono">
                    {agent.agentId}
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
      )}
    </div>
  );
}

export default AgentsPage;
