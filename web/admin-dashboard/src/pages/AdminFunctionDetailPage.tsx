import { Link, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { ROUTES } from '@/lib/constants';

interface AdminFunction {
  id: string;
  name: string;
  status?: string;
  runtime?: string;
  created_at?: string;
  updated_at?: string;
}

export function AdminFunctionDetailPage() {
  const { functionId } = useParams<{ functionId: string }>();

  const { data: functionResponse, isLoading: loadingFn } = useQuery({
    queryKey: ['admin-function-detail', functionId],
    queryFn: async () => {
      if (!functionId) {
        return { data: null, success: false, timestamp: new Date().toISOString() };
      }
      try {
        return await adminApiClient.get<AdminFunction>(`/functions/${functionId}`);
      } catch {
        return { data: null, success: false, timestamp: new Date().toISOString() };
      }
    },
    enabled: Boolean(functionId),
  });

  const { data: metricsResponse, isLoading: loadingMetrics } = useQuery({
    queryKey: ['admin-function-metrics', functionId],
    queryFn: async () => {
      if (!functionId) {
        return { data: {}, success: false, timestamp: new Date().toISOString() };
      }
      try {
        return await adminApiClient.get<Record<string, unknown>>(`/functions/${functionId}/metrics`);
      } catch {
        return { data: {}, success: false, timestamp: new Date().toISOString() };
      }
    },
    enabled: Boolean(functionId),
  });

  if (loadingFn || loadingMetrics) return <LoadingScreen />;

  const fn = functionResponse?.data;

  if (!fn) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold text-gray-900">Function not found</h1>
        <Link to={ROUTES.ADMIN_FUNCTIONS} className="text-blue-600 hover:text-blue-700">Back to functions</Link>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">{fn.name}</h1>
          <p className="mt-2 text-gray-600">Function details, status, and metrics.</p>
        </div>
        <Link to={ROUTES.ADMIN_FUNCTIONS} className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">
          Back
        </Link>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Detail label="Function ID" value={fn.id} />
        <Detail label="Status" value={fn.status || 'unknown'} />
        <Detail label="Runtime" value={fn.runtime || '-'} />
        <Detail label="Created" value={fn.created_at ? new Date(fn.created_at).toLocaleString() : '-'} />
      </div>

      <div className="bg-white rounded-lg border border-gray-200 p-4">
        <p className="text-sm text-gray-600">Metrics snapshot</p>
        <pre className="mt-2 text-xs overflow-auto bg-gray-50 rounded p-3">{JSON.stringify(metricsResponse?.data || {}, null, 2)}</pre>
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
