/**
 * FlyAssistant Centralized Types
 *
 * Single source of truth for all shared type definitions across the FlyAssistant
 * component system. These types are used across multiple components and hooks.
 *
 * @module fly-assistant/types
 */

// ============================================================================
// Core Types
// ============================================================================

/**
 * User role tier for feature gating and permissions
 */
export type UserRole = "free" | "pro" | "enterprise";

/**
 * Trust tier based on confidence score
 */
export type TrustTier = "low" | "medium" | "high" | "critical";

/**
 * Assistant operating modes
 */
export type AssistantMode = "chat" | "insight" | "marketplace" | "onboarding";

/**
 * Panel display sizes
 */
export type PanelSize = "compact" | "default" | "large" | "fullscreen";

/**
 * Panel display modes
 */
export type PanelMode = "bubble" | "sidebar" | "drawer" | "modal";

// ============================================================================
// Message Types
// ============================================================================

/**
 * Individual message in a conversation
 */
export interface Message {
  id: string;
  role: "user" | "assistant" | "system";
  content: string;
  timestamp: number;
  metadata?: MessageMetadata;
}

/**
 * Extended message metadata for rich messages
 */
export interface MessageMetadata {
  /** Quick actions attached to the message */
  actions?: QuickAction[];
  /** Insights related to the message */
  insights?: Insight[];
  /** Execution preview data if applicable */
  execution?: ExecutionPreviewData;
  /** Diff data for code changes */
  diff?: DiffData;
  /** Whether the message is streaming */
  isStreaming?: boolean;
}

/**
 * Message variant types for visual styling
 */
export type MessageVariant = "user" | "ai" | "system" | "marketplace" | "warning";

/**
 * Action that can be attached to a message
 */
export interface MessageAction {
  id: string;
  label: string;
  variant?: "primary" | "secondary" | "danger" | "ghost";
  icon?: React.ReactNode;
}

/**
 * Marketplace suggestion data for marketplace messages
 */
export interface MarketplaceSuggestion {
  id: string;
  name: string;
  description: string;
  author: string;
  rating: number;
  downloads: number;
  tags: string[];
}

// ============================================================================
// Context Types
// ============================================================================

/**
 * Current route/page context for contextual assistance
 */
export interface RouteContext {
  path: string;
  name: string;
  params?: Record<string, string>;
}

/**
 * Extended route context with computed properties
 */
export interface ExtendedRouteContext extends RouteContext {
  /** Whether currently on a function page */
  isFunctionPage: boolean;
  /** Whether currently on the marketplace */
  isMarketplacePage: boolean;
  /** Whether currently on dashboard home */
  isDashboard: boolean;
  /** Function ID if on function page */
  functionId?: string;
  /** Category if on marketplace */
  category?: string;
}

/**
 * Context information for display in badges
 */
export interface ContextInfo {
  page: string;
  function?: string;
  category?: string;
  trustTier: TrustTier;
}

// ============================================================================
// Conversation Types
// ============================================================================

/**
 * Cached conversation entry for memory persistence
 */
export interface ConversationEntry {
  id: string;
  timestamp: number;
  messages: Message[];
  context?: RouteContext;
}

/**
 * Summary of a conversation for timeline display
 */
export interface ConversationSummary {
  id: string;
  preview: string;
  timestamp: Date;
  messageCount: number;
}

// ============================================================================
// User Types
// ============================================================================

/**
 * User session state
 */
export interface UserSession {
  id: string;
  email: string;
  role: UserRole;
  orgId?: string;
  name?: string;
  avatar?: string;
}

// ============================================================================
// Insight Types
// ============================================================================

/**
 * Insight type for visual styling
 */
export type InsightType = "info" | "warning" | "success" | "error";

/**
 * Action that can be performed on an insight
 */
export interface InsightAction {
  id: string;
  label: string;
  variant?: "primary" | "secondary" | "ghost" | "destructive";
  icon?: React.ReactNode;
  onClick: () => void;
}

/**
 * Proactive insight data structure
 */
export interface Insight {
  id: string;
  title: string;
  description: string;
  type: InsightType;
  dismissible: boolean;
  actions?: InsightAction[];
}

// ============================================================================
// Marketplace Types
// ============================================================================

/**
 * Latency indicator for marketplace functions
 */
export type LatencyIndicator = "fast" | "medium" | "slow";

/**
 * Marketplace function information
 */
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
  rating?: number;
  downloads?: number;
  tags?: string[];
}

// ============================================================================
// Action Types
// ============================================================================

/**
 * Action variant for styling
 */
export type ActionVariant = "primary" | "secondary" | "danger" | "ghost";

/**
 * Action state for async operations
 */
export type ActionState = "idle" | "loading" | "success" | "error";

/**
 * Available icon names for quick actions
 */
export type ActionIconName =
  | "deploy"
  | "rollback"
  | "config"
  | "test"
  | "delete"
  | "clone"
  | "share"
  | "settings";

/**
 * Quick action button configuration
 */
export interface QuickAction {
  id: string;
  label: string;
  icon: string;
  variant: ActionVariant;
  loading?: boolean;
  disabled?: boolean;
  onClick?: () => void;
}

// ============================================================================
// Execution Types
// ============================================================================

/**
 * Risk level for execution preview
 */
export type RiskLevel = "low" | "medium" | "high" | "critical";

/**
 * Affected function information in execution preview
 */
export interface AffectedFunction {
  id: string;
  name: string;
  status: "active" | "inactive" | "error";
  impact: string;
  invocationsAffected?: number;
}

/**
 * Execution preview data for safety review
 */
export interface ExecutionPreviewData {
  title: string;
  description: string;
  affectedFunctions: AffectedFunction[];
  estimatedCost: number;
  estimatedLatency: number;
  riskLevel: RiskLevel;
  trustImpact?: number;
}

/**
 * Diff view modes
 */
export type DiffViewMode = "split" | "unified";

/**
 * Type of change in a diff
 */
export type ChangeType = "addition" | "deletion" | "context";

/**
 * Single line in a diff
 */
export interface DiffLine {
  type: ChangeType;
  content: string;
  lineNumber?: number;
  oldLineNumber?: number;
  newLineNumber?: number;
}

/**
 * Hunk of changes in a diff
 */
export interface DiffHunk {
  oldStart: number;
  oldLength: number;
  newStart: number;
  newLength: number;
  lines: DiffLine[];
  header?: string;
}

/**
 * Diff data for before/after comparison
 */
export interface DiffData {
  before: string;
  after: string;
  beforeLabel?: string;
  afterLabel?: string;
  language?: string;
}

// ============================================================================
// Event Tracking Types
// ============================================================================

/**
 * Event types that can be tracked
 */
export type TrackedEventType =
  | "error"
  | "deployment"
  | "trust_change"
  | "marketplace_view"
  | "marketplace_install"
  | "marketplace_rate"
  | "assistant_message"
  | "assistant_action"
  | "assistant_open"
  | "assistant_close"
  | "execution_preview"
  | "execution_confirm"
  | "execution_cancel"
  | "page_view"
  | "custom";

/**
 * Tracked event structure
 */
export interface TrackedEvent {
  id: string;
  type: TrackedEventType;
  timestamp: number;
  data: Record<string, unknown>;
  context: {
    route: string;
    function?: string;
    userRole: UserRole;
    sessionId?: string;
  };
  sequenceNumber: number;
}

/**
 * Error event data
 */
export interface ErrorEventData {
  message: string;
  stack?: string;
  source?: string;
  lineno?: number;
  colno?: number;
}

/**
 * Deployment event data
 */
export interface DeploymentEventData {
  functionId: string;
  version: string;
  environment: string;
  success: boolean;
  duration?: number;
  error?: string;
}

/**
 * Trust change event data
 */
export interface TrustChangeEventData {
  previousScore: number;
  newScore: number;
  reason?: string;
}

/**
 * Marketplace event data
 */
export interface MarketplaceEventData {
  functionId: string;
  functionName: string;
  category?: string;
  rating?: number;
}

/**
 * Assistant usage event data
 */
export interface AssistantUsageEventData {
  messageCount?: number;
  actionType?: string;
  mode?: string;
}

// ============================================================================
// Permission Types
// ============================================================================

/**
 * Permission check result
 */
export interface PermissionResult {
  hasPermission: boolean;
  isPro: boolean;
  isEnterprise: boolean;
  isFree: boolean;
  currentTier: UserRole;
  nextTier: UserRole | null;
  showUpgradePrompt: () => void;
  getMissingTierMessage: (featureName: string) => string;
}

/**
 * Tier comparison utilities
 */
export interface TierComparisonResult {
  isAtLeast: (tier: UserRole) => boolean;
  isExactly: (tier: UserRole) => boolean;
  isLowerThan: (tier: UserRole) => boolean;
  nextTier: UserRole | null;
}

// ============================================================================
// Provider Types
// ============================================================================

/**
 * Complete FlyAssistant state (from Zustand store)
 */
export interface FlyAssistantState {
  // UI State
  isOpen: boolean;
  isMinimized: boolean;
  isFullscreen: boolean;

  // User Context
  userSession: UserSession | null;
  currentRoute: RouteContext | null;

  // Trust & Quality
  trustScore: number;
  trustTier: TrustTier;

  // Error State
  hasError: boolean;
  errorMessage: string | null;

  // Insights & Notifications
  hasInsights: boolean;
  notificationCount: number;

  // Memory Cache
  conversationCache: Map<string, ConversationEntry>;
  currentConversationId: string | null;

  // Actions
  open: () => void;
  close: () => void;
  toggle: () => void;
  minimize: () => void;
  expand: () => void;
  setFullscreen: (value: boolean) => void;
  setUserSession: (session: UserSession | null) => void;
  setCurrentRoute: (route: RouteContext | null) => void;
  setTrustScore: (score: number) => void;
  setError: (error: string | null) => void;
  setHasInsights: (value: boolean) => void;
  setNotificationCount: (count: number) => void;
  addToCache: (entry: ConversationEntry) => void;
  clearCache: () => void;
  setCurrentConversation: (id: string | null) => void;
}

/**
 * FlyAssistant UI actions (subset of state)
 */
export interface FlyAssistantActions {
  open: () => void;
  close: () => void;
  toggle: () => void;
  minimize: () => void;
  expand: () => void;
  setFullscreen: (value: boolean) => void;
}

/**
 * FlyAssistant user state
 */
export interface FlyAssistantUserState {
  userSession: UserSession | null;
  currentRoute: RouteContext | null;
  setUserSession: (session: UserSession | null) => void;
  setCurrentRoute: (route: RouteContext | null) => void;
}

/**
 * FlyAssistant status state
 */
export interface FlyAssistantStatusState {
  trustScore: number;
  trustTier: TrustTier;
  hasError: boolean;
  errorMessage: string | null;
  hasInsights: boolean;
  notificationCount: number;
  setTrustScore: (score: number) => void;
  setError: (error: string | null) => void;
  setHasInsights: (value: boolean) => void;
  setNotificationCount: (count: number) => void;
}

/**
 * FlyAssistant cache state
 */
export interface FlyAssistantCacheState {
  conversationCache: Map<string, ConversationEntry>;
  currentConversationId: string | null;
  addToCache: (entry: ConversationEntry) => void;
  clearCache: () => void;
  setCurrentConversation: (id: string | null) => void;
}

// ============================================================================
// Hook Return Types
// ============================================================================

/**
 * Return type for useRouteTracking hook
 */
export interface UseRouteTrackingReturn {
  route: ExtendedRouteContext | null;
  pageName: string;
  path: string;
  params: Record<string, string>;
  isReady: boolean;
  refresh: () => void;
}

/**
 * Return type for useEventTracking hook
 */
export interface UseEventTrackingReturn {
  track: (eventName: string, data?: Record<string, unknown>) => void;
  trackError: (error: Error, context?: Record<string, unknown>) => void;
  trackDeployment: (data: DeploymentEventData) => void;
  trackTrustChange: (data: TrustChangeEventData) => void;
  trackMarketplaceView: (data: MarketplaceEventData) => void;
  trackMarketplaceInstall: (data: MarketplaceEventData) => void;
  trackMarketplaceRate: (data: MarketplaceEventData) => void;
  trackAssistantMessage: (data?: AssistantUsageEventData) => void;
  trackAssistantAction: (actionType: string, data?: Record<string, unknown>) => void;
  trackAssistantOpen: (mode?: string) => void;
  trackAssistantClose: () => void;
  trackPageView: (pageName: string, data?: Record<string, unknown>) => void;
  flush: () => void;
  isEnabled: boolean;
  pendingCount: number;
}

// ============================================================================
// Component Props Types
// ============================================================================

/**
 * Provider component props
 */
export interface FlyAssistantProviderProps {
  children: React.ReactNode;
  initialSession?: UserSession | null;
  initialRoute?: RouteContext | null;
}

/**
 * Portal component props
 */
export interface FlyAssistantPortalProps {
  children: React.ReactNode;
  containerId?: string;
  zIndex?: number;
}

/**
 * Bubble trigger props
 */
export interface FlyBubbleTriggerProps {
  onClick?: () => void;
  notificationCount?: number;
  isOpen?: boolean;
  className?: string;
  position?: "bottom-right" | "bottom-left" | "top-right" | "top-left";
}

/**
 * Panel component props
 */
export interface FlyAssistantPanelProps {
  children: React.ReactNode;
  mode?: PanelMode;
  size?: PanelSize;
  isOpen?: boolean;
  isMinimized?: boolean;
  isFullscreen?: boolean;
  onClose?: () => void;
  onMinimize?: () => void;
  onExpand?: () => void;
  onFullscreenToggle?: (isFullscreen: boolean) => void;
  className?: string;
}

/**
 * Route listener props
 */
export interface FlyRouteListenerProps {
  onRouteChange?: (context: RouteContext) => void;
  debounceMs?: number;
}

/**
 * Event tracker props
 */
export interface FlyEventTrackerProps {
  children: React.ReactNode;
  enabled?: boolean;
  endpoint?: string;
  batchSize?: number;
  flushIntervalMs?: number;
  onError?: (error: Error) => void;
}

/**
 * Permission guard props
 */
export interface FlyPermissionGuardProps {
  requiredTier: UserRole;
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

/**
 * Execution guard props
 */
export interface ExecutionGuardProps {
  requiredTier: UserRole;
  children: React.ReactNode;
  onUnauthorized?: () => void;
}

/**
 * Tier badge props
 */
export interface TierBadgeProps {
  tier: UserRole;
  showLabel?: boolean;
  size?: "sm" | "md" | "lg";
}

// ============================================================================
// Context Badge Props
// ============================================================================

export interface FlyContextBadgeProps {
  context: ContextInfo;
  showTrustScore?: boolean;
  className?: string;
}

// ============================================================================
// Insight Card Props
// ============================================================================

export interface FlyInsightCardProps {
  insight: Insight;
  onDismiss?: (id: string, dontShowAgain: boolean) => void;
  className?: string;
}

// ============================================================================
// Trust Score Widget Props
// ============================================================================

export interface FlyTrustScoreWidgetProps {
  score: number;
  previousScore?: number;
  showTrend?: boolean;
  size?: "sm" | "md" | "lg";
  className?: string;
}

// ============================================================================
// Marketplace Preview Props
// ============================================================================

export interface FlyMarketplacePreviewProps {
  functions: MarketplaceFunction[];
  onFunctionClick?: (func: MarketplaceFunction) => void;
  onInstall?: (func: MarketplaceFunction) => void;
  className?: string;
}

// ============================================================================
// Quick Actions Props
// ============================================================================

export interface FlyQuickActionsProps {
  actions: QuickAction[];
  layout?: "horizontal" | "vertical" | "grid";
  size?: "sm" | "md" | "lg";
  className?: string;
}

// ============================================================================
// Execution Preview Props
// ============================================================================

export interface FlyExecutionPreviewProps {
  title: string;
  description: string;
  affectedFunctions: AffectedFunction[];
  estimatedCost: number;
  estimatedLatency: number;
  riskLevel: RiskLevel;
  trustImpact?: number;
  onConfirm: () => void;
  onCancel: () => void;
  countdownSeconds?: number;
  className?: string;
}

// ============================================================================
// Diff Viewer Props
// ============================================================================

export interface FlyDiffViewerProps {
  before: string;
  after: string;
  beforeLabel?: string;
  afterLabel?: string;
  language?: string;
  viewMode?: DiffViewMode;
  showLineNumbers?: boolean;
  highlightSyntax?: boolean;
  className?: string;
}

// ============================================================================
// Chat Window Props
// ============================================================================

export interface FlyChatWindowProps {
  messages: Message[];
  isTyping?: boolean;
  onSendMessage?: (message: string) => void;
  onActionClick?: (actionId: string, messageId: string) => void;
  className?: string;
}

// ============================================================================
// Message Component Props
// ============================================================================

export interface FlyMessageProps {
  variant: MessageVariant;
  content: string;
  timestamp?: Date | number;
  actions?: MessageAction[];
  onAction?: (actionId: string) => void;
  onFeedback?: (type: "positive" | "negative") => void;
  isStreaming?: boolean;
  suggestion?: MarketplaceSuggestion;
  className?: string;
  userAvatar?: string;
  aiAvatar?: string;
}

// ============================================================================
// Typing Indicator Props
// ============================================================================

export interface FlyTypingIndicatorProps {
  text?: string;
  className?: string;
}

export interface FlyTypingIndicatorCompactProps {
  className?: string;
}

export interface FlyTypingSkeletonProps {
  lines?: number;
  className?: string;
}

// ============================================================================
// Input Bar Props
// ============================================================================

/**
 * Slash command definition
 */
export interface SlashCommand {
  id: string;
  command: string;
  description: string;
  icon?: string;
}

/**
 * Attachment file info
 */
export interface Attachment {
  id: string;
  name: string;
  size: number;
  type: string;
  url?: string;
}

export interface FlyInputBarProps {
  onSend?: (message: string, attachments?: Attachment[]) => void;
  onSlashCommand?: (command: string) => void;
  slashCommands?: SlashCommand[];
  disabled?: boolean;
  placeholder?: string;
  className?: string;
}

// ============================================================================
// UI Component Props
// ============================================================================

export interface FlyNotificationDotProps {
  count: number;
  maxCount?: number;
  showZero?: boolean;
  pulse?: boolean;
  className?: string;
}

export interface FlyAutoNudgeProps {
  title: string;
  description: string;
  actionLabel?: string;
  onAction?: () => void;
  onDismiss?: () => void;
  delayMs?: number;
  className?: string;
}

export interface FlyMemoryTimelineProps {
  conversations: ConversationSummary[];
  onConversationClick?: (id: string) => void;
  onClearAll?: () => void;
  className?: string;
}

export interface FlyCommandPaletteProps {
  commands: Array<{
    id: string;
    label: string;
    shortcut?: string;
    icon?: React.ReactNode;
    onExecute: () => void;
  }>;
  isOpen: boolean;
  onClose: () => void;
  className?: string;
}

export interface FlyModeSwitcherProps {
  currentMode: AssistantMode;
  onModeChange: (mode: AssistantMode) => void;
  availableModes?: AssistantMode[];
  disabledModes?: AssistantMode[];
  className?: string;
}
