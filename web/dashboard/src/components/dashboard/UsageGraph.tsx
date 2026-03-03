import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

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
  return (
    <Card className={cn("border-theme bg-card", className)}>
      {title && (
        <CardHeader className="pb-2">
          <CardTitle className="text-base font-semibold text-text-primary">
            {title}
          </CardTitle>
        </CardHeader>
      )}
      <CardContent className={title ? "pt-0" : undefined}>
        <div className="h-[200px] min-h-[200px] w-full min-w-0">
          <ResponsiveContainer width="100%" height="100%" minHeight={200}>
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
