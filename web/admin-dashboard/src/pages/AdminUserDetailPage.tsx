import { useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { ROUTES } from '@/lib/constants';
import { type AdminUser } from '@/types';
import { Pencil, X, Crown } from 'lucide-react';

const PLAN_OPTIONS = ['free', 'starter', 'pro', 'enterprise', 'agent_starter', 'agent_scale', 'agent_pro', 'agent_enterprise'] as const;
const PLAN_LABELS: Record<string, string> = {
  free: 'Free',
  starter: 'Starter',
  pro: 'Pro',
  enterprise: 'Enterprise',
  agent_starter: 'Agent Starter',
  agent_scale: 'Agent Scale',
  agent_pro: 'Agent Pro',
  agent_enterprise: 'Agent Enterprise',
};

const ROLE_OPTIONS = [
  { value: 'super_admin', label: 'Super Admin' },
  { value: 'admin', label: 'Admin' },
  { value: 'moderator', label: 'Moderator' },
  { value: 'user', label: 'User' },
];

export function AdminUserDetailPage() {
  const { userId } = useParams<{ userId: string }>();
  const queryClient = useQueryClient();
  const [editing, setEditing] = useState(false);
  const [editingField, setEditingField] = useState<'role' | 'plan' | null>(null);
  const [editName, setEditName] = useState('');
  const [editEmail, setEditEmail] = useState('');
  const [editRole, setEditRole] = useState('');
  const [editPlan, setEditPlan] = useState('');

  const { data, isLoading } = useQuery({
    queryKey: ['admin-user-detail', userId],
    queryFn: async () => {
      if (!userId) return null;
      try {
        return await adminApiClient.get<AdminUser>(`/users/${userId}`);
      } catch {
        return null;
      }
    },
    enabled: Boolean(userId),
  });

  const updateMutation = useMutation({
    mutationFn: async (payload: { name?: string; email?: string; role?: string; plan?: string }) => {
      const res = await adminApiClient.patch<AdminUser>(`/users/${userId!}`, payload);
      // res is AdminAPIResponse<AdminUser> = { data: AdminUser, success: boolean, ... }
      return (res as { data?: AdminUser })?.data ?? res as AdminUser;
    },
    onSuccess: (updated) => {
      if (updated) {
        queryClient.setQueryData(['admin-user-detail', userId], updated);
        queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      }
      setEditingField(null);
    },
  });

  const startEditing = (user: AdminUser) => {
    setEditName(user.name ?? '');
    setEditEmail(user.email ?? '');
    setEditRole(user.role ?? 'user');
    setEditing(true);
    setEditingField('role');
  };

  const startEditingPlan = (user: AdminUser) => {
    setEditPlan(user.plan ?? 'free');
    setEditingField('plan');
  };

  const cancelEditing = () => {
    setEditing(false);
    setEditingField(null);
    updateMutation.reset();
  };

  if (isLoading) return <LoadingScreen />;

  // Backend returns the user object directly (not wrapped in { data }); support both shapes.
  const user =
    data && typeof data === 'object' && 'email' in data && 'id' in data
      ? (data as unknown as AdminUser)
      : (data as { data?: AdminUser } | undefined)?.data;

  if (!user) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">User not found</h1>
        <Link to={ROUTES.ADMIN_USERS} className="text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300">Back to users</Link>
      </div>
    );
  }

  if (editingField === 'plan') {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">Change Plan</h1>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={cancelEditing}
              className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 flex items-center gap-2 text-gray-900 dark:text-gray-100"
            >
              <X className="w-4 h-4" /> Cancel
            </button>
            <button
              type="button"
              onClick={() => updateMutation.mutate({ plan: editPlan })}
              disabled={updateMutation.isPending}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 flex items-center gap-2"
            >
              {updateMutation.isPending ? 'Saving...' : 'Save plan'}
            </button>
          </div>
        </div>

        {updateMutation.isError && (
          <div className="p-4 bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded-lg text-red-800 dark:text-red-300 text-sm">
            {updateMutation.error instanceof Error ? updateMutation.error.message : 'Failed to update plan'}
          </div>
        )}

        <div className="bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700 p-6 max-w-xl">
          <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">
            Changing plan for <strong>{user.email}</strong>
          </p>
          <div className="grid grid-cols-2 gap-3">
            {PLAN_OPTIONS.map((p) => (
              <button
                key={p}
                onClick={() => setEditPlan(p)}
                className={`p-4 rounded-lg border-2 text-left transition-all ${
                  editPlan === p
                    ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/30'
                    : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
                }`}
              >
                <div className="flex items-center gap-2">
                  <Crown className={`w-5 h-5 ${editPlan === p ? 'text-blue-600 dark:text-blue-400' : 'text-gray-400'}`} />
                  <span className="font-medium text-gray-900 dark:text-gray-100">{PLAN_LABELS[p]}</span>
                </div>
                <p className="text-xs text-gray-500 dark:text-gray-400 mt-1 capitalize">{p}</p>
              </button>
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (editingField === 'role') {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">Edit user</h1>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={cancelEditing}
              className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 flex items-center gap-2 text-gray-900 dark:text-gray-100"
            >
              <X className="w-4 h-4" /> Cancel
            </button>
            <button
              type="button"
              onClick={() => updateMutation.mutate({ name: editName.trim() || undefined, email: editEmail.trim(), role: editRole })}
              disabled={updateMutation.isPending}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 flex items-center gap-2"
            >
              {updateMutation.isPending ? 'Saving...' : 'Save changes'}
            </button>
          </div>
        </div>

        {updateMutation.isError && (
          <div className="p-4 bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded-lg text-red-800 dark:text-red-300 text-sm">
            {updateMutation.error instanceof Error ? updateMutation.error.message : 'Failed to update user'}
          </div>
        )}

        <div className="bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700 p-6 space-y-4 max-w-xl">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Name</label>
            <input
              type="text"
              value={editName}
              onChange={(e) => setEditName(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
              placeholder="Display name"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Email</label>
            <input
              type="email"
              value={editEmail}
              onChange={(e) => setEditEmail(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
              placeholder="user@example.com"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Role</label>
            <select
              value={editRole}
              onChange={(e) => setEditRole(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
            >
              {ROLE_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">{user.name?.trim() || user.email}</h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">Admin user profile and permissions.</p>
        </div>
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => startEditing(user)}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 flex items-center gap-2"
          >
            <Pencil className="w-4 h-4" /> Edit user
          </button>
          <Link to={ROUTES.ADMIN_USERS} className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 text-gray-900 dark:text-gray-100">
            Back
          </Link>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Detail label="User ID" value={user.id} />
        <Detail label="Email" value={user.email} />
        <Detail label="Username" value={user.username ?? '—'} />
        <DetailWithAction label="Role" value={(user.role ?? 'user').replace(/_/g, ' ')} capitalize onEdit={() => setEditingField('role')} />
        <DetailWithAction label="Plan" value={PLAN_LABELS[user.plan ?? 'free'] ?? user.plan ?? 'free'} onEdit={() => startEditingPlan(user)} />
        <Detail label="MFA" value={user.mfa_enabled ? 'Enabled' : 'Disabled'} />
        <Detail label="Tenant" value={user.tenant_name ?? user.tenant_id ?? 'None'} />
        <Detail label="Created" value={new Date(user.created_at).toLocaleString()} />
        <Detail label="Updated" value={new Date(user.updated_at).toLocaleString()} />
      </div>

      <div className="bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700 p-4">
        <p className="text-sm text-gray-600 dark:text-gray-400">Permissions</p>
        <div className="mt-2 flex flex-wrap gap-2">
          {(user.permissions?.length ?? 0) === 0 ? (
            <span className="text-sm text-gray-500 dark:text-gray-400">No permissions</span>
          ) : (
            (user.permissions ?? []).map((perm) => (
              <span key={perm} className="px-2 py-1 text-xs rounded bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">
                {perm}
              </span>
            ))
          )}
        </div>
      </div>
    </div>
  );
}

function Detail({ label, value, capitalize }: { label: string; value: string; capitalize?: boolean }) {
  return (
    <div className="bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700 p-4">
      <p className="text-sm text-gray-600 dark:text-gray-400">{label}</p>
      <p className={`text-sm font-medium text-gray-900 dark:text-gray-100 mt-1 break-all ${capitalize ? 'capitalize' : ''}`}>{value}</p>
    </div>
  );
}

function DetailWithAction({ label, value, capitalize, onEdit }: { label: string; value: string; capitalize?: boolean; onEdit: () => void }) {
  return (
    <div className="bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700 p-4 flex items-center justify-between">
      <div>
        <p className="text-sm text-gray-600 dark:text-gray-400">{label}</p>
        <p className={`text-sm font-medium text-gray-900 dark:text-gray-100 mt-1 ${capitalize ? 'capitalize' : ''}`}>{value}</p>
      </div>
      <button
        onClick={onEdit}
        className="ml-4 p-2 text-gray-400 hover:text-blue-600 dark:hover:text-blue-400"
        title={`Edit ${label}`}
      >
        <Pencil className="w-4 h-4" />
      </button>
    </div>
  );
}
