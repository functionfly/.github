import { agentApi, type AgentFinancialTransaction } from '@/api/agent';
import { notificationsApi } from '@/api/notifications';
import { Badge } from '@/components/ui/badge';
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
import { Progress } from '@/components/ui/progress';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useNotificationStore } from '@/stores/notificationStore';
import {
  ArrowDownRight,
  ArrowUpRight,
  CreditCard,
  DollarSign,
  Download,
  Loader2,
  Plus,
  Shield,
  TrendingDown,
  TrendingUp,
  Wallet,
} from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';
import { toast } from 'sonner';

// Types for Wallet Dashboard
interface WalletData {
  balance: number;
  escrowBalance: number;
  totalEarned: number;
  totalSpent: number;
  lastEarning: string;
  lastSpending: string;
}

interface Transaction {
  id: string;
  type:
    | 'credit'
    | 'debit'
    | 'transfer'
    | 'credit_purchase'
    | 'execution_debit'
    | 'transfer_in'
    | 'transfer_out'
    | 'adjustment'
    | 'refund';
  amount: number;
  description: string;
  counterparty?: string;
  timestamp: string;
  status: 'completed' | 'pending' | 'failed';
}

interface RevenueBreakdown {
  category: string;
  amount: number;
  percentage: number;
}

interface SpendingBreakdown {
  category: string;
  amount: number;
  percentage: number;
}

export function WalletDashboard({ agentId }: { agentId: string }) {
  const updateUnreadCounts = useNotificationStore((s) => s.updateUnreadCounts);
  const [wallet, setWallet] = useState<WalletData | null>(null);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [revenueBreakdown, setRevenueBreakdown] = useState<RevenueBreakdown[]>([]);
  const [spendingBreakdown, setSpendingBreakdown] = useState<SpendingBreakdown[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Add Funds dialog state
  const [showAddFunds, setShowAddFunds] = useState(false);
  const [addFundsAmount, setAddFundsAmount] = useState('10.00');
  const [addFundsLoading, setAddFundsLoading] = useState(false);
  const [addFundsError, setAddFundsError] = useState<string | null>(null);

  // Transactions pagination
  const [txLimit, setTxLimit] = useState(20);
  const [txOffset, setTxOffset] = useState(0);
  const [txTotal, setTxTotal] = useState(0);

  const mapTxKind = (kind: string): Transaction['type'] => {
    switch (kind) {
      case 'credit_purchase':
        return 'credit_purchase';
      case 'execution_debit':
        return 'execution_debit';
      case 'transfer_in':
        return 'transfer_in';
      case 'transfer_out':
        return 'transfer_out';
      case 'adjustment':
        return 'adjustment';
      case 'refund':
        return 'refund';
      default:
        return 'credit';
    }
  };

  const formatTxDescription = (tx: AgentFinancialTransaction): string => {
    switch (tx.kind) {
      case 'credit_purchase':
        return 'Credits purchased';
      case 'execution_debit':
        return 'Execution cost';
      case 'transfer_in':
        return 'Transfer received';
      case 'transfer_out':
        return 'Transfer sent';
      case 'adjustment':
        return 'Balance adjustment';
      case 'refund':
        return 'Refund';
      default:
        return 'Transaction';
    }
  };

  const fetchData = useCallback(async () => {
    if (!agentId) {
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const [walletRes, costRes, txRes] = await Promise.allSettled([
        agentApi.getWallet(agentId),
        agentApi.getCostBreakdown(agentId),
        agentApi.listWalletTransactions(agentId, { limit: txLimit, offset: txOffset }),
      ]);

      const w = walletRes.status === 'fulfilled' ? walletRes.value?.wallet : null;
      const walletObj = w != null ? (w as unknown as Record<string, unknown>) : null;
      if (walletObj) {
        const balance = Number(walletObj.balance_usd ?? walletObj.balanceUSD ?? 0);
        const escrow = Number(walletObj.escrow_balance_usd ?? walletObj.escrowBalanceUSD ?? 0);
        const earned = Number(walletObj.total_earned_usd ?? walletObj.totalEarnedUSD ?? 0);
        const spent = Number(walletObj.total_spent_usd ?? walletObj.totalSpentUSD ?? 0);
        const lastEarning = (walletObj.last_earning_at ?? walletObj.lastEarningAt ?? '') as string;
        const lastSpending = (walletObj.last_spending_at ??
          walletObj.lastSpendingAt ??
          '') as string;
        setWallet({
          balance,
          escrowBalance: escrow,
          totalEarned: earned,
          totalSpent: spent,
          lastEarning,
          lastSpending,
        });
        if (earned > 0) {
          setRevenueBreakdown([{ category: 'Earnings', amount: earned, percentage: 100 }]);
        } else {
          setRevenueBreakdown([]);
        }
      } else {
        setWallet(null);
        setRevenueBreakdown([]);
      }

      const breakdown = costRes.status === 'fulfilled' ? costRes.value?.breakdown : null;
      const breakdownObj =
        breakdown != null ? (breakdown as unknown as Record<string, unknown>) : null;
      const byFn = (breakdown?.byFunction ?? breakdownObj?.by_function) as
        | Record<string, number>
        | undefined;
      if (byFn && typeof byFn === 'object') {
        const entries = Object.entries(byFn).filter(([, v]) => Number(v) > 0);
        const total = entries.reduce((sum, [, v]) => sum + Number(v), 0);
        const list: SpendingBreakdown[] =
          total > 0
            ? entries.map(([category, amount]) => ({
                category: category.replace(/_/g, ' '),
                amount: Number(amount),
                percentage: Math.round((Number(amount) / total) * 100),
              }))
            : [];
        setSpendingBreakdown(list);
      } else {
        setSpendingBreakdown([]);
      }

      if (txRes.status === 'fulfilled' && txRes.value?.transactions) {
        const mapped: Transaction[] = txRes.value.transactions.map((t) => ({
          id: t.id,
          type: mapTxKind(t.kind),
          amount: t.amount_usd,
          description: formatTxDescription(t),
          timestamp: t.created_at,
          status: t.status,
          counterparty: t.provider,
        }));
        setTransactions(mapped);
        setTxTotal(txRes.value.total ?? 0);
      } else {
        setTransactions([]);
        setTxTotal(0);
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load wallet data');
      setWallet(null);
      setTransactions([]);
      setRevenueBreakdown([]);
      setSpendingBreakdown([]);
    } finally {
      setLoading(false);
    }
  }, [agentId, txLimit, txOffset]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const refreshNotificationBell = useCallback(async () => {
    try {
      const counts = await notificationsApi.fetchUnreadCounts();
      const byCategory = counts?.byCategory || {};
      updateUnreadCounts({
        all: counts?.total || 0,
        trust: byCategory.trust || 0,
        revenue: byCategory.revenue || 0,
        issues: byCategory.issues || 0,
        messages: byCategory.messages || 0,
        security: byCategory.security || 0,
      });
    } catch {
      /* silent – badge catches up on next poll */
    }
  }, [updateUnreadCounts]);

  const handleAddFunds = async () => {
    setAddFundsError(null);
    const amount = parseFloat(addFundsAmount);
    if (isNaN(amount) || amount <= 0) {
      setAddFundsError('Please enter a valid amount');
      return;
    }
    if (amount < 0.5) {
      setAddFundsError('Minimum amount is $0.50');
      return;
    }

    setAddFundsLoading(true);
    try {
      const successUrl = `${window.location.origin}/wallet/agents/${encodeURIComponent(agentId)}?credits=success`;
      const cancelUrl = `${window.location.origin}/wallet/agents/${encodeURIComponent(agentId)}?credits=cancel`;
      const res = await agentApi.createCreditsCheckout(agentId, amount, successUrl, cancelUrl);
      if (res?.url) {
        // Redirect to Stripe – notification bell refresh happens in WalletPage on return
        window.location.href = res.url;
      } else {
        setAddFundsError('Failed to create checkout session');
      }
    } catch (e: unknown) {
      const ax = e as {
        response?: { data?: { error?: { code?: string; message?: string } } };
        message?: string;
      };
      const errObj = ax.response?.data?.error;

      // If Stripe is not configured the backend falls back to a simulated credit purchase
      if (errObj?.code === 'PAYMENTS_NOT_CONFIGURED') {
        try {
          await agentApi.purchaseCredits(agentId, amount);
          setShowAddFunds(false);
          toast.success('Funds added successfully', {
            description: `$${amount.toFixed(2)} has been added to your wallet.`,
          });
          fetchData();
          refreshNotificationBell();
          return;
        } catch (innerErr: unknown) {
          const msg = innerErr instanceof Error ? innerErr.message : 'Failed to add funds';
          setAddFundsError(msg);
          return;
        }
      }

      const apiMsg =
        errObj?.message && (errObj.code ? `${errObj.code}: ${errObj.message}` : errObj.message);
      setAddFundsError(apiMsg ?? (e instanceof Error ? e.message : 'Failed to initiate checkout'));
    } finally {
      setAddFundsLoading(false);
    }
  };

  const handleExport = () => {
    if (!transactions.length) return;

    const headers = ['Date', 'Type', 'Description', 'Amount', 'Status'];
    const rows = transactions.map((t) => [
      new Date(t.timestamp).toISOString(),
      t.type,
      t.description,
      t.amount.toString(),
      t.status,
    ]);
    const csv = [headers.join(','), ...rows.map((r) => r.join(','))].join('\n');
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `wallet-transactions-${agentId}-${new Date().toISOString().split('T')[0]}.csv`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(amount);
  };

  const getTransactionIcon = (type: string) => {
    switch (type) {
      case 'credit':
      case 'credit_purchase':
      case 'transfer_in':
      case 'refund':
        return <ArrowDownRight className="h-4 w-4 text-green-500" />;
      case 'debit':
      case 'execution_debit':
      case 'transfer_out':
        return <ArrowUpRight className="h-4 w-4 text-red-500" />;
      case 'transfer':
      case 'adjustment':
        return <TrendingUp className="h-4 w-4 text-blue-500" />;
      default:
        return <DollarSign className="h-4 w-4 text-gray-500" />;
    }
  };

  const getAmountPrefix = (type: string) => {
    switch (type) {
      case 'credit':
      case 'credit_purchase':
      case 'transfer_in':
      case 'refund':
        return '+';
      case 'debit':
      case 'execution_debit':
      case 'transfer_out':
        return '-';
      default:
        return '';
    }
  };

  const getAmountColor = (type: string) => {
    switch (type) {
      case 'credit':
      case 'credit_purchase':
      case 'transfer_in':
      case 'refund':
        return 'text-green-500';
      case 'debit':
      case 'execution_debit':
      case 'transfer_out':
        return 'text-red-500';
      default:
        return 'text-blue-500';
    }
  };

  const getStatusBadge = (status: string) => {
    const colors = {
      completed: 'bg-green-500',
      pending: 'bg-yellow-500',
      failed: 'bg-red-500',
    };
    return (
      <Badge className={`${colors[status as keyof typeof colors]} text-white text-xs`}>
        {status}
      </Badge>
    );
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4 text-destructive">
        <p className="font-medium">Failed to load wallet data</p>
        <p className="text-sm mt-1">{error}</p>
        <Button variant="outline" size="sm" className="mt-3" onClick={() => fetchData()}>
          Retry
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-2">
            <Wallet className="h-8 w-8" />
            Agent Wallet
          </h1>
          <p className="text-muted-foreground mt-1">
            Manage your agent's finances and transactions
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" onClick={() => setShowAddFunds(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Add Funds
          </Button>
          <Button variant="outline" onClick={handleExport} disabled={transactions.length === 0}>
            <Download className="h-4 w-4 mr-2" />
            Export
          </Button>
        </div>
      </div>

      {/* Balance Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="bg-gradient-to-br from-green-500/10 to-green-500/5 border-green-500/20">
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Available Balance</p>
                <p className="text-3xl font-bold text-green-600">
                  {formatCurrency(wallet?.balance ?? 0)}
                </p>
              </div>
              <Wallet className="h-8 w-8 text-green-500" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Escrow Balance</p>
                <p className="text-3xl font-bold">{formatCurrency(wallet?.escrowBalance ?? 0)}</p>
              </div>
              <Shield className="h-8 w-8 text-blue-500" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Total Earned</p>
                <p className="text-3xl font-bold text-green-600">
                  {formatCurrency(wallet?.totalEarned ?? 0)}
                </p>
              </div>
              <TrendingUp className="h-8 w-8 text-green-500" />
            </div>
            {wallet?.lastEarning ? (
              <p className="text-xs text-muted-foreground mt-2">
                Last: {new Date(wallet.lastEarning).toLocaleDateString()}
              </p>
            ) : null}
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Total Spent</p>
                <p className="text-3xl font-bold text-red-600">
                  {formatCurrency(wallet?.totalSpent ?? 0)}
                </p>
              </div>
              <TrendingDown className="h-8 w-8 text-red-500" />
            </div>
            {wallet?.lastSpending ? (
              <p className="text-xs text-muted-foreground mt-2">
                Last: {new Date(wallet.lastSpending).toLocaleDateString()}
              </p>
            ) : null}
          </CardContent>
        </Card>
      </div>

      {/* Add Funds Dialog */}
      <Dialog open={showAddFunds} onOpenChange={setShowAddFunds}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Add Funds</DialogTitle>
            <DialogDescription>
              Purchase execution credits for your agent. You'll be redirected to Stripe to complete
              the payment.
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="amount">Amount (USD)</Label>
              <Input
                id="amount"
                type="number"
                step="0.01"
                min="0.50"
                placeholder="10.00"
                value={addFundsAmount}
                onChange={(e) => setAddFundsAmount(e.target.value)}
              />
              <p className="text-xs text-muted-foreground">Minimum amount: $0.50 USD</p>
            </div>
            {addFundsError && <p className="text-sm text-destructive">{addFundsError}</p>}
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowAddFunds(false)}
              disabled={addFundsLoading}
            >
              Cancel
            </Button>
            <Button onClick={handleAddFunds} disabled={addFundsLoading}>
              {addFundsLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Processing...
                </>
              ) : (
                'Continue to Checkout'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Tabs */}
      <Tabs defaultValue="transactions" className="space-y-4">
        <TabsList>
          <TabsTrigger value="transactions">Transactions</TabsTrigger>
          <TabsTrigger value="revenue">Revenue Breakdown</TabsTrigger>
          <TabsTrigger value="spending">Spending Breakdown</TabsTrigger>
        </TabsList>

        <TabsContent value="transactions">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <div>
                <CardTitle>Recent Transactions</CardTitle>
                <CardDescription>Your agent's financial activity</CardDescription>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-sm text-muted-foreground">
                  {txTotal > 0
                    ? `Showing ${Math.min(txOffset + 1, txTotal)}-${Math.min(txOffset + transactions.length, txTotal)} of ${txTotal}`
                    : 'No transactions'}
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setTxOffset(Math.max(0, txOffset - txLimit))}
                  disabled={txOffset === 0}
                >
                  Previous
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setTxOffset(txOffset + txLimit)}
                  disabled={txOffset + txLimit >= txTotal}
                >
                  Next
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              {transactions.length === 0 ? (
                <p className="text-sm text-muted-foreground py-8 text-center">
                  No transaction history available. Add funds to get started.
                </p>
              ) : (
                <div className="space-y-4">
                  {transactions.map((tx) => (
                    <div
                      key={tx.id}
                      className="flex items-center justify-between p-4 border rounded-lg"
                    >
                      <div className="flex items-center gap-4">
                        <div className="p-2 bg-muted rounded-lg">{getTransactionIcon(tx.type)}</div>
                        <div>
                          <p className="font-medium">{tx.description}</p>
                          <div className="flex items-center gap-2 mt-1">
                            {tx.counterparty && (
                              <span className="text-sm text-muted-foreground">
                                {tx.counterparty}
                              </span>
                            )}
                            <span className="text-xs text-muted-foreground">•</span>
                            <span className="text-xs text-muted-foreground">
                              {new Date(tx.timestamp).toLocaleString()}
                            </span>
                          </div>
                        </div>
                      </div>
                      <div className="flex items-center gap-4">
                        <span className={`font-bold ${getAmountColor(tx.type)}`}>
                          {getAmountPrefix(tx.type)}
                          {formatCurrency(tx.amount)}
                        </span>
                        {getStatusBadge(tx.status)}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="revenue">
          <Card>
            <CardHeader>
              <CardTitle>Revenue Sources</CardTitle>
              <CardDescription>Where your agent's earnings come from</CardDescription>
            </CardHeader>
            <CardContent>
              {revenueBreakdown.length === 0 ? (
                <p className="text-sm text-muted-foreground py-8 text-center">
                  No revenue data yet. Earnings will appear here as your agent earns.
                </p>
              ) : (
                <div className="space-y-4">
                  {revenueBreakdown.map((item) => (
                    <div key={item.category} className="space-y-2">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <DollarSign className="h-4 w-4 text-green-500" />
                          <span className="font-medium">{item.category}</span>
                        </div>
                        <div className="flex items-center gap-4">
                          <span className="text-green-600 font-bold">
                            {formatCurrency(item.amount)}
                          </span>
                          <span className="text-muted-foreground">({item.percentage}%)</span>
                        </div>
                      </div>
                      <Progress value={item.percentage} className="h-2" />
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="spending">
          <Card>
            <CardHeader>
              <CardTitle>Spending Categories</CardTitle>
              <CardDescription>Where your agent's funds are spent</CardDescription>
            </CardHeader>
            <CardContent>
              {spendingBreakdown.length === 0 ? (
                <p className="text-sm text-muted-foreground py-8 text-center">
                  No spending breakdown yet. Cost by function will appear when available.
                </p>
              ) : (
                <div className="space-y-4">
                  {spendingBreakdown.map((item) => (
                    <div key={item.category} className="space-y-2">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <CreditCard className="h-4 w-4 text-red-500" />
                          <span className="font-medium">{item.category}</span>
                        </div>
                        <div className="flex items-center gap-4">
                          <span className="text-red-600 font-bold">
                            {formatCurrency(item.amount)}
                          </span>
                          <span className="text-muted-foreground">({item.percentage}%)</span>
                        </div>
                      </div>
                      <Progress value={item.percentage} className="h-2" />
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

export default WalletDashboard;
