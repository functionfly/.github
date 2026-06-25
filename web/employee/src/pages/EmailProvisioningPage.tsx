import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { emailApi, type EmailAccount } from '@/api/email';
import { Mail, Plus, Search, Globe, Users, Ban, CheckCircle, Pause } from 'lucide-react';
import { toast } from 'sonner';

const statusColors: Record<string, string> = {
  active: 'bg-green-500/20 text-green-400',
  suspended: 'bg-yellow-500/20 text-yellow-400',
  deactivated: 'bg-red-500/20 text-red-400',
  provisioning: 'bg-blue-500/20 text-blue-400',
};

const statusIcons: Record<string, typeof CheckCircle> = {
  active: CheckCircle,
  suspended: Pause,
  deactivated: Ban,
};

export function EmailProvisioningPage() {
  const queryClient = useQueryClient();
  const [showProvision, setShowProvision] = useState(false);
  const [search, setSearch] = useState('');
  const [form, setForm] = useState({ employee_id: '', display_name: '', aliases: '' });

  const { data, isLoading } = useQuery({
    queryKey: ['email-accounts'],
    queryFn: () => emailApi.list(),
  });

  const provisionMutation = useMutation({
    mutationFn: () => {
      const aliases = form.aliases
        .split(',')
        .map((a) => a.trim())
        .filter(Boolean);
      return emailApi.provision(form.employee_id, {
        display_name: form.display_name || undefined,
        aliases: aliases.length > 0 ? aliases : undefined,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['email-accounts'] });
      toast.success('Email account provisioned');
      setShowProvision(false);
      setForm({ employee_id: '', display_name: '', aliases: '' });
    },
    onError: () => toast.error('Failed to provision email'),
  });

  const statusMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) => emailApi.updateStatus(id, status),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['email-accounts'] });
      toast.success('Status updated');
    },
    onError: () => toast.error('Failed to update status'),
  });

  const accounts = (data?.data?.accounts || []).filter(
    (a) =>
      !search ||
      a.email.toLowerCase().includes(search.toLowerCase()) ||
      (a.display_name && a.display_name.toLowerCase().includes(search.toLowerCase())),
  );

  const activeCount = (data?.data?.accounts || []).filter((a) => a.status === 'active').length;
  const suspendedCount = (data?.data?.accounts || []).filter((a) => a.status === 'suspended').length;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Mail className="h-6 w-6 text-blue-400" />
          <h1 className="text-2xl font-bold">Email Provisioning</h1>
        </div>
        <button
          onClick={() => setShowProvision(true)}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          <Plus className="h-4 w-4" />
          Provision Email
        </button>
      </div>

      <div className="grid grid-cols-3 gap-3">
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <div className="flex items-center gap-2 text-gray-400">
            <Mail className="h-4 w-4 text-blue-400" />
            <span className="text-sm">Total Accounts</span>
          </div>
          <p className="mt-1 text-2xl font-bold">{data?.data?.accounts?.length || 0}</p>
        </div>
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <div className="flex items-center gap-2 text-gray-400">
            <CheckCircle className="h-4 w-4 text-green-400" />
            <span className="text-sm">Active</span>
          </div>
          <p className="mt-1 text-2xl font-bold text-green-400">{activeCount}</p>
        </div>
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <div className="flex items-center gap-2 text-gray-400">
            <Pause className="h-4 w-4 text-yellow-400" />
            <span className="text-sm">Suspended</span>
          </div>
          <p className="mt-1 text-2xl font-bold text-yellow-400">{suspendedCount}</p>
        </div>
      </div>

      <div className="flex items-center gap-3">
        <Search className="h-4 w-4 text-gray-400" />
        <input
          type="text"
          placeholder="Search accounts..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-64 rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
        />
        <div className="ml-auto flex items-center gap-2 text-sm text-gray-400">
          <Globe className="h-4 w-4" />
          Provider: <span className="font-medium text-gray-200">Spacemail</span>
        </div>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : accounts.length === 0 ? (
        <div className="rounded-xl border border-gray-800 bg-gray-900 py-16 text-center text-gray-500">
          No email accounts found
        </div>
      ) : (
        <div className="overflow-hidden rounded-xl border border-gray-800">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-800 bg-gray-900/50 text-left text-sm text-gray-400">
                <th className="px-4 py-3 font-medium">Email</th>
                <th className="px-4 py-3 font-medium">Display Name</th>
                <th className="px-4 py-3 font-medium">Employee ID</th>
                <th className="px-4 py-3 font-medium">Aliases</th>
                <th className="px-4 py-3 font-medium">Groups</th>
                <th className="px-4 py-3 font-medium">Status</th>
                <th className="px-4 py-3 font-medium">Actions</th>
              </tr>
            </thead>
            <tbody>
              {accounts.map((account) => {
                const StatusIcon = statusIcons[account.status] || CheckCircle;
                return (
                  <tr key={account.id} className="border-b border-gray-800/50 hover:bg-gray-900/30">
                    <td className="px-4 py-3">
                      <span className="font-medium">{account.email}</span>
                    </td>
                    <td className="px-4 py-3 text-gray-400">{account.display_name || '—'}</td>
                    <td className="px-4 py-3">
                      <code className="rounded bg-gray-800 px-2 py-0.5 text-xs">{account.employee_id}</code>
                    </td>
                    <td className="px-4 py-3">
                      {account.aliases.length > 0 ? (
                        <div className="flex flex-wrap gap-1">
                          {account.aliases.map((alias) => (
                            <span key={alias} className="rounded bg-gray-800 px-2 py-0.5 text-xs text-gray-400">
                              {alias}
                            </span>
                          ))}
                        </div>
                      ) : (
                        <span className="text-gray-600">—</span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      {account.groups.length > 0 ? (
                        <div className="flex flex-wrap gap-1">
                          {account.groups.map((g) => (
                            <span key={g} className="flex items-center gap-1 rounded bg-blue-500/10 px-2 py-0.5 text-xs text-blue-400">
                              <Users className="h-3 w-3" />
                              {g}
                            </span>
                          ))}
                        </div>
                      ) : (
                        <span className="text-gray-600">—</span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium ${statusColors[account.status] || 'bg-gray-500/20 text-gray-400'}`}
                      >
                        <StatusIcon className="h-3 w-3" />
                        {account.status}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <select
                        value={account.status}
                        onChange={(e) => statusMutation.mutate({ id: account.id, status: e.target.value })}
                        className="rounded border border-gray-700 bg-gray-800 px-2 py-1 text-xs text-gray-200"
                      >
                        <option value="active">Active</option>
                        <option value="suspended">Suspended</option>
                        <option value="deactivated">Deactivated</option>
                      </select>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {showProvision && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
          <div className="w-full max-w-lg rounded-xl border border-gray-700 bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Provision Email Account</h2>
            <div className="space-y-4">
              <div>
                <label className="mb-1 block text-sm text-gray-400">Employee ID</label>
                <input
                  type="text"
                  value={form.employee_id}
                  onChange={(e) => setForm({ ...form, employee_id: e.target.value })}
                  placeholder="Enter employee ID"
                  className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
                />
              </div>
              <div>
                <label className="mb-1 block text-sm text-gray-400">Display Name</label>
                <input
                  type="text"
                  value={form.display_name}
                  onChange={(e) => setForm({ ...form, display_name: e.target.value })}
                  placeholder="e.g. John Smith"
                  className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
                />
              </div>
              <div>
                <label className="mb-1 block text-sm text-gray-400">Aliases (comma-separated)</label>
                <input
                  type="text"
                  value={form.aliases}
                  onChange={(e) => setForm({ ...form, aliases: e.target.value })}
                  placeholder="e.g. j.smith, johnsmith"
                  className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
                />
              </div>
              <div className="flex items-center gap-2 text-sm text-gray-400">
                <Globe className="h-4 w-4" />
                Provider: Spacemail
              </div>
            </div>
            <div className="mt-6 flex justify-end gap-3">
              <button
                onClick={() => setShowProvision(false)}
                className="rounded-lg border border-gray-700 px-4 py-2 text-sm text-gray-300 hover:bg-gray-800"
              >
                Cancel
              </button>
              <button
                onClick={() => provisionMutation.mutate()}
                disabled={!form.employee_id || provisionMutation.isPending}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                {provisionMutation.isPending ? 'Provisioning...' : 'Provision'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
