import {
  cancelSubscription,
  createBillingPortalSession,
  getBillingPortalErrorMessage,
  getSubscription,
  getSubscriptionErrorMessage,
  getTopUpErrorMessage,
  getUsage,
  getWalletInfo,
  listInvoices,
  topUpWallet,
  type Invoice,
  type Subscription,
} from '@/api/billing';
import { getCostSummary } from '@/api/usageAnalytics';
import { EnterpriseSettingsSection } from '@/components/enterprise';
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
import { Chamber, CornerBrace, FrameButton, SealedButton, StatusPill } from '@/components/sc';
import { PLANS } from '@/lib/constants';
import { useAuthStore } from '@/stores/authStore';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { AlertCircle, Trash2, Zap } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';
import { formatCurrency, formatDate } from '../../settings-utils';
import { BundleStatusSection } from './BundleStatusSection';
import { CurrentPlanSection } from './CurrentPlanSection';
import { FreeTierSection } from './FreeTierSection';
import { InvoicesSection } from './InvoicesSection';
import { MAX_WALLET_TOP_UP_USD, MIN_WALLET_TOP_UP_USD, WalletSection } from './WalletSection';

export interface BillingSettingsTabProps {
  /** Return URL after Stripe billing portal (e.g. current page). */
  returnUrl: string;
  /** Plan name to show when there is no subscription (from parent me query). */
  displayPlan: string;
}

interface UsageMetric {
  label: string;
  current: number;
  limit: number;
  unit: string;
  icon: React.ReactNode;
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

  const { data: usageData, isLoading: usageLoading } = useQuery({
    queryKey: ['billing', 'usage'],
    queryFn: async () => {
      try {
        return await getUsage();
      } catch {
        return null;
      }
    },
    retry: false,
  });

  const { data: costData, isLoading: costLoading } = useQuery({
    queryKey: ['billing', 'cost-summary'],
    queryFn: async () => {
      try {
        return await getCostSummary();
      } catch {
        return null;
      }
    },
    retry: false,
  });

  const subscription = subscriptionData as Subscription | null | undefined;

  const usageMetrics = useMemo((): UsageMetric[] => {
    const planKey = subscription?.plan?.toUpperCase() as keyof typeof PLANS;
    const planLimits = planKey ? PLANS[planKey]?.limits : null;

    if (!planLimits) return [];

    const totalExecutions = usageData?.total_events ?? 0;
    const requestsLimit = planLimits.requests === Infinity ? -1 : planLimits.requests;

    const metrics: UsageMetric[] = [];

    if (requestsLimit > 0) {
      metrics.push({
        label: 'API Requests',
        current: totalExecutions,
        limit: requestsLimit,
        unit: 'requests',
        icon: <Zap className="w-4 h-4" />,
      });
    }

    return metrics;
  }, [subscription?.plan, usageData]);

  const projectedBilling = useMemo(() => {
    if (!subscription?.current_period_end) return null;

    const periodEnd = new Date(subscription.current_period_end);
    const now = new Date();
    const daysRemaining = Math.ceil((periodEnd.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
    const totalDays = subscription.current_period_start
      ? Math.ceil(
          (periodEnd.getTime() - new Date(subscription.current_period_start).getTime()) /
            (1000 * 60 * 60 * 24)
        )
      : 30;

    const daysPassed = totalDays - daysRemaining;
    const usagePercent = totalDays > 0 ? (daysPassed / totalDays) * 100 : 0;

    const currentUsage = usageData?.total_events ?? 0;
    const dailyRate = daysPassed > 0 ? currentUsage / daysPassed : 0;
    const projectedTotal = Math.round(dailyRate * totalDays);

    return {
      periodEnd,
      daysRemaining: Math.max(0, daysRemaining),
      usagePercent: Math.min(100, usagePercent),
      projectedTotal,
      dailyRate,
      currentUsage,
    };
  }, [subscription, usageData]);

  const planOptions = useMemo(() => {
    const currentPlan = subscription?.plan?.toLowerCase() || displayPlan.toLowerCase();
    const plans = [
      { id: 'free', name: 'Free', tier: 0 },
      { id: 'starter', name: 'Starter', tier: 1 },
      { id: 'professional', name: 'Professional', tier: 2 },
      { id: 'enterprise', name: 'Enterprise', tier: 3 },
    ];
    const currentTier = plans.find((p) => p.id === currentPlan)?.tier ?? 0;

    return plans.map((plan) => ({
      ...plan,
      isCurrent: plan.id === currentPlan,
      isUpgrade: plan.tier > currentTier,
      isDowngrade: plan.tier < currentTier && plan.tier > 0,
    }));
  }, [subscription?.plan, displayPlan]);

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
      <BundleStatusSection />
      <Chamber>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <div className="mb-4">
          <h3 className="font-display text-lg font-semibold" style={{ color: 'var(--text)' }}>
            Current Plan
          </h3>
          <p className="text-sm mt-1" style={{ color: 'var(--text-dim)' }}>
            Manage your subscription
          </p>
        </div>
        <div>
          {subscriptionLoading ? (
            <div className="flex items-center justify-center p-4">
              <div
                className="animate-spin rounded-full h-6 w-6 border-b-2"
                style={{ borderColor: 'var(--accent)' }}
              />
            </div>
          ) : subscriptionError ? (
            <div
              className="flex items-center gap-2 p-4 rounded-lg"
              style={{
                background: 'rgba(232, 196, 104, 0.06)',
                border: '1px solid rgba(232, 196, 104, 0.3)',
              }}
            >
              <AlertCircle className="w-5 h-5" style={{ color: 'var(--status-pending)' }} />
              <p className="text-sm" style={{ color: 'var(--status-pending)' }}>
                {getSubscriptionErrorMessage(subscriptionError)}
              </p>
            </div>
          ) : subscription ? (
            <CurrentPlanSection
              subscription={subscription}
              planOptions={planOptions}
              projectedBilling={projectedBilling}
              usageMetrics={usageMetrics}
              usageLoading={usageLoading}
              costData={costData}
              costLoading={costLoading}
              billingPortalLoading={billingPortalLoading}
              returnUrl={returnUrl}
              openPortal={openPortal}
              formatDate={formatDate}
            />
          ) : (
            <FreeTierSection
              displayPlan={displayPlan}
              planOptions={planOptions}
              billingPortalLoading={billingPortalLoading}
              openPortal={openPortal}
            />
          )}

          <div className="mt-6 flex flex-wrap gap-3">
            <SealedButton
              onClick={() => openPortal(returnUrl)}
              disabled={billingPortalLoading}
              loading={billingPortalLoading}
            >
              Manage billing
            </SealedButton>
            <FrameButton
              disabled={billingPortalLoading}
              onClick={() => (window.location.href = '/pricing')}
            >
              Upgrade Plan
            </FrameButton>
            <Dialog open={contactModalOpen} onOpenChange={setContactModalOpen}>
              <DialogTrigger asChild>
                <FrameButton>Contact Sales</FrameButton>
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
                  <FrameButton onClick={() => setContactModalOpen(false)}>Cancel</FrameButton>
                  <SealedButton onClick={handleContactSales} disabled={contactSubmitting}>
                    {contactSubmitting ? 'Sending...' : 'Send Message'}
                  </SealedButton>
                </DialogFooter>
              </DialogContent>
            </Dialog>
            {subscription &&
              subscription.status === 'active' &&
              !subscription.cancel_at_period_end && (
                <Dialog open={cancelModalOpen} onOpenChange={setCancelModalOpen}>
                  <DialogTrigger asChild>
                    <FrameButton iconLeft={<Trash2 className="w-4 h-4" />}>
                      Cancel Subscription
                    </FrameButton>
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
                      <div
                        className="p-4 rounded-lg"
                        style={{
                          background: 'rgba(232, 196, 104, 0.06)',
                          border: '1px solid rgba(232, 196, 104, 0.3)',
                        }}
                      >
                        <p className="text-sm" style={{ color: 'var(--status-pending)' }}>
                          Your subscription will remain active until{' '}
                          {formatDate(subscription.current_period_end)}. After that, you'll be
                          downgraded to the free plan.
                        </p>
                      </div>
                    </div>
                    <DialogFooter>
                      <FrameButton onClick={() => setCancelModalOpen(false)}>
                        Keep Subscription
                      </FrameButton>
                      <SealedButton
                        onClick={handleCancelSubscription}
                        disabled={cancelSubmitting}
                      >
                        {cancelSubmitting ? 'Cancelling...' : 'Confirm Cancellation'}
                      </SealedButton>
                    </DialogFooter>
                  </DialogContent>
                </Dialog>
              )}
          </div>
        </div>
      </Chamber>

      <WalletSection
        walletData={walletData}
        walletLoading={walletLoading}
        walletError={walletError}
        topUpAmountInput={topUpAmountInput}
        topUpSubmitting={topUpSubmitting}
        topUpAmountValid={topUpAmountValid}
        onTopUpAmountChange={setTopUpAmountInput}
        onWalletTopUp={handleWalletTopUp}
      />

      <InvoicesSection
        invoices={invoices}
        invoicesLoading={invoicesLoading}
        invoicesError={invoicesError}
        formatCurrency={formatCurrency}
        formatDate={formatDate}
      />
    </div>
  );
}
