import type { WalletInfo } from '@/api/billing';
import { getWalletErrorMessage } from '@/api/billing';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Loader2, Wallet } from 'lucide-react';
import {
  MAX_WALLET_TOP_UP_USD,
  MIN_WALLET_TOP_UP_USD,
  WALLET_TOP_UP_PRESETS,
  formatUsd,
} from './WalletSection.constants';

export { MAX_WALLET_TOP_UP_USD, MIN_WALLET_TOP_UP_USD };

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
    <div
      className="rounded-lg p-5"
      style={{
        background: 'var(--panel)',
        border: '1px solid var(--panel-edge)',
        boxShadow: 'var(--shadow-chamber)',
      }}
    >
      <div className="mb-4">
        <h3
          className="font-display flex items-center gap-2 text-lg font-semibold"
          style={{ color: 'var(--text)' }}
        >
          <Wallet className="h-5 w-5" style={{ color: 'var(--accent)' }} />
          Registry credits
        </h3>
        <p className="text-sm mt-1" style={{ color: 'var(--text-dim)' }}>
          Prepaid balance for registry publish fees and platform charges. Top up with a card via
          Stripe (test or live keys on the server).
        </p>
      </div>
      <div className="space-y-4">
        {walletLoading ? (
          <div
            className="flex items-center justify-center gap-2 p-4"
            style={{ color: 'var(--text-dim)' }}
          >
            <Loader2 className="h-5 w-5 animate-spin" />
            <span>Loading balance…</span>
          </div>
        ) : walletError ? (
          <div
            className="flex items-center gap-2 p-4 rounded-lg"
            style={{
              background: 'rgba(232, 196, 104, 0.06)',
              border: '1px solid rgba(232, 196, 104, 0.3)',
            }}
          >
            <span className="w-5 h-5 shrink-0" style={{ color: 'var(--status-pending)' }}>
              ⚠️
            </span>
            <p className="text-sm" style={{ color: 'var(--status-pending)' }}>
              {getWalletErrorMessage(walletError)}
            </p>
          </div>
        ) : (
          <>
            <div
              className="grid gap-3 sm:grid-cols-3 rounded-lg p-4"
              style={{ background: 'var(--panel-raised)', border: '1px solid var(--panel-edge)' }}
            >
              <div>
                <p
                  className="text-xs uppercase tracking-wide"
                  style={{ color: 'var(--text-faint)' }}
                >
                  Balance
                </p>
                <p
                  className="text-lg font-semibold font-mono tabular-nums"
                  style={{ color: 'var(--status-pending)' }}
                >
                  {formatUsd(walletData?.balance_usd ?? 0)}
                </p>
              </div>
              <div>
                <p
                  className="text-xs uppercase tracking-wide"
                  style={{ color: 'var(--text-faint)' }}
                >
                  Lifetime earned
                </p>
                <p
                  className="text-lg font-medium font-mono tabular-nums"
                  style={{ color: 'var(--text)' }}
                >
                  {formatUsd(walletData?.lifetime_earnings_usd ?? 0)}
                </p>
              </div>
              <div>
                <p
                  className="text-xs uppercase tracking-wide"
                  style={{ color: 'var(--text-faint)' }}
                >
                  Fees paid
                </p>
                <p
                  className="text-lg font-medium font-mono tabular-nums"
                  style={{ color: 'var(--text)' }}
                >
                  {formatUsd(walletData?.lifetime_fees_usd ?? 0)}
                </p>
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="wallet-top-up-amount" style={{ color: 'var(--text)' }}>
                Add funds (USD)
              </Label>
              <div className="flex flex-wrap gap-2">
                {WALLET_TOP_UP_PRESETS.map((n) => (
                  <Button
                    key={n}
                    type="button"
                    variant="outline"
                    size="sm"
                    className="tabular-nums transition-colors duration-150"
                    style={{ borderColor: 'var(--steel)', color: 'var(--text)' }}
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
              <p className="text-xs" style={{ color: 'var(--text-faint)' }}>
                Minimum ${MIN_WALLET_TOP_UP_USD.toFixed(2)} · maximum $
                {MAX_WALLET_TOP_UP_USD.toLocaleString()} per top-up.
                {import.meta.env.DEV && (
                  <>
                    {' '}
                    Local dev: run{' '}
                    <code
                      className="rounded px-1 py-0.5 text-[11px]"
                      style={{ background: 'var(--panel-raised)' }}
                    >
                      stripe listen --forward-to localhost:8080/v1/webhooks/stripe
                    </code>{' '}
                    and match{' '}
                    <code
                      className="rounded px-1 py-0.5 text-[11px]"
                      style={{ background: 'var(--panel-raised)' }}
                    >
                      STRIPE_WEBHOOK_SECRET
                    </code>{' '}
                    to the CLI signing secret so balance updates after payment.
                  </>
                )}
              </p>
            </div>
            <Button
              type="button"
              disabled={topUpSubmitting || !topUpAmountValid}
              onClick={onWalletTopUp}
              style={{
                background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
                color: 'var(--text-on-light)',
                boxShadow: 'var(--shadow-btn-primary-rest)',
              }}
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
      </div>
    </div>
  );
}
