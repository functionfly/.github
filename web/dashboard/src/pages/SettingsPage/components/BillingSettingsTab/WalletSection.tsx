import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Loader2, Wallet } from 'lucide-react';
import type { WalletInfo } from '@/api/billing';
import { getWalletErrorMessage } from '@/api/billing';

const WALLET_TOP_UP_PRESETS = [10, 25, 50, 100] as const;
const MIN_WALLET_TOP_UP_USD = 1;
const MAX_WALLET_TOP_UP_USD = 10_000;

function formatUsd(amount: number): string {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(amount);
}

interface WalletSectionProps {
  walletData: WalletInfo | undefined;
  walletLoading: boolean;
  walletError: Error | null;
  topUpAmountInput: string;
  topUpSubmitting: boolean;
  topUpAmountValid: boolean;
  onTopUpAmountChange: (value: string) => void;
  onWalletTopUp: () => void;
}

export function WalletSection({
  walletData,
  walletLoading,
  walletError,
  topUpAmountInput,
  topUpSubmitting,
  topUpAmountValid,
  onTopUpAmountChange,
  onWalletTopUp,
}: WalletSectionProps) {
  return (
    <Card className="ff-card-velocity">
      <CardHeader>
        <CardTitle className="font-display flex items-center gap-2">
          <Wallet className="h-5 w-5 text-brand-500" />
          Registry credits
        </CardTitle>
        <CardDescription className="text-text-secondary">
          Prepaid balance for registry publish fees and platform charges. Top up with a card via
          Stripe (test or live keys on the server).
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {walletLoading ? (
          <div className="flex items-center justify-center gap-2 p-4 text-text-muted">
            <Loader2 className="h-5 w-5 animate-spin" />
            <span>Loading balance…</span>
          </div>
        ) : walletError ? (
          <div className="flex items-center gap-2 p-4 rounded-lg bg-amber-500/10 border border-amber-500/20">
            <span className="w-5 h-5 text-amber-500 shrink-0">⚠️</span>
            <p className="text-amber-500 text-sm">{getWalletErrorMessage(walletError)}</p>
          </div>
        ) : (
          <>
            <div className="grid gap-3 sm:grid-cols-3 rounded-lg bg-bg-secondary border border-border-default p-4">
              <div>
                <p className="text-xs text-text-muted uppercase tracking-wide">Balance</p>
                <p className="text-lg font-semibold font-mono text-amber-500 tabular-nums">
                  {formatUsd(walletData?.balance_usd ?? 0)}
                </p>
              </div>
              <div>
                <p className="text-xs text-text-muted uppercase tracking-wide">Lifetime earned</p>
                <p className="text-lg font-medium font-mono text-text-primary tabular-nums">
                  {formatUsd(walletData?.lifetime_earnings_usd ?? 0)}
                </p>
              </div>
              <div>
                <p className="text-xs text-text-muted uppercase tracking-wide">Fees paid</p>
                <p className="text-lg font-medium font-mono text-text-primary tabular-nums">
                  {formatUsd(walletData?.lifetime_fees_usd ?? 0)}
                </p>
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="wallet-top-up-amount">Add funds (USD)</Label>
              <div className="flex flex-wrap gap-2">
                {WALLET_TOP_UP_PRESETS.map((n) => (
                  <Button
                    key={n}
                    type="button"
                    variant="outline"
                    size="sm"
                    className="border-border-strong tabular-nums hover:border-brand-500 hover:bg-brand-500/10 hover:text-brand-400 transition-colors duration-150"
                    onClick={() => onTopUpAmountChange(String(n))}
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
                onChange={(e) => onTopUpAmountChange(e.target.value)}
                placeholder="25.00"
                className="max-w-[200px]"
              />
              <p className="text-xs text-text-muted">
                Minimum ${MIN_WALLET_TOP_UP_USD.toFixed(2)} · maximum $
                {MAX_WALLET_TOP_UP_USD.toLocaleString()} per top-up.
                {import.meta.env.DEV && (
                  <>
                    {' '}
                    Local dev: run{' '}
                    <code className="rounded bg-bg-tertiary px-1 py-0.5 text-[11px]">
                      stripe listen --forward-to localhost:8080/v1/webhooks/stripe
                    </code>{' '}
                    and match{' '}
                    <code className="rounded bg-bg-tertiary px-1 py-0.5 text-[11px]">
                      STRIPE_WEBHOOK_SECRET
                    </code>{' '}
                    to the CLI signing secret so balance updates after payment.
                  </>
                )}
              </p>
            </div>
            <Button
              type="button"
              className="ff-btn-velocity"
              disabled={topUpSubmitting || !topUpAmountValid}
              onClick={onWalletTopUp}
            >
              {topUpSubmitting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Redirecting to checkout…
                </>
              ) : (
                <>
                  <Wallet className="mr-2 h-4 w-4" />
                  Buy credits
                </>
              )}
            </Button>
          </>
        )}
      </CardContent>
    </Card>
  );
}

export { WALLET_TOP_UP_PRESETS, MIN_WALLET_TOP_UP_USD, MAX_WALLET_TOP_UP_USD, formatUsd };