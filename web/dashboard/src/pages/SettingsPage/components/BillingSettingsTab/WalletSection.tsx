import type { WalletInfo } from '@/api/billing';
import { getWalletErrorMessage } from '@/api/billing';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Chamber, CornerBrace, FrameButton, SealedButton, StatusPill } from '@/components/sc';
import { Loader2, Wallet } from 'lucide-react';
import { useTranslation } from 'react-i18next';
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
  const { t } = useTranslation();
  return (
    <Chamber>
      <CornerBrace position="tl" />
      <CornerBrace position="br" />
      <div className="mb-4">
        <h3
          className="font-display flex items-center gap-2 text-lg font-semibold"
          style={{ color: 'var(--text)' }}
        >
          <Wallet className="h-5 w-5" style={{ color: 'var(--accent)' }} />
          {t('billingSettings.wallet.title')}
        </h3>
        <p className="text-sm mt-1" style={{ color: 'var(--text-dim)' }}>
          {t('billingSettings.wallet.description')}
        </p>
      </div>
      <div className="space-y-4">
        {walletLoading ? (
          <div
            className="flex items-center justify-center gap-2 p-4"
            style={{ color: 'var(--text-dim)' }}
          >
            <Loader2 className="h-5 w-5 animate-spin" />
            <span>{t('billingSettings.wallet.loading')}</span>
          </div>
        ) : walletError ? (
          <div
            className="flex items-center gap-2 p-4 rounded-lg"
            style={{
              background: 'rgba(232, 196, 104, 0.06)',
              border: '1px solid rgba(232, 196, 104, 0.3)',
            }}
          >
            <StatusPill status="pending" label="Warning" />
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
                  {t('billingSettings.wallet.balance')}
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
                  {t('billingSettings.wallet.lifetimeEarned')}
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
                  {t('billingSettings.wallet.feesPaid')}
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
                {t('billingSettings.wallet.addFunds')}
              </Label>
              <div className="flex flex-wrap gap-2">
                {WALLET_TOP_UP_PRESETS.map((n) => (
                  <FrameButton
                    key={n}
                    size="sm"
                    onClick={() => onTopUpAmountChange(String(n))}
                  >
                    ${n}
                  </FrameButton>
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
                {t('billingSettings.wallet.minMax', {
                  min: `$${MIN_WALLET_TOP_UP_USD.toFixed(2)}`,
                  max: `$${MAX_WALLET_TOP_UP_USD.toLocaleString()}`
                })}
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
            <SealedButton
              disabled={topUpSubmitting || !topUpAmountValid}
              onClick={onWalletTopUp}
              loading={topUpSubmitting}
              iconLeft={<Wallet className="h-4 w-4" />}
            >
              {t('billingSettings.wallet.buyCredits')}
            </SealedButton>
          </>
        )}
      </div>
    </Chamber>
  );
}
