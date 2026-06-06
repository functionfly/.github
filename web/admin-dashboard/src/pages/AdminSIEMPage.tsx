/**
 * Admin SIEM Configuration Page
 * Manage SIEM export configurations for security event forwarding
 */

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { 
  Plus, 
  Search, 
  Trash2, 
  Edit2, 
  Play,
  CheckCircle,
  XCircle,
  Clock,
  RefreshCw,
  ExternalLink,
  Copy,
  Eye,
  EyeOff
} from 'lucide-react';
import { LoadingScreen } from '@/components/common/LoadingScreen';
interface SIEMDestinationType {
  value: string;
  label: string;
  icon: string;
}

const SIEM_DESTINATION_TYPES: SIEMDestinationType[] = [
  { value: 'webhook', label: 'Generic webhook', icon: '🔗' },
  { value: 'splunk_hec', label: 'Splunk HEC', icon: '🟠' },
  { value: 'datadog_logs', label: 'Datadog Logs', icon: '🐶' },
  { value: 'cloudwatch', label: 'AWS CloudWatch', icon: '☁️' },
  { value: 'azure', label: 'Azure Monitor', icon: '🔷' },
  { value: 'gcp', label: 'GCP Logging', icon: '🌈' },
];

interface SIEMFormatType {
  value: string;
  label: string;
}

const SIEM_FORMATS: SIEMFormatType[] = [
  { value: 'json', label: 'JSON' },
  { value: 'cef', label: 'CEF' },
  { value: 'leef', label: 'LEEF' },
];
import type { 
  SIEMConfig, 
  SIEMConfigCreateInput,
  SIEMDestination,
  SIEMFormat,
  SIEMTestResult,
  Tenant 
} from '@/types';

export function AdminSIEMPage() {
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedTenant, setSelectedTenant] = useState<string>('all');
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  const [isTestDialogOpen, setIsTestDialogOpen] = useState(false);
  const [selectedConfig, setSelectedConfig] = useState<SIEMConfig | null>(null);
  const [testResult, setTestResult] = useState<SIEMTestResult | null>(null);
  const [showSecrets, setShowSecrets] = useState<Record<string, boolean>>({});
  
  const queryClient = useQueryClient();

  // Form state
  const [formData, setFormData] = useState<SIEMConfigCreateInput & { enabled: boolean }>({
    tenant_id: '',
    name: '',
    destination_type: 'webhook',
    format: 'json',
    enabled: true,
    config: {},
  } as any);

  // Fetch tenants
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

  // Fetch SIEM configs
  const { data: configsResponse, isLoading, isError, refetch, isFetching } = useQuery({
    queryKey: ['admin-siem-configs', selectedTenant],
    queryFn: async () => {
      try {
        const tenantId = selectedTenant !== 'all' ? selectedTenant : '';
        const endpoint = tenantId 
          ? `/tenants/${tenantId}/siem-configs`
          : '/admin/siem-configs';
        return await adminApiClient.get<SIEMConfig[]>(endpoint);
      } catch {
        return { data: [] };
      }
    },
    staleTime: 1000 * 60,
  });

  const configs = configsResponse?.data || [];

  // Create mutation
  const createMutation = useMutation({
    mutationFn: (data: SIEMConfigCreateInput) =>
      adminApiClient.post(`/tenants/${data.tenant_id}/siem-configs`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-siem-configs'] });
      setIsCreateDialogOpen(false);
      resetFormData();
    },
  });

  // Update mutation
  const updateMutation = useMutation({
    mutationFn: ({ id, updates }: { id: string; updates: Partial<SIEMConfig> }) =>
      adminApiClient.put(`/siem-configs/${id}`, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-siem-configs'] });
      setIsEditDialogOpen(false);
      setSelectedConfig(null);
    },
  });

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: (id: string) =>
      adminApiClient.delete(`/siem-configs/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-siem-configs'] });
    },
  });

  // Test connection mutation
  const testMutation = useMutation({
    mutationFn: (id: string) =>
      adminApiClient.post<SIEMTestResult>(`/siem-configs/${id}/test`),
    onSuccess: (response) => {
      setTestResult(response.data);
    },
  });

  const resetFormData = () => {
    setFormData({
      tenant_id: tenants[0]?.id || '',
      name: '',
      destination_type: 'webhook',
      format: 'json',
      enabled: true,
      config: {},
    } as any);
  };

  const filteredConfigs = configs.filter((config) => {
    const matchesSearch = searchTerm
      ? config.name.toLowerCase().includes(searchTerm.toLowerCase())
      : true;
    return matchesSearch;
  });

  const handleCreateConfig = () => {
    if (formData.tenant_id && formData.name.trim()) {
      createMutation.mutate(formData);
    }
  };

  const handleUpdateConfig = () => {
    if (selectedConfig) {
      updateMutation.mutate({
        id: selectedConfig.id,
        updates: {
          name: formData.name,
          destination_type: formData.destination_type as any,
          format: formData.format as any,
          enabled: formData.enabled,
          config: formData.config,
        },
      });
    }
  };

  const handleDeleteConfig = (config: SIEMConfig) => {
    if (confirm(`Are you sure you want to delete "${config.name}"?`)) {
      deleteMutation.mutate(config.id);
    }
  };

  const handleTestConnection = (config: SIEMConfig) => {
    setSelectedConfig(config);
    setTestResult(null);
    setIsTestDialogOpen(true);
    testMutation.mutate(config.id);
  };

  const openEditDialog = (config: SIEMConfig) => {
    setSelectedConfig(config);
    setFormData({
      tenant_id: config.tenant_id,
      name: config.name,
      destination_type: config.destination_type,
      format: config.format,
      enabled: config.enabled,
      config: config.config,
    } as any);
  };

  const getDestinationIcon = (type: SIEMDestination) => {
    const dest = SIEM_DESTINATION_TYPES.find((d) => d.value === type);
    return dest?.icon || '📊';
  };

  const getDestinationLabel = (type: SIEMDestination) => {
    const dest = SIEM_DESTINATION_TYPES.find((d) => d.value === type);
    return dest?.label || type;
  };

  const toggleShowSecret = (id: string) => {
    setShowSecrets((prev) => ({ ...prev, [id]: !prev[id] }));
  };

  const maskSecret = (value: string | undefined) => {
    if (!value) return '-';
    return '••••••••••••';
  };

  if (isLoading) {
    return <LoadingScreen />;
  }

  if (isError) {
    return (
      <div className="p-8 bg-red-50 border border-red-200 rounded-lg dark:bg-red-900/20 dark:border-red-800">
        <h3 className="font-semibold text-red-900 dark:text-red-200">Error loading SIEM configurations</h3>
        <p className="text-red-700 dark:text-red-300 mt-2">Failed to fetch SIEM data.</p>
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
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">SIEM Configuration</h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">
            Configure security event exports to your SIEM systems
          </p>
        </div>

        <div className="flex gap-2">
          <button
            onClick={() => refetch()}
            disabled={isFetching}
            className="flex items-center gap-2 px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700"
          >
            <RefreshCw className={`w-5 h-5 ${isFetching ? 'animate-spin' : ''}`} />
            Refresh
          </button>
          <button
            onClick={() => {
              resetFormData();
              setIsCreateDialogOpen(true);
            }}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            <Plus className="w-5 h-5" />
            Add Configuration
          </button>
        </div>
      </div>

      {/* Filters */}
      <div className="flex gap-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-3 w-5 h-5 text-gray-400" />
          <input
            type="text"
            placeholder="Search configurations..."
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

      {/* Configurations Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {filteredConfigs.length === 0 ? (
          <div className="col-span-2 p-8 text-center text-gray-500 dark:text-gray-400 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
            No SIEM configurations found
          </div>
        ) : (
          filteredConfigs.map((config) => {
            const tenant = tenants.find((t) => t.id === config.tenant_id);

            return (
              <div
                key={config.id}
                className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4"
              >
                {/* Header */}
                <div className="flex items-start justify-between mb-4">
                  <div className="flex items-center gap-3">
                    <span className="text-2xl">{getDestinationIcon(config.destination_type)}</span>
                    <div>
                      <h3 className="font-semibold text-gray-900 dark:text-white">{config.name}</h3>
                      <p className="text-sm text-gray-500 dark:text-gray-400">
                        {getDestinationLabel(config.destination_type)} • {config.format.toUpperCase()}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {config.enabled ? (
                      <span className="inline-flex items-center gap-1 px-2 py-1 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 rounded text-xs">
                        <CheckCircle className="w-3 h-3" />
                        Active
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1 px-2 py-1 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 rounded text-xs">
                        <XCircle className="w-3 h-3" />
                        Disabled
                      </span>
                    )}
                  </div>
                </div>

                {/* Config Details */}
                <div className="space-y-2 mb-4 text-sm">
                  <div className="flex justify-between">
                    <span className="text-gray-500 dark:text-gray-400">Tenant:</span>
                    <span className="text-gray-900 dark:text-white">{tenant?.name || 'Unknown'}</span>
                  </div>

                  {/* Destination-specific details */}
                  {config.destination_type === 'webhook' && (
                    <div className="flex justify-between items-center">
                      <span className="text-gray-500 dark:text-gray-400">URL:</span>
                      <span className="text-gray-900 dark:text-white font-mono text-xs truncate max-w-[200px]">
                        {config.config.webhook_url || '-'}
                      </span>
                    </div>
                  )}

                  {config.destination_type === 'cloudwatch' && (
                    <>
                      <div className="flex justify-between">
                        <span className="text-gray-500 dark:text-gray-400">Log Group:</span>
                        <span className="text-gray-900 dark:text-white font-mono text-xs">
                          {config.config.cloudwatch_log_group || '-'}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-gray-500 dark:text-gray-400">Region:</span>
                        <span className="text-gray-900 dark:text-white">
                          {config.config.cloudwatch_region || '-'}
                        </span>
                      </div>
                    </>
                  )}

                  {config.destination_type === 'azure' && (
                    <div className="flex justify-between items-center">
                      <span className="text-gray-500 dark:text-gray-400">Workspace ID:</span>
                      <span className="text-gray-900 dark:text-white font-mono text-xs">
                        {showSecrets[config.id] 
                          ? (config.config.azure_workspace_id || '-')
                          : maskSecret(config.config.azure_workspace_id)}
                      </span>
                      <button
                        onClick={() => toggleShowSecret(config.id)}
                        className="ml-2 text-gray-400 hover:text-gray-600"
                      >
                        {showSecrets[config.id] ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                      </button>
                    </div>
                  )}

                  {config.destination_type === 'splunk' && (
                    <>
                      <div className="flex justify-between">
                        <span className="text-gray-500 dark:text-gray-400">Host:</span>
                        <span className="text-gray-900 dark:text-white font-mono text-xs">
                          {config.config.splunk_host || '-'}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-gray-500 dark:text-gray-400">Index:</span>
                        <span className="text-gray-900 dark:text-white">
                          {config.config.splunk_index || '-'}
                        </span>
                      </div>
                    </>
                  )}

                  {config.destination_type === 'gcp' && (
                    <>
                      <div className="flex justify-between">
                        <span className="text-gray-500 dark:text-gray-400">Project:</span>
                        <span className="text-gray-900 dark:text-white font-mono text-xs">
                          {config.config.gcp_project_id || '-'}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-gray-500 dark:text-gray-400">Log Name:</span>
                        <span className="text-gray-900 dark:text-white">
                          {config.config.gcp_log_name || '-'}
                        </span>
                      </div>
                    </>
                  )}

                  {config.last_export_at && (
                    <div className="flex justify-between">
                      <span className="text-gray-500 dark:text-gray-400">Last Export:</span>
                      <span className="text-gray-900 dark:text-white">
                        {new Date(config.last_export_at).toLocaleString()}
                      </span>
                    </div>
                  )}
                </div>

                {/* Actions */}
                <div className="flex gap-2 pt-3 border-t border-gray-200 dark:border-gray-700">
                  <button
                    onClick={() => handleTestConnection(config)}
                    disabled={testMutation.isPending}
                    className="flex items-center gap-1 px-3 py-1.5 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 rounded text-sm hover:bg-green-200 dark:hover:bg-green-900/50"
                  >
                    <Play className="w-4 h-4" />
                    Test
                  </button>
                  <button
                    onClick={() => openEditDialog(config)}
                    className="flex items-center gap-1 px-3 py-1.5 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded text-sm hover:bg-gray-200 dark:hover:bg-gray-600"
                  >
                    <Edit2 className="w-4 h-4" />
                    Edit
                  </button>
                  <button
                    onClick={() => handleDeleteConfig(config)}
                    disabled={deleteMutation.isPending}
                    className="flex items-center gap-1 px-3 py-1.5 bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400 rounded text-sm hover:bg-red-200 dark:hover:bg-red-900/50"
                  >
                    <Trash2 className="w-4 h-4" />
                    Delete
                  </button>
                </div>
              </div>
            );
          })
        )}
      </div>

      {/* Create Dialog */}
      {isCreateDialogOpen && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-bold mb-4 text-gray-900 dark:text-white">Add SIEM Configuration</h2>
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
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
                    Configuration Name
                  </label>
                  <input
                    type="text"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    placeholder="e.g., Production Splunk"
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Destination Type
                  </label>
                  <select
                    value={formData.destination_type}
                    onChange={(e) => setFormData({ 
                      ...formData, 
                      destination_type: e.target.value as SIEMDestination,
                      config: {} 
                    })}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                  >
                    {SIEM_DESTINATION_TYPES.map((type) => (
                      <option key={type.value} value={type.value}>
                        {type.icon} {type.label}
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Export Format
                  </label>
                  <select
                    value={formData.format}
                    onChange={(e) => setFormData({ ...formData, format: e.target.value as SIEMFormat })}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                  >
                    {SIEM_FORMATS.map((format) => (
                      <option key={format.value} value={format.value}>
                        {format.label}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              {/* Destination-specific config fields */}
              {formData.destination_type === 'webhook' && (
                <div className="space-y-4 p-4 bg-gray-50 dark:bg-gray-900 rounded-lg">
                  <h4 className="font-medium text-gray-900 dark:text-white">Webhook Configuration</h4>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                      Webhook URL
                    </label>
                    <input
                      type="url"
                      value={formData.config.webhook_url || ''}
                      onChange={(e) => setFormData({ 
                        ...formData, 
                        config: { ...formData.config, webhook_url: e.target.value } 
                      })}
                      placeholder="https://your-siem.example.com/webhook"
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                    />
                  </div>
                </div>
              )}

              {formData.destination_type === 'cloudwatch' && (
                <div className="space-y-4 p-4 bg-gray-50 dark:bg-gray-900 rounded-lg">
                  <h4 className="font-medium text-gray-900 dark:text-white">CloudWatch Configuration</h4>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        Log Group Name
                      </label>
                      <input
                        type="text"
                        value={formData.config.cloudwatch_log_group || ''}
                        onChange={(e) => setFormData({ 
                          ...formData, 
                          config: { ...formData.config, cloudwatch_log_group: e.target.value } 
                        })}
                        placeholder="/functionfly/security-events"
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        Region
                      </label>
                      <input
                        type="text"
                        value={formData.config.cloudwatch_region || ''}
                        onChange={(e) => setFormData({ 
                          ...formData, 
                          config: { ...formData.config, cloudwatch_region: e.target.value } 
                        })}
                        placeholder="us-east-1"
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                      />
                    </div>
                  </div>
                </div>
              )}

              {formData.destination_type === 'azure' && (
                <div className="space-y-4 p-4 bg-gray-50 dark:bg-gray-900 rounded-lg">
                  <h4 className="font-medium text-gray-900 dark:text-white">Azure Log Analytics Configuration</h4>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        Workspace ID
                      </label>
                      <input
                        type="text"
                        value={formData.config.azure_workspace_id || ''}
                        onChange={(e) => setFormData({ 
                          ...formData, 
                          config: { ...formData.config, azure_workspace_id: e.target.value } 
                        })}
                        placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        Shared Key
                      </label>
                      <input
                        type="password"
                        value={formData.config.azure_shared_key || ''}
                        onChange={(e) => setFormData({ 
                          ...formData, 
                          config: { ...formData.config, azure_shared_key: e.target.value } 
                        })}
                        placeholder="Enter shared key"
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                      />
                    </div>
                  </div>
                </div>
              )}

              {formData.destination_type === 'gcp' && (
                <div className="space-y-4 p-4 bg-gray-50 dark:bg-gray-900 rounded-lg">
                  <h4 className="font-medium text-gray-900 dark:text-white">Google Cloud Logging Configuration</h4>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        Project ID
                      </label>
                      <input
                        type="text"
                        value={formData.config.gcp_project_id || ''}
                        onChange={(e) => setFormData({ 
                          ...formData, 
                          config: { ...formData.config, gcp_project_id: e.target.value } 
                        })}
                        placeholder="my-project-id"
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        Log Name
                      </label>
                      <input
                        type="text"
                        value={formData.config.gcp_log_name || ''}
                        onChange={(e) => setFormData({ 
                          ...formData, 
                          config: { ...formData.config, gcp_log_name: e.target.value } 
                        })}
                        placeholder="functionfly-security"
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                      />
                    </div>
                  </div>
                </div>
              )}

              {formData.destination_type === 'splunk' && (
                <div className="space-y-4 p-4 bg-gray-50 dark:bg-gray-900 rounded-lg">
                  <h4 className="font-medium text-gray-900 dark:text-white">Splunk Configuration</h4>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        Splunk Host
                      </label>
                      <input
                        type="text"
                        value={formData.config.splunk_host || ''}
                        onChange={(e) => setFormData({ 
                          ...formData, 
                          config: { ...formData.config, splunk_host: e.target.value } 
                        })}
                        placeholder="https://splunk.example.com:8088"
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        Index
                      </label>
                      <input
                        type="text"
                        value={formData.config.splunk_index || ''}
                        onChange={(e) => setFormData({ 
                          ...formData, 
                          config: { ...formData.config, splunk_index: e.target.value } 
                        })}
                        placeholder="security"
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                      />
                    </div>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                      HEC Token
                    </label>
                    <input
                      type="password"
                      value={formData.config.splunk_token || ''}
                      onChange={(e) => setFormData({ 
                        ...formData, 
                        config: { ...formData.config, splunk_token: e.target.value } 
                      })}
                      placeholder="Enter HEC token"
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                    />
                  </div>
                </div>
              )}

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="enabled"
                  checked={formData.enabled}
                  onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
                  className="rounded border-gray-300 dark:border-gray-600"
                />
                <label htmlFor="enabled" className="text-sm text-gray-700 dark:text-gray-300">
                  Enable this configuration immediately
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
                  onClick={handleCreateConfig}
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

      {/* Edit Dialog */}
      {isEditDialogOpen && selectedConfig && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-bold mb-4 text-gray-900 dark:text-white">Edit SIEM Configuration</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Configuration Name
                </label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Destination Type
                  </label>
                  <select
                    value={formData.destination_type}
                    onChange={(e) => setFormData({ 
                      ...formData, 
                      destination_type: e.target.value as SIEMDestination,
                      config: {} 
                    })}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                  >
                    {SIEM_DESTINATION_TYPES.map((type) => (
                      <option key={type.value} value={type.value}>
                        {type.icon} {type.label}
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Export Format
                  </label>
                  <select
                    value={formData.format}
                    onChange={(e) => setFormData({ ...formData, format: e.target.value as SIEMFormat })}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                  >
                    {SIEM_FORMATS.map((format) => (
                      <option key={format.value} value={format.value}>
                        {format.label}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              {/* Webhook URL for edit */}
              {formData.destination_type === 'webhook' && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Webhook URL
                  </label>
                  <input
                    type="url"
                    value={formData.config.webhook_url || ''}
                    onChange={(e) => setFormData({ 
                      ...formData, 
                      config: { ...formData.config, webhook_url: e.target.value } 
                    })}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                  />
                </div>
              )}

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="edit_enabled"
                  checked={formData.enabled}
                  onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
                  className="rounded border-gray-300 dark:border-gray-600"
                />
                <label htmlFor="edit_enabled" className="text-sm text-gray-700 dark:text-gray-300">
                  Configuration is enabled
                </label>
              </div>

              <div className="flex gap-3 pt-4">
                <button
                  onClick={() => {
                    setIsEditDialogOpen(false);
                    setSelectedConfig(null);
                  }}
                  className="flex-1 px-4 py-2 bg-gray-200 dark:bg-gray-700 text-gray-800 dark:text-gray-200 rounded-lg hover:bg-gray-300 dark:hover:bg-gray-600"
                >
                  Cancel
                </button>
                <button
                  onClick={handleUpdateConfig}
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

      {/* Test Result Dialog */}
      {isTestDialogOpen && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 max-w-md w-full">
            <h2 className="text-xl font-bold mb-4 text-gray-900 dark:text-white">
              Test Connection: {selectedConfig?.name}
            </h2>

            {testMutation.isPending ? (
              <div className="flex flex-col items-center py-8">
                <RefreshCw className="w-8 h-8 animate-spin text-blue-600 mb-4" />
                <p className="text-gray-600 dark:text-gray-400">Testing connection...</p>
              </div>
            ) : testResult ? (
              <div className="space-y-4">
                <div className={`p-4 rounded-lg ${
                  testResult.success 
                    ? 'bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800' 
                    : 'bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800'
                }`}>
                  <div className="flex items-center gap-2">
                    {testResult.success ? (
                      <CheckCircle className="w-6 h-6 text-green-600 dark:text-green-400" />
                    ) : (
                      <XCircle className="w-6 h-6 text-red-600 dark:text-red-400" />
                    )}
                    <span className={`font-medium ${
                      testResult.success 
                        ? 'text-green-800 dark:text-green-200' 
                        : 'text-red-800 dark:text-red-200'
                    }`}>
                      {testResult.success ? 'Connection Successful' : 'Connection Failed'}
                    </span>
                  </div>
                  <p className={`mt-2 text-sm ${
                    testResult.success 
                      ? 'text-green-700 dark:text-green-300' 
                      : 'text-red-700 dark:text-red-300'
                  }`}>
                    {testResult.message}
                  </p>
                  {testResult.latency_ms && (
                    <p className="mt-2 text-xs text-gray-500 dark:text-gray-400">
                      Latency: {testResult.latency_ms}ms
                    </p>
                  )}
                </div>
              </div>
            ) : null}

            <button
              onClick={() => {
                setIsTestDialogOpen(false);
                setTestResult(null);
                setSelectedConfig(null);
              }}
              className="w-full mt-4 px-4 py-2 bg-gray-200 dark:bg-gray-700 text-gray-800 dark:text-gray-200 rounded-lg hover:bg-gray-300 dark:hover:bg-gray-600"
            >
              Close
            </button>
          </div>
        </div>
      )}

      {/* Summary Stats */}
      <div className="grid grid-cols-4 gap-4">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
          <p className="text-gray-600 dark:text-gray-400 text-sm">Total Configurations</p>
          <p className="text-2xl font-bold text-gray-900 dark:text-white">{configs.length}</p>
        </div>
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
          <p className="text-gray-600 dark:text-gray-400 text-sm">Active</p>
          <p className="text-2xl font-bold text-green-600 dark:text-green-400">
            {configs.filter((c) => c.enabled).length}
          </p>
        </div>
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
          <p className="text-gray-600 dark:text-gray-400 text-sm">Disabled</p>
          <p className="text-2xl font-bold text-gray-600 dark:text-gray-400">
            {configs.filter((c) => !c.enabled).length}
          </p>
        </div>
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
          <p className="text-gray-600 dark:text-gray-400 text-sm">Destinations</p>
          <p className="text-2xl font-bold text-blue-600 dark:text-blue-400">
            {new Set(configs.map((c) => c.destination_type)).size}
          </p>
        </div>
      </div>
    </div>
  );
}

export default AdminSIEMPage;
