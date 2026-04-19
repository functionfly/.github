import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  getSubscription,
  listInvoices,
  getUsage,
  getWalletInfo,
  listPlatformFees,
  topUpWallet,
  createBillingPortalSession,
  createCheckoutSession,
  cancelSubscription,
  getBundles,
  getBundle,
  listStateFabricAddOnCatalog,
  getStateFabricAddOnEntitlements,
  createStateFabricAddOnCheckout,
  getFounderModeStatus,
  getDeferredBillingStatus,
  convertToPaid,
  type Bundle,
} from '@/api/billing';

// Query keys
export const billingKeys = {
  all: ['billing'] as const,
  subscription: () => [...billingKeys.all, 'subscription'] as const,
  invoices: (params?: { limit?: number; offset?: number }) =>
    [...billingKeys.all, 'invoices', params] as const,
  usage: (params?: { startDate?: string; endDate?: string }) =>
    [...billingKeys.all, 'usage', params] as const,
  wallet: () => [...billingKeys.all, 'wallet'] as const,
  fees: (params?: { limit?: number; offset?: number }) =>
    [...billingKeys.all, 'fees', params] as const,
  bundles: () => [...billingKeys.all, 'bundles'] as const,
  bundle: (slug: string) => [...billingKeys.all, 'bundle', slug] as const,
  stateFabricAddOns: () => [...billingKeys.all, 'state-fabric-addons'] as const,
  stateFabricEntitlements: () => [...billingKeys.all, 'state-fabric-entitlements'] as const,
  founderMode: () => [...billingKeys.all, 'founder-mode'] as const,
  deferredStatus: () => [...billingKeys.all, 'deferred-status'] as const,
};

// Get subscription
export function useSubscription() {
  return useQuery({
    queryKey: billingKeys.subscription(),
    queryFn: getSubscription,
    staleTime: 1000 * 60 * 5,
  });
}

// Get invoices
export function useInvoices(params?: { limit?: number; offset?: number }) {
  return useQuery({
    queryKey: billingKeys.invoices(params),
    queryFn: () => listInvoices(params?.limit, params?.offset),
    staleTime: 1000 * 60,
  });
}

// Get usage
export function useUsage(params?: { startDate?: string; endDate?: string }) {
  return useQuery({
    queryKey: billingKeys.usage(params),
    queryFn: () => getUsage(params?.startDate, params?.endDate),
    staleTime: 1000 * 60 * 5,
  });
}

// Get wallet info
export function useWallet() {
  return useQuery({
    queryKey: billingKeys.wallet(),
    queryFn: getWalletInfo,
    staleTime: 1000 * 60,
  });
}

// Get platform fees
export function usePlatformFees(params?: { limit?: number; offset?: number }) {
  return useQuery({
    queryKey: billingKeys.fees(params),
    queryFn: () => listPlatformFees(params?.limit, params?.offset),
    staleTime: 1000 * 60,
  });
}

// Create billing portal session
export function useCreateBillingPortal() {
  return useMutation({
    mutationFn: (returnUrl?: string) => createBillingPortalSession(returnUrl),
    onError: (error: Error) => {
      toast.error(`Failed to open billing portal: ${error.message}`);
    },
  });
}

// Create checkout session
export function useCreateCheckout() {
  return useMutation({
    mutationFn: ({
      priceId,
      successUrl,
      cancelUrl,
    }: {
      priceId: string;
      successUrl?: string;
      cancelUrl?: string;
    }) => createCheckoutSession(priceId, successUrl, cancelUrl),
    onError: (error: Error) => {
      toast.error(`Failed to create checkout: ${error.message}`);
    },
  });
}

// Cancel subscription
export function useCancelSubscription() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (immediately?: boolean) => cancelSubscription(immediately),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: billingKeys.subscription() });
      toast.success('Subscription cancelled successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to cancel subscription: ${error.message}`);
    },
  });
}

// Top up wallet
export function useTopUpWallet() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      amountUsd,
      successUrl,
      cancelUrl,
    }: {
      amountUsd: number;
      successUrl?: string;
      cancelUrl?: string;
    }) => topUpWallet(amountUsd, successUrl, cancelUrl),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: billingKeys.wallet() });
      toast.success('Wallet top-up initiated');
    },
    onError: (error: Error) => {
      toast.error(`Failed to top up wallet: ${error.message}`);
    },
  });
}

// Get bundles
export function useBundles() {
  return useQuery({
    queryKey: billingKeys.bundles(),
    queryFn: getBundles,
    staleTime: 1000 * 60 * 5,
  });
}

// Get single bundle
export function useBundle(slug: string) {
  return useQuery({
    queryKey: billingKeys.bundle(slug),
    queryFn: () => getBundle(slug),
    enabled: !!slug,
    staleTime: 1000 * 60 * 5,
  });
}

// Get State Fabric add-ons catalog
export function useStateFabricAddOns() {
  return useQuery({
    queryKey: billingKeys.stateFabricAddOns(),
    queryFn: listStateFabricAddOnCatalog,
    staleTime: 1000 * 60 * 5,
  });
}

// Get State Fabric entitlements
export function useStateFabricEntitlements() {
  return useQuery({
    queryKey: billingKeys.stateFabricEntitlements(),
    queryFn: getStateFabricAddOnEntitlements,
    staleTime: 1000 * 60,
  });
}

// Create State Fabric add-on checkout
export function useCreateStateFabricAddOnCheckout() {
  return useMutation({
    mutationFn: ({
      addonId,
      successUrl,
      cancelUrl,
    }: {
      addonId: string;
      successUrl?: string;
      cancelUrl?: string;
    }) => createStateFabricAddOnCheckout(addonId, successUrl, cancelUrl),
    onError: (error: Error) => {
      toast.error(`Failed to create checkout: ${error.message}`);
    },
  });
}

// Get founder mode status
export function useFounderModeStatus() {
  return useQuery({
    queryKey: billingKeys.founderMode(),
    queryFn: getFounderModeStatus,
    staleTime: 1000 * 60,
  });
}

// Get deferred billing status
export function useDeferredBillingStatus() {
  return useQuery({
    queryKey: billingKeys.deferredStatus(),
    queryFn: getDeferredBillingStatus,
    staleTime: 1000 * 60,
  });
}

// Convert founder mode to paid
export function useConvertToPaid() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (bundleId: string) => convertToPaid(bundleId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: billingKeys.founderMode() });
      queryClient.invalidateQueries({ queryKey: billingKeys.subscription() });
      toast.success('Converting to paid subscription');
    },
    onError: (error: Error) => {
      toast.error(`Failed to convert: ${error.message}`);
    },
  });
}
