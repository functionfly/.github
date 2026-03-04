/**
 * FlyExecutionPreview.tsx
 *
 * Safety preview component shown before executing assistant-suggested actions.
 * Provides clear visibility into what will happen, estimated costs, risks,
 * and a safety countdown before confirmation.
 */

import React, { useCallback, useEffect, useMemo, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import {
  AlertTriangle,
  CheckCircle,
  XCircle,
  Clock,
  DollarSign,
  Zap,
  Shield,
  TrendingUp,
  TrendingDown,
  Minus,
  FunctionSquare,
  ChevronRight,
  Loader2,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";

// ============================================================================
// Types & Interfaces
// ============================================================================

/**
 * Risk level for execution preview
 */
export type RiskLevel = "low" | "medium" | "high";

/**
 * Affected function information
 */
export interface AffectedFunction {
  /** Function ID */
  id: string;
  /** Function name */
  name: string;
  /** Current status */
  status: "active" | "inactive" | "error";
  /** Impact description */
  impact: string;
  /** Estimated invocations affected */
  invocationsAffected?: number;
}

/**
 * Props for the FlyExecutionPreview component
 */
export interface FlyExecutionPreviewProps {
  /** Preview title */
  title: string;
  /** Detailed description of what will happen */
  description: string;
  /** List of affected functions */
  affectedFunctions: AffectedFunction[];
  /** Estimated cost impact (in cents) */
  estimatedCost: number;
  /** Estimated latency in milliseconds */
  estimatedLatency: number;
  /** Risk assessment level */
  riskLevel: RiskLevel;
  /** Trust score impact (+/- or 0) */
  trustImpact?: number;
  /** Callback when user confirms */
  onConfirm: () => void;
  /** Callback when user cancels */
  onCancel: () => void;
  /** Countdown seconds before confirm is enabled */
  countdownSeconds?: number;
  /** Custom className */
  className?: string;
  /** Whether execution is in progress */
  isExecuting?: boolean;
}

// ============================================================================
// Risk Level Configuration
// ============================================================================

interface RiskConfig {
  color: string;
  bgColor: string;
  borderColor: string;
  icon: React.ReactNode;
  label: string;
}

const RISK_CONFIG: Record<RiskLevel, RiskConfig> = {
  low: {
    color: "text-emerald-500",
    bgColor: "bg-emerald-500/10",
    borderColor: "border-emerald-500/30",
    icon: <Shield className="h-5 w-5" />,
    label: "Low Risk",
  },
  medium: {
    color: "text-amber-500",
    bgColor: "bg-amber-500/10",
    borderColor: "border-amber-500/30",
    icon: <AlertTriangle className="h-5 w-5" />,
    label: "Medium Risk",
  },
  high: {
    color: "text-red-500",
    bgColor: "bg-red-500/10",
    borderColor: "border-red-500/30",
    icon: <AlertTriangle className="h-5 w-5" />,
    label: "High Risk",
  },
};

// ============================================================================
// Helper Functions
// ============================================================================

function formatCost(costInCents: number): string {
  if (costInCents < 100) {
    return `${costInCents}¢`;
  }
  return `$${(costInCents / 100).toFixed(2)}`;
}

function formatLatency(ms: number): string {
  if (ms < 1000) {
    return `${ms}ms`;
  }
  return `${(ms / 1000).toFixed(1)}s`;
}

function formatTrustImpact(impact: number): string {
  if (impact > 0) return `+${impact}`;
  if (impact < 0) return `${impact}`;
  return "No change";
}

// ============================================================================
// Countdown Hook
// ============================================================================

function useCountdown(
  initialSeconds: number,
  onComplete?: () => void
): [number, boolean] {
  const [seconds, setSeconds] = useState(initialSeconds);
  const [isComplete, setIsComplete] = useState(initialSeconds <= 0);

  useEffect(() => {
    if (initialSeconds <= 0) {
      setIsComplete(true);
      return;
    }

    setSeconds(initialSeconds);
    setIsComplete(false);

    const interval = setInterval(() => {
      setSeconds((prev) => {
        if (prev <= 1) {
          clearInterval(interval);
          setIsComplete(true);
          onComplete?.();
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(interval);
  }, [initialSeconds, onComplete]);

  return [seconds, isComplete];
}

// ============================================================================
// Sub-Components
// ============================================================================

interface FunctionGraphItemProps {
  func: AffectedFunction;
  index: number;
}

const FunctionGraphItem = React.memo<FunctionGraphItemProps>(
  ({ func, index }) => {
    const statusColors = {
      active: "bg-emerald-500",
      inactive: "bg-gray-400",
      error: "bg-red-500",
    };

    return (
      <motion.div
        initial={{ opacity: 0, x: -20 }}
        animate={{ opacity: 1, x: 0 }}
        transition={{ delay: index * 0.1 }}
        className="flex items-center gap-3 p-2 rounded-lg bg-[var(--color-bg-primary)]/50 border border-[var(--color-border)]/50"
      >
        <div className="relative">
          <FunctionSquare className="h-8 w-8 text-[var(--color-brand-500)]" />
          <span
            className={cn(
              "absolute -top-1 -right-1 h-2.5 w-2.5 rounded-full border-2 border-[var(--color-bg-primary)]",
              statusColors[func.status]
            )}
          />
        </div>
        <div className="flex-1 min-w-0">
          <p className="font-medium text-sm text-[var(--color-text-primary)] truncate">
            {func.name}
          </p>
          <p className="text-xs text-[var(--color-text-secondary)] truncate">
            {func.impact}
          </p>
        </div>
        {func.invocationsAffected !== undefined && (
          <Badge variant="secondary" className="text-xs shrink-0">
            {func.invocationsAffected.toLocaleString()} invocations
          </Badge>
        )}
      </motion.div>
    );
  }
);

FunctionGraphItem.displayName = "FunctionGraphItem";

// ============================================================================
// Main Component
// ============================================================================

export const FlyExecutionPreview = React.memo<FlyExecutionPreviewProps>(
  ({
    title,
    description,
    affectedFunctions,
    estimatedCost,
    estimatedLatency,
    riskLevel,
    trustImpact = 0,
    onConfirm,
    onCancel,
    countdownSeconds = 3,
    className,
    isExecuting = false,
  }) => {
    const [countdown, canConfirm] = useCountdown(countdownSeconds);
    const riskConfig = RISK_CONFIG[riskLevel];

    const handleConfirm = useCallback(() => {
      if (canConfirm && !isExecuting) {
        onConfirm();
      }
    }, [canConfirm, isExecuting, onConfirm]);

    // Trust impact visualization
    const trustImpactConfig = useMemo(() => {
      if (trustImpact > 0) {
        return {
          icon: <TrendingUp className="h-4 w-4" />,
          color: "text-emerald-500",
          bgColor: "bg-emerald-500/10",
          label: "Positive",
        };
      }
      if (trustImpact < 0) {
        return {
          icon: <TrendingDown className="h-4 w-4" />,
          color: "text-red-500",
          bgColor: "bg-red-500/10",
          label: "Negative",
        };
      }
      return {
        icon: <Minus className="h-4 w-4" />,
        color: "text-gray-500",
        bgColor: "bg-gray-500/10",
        label: "Neutral",
      };
    }, [trustImpact]);

    return (
      <motion.div
        initial={{ opacity: 0, y: 20, scale: 0.95 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        exit={{ opacity: 0, y: -20, scale: 0.95 }}
        transition={{ duration: 0.3, ease: "easeOut" }}
        className={cn(
          "w-full max-w-lg rounded-xl overflow-hidden",
          "bg-[var(--color-bg-secondary)]",
          "border border-[var(--color-border)]",
          "shadow-lg shadow-black/10",
          className
        )}
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="execution-preview-title"
        aria-describedby="execution-preview-description"
      >
        {/* Warning Header */}
        <div
          className={cn(
            "px-4 py-3",
            "bg-amber-500/10",
            "border-b border-amber-500/30",
            "flex items-center gap-3"
          )}
        >
          <AlertTriangle className="h-5 w-5 text-amber-400 shrink-0" />
          <div className="flex-1">
            <h3
              id="execution-preview-title"
              className="font-semibold text-amber-400 text-sm"
            >
              Review Before Executing
            </h3>
            <p className="text-xs text-amber-400/80">
              This action will affect your functions
            </p>
          </div>
        </div>

        {/* Content */}
        <div className="p-4 space-y-4">
          {/* Description */}
          <p
            id="execution-preview-description"
            className="text-sm text-[var(--color-text-primary)] leading-relaxed"
          >
            {description}
          </p>

          {/* Function Graph Summary */}
          {affectedFunctions.length > 0 && (
            <div className="space-y-2">
              <h4 className="text-xs font-medium text-[var(--color-text-secondary)] uppercase tracking-wide">
                Affected Functions ({affectedFunctions.length})
              </h4>
              <div className="space-y-2 max-h-40 overflow-y-auto scrollbar-thin">
                <AnimatePresence>
                  {affectedFunctions.map((func, index) => (
                    <FunctionGraphItem
                      key={func.id}
                      func={func}
                      index={index}
                    />
                  ))}
                </AnimatePresence>
              </div>
            </div>
          )}

          {/* Estimates Grid */}
          <div className="grid grid-cols-3 gap-3">
            {/* Cost */}
            <div className="p-3 rounded-lg bg-[var(--color-bg-primary)] border border-[var(--color-border)]">
              <div className="flex items-center gap-1.5 text-[var(--color-text-secondary)] mb-1">
                <DollarSign className="h-3.5 w-3.5" />
                <span className="text-xs">Est. Cost</span>
              </div>
              <p className="font-semibold text-sm text-[var(--color-text-primary)]">
                {formatCost(estimatedCost)}
              </p>
            </div>

            {/* Latency */}
            <div className="p-3 rounded-lg bg-[var(--color-bg-primary)] border border-[var(--color-border)]">
              <div className="flex items-center gap-1.5 text-[var(--color-text-secondary)] mb-1">
                <Zap className="h-3.5 w-3.5" />
                <span className="text-xs">Est. Time</span>
              </div>
              <p className="font-semibold text-sm text-[var(--color-text-primary)]">
                {formatLatency(estimatedLatency)}
              </p>
            </div>

            {/* Risk Level */}
            <div
              className={cn(
                "p-3 rounded-lg border",
                riskConfig.bgColor,
                riskConfig.borderColor
              )}
            >
              <div
                className={cn(
                  "flex items-center gap-1.5 mb-1",
                  riskConfig.color
                )}
              >
                {riskConfig.icon}
                <span className="text-xs">Risk</span>
              </div>
              <p className={cn("font-semibold text-sm", riskConfig.color)}>
                {riskConfig.label}
              </p>
            </div>
          </div>

          {/* Trust Impact */}
          {trustImpact !== 0 && (
            <div
              className={cn(
                "flex items-center justify-between p-2 rounded-lg",
                trustImpactConfig.bgColor
              )}
            >
              <div className="flex items-center gap-2">
                <span className={trustImpactConfig.color}>
                  {trustImpactConfig.icon}
                </span>
                <span className="text-sm text-[var(--color-text-secondary)]">
                  Trust Score Impact
                </span>
              </div>
              <Badge
                variant="outline"
                className={cn("text-xs", trustImpactConfig.color)}
              >
                {formatTrustImpact(trustImpact)}
              </Badge>
            </div>
          )}
        </div>

        {/* Action Buttons */}
        <div className="px-4 py-3 border-t border-[var(--color-border)] flex items-center justify-end gap-3">
          <Button
            variant="ghost"
            size="sm"
            onClick={onCancel}
            disabled={isExecuting}
            className="text-[var(--color-text-secondary)]"
          >
            Cancel
          </Button>

          <Button
            variant="default"
            size="sm"
            onClick={handleConfirm}
            disabled={!canConfirm || isExecuting}
            isLoading={isExecuting}
            className={cn(
              "min-w-[120px] transition-all duration-300",
              !canConfirm && "opacity-80"
            )}
          >
            {isExecuting ? (
              "Executing..."
            ) : !canConfirm ? (
              <span className="flex items-center gap-2">
                <Clock className="h-4 w-4" />
                Confirm ({countdown}s)
              </span>
            ) : (
              <span className="flex items-center gap-2">
                <CheckCircle className="h-4 w-4" />
                Confirm
              </span>
            )}
          </Button>
        </div>
      </motion.div>
    );
  }
);

FlyExecutionPreview.displayName = "FlyExecutionPreview";

export default FlyExecutionPreview;
