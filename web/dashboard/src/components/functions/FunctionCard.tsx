/**
 * FunctionCard Component
 *
 * A comprehensive card component for displaying function information in the FunctionFly dashboard.
 * Supports three variants: compact, expanded, and analytics.
 *
 * @example
 * // Compact variant for marketplace grid
 * <FunctionCard
 *   data={functionData}
 *   variant="compact"
 *   onView={(id) => router.push(`/functions/${id}`)}
 *   onFavorite={(id, isFav) => toggleFavorite(id, isFav)}
 * />
 *
 * @example
 * // Expanded variant for search results
 * <FunctionCard
 *   data={functionData}
 *   variant="expanded"
 *   onView={handleView}
 *   onExecute={handleExecute}
 *   onShare={handleShare}
 * />
 *
 * @example
 * // Analytics variant for internal dashboard
 * <FunctionCard
 *   data={functionData}
 *   variant="analytics"
 *   onAdminAction={(id, action) => handleAdminAction(id, action)}
 * />
 */

import * as React from "react";
import { useState } from "react";
import { cva, type VariantProps } from "class-variance-authority";
import {
  CheckCircle,
  Code,
  DollarSign,
  Heart,
  Loader2,
  MoreHorizontal,
  Play,
  Share2,
  Shield,
  Star,
  TrendingUp,
  Zap,
  Activity,
  BarChart3,
  Clock,
  Edit,
  Trash2,
  Ban,
  Check,
  X,
  ExternalLink,
} from "lucide-react";
import { toast } from "sonner";
import { cn, formatNumber } from "@/lib/utils";
import type {
  FunctionCardData,
  FunctionCardProps,
  FunctionCardVariant,
} from "@/types";
import { Card, CardContent, CardHeader, CardFooter } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { TrendSparkline } from "@/components/dashboard/TrendSparkline";

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Format price with appropriate currency symbol
 */
function formatPrice(price: number, currency: string = "USD"): string {
  if (price === 0) return "Free";
  const formatter = new Intl.NumberFormat("en-US", {
    style: "currency",
    currency,
    minimumFractionDigits: price < 1 ? 4 : 2,
    maximumFractionDigits: price < 1 ? 4 : 2,
  });
  return formatter.format(price);
}

/**
 * Format price per call for display
 */
function formatPricePerCall(price: number, currency: string = "USD"): string {
  if (price === 0) return "Free";
  return `${formatPrice(price, currency)}/call`;
}

/**
 * Get trust score color based on value
 */
function getTrustScoreColor(score: number): string {
  if (score >= 80) return "text-emerald-400";
  if (score >= 60) return "text-brand-400";
  if (score >= 40) return "text-amber-400";
  return "text-red-400";
}

/**
 * Get trust score background color
 */
function getTrustScoreBgColor(score: number): string {
  if (score >= 80) return "bg-emerald-500/10";
  if (score >= 60) return "bg-brand-500/10";
  if (score >= 40) return "bg-amber-500/10";
  return "bg-red-500/10";
}

/**
 * Get star rating color
 */
function getStarColor(rating: number): string {
  if (rating >= 4.5) return "text-amber-400";
  if (rating >= 4.0) return "text-yellow-400";
  if (rating >= 3.0) return "text-orange-400";
  return "text-gray-400";
}

// ============================================================================
// Sub-components
// ============================================================================

/**
 * Verified Badge Component
 */
function VerifiedBadge({ className }: { className?: string }) {
  return (
    <Badge
      variant="success"
      className={cn("gap-1 px-1.5 py-0.5", className)}
      aria-label="Verified function"
    >
      <CheckCircle className="h-3 w-3" />
      <span className="sr-only">Verified</span>
    </Badge>
  );
}

/**
 * Deterministic Badge Component
 */
function DeterministicBadge({
  isDeterministic,
  className,
}: {
  isDeterministic: boolean;
  className?: string;
}) {
  if (isDeterministic) {
    return (
      <Badge
        variant="secondary"
        className={cn("gap-1 px-1.5 py-0.5 text-xs", className)}
        aria-label="Deterministic function"
      >
        <Zap className="h-3 w-3" />
        <span>Deterministic</span>
      </Badge>
    );
  }
   return (
     <Badge
       variant="outline"
       className={cn("gap-1 px-1.5 py-0.5 text-xs text-text-secondary", className)}
       aria-label="Non-deterministic function"
     >
      <Zap className="h-3 w-3" />
      <span>Non-deterministic</span>
    </Badge>
  );
}

/**
 * Star Rating Component
 */
function StarRating({
  rating,
  count,
  showCount = true,
  size = "sm",
  className,
}: {
  rating: number;
  count: number;
  showCount?: boolean;
  size?: "sm" | "md" | "lg";
  className?: string;
}) {
  const sizeClasses = {
    sm: "h-3.5 w-3.5",
    md: "h-4 w-4",
    lg: "h-5 w-5",
  };

  const textSizes = {
    sm: "text-xs",
    md: "text-sm",
    lg: "text-base",
  };

  return (
    <div
      className={cn("flex items-center gap-1.5", className)}
      aria-label={`Rating: ${rating.toFixed(1)} out of 5 stars, ${count} reviews`}
    >
      <div className="flex items-center">
        {[1, 2, 3, 4, 5].map((star) => (
          <Star
            key={star}
            className={cn(
              sizeClasses[size],
              star <= Math.round(rating)
                ? getStarColor(rating)
                : "text-text-muted fill-text-muted/20"
            )}
            aria-hidden="true"
          />
        ))}
      </div>
      <span className={cn("font-medium", textSizes[size], getStarColor(rating))}>
        {rating.toFixed(1)}
      </span>
      {showCount && (
        <span className={cn(textSizes[size], "text-text-muted")}>
          ({formatNumber(count)})
        </span>
      )}
    </div>
  );
}

/**
 * Trust Score Indicator Component
 */
function TrustScoreIndicator({
  score,
  showLabel = false,
  size = "sm",
  className,
}: {
  score: number;
  showLabel?: boolean;
  size?: "sm" | "md" | "lg";
  className?: string;
}) {
  const sizeClasses = {
    sm: "h-8 w-8 text-xs",
    md: "h-10 w-10 text-sm",
    lg: "h-12 w-12 text-base",
  };

  const circumference = 2 * Math.PI * 18;
  const strokeDashoffset = circumference - (score / 100) * circumference;

  return (
    <div
      className={cn("flex items-center gap-2", className)}
      aria-label={`Trust score: ${score}%`}
    >
      <div className={cn("relative flex items-center justify-center", sizeClasses[size])}>
        <svg className="h-full w-full -rotate-90" viewBox="0 0 40 40">
          {/* Background circle */}
          <circle
            cx="20"
            cy="20"
            r="18"
            fill="none"
            stroke="currentColor"
            strokeWidth="3"
            className="text-border-subtle"
          />
          {/* Progress circle */}
          <circle
            cx="20"
            cy="20"
            r="18"
            fill="none"
            stroke="currentColor"
            strokeWidth="3"
            strokeLinecap="round"
            className={getTrustScoreColor(score)}
            style={{
              strokeDasharray: circumference,
              strokeDashoffset,
              transition: "stroke-dashoffset 0.5s ease",
            }}
          />
        </svg>
        <Shield className={cn("absolute h-3.5 w-3.5", getTrustScoreColor(score))} />
      </div>
      {showLabel && (
        <div className="flex flex-col">
          <span className={cn("font-semibold", getTrustScoreColor(score))}>{score}%</span>
          <span className="text-xs text-text-muted">Trust Score</span>
        </div>
      )}
    </div>
  );
}

/**
 * Author Info Component
 */
function AuthorInfo({
  author,
  className,
}: {
  author: FunctionCardData["author"];
  className?: string;
}) {
  return (
    <a
      href={author.profileUrl || `#/users/${author.id}`}
      className={cn(
        "flex items-center gap-2 group",
        author.profileUrl && "hover:opacity-80 transition-opacity"
      )}
      aria-label={`View profile of ${author.name}`}
    >
      {author.avatar ? (
        <img
          src={author.avatar}
          alt={`${author.name}'s avatar`}
          className="h-6 w-6 rounded-full border border-border-subtle"
          loading="lazy"
        />
      ) : (
        <div className="h-6 w-6 rounded-full bg-gradient-to-br from-brand-500 to-purple-500 flex items-center justify-center text-xs font-medium text-white">
          {author.name.charAt(0).toUpperCase()}
        </div>
      )}
      <span className="text-sm text-text-secondary group-hover:text-text-primary transition-colors">
        {author.username}
      </span>
    </a>
  );
}

/**
 * Execution Count Component
 */
function ExecutionCount({
  count,
  trend,
  showTrend = true,
  className,
}: {
  count: number;
  trend?: number[];
  showTrend?: boolean;
  className?: string;
}) {
  const trendDirection: "up" | "down" | "neutral" =
    trend && trend.length >= 2
      ? trend[trend.length - 1] > trend[trend.length - 2]
        ? "up"
        : trend[trend.length - 1] < trend[trend.length - 2]
        ? "down"
        : "neutral"
      : "neutral";

  return (
    <div className={cn("flex items-center gap-2", className)}>
      <Activity className="h-4 w-4 text-text-muted" aria-hidden="true" />
      <span className="text-sm text-text-secondary">{formatNumber(count)} calls</span>
      {showTrend && trend && trend.length > 0 && (
        <div className="h-6 w-16">
          <TrendSparkline data={trend} trend={trendDirection} variant="line" />
        </div>
      )}
    </div>
  );
}

// ============================================================================
// Variant Styles (CVA)
// ============================================================================

const cardVariants = cva("transition-all duration-200", {
  variants: {
    variant: {
      compact: "hover:shadow-lg hover:shadow-brand-500/5 hover:-translate-y-0.5",
      expanded: "hover:shadow-xl hover:shadow-brand-500/5",
      analytics: "border-border-default",
    },
  },
  defaultVariants: {
    variant: "compact",
  },
});

// ============================================================================
// Main Component
// ============================================================================

/**
 * FunctionCard Component
 *
 * Displays function information with three layout variants optimized for different contexts.
 */
const FunctionCard = React.forwardRef<HTMLDivElement, FunctionCardProps>(
  (
    {
      data,
      variant = "compact",
      className,
      onView,
      onExecute,
      onFavorite,
      onShare,
      onEdit,
      onDelete,
      onAdminAction,
    },
    ref
  ) => {
    const [isHovered, setIsHovered] = React.useState(false);
    const [isFavorited, setIsFavorited] = React.useState(data.isFavorite ?? false);

    const handleFavorite = (e: React.MouseEvent) => {
      e.preventDefault();
      e.stopPropagation();
      const newState = !isFavorited;
      setIsFavorited(newState);
      onFavorite?.(data.id, newState);
    };

    const handleView = () => {
      onView?.(data.id);
    };

    const handleExecute = (e: React.MouseEvent) => {
      e.stopPropagation();
      onExecute?.(data.id);
    };

    const [isSharing, setIsSharing] = useState(false);

    const handleShare = async (e: React.MouseEvent) => {
      e.stopPropagation();
      e.preventDefault();
      setIsSharing(true);

      const shareUrl = `${window.location.origin}/fx/${data.author.username}/${data.name}`;
      const shareData = {
        title: `${data.author.username}/${data.name}`,
        text: data.description || `Check out ${data.name} on FunctionFly`,
        url: shareUrl,
      };

      try {
        if (navigator.share) {
          await navigator.share(shareData);
          toast.success('Shared successfully');
        } else {
          await navigator.clipboard.writeText(shareUrl);
          toast.success('Link copied to clipboard');
        }
        onShare?.(data.id);
      } catch (err) {
        if ((err as Error).name !== 'AbortError') {
          toast.error('Failed to share');
        }
      } finally {
        setIsSharing(false);
      }
    };

    // ============================================================================
    // Compact Variant (Marketplace Grid)
    // ============================================================================
    if (variant === "compact") {
      return (
        <Card
          ref={ref}
          className={cn(
            cardVariants({ variant: "compact" }),
            "relative overflow-hidden cursor-pointer group",
            className
          )}
          onMouseEnter={() => setIsHovered(true)}
          onMouseLeave={() => setIsHovered(false)}
          onClick={handleView}
          role="article"
          aria-label={`${data.name} by ${data.author.name}`}
          tabIndex={0}
          onKeyDown={(e) => {
            if (e.key === "Enter" || e.key === " ") {
              e.preventDefault();
              handleView();
            }
          }}
        >
          {/* Quick Actions Overlay */}
          <div
            className={cn(
              "absolute top-2 right-2 flex gap-1 transition-opacity duration-200 z-10",
              isHovered ? "opacity-100" : "opacity-0"
            )}
          >
            {onFavorite && (
              <Button
                variant="ghost"
                size="icon"
                className={cn(
                  "h-8 w-8 rounded-full bg-bg-glass backdrop-blur-sm",
                  isFavorited && "text-red-500 hover:text-red-600"
                )}
                onClick={handleFavorite}
                aria-label={isFavorited ? "Remove from favorites" : "Add to favorites"}
              >
                <Heart
                  className={cn("h-4 w-4", isFavorited && "fill-current")}
                  aria-hidden="true"
                />
              </Button>
            )}
          </div>

          <CardHeader className="pb-3">
            <div className="flex items-start justify-between gap-2">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <h3 className="font-semibold text-text-primary truncate">{data.name}</h3>
                  {data.isVerified && <VerifiedBadge />}
                </div>
                <AuthorInfo author={data.author} className="mt-1" />
              </div>
            </div>
          </CardHeader>

          <CardContent className="pt-0 pb-4">
            <div className="flex items-center justify-between">
              <StarRating
                rating={data.rating.average}
                count={data.rating.count}
                showCount={false}
                size="sm"
              />
              <div className="flex items-center gap-1.5">
                <DollarSign className="h-3.5 w-3.5 text-brand-400" aria-hidden="true" />
                <span className="text-sm font-medium text-text-primary">
                  {data.pricing.model === "free"
                    ? "Free"
                    : formatPricePerCall(
                        (data.pricing.pricePerCall || 0) / 100,
                        data.pricing.currency
                      )}
                </span>
              </div>
            </div>

            {/* Hover Actions */}
            <div
              className={cn(
                "flex gap-2 mt-3 transition-all duration-200",
                isHovered ? "opacity-100 translate-y-0" : "opacity-0 translate-y-2"
              )}
            >
              {onExecute && (
                <Button
                  size="sm"
                  className="flex-1 h-8 text-xs"
                  onClick={handleExecute}
                >
                  <Play className="h-3.5 w-3.5 mr-1" />
                  Execute
                </Button>
              )}
              {onView && (
                <Button
                  variant="outline"
                  size="sm"
                  className="flex-1 h-8 text-xs"
                  onClick={handleView}
                >
                  <ExternalLink className="h-3.5 w-3.5 mr-1" />
                  Details
                </Button>
              )}
            </div>
          </CardContent>
        </Card>
      );
    }

    // ============================================================================
    // Expanded Variant (Search Results)
    // ============================================================================
    if (variant === "expanded") {
      return (
        <Card
          ref={ref}
          className={cn(
            cardVariants({ variant: "expanded" }),
            "overflow-hidden",
            className
          )}
          role="article"
          aria-label={`${data.name} by ${data.author.name}`}
        >
          <CardHeader className="pb-4">
            <div className="flex items-start justify-between gap-4">
              <div className="flex-1">
                <div className="flex items-center gap-3 mb-2">
                  <div className="h-10 w-10 rounded-lg bg-gradient-to-br from-brand-500 to-purple-500 flex items-center justify-center">
                    <Code className="h-5 w-5 text-white" aria-hidden="true" />
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <h3 className="text-lg font-semibold text-text-primary">{data.name}</h3>
                      {data.isVerified && <VerifiedBadge />}
                       {data.isFeatured && (
                         <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold border border-transparent bg-brand-700 text-white">
                           <Star className="w-3 h-3" />
                           Featured
                         </span>
                       )}
                    </div>
                    <AuthorInfo author={data.author} />
                  </div>
                </div>

                {data.description && (
                  <p className="text-sm text-text-secondary mt-2 line-clamp-2">
                    {data.description}
                  </p>
                )}

                {data.tags && data.tags.length > 0 && (
                  <div className="flex flex-wrap gap-1.5 mt-3">
                    {data.tags.slice(0, 5).map((tag) => (
                      <Badge key={tag} variant="secondary" className="text-xs">
                        {tag}
                      </Badge>
                    ))}
                    {data.tags.length > 5 && (
                      <Badge variant="outline" className="text-xs">
                        +{data.tags.length - 5}
                      </Badge>
                    )}
                  </div>
                )}
              </div>

              <div className="flex flex-col items-end gap-2">
                <TrustScoreIndicator score={data.trustScore} showLabel />
                <DeterministicBadge isDeterministic={data.isDeterministic} />
              </div>
            </div>
          </CardHeader>

          <CardContent className="pt-0 pb-4">
            <div className="grid grid-cols-3 gap-4 mb-4">
              <div className="p-3 rounded-lg bg-bg-secondary">
                <div className="flex items-center gap-2 text-text-muted mb-1">
                  <Star className="h-4 w-4" aria-hidden="true" />
                  <span className="text-xs">Rating</span>
                </div>
                <StarRating
                  rating={data.rating.average}
                  count={data.rating.count}
                  size="md"
                />
              </div>

              <div className="p-3 rounded-lg bg-bg-secondary">
                <div className="flex items-center gap-2 text-text-muted mb-1">
                  <Activity className="h-4 w-4" aria-hidden="true" />
                  <span className="text-xs">Executions</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-lg font-semibold text-text-primary">
                    {formatNumber(data.metrics.executionCount)}
                  </span>
                  {data.metrics.executionTrend && (
                    <div className="h-8 w-16">
                      <TrendSparkline
                        data={data.metrics.executionTrend}
                        trend="neutral"
                        variant="line"
                      />
                    </div>
                  )}
                </div>
              </div>

              <div className="p-3 rounded-lg bg-bg-secondary">
                <div className="flex items-center gap-2 text-text-muted mb-1">
                  <DollarSign className="h-4 w-4" aria-hidden="true" />
                  <span className="text-xs">Price</span>
                </div>
                <span className="text-lg font-semibold text-text-primary">
                  {data.pricing.model === "free"
                    ? "Free"
                    : formatPricePerCall(
                        (data.pricing.pricePerCall || 0) / 100,
                        data.pricing.currency
                      )}
                </span>
              </div>
            </div>
          </CardContent>

          <CardFooter className="pt-0 border-t border-border-subtle">
            <div className="flex items-center justify-between w-full">
              <div className="flex items-center gap-2">
                {data.language && (
                  <Badge variant="outline" className="text-xs">
                    {data.language}
                  </Badge>
                )}
                {data.version && (
                  <Badge variant="secondary" className="text-xs">
                    v{data.version}
                  </Badge>
                )}
                {data.lastUpdated && (
                  <span className="text-xs text-text-muted">
                    Updated {new Date(data.lastUpdated).toLocaleDateString()}
                  </span>
                )}
              </div>

              <div className="flex items-center gap-2 mt-4">
                {onFavorite && (
                  <Button
                    variant="ghost"
                    size="icon"
                    className={cn(
                      "h-9 w-9",
                      isFavorited && "text-red-500 hover:text-red-600"
                    )}
                    onClick={handleFavorite}
                    aria-label={isFavorited ? "Remove from favorites" : "Add to favorites"}
                  >
                    <Heart
                      className={cn("h-4 w-4", isFavorited && "fill-current")}
                      aria-hidden="true"
                    />
                  </Button>
                )}
                {onShare && (
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-9 w-9"
                    onClick={handleShare}
                    disabled={isSharing}
                    aria-label="Share function"
                  >
                    {isSharing ? (
                      <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
                    ) : (
                      <Share2 className="h-4 w-4" aria-hidden="true" />
                    )}
                  </Button>
                )}
                {onView && (
                  <Button variant="outline" onClick={handleView} className="gap-2">
                    <ExternalLink className="h-4 w-4" />
                    View Details
                  </Button>
                )}
                {onExecute && (
                  <Button variant="outline" onClick={handleExecute} className="gap-2">
                    <Play className="h-4 w-4" />
                    Execute
                  </Button>
                )}
              </div>
            </div>
          </CardFooter>
        </Card>
      );
    }

    // ============================================================================
    // Analytics Variant (Internal Dashboard)
    // ============================================================================
    if (variant === "analytics") {
      return (
        <Card
          ref={ref}
          className={cn(cardVariants({ variant: "analytics" }), className)}
          role="article"
          aria-label={`${data.name} analytics dashboard`}
        >
          <CardHeader className="pb-4">
            <div className="flex items-start justify-between gap-4">
              <div className="flex items-center gap-3">
                <div className="h-12 w-12 rounded-xl bg-gradient-to-br from-brand-500 to-purple-500 flex items-center justify-center">
                  <BarChart3 className="h-6 w-6 text-white" aria-hidden="true" />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="text-lg font-semibold text-text-primary">{data.name}</h3>
                    {data.isVerified && <VerifiedBadge />}
                  </div>
                  <div className="flex items-center gap-2 mt-0.5">
                    <AuthorInfo author={data.author} />
                    <span className="text-text-muted">•</span>
                    <span className="text-xs text-text-muted">ID: {data.id.slice(0, 8)}...</span>
                  </div>
                </div>
              </div>

              {onAdminAction && (
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon" className="h-8 w-8" aria-label="Open admin menu">
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem onClick={() => onAdminAction(data.id, "feature")}>
                      <Star className="h-4 w-4 mr-2" />
                      {data.isFeatured ? "Unfeature" : "Feature"}
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={() => onAdminAction(data.id, "verify")}>
                      <CheckCircle className="h-4 w-4 mr-2" />
                      {data.isVerified ? "Revoke Verification" : "Verify"}
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    {onEdit && (
                      <DropdownMenuItem onClick={() => onEdit(data.id)}>
                        <Edit className="h-4 w-4 mr-2" />
                        Edit
                      </DropdownMenuItem>
                    )}
                    {onDelete && (
                      <DropdownMenuItem
                        className="text-red-500 focus:text-red-500"
                        onClick={() => onDelete(data.id)}
                      >
                        <Trash2 className="h-4 w-4 mr-2" />
                        Delete
                      </DropdownMenuItem>
                    )}
                  </DropdownMenuContent>
                </DropdownMenu>
              )}
            </div>
          </CardHeader>

          <CardContent className="pt-0 pb-4">
            {/* Key Metrics Grid */}
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 mb-4">
              {/* Trust Score */}
              <div className="p-3 rounded-lg bg-bg-secondary border border-border-subtle">
                <div className="flex items-center gap-2 text-text-muted mb-2">
                  <Shield className="h-4 w-4" aria-hidden="true" />
                  <span className="text-xs uppercase tracking-wider">Trust Score</span>
                </div>
                <div className="flex items-baseline gap-2">
                  <span className={cn("text-2xl font-bold", getTrustScoreColor(data.trustScore))}>
                    {data.trustScore}%
                  </span>
                  {data.trustScore >= 80 && (
                    <Badge variant="success" className="text-xs">
                      Excellent
                    </Badge>
                  )}
                </div>
              </div>

              {/* Execution Count */}
              <div className="p-3 rounded-lg bg-bg-secondary border border-border-subtle">
                <div className="flex items-center gap-2 text-text-muted mb-2">
                  <TrendingUp className="h-4 w-4" aria-hidden="true" />
                  <span className="text-xs uppercase tracking-wider">Executions</span>
                </div>
                <div className="flex items-baseline gap-2">
                  <span className="text-2xl font-bold text-text-primary">
                    {formatNumber(data.metrics.executionCount)}
                  </span>
                </div>
                {data.metrics.executionTrend && (
                  <div className="h-10 mt-2">
                    <TrendSparkline data={data.metrics.executionTrend} trend="neutral" />
                  </div>
                )}
              </div>

              {/* Rating */}
              <div className="p-3 rounded-lg bg-bg-secondary border border-border-subtle">
                <div className="flex items-center gap-2 text-text-muted mb-2">
                  <Star className="h-4 w-4" aria-hidden="true" />
                  <span className="text-xs uppercase tracking-wider">Rating</span>
                </div>
                <div className="flex items-baseline gap-2">
                  <span className={cn("text-2xl font-bold", getStarColor(data.rating.average))}>
                    {data.rating.average.toFixed(1)}
                  </span>
                  <span className="text-sm text-text-muted">
                    /5 ({formatNumber(data.rating.count)})
                  </span>
                </div>
              </div>

              {/* Revenue */}
              <div className="p-3 rounded-lg bg-bg-secondary border border-border-subtle">
                <div className="flex items-center gap-2 text-text-muted mb-2">
                  <DollarSign className="h-4 w-4" aria-hidden="true" />
                  <span className="text-xs uppercase tracking-wider">Price Model</span>
                </div>
                <div className="flex flex-col">
                  <span className="text-lg font-bold text-text-primary capitalize">
                    {data.pricing.model.replace("_", " ")}
                  </span>
                  {data.pricing.model !== "free" && data.pricing.pricePerCall && (
                    <span className="text-xs text-text-muted">
                      {formatPricePerCall(data.pricing.pricePerCall / 100, data.pricing.currency)}
                    </span>
                  )}
                </div>
              </div>
            </div>

            {/* Additional Metrics */}
            <div className="grid grid-cols-3 gap-4 p-3 rounded-lg bg-bg-secondary/50">
              <div className="text-center">
                <div className="text-xs text-text-muted mb-1">Average Latency</div>
                <div className="text-sm font-medium text-text-primary">
                  {data.metrics.averageLatency ? `${data.metrics.averageLatency}ms` : "N/A"}
                </div>
              </div>
              <div className="text-center border-x border-border-subtle">
                <div className="text-xs text-text-muted mb-1">Error Rate</div>
                <div
                  className={cn(
                    "text-sm font-medium",
                    (data.metrics.errorRate || 0) > 5
                      ? "text-red-400"
                      : (data.metrics.errorRate || 0) > 1
                      ? "text-amber-400"
                      : "text-emerald-400"
                  )}
                >
                  {data.metrics.errorRate ? `${data.metrics.errorRate.toFixed(2)}%` : "0%"}
                </div>
              </div>
              <div className="text-center">
                <div className="text-xs text-text-muted mb-1">Deterministic</div>
                <div className="flex justify-center">
                  {data.isDeterministic ? (
                    <Badge variant="success" className="text-xs gap-1">
                      <Check className="h-3 w-3" />
                      Yes
                    </Badge>
                  ) : (
                    <Badge variant="outline" className="text-xs gap-1 text-text-secondary">
                      <X className="h-3 w-3" />
                      No
                    </Badge>
                  )}
                </div>
              </div>
            </div>

            {/* Status Badges */}
            <div className="flex flex-wrap gap-2 mt-4">
              <DeterministicBadge isDeterministic={data.isDeterministic} />
              {data.category && (
                <Badge variant="secondary" className="text-xs">
                  {data.category}
                </Badge>
              )}
              {data.language && (
                <Badge variant="outline" className="text-xs">
                  {data.language}
                </Badge>
              )}
              {data.isFeatured && (
                <Badge variant="default" className="text-xs gap-1">
                  <Star className="h-3 w-3" />
                  Featured
                </Badge>
              )}
            </div>
          </CardContent>

          <CardFooter className="pt-0 border-t border-border-subtle">
            <div className="flex items-center justify-between w-full">
              <div className="flex items-center gap-4 text-xs text-text-muted">
                {data.version && (
                  <span className="flex items-center gap-1">
                    <Code className="h-3.5 w-3.5" />
                    v{data.version}
                  </span>
                )}
                {data.lastUpdated && (
                  <span className="flex items-center gap-1">
                    <Clock className="h-3.5 w-3.5" />
                    Updated {new Date(data.lastUpdated).toLocaleDateString()}
                  </span>
                )}
              </div>

              <div className="flex items-center gap-2">
                {onView && (
                  <Button variant="outline" size="sm" onClick={handleView}>
                    <ExternalLink className="h-4 w-4 mr-2" />
                    Details
                  </Button>
                )}
                {onExecute && (
                  <Button size="sm" onClick={handleExecute}>
                    <Play className="h-4 w-4 mr-2" />
                    Execute
                  </Button>
                )}
              </div>
            </div>
          </CardFooter>
        </Card>
      );
    }

    return null;
  }
);

FunctionCard.displayName = "FunctionCard";

// ============================================================================
// Export Variants for Convenience
// ============================================================================

/**
 * Compact FunctionCard for marketplace grid view
 */
export function FunctionCardCompact(
  props: Omit<FunctionCardProps, "variant">
) {
  return <FunctionCard {...props} variant="compact" />;
}

/**
 * Expanded FunctionCard for search results
 */
export function FunctionCardExpanded(
  props: Omit<FunctionCardProps, "variant">
) {
  return <FunctionCard {...props} variant="expanded" />;
}

/**
 * Analytics FunctionCard for internal dashboard
 */
export function FunctionCardAnalytics(
  props: Omit<FunctionCardProps, "variant">
) {
  return <FunctionCard {...props} variant="analytics" />;
}

export { FunctionCard };
export type { FunctionCardProps, FunctionCardData };
