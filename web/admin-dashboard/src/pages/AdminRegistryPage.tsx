import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';

interface RegistryStats {
  total_functions?: number;
  active_functions?: number;
  flagged_functions?: number;
}

interface RegistryFunction {
  id: string;
  name: string;
  visibility?: string;
  status?: string;
  updated_at?: string;
}

export function AdminRegistryPage() {
  const { data: statsResponse, isLoading: loadingStats } = useQuery({
    queryKey: ['admin-registry-stats'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<RegistryStats>('/registry/stats');
      } catch {
        return { data: {}, success: false, timestamp: new Date().toISOString() };
      }
    },
  });

  const { data: functionsResponse, isLoading: loadingFunctions } = useQuery({
    queryKey: ['admin-registry-functions'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<RegistryFunction[]>('/registry/functions');
      } catch {
        return { data: [], success: false, timestamp: new Date().toISOString() };
      }
    },
  });

  if (loadingStats || loadingFunctions) {
    return <LoadingScreen />;
  }

  const stats = statsResponse?.data || {};
  const functions = functionsResponse?.data || [];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Registry</h1>
        <p className="mt-2 text-gray-600 dark:text-gray-400">Review and manage marketplace registry functions.</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <StatCard label="Total Functions" value={stats.total_functions ?? functions.length} />
        <StatCard label="Active" value={stats.active_functions ?? 0} />
        <StatCard label="Flagged" value={stats.flagged_functions ?? 0} />
      </div>

      <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Name</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Visibility</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Status</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Updated</th>
            </tr>
          </thead>
          <tbody>
            {functions.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-6 py-8 text-center text-gray-500 dark:text-gray-400">No registry functions found.</td>
              </tr>
            ) : (
              functions.map((fn) => (
                <tr key={fn.id} className="border-b border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800">
                  <td className="px-6 py-4 text-sm font-medium text-gray-900 dark:text-gray-100">{fn.name}</td>
                  <td className="px-6 py-4 text-sm text-gray-600 dark:text-gray-400">{fn.visibility || 'public'}</td>
                  <td className="px-6 py-4 text-sm text-gray-600 dark:text-gray-400">{fn.status || 'unknown'}</td>
                  <td className="px-6 py-4 text-sm text-gray-600 dark:text-gray-400">{fn.updated_at ? new Date(fn.updated_at).toLocaleString() : '-'}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function StatCard({ label, value }: { label: string; value: number }) {
  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4">
      <p className="text-sm text-gray-600 dark:text-gray-400">{label}</p>
      <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">{value}</p>
    </div>
  );
}
