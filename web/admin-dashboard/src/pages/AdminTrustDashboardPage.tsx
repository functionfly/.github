import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { Shield, AlertTriangle, TrendingUp, Settings, Save, RotateCcw } from 'lucide-react';
import { useState } from 'react';

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

interface TrustScoreWeights {
  id: string;
  name: string;
  description: string;
  reliability: number;
  latency: number;
  error_rate: number;
  user_rating: number;
  verification: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
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

  const { data: weights, isLoading: weightsLoading } = useQuery({
    queryKey: ['trust-score-weights'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<TrustScoreWeights>('/reputation/trust-weights');
      } catch {
        return null;
      }
    },
  });

  const queryClient = useQueryClient();
  const [editMode, setEditMode] = useState(false);
  const [editedWeights, setEditedWeights] = useState<TrustScoreWeights | null>(null);

  const updateWeightsMutation = useMutation({
    mutationFn: async (weights: TrustScoreWeights) => {
      return await adminApiClient.put('/reputation/trust-weights', weights);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['trust-score-weights'] });
      setEditMode(false);
    },
  });

  const detectFarmingMutation = useMutation({
    mutationFn: async () => {
      return await adminApiClient.post('/reputation/detect-farming', {});
    },
  });

  const cleanupHistoryMutation = useMutation({
    mutationFn: async (retentionDays: number) => {
      return await adminApiClient.post('/reputation/cleanup-trust-history', {}, { params: { retentionDays } });
    },
  });

  const apiWeights = (weights as { data?: TrustScoreWeights })?.data ?? weights as TrustScoreWeights | null;

  const handleEdit = () => {
    if (apiWeights) {
      setEditedWeights({ ...apiWeights });
      setEditMode(true);
    }
  };

  const handleSave = () => {
    if (editedWeights) {
      updateWeightsMutation.mutate(editedWeights);
    }
  };

  const handleCancel = () => {
    setEditedWeights(null);
    setEditMode(false);
  };

  const handleWeightChange = (field: keyof TrustScoreWeights, value: number) => {
    if (editedWeights) {
      setEditedWeights({ ...editedWeights, [field]: value });
    }
  };

  const totalWeight = editedWeights
    ? editedWeights.reliability + editedWeights.latency + editedWeights.error_rate + editedWeights.user_rating + editedWeights.verification
    : 0;

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

      {/* Trust Score Weights Configuration */}
      <div>
        <h2 className="text-lg font-semibold text-gray-900 mb-3 flex items-center gap-2">
          <Settings className="w-5 h-5 text-blue-600" />
          Trust Score Weights Configuration
        </h2>
        <div className="bg-white border border-gray-200 rounded-lg p-6">
          {weightsLoading ? (
            <p className="text-gray-500">Loading weights...</p>
          ) : editMode && editedWeights ? (
            <div className="space-y-4">
              <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
                <WeightInput
                  label="Reliability"
                  value={editedWeights.reliability}
                  onChange={(v) => handleWeightChange('reliability', v)}
                />
                <WeightInput
                  label="Latency"
                  value={editedWeights.latency}
                  onChange={(v) => handleWeightChange('latency', v)}
                />
                <WeightInput
                  label="Error Rate"
                  value={editedWeights.error_rate}
                  onChange={(v) => handleWeightChange('error_rate', v)}
                />
                <WeightInput
                  label="User Rating"
                  value={editedWeights.user_rating}
                  onChange={(v) => handleWeightChange('user_rating', v)}
                />
                <WeightInput
                  label="Verification"
                  value={editedWeights.verification}
                  onChange={(v) => handleWeightChange('verification', v)}
                />
              </div>
              <div className="flex items-center justify-between">
                <div className="text-sm">
                  <span className="font-medium">Total:</span>{' '}
                  <span className={Math.abs(totalWeight - 1) < 0.001 ? 'text-green-600' : 'text-red-600'}>
                    {(totalWeight * 100).toFixed(1)}%
                  </span>
                  {Math.abs(totalWeight - 1) >= 0.001 && (
                    <span className="text-red-600 ml-2">Weights must sum to 100%</span>
                  )}
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={handleCancel}
                    className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 flex items-center gap-2"
                  >
                    <RotateCcw className="w-4 h-4" />
                    Cancel
                  </button>
                  <button
                    onClick={handleSave}
                    disabled={Math.abs(totalWeight - 1) >= 0.001 || updateWeightsMutation.isPending}
                    className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 flex items-center gap-2"
                  >
                    <Save className="w-4 h-4" />
                    {updateWeightsMutation.isPending ? 'Saving...' : 'Save Changes'}
                  </button>
                </div>
              </div>
            </div>
          ) : apiWeights ? (
            <div className="space-y-4">
              <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
                <WeightDisplay label="Reliability" value={apiWeights.reliability} />
                <WeightDisplay label="Latency" value={apiWeights.latency} />
                <WeightDisplay label="Error Rate" value={apiWeights.error_rate} />
                <WeightDisplay label="User Rating" value={apiWeights.user_rating} />
                <WeightDisplay label="Verification" value={apiWeights.verification} />
              </div>
              <div className="flex items-center justify-between">
                <div className="text-sm text-gray-500">
                  Last updated: {apiWeights.updated_at ? new Date(apiWeights.updated_at).toLocaleString() : 'N/A'}
                </div>
                <button
                  onClick={handleEdit}
                  className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 flex items-center gap-2"
                >
                  <Settings className="w-4 h-4" />
                  Edit Weights
                </button>
              </div>
            </div>
          ) : (
            <p className="text-gray-500">Unable to load trust score weights.</p>
          )}
        </div>
      </div>

      {/* Reputation Farming Detection */}
      <div>
        <h2 className="text-lg font-semibold text-gray-900 mb-3 flex items-center gap-2">
          <AlertTriangle className="w-5 h-5 text-amber-600" />
          Reputation Farming Detection
        </h2>
        <div className="bg-white border border-gray-200 rounded-lg p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium text-gray-900">Run Detection</p>
              <p className="text-sm text-gray-500">Analyze patterns for potential reputation farming</p>
            </div>
            <button
              onClick={() => detectFarmingMutation.mutate()}
              disabled={detectFarmingMutation.isPending}
              className="px-4 py-2 text-sm font-medium text-white bg-amber-600 rounded-lg hover:bg-amber-700 disabled:opacity-50"
            >
              {detectFarmingMutation.isPending ? 'Running...' : 'Run Detection'}
            </button>
          </div>
          {detectFarmingMutation.isSuccess && (
            <div className="mt-4 p-4 bg-green-50 border border-green-200 rounded-lg">
              <p className="text-green-800 font-medium">Detection complete!</p>
              <p className="text-sm text-green-700">
                {detectFarmingMutation.data?.alerts?.length ?? 0} alert(s) generated
              </p>
            </div>
          )}
          {detectFarmingMutation.isError && (
            <div className="mt-4 p-4 bg-red-50 border border-red-200 rounded-lg">
              <p className="text-red-800 font-medium">Detection failed</p>
            </div>
          )}
        </div>
      </div>

      {/* Trust History Cleanup */}
      <div>
        <h2 className="text-lg font-semibold text-gray-900 mb-3 flex items-center gap-2">
          <TrendingUp className="w-5 h-5 text-purple-600" />
          Trust History Data Retention
        </h2>
        <div className="bg-white border border-gray-200 rounded-lg p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium text-gray-900">Cleanup Old Trust History</p>
              <p className="text-sm text-gray-500">Remove trust history entries older than retention period</p>
            </div>
            <button
              onClick={() => cleanupHistoryMutation.mutate(90)}
              disabled={cleanupHistoryMutation.isPending}
              className="px-4 py-2 text-sm font-medium text-white bg-purple-600 rounded-lg hover:bg-purple-700 disabled:opacity-50"
            >
              {cleanupHistoryMutation.isPending ? 'Cleaning...' : 'Cleanup (90 days)'}
            </button>
          </div>
          {cleanupHistoryMutation.isSuccess && (
            <div className="mt-4 p-4 bg-green-50 border border-green-200 rounded-lg">
              <p className="text-green-800 font-medium">Cleanup complete!</p>
              <p className="text-sm text-green-700">
                {cleanupHistoryMutation.data?.deleted_entries ?? 0} entries removed
              </p>
            </div>
          )}
        </div>
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

function WeightDisplay({ label, value }: { label: string; value: number }) {
  return (
    <div className="text-center">
      <p className="text-sm text-gray-500 mb-1">{label}</p>
      <p className="text-2xl font-bold text-gray-900">{(value * 100).toFixed(0)}%</p>
    </div>
  );
}

function WeightInput({
  label,
  value,
  onChange,
}: {
  label: string;
  value: number;
  onChange: (value: number) => void;
}) {
  return (
    <div>
      <label className="block text-sm text-gray-500 mb-1">{label}</label>
      <input
        type="number"
        min="0"
        max="1"
        step="0.01"
        value={value}
        onChange={(e) => onChange(parseFloat(e.target.value) || 0)}
        className="w-full px-3 py-2 border border-gray-300 rounded-lg text-center font-medium"
      />
    </div>
  );
}
