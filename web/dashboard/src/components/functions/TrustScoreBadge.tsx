import React, { useMemo } from "react";
import {
  Shield,
  Zap,
  Clock,
  Target,
  Users,
  AlertTriangle,
  CheckCircle,
  Info,
  TrendingUp,
  Activity,
} from "lucide-react";
import { cn } from "@/lib/utils";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import type { TrustMetrics, TrustScoreBadgeProps, TrustScoreBand, FraudRiskLevel } from "@/types";

/**
 * Get the trust score band based on the overall score
 */
export function getTrustScoreBand(score: number): TrustScoreBand {
  if (score >= 90) return "excellent";
  if (score >= 70) return "good";
  if (score >= 50) return "fair";
  return "poor";
}

/**
 * Get color configuration based on trust score band
 */
export function getTrustColorConfig(band: TrustScoreBand) {
  const configs = {
    excellent: {
      primary: "#10b981", // emerald-500
      bg: "bg-emerald-500",
      bgLight: "bg-emerald-500/10",
      text: "text-emerald-500",
      border: "border-emerald-500/30",
      shadow: "shadow-emerald-500/20",
      label: "Excellent",
    },
    good: {
      primary: "#3b82f6", // blue-500
      bg: "bg-blue-500",
      bgLight: "bg-blue-500/10",
      text: "text-blue-500",
      border: "border-blue-500/30",
      shadow: "shadow-blue-500/20",
      label: "Good",
    },
    fair: {
      primary: "#f59e0b", // amber-500
      bg: "bg-amber-500",
      bgLight: "bg-amber-500/10",
      text: "text-amber-500",
      border: "border-amber-500/30",
      shadow: "shadow-amber-500/20",
      label: "Fair",
    },
    poor: {
      primary: "#ef4444", // red-500
      bg: "bg-red-500",
      bgLight: "bg-red-500/10",
      text: "text-red-500",
      border: "border-red-500/30",
      shadow: "shadow-red-500/20",
      label: "Poor",
    },
  };
  return configs[band];
}

/**
 * Get fraud risk configuration
 */
function getFraudRiskConfig(risk: FraudRiskLevel) {
  const configs = {
    low: {
      icon: CheckCircle,
      color: "text-emerald-500",
      bgColor: "bg-emerald-500/10",
      label: "Low Risk",
    },
    medium: {
      icon: AlertTriangle,
      color: "text-amber-500",
      bgColor: "bg-amber-500/10",
      label: "Medium Risk",
    },
    high: {
      icon: AlertTriangle,
      color: "text-red-500",
      bgColor: "bg-red-500/10",
      label: "High Risk",
    },
  };
  return configs[risk];
}

/**
 * Individual metric item with progress bar
 */
interface MetricItemProps {
  label: string;
  value: number;
  icon: React.ElementType;
  colorClass: string;
  description?: string;
}

function MetricItem({ label, value, icon: Icon, colorClass, description }: MetricItemProps) {
  const IconComponent = Icon as React.ComponentType<{ className?: string }>;
  return (
    <div className="space-y-1.5">
      <div className="flex items-center justify-between text-xs">
        <div className="flex items-center gap-1.5">
          <IconComponent className={cn("h-3.5 w-3.5", colorClass)} />
          <span className="text-text-secondary">{label}</span>
        </div>
        <span className={cn("font-semibold", colorClass)}>{Math.round(value)}%</span>
      </div>
      <div className="h-1.5 bg-bg-secondary rounded-full overflow-hidden">
        <div
          className={cn("h-full rounded-full transition-all duration-700 ease-out", colorClass.replace("text-", "bg-"))}
          style={{ width: `${Math.max(0, Math.min(100, value))}%` }}
        />
      </div>
      {description && (
        <p className="text-[10px] text-text-muted leading-tight">{description}</p>
      )}
    </div>
  );
}

/**
 * Circular progress indicator for overall score
 */
function CircularScore({
  score,
  color,
  size = "md",
}: {
  score: number;
  color: string;
  size?: "sm" | "md" | "lg";
}) {
  const sizeConfig = {
    sm: { width: 40, strokeWidth: 4, fontSize: "text-[10px]" },
    md: { width: 56, strokeWidth: 5, fontSize: "text-xs" },
    lg: { width: 80, strokeWidth: 6, fontSize: "text-base" },
  };

  const config = sizeConfig[size];
  const normalizedScore = Math.max(0, Math.min(100, score));
  const radius = (config.width - config.strokeWidth) / 2;
  const circumference = 2 * Math.PI * radius;
  const strokeDashoffset = circumference - (normalizedScore / 100) * circumference;

  return (
    <div className="relative inline-flex items-center justify-center">
      <svg
        width={config.width}
        height={config.width}
        className="transform -rotate-90"
        aria-hidden="true"
      >
        <circle
          cx={config.width / 2}
          cy={config.width / 2}
          r={radius}
          fill="none"
          stroke="currentColor"
          strokeWidth={config.strokeWidth}
          className="text-bg-secondary"
        />
        <circle
          cx={config.width / 2}
          cy={config.width / 2}
          r={radius}
          fill="none"
          stroke={color}
          strokeWidth={config.strokeWidth}
          strokeLinecap="round"
          strokeDasharray={circumference}
          strokeDashoffset={strokeDashoffset}
          className="transition-all duration-700 ease-out"
        />
      </svg>
      <span
        className={cn("absolute font-bold", config.fontSize)}
        style={{ color }}
        aria-label={`Trust score ${Math.round(score)} percent`}
      >
        {Math.round(score)}
      </span>
    </div>
  );
}

/**
 * Mini variant - compact inline display
 */
function MiniVariant({ metrics, colorConfig }: { metrics: TrustMetrics; colorConfig: ReturnType<typeof getTrustColorConfig> }) {
  return (
    <div
      className={cn(
        "inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full",
        "bg-bg-secondary border border-border-subtle",
        "transition-colors duration-300"
      )}
      aria-label={`Trust score ${Math.round(metrics.overallScore)}% - ${colorConfig.label}`}
    >
      <Shield className={cn("h-3 w-3", colorConfig.text)} />
      <span className={cn("text-xs font-medium", colorConfig.text)}>
        {Math.round(metrics.overallScore)}%
      </span>
    </div>
  );
}

/**
 * Compact variant - for small spaces like FunctionCard
 */
function CompactVariant({
  metrics,
  colorConfig,
  band,
}: {
  metrics: TrustMetrics;
  colorConfig: ReturnType<typeof getTrustColorConfig>;
  band: TrustScoreBand;
}) {
  const fraudConfig = getFraudRiskConfig(metrics.fraudRisk);
  const FraudIcon = fraudConfig.icon;

  return (
    <div
      className={cn(
        "inline-flex items-center gap-2 px-3 py-1.5 rounded-lg",
        "bg-bg-secondary border border-border-subtle",
        "transition-all duration-300"
      )}
    >
      <CircularScore score={metrics.overallScore} color={colorConfig.primary} size="sm" />
      <div className="flex flex-col">
        <span className={cn("text-xs font-semibold", colorConfig.text)}>
          {colorConfig.label}
        </span>
        <div className="flex items-center gap-1.5">
          <FraudIcon className={cn("h-3 w-3", fraudConfig.color)} />
          <span className="text-[10px] text-text-muted">{fraudConfig.label}</span>
        </div>
      </div>
    </div>
  );
}

/**
 * Expanded variant - detailed view for FunctionHeader
 */
function ExpandedVariant({
  metrics,
  colorConfig,
  band,
}: {
  metrics: TrustMetrics;
  colorConfig: ReturnType<typeof getTrustColorConfig>;
  band: TrustScoreBand;
}) {
  const fraudConfig = getFraudRiskConfig(metrics.fraudRisk);
  const FraudIcon = fraudConfig.icon;

  const getScoreColor = (score: number): string => {
    if (score >= 90) return "text-emerald-500";
    if (score >= 70) return "text-blue-500";
    if (score >= 50) return "text-amber-500";
    return "text-red-500";
  };

  return (
    <div
      className={cn(
        "p-4 rounded-xl border",
        "bg-bg-secondary border-border-subtle",
        "transition-all duration-300"
      )}
    >
      {/* Header with overall score */}
      <div className="flex items-start justify-between mb-4">
        <div className="flex items-center gap-3">
          <CircularScore score={metrics.overallScore} color={colorConfig.primary} size="lg" />
          <div>
            <h4 className={cn("text-lg font-bold", colorConfig.text)}>
              {colorConfig.label} Trust
            </h4>
            <p className="text-sm text-text-muted">
              Overall Score: {Math.round(metrics.overallScore)}%
            </p>
          </div>
        </div>
        <div
          className={cn(
            "flex items-center gap-1.5 px-2.5 py-1 rounded-full",
            fraudConfig.bgColor
          )}
        >
          <FraudIcon className={cn("h-4 w-4", fraudConfig.color)} />
          <span className={cn("text-xs font-medium", fraudConfig.color)}>
            {fraudConfig.label}
          </span>
        </div>
      </div>

      {/* Metrics grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <MetricItem
          label="Reliability"
          value={metrics.reliability}
          icon={Activity}
          colorClass={getScoreColor(metrics.reliability)}
          description="Uptime and execution success rate"
        />
        <MetricItem
          label="Latency"
          value={metrics.latency}
          icon={Zap}
          colorClass={getScoreColor(metrics.latency)}
          description="Response time performance"
        />
        <MetricItem
          label="Determinism"
          value={metrics.determinism}
          icon={Target}
          colorClass={getScoreColor(metrics.determinism)}
          description="Consistency of outputs"
        />
        <MetricItem
          label="Community"
          value={metrics.communityReputation}
          icon={Users}
          colorClass={getScoreColor(metrics.communityReputation)}
          description="User ratings and votes"
        />
      </div>

      {/* Details footer */}
      {metrics.details && (
        <div className="mt-4 pt-3 border-t border-border-subtle">
          <div className="flex items-center gap-1.5 text-xs text-text-muted">
            <Info className="h-3.5 w-3.5" />
            <span>
              Based on {metrics.details.totalExecutions?.toLocaleString() ?? "N/A"} executions
              {metrics.details.lastUpdated && (
                <> · Updated {new Date(metrics.details.lastUpdated).toLocaleDateString()}</>
              )}
            </span>
          </div>
        </div>
      )}
    </div>
  );
}

/**
 * Detailed tooltip content
 */
function TooltipContentDetailed({ metrics }: { metrics: TrustMetrics }) {
  const colorConfig = getTrustColorConfig(getTrustScoreBand(metrics.overallScore));
  const fraudConfig = getFraudRiskConfig(metrics.fraudRisk);
  const FraudIcon = fraudConfig.icon;

  const getScoreColor = (score: number): string => {
    if (score >= 90) return "text-emerald-500";
    if (score >= 70) return "text-blue-500";
    if (score >= 50) return "text-amber-500";
    return "text-red-500";
  };

  return (
    <div className="w-64 space-y-3">
      {/* Overall score */}
      <div className="flex items-center justify-between pb-2 border-b border-border-subtle">
        <span className="text-sm font-medium text-text-primary">Trust Score</span>
        <span className={cn("text-sm font-bold", colorConfig.text)}>
          {Math.round(metrics.overallScore)}% - {colorConfig.label}
        </span>
      </div>

      {/* Individual metrics */}
      <div className="space-y-2">
        <TooltipMetricRow
          label="Reliability"
          value={metrics.reliability}
          icon={Activity}
          colorClass={getScoreColor(metrics.reliability)}
        />
        <TooltipMetricRow
          label="Latency"
          value={metrics.latency}
          icon={Zap}
          colorClass={getScoreColor(metrics.latency)}
        />
        <TooltipMetricRow
          label="Determinism"
          value={metrics.determinism}
          icon={Target}
          colorClass={getScoreColor(metrics.determinism)}
        />
        <TooltipMetricRow
          label="Community"
          value={metrics.communityReputation}
          icon={Users}
          colorClass={getScoreColor(metrics.communityReputation)}
        />
      </div>

      {/* Fraud risk */}
      <div className="flex items-center justify-between pt-2 border-t border-border-subtle">
        <span className="text-xs text-text-muted">Fraud Risk</span>
        <div className="flex items-center gap-1">
          <FraudIcon className={cn("h-3.5 w-3.5", fraudConfig.color)} />
          <span className={cn("text-xs font-medium", fraudConfig.color)}>
            {fraudConfig.label}
          </span>
        </div>
      </div>

      {/* Details */}
      {metrics.details && (
        <div className="text-[10px] text-text-muted pt-1">
          {metrics.details.totalExecutions && (
            <p>{metrics.details.totalExecutions.toLocaleString()} total executions</p>
          )}
          {metrics.details.averageResponseTimeMs && (
            <p>Avg response: {metrics.details.averageResponseTimeMs}ms</p>
          )}
          {metrics.details.voteCount && <p>{metrics.details.voteCount} community votes</p>}
        </div>
      )}
    </div>
  );
}

/**
 * Simple metric row for tooltip
 */
function TooltipMetricRow({
  label,
  value,
  icon: Icon,
  colorClass,
}: {
  label: string;
  value: number;
  icon: React.ElementType;
  colorClass: string;
}) {
  const IconComponent = Icon as React.ComponentType<{ className?: string }>;
  return (
    <div className="flex items-center justify-between text-xs">
      <div className="flex items-center gap-1.5">
        <IconComponent className={cn("h-3.5 w-3.5", colorClass)} />
        <span className="text-text-secondary">{label}</span>
      </div>
      <span className={cn("font-medium", colorClass)}>{Math.round(value)}%</span>
    </div>
  );
}

/**
 * TrustScoreBadge Component
 *
 * Displays comprehensive trust metrics with dynamic color coding.
 * Supports three variants: mini, compact, and expanded.
 */
export function TrustScoreBadge({
  metrics,
  variant = "compact",
  showDetails = true,
  className,
  onClick,
}: TrustScoreBadgeProps) {
  const band = useMemo(() => getTrustScoreBand(metrics.overallScore), [metrics.overallScore]);
  const colorConfig = useMemo(() => getTrustColorConfig(band), [band]);

  const handleClick = () => {
    if (onClick) {
      onClick();
    }
  };

  const badgeContent = () => {
    switch (variant) {
      case "mini":
        return <MiniVariant metrics={metrics} colorConfig={colorConfig} />;
      case "expanded":
        return <ExpandedVariant metrics={metrics} colorConfig={colorConfig} band={band} />;
      case "compact":
      default:
        return <CompactVariant metrics={metrics} colorConfig={colorConfig} band={band} />;
    }
  };

  const content = (
    <div
      className={cn(
        "inline-block",
        onClick && "cursor-pointer hover:opacity-90 transition-opacity",
        className
      )}
      onClick={handleClick}
      role={onClick ? "button" : undefined}
      tabIndex={onClick ? 0 : undefined}
      onKeyDown={
        onClick
          ? (e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                handleClick();
              }
            }
          : undefined
      }
      aria-label={`Trust score badge - ${Math.round(metrics.overallScore)}% ${colorConfig.label}`}
    >
      {badgeContent()}
    </div>
  );

  if (!showDetails) {
    return content;
  }

  return (
    <TooltipProvider delayDuration={200}>
      <Tooltip>
        <TooltipTrigger asChild>{content}</TooltipTrigger>
        <TooltipContent
          side="top"
          sideOffset={8}
          className="bg-bg-primary border-border-subtle p-3 shadow-lg"
        >
          <TooltipContentDetailed metrics={metrics} />
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

/**
 * TrustScoreBadge Skeleton for loading states
 */
export function TrustScoreBadgeSkeleton({ variant = "compact" }: { variant?: "compact" | "expanded" | "mini" }) {
  if (variant === "mini") {
    return (
      <div className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-bg-secondary animate-pulse">
        <div className="w-3 h-3 rounded-full bg-bg-hover" />
        <div className="w-8 h-3 rounded bg-bg-hover" />
      </div>
    );
  }

  if (variant === "expanded") {
    return (
      <div className="p-4 rounded-xl bg-bg-secondary animate-pulse space-y-4">
        <div className="flex items-start gap-3">
          <div className="w-20 h-20 rounded-full bg-bg-hover" />
          <div className="space-y-2">
            <div className="w-32 h-5 rounded bg-bg-hover" />
            <div className="w-24 h-3 rounded bg-bg-hover" />
          </div>
        </div>
        <div className="grid grid-cols-2 gap-3">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="space-y-1">
              <div className="w-full h-3 rounded bg-bg-hover" />
              <div className="w-full h-1.5 rounded bg-bg-hover" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg bg-bg-secondary animate-pulse">
      <div className="w-10 h-10 rounded-full bg-bg-hover" />
      <div className="space-y-1">
        <div className="w-16 h-3 rounded bg-bg-hover" />
        <div className="w-20 h-2 rounded bg-bg-hover" />
      </div>
    </div>
  );
}

export default TrustScoreBadge;
