import { useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { Button } from '@/components/ui/Button';
import { ROUTES } from '@/lib/constants';
import { AlertTriangle, ArrowLeft, ChevronDown, ChevronRight, Code } from 'lucide-react';

interface RegistryFunctionDetail {
  id: string;
  author: string;
  name: string;
  title: string;
  description: string;
  category: string;
  visibility: string;
  price_per_call: number;
  latest_version: string;
  created_at: string;
  updated_at: string;
  [key: string]: unknown;
}

interface FunctionVersion {
  id: string;
  function_id: string;
  version: string;
  manifest: unknown;
  runtime: string;
  timeout_ms: number;
  memory_mb: number;
  deterministic: boolean;
  cache_ttl: number;
  capabilities: string[];
  side_effects: boolean;
  idempotent: boolean;
  deployment_id: string;
  backend_id: string;
  content_hash: string;
  source_hash: string;
  source_code: string;
  bundle_size: number;
  published_at: string;
  updated_at: string;
  is_active: boolean;
}

interface FunctionDetailResponse {
  function: RegistryFunctionDetail;
  versions: FunctionVersion[];
}

export function AdminFunctionDetailPage() {
  const { functionId } = useParams<{ functionId: string }>();
  const [expandedVersions, setExpandedVersions] = useState<Set<string>>(new Set());

  const { data: functionData, isLoading: loadingFn } = useQuery<FunctionDetailResponse | null>({
    queryKey: ['admin-function-detail', functionId],
    queryFn: async () => {
      if (!functionId) {
        return null;
      }
      const resp = await adminApiClient.get<FunctionDetailResponse>(`/registry/functions/${functionId}`);
      const inner = (resp as { data?: FunctionDetailResponse }).data;
      return inner ?? null;
    },
    enabled: Boolean(functionId),
  });

  const { data: metricsResponse, isLoading: loadingMetrics } = useQuery<Record<string, unknown>>({
    queryKey: ['admin-function-metrics', functionId],
    queryFn: async () => {
      if (!functionId) return {};
      try {
        const resp = await adminApiClient.get<Record<string, unknown>>(`/functions/${functionId}/metrics`);
        const inner = (resp as { data?: Record<string, unknown> }).data;
        return inner ?? {};
      } catch {
        return {};
      }
    },
    enabled: Boolean(functionId),
  });

  const toggleVersion = (versionId: string) => {
    setExpandedVersions((prev) => {
      const next = new Set(prev);
      if (next.has(versionId)) {
        next.delete(versionId);
      } else {
        next.add(versionId);
      }
      return next;
    });
  };

  if (loadingFn || loadingMetrics) return <LoadingScreen />;

  const fn = functionData?.function;
  const versions = functionData?.versions ?? [];

  if (!fn) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[400px] text-center space-y-6">
        <div className="p-4 bg-red-100 dark:bg-red-900/20 rounded-full">
          <AlertTriangle className="w-12 h-12 text-red-600 dark:text-red-400" />
        </div>
        <div className="space-y-2">
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Function not found</h1>
          <p className="text-gray-500 dark:text-gray-400 max-w-md">
            The function you're looking for doesn't exist or may have been removed.
          </p>
        </div>
        <Link to={ROUTES.ADMIN_FUNCTIONS}>
          <Button variant="outline" className="gap-2">
            <ArrowLeft className="w-4 h-4" />
            Back to functions
          </Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">{fn.name}</h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">Function details, status, and metrics.</p>
        </div>
        <Link to={ROUTES.ADMIN_FUNCTIONS} className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300">
          Back
        </Link>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Detail label="Function ID" value={fn.id} />
        <Detail label="Author" value={fn.author || '-'} />
        <Detail label="Title" value={fn.title || '-'} />
        <Detail label="Visibility" value={fn.visibility || '-'} />
        <Detail label="Category" value={fn.category || '-'} />
        <Detail label="Latest Version" value={fn.latest_version || '-'} />
        <Detail label="Created" value={fn.created_at ? new Date(fn.created_at).toLocaleString() : '-'} />
      </div>

      <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4">
        <p className="text-sm text-gray-600 dark:text-gray-400">Metrics snapshot</p>
        <pre className="mt-2 text-xs overflow-auto bg-gray-50 dark:bg-gray-900 rounded p-3">{JSON.stringify(metricsResponse || {}, null, 2)}</pre>
      </div>

      {versions.length > 0 && (
        <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4">
          <div className="flex items-center gap-2 mb-4">
            <Code className="w-5 h-5 text-gray-600 dark:text-gray-400" />
            <p className="text-sm font-medium text-gray-700 dark:text-gray-300">Source Code ({versions.length} version{versions.length !== 1 ? 's' : ''})</p>
          </div>
          <div className="space-y-3">
            {versions.map((version: FunctionVersion) => {
              const isExpanded = expandedVersions.has(version.id);
              return (
                <div key={version.id} className="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
                  <button
                    onClick={() => toggleVersion(version.id)}
                    className="w-full flex items-center justify-between px-4 py-3 bg-gray-50 dark:bg-gray-900 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                  >
                    <div className="flex items-center gap-3">
                      <span className="font-mono text-sm font-medium text-gray-900 dark:text-gray-100">v{version.version}</span>
                      <span className="text-xs text-gray-500 dark:text-gray-400">{version.runtime}</span>
                      <span className="text-xs text-gray-500 dark:text-gray-400">{version.bundle_size > 0 ? `${(version.bundle_size / 1024).toFixed(1)} KB` : '-'}</span>
                    </div>
                    {isExpanded ? (
                      <ChevronDown className="w-4 h-4 text-gray-500 dark:text-gray-400" />
                    ) : (
                      <ChevronRight className="w-4 h-4 text-gray-500 dark:text-gray-400" />
                    )}
                  </button>
                  {isExpanded && version.source_code && (
                    <pre className="px-4 py-3 text-xs font-mono bg-gray-900 dark:bg-black text-gray-100 overflow-auto max-h-96 whitespace-pre-wrap">{version.source_code}</pre>
                  )}
                  {isExpanded && !version.source_code && (
                    <div className="px-4 py-3 space-y-2">
                      <div className="text-xs text-gray-500 dark:text-gray-400 bg-gray-50 dark:bg-gray-900 rounded p-3">
                        No source code stored on-chain. This function ships as a WASM binary only.
                      </div>
                      <div className="grid grid-cols-2 gap-2 text-xs">
                        <div>
                          <span className="text-gray-500 dark:text-gray-400">Runtime: </span>
                          <span className="font-mono text-gray-700 dark:text-gray-300">{version.runtime}</span>
                        </div>
                        <div>
                          <span className="text-gray-500 dark:text-gray-400">Bundle: </span>
                          <span className="font-mono text-gray-700 dark:text-gray-300">{version.bundle_size > 0 ? `${(version.bundle_size / 1024).toFixed(1)} KB` : '-'}</span>
                        </div>
                        <div className="col-span-2">
                          <span className="text-gray-500 dark:text-gray-400">Content Hash: </span>
                          <span className="font-mono text-gray-700 dark:text-gray-300 break-all">{version.content_hash || '-'}</span>
                        </div>
                        <div>
                          <span className="text-gray-500 dark:text-gray-400">Memory: </span>
                          <span className="font-mono text-gray-700 dark:text-gray-300">{version.memory_mb} MB</span>
                        </div>
                        <div>
                          <span className="text-gray-500 dark:text-gray-400">Timeout: </span>
                          <span className="font-mono text-gray-700 dark:text-gray-300">{version.timeout_ms} ms</span>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}

function Detail({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4">
      <p className="text-sm text-gray-600 dark:text-gray-400">{label}</p>
      <p className="text-sm font-medium text-gray-900 dark:text-gray-100 mt-1 break-all">{value}</p>
    </div>
  );
}
