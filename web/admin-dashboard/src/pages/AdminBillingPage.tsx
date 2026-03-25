/**
 * Admin Billing Page
 * Manage billing, subscriptions, and invoicing
 */

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { Download, Search, DollarSign, TrendingUp, AlertCircle } from 'lucide-react';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { toast } from 'sonner';

interface BillingData {
  totalRevenue: number;
  activeSubscriptions: number;
  pendingInvoices: number;
  overdue: number;
}

export function AdminBillingPage() {
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [tenantIdInput, setTenantIdInput] = useState('');
  const [selectedAddon, setSelectedAddon] = useState('hot_cache_booster');
  const [selectedStatus, setSelectedStatus] = useState('active');
  const [overrideBusy, setOverrideBusy] = useState(false);

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
          <h1 className="text-3xl font-bold text-gray-900">Billing & Invoicing</h1>
          <p className="mt-2 text-gray-600">Manage subscriptions, invoices, and revenue</p>
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
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-600 text-sm">Total Revenue</p>
              <p className="text-2xl font-bold text-gray-900">
                ${billingData?.totalRevenue.toLocaleString() || '0'}
              </p>
            </div>
            <DollarSign className="w-8 h-8 text-green-600" />
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-600 text-sm">Active Subscriptions</p>
              <p className="text-2xl font-bold text-gray-900">
                {billingData?.activeSubscriptions || '0'}
              </p>
            </div>
            <TrendingUp className="w-8 h-8 text-blue-600" />
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-600 text-sm">Pending Invoices</p>
              <p className="text-2xl font-bold text-gray-900">
                {billingData?.pendingInvoices || '0'}
              </p>
            </div>
            <AlertCircle className="w-8 h-8 text-yellow-600" />
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-600 text-sm">Overdue</p>
              <p className="text-2xl font-bold text-red-600">
                ${billingData?.overdue || '0'}
              </p>
            </div>
            <AlertCircle className="w-8 h-8 text-red-600" />
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="flex gap-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-3 w-5 h-5 text-gray-400" />
          <input
            type="text"
            placeholder="Search invoices or customers..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>

        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        >
          <option value="all">All Status</option>
          <option value="paid">Paid</option>
          <option value="pending">Pending</option>
          <option value="overdue">Overdue</option>
          <option value="failed">Failed</option>
        </select>
      </div>

      {/* Invoices List */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200">
        <div className="p-6 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">Recent Invoices</h2>
        </div>
        <div className="divide-y divide-gray-200">
          <div className="p-6 text-center text-gray-500">
            Loading invoice data...
          </div>
        </div>
      </div>

      {/* State Fabric Add-on Entitlements (Support Override) */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200">
        <div className="p-6 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">State Fabric Add-on Entitlements</h2>
          <p className="text-sm text-gray-600 mt-1">
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
              className="px-3 py-2 border border-gray-300 rounded-lg"
            />
            <select
              value={selectedAddon}
              onChange={(e) => setSelectedAddon(e.target.value)}
              className="px-3 py-2 border border-gray-300 rounded-lg"
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
              className="px-3 py-2 border border-gray-300 rounded-lg"
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
          <div className="text-sm text-gray-600">
            Current entitlements ({(tenantEntitlements?.data ?? []).length})
          </div>
          <div className="space-y-2">
            {(tenantEntitlements?.data ?? []).map((e) => (
              <div
                key={`${e.addon_id}-${e.stripe_subscription_id ?? 'manual'}`}
                className="flex items-center justify-between border border-gray-200 rounded px-3 py-2"
              >
                <span className="font-mono text-xs">{e.addon_id}</span>
                <span className="text-xs">{e.status}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
