import {
  PieChart as RechartsPieChart,
  Pie,
  Cell,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from "recharts";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

interface PieChartData {
  name: string;
  value: number;
  color?: string;
}

interface PieChartProps {
  data: PieChartData[];
  title?: string;
  height?: number;
  showLegend?: boolean;
  showLabels?: boolean;
  innerRadius?: number;
  outerRadius?: number;
  className?: string;
  tooltipFormatter?: (value: any, name: string) => [string, string];
  labelFormatter?: (entry: PieChartData) => string;
}

const COLORS = [
  "#6366f1", // indigo
  "#10b981", // emerald
  "#f59e0b", // amber
  "#ef4444", // red
  "#8b5cf6", // violet
  "#06b6d4", // cyan
  "#84cc16", // lime
  "#f97316", // orange
  "#ec4899", // pink
  "#6b7280", // gray
];

const CustomTooltip = ({ active, payload, formatter }: any) => {
  if (active && payload && payload.length) {
    const data = payload[0];
    const [formattedValue, formattedName] = formatter
      ? formatter(data.value, data.name)
      : [data.value, data.name];

    return (
      <div className="chart-tooltip">
        <p className="chart-tooltip-label">{data.name}</p>
        <p
          className="chart-tooltip-value"
          style={{ color: data.color }}
        >
          {formattedName}: {formattedValue}
        </p>
      </div>
    );
  }
  return null;
};

const CustomLabel = ({ cx, cy, midAngle, innerRadius, outerRadius, percent, index, formatter, data }: any) => {
  if (percent < 0.05) return null; // Don't show labels for slices smaller than 5%

  const RADIAN = Math.PI / 180;
  const radius = innerRadius + (outerRadius - innerRadius) * 0.5;
  const x = cx + radius * Math.cos(-midAngle * RADIAN);
  const y = cy + radius * Math.sin(-midAngle * RADIAN);

  const labelText = formatter ? formatter(data[index]) : `${(percent * 100).toFixed(0)}%`;

  return (
    <text
      x={x}
      y={y}
      fill="var(--text-primary)"
      textAnchor={x > cx ? 'start' : 'end'}
      dominantBaseline="central"
      fontSize={12}
      fontWeight={500}
    >
      {labelText}
    </text>
  );
};

export function PieChart({
  data,
  title,
  height = 300,
  showLegend = true,
  showLabels = false,
  innerRadius = 0,
  outerRadius = 80,
  className,
  tooltipFormatter,
  labelFormatter,
}: PieChartProps) {
  // Ensure each data point has a color
  const dataWithColors = data.map((item, index) => ({
    ...item,
    color: item.color || COLORS[index % COLORS.length],
  }));

  return (
    <Card className={cn("chart-container", className)}>
      {title && (
        <CardHeader>
          <CardTitle className="chart-title">{title}</CardTitle>
        </CardHeader>
      )}
      <CardContent>
        <ResponsiveContainer width="100%" height={height}>
          <RechartsPieChart>
            <Pie
              data={dataWithColors}
              cx="50%"
              cy="50%"
              innerRadius={innerRadius}
              outerRadius={outerRadius}
              paddingAngle={2}
              dataKey="value"
              label={showLabels ? (props) => (
                <CustomLabel {...props} formatter={labelFormatter} data={dataWithColors} />
              ) : false}
            >
              {dataWithColors.map((entry, index) => (
                <Cell key={`cell-${index}`} fill={entry.color} />
              ))}
            </Pie>
            <Tooltip
              content={
                <CustomTooltip formatter={tooltipFormatter} />
              }
            />
            {showLegend && (
              <Legend
                wrapperStyle={{ paddingTop: "20px" }}
                iconType="circle"
              />
            )}
          </RechartsPieChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}