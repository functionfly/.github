/**
 * Factory Reviews Tab Component
 * Displays pending reviews awaiting approval/rejection with bulk actions
 */

import { useState } from 'react';
import { RefreshCw, CheckCircle, XCircle, Square, CheckSquare, SquareCheck } from 'lucide-react';
import type { PendingReview } from '@/lib/api/factory';
import clsx from 'clsx';

function Skeleton({ className }: { className?: string }) {
  return (
    <div className={clsx('animate-pulse bg-gray-200 dark:bg-gray-700 rounded', className)} />
  );
}

interface FactoryReviewsTabProps {
  reviews: PendingReview[];
  total: number;
  isLoading: boolean;
  onRefresh: () => void;
  onApprove: (review: PendingReview) => void;
  onReject: (review: PendingReview) => void;
  onBulkApprove?: (ids: string[]) => void;
  onBulkReject?: (reviews: PendingReview[]) => void;
  isApproving: boolean;
}

export function FactoryReviewsTab({
  reviews,
  total,
  isLoading,
  onRefresh,
  onApprove,
  onReject,
  onBulkApprove,
  onBulkReject,
  isApproving,
}: FactoryReviewsTabProps) {
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());

  const toggleSelect = (id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const toggleSelectAll = () => {
    if (selectedIds.size === reviews.length) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(reviews.map((r) => r.id)));
    }
  };

  const handleBulkApprove = () => {
    if (onBulkApprove && selectedIds.size > 0) {
      onBulkApprove(Array.from(selectedIds));
      setSelectedIds(new Set());
    }
  };

  const handleBulkReject = () => {
    if (onBulkReject && selectedIds.size > 0) {
      const selected = reviews.filter((r) => selectedIds.has(r.id));
      onBulkReject(selected);
      setSelectedIds(new Set());
    }
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

  const getScoreColor = (score?: number) => {
    if (score === undefined) return 'text-gray-400 dark:text-gray-500';
    if (score >= 80) return 'text-green-600 dark:text-green-400';
    if (score >= 60) return 'text-yellow-600 dark:text-yellow-400';
    return 'text-red-600 dark:text-red-400';
  };

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg">
      <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex items-start justify-between gap-4">
        <div>
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Pending Reviews</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            Opportunities awaiting manual review before publishing
          </p>
        </div>
        <div className="flex items-center gap-3">
          {selectedIds.size > 0 && (
            <div className="flex items-center gap-2 px-3 py-1.5 bg-blue-50 dark:bg-blue-900/30 rounded-lg">
              <span className="text-sm text-blue-700 dark:text-blue-300 font-medium">
                {selectedIds.size} selected
              </span>
              {onBulkApprove && (
                <button
                  onClick={handleBulkApprove}
                  disabled={isApproving}
                  className="flex items-center gap-1 px-2 py-1 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 transition-colors"
                >
                  <CheckCircle className="h-3 w-3" />
                  Approve All
                </button>
              )}
              {onBulkReject && (
                <button
                  onClick={handleBulkReject}
                  disabled={isApproving}
                  className="flex items-center gap-1 px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 transition-colors"
                >
                  <XCircle className="h-3 w-3" />
                  Reject All
                </button>
              )}
            </div>
          )}
          <button
            onClick={onRefresh}
            disabled={isLoading}
            className="flex items-center gap-2 px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors disabled:opacity-50 shrink-0 text-gray-700 dark:text-gray-300"
          >
            <RefreshCw className={clsx('h-4 w-4', isLoading && 'animate-spin')} />
            Refresh
          </button>
        </div>
      </div>
      <div className="px-6 py-4">
        {isLoading ? (
          <div className="space-y-3">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="flex items-center gap-4 p-3 border-b border-gray-100 dark:border-gray-700">
                <Skeleton className="h-5 w-5" />
                <div className="flex-1">
                  <Skeleton className="h-4 w-56 mb-2" />
                  <Skeleton className="h-3 w-24" />
                </div>
                <Skeleton className="h-5 w-12" />
                <Skeleton className="h-5 w-12" />
                <Skeleton className="h-5 w-20" />
                <div className="flex gap-2">
                  <Skeleton className="h-8 w-20" />
                  <Skeleton className="h-8 w-20" />
                </div>
              </div>
            ))}
          </div>
        ) : reviews.length === 0 ? (
          <div className="text-center py-12">
            <CheckCircle className="h-12 w-12 text-green-500 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100">All Caught Up!</h3>
            <p className="text-gray-500 dark:text-gray-400">No pending reviews at the moment.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-gray-200 dark:border-gray-700">
                  <th className="py-3 px-4 text-left">
                    <button
                      onClick={toggleSelectAll}
                      className="p-1 hover:bg-gray-100 dark:hover:bg-gray-700 rounded transition-colors"
                      title={selectedIds.size === reviews.length ? 'Deselect all' : 'Select all'}
                    >
                      {selectedIds.size === reviews.length && reviews.length > 0 ? (
                        <CheckSquare className="h-4 w-4 text-blue-600 dark:text-blue-400" />
                      ) : (
                        <Square className="h-4 w-4 text-gray-400 dark:text-gray-500" />
                      )}
                    </button>
                  </th>
                  <th className="text-left py-3 px-4 text-sm font-medium text-gray-500 dark:text-gray-400">Title</th>
                  <th className="text-left py-3 px-4 text-sm font-medium text-gray-500 dark:text-gray-400">Source</th>
                  <th className="text-left py-3 px-4 text-sm font-medium text-gray-500 dark:text-gray-400">Quality</th>
                  <th className="text-left py-3 px-4 text-sm font-medium text-gray-500 dark:text-gray-400">Tests</th>
                  <th className="text-left py-3 px-4 text-sm font-medium text-gray-500 dark:text-gray-400">Submitted</th>
                  <th className="text-right py-3 px-4 text-sm font-medium text-gray-500 dark:text-gray-400">Actions</th>
                </tr>
              </thead>
              <tbody>
                {reviews.map((review) => (
                  <tr key={review.id} className="border-b border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700/50">
                    <td className="py-3 px-4">
                      <button
                        onClick={() => toggleSelect(review.id)}
                        className="p-1 hover:bg-gray-100 dark:hover:bg-gray-700 rounded transition-colors"
                      >
                        {selectedIds.has(review.id) ? (
                          <CheckSquare className="h-4 w-4 text-blue-600 dark:text-blue-400" />
                        ) : (
                          <Square className="h-4 w-4 text-gray-400 dark:text-gray-500" />
                        )}
                      </button>
                    </td>
                    <td className="py-3 px-4">
                      <p className="font-medium text-gray-900 dark:text-gray-100 max-w-xs truncate">{review.title}</p>
                    </td>
                    <td className="py-3 px-4">
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-200">
                        {review.source}
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      <span className={getScoreColor(review.quality_score)}>
                        {review.quality_score ? `${review.quality_score}%` : 'N/A'}
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      <span className={getScoreColor(review.test_score)}>
                        {review.test_score ? `${review.test_score}%` : 'N/A'}
                      </span>
                    </td>
                    <td className="py-3 px-4 text-gray-500 dark:text-gray-400">
                      {formatRelativeTime(review.review_requested_at || review.created_at)}
                    </td>
                    <td className="py-3 px-4 text-right">
                      <div className="flex items-center justify-end gap-2">
                        <button
                          onClick={() => onReject(review)}
                          className="flex items-center gap-1 px-3 py-1.5 text-sm border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors text-gray-700 dark:text-gray-300"
                        >
                          <XCircle className="h-4 w-4" />
                          Reject
                        </button>
                        <button
                          onClick={() => onApprove(review)}
                          disabled={isApproving}
                          className="flex items-center gap-1 px-3 py-1.5 text-sm bg-blue-600 dark:bg-blue-600 text-white rounded-lg hover:bg-blue-700 dark:hover:bg-blue-700 transition-colors disabled:opacity-50"
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
  );
}
