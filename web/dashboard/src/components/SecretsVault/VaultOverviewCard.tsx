/**
 * VaultOverviewCard - Summary card displaying vault health and key metrics
 *
 * Displays essential vault statistics including total secrets count,
 * vault health status, and last rotation date in a compact card format.
 *
 * @example
 * ```tsx
 * // Basic usage
 * <VaultOverviewCard
 *   totalSecrets={42}
 *   healthStatus="healthy"
 *   lastRotationDate="2024-03-01T10:00:00Z"
 * />
 *
 * // With loading state
 * <VaultOverviewCard isLoading />
 *
 * // With custom className
 * <VaultOverviewCard
 *   totalSecrets={42}
 *   healthStatus="warning"
 *   lastRotationDate="2024-03-01T10:00:00Z"
 *   className="col-span-2"
 * />
 * ```
 */

import { Shield, Key, RefreshCw, AlertTriangle, CheckCircle, XCircle } from "lucide-react";
import { cn, formatDateTime } from "@/lib/utils";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";

/** Vault health status levels */
export type VaultHealthStatus = "healthy" | "warning" | "error" | "unknown";

export interface VaultOverviewCardProps {
  /** Total number of secrets in the vault */
  totalSecrets?: number;
  /** Current health status of the vault */
  healthStatus?: VaultHealthStatus;
  /** ISO timestamp of the last key rotation */
  lastRotationDate?: string;
  /** Whether the component is in loading state */
  isLoading?: boolean;
  /** Optional callback when refresh is clicked */
  onRefresh?: () => void;
  /** Additional CSS classes */
  className?: string;
}

/** Health status configuration with icons and styling */
const healthConfig: Record<
  VaultHealthStatus,
  { icon: typeof Shield; label: string; variant: "default" | "secondary" | "destructive" | "outline" | "success" | "warning" | "error" }
> = {
  healthy: { icon: CheckCircle, label: "Healthy", variant: "success" },
  warning: { icon: AlertTriangle, label: "Warning", variant: "warning" },
  error: { icon: XCircle, label: "Error", variant: "destructive" },
  unknown: { icon: Shield, label: "Unknown", variant: "secondary" },
};

/**
 * Skeleton loader for the overview card
 */
function VaultOverviewCardSkeleton({ className }: { className?: string }) {
  return (
    <Card className={cn("overflow-hidden", className)}>
      <CardHeader className="pb-2">
        <Skeleton className="h-5 w-32" />
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center gap-4">
          <Skeleton className="h-10 w-10 rounded-lg" />
          <Skeleton className="h-8 w-16" />
        </div>
        <div className="space-y-2">
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-3/4" />
        </div>
      </CardContent>
    </Card>
  );
}

/**
 * VaultOverviewCard component
 *
 * Renders a summary card with vault health metrics, secret count,
 * and last rotation information.
 */
export function VaultOverviewCard({
  totalSecrets,
  healthStatus = "unknown",
  lastRotationDate,
  isLoading = false,
  onRefresh,
  className,
}: VaultOverviewCardProps) {
  if (isLoading) {
    return <VaultOverviewCardSkeleton className={className} />;
  }

  const health = healthConfig[healthStatus];
  const HealthIcon = health.icon;

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
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-medium text-(--color-text-secondary)">
            Vault Overview
          </CardTitle>
          {onRefresh && (
            <button
              onClick={onRefresh}
              className={cn(
                "p-1.5 rounded-md transition-colors",
                "text-(--color-text-muted)",
                "hover:text-(--color-text-primary)",
                "hover:bg-(--color-bg-tertiary)"
              )}
              aria-label="Refresh vault data"
            >
              <RefreshCw className="h-3.5 w-3.5" />
            </button>
          )}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Secret Count */}
        <div className="flex items-center gap-4">
          <div
            className={cn(
              "flex h-10 w-10 items-center justify-center rounded-lg",
              "bg-gradient-to-br from-(--color-brand-500) to-purple-500"
            )}
          >
            <Key className="h-5 w-5 text-white" />
          </div>
          <div>
            <div className="text-2xl font-bold text-(--color-text-primary)">
              {totalSecrets?.toLocaleString() ?? "-"}
            </div>
            <div className="text-xs text-(--color-text-muted)">
              Total Secrets
            </div>
          </div>
        </div>

        {/* Health Status */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <HealthIcon
              className={cn("h-4 w-4", {
                "text-green-500": healthStatus === "healthy",
                "text-yellow-500": healthStatus === "warning",
                "text-[var(--status-revoked)]": healthStatus === "error",
                "text-(--color-text-muted)": healthStatus === "unknown",
              })}
            />
            <span className="text-sm text-(--color-text-secondary)">
              Status
            </span>
          </div>
          <Badge variant={health.variant}>{health.label}</Badge>
        </div>

        {/* Last Rotation */}
        {lastRotationDate && (
          <div className="flex items-center justify-between pt-2 border-t border-(--border-subtle)">
            <div className="flex items-center gap-2">
              <RefreshCw className="h-4 w-4 text-(--color-text-muted)" />
              <span className="text-sm text-(--color-text-secondary)">
                Last Rotation
              </span>
            </div>
            <span className="text-sm text-(--color-text-primary)">
              {formatDateTime(lastRotationDate)}
            </span>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export default VaultOverviewCard;
