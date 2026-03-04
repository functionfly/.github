/**
 * FlyTrustScoreWidget.tsx
 *
 * Mini gamified trust score display with animated progress ring,
 * current tier badge, and delta to next tier visualization.
 */

import React, { useEffect, useState } from "react";
import { motion, useSpring, useTransform } from "framer-motion";
import { Shield, TrendingUp, Award, Lock } from "lucide-react";
import { cn } from "@/lib/utils";
import type { TrustTier } from "../FlyAssistantProvider";

// ============================================================================
// Types & Interfaces
// ============================================================================

export interface FlyTrustScoreWidgetProps {
  score: number;
  tier: TrustTier;
  nextTierThreshold?: number;
  variant?: "default" | "compact" | "inline";
  showDelta?: boolean;
  className?: string;
  onClick?: () => void;
}

// ============================================================================
// Constants & Configuration
// ============================================================================

const TIER_THRESHOLDS: Record<TrustTier, { min: number; max: number; next: TrustTier | null }> = {
  low: { min: 0, max: 40, next: "medium" },
  medium: { min: 40, max: 70, next: "high" },
  high: { min: 70, max: 90, next: "critical" },
  critical: { min: 90, max: 100, next: null },
};

const TIER_CONFIG: Record<TrustTier, {
  color: string;
  bgColor: string;
  label: string;
  icon: React.ReactNode;
}> = {
  low: {
    color: "#ef4444", // red-500
    bgColor: "bg-red-500",
    label: "Low",
    icon: <Shield className="h-4 w-4" />,
  },
  medium: {
    color: "#f59e0b", // amber-500
    bgColor: "bg-amber-500",
    label: "Medium",
    icon: <Shield className="h-4 w-4" />,
  },
  high: {
    color: "#6366f1", // indigo-500
    bgColor: "bg-indigo-500",
    label: "High",
    icon: <Shield className="h-4 w-4" />,
  },
  critical: {
    color: "#10b981", // emerald-500
    bgColor: "bg-emerald-500",
    label: "Critical",
    icon: <Award className="h-4 w-4" />,
  },
};

// ============================================================================
// Helper Functions
// ============================================================================

function getTierFromScore(score: number): TrustTier {
  if (score >= 90) return "critical";
  if (score >= 70) return "high";
  if (score >= 40) return "medium";
  return "low";
}

function calculateProgress(score: number): number {
  return Math.min(Math.max(score, 0), 100);
}

function calculateDelta(score: number, tier: TrustTier, nextTierThreshold?: number): number | null {
  if (tier === "critical") return null;
  const threshold = nextTierThreshold ?? TIER_THRESHOLDS[tier].max;
  return Math.max(0, threshold - score);
}

// ============================================================================
// Progress Ring Component
// ============================================================================

interface ProgressRingProps {
  progress: number;
  size: number;
  strokeWidth: number;
  color: string;
  animated?: boolean;
}

const ProgressRing: React.FC<ProgressRingProps> = ({
  progress,
  size,
  strokeWidth,
  color,
  animated = true,
}) => {
  const radius = (size - strokeWidth) / 2;
  const circumference = radius * 2 * Math.PI;
  const progressOffset = circumference - (progress / 100) * circumference;

  const springProgress = useSpring(0, {
    stiffness: 50,
    damping: 20,
    duration: 1.5,
  });

  const strokeDashoffset = useTransform(
    springProgress,
    [0, 100],
    [circumference, 0]
  );

  useEffect(() => {
    if (animated) {
      springProgress.set(progress);
    }
  }, [progress, animated, springProgress]);

  return (
    <svg width={size} height={size} className="transform -rotate-90">
      {/* Background Circle */}
      <circle
        cx={size / 2}
        cy={size / 2}
        r={radius}
        fill="none"
        stroke="currentColor"
        strokeWidth={strokeWidth}
        className="text-[var(--color-bg-tertiary)]"
      />
      {/* Progress Circle */}
      <motion.circle
        cx={size / 2}
        cy={size / 2}
        r={radius}
        fill="none"
        stroke={color}
        strokeWidth={strokeWidth}
        strokeLinecap="round"
        strokeDasharray={circumference}
        style={{ strokeDashoffset: animated ? strokeDashoffset : progressOffset }}
        initial={animated ? { strokeDashoffset: circumference } : false}
        transition={{ duration: 1.5, ease: "easeOut" }}
      />
    </svg>
  );
};

// ============================================================================
// Main Component
// ============================================================================

export const FlyTrustScoreWidget: React.FC<FlyTrustScoreWidgetProps> = ({
  score,
  tier,
  nextTierThreshold,
  variant = "default",
  showDelta = true,
  className,
  onClick,
}) => {
  const [displayScore, setDisplayScore] = useState(0);
  const config = TIER_CONFIG[tier];
  const delta = calculateDelta(score, tier, nextTierThreshold);
  const normalizedProgress = calculateProgress(score);

  // Animate score number on load
  useEffect(() => {
    const duration = 1500;
    const startTime = Date.now();
    const startValue = 0;

    const animate = () => {
      const elapsed = Date.now() - startTime;
      const progress = Math.min(elapsed / duration, 1);
      // Ease out cubic
      const easeProgress = 1 - Math.pow(1 - progress, 3);
      const currentValue = Math.round(startValue + (score - startValue) * easeProgress);

      setDisplayScore(currentValue);

      if (progress < 1) {
        requestAnimationFrame(animate);
      }
    };

    requestAnimationFrame(animate);
  }, [score]);

  // Variant configurations
  const variantConfig = {
    default: { size: 80, strokeWidth: 6, showText: true },
    compact: { size: 48, strokeWidth: 4, showText: false },
    inline: { size: 36, strokeWidth: 3, showText: false },
  };

  const { size, strokeWidth, showText } = variantConfig[variant];

  if (variant === "inline") {
    return (
      <motion.div
        className={cn(
          "inline-flex items-center gap-2 px-2 py-1 rounded-full",
          "bg-[var(--color-bg-tertiary)] border border-[var(--color-border)]",
          onClick && "cursor-pointer hover:border-[var(--color-brand-500)]/30 transition-colors",
          className
        )}
        onClick={onClick}
        whileHover={onClick ? { scale: 1.02 } : undefined}
        whileTap={onClick ? { scale: 0.98 } : undefined}
      >
        <div className={cn("w-2 h-2 rounded-full", config.bgColor)} />
        <span className="text-xs font-medium text-[var(--color-text-primary)]">
          {displayScore}
        </span>
        <span className="text-[10px] text-[var(--color-text-secondary)] capitalize">
          {tier}
        </span>
      </motion.div>
    );
  }

  return (
    <motion.div
      className={cn(
        "flex items-center gap-4",
        variant === "compact" && "gap-3",
        onClick && "cursor-pointer",
        className
      )}
      onClick={onClick}
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      {/* Progress Ring */}
      <motion.div
        className="relative"
        whileHover={onClick ? { scale: 1.05 } : undefined}
        whileTap={onClick ? { scale: 0.95 } : undefined}
      >
        <ProgressRing
          progress={normalizedProgress}
          size={size}
          strokeWidth={strokeWidth}
          color={config.color}
        />
        {/* Center Content */}
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="flex flex-col items-center">
            <span className={cn(
              "font-bold text-[var(--color-text-primary)]",
              variant === "default" ? "text-lg" : "text-sm"
            )}>
              {displayScore}
            </span>
            {variant === "default" && (
              <span className="text-[10px] text-[var(--color-text-muted)]">/100</span>
            )}
          </div>
        </div>

        {/* Pulse Effect on Score Change */}
        <motion.div
          className="absolute inset-0 rounded-full pointer-events-none"
          style={{ border: `2px solid ${config.color}` }}
          initial={{ scale: 1, opacity: 0 }}
          animate={{
            scale: [1, 1.2, 1.4],
            opacity: [0.5, 0.2, 0],
          }}
          transition={{
            duration: 1,
            repeat: Infinity,
            repeatDelay: 2,
          }}
        />
      </motion.div>

      {/* Info Section */}
      {showText && (
        <div className="flex flex-col gap-1">
          {/* Tier Badge */}
          <motion.div
            className={cn(
              "inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full",
              "text-xs font-medium border",
              tier === "low" && "bg-red-500/10 border-red-500/30 text-red-600",
              tier === "medium" && "bg-amber-500/10 border-amber-500/30 text-amber-600",
              tier === "high" && "bg-indigo-500/10 border-indigo-500/30 text-indigo-600",
              tier === "critical" && "bg-emerald-500/10 border-emerald-500/30 text-emerald-600"
            )}
            initial={{ opacity: 0, x: -10 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: 0.2 }}
          >
            {config.icon}
            <span className="capitalize">{tier} Trust</span>
          </motion.div>

          {/* Delta to Next Tier */}
          {showDelta && delta !== null && (
            <motion.div
              className="flex items-center gap-1 text-xs text-[var(--color-text-secondary)]"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: 0.3 }}
            >
              <TrendingUp className="h-3 w-3" />
              <span>
                +{delta} to reach{" "}
                <span className="font-medium text-[var(--color-text-primary)] capitalize">
                  {TIER_THRESHOLDS[tier].next}
                </span>{" "}
                tier
              </span>
            </motion.div>
          )}

          {/* Max Tier Message */}
          {tier === "critical" && (
            <motion.div
              className="flex items-center gap-1 text-xs text-emerald-600"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: 0.3 }}
            >
              <Lock className="h-3 w-3" />
              <span>Maximum tier achieved</span>
            </motion.div>
          )}
        </div>
      )}

      {/* Compact Tier Badge */}
      {variant === "compact" && (
        <motion.div
          className={cn(
            "flex items-center justify-center w-6 h-6 rounded-full",
            config.bgColor,
            "text-white"
          )}
          initial={{ scale: 0 }}
          animate={{ scale: 1 }}
          transition={{ delay: 0.2, type: "spring", stiffness: 300 }}
        >
          {config.icon}
        </motion.div>
      )}
    </motion.div>
  );
};

export default FlyTrustScoreWidget;
