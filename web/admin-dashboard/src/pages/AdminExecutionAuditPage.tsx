import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';

export function AdminExecutionAuditPage() {
  const { data } = useQuery({
    queryKey: ['admin-oversight-execution-audit'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<Record<string, unknown>>('/oversight/execution-audit');
      } catch {
        return { data: {}, success: false, timestamp: new Date().toISOString() };
      }
    },
  });

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Execution Audit</h1>
        <p className="mt-2 text-gray-600">Execution-level audit summaries and anomalies.</p>
      </div>
      <pre className="bg-white border border-gray-200 rounded-lg p-4 text-xs overflow-auto">
        {JSON.stringify(data?.data || {}, null, 2)}
      </pre>
    </div>
  );
}
