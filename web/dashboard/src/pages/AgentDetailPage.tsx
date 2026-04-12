import { agentApi, type AgentIdentity, type BehavioralPolicy } from '@/api/agent';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Textarea } from '@/components/ui/textarea';
import { useAgent, useAgentPolicy, useAgentUsage, useUpdateAgent, useUpdateAgentPolicy } from '@/hooks/useAgent';
import { useAgentMemories } from '@/hooks/useAgentMemory';
import { ROUTES } from '@/lib/constants';
import {
  ArrowLeft,
  Bot,
  Brain,
  Loader2,
  MemoryStick,
  Save,
  Settings,
  Trash2,
  Wallet,
  Zap,
} from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import { AgentMemoryGraph } from './AgentDetailPage/components/AgentMemoryGraph';

function sanitizeAgentIdParam(raw: string | undefined): string | null {
  const t = raw?.trim();
  if (!t || t === 'undefined' || t === 'null') return null;
  return t;
}

export function AgentDetailPage() {
  const { agentId: pathAgentId } = useParams<{ agentId: string }>();
  const navigate = useNavigate();
  const agentId = sanitizeAgentIdParam(pathAgentId);

  const [activeTab, setActiveTab] = useState('overview');
  const [deleting, setDeleting] = useState(false);

  // Form states for settings
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [autonomousEnabled, setAutonomousEnabled] = useState(false);
  const [evolutionEnabled, setEvolutionEnabled] = useState(false);
  const [maxExecutionDepth, setMaxExecutionDepth] = useState(5);
  const [maxRecursionDepth, setMaxRecursionDepth] = useState(3);
  const [maxWallTimeMs, setMaxWallTimeMs] = useState(30000);
  const [maxMemoryGrowthMB, setMaxMemoryGrowthMB] = useState(512);
  const [saving, setSaving] = useState(false);

  const { data: agentData, isLoading: loading, error } = useAgent(agentId ?? '');
  const { data: policyData } = useAgentPolicy(agentId ?? '');
  const { data: usageData } = useAgentUsage(agentId ?? '');
  const { data: memoriesData } = useAgentMemories({ agent_id: agentId ?? undefined }, !!agentId);
  const updateAgent = useUpdateAgent(agentId ?? '');
  const updatePolicy = useUpdateAgentPolicy(agentId ?? '');

  const agent = agentData?.agent ?? null;
  const policy = policyData?.policy;
  const usage = usageData?.usage;
  const memories = memoriesData?.memories ?? [];

  // Initialize form states from agent data
  useEffect(() => {
    if (agent) {
      setName(agent.name || '');
      setDescription(agent.description || '');
      setAutonomousEnabled(agent.autonomousEnabled ?? false);
      setEvolutionEnabled(agent.evolutionEnabled ?? false);
    }
  }, [agent]);

  // Initialize policy form states
  useEffect(() => {
    if (policy) {
      setMaxExecutionDepth(policy.maxExecutionDepth ?? 5);
      setMaxRecursionDepth(policy.maxRecursionDepth ?? 3);
      setMaxWallTimeMs(policy.maxWallTimeMs ?? 30000);
      setMaxMemoryGrowthMB(policy.maxMemoryGrowthMB ?? 512);
    }
  }, [policy]);

  const handleDelete = async () => {
    if (!agentId || !confirm(`Delete agent "${agent?.name ?? agentId}"? This cannot be undone.`)) {
      return;
    }
    setDeleting(true);
    try {
      await agentApi.deleteAgent(agentId);
      toast.success('Agent deleted.');
      navigate(ROUTES.AGENTS);
    } catch {
      toast.error('Failed to delete agent.');
    } finally {
      setDeleting(false);
    }
  };

  const handleSaveSettings = async () => {
    if (!agentId) return;
    setSaving(true);
    try {
      await updateAgent.mutateAsync({
        name,
        description,
      });

      await updatePolicy.mutateAsync({
        agentId,
        maxExecutionDepth,
        maxRecursionDepth,
        maxWallTimeMs,
        maxMemoryGrowthMB,
      });

      toast.success('Settings saved successfully');
    } catch {
      toast.error('Failed to save settings');
    } finally {
      setSaving(false);
    }
  };

  if (!agentId) {
    return (
      <div className="space-y-4">
        <Button variant="ghost" size="sm" asChild>
          <Link to={ROUTES.AGENTS}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to agents
          </Link>
        </Button>
        <p className="text-sm text-muted-foreground">Invalid agent link.</p>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center gap-3 py-16 text-muted-foreground">
        <Loader2 className="h-8 w-8 animate-spin" />
        <p className="text-sm">Loading agent…</p>
      </div>
    );
  }

  if (error || !agent) {
    return (
      <div className="space-y-4">
        <Button variant="ghost" size="sm" asChild>
          <Link to={ROUTES.AGENTS}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to agents
          </Link>
        </Button>
        <Card className="border-destructive/40">
          <CardHeader>
            <CardTitle>Agent not found</CardTitle>
            <CardDescription>
              {error instanceof Error ? error.message : 'Could not load this agent. It may have been deleted or you may not have access.'}
            </CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  const caps = agent.capabilities ? Object.keys(agent.capabilities) : [];
  const enc = encodeURIComponent(agent.agentId);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="space-y-2">
          <Button variant="ghost" size="sm" className="-ml-2 w-fit" asChild>
            <Link to={ROUTES.AGENTS}>
              <ArrowLeft className="h-4 w-4 mr-2" />
              Back to agents
            </Link>
          </Button>
          <div className="flex items-center gap-2">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-brand-500/15 text-brand-500">
              <Bot className="h-5 w-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold">{agent.name || agent.agentId}</h1>
              <p className="font-mono text-xs text-muted-foreground">{agent.agentId}</p>
            </div>
          </div>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button variant="outline" size="sm" asChild>
            <Link to={`/evolution/${enc}`}>
              <Zap className="h-4 w-4 mr-2" />
              Evolution
            </Link>
          </Button>
          <Button variant="outline" size="sm" asChild>
            <Link to={`/wallet/agents/${enc}`}>
              <Wallet className="h-4 w-4 mr-2" />
              Wallet
            </Link>
          </Button>
          <Button variant="outline" size="sm" asChild>
            <Link to={ROUTES.SDK_INTEGRATIONS}>SDK setup</Link>
          </Button>
          <Button
            variant="outline"
            size="sm"
            className="text-destructive hover:text-destructive gap-2"
            onClick={() => void handleDelete()}
            disabled={deleting}
          >
            {deleting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Trash2 className="h-4 w-4" />}
            Delete
          </Button>
        </div>
      </div>

      {/* Status Badges */}
      <div className="flex flex-wrap gap-2">
        <Badge className={agent.status === 'active' ? 'bg-green-600' : ''}>{agent.status}</Badge>
        {agent.swarmRole && <Badge variant="secondary">{agent.swarmRole}</Badge>}
        {agent.parentAgentId && (
          <Badge variant="outline" className="text-amber-600 border-amber-500/40">
            Child of{' '}
            <Link
              className="ml-1 underline font-mono text-xs"
              to={`${ROUTES.AGENTS}/${encodeURIComponent(agent.parentAgentId)}`}
            >
              {agent.parentAgentId}
            </Link>
          </Badge>
        )}
        {agent.autonomousEnabled && <Badge variant="outline">Autonomous</Badge>}
        {agent.evolutionEnabled && <Badge variant="outline">Evolution</Badge>}
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid w-full max-w-md grid-cols-3">
          <TabsTrigger value="overview">
            <Brain className="h-4 w-4 mr-2" />
            Overview
          </TabsTrigger>
          <TabsTrigger value="settings">
            <Settings className="h-4 w-4 mr-2" />
            Settings
          </TabsTrigger>
          <TabsTrigger value="memory">
            <MemoryStick className="h-4 w-4 mr-2" />
            Memory Graph
          </TabsTrigger>
        </TabsList>

        {/* Overview Tab */}
        <TabsContent value="overview" className="space-y-6">
          <div className="grid gap-6 md:grid-cols-2">
            {/* Agent Info Card */}
            <Card>
              <CardHeader>
                <CardTitle>Overview</CardTitle>
                <CardDescription>Identity and capabilities for this agent.</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {agent.description ? (
                  <p className="text-sm text-muted-foreground">{agent.description}</p>
                ) : (
                  <p className="text-sm text-muted-foreground italic">No description.</p>
                )}
                {caps.length > 0 && (
                  <div>
                    <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-2">
                      Capabilities
                    </p>
                    <div className="flex flex-wrap gap-1.5">
                      {caps.map((c) => (
                        <Badge key={c} variant="secondary" className="text-xs">
                          {c.replace(/_/g, ' ')}
                        </Badge>
                      ))}
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Usage Card */}
            <Card>
              <CardHeader>
                <CardTitle>Current Usage</CardTitle>
                <CardDescription>Real-time agent metrics.</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {usage ? (
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-1">
                      <p className="text-xs text-muted-foreground">Calls This Minute</p>
                      <p className="text-lg font-semibold">{usage.callsThisMinute}</p>
                    </div>
                    <div className="space-y-1">
                      <p className="text-xs text-muted-foreground">Concurrent Executions</p>
                      <p className="text-lg font-semibold">{usage.concurrentExecutions}</p>
                    </div>
                    <div className="space-y-1">
                      <p className="text-xs text-muted-foreground">Memory Usage</p>
                      <p className="text-lg font-semibold">{usage.memoryUsageMB} MB</p>
                    </div>
                    <div className="space-y-1">
                      <p className="text-xs text-muted-foreground">Avg Execution Time</p>
                      <p className="text-lg font-semibold">{usage.executionTimeMs} ms</p>
                    </div>
                    <div className="col-span-2 space-y-1">
                      <p className="text-xs text-muted-foreground">Spend Today</p>
                      <p className="text-lg font-semibold">${usage.spendToday.toFixed(4)}</p>
                    </div>
                  </div>
                ) : (
                  <p className="text-sm text-muted-foreground italic">No usage data available.</p>
                )}
              </CardContent>
            </Card>
          </div>

          {/* Policy Card */}
          {policy && (
            <Card>
              <CardHeader>
                <CardTitle>Behavioral Policy</CardTitle>
                <CardDescription>Safety limits and constraints.</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
                  <div className="space-y-1">
                    <p className="text-xs text-muted-foreground">Max Execution Depth</p>
                    <p className="text-lg font-semibold">{policy.maxExecutionDepth}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-xs text-muted-foreground">Max Recursion Depth</p>
                    <p className="text-lg font-semibold">{policy.maxRecursionDepth}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-xs text-muted-foreground">Max Wall Time</p>
                    <p className="text-lg font-semibold">{policy.maxWallTimeMs}ms</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-xs text-muted-foreground">Max Memory Growth</p>
                    <p className="text-lg font-semibold">{policy.maxMemoryGrowthMB}MB</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        {/* Settings Tab */}
        <TabsContent value="settings" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Agent Settings</CardTitle>
              <CardDescription>Configure agent identity and behavior.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              {/* Basic Info */}
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="name">Name</Label>
                  <Input
                    id="name"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="Enter agent name"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="description">Description</Label>
                  <Textarea
                    id="description"
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    placeholder="Enter agent description"
                    rows={3}
                  />
                </div>
              </div>

              {/* Toggles */}
              <div className="flex flex-wrap gap-6">
                <div className="flex items-center space-x-2">
                  <Switch
                    id="autonomous"
                    checked={autonomousEnabled}
                    onCheckedChange={setAutonomousEnabled}
                  />
                  <Label htmlFor="autonomous">Autonomous Mode</Label>
                </div>
                <div className="flex items-center space-x-2">
                  <Switch
                    id="evolution"
                    checked={evolutionEnabled}
                    onCheckedChange={setEvolutionEnabled}
                  />
                  <Label htmlFor="evolution">Evolution Enabled</Label>
                </div>
              </div>

              {/* Policy Settings */}
              <div className="space-y-4 border-t pt-4">
                <h4 className="text-sm font-medium">Behavioral Limits</h4>
                <div className="grid gap-4 sm:grid-cols-2">
                  <div className="space-y-2">
                    <Label htmlFor="maxExecutionDepth">Max Execution Depth</Label>
                    <Input
                      id="maxExecutionDepth"
                      type="number"
                      value={maxExecutionDepth}
                      onChange={(e) => setMaxExecutionDepth(parseInt(e.target.value) || 0)}
                      min={1}
                      max={20}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="maxRecursionDepth">Max Recursion Depth</Label>
                    <Input
                      id="maxRecursionDepth"
                      type="number"
                      value={maxRecursionDepth}
                      onChange={(e) => setMaxRecursionDepth(parseInt(e.target.value) || 0)}
                      min={0}
                      max={10}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="maxWallTimeMs">Max Wall Time (ms)</Label>
                    <Input
                      id="maxWallTimeMs"
                      type="number"
                      value={maxWallTimeMs}
                      onChange={(e) => setMaxWallTimeMs(parseInt(e.target.value) || 0)}
                      min={1000}
                      step={1000}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="maxMemoryGrowthMB">Max Memory Growth (MB)</Label>
                    <Input
                      id="maxMemoryGrowthMB"
                      type="number"
                      value={maxMemoryGrowthMB}
                      onChange={(e) => setMaxMemoryGrowthMB(parseInt(e.target.value) || 0)}
                      min={64}
                      step={64}
                    />
                  </div>
                </div>
              </div>

              <Button onClick={handleSaveSettings} disabled={saving}>
                {saving ? (
                  <>
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    Saving...
                  </>
                ) : (
                  <>
                    <Save className="h-4 w-4 mr-2" />
                    Save Settings
                  </>
                )}
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Memory Graph Tab */}
        <TabsContent value="memory" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Memory Visualization</CardTitle>
              <CardDescription>Graph view of agent memory relationships and access patterns.</CardDescription>
            </CardHeader>
            <CardContent>
              {memories.length > 0 ? (
                <AgentMemoryGraph memories={memories} agentId={agentId} />
              ) : (
                <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
                  <Brain className="h-12 w-12 mb-4 opacity-50" />
                  <p>No memories found for this agent.</p>
                  <p className="text-sm">Memories will appear here when the agent stores information.</p>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

export default AgentDetailPage;
