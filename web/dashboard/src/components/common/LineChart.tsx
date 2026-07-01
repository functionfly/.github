import {
  LineChart as RechartsLineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from "recharts";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

interface LineChartData {
  [key: string]: any;
}

interface LineSeries {
  key: string;
  name: string;
  color: string;
  strokeWidth?: number;
}

interface LineChartProps {
  data: LineChartData[];
  series: LineSeries[];
  title?: string;
  xAxisKey?: string;
  height?: number;
  showGrid?: boolean;
  showLegend?: boolean;
  className?: string;
  tooltipFormatter?: (value: any, name: string) => [string, string];
  yAxisFormatter?: (value: any) => string;
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

export function LineChart({
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
}: LineChartProps) {
  return (
    <Card className={cn("chart-container", className)}>
      {title && (
        <CardHeader>
          <CardTitle className="chart-title">{title}</CardTitle>
        </CardHeader>
      )}
      <CardContent>
        <ResponsiveContainer width="100%" height={height}>
          <RechartsLineChart
            data={data}
            margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
          >
            {showGrid && (
              <CartesianGrid
                strokeDasharray="3 3"
                stroke="var(--panel-edge)"
                opacity={0.5}
              />
            )}
            <XAxis
              dataKey={xAxisKey}
              stroke="var(--text-dim)"
              fontSize={12}
              tickLine={false}
              axisLine={false}
            />
            <YAxis
              stroke="var(--text-dim)"
              fontSize={12}
              tickLine={false}
              axisLine={false}
              tickFormatter={yAxisFormatter}
            />
            <Tooltip
              content={
                <CustomTooltip formatter={tooltipFormatter} />
              }
            />
            {showLegend && <Legend />}
            {series.map((line) => (
              <Line
                key={line.key}
                type="monotone"
                dataKey={line.key}
                stroke={line.color}
                strokeWidth={line.strokeWidth || 2}
                dot={{ fill: line.color, strokeWidth: 0, r: 4 }}
                activeDot={{ r: 6, stroke: line.color, strokeWidth: 2, fill: "var(--bg)" }}
                name={line.name}
                isAnimationActive={false}
              />
            ))}
          </RechartsLineChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}