/**
 * FlyNotificationDot.tsx
 *
 * Small pulsing indicator for showing notifications on triggers.
 * Supports multiple variants and sizes with a subtle pulse animation.
 *
 * @module fly-assistant/ui
 */

import React from "react";
import { motion } from "framer-motion";
import { cn } from "@/lib/utils";

// ============================================================================
// Types & Interfaces
// ============================================================================

export interface FlyNotificationDotProps {
  /** Visual variant of the notification */
  variant?: "info" | "warning" | "error" | "success";
  /** Optional count to display (if > 0, shows as badge) */
  count?: number;
  /** Size of the notification dot */
  size?: "sm" | "md" | "lg";
  /** Whether to show pulse animation */
  pulse?: boolean;
  /** Custom className */
  className?: string;
}

// ============================================================================
// Variant Configurations
// ============================================================================

const variantStyles = {
  info: {
    bg: "bg-blue-500",
    shadow: "rgba(59, 130, 246, 0.7)",
  },
  warning: {
    bg: "bg-amber-500",
    shadow: "rgba(245, 158, 11, 0.7)",
  },
  error: {
    bg: "bg-red-500",
    shadow: "rgba(239, 68, 68, 0.7)",
  },
  success: {
    bg: "bg-emerald-500",
    shadow: "rgba(16, 185, 129, 0.7)",
  },
};

const sizeStyles = {
  sm: {
    dot: "w-[6px] h-[6px]",
    badge: "min-w-[14px] h-[14px] text-[9px] px-1",
  },
  md: {
    dot: "w-[10px] h-[10px]",
    badge: "min-w-[18px] h-[18px] text-[10px] px-1.5",
  },
  lg: {
    dot: "w-[14px] h-[14px]",
    badge: "min-w-[22px] h-[22px] text-[11px] px-2",
  },
};

// ============================================================================
// Component
// ============================================================================

export const FlyNotificationDot: React.FC<FlyNotificationDotProps> = ({
  variant = "info",
  count,
  size = "md",
  pulse = true,
  className,
}) => {
  const styles = variantStyles[variant];
  const sizeConfig = sizeStyles[size];
  const hasCount = count !== undefined && count > 0;

  return (
    <motion.div
      className={cn(
        "absolute -top-0.5 -right-0.5 z-10 flex items-center justify-center rounded-full",
        styles.bg,
        hasCount ? sizeConfig.badge : sizeConfig.dot,
        className
      )}
      initial={{ scale: 0 }}
      animate={{ scale: 1 }}
      exit={{ scale: 0 }}
      transition={{ type: "spring", stiffness: 500, damping: 25 }}
      aria-label={hasCount ? `${count} notifications` : "New notification"}
      role="status"
    >
      {/* Pulse ring animation */}
      {pulse && !hasCount && (
        <motion.span
          className={cn(
            "absolute inset-0 rounded-full",
            styles.bg
          )}
          initial={{ opacity: 0.7, scale: 1 }}
          animate={{ opacity: 0, scale: 2 }}
          transition={{
            duration: 2,
            repeat: Infinity,
            ease: "easeOut",
          }}
          style={{
            boxShadow: `0 0 0 0 ${styles.shadow}`,
          }}
        />
      )}

      {/* Count badge text */}
      {hasCount && (
        <span className="text-white font-semibold leading-none">
          {count > 99 ? "99+" : count}
        </span>
      )}
    </motion.div>
  );
};

export default FlyNotificationDot;
