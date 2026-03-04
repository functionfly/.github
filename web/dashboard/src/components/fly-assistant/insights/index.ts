/**
 * FlyAssistant Insights Components
 *
 * Context-aware intelligence components that make the assistant NOT generic.
 * These components provide contextual awareness, proactive insights,
 * trust scoring, and marketplace integration.
 *
 * @module fly-assistant/insights
 */

// ============================================================================
// Context Badge - Displays current assistant context
// ============================================================================

export { FlyContextBadge } from "./FlyContextBadge";
export type {
  FlyContextBadgeProps,
  ContextInfo,
} from "./FlyContextBadge";

// ============================================================================
// Insight Card - Proactive nudges and insights
// ============================================================================

export { FlyInsightCard, InsightPresets } from "./FlyInsightCard";
export type {
  FlyInsightCardProps,
  Insight,
  InsightType,
  InsightAction,
} from "./FlyInsightCard";

// ============================================================================
// Trust Score Widget - Gamified trust score display
// ============================================================================

export { FlyTrustScoreWidget } from "./FlyTrustScoreWidget";
export type {
  FlyTrustScoreWidgetProps,
} from "./FlyTrustScoreWidget";

// ============================================================================
// Marketplace Preview - Function suggestions from marketplace
// ============================================================================

export { FlyMarketplacePreview } from "./FlyMarketplacePreview";
export type {
  FlyMarketplacePreviewProps,
  MarketplaceFunction,
  LatencyIndicator,
} from "./FlyMarketplacePreview";
