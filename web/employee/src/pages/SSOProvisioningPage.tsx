import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import apiClient from '@/api/client';
import { Key, Plus, Search, RefreshCw, ArrowRight, AlertCircle, CheckCircle, XCircle } from 'lucide-react';
import { toast } from 'sonner';

interface SSOConfig {
  id: string;
  provider: string;
  display_name?: string;
  entity_url?: string;
  sso_url?: string;
  auto_create_users: boolean;
  field_mappings: Record<string, string>;
  status: string;
  last_synced_at?: string;
}

interface ProvisioningLogEntry {
  id: string;
  sso_config_id: string;
  employee_id: string;
  action: string;
  details?: string;
  created_at: string;
}

const providerColors: Record<string, string> = {
  okta: 'bg-blue-500/20 text-blue-400',
  azure_ad: 'bg-cyan-500/20 text-cyan-400',
  google_workspace: 'bg-yellow-500/20 text-yellow-400',
  onelogin: 'bg-purple-500/20 text-purple-400',
  saml: 'bg-green-500/20 text-green-400',
  oidc: 'bg-pink-500/20 text-pink-400',
};

const actionColors: Record<string, string> = {
  create: 'bg-green-500/20 text-green-400',
  update: 'bg-blue-500/20 text-blue-400',
  deactivate: 'bg-yellow-500/20 text-yellow-400',
  error: 'bg-red-500/20 text-red-400',
};

const actionIcons: Record<string, typeof CheckCircle> = {
  create: CheckCircle,
  update: RefreshCw,
  deactivate: XCircle,
  error: AlertCircle,
};

export function SSOProvisioningPage() {
  const queryClient = useQueryClient();
  const [showCreate, setShowCreate] = useState(false);
  const [search, setSearch] = useState('');
  const [form, setForm] = useState({
    provider: 'okta',
    display_name: '',
    entity_url: '',
    sso_url: '',
    auto_create_users: true,
    field_mappings: '{\n  "email": "email",\n  "name": "displayName"\n}',
  });

  const { data, isLoading } = useQuery({
    queryKey: ['sso-configs'],
    queryFn: () => apiClient.get<{ configs: SSOConfig[] }>('/v1/sso/configs'),
  });

  const { data: logData } = useQuery({
    queryKey: ['sso-provisioning-log'],
    queryFn: () => apiClient.get<{ entries: ProvisioningLogEntry[] }>('/v1/sso/provisioning-log'),
  });

  const createMutation = useMutation({
    mutationFn: () => {
      let mappings: Record<string, string> = {};
      try {
        mappings = JSON.parse(form.field_mappings);
      } catch {
        toast.error('Invalid field mappings JSON');
        return Promise.reject(new Error('Invalid JSON'));
      }
      return apiClient.post('/v1/sso/configs', {
        provider: form.provider,
        display_name: form.display_name || undefined,
        entity_url: form.entity_url || undefined,
        sso_url: form.sso_url || undefined,
        auto_create_users: form.auto_create_users,
        field_mappings: mappings,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sso-configs'] });
      toast.success('SSO config created');
      setShowCreate(false);
      setForm({ provider: 'okta', display_name: '', entity_url: '', sso_url: '', auto_create_users: true, field_mappings: '{\n  "email": "email",\n  "name": "displayName"\n}' });
    },
    onError: () => toast.error('Failed to create SSO config'),
  });

  const syncMutation = useMutation({
    mutationFn: (configId: string) => apiClient.post(`/v1/sso/configs/${configId}/sync`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sso-configs'] });
      queryClient.invalidateQueries({ queryKey: ['sso-provisioning-log'] });
      toast.success('Sync completed');
    },
    onError: () => toast.error('Sync failed'),
  });

  const configs = (data?.data?.configs || []).filter(
    (c) =>
      !search ||
      c.provider.toLowerCase().includes(search.toLowerCase()) ||
      (c.display_name && c.display_name.toLowerCase().includes(search.toLowerCase())),
  );

  const logEntries = logData?.data?.entries || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Key className="h-6 w-6 text-amber-400" />
          <h1 className="text-2xl font-bold">SSO Provisioning</h1>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          <Plus className="h-4 w-4" />
          Create Config
        </button>
      </div>

      <div className="flex items-center gap-3">
        <Search className="h-4 w-4 text-gray-400" />
        <input
          type="text"
          placeholder="Search SSO configs..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-64 rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
        />
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : configs.length === 0 ? (
        <div className="rounded-xl border border-gray-800 bg-gray-900 py-16 text-center text-gray-500">
          No SSO configurations found
        </div>
      ) : (
        <div className="space-y-3">
          {configs.map((config) => (
            <div
              key={config.id}
              className="flex items-center justify-between rounded-xl border border-gray-800 bg-gray-900 p-4"
            >
              <div className="flex items-center gap-4">
                <span
                  className={`rounded-lg px-3 py-1.5 text-xs font-semibold uppercase ${providerColors[config.provider] || 'bg-gray-500/20 text-gray-400'}`}
                >
                  {config.provider}
                </span>
                <div>
                  <p className="font-medium">{config.display_name || config.provider}</p>
                  <p className="text-xs text-gray-500">{config.entity_url || 'No entity URL'}</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                {config.auto_create_users && (
                  <span className="rounded bg-blue-500/10 px-2 py-0.5 text-xs text-blue-400">Auto-create</span>
                )}
                <span
                  className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
                    config.status === 'active' ? 'bg-green-500/20 text-green-400' : 'bg-gray-500/20 text-gray-400'
                  }`}
                >
                  {config.status}
                </span>
                {config.last_synced_at && (
                  <span className="text-xs text-gray-500">
                    Synced {new Date(config.last_synced_at).toLocaleString()}
                  </span>
                )}
                <button
                  onClick={() => syncMutation.mutate(config.id)}
                  disabled={syncMutation.isPending}
                  className="flex items-center gap-1.5 rounded-lg border border-gray-700 px-3 py-1.5 text-xs text-gray-300 hover:bg-gray-800 disabled:opacity-50"
                >
                  <RefreshCw className={`h-3 w-3 ${syncMutation.isPending ? 'animate-spin' : ''}`} />
                  Sync
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      <div>
        <h2 className="mb-3 text-lg font-semibold">Provisioning Log</h2>
        {logEntries.length === 0 ? (
          <div className="rounded-xl border border-gray-800 bg-gray-900 py-12 text-center text-gray-500">
            No provisioning activity
          </div>
        ) : (
          <div className="overflow-hidden rounded-xl border border-gray-800">
            <table className="w-full">
              <thead>
                <tr className="border-b border-gray-800 bg-gray-900/50 text-left text-sm text-gray-400">
                  <th className="px-4 py-3 font-medium">Action</th>
                  <th className="px-4 py-3 font-medium">Employee</th>
                  <th className="px-4 py-3 font-medium">Config</th>
                  <th className="px-4 py-3 font-medium">Details</th>
                  <th className="px-4 py-3 font-medium">Time</th>
                </tr>
              </thead>
              <tbody>
                {logEntries.slice(0, 20).map((entry) => {
                  const ActionIcon = actionIcons[entry.action] || ArrowRight;
                  return (
                    <tr key={entry.id} className="border-b border-gray-800/50 hover:bg-gray-900/30">
                      <td className="px-4 py-3">
                        <span
                          className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium ${actionColors[entry.action] || 'bg-gray-500/20 text-gray-400'}`}
                        >
                          <ActionIcon className="h-3 w-3" />
                          {entry.action}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <code className="rounded bg-gray-800 px-2 py-0.5 text-xs">{entry.employee_id}</code>
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-400">{entry.sso_config_id}</td>
                      <td className="max-w-xs truncate px-4 py-3 text-sm text-gray-500">{entry.details || '—'}</td>
                      <td className="px-4 py-3 text-xs text-gray-500">{new Date(entry.created_at).toLocaleString()}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
          <div className="w-full max-w-lg rounded-xl border border-gray-700 bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Create SSO Configuration</h2>
            <div className="space-y-4">
              <div>
                <label className="mb-1 block text-sm text-gray-400">Provider</label>
                <select
                  value={form.provider}
                  onChange={(e) => setForm({ ...form, provider: e.target.value })}
                  className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-200"
                >
                  <option value="okta">Okta</option>
                  <option value="azure_ad">Azure AD</option>
                  <option value="google_workspace">Google Workspace</option>
                  <option value="onelogin">OneLogin</option>
                  <option value="saml">Generic SAML</option>
                  <option value="oidc">OIDC</option>
                </select>
              </div>
              <div>
                <label className="mb-1 block text-sm text-gray-400">Display Name</label>
                <input
                  type="text"
                  value={form.display_name}
                  onChange={(e) => setForm({ ...form, display_name: e.target.value })}
                  placeholder="e.g. Company Okta"
                  className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
                />
              </div>
              <div>
                <label className="mb-1 block text-sm text-gray-400">Entity URL</label>
                <input
                  type="text"
                  value={form.entity_url}
                  onChange={(e) => setForm({ ...form, entity_url: e.target.value })}
                  placeholder="https://..."
                  className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
                />
              </div>
              <div>
                <label className="mb-1 block text-sm text-gray-400">SSO URL</label>
                <input
                  type="text"
                  value={form.sso_url}
                  onChange={(e) => setForm({ ...form, sso_url: e.target.value })}
                  placeholder="https://..."
                  className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
                />
              </div>
              <label className="flex items-center gap-2 text-sm text-gray-300">
                <input
                  type="checkbox"
                  checked={form.auto_create_users}
                  onChange={(e) => setForm({ ...form, auto_create_users: e.target.checked })}
                  className="rounded border-gray-600"
                />
                Auto-create users on first login
              </label>
              <div>
                <label className="mb-1 block text-sm text-gray-400">Field Mappings (JSON)</label>
                <textarea
                  value={form.field_mappings}
                  onChange={(e) => setForm({ ...form, field_mappings: e.target.value })}
                  rows={4}
                  className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 font-mono text-xs text-gray-100"
                />
              </div>
            </div>
            <div className="mt-6 flex justify-end gap-3">
              <button
                onClick={() => setShowCreate(false)}
                className="rounded-lg border border-gray-700 px-4 py-2 text-sm text-gray-300 hover:bg-gray-800"
              >
                Cancel
              </button>
              <button
                onClick={() => createMutation.mutate()}
                disabled={createMutation.isPending}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                {createMutation.isPending ? 'Creating...' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
