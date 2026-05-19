/**
 * @functionfly/ui-ai
 * AI Confidence Meter - Visualize AI response confidence
 */

import * as React from "react";
import { cn } from "@functionfly/ui-core";
import { Badge } from "@functionfly/ui-core";
import { Gauge, AlertTriangle, CheckCircle2, HelpCircle } from "lucide-react";

export interface AIConfidenceMeterProps {
  confidence: number; // 0-1
  thresholds?: {
    high?: number;
    medium?: number;
  };
  label?: string;
  showValue?: boolean;
  showHistory?: boolean;
  history?: number[];
  className?: string;
}

export function AIConfidenceMeter({
  confidence,
  thresholds = { high: 0.8, medium: 0.5 },
  label = "Confidence",
  showValue = true,
  showHistory = false,
  history = [],
  className,
}: AIConfidenceMeterProps) {
  const highThreshold = thresholds.high ?? 0.8;
  const mediumThreshold = thresholds.medium ?? 0.5;

  const getLevel = (): "high" | "medium" | "low" => {
    if (confidence >= highThreshold) return "high";
    if (confidence >= mediumThreshold) return "medium";
    return "low";
  };

  const level = getLevel();

  const levelConfig = {
    high: {
      color: "text-success",
      bg: "bg-success/10",
      border: "border-success/20",
      icon: CheckCircle2,
      label: "High confidence",
    },
    medium: {
      color: "text-warning",
      bg: "bg-warning/10",
      border: "border-warning/20",
      icon: AlertTriangle,
      label: "Medium confidence",
    },
    low: {
      color: "text-error",
      bg: "bg-error/10",
      border: "border-error/20",
      icon: HelpCircle,
      label: "Low confidence",
    },
  };

  const config = levelConfig[level];
  const Icon = config.icon;
  const percentage = Math.round(confidence * 100);

  // Calculate color gradient based on confidence
  const getGradientColor = () => {
    if (confidence >= 0.8) return "from-success to-success";
    if (confidence >= 0.5) return "from-warning to-brand-500";
    return "from-error to-warning";
  };

  return (
    <div className={cn("flex flex-col", className)}>
      {/* Header */}
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <Gauge className={cn("size-4", config.color)} />
          <span className="text-xs font-medium text-text-primary">{label}</span>
        </div>
        {showValue && (
          <Badge className={cn("text-xs", config.bg, config.color)} variant="outline" size="sm">
            {percentage}%
          </Badge>
        )}
      </div>

      {/* Meter */}
      <div className="relative h-3 rounded-full bg-bg-tertiary overflow-hidden">
        <div
          className={cn(
            "absolute left-0 top-0 h-full rounded-full transition-all duration-500",
            `bg-gradient-to-r ${getGradientColor()}`
          )}
          style={{ width: `${percentage}%` }}
        />
        
        {/* Threshold markers */}
        <div
          className="absolute top-0 bottom-0 w-[1px] bg-text-muted/30"
          style={{ left: `${mediumThreshold * 100}%` }}
        />
        <div
          className="absolute top-0 bottom-0 w-[1px] bg-text-muted/30"
          style={{ left: `${highThreshold * 100}%` }}
        />
      </div>

      {/* Level Label */}
      <div className={cn("flex items-center gap-1.5 mt-2", config.color)}>
        <Icon className="size-3" />
        <span className="text-[10px] font-medium">{config.label}</span>
      </div>

      {/* History Sparkline */}
      {showHistory && history.length > 0 && (
        <div className="mt-3 pt-3 border-t border-border-subtle">
          <div className="flex items-center gap-2 mb-1">
            <span className="text-[10px] text-text-muted">Recent</span>
          </div>
          <div className="flex items-center gap-1 h-8">
            {history.slice(-10).map((value, i) => (
              <div
                key={i}
                className={cn(
                  "flex-1 rounded-full transition-all",
                  value >= highThreshold ? "bg-success/40" :
                  value >= mediumThreshold ? "bg-warning/40" : "bg-error/40"
                )}
                style={{ height: `${Math.max(20, value * 100)}%` }}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
