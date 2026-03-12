import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { TrendingUp, AlertTriangle, Zap } from 'lucide-react';

interface RevenueGenerator {
  id: string;
  rank: number;
  tenantFunction: string;
  revenue30d: number;
  executionCount: number;
  growthRate: number;
}

interface SuspiciousGrowthAlert {
  id: string;
  tenantFunction: string;
  pattern: string;
  details: string;
  detectedAt: string;
}

interface ArtificialBoosting {
  id: string;
  function: string;
  detectedPattern: string;
  confidence: number;
  relatedAccounts: string[];
}

interface EconomicLeaderboardPayload {
  topRevenueGenerators?: RevenueGenerator[];
  suspiciousGrowth?: SuspiciousGrowthAlert[];
  artificialBoosting?: ArtificialBoosting[];
}

export function AdminEconomicLeaderboardPage() {
  const { data: raw, isLoading, isError } = useQuery({
    queryKey: ['admin-oversight-economic-leaderboard'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<EconomicLeaderboardPayload>('/oversight/economic-leaderboard');
      } catch {
        return null;
      }
    },
  });

  const hasDataWrapper = raw && typeof raw === 'object' && 'data' in raw;
  const payload: EconomicLeaderboardPayload = hasDataWrapper
    ? (raw as { data?: EconomicLeaderboardPayload }).data ?? {}
    : ((raw ?? {}) as unknown as EconomicLeaderboardPayload);

  const topRevenue = payload.topRevenueGenerators ?? [];
  const suspiciousGrowth = payload.suspiciousGrowth ?? [];
  const artificialBoosting = payload.artificialBoosting ?? [];

  if (isLoading) return <LoadingScreen />;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Economic Leaderboard</h1>
        <p className="mt-2 text-gray-600">Top economic performers and trend indicators.</p>
      </div>

      {isError || raw == null ? (
        <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 text-amber-800">
          <p className="font-medium">Unable to load economic leaderboard data.</p>
          <p className="text-sm mt-1">The oversight service or registry may be unavailable.</p>
        </div>
      ) : (
        <>
          <div>
            <h2 className="text-lg font-semibold text-gray-900 mb-3 flex items-center gap-2">
              <TrendingUp className="w-5 h-5 text-emerald-600" />
              Top revenue generators
            </h2>
            {topRevenue.length === 0 ? (
              <div className="bg-white border border-gray-200 rounded-lg p-6 text-center text-gray-500 text-sm">
                No revenue data yet.
              </div>
            ) : (
              <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="bg-gray-50 border-b border-gray-200">
                      <th className="px-4 py-3 text-left font-semibold text-gray-700">Rank</th>
                      <th className="px-4 py-3 text-left font-semibold text-gray-700">Tenant / Function</th>
                      <th className="px-4 py-3 text-left font-semibold text-gray-700">Revenue (30d)</th>
                      <th className="px-4 py-3 text-left font-semibold text-gray-700">Executions</th>
                      <th className="px-4 py-3 text-left font-semibold text-gray-700">Growth</th>
                    </tr>
                  </thead>
                  <tbody>
                    {topRevenue.map((r) => (
                      <tr key={r.id} className="border-b border-gray-100 hover:bg-gray-50">
                        <td className="px-4 py-3 font-medium text-gray-900">#{r.rank ?? '—'}</td>
                        <td className="px-4 py-3 text-gray-600">{r.tenantFunction ?? '—'}</td>
                        <td className="px-4 py-3 text-gray-900">
                          {typeof r.revenue30d === 'number' ? `$${r.revenue30d.toFixed(2)}` : '—'}
                        </td>
                        <td className="px-4 py-3 text-gray-600">{r.executionCount ?? '—'}</td>
                        <td className="px-4 py-3">
                          {typeof r.growthRate === 'number' ? (
                            <span className={r.growthRate >= 0 ? 'text-green-600' : 'text-red-600'}>
                              {r.growthRate >= 0 ? '+' : ''}{(r.growthRate * 100).toFixed(1)}%
                            </span>
                          ) : (
                            '—'
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>

          <div>
            <h2 className="text-lg font-semibold text-gray-900 mb-3 flex items-center gap-2">
              <AlertTriangle className="w-5 h-5 text-amber-600" />
              Suspicious growth
            </h2>
            {suspiciousGrowth.length === 0 ? (
              <div className="bg-white border border-gray-200 rounded-lg p-6 text-center text-gray-500 text-sm">
                No suspicious growth alerts.
              </div>
            ) : (
              <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="bg-gray-50 border-b border-gray-200">
                      <th className="px-4 py-3 text-left font-semibold text-gray-700">Tenant / Function</th>
                      <th className="px-4 py-3 text-left font-semibold text-gray-700">Pattern</th>
                      <th className="px-4 py-3 text-left font-semibold text-gray-700">Details</th>
                      <th className="px-4 py-3 text-left font-semibold text-gray-700">Detected</th>
                    </tr>
                  </thead>
                  <tbody>
                    {suspiciousGrowth.map((a) => (
                      <tr key={a.id} className="border-b border-gray-100 hover:bg-gray-50">
                        <td className="px-4 py-3 text-gray-900">{a.tenantFunction ?? '—'}</td>
                        <td className="px-4 py-3 text-gray-600">{a.pattern ?? '—'}</td>
                        <td className="px-4 py-3 text-gray-600 max-w-xs truncate" title={a.details}>{a.details ?? '—'}</td>
                        <td className="px-4 py-3 text-gray-500 text-xs">
                          {a.detectedAt ? new Date(a.detectedAt).toLocaleString() : '—'}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>

          <div>
            <h2 className="text-lg font-semibold text-gray-900 mb-3 flex items-center gap-2">
              <Zap className="w-5 h-5 text-violet-600" />
              Artificial boosting
            </h2>
            {artificialBoosting.length === 0 ? (
              <div className="bg-white border border-gray-200 rounded-lg p-6 text-center text-gray-500 text-sm">
                No artificial boosting alerts.
              </div>
            ) : (
              <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="bg-gray-50 border-b border-gray-200">
                      <th className="px-4 py-3 text-left font-semibold text-gray-700">Function</th>
                      <th className="px-4 py-3 text-left font-semibold text-gray-700">Pattern</th>
                      <th className="px-4 py-3 text-left font-semibold text-gray-700">Confidence</th>
                      <th className="px-4 py-3 text-left font-semibold text-gray-700">Related accounts</th>
                    </tr>
                  </thead>
                  <tbody>
                    {artificialBoosting.map((b) => (
                      <tr key={b.id} className="border-b border-gray-100 hover:bg-gray-50">
                        <td className="px-4 py-3 text-gray-900">{b.function ?? '—'}</td>
                        <td className="px-4 py-3 text-gray-600">{b.detectedPattern ?? '—'}</td>
                        <td className="px-4 py-3">
                          <span className={b.confidence >= 80 ? 'text-red-600 font-medium' : 'text-gray-600'}>
                            {b.confidence ?? 0}%
                          </span>
                        </td>
                        <td className="px-4 py-3 text-gray-500 text-xs">
                          {(b.relatedAccounts ?? []).length} account(s)
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </>
      )}
    </div>
  );
}
