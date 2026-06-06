import { useState } from "react";
import { GlassCard, Badge, Spinner } from "@functionfly/ui-core";
import { Activity, AlertTriangle, Clock, Zap, BarChart3, TrendingUp, TrendingDown, Minus } from "lucide-react";
import { type Plugin, usePluginTelemetry } from "@/hooks/usePlugin";

interface PluginTelemetryPanelProps {
  plugins: Plugin[];
}

type TimeRange = "7d" | "30d" | "90d";

function formatNumber(n: number) {
  if (n >= 1000000) return (n / 1000000).toFixed(1) + "M";
  if (n >= 1000) return (n / 1000).toFixed(1) + "k";
  return n.toString();
}

function formatDuration(seconds: number) {
  if (seconds >= 3600) return (seconds / 3600).toFixed(1) + "h";
  if (seconds >= 60) return (seconds / 60).toFixed(1) + "m";
  return seconds + "s";
}

function getTrendIcon(trend: "up" | "down" | "stable") {
  if (trend === "up") return <TrendingUp className="w-3 h-3 text-red-400" />;
  if (trend === "down") return <TrendingDown className="w-3 h-3 text-green-400" />;
  return <Minus className="w-3 h-3 text-yellow-400" />;
}

function getTrendPercentage(current: number, previous: number) {
  if (previous === 0) return 0;
  return ((current - previous) / previous) * 100;
}

export function PluginTelemetryPanel({ plugins }: PluginTelemetryPanelProps) {
  const [selectedPlugin, setSelectedPlugin] = useState<Plugin | null>(plugins[0] || null);
  const [timeRange, setTimeRange] = useState<TimeRange>("7d");

  const { data: telemetryData, isLoading } = usePluginTelemetry(
    selectedPlugin?.id || "",
    timeRange
  );

  const telemetry = telemetryData?.telemetry;

  const timeRanges: { value: TimeRange; label: string }[] = [
    { value: "7d", label: "7 Days" },
    { value: "30d", label: "30 Days" },
    { value: "90d", label: "90 Days" },
  ];

  return (
    <div className="space-y-4">
      <div className="flex gap-4">
        <div className="w-64 space-y-2">
          <h3 className="text-sm font-medium text-white/60">Select Plugin</h3>
          {plugins.length === 0 ? (
            <p className="text-sm text-white/40">No plugins installed</p>
          ) : (
            plugins.map((plugin) => (
              <button
                key={plugin.id}
                onClick={() => setSelectedPlugin(plugin)}
                className={`w-full text-left p-3 rounded-lg border transition-colors ${
                  selectedPlugin?.id === plugin.id
                    ? "bg-white/10 border-white/20"
                    : "bg-white/5 border-white/10 hover:bg-white/10"
                }`}
              >
                <div className="font-medium text-white text-sm">{plugin.name}</div>
                <div className="text-xs text-white/60">v{plugin.version}</div>
              </button>
            ))
          )}
        </div>

        <div className="flex-1">
          {selectedPlugin && isLoading ? (
            <div className="flex items-center justify-center h-64">
              <Spinner className="w-8 h-8" />
            </div>
          ) : selectedPlugin && telemetry ? (
            <GlassCard className="p-4 space-y-4">
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-2">
                  <BarChart3 className="w-5 h-5 text-white/60" />
                  <h3 className="font-medium text-white">Telemetry</h3>
                </div>
                <div className="flex items-center gap-1 p-0.5 rounded-lg bg-white/5 border border-white/10">
                  {timeRanges.map((range) => (
                    <button
                      key={range.value}
                      onClick={() => setTimeRange(range.value)}
                      className={`px-3 py-1 rounded text-xs font-medium transition-colors ${
                        timeRange === range.value
                          ? "bg-white/10 text-white"
                          : "text-white/60 hover:text-white"
                      }`}
                    >
                      {range.label}
                    </button>
                  ))}
                </div>
              </div>

              <div className="grid grid-cols-4 gap-4">
                <div className="p-4 bg-white/5 rounded-lg">
                  <div className="flex items-center gap-2 text-white/60 text-sm mb-2">
                    <Zap className="w-4 h-4" /> Executions
                    {getTrendIcon(telemetry.executions_trend)}
                  </div>
                  <div className="text-2xl font-bold text-white">{formatNumber(telemetry.executions)}</div>
                  <div className={`text-xs mt-1 flex items-center gap-1 ${
                    telemetry.executions_trend === "up" ? "text-green-400" :
                    telemetry.executions_trend === "down" ? "text-red-400" : "text-white/40"
                  }`}>
                    {telemetry.executions_trend === "up" ? "+" : telemetry.executions_trend === "down" ? "-" : ""}
                    {getTrendPercentage(telemetry.executions, telemetry.previous_executions).toFixed(1)}% from last period
                  </div>
                </div>
                <div className="p-4 bg-white/5 rounded-lg">
                  <div className="flex items-center gap-2 text-white/60 text-sm mb-2">
                    <AlertTriangle className="w-4 h-4" /> Errors
                  </div>
                  <div className="text-2xl font-bold text-white">{telemetry.errors}</div>
                  <div className="text-xs text-white/40 mt-1">{telemetry.error_rate.toFixed(2)}% error rate</div>
                </div>
                <div className="p-4 bg-white/5 rounded-lg">
                  <div className="flex items-center gap-2 text-white/60 text-sm mb-2">
                    <Clock className="w-4 h-4" /> Avg Latency
                    {getTrendIcon(telemetry.latency_trend)}
                  </div>
                  <div className="text-2xl font-bold text-white">{Math.round(telemetry.avg_latency_ms)}ms</div>
                  <div className="text-xs text-white/40 mt-1">per execution</div>
                </div>
                <div className="p-4 bg-white/5 rounded-lg">
                  <div className="flex items-center gap-2 text-white/60 text-sm mb-2">
                    <Activity className="w-4 h-4" /> CPU Time
                  </div>
                  <div className="text-2xl font-bold text-white">{formatDuration(telemetry.cpu_usage_seconds)}</div>
                  <div className="text-xs text-white/40 mt-1">total usage</div>
                </div>
              </div>

              <div className="p-8 border-2 border-dashed border-white/10 rounded-lg text-center">
                <BarChart3 className="w-12 h-12 text-white/20 mx-auto mb-3" />
                <p className="text-white/60">Execution trend chart</p>
                <p className="text-xs text-white/40 mt-1">
                  Visualize your plugin's performance over time
                </p>
              </div>
            </GlassCard>
          ) : (
            <div className="flex items-center justify-center h-64 text-white/40">
              Select a plugin to view telemetry
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
