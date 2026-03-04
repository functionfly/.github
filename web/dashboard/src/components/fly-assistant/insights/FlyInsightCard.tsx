/**
 * FlyInsightCard.tsx
 *
 * Proactive insight cards for performance, monetization, quality,
 * security, and usage insights. Supports multiple visual variants
 * with interactive actions and dismissal.
 */

import React, { useCallback, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import {
  Zap,
  DollarSign,
  Sparkles,
  ShieldAlert,
  BarChart3,
  X,
  ArrowRight,
  Check,
  Info,
  AlertTriangle,
  CheckCircle2,
  AlertOctagon,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";

// ============================================================================
// Types & Interfaces
// ============================================================================

export type InsightType = "info" | "warning" | "success" | "error";

export interface InsightAction {
  id: string;
  label: string;
  variant?: "primary" | "secondary" | "ghost" | "destructive";
  icon?: React.ReactNode;
  onClick: () => void;
}

export interface Insight {
  id: string;
  title: string;
  description: string;
  type: InsightType;
  dismissible: boolean;
  actions?: InsightAction[];
}

export interface FlyInsightCardProps {
  insight: Insight;
  onDismiss?: (id: string, dontShowAgain: boolean) => void;
  className?: string;
}

// ============================================================================
// Configuration
// ============================================================================

interface InsightConfig {
  icon: React.ReactNode;
  colors: {
    bg: string;
    border: string;
    iconBg: string;
    iconColor: string;
    title: string;
    description: string;
  };
}

const insightConfigs: Record<InsightType, InsightConfig> = {
  info: {
    icon: <Info className="h-5 w-5" />,
    colors: {
      bg: "bg-blue-500/10",
      border: "border-blue-500/30",
      iconBg: "bg-blue-500/20",
      iconColor: "text-blue-500",
      title: "text-blue-900 dark:text-blue-100",
      description: "text-blue-700 dark:text-blue-300",
    },
  },
  warning: {
    icon: <AlertTriangle className="h-5 w-5" />,
    colors: {
      bg: "bg-amber-500/10",
      border: "border-amber-500/30",
      iconBg: "bg-amber-500/20",
      iconColor: "text-amber-500",
      title: "text-amber-900 dark:text-amber-100",
      description: "text-amber-700 dark:text-amber-300",
    },
  },
  success: {
    icon: <CheckCircle2 className="h-5 w-5" />,
    colors: {
      bg: "bg-emerald-500/10",
      border: "border-emerald-500/30",
      iconBg: "bg-emerald-500/20",
      iconColor: "text-emerald-500",
      title: "text-emerald-900 dark:text-emerald-100",
      description: "text-emerald-700 dark:text-emerald-300",
    },
  },
  error: {
    icon: <AlertOctagon className="h-5 w-5" />,
    colors: {
      bg: "bg-red-500/10",
      border: "border-red-500/30",
      iconBg: "bg-red-500/20",
      iconColor: "text-red-500",
      title: "text-red-900 dark:text-red-100",
      description: "text-red-700 dark:text-red-300",
    },
  },
};

// ============================================================================
// Helper Functions
// ============================================================================

function getInsightIcon(type: InsightType): React.ReactNode {
  const icons: Record<InsightType, React.ReactNode> = {
    info: <Info className="h-5 w-5" />,
    warning: <AlertTriangle className="h-5 w-5" />,
    success: <CheckCircle2 className="h-5 w-5" />,
    error: <AlertOctagon className="h-5 w-5" />,
  };
  return icons[type];
}

// ============================================================================
// Component
// ============================================================================

export const FlyInsightCard: React.FC<FlyInsightCardProps> = ({
  insight,
  onDismiss,
  className,
}) => {
  const { id, title, description, type, dismissible, actions } = insight;
  const config = insightConfigs[type];
  const [isDismissed, setIsDismissed] = useState(false);
  const [dontShowAgain, setDontShowAgain] = useState(false);
  const [showDismissConfirm, setShowDismissConfirm] = useState(false);

  const handleDismiss = useCallback(() => {
    if (dontShowAgain) {
      onDismiss?.(id, true);
      setIsDismissed(true);
    } else {
      setShowDismissConfirm(true);
    }
  }, [id, dontShowAgain, onDismiss]);

  const handleConfirmDismiss = useCallback(() => {
    onDismiss?.(id, dontShowAgain);
    setIsDismissed(true);
  }, [id, dontShowAgain, onDismiss]);

  const handleCancelDismiss = useCallback(() => {
    setShowDismissConfirm(false);
    setDontShowAgain(false);
  }, []);

  if (isDismissed) {
    return null;
  }

  return (
    <AnimatePresence>
      <motion.div
        className={cn(
          "relative overflow-hidden",
          "rounded-xl p-4",
          "border",
          config.colors.bg,
          config.colors.border,
          className
        )}
        initial={{ opacity: 0, y: 10, scale: 0.98 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        exit={{ opacity: 0, scale: 0.95, y: -10 }}
        transition={{ duration: 0.25, ease: [0.23, 1, 0.32, 1] }}
        role="alert"
        aria-live="polite"
      >
        {/* Dismiss Confirmation Overlay */}
        <AnimatePresence>
          {showDismissConfirm && (
            <motion.div
              className="absolute inset-0 z-10 flex flex-col items-center justify-center gap-3 bg-inherit backdrop-blur-sm"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
            >
              <p className="text-sm font-medium text-[var(--color-text-primary)]">
                Dismiss this insight?
              </p>
              <div className="flex items-center gap-2">
                <Checkbox
                  id={`dont-show-${id}`}
                  checked={dontShowAgain}
                  onCheckedChange={(checked) => setDontShowAgain(checked === true)}
                />
                <label
                  htmlFor={`dont-show-${id}`}
                  className="text-xs text-[var(--color-text-secondary)] cursor-pointer"
                >
                  Don't show again
                </label>
              </div>
              <div className="flex items-center gap-2">
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={handleCancelDismiss}
                >
                  Cancel
                </Button>
                <Button
                  size="sm"
                  variant="default"
                  onClick={handleConfirmDismiss}
                >
                  <Check className="h-3 w-3 mr-1" />
                  Dismiss
                </Button>
              </div>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Main Content */}
        <div className="flex gap-3">
          {/* Icon */}
          <motion.div
            className={cn(
              "flex-shrink-0 w-10 h-10 rounded-lg flex items-center justify-center",
              config.colors.iconBg,
              config.colors.iconColor
            )}
            initial={{ scale: 0.8, rotate: -10 }}
            animate={{ scale: 1, rotate: 0 }}
            transition={{ delay: 0.1, type: "spring", stiffness: 300 }}
          >
            {getInsightIcon(type)}
          </motion.div>

          {/* Content */}
          <div className="flex-1 min-w-0">
            <div className="flex items-start justify-between gap-2">
              <h4 className={cn("text-sm font-semibold", config.colors.title)}>
                {title}
              </h4>
              {dismissible && (
                <button
                  onClick={handleDismiss}
                  className={cn(
                    "flex-shrink-0 p-1 rounded transition-colors",
                    "text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)]",
                    "hover:bg-[var(--color-bg-tertiary)]"
                  )}
                  aria-label="Dismiss insight"
                >
                  <X className="h-4 w-4" />
                </button>
              )}
            </div>
            <p className={cn("text-xs mt-1 leading-relaxed", config.colors.description)}>
              {description}
            </p>

            {/* Actions */}
            {actions && actions.length > 0 && (
              <motion.div
                className="flex flex-wrap items-center gap-2 mt-3"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ delay: 0.15 }}
              >
                {actions.map((action, index) => (
                  <motion.div
                    key={action.id}
                    initial={{ opacity: 0, x: -10 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: 0.2 + index * 0.05 }}
                  >
                    <Button
                      size="sm"
                      variant={action.variant === "primary" ? "default" : action.variant === "ghost" ? "ghost" : "outline"}
                      onClick={action.onClick}
                      className={cn(
                        "h-7 text-xs",
                        action.variant === "secondary" && "border-current"
                      )}
                    >
                      {action.icon && <span className="mr-1.5">{action.icon}</span>}
                      {action.label}
                      {action.variant === "primary" && (
                        <ArrowRight className="h-3 w-3 ml-1.5" />
                      )}
                    </Button>
                  </motion.div>
                ))}
              </motion.div>
            )}
          </div>
        </div>
      </motion.div>
    </AnimatePresence>
  );
};

/**
 * Pre-configured insight presets for common use cases
 */
export const InsightPresets = {
  performance: (overrides?: Partial<Insight>): Insight => ({
    id: "performance-insight",
    title: "Performance Alert",
    description: "Latency above top 10% threshold. Consider optimizing your function.",
    type: "warning",
    dismissible: true,
    actions: [
      {
        id: "view-metrics",
        label: "View Metrics",
        variant: "secondary",
        onClick: () => {},
      },
      {
        id: "optimize",
        label: "Optimize Now",
        variant: "primary",
        onClick: () => {},
      },
    ],
    ...overrides,
  }),

  monetization: (overrides?: Partial<Insight>): Insight => ({
    id: "monetization-insight",
    title: "Enable Monetization",
    description: "Enable pricing to start earning from your function usage.",
    type: "info",
    dismissible: true,
    actions: [
      {
        id: "learn-more",
        label: "Learn More",
        variant: "secondary",
        onClick: () => {},
      },
      {
        id: "enable",
        label: "Enable Pricing",
        variant: "primary",
        onClick: () => {},
      },
    ],
    ...overrides,
  }),

  quality: (overrides?: Partial<Insight>): Insight => ({
    id: "quality-insight",
    title: "Improve Discoverability",
    description: "Add more metadata to improve your function's discoverability in the marketplace.",
    type: "info",
    dismissible: true,
    actions: [
      {
        id: "edit-metadata",
        label: "Edit Metadata",
        variant: "primary",
        onClick: () => {},
      },
    ],
    ...overrides,
  }),

  security: (overrides?: Partial<Insight>): Insight => ({
    id: "security-insight",
    title: "Security Update Available",
    description: "Update dependencies to address known vulnerabilities.",
    type: "error",
    dismissible: false,
    actions: [
      {
        id: "view-details",
        label: "View Details",
        variant: "secondary",
        onClick: () => {},
      },
      {
        id: "update",
        label: "Update Now",
        variant: "primary",
        onClick: () => {},
      },
    ],
    ...overrides,
  }),

  usage: (overrides?: Partial<Insight>): Insight => ({
    id: "usage-insight",
    title: "High Traffic Detected",
    description: "Your function is receiving 3x normal traffic. Consider scaling up.",
    type: "success",
    dismissible: true,
    actions: [
      {
        id: "view-usage",
        label: "View Usage",
        variant: "secondary",
        onClick: () => {},
      },
      {
        id: "scale",
        label: "Scale Up",
        variant: "primary",
        onClick: () => {},
      },
    ],
    ...overrides,
  }),
};

export default FlyInsightCard;
