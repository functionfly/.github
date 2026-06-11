/**
 * NotificationsPage
 *
 * Full-page notifications view with TrustAlertBanner, NotificationCenter,
 * and RealtimeActivityFeed in a responsive layout.
 */

import './styles.css';

import { useEffect, useState, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { usePageTitle } from '@/hooks';
import { motion } from 'framer-motion';
import { ArrowLeft, Bell } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Skeleton } from '@/components/ui/skeleton';
import { PageHeader } from '@/components/layout/PageHeader';
import { NotificationCenter } from '@/components/notifications';
import { RealtimeActivityFeed } from '@/components/notifications';
import { TrustAlertBanner } from '@/components/notifications';
import { useNotificationRealtime } from '@/hooks/useNotificationRealtime';
import { useNotificationStore } from '@/stores/notificationStore';
import { useAuthStore } from '@/stores/authStore';
import { notificationsApi } from '@/api/notifications';
import { unreadPartialFromServerCount } from '@/lib/notification-unread-sync';
import type { TrustAlert, Notification } from '@/types/notifications';

// ============================================================================
// Loading Skeleton Component
// ============================================================================

function NotificationsPageSkeleton() {
  return (
    <div className="notifications-skeleton">
      {/* Header Skeleton */}
      <div className="notifications-skeleton-header">
        <Skeleton className="h-8 w-48 notifications-skeleton-shimmer" />
        <Skeleton className="h-4 w-96 notifications-skeleton-shimmer" />
      </div>

      {/* Trust Alert Skeleton */}
      <Skeleton className="notifications-skeleton-banner notifications-skeleton-shimmer" />

      {/* Content Grid Skeleton */}
      <div className="notifications-skeleton-grid">
        {/* Main Content */}
        <div className="notifications-skeleton-main">
          <Skeleton className="notifications-skeleton-tabs notifications-skeleton-shimmer" />
          <Skeleton className="notifications-skeleton-list notifications-skeleton-shimmer" />
        </div>

        {/* Sidebar */}
        <div className="notifications-skeleton-sidebar">
          <Skeleton className="notifications-skeleton-sidebar-header notifications-skeleton-shimmer" />
          <Skeleton className="notifications-skeleton-sidebar-content notifications-skeleton-shimmer" />
        </div>
      </div>
    </div>
  );
}

// ============================================================================
// Main Page Component
// ============================================================================

export function NotificationsPage() {
  usePageTitle('Notifications');
  const { t } = useTranslation();
  const navigate = useNavigate();
  const user = useAuthStore((state) => state.user);
  const { updateUnreadCounts } = useNotificationStore();

  // Local state
  const [isLoading, setIsLoading] = useState(true);
  const [criticalAlert, setCriticalAlert] = useState<TrustAlert | undefined>();
  const [dismissedAlertIds, setDismissedAlertIds] = useState<string[]>([]);
  const dismissedAlertIdsRef = useRef<string[]>([]);

  // Fetch initial data
  useEffect(() => {
    const fetchNotifications = async () => {
      if (!user?.id) return;

      try {
        setIsLoading(true);

        // Fetch unread counts
        const counts = await notificationsApi.fetchUnreadCounts();
        updateUnreadCounts(unreadPartialFromServerCount(counts));

        // Fetch notifications to check for critical trust alerts in metadata
        const result = await notificationsApi.fetchNotifications({
          category: 'trust',
          limit: 10,
        });

        // Look for critical trust-related notifications
        const criticalNotification = result.notifications.find(
          (notification) =>
            notification.priority === 'critical' &&
            (notification.type === 'trust_drop' ||
              notification.type === 'replay_failed' ||
              notification.type === 'determinism_broken')
        );

        if (criticalNotification?.metadata?.trustAlert) {
          setCriticalAlert(criticalNotification.metadata.trustAlert as TrustAlert);
        }
      } catch (error) {
        console.error('Failed to fetch notifications:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchNotifications();
  }, [user?.id, updateUnreadCounts]);

  // Real-time notification handling
  useNotificationRealtime({
    enabled: !!user?.id,
    onNewNotification: (notification: Notification) => {
      // Update unread count for the category
      const { unreadCounts, updateUnreadCount } = useNotificationStore.getState();
      const category = notification.category === 'all' ? 'all' : notification.category;
      updateUnreadCount(category, unreadCounts[category] + 1);
    },
    onTrustAlert: (alert: TrustAlert) => {
      // Show critical alerts immediately
      if (
        (alert.severity === 'critical' || alert.severity === 'emergency') &&
        !alert.acknowledged &&
        !dismissedAlertIdsRef.current.includes(alert.id)
      ) {
        setCriticalAlert(alert);
      }
    },
  });

  // Handle alert dismissal
  const handleDismissAlert = (alertId: string) => {
    const newDismissed = [...dismissedAlertIdsRef.current, alertId];
    dismissedAlertIdsRef.current = newDismissed;
    setDismissedAlertIds(newDismissed);
    setCriticalAlert(undefined);

    // Persist dismissal in store
    const { dismissAlert } = useNotificationStore.getState();
    dismissAlert(alertId);
  };

  // Handle notification click - navigate to relevant page
  const handleNotificationClick = (notification: Notification) => {
    if (notification.actionUrl) {
      navigate(notification.actionUrl);
    }
  };

  // Handle back button
  const handleBack = () => {
    navigate(-1);
  };

  if (isLoading) {
    return (
      <div className="notifications-page">
        <NotificationsPageSkeleton />
      </div>
    );
  }

  return (
    <div className="notifications-page">
      {/* Page Header */}
      <PageHeader
        title={t('notificationsPage.title')}
        subtitle={t('notificationsPage.subtitle')}
        breadcrumbs={[
          { label: t('notificationsPage.home'), path: '/dashboard' },
          { label: t('notificationsPage.breadcrumbNotifications') },
        ]}
        actions={[
          {
            label: t('notificationsPage.back'),
            onClick: handleBack,
            variant: 'outline',
            icon: ArrowLeft,
          },
        ]}
      />

      {/* Trust Alert Banner (if critical alerts exist) */}
      <AnimatePresence>
        {criticalAlert && (
          <motion.div
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            className="trust-alert-container"
          >
            <TrustAlertBanner
              alert={criticalAlert}
              onDismiss={handleDismissAlert}
              onAction={(action, alert) => {
                if (action === 'view_details' && alert.actionUrl) {
                  navigate(alert.actionUrl);
                }
              }}
            />
          </motion.div>
        )}
      </AnimatePresence>

      {/* Main Content Grid */}
      <div className="notifications-content-grid">
        {/* Notification Center - Takes 2/3 width on desktop */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="lg:col-span-2 notification-center-panel"
        >
          <NotificationCenter
            maxHeight="calc(100vh - 300px)"
            onNotificationClick={handleNotificationClick}
          />
        </motion.div>

        {/* Realtime Activity Feed Sidebar - Takes 1/3 width on desktop */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
          className="lg:col-span-1 activity-feed-sidebar"
        >
          <div className="bg-bg-secondary/50 backdrop-blur-sm border border-white/12 rounded-lg overflow-hidden activity-feed-container">
            <div className="activity-feed-header">
              <Bell className="h-5 w-5 activity-feed-icon" />
              <h2 className="text-lg font-semibold text-text-primary activity-feed-title">
                {t('notificationsPage.activityFeed')}
              </h2>
            </div>
            <ScrollArea className="activity-feed-scroll">
              <RealtimeActivityFeed
                compact
                showHeader={false}
                maxItems={50}
                onActivityClick={(activity) => {
                  if (activity.target?.url) {
                    navigate(activity.target.url);
                  }
                }}
              />
            </ScrollArea>
          </div>
        </motion.div>
      </div>
    </div>
  );
}

// ============================================================================
// Animations
// ============================================================================

function AnimatePresence({ children }: { children: React.ReactNode }) {
  return (
    <motion.div initial={false} animate="animate" exit="exit">
      {children}
    </motion.div>
  );
}

export default NotificationsPage;
