/**
 * SecretLeaseTimer - Visual countdown timer for secret leases
 *
 * Displays a circular or linear progress indicator with time remaining
 * in HH:MM:SS format. Color changes as time expires (green → yellow → red),
 * with warning states at configurable thresholds. Includes renew/extend
 * lease button, expired state with visual indicator, support for different
 * lease durations, and tooltip with lease details.
 *
 * @example
 * ```tsx
 * // Basic usage with expiration
 * <SecretLeaseTimer
 *   expiresAt="2024-12-31T23:59:59Z"
 *   onRenew={() => extendLease()}
 * />
 *
 * // With custom duration and thresholds
 * <SecretLeaseTimer
 *   expiresAt={expiresAt}
 *   duration={3600}
 *   warningThreshold={300}
 *   criticalThreshold={60}
 *   onRenew={handleRenew}
 *   allowRenewal
 * />
 *
 * // Linear variant
 * <SecretLeaseTimer
 *   expiresAt={expiresAt}
 *   variant="linear"
 *   size="lg"
 * />
 *
 * // Already expired
 * <SecretLeaseTimer
 *   expiresAt="2024-01-01T00:00:00Z"
 *   showExpiredState
 * />
 * ```
 */

import { useState, useEffect, useMemo, useCallback } from "react";
import {
  Clock,
  AlertTriangle,
  AlertCircle,
  RefreshCw,
  Hourglass,
  CheckCircle,
  Info,
  Calendar,
  Timer,
} from "lucide-react";
import { format, formatDistanceToNow, isPast, differenceInSeconds } from "date-fns";
import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";

/** Timer visual variant */
export type TimerVariant = "circular" | "linear" | "compact";

/** Timer size */
export type TimerSize = "sm" | "md" | "lg";

/** Lease state derived from time remaining */
export type LeaseState = "active" | "warning" | "critical" | "expired";

export interface SecretLeaseTimerProps {
  /** Expiration timestamp (ISO 8601) */
  expiresAt: string;
  /** Total lease duration in seconds (optional - calculated if not provided) */
  duration?: number;
  /** Warning threshold in seconds (default: 300 = 5 minutes) */
  warningThreshold?: number;
  /** Critical threshold in seconds (default: 60 = 1 minute) */
  criticalThreshold?: number;
  /** Visual variant */
  variant?: TimerVariant;
  /** Timer size */
  size?: TimerSize;
  /** Whether to allow lease renewal */
  allowRenewal?: boolean;
  /** Whether renewal is in progress */
  isRenewing?: boolean;
  /** Whether to show expired state visually */
  showExpiredState?: boolean;
  /** Whether to show detailed tooltip */
  showTooltip?: boolean;
  /** Callback when renew is clicked */
  onRenew?: () => void | Promise<void>;
  /** Callback when timer expires */
  onExpire?: () => void;
  /** Custom label */
  label?: string;
  /** Additional CSS classes */
  className?: string;
}

/**
 * Format seconds to HH:MM:SS
 */
function formatDuration(totalSeconds: number): string {
  if (totalSeconds <= 0) return "00:00:00";

  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;

  if (hours > 0) {
    return `${hours.toString().padStart(2, "0")}:${minutes
      .toString()
      .padStart(2, "0")}:${seconds.toString().padStart(2, "0")}`;
  }

  return `${minutes.toString().padStart(2, "0")}:${seconds
    .toString()
    .padStart(2, "0")}`;
}

/**
 * Calculate lease state based on time remaining
 */
function calculateLeaseState(
  secondsRemaining: number,
  warningThreshold: number,
  criticalThreshold: number
): LeaseState {
  if (secondsRemaining <= 0) return "expired";
  if (secondsRemaining <= criticalThreshold) return "critical";
  if (secondsRemaining <= warningThreshold) return "warning";
  return "active";
}

/**
 * Get color classes based on lease state
 */
function getStateColors(state: LeaseState): {
  text: string;
  bg: string;
  border: string;
  progress: string;
  icon: typeof CheckCircle;
} {
  switch (state) {
    case "expired":
      return {
        text: "text-error",
        bg: "bg-error-glow",
        border: "border-error/30",
        progress: "bg-error",
        icon: AlertCircle,
      };
    case "critical":
      return {
        text: "text-error",
        bg: "bg-error/10",
        border: "border-error/30",
        progress: "bg-error",
        icon: AlertTriangle,
      };
    case "warning":
      return {
        text: "text-warning",
        bg: "bg-warning-glow",
        border: "border-warning/30",
        progress: "bg-warning",
        icon: AlertTriangle,
      };
    default:
      return {
        text: "text-success",
        bg: "bg-success-glow",
        border: "border-success/30",
        progress: "bg-success",
        icon: CheckCircle,
      };
  }
}

/**
 * Circular progress indicator
 */
function CircularProgress({
  progress,
  size,
  state,
  children,
}: {
  progress: number;
  size: TimerSize;
  state: LeaseState;
  children: React.ReactNode;
}) {
  const colors = getStateColors(state);
  const strokeWidth = size === "lg" ? 4 : size === "md" ? 3 : 2;
  const radius = size === "lg" ? 50 : size === "md" ? 36 : 24;
  const normalizedRadius = radius - strokeWidth / 2;
  const circumference = normalizedRadius * 2 * Math.PI;
  const strokeDashoffset = circumference - (progress / 100) * circumference;

  const svgSize = radius * 2 + strokeWidth;

  return (
    <div
      className={cn(
        "relative inline-flex items-center justify-center",
        size === "lg" && "w-28 h-28",
        size === "md" && "w-20 h-20",
        size === "sm" && "w-14 h-14"
      )}
    >
      <svg
        height={svgSize}
        width={svgSize}
        className="transform -rotate-90"
      >
        {/* Background circle */}
        <circle
          stroke="currentColor"
          fill="transparent"
          strokeWidth={strokeWidth}
          r={normalizedRadius}
          cx={svgSize / 2}
          cy={svgSize / 2}
          className="text-border-subtle"
        />
        {/* Progress circle */}
        <circle
          stroke="currentColor"
          fill="transparent"
          strokeWidth={strokeWidth}
          strokeDasharray={circumference}
          strokeDashoffset={strokeDashoffset}
          strokeLinecap="round"
          r={normalizedRadius}
          cx={svgSize / 2}
          cy={svgSize / 2}
          className={cn(colors.progress, "transition-all duration-1000")}
        />
      </svg>
      <div className="absolute inset-0 flex items-center justify-center">
        {children}
      </div>
    </div>
  );
}

/**
 * SecretLeaseTimer component
 */
export function SecretLeaseTimer({
  expiresAt,
  duration,
  warningThreshold = 300,
  criticalThreshold = 60,
  variant = "circular",
  size = "md",
  allowRenewal = false,
  isRenewing = false,
  showExpiredState = true,
  showTooltip = true,
  onRenew,
  onExpire,
  label = "Lease expires in",
  className,
}: SecretLeaseTimerProps) {
  const [now, setNow] = useState(new Date());
  const [hasExpired, setHasExpired] = useState(false);

  // Update timer every second
  useEffect(() => {
    const interval = setInterval(() => {
      setNow(new Date());
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  // Calculate time values
  const expiresDate = useMemo(() => new Date(expiresAt), [expiresAt]);
  const secondsRemaining = useMemo(
    () => Math.max(0, differenceInSeconds(expiresDate, now)),
    [expiresDate, now]
  );
  const isExpired = secondsRemaining === 0 || isPast(expiresDate);

  // Calculate total duration
  const totalDuration = useMemo(() => {
    if (duration) return duration;
    // Estimate from expiresAt and createdAt (if we had it)
    // Default to 1 hour if not calculable
    return 3600;
  }, [duration]);

  // Calculate progress percentage
  const progress = useMemo(() => {
    if (isExpired) return 0;
    return Math.max(0, Math.min(100, (secondsRemaining / totalDuration) * 100));
  }, [secondsRemaining, totalDuration, isExpired]);

  // Determine lease state
  const leaseState = useMemo(
    () => calculateLeaseState(secondsRemaining, warningThreshold, criticalThreshold),
    [secondsRemaining, warningThreshold, criticalThreshold]
  );

  // Get state colors
  const colors = getStateColors(leaseState);

  // Handle expiration
  useEffect(() => {
    if (isExpired && !hasExpired) {
      setHasExpired(true);
      onExpire?.();
    }
  }, [isExpired, hasExpired, onExpire]);

  // Handle renew click
  const handleRenew = useCallback(() => {
    if (onRenew && !isRenewing) {
      onRenew();
    }
  }, [onRenew, isRenewing]);

  // Format display time
  const displayTime = formatDuration(secondsRemaining);
  const timeParts = displayTime.split(":");

  // Compact variant
  if (variant === "compact") {
    const StateIcon = colors.icon;

    return (
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <div
              className={cn(
                "inline-flex items-center gap-2 px-3 py-1.5 rounded-full border",
                colors.bg,
                colors.border,
                className
              )}
            >
              <StateIcon className={cn("h-4 w-4", colors.text)} />
              <span className={cn("text-sm font-medium tabular-nums", colors.text)}>
                {isExpired ? "Expired" : displayTime}
              </span>
              {allowRenewal && !isExpired && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-5 w-5 ml-1"
                  onClick={handleRenew}
                  disabled={isRenewing}
                  aria-label="Renew lease"
                >
                  <RefreshCw className={cn("h-3 w-3", isRenewing && "animate-spin")} />
                </Button>
              )}
            </div>
          </TooltipTrigger>
          {showTooltip && (
            <TooltipContent>
              <LeaseDetails
                expiresAt={expiresAt}
                secondsRemaining={secondsRemaining}
                leaseState={leaseState}
              />
            </TooltipContent>
          )}
        </Tooltip>
      </TooltipProvider>
    );
  }

  // Linear variant
  if (variant === "linear") {
    return (
      <TooltipProvider>
        <div className={cn("space-y-2", className)}>
          <div className="flex items-center justify-between text-sm">
            <div className="flex items-center gap-2">
              <Clock className={cn("h-4 w-4", colors.text)} />
              <span className="text-[var(--text-dim)]">{label}</span>
            </div>
            <div className="flex items-center gap-2">
              <span
                className={cn(
                  "font-medium tabular-nums",
                  isExpired ? "text-error" : colors.text
                )}
              >
                {isExpired ? "Expired" : displayTime}
              </span>
              {allowRenewal && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleRenew}
                  disabled={isRenewing}
                  className="h-7 px-2"
                >
                  <RefreshCw
                    className={cn("h-3.5 w-3.5 mr-1", isRenewing && "animate-spin")}
                  />
                  Renew
                </Button>
              )}
            </div>
          </div>
          <Tooltip>
            <TooltipTrigger asChild>
              <div>
                <Progress
                  value={progress}
                  className={cn("h-2", leaseState === "critical" && "animate-pulse")}
                />
              </div>
            </TooltipTrigger>
            {showTooltip && (
              <TooltipContent>
                <LeaseDetails
                  expiresAt={expiresAt}
                  secondsRemaining={secondsRemaining}
                  leaseState={leaseState}
                />
              </TooltipContent>
            )}
          </Tooltip>
        </div>
      </TooltipProvider>
    );
  }

  // Circular variant (default)
  const StateIcon = colors.icon;

  return (
    <TooltipProvider>
      <div className={cn("flex flex-col items-center gap-3", className)}>
        <Tooltip>
          <TooltipTrigger asChild>
            <div className="relative">
              <CircularProgress
                progress={progress}
                size={size}
                state={leaseState}
              >
                <div className="flex flex-col items-center">
                  {isExpired && showExpiredState ? (
                    <div className="flex flex-col items-center">
                      <Hourglass className="h-6 w-6 text-error" />
                      <span className="text-xs font-medium text-error mt-1">
                        Expired
                      </span>
                    </div>
                  ) : (
                    <div className="flex flex-col items-center">
                      <div
                        className={cn(
                          "flex items-baseline gap-0.5",
                          size === "lg" && "text-2xl",
                          size === "md" && "text-lg",
                          size === "sm" && "text-sm"
                        )}
                      >
                        {timeParts.map((part, i) => (
                          <span key={i} className={cn("font-bold tabular-nums", colors.text)}>
                            {part}
                            {i < timeParts.length - 1 && (
                              <span className="text-[var(--text-faint)]">:</span>
                            )}
                          </span>
                        ))}
                      </div>
                      {size !== "sm" && (
                        <span className="text-[10px] text-[var(--text-faint)]">
                          {timeParts.length > 2 ? "HH:MM:SS" : "MM:SS"}
                        </span>
                      )}
                    </div>
                  )}
                </div>
              </CircularProgress>

              {/* State indicator badge */}
              <Badge
                variant={
                  leaseState === "expired"
                    ? "error"
                    : leaseState === "critical"
                    ? "error"
                    : leaseState === "warning"
                    ? "warning"
                    : "success"
                }
                className="absolute -top-1 -right-1 h-5 w-5 p-0 flex items-center justify-center"
              >
                <StateIcon className="h-3 w-3" />
              </Badge>
            </div>
          </TooltipTrigger>
          {showTooltip && (
            <TooltipContent>
              <LeaseDetails
                expiresAt={expiresAt}
                secondsRemaining={secondsRemaining}
                leaseState={leaseState}
              />
            </TooltipContent>
          )}
        </Tooltip>

        {/* Label and renew button */}
        <div className="flex flex-col items-center gap-2">
          <span className="text-sm text-[var(--text-dim)]">{label}</span>
          {allowRenewal && (
            <Button
              variant="outline"
              size="sm"
              onClick={handleRenew}
              disabled={isRenewing}
              className="gap-2"
            >
              <RefreshCw className={cn("h-4 w-4", isRenewing && "animate-spin")} />
              Renew Lease
            </Button>
          )}
        </div>
      </div>
    </TooltipProvider>
  );
}

/**
 * Lease details component for tooltip
 */
interface LeaseDetailsProps {
  expiresAt: string;
  secondsRemaining: number;
  leaseState: LeaseState;
}

function LeaseDetails({ expiresAt, secondsRemaining, leaseState }: LeaseDetailsProps) {
  const stateLabels: Record<LeaseState, string> = {
    active: "Active",
    warning: "Expiring Soon",
    critical: "Critical",
    expired: "Expired",
  };

  const stateDescriptions: Record<LeaseState, string> = {
    active: "This lease is active and has plenty of time remaining.",
    warning: "This lease will expire soon. Consider renewing it.",
    critical: "This lease is about to expire! Renew immediately to avoid interruption.",
    expired: "This lease has expired and is no longer valid.",
  };

  return (
    <div className="space-y-3 max-w-xs">
      <div className="flex items-center gap-2">
        <Badge
          variant={
            leaseState === "expired"
              ? "error"
              : leaseState === "critical"
              ? "error"
              : leaseState === "warning"
              ? "warning"
              : "success"
          }
        >
          {stateLabels[leaseState]}
        </Badge>
        <span className="text-sm font-medium">
          {secondsRemaining > 0 ? formatDuration(secondsRemaining) : "Expired"}
        </span>
      </div>

      <p className="text-sm text-[var(--text-dim)]">{stateDescriptions[leaseState]}</p>

      <div className="space-y-1 text-xs text-[var(--text-faint)]">
        <div className="flex items-center gap-2">
          <Calendar className="h-3.5 w-3.5" />
          <span>Expires: {format(new Date(expiresAt), "MMM d, yyyy HH:mm")}</span>
        </div>
        <div className="flex items-center gap-2">
          <Timer className="h-3.5 w-3.5" />
          <span>{formatDistanceToNow(new Date(expiresAt), { addSuffix: true })}</span>
        </div>
      </div>
    </div>
  );
}

export type { LeaseDetailsProps };
