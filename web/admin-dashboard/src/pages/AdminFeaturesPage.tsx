import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';

interface SecurityMeasure {
  id?: string;
  name?: string;
  enabled?: boolean;
  description?: string;
}

export function AdminFeaturesPage() {
  const { data } = useQuery({
    queryKey: ['admin-feature-measures'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<SecurityMeasure[]>('/security/measures');
      } catch {
        return { data: [], success: false, timestamp: new Date().toISOString() };
      }
    },
  });

  const measures = data?.data || [];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Features</h1>
        <p className="mt-2 text-gray-600">Platform security and operational feature controls.</p>
      </div>

      <div className="bg-white rounded-lg border border-gray-200 divide-y divide-gray-100">
        {measures.length === 0 ? (
          <div className="p-6 text-gray-500">No feature measures returned.</div>
        ) : (
          measures.map((measure, idx) => (
            <div key={measure.id || idx} className="p-6 flex items-center justify-between">
              <div>
                <h2 className="text-sm font-semibold text-gray-900">{measure.name || 'Unnamed measure'}</h2>
                <p className="text-sm text-gray-600 mt-1">{measure.description || 'No description provided.'}</p>
              </div>
              <span className={`text-xs px-2 py-1 rounded ${measure.enabled ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-700'}`}>
                {measure.enabled ? 'Enabled' : 'Disabled'}
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
