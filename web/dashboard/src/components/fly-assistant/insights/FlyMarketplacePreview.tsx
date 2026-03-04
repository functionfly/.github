/**
 * FlyMarketplacePreview.tsx
 *
 * Marketplace function preview card for when the assistant suggests
 * functions from the marketplace. Includes trust score badge, latency
 * indicator, and quick actions.
 */

import React, { useCallback } from "react";
import { motion } from "framer-motion";
import {
  FunctionSquare,
  Zap,
  Gauge,
  Turtle,
  ExternalLink,
  GitCompare,
  Download,
  DollarSign,
  Star,
  DownloadIcon,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { FlyTrustScoreWidget } from "./FlyTrustScoreWidget";
import type { TrustTier } from "../FlyAssistantProvider";

// ============================================================================
// Types & Interfaces
// ============================================================================

export type LatencyIndicator = "fast" | "medium" | "slow";

export interface MarketplaceFunction {
  id: string;
  name: string;
  description?: string;
  trustScore: number;
  latency: LatencyIndicator;
  price?: number;
  currency?: string;
  icon?: string;
  author?: string;
  downloads?: number;
  rating?: number;
}

export interface FlyMarketplacePreviewProps {
  func: MarketplaceFunction;
  onCompare?: (func: MarketplaceFunction) => void;
  onView?: (func: MarketplaceFunction) => void;
  onUse?: (func: MarketplaceFunction) => void;
  className?: string;
  variant?: "default" | "compact";
}

// ============================================================================
// Helper Functions
// ============================================================================

function getLatencyIcon(latency: LatencyIndicator): React.ReactNode {
  const icons: Record<LatencyIndicator, React.ReactNode> = {
    fast: <Zap className="h-3.5 w-3.5" />,
    medium: <Gauge className="h-3.5 w-3.5" />,
    slow: <Turtle className="h-3.5 w-3.5" />,
  };
  return icons[latency];
}

function getLatencyColor(latency: LatencyIndicator): string {
  const colors: Record<LatencyIndicator, string> = {
    fast: "text-emerald-500 bg-emerald-500/10 border-emerald-500/30",
    medium: "text-amber-500 bg-amber-500/10 border-amber-500/30",
    slow: "text-red-500 bg-red-500/10 border-red-500/30",
  };
  return colors[latency];
}

function getLatencyLabel(latency: LatencyIndicator): string {
  const labels: Record<LatencyIndicator, string> = {
    fast: "Fast",
    medium: "Medium",
    slow: "Slow",
  };
  return labels[latency];
}

function getTierFromScore(score: number): TrustTier {
  if (score >= 90) return "critical";
  if (score >= 70) return "high";
  if (score >= 40) return "medium";
  return "low";
}

function formatPrice(price: number, currency: string = "USD"): string {
  if (price === 0) return "Free";
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency,
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(price);
}

function formatDownloads(count: number): string {
  if (count >= 1000000) return `${(count / 1000000).toFixed(1)}M`;
  if (count >= 1000) return `${(count / 1000).toFixed(1)}K`;
  return count.toString();
}

// ============================================================================
// Component
// ============================================================================

export const FlyMarketplacePreview: React.FC<FlyMarketplacePreviewProps> = ({
  func,
  onCompare,
  onView,
  onUse,
  className,
  variant = "default",
}) => {
  const {
    name,
    description,
    trustScore,
    latency,
    price,
    currency,
    author,
    downloads,
    rating,
  } = func;

  const tier = getTierFromScore(trustScore);
  const isFree = price === undefined || price === 0;

  const handleCompare = useCallback(() => {
    onCompare?.(func);
  }, [func, onCompare]);

  const handleView = useCallback(() => {
    onView?.(func);
  }, [func, onView]);

  const handleUse = useCallback(() => {
    onUse?.(func);
  }, [func, onUse]);

  return (
    <TooltipProvider>
      <motion.div
        className={cn(
          "relative overflow-hidden",
          "bg-[var(--color-bg-secondary)]",
          "border border-[var(--color-border)]",
          "rounded-xl",
          "transition-all duration-200",
          "hover:border-[var(--color-brand-500)]/50 hover:shadow-lg",
          "hover:shadow-[var(--color-brand-500)]/5",
          variant === "compact" && "p-3",
          variant === "default" && "p-4",
          className
        )}
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
        whileHover={{ y: -2 }}
      >
        {/* Glow Effect on Hover */}
        <div className="absolute inset-0 opacity-0 hover:opacity-100 transition-opacity duration-300 pointer-events-none">
          <div className="absolute inset-0 bg-gradient-to-br from-[var(--color-brand-500)]/5 to-transparent" />
        </div>

        {/* Header */}
        <div className="relative flex items-start justify-between gap-3">
          <div className="flex items-center gap-3 min-w-0">
            {/* Function Icon */}
            <motion.div
              className={cn(
                "flex-shrink-0 flex items-center justify-center rounded-lg",
                "bg-[var(--color-bg-tertiary)] border border-[var(--color-border)]",
                variant === "compact" ? "w-10 h-10" : "w-12 h-12"
              )}
              whileHover={{ scale: 1.05, rotate: 5 }}
              transition={{ type: "spring", stiffness: 300 }}
            >
              <FunctionSquare
                className={cn(
                  "text-[var(--color-brand-500)]",
                  variant === "compact" ? "h-5 w-5" : "h-6 w-6"
                )}
              />
            </motion.div>

            {/* Name & Meta */}
            <div className="min-w-0 flex-1">
              <h4
                className={cn(
                  "font-semibold text-[var(--color-text-primary)] truncate",
                  variant === "compact" ? "text-sm" : "text-base"
                )}
              >
                {name}
              </h4>
              {author && (
                <p className="text-xs text-[var(--color-text-secondary)] truncate">
                  by {author}
                </p>
              )}
            </div>
          </div>

          {/* Trust Score Badge */}
          <FlyTrustScoreWidget
            score={trustScore}
            tier={tier}
            variant={variant === "compact" ? "inline" : "compact"}
            showDelta={false}
          />
        </div>

        {/* Description */}
        {description && variant === "default" && (
          <motion.p
            className="relative mt-3 text-sm text-[var(--color-text-secondary)] line-clamp-2"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.1 }}
          >
            {description}
          </motion.p>
        )}

        {/* Meta Row */}
        <div className="relative flex flex-wrap items-center gap-2 mt-3">
          {/* Latency Badge */}
          <Tooltip>
            <TooltipTrigger asChild>
              <Badge
                variant="outline"
                className={cn(
                  "gap-1.5 font-medium",
                  getLatencyColor(latency)
                )}
              >
                {getLatencyIcon(latency)}
                <span className="text-[10px] uppercase">{getLatencyLabel(latency)}</span>
              </Badge>
            </TooltipTrigger>
            <TooltipContent>
              <p className="text-xs">
                {latency === "fast" && "< 50ms response time"}
                {latency === "medium" && "50-200ms response time"}
                {latency === "slow" && "> 200ms response time"}
              </p>
            </TooltipContent>
          </Tooltip>

          {/* Price Badge */}
          <Badge
            variant={isFree ? "default" : "secondary"}
            className={cn(
              "gap-1 text-[10px] font-medium",
              isFree && "bg-emerald-500/20 text-emerald-600 hover:bg-emerald-500/30"
            )}
          >
            <DollarSign className="h-3 w-3" />
            {isFree ? "Free" : formatPrice(price!, currency)}
          </Badge>

          {/* Downloads */}
          {downloads !== undefined && (
            <Badge
              variant="outline"
              className="gap-1 text-[10px] font-medium text-[var(--color-text-secondary)]"
            >
              <DownloadIcon className="h-3 w-3" />
              {formatDownloads(downloads)}
            </Badge>
          )}

          {/* Rating */}
          {rating !== undefined && (
            <Badge
              variant="outline"
              className="gap-1 text-[10px] font-medium text-amber-500 border-amber-500/30"
            >
              <Star className="h-3 w-3 fill-current" />
              {rating.toFixed(1)}
            </Badge>
          )}
        </div>

        {/* Actions */}
        <motion.div
          className={cn(
            "relative flex items-center gap-2 mt-4",
            variant === "compact" && "mt-3"
          )}
          initial={{ opacity: 0, y: 5 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.15 }}
        >
          {/* Compare Button */}
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                size="sm"
                variant="outline"
                className="flex-1 h-8 text-xs"
                onClick={handleCompare}
              >
                <GitCompare className="h-3.5 w-3.5 mr-1.5" />
                Compare
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              <p className="text-xs">Compare with current function</p>
            </TooltipContent>
          </Tooltip>

          {/* View Button */}
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                size="sm"
                variant="outline"
                className="flex-1 h-8 text-xs"
                onClick={handleView}
              >
                <ExternalLink className="h-3.5 w-3.5 mr-1.5" />
                View
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              <p className="text-xs">View function details</p>
            </TooltipContent>
          </Tooltip>

          {/* Use Button */}
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                size="sm"
                className="flex-1 h-8 text-xs bg-[var(--color-brand-500)] hover:bg-[var(--color-brand-600)]"
                onClick={handleUse}
              >
                <Download className="h-3.5 w-3.5 mr-1.5" />
                Use
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              <p className="text-xs">Quick import this function</p>
            </TooltipContent>
          </Tooltip>
        </motion.div>
      </motion.div>
    </TooltipProvider>
  );
};

export default FlyMarketplacePreview;
