import { agentApi, type AgentIdentity } from '@/api/agent';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { ROUTES } from '@/lib/constants';
import { ArrowLeft, Bot, Loader2, Trash2, Wallet, Zap } from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';

function sanitizeAgentIdParam(raw: string | undefined): string | null {
  const t = raw?.trim();
  if (!t || t === 'undefined' || t === 'null') return null;
  return t;
}

export function AgentDetailPage() {
  const { agentId: pathAgentId } = useParams<{ agentId: string }>();
  const navigate = useNavigate();
  const agentId = sanitizeAgentIdParam(pathAgentId);

  const [agent, setAgent] = useState<AgentIdentity | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [deleting, setDeleting] = useState(false);

  const load = useCallback(async () => {
    if (!agentId) return;
    setLoading(true);
    setError(null);
    try {
      const res = await agentApi.getAgent(agentId);
      setAgent(res.agent);
    } catch {
      setError('Could not load this agent. It may have been deleted or you may not have access.');
      setAgent(null);
    } finally {
      setLoading(false);
    }
  }, [agentId]);

  useEffect(() => {
    if (!agentId) {
      setLoading(false);
      setError('Missing agent ID.');
      return;
    }
    void load();
  }, [agentId, load]);

  const handleDelete = async () => {
    if (!agentId || !confirm(`Delete agent “${agent?.name ?? agentId}”? This cannot be undone.`)) {
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
            <CardDescription>{error}</CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  const caps = agent.capabilities ? Object.keys(agent.capabilities) : [];
  const enc = encodeURIComponent(agent.agentId);

  return (
    <div className="space-y-6">
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
            <Link to={`/wallet/${enc}`}>
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
    </div>
  );
}

export default AgentDetailPage;
