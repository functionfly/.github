/**
 * Trust Components Index
 *
 * Re-exports all trust-related UI components for the FunctionFly dashboard.
 * These components provide trust visualization, verification status, and trust score displays.
 */

export {
  TrustBadge,
  TrustBadgeSkeleton,
  type TrustBadgeProps,
  type VerificationLevel,
} from './TrustBadge';
export {
  TrustDashboardWidget,
  TrustDashboardWidgetSkeleton,
  type TrustDashboardWidgetProps,
  type TrustMetricsSummary,
} from './TrustDashboardWidget';
export {
  TrustHistory,
  TrustHistorySkeleton,
  type TrustHistoryDataPoint,
  type TrustHistoryProps,
} from './TrustHistory';
export {
  TrustScore,
  TrustScoreSkeleton,
  type TrustScoreProps,
  type TrustTrend,
} from './TrustScore';
export {
  VerificationStatus,
  VerificationStatusSkeleton,
  type VerificationItem,
  type VerificationStatusProps,
  type VerificationStatusType,
} from './VerificationStatus';

// Re-export utility functions from existing trust components
export { getTrustColorConfig, getTrustScoreBand } from '@/components/functions/TrustScoreBadge';
