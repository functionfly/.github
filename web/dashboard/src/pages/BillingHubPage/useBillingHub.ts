import {
  cancelSubscription,
  createBillingPortalSession,
  createStateFabricAddOnCheckout,
  listStateFabricAddOnCatalog,
  getStateFabricAddOnEntitlements,
  getSubscription,
  getUsage,
  getWalletInfo,
  getWalletTransactions,
  listInvoices,
  listPaymentMethods,
  topUpWallet,
  type Invoice,
  type PaymentMethod,
  type StateFabricAddOnCatalogResponse,
  type Subscription,
  type WalletInfo,
  type WalletTransaction,
} from '@/api/billing';
import { getCostSummary, type CostSummary } from '@/api/usageAnalytics';
import { getPlanLimits, hasFeature } from '@/lib/plan-utils';
import { useAuthStore } from '@/stores/authStore';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';

export interface UsageMetric {
  label: string;
  current: number;
  limit: number;
  unit: string;
  icon: React.ReactNode;
}

export interface ProjectedBilling {
  periodEnd: Date;
  daysRemaining: number;
  usagePercent: number;
  projectedTotal: number;
  dailyRate: number;
  currentUsage: number;
}

export interface BillingHubState {
  subscription: Subscription | null;
  invoices: Invoice[];
  walletInfo: WalletInfo | null;
  walletTransactions: WalletTransaction[];
  paymentMethods: PaymentMethod[];
  addOnCatalog: StateFabricAddOnCatalogResponse['add_ons'];
  entitledAddOnIds: string[];
  usageData: { total_events: number; total_cost_usd: number } | null;
  costData: CostSummary | null;
}

export interface BillingHubActions {
  openBillingPortal: () => Promise<void>;
  handleTopUp: (amountUsd: number) => Promise<void>;
  handleCancelSubscription: () => Promise<void>;
  handleAddOnPurchase: (addonId: string) => Promise<void>;
  refreshAll: () => void;
}

function useBillingHub(): {
  state: BillingHubState;
  actions: BillingHubActions;
  isLoading: {
    subscription: boolean;
    invoices: boolean;
    wallet: boolean;
    paymentMethods: boolean;
    addOns: boolean;
    usage: boolean;
  };
  errors: {
    subscription: Error | null;
    invoices: Error | null;
    wallet: Error | null;
    paymentMethods: Error | null;
    addOns: Error | null;
    usage: Error | null;
  };
  projectedBilling: ProjectedBilling | null;
  usageMetrics: UsageMetric[];
  planLimits: ReturnType<typeof getPlanLimits>;
} {
  const user = useAuthStore((s) => s.user);
  const queryClient = useQueryClient();
  const [searchParams, setSearchParams] = useSearchParams();
  const [portalLoading, setPortalLoading] = useState(false);
  const [topUpSubmitting, setTopUpSubmitting] = useState(false);
  const [addOnCheckoutLoading, setAddOnCheckoutLoading] = useState<string | null>(null);

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
    queryFn: () => listInvoices(20, 0),
    retry: false,
  });

  const {
    data: walletInfoData,
    isLoading: walletLoading,
    error: walletError,
  } = useQuery({
    queryKey: ['billing-wallet-info', user?.id],
    queryFn: getWalletInfo,
    enabled: !!user,
    staleTime: 30_000,
    retry: false,
  });

  const {
    data: walletTransactionsData,
    isLoading: walletTransactionsLoading,
  } = useQuery({
    queryKey: ['billing-wallet-transactions', user?.id],
    queryFn: () => getWalletTransactions(50),
    enabled: !!user,
    staleTime: 30_000,
    retry: false,
  });

  const {
    data: paymentMethodsData,
    isLoading: paymentMethodsLoading,
    error: paymentMethodsError,
  } = useQuery({
    queryKey: ['billing', 'payment-methods'],
    queryFn: listPaymentMethods,
    retry: false,
  });

  const {
    data: addOnCatalogData,
    isLoading: addOnCatalogLoading,
    error: addOnCatalogError,
  } = useQuery({
    queryKey: ['billing', 'state-fabric-add-ons-catalog'],
    queryFn: listStateFabricAddOnCatalog,
    retry: false,
  });

  const {
    data: addOnEntitlementsData,
    isLoading: addOnEntitlementsLoading,
    error: addOnEntitlementsError,
  } = useQuery({
    queryKey: ['billing', 'state-fabric-add-ons-entitlements'],
    queryFn: getStateFabricAddOnEntitlements,
    retry: false,
  });

  const {
    data: usageData,
    isLoading: usageLoading,
    error: usageError,
  } = useQuery({
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

  const {
    data: costData,
    isLoading: costLoading,
  } = useQuery({
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

  const subscription = subscriptionData as Subscription | null;
  const invoices = (invoicesData as { invoices: Invoice[] } | undefined)?.invoices ?? [];
  const walletInfo = walletInfoData as WalletInfo | null;
  const walletTransactions = (walletTransactionsData as { transactions: WalletTransaction[] } | undefined)?.transactions ?? [];
  const paymentMethods = (paymentMethodsData as { payment_methods: PaymentMethod[] } | undefined)?.payment_methods ?? [];
  const addOnCatalog = addOnCatalogData?.add_ons ?? [];
  const entitledAddOnIds = (addOnEntitlementsData as { addon_ids: string[] } | undefined)?.addon_ids ?? [];

  const planLimits = useMemo(() => {
    return getPlanLimits(subscription?.plan ?? user?.plan);
  }, [subscription?.plan, user?.plan]);

  const usageMetrics = useMemo((): UsageMetric[] => {
    if (!planLimits) return [];
    const totalExecutions = usageData?.total_events ?? 0;
    const requestsLimit = planLimits.requests === Infinity ? -1 : planLimits.requests;
    if (requestsLimit <= 0) return [];

    return [{
      label: 'API Requests',
      current: totalExecutions,
      limit: requestsLimit,
      unit: 'requests',
      icon: null,
    }];
  }, [planLimits, usageData]);

  const projectedBilling = useMemo((): ProjectedBilling | null => {
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

  const returnUrl = useMemo(() => {
    return window.location.origin + '/billing';
  }, []);

  const buildReturnUrl = useCallback((outcome: string) => {
    try {
      const u = new URL(returnUrl);
      u.searchParams.set('billing', outcome);
      return u.toString();
    } catch {
      const u = new URL(window.location.href);
      u.searchParams.set('billing', outcome);
      return u.toString();
    }
  }, [returnUrl]);

  const openBillingPortal = useCallback(async () => {
    setPortalLoading(true);
    try {
      const { url } = await createBillingPortalSession(returnUrl);
      window.location.href = url;
    } catch (e: unknown) {
      setPortalLoading(false);
      toast.error('Could not open billing portal. Please try again.');
    }
  }, [returnUrl]);

  const handleTopUp = useCallback(async (amountUsd: number) => {
    setTopUpSubmitting(true);
    try {
      const { checkout_url: checkoutUrl } = await topUpWallet(
        amountUsd,
        buildReturnUrl('wallet-success'),
        buildReturnUrl('wallet-cancel')
      );
      window.location.href = checkoutUrl;
    } catch (e: unknown) {
      setTopUpSubmitting(false);
      toast.error('Could not create top-up session. Please try again.');
    }
  }, [buildReturnUrl]);

  const handleCancelSubscription = useCallback(async () => {
    try {
      await cancelSubscription(false);
      toast.success('Subscription will be cancelled at the end of the billing period');
      queryClient.invalidateQueries({ queryKey: ['billing', 'subscription'] });
    } catch {
      toast.error('Could not cancel subscription. Please try again.');
    }
  }, [queryClient]);

  const handleAddOnPurchase = useCallback(async (addonId: string) => {
    setAddOnCheckoutLoading(addonId);
    try {
      const { url } = await createStateFabricAddOnCheckout(
        addonId,
        buildReturnUrl('addon-success'),
        buildReturnUrl('addon-cancel')
      );
      window.location.href = url;
    } catch {
      toast.error('Could not create add-on checkout. Please try again.');
      setAddOnCheckoutLoading(null);
    }
  }, [buildReturnUrl]);

  const refreshAll = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: ['billing'] });
    queryClient.invalidateQueries({ queryKey: ['billing-wallet-info'] });
    queryClient.invalidateQueries({ queryKey: ['billing-wallet-transactions'] });
  }, [queryClient]);

  // Handle return URL params
  useEffect(() => {
    const billingParam = searchParams.get('billing');
    if (!billingParam) return;

    if (billingParam === 'wallet-success') {
      toast.success('Payment completed', {
        description: 'Your registry balance updates after Stripe confirms the payment.',
      });
      queryClient.invalidateQueries({ queryKey: ['billing-wallet-info'] });
    } else if (billingParam === 'wallet-cancel') {
      toast.message('Top-up cancelled');
    } else if (billingParam === 'addon-success') {
      toast.success('Add-on purchased', {
        description: 'Your add-on subscription is now active.',
      });
      queryClient.invalidateQueries({ queryKey: ['billing', 'state-fabric-add-ons-entitlements'] });
    } else if (billingParam === 'addon-cancel') {
      toast.message('Add-on purchase cancelled');
    }

    const next = new URLSearchParams(searchParams);
    next.delete('billing');
    setSearchParams(next, { replace: true });
  }, [searchParams, setSearchParams, queryClient]);

  return {
    state: {
      subscription,
      invoices,
      walletInfo,
      walletTransactions,
      paymentMethods,
      addOnCatalog,
      entitledAddOnIds,
      usageData: usageData ?? null,
      costData: costData ?? null,
    },
    actions: {
      openBillingPortal,
      handleTopUp,
      handleCancelSubscription,
      handleAddOnPurchase,
      refreshAll,
    },
    isLoading: {
      subscription: subscriptionLoading,
      invoices: invoicesLoading,
      wallet: walletLoading || walletTransactionsLoading,
      paymentMethods: paymentMethodsLoading,
      addOns: addOnCatalogLoading || addOnEntitlementsLoading,
      usage: usageLoading || costLoading,
    },
    errors: {
      subscription: subscriptionError as Error | null,
      invoices: invoicesError as Error | null,
      wallet: walletError as Error | null,
      paymentMethods: paymentMethodsError as Error | null,
      addOns: (addOnCatalogError || addOnEntitlementsError) as Error | null,
      usage: usageError as Error | null,
    },
    projectedBilling,
    usageMetrics,
    planLimits,
  };
}

export { useBillingHub };