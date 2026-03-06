import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';

export function AdminFraudDetectionPage() {
  const { data } = useQuery({
    queryKey: ['admin-oversight-fraud-detection'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<Record<string, unknown>>('/oversight/fraud-detection');
      } catch {
        return { data: {}, success: false, timestamp: new Date().toISOString() };
      }
    },
  });

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Fraud Detection</h1>
        <p className="mt-2 text-gray-600">Fraud risk insights and suspicious activity signals.</p>
      </div>
      <pre className="bg-white border border-gray-200 rounded-lg p-4 text-xs overflow-auto">
        {JSON.stringify(data?.data || {}, null, 2)}
      </pre>
    </div>
  );
}
