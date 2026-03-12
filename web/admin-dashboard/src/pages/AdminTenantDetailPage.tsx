import { useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { ROUTES } from '@/lib/constants';
import { Pencil } from 'lucide-react';
import type { Tenant } from '@/types';

const PLAN_OPTIONS = ['free', 'starter', 'pro', 'enterprise', 'agent_starter', 'agent_scale', 'agent_pro', 'agent_enterprise'] as const;
const STATUS_OPTIONS: Tenant['status'][] = ['active', 'suspended'];

export function AdminTenantDetailPage() {
  const { tenantId } = useParams<{ tenantId: string }>();
  const queryClient = useQueryClient();
  const [isEditOpen, setIsEditOpen] = useState(false);
  const [editName, setEditName] = useState('');
  const [editPlan, setEditPlan] = useState('');
  const [editStatus, setEditStatus] = useState<Tenant['status']>('active');

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

  const updateMutation = useMutation({
    mutationFn: (updates: Partial<Tenant>) =>
      adminApiClient.patch(`/tenants/${tenantId!}`, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-tenant-detail', tenantId] });
      queryClient.invalidateQueries({ queryKey: ['admin-tenants'] });
      setIsEditOpen(false);
    },
  });

  if (isLoading) return <LoadingScreen />;

  // API may return { data: tenant } (new) or the tenant object directly (legacy)
  const raw = data as { data?: Tenant } | Tenant | undefined;
  const tenant =
    raw && typeof raw === 'object' && 'data' in raw && raw.data != null
      ? raw.data
      : raw && typeof raw === 'object' && 'id' in raw && 'name' in raw
        ? (raw as Tenant)
        : undefined;

  if (!tenant) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold text-gray-900">Tenant not found</h1>
        <Link to={ROUTES.ADMIN_TENANTS} className="text-blue-600 hover:text-blue-700">Back to tenants</Link>
      </div>
    );
  }

  const openEdit = () => {
    setEditName(tenant.name);
    setEditPlan(tenant.plan || 'free');
    setEditStatus(tenant.status);
    setIsEditOpen(true);
  };

  const handleEditSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    updateMutation.mutate({
      name: editName.trim(),
      plan: editPlan,
      status: editStatus,
    });
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">{tenant.name}</h1>
          <p className="mt-2 text-gray-600">Tenant details and lifecycle metadata.</p>
        </div>
        <div className="flex gap-2">
          <button
            type="button"
            onClick={openEdit}
            className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
          >
            <Pencil className="w-4 h-4" />
            Edit
          </button>
          <Link to={ROUTES.ADMIN_TENANTS} className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">
            Back
          </Link>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Detail label="Tenant ID" value={tenant.id} />
        <Detail label="Plan" value={tenant.plan} />
        <Detail label="Status" value={tenant.status} />
        <Detail label="Created" value={new Date(tenant.created_at).toLocaleString()} />
        <Detail label="Updated" value={new Date(tenant.updated_at).toLocaleString()} />
      </div>

      {isEditOpen && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl p-6 max-w-md w-full">
            <h2 className="text-xl font-bold mb-4">Edit tenant</h2>
            <form onSubmit={handleEditSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
                <input
                  type="text"
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Plan</label>
                <select
                  value={editPlan}
                  onChange={(e) => setEditPlan(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                >
                  {PLAN_OPTIONS.map((p) => (
                    <option key={p} value={p}>{p}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Status</label>
                <select
                  value={editStatus}
                  onChange={(e) => setEditStatus(e.target.value as Tenant['status'])}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                >
                  {STATUS_OPTIONS.map((s) => (
                    <option key={s} value={s}>{s}</option>
                  ))}
                </select>
              </div>
              <div className="flex gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setIsEditOpen(false)}
                  className="flex-1 px-4 py-2 bg-gray-200 text-gray-800 rounded-lg hover:bg-gray-300"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={updateMutation.isPending || !editName.trim()}
                  className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
                >
                  {updateMutation.isPending ? 'Saving...' : 'Save'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
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
