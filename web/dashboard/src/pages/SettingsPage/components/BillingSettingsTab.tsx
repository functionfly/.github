import {
  cancelSubscription,
  createBillingPortalSession,
  getBillingPortalErrorMessage,
  getInvoicesErrorMessage,
  getSubscription,
  getSubscriptionErrorMessage,
  getTopUpErrorMessage,
  getWalletErrorMessage,
  getWalletInfo,
  listInvoices,
  topUpWallet,
  type Invoice,
  type Subscription,
  type UsageSummary,
} from '@/api/billing';
import { EnterpriseSettingsSection } from '@/components/enterprise';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { useAuthStore } from '@/stores/authStore';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  AlertCircle,
  Building2,
  Calendar,
  Clock,
  CreditCard,
  Download,
  Loader2,
  Trash2,
  Wallet,
  Check,
} from 'lucide-react';
import { useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';
import { formatCurrency, formatDate } from '../settings-utils';

const WALLET_TOP_UP_PRESETS = [10, 25, 50, 100] as const;
const MIN_WALLET_TOP_UP_USD = 1;
const MAX_WALLET_TOP_UP_USD = 10_000;

/** Wallet API returns USD dollars; settings `formatCurrency` expects invoice cents — keep separate. */
function formatUsd(amount: number): string {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(amount);
}

export interface BillingSettingsTabProps {
  /** Return URL after Stripe billing portal (e.g. current page). */
  returnUrl: string;
  /** Plan name to show when there is no subscription (from parent me query). */
  displayPlan: string;
}

export function BillingSettingsTab({ returnUrl, displayPlan }: BillingSettingsTabProps) {
  const user = useAuthStore((state) => state.user);
  const queryClient = useQueryClient();
  const [searchParams, setSearchParams] = useSearchParams();
  const [billingPortalLoading, setBillingPortalLoading] = useState(false);
  const [topUpAmountInput, setTopUpAmountInput] = useState('25');
  const [topUpSubmitting, setTopUpSubmitting] = useState(false);
  const [contactModalOpen, setContactModalOpen] = useState(false);
  const [contactForm, setContactForm] = useState({
    name: user?.name || '',
    email: user?.email || '',
    company: '',
    message: '',
  });
  const [contactSubmitting, setContactSubmitting] = useState(false);
  const [cancelModalOpen, setCancelModalOpen] = useState(false);
  const [cancelSubmitting, setCancelSubmitting] = useState(false);

  const {
    data: subscriptionData,
    isLoading: subscriptionLoading,
    error: subscriptionError,
  } = useQuery({
    queryKey: ['billing', 'subscription'],
    // A 404 means the user simply has no subscription row in our DB yet.
    // Treat it as "no subscription" (null) so we don't show an error banner.
    queryFn: async () => {
      try {
        return await getSubscription();
      } catch (e: unknown) {
        const status = (e as { response?: { status?: number } })?.response?.status;
        if (status === 404) return null;
        throw e;
      }
    },
    retry: false,
  });

  const {
    data: invoicesData,
    isLoading: invoicesLoading,
    error: invoicesError,
  } = useQuery({
    queryKey: ['billing', 'invoices'],
    queryFn: () => listInvoices(10, 0),
    retry: false,
  });

  const {
    data: walletData,
    isLoading: walletLoading,
    error: walletError,
  } = useQuery({
    queryKey: ['billing-wallet-info', user?.id],
    queryFn: getWalletInfo,
    enabled: !!user,
    staleTime: 30_000,
    retry: false,
  });

  // Derived constants must be declared before hooks that depend on them
  const subscription = subscriptionData as Subscription | null | undefined;

  useEffect(() => {
    const walletFlag = searchParams.get('walletTopUp');
    const subscriptionFlag = searchParams.get('subscription');

    if (walletFlag) {
      if (walletFlag === 'success') {
        toast.success('Payment completed', {
          description:
            'Your registry balance updates after Stripe confirms the payment (usually within a few seconds).',
        });
        queryClient.invalidateQueries({ queryKey: ['billing-wallet-info'] });
      } else if (walletFlag === 'cancel') {
        toast.message('Top-up cancelled');
      }
      const next = new URLSearchParams(searchParams);
      next.delete('walletTopUp');
      setSearchParams(next, { replace: true });
    }

    if (subscriptionFlag) {
      if (subscriptionFlag === 'success') {
        toast.success('Subscription updated', {
          description: 'Your subscription has been successfully updated.',
        });
        queryClient.invalidateQueries({ queryKey: ['billing', 'subscription'] });
        queryClient.invalidateQueries({ queryKey: ['billing', 'invoices'] });
      } else if (subscriptionFlag === 'cancel') {
        toast.message('Checkout cancelled');
      }
      const next = new URLSearchParams(searchParams);
      next.delete('subscription');
      setSearchParams(next, { replace: true });
    }
  }, [searchParams, setSearchParams, queryClient]);

  const parsedTopUpUsd = parseFloat(topUpAmountInput.replace(/,/g, ''));
  const topUpAmountValid =
    !Number.isNaN(parsedTopUpUsd) &&
    parsedTopUpUsd >= MIN_WALLET_TOP_UP_USD &&
    parsedTopUpUsd <= MAX_WALLET_TOP_UP_USD;

  const buildWalletTopUpReturnUrl = (outcome: 'success' | 'cancel') => {
    try {
      const u = new URL(returnUrl);
      u.searchParams.set('walletTopUp', outcome);
      return u.toString();
    } catch {
      const u = new URL(window.location.href);
      u.searchParams.set('walletTopUp', outcome);
      return u.toString();
    }
  };

  const handleWalletTopUp = async () => {
    if (!topUpAmountValid) {
      toast.error(
        `Enter an amount between $${MIN_WALLET_TOP_UP_USD} and $${MAX_WALLET_TOP_UP_USD.toLocaleString()}.`
      );
      return;
    }
    setTopUpSubmitting(true);
    try {
      const { checkout_url: checkoutUrl } = await topUpWallet(
        Math.round(parsedTopUpUsd * 100) / 100,
        buildWalletTopUpReturnUrl('success'),
        buildWalletTopUpReturnUrl('cancel')
      );
      window.location.href = checkoutUrl;
    } catch (e: unknown) {
      setTopUpSubmitting(false);
      toast.error(getTopUpErrorMessage(e));
    }
  };

  const invoices = (invoicesData as { invoices: Invoice[] })?.invoices ?? [];

  const handleContactSales = async () => {
    if (!contactForm.name || !contactForm.email || !contactForm.message) {
      toast.error('Please fill in all required fields');
      return;
    }
    setContactSubmitting(true);
    try {
      const subject = encodeURIComponent(`Enterprise Plan Inquiry from ${contactForm.name}`);
      const body = encodeURIComponent(
        `Name: ${contactForm.name}\nEmail: ${contactForm.email}\nCompany: ${contactForm.company || 'Not provided'}\n\nMessage:\n${contactForm.message}`
      );
      window.location.href = `mailto:sales@functionfly.com?subject=${subject}&body=${body}`;
      setContactModalOpen(false);
      toast.success('Thank you for your interest! Our sales team will contact you soon.');
    } catch {
      toast.error(
        'Failed to submit contact request. Please email us directly at sales@functionfly.com'
      );
    } finally {
      setContactSubmitting(false);
    }
  };

  const handleCancelSubscription = async () => {
    setCancelSubmitting(true);
    try {
      await cancelSubscription(false);
      toast.success('Subscription will be cancelled at the end of the billing period');
      setCancelModalOpen(false);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to cancel subscription');
    } finally {
      setCancelSubmitting(false);
    }
  };

  const openPortal = async (urlPath: string) => {
    setBillingPortalLoading(true);
    try {
      const { url } = await createBillingPortalSession(urlPath);
      window.location.href = url;
    } catch (e: unknown) {
      setBillingPortalLoading(false);
      toast.error(getBillingPortalErrorMessage(e));
    }
  };

  return (
    <div className="space-y-6">
      <EnterpriseSettingsSection />
      <Card>
        <CardHeader>
          <CardTitle>Current Plan</CardTitle>
          <CardDescription className="text-text-secondary">
            Manage your subscription
          </CardDescription>
        </CardHeader>
        <CardContent>
          {subscriptionLoading ? (
            <div className="flex items-center justify-center p-4">
              <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-[#6366f1]" />
            </div>
          ) : subscriptionError ? (
            <div className="flex items-center gap-2 p-4 rounded-lg bg-amber-500/10 border border-amber-500/20">
              <AlertCircle className="w-5 h-5 text-amber-500" />
              <p className="text-amber-500 text-sm">
                {getSubscriptionErrorMessage(subscriptionError)}
              </p>
            </div>
          ) : subscription ? (
            <div className="space-y-4">
              <div className="flex items-center justify-between p-4 rounded-lg bg-linear-to-r from-[#6366f1]/10 to-[#8b5cf6]/10 border border-border-default">
                <div>
                  <h3 className="font-semibold text-text-primary capitalize">
                    {subscription.plan} Plan
                  </h3>
                  <div className="flex items-center gap-2 mt-1">
                    <Badge
                      variant={subscription.status === 'active' ? 'default' : 'secondary'}
                      className={
                        subscription.status === 'active'
                          ? 'bg-green-500/20 text-green-600 dark:text-green-400 border border-green-500/40 dark:border-green-500/30'
                          : ''
                      }
                    >
                      {subscription.status}
                    </Badge>
                    {subscription.cancel_at_period_end && (
                      <Badge
                        variant="outline"
                        className="border-amber-500/50 text-amber-600 dark:text-amber-400"
                      >
                        Cancels at period end
                      </Badge>
                    )}
                  </div>
                </div>
                <Badge>Current</Badge>
              </div>

              {/* Trial Period Display */}
              {subscription.is_trialing && (
                <div
                  className={`p-4 rounded-lg border ${
                    subscription.trial_days_remaining <= 3
                      ? 'bg-amber-500/10 border-amber-500/20'
                      : 'bg-blue-500/10 border-blue-500/20'
                  }`}
                >
                  <div className="flex items-center gap-2 mb-2">
                    <Clock className="w-5 h-5 text-text-muted" />
                    <span className="text-sm font-medium">Trial Period</span>
                  </div>
                  <p className="text-sm">
                    <span
                      className={
                        subscription.trial_days_remaining <= 3
                          ? 'text-amber-400 font-semibold'
                          : 'text-blue-400'
                      }
                    >
                      {subscription.trial_days_remaining} days remaining
                    </span>
                    {subscription.trial_end && <> · Ends {formatDate(subscription.trial_end)}</>}
                  </p>
                  {subscription.trial_days_remaining <= 3 && (
                    <p className="text-xs mt-2 text-amber-400">
                      Your trial ends soon. Choose a plan to continue using premium features.
                    </p>
                  )}
                </div>
              )}

              {(subscription.current_period_start || subscription.current_period_end) && (
                <div className="grid grid-cols-2 gap-4 p-4 rounded-lg bg-bg-secondary border border-border-default">
                  <div className="flex items-center gap-3">
                    <Calendar className="w-5 h-5 text-text-muted" />
                    <div>
                      <p className="text-sm text-text-muted">Current Period Start</p>
                      <p className="text-text-primary font-medium">
                        {formatDate(subscription.current_period_start)}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <Calendar className="w-5 h-5 text-text-muted" />
                    <div>
                      <p className="text-sm text-text-muted">Next Billing Date</p>
                      <p className="text-text-primary font-medium">
                        {formatDate(subscription.current_period_end)}
                      </p>
                    </div>
                  </div>
                </div>
              )}
              {subscription.payment_method && (
                <div className="p-4 rounded-lg bg-bg-secondary border border-border-default">
                  <p className="text-sm text-text-muted mb-2">Payment Method</p>
                  <div className="flex items-center gap-3">
                    <CreditCard className="w-5 h-5 text-[#6366f1]" />
                    <div>
                      <p className="text-text-primary font-medium capitalize">
                        {subscription.payment_method.brand} •••• {subscription.payment_method.last4}
                      </p>
                      {subscription.payment_method.exp_month &&
                        subscription.payment_method.exp_year && (
                          <p className="text-sm text-text-muted">
                            Expires {subscription.payment_method.exp_month}/
                            {subscription.payment_method.exp_year}
                          </p>
                        )}
                    </div>
                  </div>
                </div>
              )}
            </div>
          ) : (
            <div className="space-y-4">
              <div className="flex items-center justify-between p-4 rounded-lg bg-linear-to-r from-[#6366f1]/10 to-[#8b5cf6]/10 border border-border-default">
                <div>
                  <h3 className="font-semibold text-text-primary capitalize">{displayPlan} Plan</h3>
                  <p className="text-sm text-text-secondary mt-1">
                    {displayPlan === 'free' || displayPlan.toLowerCase() === 'free' ? (
                      <>
                        <Badge variant="secondary" className="mr-2">
                          Free Forever
                        </Badge>
                        Basic features included
                      </>
                    ) : (
                      <>
                        <Badge variant="secondary" className="mr-2">
                          {displayPlan}
                        </Badge>
                        Active
                      </>
                    )}
                  </p>
                </div>
                <Badge>Current</Badge>
              </div>

              {/* Free tier features list */}
              {(displayPlan === 'free' || displayPlan.toLowerCase() === 'free') && (
                <>
                  <div className="p-4 rounded-lg bg-bg-secondary border border-border-default">
                    <p className="font-medium text-text-primary mb-2">Your Free Plan includes:</p>
                    <ul className="space-y-1 text-sm text-text-secondary">
                      <li className="flex items-center gap-2">
                        <span className="w-5 h-5 rounded-full bg-green-500/20 border border-green-500/30 flex items-center justify-center shrink-0">
                          <Check className="w-3.5 h-3.5 text-green-400" />
                        </span>
                        Basic function deployment
                      </li>
                      <li className="flex items-center gap-2">
                        <span className="w-5 h-5 rounded-full bg-green-500/20 border border-green-500/30 flex items-center justify-center shrink-0">
                          <Check className="w-3.5 h-3.5 text-green-400" />
                        </span>
                        Community support
                      </li>
                      <li className="flex items-center gap-2">
                        <span className="w-5 h-5 rounded-full bg-green-500/20 border border-green-500/30 flex items-center justify-center shrink-0">
                          <Check className="w-3.5 h-3.5 text-green-400" />
                        </span>
                        Registry access
                      </li>
                      <li className="flex items-center gap-2">
                        <span className="w-5 h-5 rounded-full bg-green-500/20 border border-green-500/30 flex items-center justify-center shrink-0">
                          <Check className="w-3.5 h-3.5 text-green-400" />
                        </span>
                        Up to 5 functions
                      </li>
                    </ul>
                  </div>

                  {/* Upgrade prompt for free users */}
                  <div className="p-4 rounded-lg bg-gradient-to-r from-indigo-500/10 to-purple-500/10 border border-indigo-500/20">
                    <p className="text-sm font-medium text-text-primary mb-2">
                      Ready to unlock more?
                    </p>
                    <ul className="space-y-1 text-xs text-text-secondary mb-3">
                      <li>- Unlimited executions</li>
                      <li>- Priority support</li>
                      <li>- Advanced analytics</li>
                    </ul>
                    <Button
                      size="sm"
                      onClick={() => openPortal(`${window.location.origin}/pricing`)}
                    >
                      View Plans & Pricing
                    </Button>
                  </div>
                </>
              )}
            </div>
          )}
          <div className="mt-6 flex flex-wrap gap-3">
            <Button
              variant="default"
              onClick={() => openPortal(returnUrl)}
              disabled={billingPortalLoading}
            >
              {billingPortalLoading ? 'Opening…' : 'Manage billing'}
            </Button>
            <Button
              variant="outline"
              className="settings-upgrade-btn border-border-strong"
              disabled={billingPortalLoading}
              onClick={() => openPortal(`${window.location.origin}/pricing`)}
            >
              {billingPortalLoading ? 'Opening…' : 'Upgrade Plan'}
            </Button>
            <Dialog open={contactModalOpen} onOpenChange={setContactModalOpen}>
              <DialogTrigger asChild>
                <Button
                  variant="outline"
                  className="border-border-strong border-[#6366f1]/50 text-[#6366f1] hover:bg-[#6366f1]/10"
                >
                  <Building2 className="w-4 h-4 mr-2" />
                  Contact Sales
                </Button>
              </DialogTrigger>
              <DialogContent className="sm:max-w-[500px]">
                <DialogHeader>
                  <DialogTitle>Contact Sales</DialogTitle>
                  <DialogDescription>
                    Interested in our Enterprise plan? Fill out the form below and our team will get
                    back to you within 24 hours.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="contact-name">Name *</Label>
                    <Input
                      id="contact-name"
                      value={contactForm.name}
                      onChange={(e) => setContactForm({ ...contactForm, name: e.target.value })}
                      placeholder="Your name"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="contact-email">Email *</Label>
                    <Input
                      id="contact-email"
                      type="email"
                      value={contactForm.email}
                      onChange={(e) => setContactForm({ ...contactForm, email: e.target.value })}
                      placeholder="your@email.com"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="contact-company">Company</Label>
                    <Input
                      id="contact-company"
                      value={contactForm.company}
                      onChange={(e) => setContactForm({ ...contactForm, company: e.target.value })}
                      placeholder="Your company name"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="contact-message">Message *</Label>
                    <Textarea
                      id="contact-message"
                      value={contactForm.message}
                      onChange={(e) => setContactForm({ ...contactForm, message: e.target.value })}
                      placeholder="Tell us about your needs..."
                      rows={4}
                    />
                  </div>
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setContactModalOpen(false)}>
                    Cancel
                  </Button>
                  <Button
                    onClick={handleContactSales}
                    disabled={contactSubmitting}
                    className="bg-[#6366f1] hover:bg-[#6366f1]/90"
                  >
                    {contactSubmitting ? 'Sending...' : 'Send Message'}
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
            {subscription &&
              subscription.status === 'active' &&
              !subscription.cancel_at_period_end && (
                <Dialog open={cancelModalOpen} onOpenChange={setCancelModalOpen}>
                  <DialogTrigger asChild>
                    <Button
                      variant="outline"
                      className="border-border-strong border-red-500/50 text-red-400 hover:bg-red-500/10"
                    >
                      <Trash2 className="w-4 h-4 mr-2" />
                      Cancel Subscription
                    </Button>
                  </DialogTrigger>
                  <DialogContent className="sm:max-w-[500px]">
                    <DialogHeader>
                      <DialogTitle>Cancel Subscription</DialogTitle>
                      <DialogDescription>
                        Are you sure you want to cancel your subscription? You'll lose access to
                        premium features at the end of your billing period.
                      </DialogDescription>
                    </DialogHeader>
                    <div className="py-4">
                      <div className="p-4 rounded-lg bg-amber-500/10 border border-amber-500/20">
                        <p className="text-amber-400 text-sm">
                          Your subscription will remain active until{' '}
                          {formatDate(subscription.current_period_end)}. After that, you'll be
                          downgraded to the free plan.
                        </p>
                      </div>
                    </div>
                    <DialogFooter>
                      <Button variant="outline" onClick={() => setCancelModalOpen(false)}>
                        Keep Subscription
                      </Button>
                      <Button
                        onClick={handleCancelSubscription}
                        disabled={cancelSubmitting}
                        variant="destructive"
                      >
                        {cancelSubmitting ? 'Cancelling...' : 'Confirm Cancellation'}
                      </Button>
                    </DialogFooter>
                  </DialogContent>
                </Dialog>
              )}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Wallet className="h-5 w-5 text-[#6366f1]" />
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
              <AlertCircle className="w-5 h-5 text-amber-500 shrink-0" />
              <p className="text-amber-500 text-sm">{getWalletErrorMessage(walletError)}</p>
            </div>
          ) : (
            <>
              <div className="grid gap-3 sm:grid-cols-3 rounded-lg bg-bg-secondary border border-border-default p-4">
                <div>
                  <p className="text-xs text-text-muted uppercase tracking-wide">Balance</p>
                  <p className="text-lg font-semibold text-amber-500 tabular-nums">
                    {formatUsd(walletData?.balance_usd ?? 0)}
                  </p>
                </div>
                <div>
                  <p className="text-xs text-text-muted uppercase tracking-wide">Lifetime earned</p>
                  <p className="text-lg font-medium text-text-primary tabular-nums">
                    {formatUsd(walletData?.lifetime_earnings_usd ?? 0)}
                  </p>
                </div>
                <div>
                  <p className="text-xs text-text-muted uppercase tracking-wide">Fees paid</p>
                  <p className="text-lg font-medium text-text-primary tabular-nums">
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
                      className="border-border-strong tabular-nums"
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
                className="bg-[#6366f1] hover:bg-[#6366f1]/90"
                disabled={topUpSubmitting || !topUpAmountValid}
                onClick={handleWalletTopUp}
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

      <Card>
        <CardHeader>
          <CardTitle>Invoices</CardTitle>
          <CardDescription className="text-text-secondary">
            View and download your past invoices
          </CardDescription>
        </CardHeader>
        <CardContent>
          {invoicesLoading ? (
            <div className="flex items-center justify-center p-4">
              <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-[#6366f1]" />
            </div>
          ) : invoicesError ? (
            <div className="flex items-center gap-2 p-4 rounded-lg bg-amber-500/10 border border-amber-500/20">
              <AlertCircle className="w-5 h-5 text-amber-500" />
              <p className="text-amber-500 text-sm">{getInvoicesErrorMessage(invoicesError)}</p>
            </div>
          ) : invoices.length === 0 ? (
            <div className="text-center p-6">
              <CreditCard className="w-12 h-12 text-text-muted mx-auto mb-3" />
              <p className="text-text-muted">No invoices yet</p>
              <p className="text-sm text-text-muted">
                Your invoices will appear here after your first payment
              </p>
            </div>
          ) : (
            <div className="space-y-3">
              {invoices.map((invoice) => (
                <div
                  key={invoice.id}
                  className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-border-default hover:border-border-strong transition-colors"
                >
                  <div className="flex items-center gap-4">
                    <div className="w-10 h-10 rounded-lg bg-[#6366f1]/20 flex items-center justify-center">
                      <CreditCard className="w-5 h-5 text-[#6366f1]" />
                    </div>
                    <div>
                      <p className="font-medium text-text-primary">
                        {formatCurrency(invoice.amount, invoice.currency)}
                      </p>
                      <p className="text-sm text-text-muted">{formatDate(invoice.invoice_date)}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <Badge
                      variant={invoice.status === 'paid' ? 'default' : 'secondary'}
                      className={
                        invoice.status === 'paid'
                          ? 'bg-green-500/20 text-green-400 border-green-500/30'
                          : ''
                      }
                    >
                      {invoice.status}
                    </Badge>
                    {invoice.invoice_pdf || invoice.hosted_invoice_url ? (
                      <Button variant="ghost" size="sm" asChild>
                        <a
                          href={invoice.invoice_pdf || invoice.hosted_invoice_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="flex items-center gap-1"
                        >
                          <Download className="w-4 h-4" />
                          Download
                        </a>
                      </Button>
                    ) : invoice.status === 'paid' ? (
                      <span
                        className="text-xs text-text-muted"
                        title="Invoice PDF will be available shortly"
                      >
                        Processing...
                      </span>
                    ) : null}
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
