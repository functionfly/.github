'use client';

import { useParams, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { ArrowLeft, Bot } from 'lucide-react';
import { ROUTES } from '@/lib/constants';
import AgentAnalyticsComponent from '@/components/analytics';

function sanitizeId(raw: string | undefined): string | null {
  const t = raw?.trim();
  if (!t || t === 'undefined' || t === 'null') return null;
  return t;
}

export function AgentAnalyticsPage() {
  const { t } = useTranslation();
  const { id: agentId } = useParams<{ id: string }>();
  const id = sanitizeId(agentId);

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

  return (
    <div className="space-y-6">
      <Button variant="ghost" size="sm" asChild>
        <Link to={ROUTES.agentPath(id ?? '')}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Agent
        </Link>
      </Button>

      <div className="flex items-center gap-3">
        <div className="h-10 w-10 rounded-lg bg-gradient-to-br from-brand-500 to-purple-500 flex items-center justify-center">
          <Bot className="h-5 w-5 text-white" />
        </div>
        <div>
          <h1 className="text-2xl font-bold">Agent Analytics</h1>
          <p className="text-muted-foreground font-mono text-sm">{id}</p>
        </div>
      </div>

      <AgentAnalyticsComponent agentId={id} />
    </div>
  );
}

export default AgentAnalyticsPage;