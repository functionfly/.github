import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';
import { Shield, ShieldAlert, ShieldCheck, ShieldX, type LucideIcon } from 'lucide-react';

/**
 * Verification level types for trust badges
 */
export type VerificationLevel = 'verified' | 'highly_trusted' | 'untrusted' | 'pending' | 'basic';

/**
 * Trust tier mapping to verification level
 */
export const TRUST_TIER_TO_VERIFICATION: Record<string, VerificationLevel> = {
  critical: 'highly_trusted',
  high: 'verified',
  medium: 'basic',
  low: 'pending',
  untrusted: 'untrusted',
};

/**
 * Get badge configuration based on verification level
 */
function getVerificationConfig(level: VerificationLevel): {
  icon: LucideIcon;
  label: string;
  colorClass: string;
  bgClass: string;
  borderClass: string;
  description: string;
} {
  const configs = {
    highly_trusted: {
      icon: ShieldCheck,
      label: 'Highly Trusted',
      colorClass: 'text-emerald-400',
      bgClass: 'bg-emerald-500/10',
      borderClass: 'border-emerald-500/40',
      description:
        'Maximum security - Fully verified with multi-factor authentication and identity proof',
    },
    verified: {
      icon: ShieldCheck,
      label: 'Verified',
      colorClass: 'text-blue-400',
      bgClass: 'bg-blue-500/10',
      borderClass: 'border-blue-500/40',
      description: 'Identity verified - Core verification completed',
    },
    basic: {
      icon: Shield,
      label: 'Basic',
      colorClass: 'text-amber-400',
      bgClass: 'bg-amber-500/10',
      borderClass: 'border-amber-500/40',
      description: 'Basic verification - Minimal identity checks completed',
    },
    pending: {
      icon: ShieldAlert,
      label: 'Pending',
      colorClass: 'text-orange-400',
      bgClass: 'bg-orange-500/10',
      borderClass: 'border-orange-500/40',
      description: 'Verification pending - Awaiting review',
    },
    untrusted: {
      icon: ShieldX,
      label: 'Unverified',
      colorClass: 'text-red-400',
      bgClass: 'bg-red-500/10',
      borderClass: 'border-red-500/40',
      description: 'Not verified - Exercise caution',
    },
  };
  return configs[level];
}

/**
 * TrustBadge Component Props
 */
export interface TrustBadgeProps {
  /** Verification level to display */
  level: VerificationLevel;
  /** Display variant */
  variant?: 'inline' | 'full' | 'icon-only';
  /** Size variant */
  size?: 'sm' | 'md' | 'lg';
  /** Additional CSS classes */
  className?: string;
  /** Show tooltip with details */
  showTooltip?: boolean;
  /** Custom click handler */
  onClick?: () => void;
}

/**
 * TrustBadge Component
 *
 * Displays verification level badge with appropriate icon and styling.
 * Supports three variants: inline (compact text), full (with label), icon-only.
 *
 * @example
 * // Full badge with tooltip
 * <TrustBadge level="verified" variant="full" showTooltip />
 *
 * // Compact inline badge
 * <TrustBadge level="highly_trusted" variant="inline" size="sm" />
 *
 * // Icon only with click handler
 * <TrustBadge level="untrusted" variant="icon-only" onClick={handleClick} />
 */
export function TrustBadge({
  level,
  variant = 'inline',
  size = 'md',
  className,
  showTooltip = false,
  onClick,
}: TrustBadgeProps) {
  const config = getVerificationConfig(level);
  const Icon = config.icon;

  const sizeClasses = {
    sm: 'px-1.5 py-0.5 text-[10px]',
    md: 'px-2 py-1 text-xs',
    lg: 'px-3 py-1.5 text-sm',
  };

  const iconSizes = {
    sm: 'h-3 w-3',
    md: 'h-3.5 w-3.5',
    lg: 'h-4 w-4',
  };

  const badgeContent = (
    <div
      className={cn(
        // Base styles
        'inline-flex items-center gap-1.5 rounded-full border font-medium transition-all duration-200',
        // Variant styles
        variant === 'inline' && sizeClasses[size],
        variant === 'full' && 'px-3 py-1.5',
        variant === 'icon-only' && 'p-1.5',
        // Color styles
        config.bgClass,
        config.borderClass,
        // Hover effect
        'hover:shadow-md',
        // Clickable
        onClick && 'cursor-pointer hover:scale-105',
        className
      )}
      onClick={onClick}
      role={onClick ? 'button' : undefined}
      tabIndex={onClick ? 0 : undefined}
      onKeyDown={onClick ? (e) => e.key === 'Enter' && onClick() : undefined}
    >
      <Icon className={cn(iconSizes[size], config.colorClass)} aria-hidden="true" />

      {variant !== 'icon-only' && <span className={cn(config.colorClass)}>{config.label}</span>}
    </div>
  );

  if (showTooltip) {
    return (
      <TooltipProvider delayDuration={200}>
        <Tooltip>
          <TooltipTrigger asChild>{badgeContent}</TooltipTrigger>
          <TooltipContent side="top" className="max-w-xs">
            <div className="space-y-1">
              <p className="font-semibold">{config.label}</p>
              <p className="text-xs text-muted-foreground">{config.description}</p>
            </div>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    );
  }

  return badgeContent;
}

/**
 * TrustBadgeSkeleton Component
 * Loading placeholder for TrustBadge
 */
export function TrustBadgeSkeleton({
  variant = 'inline',
  className,
}: {
  variant?: 'inline' | 'full' | 'icon-only';
  className?: string;
}) {
  return (
    <div
      className={cn(
        'animate-pulse bg-muted rounded-full',
        variant === 'icon-only' && 'h-8 w-8',
        variant === 'inline' && 'h-6 w-24',
        variant === 'full' && 'h-8 w-32',
        className
      )}
      aria-hidden="true"
    />
  );
}

export default TrustBadge;
