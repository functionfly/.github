import React from 'react';
import { TrustScoreBadge, TrustLevel } from './TrustScoreBadge';
import { TrustScoreGauge, TrustScoreBar } from './TrustScoreGauge';
import {
  Shield,
  Zap,
  Clock,
  AlertTriangle,
  Users,
  Building2,
  User,
  Activity
} from 'lucide-react';

interface TrustMetrics {
  trust_score: number;
  trust_level?: TrustLevel;
  success_rate: number;
  p50_latency_ms: number;
  p95_latency_ms: number;
  timeout_rate: number;
  error_rate: number;
  consumer_diversity: number;
  tenant_diversity: number;
  user_diversity: number;
}

interface TrustScoreCardProps {
  metrics: TrustMetrics;
  showGauge?: boolean;
  showDetails?: boolean;
  compact?: boolean;
}

const MetricRow: React.FC<{
  icon: React.ReactNode;
  label: string;
  value: string | number;
  subValue?: string;
  trend?: 'up' | 'down' | 'stable';
}> = ({ icon, label, value, subValue, trend }) => (
  <div className="flex items-center justify-between py-2 border-b border-gray-100 last:border-0">
    <div className="flex items-center gap-2 text-gray-600">
      <span className="w-4 h-4">{icon}</span>
      <span className="text-sm">{label}</span>
    </div>
    <div className="text-right">
      <span className="font-semibold text-gray-900">{value}</span>
      {subValue && <span className="text-xs text-gray-500 ml-1">{subValue}</span>}
    </div>
  </div>
);

export function TrustScoreCard({
  metrics,
  showGauge = true,
  showDetails = true,
  compact = false
}: TrustScoreCardProps) {
  const {
    trust_score,
    trust_level,
    success_rate,
    p50_latency_ms,
    p95_latency_ms,
    timeout_rate,
    error_rate,
    consumer_diversity,
    tenant_diversity,
    user_diversity,
  } = metrics;

  if (compact) {
    return (
      <div className="flex items-center gap-3">
        <TrustScoreBadge trustScore={trust_score} trustLevel={trust_level} size="sm" />
        <div className="flex-1">
          <TrustScoreBar score={trust_score} height="sm" showLabel={false} />
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
      {/* Header */}
      <div className="bg-gradient-to-r from-gray-50 to-gray-100 px-4 py-3 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Shield className="w-5 h-5 text-gray-700" />
            <span className="font-semibold text-gray-900">Trust & Safety Score</span>
          </div>
          <TrustScoreBadge trustScore={trust_score} trustLevel={trust_level} showScore />
        </div>
      </div>

      <div className="p-4">
        {/* Gauge Section */}
        {showGauge && (
          <div className="flex justify-center mb-6">
            <TrustScoreGauge score={trust_score} size="lg" />
          </div>
        )}

        {/* Trust Level */}
        <div className="text-center mb-4">
          <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full text-sm font-medium bg-gray-100 text-gray-700">
            <Activity className="w-4 h-4" />
            Level: {trust_level?.replace('_', ' ') || 'N/A'}
          </span>
        </div>

        {/* Metrics Grid */}
        {showDetails && (
          <div className="grid grid-cols-2 gap-4 mt-4">
            {/* Performance Metrics */}
            <div className="bg-gray-50 rounded-lg p-3">
              <h4 className="text-xs font-semibold text-gray-500 uppercase mb-2">Performance</h4>
              <MetricRow
                icon={<Zap className="w-4 h-4 text-yellow-500" />}
                label="Success Rate"
                value={`${success_rate.toFixed(1)}%`}
              />
              <MetricRow
                icon={<Clock className="w-4 h-4 text-blue-500" />}
                label="P50 Latency"
                value={`${p50_latency_ms}ms`}
              />
              <MetricRow
                icon={<Clock className="w-4 h-4 text-purple-500" />}
                label="P95 Latency"
                value={`${p95_latency_ms}ms`}
              />
            </div>

            {/* Reliability Metrics */}
            <div className="bg-gray-50 rounded-lg p-3">
              <h4 className="text-xs font-semibold text-gray-500 uppercase mb-2">Reliability</h4>
              <MetricRow
                icon={<AlertTriangle className="w-4 h-4 text-orange-500" />}
                label="Timeout Rate"
                value={`${timeout_rate.toFixed(2)}%`}
              />
              <MetricRow
                icon={<AlertTriangle className="w-4 h-4 text-red-500" />}
                label="Error Rate"
                value={`${error_rate.toFixed(2)}%`}
              />
            </div>

            {/* Diversity Metrics */}
            <div className="bg-gray-50 rounded-lg p-3 col-span-2">
              <h4 className="text-xs font-semibold text-gray-500 uppercase mb-2">Consumer Diversity</h4>
              <div className="grid grid-cols-3 gap-2">
                <div className="text-center">
                  <div className="flex items-center justify-center gap-1 text-gray-600 mb-1">
                    <Building2 className="w-3 h-3" />
                    <span className="text-xs">Tenants</span>
                  </div>
                  <span className="font-semibold text-gray-900">{tenant_diversity}</span>
                </div>
                <div className="text-center">
                  <div className="flex items-center justify-center gap-1 text-gray-600 mb-1">
                    <User className="w-3 h-3" />
                    <span className="text-xs">Users</span>
                  </div>
                  <span className="font-semibold text-gray-900">{user_diversity}</span>
                </div>
                <div className="text-center">
                  <div className="flex items-center justify-center gap-1 text-gray-600 mb-1">
                    <Users className="w-3 h-3" />
                    <span className="text-xs">IP Diversity</span>
                  </div>
                  <span className="font-semibold text-gray-900">{consumer_diversity.toFixed(1)}%</span>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Trust Score Bar */}
        <div className="mt-4 pt-4 border-t border-gray-200">
          <div className="flex items-center justify-between text-sm text-gray-600 mb-2">
            <span>Overall Trust Score</span>
            <span className="font-semibold">{trust_score.toFixed(1)} / 100</span>
          </div>
          <TrustScoreBar score={trust_score} />
        </div>
      </div>
    </div>
  );
}

// Compact version for function list items
export function TrustScoreCompact({ metrics }: { metrics: TrustMetrics }) {
  return (
    <div className="flex items-center gap-2">
      <TrustScoreGauge score={metrics.trust_score} size="sm" showLabel={false} />
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between">
          <span className="text-sm font-medium text-gray-900">
            {Math.round(metrics.trust_score)}
          </span>
          <span className="text-xs text-gray-500">
            {metrics.success_rate.toFixed(0)}% ✓
          </span>
        </div>
        <TrustScoreBar score={metrics.trust_score} height="sm" showLabel={false} />
      </div>
    </div>
  );
}
