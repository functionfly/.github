import { useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { ROUTES } from '@/lib/constants';
import type { AdminUser } from '@/types';
import { Pencil, X } from 'lucide-react';

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
  const [editName, setEditName] = useState('');
  const [editEmail, setEditEmail] = useState('');
  const [editRole, setEditRole] = useState('');

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
    mutationFn: async (payload: { name?: string; email?: string; role?: string }) => {
      const res = await adminApiClient.patch<AdminUser>(`/users/${userId!}`, payload);
      return res && typeof res === 'object' && 'email' in res ? (res as unknown as AdminUser) : (res as { data?: AdminUser })?.data;
    },
    onSuccess: (updated) => {
      queryClient.setQueryData(['admin-user-detail', userId], updated ?? data);
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      setEditing(false);
    },
  });

  const startEditing = (user: AdminUser) => {
    setEditName(user.name ?? '');
    setEditEmail(user.email ?? '');
    setEditRole(user.role ?? 'user');
    setEditing(true);
  };

  const cancelEditing = () => {
    setEditing(false);
    updateMutation.reset();
  };

  const handleSave = () => {
    const payload: { name?: string; email?: string; role?: string } = {};
    if (editName !== (user?.name ?? '')) payload.name = editName.trim() || undefined;
    if (editEmail !== user?.email) payload.email = editEmail.trim();
    if (editRole !== user?.role) payload.role = editRole;
    if (Object.keys(payload).length === 0) {
      setEditing(false);
      return;
    }
    updateMutation.mutate(payload);
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
        <h1 className="text-2xl font-bold text-gray-900">User not found</h1>
        <Link to={ROUTES.ADMIN_USERS} className="text-blue-600 hover:text-blue-700">Back to users</Link>
      </div>
    );
  }

  if (editing) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900">Edit user</h1>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={cancelEditing}
              className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 flex items-center gap-2"
            >
              <X className="w-4 h-4" /> Cancel
            </button>
            <button
              type="button"
              onClick={handleSave}
              disabled={updateMutation.isPending}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 flex items-center gap-2"
            >
              {updateMutation.isPending ? 'Saving...' : 'Save changes'}
            </button>
          </div>
        </div>

        {updateMutation.isError && (
          <div className="p-4 bg-red-50 border border-red-200 rounded-lg text-red-800 text-sm">
            {updateMutation.error instanceof Error ? updateMutation.error.message : 'Failed to update user'}
          </div>
        )}

        <div className="bg-white rounded-lg border border-gray-200 p-6 space-y-4 max-w-xl">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <input
              type="text"
              value={editName}
              onChange={(e) => setEditName(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="Display name"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
            <input
              type="email"
              value={editEmail}
              onChange={(e) => setEditEmail(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="user@example.com"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Role</label>
            <select
              value={editRole}
              onChange={(e) => setEditRole(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
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
          <h1 className="text-3xl font-bold text-gray-900">{user.name?.trim() || user.email}</h1>
          <p className="mt-2 text-gray-600">Admin user profile and permissions.</p>
        </div>
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => startEditing(user)}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 flex items-center gap-2"
          >
            <Pencil className="w-4 h-4" /> Edit user
          </button>
          <Link to={ROUTES.ADMIN_USERS} className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">
            Back
          </Link>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Detail label="User ID" value={user.id} />
        <Detail label="Email" value={user.email} />
        <Detail label="Role" value={user.role} />
        <Detail label="MFA" value={user.mfa_enabled ? 'Enabled' : 'Disabled'} />
        <Detail label="Created" value={new Date(user.created_at).toLocaleString()} />
        <Detail label="Updated" value={new Date(user.updated_at).toLocaleString()} />
      </div>

      <div className="bg-white rounded-lg border border-gray-200 p-4">
        <p className="text-sm text-gray-600">Permissions</p>
        <div className="mt-2 flex flex-wrap gap-2">
          {(user.permissions?.length ?? 0) === 0 ? (
            <span className="text-sm text-gray-500">No permissions</span>
          ) : (
            (user.permissions ?? []).map((perm) => (
              <span key={perm} className="px-2 py-1 text-xs rounded bg-blue-100 text-blue-800">
                {perm}
              </span>
            ))
          )}
        </div>
      </div>
    </div>
  );
}

function Detail({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-white rounded-lg border border-gray-200 p-4">
      <p className="text-sm text-gray-600">{label}</p>
      <p className="text-sm font-medium text-gray-900 mt-1 break-all">{value}</p>
    </div>
  );
}
