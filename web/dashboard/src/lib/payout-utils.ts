/** React Query keys for payout APIs (invalidate from multiple surfaces). */
export const payoutQueryKeys = {
  connect: ['payouts', 'connect'] as const,
  balance: ['payouts', 'balance'] as const,
  requests: ['payouts', 'requests'] as const,
  ledger: ['payouts', 'ledger'] as const,
};

export const MIN_PAYOUT_USD = 10;
export const MAX_PAYOUT_USD = 50_000;

export function formatPayoutUsd(amount: number): string {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(amount);
}

export function formatPayoutCentsSigned(cents: number, currency: string = 'usd'): string {
  const sign = cents < 0 ? '-' : '';
  const abs = Math.abs(cents);
  return (
    sign +
    new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency.toUpperCase(),
    }).format(abs / 100)
  );
}

export function getPayoutApiErrorMessage(err: unknown): string {
  const res = (err as { response?: { data?: unknown; status?: number } })?.response;
  const data = res?.data;
  if (data && typeof data === 'object' && data !== null && 'error' in data) {
    const e = (data as { error?: unknown }).error;
    if (typeof e === 'string' && e) return e;
    if (e && typeof e === 'object' && 'message' in e) {
      const m = (e as { message?: unknown }).message;
      if (typeof m === 'string' && m) return m;
    }
  }
  return 'Something went wrong.';
}

export function payoutStatusBadgeVariant(
  status: string
): 'default' | 'secondary' | 'destructive' | 'outline' {
  const s = status.toLowerCase();
  if (s === 'completed' || s === 'active') return 'default';
  if (s === 'failed' || s === 'disabled' || s === 'restricted') return 'destructive';
  if (s === 'processing' || s === 'pending' || s === 'onboarding') return 'secondary';
  return 'outline';
}

export function payoutEntryTypeLabel(type: string): string {
  switch (type) {
    case 'earning_credit':
      return 'Earning';
    case 'payout_debit':
      return 'Payout';
    case 'payout_reversal':
      return 'Reversal';
    case 'adjustment':
      return 'Adjustment';
    default:
      return type.replace(/_/g, ' ');
  }
}

/** Integer cents from dollar string; null if invalid or out of bounds. */
export function parseUsdToPayoutCents(
  raw: string,
  availableUsd: number
): { ok: true; cents: number } | { ok: false } {
  const trimmed = raw.replace(/,/g, '').trim();
  if (!/^\d+(\.\d{1,2})?$/.test(trimmed)) {
    return { ok: false };
  }
  const dollars = parseFloat(trimmed);
  if (Number.isNaN(dollars)) return { ok: false };
  const cents = Math.round(dollars * 100);
  const minCents = MIN_PAYOUT_USD * 100;
  const maxCents = MAX_PAYOUT_USD * 100;
  const availableCents = Math.round(availableUsd * 100);
  if (cents < minCents || cents > maxCents) return { ok: false };
  if (cents > availableCents) return { ok: false };
  return { ok: true, cents };
}
