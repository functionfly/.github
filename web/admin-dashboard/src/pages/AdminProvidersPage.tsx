/**
 * Admin Providers Page
 * Manage external providers and service integrations
 */

import { useState, useRef, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { Plus, Search, MoreVertical, Trash2 } from 'lucide-react';
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

export function AdminProvidersPage() {
  const [searchTerm, setSearchTerm] = useState('');
  const [typeFilter, setTypeFilter] = useState<string>('all');
  const [deleteProviderId, setDeleteProviderId] = useState<string | null>(null);
  const [openMenuId, setOpenMenuId] = useState<string | null>(null);
  const queryClient = useQueryClient();

  const { data: providersResponse, isLoading, isError } = useQuery({
    queryKey: ['admin-providers'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<Provider[]>('/providers');
      } catch {
        return { providers: [], success: false };
      }
    },
    staleTime: 1000 * 60,
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

  const providers = providersResponse?.providers || [];

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

  const filteredProviders = providers.filter((provider) => {
    const matchesSearch = (provider.tenant_name || provider.user_email).toLowerCase().includes(searchTerm.toLowerCase());
    const matchesType = typeFilter === 'all' || provider.provider === typeFilter;
    return matchesSearch && matchesType;
  });

  const providerTypes = [...new Set(providers.map((p) => p.provider))];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Service Providers</h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">Manage external integrations and service providers</p>
        </div>
        <button className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
          <Plus className="w-5 h-5" />
          Add Provider
        </button>
      </div>

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
              filteredProviders.map((provider) => (
                <tr key={provider.id} className="border-b border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800">
                  <td className="px-6 py-4 text-sm font-medium text-gray-900 dark:text-gray-100">{provider.tenant_name || provider.user_email}</td>
                  <td className="px-6 py-4 text-sm text-gray-600 dark:text-gray-400">{provider.provider}</td>
                  <td className="px-6 py-4 text-sm">
                    <span className={`inline-block px-2 py-1 rounded text-xs font-medium ${
                      provider.status === 'connected' ? 'bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-400' :
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
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
