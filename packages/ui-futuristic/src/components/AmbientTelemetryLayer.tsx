/**
 * @functionfly/ui-futuristic
 * AmbientTelemetryLayer - Ambient background telemetry display
 */

import React, { useState, useEffect } from "react";
import { cn } from "@functionfly/ui-core";
import { TrendingUp, TrendingDown, Minus, Radio } from "lucide-react";
import type { AmbientTelemetryLayerProps, TelemetryMetric, TelemetryTrend } from "../types";

export const AmbientTelemetryLayer: React.FC<AmbientTelemetryLayerProps> = ({
  metrics = [
    { label: "CPU", value: 67, unit: "%", trend: "up" },
    { label: "Memory", value: 45, unit: "%", trend: "stable" },
    { label: "Network", value: 128, unit: "MB/s", trend: "down" },
    { label: "Requests", value: 2847, unit: "/min", trend: "up" },
    { label: "Latency", value: 23, unit: "ms", trend: "down" },
    { label: "Error Rate", value: 0.12, unit: "%", trend: "stable" },
  ],
  isActive = true,
  opacity = 0.7,
  onMetricClick,
  className,
}) => {
  const [pulseState, setPulseState] = useState(0);

  useEffect(() => {
    if (!isActive) return;
    const interval = setInterval(() => {
      setPulseState((prev) => (prev + 1) % 360);
    }, 50);
    return () => clearInterval(interval);
  }, [isActive]);

  const getTrendIcon = (trend: TelemetryTrend) => {
    switch (trend) {
      case "up":
        return <TrendingUp className="w-3 h-3 text-green-400" />;
      case "down":
        return <TrendingDown className="w-3 h-3 text-cyan-400" />;
      case "stable":
        return <Minus className="w-3 h-3 text-slate-400" />;
    }
  };

  const getTrendColor = (value: number, trend: TelemetryTrend) => {
    if (trend === "stable") return "text-slate-300";
    if (value > 80) return trend === "up" ? "text-red-400" : "text-green-400";
    if (value > 60) return trend === "up" ? "text-amber-400" : "text-green-400";
    return "text-cyan-400";
  };

  return (
    <div
      className={cn(
        "relative p-6 rounded-2xl",
        "bg-gradient-to-br from-slate-900/95 via-slate-800/90 to-slate-900/95",
        "backdrop-blur-xl",
        "border border-slate-700/30",
        "overflow-hidden",
        className,
      )}
      style={{ opacity }}
    >
      {/* Ambient effects */}
      <div className="absolute inset-0 pointer-events-none">
        <div
          className="absolute -top-20 -right-20 w-40 h-40 rounded-full bg-cyan-500/10 blur-3xl"
          style={{
            transform: `scale(${1 + Math.sin(pulseState * 0.02) * 0.3})`,
            transition: "transform 0.5s ease-out",
          }}
        />
        <div
          className="absolute -bottom-20 -left-20 w-40 h-40 rounded-full bg-purple-500/10 blur-3xl"
          style={{
            transform: `scale(${1 + Math.cos(pulseState * 0.02) * 0.3})`,
            transition: "transform 0.5s ease-out",
          }}
        />

        {/* Scan line effect */}
        <div className="absolute inset-0 pointer-events-none overflow-hidden">
          <div
            className="absolute left-0 right-0 h-px bg-gradient-to-r from-transparent via-cyan-500/50 to-transparent"
            style={{
              top: `${pulseState % 100}%`,
              animation: "scan-line 3s linear infinite",
            }}
          />
        </div>
      </div>

      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <Radio className="w-5 h-5 text-cyan-400" />
          <span className="text-sm text-cyan-300 font-mono tracking-wider">
            AMBIENT TELEMETRY
          </span>
        </div>
        <div
          className={cn(
            "flex items-center gap-1.5 px-2 py-1 rounded-full",
            "bg-slate-800/80 border border-cyan-500/30",
          )}
        >
          <Radio
            className={cn("w-3 h-3 text-cyan-400", isActive && "animate-pulse")}
          />
          <span className="text-[10px] text-cyan-300">
            {isActive ? "LIVE" : "OFFLINE"}
          </span>
        </div>
      </div>

      {/* Metrics grid */}
      <div className="grid grid-cols-3 gap-3">
        {metrics.map((metric: TelemetryMetric, index: number) => (
          <div
            key={metric.label}
            onClick={() => onMetricClick?.(metric)}
            className={cn(
              "group relative p-4 rounded-xl",
              "bg-slate-800/50 backdrop-blur-sm",
              "border border-slate-700/50",
              "hover:border-cyan-500/50 hover:bg-slate-800/80",
              "transition-all duration-300 cursor-pointer",
              "hover:shadow-[0_0_20px_rgba(6,182,212,0.2)]",
            )}
            style={{
              animationDelay: `${index * 100}ms`,
            }}
          >
            {/* Trend indicator */}
            <div className="absolute top-2 right-2">
              {getTrendIcon(metric.trend)}
            </div>

            {/* Label */}
            <div className="text-[10px] text-slate-500 mb-1">
              {metric.label}
            </div>

            {/* Value */}
            <div className="flex items-baseline gap-1">
              <span
                className={cn(
                  "text-2xl font-bold font-mono",
                  getTrendColor(metric.value, metric.trend),
                )}
              >
                {metric.value.toLocaleString()}
              </span>
              <span className="text-xs text-slate-500">{metric.unit}</span>
            </div>

            {/* Progress bar */}
            <div className="mt-2 h-1 bg-slate-700/50 rounded-full overflow-hidden">
              <div
                className={cn(
                  "h-full rounded-full transition-all duration-500",
                  metric.value > 80
                    ? "bg-red-500"
                    : metric.value > 60
                      ? "bg-amber-500"
                      : "bg-cyan-500",
                )}
                style={{
                  width: `${Math.min(metric.value, 100)}%`,
                  boxShadow: `0 0 8px ${
                    metric.value > 80
                      ? "rgba(239,68,68,0.5)"
                      : metric.value > 60
                        ? "rgba(245,158,11,0.5)"
                        : "rgba(6,182,212,0.5)"
                  }`,
                }}
              />
            </div>
          </div>
        ))}
      </div>

      <style>{`
        @keyframes scan-line {
          0% { transform: translateY(-100%); }
          100% { transform: translateY(400%); }
        }
      `}</style>
    </div>
  );
};
