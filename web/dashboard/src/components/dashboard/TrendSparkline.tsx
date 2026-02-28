import {
  LineChart,
  Line,
  ResponsiveContainer,
  AreaChart,
  Area,
  YAxis,
} from "recharts";
import { cn } from "@/lib/utils";

export type TrendDirection = "up" | "down" | "neutral";

export interface TrendSparklineProps {
  data: number[];
  trend?: TrendDirection;
  /** Use area fill under the line */
  variant?: "line" | "area";
  className?: string;
}

const trendColors: Record<TrendDirection, string> = {
  up: "var(--color-success)",
  down: "var(--color-error)",
  neutral: "var(--color-text-muted)",
};

export function TrendSparkline({
  data,
  trend = "neutral",
  variant = "area",
  className,
}: TrendSparklineProps) {
  const color = trendColors[trend];
  const chartData = data.map((value, index) => ({ value, index }));

  return (
    <div className={cn("flex items-center", className)}>
      <ResponsiveContainer width="100%" height="100%">
        {variant === "area" ? (
          <AreaChart
            data={chartData}
            margin={{ top: 2, right: 0, left: 0, bottom: 2 }}
          >
            <defs>
              <linearGradient id={`sparkline-fill-${trend}`} x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor={color} stopOpacity={0.35} />
                <stop offset="100%" stopColor={color} stopOpacity={0} />
              </linearGradient>
            </defs>
            <YAxis hide domain={["auto", "auto"]} />
            <Area
              type="monotone"
              dataKey="value"
              stroke={color}
              strokeWidth={1.5}
              fill={`url(#sparkline-fill-${trend})`}
              dot={false}
              activeDot={false}
              isAnimationActive={true}
              animationDuration={400}
            />
          </AreaChart>
        ) : (
          <LineChart
            data={chartData}
            margin={{ top: 2, right: 0, left: 0, bottom: 2 }}
          >
            <Line
              type="monotone"
              dataKey="value"
              stroke={color}
              strokeWidth={1.5}
              dot={false}
              activeDot={false}
              isAnimationActive={true}
              animationDuration={400}
            />
          </LineChart>
        )}
      </ResponsiveContainer>
    </div>
  );
}
