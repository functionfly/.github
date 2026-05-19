import { useState, useEffect } from "react";
import { GlassCard, Badge, Button } from "@functionfly/ui-core";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Progress } from "@/components/ui/progress";
import { cn } from "@/lib/utils";
import {
  Activity, Cpu, MemoryStick, Gauge, Zap, Clock, TrendingUp,
  AlertTriangle, CheckCircle2, XCircle, RefreshCw, Pause, Play
} from "lucide-react";

interface MetricData {
  name: string;
  value: number;
  unit: string;
  max: number;
  color: string;
  history: number[];
}

interface ProcessInfo {
  id: string;
  name: string;
  cpu: number;
  memory: number;
  status: "running" | "paused" | "stopped";
}

export function StudioPerformanceProfiler() {
  const [activeTab, setActiveTab] = useState("overview");
  const [isProfiling, setIsProfiling] = useState(true);
  const [metrics, setMetrics] = useState<Record<string, MetricData>>({
    cpu: {
      name: "CPU Usage",
      value: 42,
      unit: "%",
      max: 100,
      color: "#f97316",
      history: [35, 40, 38, 45, 42, 48, 44, 42, 40, 45, 38, 42],
    },
    memory: {
      name: "Memory",
      value: 2.4,
      unit: "GB",
      max: 16,
      color: "#3b82f6",
      history: [2.1, 2.2, 2.3, 2.4, 2.2, 2.5, 2.4, 2.3, 2.4, 2.5, 2.3, 2.4],
    },
    gpu: {
      name: "GPU",
      value: 67,
      unit: "%",
      max: 100,
      color: "#22c55e",
      history: [60, 65, 62, 68, 70, 65, 67, 69, 68, 65, 66, 67],
    },
    latency: {
      name: "Render Latency",
      value: 12,
      unit: "ms",
      max: 100,
      color: "#8b5cf6",
      history: [15, 12, 14, 11, 13, 12, 10, 14, 12, 11, 13, 12],
    },
  });

  const [processes] = useState<ProcessInfo[]>([
    { id: "p1", name: "Graph Renderer", cpu: 15.2, memory: 340, status: "running" },
    { id: "p2", name: "Plugin Host", cpu: 8.4, memory: 512, status: "running" },
    { id: "p3", name: "Code Editor", cpu: 12.1, memory: 890, status: "running" },
    { id: "p4", name: "AI Engine", cpu: 28.5, memory: 1200, status: "running" },
    { id: "p5", name: "State Manager", cpu: 3.2, memory: 256, status: "running" },
  ]);

  useEffect(() => {
    if (!isProfiling) return;
    const interval = setInterval(() => {
      setMetrics((prev) => {
        const updated = { ...prev };
        Object.keys(updated).forEach((key) => {
          const m = updated[key];
          const variance = m.max * 0.1;
          const newValue = Math.max(0, Math.min(m.max, m.value + (Math.random() - 0.5) * variance));
          updated[key] = {
            ...m,
            value: Number(newValue.toFixed(key === "memory" ? 1 : 0)),
            history: [...m.history.slice(1), Number(newValue.toFixed(key === "memory" ? 1 : 0))],
          };
        });
        return updated;
      });
    }, 1000);
    return () => clearInterval(interval);
  }, [isProfiling]);

  const renderSparkline = (history: number[], color: string) => {
    const width = 120;
    const height = 32;
    const max = Math.max(...history);
    const min = Math.min(...history);
    const range = max - min || 1;
    const points = history
      .map((v, i) => {
        const x = (i / (history.length - 1)) * width;
        const y = height - ((v - min) / range) * height;
        return `${x},${y}`;
      })
      .join(" ");
    return (
      <svg width={width} height={height} className="overflow-visible">
        <polyline
          points={points}
          fill="none"
          stroke={color}
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    );
  };

  const overallHealth = Object.values(metrics).reduce((acc, m) => {
    const healthPercent = 100 - (m.value / m.max) * 100;
    return acc + healthPercent;
  }, 0) / Object.keys(metrics).length;

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between p-5 border-b border-white/10">
        <div className="flex items-center gap-3">
          <div
            className={cn(
              "w-3 h-3 rounded-full",
              isProfiling ? "bg-emerald-400 animate-pulse" : "bg-yellow-400"
            )}
          />
          <h2 className="text-xl font-semibold text-white">Performance Profiler</h2>
          {isProfiling && (
            <Badge variant="outline" className="text-emerald-400 border-emerald-400/30">
              <Activity className="w-3 h-3 mr-1" />
              Live
            </Badge>
          )}
        </div>
        <div className="flex items-center gap-3">
          <Button
            size="sm"
            variant="outline"
            onClick={() => setIsProfiling(!isProfiling)}
            className="gap-2"
          >
            {isProfiling ? (
              <>
                <Pause className="w-4 h-4" />
                Pause
              </>
            ) : (
              <>
                <Play className="w-4 h-4" />
                Resume
              </>
            )}
          </Button>
          <Button size="sm" variant="outline" className="gap-2">
            <RefreshCw className="w-4 h-4" />
            Refresh
          </Button>
        </div>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col">
        <div className="px-5 pt-5">
          <TabsList className="inline-flex h-auto flex-wrap gap-1 rounded-xl border border-white/10 bg-white/5 p-1.5 text-white/60">
            <TabsTrigger
              value="overview"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <Gauge className="h-4 w-4 shrink-0" />
              Overview
            </TabsTrigger>
            <TabsTrigger
              value="cpu"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <Cpu className="h-4 w-4 shrink-0" />
              CPU
            </TabsTrigger>
            <TabsTrigger
              value="memory"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <MemoryStick className="h-4 w-4 shrink-0" />
              Memory
            </TabsTrigger>
            <TabsTrigger
              value="processes"
              className="gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-all duration-200 data-[state=active]:text-white data-[state=active]:bg-white/10"
            >
              <Activity className="h-4 w-4 shrink-0" />
              Processes
            </TabsTrigger>
          </TabsList>
        </div>

        <div className="flex-1 overflow-auto p-4">
          <TabsContent value="overview" className="mt-0">
            <div className="space-y-6">
              <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
                {Object.entries(metrics).map(([key, metric]) => (
                  <GlassCard key={key} className="p-4">
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-sm text-white/60">{metric.name}</span>
                      <div
                        className="w-8 h-8 rounded-lg flex items-center justify-center"
                        style={{ backgroundColor: `${metric.color}20` }}
                      >
                        {key === "cpu" && <Cpu className="w-4 h-4" style={{ color: metric.color }} />}
                        {key === "memory" && <MemoryStick className="w-4 h-4" style={{ color: metric.color }} />}
                        {key === "gpu" && <Zap className="w-4 h-4" style={{ color: metric.color }} />}
                        {key === "latency" && <Clock className="w-4 h-4" style={{ color: metric.color }} />}
                      </div>
                    </div>
                    <div className="flex items-end justify-between">
                      <div>
                        <p className="text-2xl font-bold text-white">
                          {metric.value}
                          <span className="text-sm font-normal text-white/60 ml-1">{metric.unit}</span>
                        </p>
                        <p className="text-xs text-white/40 mt-1">
                          of {metric.max} {metric.unit}
                        </p>
                      </div>
                      {renderSparkline(metric.history, metric.color)}
                    </div>
                    <Progress
                      value={(metric.value / metric.max) * 100}
                      className="h-1.5 mt-3"
                      style={{ ["--progress-color" as string]: metric.color }}
                    />
                  </GlassCard>
                ))}
              </div>

              <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                <GlassCard className="p-5">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="font-semibold text-white">Studio Health</h3>
                    <Badge
                      variant="outline"
                      className={cn(
                        overallHealth > 70
                          ? "text-emerald-400 border-emerald-400/30"
                          : overallHealth > 40
                          ? "text-yellow-400 border-yellow-400/30"
                          : "text-red-400 border-red-400/30"
                      )}
                    >
                      {overallHealth > 70 ? (
                        <>
                          <CheckCircle2 className="w-3 h-3 mr-1" />
                          Healthy
                        </>
                      ) : overallHealth > 40 ? (
                        <>
                          <AlertTriangle className="w-3 h-3 mr-1" />
                          Warning
                        </>
                      ) : (
                        <>
                          <XCircle className="w-3 h-3 mr-1" />
                          Critical
                        </>
                      )}
                    </Badge>
                  </div>
                  <div className="relative w-32 h-32 mx-auto">
                    <svg className="w-full h-full -rotate-90">
                      <circle
                        cx="64"
                        cy="64"
                        r="56"
                        fill="none"
                        stroke="rgba(255,255,255,0.1)"
                        strokeWidth="12"
                      />
                      <circle
                        cx="64"
                        cy="64"
                        r="56"
                        fill="none"
                        stroke={overallHealth > 70 ? "#22c55e" : overallHealth > 40 ? "#eab308" : "#ef4444"}
                        strokeWidth="12"
                        strokeLinecap="round"
                        strokeDasharray={`${(overallHealth / 100) * 352} 352`}
                      />
                    </svg>
                    <div className="absolute inset-0 flex items-center justify-center">
                      <span className="text-3xl font-bold text-white">{Math.round(overallHealth)}%</span>
                    </div>
                  </div>
                </GlassCard>

                <GlassCard className="p-5">
                  <h3 className="font-semibold text-white mb-4">Recent Events</h3>
                  <div className="space-y-3">
                    {[
                      { time: "2s ago", event: "Graph render completed", type: "success" },
                      { time: "15s ago", event: "Plugin memory spike detected", type: "warning" },
                      { time: "1m ago", event: "Auto-save triggered", type: "info" },
                      { time: "3m ago", event: "Cache cleared successfully", type: "success" },
                    ].map((item, i) => (
                      <div key={i} className="flex items-center gap-3 text-sm">
                        <div
                          className={cn(
                            "w-2 h-2 rounded-full",
                            item.type === "success" && "bg-emerald-400",
                            item.type === "warning" && "bg-yellow-400",
                            item.type === "info" && "bg-blue-400"
                          )}
                        />
                        <span className="text-white/60">{item.time}</span>
                        <span className="text-white/80 flex-1">{item.event}</span>
                      </div>
                    ))}
                  </div>
                </GlassCard>
              </div>
            </div>
          </TabsContent>

          <TabsContent value="cpu" className="mt-0">
            <GlassCard className="p-6">
              <h3 className="font-semibold text-white mb-4">CPU Usage Over Time</h3>
              <div className="h-48 flex items-center justify-center">
                {renderSparkline(metrics.cpu.history, metrics.cpu.color)}
              </div>
              <div className="grid grid-cols-3 gap-4 mt-6">
                <div className="p-4 rounded-lg bg-white/5">
                  <p className="text-xs text-white/60">Average</p>
                  <p className="text-xl font-bold text-white">
                    {(metrics.cpu.history.reduce((a, b) => a + b, 0) / metrics.cpu.history.length).toFixed(1)}%
                  </p>
                </div>
                <div className="p-4 rounded-lg bg-white/5">
                  <p className="text-xs text-white/60">Peak</p>
                  <p className="text-xl font-bold text-white">{Math.max(...metrics.cpu.history)}%</p>
                </div>
                <div className="p-4 rounded-lg bg-white/5">
                  <p className="text-xs text-white/60">Current</p>
                  <p className="text-xl font-bold text-white">{metrics.cpu.value}%</p>
                </div>
              </div>
            </GlassCard>
          </TabsContent>

          <TabsContent value="memory" className="mt-0">
            <GlassCard className="p-6">
              <h3 className="font-semibold text-white mb-4">Memory Usage Over Time</h3>
              <div className="h-48 flex items-center justify-center">
                {renderSparkline(
                  metrics.memory.history.map((v) => v * 100),
                  metrics.memory.color
                )}
              </div>
              <div className="space-y-3 mt-6">
                {processes.map((proc) => (
                  <div key={proc.id} className="flex items-center gap-4">
                    <span className="text-sm text-white/80 w-32 truncate">{proc.name}</span>
                    <Progress value={(proc.memory / 2048) * 100} className="flex-1 h-2" />
                    <span className="text-sm text-white/60 w-16 text-right">{proc.memory} MB</span>
                  </div>
                ))}
              </div>
            </GlassCard>
          </TabsContent>

          <TabsContent value="processes" className="mt-0">
            <div className="space-y-3">
              {processes.map((proc) => (
                <GlassCard key={proc.id} className="p-4">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-4">
                      <div
                        className={cn(
                          "w-3 h-3 rounded-full",
                          proc.status === "running" && "bg-emerald-400",
                          proc.status === "paused" && "bg-yellow-400",
                          proc.status === "stopped" && "bg-red-400"
                        )}
                      />
                      <span className="font-medium text-white">{proc.name}</span>
                    </div>
                    <div className="flex items-center gap-6">
                      <div className="text-right">
                        <p className="text-xs text-white/60">CPU</p>
                        <p className="text-sm font-medium text-white">{proc.cpu}%</p>
                      </div>
                      <div className="text-right">
                        <p className="text-xs text-white/60">Memory</p>
                        <p className="text-sm font-medium text-white">{proc.memory} MB</p>
                      </div>
                      <Badge
                        variant="outline"
                        className={cn(
                          proc.status === "running" && "text-emerald-400 border-emerald-400/30",
                          proc.status === "paused" && "text-yellow-400 border-yellow-400/30",
                          proc.status === "stopped" && "text-red-400 border-red-400/30"
                        )}
                      >
                        {proc.status}
                      </Badge>
                    </div>
                  </div>
                </GlassCard>
              ))}
            </div>
          </TabsContent>
        </div>
      </Tabs>
    </div>
  );
}