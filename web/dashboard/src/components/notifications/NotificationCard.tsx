/**
 * NotificationCard Component
 *
 * Displays individual notifications with type-specific styling,
 * icons, and interactive features including context menu and animations.
 */

'use client';

import { isNotificationUnreadStatus } from '@/api/notifications';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { cn } from '@/lib/utils';
import {
  type Notification,
  type NotificationPriority,
  type NotificationType,
} from '@/types/notifications';
import { formatDistanceToNow } from 'date-fns';
import { motion } from 'framer-motion';
import {
  AlertTriangle,
  Archive,
  BadgeCheck,
  Bell,
  Check,
  Coins,
  GitBranch,
  MailOpen,
  MoreVertical,
  Shield,
  TrendingUp,
} from 'lucide-react';
import React, { useCallback, useState } from 'react';

// ============================================================================
// Props Interface
// ============================================================================

export interface NotificationCardProps {
  notification: Notification;
  onClick?: (notification: Notification) => void;
  onMarkAsRead?: (id: string) => void;
  onArchive?: (id: string) => void;
  className?: string;
  compact?: boolean;
}

// ============================================================================
// Type Definitions for Notification Config
// ============================================================================

interface NotificationTypeConfig {
  icon: React.ReactNode;
  accentColor: string;
  bgGradient: string;
  label: string;
}

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Format timestamp to relative time (e.g., "2 minutes ago") using date-fns for i18n-ready formatting
 */
function formatRelativeTime(timestamp: string): string {
  const date = new Date(timestamp);
  if (Date.now() - date.getTime() < 60_000) return 'Just now';
  return formatDistanceToNow(date, { addSuffix: true });
}

/**
 * Get priority color for badge and border
 */
function getPriorityColor(priority: NotificationPriority): string {
  switch (priority) {
    case 'critical':
      return 'bg-red-500 text-white border-red-500';
    case 'high':
      return 'bg-orange-500 text-white border-orange-500';
    case 'medium':
      return 'bg-blue-500 text-white border-blue-500';
    case 'low':
      return 'bg-green-500 text-white border-green-500';
    default:
      return 'bg-gray-500 text-white border-gray-500';
  }
}

/**
 * Get priority border color for subtle indicator
 */
function getPriorityBorderColor(priority: NotificationPriority): string {
  switch (priority) {
    case 'critical':
      return 'border-l-red-500';
    case 'high':
      return 'border-l-orange-500';
    case 'medium':
      return 'border-l-blue-500';
    case 'low':
      return 'border-l-green-500';
    default:
      return 'border-l-transparent';
  }
}

/**
 * Get notification type configuration
 */
function getNotificationTypeConfig(
  type: NotificationType,
  notification: Notification
): NotificationTypeConfig {
  switch (type) {
    case 'reputation_gained': {
      const amount = notification.metadata.reputation?.amount ?? 0;
      return {
        icon: <TrendingUp className="h-5 w-5 text-emerald-500" />,
        accentColor: 'text-emerald-500',
        bgGradient:
          'group-hover:bg-gradient-to-r group-hover:from-emerald-500/5 group-hover:to-transparent',
        label: `+${amount} reputation`,
      };
    }

    case 'function_error_spike': {
      return {
        icon: <AlertTriangle className="h-5 w-5 text-red-500" />,
        accentColor: 'text-red-500',
        bgGradient:
          'group-hover:bg-gradient-to-r group-hover:from-red-500/5 group-hover:to-transparent',
        label: 'Error Spike',
      };
    }

    case 'trust_change': {
      const delta = notification.metadata.trust?.trustDelta ?? 0;
      const isPositive = delta >= 0;
      return {
        icon: <Shield className={cn('h-5 w-5', isPositive ? 'text-blue-500' : 'text-amber-500')} />,
        accentColor: isPositive ? 'text-blue-500' : 'text-amber-500',
        bgGradient: isPositive
          ? 'group-hover:bg-gradient-to-r group-hover:from-blue-500/5 group-hover:to-transparent'
          : 'group-hover:bg-gradient-to-r group-hover:from-amber-500/5 group-hover:to-transparent',
        label: `${isPositive ? '+' : ''}${delta} trust`,
      };
    }

    case 'issue_assigned': {
      return {
        icon: <GitBranch className="h-5 w-5 text-violet-500" />,
        accentColor: 'text-violet-500',
        bgGradient:
          'group-hover:bg-gradient-to-r group-hover:from-violet-500/5 group-hover:to-transparent',
        label: 'Issue Assigned',
      };
    }

    case 'fxcert_verified': {
      return {
        icon: <BadgeCheck className="h-5 w-5 text-indigo-500" />,
        accentColor: 'text-indigo-500',
        bgGradient:
          'group-hover:bg-gradient-to-r group-hover:from-indigo-500/5 group-hover:to-transparent',
        label: 'Certification Verified',
      };
    }

    case 'bounty_claimed': {
      return {
        icon: <Coins className="h-5 w-5 text-yellow-500" />,
        accentColor: 'text-yellow-500',
        bgGradient:
          'group-hover:bg-gradient-to-r group-hover:from-yellow-500/5 group-hover:to-transparent',
        label: 'Bounty Claimed',
      };
    }

    default:
      return {
        icon: <Bell className="h-5 w-5 text-gray-500" />,
        accentColor: 'text-gray-500',
        bgGradient:
          'group-hover:bg-gradient-to-r group-hover:from-gray-500/5 group-hover:to-transparent',
        label: 'Notification',
      };
  }
}

/**
 * Get type-specific content details
 */
function getTypeDetails(
  notification: Notification
): { primary: string; secondary?: string } | null {
  switch (notification.type) {
    case 'reputation_gained': {
      const meta = notification.metadata.reputation;
      if (!meta) return null;
      return {
        primary: `+${meta.amount} reputation points`,
        secondary: meta.source,
      };
    }

    case 'function_error_spike': {
      const meta = notification.metadata.errorSpike;
      if (!meta) return null;
      return {
        primary: `${meta.errorCount} errors detected`,
        secondary: meta.functionName,
      };
    }

    case 'trust_change': {
      const meta = notification.metadata.trust;
      if (!meta) return null;
      const isPositive = meta.trustDelta >= 0;
      return {
        primary: `${isPositive ? '+' : ''}${meta.trustDelta} trust score`,
        secondary: meta.reason,
      };
    }

    case 'issue_assigned': {
      const meta = notification.metadata.issue;
      if (!meta) return null;
      return {
        primary: `${meta.issueId}: ${meta.issueTitle}`,
        secondary: `Priority: ${meta.severity ?? 'normal'}`,
      };
    }

    case 'fxcert_verified': {
      const meta = notification.metadata.fxCert;
      if (!meta) return null;
      return {
        primary: meta.certType,
        secondary: meta.verifierName ? `Verified by ${meta.verifierName}` : undefined,
      };
    }

    case 'bounty_claimed': {
      const meta = notification.metadata.bounty;
      if (!meta) return null;
      return {
        primary: `${meta.bountyAmount} ${meta.currency} earned`,
        secondary: meta.functionName,
      };
    }

    default:
      return null;
  }
}

// ============================================================================
// Component
// ============================================================================

export const NotificationCard = React.memo(function NotificationCard({
  notification,
  onClick,
  onMarkAsRead,
  onArchive,
  className,
  compact = false,
}: NotificationCardProps) {
  const [isContextMenuOpen, setIsContextMenuOpen] = useState(false);
  const isUnread = isNotificationUnreadStatus(notification.status);
  const typeConfig = getNotificationTypeConfig(notification.type, notification);
  const typeDetails = getTypeDetails(notification);

  const handleClick = useCallback(() => {
    if (isUnread && onMarkAsRead) {
      onMarkAsRead(notification.id);
    }
    onClick?.(notification);
  }, [notification, isUnread, onClick, onMarkAsRead]);

  const handleMarkAsRead = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation();
      onMarkAsRead?.(notification.id);
    },
    [notification.id, onMarkAsRead]
  );

  const handleArchive = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation();
      onArchive?.(notification.id);
    },
    [notification.id, onArchive]
  );

  const handleContextMenu = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    setIsContextMenuOpen(true);
  }, []);

  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: isUnread ? 1 : 0.7, y: 0 }}
      exit={{ opacity: 0, x: -100 }}
      whileHover={{ scale: 1.02 }}
      transition={{
        duration: 0.3,
        ease: [0.25, 0.46, 0.45, 0.94],
        scale: { duration: 0.2 },
      }}
      className={cn(
        'group relative flex gap-3 cursor-pointer',
        'border-b border-border-subtle last:border-b-0',
        'transition-all duration-300 ease-out',
        'hover:shadow-md hover:shadow-black/5',
        typeConfig.bgGradient,
        getPriorityBorderColor(notification.priority),
        'border-l-2',
        isUnread ? 'bg-brand-500/5' : 'bg-transparent',
        compact ? 'p-3' : 'p-4',
        className
      )}
      onClick={handleClick}
      onContextMenu={handleContextMenu}
    >
      {/* Unread Indicator Dot */}
      {isUnread && (
        <motion.div
          initial={{ scale: 0 }}
          animate={{ scale: 1 }}
          className="absolute left-2 top-1/2 -translate-y-1/2 w-2 h-2 bg-blue-500 rounded-full"
        />
      )}

      {/* Icon Container */}
      <div
        className={cn(
          'flex-shrink-0 rounded-full flex items-center justify-center',
          'bg-bg-tertiary border border-border-subtle',
          'transition-transform duration-200 group-hover:scale-110',
          compact ? 'w-9 h-9' : 'w-10 h-10'
        )}
      >
        {typeConfig.icon}
      </div>

      {/* Content */}
      <div className="flex-1 min-w-0">
        {/* Header Row */}
        <div className="flex items-start justify-between gap-2">
          <div className="flex items-center gap-2 min-w-0">
            <h4
              className={cn(
                'text-sm line-clamp-1',
                isUnread ? 'font-semibold text-text-primary' : 'font-medium text-text-primary'
              )}
            >
              {notification.title}
            </h4>
            <Badge
              variant="outline"
              className={cn(
                'text-[10px] px-1.5 py-0 flex-shrink-0',
                getPriorityColor(notification.priority)
              )}
            >
              {typeConfig.label}
            </Badge>
          </div>

          {/* Time and Actions */}
          <div className="flex items-center gap-1 flex-shrink-0">
            <span className="text-xs text-text-muted">
              {formatRelativeTime(notification.timestamp)}
            </span>

            {/* Context Menu */}
            <DropdownMenu open={isContextMenuOpen} onOpenChange={setIsContextMenuOpen}>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7 opacity-0 group-hover:opacity-100 transition-opacity"
                  onClick={(e) => e.stopPropagation()}
                >
                  <MoreVertical className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" onClick={(e) => e.stopPropagation()}>
                {isUnread && (
                  <DropdownMenuItem onClick={handleMarkAsRead}>
                    <MailOpen className="h-4 w-4 mr-2" />
                    Mark as read
                  </DropdownMenuItem>
                )}
                <DropdownMenuItem onClick={handleArchive}>
                  <Archive className="h-4 w-4 mr-2" />
                  Archive
                </DropdownMenuItem>
                {isUnread && (
                  <>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem onClick={handleMarkAsRead}>
                      <Check className="h-4 w-4 mr-2" />
                      Mark as read & close
                    </DropdownMenuItem>
                  </>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

        {/* Message */}
        <p className="text-sm text-text-secondary line-clamp-2 mt-1">{notification.message}</p>

        {/* Type-specific Details */}
        {typeDetails && !compact && (
          <div className="flex items-center gap-2 mt-2">
            <span className={cn('text-xs font-medium', typeConfig.accentColor)}>
              {typeDetails.primary}
            </span>
            {typeDetails.secondary && (
              <>
                <span className="text-xs text-text-muted">•</span>
                <span className="text-xs text-text-muted">{typeDetails.secondary}</span>
              </>
            )}
          </div>
        )}

        {/* Action Link */}
        {notification.actionUrl && (
          <div className="mt-2">
            <span className={cn('text-xs hover:underline', typeConfig.accentColor)}>
              View details →
            </span>
          </div>
        )}
      </div>
    </motion.div>
  );
});

export default NotificationCard;
