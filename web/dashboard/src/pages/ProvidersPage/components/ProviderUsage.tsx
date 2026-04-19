import { BarChart3, Activity, Globe, Clock } from 'lucide-react';

interface ProviderUsageStats {
  functionCount: number;
  requestCount: number;
  requestChangePercent: number;
  avgLatency: number;
  errorRate: number;
}

interface ProviderUsageProps {
  stats: ProviderUsageStats;
  compact?: boolean;
}

export function ProviderUsage({ stats, compact = false }: ProviderUsageProps) {
  const isPositiveChange = stats.requestChangePercent >= 0;

  if (compact) {
    return (
      <div className="flex items-center gap-3 text-sm">
        <div className="flex items-center gap-1.5" title="Active functions">
          <BarChart3 className="w-3.5 h-3.5 text-text-muted" />
          <span className="text-text-secondary">
            {stats.functionCount} <span className="text-text-tertiary">funcs</span>
          </span>
        </div>
        <div className="flex items-center gap-1.5" title="Requests (24h)">
          <Activity className="w-3.5 h-3.5 text-text-muted" />
          <span className="text-text-secondary">
            {formatNumber(stats.requestCount)}
          </span>
          <span
            className={`text-xs ${isPositiveChange ? 'text-emerald-500' : 'text-red-500'}`}
          >
            {isPositiveChange ? '+' : ''}{stats.requestChangePercent.toFixed(0)}%
          </span>
        </div>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-2 gap-3">
      <div className="p-2.5 rounded-lg bg-bg-secondary/50 border border-border-subtle">
        <div className="flex items-center gap-2 mb-1">
          <BarChart3 className="w-4 h-4 text-text-muted" />
          <span className="text-xs text-text-tertiary">Functions</span>
        </div>
        <p className="text-lg font-semibold text-text-primary">{stats.functionCount}</p>
      </div>

      <div className="p-2.5 rounded-lg bg-bg-secondary/50 border border-border-subtle">
        <div className="flex items-center gap-2 mb-1">
          <Activity className="w-4 h-4 text-text-muted" />
          <span className="text-xs text-text-tertiary">Requests (24h)</span>
        </div>
        <div className="flex items-baseline gap-2">
          <p className="text-lg font-semibold text-text-primary">{formatNumber(stats.requestCount)}</p>
          <span
            className={`text-xs ${isPositiveChange ? 'text-emerald-500' : 'text-red-500'}`}
          >
            {isPositiveChange ? '+' : ''}{stats.requestChangePercent.toFixed(0)}%
          </span>
        </div>
      </div>

      <div className="p-2.5 rounded-lg bg-bg-secondary/50 border border-border-subtle">
        <div className="flex items-center gap-2 mb-1">
          <Clock className="w-4 h-4 text-text-muted" />
          <span className="text-xs text-text-tertiary">Avg Latency</span>
        </div>
        <p className="text-lg font-semibold text-text-primary">{stats.avgLatency}ms</p>
      </div>

      <div className="p-2.5 rounded-lg bg-bg-secondary/50 border border-border-subtle">
        <div className="flex items-center gap-2 mb-1">
          <Globe className="w-4 h-4 text-text-muted" />
          <span className="text-xs text-text-tertiary">Error Rate</span>
        </div>
        <p className={`text-lg font-semibold ${stats.errorRate > 1 ? 'text-red-500' : 'text-text-primary'}`}>
          {stats.errorRate.toFixed(2)}%
        </p>
      </div>
    </div>
  );
}

function formatNumber(num: number): string {
  if (num >= 1_000_000) return (num / 1_000_000).toFixed(1) + 'M';
  if (num >= 1_000) return (num / 1_000).toFixed(1) + 'K';
  return num.toString();
}

// Mock data generator for development
export function generateMockUsageStats(): ProviderUsageStats {
  return {
    functionCount: Math.floor(Math.random() * 20) + 1,
    requestCount: Math.floor(Math.random() * 100000),
    requestChangePercent: (Math.random() - 0.5) * 40,
    avgLatency: Math.floor(Math.random() * 150) + 20,
    errorRate: Math.random() * 2,
  };
}
