import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { motion } from "framer-motion";
import { Globe2, MapPin, Server } from "lucide-react";
import {
    Cell,
    Pie,
    PieChart,
    ResponsiveContainer,
    Tooltip,
} from "recharts";

export interface RegionData {
  name: string;
  value: number;
  code: string;
  provider?: string;
}

export interface RegionDistributionWidgetProps {
  regions: RegionData[];
  totalFunctions: number;
  className?: string;
}

const regionColors = [
  "var(--color-aviation-amber)",
  "var(--color-aviation-cyan)",
  "var(--color-success)",
  "var(--color-aviation-amber-light)",
  "var(--color-aviation-cyan-light)",
  "#8b5cf6",
  "#ec4899",
  "#14b8a6",
];

export function RegionDistributionWidget({
  regions,
  totalFunctions,
  className,
}: RegionDistributionWidgetProps) {
  const sortedRegions = [...regions].sort((a, b) => b.value - a.value);
  const topRegions = sortedRegions.slice(0, 5);
  const otherValue = sortedRegions
    .slice(5)
    .reduce((sum, r) => sum + r.value, 0);

  const chartData =
    otherValue > 0
      ? [...topRegions, { name: "Other", value: otherValue, code: "OT" }]
      : topRegions;

  const hasData = regions.length > 0 && totalFunctions > 0;

  return (
    <Card className={cn("overflow-hidden", className)}>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-medium text-text-secondary">
            Regional Distribution
          </CardTitle>
          <div className="flex items-center gap-1.5 text-xs text-text-muted">
            <Globe2 className="w-3.5 h-3.5" />
            <span>{regions.length} regions</span>
          </div>
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        {!hasData ? (
          <div className="flex flex-col items-center justify-center py-8 text-center">
            <div className="w-12 h-12 rounded-full bg-bg-tertiary flex items-center justify-center mb-3">
              <MapPin className="w-6 h-6 text-text-muted" />
            </div>
            <p className="text-sm text-text-secondary">No regions active</p>
            <p className="text-xs text-text-muted mt-1">
              Deploy functions to see regional distribution
            </p>
          </div>
        ) : (
          <div className="flex items-center gap-4">
            <div className="relative w-[140px] h-[140px] shrink-0">
              <ResponsiveContainer width="100%" height="100%" minWidth={100} minHeight={100}>
                <PieChart>
                  <Pie
                    data={chartData}
                    cx="50%"
                    cy="50%"
                    innerRadius={45}
                    outerRadius={65}
                    paddingAngle={2}
                    dataKey="value"
                    stroke="none"
                  >
                    {chartData.map((_, index) => (
                      <Cell
                        key={`cell-${index}`}
                        fill={regionColors[index % regionColors.length]}
                        fillOpacity={0.85}
                      />
                    ))}
                  </Pie>
                  <Tooltip
                    contentStyle={{
                      backgroundColor: "var(--color-bg-secondary)",
                      border: "1px solid var(--color-border)",
                      borderRadius: "8px",
                      fontSize: 12,
                    }}
                    itemStyle={{ fontSize: 12 }}
                  />
                </PieChart>
              </ResponsiveContainer>
              <div className="absolute inset-0 flex flex-col items-center justify-center pointer-events-none">
                <span className="text-2xl font-bold text-text-primary">
                  {totalFunctions}
                </span>
                <span className="text-xs text-text-muted">functions</span>
              </div>
            </div>

            <div className="flex-1 min-w-0 space-y-2">
              {chartData.slice(0, 4).map((region, index) => {
                const percentage = Math.round(
                  (region.value / totalFunctions) * 100
                );
                return (
                  <motion.div
                    key={region.code}
                    initial={{ opacity: 0, x: 10 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: index * 0.1 }}
                    className="flex items-center justify-between"
                  >
                    <div className="flex items-center gap-2">
                      <div
                        className="w-2.5 h-2.5 rounded-full"
                        style={{
                          backgroundColor:
                            regionColors[index % regionColors.length],
                        }}
                      />
                      <span className="text-xs text-text-secondary truncate">
                        {region.name}
                      </span>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-medium text-text-primary tabular-nums">
                        {region.value}
                      </span>
                      <span className="text-xs text-text-muted w-8 text-right">
                        {percentage}%
                      </span>
                    </div>
                  </motion.div>
                );
              })}
            </div>
          </div>
        )}

        {hasData && (
          <div className="mt-4 pt-3 border-t border-border">
            <div className="flex items-center gap-4 text-xs text-text-muted">
              <div className="flex items-center gap-1.5">
                <Server className="w-3.5 h-3.5" />
                <span>Multi-region deployment</span>
              </div>
              <div className="flex items-center gap-1.5">
                <Globe2 className="w-3.5 h-3.5" />
                <span>Edge optimized</span>
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
