/**
 * NotificationsPage
 *
 * Full-page notifications view with TrustAlertBanner, NotificationCenter,
 * and RealtimeActivityFeed in a responsive layout.
 */

import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
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
import type { TrustAlert, Notification } from '@/types/notifications';

// ============================================================================
// Loading Skeleton Component
// ============================================================================

function NotificationsPageSkeleton() {
  return (
    <div className="space-y-6">
      {/* Header Skeleton */}
      <div className="space-y-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-4 w-96" />
      </div>

      {/* Trust Alert Skeleton */}
      <Skeleton className="h-24 w-full rounded-lg" />

      {/* Content Grid Skeleton */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-4">
          <Skeleton className="h-12 w-full rounded-lg" />
          <Skeleton className="h-96 w-full rounded-lg" />
        </div>

        {/* Sidebar */}
        <div className="space-y-4">
          <Skeleton className="h-8 w-32" />
          <Skeleton className="h-96 w-full rounded-lg" />
        </div>
      </div>
    </div>
  );
}

// ============================================================================
// Main Page Component
// ============================================================================

export function NotificationsPage() {
  const navigate = useNavigate();
  const user = useAuthStore((state) => state.user);
  const { updateUnreadCounts } = useNotificationStore();

  // Local state
  const [isLoading, setIsLoading] = useState(true);
  const [criticalAlert, setCriticalAlert] = useState<TrustAlert | undefined>();
  const [dismissedAlertIds, setDismissedAlertIds] = useState<string[]>([]);

  // Fetch initial data
  useEffect(() => {
    const fetchNotifications = async () => {
      if (!user?.id) return;

      try {
        setIsLoading(true);

        // Fetch unread counts
        const counts = await notificationsApi.fetchUnreadCounts();
        const byCategory = counts?.byCategory || {};
        updateUnreadCounts({
          all: counts?.total || 0,
          trust: byCategory.trust || 0,
          revenue: byCategory.revenue || 0,
          issues: byCategory.issues || 0,
          messages: byCategory.messages || 0,
          security: byCategory.security || 0,
        });

        // Fetch notifications to check for critical trust alerts in metadata
        const notifications = await notificationsApi.fetchNotifications({
          category: 'trust',
          limit: 10,
        });

        // Look for critical trust-related notifications
        const criticalNotification = notifications.find(
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
  }, [user?.id, updateUnreadCounts, dismissedAlertIds]);

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
        !dismissedAlertIds.includes(alert.id)
      ) {
        setCriticalAlert(alert);
      }
    },
  });

  // Handle alert dismissal
  const handleDismissAlert = (alertId: string) => {
    setDismissedAlertIds((prev) => [...prev, alertId]);
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
      <div className="min-h-screen p-4 lg:p-6">
        <NotificationsPageSkeleton />
      </div>
    );
  }

  return (
    <div className="min-h-screen p-4 lg:p-6">
      {/* Page Header */}
      <PageHeader
        title="Notifications"
        subtitle="Stay updated on your functions, trust scores, and system alerts"
        breadcrumbs={[
          { label: 'Home', path: '/dashboard' },
          { label: 'Notifications' },
        ]}
        actions={[
          {
            label: 'Back',
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
            className="mt-6"
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
      <div className="mt-6 grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Notification Center - Takes 2/3 width on desktop */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="lg:col-span-2"
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
          className="lg:col-span-1"
        >
          <div className="bg-bg-secondary/50 backdrop-blur-sm border border-white/12 rounded-lg overflow-hidden">
            <div className="flex items-center gap-3 px-4 py-3 border-b border-white/8">
              <Bell className="h-5 w-5 text-brand-400" />
              <h2 className="text-lg font-semibold text-text-primary">
                Activity Feed
              </h2>
            </div>
            <ScrollArea className="h-[calc(100vh-300px)]">
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
    <motion.div
      initial={false}
      animate="animate"
      exit="exit"
    >
      {children}
    </motion.div>
  );
}

export default NotificationsPage;
