import { getWalletInfo, topUpWallet } from '@/api/billing';
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
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { CreditCard, DollarSign, Loader2, Plus, TrendingUp, Wallet } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';

interface WalletInfo {
  balance_usd: number;
  lifetime_earnings_usd: number;
  lifetime_fees_usd: number;
  user_id: string;
}

export function PlatformWalletPage() {
  const queryClient = useQueryClient();
  const [searchParams, setSearchParams] = useSearchParams();
  const [showAddFunds, setShowAddFunds] = useState(false);
  const [addFundsAmount, setAddFundsAmount] = useState('10.00');
  const [addFundsLoading, setAddFundsLoading] = useState(false);

  const {
    data: wallet,
    isLoading,
    error,
  } = useQuery<WalletInfo>({
    queryKey: ['platform-wallet'],
    queryFn: getWalletInfo,
    retry: 2,
  });

  // Handle return from Stripe checkout
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
      queryClient.invalidateQueries({ queryKey: ['platform-wallet'] });
    } else if (walletTopUp === 'cancel') {
      toast.info('Payment cancelled', {
        description: 'No funds were added to your wallet.',
      });
    }
  }, [searchParams, setSearchParams, queryClient]);

  const handleAddFunds = async () => {
    const amount = parseFloat(addFundsAmount);
    if (isNaN(amount) || amount < 1) {
      toast.error('Invalid amount', { description: 'Minimum top-up is $1.00' });
      return;
    }
    if (amount > 10000) {
      toast.error('Invalid amount', { description: 'Maximum top-up is $10,000.00' });
      return;
    }

    setAddFundsLoading(true);
    try {
      const successUrl = `${window.location.origin}/platform-wallet?walletTopUp=success`;
      const cancelUrl = `${window.location.origin}/platform-wallet?walletTopUp=cancel`;

      const result = await topUpWallet(amount, successUrl, cancelUrl);

      if (result.checkout_url) {
        window.location.href = result.checkout_url;
      } else {
        toast.error('Failed to start checkout');
      }
    } catch (e) {
      toast.error('Failed to initiate top-up', {
        description: e instanceof Error ? e.message : 'Please try again',
      });
    } finally {
      setAddFundsLoading(false);
    }
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(amount);
  };

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center gap-3 p-12 text-muted-foreground">
        <Loader2 className="h-8 w-8 animate-spin" />
        <p className="text-sm">Loading wallet…</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="mx-auto max-w-md space-y-4 rounded-lg border border-border p-6">
        <h1 className="text-xl font-semibold">Platform Wallet</h1>
        <p className="text-sm text-muted-foreground">
          Failed to load wallet. Please try again later.
        </p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-4xl space-y-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold flex items-center gap-2">
            <Wallet className="h-6 w-6" />
            Platform Wallet
          </h1>
          <p className="text-muted-foreground">Manage your platform credits and balance</p>
        </div>
        <Button onClick={() => setShowAddFunds(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Add Funds
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Current Balance</CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatCurrency(wallet?.balance_usd ?? 0)}</div>
            <p className="text-xs text-muted-foreground">Available for platform fees</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Lifetime Earnings</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatCurrency(wallet?.lifetime_earnings_usd ?? 0)}
            </div>
            <p className="text-xs text-muted-foreground">Total earnings from registry</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Lifetime Fees</CardTitle>
            <CreditCard className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatCurrency(wallet?.lifetime_fees_usd ?? 0)}
            </div>
            <p className="text-xs text-muted-foreground">Total platform fees paid</p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>About Your Wallet</CardTitle>
          <CardDescription>
            Your platform wallet is used for registry fees, function publishing, and other platform charges.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="rounded-lg border bg-muted/50 p-4">
            <h4 className="font-medium mb-2">How it works</h4>
            <ul className="list-disc list-inside space-y-1 text-sm text-muted-foreground">
              <li>Add funds to your wallet to pay for platform fees</li>
              <li>Wallet balance is used for registry publishing and updates</li>
              <li>Refunds can be processed within 30 days of purchase</li>
              <li>Your wallet balance never expires</li>
            </ul>
          </div>
        </CardContent>
      </Card>

      {/* Add Funds Dialog */}
      <Dialog open={showAddFunds} onOpenChange={setShowAddFunds}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Add Funds to Wallet</DialogTitle>
            <DialogDescription>
              Add credits to your platform wallet. Minimum $1.00, maximum $10,000.00.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="amount">Amount (USD)</Label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground">$</span>
                <Input
                  id="amount"
                  type="number"
                  min="1"
                  max="10000"
                  step="0.01"
                  value={addFundsAmount}
                  onChange={(e) => setAddFundsAmount(e.target.value)}
                  className="pl-7"
                  placeholder="10.00"
                />
              </div>
            </div>
            <div className="flex gap-2">
              {['10', '25', '50', '100'].map((amt) => (
                <Button
                  key={amt}
                  variant="outline"
                  size="sm"
                  onClick={() => setAddFundsAmount(amt)}
                  className={addFundsAmount === amt ? 'border-primary' : ''}
                >
                  ${amt}
                </Button>
              ))}
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowAddFunds(false)} disabled={addFundsLoading}>
              Cancel
            </Button>
            <Button onClick={handleAddFunds} disabled={addFundsLoading}>
              {addFundsLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Processing...
                </>
              ) : (
                'Proceed to Checkout'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export default PlatformWalletPage;
