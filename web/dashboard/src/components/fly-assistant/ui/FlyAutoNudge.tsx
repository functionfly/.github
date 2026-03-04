/**
 * FlyAutoNudge.tsx
 *
 * Subtle slide-in micro-suggestion that appears without fully opening chat.
 * Perfect for proactive notifications and gentle user guidance.
 *
 * @module fly-assistant/ui
 */

import React, { useEffect, useCallback } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { X, ArrowRight, AlertTriangle, CheckCircle, Info } from "lucide-react";
import { cn } from "@/lib/utils";

// ============================================================================
// Types & Interfaces
// ============================================================================

export interface FlyAutoNudgeProps {
  /** Short message text (max 60 chars recommended) */
  message: string;
  /** Optional action button label */
  actionLabel?: string;
  /** Callback when action is clicked */
  onAction?: () => void;
  /** Callback when dismissed */
  onDismiss?: () => void;
  /** Auto-dismiss delay in ms (0 to disable) */
  autoDismissDelay?: number;
  /** Visual variant */
  variant?: "info" | "warning" | "success";
  /** Whether the nudge is visible */
  isOpen?: boolean;
  /** Custom className */
  className?: string;
}

// ============================================================================
// Variant Configurations
// ============================================================================

const variantStyles = {
  info: {
    icon: Info,
    iconColor: "text-blue-400",
    borderColor: "border-blue-500/30",
    bgColor: "bg-[var(--color-bg-tertiary)]",
  },
  warning: {
    icon: AlertTriangle,
    iconColor: "text-amber-400",
    borderColor: "border-amber-500/30",
    bgColor: "bg-[var(--color-bg-tertiary)]",
  },
  success: {
    icon: CheckCircle,
    iconColor: "text-emerald-400",
    borderColor: "border-emerald-500/30",
    bgColor: "bg-[var(--color-bg-tertiary)]",
  },
};

// ============================================================================
// Component
// ============================================================================

export const FlyAutoNudge: React.FC<FlyAutoNudgeProps> = ({
  message,
  actionLabel,
  onAction,
  onDismiss,
  autoDismissDelay = 5000,
  variant = "info",
  isOpen = true,
  className,
}) => {
  const styles = variantStyles[variant];
  const Icon = styles.icon;

  // Auto-dismiss timer
  useEffect(() => {
    if (!isOpen || autoDismissDelay <= 0) return;

    const timer = setTimeout(() => {
      onDismiss?.();
    }, autoDismissDelay);

    return () => clearTimeout(timer);
  }, [isOpen, autoDismissDelay, onDismiss]);

  // Handle action click
  const handleAction = useCallback(() => {
    onAction?.();
  }, [onAction]);

  // Handle dismiss
  const handleDismiss = useCallback(() => {
    onDismiss?.();
  }, [onDismiss]);

  return (
    <AnimatePresence>
      {isOpen && (
        <motion.div
          initial={{ opacity: 0, y: 20, scale: 0.95 }}
          animate={{ opacity: 1, y: 0, scale: 1 }}
          exit={{ opacity: 0, y: 10, scale: 0.95 }}
          transition={{
            type: "spring",
            stiffness: 400,
            damping: 30,
          }}
          className={cn(
            "relative w-full max-w-[280px] rounded-lg border px-3.5 py-2.5 shadow-lg",
            "bg-[var(--color-bg-tertiary)] border-[var(--color-border)]",
            styles.borderColor,
            className
          )}
          role="alert"
          aria-live="polite"
        >
          {/* Arrow pointing down to trigger */}
          <div
            className="absolute -bottom-1.5 right-5 w-3 h-3 rotate-45"
            style={{
              backgroundColor: "var(--color-bg-tertiary)",
              borderRight: "1px solid var(--color-border)",
              borderBottom: "1px solid var(--color-border)",
            }}
            aria-hidden="true"
          />

          {/* Content */}
          <div className="flex items-start gap-2.5">
            {/* Icon */}
            <Icon
              className={cn("w-4 h-4 shrink-0 mt-0.5", styles.iconColor)}
              aria-hidden="true"
            />

            {/* Message */}
            <div className="flex-1 min-w-0">
              <p className="text-sm text-[var(--color-text-primary)] leading-relaxed">
                {message}
              </p>

              {/* Action button */}
              {actionLabel && (
                <button
                  onClick={handleAction}
                  className={cn(
                    "mt-2 flex items-center gap-1 text-xs font-medium",
                    "text-[var(--color-brand-500)] hover:text-[var(--color-brand-400)]",
                    "transition-colors focus:outline-none focus:ring-2",
                    "focus:ring-[var(--color-brand-500)]/30 rounded px-1 -ml-1"
                  )}
                  aria-label={`${actionLabel}: ${message}`}
                >
                  {actionLabel}
                  <ArrowRight className="w-3 h-3" aria-hidden="true" />
                </button>
              )}
            </div>

            {/* Dismiss button */}
            <button
              onClick={handleDismiss}
              className={cn(
                "shrink-0 p-0.5 rounded",
                "text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)]",
                "hover:bg-[var(--color-bg-secondary)] transition-colors",
                "focus:outline-none focus:ring-2 focus:ring-[var(--color-border-focus)]"
              )}
              aria-label="Dismiss notification"
            >
              <X className="w-3.5 h-3.5" aria-hidden="true" />
            </button>
          </div>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default FlyAutoNudge;
