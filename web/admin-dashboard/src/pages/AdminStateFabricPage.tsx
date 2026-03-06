import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';

interface StateFabricItem {
  id: string;
  name?: string;
  tenant_id?: string;
  status?: 'running' | 'suspended' | string;
}

interface StateFabricStats {
  total?: number;
  active?: number;
  suspended?: number;
}

export function AdminStateFabricPage() {
  const queryClient = useQueryClient();

  const { data: statsResponse, isLoading: loadingStats } = useQuery({
    queryKey: ['admin-state-fabrics-stats'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<StateFabricStats>('/state-fabrics/stats');
      } catch {
        return { data: {}, success: false, timestamp: new Date().toISOString() };
      }
    },
  });

  const { data: listResponse, isLoading: loadingList } = useQuery({
    queryKey: ['admin-state-fabrics-list'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<StateFabricItem[]>('/state-fabrics');
      } catch {
        return { data: [], success: false, timestamp: new Date().toISOString() };
      }
    },
  });

  const suspendMutation = useMutation({
    mutationFn: (id: string) => adminApiClient.post(`/state-fabrics/${id}/suspend`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin-state-fabrics-list'] }),
  });

  const resumeMutation = useMutation({
    mutationFn: (id: string) => adminApiClient.post(`/state-fabrics/${id}/resume`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin-state-fabrics-list'] }),
  });

  if (loadingStats || loadingList) {
    return <LoadingScreen />;
  }

  const stats = statsResponse?.data || {};
  const fabrics = listResponse?.data || [];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">State Fabric</h1>
        <p className="mt-2 text-gray-600">Monitor and control state fabrics across tenants.</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <StatCard label="Total" value={stats.total ?? fabrics.length} />
        <StatCard label="Active" value={stats.active ?? fabrics.filter((f) => f.status === 'running').length} />
        <StatCard label="Suspended" value={stats.suspended ?? fabrics.filter((f) => f.status === 'suspended').length} />
      </div>

      <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="bg-gray-50 border-b border-gray-200">
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">ID</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Name</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Tenant</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Status</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Action</th>
            </tr>
          </thead>
          <tbody>
            {fabrics.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-6 py-8 text-center text-gray-500">No state fabrics found.</td>
              </tr>
            ) : (
              fabrics.map((fabric) => {
                const suspended = fabric.status === 'suspended';
                return (
                  <tr key={fabric.id} className="border-b border-gray-100 hover:bg-gray-50">
                    <td className="px-6 py-4 text-sm text-gray-900">{fabric.id}</td>
                    <td className="px-6 py-4 text-sm text-gray-600">{fabric.name || '-'}</td>
                    <td className="px-6 py-4 text-sm text-gray-600">{fabric.tenant_id || '-'}</td>
                    <td className="px-6 py-4 text-sm text-gray-600">{fabric.status || 'unknown'}</td>
                    <td className="px-6 py-4 text-sm">
                      <button
                        type="button"
                        disabled={suspendMutation.isPending || resumeMutation.isPending}
                        onClick={() => suspended ? resumeMutation.mutate(fabric.id) : suspendMutation.mutate(fabric.id)}
                        className="px-3 py-1 rounded bg-blue-100 text-blue-800 hover:bg-blue-200 disabled:opacity-50"
                      >
                        {suspended ? 'Resume' : 'Suspend'}
                      </button>
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

function StatCard({ label, value }: { label: string; value: number }) {
  return (
    <div className="bg-white rounded-lg border border-gray-200 p-4">
      <p className="text-sm text-gray-600">{label}</p>
      <p className="text-2xl font-bold text-gray-900">{value}</p>
    </div>
  );
}
