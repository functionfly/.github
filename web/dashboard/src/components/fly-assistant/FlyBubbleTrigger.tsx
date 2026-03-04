/**
 * FlyBubbleTrigger.tsx
 *
 * The floating action button that triggers the FlyAssistant panel.
 * Features animated pulse effects, notification badges, trust score glow,
 * and error states with full accessibility support.
 */

import React, { useMemo } from "react";
import { motion, AnimatePresence, type Variants, type TargetAndTransition } from "framer-motion";
import { Sparkles, AlertCircle } from "lucide-react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { type TrustTier } from "./FlyAssistantProvider";

// ============================================================================
// Types & Interfaces
// ============================================================================

export interface FlyBubbleTriggerProps {
  /** Click handler for opening the assistant */
  onClick: () => void;
  /** Whether insights are available to view */
  hasInsights?: boolean;
  /** Number of unread notifications */
  notificationCount?: number;
  /** Current trust tier (affects glow color) */
  trustTier?: TrustTier;
  /** Whether there's an active error state */
  hasError?: boolean;
  /** Optional CSS class names */
  className?: string;
  /** Optional test ID */
  testId?: string;
  /** Optional aria-label override */
  ariaLabel?: string;
  /** Whether the button is disabled */
  disabled?: boolean;
}

// ============================================================================
// Animation Variants
// ============================================================================

const buttonVariants: Variants = {
  initial: { scale: 0, opacity: 0 },
  animate: {
    scale: 1,
    opacity: 1,
    transition: {
      type: "spring" as const,
      stiffness: 260,
      damping: 20,
    }
  },
  exit: {
    scale: 0,
    opacity: 0,
    transition: { duration: 0.2 }
  },
  hover: {
    scale: 1.1,
    transition: { duration: 0.2 }
  },
  tap: {
    scale: 0.95,
    transition: { duration: 0.1 }
  },
};

const pulseVariants: Variants = {
  initial: { scale: 1, opacity: 0.5 },
  animate: {
    scale: 1.5,
    opacity: 0,
    transition: {
      duration: 1.5,
      repeat: Infinity,
      ease: "easeOut" as const,
    },
  },
};

const glowVariants: Record<TrustTier, TargetAndTransition> = {
  low: {
    boxShadow: "0 0 20px rgba(239, 68, 68, 0.5), 0 0 40px rgba(239, 68, 68, 0.3)",
  },
  medium: {
    boxShadow: "0 0 20px rgba(245, 158, 11, 0.5), 0 0 40px rgba(245, 158, 11, 0.3)",
  },
  high: {
    boxShadow: "0 0 20px rgba(99, 102, 241, 0.6), 0 0 40px rgba(99, 102, 241, 0.4)",
  },
  critical: {
    boxShadow: "0 0 30px rgba(16, 185, 129, 0.7), 0 0 60px rgba(16, 185, 129, 0.5)",
  },
};

const iconVariants: Variants = {
  initial: { rotate: -180, opacity: 0 },
  animate: {
    rotate: 0,
    opacity: 1,
    transition: {
      type: "spring" as const,
      stiffness: 200,
      damping: 15,
    }
  },
};

const badgeVariants: Variants = {
  initial: { scale: 0, y: 10 },
  animate: {
    scale: 1,
    y: 0,
    transition: {
      type: "spring" as const,
      stiffness: 500,
      damping: 25,
    }
  },
  exit: {
    scale: 0,
    y: 10,
    transition: { duration: 0.15 }
  },
};

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Get ring color class based on trust tier
 */
function getTrustTierColor(tier: TrustTier): string {
  switch (tier) {
    case "low":
      return "ring-2 ring-error ring-offset-2 ring-offset-bg-primary";
    case "medium":
      return "ring-2 ring-warning ring-offset-2 ring-offset-bg-primary";
    case "high":
      return "ring-2 ring-brand-500 ring-offset-2 ring-offset-bg-primary";
    case "critical":
      return "ring-3 ring-success ring-offset-2 ring-offset-bg-primary";
    default:
      return "ring-2 ring-brand-500 ring-offset-2 ring-offset-bg-primary";
  }
}

/**
 * Get glow animation variant based on trust tier
 */
function getGlowVariant(tier: TrustTier): TargetAndTransition {
  return glowVariants[tier] || glowVariants.high;
}

/**
 * Format notification count (max 99+)
 */
function formatNotificationCount(count: number): string {
  if (count > 99) return "99+";
  return String(count);
}

// ============================================================================
// Component
// ============================================================================

/**
 * FlyBubbleTrigger - Floating action button for FlyAssistant
 *
 * A fixed-position button in the bottom-right corner that triggers the
 * assistant panel. Features include:
 * - Animated pulse when insights are available
 * - Notification badge with count
 * - Trust score glow ring (color-coded)
 * - Error alert state (red indicator)
 * - Smooth hover/tap animations
 * - Full keyboard accessibility
 *
 * @example
 * ```tsx
 * <FlyBubbleTrigger
 *   onClick={() => setIsOpen(true)}
 *   hasInsights={true}
 *   notificationCount={3}
 *   trustTier="high"
 * />
 * ```
 */
export function FlyBubbleTrigger({
  onClick,
  hasInsights = false,
  notificationCount = 0,
  trustTier = "medium",
  hasError = false,
  className,
  testId = "fly-bubble-trigger",
  ariaLabel,
  disabled = false,
}: FlyBubbleTriggerProps) {
  // Determine effective trust tier (error overrides)
  const effectiveTrustTier = hasError ? "low" : trustTier;

  // Memoize computed values
  const showBadge = notificationCount > 0;
  const showPulse = hasInsights && !hasError;
  const buttonLabel = useMemo(() => {
    if (ariaLabel) return ariaLabel;
    if (hasError) return "FlyAssistant - Error state, click to view";
    if (notificationCount > 0) {
      return `FlyAssistant - ${notificationCount} notification${notificationCount > 1 ? "s" : ""}`;
    }
    return "Open FlyAssistant AI Assistant";
  }, [ariaLabel, hasError, notificationCount]);

  return (
    <motion.button
      data-testid={testId}
      onClick={onClick}
      disabled={disabled}
      className={cn(
        // Base styles
        "fixed bottom-6 right-6 z-50",
        "w-14 h-14 rounded-full",
        "flex items-center justify-center",
        "bg-gradient-to-r from-brand-500 via-purple-500 to-pink-500",
        "text-white shadow-lg",
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-border-focus focus-visible:ring-offset-2",
        "disabled:opacity-50 disabled:cursor-not-allowed",
        "pointer-events-auto", // Ensure clicks work in portal
        // Trust tier ring
        getTrustTierColor(effectiveTrustTier),
        className
      )}
      style={{
        // @ts-expect-error CSS custom property
        "--tw-ring-offset-color": "var(--color-bg-primary)",
      }}
      variants={buttonVariants}
      initial="initial"
      animate="animate"
      exit="exit"
      whileHover={disabled ? undefined : "hover"}
      whileTap={disabled ? undefined : "tap"}
      aria-label={buttonLabel}
      aria-haspopup="dialog"
      aria-expanded={false}
    >
      {/* Pulse animation ring (when insights available) */}
      <AnimatePresence>
        {showPulse && (
          <motion.span
            className="absolute inset-0 rounded-full bg-brand-500"
            variants={pulseVariants}
            initial="initial"
            animate="animate"
            exit={{ opacity: 0 }}
            aria-hidden="true"
          />
        )}
      </AnimatePresence>

      {/* Glow effect based on trust tier */}
      <motion.div
        className="absolute inset-0 rounded-full"
        animate={getGlowVariant(effectiveTrustTier)}
        transition={{ duration: 0.3 }}
        aria-hidden="true"
      />

      {/* Icon container */}
      <motion.div
        className="relative flex items-center justify-center"
        variants={iconVariants}
      >
        {hasError ? (
          <AlertCircle className="w-6 h-6 text-white" aria-hidden="true" />
        ) : (
          <Sparkles className="w-6 h-6 text-white" aria-hidden="true" />
        )}
      </motion.div>

      {/* Notification badge */}
      <AnimatePresence>
        {showBadge && (
          <motion.div
            className="absolute -top-1 -right-1"
            variants={badgeVariants}
            initial="initial"
            animate="animate"
            exit="exit"
          >
            <Badge
              variant="destructive"
              className="h-5 min-w-[20px] px-1.5 text-[10px] font-bold flex items-center justify-center"
            >
              {formatNotificationCount(notificationCount)}
            </Badge>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Error indicator dot */}
      <AnimatePresence>
        {hasError && !showBadge && (
          <motion.span
            className="absolute -top-1 -right-1 w-3 h-3 rounded-full bg-error"
            initial={{ scale: 0 }}
            animate={{
              scale: 1,
              transition: { type: "spring", stiffness: 500, damping: 25 }
            }}
            exit={{ scale: 0 }}
            aria-hidden="true"
          />
        )}
      </AnimatePresence>
    </motion.button>
  );
}

export default FlyBubbleTrigger;
