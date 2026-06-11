import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import type { WalletInfo, WalletTransaction } from '@/api/billing';
import { getWalletErrorMessage } from '@/api/billing';
import { Wallet, Plus, Loader2, ArrowUpRight, ArrowDownRight, RefreshCw } from 'lucide-react';
import { useState } from 'react';

interface WalletTabProps {
  walletInfo: WalletInfo | null;
  walletTransactions: WalletTransaction[];
  isLoading: boolean;
  error: Error | null;
  onTopUp: (amountUsd: number) => Promise<void>;
}

const WALLET_TOP_UP_PRESETS = [10, 25, 50, 100] as const;
const MIN_WALLET_TOP_UP_USD = 1;
const MAX_WALLET_TOP_UP_USD = 10_000;

function formatUsd(amount: number): string {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(amount);
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });
}

export function WalletTab({
  walletInfo,
  walletTransactions,
  isLoading,
  error,
  onTopUp,
}: WalletTabProps) {
  const [topUpAmountInput, setTopUpAmountInput] = useState('25');
  const [topUpSubmitting, setTopUpSubmitting] = useState(false);

  const parsedTopUpUsd = parseFloat(topUpAmountInput.replace(/,/g, ''));
  const topUpAmountValid =
    !Number.isNaN(parsedTopUpUsd) &&
    parsedTopUpUsd >= MIN_WALLET_TOP_UP_USD &&
    parsedTopUpUsd <= MAX_WALLET_TOP_UP_USD;

  const handleTopUpClick = async () => {
    if (!topUpAmountValid) return;
    setTopUpSubmitting(true);
    try {
      await onTopUp(parsedTopUpUsd);
    } finally {
      setTopUpSubmitting(false);
    }
  };

  return (
    <div className="space-y-6">
      <Card className="ff-card-velocity">
        <CardHeader>
          <CardTitle className="font-display flex items-center gap-2">
            <Wallet className="h-5 w-5 text-amber-500" />
            Registry Credits
          </CardTitle>
          <CardDescription>
            Prepaid balance for registry publish fees and platform charges
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {isLoading ? (
            <div className="grid grid-cols-3 gap-4">
              <Skeleton className="h-16 w-full" />
              <Skeleton className="h-16 w-full" />
              <Skeleton className="h-16 w-full" />
            </div>
          ) : error ? (
            <div className="flex items-center gap-2 p-4 rounded-lg bg-amber-500/10 border border-amber-500/20">
              <span className="w-5 h-5 text-amber-500 shrink-0">⚠️</span>
              <p className="text-amber-500 text-sm">{getWalletErrorMessage(error)}</p>
            </div>
          ) : (
            <>
              <div className="grid grid-cols-3 gap-4 rounded-lg bg-bg-secondary border border-border-default p-4">
                <div>
                  <p className="text-xs text-text-muted uppercase tracking-wide">Balance</p>
                  <p className="text-2xl font-bold font-mono text-amber-500">
                    {formatUsd(walletInfo?.balance_usd ?? 0)}
                  </p>
                </div>
                <div>
                  <p className="text-xs text-text-muted uppercase tracking-wide">Lifetime Earned</p>
                  <p className="text-lg font-medium font-mono">
                    {formatUsd(walletInfo?.lifetime_earnings_usd ?? 0)}
                  </p>
                </div>
                <div>
                  <p className="text-xs text-text-muted uppercase tracking-wide">Fees Paid</p>
                  <p className="text-lg font-medium font-mono">
                    {formatUsd(walletInfo?.lifetime_fees_usd ?? 0)}
                  </p>
                </div>
              </div>

              <div className="space-y-3">
                <Label htmlFor="wallet-top-up-amount">Add funds (USD)</Label>
                <div className="flex flex-wrap gap-2">
                  {WALLET_TOP_UP_PRESETS.map((n) => (
                    <Button
                      key={n}
                      type="button"
                      variant="outline"
                      size="sm"
                      className="border-border-strong tabular-nums hover:border-brand-500 hover:bg-brand-500/10 hover:text-brand-400 transition-colors duration-150"
                      onClick={() => setTopUpAmountInput(String(n))}
                    >
                      ${n}
                    </Button>
                  ))}
                </div>
                <Input
                  id="wallet-top-up-amount"
                  type="text"
                  inputMode="decimal"
                  autoComplete="transaction-amount"
                  value={topUpAmountInput}
                  onChange={(e) => setTopUpAmountInput(e.target.value)}
                  placeholder="25.00"
                  className="max-w-[200px]"
                />
                <p className="text-xs text-text-muted">
                  Minimum ${MIN_WALLET_TOP_UP_USD.toFixed(2)} · maximum $
                  {MAX_WALLET_TOP_UP_USD.toLocaleString()} per top-up.
                </p>
              </div>

              <Button
                type="button"
                className="ff-btn-velocity"
                disabled={topUpSubmitting || !topUpAmountValid}
                onClick={handleTopUpClick}
              >
                {topUpSubmitting ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Redirecting to checkout…
                  </>
                ) : (
                  <>
                    <Plus className="mr-2 h-4 w-4" />
                    Buy credits
                  </>
                )}
              </Button>
            </>
          )}
        </CardContent>
      </Card>

      <Card className="ff-card-velocity">
        <CardHeader>
          <CardTitle className="font-display flex items-center gap-2">
            <RefreshCw className="h-5 w-5 text-brand-500" />
            Transaction History
          </CardTitle>
        </CardHeader>
        <CardContent>
          {walletTransactions.length === 0 ? (
            <div className="text-center py-8">
              <RefreshCw className="h-12 w-12 text-text-muted mx-auto mb-3" />
              <p className="text-text-muted">No transactions yet</p>
            </div>
          ) : (
            <div className="space-y-2">
              {walletTransactions.slice(0, 10).map((tx) => (
                <div
                  key={tx.id}
                  className="flex items-center justify-between p-3 rounded-lg bg-bg-secondary border border-border-default"
                >
                  <div className="flex items-center gap-3">
                    <div
                      className={`w-8 h-8 rounded-lg flex items-center justify-center ${
                        tx.type === 'credit' || tx.type === 'top_up'
                          ? 'bg-green-500/20 text-green-400'
                          : 'bg-red-500/20 text-red-400'
                      }`}
                    >
                      {tx.type === 'credit' || tx.type === 'top_up' ? (
                        <ArrowUpRight className="h-4 w-4" />
                      ) : (
                        <ArrowDownRight className="h-4 w-4" />
                      )}
                    </div>
                    <div>
                      <p className="text-sm font-medium">{tx.description}</p>
                      <p className="text-xs text-text-muted">{formatDate(tx.timestamp)}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <span
                      className={`font-mono font-medium ${
                        tx.type === 'credit' || tx.type === 'top_up'
                          ? 'text-green-400'
                          : 'text-red-400'
                      }`}
                    >
                      {tx.type === 'credit' || tx.type === 'top_up' ? '+' : '-'}
                      {formatUsd(tx.amount)}
                    </span>
                    <Badge
                      variant={tx.status === 'completed' ? 'success' : 'secondary'}
                      className={tx.status === 'completed' ? 'ff-badge-success' : ''}
                    >
                      {tx.status}
                    </Badge>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}