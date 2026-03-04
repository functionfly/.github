/**
 * FlyAssistant - AI Assistant Component Library
 *
 * A comprehensive AI assistant interface for the Fly.io dashboard, providing
 * contextual chat, insights, execution previews, and marketplace integration.
 *
 * @packageDocumentation
 * @module fly-assistant
 * @example
 * ```tsx
 * import {
 *   FlyAssistantProvider,
 *   FlyAssistantPortal,
 *   FlyBubbleTrigger,
 *   FlyAssistantPanel,
 *   useFlyAssistantContext,
 * } from '@/components/fly-assistant';
 *
 * function App() {
 *   return (
 *     <FlyAssistantProvider>
 *       <FlyAssistantPortal>
 *         <FlyBubbleTrigger />
 *         <FlyAssistantPanel>
 *           // Assistant content
 *         </FlyAssistantPanel>
 *       </FlyAssistantPortal>
 *     </FlyAssistantProvider>
 *   );
 * }
 * ```
 */

// ============================================================================
// Types (export first for type-dependent components)
// ============================================================================

export * from './types';

// ============================================================================
// Composite Hook - Convenient all-in-one hook
// ============================================================================

export {
  useFlyAssistantContext,
} from './useFlyAssistantContext';

export type {
  UseFlyAssistantContextReturn,
} from './useFlyAssistantContext';

// ============================================================================
// Phase 1: Core Structural Components
// ============================================================================

/**
 * Provider & Context - Global state management for FlyAssistant
 *
 * The FlyAssistantProvider wraps your application and provides global state
 * management using Zustand. It includes several specialized hooks for
 * accessing different parts of the state.
 */
export {
  FlyAssistantProvider,
  useFlyAssistant,
  useFlyAssistantStore,
  useFlyAssistantActions,
  useFlyAssistantUser,
  useFlyAssistantStatus,
  useFlyAssistantCache,
} from './FlyAssistantProvider';

export type {
  FlyAssistantProviderProps,
  FlyAssistantState,
  FlyAssistantActions,
  FlyAssistantUserState,
  FlyAssistantStatusState,
  FlyAssistantCacheState,
} from './types';

/**
 * Portal - DOM portal for rendering outside component tree
 *
 * FlyAssistantPortal renders children into a dedicated DOM node outside
 * the normal React tree, ensuring proper z-index stacking and positioning.
 */
export {
  FlyAssistantPortal,
  useFlyAssistantPortal,
  PORTAL_CONTAINER_ID,
  PORTAL_Z_INDEX,
} from './FlyAssistantPortal';

export type {
  FlyAssistantPortalProps,
} from './types';

/**
 * Bubble Trigger - Floating action button to open assistant
 *
 * FlyBubbleTrigger is the floating button users click to open the
 * assistant panel. It includes notification badges and animations.
 */
export {
  FlyBubbleTrigger,
} from './FlyBubbleTrigger';

export type {
  FlyBubbleTriggerProps,
} from './types';

/**
 * Panel - Main assistant container with resize/minimize/fullscreen
 *
 * FlyAssistantPanel is the main container for the assistant interface,
 * supporting multiple display modes and sizes.
 */
export {
  FlyAssistantPanel,
} from './FlyAssistantPanel';

export type {
  FlyAssistantPanelProps,
  PanelMode,
  PanelSize,
} from './types';

// ============================================================================
// Phase 2: Chat System Components
// ============================================================================

/**
 * Chat Components - Internal message engine
 *
 * These components handle message display, input, and typing indicators
 * for the chat interface.
 */
export {
  // Chat Window
  FlyChatWindow,
  // Message Component
  FlyMessage,
  // Typing Indicators
  FlyTypingIndicator,
  FlyTypingIndicatorCompact,
  FlyTypingSkeleton,
  // Input Bar
  FlyInputBar,
} from './chat';

export type {
  // Chat Window Types
  FlyChatWindowProps,
  // Message Types
  FlyMessageProps,
  MessageVariant,
  MessageAction,
  MarketplaceSuggestion,
  // Typing Indicator Types
  FlyTypingIndicatorProps,
  FlyTypingIndicatorCompactProps,
  FlyTypingSkeletonProps,
  // Input Bar Types
  FlyInputBarProps,
  SlashCommand,
  Attachment,
} from './types';

// ============================================================================
// Phase 3: Context Intelligence Components
// ============================================================================

/**
 * Insights Components - Context-aware intelligence
 *
 * These components provide contextual awareness, proactive insights,
 * trust scoring, and marketplace integration.
 */
export {
  // Context Badge
  FlyContextBadge,
  // Insight Card with Presets
  FlyInsightCard,
  InsightPresets,
  // Trust Score Widget
  FlyTrustScoreWidget,
  // Marketplace Preview
  FlyMarketplacePreview,
} from './insights';

export type {
  // Context Badge Types
  FlyContextBadgeProps,
  ContextInfo,
  // Insight Card Types
  FlyInsightCardProps,
  Insight,
  InsightType,
  InsightAction,
  // Trust Score Types
  FlyTrustScoreWidgetProps,
  // Marketplace Preview Types
  FlyMarketplacePreviewProps,
  MarketplaceFunction,
  LatencyIndicator,
} from './types';

// ============================================================================
// Phase 4: Action & Execution Components
// ============================================================================

/**
 * Execution Components - Action safety and visibility
 *
 * These components connect the chat to real actions and provide safety
 * and visibility for AI-suggested executions.
 */
export {
  // Quick Actions
  FlyQuickActions,
  // Execution Preview with Safety
  FlyExecutionPreview,
  // Diff Viewer
  FlyDiffViewer,
} from './execution';

export type {
  // Quick Actions Types
  FlyQuickActionsProps,
  QuickAction,
  ActionVariant,
  ActionState,
  ActionIconName,
  // Execution Preview Types
  FlyExecutionPreviewProps,
  RiskLevel,
  AffectedFunction,
  ExecutionPreviewData,
  // Diff Viewer Types
  FlyDiffViewerProps,
  DiffViewMode,
  ChangeType,
  DiffLine,
  DiffHunk,
  DiffData,
} from './types';

// ============================================================================
// Phase 5: UX Polish Components
// ============================================================================

/**
 * UI Components - UX polish and power-user features
 *
 * These components enhance the user experience with notifications,
 * auto-nudges, memory timeline, command palette, and mode switching.
 */
export {
  // Notification Dot
  FlyNotificationDot,
  // Auto Nudge
  FlyAutoNudge,
  // Memory Timeline
  FlyMemoryTimeline,
  // Command Palette
  FlyCommandPalette,
  // Mode Switcher
  FlyModeSwitcher,
} from './ui';

export type {
  // Notification Types
  FlyNotificationDotProps,
  // Nudge Types
  FlyAutoNudgeProps,
  // Memory Types
  FlyMemoryTimelineProps,
  ConversationSummary,
  // Command Palette Types
  FlyCommandPaletteProps,
  // Mode Switcher Types
  FlyModeSwitcherProps,
  AssistantMode,
} from './types';

// ============================================================================
// Phase 6: Infrastructure Components
// ============================================================================

/**
 * Infrastructure Components - Performance, tracking, and permissions
 *
 * These components handle route listening, event tracking, and
 * permission-based access control.
 */

// Route Listener
export {
  FlyRouteListener,
  generateDeepLink,
} from './FlyRouteListener';

export type {
  FlyRouteListenerProps,
} from './types';

// Event Tracker
export {
  FlyEventTracker,
  useEventTracking as useBaseEventTracking,
} from './FlyEventTracker';

export type {
  FlyEventTrackerProps,
  TrackedEvent,
  TrackedEventType,
  ErrorEventData,
  DeploymentEventData,
  TrustChangeEventData,
  MarketplaceEventData,
  AssistantUsageEventData,
} from './types';

// Permission Guard
export {
  FlyPermissionGuard,
  ExecutionGuard,
  TierBadge,
} from './FlyPermissionGuard';

export type {
  FlyPermissionGuardProps,
  ExecutionGuardProps,
  TierBadgeProps,
} from './types';

// ============================================================================
// Hooks (Phase 6)
// ============================================================================

/**
 * Custom Hooks - Route tracking, permissions, and event tracking
 *
 * These hooks provide reusable logic for route tracking, permission
 * checking, and event tracking throughout your application.
 */
export {
  // Route tracking
  useRouteTracking,
  // Event tracking
  useEventTracking,
  // Permission management
  usePermission,
  useIsPro,
  useIsEnterprise,
  useTierComparison,
  checkTierAccess,
  getNextTier,
  getTierIndex,
} from './hooks';

export type {
  UseRouteTrackingReturn,
  UseEventTrackingReturn,
} from './types';

// ============================================================================
// Re-exports for convenience
// ============================================================================

/**
 * Core types re-exported for convenient access
 */
export type {
  // Core types
  UserRole,
  TrustTier,
  // Message types
  Message,
  MessageMetadata,
  // User types
  UserSession,
  // Conversation types
  ConversationEntry,
} from './types';
