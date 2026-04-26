/**
 * Factory Opportunity Slide-Over Component
 * Shows detailed opportunity information in a slide-over panel with expandable description
 */

import { useState } from 'react';
import { X, ExternalLink, Tag, Calendar, User, CheckCircle, XCircle, Clock, ChevronDown, ChevronUp } from 'lucide-react';
import type { Opportunity } from '@/lib/api/factory';
import clsx from 'clsx';

const DESCRIPTION_PREVIEW_LENGTH = 200;

interface FactoryOpportunitySlideOverProps {
  opportunity: Opportunity | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onApprove?: (id: string) => void;
  onReject?: (id: string) => void;
  isApproving?: boolean;
  isRejecting?: boolean;
}

export function FactoryOpportunitySlideOver({
  opportunity,
  open,
  onOpenChange,
  onApprove,
  onReject,
  isApproving = false,
  isRejecting = false,
}: FactoryOpportunitySlideOverProps) {
  const [descriptionExpanded, setDescriptionExpanded] = useState(false);

  if (!open || !opportunity) return null;

  const getStatusColor = (status?: string) => {
    switch (status?.toLowerCase()) {
      case 'approved':
      case 'published':
        return 'bg-green-100 dark:bg-green-900/50 text-green-800 dark:text-green-300';
      case 'rejected':
        return 'bg-red-100 dark:bg-red-900/50 text-red-800 dark:text-red-300';
      case 'pending_review':
        return 'bg-yellow-100 dark:bg-yellow-900/50 text-yellow-800 dark:text-yellow-300';
      case 'draft':
        return 'bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-300';
      default:
        return 'bg-blue-100 dark:bg-blue-900/50 text-blue-800 dark:text-blue-300';
    }
  };

  const getScoreColor = (score?: number) => {
    if (score === undefined) return 'text-gray-400 dark:text-gray-500';
    if (score >= 80) return 'text-green-600 dark:text-green-400';
    if (score >= 60) return 'text-yellow-600 dark:text-yellow-400';
    return 'text-red-600 dark:text-red-400';
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleString();
  };

  return (
    <div className="fixed inset-0 z-50 flex justify-end">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/30 dark:bg-black/50"
        onClick={() => onOpenChange(false)}
      />
      
      {/* Slide-over panel */}
      <div className="relative w-full max-w-xl bg-white dark:bg-gray-800 shadow-xl flex flex-col animate-in slide-in-from-right duration-300 border-l border-gray-200 dark:border-gray-700">
        {/* Header */}
        <div className="flex items-start justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <div className="flex-1 min-w-0 pr-4">
            <div className="flex items-center gap-2 mb-1">
              <span className={clsx(
                'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                getStatusColor(opportunity.status)
              )}>
                {opportunity.status || 'unknown'}
              </span>
              <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-200">
                {opportunity.source}
              </span>
            </div>
            <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100 truncate">
              {opportunity.title}
            </h2>
          </div>
          <button
            onClick={() => onOpenChange(false)}
            className="p-2 text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6">
          {/* Scores */}
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-4">
              <p className="text-sm text-gray-500 dark:text-gray-400 mb-1">Quality Score</p>
              <p className={clsx('text-2xl font-bold', getScoreColor(opportunity.quality_score))}>
                {opportunity.quality_score !== undefined ? `${opportunity.quality_score}%` : 'N/A'}
              </p>
            </div>
            <div className="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-4">
              <p className="text-sm text-gray-500 dark:text-gray-400 mb-1">Test Score</p>
              <p className={clsx('text-2xl font-bold', getScoreColor(opportunity.test_score))}>
                {opportunity.test_score !== undefined ? `${opportunity.test_score}%` : 'N/A'}
              </p>
            </div>
          </div>

          {/* Description */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400">Description</h3>
              {opportunity.description && opportunity.description.length > DESCRIPTION_PREVIEW_LENGTH && (
                <button
                  onClick={() => setDescriptionExpanded(!descriptionExpanded)}
                  className="flex items-center gap-1 text-sm text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300"
                >
                  {descriptionExpanded ? (
                    <>
                      <ChevronUp className="h-4 w-4" />
                      Show less
                    </>
                  ) : (
                    <>
                      <ChevronDown className="h-4 w-4" />
                      Show more
                    </>
                  )}
                </button>
              )}
            </div>
            <div className="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-4">
              <p className="text-gray-700 dark:text-gray-300 whitespace-pre-wrap text-sm leading-relaxed">
                {descriptionExpanded
                  ? opportunity.description
                  : opportunity.description?.slice(0, DESCRIPTION_PREVIEW_LENGTH) + (opportunity.description.length > DESCRIPTION_PREVIEW_LENGTH ? '...' : '')}
              </p>
              {!descriptionExpanded && opportunity.description && opportunity.description.length > DESCRIPTION_PREVIEW_LENGTH && (
                <button
                  onClick={() => setDescriptionExpanded(true)}
                  className="mt-2 text-sm text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300"
                >
                  Read full description
                </button>
              )}
            </div>
          </div>

          {/* Metadata */}
          <div className="grid grid-cols-2 gap-4">
            <div className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
              <Calendar className="h-4 w-4 text-gray-400 dark:text-gray-500" />
              <span>Created: {formatDate(opportunity.created_at)}</span>
            </div>
            {opportunity.reviewed_by && (
              <div className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                <User className="h-4 w-4 text-gray-400 dark:text-gray-500" />
                <span>Reviewed by: {opportunity.reviewed_by}</span>
              </div>
            )}
          </div>

          {/* Category & Tags */}
          {(opportunity.category || (opportunity.tags && opportunity.tags.length > 0)) && (
            <div>
              <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Category & Tags</h3>
              <div className="flex flex-wrap gap-2">
                {opportunity.category && (
                  <span className="inline-flex items-center gap-1 px-2.5 py-1 bg-purple-100 dark:bg-purple-900/50 text-purple-800 dark:text-purple-300 rounded-full text-xs font-medium">
                    <Tag className="h-3 w-3" />
                    {opportunity.category}
                  </span>
                )}
                {opportunity.tags?.map((tag, idx) => (
                  <span
                    key={idx}
                    className="inline-flex items-center px-2.5 py-1 bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-200 rounded-full text-xs"
                  >
                    {tag}
                  </span>
                ))}
              </div>
            </div>
          )}

          {/* Review Decision */}
          {opportunity.review_decision && (
            <div className="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-4">
              <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Review Decision</h3>
              <div className="flex items-center gap-2">
                {opportunity.review_decision === 'approved' ? (
                  <CheckCircle className="h-5 w-5 text-green-600 dark:text-green-400" />
                ) : (
                  <XCircle className="h-5 w-5 text-red-600 dark:text-red-400" />
                )}
                <span className={clsx(
                  'font-medium capitalize',
                  opportunity.review_decision === 'approved' ? 'text-green-700 dark:text-green-300' : 'text-red-700 dark:text-red-300'
                )}>
                  {opportunity.review_decision}
                </span>
                {opportunity.reviewed_at && (
                  <span className="text-gray-500 dark:text-gray-400 text-sm">
                    on {formatDate(opportunity.reviewed_at)}
                  </span>
                )}
              </div>
              {opportunity.review_reason && (
                <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
                  Reason: {opportunity.review_reason}
                </p>
              )}
            </div>
          )}

          {/* Additional Info */}
          <div className="grid grid-cols-2 gap-4 text-sm">
            {opportunity.confidence_score !== undefined && (
              <div>
                <p className="text-gray-500 dark:text-gray-400">Confidence Score</p>
                <p className="font-medium text-gray-900 dark:text-gray-100">{opportunity.confidence_score}%</p>
              </div>
            )}
            {opportunity.estimated_value !== undefined && (
              <div>
                <p className="text-gray-500 dark:text-gray-400">Estimated Value</p>
                <p className="font-medium text-gray-900 dark:text-gray-100">${opportunity.estimated_value.toFixed(2)}</p>
              </div>
            )}
            {opportunity.source_url && (
              <div className="col-span-2">
                <p className="text-gray-500 dark:text-gray-400">Source URL</p>
                <a
                  href={opportunity.source_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 flex items-center gap-1"
                >
                  {opportunity.source_url.length > 50 
                    ? `${opportunity.source_url.substring(0, 50)}...` 
                    : opportunity.source_url}
                  <ExternalLink className="h-3 w-3" />
                </a>
              </div>
            )}
          </div>

          {/* Retry Info */}
          {opportunity.retry_count !== undefined && opportunity.retry_count > 0 && (
            <div className="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-4">
              <div className="flex items-center gap-2 text-yellow-800 dark:text-yellow-300">
                <Clock className="h-4 w-4" />
                <span className="font-medium">Retry Attempt: {opportunity.retry_count}</span>
              </div>
              {opportunity.last_error && (
                <p className="mt-2 text-sm text-yellow-700 dark:text-yellow-400">
                  Last Error: {opportunity.last_error}
                </p>
              )}
            </div>
          )}
        </div>

        {/* Actions */}
        {opportunity.review_status === 'pending_review' && (
          <div className="p-6 border-t border-gray-200 dark:border-gray-700 flex justify-end gap-3">
            <button
              onClick={() => onReject?.(opportunity.id)}
              disabled={isRejecting || isApproving}
              className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors disabled:opacity-50 flex items-center gap-2 text-gray-700 dark:text-gray-300"
            >
              <XCircle className="h-4 w-4" />
              Reject
            </button>
            <button
              onClick={() => onApprove?.(opportunity.id)}
              disabled={isRejecting || isApproving}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 flex items-center gap-2"
            >
              <CheckCircle className="h-4 w-4" />
              {isApproving ? 'Approving...' : 'Approve'}
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
