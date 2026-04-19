"use client"

import { TrendingUp, TrendingDown, DollarSign } from "lucide-react"
import { cn } from "@/lib/utils"
import { formatCostUsd } from "@/api/usageAnalytics"

interface SpendingSummaryWidgetProps {
  currentSpend: number // in cents
  previousSpend?: number // in cents
  loading?: boolean
  className?: string
}

export function SpendingSummaryWidget({
  currentSpend,
  previousSpend,
  loading = false,
  className,
}: SpendingSummaryWidgetProps) {
  const trend = previousSpend
    ? ((currentSpend - previousSpend) / previousSpend) * 100
    : null
  const isIncreasing = trend !== null && trend > 0
  const isSignificant = trend !== null && Math.abs(trend) > 5

  if (loading) {
    return (
      <div className={cn("p-4 rounded-lg bg-bg-secondary border border-border-default", className)}>
        <div className="animate-pulse space-y-3">
          <div className="h-4 w-24 bg-bg-hover rounded" />
          <div className="h-8 w-32 bg-bg-hover rounded" />
          <div className="h-3 w-20 bg-bg-hover rounded" />
        </div>
      </div>
    )
  }

  return (
    <div
      className={cn(
        "p-4 rounded-lg bg-gradient-to-br from-brand-500/10 to-brand-600/5 border border-brand-500/20",
        "dark:shadow-[0_0_30px_rgba(255,107,53,0.1)]",
        className
      )}
    >
      <div className="flex items-start justify-between">
        <div className="space-y-1">
          <p className="text-xs text-text-muted uppercase tracking-wide">This Month's Spend</p>
          <p className="text-2xl font-bold text-text-primary tabular-nums">
            {formatCostUsd(currentSpend)}
          </p>
          {trend !== null && (
            <div className="flex items-center gap-1">
              {isIncreasing ? (
                <TrendingUp className={cn("w-3 h-3", isSignificant ? "text-red-400" : "text-text-muted")} />
              ) : (
                <TrendingDown className="w-3 h-3 text-green-400" />
              )}
              <span
                className={cn(
                  "text-xs font-medium",
                  isSignificant ? (isIncreasing ? "text-red-400" : "text-green-400") : "text-text-muted"
                )}
              >
                {isIncreasing ? "+" : ""}
                {trend.toFixed(1)}% vs last month
              </span>
            </div>
          )}
        </div>
        <div className="w-10 h-10 rounded-lg bg-brand-500/20 flex items-center justify-center">
          <DollarSign className="w-5 h-5 text-brand-500" />
        </div>
      </div>
    </div>
  )
}