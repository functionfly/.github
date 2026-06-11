/**
 * @functionfly/ui-futuristic
 * GlassExecutionCard - Glass-morphism execution card
 */

import React, { useState, useEffect } from "react";
import { cn } from "@functionfly/ui-core";
import { CheckCircle, XCircle, Clock, SquareIcon } from "./icons";
import { Play } from "lucide-react";
import type { GlassExecutionCardProps, ExecutionData } from "../types";

export const GlassExecutionCard: React.FC<GlassExecutionCardProps> = ({
  execution,
  onClick,
  onCancel,
  className,
}) => {
  const [isHovered, setIsHovered] = useState(false);
  const [timeLeft, setTimeLeft] = useState(execution.duration || 0);

  useEffect(() => {
    if (execution.status !== "running") return;

    const interval = setInterval(() => {
      setTimeLeft((prev) => Math.max(0, prev - 1));
    }, 1000);

    return () => clearInterval(interval);
  }, [execution.status]);

  const getStatusConfig = (): {
    bg: string;
    border: string;
    text: string;
    icon: React.ReactNode;
    glow: string;
  } => {
    switch (execution.status) {
      case "running":
        return {
          bg: "bg-cyan-500/10",
          border: "border-cyan-500/50",
          text: "text-cyan-400",
          icon: <Play className="w-3 h-3" />,
          glow: "shadow-[0_0_15px_rgba(6,182,212,0.3)]",
        };
      case "completed":
        return {
          bg: "bg-green-500/10",
          border: "border-green-500/50",
          text: "text-green-400",
          icon: <CheckCircle className="w-3 h-3" />,
          glow: "shadow-[0_0_10px_rgba(34,197,94,0.2)]",
        };
      case "failed":
        return {
          bg: "bg-red-500/10",
          border: "border-red-500/50",
          text: "text-red-400",
          icon: <XCircle className="w-3 h-3" />,
          glow: "shadow-[0_0_10px_rgba(239,68,68,0.3)]",
        };
      case "pending":
        return {
          bg: "bg-slate-500/10",
          border: "border-slate-500/50",
          text: "text-slate-400",
          icon: <Clock className="w-3 h-3" />,
          glow: "",
        };
    }
  };

  const statusConfig = getStatusConfig();
  const progress = execution.progress ?? 0;

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  };

  return (
    <div
      className={cn(
        "relative group cursor-pointer",
        "rounded-xl overflow-hidden",
        "bg-gradient-to-br from-slate-800/60 to-slate-900/80",
        "backdrop-blur-xl",
        "border border-slate-700/50",
        "transition-all duration-300",
        "hover:border-cyan-500/50 hover:shadow-[0_0_30px_rgba(6,182,212,0.2)]",
        isHovered && "scale-[1.02]",
        className,
      )}
      onClick={() => onClick?.(execution)}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      {/* Glass reflection */}
      <div className="absolute inset-0 bg-gradient-to-br from-white/5 via-transparent to-transparent pointer-events-none" />

      {/* Status glow */}
      <div
        className={cn(
          "absolute top-0 left-0 right-0 h-1",
          statusConfig.bg.replace("/10", "/30"),
          "transition-all duration-300",
        )}
      />

      <div className="p-4">
        {/* Header */}
        <div className="flex items-start justify-between mb-3">
          <div className="flex items-center gap-3">
            {/* Status indicator */}
            <div
              className={cn(
                "relative flex items-center justify-center w-10 h-10 rounded-lg",
                "bg-slate-800/80 border border-slate-600/50",
                statusConfig.glow,
              )}
            >
              <span className={statusConfig.text}>{statusConfig.icon}</span>

              {/* Running animation */}
              {execution.status === "running" && (
                <div className="absolute inset-0 rounded-lg border-2 border-cyan-400/50 animate-ping" />
              )}
            </div>

            <div>
              <h4 className="text-sm font-medium text-slate-100">
                {execution.name}
              </h4>
              <div className="flex items-center gap-2 mt-0.5">
                <span
                  className={cn("text-[10px] uppercase", statusConfig.text)}
                >
                  {execution.status}
                </span>
                {execution.startTime && (
                  <span className="text-[10px] text-slate-500">
                    {new Date(execution.startTime).toLocaleTimeString()}
                  </span>
                )}
              </div>
            </div>
          </div>

          {/* Actions */}
          {execution.status === "running" && (
            <button
              onClick={(e) => {
                e.stopPropagation();
                onCancel?.(execution.id);
              }}
              className={cn(
                "p-1.5 rounded-lg",
                "bg-red-500/20 border border-red-500/50",
                "text-red-400",
                "hover:bg-red-500/30",
                "transition-colors",
              )}
            >
              <SquareIcon className="w-3 h-3" />
            </button>
          )}
        </div>

        {/* Progress bar */}
        {execution.status === "running" && (
          <div className="mb-3">
            <div className="flex items-center justify-between mb-1">
              <span className="text-[10px] text-slate-400">Progress</span>
              <span className="text-[10px] text-cyan-400">
                {Math.round(progress * 100)}%
              </span>
            </div>
            <div className="h-1.5 bg-slate-800 rounded-full overflow-hidden">
              <div
                className={cn(
                  "h-full rounded-full transition-all duration-300",
                  "bg-gradient-to-r from-cyan-500 to-cyan-400",
                  "shadow-[0_0_10px_rgba(6,182,212,0.5)]",
                )}
                style={{ width: `${progress * 100}%` }}
              />
            </div>
          </div>
        )}

        {/* Duration */}
        {execution.duration && (
          <div className="flex items-center justify-between">
            <span className="text-[10px] text-slate-500">Duration</span>
            <span className={cn("text-xs font-mono", statusConfig.text)}>
              {formatDuration(timeLeft)}
            </span>
          </div>
        )}

        {/* Metadata */}
        {execution.metadata && Object.keys(execution.metadata).length > 0 && (
          <div className="mt-3 pt-3 border-t border-slate-700/50 flex flex-wrap gap-2">
            {Object.entries(execution.metadata)
              .slice(0, 3)
              .map(([key, value]) => (
                <span
                  key={key}
                  className="px-2 py-0.5 rounded text-[10px] bg-slate-700/50 text-slate-400"
                >
                  {key}: {String(value).slice(0, 20)}
                </span>
              ))}
          </div>
        )}
      </div>
    </div>
  );
};
