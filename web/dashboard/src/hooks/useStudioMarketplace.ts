import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { API_BASE_URL } from '@/lib/constants';

export interface MarketplaceFunction {
  id: string;
  name: string;
  description: string;
  author: string;
  version: string;
  category: string;
  downloads: number;
  rating: number;
  price: number;
  isFavorite?: boolean;
  runtime: string;
  triggers: string[];
}

export interface FunctionExecutionRequest {
  functionId: string;
  input?: Record<string, unknown>;
}

export interface SubscriptionPlan {
  id: string;
  name: string;
  price: number;
  features: string[];
  subscribers: number;
}

export interface RoyaltyEntry {
  functionId: string;
  functionName: string;
  period: string;
  calls: number;
  royaltyRate: number;
  earnings: number;
  paid: boolean;
}

export interface MarketplaceLicense {
  id: string;
  key?: string;
  type: 'open' | 'restricted' | 'commercial';
  functionId: string;
  functionName: string;
  purchaserId: string;
  purchaserName: string;
  maxActivations?: number;
  expiresAt?: string;
  createdAt: string;
  revokedAt?: string | null;
}

async function marketplaceFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    credentials: 'include',
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Request failed' }));
    throw new Error(error.message || `HTTP ${response.status}`);
  }

  return response.json();
}

export function useMarketplaceFunctions(filters?: { search?: string; category?: string }) {
  const params = new URLSearchParams();
  if (filters?.search) params.set('q', filters.search);
  if (filters?.category) params.set('category', filters.category);

  return useQuery({
    queryKey: ['marketplace', 'functions', filters],
    queryFn: () => marketplaceFetch<{ functions: MarketplaceFunction[] }>(`/v1/marketplace/functions?${params.toString()}`),
    staleTime: 1000 * 60,
  });
}

export function useExecuteFunction() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: FunctionExecutionRequest) =>
      marketplaceFetch<{ executionId: string; status: string }>(`/v1/marketplace/functions/${data.functionId}/execute`, {
        method: 'POST',
        body: JSON.stringify({ input: data.input }),
      }),
    onSuccess: () => {
      toast.success('Function executed');
    },
    onError: (error: Error) => {
      toast.error(`Failed to execute function: ${error.message}`);
    },
  });
}

export function useFavoriteFunction() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: { functionId: string; favorite: boolean }) =>
      marketplaceFetch<{ ok: boolean }>(`/v1/marketplace/functions/${params.functionId}/favorite`, {
        method: params.favorite ? 'POST' : 'DELETE',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['marketplace'] });
      toast.success('Favorite updated');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update favorite: ${error.message}`);
    },
  });
}

export function useSubscriptionPlans() {
  return useQuery({
    queryKey: ['marketplace', 'plans'],
    queryFn: () => marketplaceFetch<{ plans: SubscriptionPlan[] }>('/v1/marketplace/plans'),
    staleTime: 1000 * 60 * 5,
  });
}

export function useCreatePlan() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: Omit<SubscriptionPlan, 'id' | 'subscribers'>) =>
      marketplaceFetch<SubscriptionPlan>('/v1/marketplace/plans', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['marketplace', 'plans'] });
      toast.success('Plan created');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create plan: ${error.message}`);
    },
  });
}

export function useEditPlan() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: SubscriptionPlan) =>
      marketplaceFetch<SubscriptionPlan>(`/v1/marketplace/plans/${data.id}`, {
        method: 'PUT',
        body: JSON.stringify(data),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['marketplace', 'plans'] });
      toast.success('Plan updated');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update plan: ${error.message}`);
    },
  });
}

export function useRoyalties() {
  return useQuery({
    queryKey: ['marketplace', 'royalties'],
    queryFn: () => marketplaceFetch<{ royalties: RoyaltyEntry[]; totalEarnings: number; pendingPayout: number }>('/v1/marketplace/royalties'),
    staleTime: 1000 * 60,
  });
}

export function useRequestPayout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () =>
      marketplaceFetch<{ ok: boolean }>('/v1/marketplace/royalties/payout', {
        method: 'POST',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['marketplace', 'royalties'] });
      toast.success('Payout requested');
    },
    onError: (error: Error) => {
      toast.error(`Failed to request payout: ${error.message}`);
    },
  });
}

export function useUpdateLicense() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: { functionId: string; license: string }) =>
      marketplaceFetch<{ ok: boolean }>(`/v1/marketplace/functions/${params.functionId}/license`, {
        method: 'PUT',
        body: JSON.stringify({ license: params.license }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['marketplace'] });
      toast.success('License updated');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update license: ${error.message}`);
    },
  });
}

export function useUpdatePricing() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: { functionId: string; price: number; model: 'per_call' | 'subscription' | 'usage' }) =>
      marketplaceFetch<{ ok: boolean }>(`/v1/marketplace/functions/${params.functionId}/pricing`, {
        method: 'PUT',
        body: JSON.stringify({ price: params.price, pricing_model: params.model }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['marketplace'] });
      toast.success('Pricing updated');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update pricing: ${error.message}`);
    },
  });
}

export function useMarketplaceLicenses() {
  return useQuery({
    queryKey: ['marketplace', 'licenses'],
    queryFn: () => marketplaceFetch<{ licenses: MarketplaceLicense[] }>('/v1/marketplace/licenses'),
    staleTime: 1000 * 60,
  });
}

export function useCreateLicense() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: Omit<MarketplaceLicense, 'id' | 'key' | 'createdAt' | 'revokedAt'>) =>
      marketplaceFetch<MarketplaceLicense>('/v1/marketplace/licenses', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['marketplace', 'licenses'] });
      toast.success('License created');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create license: ${error.message}`);
    },
  });
}

export function useRevokeLicense() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (licenseId: string) =>
      marketplaceFetch<{ ok: boolean }>(`/v1/marketplace/licenses/${licenseId}/revoke`, {
        method: 'POST',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['marketplace', 'licenses'] });
      toast.success('License revoked');
    },
    onError: (error: Error) => {
      toast.error(`Failed to revoke license: ${error.message}`);
    },
  });
}