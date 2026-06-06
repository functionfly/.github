import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { cn } from '@/lib/utils';
import { motion } from 'framer-motion';
import { AlertTriangle, CheckCircle, XCircle } from 'lucide-react';
import {
    Bar,
    BarChart,
    CartesianGrid,
    Cell,
    ResponsiveContainer,
    Tooltip,
    XAxis,
    YAxis,
} from 'recharts';

export interface ErrorRateDataPoint {
  time: string;
  success: number;
  error: number;
}

export interface ErrorRateWidgetProps {
  data: ErrorRateDataPoint[];
  className?: string;
}

export function ErrorRateWidget({ data, className }: ErrorRateWidgetProps) {
  const totalSuccess = data.reduce((sum, d) => sum + d.success, 0);
  const totalError = data.reduce((sum, d) => sum + d.error, 0);
  const total = totalSuccess + totalError;
  const errorRate = total > 0 ? (totalError / total) * 100 : 0;

  const getErrorSeverity = (rate: number) => {
    if (rate < 1) return 'healthy';
    if (rate < 5) return 'warning';
    return 'critical';
  };

  const severity = getErrorSeverity(errorRate);

  const severityConfig = {
    healthy: {
      color: 'var(--color-success, #10b981)',
      bgColor: 'bg-[var(--color-success)]/10',
      textColor: 'text-[var(--color-success)]',
      icon: CheckCircle,
      label: 'Healthy',
    },
    warning: {
      color: 'var(--color-aviation-amber, #f59e0b)',
      bgColor: 'bg-[var(--color-aviation-amber)]/10',
      textColor: 'text-[var(--color-aviation-amber)]',
      icon: AlertTriangle,
      label: 'Elevated',
    },
    critical: {
      color: 'var(--color-error, #ef4444)',
      bgColor: 'bg-[var(--color-error)]/10',
      textColor: 'text-[var(--color-error)]',
      icon: XCircle,
      label: 'Critical',
    },
  };

  const config = severityConfig[severity];
  const StatusIcon = config.icon;

  return (
    <Card className={cn('overflow-hidden', className)}>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-medium text-text-secondary">Error Rate</CardTitle>
          <motion.div
            initial={{ scale: 0.8 }}
            animate={{ scale: 1 }}
            className={cn(
              'flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium',
              config.bgColor,
              config.textColor
            )}
          >
            <StatusIcon className="w-3.5 h-3.5" />
            <span>{config.label}</span>
          </motion.div>
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        <div className="flex items-baseline gap-2 mb-3">
          <span className="text-3xl font-bold text-text-primary">{errorRate.toFixed(2)}%</span>
          <span className="text-xs text-text-muted">
            {totalError.toLocaleString()} errors / {total.toLocaleString()} total
          </span>
        </div>

        <div className="h-[120px] min-h-[120px]">
          <ResponsiveContainer width="100%" height={120} minWidth={200}>
            <BarChart data={data} barGap={1} margin={{ top: 0, right: 0, bottom: 0, left: 0 }}>
              <CartesianGrid
                strokeDasharray="3 3"
                stroke="var(--color-border)"
                opacity={0.3}
                vertical={false}
              />
              <XAxis
                dataKey="time"
                tick={{ fill: 'var(--color-text-muted)', fontSize: 10 }}
                tickLine={false}
                axisLine={{ stroke: 'var(--color-border)' }}
                interval="preserveStartEnd"
                minTickGap={24}
              />
              <YAxis
                tick={{ fill: 'var(--color-text-muted)', fontSize: 10 }}
                tickLine={false}
                axisLine={false}
                width={30}
              />
              <Tooltip
                contentStyle={{
                  backgroundColor: 'var(--color-bg-secondary)',
                  border: '1px solid var(--color-border)',
                  borderRadius: '8px',
                  fontSize: 12,
                }}
                labelStyle={{ color: 'var(--color-text-secondary)' }}
                itemStyle={{ fontSize: 12 }}
              />
              <Bar dataKey="success" stackId="a" radius={[0, 0, 2, 2]}>
                {data.map((_, index) => (
                  <Cell
                    key={`success-${index}`}
                    fill="var(--color-success, #10b981)"
                    fillOpacity={0.8}
                  />
                ))}
              </Bar>
              <Bar dataKey="error" stackId="a" radius={[2, 2, 0, 0]}>
                {data.map((_, index) => (
                  <Cell
                    key={`error-${index}`}
                    fill="var(--color-error, #ef4444)"
                    fillOpacity={0.9}
                  />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>

        <div className="flex items-center justify-between mt-2 text-xs">
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 rounded-full bg-(--color-success)" />
            <span className="text-text-muted">{totalSuccess.toLocaleString()} success</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 rounded-full bg-(--color-error)" />
            <span className="text-text-muted">{totalError.toLocaleString()} errors</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
