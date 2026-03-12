import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { CheckCircle, XCircle, AlertCircle } from 'lucide-react';

interface ExecutionRecord {
  id: string;
  executionRootHash: string;
  tenant: string;
  functionName: string;
  timestamp: string;
  nodeSignature: string;
  status: string;
  duration: number;
  inputSize: number;
  outputSize: number;
  errorMessage?: string | null;
}

interface ExecutionAuditPayload {
  executions?: ExecutionRecord[];
  total?: number;
  page?: number;
  pageSize?: number;
}

export function AdminExecutionAuditPage() {
  const [page, setPage] = useState(1);
  const pageSize = 20;

  const { data: raw, isLoading, isError } = useQuery({
    queryKey: ['admin-oversight-execution-audit', page, pageSize],
    queryFn: async () => {
      try {
        return await adminApiClient.get<ExecutionAuditPayload>(
          `/oversight/execution-audit?page=${page}&pageSize=${pageSize}`
        );
      } catch {
        return null;
      }
    },
  });

  const hasDataWrapper = raw && typeof raw === 'object' && 'data' in raw;
  const payload: ExecutionAuditPayload = hasDataWrapper
    ? (raw as { data?: ExecutionAuditPayload }).data ?? {}
    : ((raw ?? {}) as unknown as ExecutionAuditPayload);

  const executions = payload.executions ?? [];
  const total = payload.total ?? 0;
  const currentPage = payload.page ?? page;
  const totalPages = payload.pageSize && payload.pageSize > 0 ? Math.ceil(total / payload.pageSize) : 1;

  if (isLoading) return <LoadingScreen />;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Execution Audit</h1>
        <p className="mt-2 text-gray-600">Execution-level audit summaries and anomalies.</p>
      </div>

      {isError || raw == null ? (
        <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 text-amber-800">
          <p className="font-medium">Unable to load execution audit data.</p>
          <p className="text-sm mt-1">The oversight service or registry may be unavailable.</p>
        </div>
      ) : (
        <>
          <div className="flex items-center justify-between">
            <p className="text-sm text-gray-600">
              Showing {executions.length} of {total} execution(s)
              {totalPages > 1 && ` · Page ${currentPage} of ${totalPages}`}
            </p>
            {totalPages > 1 && (
              <div className="flex gap-2">
                <button
                  type="button"
                  disabled={currentPage <= 1}
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  className="px-3 py-1 border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Previous
                </button>
                <button
                  type="button"
                  disabled={currentPage >= totalPages}
                  onClick={() => setPage((p) => p + 1)}
                  className="px-3 py-1 border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Next
                </button>
              </div>
            )}
          </div>

          {executions.length === 0 ? (
            <div className="bg-white border border-gray-200 rounded-lg p-8 text-center text-gray-500">
              No execution records found.
            </div>
          ) : (
            <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
              <table className="w-full">
                <thead>
                  <tr className="bg-gray-50 border-b border-gray-200">
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700">Time</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700">Function</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700">Tenant</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700">Status</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700">Duration</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700">I/O</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700">Hash</th>
                  </tr>
                </thead>
                <tbody>
                  {executions.map((ex) => (
                    <tr key={ex.id} className="border-b border-gray-100 hover:bg-gray-50">
                      <td className="px-4 py-3 text-sm text-gray-600 whitespace-nowrap">
                        {ex.timestamp ? new Date(ex.timestamp).toLocaleString() : '—'}
                      </td>
                      <td className="px-4 py-3 text-sm font-medium text-gray-900">{ex.functionName || ex.id}</td>
                      <td className="px-4 py-3 text-sm text-gray-600">{ex.tenant || '—'}</td>
                      <td className="px-4 py-3 text-sm">
                        <StatusBadge status={ex.status} errorMessage={ex.errorMessage} />
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-600">{ex.duration != null ? `${ex.duration} ms` : '—'}</td>
                      <td className="px-4 py-3 text-sm text-gray-500">
                        {ex.inputSize != null || ex.outputSize != null
                          ? `${ex.inputSize ?? 0} / ${ex.outputSize ?? 0} B`
                          : '—'}
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-400 font-mono truncate max-w-[120px]" title={ex.executionRootHash}>
                        {ex.executionRootHash ? `${ex.executionRootHash.slice(0, 12)}…` : '—'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </>
      )}
    </div>
  );
}

function StatusBadge({ status, errorMessage }: { status: string; errorMessage?: string | null }) {
  const s = (status ?? '').toLowerCase();
  if (s === 'success' || s === 'ok' || s === 'completed') {
    return (
      <span className="inline-flex items-center gap-1 text-green-700">
        <CheckCircle className="w-4 h-4" /> Success
      </span>
    );
  }
  if (s === 'error' || s === 'failed' || errorMessage) {
    return (
      <span className="inline-flex items-center gap-1 text-red-700" title={errorMessage ?? undefined}>
        <XCircle className="w-4 h-4" /> {status || 'Failed'}
      </span>
    );
  }
  return (
    <span className="inline-flex items-center gap-1 text-amber-700">
      <AlertCircle className="w-4 h-4" /> {status || 'Unknown'}
    </span>
  );
}
