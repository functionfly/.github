/**
 * Admin Billing Page
 * Manage billing, subscriptions, invoicing, and pricing tiers
 */

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { Download, Search, DollarSign, TrendingUp, AlertCircle, Pencil, X, Check, Plus, Gift, Users, Eye, Sparkles, Clock, BarChart3, UsersRound } from 'lucide-react';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { toast } from 'sonner';
import { RevenueRecognitionSection } from '@/components/revenue/RevenueRecognitionSection';
import { format } from 'date-fns';

interface BillingData {
  totalRevenue: number;
  activeSubscriptions: number;
  pendingInvoices: number;
  overdue: number;
}

interface PricingTier {
  id: string;
  name: string;
  description: string;
  price_cents: number;
  annual_price_cents: number | null;
  currency: string;
  billing_cycle: string;
  is_active: boolean;
  stripe_price_id: string | null;
  stripe_price_id_annual: string | null;
}

interface AffiliateCode {
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

interface AffiliateReferral {
  id: string;
  affiliate_code_id: string;
  referred_tenant_id: string;
  subscription_id?: string;
  utm_source?: string;
  utm_campaign?: string;
  status: 'pending' | 'converted' | 'qualified' | 'canceled';
  referred_at: string;
  converted_at?: string;
  created_at: string;
}

interface AffiliateCommission {
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
}

interface FounderModeBundleAnalytics {
  bundle_slug: string;
  total_signups: number;
  active: number;
  converted: number;
  revenue_cents: number;
  conversion_rate: number;
}

interface FounderModeTypeAnalytics {
  mode_type: 'time_based' | 'revenue_based' | 'hybrid';
  count: number;
  converted: number;
  conversion_rate: number;
}

interface FounderModeAnalytics {
  total_signups: number;
  active_founders: number;
  converted_to_paid: number;
  expired_or_canceled: number;
  conversion_rate: number;
  total_revenue_cents: number;
  avg_days_to_convert: number;
  by_bundle: FounderModeBundleAnalytics[];
  by_mode_type: FounderModeTypeAnalytics[];
  recent_signups_30d: number;
}

const MODE_TYPE_LABELS: Record<string, { label: string; description: string; icon: typeof Clock }> = {
  time_based: { label: 'Free 3 Months', description: 'Free for 3 months (Founder Mode)', icon: Clock },
  revenue_based: { label: 'Free until $1K MRR', description: 'Free until $1,000 monthly revenue', icon: DollarSign },
  hybrid: { label: 'Free until 100 Users', description: 'Free until 100 users', icon: UsersRound },
};

export function AdminBillingPage() {
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [tenantIdInput, setTenantIdInput] = useState('');
  const [selectedAddon, setSelectedAddon] = useState('hot_cache_booster');
  const [selectedStatus, setSelectedStatus] = useState('active');
  const [overrideBusy, setOverrideBusy] = useState(false);
  const [editingTierId, setEditingTierId] = useState<string | null>(null);
  const [editForm, setEditForm] = useState<{ price_cents: string; annual_price_cents: string }>({ price_cents: '', annual_price_cents: '' });

  // Affiliate/Affiliate Codes state
  const [affiliateTab, setAffiliateTab] = useState<'codes' | 'referrals' | 'commissions'>('codes');
  const [selectedCodeId, setSelectedCodeId] = useState<string | null>(null);
  const [showCreateCodeModal, setShowCreateCodeModal] = useState(false);
  const [createCodeForm, setCreateCodeForm] = useState({
    code: '',
    publisher_id: '',
    name: '',
    description: '',
    commission_type: 'percent' as 'percent' | 'fixed',
    commission_value: '',
    max_commissions: '',
    max_referrals: '',
  });

  const queryClient = useQueryClient();

  // Fetch billing data
  const { data: billingResponse, isLoading, isError } = useQuery({
    queryKey: ['admin-billing'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<BillingData>('/billing/summary');
      } catch {
        return { data: null, success: false };
      }
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
  });

  // Fetch pricing tiers
  const { data: tiersResponse, isLoading: tiersLoading } = useQuery({
    queryKey: ['admin-pricing-tiers'],
    queryFn: async () => {
      try {
        const response = await adminApiClient.get<{ tiers: PricingTier[] }>('/billing/tiers');
        return response;
      } catch {
        return { data: null, success: false };
      }
    },
    staleTime: 1000 * 60 * 5,
  });

  // Fetch founder mode analytics
  const { data: founderModeResponse, isLoading: founderModeLoading } = useQuery({
    queryKey: ['admin-founder-mode-analytics'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<FounderModeAnalytics>('/billing/founder-mode-analytics');
      } catch {
        return { data: null, success: false };
      }
    },
    staleTime: 1000 * 60 * 5,
  });

  const updateTierMutation = useMutation({
    mutationFn: async ({ tierId, updates }: { tierId: string; updates: Record<string, unknown> }) => {
      return adminApiClient.patch(`/billing/tiers/${tierId}`, updates);
    },
    onSuccess: () => {
      toast.success('Pricing tier updated successfully');
      queryClient.invalidateQueries({ queryKey: ['admin-pricing-tiers'] });
      setEditingTierId(null);
    },
    onError: () => {
      toast.error('Failed to update pricing tier');
    },
  });

  const billingData = billingResponse?.data;
  const { data: addonCatalog } = useQuery({
    queryKey: ['admin-billing-sf-addon-catalog'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<Array<{ id: string; name: string }>>(
          '/billing/state-fabric-add-ons/catalog'
        );
      } catch {
        return { data: [] };
      }
    },
  });
  const { data: tenantEntitlements, refetch: refetchEntitlements } = useQuery({
    queryKey: ['admin-billing-sf-tenant-entitlements', tenantIdInput],
    enabled: /^[0-9a-fA-F-]{36}$/.test(tenantIdInput),
    queryFn: async () => {
      try {
        return await adminApiClient.get<
          Array<{ addon_id: string; status: string; stripe_subscription_id?: string | null }>
        >(`/billing/state-fabric-add-ons/entitlements/${tenantIdInput}`);
      } catch {
        return { data: [] };
      }
    },
  });

  // Affiliate Codes query
  const { data: affiliateCodesResponse, isLoading: affiliateCodesLoading } = useQuery({
    queryKey: ['admin-affiliate-codes'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<{ affiliate_codes: AffiliateCode[] }>('/billing/affiliate-codes');
      } catch {
        return { data: { affiliate_codes: [] } };
      }
    },
    enabled: affiliateTab === 'codes',
  });

  // Affiliate Referrals query
  const { data: affiliateReferralsResponse, isLoading: affiliateReferralsLoading } = useQuery({
    queryKey: ['admin-affiliate-referrals', selectedCodeId],
    queryFn: async () => {
      if (!selectedCodeId) return { data: { referrals: [] } };
      try {
        return await adminApiClient.get<{ referrals: AffiliateReferral[] }>(`/billing/affiliate-codes/${selectedCodeId}/referrals`);
      } catch {
        return { data: { referrals: [] } };
      }
    },
    enabled: affiliateTab === 'referrals' && !!selectedCodeId,
  });

  // Affiliate Commissions query
  const { data: affiliateCommissionsResponse, isLoading: affiliateCommissionsLoading } = useQuery({
    queryKey: ['admin-affiliate-commissions', selectedCodeId],
    queryFn: async () => {
      if (!selectedCodeId) return { data: { commissions: [] } };
      try {
        return await adminApiClient.get<{ commissions: AffiliateCommission[] }>(`/billing/affiliate-codes/${selectedCodeId}/commissions`);
      } catch {
        return { data: { commissions: [] } };
      }
    },
    enabled: affiliateTab === 'commissions' && !!selectedCodeId,
  });

  // Create Affiliate Code mutation
  const createAffiliateCodeMutation = useMutation({
    mutationFn: async (data: typeof createCodeForm) => {
      return adminApiClient.post('/billing/affiliate-codes', {
        code: data.code,
        publisher_id: data.publisher_id,
        name: data.name,
        description: data.description,
        commission_type: data.commission_type,
        commission_value: parseFloat(data.commission_value) || 0,
        max_commissions: data.max_commissions ? parseInt(data.max_commissions) : undefined,
        max_referrals: data.max_referrals ? parseInt(data.max_referrals) : undefined,
      });
    },
    onSuccess: () => {
      toast.success('Affiliate code created successfully');
      queryClient.invalidateQueries({ queryKey: ['admin-affiliate-codes'] });
      setShowCreateCodeModal(false);
      setCreateCodeForm({ code: '', publisher_id: '', name: '', description: '', commission_type: 'percent', commission_value: '', max_commissions: '', max_referrals: '' });
    },
    onError: () => {
      toast.error('Failed to create affiliate code');
    },
  });

  // Update Affiliate Code mutation
  const updateAffiliateCodeMutation = useMutation({
    mutationFn: async ({ id, updates }: { id: string; updates: Partial<AffiliateCode> }) => {
      return adminApiClient.put(`/billing/affiliate-codes/${id}`, updates);
    },
    onSuccess: () => {
      toast.success('Affiliate code updated');
      queryClient.invalidateQueries({ queryKey: ['admin-affiliate-codes'] });
    },
    onError: () => {
      toast.error('Failed to update affiliate code');
    },
  });

  // Approve Commission mutation
  const approveCommissionMutation = useMutation({
    mutationFn: async (commissionId: string) => {
      return adminApiClient.post(`/billing/affiliate-commissions/${commissionId}/approve`);
    },
    onSuccess: () => {
      toast.success('Commission approved');
      queryClient.invalidateQueries({ queryKey: ['admin-affiliate-commissions'] });
    },
    onError: () => {
      toast.error('Failed to approve commission');
    },
  });

  // Mark Commission Paid mutation
  const markPaidMutation = useMutation({
    mutationFn: async (commissionId: string) => {
      return adminApiClient.post(`/billing/affiliate-commissions/${commissionId}/paid`);
    },
    onSuccess: () => {
      toast.success('Commission marked as paid');
      queryClient.invalidateQueries({ queryKey: ['admin-affiliate-commissions'] });
    },
    onError: () => {
      toast.error('Failed to mark commission as paid');
    },
  });

  if (isLoading) {
    return <LoadingScreen />;
  }

  if (isError) {
    return (
      <div className="p-8 bg-red-50 border border-red-200 rounded-lg">
        <h3 className="font-semibold text-red-900">Error loading billing data</h3>
        <p className="text-red-700 mt-2">Failed to fetch billing information.</p>
      </div>
    );
  }

  const handleExportInvoices = () => {
    const link = document.createElement('a');
    link.href = '/api/admin/billing/export';
    link.download = `invoices-${new Date().toISOString().split('T')[0]}.csv`;
    link.click();
  };

  const handleOverrideEntitlement = async () => {
    if (!/^[0-9a-fA-F-]{36}$/.test(tenantIdInput)) {
      toast.error('Enter a valid tenant UUID');
      return;
    }
    try {
      setOverrideBusy(true);
      await adminApiClient.patch(
        `/billing/state-fabric-add-ons/entitlements/${tenantIdInput}/${selectedAddon}`,
        { status: selectedStatus }
      );
      toast.success('Add-on entitlement updated');
      await refetchEntitlements();
    } catch {
      toast.error('Failed to update entitlement');
    } finally {
      setOverrideBusy(false);
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Billing & Invoicing</h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">Manage subscriptions, invoices, and revenue</p>
        </div>

        <button
          onClick={handleExportInvoices}
          className="flex items-center gap-2 px-4 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 transition-colors"
        >
          <Download className="w-5 h-5" />
          Export Invoices
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-600 dark:text-gray-400 text-sm">Total Revenue</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                ${billingData?.totalRevenue.toLocaleString() || '0'}
              </p>
            </div>
            <DollarSign className="w-8 h-8 text-green-600" />
          </div>
        </div>

        <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-600 dark:text-gray-400 text-sm">Active Subscriptions</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                {billingData?.activeSubscriptions || '0'}
              </p>
            </div>
            <TrendingUp className="w-8 h-8 text-blue-600" />
          </div>
        </div>

        <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-600 dark:text-gray-400 text-sm">Pending Invoices</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                {billingData?.pendingInvoices || '0'}
              </p>
            </div>
            <AlertCircle className="w-8 h-8 text-yellow-600" />
          </div>
        </div>

        <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-600 dark:text-gray-400 text-sm">Overdue</p>
              <p className="text-2xl font-bold text-red-600 dark:text-red-400">
                ${billingData?.overdue || '0'}
              </p>
            </div>
            <AlertCircle className="w-8 h-8 text-red-600" />
          </div>
        </div>
      </div>

      {/* Founder Mode Bundle Analytics */}
      {founderModeLoading ? (
        <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-8 text-center">
          <div className="text-gray-500 dark:text-gray-400">Loading founder mode analytics...</div>
        </div>
      ) : founderModeResponse?.data ? (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Sparkles className="w-6 h-6 text-amber-500" />
              <div>
                <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Founder Mode Bundle Analytics</h2>
                <p className="text-sm text-gray-600 dark:text-gray-400">Tracking free tier conversions: 3-month, $1K MRR, and 100-user limits</p>
              </div>
            </div>
            <div className="text-sm text-gray-500 dark:text-gray-400">
              {founderModeResponse.data.recent_signups_30d} signups in last 30 days
            </div>
          </div>

          {/* Founder Mode Stats */}
          <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
            <div className="bg-gradient-to-br from-amber-50 to-orange-50 dark:from-amber-900/20 dark:to-orange-900/20 rounded-lg shadow-sm border border-amber-200 dark:border-amber-700 p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-amber-700 dark:text-amber-400 text-sm font-medium">Total Signups</p>
                  <p className="text-2xl font-bold text-amber-900 dark:text-amber-100">
                    {founderModeResponse.data.total_signups.toLocaleString()}
                  </p>
                </div>
                <Users className="w-8 h-8 text-amber-500" />
              </div>
            </div>

            <div className="bg-gradient-to-br from-green-50 to-emerald-50 dark:from-green-900/20 dark:to-emerald-900/20 rounded-lg shadow-sm border border-green-200 dark:border-green-700 p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-green-700 dark:text-green-400 text-sm font-medium">Active Founders</p>
                  <p className="text-2xl font-bold text-green-900 dark:text-green-100">
                    {founderModeResponse.data.active_founders.toLocaleString()}
                  </p>
                </div>
                <Sparkles className="w-8 h-8 text-green-500" />
              </div>
            </div>

            <div className="bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 rounded-lg shadow-sm border border-blue-200 dark:border-blue-700 p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-blue-700 dark:text-blue-400 text-sm font-medium">Converted</p>
                  <p className="text-2xl font-bold text-blue-900 dark:text-blue-100">
                    {founderModeResponse.data.converted_to_paid.toLocaleString()}
                  </p>
                </div>
                <TrendingUp className="w-8 h-8 text-blue-500" />
              </div>
            </div>

            <div className="bg-gradient-to-br from-purple-50 to-pink-50 dark:from-purple-900/20 dark:to-pink-900/20 rounded-lg shadow-sm border border-purple-200 dark:border-purple-700 p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-purple-700 dark:text-purple-400 text-sm font-medium">Conversion Rate</p>
                  <p className="text-2xl font-bold text-purple-900 dark:text-purple-100">
                    {founderModeResponse.data.conversion_rate.toFixed(1)}%
                  </p>
                </div>
                <BarChart3 className="w-8 h-8 text-purple-500" />
              </div>
            </div>

            <div className="bg-gradient-to-br from-gray-50 to-slate-50 dark:from-gray-900/20 dark:to-slate-900/20 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-gray-700 dark:text-gray-400 text-sm font-medium">Revenue from Founders</p>
                  <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                    ${(founderModeResponse.data.total_revenue_cents / 100).toLocaleString()}
                  </p>
                </div>
                <DollarSign className="w-8 h-8 text-gray-500" />
              </div>
            </div>
          </div>

          {/* By Mode Type Breakdown */}
          <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700">
            <div className="p-4 border-b border-gray-200 dark:border-gray-700">
              <h3 className="font-medium text-gray-900 dark:text-gray-100">Breakdown by Free Tier Type</h3>
            </div>
            <div className="divide-y divide-gray-200 dark:divide-gray-700">
              {founderModeResponse.data.by_mode_type.map((modeType) => {
                const labelInfo = MODE_TYPE_LABELS[modeType.mode_type] || { label: modeType.mode_type, description: '', icon: Clock };
                const IconComponent = labelInfo.icon;
                return (
                  <div key={modeType.mode_type} className="p-4 flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="p-2 bg-amber-100 dark:bg-amber-900/30 rounded-lg">
                        <IconComponent className="w-5 h-5 text-amber-600 dark:text-amber-400" />
                      </div>
                      <div>
                        <div className="font-medium text-gray-900 dark:text-gray-100">{labelInfo.label}</div>
                        <div className="text-sm text-gray-500 dark:text-gray-400">{labelInfo.description}</div>
                      </div>
                    </div>
                    <div className="flex items-center gap-8">
                      <div className="text-center">
                        <div className="text-lg font-semibold text-gray-900 dark:text-gray-100">{modeType.count}</div>
                        <div className="text-xs text-gray-500 dark:text-gray-400">enrolled</div>
                      </div>
                      <div className="text-center">
                        <div className="text-lg font-semibold text-green-600 dark:text-green-400">{modeType.converted}</div>
                        <div className="text-xs text-gray-500 dark:text-gray-400">converted</div>
                      </div>
                      <div className="text-center">
                        <div className="text-lg font-semibold text-purple-600 dark:text-purple-400">{modeType.conversion_rate.toFixed(1)}%</div>
                        <div className="text-xs text-gray-500 dark:text-gray-400">rate</div>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* By Bundle Breakdown */}
          {founderModeResponse.data.by_bundle.length > 0 && (
            <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700">
              <div className="p-4 border-b border-gray-200 dark:border-gray-700">
                <h3 className="font-medium text-gray-900 dark:text-gray-100">Breakdown by Bundle</h3>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="bg-gray-50 dark:bg-gray-800">
                    <tr>
                      <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Bundle</th>
                      <th className="px-4 py-2 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Total</th>
                      <th className="px-4 py-2 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Active</th>
                      <th className="px-4 py-2 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Converted</th>
                      <th className="px-4 py-2 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Revenue</th>
                      <th className="px-4 py-2 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Rate</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                    {founderModeResponse.data.by_bundle.map((bundle) => (
                      <tr key={bundle.bundle_slug} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                        <td className="px-4 py-2 font-medium text-gray-900 dark:text-gray-100">{bundle.bundle_slug}</td>
                        <td className="px-4 py-2 text-right text-gray-900 dark:text-gray-100">{bundle.total_signups}</td>
                        <td className="px-4 py-2 text-right text-green-600 dark:text-green-400">{bundle.active}</td>
                        <td className="px-4 py-2 text-right text-blue-600 dark:text-blue-400">{bundle.converted}</td>
                        <td className="px-4 py-2 text-right text-gray-900 dark:text-gray-100">${(bundle.revenue_cents / 100).toLocaleString()}</td>
                        <td className="px-4 py-2 text-right">
                          <span className={`px-2 py-0.5 text-xs rounded ${
                            bundle.conversion_rate >= 20 ? 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300' :
                            bundle.conversion_rate >= 10 ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300' :
                            'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300'
                          }`}>
                            {bundle.conversion_rate.toFixed(1)}%
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      ) : null}

      {/* Filters */}
      <div className="flex gap-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-3 w-5 h-5 text-gray-400" />
          <input
            type="text"
            placeholder="Search invoices or customers..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          />
        </div>

        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
        >
          <option value="all">All Status</option>
          <option value="paid">Paid</option>
          <option value="pending">Pending</option>
          <option value="overdue">Overdue</option>
          <option value="failed">Failed</option>
        </select>
      </div>

      {/* Invoices List */}
      <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700">
        <div className="p-6 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Recent Invoices</h2>
        </div>
        <div className="divide-y divide-gray-200 dark:divide-gray-700">
          <div className="p-6 text-center text-gray-500 dark:text-gray-400">
            Loading invoice data...
          </div>
        </div>
      </div>

      {/* Pricing Tiers Management */}
      <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700">
        <div className="p-6 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Pricing Tiers</h2>
          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
            Manage monthly and annual pricing for subscription plans.
          </p>
        </div>
        {tiersLoading ? (
          <div className="p-6 text-center text-gray-500 dark:text-gray-400">Loading tiers...</div>
        ) : (
          <div className="divide-y divide-gray-200 dark:divide-gray-700">
            {(tiersResponse?.data?.tiers ?? []).map((tier) => (
              <div key={tier.id} className="p-4 flex items-center justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-3">
                    <h3 className="font-medium text-gray-900 dark:text-gray-100">{tier.name}</h3>
                    <span className={`px-2 py-0.5 text-xs rounded ${
                      tier.billing_cycle === 'annual' ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900 dark:text-emerald-300' : 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300'
                    }`}>
                      {tier.billing_cycle}
                    </span>
                    {!tier.is_active && (
                      <span className="px-2 py-0.5 text-xs rounded bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300">
                        inactive
                      </span>
                    )}
                  </div>
                  <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">{tier.description}</p>
                </div>
                <div className="flex items-center gap-6">
                  {editingTierId === tier.id ? (
                    <div className="flex items-center gap-2">
                      <div className="flex flex-col gap-1">
                        <label className="text-xs text-gray-500">Monthly ($)</label>
                        <input
                          type="number"
                          value={editForm.price_cents}
                          onChange={(e) => setEditForm({ ...editForm, price_cents: e.target.value })}
                          className="w-24 px-2 py-1 border border-gray-300 dark:border-gray-600 rounded text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
                          placeholder="cents"
                        />
                      </div>
                      <div className="flex flex-col gap-1">
                        <label className="text-xs text-gray-500">Annual ($)</label>
                        <input
                          type="number"
                          value={editForm.annual_price_cents}
                          onChange={(e) => setEditForm({ ...editForm, annual_price_cents: e.target.value })}
                          className="w-24 px-2 py-1 border border-gray-300 dark:border-gray-600 rounded text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
                          placeholder="cents"
                        />
                      </div>
                      <button
                        onClick={() => {
                          updateTierMutation.mutate({
                            tierId: tier.id,
                            updates: {
                              price_cents: parseInt(editForm.price_cents) || tier.price_cents,
                              annual_price_cents: parseInt(editForm.annual_price_cents) || tier.annual_price_cents,
                            },
                          });
                        }}
                        disabled={updateTierMutation.isPending}
                        className="p-1.5 bg-emerald-600 text-white rounded hover:bg-emerald-700"
                      >
                        <Check className="w-4 h-4" />
                      </button>
                      <button
                        onClick={() => setEditingTierId(null)}
                        className="p-1.5 bg-gray-400 text-white rounded hover:bg-gray-500"
                      >
                        <X className="w-4 h-4" />
                      </button>
                    </div>
                  ) : (
                    <>
                      <div className="text-right">
                        <div className="text-sm font-medium text-gray-900 dark:text-gray-100">
                          ${(tier.price_cents / 100).toFixed(2)}/mo
                        </div>
                        {tier.annual_price_cents !== null && (
                          <div className="text-xs text-gray-500 dark:text-gray-400">
                            ${(tier.annual_price_cents / 100).toFixed(2)}/yr
                          </div>
                        )}
                      </div>
                      <button
                        onClick={() => {
                          setEditingTierId(tier.id);
                          setEditForm({
                            price_cents: tier.price_cents.toString(),
                            annual_price_cents: tier.annual_price_cents?.toString() ?? '',
                          });
                        }}
                        className="p-2 text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-900/30 rounded"
                      >
                        <Pencil className="w-4 h-4" />
                      </button>
                    </>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* State Fabric Add-on Entitlements (Support Override) */}
      <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700">
        <div className="p-6 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">State Fabric Add-on Entitlements</h2>
          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
            Support override tool for tenant add-ons.
          </p>
        </div>
        <div className="p-6 space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
            <input
              type="text"
              placeholder="Tenant UUID"
              value={tenantIdInput}
              onChange={(e) => setTenantIdInput(e.target.value.trim())}
              className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
            />
            <select
              value={selectedAddon}
              onChange={(e) => setSelectedAddon(e.target.value)}
              className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
            >
              {(addonCatalog?.data ?? []).map((a) => (
                <option key={a.id} value={a.id}>
                  {a.name}
                </option>
              ))}
            </select>
            <select
              value={selectedStatus}
              onChange={(e) => setSelectedStatus(e.target.value)}
              className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
            >
              <option value="active">active</option>
              <option value="inactive">inactive</option>
              <option value="suspended">suspended</option>
            </select>
          </div>
          <button
            onClick={handleOverrideEntitlement}
            disabled={overrideBusy}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-60"
          >
            {overrideBusy ? 'Updating...' : 'Apply override'}
          </button>
          <div className="text-sm text-gray-600 dark:text-gray-400">
            Current entitlements ({(tenantEntitlements?.data ?? []).length})
          </div>
          <div className="space-y-2">
            {(tenantEntitlements?.data ?? []).map((e) => (
              <div
                key={`${e.addon_id}-${e.stripe_subscription_id ?? 'manual'}`}
                className="flex items-center justify-between border border-gray-200 dark:border-gray-700 rounded px-3 py-2"
              >
                <span className="font-mono text-xs text-gray-900 dark:text-gray-100">{e.addon_id}</span>
                <span className="text-xs text-gray-600 dark:text-gray-400">{e.status}</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Revenue Recognition Section */}
      <RevenueRecognitionSection />

      {/* Affiliate / Affiliate Codes Section */}
      <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700">
        <div className="p-6 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Gift className="w-6 h-6 text-purple-600" />
              <div>
                <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Affiliate & Referral Codes</h2>
                <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                  Manage promo codes, track referrals, and handle commission payouts.
                </p>
              </div>
            </div>
            <button
              onClick={() => setShowCreateCodeModal(true)}
              className="flex items-center gap-2 px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-colors"
            >
              <Plus className="w-5 h-5" />
              Create Code
            </button>
          </div>
        </div>

        {/* Tabs */}
        <div className="flex border-b border-gray-200 dark:border-gray-700">
          <button
            onClick={() => { setAffiliateTab('codes'); setSelectedCodeId(null); }}
            className={`px-6 py-3 text-sm font-medium transition-colors ${
              affiliateTab === 'codes'
                ? 'text-purple-600 border-b-2 border-purple-600'
                : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100'
            }`}
          >
            <Gift className="w-4 h-4 inline mr-2" />
            Codes ({affiliateCodesResponse?.data?.affiliate_codes?.length ?? 0})
          </button>
          <button
            onClick={() => setAffiliateTab('referrals')}
            disabled={!selectedCodeId}
            className={`px-6 py-3 text-sm font-medium transition-colors ${
              affiliateTab === 'referrals'
                ? 'text-purple-600 border-b-2 border-purple-600'
                : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100'
            } ${!selectedCodeId ? 'opacity-50 cursor-not-allowed' : ''}`}
          >
            <Users className="w-4 h-4 inline mr-2" />
            Referrals
          </button>
          <button
            onClick={() => setAffiliateTab('commissions')}
            disabled={!selectedCodeId}
            className={`px-6 py-3 text-sm font-medium transition-colors ${
              affiliateTab === 'commissions'
                ? 'text-purple-600 border-b-2 border-purple-600'
                : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100'
            } ${!selectedCodeId ? 'opacity-50 cursor-not-allowed' : ''}`}
          >
            <DollarSign className="w-4 h-4 inline mr-2" />
            Commissions
          </button>
        </div>

        <div className="p-6">
          {affiliateTab === 'codes' && (
            affiliateCodesLoading ? (
              <div className="text-center py-8 text-gray-500">Loading...</div>
            ) : (
              <div className="space-y-3">
                {(affiliateCodesResponse?.data?.affiliate_codes ?? []).map((code) => (
                  <div
                    key={code.id}
                    className={`p-4 border rounded-lg transition-colors ${
                      selectedCodeId === code.id
                        ? 'border-purple-500 bg-purple-50 dark:bg-purple-900/20'
                        : 'border-gray-200 dark:border-gray-700 hover:border-purple-300 dark:hover:border-purple-700'
                    }`}
                    onClick={() => setSelectedCodeId(code.id)}
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-3">
                          <span className="font-mono text-lg font-semibold text-purple-700 dark:text-purple-300">
                            {code.code}
                          </span>
                          <span className={`px-2 py-0.5 text-xs rounded ${
                            code.is_active
                              ? 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'
                              : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'
                          }`}>
                            {code.is_active ? 'active' : 'inactive'}
                          </span>
                          <span className={`px-2 py-0.5 text-xs rounded ${
                            code.commission_type === 'percent'
                              ? 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300'
                              : 'bg-orange-100 text-orange-700 dark:bg-orange-900 dark:text-orange-300'
                          }`}>
                            {code.commission_type === 'percent' ? `${code.commission_value}%` : `$${code.commission_value} fixed`}
                          </span>
                        </div>
                        <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">{code.name}</p>
                        {code.description && (
                          <p className="text-xs text-gray-500 dark:text-gray-500 mt-1">{code.description}</p>
                        )}
                      </div>
                      <div className="text-right space-y-1">
                        <div className="flex items-center gap-4 text-sm">
                          <div>
                            <span className="text-gray-500 dark:text-gray-400">Referrals: </span>
                            <span className="font-medium text-gray-900 dark:text-gray-100">{code.total_referrals}</span>
                          </div>
                          <div>
                            <span className="text-gray-500 dark:text-gray-400">Pending: </span>
                            <span className="font-medium text-yellow-600 dark:text-yellow-400">{code.pending_commissions}</span>
                          </div>
                        </div>
                        <div className="flex items-center gap-4 text-sm">
                          <div>
                            <span className="text-gray-500 dark:text-gray-400">Earnings: </span>
                            <span className="font-medium text-green-600 dark:text-green-400">
                              ${(code.total_earnings_cents / 100).toFixed(2)}
                            </span>
                          </div>
                          <div>
                            <span className="text-gray-500 dark:text-gray-400">Paid: </span>
                            <span className="font-medium text-gray-900 dark:text-gray-100">
                              ${(code.paid_out_earnings_cents / 100).toFixed(2)}
                            </span>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
                {(affiliateCodesResponse?.data?.affiliate_codes?.length ?? 0) === 0 && (
                  <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                    No affiliate codes yet. Create one to get started.
                  </div>
                )}
              </div>
            )
          )}

          {affiliateTab === 'referrals' && (
            affiliateReferralsLoading ? (
              <div className="text-center py-8 text-gray-500">Loading...</div>
            ) : (
              <div className="space-y-2">
                {(affiliateReferralsResponse?.data?.referrals ?? []).map((referral) => (
                  <div
                    key={referral.id}
                    className="flex items-center justify-between p-3 border border-gray-200 dark:border-gray-700 rounded-lg"
                  >
                    <div className="flex items-center gap-3">
                      <span className={`w-2 h-2 rounded-full ${
                        referral.status === 'qualified' ? 'bg-green-500' :
                        referral.status === 'converted' ? 'bg-blue-500' :
                        referral.status === 'canceled' ? 'bg-red-500' : 'bg-yellow-500'
                      }`} />
                      <div>
                        <span className="font-mono text-sm text-gray-900 dark:text-gray-100">
                          {referral.referred_tenant_id.slice(0, 8)}...
                        </span>
                        {referral.utm_source && (
                          <span className="ml-2 text-xs text-gray-500">via {referral.utm_source}</span>
                        )}
                      </div>
                    </div>
                    <div className="flex items-center gap-4 text-sm">
                      <span className={`px-2 py-0.5 text-xs rounded ${
                        referral.status === 'qualified' ? 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300' :
                        referral.status === 'converted' ? 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300' :
                        referral.status === 'canceled' ? 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300' :
                        'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300'
                      }`}>
                        {referral.status}
                      </span>
                      <span className="text-gray-500 dark:text-gray-400">
                        {format(new Date(referral.referred_at), 'MMM d, yyyy')}
                      </span>
                    </div>
                  </div>
                ))}
                {(affiliateReferralsResponse?.data?.referrals?.length ?? 0) === 0 && (
                  <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                    Select a code above to view its referrals.
                  </div>
                )}
              </div>
            )
          )}

          {affiliateTab === 'commissions' && (
            affiliateCommissionsLoading ? (
              <div className="text-center py-8 text-gray-500">Loading...</div>
            ) : (
              <div className="space-y-2">
                {(affiliateCommissionsResponse?.data?.commissions ?? []).map((commission) => (
                  <div
                    key={commission.id}
                    className="flex items-center justify-between p-3 border border-gray-200 dark:border-gray-700 rounded-lg"
                  >
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-medium text-gray-900 dark:text-gray-100">
                          ${commission.commission_usd.toFixed(2)}
                        </span>
                        <span className="text-xs text-gray-500">
                          ({commission.commission_type === 'percent' ? `${commission.commission_value}%` : `$${commission.commission_value} fixed`} of ${commission.base_amount_usd.toFixed(2)})
                        </span>
                      </div>
                      <span className="text-xs text-gray-500 dark:text-gray-400">
                        {format(new Date(commission.created_at), 'MMM d, yyyy')}
                      </span>
                    </div>
                    <div className="flex items-center gap-3">
                      <span className={`px-2 py-0.5 text-xs rounded ${
                        commission.status === 'paid' ? 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300' :
                        commission.status === 'approved' ? 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300' :
                        commission.status === 'canceled' ? 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300' :
                        'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300'
                      }`}>
                        {commission.status}
                      </span>
                      {commission.status === 'pending' && (
                        <div className="flex gap-1">
                          <button
                            onClick={() => approveCommissionMutation.mutate(commission.id)}
                            disabled={approveCommissionMutation.isPending}
                            className="px-2 py-1 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
                            title="Approve"
                          >
                            <Check className="w-3 h-3" />
                          </button>
                        </div>
                      )}
                      {commission.status === 'approved' && (
                        <button
                          onClick={() => markPaidMutation.mutate(commission.id)}
                          disabled={markPaidMutation.isPending}
                          className="px-2 py-1 text-xs bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50"
                        >
                          Mark Paid
                        </button>
                      )}
                    </div>
                  </div>
                ))}
                {(affiliateCommissionsResponse?.data?.commissions?.length ?? 0) === 0 && (
                  <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                    Select a code above to view its commissions.
                  </div>
                )}
              </div>
            )
          )}
        </div>
      </div>

      {/* Create Affiliate Code Modal */}
      {showCreateCodeModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-gray-900 rounded-lg shadow-xl p-6 w-full max-w-md">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Create Affiliate Code</h3>
              <button onClick={() => setShowCreateCodeModal(false)} className="text-gray-400 hover:text-gray-600">
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Code *</label>
                <input
                  type="text"
                  value={createCodeForm.code}
                  onChange={(e) => setCreateCodeForm({ ...createCodeForm, code: e.target.value.toUpperCase() })}
                  placeholder="SAVE20"
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 uppercase"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Publisher ID *</label>
                <input
                  type="text"
                  value={createCodeForm.publisher_id}
                  onChange={(e) => setCreateCodeForm({ ...createCodeForm, publisher_id: e.target.value })}
                  placeholder="UUID of the publisher"
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Name *</label>
                <input
                  type="text"
                  value={createCodeForm.name}
                  onChange={(e) => setCreateCodeForm({ ...createCodeForm, name: e.target.value })}
                  placeholder="Summer Sale 2026"
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
                <input
                  type="text"
                  value={createCodeForm.description}
                  onChange={(e) => setCreateCodeForm({ ...createCodeForm, description: e.target.value })}
                  placeholder="Optional description"
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Commission Type</label>
                  <select
                    value={createCodeForm.commission_type}
                    onChange={(e) => setCreateCodeForm({ ...createCodeForm, commission_type: e.target.value as 'percent' | 'fixed' })}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
                  >
                    <option value="percent">Percent (%)</option>
                    <option value="fixed">Fixed ($)</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Value *</label>
                  <input
                    type="number"
                    value={createCodeForm.commission_value}
                    onChange={(e) => setCreateCodeForm({ ...createCodeForm, commission_value: e.target.value })}
                    placeholder={createCodeForm.commission_type === 'percent' ? '20' : '10'}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Max Commissions</label>
                  <input
                    type="number"
                    value={createCodeForm.max_commissions}
                    onChange={(e) => setCreateCodeForm({ ...createCodeForm, max_commissions: e.target.value })}
                    placeholder="Unlimited"
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Max Referrals</label>
                  <input
                    type="number"
                    value={createCodeForm.max_referrals}
                    onChange={(e) => setCreateCodeForm({ ...createCodeForm, max_referrals: e.target.value })}
                    placeholder="Unlimited"
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
                  />
                </div>
              </div>
            </div>
            <div className="flex gap-3 mt-6">
              <button
                onClick={() => setShowCreateCodeModal(false)}
                className="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 text-gray-700 dark:text-gray-300"
              >
                Cancel
              </button>
              <button
                onClick={() => createAffiliateCodeMutation.mutate(createCodeForm)}
                disabled={createAffiliateCodeMutation.isPending || !createCodeForm.code || !createCodeForm.publisher_id || !createCodeForm.name}
                className="flex-1 px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 disabled:opacity-50"
              >
                {createAffiliateCodeMutation.isPending ? 'Creating...' : 'Create Code'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
