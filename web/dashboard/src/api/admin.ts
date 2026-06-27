import { apiClient } from './client';

// ============================================================================
// Types
// ============================================================================

export interface Tenant {
  id: string;
  name: string;
  plan: string;
  status: 'active' | 'suspended';
  created_at: string;
  updated_at: string;
}

export interface AuditEvent {
  id: string;
  actor_user_id?: string;
  actor_email?: string;
  tenant_id?: string;
  action: string;
  resource_type: string;
  resource_id?: string;
  request_id?: string;
  before_state?: any;
  after_state?: any;
  ip_address?: string;
  user_agent?: string;
  timestamp: string;
  success: boolean;
}

export interface SystemHealthService {
  name: string;
  status: 'healthy' | 'unhealthy' | 'degraded';
  latency_ms?: number;
  uptime_percent?: number;
}

export interface SystemHealth {
  status: 'healthy' | 'unhealthy';
  version?: string;
  timestamp?: string;
  checks?: {
    database: { status: string; healthy: boolean; response_time_ms?: number };
    api: { status: string; healthy: boolean };
    repository: { status: string; healthy: boolean; response_time_ms?: number };
    system: { status: string; healthy: boolean; goroutines?: number };
  };
  services?: SystemHealthService[];
}

export interface PricingTier {
  id: string;
  name: string;
  description: string;
  price_cents: number;
  currency: string;
  features: any;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Subscription {
  id: string;
  tenant_id: string;
  pricing_tier_id: string;
  status: string;
  current_period_start: string;
  current_period_end: string;
  cancel_at_period_end: boolean;
  created_at: string;
  updated_at: string;
}

export interface Invoice {
  id: string;
  tenant_id: string;
  subscription_id?: string;
  status: string;
  amount_due_cents: number;
  amount_paid_cents: number;
  currency: string;
  period_start?: string;
  period_end?: string;
  paid_at?: string;
  created_at: string;
  updated_at: string;
}

export interface Coupon {
  id: string;
  code: string;
  name: string;
  description: string;
  discount_type: 'percent' | 'fixed';
  discount_value: number;
  currency?: string;
  max_uses?: number;
  current_uses: number;
  starts_at?: string;
  expires_at?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface AffiliateCode {
  id: string;
  code: string;
  publisher_id: string;
  tenant_id?: string;
  name: string;
  description?: string;
  commission_type: 'percent' | 'fixed';
  commission_value: number;
  max_commissions?: number;
  max_referrals?: number;
  total_referrals: number;
  total_commissions: number;
  pending_commissions: number;
  pending_earnings_cents: number;
  total_earnings_cents: number;
  paid_out_earnings_cents: number;
  valid_from?: string;
  valid_until?: string;
  is_active: boolean;
  utm_source?: string;
  utm_campaign?: string;
  created_at: string;
  updated_at: string;
}

export interface AffiliateReferral {
  id: string;
  affiliate_code_id: string;
  referred_tenant_id: string;
  subscription_id?: string;
  utm_source?: string;
  utm_campaign?: string;
  utm_content?: string;
  utm_term?: string;
  ip_address?: string;
  user_agent?: string;
  status: 'pending' | 'converted' | 'qualified' | 'canceled';
  referred_at: string;
  converted_at?: string;
  created_at: string;
  updated_at: string;
}

export interface AffiliateCommission {
  id: string;
  affiliate_code_id: string;
  referral_id: string;
  commission_type: 'percent' | 'fixed';
  commission_value: number;
  base_amount_cents: number;
  base_amount_usd: number;
  commission_cents: number;
  commission_usd: number;
  status: 'pending' | 'approved' | 'paid' | 'canceled';
  paid_at?: string;
  payment_batch_id?: string;
  payment_batch?: string;
  subscription_id?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

// ============================================================================
// Tenant API
// ============================================================================

export const tenantApi = {
  list: async (params?: {
    limit?: number;
    offset?: number;
    status?: string;
  }): Promise<{ tenants: Tenant[]; total: number }> => {
    const searchParams = new URLSearchParams();
    if (params?.limit) searchParams.set('limit', String(params.limit));
    if (params?.offset) searchParams.set('offset', String(params.offset));
    if (params?.status) searchParams.set('status', params.status);
    const queryString = searchParams.toString();
    return apiClient.get<{ tenants: Tenant[]; total: number }>(
      `/admin/tenants${queryString ? `?${queryString}` : ''}`
    );
  },

  get: async (id: string): Promise<Tenant> => {
    return apiClient.get<Tenant>(`/admin/tenants/${id}`);
  },

  update: async (id: string, data: Partial<Tenant>): Promise<Tenant> => {
    return apiClient.put<Tenant>(`/admin/tenants/${id}`, data);
  },

  suspend: async (id: string): Promise<Tenant> => {
    return apiClient.post<Tenant>(`/admin/tenants/${id}/suspend`);
  },

  activate: async (id: string): Promise<Tenant> => {
    return apiClient.post<Tenant>(`/admin/tenants/${id}/activate`);
  },
};

// ============================================================================
// Audit API
// ============================================================================

export const auditApi = {
  list: async (params?: {
    limit?: number;
    offset?: number;
    actor_user_id?: string;
    tenant_id?: string;
    action?: string;
    resource_type?: string;
  }): Promise<{ events: AuditEvent[]; total: number }> => {
    const searchParams = new URLSearchParams();
    if (params?.limit) searchParams.set('limit', String(params.limit));
    if (params?.offset) searchParams.set('offset', String(params.offset));
    if (params?.actor_user_id) searchParams.set('actor_user_id', params.actor_user_id);
    if (params?.tenant_id) searchParams.set('tenant_id', params.tenant_id);
    if (params?.action) searchParams.set('action', params.action);
    if (params?.resource_type) searchParams.set('resource_type', params.resource_type);
    const queryString = searchParams.toString();
    return apiClient.get<{ events: AuditEvent[]; total: number }>(
      `/admin/audit${queryString ? `?${queryString}` : ''}`
    );
  },

  get: async (id: string): Promise<AuditEvent> => {
    return apiClient.get<AuditEvent>(`/admin/audit/${id}`);
  },
};

// ============================================================================
// Health API
// ============================================================================

export const healthApi = {
  getStatus: async (): Promise<SystemHealth> => {
    return apiClient.get<SystemHealth>('/admin/health');
  },
};

// ============================================================================
// Billing API
// ============================================================================

export const billingApi = {
  listPricingTiers: async (): Promise<PricingTier[]> => {
    return apiClient.get<PricingTier[]>('/admin/billing/tiers');
  },

  createPricingTier: async (
    data: Omit<PricingTier, 'id' | 'created_at' | 'updated_at'>
  ): Promise<PricingTier> => {
    return apiClient.post<PricingTier>('/admin/billing/tiers', data);
  },

  updatePricingTier: async (id: string, data: Partial<PricingTier>): Promise<PricingTier> => {
    return apiClient.put<PricingTier>(`/admin/billing/tiers/${id}`, data);
  },

  deletePricingTier: async (id: string): Promise<void> => {
    return apiClient.delete(`/admin/billing/tiers/${id}`);
  },

  listInvoices: async (tenantId: string): Promise<Invoice[]> => {
    return apiClient.get<Invoice[]>(`/admin/billing/invoices?tenant_id=${tenantId}`);
  },

  listCoupons: async (): Promise<Coupon[]> => {
    return apiClient.get<Coupon[]>('/admin/billing/coupons');
  },

  createCoupon: async (
    data: Omit<Coupon, 'id' | 'current_uses' | 'created_at' | 'updated_at'>
  ): Promise<Coupon> => {
    return apiClient.post<Coupon>('/admin/billing/coupons', data);
  },

  // Affiliate Codes
  listAffiliateCodes: async (): Promise<AffiliateCode[]> => {
    return apiClient.get<AffiliateCode[]>('/admin/billing/affiliate-codes');
  },

  createAffiliateCode: async (data: {
    code: string;
    publisher_id: string;
    tenant_id?: string;
    name: string;
    description?: string;
    commission_type: 'percent' | 'fixed';
    commission_value: number;
    max_commissions?: number;
    max_referrals?: number;
    valid_from?: string;
    valid_until?: string;
    utm_source?: string;
    utm_campaign?: string;
  }): Promise<AffiliateCode> => {
    return apiClient.post<AffiliateCode>('/admin/billing/affiliate-codes', data);
  },

  getAffiliateCode: async (id: string): Promise<AffiliateCode> => {
    return apiClient.get<AffiliateCode>(`/admin/billing/affiliate-codes/${id}`);
  },

  updateAffiliateCode: async (id: string, data: Partial<AffiliateCode>): Promise<AffiliateCode> => {
    return apiClient.put<AffiliateCode>(`/admin/billing/affiliate-codes/${id}`, data);
  },

  listAffiliateReferrals: async (codeId: string): Promise<AffiliateReferral[]> => {
    return apiClient.get<AffiliateReferral[]>(`/admin/billing/affiliate-codes/${codeId}/referrals`);
  },

  listAffiliateCommissions: async (codeId: string): Promise<AffiliateCommission[]> => {
    return apiClient.get<AffiliateCommission[]>(
      `/admin/billing/affiliate-codes/${codeId}/commissions`
    );
  },

  recordAffiliateReferral: async (data: {
    affiliate_code: string;
    tenant_id: string;
    subscription_id?: string;
    utm_source?: string;
    utm_campaign?: string;
    utm_content?: string;
    utm_term?: string;
    ip_address?: string;
    user_agent?: string;
  }): Promise<AffiliateReferral> => {
    return apiClient.post<AffiliateReferral>('/admin/billing/affiliate-referrals', data);
  },

  updateAffiliateReferralStatus: async (id: string, status: string): Promise<void> => {
    return apiClient.patch(`/admin/billing/affiliate-referrals/${id}/status`, { status });
  },

  approveAffiliateCommission: async (id: string): Promise<void> => {
    return apiClient.post(`/admin/billing/affiliate-commissions/${id}/approve`);
  },

  markAffiliateCommissionPaid: async (id: string): Promise<void> => {
    return apiClient.post(`/admin/billing/affiliate-commissions/${id}/paid`);
  },

  calculateAffiliateCommission: async (
    affiliateCode: string,
    baseAmountCents: number
  ): Promise<{
    affiliate_code: string;
    commission_type: string;
    commission_value: number;
    base_amount_cents: number;
    base_amount_usd: number;
    commission_cents: number;
    commission_usd: number;
  }> => {
    return apiClient.post(`/admin/billing/affiliate-commissions/calculate`, {
      affiliate_code: affiliateCode,
      base_amount_cents: baseAmountCents,
    });
  },
};

// ============================================================================
// City Wars Admin API
// ============================================================================

export interface WarMatch {
  id: number;
  war_id: number;
  round: 'quarterfinal' | 'semifinal' | 'final';
  position: number;
  metro_a_id: number;
  metro_a_name: string;
  metro_a_country: string;
  score_a: number;
  active_users_a: number;
  metro_b_id: number;
  metro_b_name: string;
  metro_b_country: string;
  score_b: number;
  active_users_b: number;
  winner_metro_id: number | null;
  status: 'pending' | 'in_progress' | 'completed' | 'cancelled';
  scheduled_at: string;
  completed_at: string | null;
}

export interface War {
  id: number;
  name: string;
  season: string;
  slug: string;
  status: 'draft' | 'scheduled' | 'active' | 'complete' | 'cancelled';
  starts_at: string;
  ends_at: string;
  round: 'quarterfinal' | 'semifinal' | 'final' | 'complete' | null;
  quarterfinals: WarMatch[];
  semifinals: WarMatch[];
  final: WarMatch | null;
  champion_name: string | null;
  total_active_users: number;
  created_at: string;
  updated_at: string;
  created_by: string;
}

export interface EligibleMetro {
  id: number;
  name: string;
  metro_slug: string;
  country_code: string;
  score: number;
  active_users: number;
  deployments: number;
  rank_position: number;
  is_eligible: boolean;
  eligibility_reason?: string;
}

export interface MatchResultRequest {
  match_id: number;
  metro_a_score: number;
  metro_b_score: number;
  winner_metro_id: number;
  recorded_by: string;
  recorded_at: string;
}

export interface WarBracket {
  war_id: number;
  quarterfinals: WarMatch[];
  semifinals: WarMatch[];
  final: WarMatch | null;
  champion: string | null;
}

export const cityWarsApi = {
  listWars: async (params?: {
    status?: string;
    limit?: number;
    offset?: number;
  }): Promise<{ wars: War[]; total: number }> => {
    const searchParams = new URLSearchParams();
    if (params?.status) searchParams.set('status', params.status);
    if (params?.limit) searchParams.set('limit', String(params.limit));
    if (params?.offset) searchParams.set('offset', String(params.offset));
    const queryString = searchParams.toString();
    return apiClient.get<{ wars: War[]; total: number }>(
      `/admin/city-wars${queryString ? `?${queryString}` : ''}`
    );
  },

  getWar: async (id: number): Promise<War> => {
    return apiClient.get<War>(`/admin/city-wars/${id}`);
  },

  createWar: async (data: {
    name: string;
    season: string;
    slug: string;
    starts_at: string;
    ends_at: string;
  }): Promise<War> => {
    return apiClient.post<War>('/admin/city-wars', data);
  },

  activateWar: async (id: number): Promise<War> => {
    return apiClient.post<War>(`/admin/city-wars/${id}/activate`);
  },

  cancelWar: async (id: number): Promise<War> => {
    return apiClient.post<War>(`/admin/city-wars/${id}/cancel`);
  },

  generateBracket: async (id: number): Promise<WarBracket> => {
    return apiClient.post<WarBracket>(`/admin/city-wars/${id}/bracket/generate`);
  },

  setQuarterfinals: async (
    id: number,
    matches: Omit<WarMatch, 'id' | 'war_id' | 'status' | 'completed_at'>[]
  ): Promise<WarBracket> => {
    return apiClient.post<WarBracket>(`/admin/city-wars/${id}/bracket/quarterfinals`, { matches });
  },

  advanceWar: async (id: number, round: 'semifinal' | 'final'): Promise<WarBracket> => {
    return apiClient.post<WarBracket>(`/admin/city-wars/${id}/bracket/advance`, { round });
  },

  overrideMatch: async (id: number, data: MatchResultRequest): Promise<WarMatch> => {
    return apiClient.post<WarMatch>(`/admin/city-wars/${id}/matches/override`, data);
  },

  recordMatch: async (id: number, data: MatchResultRequest): Promise<WarMatch> => {
    return apiClient.post<WarMatch>(`/admin/city-wars/${id}/matches/record`, data);
  },

  getEligibleMetros: async (): Promise<EligibleMetro[]> => {
    return apiClient.get<EligibleMetro[]>('/admin/city-wars/eligible-metros');
  },

  getWarBracket: async (id: number): Promise<WarBracket> => {
    return apiClient.get<WarBracket>(`/admin/city-wars/${id}/bracket`);
  },
};

// ============================================================================
// Analytics API
// ============================================================================

export const analyticsApi = {
  getOverview: async (params?: { start_date?: string; end_date?: string }): Promise<any> => {
    const searchParams = new URLSearchParams();
    if (params?.start_date) searchParams.set('start_date', params.start_date);
    if (params?.end_date) searchParams.set('end_date', params.end_date);
    const queryString = searchParams.toString();
    return apiClient.get<any>(`/admin/analytics/overview${queryString ? `?${queryString}` : ''}`);
  },
};

export default {
  tenantApi,
  auditApi,
  healthApi,
  billingApi,
  analyticsApi,
  cityWarsApi,
};
