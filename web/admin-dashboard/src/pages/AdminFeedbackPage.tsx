import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { toast } from 'sonner';
import { RefreshCw, Filter, Download, CheckCircle, Clock, AlertCircle, XCircle, X, MessageSquare, Mail, Globe, Monitor } from 'lucide-react';

interface FeedbackStats {
  total?: number;
  open?: number;
  in_review?: number;
  resolved?: number;
  closed?: number;
}

interface FeedbackItem {
  id: string;
  feedback_type: string;
  subject: string;
  message: string;
  priority: string;
  status: string;
  user_email: string | null;
  browser_info: string;
  created_at: string;
  updated_at: string;
}

const STATUS_CONFIG: Record<string, { label: string; color: string; icon: any }> = {
  submitted: { label: 'New', color: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300', icon: Clock },
  'in-review': { label: 'In Review', color: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300', icon: RefreshCw },
  resolved: { label: 'Resolved', color: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300', icon: CheckCircle },
  closed: { label: 'Closed', color: 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300', icon: XCircle },
};

const TYPE_COLORS: Record<string, string> = {
  bug: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
  feature: 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300',
  improvement: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300',
  general: 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300',
  launch_waitlist: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300',
};

const PRIORITY_COLORS: Record<string, string> = {
  low: 'text-gray-500',
  medium: 'text-yellow-600 dark:text-yellow-400',
  high: 'text-orange-600 dark:text-orange-400',
  critical: 'text-red-600 dark:text-red-400 font-bold',
};

export function AdminFeedbackPage() {
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [typeFilter, setTypeFilter] = useState<string>('');
  const [selectedFeedback, setSelectedFeedback] = useState<FeedbackItem | null>(null);
  const queryClient = useQueryClient();

  const { data: statsResponse, isLoading: loadingStats } = useQuery({
    queryKey: ['admin-feedback-stats'],
    queryFn: async () => {
      return await adminApiClient.get<FeedbackStats>('/feedback/stats');
    },
  });

  const { data: feedbackResponse, isLoading: loadingFeedback } = useQuery({
    queryKey: ['admin-feedback-list', statusFilter, typeFilter],
    queryFn: async () => {
      const params = new URLSearchParams();
      if (statusFilter) params.append('status', statusFilter);
      if (typeFilter) params.append('feedback_type', typeFilter);
      const query = params.toString() ? `?${params.toString()}` : '';
      return await adminApiClient.get<FeedbackItem[]>(`/feedback${query}`);
    },
  });

  const updateStatusMutation = useMutation({
    mutationFn: async ({ id, status }: { id: string; status: string }) => {
      return await adminApiClient.patch(`/feedback/${id}/status`, { status });
    },
    onSuccess: () => {
      toast.success('Status updated');
      queryClient.invalidateQueries({ queryKey: ['admin-feedback-list'] });
      queryClient.invalidateQueries({ queryKey: ['admin-feedback-stats'] });
    },
    onError: () => {
      toast.error('Failed to update status');
    },
  });

  const handleStatusChange = (id: string, newStatus: string) => {
    updateStatusMutation.mutate({ id, status: newStatus });
  };

  if (loadingStats || loadingFeedback) {
    return <LoadingScreen />;
  }

// Backend returns data directly without AdminAPIResponse wrapper
  const stats = statsResponse as { total?: number; status_breakdown?: Record<string, number> } || {};
  const statusBreakdown = stats.status_breakdown || {};

  const derivedStats = {
    total: stats.total ?? 0,
    open: statusBreakdown['submitted'] ?? 0,
    in_review: statusBreakdown['in-review'] ?? 0,
    resolved: statusBreakdown['resolved'] ?? 0,
    closed: statusBreakdown['closed'] ?? 0,
  };

  const feedbackRaw = feedbackResponse as { feedback: FeedbackItem[] } | undefined;
  const feedbackItems = feedbackRaw?.feedback || [];
  const items = feedbackItems;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Feedback</h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">Review product feedback and support signals.</p>
        </div>
        <div className="flex gap-2">
          <button className="flex items-center gap-2 px-4 py-2 text-sm bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700">
            <Download className="w-4 h-4" />
            Export
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
        <StatCard label="Total" value={derivedStats.total} color="text-gray-900 dark:text-gray-100" />
        <StatCard label="New" value={derivedStats.open} color="text-blue-600 dark:text-blue-400" />
        <StatCard label="In Review" value={derivedStats.in_review} color="text-yellow-600 dark:text-yellow-400" />
        <StatCard label="Resolved" value={derivedStats.resolved} color="text-green-600 dark:text-green-400" />
        <StatCard label="Closed" value={derivedStats.closed} color="text-gray-500" />
      </div>

      <div className="flex gap-4 items-center">
        <div className="flex items-center gap-2">
          <Filter className="w-4 h-4 text-gray-500" />
          <span className="text-sm text-gray-600 dark:text-gray-400">Filters:</span>
        </div>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="px-3 py-1.5 text-sm border border-gray-200 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
        >
          <option value="">All Statuses</option>
          <option value="submitted">New</option>
          <option value="in-review">In Review</option>
          <option value="resolved">Resolved</option>
          <option value="closed">Closed</option>
        </select>
        <select
          value={typeFilter}
          onChange={(e) => setTypeFilter(e.target.value)}
          className="px-3 py-1.5 text-sm border border-gray-200 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
        >
          <option value="">All Types</option>
          <option value="bug">Bug</option>
          <option value="feature">Feature Request</option>
          <option value="improvement">Improvement</option>
          <option value="general">General</option>
          <option value="launch_waitlist">Launch Waitlist</option>
        </select>
      </div>

      <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="bg-gray-50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-700">
              <th className="px-4 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Type</th>
              <th className="px-4 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Subject</th>
              <th className="px-4 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
              <th className="px-4 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Priority</th>
              <th className="px-4 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Email</th>
              <th className="px-4 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Date</th>
              <th className="px-4 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
            {items.length === 0 ? (
              <tr>
                <td colSpan={7} className="px-6 py-12 text-center text-gray-500 dark:text-gray-400">
                  <AlertCircle className="w-8 h-8 mx-auto mb-2 opacity-50" />
                  No feedback found matching your filters.
                </td>
              </tr>
            ) : (
              items.map((item) => {
                const statusConfig = STATUS_CONFIG[item.status] || STATUS_CONFIG.submitted;
                const StatusIcon = statusConfig.icon;
                return (
                  <tr
                    key={item.id}
                    className="hover:bg-gray-50 dark:hover:bg-gray-800/50 cursor-pointer"
                    onClick={() => setSelectedFeedback(item)}
                  >
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${TYPE_COLORS[item.feedback_type] || 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'}`}>
                        {item.feedback_type?.replace('_', ' ')}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <div className="max-w-[280px]">
                        <p className="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">{item.subject}</p>
                        <p className="text-xs text-gray-500 dark:text-gray-400 truncate mt-0.5">{item.message}</p>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium ${statusConfig.color}`}>
                        <StatusIcon className="w-3 h-3" />
                        {statusConfig.label}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`text-sm capitalize ${PRIORITY_COLORS[item.priority] || 'text-gray-500'}`}>
                        {item.priority || 'medium'}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">
                      {item.user_email || '-'}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-500 dark:text-gray-400 whitespace-nowrap">
                      {item.created_at ? new Date(item.created_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) : '-'}
                    </td>
                    <td className="px-4 py-3">
                      <select
                        value={item.status}
                        onChange={(e) => handleStatusChange(item.id, e.target.value)}
                        disabled={updateStatusMutation.isPending}
                        className="text-xs border border-gray-200 dark:border-gray-600 rounded px-2 py-1 bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-300"
                      >
                        <option value="submitted">New</option>
                        <option value="in-review">In Review</option>
                        <option value="resolved">Resolved</option>
                        <option value="closed">Closed</option>
                      </select>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      <div className="flex justify-between items-center text-sm text-gray-500 dark:text-gray-400">
        <span>Showing {items.length} feedback items</span>
        <span>Last updated: {new Date().toLocaleTimeString()}</span>
      </div>

      {selectedFeedback && (
        <div className="fixed inset-0 bg-black/50 dark:bg-black/70 flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-hidden">
            <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
              <div className="flex items-center gap-3">
                <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium ${TYPE_COLORS[selectedFeedback.feedback_type] || 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'}`}>
                  {selectedFeedback.feedback_type?.replace('_', ' ')}
                </span>
                <span className={`inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium ${STATUS_CONFIG[selectedFeedback.status]?.color || 'bg-gray-100 text-gray-800'}`}>
                  {STATUS_CONFIG[selectedFeedback.status]?.label || selectedFeedback.status}
                </span>
              </div>
              <button
                onClick={() => setSelectedFeedback(null)}
                className="p-1 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg"
              >
                <X className="w-5 h-5 text-gray-500" />
              </button>
            </div>

            <div className="p-6 overflow-y-auto max-h-[calc(90vh-8rem)]">
              <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-4">
                {selectedFeedback.subject}
              </h2>

              <div className="space-y-4">
                <div className="flex items-start gap-3">
                  <MessageSquare className="w-5 h-5 text-gray-400 mt-0.5" />
                  <div>
                    <p className="text-sm font-medium text-gray-700 dark:text-gray-300">Message</p>
                    <p className="text-gray-600 dark:text-gray-400 mt-1 whitespace-pre-wrap">{selectedFeedback.message}</p>
                  </div>
                </div>

                {selectedFeedback.user_email && (
                  <div className="flex items-start gap-3">
                    <Mail className="w-5 h-5 text-gray-400 mt-0.5" />
                    <div>
                      <p className="text-sm font-medium text-gray-700 dark:text-gray-300">Email</p>
                      <a href={`mailto:${selectedFeedback.user_email}`} className="text-blue-600 hover:underline">
                        {selectedFeedback.user_email}
                      </a>
                    </div>
                  </div>
                )}

                <div className="flex items-start gap-3">
                  <Monitor className="w-5 h-5 text-gray-400 mt-0.5" />
                  <div>
                    <p className="text-sm font-medium text-gray-700 dark:text-gray-300">Priority</p>
                    <p className={`capitalize ${PRIORITY_COLORS[selectedFeedback.priority] || 'text-gray-500'}`}>
                      {selectedFeedback.priority || 'medium'}
                    </p>
                  </div>
                </div>

                {selectedFeedback.browser_info && (
                  <div className="flex items-start gap-3">
                    <Globe className="w-5 h-5 text-gray-400 mt-0.5" />
                    <div>
                      <p className="text-sm font-medium text-gray-700 dark:text-gray-300">Browser Info</p>
                      <p className="text-gray-600 dark:text-gray-400 text-sm">{selectedFeedback.browser_info}</p>
                    </div>
                  </div>
                )}

                <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
                  <span>Submitted {new Date(selectedFeedback.created_at).toLocaleString()}</span>
                  {selectedFeedback.updated_at && selectedFeedback.updated_at !== selectedFeedback.created_at && (
                    <span>· Updated {new Date(selectedFeedback.updated_at).toLocaleString()}</span>
                  )}
                </div>
              </div>
            </div>

            <div className="flex items-center justify-end gap-3 p-4 border-t border-gray-200 dark:border-gray-700">
              <select
                value={selectedFeedback.status}
                onChange={(e) => {
                  handleStatusChange(selectedFeedback.id, e.target.value);
                  setSelectedFeedback({ ...selectedFeedback, status: e.target.value });
                }}
                disabled={updateStatusMutation.isPending}
                className="text-sm border border-gray-200 dark:border-gray-600 rounded-lg px-3 py-2 bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-300"
              >
                <option value="submitted">New</option>
                <option value="in-review">In Review</option>
                <option value="resolved">Resolved</option>
                <option value="closed">Closed</option>
              </select>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function StatCard({ label, value, color }: { label: string; value: number; color?: string }) {
  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4">
      <p className="text-sm text-gray-500 dark:text-gray-400">{label}</p>
      <p className={`text-2xl font-bold ${color || 'text-gray-900 dark:text-gray-100'}`}>{value}</p>
    </div>
  );
}
