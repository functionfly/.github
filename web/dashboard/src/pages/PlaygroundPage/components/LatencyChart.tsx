import {
  ResponsiveContainer,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
} from 'recharts';
import { useTranslation } from 'react-i18next';
import { usePlaygroundState } from '../hooks/usePlaygroundState';

interface LatencyChartProps {
  className?: string;
}

interface CustomTooltipProps {
  active?: boolean;
  payload?: Array<{ value: number; payload: { ok: boolean; timestamp: number } }>;
  label?: string;
}

function CustomTooltip({ active, payload }: CustomTooltipProps) {
  const { t } = useTranslation();
  if (!active || !payload || payload.length === 0) return null;

  const item = payload[0];
  const date = new Date(item.payload.timestamp);

  return (
    <div className="bg-bg-secondary border border-border-subtle rounded-md p-2 text-xs shadow-lg">
      <p className="font-medium text-text-primary">{item.value}ms</p>
      <p className="text-text-muted">
        {date.toLocaleTimeString()}
      </p>
      <p className={item.payload.ok ? 'text-green-400' : 'text-red-400'}>
        {item.payload.ok ? t('playground.success') : t('playground.error')}
      </p>
    </div>
  );
}

export function LatencyChart({ className }: LatencyChartProps) {
  const { t } = useTranslation();
  const { latencyHistory, averageLatency } = usePlaygroundState();

  if (latencyHistory.length < 2) {
    return (
      <div className="flex items-center justify-center h-16 text-xs text-text-muted">
        {t('playground.runFewTimesForLatency')}
      </div>
    );
  }

  return (
    <div className={className}>
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs text-text-muted">{t('playground.latencyHistory', { count: latencyHistory.length })}</span>
        {averageLatency !== null && (
          <span className="text-xs text-text-secondary">
            {t('playground.avg')} <span className="font-mono text-indigo-400">{averageLatency}ms</span>
          </span>
        )}
      </div>
      <ResponsiveContainer width="100%" height={80}>
        <AreaChart data={latencyHistory} margin={{ top: 4, right: 4, bottom: 0, left: 0 }}>
          <defs>
            <linearGradient id="latencyGradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="var(--color-brand-500)" stopOpacity={0.3} />
              <stop offset="95%" stopColor="var(--color-brand-500)" stopOpacity={0} />
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" className="stroke-border-subtle" />
          <XAxis dataKey="index" hide />
          <YAxis hide />
          <Tooltip content={<CustomTooltip />} />
          <Area
            type="monotone"
            dataKey="duration_ms"
            stroke="var(--color-brand-500)"
            strokeWidth={1.5}
            fill="url(#latencyGradient)"
            isAnimationActive={false}
            dot={(props) => {
              const { cx, cy, payload } = props;
              return (
                <circle
                  key={`dot-${cx}-${cy}`}
                  cx={cx}
                  cy={cy}
                  r={3}
                  fill={payload.ok ? 'var(--color-success)' : 'var(--color-error)'}
                  stroke="none"
                />
              );
            }}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
