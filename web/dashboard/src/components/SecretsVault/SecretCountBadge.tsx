/**
 * SecretCountBadge - Badge component for displaying secret counts with color variants
 *
 * A reusable badge that displays secret counts with visual indicators
 * for different count thresholds. Useful for lists, headers, and summary views.
 *
 * @example
 * ```tsx
 * // Default usage
 * <SecretCountBadge count={42} />
 *
 * // With label
 * <SecretCountBadge count={15} label="API Keys" />
 *
 * // Loading state
 * <SecretCountBadge isLoading />
 *
 * // Custom size
 * <SecretCountBadge count={8} size="lg" />
 *
 * // With icon
 * <SecretCountBadge count={23} showIcon />
 *
 * // As a button with onClick
 * <SecretCountBadge count={5} onClick={() => navigate('/secrets')} />
 * ```
 */

import { Key, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { cva, type VariantProps } from "class-variance-authority";
import type { SecretType } from "@/types/vault";

/** Size variants for the badge */
const sizeVariants = cva("", {
  variants: {
    size: {
      sm: "text-xs px-2 py-0.5",
      md: "text-sm px-2.5 py-0.5",
      lg: "text-base px-3 py-1",
    },
  },
  defaultVariants: {
    size: "md",
  },
});

/** Count threshold levels for automatic variant selection */
export type CountThreshold = "none" | "low" | "medium" | "high";

export interface SecretCountBadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof sizeVariants> {
  /** The secret count to display */
  count?: number;
  /** Optional label to display next to the count */
  label?: string;
  /** Whether the badge is in loading state */
  isLoading?: boolean;
  /** Override the auto-detected threshold variant */
  threshold?: CountThreshold;
  /** Thresholds for determining badge color (default: low: 10, medium: 50, high: 100) */
  thresholds?: {
    low: number;
    medium: number;
    high: number;
  };
  /** Show the key icon */
  showIcon?: boolean;
  /** Make the badge clickable */
  onClick?: () => void;
  /** Secret type for type-specific styling */
  secretType?: SecretType;
  /** Maximum count before showing "99+" format */
  maxCount?: number;
}

/**
 * Determine threshold level based on count
 */
function getThresholdLevel(
  count: number,
  thresholds: { low: number; medium: number; high: number }
): CountThreshold {
  if (count === 0) return "none";
  if (count < thresholds.low) return "low";
  if (count < thresholds.medium) return "medium";
  return "high";
}

/**
 * Get badge variant based on threshold level
 */
function getVariantForThreshold(
  threshold: CountThreshold
): React.ComponentProps<typeof Badge>["variant"] {
  switch (threshold) {
    case "none":
      return "secondary";
    case "low":
      return "outline";
    case "medium":
      return "default";
    case "high":
      return "success";
    default:
      return "secondary";
  }
}

/**
 * Get secret type specific color class
 */
function getSecretTypeColor(secretType?: SecretType): string {
  switch (secretType) {
    case "api_key":
      return "text-blue-500";
    case "oauth_token":
      return "text-green-500";
    case "password":
      return "text-yellow-500";
    case "certificate":
      return "text-purple-500";
    default:
      return "text-(--color-text-primary)";
  }
}

/**
 * SecretCountBadge component
 *
 * Renders a badge displaying a secret count with optional icon, label,
 * and automatic color coding based on count thresholds.
 */
export function SecretCountBadge({
  count,
  label,
  isLoading = false,
  threshold,
  thresholds = { low: 10, medium: 50, high: 100 },
  size = "md",
  showIcon = false,
  onClick,
  secretType,
  maxCount = 99,
  className,
  ...props
}: SecretCountBadgeProps) {
  // Determine display count (with max limit)
  const displayCount =
    count === undefined
      ? "-"
      : count > maxCount
      ? `${maxCount}+`
      : count.toString();

  // Determine threshold level
  const thresholdLevel =
    threshold ?? (count !== undefined ? getThresholdLevel(count, thresholds) : "none");

  // Get badge variant
  const variant = getVariantForThreshold(thresholdLevel);

  // Check if clickable
   const isClickable = !!onClick && !isLoading;

  const content = (
    <>
      {isLoading ? (
        <Loader2 className="h-3 w-3 animate-spin" />
      ) : showIcon ? (
        <Key
          className={cn("h-3 w-3", getSecretTypeColor(secretType))}
          aria-hidden="true"
        />
      ) : null}
      <span className="font-medium">{isLoading ? "" : displayCount}</span>
      {label && (
        <span className="text-(--color-text-muted) font-normal">{label}</span>
      )}
    </>
  );

  return (
    <Badge
      variant={variant}
      className={cn(
        "inline-flex items-center gap-1.5 transition-all duration-200",
        sizeVariants({ size }),
        {
          "cursor-pointer hover:opacity-80 active:scale-95": isClickable,
          "cursor-wait": isLoading,
        },
        className
      )}
      onClick={isClickable ? onClick : undefined}
      role={isClickable ? "button" : undefined}
      tabIndex={isClickable ? 0 : undefined}
      onKeyDown={
        isClickable
          ? (e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                onClick?.();
              }
            }
          : undefined
      }
      {...props}
    >
      {content}
    </Badge>
  );
}

/**
 * GroupedSecretCountBadge - Displays multiple count badges grouped together
 *
 * @example
 * ```tsx
 * <GroupedSecretCountBadge
 *   counts={{
 *     api_key: 15,
 *     oauth_token: 8,
 *     password: 23,
 *     certificate: 3,
 *   }}
 * />
 * ```
 */
export interface GroupedSecretCountBadgeProps {
  /** Counts per secret type */
  counts?: Partial<Record<SecretType, number>>;
  /** Whether loading */
  isLoading?: boolean;
  /** Size variant */
  size?: "sm" | "md" | "lg";
  /** Additional CSS classes */
  className?: string;
}

const secretTypeLabels: Record<SecretType, string> = {
  api_key: "API Keys",
  oauth_token: "OAuth",
  password: "Passwords",
  certificate: "Certs",
};

export function GroupedSecretCountBadge({
  counts = {},
  isLoading = false,
  size = "sm",
  className,
}: GroupedSecretCountBadgeProps) {
  const secretTypes: SecretType[] = ["api_key", "oauth_token", "password", "certificate"];

  return (
    <div className={cn("flex flex-wrap gap-2", className)}>
      {secretTypes.map((type) => (
        <SecretCountBadge
          key={type}
          count={counts[type] ?? 0}
          label={secretTypeLabels[type]}
          secretType={type}
          size={size}
          isLoading={isLoading}
          showIcon
        />
      ))}
    </div>
  );
}

export default SecretCountBadge;
