/**
 * @functionfly/ui-ai
 * Execution Narrator Component - Step-by-step execution timeline
 */

import * as React from "react";
import { cn } from "@functionfly/ui-core";
import { Badge } from "@functionfly/ui-core";
import {
  Play,
  Pause,
  CheckCircle2,
  XCircle,
  Clock,
  Loader2,
  ChevronRight,
} from "lucide-react";

export interface ExecutionStep {
  id: string;
  label: string;
  status: "pending" | "running" | "completed" | "failed" | "skipped";
  duration?: number;
  timestamp: number;
  description?: string;
  artifacts?: Array<{ name: string; type: string; url: string }>;
  error?: string;
  metadata?: Record<string, string>;
}

export interface ExecutionNarratorProps {
  steps: ExecutionStep[];
  currentStepId?: string;
  onStepClick?: (step: ExecutionStep) => void;
  className?: string;
  autoPlay?: boolean;
}

interface StatusConfig {
  icon: React.ComponentType<{ className?: string }>;
  color: string;
  bg: string;
  animate?: boolean;
}

const statusConfig: Record<string, StatusConfig> = {
  pending: { icon: Clock, color: "text-text-muted", bg: "bg-bg-tertiary" },
  running: {
    icon: Loader2,
    color: "text-brand-500",
    bg: "bg-brand-500/10",
    animate: true,
  },
  completed: { icon: CheckCircle2, color: "text-success", bg: "bg-success/10" },
  failed: { icon: XCircle, color: "text-error", bg: "bg-error/10" },
  skipped: {
    icon: ChevronRight,
    color: "text-text-muted",
    bg: "bg-bg-tertiary",
  },
};

export function ExecutionNarrator({
  steps,
  currentStepId,
  onStepClick,
  className,
  autoPlay = false,
}: ExecutionNarratorProps) {
  const [isPaused, setIsPaused] = React.useState(false);

  const getStepDuration = (step: ExecutionStep): string => {
    if (!step.duration) return "";
    if (step.duration < 1000) return `${step.duration}ms`;
    return `${(step.duration / 1000).toFixed(1)}s`;
  };

  return (
    <div className={cn("flex flex-col h-full", className)}>
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-border-subtle">
        <div className="flex items-center gap-2">
          <Play className="size-4 text-brand-500" />
          <span className="text-sm font-medium text-text-primary">
            Execution
          </span>
          <Badge variant="brand" size="sm">
            {steps.filter((s) => s.status === "completed").length}/
            {steps.length}
          </Badge>
        </div>
        <button
          onClick={() => setIsPaused((p) => !p)}
          className={cn(
            "p-1.5 rounded-lg transition-colors",
            isPaused
              ? "bg-brand-500/10 text-brand-500"
              : "hover:bg-bg-tertiary text-text-muted",
          )}
        >
          {isPaused ? (
            <Play className="size-4" />
          ) : (
            <Pause className="size-4" />
          )}
        </button>
      </div>

      {/* Steps Timeline */}
      <div className="flex-1 overflow-y-auto p-4">
        <div className="relative">
          {/* Timeline line */}
          <div className="absolute left-5 top-0 bottom-0 w-[1px] bg-border-subtle" />

          <div className="space-y-3">
            {steps.map((step, index) => {
              const config = statusConfig[step.status];
              const isCurrent = currentStepId === step.id;
              const isLast = index === steps.length - 1;
              const Icon = config.icon;

              return (
                <div
                  key={step.id}
                  onClick={() => onStepClick?.(step)}
                  className={cn(
                    "relative flex gap-4 p-3 rounded-lg transition-all cursor-pointer",
                    isCurrent
                      ? "bg-brand-500/5 border border-brand-500/20"
                      : "hover:bg-bg-hover border border-transparent",
                  )}
                >
                  {/* Status Icon */}
                  <div
                    className={cn(
                      "relative z-10 size-8 rounded-full flex items-center justify-center shrink-0",
                      config.bg,
                      config.color,
                    )}
                  >
                    {step.status === "running" && config.animate ? (
                      <Icon className="size-4 animate-spin" />
                    ) : (
                      <Icon
                        className={cn(
                          "size-4",
                          step.status === "running" && "animate-pulse",
                        )}
                      />
                    )}
                  </div>

                  {/* Content */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="text-sm font-medium text-text-primary">
                        {step.label}
                      </span>
                      {step.duration && (
                        <Badge variant="ghost" size="sm">
                          {getStepDuration(step)}
                        </Badge>
                      )}
                      {isCurrent && step.status === "running" && (
                        <Badge variant="brand" size="sm" pulse>
                          Running
                        </Badge>
                      )}
                    </div>

                    {step.description && (
                      <p className="text-xs text-text-muted mb-2">
                        {step.description}
                      </p>
                    )}

                    {/* Error */}
                    {step.error && (
                      <div className="mt-2 p-2 rounded bg-error/10 border border-error/20">
                        <p className="text-xs text-error">{step.error}</p>
                      </div>
                    )}

                    {/* Artifacts */}
                    {step.artifacts && step.artifacts.length > 0 && (
                      <div className="flex items-center gap-2 mt-2">
                        {step.artifacts.map((artifact, i) => (
                          <a
                            key={i}
                            href={artifact.url}
                            className="flex items-center gap-1 px-2 py-1 rounded bg-bg-tertiary hover:bg-bg-hover text-[10px] text-text-secondary transition-colors"
                          >
                            {artifact.name}
                          </a>
                        ))}
                      </div>
                    )}

                    {/* Metadata */}
                    {step.metadata && (
                      <div className="flex flex-wrap gap-2 mt-2">
                        {Object.entries(step.metadata).map(([key, value]) => (
                          <span
                            key={key}
                            className="text-[10px] text-text-muted"
                          >
                            <span className="font-medium">{key}:</span> {value}
                          </span>
                        ))}
                      </div>
                    )}
                  </div>

                  {/* Timestamp */}
                  <span className="text-[10px] text-text-muted shrink-0">
                    {new Date(step.timestamp).toLocaleTimeString()}
                  </span>
                </div>
              );
            })}
          </div>
        </div>
      </div>

      {/* Footer Stats */}
      <div className="px-4 py-2 border-t border-border-subtle flex items-center justify-between text-[10px] text-text-muted">
        <span>
          Duration: {steps.reduce((acc, s) => acc + (s.duration || 0), 0)}ms
        </span>
        <span>{steps.filter((s) => s.status === "failed").length} failed</span>
      </div>
    </div>
  );
}
