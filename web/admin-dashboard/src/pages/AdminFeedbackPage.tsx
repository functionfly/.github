import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';

interface FeedbackStats {
  total?: number;
  open?: number;
  resolved?: number;
}

interface FeedbackItem {
  id: string;
  category?: string;
  status?: string;
  message?: string;
  created_at?: string;
}

export function AdminFeedbackPage() {
  const { data: statsResponse, isLoading: loadingStats } = useQuery({
    queryKey: ['admin-feedback-stats'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<FeedbackStats>('/feedback/stats');
      } catch {
        return { data: {}, success: false, timestamp: new Date().toISOString() };
      }
    },
  });

  const { data: feedbackResponse, isLoading: loadingFeedback } = useQuery({
    queryKey: ['admin-feedback-list'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<FeedbackItem[]>('/feedback');
      } catch {
        return { data: [], success: false, timestamp: new Date().toISOString() };
      }
    },
  });

  if (loadingStats || loadingFeedback) {
    return <LoadingScreen />;
  }

  const stats = statsResponse?.data || {};
  const items = feedbackResponse?.data || [];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Feedback</h1>
        <p className="mt-2 text-gray-600">Review product feedback and support signals.</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <StatCard label="Total" value={stats.total ?? items.length} />
        <StatCard label="Open" value={stats.open ?? items.filter((i) => i.status === 'open').length} />
        <StatCard label="Resolved" value={stats.resolved ?? items.filter((i) => i.status === 'resolved').length} />
      </div>

      <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="bg-gray-50 border-b border-gray-200">
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">ID</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Category</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Status</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Message</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Created</th>
            </tr>
          </thead>
          <tbody>
            {items.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-6 py-8 text-center text-gray-500">No feedback found.</td>
              </tr>
            ) : (
              items.map((item) => (
                <tr key={item.id} className="border-b border-gray-100 hover:bg-gray-50">
                  <td className="px-6 py-4 text-sm text-gray-900">{item.id}</td>
                  <td className="px-6 py-4 text-sm text-gray-600">{item.category || '-'}</td>
                  <td className="px-6 py-4 text-sm text-gray-600">{item.status || '-'}</td>
                  <td className="px-6 py-4 text-sm text-gray-600 truncate max-w-[420px]">{item.message || '-'}</td>
                  <td className="px-6 py-4 text-sm text-gray-600">{item.created_at ? new Date(item.created_at).toLocaleString() : '-'}</td>
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
    <div className="bg-white rounded-lg border border-gray-200 p-4">
      <p className="text-sm text-gray-600">{label}</p>
      <p className="text-2xl font-bold text-gray-900">{value}</p>
    </div>
  );
}
