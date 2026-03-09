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

// ============================================================================
// Tenant API
// ============================================================================

export const tenantApi = {
  list: async (params?: { limit?: number; offset?: number; status?: string }): Promise<{ tenants: Tenant[]; total: number }> => {
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

  createPricingTier: async (data: Omit<PricingTier, 'id' | 'created_at' | 'updated_at'>): Promise<PricingTier> => {
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

  createCoupon: async (data: Omit<Coupon, 'id' | 'current_uses' | 'created_at' | 'updated_at'>): Promise<Coupon> => {
    return apiClient.post<Coupon>('/admin/billing/coupons', data);
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
    return apiClient.get<any>(
      `/admin/analytics/overview${queryString ? `?${queryString}` : ''}`
    );
  },
};

export default {
  tenantApi,
  auditApi,
  healthApi,
  billingApi,
  analyticsApi,
};
