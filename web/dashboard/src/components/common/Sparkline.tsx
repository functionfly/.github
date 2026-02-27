import { LineChart, Line, ResponsiveContainer } from "recharts";
import { cn } from "@/lib/utils";

interface SparklineProps {
  data: number[];
  width?: number;
  height?: number;
  color?: string;
  className?: string;
  strokeWidth?: number;
}

export function Sparkline({
  data,
  width = 120,
  height = 40,
  color = "#10b981",
  className,
  strokeWidth = 2,
}: SparklineProps) {
  // Convert array of numbers to recharts format
  const chartData = data.map((value, index) => ({
    value,
    index,
  }));

  return (
    <div className={cn("flex items-center", className)}>
      <ResponsiveContainer width={width} height={height}>
        <LineChart data={chartData} margin={{ top: 0, right: 0, left: 0, bottom: 0 }}>
          <Line
            type="monotone"
            dataKey="value"
            stroke={color}
            strokeWidth={strokeWidth}
            dot={false}
            activeDot={false}
            isAnimationActive={false}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}