/**
 * NotificationsPage — Sealed Containment
 */

import './styles.css';

import { useEffect, useState, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { usePageTitle } from '@/hooks';
import { Bell, ArrowLeft } from 'lucide-react';
import { NotificationCenter } from '@/components/notifications';
import { RealtimeActivityFeed } from '@/components/notifications';
import { TrustAlertBanner } from '@/components/notifications';
import { useNotificationRealtime } from '@/hooks/useNotificationRealtime';
import { useNotificationStore } from '@/stores/notificationStore';
import { useAuthStore } from '@/stores/authStore';
import { notificationsApi } from '@/api/notifications';
import { unreadPartialFromServerCount } from '@/lib/notification-unread-sync';
import type { TrustAlert, Notification } from '@/types/notifications';
import {
  PageGrid, Chamber, CornerBrace, TrustSeal,
  SealedButton, FrameButton, StatusPill, AnnotationTag,
} from '@/components/containment';

function NotificationsPageSkeleton() {
  return (
    <div className="notif-skeleton">
      <div className="notif-skeleton__header">
        <div className="notif-skeleton__bar notif-skeleton__bar--title" />
        <div className="notif-skeleton__bar notif-skeleton__bar--subtitle" />
      </div>
      <div className="notif-skeleton__banner" />
      <div className="notif-skeleton__grid">
        <div className="notif-skeleton__main">
          <div className="notif-skeleton__tabs" />
          <div className="notif-skeleton__list" />
        </div>
        <div className="notif-skeleton__sidebar">
          <div className="notif-skeleton__sidebar-header" />
          <div className="notif-skeleton__sidebar-content" />
        </div>
      </div>
    </div>
  );
}

export function NotificationsPage() {
  usePageTitle('Notifications');
  const { t } = useTranslation();
  const navigate = useNavigate();
  const user = useAuthStore((state) => state.user);
  const { updateUnreadCounts } = useNotificationStore();

  const [isLoading, setIsLoading] = useState(true);
  const [criticalAlert, setCriticalAlert] = useState<TrustAlert | undefined>();
  const [dismissedAlertIds, setDismissedAlertIds] = useState<string[]>([]);
  const dismissedAlertIdsRef = useRef<string[]>([]);

  useEffect(() => {
    const fetchNotifications = async () => {
      if (!user?.id) return;
      try {
        setIsLoading(true);
        const counts = await notificationsApi.fetchUnreadCounts();
        updateUnreadCounts(unreadPartialFromServerCount(counts));
        const result = await notificationsApi.fetchNotifications({ category: 'trust', limit: 10 });
        const criticalNotification = result.notifications.find(
          (n) => n.priority === 'critical' && (n.type === 'trust_drop' || n.type === 'replay_failed' || n.type === 'determinism_broken')
        );
        if (criticalNotification?.metadata?.trustAlert) {
          setCriticalAlert(criticalNotification.metadata.trustAlert as TrustAlert);
        }
      } catch (error) { console.error('Failed to fetch notifications:', error); }
      finally { setIsLoading(false); }
    };
    fetchNotifications();
  }, [user?.id, updateUnreadCounts]);

  useNotificationRealtime({
    enabled: !!user?.id,
    onNewNotification: (notification: Notification) => {
      const { unreadCounts, updateUnreadCount } = useNotificationStore.getState();
      const category = notification.category === 'all' ? 'all' : notification.category;
      updateUnreadCount(category, unreadCounts[category] + 1);
    },
    onTrustAlert: (alert: TrustAlert) => {
      if ((alert.severity === 'critical' || alert.severity === 'emergency') && !alert.acknowledged && !dismissedAlertIdsRef.current.includes(alert.id)) {
        setCriticalAlert(alert);
      }
    },
  });

  const handleDismissAlert = (alertId: string) => {
    const newDismissed = [...dismissedAlertIdsRef.current, alertId];
    dismissedAlertIdsRef.current = newDismissed;
    setDismissedAlertIds(newDismissed);
    setCriticalAlert(undefined);
    useNotificationStore.getState().dismissAlert(alertId);
  };

  const handleNotificationClick = (notification: Notification) => { if (notification.actionUrl) navigate(notification.actionUrl); };

  if (isLoading) {
    return (
      <div className="notif-page">
        <PageGrid />
        <NotificationsPageSkeleton />
      </div>
    );
  }

  return (
    <div className="notif-page">
      <PageGrid />

      {/* Hero */}
      <Chamber className="notif-hero" ribs>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="MODULE NT-01" secondary="Notifications" position="top-right" />

        <div className="notif-hero__header">
          <div className="notif-hero__title-row">
            <TrustSeal size="lg" />
            <h1 className="notif-hero__title">{t('notificationsPage.title')}</h1>
          </div>
          <p className="notif-hero__subtitle">{t('notificationsPage.subtitle')}</p>
          <div className="notif-hero__actions">
            <FrameButton size="sm" onClick={() => navigate(-1)} iconLeft={<ArrowLeft className="notif-icon-sm" />}>
              {t('notificationsPage.back')}
            </FrameButton>
          </div>
        </div>
      </Chamber>

      {/* Trust Alert Banner */}
      {criticalAlert && (
        <div className="notif-alert-container">
          <TrustAlertBanner
            alert={criticalAlert}
            onDismiss={handleDismissAlert}
            onAction={(action, alert) => { if (action === 'view_details' && alert.actionUrl) navigate(alert.actionUrl); }}
          />
        </div>
      )}

      {/* Main Content Grid */}
      <div className="notif-content-grid">
        <div className="notif-center-panel">
          <NotificationCenter maxHeight="calc(100vh - 300px)" onNotificationClick={handleNotificationClick} />
        </div>

        <Chamber className="notif-activity-sidebar">
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <div className="notif-activity__header">
            <Bell className="notif-icon-sm notif-icon-accent" />
            <h2 className="notif-activity__title">{t('notificationsPage.activityFeed')}</h2>
          </div>
          <div className="notif-activity__scroll">
            <RealtimeActivityFeed compact showHeader={false} maxItems={50}
              onActivityClick={(activity) => { if (activity.target?.url) navigate(activity.target.url); }} />
          </div>
        </Chamber>
      </div>
    </div>
  );
}

export default NotificationsPage;
