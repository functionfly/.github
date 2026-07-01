'use client';

import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { agentApi, type AgentWallet } from '@/api/agent';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { ArrowLeft, Bot, Loader2, Wallet, TrendingUp, TrendingDown, ArrowUpRight, ArrowDownRight } from 'lucide-react';
import { ROUTES } from '@/lib/constants';

function sanitizeId(raw: string | undefined): string | null {
  const t = raw?.trim();
  if (!t || t === 'undefined' || t === 'null') return null;
  return t;
}

export function AgentWalletPage() {
  const { t } = useTranslation();
  const { id: agentId } = useParams<{ id: string }>();
  const id = sanitizeId(agentId);

  const [wallet, setWallet] = useState<AgentWallet | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;
    agentApi.getWallet(id).then((res) => {
      setWallet(res.wallet);
    }).catch((err) => {
      setError(err instanceof Error ? err.message : 'Failed to load wallet');
    }).finally(() => setLoading(false));
  }, [id]);

  if (!id) {
    return (
      <div className="space-y-4">
        <Button variant="ghost" size="sm" asChild>
          <Link to={ROUTES.AGENT_LIST}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            {t('agentDetail.backToAgents')}
          </Link>
        </Button>
        <p className="text-sm text-muted-foreground">Invalid agent ID</p>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <Loader2 className="h-8 w-8 animate-spin text-brand-500" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="space-y-4">
        <Button variant="ghost" size="sm" asChild>
          <Link to={ROUTES.AGENT_LIST}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            {t('agentDetail.backToAgents')}
          </Link>
        </Button>
        <Card className="border-destructive/40">
          <CardHeader>
            <CardTitle>Error Loading Wallet</CardTitle>
            <CardDescription>{error}</CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto space-y-6">
      <Button variant="ghost" size="sm" asChild>
        <Link to={ROUTES.agentPath(id ?? '')}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          {t('agentDetail.backToAgent')}
        </Link>
      </Button>

      <div className="flex items-center gap-3">
        <div className="h-10 w-10 rounded-lg bg-gradient-to-br from-brand-500 to-purple-500 flex items-center justify-center">
          <Bot className="h-5 w-5 text-white" />
        </div>
        <div>
          <h1 className="text-2xl font-bold">{t('agentDetail.wallet')}</h1>
          <p className="text-muted-foreground font-mono text-sm">{id}</p>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{t('agentDetail.walletBalance')}</CardTitle>
            <Wallet className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">${(wallet?.balance_usd ?? 0).toFixed(4)}</div>
            <p className="text-xs text-muted-foreground">{t('agentDetail.walletBalanceHint')}</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{t('agentDetail.walletTotalEarned')}</CardTitle>
            <TrendingUp className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-500">${(wallet?.total_earned_usd ?? 0).toFixed(4)}</div>
            <p className="text-xs text-muted-foreground">
              {wallet?.last_earning_at ? `Last: ${new Date(wallet.last_earning_at).toLocaleDateString()}` : t('agentDetail.walletNoEarnings')}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{t('agentDetail.walletTotalSpent')}</CardTitle>
            <TrendingDown className="h-4 w-4 text-red-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-500">${(wallet?.total_spent_usd ?? 0).toFixed(4)}</div>
            <p className="text-xs text-muted-foreground">
              {wallet?.last_spending_at ? `Last: ${new Date(wallet.last_spending_at).toLocaleDateString()}` : t('agentDetail.walletNoSpending')}
            </p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t('agentDetail.walletDetails')}</CardTitle>
          <CardDescription>{t('agentDetail.walletDetailsDescription')}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium">{t('agentDetail.walletEscrowBalance')}</p>
              <p className="text-sm text-muted-foreground">{t('agentDetail.walletEscrowHint')}</p>
            </div>
            <Badge variant="outline">${(wallet?.escrow_balance_usd ?? 0).toFixed(4)}</Badge>
          </div>
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium">Agent ID</p>
              <p className="text-sm text-muted-foreground">Unique agent identifier</p>
            </div>
            <Badge variant="outline" className="font-mono">{wallet?.agent_id ?? id}</Badge>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

export default AgentWalletPage;