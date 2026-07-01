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
import { usePageTitle } from '@/hooks';

function sanitizeId(raw: string | undefined): string | null {
  const t = raw?.trim();
  if (!t || t === 'undefined' || t === 'null') return null;
  return t;
}

export function AgentEditPage() {
  const { t } = useTranslation();
  usePageTitle(t('agentDetail.editAgent') ?? 'Edit Agent');
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
  const [model, setModel] = useState('gpt-4o-mini');
  const [models, setModels] = useState<Array<{ id: string; name: string; provider: string; provider_label?: string; key_source?: string; tier: string; cost: string }>>([]);
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
      setModel(data.agent.model || 'gpt-4o-mini');
    }
  }, [data]);

  useEffect(() => {
    agentApi.getModels().then((res) => setModels(res.models)).catch(() => {});
  }, []);

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
      await updateAgent.mutateAsync({
        name,
        description,
        model,
        autonomous_enabled: autonomousEnabled,
        evolution_enabled: evolutionEnabled,
      });
      await updatePolicy.mutateAsync({
        agentId: id,
        maxExecutionDepth,
        maxRecursionDepth,
        maxWallTimeMs,
        maxMemoryGrowthMB,
      });
      toast.success(t('agentDetail.settingsSaved'));
      navigate(ROUTES.agentPath(id ?? ''));
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
        <Link to={ROUTES.agentPath(id ?? '')}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          {t('agentDetail.backToAgent')}
        </Link>
      </Button>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-lg bg-gradient-to-br from-brand-500 to-purple-500 flex items-center justify-center">
              <Bot className="h-5 w-5 text-white" />
            </div>
            <div>
              <CardTitle>{t('agentDetail.editAgent')}: {agent.name || agent.agentId}</CardTitle>
              <CardDescription>{t('agentDetail.editAgentDescription')}</CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-4">
            <h4 className="text-sm font-medium border-b pb-2">{t('agentDetail.identity')}</h4>
            <div className="space-y-2">
              <Label htmlFor="name">{t('agentDetail.name')}</Label>
              <Input id="name" value={name} onChange={(e) => setName(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="description">{t('agentDetail.description')}</Label>
              <Textarea id="description" value={description} onChange={(e) => setDescription(e.target.value)} rows={3} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="model">{t('agentDetail.aiModel')}</Label>
              <p className="text-xs text-muted-foreground">{t('agentDetail.aiModelHint')}</p>
              <select
                id="model"
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                value={model}
                onChange={(e) => setModel(e.target.value)}
              >
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
          </div>

          <div className="space-y-4">
            <h4 className="text-sm font-medium border-b pb-2">{t('agentDetail.capabilities')}</h4>
            <div className="flex flex-wrap gap-6">
              <div className="flex items-center space-x-2">
                <Switch id="autonomous" checked={autonomousEnabled} onCheckedChange={setAutonomousEnabled} />
                <Label htmlFor="autonomous">{t('agentDetail.autonomousMode')}</Label>
              </div>
              <div className="flex items-center space-x-2">
                <Switch id="evolution" checked={evolutionEnabled} onCheckedChange={setEvolutionEnabled} />
                <Label htmlFor="evolution">{t('agentDetail.evolutionEnabled')}</Label>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <h4 className="text-sm font-medium border-b pb-2">{t('agentDetail.behavioralLimits')}</h4>
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="maxExecutionDepth">{t('agentDetail.maxExecutionDepth')}</Label>
                <Input id="maxExecutionDepth" type="number" value={maxExecutionDepth} onChange={(e) => setMaxExecutionDepth(parseInt(e.target.value) || 0)} min={1} max={20} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="maxRecursionDepth">{t('agentDetail.maxRecursionDepth')}</Label>
                <Input id="maxRecursionDepth" type="number" value={maxRecursionDepth} onChange={(e) => setMaxRecursionDepth(parseInt(e.target.value) || 0)} min={0} max={10} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="maxWallTimeMs">{t('agentDetail.maxWallTimeMs')}</Label>
                <Input id="maxWallTimeMs" type="number" value={maxWallTimeMs} onChange={(e) => setMaxWallTimeMs(parseInt(e.target.value) || 0)} min={1000} step={1000} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="maxMemoryGrowthMB">{t('agentDetail.maxMemoryGrowthMB')}</Label>
                <Input id="maxMemoryGrowthMB" type="number" value={maxMemoryGrowthMB} onChange={(e) => setMaxMemoryGrowthMB(parseInt(e.target.value) || 0)} min={64} step={64} />
              </div>
            </div>
          </div>

          <div className="flex gap-2">
            <Button variant="outline" onClick={() => navigate(ROUTES.agentPath(id ?? ''))} className="flex-1">
              {t('agents.cancel')}
            </Button>
            <Button onClick={handleSave} disabled={saving} className="flex-1">
              {saving ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <Save className="h-4 w-4 mr-2" />}
              {saving ? t('agentDetail.saving') : t('agentDetail.saveSettings')}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

export default AgentEditPage;