/**
 * Admin IP Allowlist Page
 * Manage IP allowlists for tenant security
 */

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { 
  Plus, 
  Search, 
  Trash2, 
  Edit2, 
  Shield, 
  Globe,
  AlertTriangle,
  CheckCircle,
  XCircle,
  ChevronDown,
  ChevronUp,
  Copy
} from 'lucide-react';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import type { 
  IPAllowlist, 
  IPAllowlistEntry, 
  IPAllowlistCreateInput,
  IPEntryCreateInput,
  Tenant 
} from '@/types';

export function AdminIPAllowlistPage() {
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedTenant, setSelectedTenant] = useState<string>('all');
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  const [isAddEntryDialogOpen, setIsAddEntryDialogOpen] = useState(false);
  const [selectedAllowlist, setSelectedAllowlist] = useState<IPAllowlist | null>(null);
  const [expandedAllowlists, setExpandedAllowlists] = useState<Set<string>>(new Set());
  
  const queryClient = useQueryClient();

  // Form state for creating/editing allowlist
  const [formData, setFormData] = useState<IPAllowlistCreateInput>({
    tenant_id: '',
    name: '',
    description: '',
    default_policy: 'deny',
    mfa_bypass: false,
  });

  // Form state for adding IP entry
  const [entryData, setEntryData] = useState<IPEntryCreateInput>({
    ip_address: '',
    cidr: undefined,
    description: '',
  });

  // Fetch tenants for dropdown
  const { data: tenantsResponse } = useQuery({
    queryKey: ['admin-tenants'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<Tenant[]>('/tenants');
      } catch {
        return { data: [] };
      }
    },
    staleTime: 1000 * 60 * 5,
  });

  const tenants = tenantsResponse?.data || [];

  // Fetch IP allowlists
  const { data: allowlistsResponse, isLoading, isError, refetch } = useQuery({
    queryKey: ['admin-ip-allowlists', selectedTenant],
    queryFn: async () => {
      try {
        const tenantId = selectedTenant !== 'all' ? selectedTenant : '';
        const endpoint = tenantId 
          ? `/tenants/${tenantId}/ip-allowlists`
          : '/admin/ip-allowlists';
        return await adminApiClient.get<IPAllowlist[]>(endpoint);
      } catch {
        return { data: [] };
      }
    },
    staleTime: 1000 * 60,
  });

  const allowlists = allowlistsResponse?.data || [];

  // Create allowlist mutation
  const createMutation = useMutation({
    mutationFn: (data: IPAllowlistCreateInput) =>
      adminApiClient.post(`/tenants/${data.tenant_id}/ip-allowlists`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-ip-allowlists'] });
      setIsCreateDialogOpen(false);
      resetFormData();
    },
  });

  // Update allowlist mutation
  const updateMutation = useMutation({
    mutationFn: ({ id, updates }: { id: string; updates: Partial<IPAllowlist> }) =>
      adminApiClient.put(`/ip-allowlists/${id}`, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-ip-allowlists'] });
      setIsEditDialogOpen(false);
      setSelectedAllowlist(null);
    },
  });

  // Delete allowlist mutation
  const deleteMutation = useMutation({
    mutationFn: (id: string) =>
      adminApiClient.delete(`/ip-allowlists/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-ip-allowlists'] });
    },
  });

  // Add IP entry mutation
  const addEntryMutation = useMutation({
    mutationFn: ({ allowlistId, entry }: { allowlistId: string; entry: IPEntryCreateInput }) =>
      adminApiClient.post(`/ip-allowlists/${allowlistId}/entries`, entry),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-ip-allowlists'] });
      setIsAddEntryDialogOpen(false);
      resetEntryData();
    },
  });

  // Delete IP entry mutation
  const deleteEntryMutation = useMutation({
    mutationFn: ({ allowlistId, entryId }: { allowlistId: string; entryId: string }) =>
      adminApiClient.delete(`/ip-allowlists/${allowlistId}/entries/${entryId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-ip-allowlists'] });
    },
  });

  const resetFormData = () => {
    setFormData({
      tenant_id: tenants[0]?.id || '',
      name: '',
      description: '',
      default_policy: 'deny',
      mfa_bypass: false,
    });
  };

  const resetEntryData = () => {
    setEntryData({
      ip_address: '',
      cidr: undefined,
      description: '',
    });
  };

  const filteredAllowlists = allowlists.filter((allowlist) => {
    const matchesSearch = searchTerm
      ? allowlist.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        allowlist.description?.toLowerCase().includes(searchTerm.toLowerCase())
      : true;
    return matchesSearch;
  });

  const toggleExpand = (id: string) => {
    const newExpanded = new Set(expandedAllowlists);
    if (newExpanded.has(id)) {
      newExpanded.delete(id);
    } else {
      newExpanded.add(id);
    }
    setExpandedAllowlists(newExpanded);
  };

  const handleCreateAllowlist = () => {
    if (formData.tenant_id && formData.name.trim()) {
      createMutation.mutate(formData);
    }
  };

  const handleUpdateAllowlist = () => {
    if (selectedAllowlist) {
      updateMutation.mutate({
        id: selectedAllowlist.id,
        updates: {
          name: formData.name,
          description: formData.description,
          default_policy: formData.default_policy,
          mfa_bypass: formData.mfa_bypass,
        },
      });
    }
  };

  const handleDeleteAllowlist = (allowlist: IPAllowlist) => {
    if (confirm(`Are you sure you want to delete the allowlist "${allowlist.name}"?`)) {
      deleteMutation.mutate(allowlist.id);
    }
  };

  const handleAddEntry = () => {
    if (selectedAllowlist && entryData.ip_address.trim()) {
      addEntryMutation.mutate({
        allowlistId: selectedAllowlist.id,
        entry: entryData,
      });
    }
  };

  const handleDeleteEntry = (allowlistId: string, entryId: string) => {
    if (confirm('Are you sure you want to delete this IP entry?')) {
      deleteEntryMutation.mutate({ allowlistId, entryId });
    }
  };

  const openEditDialog = (allowlist: IPAllowlist) => {
    setSelectedAllowlist(allowlist);
    setFormData({
      tenant_id: allowlist.tenant_id,
      name: allowlist.name,
      description: allowlist.description || '',
      default_policy: allowlist.default_policy,
      mfa_bypass: allowlist.mfa_bypass,
    });
    setIsEditDialogOpen(true);
  };

  const openAddEntryDialog = (allowlist: IPAllowlist) => {
    setSelectedAllowlist(allowlist);
    resetEntryData();
    setIsAddEntryDialogOpen(true);
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  if (isLoading) {
    return <LoadingScreen />;
  }

  if (isError) {
    return (
      <div className="p-8 bg-red-50 border border-red-200 rounded-lg dark:bg-red-900/20 dark:border-red-800">
        <h3 className="font-semibold text-red-900 dark:text-red-200">Error loading IP allowlists</h3>
        <p className="text-red-700 dark:text-red-300 mt-2">Failed to fetch allowlist data.</p>
        <button 
          onClick={() => refetch()}
          className="mt-4 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">IP Allowlists</h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">
            Manage IP allowlists for tenant access control
          </p>
        </div>

        <button
          onClick={() => {
            resetFormData();
            setIsCreateDialogOpen(true);
          }}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          <Plus className="w-5 h-5" />
          Create Allowlist
        </button>
      </div>

      {/* Filters */}
      <div className="flex gap-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-3 w-5 h-5 text-gray-400" />
          <input
            type="text"
            placeholder="Search allowlists..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
          />
        </div>

        <select
          value={selectedTenant}
          onChange={(e) => setSelectedTenant(e.target.value)}
          className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
        >
          <option value="all">All Tenants</option>
          {tenants.map((tenant) => (
            <option key={tenant.id} value={tenant.id}>
              {tenant.name}
            </option>
          ))}
        </select>
      </div>

      {/* Allowlists List */}
      <div className="space-y-4">
        {filteredAllowlists.length === 0 ? (
          <div className="p-8 text-center text-gray-500 dark:text-gray-400 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
            No IP allowlists found
          </div>
        ) : (
          filteredAllowlists.map((allowlist) => {
            const isExpanded = expandedAllowlists.has(allowlist.id);
            const tenant = tenants.find((t) => t.id === allowlist.tenant_id);

            return (
              <div
                key={allowlist.id}
                className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden"
              >
                {/* Allowlist Header */}
                <div
                  className="p-4 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700/50"
                  onClick={() => toggleExpand(allowlist.id)}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-4">
                      <div className="flex items-center gap-2">
                        <Shield className={`w-5 h-5 ${allowlist.enabled ? 'text-green-600' : 'text-gray-400'}`} />
                        <span className="font-medium text-gray-900 dark:text-white">{allowlist.name}</span>
                      </div>
                      
                      <div className="flex items-center gap-2">
                        {allowlist.enabled ? (
                          <span className="inline-flex items-center gap-1 px-2 py-1 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 rounded text-xs">
                            <CheckCircle className="w-3 h-3" />
                            Enabled
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1 px-2 py-1 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 rounded text-xs">
                            <XCircle className="w-3 h-3" />
                            Disabled
                          </span>
                        )}

                        <span className={`inline-flex items-center gap-1 px-2 py-1 rounded text-xs ${
                          allowlist.default_policy === 'allow' 
                            ? 'bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400'
                            : 'bg-orange-100 dark:bg-orange-900/30 text-orange-700 dark:text-orange-400'
                        }`}>
                          {allowlist.default_policy === 'allow' ? 'Default: Allow' : 'Default: Deny'}
                        </span>

                        {allowlist.mfa_bypass && (
                          <span className="inline-flex items-center gap-1 px-2 py-1 bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-400 rounded text-xs">
                            MFA Bypass
                          </span>
                        )}
                      </div>
                    </div>

                    <div className="flex items-center gap-4">
                      <span className="text-sm text-gray-500 dark:text-gray-400">
                        {tenant?.name || 'Unknown Tenant'}
                      </span>
                      <span className="text-sm text-gray-500 dark:text-gray-400">
                        {allowlist.entries?.length || 0} IPs
                      </span>
                      {isExpanded ? (
                        <ChevronUp className="w-5 h-5 text-gray-400" />
                      ) : (
                        <ChevronDown className="w-5 h-5 text-gray-400" />
                      )}
                    </div>
                  </div>

                  {allowlist.description && (
                    <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
                      {allowlist.description}
                    </p>
                  )}
                </div>

                {/* Expanded Content */}
                {isExpanded && (
                  <div className="border-t border-gray-200 dark:border-gray-700 p-4">
                    {/* Actions */}
                    <div className="flex gap-2 mb-4">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          openAddEntryDialog(allowlist);
                        }}
                        className="flex items-center gap-1 px-3 py-1.5 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 rounded text-sm hover:bg-blue-200 dark:hover:bg-blue-900/50"
                      >
                        <Plus className="w-4 h-4" />
                        Add IP
                      </button>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          openEditDialog(allowlist);
                        }}
                        className="flex items-center gap-1 px-3 py-1.5 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded text-sm hover:bg-gray-200 dark:hover:bg-gray-600"
                      >
                        <Edit2 className="w-4 h-4" />
                        Edit
                      </button>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDeleteAllowlist(allowlist);
                        }}
                        disabled={deleteMutation.isPending}
                        className="flex items-center gap-1 px-3 py-1.5 bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400 rounded text-sm hover:bg-red-200 dark:hover:bg-red-900/50"
                      >
                        <Trash2 className="w-4 h-4" />
                        Delete
                      </button>
                    </div>

                    {/* IP Entries Table */}
                    {allowlist.entries && allowlist.entries.length > 0 ? (
                      <table className="w-full">
                        <thead>
                          <tr className="border-b border-gray-200 dark:border-gray-700">
                            <th className="px-4 py-2 text-left text-sm font-medium text-gray-700 dark:text-gray-300">IP Address</th>
                            <th className="px-4 py-2 text-left text-sm font-medium text-gray-700 dark:text-gray-300">CIDR</th>
                            <th className="px-4 py-2 text-left text-sm font-medium text-gray-700 dark:text-gray-300">Description</th>
                            <th className="px-4 py-2 text-left text-sm font-medium text-gray-700 dark:text-gray-300">Created</th>
                            <th className="px-4 py-2 text-left text-sm font-medium text-gray-700 dark:text-gray-300">Actions</th>
                          </tr>
                        </thead>
                        <tbody>
                          {allowlist.entries.map((entry) => (
                            <tr key={entry.id} className="border-b border-gray-100 dark:border-gray-700">
                              <td className="px-4 py-2">
                                <div className="flex items-center gap-2">
                                  <Globe className="w-4 h-4 text-gray-400" />
                                  <span className="font-mono text-sm text-gray-900 dark:text-white">
                                    {entry.ip_address}
                                  </span>
                                  <button
                                    onClick={() => copyToClipboard(entry.ip_address)}
                                    className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                                  >
                                    <Copy className="w-3 h-3" />
                                  </button>
                                </div>
                              </td>
                              <td className="px-4 py-2 text-sm text-gray-600 dark:text-gray-400">
                                {entry.cidr ? `/${entry.cidr}` : '-'}
                              </td>
                              <td className="px-4 py-2 text-sm text-gray-600 dark:text-gray-400">
                                {entry.description || '-'}
                              </td>
                              <td className="px-4 py-2 text-sm text-gray-500 dark:text-gray-400">
                                {new Date(entry.created_at).toLocaleDateString()}
                              </td>
                              <td className="px-4 py-2">
                                <button
                                  onClick={() => handleDeleteEntry(allowlist.id, entry.id)}
                                  disabled={deleteEntryMutation.isPending}
                                  className="p-1 text-red-600 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300"
                                >
                                  <Trash2 className="w-4 h-4" />
                                </button>
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    ) : (
                      <p className="text-sm text-gray-500 dark:text-gray-400 text-center py-4">
                        No IP entries yet. Click "Add IP" to add entries.
                      </p>
                    )}
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>

      {/* Create Allowlist Dialog */}
      {isCreateDialogOpen && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 max-w-md w-full">
            <h2 className="text-xl font-bold mb-4 text-gray-900 dark:text-white">Create IP Allowlist</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Tenant
                </label>
                <select
                  value={formData.tenant_id}
                  onChange={(e) => setFormData({ ...formData, tenant_id: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                >
                  <option value="">Select a tenant</option>
                  {tenants.map((tenant) => (
                    <option key={tenant.id} value={tenant.id}>
                      {tenant.name}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Name
                </label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="e.g., Office Network"
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Description
                </label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  placeholder="Optional description"
                  rows={2}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Default Policy
                </label>
                <select
                  value={formData.default_policy}
                  onChange={(e) => setFormData({ ...formData, default_policy: e.target.value as 'allow' | 'deny' })}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                >
                  <option value="deny">Deny (whitelist mode)</option>
                  <option value="allow">Allow (blacklist mode)</option>
                </select>
                <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  {formData.default_policy === 'deny' 
                    ? 'Only IPs in the list will be allowed access'
                    : 'All IPs except those in the list will be allowed access'}
                </p>
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="mfa_bypass"
                  checked={formData.mfa_bypass}
                  onChange={(e) => setFormData({ ...formData, mfa_bypass: e.target.checked })}
                  className="rounded border-gray-300 dark:border-gray-600"
                />
                <label htmlFor="mfa_bypass" className="text-sm text-gray-700 dark:text-gray-300">
                  Allow MFA bypass for IPs in this list
                </label>
              </div>

              <div className="flex gap-3 pt-4">
                <button
                  onClick={() => setIsCreateDialogOpen(false)}
                  className="flex-1 px-4 py-2 bg-gray-200 dark:bg-gray-700 text-gray-800 dark:text-gray-200 rounded-lg hover:bg-gray-300 dark:hover:bg-gray-600"
                >
                  Cancel
                </button>
                <button
                  onClick={handleCreateAllowlist}
                  disabled={createMutation.isPending || !formData.tenant_id || !formData.name.trim()}
                  className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
                >
                  {createMutation.isPending ? 'Creating...' : 'Create'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Edit Allowlist Dialog */}
      {isEditDialogOpen && selectedAllowlist && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 max-w-md w-full">
            <h2 className="text-xl font-bold mb-4 text-gray-900 dark:text-white">Edit IP Allowlist</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Name
                </label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Description
                </label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  rows={2}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Default Policy
                </label>
                <select
                  value={formData.default_policy}
                  onChange={(e) => setFormData({ ...formData, default_policy: e.target.value as 'allow' | 'deny' })}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                >
                  <option value="deny">Deny (whitelist mode)</option>
                  <option value="allow">Allow (blacklist mode)</option>
                </select>
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="edit_mfa_bypass"
                  checked={formData.mfa_bypass}
                  onChange={(e) => setFormData({ ...formData, mfa_bypass: e.target.checked })}
                  className="rounded border-gray-300 dark:border-gray-600"
                />
                <label htmlFor="edit_mfa_bypass" className="text-sm text-gray-700 dark:text-gray-300">
                  Allow MFA bypass for IPs in this list
                </label>
              </div>

              <div className="flex gap-3 pt-4">
                <button
                  onClick={() => {
                    setIsEditDialogOpen(false);
                    setSelectedAllowlist(null);
                  }}
                  className="flex-1 px-4 py-2 bg-gray-200 dark:bg-gray-700 text-gray-800 dark:text-gray-200 rounded-lg hover:bg-gray-300 dark:hover:bg-gray-600"
                >
                  Cancel
                </button>
                <button
                  onClick={handleUpdateAllowlist}
                  disabled={updateMutation.isPending || !formData.name.trim()}
                  className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
                >
                  {updateMutation.isPending ? 'Saving...' : 'Save'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Add IP Entry Dialog */}
      {isAddEntryDialogOpen && selectedAllowlist && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 max-w-md w-full">
            <h2 className="text-xl font-bold mb-4 text-gray-900 dark:text-white">
              Add IP Entry to "{selectedAllowlist.name}"
            </h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  IP Address
                </label>
                <input
                  type="text"
                  value={entryData.ip_address}
                  onChange={(e) => setEntryData({ ...entryData, ip_address: e.target.value })}
                  placeholder="e.g., 192.168.1.1"
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white font-mono"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  CIDR (optional)
                </label>
                <input
                  type="number"
                  value={entryData.cidr || ''}
                  onChange={(e) => setEntryData({ ...entryData, cidr: e.target.value ? parseInt(e.target.value) : undefined })}
                  placeholder="e.g., 24"
                  min={0}
                  max={128}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                />
                <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  Use CIDR notation for IP ranges (e.g., 24 for /24 subnet)
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Description (optional)
                </label>
                <input
                  type="text"
                  value={entryData.description}
                  onChange={(e) => setEntryData({ ...entryData, description: e.target.value })}
                  placeholder="e.g., Office WiFi"
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                />
              </div>

              <div className="flex gap-3 pt-4">
                <button
                  onClick={() => {
                    setIsAddEntryDialogOpen(false);
                    setSelectedAllowlist(null);
                  }}
                  className="flex-1 px-4 py-2 bg-gray-200 dark:bg-gray-700 text-gray-800 dark:text-gray-200 rounded-lg hover:bg-gray-300 dark:hover:bg-gray-600"
                >
                  Cancel
                </button>
                <button
                  onClick={handleAddEntry}
                  disabled={addEntryMutation.isPending || !entryData.ip_address.trim()}
                  className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
                >
                  {addEntryMutation.isPending ? 'Adding...' : 'Add Entry'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Summary Stats */}
      <div className="grid grid-cols-3 gap-4">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
          <p className="text-gray-600 dark:text-gray-400 text-sm">Total Allowlists</p>
          <p className="text-2xl font-bold text-gray-900 dark:text-white">{allowlists.length}</p>
        </div>
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
          <p className="text-gray-600 dark:text-gray-400 text-sm">Enabled</p>
          <p className="text-2xl font-bold text-green-600 dark:text-green-400">
            {allowlists.filter((a) => a.enabled).length}
          </p>
        </div>
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
          <p className="text-gray-600 dark:text-gray-400 text-sm">Total IP Entries</p>
          <p className="text-2xl font-bold text-blue-600 dark:text-blue-400">
            {allowlists.reduce((sum, a) => sum + (a.entries?.length || 0), 0)}
          </p>
        </div>
      </div>
    </div>
  );
}

export default AdminIPAllowlistPage;
