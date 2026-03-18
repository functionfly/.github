import { agentApi } from '@/api/agent';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  ArrowDownRight,
  ArrowUpRight,
  CreditCard,
  DollarSign,
  Download,
  Filter,
  Plus,
  Shield,
  TrendingDown,
  TrendingUp,
  Wallet,
} from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';

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
  type: 'credit' | 'debit' | 'transfer';
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
  const [wallet, setWallet] = useState<WalletData | null>(null);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [revenueBreakdown, setRevenueBreakdown] = useState<RevenueBreakdown[]>([]);
  const [spendingBreakdown, setSpendingBreakdown] = useState<SpendingBreakdown[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    if (!agentId) {
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const [walletRes, costRes] = await Promise.allSettled([
        agentApi.getWallet(agentId),
        agentApi.getCostBreakdown(agentId),
      ]);

      const w = walletRes.status === 'fulfilled' ? walletRes.value?.wallet : null;
      const walletObj = w != null ? (w as unknown as Record<string, unknown>) : null;
      if (walletObj) {
        const balance = Number(walletObj.balance_usd ?? walletObj.balanceUSD ?? 0);
        const escrow = Number(walletObj.escrow_balance_usd ?? walletObj.escrowBalanceUSD ?? 0);
        const earned = Number(walletObj.total_earned_usd ?? walletObj.totalEarnedUSD ?? 0);
        const spent = Number(walletObj.total_spent_usd ?? walletObj.totalSpentUSD ?? 0);
        setWallet({
          balance,
          escrowBalance: escrow,
          totalEarned: earned,
          totalSpent: spent,
          lastEarning: '',
          lastSpending: '',
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

      setTransactions([]);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load wallet data');
      setWallet(null);
      setTransactions([]);
      setRevenueBreakdown([]);
      setSpendingBreakdown([]);
    } finally {
      setLoading(false);
    }
  }, [agentId]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(amount);
  };

  const getTransactionIcon = (type: string) => {
    switch (type) {
      case 'credit':
        return <ArrowDownRight className="h-4 w-4 text-green-500" />;
      case 'debit':
        return <ArrowUpRight className="h-4 w-4 text-red-500" />;
      case 'transfer':
        return <TrendingUp className="h-4 w-4 text-blue-500" />;
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
          <Button variant="outline">
            <Plus className="h-4 w-4 mr-2" />
            Add Funds
          </Button>
          <Button variant="outline">
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
              <Button variant="outline" size="sm">
                <Filter className="h-4 w-4 mr-2" />
                Filter
              </Button>
            </CardHeader>
            <CardContent>
              {transactions.length === 0 ? (
                <p className="text-sm text-muted-foreground py-8 text-center">
                  No transaction history available. Transactions will appear here when the API
                  supports listing them.
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
                        <span
                          className={`font-bold ${
                            tx.type === 'credit'
                              ? 'text-green-500'
                              : tx.type === 'debit'
                                ? 'text-red-500'
                                : 'text-blue-500'
                          }`}
                        >
                          {tx.type === 'credit' ? '+' : tx.type === 'debit' ? '-' : ''}
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
