import { Link, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { ROUTES } from '@/lib/constants';
import type { Tenant } from '@/types';

export function AdminTenantDetailPage() {
  const { tenantId } = useParams<{ tenantId: string }>();

  const { data, isLoading } = useQuery({
    queryKey: ['admin-tenant-detail', tenantId],
    queryFn: async () => {
      if (!tenantId) {
        return { data: null, success: false, timestamp: new Date().toISOString() };
      }
      try {
        return await adminApiClient.get<Tenant>(`/tenants/${tenantId}`);
      } catch {
        return { data: null, success: false, timestamp: new Date().toISOString() };
      }
    },
    enabled: Boolean(tenantId),
  });

  if (isLoading) return <LoadingScreen />;

  const tenant = data?.data;

  if (!tenant) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold text-gray-900">Tenant not found</h1>
        <Link to={ROUTES.ADMIN_TENANTS} className="text-blue-600 hover:text-blue-700">Back to tenants</Link>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">{tenant.name}</h1>
          <p className="mt-2 text-gray-600">Tenant details and lifecycle metadata.</p>
        </div>
        <Link to={ROUTES.ADMIN_TENANTS} className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">
          Back
        </Link>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Detail label="Tenant ID" value={tenant.id} />
        <Detail label="Plan" value={tenant.plan} />
        <Detail label="Status" value={tenant.status} />
        <Detail label="Created" value={new Date(tenant.created_at).toLocaleString()} />
        <Detail label="Updated" value={new Date(tenant.updated_at).toLocaleString()} />
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
