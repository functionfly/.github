export interface HealthDataPoint {
  timestamp: number;
  status: 'online' | 'degraded' | 'offline';
  latency?: number;
}

interface ConnectionHealthSparklineProps {
  data: HealthDataPoint[];
  height?: number;
  width?: number;
}

export function ConnectionHealthSparkline({ data, height = 24, width = 100 }: ConnectionHealthSparklineProps) {
  if (!data || data.length === 0) {
    return (
      <div
        className="flex items-center justify-center bg-bg-secondary/50 rounded"
        style={{ height, width }}
      >
        <span className="text-xs text-text-tertiary">No data</span>
      </div>
    );
  }

  const maxLatency = Math.max(...data.map(d => d.latency || 0), 100);
  const barWidth = width / data.length;
  const gap = 1;
  const effectiveBarWidth = Math.max(barWidth - gap, 2);

  const getBarColor = (status: string, latency?: number) => {
    if (status === 'offline') return '#ef4444'; // red-500
    if (status === 'degraded') return '#f59e0b'; // amber-500
    if ((latency || 0) > maxLatency * 0.8) return '#f59e0b'; // High latency = amber
    return '#10b981'; // emerald-500
  };

  return (
    <div className="flex items-end gap-px" style={{ height, width }}>
      {data.map((point, i) => {
        const barHeight = point.latency
          ? Math.max((point.latency / maxLatency) * height, 2)
          : height * 0.5;
        const color = getBarColor(point.status, point.latency);

        return (
          <div
            key={i}
            className="rounded-sm transition-all duration-200 hover:opacity-80"
            style={{
              width: effectiveBarWidth,
              height: barHeight,
              backgroundColor: color,
              opacity: point.status === 'offline' ? 0.5 : 1,
            }}
            title={`${new Date(point.timestamp).toLocaleTimeString()}: ${point.status}${point.latency ? ` (${point.latency}ms)` : ''}`}
          />
        );
      })}
    </div>
  );
}

interface HealthSparklineCardProps {
  providerId: string;
  data?: HealthDataPoint[];
  last24hUptime?: number;
}

export function HealthSparklineCard({ data, last24hUptime }: HealthSparklineCardProps) {
  return (
    <div className="flex items-center gap-3">
      <ConnectionHealthSparkline data={data || []} />
      {last24hUptime !== undefined && (
        <span className="text-xs font-medium text-text-secondary whitespace-nowrap">
          {last24hUptime.toFixed(1)}% uptime
        </span>
      )}
    </div>
  );
}

// Helper to generate mock health data (for demo/development)
export function generateMockHealthData(hours = 24, online = true): HealthDataPoint[] {
  const data: HealthDataPoint[] = [];
  const now = Date.now();
  const hourMs = 60 * 60 * 1000;

  for (let i = hours; i >= 0; i--) {
    const baseLatency = online ? 50 + Math.random() * 100 : 200 + Math.random() * 300;
    const isDegraded = Math.random() < 0.1; // 10% chance of degraded
    const isOffline = !online && Math.random() < 0.3;

    data.push({
      timestamp: now - i * hourMs,
      status: isOffline ? 'offline' : isDegraded ? 'degraded' : 'online',
      latency: Math.round(baseLatency),
    });
  }

  return data;
}
