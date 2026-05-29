import { apiClient } from '@/api/client';
import { useQuery } from '@tanstack/react-query';

export type FunctionPurchaseRow = {
  id: string;
  agentId: string;
  functionAuthor: string;
  functionName: string;
  pricePaidUsd: number;
  status: string;
  createdAt: number;
};

export type AgentHiringRow = {
  id: string;
  agentId: string;
  taskType: string;
  budgetUsd: number;
  status: string;
  createdAt: number;
  taskPayload?: Record<string, unknown>;
};

export type BuyerLicenseRow = {
  id: string;
  type: string;
  functionId: string;
  functionName: string;
  issuerTenantId?: string;
  purchaserName: string;
  issuedAt: number;
  expiresAt?: number;
  maxActivations?: number;
  activationCount: number;
  revoked: boolean;
  keyPrefix?: string;
};

export type BuyerSubscriptionRow = {
  id: string;
  planName: string;
  creatorTenantId?: string;
  status: string;
  amount: number;
  currency: string;
  billingCycle: string;
  currentPeriodStart: number;
  currentPeriodEnd: number;
  cancelAtPeriodEnd: boolean;
};

export type MarketplacePurchasesResponse = {
  enabled: boolean;
  functions: FunctionPurchaseRow[];
  agents: AgentHiringRow[];
  licenses: BuyerLicenseRow[];
  subscriptions: BuyerSubscriptionRow[];
  totalFunctions: number;
  totalAgents: number;
  totalLicenses: number;
  totalSubscriptions: number;
  limit?: number;
  offset?: number;
};

export function useMarketplacePurchases(limit = 50, offset = 0) {
  return useQuery({
    queryKey: ['marketplace', 'purchases', limit, offset],
    queryFn: async () => {
      const params = new URLSearchParams({
        limit: String(limit),
        offset: String(offset),
      });
      return apiClient.get<MarketplacePurchasesResponse>(
        `/v1/marketplace/purchases?${params.toString()}`
      );
    },
    staleTime: 60_000,
  });
}
