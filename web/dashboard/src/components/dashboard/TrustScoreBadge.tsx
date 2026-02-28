import { Shield, CheckCircle, AlertTriangle, XCircle, Info } from "lucide-react";
import { cn } from "@/lib/utils";

export type TrustLevel =
  | "excellent"
  | "good"
  | "fair"
  | "poor"
  | "very_poor"
  | "insufficient_data";

export interface TrustScoreBadgeProps {
  trustScore: number;
  trustLevel?: TrustLevel;
  showScore?: boolean;
  size?: "sm" | "md" | "lg";
  className?: string;
}

function getLevel(score: number): TrustLevel {
  if (score >= 80) return "excellent";
  if (score >= 60) return "good";
  if (score >= 40) return "fair";
  if (score >= 20) return "poor";
  if (score > 0) return "very_poor";
  return "insufficient_data";
}

const levelConfig: Record<
  TrustLevel,
  { label: string; icon: React.ReactNode; textClass: string; bgClass: string }
> = {
  excellent: {
    label: "Excellent",
    icon: <CheckCircle className="h-4 w-4" />,
    textClass: "text-[var(--color-success)]",
    bgClass: "bg-[var(--color-success)]/15",
  },
  good: {
    label: "Good",
    icon: <Shield className="h-4 w-4" />,
    textClass: "text-emerald-400",
    bgClass: "bg-emerald-500/15",
  },
  fair: {
    label: "Fair",
    icon: <AlertTriangle className="h-4 w-4" />,
    textClass: "text-[var(--color-warning)]",
    bgClass: "bg-[var(--color-warning)]/15",
  },
  poor: {
    label: "Poor",
    icon: <AlertTriangle className="h-4 w-4" />,
    textClass: "text-orange-400",
    bgClass: "bg-orange-500/15",
  },
  very_poor: {
    label: "Very Poor",
    icon: <XCircle className="h-4 w-4" />,
    textClass: "text-[var(--color-error)]",
    bgClass: "bg-[var(--color-error)]/15",
  },
  insufficient_data: {
    label: "Insufficient Data",
    icon: <Info className="h-4 w-4" />,
    textClass: "text-text-muted",
    bgClass: "bg-bg-hover",
  },
};

const sizeClasses = {
  sm: "text-xs px-2 py-0.5 gap-1",
  md: "text-sm px-2.5 py-1 gap-1.5",
  lg: "text-base px-3 py-1.5 gap-2",
};

export function TrustScoreBadge({
  trustScore,
  trustLevel,
  showScore = true,
  size = "md",
  className,
}: TrustScoreBadgeProps) {
  const level = trustLevel ?? getLevel(trustScore);
  const config = levelConfig[level];

  return (
    <span
      className={cn(
        "inline-flex items-center font-medium rounded-full border border-theme",
        config.textClass,
        config.bgClass,
        sizeClasses[size],
        className
      )}
    >
      {config.icon}
      {showScore && <span className="tabular-nums">{Math.round(trustScore)}</span>}
      <span>{config.label}</span>
    </span>
  );
}
