import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { Shield, AlertTriangle, TrendingUp } from 'lucide-react';

interface TrustDistribution {
  excellent: number;
  good: number;
  fair: number;
  poor: number;
}

interface HighRiskFunction {
  id: string;
  name: string;
  tenant: string;
  trustScore: number;
  riskFactors: string[];
  lastUpdated: string;
}

interface TrustDashboardPayload {
  distribution?: TrustDistribution;
  highRiskFunctions?: HighRiskFunction[];
  trustSpikes?: unknown[];
  reputationFarmingAlerts?: unknown[];
}

export function AdminTrustDashboardPage() {
  const { data: raw, isLoading, isError } = useQuery({
    queryKey: ['admin-oversight-trust-dashboard'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<TrustDashboardPayload>('/oversight/trust-dashboard');
      } catch {
        return null;
      }
    },
  });

  // API returns payload at root (no .data wrapper)
  const payload = (raw as TrustDashboardPayload | undefined) ?? {};
  const hasDataWrapper = payload && 'data' in payload && typeof (payload as { data?: TrustDashboardPayload }).data === 'object';
  const dashboard: TrustDashboardPayload = hasDataWrapper
    ? (payload as { data: TrustDashboardPayload }).data ?? {}
    : payload;

  const distribution = dashboard.distribution ?? { excellent: 0, good: 0, fair: 0, poor: 0 };
  const highRisk = dashboard.highRiskFunctions ?? [];
  const total = distribution.excellent + distribution.good + distribution.fair + distribution.poor;

  if (isLoading) return <LoadingScreen />;

  if (isError || raw == null) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Trust Dashboard</h1>
          <p className="mt-2 text-gray-600">Trust and safety indicators across the platform.</p>
        </div>
        <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 text-amber-800">
          <p className="font-medium">Unable to load trust data.</p>
          <p className="text-sm mt-1">The oversight service or registry may be unavailable.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Trust Dashboard</h1>
        <p className="mt-2 text-gray-600">Trust and safety indicators across the platform.</p>
      </div>

      {/* Distribution */}
      <div>
        <h2 className="text-lg font-semibold text-gray-900 mb-3 flex items-center gap-2">
          <Shield className="w-5 h-5 text-emerald-600" />
          Trust distribution
        </h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <StatCard label="Excellent" value={distribution.excellent} total={total} color="emerald" />
          <StatCard label="Good" value={distribution.good} total={total} color="green" />
          <StatCard label="Fair" value={distribution.fair} total={total} color="amber" />
          <StatCard label="Poor" value={distribution.poor} total={total} color="red" />
        </div>
      </div>

      {/* High-risk functions */}
      <div>
        <h2 className="text-lg font-semibold text-gray-900 mb-3 flex items-center gap-2">
          <AlertTriangle className="w-5 h-5 text-amber-600" />
          High-risk functions
        </h2>
        {highRisk.length === 0 ? (
          <div className="bg-white border border-gray-200 rounded-lg p-6 text-center text-gray-500">
            No high-risk functions identified.
          </div>
        ) : (
          <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
            <table className="w-full">
              <thead>
                <tr className="bg-gray-50 border-b border-gray-200">
                  <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700">Name</th>
                  <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700">Tenant</th>
                  <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700">Trust score</th>
                  <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700">Risk factors</th>
                  <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700">Last updated</th>
                </tr>
              </thead>
              <tbody>
                {highRisk.map((fn) => (
                  <tr key={fn.id} className="border-b border-gray-100 hover:bg-gray-50">
                    <td className="px-4 py-3 text-sm font-medium text-gray-900">{fn.name || fn.id}</td>
                    <td className="px-4 py-3 text-sm text-gray-600">{fn.tenant || '—'}</td>
                    <td className="px-4 py-3 text-sm">
                      <span className={fn.trustScore < 30 ? 'text-red-600 font-medium' : fn.trustScore < 60 ? 'text-amber-600' : 'text-gray-600'}>
                        {Math.round(fn.trustScore)}%
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-600">
                      <ul className="list-disc list-inside space-y-0.5">
                        {(fn.riskFactors ?? []).slice(0, 3).map((f, i) => (
                          <li key={i}>{f}</li>
                        ))}
                        {(fn.riskFactors?.length ?? 0) > 3 && <li>…</li>}
                      </ul>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-500">
                      {fn.lastUpdated ? new Date(fn.lastUpdated).toLocaleString() : '—'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {(dashboard.trustSpikes?.length ?? 0) > 0 && (
        <div>
          <h2 className="text-lg font-semibold text-gray-900 mb-3 flex items-center gap-2">
            <TrendingUp className="w-5 h-5" />
            Trust spikes
          </h2>
          <div className="bg-white border border-gray-200 rounded-lg p-4 text-sm text-gray-600">
            {dashboard.trustSpikes!.length} spike(s) detected.
          </div>
        </div>
      )}
    </div>
  );
}

function StatCard({
  label,
  value,
  total,
  color,
}: {
  label: string;
  value: number;
  total: number;
  color: 'emerald' | 'green' | 'amber' | 'red';
}) {
  const pct = total > 0 ? Math.round((value / total) * 100) : 0;
  const colorClasses = {
    emerald: 'bg-emerald-50 border-emerald-200 text-emerald-800',
    green: 'bg-green-50 border-green-200 text-green-800',
    amber: 'bg-amber-50 border-amber-200 text-amber-800',
    red: 'bg-red-50 border-red-200 text-red-800',
  };
  return (
    <div className={`rounded-lg border p-4 ${colorClasses[color]}`}>
      <p className="text-sm font-medium opacity-90">{label}</p>
      <p className="text-2xl font-bold mt-1">{value}</p>
      {total > 0 && <p className="text-xs mt-1 opacity-80">{pct}% of total</p>}
    </div>
  );
}
