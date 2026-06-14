/**
 * Admin Providers Page
 * Manage external providers and service integrations
 */

import { useState, useRef, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApiClient, getProviderSettings, updateProviderSettings, ProviderSettings } from '@/lib/api/adminClient';
import { Plus, Search, MoreVertical, Trash2, AlertTriangle, Settings, X } from 'lucide-react';
import { LoadingScreen } from '@/components/common/LoadingScreen';

interface Provider {
  id: string;
  user_email: string;
  tenant_name: string;
  provider: string;
  status: string;
  is_shared: boolean;
  created_at: string;
  updated_at: string;
}

type StatusFilter = 'all' | 'active' | 'maintenance' | 'error';

export function AdminProvidersPage() {
  const [searchTerm, setSearchTerm] = useState('');
  const [typeFilter, setTypeFilter] = useState<string>('all');
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');
  const [deleteProviderId, setDeleteProviderId] = useState<string | null>(null);
  const [openMenuId, setOpenMenuId] = useState<string | null>(null);
  const [showSettings, setShowSettings] = useState(false);
  const [editingProvider, setEditingProvider] = useState<string | null>(null);
  const [maintenanceReason, setMaintenanceReason] = useState('');
  const queryClient = useQueryClient();

  const { data: providersResponse, isLoading, isError } = useQuery({
    queryKey: ['admin-providers'],
    queryFn: async () => {
      try {
        const res = await adminApiClient.get<Provider[]>('/providers');
        const items = Array.isArray(res) ? res : [];
        return items;
      } catch {
        return [];
      }
    },
    staleTime: 1000 * 60,
  });

  const { data: settingsResponse } = useQuery({
    queryKey: ['admin-provider-settings'],
    queryFn: getProviderSettings,
    staleTime: 1000 * 60,
    enabled: showSettings,
  });

  const updateSettingsMutation = useMutation({
    mutationFn: async ({ provider, disabled, reason }: { provider: string; disabled: boolean; reason?: string }) => {
      return updateProviderSettings(provider, { disabled, reason });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-provider-settings'] });
      queryClient.invalidateQueries({ queryKey: ['admin-providers'] });
      setEditingProvider(null);
      setMaintenanceReason('');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (providerId: string) => {
      await adminApiClient.delete(`/providers/${providerId}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-providers'] });
      setOpenMenuId(null);
    },
  });

  const providers = providersResponse ?? [];
  const settings = settingsResponse ?? [];

  const settingsMap = new Map<string, ProviderSettings>();
  settings.forEach((s) => settingsMap.set(s.provider, s));

  if (isLoading) {
    return <LoadingScreen />;
  }

  if (isError) {
    return (
      <div className="p-8 bg-red-50 border border-red-200 rounded-lg">
        <h3 className="font-semibold text-red-900">Error loading providers</h3>
      </div>
    );
  }

  const filteredProviders = providers.filter((provider: Provider) => {
    const matchesSearch = (provider.tenant_name || provider.user_email).toLowerCase().includes(searchTerm.toLowerCase());
    const matchesType = typeFilter === 'all' || provider.provider === typeFilter;

    const providerSetting = settingsMap.get(provider.provider);
    const isUnderMaintenance = providerSetting?.disabled ?? false;

    let matchesStatus = true;
    if (statusFilter === 'maintenance') {
      matchesStatus = isUnderMaintenance;
    } else if (statusFilter === 'active') {
      matchesStatus = !isUnderMaintenance && provider.status === 'active';
    } else if (statusFilter === 'error') {
      matchesStatus = provider.status === 'error';
    }

    return matchesSearch && matchesType && matchesStatus;
  });

  const providerTypes = [...new Set(providers.map((p: Provider) => p.provider))];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Service Providers</h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">Manage external integrations and service providers</p>
        </div>
        <div className="flex gap-3">
          <button
            onClick={() => setShowSettings(!showSettings)}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg border ${
              showSettings
                ? 'bg-blue-50 border-blue-300 text-blue-700 dark:bg-blue-900/30 dark:border-blue-700 dark:text-blue-300'
                : 'border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800'
            }`}
          >
            <Settings className="w-5 h-5" />
            Provider Settings
          </button>
          <button className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
            <Plus className="w-5 h-5" />
            Add Provider
          </button>
        </div>
      </div>

      {showSettings && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">Provider Maintenance Mode</h2>
          <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
            When a provider is disabled for maintenance, users will not be able to connect new providers of that type.
            Existing connections remain active.
          </p>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {providerTypes.map((type) => {
              const setting = settingsMap.get(type);
              const isDisabled = setting?.disabled ?? false;
              const isEditing = editingProvider === type;

              return (
                <div
                  key={type}
                  className={`p-4 rounded-lg border ${
                    isDisabled
                      ? 'bg-amber-50 border-amber-200 dark:bg-amber-900/20 dark:border-amber-800'
                      : 'bg-gray-50 border-gray-200 dark:bg-gray-700/50 dark:border-gray-600'
                  }`}
                >
                  <div className="flex items-center justify-between mb-2">
                    <span className="font-medium text-gray-900 dark:text-gray-100">{type}</span>
                    {isDisabled && (
                      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-300">
                        <AlertTriangle className="w-3 h-3" />
                        Maintenance
                      </span>
                    )}
                  </div>

                  {isEditing ? (
                    <div className="space-y-2">
                      <input
                        type="text"
                        placeholder="Maintenance reason..."
                        value={maintenanceReason}
                        onChange={(e) => setMaintenanceReason(e.target.value)}
                        className="w-full px-3 py-1.5 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                      />
                      <div className="flex gap-2">
                        <button
                          onClick={() => updateSettingsMutation.mutate({ provider: type, disabled: true, reason: maintenanceReason })}
                          disabled={updateSettingsMutation.isPending}
                          className="px-3 py-1 text-sm bg-amber-600 text-white rounded hover:bg-amber-700 disabled:opacity-50"
                        >
                          {updateSettingsMutation.isPending ? 'Saving...' : 'Save'}
                        </button>
                        <button
                          onClick={() => { setEditingProvider(null); setMaintenanceReason(''); }}
                          className="px-3 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-100 dark:hover:bg-gray-700"
                        >
                          Cancel
                        </button>
                      </div>
                    </div>
                  ) : (
                    <div className="space-y-2">
                      {setting?.disabled_reason && (
                        <p className="text-sm text-amber-700 dark:text-amber-300">{setting.disabled_reason}</p>
                      )}
                      <div className="flex gap-2">
                        {isDisabled ? (
                          <>
                            <button
                              onClick={() => updateSettingsMutation.mutate({ provider: type, disabled: false })}
                              disabled={updateSettingsMutation.isPending}
                              className="px-3 py-1 text-sm bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50"
                            >
                              Re-enable
                            </button>
                            <button
                              onClick={() => {
                                setEditingProvider(type);
                                setMaintenanceReason(setting?.disabled_reason || '');
                              }}
                              className="px-3 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-100 dark:hover:bg-gray-700"
                            >
                              Edit Reason
                            </button>
                          </>
                        ) : (
                          <button
                            onClick={() => setEditingProvider(type)}
                            className="px-3 py-1 text-sm border border-amber-300 text-amber-700 dark:border-amber-700 dark:text-amber-300 rounded hover:bg-amber-50 dark:hover:bg-amber-900/30"
                          >
                            Disable for Maintenance
                          </button>
                        )}
                      </div>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}

      <div className="flex gap-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-3 w-5 h-5 text-gray-400" />
          <input
            type="text"
            placeholder="Search providers..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          />
        </div>
        <select
          value={typeFilter}
          onChange={(e) => setTypeFilter(e.target.value)}
          className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
        >
          <option value="all">All Types</option>
          {providerTypes.map((type) => (
            <option key={type} value={type}>
              {type}
            </option>
          ))}
        </select>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value as StatusFilter)}
          className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
        >
          <option value="all">All Status</option>
          <option value="active">Active</option>
          <option value="maintenance">Maintenance</option>
          <option value="error">Error</option>
        </select>
      </div>

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 overflow-visible">
        <table className="w-full overflow-visible">
          <thead>
            <tr className="border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Name</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Type</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Status</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Last Sync</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredProviders.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-6 py-8 text-center text-gray-500 dark:text-gray-400">
                  No providers found
                </td>
              </tr>
            ) : (
              filteredProviders.map((provider: Provider) => {
                const providerSetting = settingsMap.get(provider.provider);
                const isUnderMaintenance = providerSetting?.disabled ?? false;

                return (
                  <tr key={provider.id} className="border-b border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800">
                    <td className="px-6 py-4 text-sm font-medium text-gray-900 dark:text-gray-100">{provider.tenant_name || provider.user_email}</td>
                    <td className="px-6 py-4 text-sm">
                      <div className="flex items-center gap-2">
                        <span className="text-gray-600 dark:text-gray-400">{provider.provider}</span>
                        {isUnderMaintenance && (
                          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-300">
                            <AlertTriangle className="w-3 h-3" />
                            Maintenance
                          </span>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4 text-sm">
                      <span className={`inline-block px-2 py-1 rounded text-xs font-medium ${
                        provider.status === 'connected' || provider.status === 'active' ? 'bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-400' :
                        provider.status === 'error' ? 'bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-400' :
                        'bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-400'
                      }`}>
                        {provider.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-600 dark:text-gray-400">
                      {new Date(provider.updated_at).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-4 text-sm">
                      <div className="relative inline-block">
                        <button
                          onClick={() => setOpenMenuId(openMenuId === provider.id ? null : provider.id)}
                          className="p-1 hover:bg-gray-100 dark:hover:bg-gray-700 rounded"
                          data-dropdown
                        >
                          <MoreVertical className="w-4 h-4 text-gray-500 dark:text-gray-400" />
                        </button>
                        {openMenuId === provider.id && (
                          <div className="absolute left-0 mt-1 w-36 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-md shadow-lg z-50" data-dropdown>
                            <button
                              onClick={() => {
                                if (confirm('Are you sure you want to delete this provider?')) {
                                  deleteMutation.mutate(provider.id);
                                }
                              }}
                              className="w-full px-4 py-2 text-left text-sm text-red-600 hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2 rounded-md"
                            >
                              <Trash2 className="w-4 h-4" /> Delete
                            </button>
                          </div>
                        )}
                      </div>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
