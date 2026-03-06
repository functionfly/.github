import { Link, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { ROUTES } from '@/lib/constants';
import type { AdminUser } from '@/types';

export function AdminUserDetailPage() {
  const { userId } = useParams<{ userId: string }>();

  const { data, isLoading } = useQuery({
    queryKey: ['admin-user-detail', userId],
    queryFn: async () => {
      if (!userId) {
        return { data: null, success: false, timestamp: new Date().toISOString() };
      }
      try {
        return await adminApiClient.get<AdminUser>(`/users/${userId}`);
      } catch {
        return { data: null, success: false, timestamp: new Date().toISOString() };
      }
    },
    enabled: Boolean(userId),
  });

  if (isLoading) return <LoadingScreen />;

  const user = data?.data;

  if (!user) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold text-gray-900">User not found</h1>
        <Link to={ROUTES.ADMIN_USERS} className="text-blue-600 hover:text-blue-700">Back to users</Link>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">{user.name}</h1>
          <p className="mt-2 text-gray-600">Admin user profile and permissions.</p>
        </div>
        <Link to={ROUTES.ADMIN_USERS} className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">
          Back
        </Link>
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
          {user.permissions.length === 0 ? (
            <span className="text-sm text-gray-500">No permissions</span>
          ) : (
            user.permissions.map((perm) => (
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
