import {
  BarChart as RechartsBarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from "recharts";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

interface BarChartData {
  [key: string]: any;
}

interface BarSeries {
  key: string;
  name: string;
  color: string;
  radius?: number | [number, number, number, number];
}

interface BarChartProps {
  data: BarChartData[];
  series: BarSeries[];
  title?: string;
  xAxisKey?: string;
  height?: number;
  showGrid?: boolean;
  showLegend?: boolean;
  className?: string;
  tooltipFormatter?: (value: any, name: string) => [string, string];
  yAxisFormatter?: (value: any) => string;
  layout?: "horizontal" | "vertical";
}

const CustomTooltip = ({ active, payload, label, formatter }: any) => {
  if (active && payload && payload.length) {
    return (
      <div className="chart-tooltip">
        <p className="chart-tooltip-label">{label}</p>
        {payload.map((entry: any, index: number) => {
          const [formattedValue, formattedName] = formatter
            ? formatter(entry.value, entry.name)
            : [entry.value, entry.name];

          return (
            <p
              key={index}
              className="chart-tooltip-value"
              style={{ color: entry.color }}
            >
              {formattedName}: {formattedValue}
            </p>
          );
        })}
      </div>
    );
  }
  return null;
};

export function BarChart({
  data,
  series,
  title,
  xAxisKey = "name",
  height = 300,
  showGrid = true,
  showLegend = true,
  className,
  tooltipFormatter,
  yAxisFormatter,
  layout = "vertical",
}: BarChartProps) {
  const isHorizontal = layout === "horizontal";

  return (
    <Card className={cn("chart-container", className)}>
      {title && (
        <CardHeader>
          <CardTitle className="chart-title">{title}</CardTitle>
        </CardHeader>
      )}
      <CardContent>
        <ResponsiveContainer width="100%" height={height}>
          <RechartsBarChart
            data={data}
            layout={layout}
            margin={{
              top: 5,
              right: 30,
              left: isHorizontal ? 60 : 20,
              bottom: isHorizontal ? 5 : 5
            }}
          >
            {showGrid && (
              <CartesianGrid
                strokeDasharray="3 3"
                stroke="var(--border-subtle)"
                opacity={0.5}
              />
            )}
            <XAxis
              type={isHorizontal ? "number" : "category"}
              dataKey={isHorizontal ? undefined : xAxisKey}
              stroke="var(--text-secondary)"
              fontSize={12}
              tickLine={false}
              axisLine={false}
            />
            <YAxis
              type={isHorizontal ? "category" : "number"}
              dataKey={isHorizontal ? xAxisKey : undefined}
              stroke="var(--text-secondary)"
              fontSize={12}
              tickLine={false}
              axisLine={false}
              tickFormatter={yAxisFormatter}
              width={isHorizontal ? 80 : 40}
            />
            <Tooltip
              content={
                <CustomTooltip formatter={tooltipFormatter} />
              }
            />
            {showLegend && <Legend />}
            {series.map((bar, index) => (
              <Bar
                key={bar.key}
                dataKey={bar.key}
                fill={bar.color}
                name={bar.name}
                radius={bar.radius || [2, 2, 0, 0]}
              />
            ))}
          </RechartsBarChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}