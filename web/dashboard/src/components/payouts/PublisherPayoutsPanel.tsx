import {
  cancelPayout,
  getConnectAccountStatus,
  getPayoutBalance,
  listPayoutLedger,
  listPayoutRequests,
  refreshConnectAccount,
  requestPayout,
  startConnectOnboarding,
  type ConnectAccountStatus,
  type PayoutBalance,
  type PayoutLedgerEntry,
  type PayoutRequest,
} from '@/api/payouts';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { ROUTES } from '@/lib/constants';
import {
  formatPayoutCentsSigned,
  formatPayoutUsd,
  getPayoutApiErrorMessage,
  MAX_PAYOUT_USD,
  MIN_PAYOUT_USD,
  parseUsdToPayoutCents,
  payoutEntryTypeLabel,
  payoutQueryKeys,
  payoutStatusBadgeVariant,
} from '@/lib/payout-utils';
import { navigateToStripeHostedUrl } from '@/lib/stripe-redirect';
import { formatDate } from '@/pages/SettingsPage/settings-utils';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Banknote, ExternalLink, Info, Loader2, RefreshCw, Shield, Wallet, XCircle } from 'lucide-react';
import { useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { toast } from 'sonner';

export interface PublisherPayoutsPanelProps {
  /** `wallet`: compact security note + id prefix for a11y when embedded in wallet. */
  variant?: 'default' | 'wallet';
  /** Prefix for form control ids (e.g. `wallet` → `wallet-payout-amount`). */
  idPrefix?: string;
}

/**
 * Publisher earnings → Stripe Connect → bank. Uses Bearer auth + server-side idempotency.
 * Stripe redirects are allowlisted in `lib/stripe-redirect.ts` before `location.assign`.
 */
export function PublisherPayoutsPanel({
  variant = 'default',
  idPrefix = 'payout',
}: PublisherPayoutsPanelProps) {
  const queryClient = useQueryClient();
  const [amountUsd, setAmountUsd] = useState('');
  const amountFieldId = `${idPrefix}-payout-amount`;

  const connectQuery = useQuery({
    queryKey: payoutQueryKeys.connect,
    queryFn: getConnectAccountStatus,
    retry: false,
  });

  const notConfigured =
    connectQuery.isError &&
    (connectQuery.error as { response?: { status?: number } })?.response?.status === 503;

  const balanceQuery = useQuery({
    queryKey: payoutQueryKeys.balance,
    queryFn: getPayoutBalance,
    retry: false,
    enabled: !notConfigured,
  });

  const requestsQuery = useQuery({
    queryKey: payoutQueryKeys.requests,
    queryFn: () => listPayoutRequests(25, 0),
    retry: false,
    enabled: !notConfigured,
  });

  const ledgerQuery = useQuery({
    queryKey: payoutQueryKeys.ledger,
    queryFn: () => listPayoutLedger(40, 0),
    retry: false,
    enabled: !notConfigured,
  });

  const onboardingMutation = useMutation({
    mutationFn: startConnectOnboarding,
    onSuccess: (data) => {
      if (data.onboarding_url) {
        if (!navigateToStripeHostedUrl(data.onboarding_url)) {
          toast.error('Could not open Stripe (invalid link). Contact support if this persists.');
        }
        return;
      }
      toast.success('Connect account is ready');
      queryClient.invalidateQueries({ queryKey: payoutQueryKeys.connect });
    },
    onError: (err) => toast.error(getPayoutApiErrorMessage(err)),
  });

  const refreshMutation = useMutation({
    mutationFn: refreshConnectAccount,
    onSuccess: () => {
      toast.success('Account status refreshed');
      queryClient.invalidateQueries({ queryKey: payoutQueryKeys.connect });
      queryClient.invalidateQueries({ queryKey: payoutQueryKeys.balance });
    },
    onError: (err) => toast.error(getPayoutApiErrorMessage(err)),
  });

  const payoutMutation = useMutation({
    mutationFn: ({ cents, key }: { cents: number; key: string }) => requestPayout(cents, key),
    onSuccess: (data) => {
      const netUsd = (data.fee.net_amount_cents / 100).toFixed(2);
      const feeUsd = (data.fee.fee_amount_cents / 100).toFixed(2);
      const desc =
        data.fee.fee_amount_cents > 0
          ? `$${netUsd} after $${feeUsd} fee — ${data.payout.status}`
          : `$${netUsd} — ${data.payout.status}`;
      toast.success('Payout submitted', { description: desc });
      setAmountUsd('');
      queryClient.invalidateQueries({ queryKey: payoutQueryKeys.balance });
      queryClient.invalidateQueries({ queryKey: payoutQueryKeys.requests });
      queryClient.invalidateQueries({ queryKey: payoutQueryKeys.ledger });
    },
    onError: (err) => toast.error(getPayoutApiErrorMessage(err)),
  });

  const cancelMutation = useMutation({
    mutationFn: ({ payoutId }: { payoutId: string }) => cancelPayout(payoutId, 'Cancelled by user'),
    onSuccess: () => {
      toast.success('Payout cancelled — funds returned to balance');
      queryClient.invalidateQueries({ queryKey: payoutQueryKeys.balance });
      queryClient.invalidateQueries({ queryKey: payoutQueryKeys.requests });
      queryClient.invalidateQueries({ queryKey: payoutQueryKeys.ledger });
    },
    onError: (err) => toast.error(getPayoutApiErrorMessage(err)),
  });

  const status: ConnectAccountStatus | undefined = connectQuery.data;
  const balance: PayoutBalance | undefined = balanceQuery.data;
  const availableUsd = balance?.available_balance_usd ?? 0;
  const maxWithdrawUsd = Math.min(MAX_PAYOUT_USD, availableUsd);

  const parsedCents = useMemo(
    () => parseUsdToPayoutCents(amountUsd, availableUsd),
    [amountUsd, availableUsd]
  );

  const amountValid = parsedCents.ok;
  const canRequestPayout =
    status?.payouts_enabled &&
    !status?.needs_onboarding &&
    amountValid &&
    !payoutMutation.isPending;

  const openStripeOnboarding = (url: string) => {
    if (!navigateToStripeHostedUrl(url)) {
      toast.error('Could not open Stripe (invalid link). Contact support if this persists.');
    }
  };

  const requests: PayoutRequest[] = requestsQuery.data?.requests ?? [];
  const ledger: PayoutLedgerEntry[] = ledgerQuery.data?.entries ?? [];

  const isWallet = variant === 'wallet';

  return (
    <div className={isWallet ? 'space-y-4' : 'space-y-6'}>
      {isWallet && (
        <Alert className="border-border-default bg-bg-secondary/40">
          <Shield className="h-4 w-4" />
          <AlertTitle className="text-sm">Secure payouts</AlertTitle>
          <AlertDescription className="text-xs text-text-secondary sm:text-sm">
            Actions use your logged-in session (Bearer token), server-side limits, and per-request
            idempotency keys. Stripe onboarding only accepts verified Stripe-hosted links. Finishing
            Connect in Stripe returns to{' '}
            <Link
              to={ROUTES.PAYOUTS}
              className="font-medium text-primary underline-offset-2 hover:underline"
            >
              {ROUTES.PAYOUTS}
            </Link>
            .
          </AlertDescription>
        </Alert>
      )}

      <Alert className="border-border-default bg-muted/40">
        <Info className="h-4 w-4 shrink-0" />
        <AlertTitle className="text-sm">
          Registry wallet and publisher earnings are separate
        </AlertTitle>
        <AlertDescription
          className={`space-y-2 text-text-secondary ${isWallet ? 'text-xs sm:text-sm' : 'text-sm'}`}
        >
          <p>
            Your <strong>registry wallet</strong> is the balance shown next to your profile. It pays
            for runs, publish fees, and top-ups inside FunctionFly — it is not withdrawn through
            this payout flow.
          </p>
          <p>
            <strong>Publisher payout balance</strong> comes only from the{' '}
            <strong>payout ledger</strong>: withdrawable earnings when revenue from <em>your</em>{' '}
            published functions is credited for payouts. What you see in your payment provider
            dashboard is not mirrored here line for line.
          </p>
          <p>
            If nothing has been credited to your payout ledger yet, your withdrawable balance stays
            at zero even when your registry wallet shows a balance.
          </p>
        </AlertDescription>
      </Alert>

      {notConfigured && (
        <Alert>
          <AlertTitle>Payouts unavailable</AlertTitle>
          <AlertDescription>
            Publisher payouts are not configured on this deployment (Stripe Connect is required).
          </AlertDescription>
        </Alert>
      )}

      {!notConfigured && connectQuery.isLoading && (
        <div className="flex justify-center py-8">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" aria-hidden />
        </div>
      )}

      {!notConfigured && connectQuery.isError && !connectQuery.isLoading && (
        <Alert variant="destructive">
          <AlertTitle>Could not load payout account</AlertTitle>
          <AlertDescription>{getPayoutApiErrorMessage(connectQuery.error)}</AlertDescription>
        </Alert>
      )}

      {status && !connectQuery.isLoading && (
        <>
          <Card className="border-border-default">
            <CardHeader className={isWallet ? 'p-4 pb-2' : undefined}>
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div className="flex items-center gap-2">
                  <Banknote className="h-5 w-5 text-text-muted" />
                  <div>
                    <CardTitle className={isWallet ? 'text-base' : 'text-lg'}>
                      Bank account (Stripe Connect)
                    </CardTitle>
                    <CardDescription className="text-text-secondary">
                      Express account for receiving transfers from FunctionFly.
                    </CardDescription>
                  </div>
                </div>
                <Badge variant={payoutStatusBadgeVariant(status.status)}>{status.status}</Badge>
              </div>
            </CardHeader>
            <CardContent className={`space-y-4 ${isWallet ? 'p-4 pt-2' : ''}`}>
              {status.has_account && (status.bank_name || status.bank_last4) && (
                <p className="text-sm text-text-primary">
                  {status.bank_name && <span>{status.bank_name}</span>}
                  {status.bank_last4 && <span className="font-mono"> ••••{status.bank_last4}</span>}
                  {status.country && <span className="text-text-muted"> · {status.country}</span>}
                </p>
              )}

              <div className="flex flex-wrap gap-2">
                {!status.has_account && (
                  <Button
                    onClick={() => onboardingMutation.mutate()}
                    disabled={onboardingMutation.isPending}
                  >
                    {onboardingMutation.isPending ? (
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    ) : null}
                    Set up payouts
                  </Button>
                )}

                {status.has_account && status.needs_onboarding && (
                  <>
                    <Button
                      onClick={() => onboardingMutation.mutate()}
                      disabled={onboardingMutation.isPending}
                    >
                      {onboardingMutation.isPending ? (
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      ) : null}
                      Continue onboarding
                    </Button>
                    {status.onboarding_url && (
                      <Button
                        variant="secondary"
                        onClick={() => openStripeOnboarding(status.onboarding_url!)}
                      >
                        <ExternalLink className="mr-2 h-4 w-4" />
                        Open Stripe link
                      </Button>
                    )}
                  </>
                )}

                {status.has_account && (
                  <Button
                    variant="outline"
                    onClick={() => refreshMutation.mutate()}
                    disabled={refreshMutation.isPending}
                  >
                    {refreshMutation.isPending ? (
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    ) : (
                      <RefreshCw className="mr-2 h-4 w-4" />
                    )}
                    Sync from Stripe
                  </Button>
                )}
              </div>

              {status.payouts_enabled && (
                <p className="text-sm text-green-600 dark:text-green-400">
                  Payouts are enabled — you can withdraw available balance below.
                </p>
              )}
            </CardContent>
          </Card>

          <Card className="border-border-default">
            <CardHeader className={isWallet ? 'p-4 pb-2' : undefined}>
              <div className="flex items-center gap-2">
                <Wallet className="h-5 w-5 text-text-muted" />
                <div>
                  <CardTitle className={isWallet ? 'text-base' : 'text-lg'}>Balance</CardTitle>
                  <CardDescription>
                    Payout ledger only — withdrawable publisher earnings, not registry wallet
                    credits.
                  </CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent className={`space-y-4 ${isWallet ? 'p-4 pt-2' : ''}`}>
              {balanceQuery.isLoading ? (
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              ) : balanceQuery.isError ? (
                <p className="text-sm text-destructive">
                  {getPayoutApiErrorMessage(balanceQuery.error)}
                </p>
              ) : balance ? (
                <>
                  <div className={`grid gap-3 ${isWallet ? 'sm:grid-cols-2' : 'sm:grid-cols-2'}`}>
                    <div className="rounded-lg border border-border-default bg-bg-secondary/50 p-4">
                      <p className="text-xs font-medium uppercase tracking-wide text-text-muted">
                        Available
                      </p>
                      <p className="mt-1 text-2xl font-semibold text-text-primary">
                        {formatPayoutUsd(balance.available_balance_usd)}
                      </p>
                    </div>
                    <div className="rounded-lg border border-border-default bg-bg-secondary/50 p-4">
                      <p className="text-xs font-medium uppercase tracking-wide text-text-muted">
                        Lifetime earnings
                      </p>
                      <p className="mt-1 text-2xl font-semibold text-text-primary">
                        {formatPayoutUsd(balance.total_earnings_usd)}
                      </p>
                    </div>
                    <div className="rounded-lg border border-border-default bg-bg-secondary/50 p-4">
                      <p className="text-xs font-medium uppercase tracking-wide text-text-muted">
                        Paid out
                      </p>
                      <p className="mt-1 text-xl font-medium text-text-primary">
                        {formatPayoutUsd(balance.total_paid_out_usd)}
                      </p>
                    </div>
                    {balance.pending_balance_usd > 0 && (
                      <div className="rounded-lg border border-border-default bg-bg-secondary/50 p-4">
                        <p className="text-xs font-medium uppercase tracking-wide text-text-muted">
                          Pending
                        </p>
                        <p className="mt-1 text-xl font-medium text-text-primary">
                          {formatPayoutUsd(balance.pending_balance_usd)}
                        </p>
                      </div>
                    )}
                  </div>

                  <div className="space-y-2 border-t border-border-subtle pt-4">
                    <Label htmlFor={amountFieldId}>Withdraw amount (USD)</Label>
                    <div className="flex max-w-md flex-col gap-2 sm:flex-row sm:items-end">
                      <Input
                        id={amountFieldId}
                        type="text"
                        inputMode="decimal"
                        autoComplete="off"
                        spellCheck={false}
                        placeholder={`${MIN_PAYOUT_USD} – ${maxWithdrawUsd.toFixed(2)}`}
                        value={amountUsd}
                        onChange={(e) => setAmountUsd(e.target.value)}
                        disabled={!status.payouts_enabled}
                        className="font-mono"
                      />
                      <Button
                        disabled={!canRequestPayout}
                        onClick={() => {
                          if (!parsedCents.ok) return;
                          payoutMutation.mutate({
                            cents: parsedCents.cents,
                            key: crypto.randomUUID(),
                          });
                        }}
                      >
                        {payoutMutation.isPending ? (
                          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        ) : null}
                        Request payout
                      </Button>
                    </div>
                    <p className="text-xs text-text-muted">
                      Minimum {formatPayoutUsd(MIN_PAYOUT_USD)}, maximum{' '}
                      {formatPayoutUsd(MAX_PAYOUT_USD)} per request. Cannot exceed available balance
                      ({formatPayoutUsd(availableUsd)}).
                    </p>
                  </div>
                </>
              ) : null}
            </CardContent>
          </Card>

          <Card className="border-border-default">
            <CardHeader className={isWallet ? 'p-4 pb-2' : undefined}>
              <CardTitle className={isWallet ? 'text-base' : 'text-lg'}>Payout history</CardTitle>
              <CardDescription>Transfer requests to your connected account</CardDescription>
            </CardHeader>
            <CardContent className={isWallet ? 'p-4 pt-0' : undefined}>
              {requestsQuery.isLoading ? (
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              ) : requestsQuery.isError ? (
                <p className="text-sm text-destructive">
                  {getPayoutApiErrorMessage(requestsQuery.error)}
                </p>
              ) : requests.length === 0 ? (
                <p className="text-sm text-text-muted">No payout requests yet.</p>
              ) : (
                <div className="max-w-full overflow-x-auto">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Date</TableHead>
                        <TableHead>Amount</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead className="hidden md:table-cell">Note</TableHead>
                        <TableHead className="w-[60px]" />
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {requests.map((r) => (
                        <TableRow key={r.id}>
                          <TableCell className="whitespace-nowrap text-text-secondary">
                            {formatDate(r.created_at)}
                          </TableCell>
                          <TableCell className="font-mono">
                            {formatPayoutCentsSigned(r.amount_cents, r.currency)}
                          </TableCell>
                          <TableCell>
                            <Badge variant={payoutStatusBadgeVariant(r.status)}>{r.status}</Badge>
                          </TableCell>
                          <TableCell className="hidden max-w-[240px] truncate text-xs text-text-muted md:table-cell">
                            {r.failure_reason ?? '—'}
                          </TableCell>
                          <TableCell>
                            {(r.status === 'pending' || r.status === 'processing') && (
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-7 w-7 text-text-muted hover:text-destructive"
                                disabled={cancelMutation.isPending}
                                onClick={() => cancelMutation.mutate({ payoutId: r.id })}
                                title="Cancel payout"
                              >
                                <XCircle className="h-4 w-4" />
                              </Button>
                            )}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              )}
            </CardContent>
          </Card>

          <Card className="border-border-default">
            <CardHeader className={isWallet ? 'p-4 pb-2' : undefined}>
              <CardTitle className={isWallet ? 'text-base' : 'text-lg'}>Ledger</CardTitle>
              <CardDescription>Credits, debits, and adjustments</CardDescription>
            </CardHeader>
            <CardContent className={isWallet ? 'p-4 pt-0' : undefined}>
              {ledgerQuery.isLoading ? (
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              ) : ledgerQuery.isError ? (
                <p className="text-sm text-destructive">
                  {getPayoutApiErrorMessage(ledgerQuery.error)}
                </p>
              ) : ledger.length === 0 ? (
                <p className="text-sm text-text-muted">No ledger entries yet.</p>
              ) : (
                <div className="max-w-full overflow-x-auto">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Date</TableHead>
                        <TableHead>Type</TableHead>
                        <TableHead>Amount</TableHead>
                        <TableHead className="hidden sm:table-cell">Balance after</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {ledger.map((e) => (
                        <TableRow key={e.id}>
                          <TableCell className="whitespace-nowrap text-text-secondary">
                            {formatDate(e.created_at)}
                          </TableCell>
                          <TableCell>{payoutEntryTypeLabel(e.entry_type)}</TableCell>
                          <TableCell className="font-mono">
                            {formatPayoutCentsSigned(e.amount_cents)}
                          </TableCell>
                          <TableCell className="hidden font-mono sm:table-cell">
                            {formatPayoutCentsSigned(e.balance_after_cents)}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              )}
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}
