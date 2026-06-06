import { CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import * as React from "react";
import { TrendSparkline } from "./TrendSparkline";

/**
 * Semantic tone for a metric card. Drives the icon-tile background,
 * the icon color, the accent bar, the value color, and the card's
 * hairline border so each card has a distinct, meaningful identity
 * instead of a uniform gray floating panel.
 *
 * Colors are sourced from CSS variables defined in index.css so they
 * flip automatically with [data-theme='light'].
 */
export type MetricTone =
  | "neutral"
  | "indigo"
  | "amber"
  | "emerald"
  | "cyan"
  | "violet"
  | "rose";

export interface MetricCardProps {
  title: string;
  value: string | number;
  /** Short label under the value, e.g. "vs last 7d" */
  changeLabel?: string;
  /** Percent change; positive = up, negative = down */
  changePercent?: number;
  /** Optional sparkline data (e.g. last 7–30 points) */
  sparklineData?: number[];
  icon?: React.ReactNode;
  /**
   * Semantic color tone. When omitted, falls back to the legacy
   * neutral gray look so existing call-sites keep working.
   */
  tone?: MetricTone;
  className?: string;
}

/**
 * Map tone → CSS variable names. Variables are defined in index.css
 * with paired light/dark values, so we just reference them here.
 */
const TONE_VARS: Record<
  MetricTone,
  { tint: string; border: string; value: string }
> = {
  neutral: {
    tint: "var(--metric-tint-neutral)",
    border: "var(--metric-border-neutral)",
    value: "var(--metric-value-neutral)",
  },
  indigo: {
    tint: "var(--metric-tint-indigo)",
    border: "var(--metric-border-indigo)",
    value: "var(--metric-value-indigo)",
  },
  amber: {
    tint: "var(--metric-tint-amber)",
    border: "var(--metric-border-amber)",
    value: "var(--metric-value-amber)",
  },
  emerald: {
    tint: "var(--metric-tint-emerald)",
    border: "var(--metric-border-emerald)",
    value: "var(--metric-value-emerald)",
  },
  cyan: {
    tint: "var(--metric-tint-cyan)",
    border: "var(--metric-border-cyan)",
    value: "var(--metric-value-cyan)",
  },
  violet: {
    tint: "var(--metric-tint-violet)",
    border: "var(--metric-border-violet)",
    value: "var(--metric-value-violet)",
  },
  rose: {
    tint: "var(--metric-tint-rose)",
    border: "var(--metric-border-rose)",
    value: "var(--metric-value-rose)",
  },
};

export function MetricCard({
  title,
  value,
  changeLabel,
  changePercent,
  sparklineData,
  icon,
  tone = "neutral",
  className,
}: MetricCardProps) {
  const trend = changePercent == null ? "neutral" : changePercent >= 0 ? "up" : "down";
  const vars = TONE_VARS[tone];

  return (
    // Plain div instead of Card so we can fully control the surface
    // (no bg-card, no border-theme utility overriding our tone).
    <div
      className={cn(
        "relative overflow-hidden rounded-xl border bg-transparent",
        "transition-colors duration-200",
        className
      )}
      style={{
        borderColor: vars.border,
        boxShadow: `inset 0 1px 0 0 ${vars.tint}`,
      }}
    >
      {/* Left accent bar — 2px, full height, color matches tone */}
      <span
        aria-hidden
        className="pointer-events-none absolute left-0 top-3 bottom-3 w-[2px] rounded-full"
        style={{ backgroundColor: vars.border }}
      />
      <CardContent className="p-5">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0 flex-1">
            <p className="text-xs font-medium uppercase tracking-wider text-text-muted">
              {title}
            </p>
            <div className="mt-2 flex items-baseline gap-2">
              <span
                className="text-2xl font-bold tabular-nums"
                style={{ color: vars.value }}
              >
                {value}
              </span>
              {changePercent != null && (
                <span
                  className={cn(
                    "text-xs font-medium tabular-nums",
                    trend === "up" && "text-[var(--color-success)]",
                    trend === "down" && "text-[var(--color-error)]",
                    trend === "neutral" && "text-text-muted"
                  )}
                >
                  {changePercent >= 0 ? "+" : ""}
                  {changePercent}%
                </span>
              )}
            </div>
            {changeLabel && (
              <p className="mt-0.5 text-xs text-text-muted">{changeLabel}</p>
            )}
            {sparklineData && sparklineData.length > 0 && (
              <div className="mt-3 h-8 w-full max-w-[140px]">
                <TrendSparkline
                  data={sparklineData}
                  trend={trend}
                  className="h-full w-full"
                />
              </div>
            )}
          </div>
          {icon && (
            <div
              className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border"
              style={{
                backgroundColor: vars.tint,
                borderColor: vars.border,
                color: vars.value,
              }}
            >
              {icon}
            </div>
          )}
        </div>
      </CardContent>
    </div>
  );
}
