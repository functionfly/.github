/**
 * Factory Functions Tab Component
 * Browse and manage factory generated functions
 */

import { useState, Fragment } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Package, RefreshCw, ExternalLink, Code, ChevronDown, ChevronRight } from 'lucide-react';
import { factoryApi, type FactoryVersion, type PublishedFunction } from '@/lib/api/factory';
import clsx from 'clsx';

function Skeleton({ className }: { className?: string }) {
  return (
    <div className={clsx('animate-pulse bg-gray-200 dark:bg-gray-700 rounded', className)} />
  );
}

interface FactoryFunctionsTabProps {
  onViewFunction: (fn: PublishedFunction) => void;
}

export function FactoryFunctionsTab({ onViewFunction }: FactoryFunctionsTabProps) {
  const [expandedVersions, setExpandedVersions] = useState<Set<string>>(new Set());

  const {
    data: functionsData,
    isLoading,
    refetch,
  } = useQuery({
    queryKey: ['factory-functions'],
    queryFn: async () => {
      const data = await factoryApi.getFactoryFunctions({ limit: 50 });
      return data ?? { versions: [], total_versions: 0, limit: 50, offset: 0 };
    },
  });

  const publishedFunctions = functionsData?.published_functions ?? [];
  const versions = functionsData?.versions ?? [];
  const isEmpty = publishedFunctions.length === 0 && versions.length === 0;

  const toggleVersion = (id: string) => {
    setExpandedVersions((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleString();
  };

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg">
          <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
            <div>
              <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Factory Functions</h2>
              <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                Generated and published functions from the AI Factory
              </p>
            </div>
            <button
              onClick={() => refetch()}
              disabled={true}
              className="flex items-center gap-2 px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded-lg opacity-50 shrink-0 text-gray-700 dark:text-gray-300"
            >
              <RefreshCw className="h-4 w-4 animate-spin" />
              Refresh
            </button>
          </div>
          <div className="space-y-3 p-4">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="flex items-center gap-4 p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
                <Skeleton className="h-5 w-5" />
                <div className="flex-1">
                  <Skeleton className="h-4 w-48 mb-2" />
                  <Skeleton className="h-3 w-32" />
                </div>
                <Skeleton className="h-5 w-20" />
                <Skeleton className="h-5 w-16" />
              </div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (isEmpty) {
    return (
      <div className="space-y-4">
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg">
          <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
            <div>
              <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Factory Functions</h2>
              <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                Generated and published functions from the AI Factory
              </p>
            </div>
            <button
              onClick={() => refetch()}
              disabled={false}
              className="flex items-center gap-2 px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors shrink-0 text-gray-700 dark:text-gray-300"
            >
              <RefreshCw className="h-4 w-4" />
              Refresh
            </button>
          </div>
          <div className="text-center py-12">
            <Package className="h-12 w-12 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100">No functions found</h3>
            <p className="text-gray-500 dark:text-gray-400">
              Factory generated functions will appear here once created.
            </p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg">
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Factory Functions</h2>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
              Generated and published functions from the AI Factory
            </p>
          </div>
          <button
            onClick={() => refetch()}
            className="flex items-center gap-2 px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors shrink-0 text-gray-700 dark:text-gray-300"
          >
            <RefreshCw className={clsx('h-4 w-4', isLoading && 'animate-spin')} />
            Refresh
          </button>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50">
                <th className="text-left py-3 px-4 text-sm font-medium text-gray-500 dark:text-gray-400">Function ID</th>
                <th className="text-left py-3 px-4 text-sm font-medium text-gray-500 dark:text-gray-400">Model</th>
                <th className="text-left py-3 px-4 text-sm font-medium text-gray-500 dark:text-gray-400">Quality</th>
                <th className="text-left py-3 px-4 text-sm font-medium text-gray-500 dark:text-gray-400">Tests</th>
                <th className="text-left py-3 px-4 text-sm font-medium text-gray-500 dark:text-gray-400">Review</th>
                <th className="text-left py-3 px-4 text-sm font-medium text-gray-500 dark:text-gray-400">Created</th>
                <th className="text-right py-3 px-4 text-sm font-medium text-gray-500 dark:text-gray-400">Actions</th>
              </tr>
            </thead>
            <tbody>
              {versions.map((fn: FactoryVersion) => {
                const isExpanded = expandedVersions.has(fn.id);
                return (
                  <Fragment key={fn.id}>
                    <tr className="border-b border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700/50">
                      <td className="py-3 px-4">
                        <div className="flex items-center gap-2">
                          <button onClick={() => toggleVersion(fn.id)} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
                            {isExpanded ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
                          </button>
                          <div>
                            <p className="font-mono text-xs font-medium text-gray-900 dark:text-gray-100">{fn.function_id}</p>
                            <p className="text-xs text-gray-500 dark:text-gray-400">{fn.opportunity_id.slice(0, 8)}...</p>
                          </div>
                        </div>
                      </td>
                      <td className="py-3 px-4">
                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 dark:bg-purple-900/30 text-purple-800 dark:text-purple-300">
                          {fn.model_used || 'N/A'}
                        </span>
                      </td>
                      <td className="py-3 px-4">
                        <span className={clsx('text-sm font-medium', fn.quality_score >= 80 ? 'text-green-600 dark:text-green-400' : fn.quality_score >= 50 ? 'text-yellow-600 dark:text-yellow-400' : 'text-red-600 dark:text-red-400')}>
                          {fn.quality_score.toFixed(0)}%
                        </span>
                      </td>
                      <td className="py-3 px-4">
                        <span className={clsx('text-sm font-medium', fn.test_score >= 80 ? 'text-green-600 dark:text-green-400' : fn.test_score >= 50 ? 'text-yellow-600 dark:text-yellow-400' : 'text-red-600 dark:text-red-400')}>
                          {fn.test_score.toFixed(0)}%
                        </span>
                      </td>
                      <td className="py-3 px-4">
                        {fn.review_required ? (
                          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-amber-100 dark:bg-amber-900/30 text-amber-800 dark:text-amber-300">
                            Pending
                          </span>
                        ) : (
                          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-300">
                            Auto
                          </span>
                        )}
                      </td>
                      <td className="py-3 px-4 text-gray-500 dark:text-gray-400 text-sm">
                        {formatDate(fn.created_at)}
                      </td>
                      <td className="py-3 px-4 text-right">
                        <button
                          onClick={() => onViewFunction({ id: fn.function_id, author: 'factory', name: fn.function_id, version: '1.0', runtime: 'python', created_at: fn.created_at, status: 'published' })}
                          className="inline-flex items-center gap-1 px-3 py-1.5 text-sm text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/30 rounded-lg transition-colors"
                        >
                          <ExternalLink className="h-4 w-4" />
                          Registry
                        </button>
                      </td>
                    </tr>
                    {isExpanded && (
                      <tr className="border-b border-gray-100 dark:border-gray-700">
                        <td colSpan={7} className="px-4 py-3 bg-gray-50 dark:bg-gray-900">
                          <div className="space-y-2">
                            <div className="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
                              <Code className="h-3 w-3" />
                              Generated Source Code
                            </div>
                            <pre className="text-xs font-mono bg-gray-900 dark:bg-black text-gray-100 rounded p-3 overflow-auto max-h-64 whitespace-pre-wrap">{fn.generated_code || 'No source code available'}</pre>
                          </div>
                        </td>
                      </tr>
                    )}
                  </Fragment>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
