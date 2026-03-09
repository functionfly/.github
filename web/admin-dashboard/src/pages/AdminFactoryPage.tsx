/**
 * Admin Factory Page
 * AI Function Factory status monitoring and review queue management
 */

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Activity,
  Play,
  RefreshCw,
  CheckCircle,
  XCircle,
  Clock,
  Package,
  Search,
  Zap,
  AlertTriangle,
} from 'lucide-react';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import {
  factoryApi,
  type FactoryStatus,
  type PendingReview,
  type FactoryRun,
} from '@/lib/api/factory';
import clsx from 'clsx';

/**
 * AdminFactoryPage - Factory status monitoring and review queue management
 */
export function AdminFactoryPage() {
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState<'status' | 'reviews'>('status');
  const [selectedReview, setSelectedReview] = useState<PendingReview | null>(null);
  const [rejectReason, setRejectReason] = useState('');
  const [showRejectDialog, setShowRejectDialog] = useState(false);

  // Fetch factory status (queryFn must not return undefined for React Query v5)
  const {
    data: factoryStatus,
    isLoading: statusLoading,
    error: statusError,
    refetch: refetchStatus,
  } = useQuery({
    queryKey: ['factory-status'],
    queryFn: async (): Promise<FactoryStatus | null> => {
      const data = await factoryApi.getStatus();
      return data ?? null;
    },
    refetchInterval: 30000,
  });

  // Fetch pending reviews (queryFn must not return undefined for React Query v5)
  const {
    data: pendingReviews,
    isLoading: reviewsLoading,
    refetch: refetchReviews,
  } = useQuery({
    queryKey: ['factory-pending-reviews'],
    queryFn: async () => {
      const data = await factoryApi.listPendingReviews({ limit: 50 });
      return data ?? { reviews: [], total: 0, limit: 50, offset: 0 };
    },
    refetchInterval: 15000,
  });

  // Pipeline run mutation
  const runPipelineMutation = useMutation({
    mutationFn: factoryApi.triggerPipelineRun,
    onSuccess: (data) => {
      alert(`Pipeline run initiated: ${data.run?.id || 'Started successfully'}`);
      refetchStatus();
    },
    onError: (error: Error) => {
      alert(`Failed to start pipeline: ${error.message}`);
    },
  });

  // Approve opportunity mutation
  const approveMutation = useMutation({
    mutationFn: ({ id }: { id: string }) => factoryApi.approveOpportunity(id),
    onSuccess: () => {
      alert('Opportunity approved');
      queryClient.invalidateQueries({ queryKey: ['factory-pending-reviews'] });
      queryClient.invalidateQueries({ queryKey: ['factory-status'] });
      setSelectedReview(null);
    },
    onError: (error: Error) => {
      alert(`Failed to approve: ${error.message}`);
    },
  });

  // Reject opportunity mutation
  const rejectMutation = useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) =>
      factoryApi.rejectOpportunity(id, reason),
    onSuccess: () => {
      alert('Opportunity rejected');
      queryClient.invalidateQueries({ queryKey: ['factory-pending-reviews'] });
      queryClient.invalidateQueries({ queryKey: ['factory-status'] });
      setShowRejectDialog(false);
      setSelectedReview(null);
      setRejectReason('');
    },
    onError: (error: Error) => {
      alert(`Failed to reject: ${error.message}`);
    },
  });

  const handleRunPipeline = () => {
    runPipelineMutation.mutate();
  };

  const handleApprove = (review: PendingReview) => {
    approveMutation.mutate({ id: review.id });
  };

  const handleRejectClick = (review: PendingReview) => {
    setSelectedReview(review);
    setShowRejectDialog(true);
  };

  const handleConfirmReject = () => {
    if (!selectedReview || !rejectReason.trim()) return;
    rejectMutation.mutate({ id: selectedReview.id, reason: rejectReason });
  };

  const getStatusColor = (status?: string) => {
    switch (status) {
      case 'healthy':
      case 'completed':
      case 'approved':
        return 'bg-green-100 text-green-800';
      case 'running':
        return 'bg-blue-100 text-blue-800';
      case 'failed':
      case 'rejected':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-yellow-100 text-yellow-800';
    }
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleString();
  };

  const formatRelativeTime = (dateString?: string) => {
    if (!dateString) return 'N/A';
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    return `${diffDays}d ago`;
  };

  const calculateSuccessRate = (run?: FactoryRun) => {
    if (!run || run.functions_created === 0) return 0;
    const total = run.functions_created + run.functions_failed;
    if (total === 0) return 0;
    return Math.round((run.functions_created / total) * 100);
  };

  if (statusLoading) {
    return <LoadingScreen />;
  }

  return (
    <div className="p-6 max-w-7xl mx-auto">
      {/* Page Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">AI Function Factory</h1>
          <p className="text-gray-600 mt-1">
            Monitor factory status and manage opportunity reviews
          </p>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={() => refetchStatus()}
            disabled={statusLoading}
            className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors disabled:opacity-50"
          >
            <RefreshCw className={clsx('h-4 w-4', statusLoading && 'animate-spin')} />
            Refresh
          </button>
          <button
            onClick={handleRunPipeline}
            disabled={runPipelineMutation.isPending}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50"
          >
            <Play className={clsx('h-4 w-4', runPipelineMutation.isPending && 'animate-spin')} />
            {runPipelineMutation.isPending ? 'Starting...' : 'Run Pipeline'}
          </button>
        </div>
      </div>

      {/* Error Alert */}
      {statusError && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
          <div className="flex items-center gap-2 text-red-800">
            <AlertTriangle className="h-5 w-5" />
            <span>Failed to load factory status: {(statusError as Error).message}</span>
          </div>
        </div>
      )}

      {/* Tabs */}
      <div className="border-b border-gray-200 mb-6">
        <nav className="-mb-px flex space-x-8">
          <button
            onClick={() => setActiveTab('status')}
            className={clsx(
              'py-4 px-1 border-b-2 font-medium text-sm transition-colors',
              activeTab === 'status'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            )}
          >
            Status
          </button>
          <button
            onClick={() => setActiveTab('reviews')}
            className={clsx(
              'py-4 px-1 border-b-2 font-medium text-sm transition-colors',
              activeTab === 'reviews'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            )}
          >
            Reviews
            {pendingReviews?.total ? (
              <span className="ml-2 px-2 py-0.5 bg-red-100 text-red-600 text-xs rounded-full">
                {pendingReviews.total}
              </span>
            ) : null}
          </button>
        </nav>
      </div>

      {/* Status Tab */}
      {activeTab === 'status' && (
        <div className="space-y-6">
          {factoryStatus ? (
            <>
              {/* Stats Cards */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <StatCard
                  title="Total Runs"
                  value={factoryStatus.totals.runs}
                  icon={Activity}
                  description="Pipeline executions"
                />
                <StatCard
                  title="Published Functions"
                  value={factoryStatus.totals.published}
                  icon={Package}
                  description="Functions in registry"
                />
                <StatCard
                  title="Opportunities"
                  value={factoryStatus.totals.opportunities}
                  icon={Search}
                  description="Discovered opportunities"
                />
                <StatCard
                  title="Auto-Publish"
                  value={factoryStatus.totals.autopublish ? 'Enabled' : 'Disabled'}
                  icon={Zap}
                  description={`Quality: ${factoryStatus.totals.quality_minimum}% | Tests: ${factoryStatus.totals.test_minimum}%`}
                  status={factoryStatus.totals.autopublish ? 'success' : 'warning'}
                />
              </div>

              {/* Latest Run Card */}
              {factoryStatus.latest_run && (
                <div className="bg-white border border-gray-200 rounded-lg">
                  <div className="px-6 py-4 border-b border-gray-200">
                    <div className="flex items-center gap-2">
                      <Clock className="h-5 w-5 text-gray-400" />
                      <h2 className="text-lg font-semibold text-gray-900">Latest Pipeline Run</h2>
                    </div>
                    <p className="text-sm text-gray-500 mt-1">ID: {factoryStatus.latest_run.id}</p>
                  </div>
                  <div className="px-6 py-4">
                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
                      <div className="space-y-1">
                        <p className="text-sm text-gray-500">Status</p>
                        <span
                          className={clsx(
                            'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                            getStatusColor(factoryStatus.latest_run.status)
                          )}
                        >
                          {factoryStatus.latest_run.status}
                        </span>
                      </div>
                      <div className="space-y-1">
                        <p className="text-sm text-gray-500">Started</p>
                        <p className="font-medium text-gray-900">{formatDate(factoryStatus.latest_run.started_at)}</p>
                      </div>
                      <div className="space-y-1">
                        <p className="text-sm text-gray-500">Completed</p>
                        <p className="font-medium text-gray-900">
                          {factoryStatus.latest_run.completed_at
                            ? formatDate(factoryStatus.latest_run.completed_at)
                            : 'In Progress'}
                        </p>
                      </div>
                      <div className="space-y-1">
                        <p className="text-sm text-gray-500">Success Rate</p>
                        <p
                          className={clsx(
                            'font-medium',
                            calculateSuccessRate(factoryStatus.latest_run) >= 80
                              ? 'text-green-600'
                              : calculateSuccessRate(factoryStatus.latest_run) >= 50
                              ? 'text-yellow-600'
                              : 'text-red-600'
                          )}
                        >
                          {calculateSuccessRate(factoryStatus.latest_run)}%
                        </p>
                      </div>
                    </div>

                    <div className="grid grid-cols-2 md:grid-cols-5 gap-4 pt-4 border-t">
                      <div className="text-center">
                        <p className="text-2xl font-bold">{factoryStatus.latest_run.opportunities_found}</p>
                        <p className="text-xs text-gray-500">Found</p>
                      </div>
                      <div className="text-center">
                        <p className="text-2xl font-bold text-green-600">
                          {factoryStatus.latest_run.opportunities_approved}
                        </p>
                        <p className="text-xs text-gray-500">Approved</p>
                      </div>
                      <div className="text-center">
                        <p className="text-2xl font-bold text-red-600">
                          {factoryStatus.latest_run.opportunities_rejected}
                        </p>
                        <p className="text-xs text-gray-500">Rejected</p>
                      </div>
                      <div className="text-center">
                        <p className="text-2xl font-bold text-blue-600">
                          {factoryStatus.latest_run.functions_created}
                        </p>
                        <p className="text-xs text-gray-500">Created</p>
                      </div>
                      <div className="text-center">
                        <p className="text-2xl font-bold text-red-600">
                          {factoryStatus.latest_run.functions_failed}
                        </p>
                        <p className="text-xs text-gray-500">Failed</p>
                      </div>
                    </div>

                    {factoryStatus.latest_run.error && (
                      <div className="mt-4 p-4 bg-red-50 border border-red-200 rounded-lg">
                        <div className="flex items-center gap-2 text-red-800">
                          <AlertTriangle className="h-4 w-4" />
                          <span>{factoryStatus.latest_run.error}</span>
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              )}

              {/* Config Card */}
              <div className="bg-white border border-gray-200 rounded-lg">
                <div className="px-6 py-4 border-b border-gray-200">
                  <h2 className="text-lg font-semibold text-gray-900">Configuration</h2>
                  <p className="text-sm text-gray-500 mt-1">Current factory settings</p>
                </div>
                <div className="px-6 py-4">
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                    <div className="space-y-1">
                      <p className="text-sm text-gray-500">Agent ID</p>
                      <p className="font-mono text-sm text-gray-900">{factoryStatus.config.agent_id}</p>
                    </div>
                    <div className="space-y-1">
                      <p className="text-sm text-gray-500">Discovery Batch Size</p>
                      <p className="font-medium text-gray-900">{factoryStatus.config.discovery_batch_size}</p>
                    </div>
                    <div className="space-y-1">
                      <p className="text-sm text-gray-500">Min Quality Score</p>
                      <p className="font-medium text-gray-900">{factoryStatus.config.minimum_quality_score}%</p>
                    </div>
                    <div className="space-y-1">
                      <p className="text-sm text-gray-500">Min Test Score</p>
                      <p className="font-medium text-gray-900">{factoryStatus.config.minimum_test_score}%</p>
                    </div>
                  </div>
                </div>
              </div>
            </>
          ) : (
            <div className="text-center py-12 text-gray-500">No factory status available</div>
          )}
        </div>
      )}

      {/* Reviews Tab */}
      {activeTab === 'reviews' && (
        <div className="bg-white border border-gray-200 rounded-lg">
          <div className="px-6 py-4 border-b border-gray-200 flex items-start justify-between gap-4">
            <div>
              <h2 className="text-lg font-semibold text-gray-900">Pending Reviews</h2>
              <p className="text-sm text-gray-500 mt-1">
                Opportunities awaiting manual review before publishing
              </p>
            </div>
            <button
              onClick={() => refetchReviews()}
              disabled={reviewsLoading}
              className="flex items-center gap-2 px-3 py-1.5 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors disabled:opacity-50 shrink-0"
            >
              <RefreshCw className={clsx('h-4 w-4', reviewsLoading && 'animate-spin')} />
              Refresh
            </button>
          </div>
          <div className="px-6 py-4">
            {reviewsLoading ? (
              <div className="text-center py-8 text-gray-500">Loading reviews...</div>
            ) : pendingReviews?.reviews.length === 0 ? (
              <div className="text-center py-12">
                <CheckCircle className="h-12 w-12 text-green-500 mx-auto mb-4" />
                <h3 className="text-lg font-medium text-gray-900">All Caught Up!</h3>
                <p className="text-gray-500">No pending reviews at the moment.</p>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-gray-200">
                      <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">Title</th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">Source</th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">Quality</th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">Tests</th>
                      <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">Submitted</th>
                      <th className="text-right py-3 px-4 text-sm font-medium text-gray-500">Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {pendingReviews?.reviews.map((review) => (
                      <tr key={review.id} className="border-b border-gray-100 hover:bg-gray-50">
                        <td className="py-3 px-4">
                          <p className="font-medium text-gray-900 max-w-xs truncate">{review.title}</p>
                        </td>
                        <td className="py-3 px-4">
                          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                            {review.source}
                          </span>
                        </td>
                        <td className="py-3 px-4">
                          <span
                            className={clsx(
                              review.quality_score
                                ? review.quality_score >= 80
                                  ? 'text-green-600'
                                  : review.quality_score >= 60
                                  ? 'text-yellow-600'
                                  : 'text-red-600'
                                : 'text-gray-400'
                            )}
                          >
                            {review.quality_score ? `${review.quality_score}%` : 'N/A'}
                          </span>
                        </td>
                        <td className="py-3 px-4">
                          <span
                            className={clsx(
                              review.test_score
                                ? review.test_score >= 80
                                  ? 'text-green-600'
                                  : review.test_score >= 60
                                  ? 'text-yellow-600'
                                  : 'text-red-600'
                                : 'text-gray-400'
                            )}
                          >
                            {review.test_score ? `${review.test_score}%` : 'N/A'}
                          </span>
                        </td>
                        <td className="py-3 px-4 text-gray-500">
                          {formatRelativeTime(review.review_requested_at || review.created_at)}
                        </td>
                        <td className="py-3 px-4 text-right">
                          <div className="flex items-center justify-end gap-2">
                            <button
                              onClick={() => handleRejectClick(review)}
                              className="flex items-center gap-1 px-3 py-1.5 text-sm border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
                            >
                              <XCircle className="h-4 w-4" />
                              Reject
                            </button>
                            <button
                              onClick={() => handleApprove(review)}
                              disabled={approveMutation.isPending}
                              className="flex items-center gap-1 px-3 py-1.5 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50"
                            >
                              <CheckCircle className="h-4 w-4" />
                              Approve
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Reject Dialog */}
      {showRejectDialog && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
            <div className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-semibold text-gray-900">Reject Opportunity</h3>
              <p className="text-sm text-gray-500 mt-1">
                Please provide a reason for rejecting this opportunity.
              </p>
            </div>
            <div className="px-6 py-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Rejection Reason
              </label>
              <textarea
                value={rejectReason}
                onChange={(e) => setRejectReason(e.target.value)}
                placeholder="Enter the reason for rejection..."
                rows={4}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>
            <div className="px-6 py-4 border-t border-gray-200 flex justify-end gap-3">
              <button
                onClick={() => setShowRejectDialog(false)}
                className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleConfirmReject}
                disabled={!rejectReason.trim() || rejectMutation.isPending}
                className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors disabled:opacity-50"
              >
                {rejectMutation.isPending ? 'Rejecting...' : 'Reject Opportunity'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

/**
 * StatCard - Display a single statistic
 */
interface StatCardProps {
  title: string;
  value: string | number;
  icon: React.ComponentType<{ className?: string }>;
  description?: string;
  status?: 'success' | 'warning' | 'error';
}

function StatCard({ title, value, icon: Icon, description, status }: StatCardProps) {
  return (
    <div className="bg-white border border-gray-200 rounded-lg p-5">
      <div className="flex items-start justify-between">
        <div className="space-y-1">
          <p className="text-sm text-gray-500">{title}</p>
          <p
            className={clsx(
              'text-2xl font-bold',
              status === 'success' && 'text-green-600',
              status === 'warning' && 'text-yellow-600',
              status === 'error' && 'text-red-600',
              !status && 'text-gray-900'
            )}
          >
            {value}
          </p>
          {description && <p className="text-xs text-gray-500">{description}</p>}
        </div>
        <div
          className={clsx(
            'p-2 rounded-lg',
            status === 'success' && 'bg-green-100 text-green-600',
            status === 'warning' && 'bg-yellow-100 text-yellow-600',
            status === 'error' && 'bg-red-100 text-red-600',
            !status && 'bg-blue-100 text-blue-600'
          )}
        >
          <Icon className="h-5 w-5" />
        </div>
      </div>
    </div>
  );
}

export default AdminFactoryPage;
