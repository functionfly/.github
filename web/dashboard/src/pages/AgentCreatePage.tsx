'use client';

import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { agentApi } from '@/api/agent';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { ArrowLeft, Bot, Loader2, Copy, Check } from 'lucide-react';
import { Link } from 'react-router-dom';
import { ROUTES } from '@/lib/constants';
import { toast } from 'sonner';

function slugFrom(s: string) {
  return (s ?? '')
    .trim()
    .toLowerCase()
    .replace(/\s+/g, '-')
    .replace(/[^a-z0-9-]/g, '');
}

export function AgentCreatePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [form, setForm] = useState({ agentId: '', name: '', description: '' });
  const [submitting, setSubmitting] = useState(false);
  const [apiKey, setApiKey] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const agentIdRaw = form.agentId ?? '';
  const agentIdInvalid =
    agentIdRaw.trim() !== '' && /[^a-z0-9-]/.test(agentIdRaw.trim().toLowerCase());

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const agentId = slugFrom(agentIdRaw);
    const name = (form.name ?? '').trim();
    if (!agentId || !name) {
      toast.error(t('agents.agentIdNameRequired'));
      return;
    }
    setSubmitting(true);
    try {
      const res = await agentApi.registerAgent({ agentId, name, description: form.description.trim() || undefined });
      setApiKey(res.api_key);
      toast.success(t('agents.agentCreatedSuccess'));
    } catch (err: unknown) {
      const errMsg = err instanceof Error ? err.message : String(err);
      toast.error(t('agents.failedToCreate') + ': ' + errMsg);
    } finally {
      setSubmitting(false);
    }
  };

  const handleCopy = () => {
    if (apiKey) {
      navigator.clipboard.writeText(apiKey);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  if (apiKey) {
    return (
      <div className="max-w-xl mx-auto space-y-6">
        <Button variant="ghost" size="sm" asChild>
          <Link to={ROUTES.AGENT_LIST}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Agents
          </Link>
        </Button>
        <Card className="border-green-500/30 bg-green-500/5">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Check className="h-5 w-5 text-green-500" />
              Agent Created Successfully
            </CardTitle>
            <CardDescription>
              Copy your API key below — it won't be shown again.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center gap-2 p-3 bg-bg-secondary rounded-lg font-mono text-sm break-all">
              {apiKey}
            </div>
            <Button onClick={handleCopy} variant="outline" className="w-full">
              {copied ? <Check className="h-4 w-4 mr-2" /> : <Copy className="h-4 w-4 mr-2" />}
              {copied ? t('agents.copied') : t('agents.copy')} API Key
            </Button>
            <Button onClick={() => navigate(ROUTES.AGENT_LIST)} className="w-full">
              Go to Agents
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="max-w-xl mx-auto space-y-6">
      <Button variant="ghost" size="sm" asChild>
        <Link to={ROUTES.AGENT_LIST}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Agents
        </Link>
      </Button>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-lg bg-gradient-to-br from-brand-500 to-purple-500 flex items-center justify-center">
              <Bot className="h-5 w-5 text-white" />
            </div>
            <div>
              <CardTitle>Create New Agent</CardTitle>
              <CardDescription>
                Register a new AI agent. You'll get an API key to authenticate requests.
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="agentId">Agent ID</Label>
              <Input
                id="agentId"
                placeholder="e.g. my-agent"
                value={form.agentId}
                onChange={(e) => setForm((f) => ({ ...f, agentId: e.target.value }))}
                className="font-mono"
              />
              <p className="text-xs text-text-muted">
                Lowercase letters, numbers, and hyphens only. Used in API calls and URLs.
              </p>
              {agentIdInvalid && (
                <p className="text-xs text-destructive">{t('agents.agentIdChars')}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                placeholder="e.g. My Agent"
                value={form.name}
                onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="description">Description (optional)</Label>
              <Textarea
                id="description"
                placeholder="e.g. Handles support queries and triages tickets"
                value={form.description}
                onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
                rows={3}
              />
            </div>

            <div className="flex gap-2">
              <Button type="button" variant="outline" onClick={() => navigate(ROUTES.AGENT_LIST)} className="flex-1">
                Cancel
              </Button>
              <Button type="submit" disabled={submitting || !form.name || !form.agentId || agentIdInvalid} className="flex-1">
                {submitting ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : null}
                Create Agent
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}

export default AgentCreatePage;