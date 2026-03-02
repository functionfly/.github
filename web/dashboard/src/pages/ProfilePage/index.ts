/**
 * ProfilePage Module
 *
 * Barrel export for all ProfilePage components and utilities.
 */

// Transformers
export {
  transformRegistryFunction,
  generateEmptyContributionGraph,
  transformToUserProfile,
} from "./transformers";

// Animations
export {
  containerVariants,
  itemVariants,
  tabContentVariants,
} from "./animations";

// Skeleton components
export {
  ProfileHeaderSkeleton,
  StatsOverviewSkeleton,
  TabContentSkeleton,
} from "./components/Skeletons";

// Main components
export { ProfileHeader } from "./components/ProfileHeader";
export type { ProfileHeaderProps } from "./components/ProfileHeader";

export { StatsOverview } from "./components/StatsOverview";
export type { StatsOverviewProps } from "./components/StatsOverview";

export { ContributionGraph } from "./components/ContributionGraph";
export type { ContributionGraphProps } from "./components/ContributionGraph";

export { AchievementsSection } from "./components/AchievementsSection";
export type { AchievementsSectionProps } from "./components/AchievementsSection";

export { SkillsSection } from "./components/SkillsSection";
export type { SkillsSectionProps } from "./components/SkillsSection";

export { TrustMetricsSection } from "./components/TrustMetricsSection";
export type { TrustMetricsSectionProps } from "./components/TrustMetricsSection";

export { ActivityTimeline } from "./components/ActivityTimeline";
export type { ActivityTimelineProps } from "./components/ActivityTimeline";

// Tab components
export { OverviewTab } from "./components/tabs/OverviewTab";
export type { OverviewTabProps } from "./components/tabs/OverviewTab";

export { FunctionsTab } from "./components/tabs/FunctionsTab";
export type { FunctionsTabProps } from "./components/tabs/FunctionsTab";

export { ActivityTab } from "./components/tabs/ActivityTab";
export type { ActivityTabProps } from "./components/tabs/ActivityTab";

export { AnalyticsTab } from "./components/tabs/AnalyticsTab";
export type { AnalyticsTabProps } from "./components/tabs/AnalyticsTab";

export { AboutTab } from "./components/tabs/AboutTab";
export type { AboutTabProps } from "./components/tabs/AboutTab";

export { SettingsTab } from "./components/tabs/SettingsTab";

// Main page component
export { ProfilePage, default } from "./ProfilePage";
export type { ProfilePageProps } from "./ProfilePage";
