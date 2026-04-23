import { adminApiClient } from '@/lib/api/adminClient';
import { API_ROUTES } from '@/lib/constants';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Clock, Database, HardDrive, RefreshCw, Trash2, Zap } from 'lucide-react';

// Types
interface CacheStats {
  total_entries: number;
  total_size_bytes: number;
  hit_rate: number;
  miss_rate: number;
  evictions: number;
  memory_used: number;
  memory_available: number;
  by_function: {
    function_id: string;
    function_name: string;
    entries: number;
    size_bytes: number;
  }[];
}

type RawCacheStats = {
  memory_cache?: {
    hits?: number;
    misses?: number;
    hit_ratio?: number;
    size_bytes?: number;
    evictions?: number;
  };
  disk_cache?: {
    total_entries?: number;
    total_size_bytes?: number;
    total_hits?: number;
    expired_entries?: number;
    hit_ratio?: number;
  };
};

/**
 * Admin Cache Management Page
 * Manages cache settings and monitoring
 */
export function AdminCachePage() {
  const queryClient = useQueryClient();

  const normalizeCacheStats = (raw: RawCacheStats): CacheStats => {
    const mem = raw.memory_cache;
    const disk = raw.disk_cache;

    const totalEntries = Number(disk?.total_entries ?? 0);
    const totalSizeBytes = Number(disk?.total_size_bytes ?? mem?.size_bytes ?? 0);
    const hitRate = Number(mem?.hit_ratio ?? disk?.hit_ratio ?? 0);
    const misses = Number(mem?.misses ?? 0);
    const hits = Number(mem?.hits ?? 0);
    const totalLookups = hits + misses;
    const missRate = totalLookups > 0 ? misses / totalLookups : 0;

    return {
      total_entries: Number.isFinite(totalEntries) ? totalEntries : 0,
      total_size_bytes: Number.isFinite(totalSizeBytes) ? totalSizeBytes : 0,
      hit_rate: Number.isFinite(hitRate) ? hitRate : 0,
      miss_rate: Number.isFinite(missRate) ? missRate : 0,
      evictions: Number(mem?.evictions ?? 0) || 0,
      memory_used: 0,
      memory_available: 0,
      by_function: [],
    };
  };

  // Fetch cache stats
  const {
    data: stats,
    isLoading: statsLoading,
    refetch: refetchStats,
  } = useQuery<CacheStats>({
    queryKey: ['cache-stats'],
    queryFn: async () => {
      // Backend returns a raw JSON object (not the { data, success } envelope).
      // React Query requires queryFn to return a non-undefined value.
      const rawUnknown = await adminApiClient.get<unknown>(API_ROUTES.ADMIN_CACHE_STATS);
      const raw =
        rawUnknown && typeof rawUnknown === 'object'
          ? (rawUnknown as RawCacheStats)
          : ({} as RawCacheStats);
      return normalizeCacheStats(raw);
    },
    refetchInterval: 30000, // Refresh every 30 seconds
  });

  // Purge all cache mutation
  const purgeAllMutation = useMutation({
    mutationFn: async () => {
      return adminApiClient.delete(API_ROUTES.ADMIN_CACHE_STATS);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cache-stats'] });
    },
  });

  // Purge function cache mutation
  const purgeFunctionMutation = useMutation({
    mutationFn: async (functionId: string) => {
      return adminApiClient.delete(`/cache/${functionId}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cache-stats'] });
    },
  });

  const handlePurgeAll = () => {
    if (confirm('Are you sure you want to purge all cache? This will clear all cached data.')) {
      purgeAllMutation.mutate();
    }
  };

  const handlePurgeFunction = (functionId: string) => {
    if (confirm(`Are you sure you want to purge cache for function ${functionId}?`)) {
      purgeFunctionMutation.mutate(functionId);
    }
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatPercentage = (value: number) => {
    return (value * 100).toFixed(2) + '%';
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Cache Management</h1>
          <p className="text-gray-600 dark:text-gray-400 mt-1">Monitor and manage system cache</p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => refetchStats()}
            disabled={statsLoading}
            className="flex items-center gap-2 px-4 py-2 bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 disabled:opacity-50"
          >
            <RefreshCw className={`w-4 h-4 ${statsLoading ? 'animate-spin' : ''}`} />
            Refresh
          </button>
          <button
            onClick={handlePurgeAll}
            disabled={purgeAllMutation.isPending}
            className="flex items-center gap-2 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 disabled:opacity-50"
          >
            <Trash2 className="w-4 h-4" />
            Purge All Cache
          </button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        {/* Total Entries */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Total Entries</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                {stats?.total_entries?.toLocaleString() || 0}
              </p>
            </div>
            <Database className="w-8 h-8 text-blue-500" />
          </div>
        </div>

        {/* Total Size */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Total Size</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                {formatBytes(stats?.total_size_bytes || 0)}
              </p>
            </div>
            <HardDrive className="w-8 h-8 text-purple-500" />
          </div>
        </div>

        {/* Hit Rate */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Hit Rate</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                {formatPercentage(stats?.hit_rate || 0)}
              </p>
            </div>
            <Zap className="w-8 h-8 text-green-500" />
          </div>
        </div>

        {/* Evictions */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Evictions</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                {stats?.evictions?.toLocaleString() || 0}
              </p>
            </div>
            <Clock className="w-8 h-8 text-yellow-500" />
          </div>
        </div>
      </div>

      {/* Memory Usage */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">Memory Usage</h2>
        <div className="w-full bg-gray-200 dark:bg-gray-600 rounded-full h-4">
          <div
            className="bg-blue-600 h-4 rounded-full"
            style={{
              width: `${
                stats?.memory_used && stats?.memory_available
                  ? (stats.memory_used / stats.memory_available) * 100
                  : 0
              }%`,
            }}
          />
        </div>
        <div className="mt-2 flex justify-between text-sm text-gray-600 dark:text-gray-400">
          <span>Used: {formatBytes(stats?.memory_used || 0)}</span>
          <span>Available: {formatBytes(stats?.memory_available || 0)}</span>
        </div>
      </div>

      {/* Cache by Function */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Cache by Function</h2>
        </div>
        {statsLoading ? (
          <div className="p-8 text-center text-gray-500 dark:text-gray-400">Loading...</div>
        ) : stats?.by_function?.length === 0 ? (
          <div className="p-8 text-center text-gray-500 dark:text-gray-400">No cached functions</div>
        ) : (
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
            <thead className="bg-gray-50 dark:bg-gray-800">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  Function
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  Entries
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  Size
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
              {stats?.by_function?.map((func) => (
                <tr key={func.function_id}>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm font-medium text-gray-900 dark:text-gray-100">{func.function_name}</div>
                    <div className="text-sm text-gray-500 dark:text-gray-400">{func.function_id}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                    {func.entries.toLocaleString()}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                    {formatBytes(func.size_bytes)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <button
                      onClick={() => handlePurgeFunction(func.function_id)}
                      className="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}

export default AdminCachePage;
