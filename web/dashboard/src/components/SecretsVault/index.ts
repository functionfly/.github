/**
 * Secrets Vault Components - Barrel Export
 *
 * This module exports all components for the Secrets Vault feature,
 * which provides secure client-side encrypted secret storage.
 *
 * @example
 * ```tsx
 * import { SecretList, SecretForm, SecretDetail } from '@/components/SecretsVault';
 *
 * function SecretsPage() {
 *   return <SecretList />;
 * }
 * ```
 */

// Main components
export { SecretList } from "./SecretList";
export { SecretForm } from "./SecretForm";
export { SecretDetail } from "./SecretDetail";
export { TokenGenerator } from "./TokenGenerator";
export { AuditLog } from "./AuditLog";

// Dashboard components
export { VaultOverviewCard } from "./VaultOverviewCard";
export { VaultUsageChart } from "./VaultUsageChart";
export { SecretCountBadge } from "./SecretCountBadge";
export { VaultSecurityLevelIndicator } from "./VaultSecurityLevelIndicator";

// Secret Management components - Part 1
export { SecretCard } from "./SecretCard";
export { SecretRow } from "./SecretRow";
export { SecretRevealGate } from "./SecretRevealGate";

// Secret Management components - Part 2
export { SecretScopeSelector } from "./SecretScopeSelector";
export { SecretUsageTimeline } from "./SecretUsageTimeline";
export { SecretRotationModal } from "./SecretRotationModal";

// Types
export type { SecretListProps } from "./SecretList";
export type { SecretFormProps } from "./SecretForm";
export type { SecretDetailProps } from "./SecretDetail";
export type { TokenGeneratorProps } from "./TokenGenerator";
export type { AuditLogProps } from "./AuditLog";

// Dashboard component types
export type { VaultOverviewCardProps, VaultHealthStatus } from "./VaultOverviewCard";
export type { VaultUsageChartProps, UsageDataPoint, TrendDirection } from "./VaultUsageChart";
export type {
  SecretCountBadgeProps,
  GroupedSecretCountBadgeProps,
  CountThreshold,
} from "./SecretCountBadge";
export type {
  VaultSecurityLevelIndicatorProps,
  SecurityLevel,
  SecurityFactors,
} from "./VaultSecurityLevelIndicator";

// Secret Management component types - Part 1
export type {
  SecretCardProps,
  SecretWithStatus,
  SecretStatus,
} from "./SecretCard";
export type {
  SecretRowProps,
  SecretRowData,
  SecretRowStatus,
} from "./SecretRow";
export type {
  SecretRevealGateProps,
  SecurityLevel as RevealGateSecurityLevel,
  AuthMethod,
  VerificationResult,
} from "./SecretRevealGate";

// Secret Management component types - Part 2
export type {
  SecretScopeSelectorProps,
  Scope,
  ScopeType,
  ScopeWithSelection,
} from "./SecretScopeSelector";
export type {
  SecretUsageTimelineProps,
  TimelineEvent,
  TimelineEventType,
  TimelineActor,
  TimelineLocation,
  TimelineFilters,
} from "./SecretUsageTimeline";
export type {
  SecretRotationModalProps,
  RotationType,
  RotationStep,
  RotationRequest,
  RotationResult,
  ImpactedService,
  RotationOptions,
} from "./SecretRotationModal";

// Secret Management components - Part 3
export { SecretAuditDrawer } from "./SecretAuditDrawer";
export { SecretPermissionMatrix } from "./SecretPermissionMatrix";
export { SecretLeaseTimer } from "./SecretLeaseTimer";
export { SecretRevokeButton } from "./SecretRevokeButton";

// Secret Management component types - Part 3
export type {
  SecretAuditDrawerProps,
  ExportFormat,
  DateRange,
} from "./SecretAuditDrawer";
export type {
  SecretPermissionMatrixProps,
  PermissionType,
  PermissionUser,
  PermissionRole,
  PermissionEntry,
} from "./SecretPermissionMatrix";
export type {
  SecretLeaseTimerProps,
  TimerVariant,
  TimerSize,
  LeaseState,
  LeaseDetailsProps,
} from "./SecretLeaseTimer";
export type {
  SecretRevokeButtonProps,
  RevokeMode,
  ImpactedService as RevokeImpactedService,
  RevokeOptions,
  RevokeRequest,
  RevokeResult,
  RevokeStep,
} from "./SecretRevokeButton";
