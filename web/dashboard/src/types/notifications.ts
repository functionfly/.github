/**
 * Notification System Type Definitions
 *
 * TypeScript types and interfaces for the FunctionFly notification system,
 * including real-time alerts, activity feeds, and WebSocket events.
 */

// ============================================================================
// Notification Enums
// ============================================================================

/**
 * Categories for organizing notifications in the UI
 */
export type NotificationCategory =
  | 'all'
  | 'trust'
  | 'revenue'
  | 'issues'
  | 'messages'
  | 'security';

/**
 * Specific notification types for different events
 */
export type NotificationType =
  | 'reputation_gained'
  | 'function_error_spike'
  | 'trust_change'
  | 'issue_assigned'
  | 'fxcert_verified'
  | 'bounty_claimed'
  | 'trust_drop'
  | 'replay_failed'
  | 'determinism_broken';

/**
 * Priority levels for notifications
 */
export type NotificationPriority = 'low' | 'medium' | 'high' | 'critical';

/**
 * Status of a notification in the user's inbox
 */
export type NotificationStatus = 'unread' | 'read' | 'archived';

// ============================================================================
// Core Interfaces
// ============================================================================

/**
 * Metadata for reputation-related notifications
 */
export interface ReputationMetadata {
  amount: number;
  source: string;
  previousScore: number;
  newScore: number;
}

/**
 * Metadata for trust-related notifications
 */
export interface TrustMetadata {
  trustDelta: number;
  previousTrust: number;
  newTrust: number;
  reason?: string;
  functionId?: string;
  functionName?: string;
}

/**
 * Metadata for issue-related notifications
 */
export interface IssueMetadata {
  issueId: string;
  issueTitle: string;
  issueUrl?: string;
  assignerName?: string;
  severity?: 'low' | 'medium' | 'high' | 'critical';
}

/**
 * Metadata for bounty-related notifications
 */
export interface BountyMetadata {
  bountyAmount: number;
  currency: string;
  functionId?: string;
  functionName?: string;
  claimerName?: string;
}

/**
 * Metadata for FxCert verification notifications
 */
export interface FxCertMetadata {
  certId: string;
  certType: string;
  verifierName?: string;
  verificationUrl?: string;
}

/**
 * Metadata for function error spike notifications
 */
export interface ErrorSpikeMetadata {
  functionId: string;
  functionName: string;
  errorCount: number;
  timeWindow: number; // in minutes
  errorRate: number;
  sampleError?: string;
}

/**
 * Flexible metadata object for type-specific notification data
 */
export interface NotificationMetadata {
  // Reputation
  reputation?: ReputationMetadata;

  // Trust
  trust?: TrustMetadata;

  // Issues
  issue?: IssueMetadata;

  // Bounties
  bounty?: BountyMetadata;

  // FxCert
  fxCert?: FxCertMetadata;

  // Error spikes
  errorSpike?: ErrorSpikeMetadata;

  // Trust alerts (for TrustAlertBanner)
  affectedFunctions?: string[];
  affectedFunctionNames?: string[];
  replayAttemptCount?: number;
  lastReplayTimestamp?: string;
  determinismViolationDetails?: {
    expectedHash: string;
    actualHash: string;
    inputSample?: string;
  };

  // Generic additional data
  [key: string]: unknown;
}

/**
 * Base notification interface
 */
export interface Notification {
  /** Unique identifier for the notification */
  id: string;

  /** Type of notification event */
  type: NotificationType;

  /** Category for UI organization */
  category: NotificationCategory;

  /** Short title/summary */
  title: string;

  /** Detailed message content */
  message: string;

  /** ISO 8601 timestamp when the notification was created */
  timestamp: string;

  /** Priority level */
  priority: NotificationPriority;

  /** Current status in the user's inbox */
  status: NotificationStatus;

  /** Type-specific metadata */
  metadata: NotificationMetadata;

  /** ID of the user who should receive this notification */
  userId: string;

  /** ID of the tenant/organization */
  tenantId: string;

  /** Optional URL for taking action on the notification */
  actionUrl?: string;

  /** Optional icon identifier for the UI */
  icon?: string;

  /** When the notification was read (if applicable) */
  readAt?: string;

  /** When the notification was archived (if applicable) */
  archivedAt?: string;
}

/**
 * Trust alert interface for critical trust-related issues
 */
export interface TrustAlert {
  /** Unique identifier */
  id: string;

  /** Alert type for TrustAlertBanner compatibility */
  type: 'trust_drop' | 'replay_failed' | 'determinism_broken';

  /** Severity level */
  severity: 'warning' | 'critical' | 'emergency';

  /** Human-readable title */
  title: string;

  /** Detailed description */
  description: string;

  /** Functions affected by this trust issue */
  affectedFunctions: string[];

  /** Human-readable function names */
  affectedFunctionNames?: string[];

  /** Recommended action to resolve */
  recommendedAction: string;

  /** Action URL for resolution */
  actionUrl?: string;

  /** When the alert was triggered */
  triggeredAt: string;

  /** Whether the alert has been acknowledged */
  acknowledged: boolean;

  /** When the alert was acknowledged */
  acknowledgedAt?: string;

  /** User ID who acknowledged */
  acknowledgedBy?: string;

  /** Current trust score (if applicable) */
  currentTrustScore?: number;

  /** Trust score drop amount (if applicable) */
  trustDropAmount?: number;
}

/**
 * Activity feed item for real-time updates
 */
export interface ActivityFeedItem {
  /** Unique identifier */
  id: string;

  /** The entity performing the action */
  actor: {
    id: string;
    type: 'user' | 'system' | 'function' | 'agent';
    name: string;
    avatarUrl?: string;
  };

  /** The action performed */
  action:
    | 'created'
    | 'updated'
    | 'deleted'
    | 'deployed'
    | 'verified'
    | 'claimed'
    | 'earned'
    | 'lost'
    | 'assigned'
    | 'resolved'
    | 'commented'
    | 'starred'
    | 'forked'
    | 'published';

  /** The entity being acted upon */
  target: {
    id: string;
    type: 'function' | 'issue' | 'bounty' | 'user' | 'cert' | 'deployment' | 'app';
    name: string;
    url?: string;
  };

  /** ISO 8601 timestamp */
  timestamp: string;

  /** Optional additional context */
  context?: {
    description?: string;
    metadata?: Record<string, unknown>;
    icon?: string;
  };

  /** Tenant ID for filtering */
  tenantId?: string;
}

/**
 * Filter options for notification queries
 */
export interface NotificationFilter {
  /** Filter by category */
  category?: NotificationCategory;

  /** Filter by status */
  status?: NotificationStatus;

  /** Filter by priority */
  priority?: NotificationPriority;

  /** Filter by notification type */
  type?: NotificationType;

  /** Start date for date range filter (ISO 8601) */
  startDate?: string;

  /** End date for date range filter (ISO 8601) */
  endDate?: string;

  /** Search query for text matching */
  searchQuery?: string;

  /** Filter by read status */
  isRead?: boolean;

  /** Include archived notifications */
  includeArchived?: boolean;
}

/**
 * User preferences for notification settings
 */
export interface NotificationPreferences {
  /** User ID */
  userId: string;

  /** Tenant ID */
  tenantId: string;

  /** Email notification settings */
  email: {
    enabled: boolean;
    digestFrequency: 'immediate' | 'hourly' | 'daily' | 'weekly';
    categories: NotificationCategory[];
    minimumPriority: NotificationPriority;
  };

  /** In-app notification settings */
  inApp: {
    enabled: boolean;
    soundEnabled: boolean;
    desktopNotifications: boolean;
    categories: NotificationCategory[];
    minimumPriority: NotificationPriority;
  };

  /** Push notification settings */
  push: {
    enabled: boolean;
    categories: NotificationCategory[];
    minimumPriority: NotificationPriority;
    quietHoursStart?: string; // 24-hour format "HH:mm"
    quietHoursEnd?: string; // 24-hour format "HH:mm"
  };

  /** Webhook notification settings */
  webhook?: {
    enabled: boolean;
    url: string;
    secret?: string;
    categories: NotificationCategory[];
    minimumPriority: NotificationPriority;
    events: NotificationType[];
  };

  /** Category-specific overrides */
  categoryOverrides?: Partial<
    Record<
      NotificationCategory,
      {
        email?: boolean;
        inApp?: boolean;
        push?: boolean;
        minimumPriority?: NotificationPriority;
      }
    >
  >;

  /** When preferences were last updated */
  updatedAt: string;
}

// ============================================================================
// WebSocket Event Types
// ============================================================================

/**
 * Base WebSocket event interface
 */
export interface WebSocketEvent {
  /** Event type identifier */
  eventType: string;

  /** ISO 8601 timestamp */
  timestamp: string;

  /** Tenant ID for routing */
  tenantId?: string;

  /** User ID for routing */
  userId?: string;
}

/**
 * WebSocket message structure for new notifications
 */
export interface NotificationEvent extends WebSocketEvent {
  eventType: 'notification';

  /** The notification payload */
  notification: Notification;

  /** Whether this requires immediate user attention */
  urgent: boolean;

  /** Badge count update */
  unreadCount: number;
}

/**
 * WebSocket message for live activity updates
 */
export interface ActivityEvent extends WebSocketEvent {
  eventType: 'activity';

  /** The activity item */
  activity: ActivityFeedItem;

  /** Batch of activities for initial load */
  activityBatch?: ActivityFeedItem[];

  /** Whether this is an update to existing activity */
  isUpdate: boolean;

  /** For updates, the previous state */
  previousState?: Partial<ActivityFeedItem>;
}

/**
 * WebSocket message for trust alerts
 */
export interface TrustAlertEvent extends WebSocketEvent {
  eventType: 'trust_alert';

  /** The trust alert payload */
  alert: TrustAlert;

  /** Previous trust score for comparison */
  previousTrustScore?: number;

  /** Current trust score */
  currentTrustScore?: number;

  /** Whether this is a new alert or an update */
  isNew: boolean;

  /** Active alert count for the tenant */
  activeAlertCount: number;

  /** Requires immediate attention */
  requiresAction: boolean;
}

/**
 * WebSocket message for bulk notification updates
 */
export interface BulkNotificationEvent extends WebSocketEvent {
  eventType: 'bulk_notification';

  /** Operation type */
  operation: 'mark_read' | 'mark_unread' | 'archive' | 'delete';

  /** Affected notification IDs */
  notificationIds: string[];

  /** Updated unread count */
  unreadCount: number;

  /** Updated archive count */
  archivedCount?: number;
}

/**
 * WebSocket message for preference updates
 */
export interface PreferencesUpdateEvent extends WebSocketEvent {
  eventType: 'preferences_updated';

  /** Updated preferences */
  preferences: NotificationPreferences;

  /** Which sections were updated */
  updatedSections: Array<'email' | 'inApp' | 'push' | 'webhook' | 'categoryOverrides'>;
}

// ============================================================================
// Request/Response Types
// ============================================================================

/**
 * Request to mark notifications as read
 */
export interface MarkAsReadRequest {
  notificationIds: string[];
  markAll?: boolean;
  category?: NotificationCategory;
}

/**
 * Response from mark as read operation
 */
export interface MarkAsReadResponse {
  success: boolean;
  markedCount: number;
  unreadCount: number;
}

/**
 * Request to archive notifications
 */
export interface ArchiveRequest {
  notificationIds: string[];
  archiveAllRead?: boolean;
}

/**
 * Response from archive operation
 */
export interface ArchiveResponse {
  success: boolean;
  archivedCount: number;
}

/**
 * Request to update notification preferences
 */
export interface UpdatePreferencesRequest {
  email?: Partial<NotificationPreferences['email']>;
  inApp?: Partial<NotificationPreferences['inApp']>;
  push?: Partial<NotificationPreferences['push']>;
  webhook?: Partial<NotificationPreferences['webhook']>;
  categoryOverrides?: NotificationPreferences['categoryOverrides'];
}

/**
 * Notification list response
 */
export interface NotificationListResponse {
  notifications: Notification[];
  totalCount: number;
  unreadCount: number;
  hasMore: boolean;
  nextCursor?: string;
}

/**
 * Activity feed response
 */
export interface ActivityFeedResponse {
  activities: ActivityFeedItem[];
  hasMore: boolean;
  nextCursor?: string;
}

/**
 * Trust alerts response
 */
export interface TrustAlertsResponse {
  alerts: TrustAlert[];
  activeCount: number;
  acknowledgedCount: number;
  hasCritical: boolean;
}

// ============================================================================
// Utility Types
// ============================================================================

/**
 * Type guard for notification types
 */
export function isNotificationType(type: string): type is NotificationType {
  const validTypes: NotificationType[] = [
    'reputation_gained',
    'function_error_spike',
    'trust_change',
    'issue_assigned',
    'fxcert_verified',
    'bounty_claimed',
    'trust_drop',
    'replay_failed',
    'determinism_broken',
  ];
  return validTypes.includes(type as NotificationType);
}

/**
 * Type guard for notification categories
 */
export function isNotificationCategory(category: string): category is NotificationCategory {
  const validCategories: NotificationCategory[] = [
    'all',
    'trust',
    'revenue',
    'issues',
    'messages',
    'security',
  ];
  return validCategories.includes(category as NotificationCategory);
}

/**
 * Priority weight for sorting (higher = more important)
 */
export const PRIORITY_WEIGHTS: Record<NotificationPriority, number> = {
  low: 1,
  medium: 2,
  high: 3,
  critical: 4,
};

/**
 * Default notification preferences
 */
export const DEFAULT_NOTIFICATION_PREFERENCES: Omit<
  NotificationPreferences,
  'userId' | 'tenantId' | 'updatedAt'
> = {
  email: {
    enabled: true,
    digestFrequency: 'daily',
    categories: ['all', 'trust', 'security', 'revenue'],
    minimumPriority: 'medium',
  },
  inApp: {
    enabled: true,
    soundEnabled: true,
    desktopNotifications: true,
    categories: ['all', 'trust', 'revenue', 'issues', 'messages', 'security'],
    minimumPriority: 'low',
  },
  push: {
    enabled: false,
    categories: ['trust', 'security'],
    minimumPriority: 'high',
  },
};
