import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { BarChart3 } from "lucide-react";
import {
    Area,
    AreaChart,
    CartesianGrid,
    ResponsiveContainer,
    Tooltip,
    XAxis,
    YAxis,
} from "recharts";

export interface UsageGraphDataPoint {
  time: string;
  value: number;
  label?: string;
}

export interface UsageGraphProps {
  data: UsageGraphDataPoint[];
  title?: string;
  valueLabel?: string;
  color?: string;
  className?: string;
}

export function UsageGraph({
  data,
  title = "Usage",
  valueLabel = "Usage",
  color = "var(--color-brand-500)",
  className,
}: UsageGraphProps) {
  // Don't render chart if no data
  if (!data || data.length === 0) {
    return (
      <Card className={cn("card", className)}>
        {title && (
          <CardHeader className="pb-2">
            <CardTitle className="text-base font-semibold text-text-primary">
              {title}
            </CardTitle>
          </CardHeader>
        )}
        <CardContent className={title ? "pt-0" : undefined}>
          <div className="h-[200px] min-h-[200px] w-full min-w-0 flex flex-col items-center justify-center gap-3 text-center">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-bg-secondary border border-border-subtle">
              <BarChart3 className="h-5 w-5 text-text-muted" />
            </div>
            <div>
              <p className="text-sm font-medium text-text-secondary">No usage yet</p>
              <p className="text-xs text-text-muted mt-0.5">Data will appear once your functions start receiving requests.</p>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className={cn("card", className)}>
      {title && (
        <CardHeader className="pb-2">
          <CardTitle className="text-base font-semibold text-text-primary">
            {title}
          </CardTitle>
        </CardHeader>
      )}
      <CardContent className={title ? "pt-0" : undefined}>
        <div className="h-[200px] min-h-[200px] w-full min-w-0">
          <ResponsiveContainer width="100%" height={200} minWidth={300}>
            <AreaChart
              data={data}
              margin={{ top: 8, right: 8, left: 0, bottom: 0 }}
            >
              <defs>
                <linearGradient id="usage-gradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor={color} stopOpacity={0.4} />
                  <stop offset="100%" stopColor={color} stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid
                strokeDasharray="3 3"
                stroke="var(--color-border-subtle)"
                vertical={false}
              />
              <XAxis
                dataKey="time"
                tick={{ fill: "var(--color-text-muted)", fontSize: 11 }}
                tickLine={false}
                axisLine={false}
              />
              <YAxis
                tick={{ fill: "var(--color-text-muted)", fontSize: 11 }}
                tickLine={false}
                axisLine={false}
                tickFormatter={(v) => (v >= 1000 ? `${(v / 1000).toFixed(1)}k` : String(v))}
              />
              <Tooltip
                contentStyle={{
                  backgroundColor: "var(--color-bg-tertiary)",
                  border: "1px solid var(--color-border-default)",
                  borderRadius: "8px",
                }}
                labelStyle={{ color: "var(--color-text-primary)" }}
                formatter={(value: number) => [value, valueLabel]}
                labelFormatter={(label) => (typeof label === "string" ? label : "")}
              />
              <Area
                type="monotone"
                dataKey="value"
                stroke={color}
                strokeWidth={2}
                fill="url(#usage-gradient)"
                isAnimationActive={false}
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  );
}
