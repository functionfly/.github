'use client';

import { useEffect, useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAgent, useAgentPolicy, useUpdateAgent, useUpdateAgentPolicy } from '@/hooks/useAgent';
import { agentApi, type BehavioralPolicy } from '@/api/agent';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Textarea } from '@/components/ui/textarea';
import { ArrowLeft, Bot, Loader2, Save } from 'lucide-react';
import { ROUTES } from '@/lib/constants';
import { toast } from 'sonner';

function sanitizeId(raw: string | undefined): string | null {
  const t = raw?.trim();
  if (!t || t === 'undefined' || t === 'null') return null;
  return t;
}

export function AgentEditPage() {
  const { t } = useTranslation();
  const { id: agentId } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const id = sanitizeId(agentId);

  const { data, isLoading, error } = useAgent(id ?? '');
  const { data: policyData } = useAgentPolicy(id ?? '');
  const updateAgent = useUpdateAgent(id ?? '');
  const updatePolicy = useUpdateAgentPolicy(id ?? '');

  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [autonomousEnabled, setAutonomousEnabled] = useState(false);
  const [evolutionEnabled, setEvolutionEnabled] = useState(false);
  const [maxExecutionDepth, setMaxExecutionDepth] = useState(5);
  const [maxRecursionDepth, setMaxRecursionDepth] = useState(3);
  const [maxWallTimeMs, setMaxWallTimeMs] = useState(30000);
  const [maxMemoryGrowthMB, setMaxMemoryGrowthMB] = useState(512);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (data?.agent) {
      setName(data.agent.name || '');
      setDescription(data.agent.description || '');
      setAutonomousEnabled(data.agent.autonomousEnabled ?? false);
      setEvolutionEnabled(data.agent.evolutionEnabled ?? false);
    }
  }, [data]);

  useEffect(() => {
    if (policyData?.policy) {
      setMaxExecutionDepth(policyData.policy.maxExecutionDepth ?? 5);
      setMaxRecursionDepth(policyData.policy.maxRecursionDepth ?? 3);
      setMaxWallTimeMs(policyData.policy.maxWallTimeMs ?? 30000);
      setMaxMemoryGrowthMB(policyData.policy.maxMemoryGrowthMB ?? 512);
    }
  }, [policyData?.policy]);

  const handleSave = async () => {
    if (!id) return;
    setSaving(true);
    try {
      await updateAgent.mutateAsync({ name, description });
      await updatePolicy.mutateAsync({
        agentId: id,
        maxExecutionDepth,
        maxRecursionDepth,
        maxWallTimeMs,
        maxMemoryGrowthMB,
      });
      toast.success(t('agentDetail.settingsSaved'));
      navigate(ROUTES.AGENT_DETAIL.replace(':id', id));
    } catch {
      toast.error(t('agentDetail.failedToSave'));
    } finally {
      setSaving(false);
    }
  };

  if (!id) {
    return (
      <div className="space-y-4">
        <Button variant="ghost" size="sm" asChild>
          <Link to={ROUTES.AGENT_LIST}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Agents
          </Link>
        </Button>
        <p className="text-sm text-muted-foreground">Invalid agent ID</p>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-8">
        <Loader2 className="h-8 w-8 animate-spin text-brand-500" />
      </div>
    );
  }

  if (error || !data?.agent) {
    return (
      <div className="space-y-4">
        <Button variant="ghost" size="sm" asChild>
          <Link to={ROUTES.AGENT_LIST}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Agents
          </Link>
        </Button>
        <Card className="border-destructive/40">
          <CardHeader>
            <CardTitle>{t('agentDetail.agentNotFound')}</CardTitle>
            <CardDescription>
              {error instanceof Error ? error.message : t('agentDetail.agentNotFoundDescription')}
            </CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  const agent = data.agent;

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <Button variant="ghost" size="sm" asChild>
        <Link to={ROUTES.AGENT_DETAIL.replace(':id', id)}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Agent
        </Link>
      </Button>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-lg bg-gradient-to-br from-brand-500 to-purple-500 flex items-center justify-center">
              <Bot className="h-5 w-5 text-white" />
            </div>
            <div>
              <CardTitle>Edit Agent: {agent.name || agent.agentId}</CardTitle>
              <CardDescription>Configure identity and behavioral settings for this agent.</CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-4">
            <h4 className="text-sm font-medium border-b pb-2">Identity</h4>
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input id="name" value={name} onChange={(e) => setName(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Textarea id="description" value={description} onChange={(e) => setDescription(e.target.value)} rows={3} />
            </div>
          </div>

          <div className="space-y-4">
            <h4 className="text-sm font-medium border-b pb-2">Capabilities</h4>
            <div className="flex flex-wrap gap-6">
              <div className="flex items-center space-x-2">
                <Switch id="autonomous" checked={autonomousEnabled} onCheckedChange={setAutonomousEnabled} />
                <Label htmlFor="autonomous">Autonomous Mode</Label>
              </div>
              <div className="flex items-center space-x-2">
                <Switch id="evolution" checked={evolutionEnabled} onCheckedChange={setEvolutionEnabled} />
                <Label htmlFor="evolution">Evolution Enabled</Label>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <h4 className="text-sm font-medium border-b pb-2">Behavioral Limits</h4>
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="maxExecutionDepth">Max Execution Depth</Label>
                <Input id="maxExecutionDepth" type="number" value={maxExecutionDepth} onChange={(e) => setMaxExecutionDepth(parseInt(e.target.value) || 0)} min={1} max={20} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="maxRecursionDepth">Max Recursion Depth</Label>
                <Input id="maxRecursionDepth" type="number" value={maxRecursionDepth} onChange={(e) => setMaxRecursionDepth(parseInt(e.target.value) || 0)} min={0} max={10} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="maxWallTimeMs">Max Wall Time (ms)</Label>
                <Input id="maxWallTimeMs" type="number" value={maxWallTimeMs} onChange={(e) => setMaxWallTimeMs(parseInt(e.target.value) || 0)} min={1000} step={1000} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="maxMemoryGrowthMB">Max Memory Growth (MB)</Label>
                <Input id="maxMemoryGrowthMB" type="number" value={maxMemoryGrowthMB} onChange={(e) => setMaxMemoryGrowthMB(parseInt(e.target.value) || 0)} min={64} step={64} />
              </div>
            </div>
          </div>

          <div className="flex gap-2">
            <Button variant="outline" onClick={() => navigate(ROUTES.AGENT_DETAIL.replace(':id', id))} className="flex-1">
              Cancel
            </Button>
            <Button onClick={handleSave} disabled={saving} className="flex-1">
              {saving ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <Save className="h-4 w-4 mr-2" />}
              Save Changes
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

export default AgentEditPage;