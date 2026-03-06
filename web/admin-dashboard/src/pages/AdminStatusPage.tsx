import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';

interface HealthCheckResponse {
  status?: string;
  version?: string;
  timestamp?: string;
  checks?: Record<string, { healthy?: boolean; status?: string }>;
}

export function AdminStatusPage() {
  const { data, isLoading } = useQuery({
    queryKey: ['admin-health-status'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<HealthCheckResponse>('/health');
      } catch {
        return { data: {}, success: false, timestamp: new Date().toISOString() };
      }
    },
    refetchInterval: 30000,
  });

  if (isLoading) {
    return <LoadingScreen />;
  }

  const health = data?.data || {};
  const checks = health.checks || {};

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Platform Status</h1>
        <p className="mt-2 text-gray-600">Live system health and subsystem checks.</p>
      </div>

      <div className="bg-white rounded-lg border border-gray-200 p-6">
        <div className="flex items-center justify-between">
          <p className="text-sm text-gray-600">Overall status</p>
          <span className={`text-xs px-2 py-1 rounded ${health.status === 'healthy' ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'}`}>
            {health.status || 'unknown'}
          </span>
        </div>
        <p className="mt-2 text-sm text-gray-600">Version: {health.version || '-'}</p>
      </div>

      <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="bg-gray-50 border-b border-gray-200">
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Component</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Status</th>
            </tr>
          </thead>
          <tbody>
            {Object.keys(checks).length === 0 ? (
              <tr>
                <td colSpan={2} className="px-6 py-8 text-center text-gray-500">No health checks available.</td>
              </tr>
            ) : (
              Object.entries(checks).map(([name, check]) => (
                <tr key={name} className="border-b border-gray-100">
                  <td className="px-6 py-4 text-sm text-gray-900">{name}</td>
                  <td className="px-6 py-4 text-sm">
                    <span className={`px-2 py-1 rounded text-xs ${check.healthy ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
                      {check.status || (check.healthy ? 'healthy' : 'unhealthy')}
                    </span>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
