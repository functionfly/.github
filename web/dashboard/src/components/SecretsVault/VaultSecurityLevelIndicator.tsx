/**
 * VaultSecurityLevelIndicator - Visual indicator of vault security level
 *
 * Displays the security level of the vault (Low/Medium/High/Critical)
 * with visual indicators including progress bars, icons, and color coding.
 * Provides contextual recommendations based on the security level.
 *
 * @example
 * ```tsx
 * // Basic usage
 * <VaultSecurityLevelIndicator level="high" />
 *
 * // With detailed breakdown
 * <VaultSecurityLevelIndicator
 *   level="medium"
 *   factors={{
 *     encryption: true,
 *     rotation: false,
 *     mfa: true,
 *     audit: true,
 *   }}
 * />
 *
 * // Loading state
 * <VaultSecurityLevelIndicator isLoading />
 *
 * // Compact variant
 * <VaultSecurityLevelIndicator level="critical" variant="compact" />
 *
 * // With custom action
 * <VaultSecurityLevelIndicator
 *   level="low"
 *   onImprove={() => navigate('/vault/security')}
 * />
 * ```
 */

import {
  Shield,
  ShieldAlert,
  ShieldCheck,
  ShieldX,
  Lock,
  RefreshCw,
  Fingerprint,
  FileText,
  AlertTriangle,
  ChevronRight,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

/** Security level tiers */
export type SecurityLevel = "low" | "medium" | "high" | "critical";

/** Security factor configuration */
export interface SecurityFactors {
  /** Whether strong encryption is enabled */
  encryption?: boolean;
  /** Whether key rotation is configured */
  rotation?: boolean;
  /** Whether MFA is required for access */
  mfa?: boolean;
  /** Whether audit logging is enabled */
  audit?: boolean;
}

export interface VaultSecurityLevelIndicatorProps {
  /** Current security level of the vault */
  level?: SecurityLevel;
  /** Security factor states (optional, for detailed view) */
  factors?: SecurityFactors;
  /** Whether the component is in loading state */
  isLoading?: boolean;
  /** Visual variant */
  variant?: "default" | "compact" | "detailed";
  /** Callback when user wants to improve security */
  onImprove?: () => void;
  /** Additional CSS classes */
  className?: string;
}

/** Security level configuration with colors, icons, and recommendations */
const securityLevelConfig: Record<
  SecurityLevel,
  {
    icon: typeof Shield;
    label: string;
    color: string;
    bgColor: string;
    progressColor: string;
    description: string;
    recommendation: string;
    score: number;
  }
> = {
  low: {
    icon: ShieldX,
    label: "Low Security",
    color: "text-[var(--status-revoked)]",
    bgColor: "rgba(255,107,107,0.06)",
    progressColor: "bg-[var(--status-revoked)]",
    description: "Basic protection with minimal security measures",
    recommendation: "Enable MFA and configure key rotation immediately",
    score: 25,
  },
  medium: {
    icon: ShieldAlert,
    label: "Medium Security",
    color: "text-yellow-500",
    bgColor: "bg-yellow-500/10",
    progressColor: "bg-yellow-500",
    description: "Standard security with some protective measures",
    recommendation: "Consider enabling audit logging and regular key rotation",
    score: 50,
  },
  high: {
    icon: ShieldCheck,
    label: "High Security",
    color: "text-green-500",
    bgColor: "bg-green-500/10",
    progressColor: "bg-green-500",
    description: "Strong security with comprehensive protection",
    recommendation: "Regular security audits recommended",
    score: 75,
  },
  critical: {
    icon: Shield,
    label: "Critical Security",
    color: "text-(--color-brand-500)",
    bgColor: "bg-(--color-brand-500)/10",
    progressColor: "bg-(--color-brand-500)",
    description: "Maximum security with all protective measures enabled",
    recommendation: "Maintain current security posture",
    score: 100,
  },
};

/** Security factor configurations */
const factorConfig: Record<
  keyof SecurityFactors,
  { icon: typeof Lock; label: string }
> = {
  encryption: { icon: Lock, label: "Encryption" },
  rotation: { icon: RefreshCw, label: "Key Rotation" },
  mfa: { icon: Fingerprint, label: "Multi-Factor Auth" },
  audit: { icon: FileText, label: "Audit Logging" },
};

/**
 * Calculate security score from factors
 */
function calculateScoreFromFactors(factors: SecurityFactors): number {
  const factorKeys = Object.keys(factors) as (keyof SecurityFactors)[];
  const enabledCount = factorKeys.filter((key) => factors[key]).length;
  if (enabledCount === 0) return 0;
  if (enabledCount === 1) return 25;
  if (enabledCount === 2) return 50;
  if (enabledCount === 3) return 75;
  return 100;
}

/**
 * Determine security level from factors
 */
function getLevelFromFactors(factors: SecurityFactors): SecurityLevel {
  const score = calculateScoreFromFactors(factors);
  if (score <= 25) return "low";
  if (score <= 50) return "medium";
  if (score <= 75) return "high";
  return "critical";
}

/**
 * Skeleton loader for the security indicator
 */
function SecurityLevelSkeleton({
  variant,
  className,
}: {
  variant: VaultSecurityLevelIndicatorProps["variant"];
  className?: string;
}) {
  if (variant === "compact") {
    return (
      <div className={cn("flex items-center gap-3", className)}>
        <Skeleton className="h-8 w-8 rounded-lg" />
        <Skeleton className="h-4 w-24" />
      </div>
    );
  }

  return (
    <Card className={cn("overflow-hidden", className)}>
      <CardHeader className="pb-2">
        <Skeleton className="h-5 w-32" />
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center gap-4">
          <Skeleton className="h-12 w-12 rounded-lg" />
          <div className="space-y-2">
            <Skeleton className="h-5 w-32" />
            <Skeleton className="h-3 w-48" />
          </div>
        </div>
        <Skeleton className="h-2 w-full" />
        <Skeleton className="h-16 w-full" />
      </CardContent>
    </Card>
  );
}

/**
 * Compact security level badge
 */
function CompactSecurityIndicator({
  level,
  className,
}: {
  level: SecurityLevel;
  className?: string;
}) {
  const config = securityLevelConfig[level];
  const Icon = config.icon;

  return (
    <Badge
      variant="outline"
      className={cn(
        "inline-flex items-center gap-2 px-3 py-1.5",
        config.bgColor,
        config.color,
        className
      )}
    >
      <Icon className="h-4 w-4" />
      <span className="font-medium">{config.label}</span>
    </Badge>
  );
}

/**
 * Security factors checklist component
 */
function SecurityFactorsList({
  factors,
  isLoading,
}: {
  factors?: SecurityFactors;
  isLoading?: boolean;
}) {
  if (!factors || isLoading) {
    return (
      <div className="space-y-2">
        {[1, 2, 3, 4].map((i) => (
          <Skeleton key={i} className="h-6 w-full" />
        ))}
      </div>
    );
  }

  const factorEntries = Object.entries(factors) as [keyof SecurityFactors, boolean][];

  return (
    <div className="space-y-2">
      {factorEntries.map(([key, enabled]) => {
        const config = factorConfig[key];
        const Icon = config.icon;
        return (
          <div
            key={key}
            className={cn(
              "flex items-center justify-between py-1.5 px-2 rounded-md",
              "transition-colors",
              enabled ? "bg-green-500/5" : "bg-(--color-bg-tertiary)"
            )}
          >
            <div className="flex items-center gap-2">
              <Icon
                className={cn(
                  "h-4 w-4",
                  enabled ? "text-green-500" : "text-(--color-text-muted)"
                )}
              />
              <span
                className={cn(
                  "text-sm",
                  enabled
                    ? "text-(--color-text-primary)"
                    : "text-(--color-text-muted)"
                )}
              >
                {config.label}
              </span>
            </div>
            {enabled ? (
              <ShieldCheck className="h-4 w-4 text-green-500" />
            ) : (
              <AlertTriangle className="h-4 w-4 text-yellow-500" />
            )}
          </div>
        );
      })}
    </div>
  );
}

/**
 * VaultSecurityLevelIndicator component
 *
 * Renders a comprehensive security level indicator with visual
 * progress bar, icon, description, and actionable recommendations.
 */
export function VaultSecurityLevelIndicator({
  level: propLevel,
  factors,
  isLoading = false,
  variant = "default",
  onImprove,
  className,
}: VaultSecurityLevelIndicatorProps) {
  // Determine level from factors if not explicitly provided
  const level = propLevel ?? (factors ? getLevelFromFactors(factors) : "low");
  const config = securityLevelConfig[level];
  const Icon = config.icon;

  // Calculate score from factors if available, otherwise use level score
  const score = factors ? calculateScoreFromFactors(factors) : config.score;

  if (isLoading) {
    return <SecurityLevelSkeleton variant={variant} className={className} />;
  }

  if (variant === "compact") {
    return <CompactSecurityIndicator level={level} className={className} />;
  }

  return (
    <Card
      className={cn(
        "overflow-hidden transition-all duration-200",
        "border border-(--border-subtle)",
        "bg-(--color-bg-primary)",
        className
      )}
    >
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium text-(--color-text-secondary)">
          Security Level
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Main indicator */}
        <div className="flex items-center gap-4">
          <div
            className={cn(
              "flex h-12 w-12 items-center justify-center rounded-xl",
              config.bgColor
            )}
          >
            <Icon className={cn("h-6 w-6", config.color)} />
          </div>
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <h3 className="text-lg font-semibold text-(--color-text-primary)">
                {config.label}
              </h3>
              <Badge variant="outline" className={cn("text-xs", config.color)}>
                {score}%
              </Badge>
            </div>
            <p className="text-sm text-(--color-text-muted) truncate">
              {config.description}
            </p>
          </div>
        </div>

        {/* Progress bar */}
        <div className="space-y-1.5">
          <div className="flex justify-between text-xs">
            <span className="text-(--color-text-muted)">Security Score</span>
            <span className={cn("font-medium", config.color)}>{score}%</span>
          </div>
          <div className="h-2 w-full rounded-full bg-(--color-bg-tertiary) overflow-hidden">
            <div
              className={cn(
                "h-full rounded-full transition-all duration-500 ease-out",
                config.progressColor
              )}
              style={{ width: `${score}%` }}
            />
          </div>
        </div>

        {/* Factors (detailed view only) */}
        {variant === "detailed" && (
          <div className="pt-2 border-t border-(--border-subtle)">
            <h4 className="text-xs font-medium text-(--color-text-secondary) mb-2">
              Security Factors
            </h4>
            <SecurityFactorsList factors={factors} />
          </div>
        )}

        {/* Recommendation */}
        <div className="flex items-start gap-2 pt-2 border-t border-(--border-subtle)">
          <AlertTriangle
            className={cn(
              "h-4 w-4 mt-0.5 flex-shrink-0",
              level === "critical" ? "text-green-500" : "text-yellow-500"
            )}
          />
          <div className="flex-1 min-w-0">
            <p className="text-xs text-(--color-text-secondary)">
              {config.recommendation}
            </p>
          </div>
        </div>

        {/* Action button */}
        {onImprove && level !== "critical" && (
          <Button
            variant="outline"
            size="sm"
            className="w-full gap-1"
            onClick={onImprove}
          >
            Improve Security
            <ChevronRight className="h-4 w-4" />
          </Button>
        )}
      </CardContent>
    </Card>
  );
}

export default VaultSecurityLevelIndicator;
