/**
 * NotificationCenter Component
 *
 * Main notification hub with category tabs, real-time updates,
 * and comprehensive notification management.
 */

'use client';

import { isNotificationUnreadStatus, notificationsApi } from '@/api/notifications';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Skeleton } from '@/components/ui/skeleton';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { NewNotificationEvent } from '@/hooks/types';
import { useRealtimeSubscription } from '@/hooks/useRealtimeSubscription';
import { unreadPartialFromServerCount } from '@/lib/notification-unread-sync';
import { cn } from '@/lib/utils';
import { useNotificationStore } from '@/stores/notificationStore';
import {
  type Notification,
  type NotificationCategory,
  type NotificationPriority,
  type NotificationStatus,
} from '@/types/notifications';
import { AnimatePresence, motion } from 'framer-motion';
import {
  AlertTriangle,
  Bell,
  Check,
  CheckCheck,
  DollarSign,
  Filter,
  Inbox,
  Loader2,
  MessageSquare,
  Settings,
  Shield,
  WifiOff,
} from 'lucide-react';
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import i18next from 'i18next';
import { toast } from 'sonner';
import { NotificationCard } from './NotificationCard';

// ============================================================================
// Types & Interfaces
// ============================================================================

interface NotificationCenterProps {
  className?: string;
  maxHeight?: string;
  onNotificationClick?: (notification: Notification) => void;
  onSettingsClick?: () => void;
}

interface TabConfig {
  id: NotificationCategory;
  label: string;
  icon: React.ReactNode;
}

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Group notifications by date (Today, Yesterday, Earlier)
 */
function groupNotificationsByDate(notifications: Notification[]): {
  today: Notification[];
  yesterday: Notification[];
  earlier: Notification[];
} {
  const now = new Date();
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const yesterday = new Date(today);
  yesterday.setDate(yesterday.getDate() - 1);

  return notifications.reduce(
    (groups, notification) => {
      const notificationDate = new Date(notification.timestamp);
      const notificationDay = new Date(
        notificationDate.getFullYear(),
        notificationDate.getMonth(),
        notificationDate.getDate()
      );

      if (notificationDay.getTime() === today.getTime()) {
        groups.today.push(notification);
      } else if (notificationDay.getTime() === yesterday.getTime()) {
        groups.yesterday.push(notification);
      } else {
        groups.earlier.push(notification);
      }

      return groups;
    },
    { today: [], yesterday: [], earlier: [] } as {
      today: Notification[];
      yesterday: Notification[];
      earlier: Notification[];
    }
  );
}

// ============================================================================
// EmptyState Component
// ============================================================================

function EmptyState({ category }: { category: string }) {
  const { t } = useTranslation();

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      className="flex flex-col items-center justify-center py-12 px-4 text-center"
    >
      <div className="w-16 h-16 rounded-full bg-bg-tertiary border border-border-subtle flex items-center justify-center mb-4">
        <Inbox className="h-8 w-8 text-text-muted" />
      </div>
      <h3 className="text-lg font-medium text-text-primary mb-1">{t('notifCenter.emptyTitle')}</h3>
      <p className="text-sm text-text-muted max-w-xs">
        {category === 'all'
          ? t('notifCenter.emptyAll')
          : t('notifCenter.emptyCategory', { category })}
      </p>
    </motion.div>
  );
}

// ============================================================================
// ErrorState Component
// ============================================================================

function ErrorState({ onRetry }: { onRetry: () => void }) {
  const { t } = useTranslation();

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      className="flex flex-col items-center justify-center py-12 px-4 text-center"
    >
      <div className="w-16 h-16 rounded-full bg-error/10 border border-error/20 flex items-center justify-center mb-4">
        <AlertTriangle className="h-8 w-8 text-error" />
      </div>
      <h3 className="text-lg font-medium text-text-primary mb-1">{t('notifCenter.errorTitle')}</h3>
      <p className="text-sm text-text-muted max-w-xs mb-4">
        {t('notifCenter.errorDescription')}
      </p>
      <Button variant="outline" size="sm" onClick={onRetry}>
        {t('notifCenter.tryAgain')}
      </Button>
    </motion.div>
  );
}

// ============================================================================
// NotificationGroup Component
// ============================================================================

interface NotificationGroupProps {
  title: string;
  notifications: Notification[];
  onNotificationClick?: (notification: Notification) => void;
  onMarkAsRead?: (id: string) => void;
  onArchive?: (id: string) => void;
}

function NotificationGroup({
  title,
  notifications,
  onNotificationClick,
  onMarkAsRead,
  onArchive,
}: NotificationGroupProps) {
  if (notifications.length === 0) return null;

  return (
    <div className="mb-4">
      <h5 className="text-xs font-semibold text-text-muted uppercase tracking-wider px-4 py-2 sticky top-0 bg-bg-primary/95 backdrop-blur-sm z-10">
        {title}
      </h5>
      <div className="divide-y divide-border-subtle">
        <AnimatePresence mode="popLayout">
          {notifications.map((notification) => (
            <NotificationCard
              key={notification.id}
              notification={notification}
              onClick={onNotificationClick}
              onMarkAsRead={onMarkAsRead}
              onArchive={onArchive}
            />
          ))}
        </AnimatePresence>
      </div>
    </div>
  );
}

// ============================================================================
// LoadingSkeleton Component
// ============================================================================

function LoadingSkeleton() {
  return (
    <div className="p-4 space-y-4">
      {[...Array(5)].map((_, i) => (
        <div key={i} className="flex gap-3">
          <Skeleton className="h-10 w-10 rounded-full flex-shrink-0" />
          <div className="flex-1 space-y-2">
            <Skeleton className="h-4 w-3/4" />
            <Skeleton className="h-3 w-full" />
            <Skeleton className="h-3 w-1/4" />
          </div>
        </div>
      ))}
    </div>
  );
}

// ============================================================================
// Custom Hook for Notifications with Real-time Updates
// ============================================================================

interface NotificationEvent extends NewNotificationEvent {
  id: string;
  notification_type: 'info' | 'warning' | 'error' | 'success';
  title: string;
  message: string;
  priority: NotificationPriority;
  category: NotificationCategory;
  action_url?: string;
}

function useNotificationCenter() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isMarkingAllRead, setIsMarkingAllRead] = useState(false);

  // Fetch initial notifications
  const fetchNotifications = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const data = await notificationsApi.fetchNotifications({ limit: 50 });
      setNotifications(data);
      const unread = data.filter((n) => isNotificationUnreadStatus(n.status)).length;
      setUnreadCount(unread);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load notifications');
      console.error('Error fetching notifications:', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Initial fetch
  useEffect(() => {
    fetchNotifications();
  }, [fetchNotifications]);

  // Subscribe to real-time notifications
  const { isConnected } = useRealtimeSubscription<NotificationEvent>(
    'user_notifications',
    'new_notification',
    (event) => {
      // Create a new notification from the event
      // Note: The event structure may vary based on backend implementation
      const newNotification: Notification = {
        id: event.record_id || event.notification_id || String(Date.now()),
        type: (event.data?.type || event.notification_type || 'info') as Notification['type'],
        category: event.data?.category || 'all',
        title: event.data?.title || event.title || i18next.t('notifCenter.toastNewNotification'),
        message: event.data?.message || event.message || '',
        timestamp: event.timestamp || new Date().toISOString(),
        priority: event.data?.priority || 'medium',
        status: 'unread',
        metadata: event.data?.metadata || {},
        userId: event.user_id || '',
        tenantId: event.tenant_id || '',
        actionUrl: event.data?.action_url || event.action_url,
      };

      setNotifications((prev) => [newNotification, ...prev]);
      setUnreadCount((prev) => prev + 1);

      // Show toast for new notification
      toast.info(newNotification.title, {
        description: newNotification.message,
        action: newNotification.actionUrl
          ? {
              label: i18next.t('notifCenter.viewAction'),
              onClick: () => {
                window.location.href = newNotification.actionUrl!;
              },
            }
          : undefined,
      });
    }
  );

  const syncBellUnreadFromServer = useCallback(async () => {
    try {
      const counts = await notificationsApi.fetchUnreadCounts();
      useNotificationStore.getState().updateUnreadCounts(unreadPartialFromServerCount(counts));
    } catch {
      /* bell stays stale; avoid toast spam */
    }
  }, []);

  const markAsRead = useCallback(
    async (id: string) => {
      try {
        await notificationsApi.markNotificationAsRead(id);
        setNotifications((prev) =>
          prev.map((n) =>
            n.id === id
              ? { ...n, status: 'read' as NotificationStatus, readAt: new Date().toISOString() }
              : n
          )
        );
        setUnreadCount((prev) => Math.max(0, prev - 1));
        void syncBellUnreadFromServer();
      } catch (err) {
        console.error('Error marking notification as read:', err);
        toast.error(i18next.t('notifCenter.toastMarkAsReadFailed'));
      }
    },
    [syncBellUnreadFromServer]
  );

  const markAllAsRead = useCallback(async () => {
    if (unreadCount === 0) return;

    setIsMarkingAllRead(true);
    try {
      const count = await notificationsApi.markAllNotificationsAsRead();
      setNotifications((prev) =>
        prev.map((n) =>
          isNotificationUnreadStatus(n.status)
            ? { ...n, status: 'read' as NotificationStatus, readAt: new Date().toISOString() }
            : n
        )
      );
      setUnreadCount(0);
      void syncBellUnreadFromServer();
      toast.success(i18next.t('notifCenter.toastMarkAllSuccess', { count }));
    } catch (err) {
      console.error('Error marking all notifications as read:', err);
      toast.error(i18next.t('notifCenter.toastMarkAllFailed'));
      throw err;
    } finally {
      setIsMarkingAllRead(false);
    }
  }, [unreadCount, syncBellUnreadFromServer]);

  const archiveNotification = useCallback(
    async (id: string) => {
      try {
        await notificationsApi.archiveNotification(id);
        let wasUnread = false;
        setNotifications((prev) => {
          const cur = prev.find((n) => n.id === id);
          wasUnread = !!(cur && isNotificationUnreadStatus(cur.status));
          return prev.map((n) =>
            n.id === id
              ? { ...n, status: 'archived' as NotificationStatus, readAt: new Date().toISOString() }
              : n
          );
        });
        if (wasUnread) {
          setUnreadCount((prev) => Math.max(0, prev - 1));
        }
        void syncBellUnreadFromServer();
        toast.success(i18next.t('notifCenter.toastArchived'));
      } catch (err) {
        console.error('Error archiving notification:', err);
        toast.error(i18next.t('notifCenter.toastArchiveFailed'));
      }
    },
    [syncBellUnreadFromServer]
  );

  return {
    notifications,
    unreadCount,
    isLoading,
    error,
    isConnected,
    isMarkingAllRead,
    markAsRead,
    markAllAsRead,
    archiveNotification,
    refresh: fetchNotifications,
  };
}

// ============================================================================
// Main NotificationCenter Component
// ============================================================================

export function NotificationCenter({
  className,
  maxHeight = '600px',
  onNotificationClick,
  onSettingsClick,
}: NotificationCenterProps) {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<NotificationCategory>('all');
  const [priorityFilter, setPriorityFilter] = useState<NotificationPriority | 'all'>('all');

  const TABS: TabConfig[] = useMemo(() => [
    { id: 'all', label: t('notifCenter.tabAll'), icon: <Inbox className="h-4 w-4" /> },
    { id: 'trust', label: t('notifCenter.tabTrust'), icon: <Shield className="h-4 w-4" /> },
    { id: 'revenue', label: t('notifCenter.tabRevenue'), icon: <DollarSign className="h-4 w-4" /> },
    { id: 'issues', label: t('notifCenter.tabIssues'), icon: <AlertTriangle className="h-4 w-4" /> },
    { id: 'messages', label: t('notifCenter.tabMessages'), icon: <MessageSquare className="h-4 w-4" /> },
    { id: 'security', label: t('notifCenter.tabSecurity'), icon: <Shield className="h-4 w-4" /> },
  ], [t]);

  const PRIORITY_OPTIONS: { value: NotificationPriority | 'all'; label: string }[] = useMemo(() => [
    { value: 'all', label: t('notifCenter.priorityAll') },
    { value: 'critical', label: t('notifCenter.priorityCritical') },
    { value: 'high', label: t('notifCenter.priorityHigh') },
    { value: 'medium', label: t('notifCenter.priorityMedium') },
    { value: 'low', label: t('notifCenter.priorityLow') },
  ], [t]);

  const {
    notifications,
    unreadCount,
    isLoading,
    error,
    isConnected,
    isMarkingAllRead,
    markAsRead,
    markAllAsRead,
    archiveNotification,
    refresh,
  } = useNotificationCenter();

  // Filter notifications based on active tab and priority
  const filteredNotifications = useMemo(() => {
    return notifications.filter((notification) => {
      const categoryMatch = activeTab === 'all' || notification.category === activeTab;
      const priorityMatch = priorityFilter === 'all' || notification.priority === priorityFilter;
      return categoryMatch && priorityMatch;
    });
  }, [notifications, activeTab, priorityFilter]);

  // Group notifications by date
  const groupedNotifications = useMemo(() => {
    return groupNotificationsByDate(filteredNotifications);
  }, [filteredNotifications]);

  // Calculate unread counts per category
  const unreadCounts = useMemo(() => {
    const counts: Record<string, number> = { all: 0 };
    TABS.forEach((tab) => {
      counts[tab.id] = notifications.filter(
        (n) => isNotificationUnreadStatus(n.status) && (tab.id === 'all' || n.category === tab.id)
      ).length;
    });
    counts.all = notifications.filter((n) => isNotificationUnreadStatus(n.status)).length;
    return counts;
  }, [notifications]);

  // Handle mark all as read
  const handleMarkAllAsRead = useCallback(async () => {
    try {
      await markAllAsRead();
    } catch {
      // Error is already handled in the hook
    }
  }, [markAllAsRead]);

  return (
    <TooltipProvider>
      <div
        className={cn(
          'flex flex-col bg-bg-primary border border-border-default rounded-xl overflow-hidden shadow-xl',
          className
        )}
      >
        {/* Glassmorphism Header */}
        <div className="relative bg-bg-secondary/80 backdrop-blur-md border-b border-border-subtle">
          {/* Header Actions */}
          <div className="flex items-center justify-between px-4 py-3">
            <div className="flex items-center gap-2">
              <div className="relative">
                <Bell className="h-5 w-5 text-text-primary" />
                {unreadCount > 0 && (
                  <span className="absolute -top-1 -right-1 w-2.5 h-2.5 bg-brand-500 rounded-full animate-pulse" />
                )}
              </div>
              <h2 className="text-lg font-semibold text-text-primary">{t('notifCenter.title')}</h2>
              {unreadCount > 0 && (
                <Badge variant="secondary" className="text-xs">
                  {t('notifCenter.unreadBadge', { count: unreadCount })}
                </Badge>
              )}
            </div>

            <div className="flex items-center gap-1">
              {/* Filter Dropdown */}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8"
                    aria-label={t('notifCenter.filterByPriority')}
                  >
                    <Filter className="h-4 w-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-48">
                  <DropdownMenuLabel>{t('notifCenter.filterByPriority')}</DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  {PRIORITY_OPTIONS.map((option) => (
                    <DropdownMenuItem
                      key={option.value}
                      onClick={() => setPriorityFilter(option.value)}
                      className={cn(
                        'flex items-center justify-between',
                        priorityFilter === option.value && 'bg-bg-secondary'
                      )}
                    >
                      {option.label}
                      {priorityFilter === option.value && <Check className="h-4 w-4" />}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>

              {/* Mark All as Read */}
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8"
                    onClick={handleMarkAllAsRead}
                    disabled={unreadCount === 0 || isMarkingAllRead}
                    aria-label={t('notifCenter.markAllAsRead')}
                  >
                    {isMarkingAllRead ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <CheckCheck className="h-4 w-4" />
                    )}
                  </Button>
                </TooltipTrigger>
                <TooltipContent>{t('notifCenter.markAllAsRead')}</TooltipContent>
              </Tooltip>

              {/* Settings */}
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8"
                    onClick={onSettingsClick}
                    aria-label={t('notifCenter.settings')}
                  >
                    <Settings className="h-4 w-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>{t('notifCenter.settings')}</TooltipContent>
              </Tooltip>
            </div>
          </div>

          {/* Connection Status */}
          {!isConnected && (
            <div className="px-4 pb-2">
              <div className="flex items-center gap-2 text-xs text-warning bg-warning-glow/20 px-3 py-1.5 rounded-md">
                <WifiOff className="h-3 w-3" />
                <span>{t('notifCenter.reconnecting')}</span>
              </div>
            </div>
          )}

          {/* Category Tabs */}
          <Tabs
            value={activeTab}
            onValueChange={(value) => setActiveTab(value as NotificationCategory)}
            className="w-full"
          >
            <TabsList className="w-full justify-start rounded-none bg-transparent border-t border-border-subtle p-0 h-11">
              {TABS.map((tab) => (
                <TabsTrigger
                  key={tab.id}
                  value={tab.id}
                  className={cn(
                    'flex-1 rounded-none border-b-2 border-transparent data-[state=active]:border-brand-500 data-[state=active]:bg-transparent data-[state=active]:shadow-none py-3 px-2 gap-1.5 text-xs font-medium text-text-muted data-[state=active]:text-text-primary transition-colors'
                  )}
                >
                  <span className="hidden sm:inline">{tab.label}</span>
                  <span className="sm:hidden">{tab.icon}</span>
                  {unreadCounts[tab.id] > 0 && (
                    <Badge
                      variant="outline"
                      className="h-5 min-w-5 px-1 text-[10px] bg-brand-500/10 text-brand-500 border-brand-500/20"
                    >
                      {unreadCounts[tab.id] > 99 ? '99+' : unreadCounts[tab.id]}
                    </Badge>
                  )}
                </TabsTrigger>
              ))}
            </TabsList>
          </Tabs>
        </div>

        {/* Notification List */}
        <ScrollArea className="flex-1" style={{ maxHeight }}>
          <AnimatePresence mode="wait">
            {isLoading ? (
              <LoadingSkeleton />
            ) : error ? (
              <ErrorState onRetry={refresh} />
            ) : filteredNotifications.length === 0 ? (
              <EmptyState category={activeTab} />
            ) : (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 0.2 }}
              >
                <NotificationGroup
                  title={t('notifCenter.groupToday')}
                  notifications={groupedNotifications.today}
                  onNotificationClick={onNotificationClick}
                  onMarkAsRead={markAsRead}
                  onArchive={archiveNotification}
                />
                <NotificationGroup
                  title={t('notifCenter.groupYesterday')}
                  notifications={groupedNotifications.yesterday}
                  onNotificationClick={onNotificationClick}
                  onMarkAsRead={markAsRead}
                  onArchive={archiveNotification}
                />
                <NotificationGroup
                  title={t('notifCenter.groupEarlier')}
                  notifications={groupedNotifications.earlier}
                  onNotificationClick={onNotificationClick}
                  onMarkAsRead={markAsRead}
                  onArchive={archiveNotification}
                />
              </motion.div>
            )}
          </AnimatePresence>
        </ScrollArea>
      </div>
    </TooltipProvider>
  );
}

export default NotificationCenter;
