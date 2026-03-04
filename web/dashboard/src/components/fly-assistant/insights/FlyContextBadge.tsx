/**
 * FlyContextBadge.tsx
 *
 * Displays current context the assistant is aware of including
 * current page, function, marketplace category, and trust tier.
 * Provides subtle animations when context changes.
 */

import React, { useCallback } from "react";
import { motion } from "framer-motion";
import {
  LayoutDashboard,
  FunctionSquare,
  Tag,
  Shield,
  ChevronRight,
  Settings,
  Store,
  FileText,
  BarChart3,
} from "lucide-react";
import { cn } from "@/lib/utils";
import type { TrustTier } from "../FlyAssistantProvider";

// ============================================================================
// Types & Interfaces
// ============================================================================

export interface ContextInfo {
  page: string;
  function?: string;
  category?: string;
  trustTier?: TrustTier;
}

export interface FlyContextBadgeProps {
  context: ContextInfo;
  onClick?: () => void;
  className?: string;
  showTrustTier?: boolean;
}

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Get icon based on page name
 */
function getPageIcon(page: string): React.ReactNode {
  const normalized = page.toLowerCase();

  if (normalized.includes("dashboard")) return <LayoutDashboard className="h-3 w-3" />;
  if (normalized.includes("function")) return <FunctionSquare className="h-3 w-3" />;
  if (normalized.includes("marketplace")) return <Store className="h-3 w-3" />;
  if (normalized.includes("settings")) return <Settings className="h-3 w-3" />;
  if (normalized.includes("analytics") || normalized.includes("metrics")) {
    return <BarChart3 className="h-3 w-3" />;
  }
  if (normalized.includes("docs") || normalized.includes("documentation")) {
    return <FileText className="h-3 w-3" />;
  }

  return <LayoutDashboard className="h-3 w-3" />;
}

/**
 * Get color for trust tier
 */
function getTrustTierColor(tier: TrustTier): string {
  const colors: Record<TrustTier, string> = {
    low: "text-red-500 bg-red-500/10 border-red-500/30",
    medium: "text-amber-500 bg-amber-500/10 border-amber-500/30",
    high: "text-indigo-500 bg-indigo-500/10 border-indigo-500/30",
    critical: "text-emerald-500 bg-emerald-500/10 border-emerald-500/30",
  };
  return colors[tier];
}

/**
 * Get trust tier icon
 */
function getTrustTierIcon(tier: TrustTier): React.ReactNode {
  return <Shield className="h-3 w-3" />;
}

// ============================================================================
// Component
// ============================================================================

export const FlyContextBadge: React.FC<FlyContextBadgeProps> = ({
  context,
  onClick,
  className,
  showTrustTier = true,
}) => {
  const { page, function: funcName, category, trustTier } = context;

  const handleClick = useCallback(() => {
    onClick?.();
  }, [onClick]);

  const hasContext = Boolean(page || funcName || category);

  if (!hasContext) {
    return null;
  }

  return (
    <motion.div
      className={cn(
        "inline-flex flex-wrap items-center gap-2",
        "bg-[var(--color-bg-tertiary)]",
        "border border-[var(--color-border)]",
        "rounded-md px-2 py-1",
        onClick && "cursor-pointer hover:border-[var(--color-brand-500)]/30 transition-colors",
        className
      )}
      onClick={handleClick}
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ duration: 0.2 }}
      whileHover={onClick ? { scale: 1.01 } : undefined}
      whileTap={onClick ? { scale: 0.98 } : undefined}
      role="button"
      tabIndex={onClick ? 0 : -1}
      aria-label="Current context"
      onKeyDown={(e) => {
        if (onClick && (e.key === "Enter" || e.key === " ")) {
          handleClick();
        }
      }}
    >
      {/* Page Context */}
      {page && (
        <motion.div
          key={`page-${page}`}
          className="flex items-center gap-1 text-xs text-[var(--color-text-secondary)]"
          initial={{ opacity: 0, x: -5 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.2, delay: 0.05 }}
        >
          <span className="text-[var(--color-text-tertiary)]">{getPageIcon(page)}</span>
          <span className="font-medium">Page</span>
          <ChevronRight className="h-3 w-3 text-[var(--color-text-muted)]" />
          <span className="text-[var(--color-text-primary)]">{page}</span>
        </motion.div>
      )}

      {/* Function Context */}
      {funcName && (
        <>
          {page && <span className="text-[var(--color-border)]">|</span>}
          <motion.div
            key={`func-${funcName}`}
            className="flex items-center gap-1 text-xs text-[var(--color-text-secondary)]"
            initial={{ opacity: 0, x: -5 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.2, delay: 0.1 }}
          >
            <FunctionSquare className="h-3 w-3 text-[var(--color-text-tertiary)]" />
            <span className="font-medium">Function</span>
            <ChevronRight className="h-3 w-3 text-[var(--color-text-muted)]" />
            <span className="text-[var(--color-text-primary)] truncate max-w-[120px]">
              {funcName}
            </span>
          </motion.div>
        </>
      )}

      {/* Category Context */}
      {category && (
        <>
          {(page || funcName) && <span className="text-[var(--color-border)]">|</span>}
          <motion.div
            key={`cat-${category}`}
            className="flex items-center gap-1 text-xs text-[var(--color-text-secondary)]"
            initial={{ opacity: 0, x: -5 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.2, delay: 0.15 }}
          >
            <Tag className="h-3 w-3 text-[var(--color-text-tertiary)]" />
            <span className="font-medium">Category</span>
            <ChevronRight className="h-3 w-3 text-[var(--color-text-muted)]" />
            <span className="text-[var(--color-text-primary)]">{category}</span>
          </motion.div>
        </>
      )}

      {/* Trust Tier */}
      {showTrustTier && trustTier && (
        <>
          {(page || funcName || category) && <span className="text-[var(--color-border)]">|</span>}
          <motion.div
            key={`tier-${trustTier}`}
            className={cn(
              "flex items-center gap-1 text-xs px-1.5 py-0.5 rounded",
              "border",
              getTrustTierColor(trustTier)
            )}
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.2, delay: 0.2 }}
            whileHover={{ scale: 1.05 }}
          >
            {getTrustTierIcon(trustTier)}
            <span className="capitalize">{trustTier}</span>
          </motion.div>
        </>
      )}
    </motion.div>
  );
};

export default FlyContextBadge;
