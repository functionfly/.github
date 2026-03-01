import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { 
  Wallet, 
  TrendingUp, 
  TrendingDown, 
  ArrowUpRight,
  ArrowDownRight,
  Clock,
  DollarSign,
  CreditCard,
  History,
  Shield,
  Plus,
  Download,
  Filter
} from 'lucide-react';

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

  useEffect(() => {
    // Mock data - in production fetch from API
    setTimeout(() => {
      setWallet({
        balance: 1250.00,
        escrowBalance: 250.00,
        totalEarned: 3420.50,
        totalSpent: 2170.50,
        lastEarning: '2026-03-01T14:30:00Z',
        lastSpending: '2026-03-01T12:15:00Z'
      });
      setTransactions([
        {
          id: '1',
          type: 'credit',
          amount: 25.00,
          description: 'Function call payment',
          counterparty: 'data-processor-agent',
          timestamp: '2026-03-01T14:30:00Z',
          status: 'completed'
        },
        {
          id: '2',
          type: 'debit',
          amount: 5.00,
          description: 'Function execution',
          counterparty: 'json-transform-fn',
          timestamp: '2026-03-01T12:15:00Z',
          status: 'completed'
        },
        {
          id: '3',
          type: 'credit',
          amount: 50.00,
          description: 'Delegation revenue share',
          counterparty: 'analytics-manager',
          timestamp: '2026-03-01T10:00:00Z',
          status: 'completed'
        },
        {
          id: '4',
          type: 'transfer',
          amount: 100.00,
          description: 'Escrow deposit',
          timestamp: '2026-02-28T16:45:00Z',
          status: 'completed'
        },
        {
          id: '5',
          type: 'credit',
          amount: 15.00,
          description: 'Subscription payment',
          timestamp: '2026-02-28T09:00:00Z',
          status: 'completed'
        }
      ]);
      setRevenueBreakdown([
        { category: 'Function Calls', amount: 1520.50, percentage: 44 },
        { category: 'Delegation Revenue', amount: 1200.00, percentage: 35 },
        { category: 'Subscriptions', amount: 500.00, percentage: 15 },
        { category: 'Other', amount: 200.00, percentage: 6 }
      ]);
      setSpendingBreakdown([
        { category: 'Function Execution', amount: 1200.00, percentage: 55 },
        { category: 'Child Agent Payments', amount: 700.00, percentage: 32 },
        { category: 'Escrow', amount: 270.50, percentage: 13 }
      ]);
      setLoading(false);
    }, 800);
  }, [agentId]);

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
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
      failed: 'bg-red-500'
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
                <p className="text-3xl font-bold text-green-600">{formatCurrency(wallet?.balance ?? 0)}</p>
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
                <p className="text-3xl font-bold text-green-600">{formatCurrency(wallet?.totalEarned ?? 0)}</p>
              </div>
              <TrendingUp className="h-8 w-8 text-green-500" />
            </div>
            <p className="text-xs text-muted-foreground mt-2">
              Since: {new Date(wallet?.lastEarning ?? '').toLocaleDateString()}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Total Spent</p>
                <p className="text-3xl font-bold text-red-600">{formatCurrency(wallet?.totalSpent ?? 0)}</p>
              </div>
              <TrendingDown className="h-8 w-8 text-red-500" />
            </div>
            <p className="text-xs text-muted-foreground mt-2">
              Since: {new Date(wallet?.lastSpending ?? '').toLocaleDateString()}
            </p>
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
                <CardDescription>
                  Your agent's financial activity
                </CardDescription>
              </div>
              <Button variant="outline" size="sm">
                <Filter className="h-4 w-4 mr-2" />
                Filter
              </Button>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {transactions.map((tx) => (
                  <div 
                    key={tx.id} 
                    className="flex items-center justify-between p-4 border rounded-lg"
                  >
                    <div className="flex items-center gap-4">
                      <div className="p-2 bg-muted rounded-lg">
                        {getTransactionIcon(tx.type)}
                      </div>
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
                      <span className={`font-bold ${
                        tx.type === 'credit' ? 'text-green-500' : 
                        tx.type === 'debit' ? 'text-red-500' : 'text-blue-500'
                      }`}>
                        {tx.type === 'credit' ? '+' : tx.type === 'debit' ? '-' : ''}
                        {formatCurrency(tx.amount)}
                      </span>
                      {getStatusBadge(tx.status)}
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="revenue">
          <Card>
            <CardHeader>
              <CardTitle>Revenue Sources</CardTitle>
              <CardDescription>
                Where your agent's earnings come from
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {revenueBreakdown.map((item) => (
                  <div key={item.category} className="space-y-2">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <DollarSign className="h-4 w-4 text-green-500" />
                        <span className="font-medium">{item.category}</span>
                      </div>
                      <div className="flex items-center gap-4">
                        <span className="text-green-600 font-bold">{formatCurrency(item.amount)}</span>
                        <span className="text-muted-foreground">({item.percentage}%)</span>
                      </div>
                    </div>
                    <Progress value={item.percentage} className="h-2" />
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="spending">
          <Card>
            <CardHeader>
              <CardTitle>Spending Categories</CardTitle>
              <CardDescription>
                Where your agent's funds are spent
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {spendingBreakdown.map((item) => (
                  <div key={item.category} className="space-y-2">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <CreditCard className="h-4 w-4 text-red-500" />
                        <span className="font-medium">{item.category}</span>
                      </div>
                      <div className="flex items-center gap-4">
                        <span className="text-red-600 font-bold">{formatCurrency(item.amount)}</span>
                        <span className="text-muted-foreground">({item.percentage}%)</span>
                      </div>
                    </div>
                    <Progress value={item.percentage} className="h-2" />
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

export default WalletDashboard;
