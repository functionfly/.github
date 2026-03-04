/**
 * FlyPermissionGuard.tsx
 *
 * Permission-based feature gating for FlyAssistant.
 * Prevents non-pro users from advanced features and requires
 * confirmation for dangerous actions.
 *
 * @module fly-assistant
 */

import React, { useState, useCallback } from "react";
import { motion, AnimatePresence } from "framer-motion";
import {
  Lock,
  Crown,
  Building2,
  AlertTriangle,
  CheckCircle,
  X,
  ArrowRight,
  Shield,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { UserRole } from "./FlyAssistantProvider";

// ============================================================================
// Types & Interfaces
// ============================================================================

/**
 * Props for FlyPermissionGuard component
 */
export interface FlyPermissionGuardProps {
  /** Required tier for access */
  requiredTier: UserRole;
  /** Current user's tier */
  currentTier: UserRole;
  /** Name of the feature being gated */
  featureName: string;
  /** Content to render if permitted */
  children: React.ReactNode;
  /** Optional fallback content when not permitted */
  fallback?: React.ReactNode;
  /** Whether to show upgrade prompt overlay */
  showUpgradePrompt?: boolean;
  /** Additional CSS classes */
  className?: string;
  /** Callback when upgrade CTA is clicked */
  onUpgrade?: () => void;
}

/**
 * Props for ExecutionGuard component
 */
export interface ExecutionGuardProps {
  /** Action name being performed */
  action: string;
  /** Description of the action */
  description: string;
  /** Risk level of the action */
  riskLevel: "low" | "medium" | "high";
  /** Callback when confirmed */
  onConfirm: () => void;
  /** Callback when cancelled */
  onCancel: () => void;
  /** Custom confirm button text */
  confirmText?: string;
  /** Whether to require re-authentication */
  requireReauthentication?: boolean;
  /** Whether the dialog is open */
  isOpen: boolean;
  /** Affected resources */
  affectedResources?: string[];
}

/**
 * Tier configuration
 */
interface TierConfig {
  icon: React.ReactNode;
  label: string;
  description: string;
  color: string;
  bgColor: string;
  borderColor: string;
}

// ============================================================================
// Tier Configuration
// ============================================================================

const TIER_ORDER: UserRole[] = ["free", "pro", "enterprise"];

const TIER_CONFIG: Record<UserRole, TierConfig> = {
  free: {
    icon: <Lock className="h-5 w-5" />,
    label: "Free",
    description: "Basic features only",
    color: "text-slate-400",
    bgColor: "bg-slate-500/10",
    borderColor: "border-slate-500/30",
  },
  pro: {
    icon: <Crown className="h-5 w-5" />,
    label: "Pro",
    description: "All insights and quick actions",
    color: "text-amber-400",
    bgColor: "bg-amber-500/10",
    borderColor: "border-amber-500/30",
  },
  enterprise: {
    icon: <Building2 className="h-5 w-5" />,
    label: "Enterprise",
    description: "All features + advanced analytics",
    color: "text-indigo-400",
    bgColor: "bg-indigo-500/10",
    borderColor: "border-indigo-500/30",
  },
};

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Check if current tier meets required tier
 */
function hasRequiredTier(currentTier: UserRole, requiredTier: UserRole): boolean {
  const currentIndex = TIER_ORDER.indexOf(currentTier);
  const requiredIndex = TIER_ORDER.indexOf(requiredTier);
  return currentIndex >= requiredIndex;
}

/**
 * Get next tier upgrade
 */
function getNextTier(currentTier: UserRole): UserRole | null {
  const currentIndex = TIER_ORDER.indexOf(currentTier);
  if (currentIndex < TIER_ORDER.length - 1) {
    return TIER_ORDER[currentIndex + 1];
  }
  return null;
}

// ============================================================================
// Components
// ============================================================================

/**
 * FlyPermissionGuard - Feature gating component
 *
 * Wraps content and conditionally renders based on user tier.
 * Shows upgrade prompts for restricted features.
 *
 * @example
 * ```tsx
 * <FlyPermissionGuard
 *   requiredTier="pro"
 *   currentTier={user.role}
 *   featureName="Advanced Insights"
 *   showUpgradePrompt={true}
 *   onUpgrade={() => navigate("/upgrade")}
 * >
 *   <PremiumFeature />
 * </FlyPermissionGuard>
 * ```
 */
export const FlyPermissionGuard: React.FC<FlyPermissionGuardProps> = ({
  requiredTier,
  currentTier,
  featureName,
  children,
  fallback,
  showUpgradePrompt = true,
  className,
  onUpgrade,
}) => {
  const hasAccess = hasRequiredTier(currentTier, requiredTier);

  if (hasAccess) {
    return <>{children}</>;
  }

  // Render custom fallback if provided
  if (fallback) {
    return <>{fallback}</>;
  }

  // Render upgrade prompt
  if (showUpgradePrompt) {
    return (
      <UpgradePrompt
        featureName={featureName}
        requiredTier={requiredTier}
        currentTier={currentTier}
        onUpgrade={onUpgrade}
        className={className}
      />
    );
  }

  return null;
};

/**
 * UpgradePrompt - Visual upgrade CTA component
 */
interface UpgradePromptProps {
  featureName: string;
  requiredTier: UserRole;
  currentTier: UserRole;
  onUpgrade?: () => void;
  className?: string;
}

const UpgradePrompt: React.FC<UpgradePromptProps> = ({
  featureName,
  requiredTier,
  currentTier,
  onUpgrade,
  className,
}) => {
  const nextTier = getNextTier(currentTier);
  const targetTier = hasRequiredTier(nextTier || "free", requiredTier)
    ? requiredTier
    : nextTier;

  if (!targetTier) return null;

  const config = TIER_CONFIG[targetTier];

  return (
    <div
      className={cn(
        "relative overflow-hidden rounded-xl border p-6 text-center",
        "bg-[var(--color-bg-secondary)]/95 backdrop-blur-sm",
        "border-[var(--color-brand-500)]/30",
        className
      )}
      role="region"
      aria-label={`${featureName} - Upgrade required`}
    >
      {/* Background gradient */}
      <div className="absolute inset-0 bg-gradient-to-br from-[var(--color-brand-500)]/5 via-transparent to-transparent" />

      {/* Content */}
      <div className="relative z-10">
        {/* Lock Icon */}
        <div
          className={cn(
            "mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full",
            "bg-[var(--color-brand-500)]/10"
          )}
        >
          <Lock
            className="h-6 w-6 text-[var(--color-brand-500)]"
            aria-hidden="true"
          />
        </div>

        {/* Title */}
        <h3 className="mb-2 text-lg font-semibold text-[var(--color-text-primary)]">
          {featureName}
        </h3>

        {/* Description */}
        <p className="mb-4 text-sm text-[var(--color-text-secondary)]">
          This feature requires a {config.label} plan or higher.
          <br />
          You are currently on the {TIER_CONFIG[currentTier].label} plan.
        </p>

        {/* Tier Badge */}
        <div
          className={cn(
            "mb-4 inline-flex items-center gap-2 rounded-full px-3 py-1.5",
            config.bgColor,
            "border",
            config.borderColor
          )}
        >
          <span className={config.color}>{config.icon}</span>
          <span className={cn("text-sm font-medium", config.color)}>
            {config.label} Required
          </span>
        </div>

        {/* CTA Button */}
        {onUpgrade && (
          <Button
            onClick={onUpgrade}
            className="group bg-[var(--color-brand-600)] text-white hover:bg-[var(--color-brand-500)]"
          >
            Upgrade to {config.label}
            <ArrowRight className="ml-2 h-4 w-4 transition-transform group-hover:translate-x-1" />
          </Button>
        )}
      </div>
    </div>
  );
};

/**
 * ExecutionGuard - Dangerous action confirmation dialog
 *
 * Shows a confirmation dialog for high-risk actions with clear
 * risk indicators and safety checks.
 *
 * @example
 * ```tsx
 * <ExecutionGuard
 *   action="Delete Function"
 *   description="This will permanently delete the function and all its data."
 *   riskLevel="high"
 *   isOpen={showConfirm}
 *   onConfirm={handleDelete}
 *   onCancel={() => setShowConfirm(false)}
 *   affectedResources={["my-function", "prod-deployment"]}
 * />
 * ```
 */
export const ExecutionGuard: React.FC<ExecutionGuardProps> = ({
  action,
  description,
  riskLevel,
  onConfirm,
  onCancel,
  confirmText: confirmButtonText,
  requireReauthentication = false,
  isOpen,
  affectedResources,
}) => {
  const [isReauthenticated, setIsReauthenticated] = useState(!requireReauthentication);
  const [typedConfirmText, setTypedConfirmText] = useState("");

  const handleConfirm = useCallback(() => {
    if (requireReauthentication && !isReauthenticated) {
      return;
    }
    onConfirm();
    setTypedConfirmText("");
    setIsReauthenticated(!requireReauthentication);
  }, [onConfirm, requireReauthentication, isReauthenticated]);

  const handleCancel = useCallback(() => {
    onCancel();
    setTypedConfirmText("");
    setIsReauthenticated(!requireReauthentication);
  }, [onCancel, requireReauthentication]);

  // Risk level configuration
  const riskConfig = {
    low: {
      icon: <CheckCircle className="h-5 w-5 text-emerald-500" />,
      title: "Confirm Action",
      buttonVariant: "default" as const,
      buttonClass: "bg-emerald-600 hover:bg-emerald-500",
    },
    medium: {
      icon: <AlertTriangle className="h-5 w-5 text-amber-500" />,
      title: "Review Before Proceeding",
      buttonVariant: "default" as const,
      buttonClass: "bg-amber-600 hover:bg-amber-500",
    },
    high: {
      icon: <Shield className="h-5 w-5 text-red-500" />,
      title: "High Risk Action",
      buttonVariant: "destructive" as const,
      buttonClass: "",
    },
  };

  const config = riskConfig[riskLevel];
  const requiredConfirmText = action.toLowerCase().replace(/\s+/g, "-");
  const isConfirmValid = typedConfirmText.toLowerCase() === requiredConfirmText;

  return (
    <Dialog open={isOpen} onOpenChange={handleCancel}>
      <DialogContent
        className="sm:max-w-md"
        aria-describedby="execution-description"
      >
        <DialogHeader>
          <div className="flex items-center gap-3">
            {config.icon}
            <DialogTitle className="text-[var(--color-text-primary)]">
              {config.title}
            </DialogTitle>
          </div>
          <DialogDescription
            id="execution-description"
            className="pt-2 text-[var(--color-text-secondary)]"
          >
            {description}
          </DialogDescription>
        </DialogHeader>

        {/* Affected Resources */}
        {affectedResources && affectedResources.length > 0 && (
          <div className="rounded-lg bg-[var(--color-bg-tertiary)] p-3">
            <p className="mb-2 text-xs font-medium uppercase tracking-wide text-[var(--color-text-muted)]">
              Affected Resources
            </p>
            <ul className="space-y-1">
              {affectedResources.map((resource, index) => (
                <li
                  key={index}
                  className="flex items-center gap-2 text-sm text-[var(--color-text-secondary)]"
                >
                  <span className="h-1.5 w-1.5 rounded-full bg-[var(--color-brand-500)]" />
                  {resource}
                </li>
              ))}
            </ul>
          </div>
        )}

        {/* Risk Warning for high risk */}
        {riskLevel === "high" && (
          <div className="rounded-lg border border-red-500/30 bg-red-500/10 p-3">
            <p className="flex items-center gap-2 text-sm font-medium text-red-600 dark:text-red-400">
              <AlertTriangle className="h-4 w-4" />
              This action cannot be undone
            </p>
          </div>
        )}

        {/* Re-authentication */}
        {requireReauthentication && (
          <div className="space-y-3">
            <p className="text-sm text-[var(--color-text-secondary)]">
              Please type <code className="rounded bg-[var(--color-bg-tertiary)] px-1.5 py-0.5 text-[var(--color-brand-400)]">
                {requiredConfirmText}
              </code> to confirm:
            </p>
            <input
              type="text"
              value={typedConfirmText}
              onChange={(e) => setTypedConfirmText(e.target.value)}
              placeholder={`Type "${requiredConfirmText}" to confirm`}
              className={cn(
                "w-full rounded-lg border bg-transparent px-3 py-2 text-sm",
                "border-[var(--color-border-primary)]",
                "text-[var(--color-text-primary)]",
                "placeholder:text-[var(--color-text-muted)]",
                "focus:border-[var(--color-brand-500)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand-500)]"
              )}
              aria-label="Confirmation text"
            />
          </div>
        )}

        <DialogFooter className="gap-2 sm:gap-0">
          <Button
            variant="outline"
            onClick={handleCancel}
            className="border-[var(--color-border-primary)] text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]"
          >
            <X className="mr-2 h-4 w-4" />
            Cancel
          </Button>
          <Button
            variant={config.buttonVariant}
            onClick={handleConfirm}
            disabled={requireReauthentication && !isConfirmValid}
            className={cn(
              "text-white",
              config.buttonClass
            )}
          >
            <CheckCircle className="mr-2 h-4 w-4" />
            {confirmButtonText || `Confirm ${action}`}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

/**
 * TierBadge - Display current user tier
 */
export interface TierBadgeProps {
  tier: UserRole;
  showLabel?: boolean;
  size?: "sm" | "md" | "lg";
  className?: string;
}

export const TierBadge: React.FC<TierBadgeProps> = ({
  tier,
  showLabel = true,
  size = "md",
  className,
}) => {
  const config = TIER_CONFIG[tier];

  const sizeClasses = {
    sm: "h-6 px-2 text-xs",
    md: "h-8 px-3 text-sm",
    lg: "h-10 px-4 text-base",
  };

  const iconSizes = {
    sm: "h-3 w-3",
    md: "h-4 w-4",
    lg: "h-5 w-5",
  };

  return (
    <div
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full font-medium",
        sizeClasses[size],
        config.bgColor,
        config.borderColor,
        "border",
        config.color,
        className
      )}
    >
      <span className={iconSizes[size]}>{config.icon}</span>
      {showLabel && <span>{config.label}</span>}
    </div>
  );
};

export default FlyPermissionGuard;
