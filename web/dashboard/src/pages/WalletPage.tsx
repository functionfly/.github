import { agentApi, type AgentIdentity } from '@/api/agent';
import { getWalletInfo, getWalletTransactions, topUpWallet } from '@/api/billing';
import { notificationsApi } from '@/api/notifications';
import { WalletDashboard } from '@/components/swarm/WalletDashboard';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { trackEvent } from '@/lib/analytics';
import { cn } from '@/lib/utils';
import { useNotificationStore } from '@/stores/notificationStore';
import { useQuery } from '@tanstack/react-query';
import {
  ArrowDownRight,
  ArrowUpRight,
  ChevronLeft,
  CreditCard,
  Loader2,
  Plus,
  Wallet,
} from 'lucide-react';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';

const LAST_WALLET_AGENT_KEY = 'ff-last-wallet-agent-id';
const FUND_PRESETS = [10, 25, 50, 100];

function persistLastWalletAgent(id: string) {
  try {
    localStorage.setItem(LAST_WALLET_AGENT_KEY, id);
  } catch {
    /* ignore */
  }
}

function sanitizeWalletAgentIdParam(raw: string | null | undefined): string | null {
  const t = raw?.trim();
  if (!t || t === 'undefined' || t === 'null') return null;
  return t;
}

interface PlatformWalletTransaction {
  id: string;
  type: 'credit' | 'debit' | 'top_up' | 'refund' | 'payout';
  amount: number;
  description: string;
  timestamp: string;
  status: 'completed' | 'pending' | 'failed';
}

// ──────────────────────────────────────────────────────────────────────────────
// Platform Wallet — shown when no agents exist, or when accessed without an agent
// ──────────────────────────────────────────────────────────────────────────────
function PlatformWalletView() {
  const { t } = useTranslation();
  const [searchParams, setSearchParams] = useSearchParams();
  const [showAddFunds, setShowAddFunds] = useState(false);
  const [addFundsAmount, setAddFundsAmount] = useState('25.00');
  const [addFundsLoading, setAddFundsLoading] = useState(false);
  const [transactionFilter, setTransactionFilter] = useState<'all' | 'credit' | 'debit'>('all');

  const {
    data: wallet,
    isLoading: walletLoading,
    refetch: refetchWallet,
  } = useQuery({
    queryKey: ['platform-wallet'],
    queryFn: getWalletInfo,
    retry: 2,
  });

  const {
    data: txData,
    isLoading: txLoading,
  } = useQuery({
    queryKey: ['platform-wallet-transactions'],
    queryFn: () => getWalletTransactions(50),
    retry: 2,
  });

  const transactions: PlatformWalletTransaction[] = txData?.transactions ?? [];

  useEffect(() => {
    const walletTopUp = searchParams.get('walletTopUp');
    if (!walletTopUp) return;
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        next.delete('walletTopUp');
        return next;
      },
      { replace: true }
    );
    if (walletTopUp === 'success') {
      toast.success('Funds added successfully', {
        description: 'Your wallet has been topped up.',
      });
      refetchWallet();
    } else if (walletTopUp === 'cancel') {
      toast.info('Payment cancelled');
    }
  }, [searchParams, setSearchParams, refetchWallet]);

  const handleAddFunds = async () => {
    const amount = parseFloat(addFundsAmount);
    if (isNaN(amount) || amount <= 0) {
      toast.error('Please enter a valid amount');
      return;
    }
    trackEvent('wallet_checkout_initiated', { amount });
    setAddFundsLoading(true);
    try {
      const origin = window.location.origin;
      const { checkout_url } = await topUpWallet(
        amount,
        `${origin}/wallet?walletTopUp=success`,
        `${origin}/wallet?walletTopUp=cancel`
      );
      window.location.href = checkout_url;
    } catch {
      trackEvent('wallet_checkout_failed');
      toast.error('Could not initiate checkout. Please try again.');
    } finally {
      setAddFundsLoading(false);
    }
  };

  const filteredTx = transactions.filter((tx) => {
    if (transactionFilter === 'all') return true;
    if (transactionFilter === 'credit') return tx.type === 'credit' || tx.type === 'top_up' || tx.type === 'refund';
    if (transactionFilter === 'debit') return tx.type === 'debit' || tx.type === 'payout';
    return true;
  });

  const formatAmount = (amount: number, type: string) => {
    const sign = type === 'credit' || type === 'top_up' || type === 'refund' ? '+' : '-';
    return `${sign}$${Math.abs(amount).toFixed(2)}`;
  };

  const txIcon = (type: string) => {
    if (type === 'credit' || type === 'top_up' || type === 'refund') {
      return <ArrowDownRight className="h-4 w-4 text-emerald-400" />;
    }
    return <ArrowUpRight className="h-4 w-4 text-red-400" />;
  };

  if (walletLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="space-y-4 w-full">
      {/* Back + Header */}
      <div className="flex items-center gap-2">
        <Link to="/dashboard" className="p-1.5 rounded-lg hover:bg-surface-elevated transition-colors">
          <ChevronLeft className="h-4 w-4 text-text-muted" />
        </Link>
        <div>
          <h1 className="text-xl font-bold font-display text-text-primary">My Wallet</h1>
          <p className="text-xs text-text-secondary">Platform wallet for purchases and deposits</p>
        </div>
      </div>

      {/* Balance + Quick Actions */}
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-3">
        {/* Main Balance Card */}
        <Card className="lg:col-span-3">
          <CardHeader className="pb-2 pt-4 px-4">
            <CardTitle className="flex items-center gap-2 text-sm">
              <Wallet className="h-4 w-4 text-brand-500" />
              Balance
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3 px-4 pb-4">
            <div className="flex items-baseline gap-1.5">
              <span className="text-3xl font-bold text-text-primary font-display">
                ${wallet?.balance_usd?.toFixed(2) ?? '0.00'}
              </span>
              <span className="text-text-muted text-xs">USD</span>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="rounded-lg bg-surface-elevated p-2.5">
                <p className="text-[11px] text-text-muted mb-0.5">Lifetime Earnings</p>
                <p className="text-sm font-semibold text-emerald-400">
                  +${wallet?.lifetime_earnings_usd?.toFixed(2) ?? '0.00'}
                </p>
              </div>
              <div className="rounded-lg bg-surface-elevated p-2.5">
                <p className="text-[11px] text-text-muted mb-0.5">Lifetime Fees</p>
                <p className="text-sm font-semibold text-red-400">
                  -${wallet?.lifetime_fees_usd?.toFixed(2) ?? '0.00'}
                </p>
              </div>
            </div>
            {/* Spending bar */}
            <div className="space-y-1">
              <div className="flex justify-between text-[11px] text-text-muted">
                <span>Spending vs Earnings</span>
                <span>
                  {wallet?.lifetime_earnings_usd
                    ? ((wallet.lifetime_fees_usd / wallet.lifetime_earnings_usd) * 100).toFixed(0)
                    : 0}%
                </span>
              </div>
              <div className="h-1.5 w-full rounded-full bg-surface-elevated overflow-hidden">
                <div
                  className="h-full rounded-full bg-gradient-to-r from-brand-500 to-emerald-500"
                  style={{
                    width: wallet?.lifetime_earnings_usd
                      ? `${Math.min((wallet.lifetime_fees_usd / wallet.lifetime_earnings_usd) * 100, 100)}%`
                      : '0%',
                  }}
                />
              </div>
            </div>
            <Button onClick={() => {
              trackEvent('wallet_add_funds_opened');
              setShowAddFunds(true);
            }} className="gap-2 w-full sm:w-auto h-8 text-sm">
              <Plus className="h-3.5 w-3.5" />
              Add Funds
            </Button>
          </CardContent>
        </Card>

        {/* Quick Links Sidebar */}
        <Card>
          <CardHeader className="pb-2 pt-4 px-4">
            <CardTitle className="text-sm">Quick Links</CardTitle>
          </CardHeader>
          <CardContent className="space-y-1.5 px-4 pb-4">
            <Link to="/agents">
              <Button variant="outline" className="w-full justify-start gap-2 text-xs h-8">
                <Wallet className="h-3.5 w-3.5" />
                Agent Wallets
              </Button>
            </Link>
            <Link to="/settings#billing">
              <Button variant="outline" className="w-full justify-start gap-2 text-xs h-8">
                <CreditCard className="h-3.5 w-3.5" />
                Billing Settings
              </Button>
            </Link>
          </CardContent>
        </Card>
      </div>

      {/* Transaction History */}
      <Card>
        <CardHeader className="pb-2 pt-4 px-4">
          <CardTitle className="text-sm">Transaction History</CardTitle>
          <CardDescription className="text-xs">Recent wallet activity</CardDescription>
        </CardHeader>
        <CardContent className="px-4 pb-4">
          <Tabs
            value={transactionFilter}
            onValueChange={(v) => setTransactionFilter(v as typeof transactionFilter)}
          >
            <TabsList className="grid w-full grid-cols-3 h-8">
              <TabsTrigger value="all" className="text-xs">All</TabsTrigger>
              <TabsTrigger value="credit" className="text-xs">Credits</TabsTrigger>
              <TabsTrigger value="debit" className="text-xs">Debits</TabsTrigger>
            </TabsList>

            <TabsContent value={transactionFilter} className="pt-2">
              {txLoading ? (
                <div className="flex items-center justify-center py-6">
                  <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
                </div>
              ) : filteredTx.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-6 text-center">
                  <Wallet className="h-7 w-7 text-text-muted mb-1.5" />
                  <p className="text-sm text-text-muted">No transactions yet</p>
                  <p className="text-xs text-text-muted mt-0.5">Add funds to get started</p>
                </div>
              ) : (
                <div className="space-y-0.5">
                  {filteredTx.map((tx) => (
                    <div
                      key={tx.id}
                      className="flex items-center justify-between py-2 border-b border-border-subtle last:border-0"
                    >
                      <div className="flex items-center gap-2.5">
                        <div
                          className={cn(
                            'flex items-center justify-center w-7 h-7 rounded-full',
                            tx.type === 'credit' || tx.type === 'top_up' || tx.type === 'refund'
                              ? 'bg-emerald-500/10'
                              : 'bg-red-500/10'
                          )}
                        >
                          {txIcon(tx.type)}
                        </div>
                        <div>
                          <p className="text-sm font-medium text-text-primary">{tx.description}</p>
                          <p className="text-[11px] text-text-muted">
                            {new Date(tx.timestamp).toLocaleDateString('en-US', {
                              month: 'short',
                              day: 'numeric',
                              hour: '2-digit',
                              minute: '2-digit',
                            })}
                          </p>
                        </div>
                      </div>
                      <span
                        className={cn(
                          'text-sm font-semibold font-mono',
                          tx.type === 'credit' || tx.type === 'top_up' || tx.type === 'refund'
                            ? 'text-emerald-400'
                            : 'text-red-400'
                        )}
                      >
                        {formatAmount(tx.amount, tx.type)}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>

      {/* Add Funds Dialog */}
      <Dialog open={showAddFunds} onOpenChange={setShowAddFunds}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Add Funds</DialogTitle>
            <DialogDescription>
              Select or enter an amount to add to your wallet
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            {/* Preset quick-select buttons */}
            <div className="flex gap-2">
              {FUND_PRESETS.map((preset) => (
                <button
                  key={preset}
                  onClick={() => {
                    trackEvent('wallet_preset_selected', { preset });
                    setAddFundsAmount(String(preset));
                  }}
                  className={cn(
                    'flex-1 py-2 rounded-lg border text-sm font-medium transition-all',
                    addFundsAmount === String(preset)
                      ? 'border-brand-500 bg-brand-500/10 text-brand-400'
                      : 'border-border-subtle bg-surface-elevated text-text-secondary hover:border-brand-500/40'
                  )}
                >
                  ${preset}
                </button>
              ))}
            </div>
            <div className="space-y-2">
              <Label htmlFor="amount">Amount (USD)</Label>
              <Input
                id="amount"
                type="number"
                min="1"
                step="0.01"
                value={addFundsAmount}
                onChange={(e) => setAddFundsAmount(e.target.value)}
                placeholder="25.00"
              />
            </div>
            {addFundsAmount && !isNaN(parseFloat(addFundsAmount)) && parseFloat(addFundsAmount) > 0 && (
              <p className="text-xs text-text-muted">
                You will be charged <span className="font-medium text-text-primary">${parseFloat(addFundsAmount).toFixed(2)}</span> via Stripe.
              </p>
            )}
          </div>
          <DialogFooter className="gap-2 sm:gap-0">
            <Button variant="outline" onClick={() => setShowAddFunds(false)}>
              Cancel
            </Button>
            <Button onClick={handleAddFunds} disabled={addFundsLoading || !addFundsAmount}>
              {addFundsLoading ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Redirecting...
                </>
              ) : (
                <>
                  <CreditCard className="h-4 w-4" />
                  Checkout
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

// ──────────────────────────────────────────────────────────────────────────────
// WalletPage — routes: /wallet, /wallet/:agentId
// Falls back to platform wallet when no agents exist
// ──────────────────────────────────────────────────────────────────────────────
export function WalletPage() {
  const { t } = useTranslation();
  const { slug: pathAgentId } = useParams<{ slug: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();
  const queryAgent = searchParams.get('agent')?.trim() || undefined;
  const updateUnreadCounts = useNotificationStore((s) => s.updateUnreadCounts);

  const explicitId =
    sanitizeWalletAgentIdParam(pathAgentId) ?? sanitizeWalletAgentIdParam(queryAgent) ?? null;

  const [phase, setPhase] = useState<'idle' | 'resolving' | 'pick' | 'platform'>('idle');
  const [agents, setAgents] = useState<AgentIdentity[]>([]);

  // Handle Stripe return
  useEffect(() => {
    const credits = searchParams.get('credits');
    if (!credits) return;

    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        next.delete('credits');
        return next;
      },
      { replace: true }
    );

    if (credits === 'success') {
      toast.success(t('walletPage.fundsAddedSuccess'), {
        description: t('walletPage.fundsAddedDescription'),
      });
      notificationsApi
        .fetchUnreadCounts()
        .then((counts) => {
          const byCategory = counts?.byCategory || {};
          updateUnreadCounts({
            all: counts?.total || 0,
            trust: byCategory.trust || 0,
            revenue: byCategory.revenue || 0,
            issues: byCategory.issues || 0,
            messages: byCategory.messages || 0,
            security: byCategory.security || 0,
          });
        })
        .catch(() => { /* silent */ });
    } else if (credits === 'cancel') {
      toast.info(t('walletPage.paymentCancelled'), {
        description: t('walletPage.paymentCancelledDescription'),
      });
    }
  }, [searchParams, setSearchParams, updateUnreadCounts]);

  useEffect(() => {
    if (explicitId) {
      persistLastWalletAgent(explicitId);
      setPhase('idle');
      return;
    }

    let cancelled = false;
    setPhase('resolving');

    (async () => {
      try {
        const res = await agentApi.listAgents({ limit: 100 });
        if (cancelled) return;
        const list = res.agents ?? [];

        if (list.length === 0) {
          setPhase('platform');
          return;
        }

        if (list.length === 1) {
          const id = list[0].agentId;
          persistLastWalletAgent(id);
          navigate(`/wallet/agents/${encodeURIComponent(id)}`, { replace: true });
          return;
        }

        let last: string | null = null;
        try {
          last = localStorage.getItem(LAST_WALLET_AGENT_KEY);
        } catch { /* ignore */ }

        if (last && list.some((a) => a.agentId === last)) {
          navigate(`/wallet/agents/${encodeURIComponent(last)}`, { replace: true });
          return;
        }

        setAgents(list);
        setPhase('pick');
      } catch {
        if (!cancelled) {
          setPhase('platform');
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [explicitId, navigate]);

  if (explicitId) {
    return <WalletDashboard agentId={explicitId} />;
  }

  if (phase === 'platform') {
    return <PlatformWalletView />;
  }

  if (phase === 'resolving' || phase === 'idle') {
    return (
      <div className="flex flex-col items-center justify-center gap-3 p-12 text-muted-foreground">
        <Loader2 className="h-8 w-8 animate-spin" />
        <p className="text-sm">{t('walletPage.loadingWallet')}</p>
      </div>
    );
  }

  if (phase === 'pick') {
    return (
      <div className="mx-auto max-w-lg space-y-4 p-6">
        {/* Back link */}
        <Link
          to="/wallet"
          className="flex items-center gap-2 text-sm text-text-muted hover:text-text-primary transition-colors"
        >
          <ChevronLeft className="h-4 w-4" />
          Back to My Wallet
        </Link>

        <h1 className="text-2xl font-bold font-display">{t('walletPage.chooseAgent')}</h1>
        <p className="text-sm text-text-secondary">
          Select an agent to view its wallet, or{' '}
          <Link to="/wallet" className="text-brand-500 hover:underline">
            view your platform wallet
          </Link>
          .
        </p>

        <div className="flex flex-col gap-2">
          {agents.map((a) => (
            <Button
              key={a.agentId}
              variant="outline"
              className="h-auto w-full justify-start py-3"
              asChild
            >
              <Link
                to={`/wallet/agents/${encodeURIComponent(a.agentId)}`}
                onClick={() => persistLastWalletAgent(a.agentId)}
              >
                <span className="font-medium">{a.name || a.agentId}</span>
                <span className="ml-2 text-xs text-muted-foreground font-mono">{a.agentId}</span>
              </Link>
            </Button>
          ))}
        </div>

        <div className="flex items-center gap-3 pt-4 border-t">
          <Link to="/wallet" className="flex-1">
            <Button variant="ghost" className="gap-2 w-full justify-start">
              <Wallet className="h-4 w-4" />
              My Platform Wallet
            </Button>
          </Link>
          <Link to="/agents">
            <Button variant="outline" size="sm" className="gap-1.5">
              <Wallet className="h-4 w-4" />
              Go to Agents
            </Button>
          </Link>
        </div>
      </div>
    );
  }

  return null;
}

export default WalletPage;
