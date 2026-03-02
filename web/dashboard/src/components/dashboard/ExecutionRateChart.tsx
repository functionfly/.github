import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
} from "recharts";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

export interface ExecutionRateDataPoint {
  time: string;
  rate: number;
  /** Optional threshold line value */
  threshold?: number;
}

export interface ExecutionRateChartProps {
  data: ExecutionRateDataPoint[];
  title?: string;
  /** Unit label in tooltip, e.g. "exec/s" */
  unit?: string;
  /** Optional target or limit line */
  targetRate?: number;
  className?: string;
}

export function ExecutionRateChart({
  data,
  title = "Execution rate",
  unit = "exec/s",
  targetRate,
  className,
}: ExecutionRateChartProps) {
  const strokeColor = "var(--color-brand-500)";

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
            <LineChart
              data={data}
              margin={{ top: 8, right: 8, left: 0, bottom: 0 }}
            >
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
                formatter={(value: number) => [value, unit]}
                labelFormatter={(label) => (typeof label === "string" ? label : "")}
              />
              {targetRate != null && (
                <ReferenceLine
                  y={targetRate}
                  stroke="var(--color-warning)"
                  strokeDasharray="4 4"
                  strokeWidth={1}
                />
              )}
              <Line
                type="monotone"
                dataKey="rate"
                stroke={strokeColor}
                strokeWidth={2}
                dot={false}
                activeDot={{ r: 4, fill: strokeColor }}
                isAnimationActive={true}
                animationDuration={500}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  );
}
