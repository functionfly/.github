/**
 * Admin Signup Invites Page
 * Manage invite-only beta signup codes.
 */

import { LoadingScreen } from '@/components/common/LoadingScreen';
import { adminApiClient } from '@/lib/api/adminClient';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Ban, Copy, Plus, RefreshCw } from 'lucide-react';
import { useState } from 'react';

interface SignupInvite {
  id: string;
  label: string;
  maxUses?: number;
  usesCount: number;
  expiresAt?: string;
  revokedAt?: string;
  createdBy?: string;
  createdAt: string;
}

interface CreateInviteResponse {
  code: string;
  id: string;
}

export function AdminSignupInvitesPage() {
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [label, setLabel] = useState('');
  const [maxUses, setMaxUses] = useState('');
  const [expiresAt, setExpiresAt] = useState('');
  const [newCode, setNewCode] = useState<string | null>(null);
  const [createError, setCreateError] = useState<string | null>(null);
  const queryClient = useQueryClient();

  const { data, isLoading, isError } = useQuery({
    queryKey: ['admin-signup-invites'],
    queryFn: async () => {
      const res = await adminApiClient.get<{ data: SignupInvite[] }>('/signup-invites');
      return res;
    },
    staleTime: 1000 * 60,
  });

  const invites: SignupInvite[] = data?.data?.data ?? [];

  const createMutation = useMutation({
    mutationFn: async (body: { label: string; maxUses?: number; expiresAt?: string }) => {
      return adminApiClient.post<CreateInviteResponse>('/signup-invites', body);
    },
    onSuccess: (res) => {
      // Backend returns flat {id, code, label} — not wrapped in AdminAPIResponse
      const flat = res as unknown as CreateInviteResponse;
      const code = flat?.code;
      if (code) setNewCode(code);
      queryClient.invalidateQueries({ queryKey: ['admin-signup-invites'] });
      setCreateError(null);
      setLabel('');
      setMaxUses('');
      setExpiresAt('');
      setIsCreateOpen(false);
    },
    onError: (err: unknown) => {
      const message =
        err instanceof Error ? err.message : 'Failed to create invite. Please try again.';
      setCreateError(message);
    },
  });

  const revokeMutation = useMutation({
    mutationFn: (id: string) => adminApiClient.post(`/signup-invites/${id}/revoke`, {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-signup-invites'] });
    },
  });

  const handleCreate = () => {
    setCreateError(null);
    const body: { label: string; maxUses?: number; expiresAt?: string } = { label };
    if (maxUses) body.maxUses = parseInt(maxUses, 10);
    if (expiresAt) body.expiresAt = new Date(expiresAt).toISOString();
    createMutation.mutate(body);
  };

  const handleCancelCreate = () => {
    setIsCreateOpen(false);
    setCreateError(null);
  };

  const copyCode = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  if (isLoading) return <LoadingScreen />;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Signup Invites</h1>
          <p className="text-gray-500 mt-1">Manage invite codes for the private beta.</p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => queryClient.invalidateQueries({ queryKey: ['admin-signup-invites'] })}
            className="inline-flex items-center gap-2 px-3 py-2 text-sm rounded-lg border hover:bg-gray-50"
          >
            <RefreshCw className="w-4 h-4" /> Refresh
          </button>
          <button
            onClick={() => {
              setCreateError(null);
              setIsCreateOpen(true);
            }}
            className="inline-flex items-center gap-2 px-4 py-2 text-sm rounded-lg bg-indigo-600 text-white hover:bg-indigo-700"
          >
            <Plus className="w-4 h-4" /> Create Invite
          </button>
        </div>
      </div>

      {/* New code banner */}
      {newCode && (
        <div className="rounded-lg bg-green-50 border border-green-200 p-4 flex items-center justify-between">
          <div>
            <p className="text-sm font-medium text-green-800">
              Invite code created — copy it now, it won't be shown again.
            </p>
            <code className="text-sm font-mono text-green-900 select-all">{newCode}</code>
          </div>
          <button
            onClick={() => copyCode(newCode)}
            className="inline-flex items-center gap-1 px-3 py-1.5 text-sm rounded border border-green-300 text-green-700 hover:bg-green-100"
          >
            <Copy className="w-3.5 h-3.5" /> Copy
          </button>
        </div>
      )}

      {/* Create dialog */}
      {isCreateOpen && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl shadow-xl w-full max-w-md p-6 space-y-4">
            <h2 className="text-lg font-semibold">Create Invite Code</h2>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Label</label>
              <input
                type="text"
                value={label}
                onChange={(e) => setLabel(e.target.value)}
                placeholder="e.g. Cohort 1, Partner Acme"
                className="w-full border rounded-lg px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Max Uses (optional)
              </label>
              <input
                type="number"
                value={maxUses}
                onChange={(e) => setMaxUses(e.target.value)}
                placeholder="Unlimited"
                className="w-full border rounded-lg px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Expires At (optional)
              </label>
              <input
                type="datetime-local"
                value={expiresAt}
                onChange={(e) => setExpiresAt(e.target.value)}
                className="w-full border rounded-lg px-3 py-2 text-sm"
              />
            </div>
            {createError && (
              <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
                <p className="text-sm text-red-700">{createError}</p>
              </div>
            )}
            <div className="flex gap-2 justify-end pt-2">
              <button
                onClick={handleCancelCreate}
                className="px-4 py-2 text-sm rounded-lg border hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                onClick={handleCreate}
                disabled={!label || createMutation.isPending}
                className="px-4 py-2 text-sm rounded-lg bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-50"
              >
                {createMutation.isPending ? 'Creating…' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}

      {isError && (
        <div className="rounded-lg bg-red-50 border border-red-200 p-4">
          <p className="text-sm text-red-700">Failed to load invite codes.</p>
        </div>
      )}

      {/* Table */}
      <div className="bg-white rounded-xl border overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 border-b">
            <tr>
              <th className="px-4 py-3 text-left font-medium text-gray-500">Label</th>
              <th className="px-4 py-3 text-left font-medium text-gray-500">Uses</th>
              <th className="px-4 py-3 text-left font-medium text-gray-500">Expires</th>
              <th className="px-4 py-3 text-left font-medium text-gray-500">Status</th>
              <th className="px-4 py-3 text-left font-medium text-gray-500">Created</th>
              <th className="px-4 py-3 text-right font-medium text-gray-500">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {invites.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-gray-400">
                  No invite codes yet. Create one to start the private beta.
                </td>
              </tr>
            ) : (
              invites.map((inv) => {
                const isRevoked = !!inv.revokedAt;
                const isExpired = inv.expiresAt && new Date(inv.expiresAt) < new Date();
                const isActive = !isRevoked && !isExpired;

                return (
                  <tr key={inv.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 font-medium text-gray-900">{inv.label || '—'}</td>
                    <td className="px-4 py-3 text-gray-600">
                      {inv.usesCount}
                      {inv.maxUses != null ? ` / ${inv.maxUses}` : ''}
                    </td>
                    <td className="px-4 py-3 text-gray-600">
                      {inv.expiresAt ? new Date(inv.expiresAt).toLocaleDateString() : 'Never'}
                    </td>
                    <td className="px-4 py-3">
                      {isActive && (
                        <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                          Active
                        </span>
                      )}
                      {isRevoked && (
                        <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">
                          Revoked
                        </span>
                      )}
                      {isExpired && !isRevoked && (
                        <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800">
                          Expired
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-gray-500">
                      {new Date(inv.createdAt).toLocaleDateString()}
                    </td>
                    <td className="px-4 py-3 text-right">
                      {isActive && (
                        <button
                          onClick={() => revokeMutation.mutate(inv.id)}
                          disabled={revokeMutation.isPending}
                          className="inline-flex items-center gap-1 px-2.5 py-1.5 text-xs rounded border border-red-200 text-red-600 hover:bg-red-50 disabled:opacity-50"
                        >
                          <Ban className="w-3 h-3" /> Revoke
                        </button>
                      )}
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
