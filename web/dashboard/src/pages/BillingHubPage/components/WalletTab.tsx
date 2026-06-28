import { useState } from 'react';
import {
  Chamber,
  SealedButton,
  FrameButton,
  StatusPill,
  GaugeStrip,
  Gauge,
  Input,
} from '@/components/containment';
import type { WalletInfo, WalletTransaction } from '@/api/billing';
import { getWalletErrorMessage } from '@/api/billing';
import { Wallet, Plus, ArrowUpRight, ArrowDownRight, RefreshCw, AlertCircle } from 'lucide-react';

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
    <div className="sc-billing-fade-in" style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
      {/* Credits */}
      <Chamber nested>
        <div className="sc-billing-card-header" style={{ margin: 'calc(-1 * var(--space-5))', marginBottom: 'var(--space-5)', padding: 'var(--space-4) var(--space-5)' }}>
          <div className="sc-billing-card-title">
            <Wallet style={{ width: 14, height: 14, color: 'var(--status-pending)' }} />
            Registry Credits
          </div>
          <div className="sc-billing-card-description">Prepaid balance for registry publish fees and platform charges</div>
        </div>

        {isLoading ? (
          <div className="sc-billing-grid sc-billing-grid-3">
            {[1, 2, 3].map((i) => (
              <div key={i} style={{ height: 64, background: 'var(--panel)', borderRadius: 'var(--radius)' }} />
            ))}
          </div>
        ) : error ? (
          <div className="sc-billing-info sc-billing-info-warning">
            <AlertCircle style={{ width: 18, height: 18 }} />
            <div className="sc-billing-info-content">
              <div className="sc-billing-info-text">{getWalletErrorMessage(error)}</div>
            </div>
          </div>
        ) : (
          <>
            <GaugeStrip>
              <Gauge data={{ value: formatUsd(walletInfo?.balance_usd ?? 0), label: 'Balance' }} isFirst />
              <Gauge data={{ value: formatUsd(walletInfo?.lifetime_earnings_usd ?? 0), label: 'Lifetime Earned' }} />
              <Gauge data={{ value: formatUsd(walletInfo?.lifetime_fees_usd ?? 0), label: 'Fees Paid' }} />
            </GaugeStrip>

            <div style={{ marginTop: 'var(--space-5)' }}>
              <label style={{ display: 'block', fontFamily: 'var(--font-mono)', fontSize: 11, fontWeight: 500, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--text-faint)', marginBottom: 'var(--space-2)' }}>
                Add funds (USD)
              </label>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 'var(--space-2)', marginBottom: 'var(--space-3)' }}>
                {WALLET_TOP_UP_PRESETS.map((n) => (
                  <FrameButton
                    key={n}
                    size="sm"
                    onClick={() => setTopUpAmountInput(String(n))}
                  >
                    ${n}
                  </FrameButton>
                ))}
              </div>
              <Input
                type="text"
                inputMode="decimal"
                value={topUpAmountInput}
                onChange={(e) => setTopUpAmountInput(e.target.value)}
                placeholder="25.00"
                style={{ maxWidth: 200 }}
              />
              <p style={{ fontSize: 11, color: 'var(--text-faint)', marginTop: 'var(--space-2)' }}>
                Minimum ${MIN_WALLET_TOP_UP_USD.toFixed(2)} · maximum ${MAX_WALLET_TOP_UP_USD.toLocaleString()} per top-up.
              </p>
            </div>

            <SealedButton
              loading={topUpSubmitting}
              disabled={!topUpAmountValid}
              onClick={handleTopUpClick}
              iconLeft={<Plus style={{ width: 14, height: 14 }} />}
            >
              Buy credits
            </SealedButton>
          </>
        )}
      </Chamber>

      {/* Transaction History */}
      <Chamber nested>
        <div className="sc-billing-card-header" style={{ margin: 'calc(-1 * var(--space-5))', marginBottom: 'var(--space-5)', padding: 'var(--space-4) var(--space-5)' }}>
          <div className="sc-billing-card-title">
            <RefreshCw style={{ width: 14, height: 14 }} />
            Transaction History
          </div>
        </div>

        {walletTransactions.length === 0 ? (
          <div className="empty-state" style={{ minHeight: 120, flexDirection: 'column', gap: 'var(--space-3)' }}>
            <RefreshCw style={{ width: 48, height: 48, color: 'var(--text-faint)' }} />
            <p style={{ color: 'var(--text-faint)', fontFamily: 'var(--font-mono)', fontSize: 11, fontWeight: 500, textTransform: 'uppercase', letterSpacing: '0.06em' }}>No transactions yet</p>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
            {walletTransactions.slice(0, 10).map((tx) => {
              const isCredit = tx.type === 'credit' || tx.type === 'top_up';
              return (
                <div
                  key={tx.id}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    padding: 'var(--space-3)',
                    borderRadius: 'var(--radius)',
                    background: 'var(--panel)',
                    border: '1px solid var(--panel-edge)',
                  }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
                    <div style={{
                      width: 32, height: 32, borderRadius: 'var(--radius)',
                      display: 'flex', alignItems: 'center', justifyContent: 'center',
                      background: isCredit ? 'rgba(143, 255, 208, 0.1)' : 'rgba(255, 107, 107, 0.1)',
                      color: isCredit ? 'var(--status-ok)' : 'var(--status-revoked)',
                    }}>
                      {isCredit ? <ArrowUpRight style={{ width: 14, height: 14 }} /> : <ArrowDownRight style={{ width: 14, height: 14 }} />}
                    </div>
                    <div>
                      <p style={{ fontSize: 13, fontWeight: 500, color: 'var(--text)' }}>{tx.description}</p>
                      <p style={{ fontSize: 11, color: 'var(--text-dim)' }}>{formatDate(tx.timestamp)}</p>
                    </div>
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
                    <span style={{ fontFamily: 'var(--font-mono)', fontSize: 13, fontWeight: 600, color: isCredit ? 'var(--status-ok)' : 'var(--status-revoked)' }}>
                      {isCredit ? '+' : '-'}{formatUsd(tx.amount)}
                    </span>
                    <StatusPill status={tx.status === 'completed' ? 'live' : tx.status === 'pending' ? 'pending' : 'revoked'} label={tx.status} />
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </Chamber>
    </div>
  );
}
