/**
 * NotificationBell Component
 *
 * A bell icon button that displays unread notification count
 * and opens a Sheet/drawer with the NotificationCenter on click.
 */

import { useState, useEffect, useCallback, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { Bell } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import CountUp from 'react-countup';
import { Button } from '@/components/ui/button';
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
  SheetClose,
} from '@/components/ui/sheet';
import { ScrollArea } from '@/components/ui/scroll-area';
import { NotificationCenter } from './NotificationCenter';
import { useNotificationStore } from '@/stores/notificationStore';
import { useNotificationRealtime } from '@/hooks/useNotificationRealtime';
import { useAuthStore } from '@/stores/authStore';
import { notificationsApi } from '@/api/notifications';
import type { Notification } from '@/types/notifications';

// ============================================================================
// Types & Interfaces
// ============================================================================

interface NotificationBellProps {
  /** Optional className for styling */
  className?: string;
  /** Variant for the button */
  variant?: 'ghost' | 'outline' | 'default';
  /** Size of the bell icon */
  size?: 'sm' | 'md' | 'lg';
}

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Get the size class for the bell icon
 */
function getIconSize(size: NotificationBellProps['size']): string {
  switch (size) {
    case 'sm':
      return 'w-4 h-4';
    case 'lg':
      return 'w-6 h-6';
    case 'md':
    default:
      return 'w-5 h-5';
  }
}

/**
 * Get the badge size class based on bell size
 */
function getBadgeSize(size: NotificationBellProps['size']): string {
  switch (size) {
    case 'sm':
      return 'min-w-[14px] h-[14px] text-[10px]';
    case 'lg':
      return 'min-w-[22px] h-[22px] text-sm';
    case 'md':
    default:
      return 'min-w-[18px] h-[18px] text-xs';
  }
}

/** Play a short notification ping using Web Audio (no external URL, avoids 403 hotlink issues). */
function playNotificationPing(): void {
  if (typeof window === 'undefined' || !window.AudioContext) return;
  try {
    const ctx = new window.AudioContext();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    osc.connect(gain);
    gain.connect(ctx.destination);
    osc.frequency.value = 800;
    osc.type = 'sine';
    gain.gain.setValueAtTime(0.15, ctx.currentTime);
    gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.12);
    osc.start(ctx.currentTime);
    osc.stop(ctx.currentTime + 0.12);
  } catch {
    // Ignore (e.g. autoplay policy, AudioContext not allowed)
  }
}

// ============================================================================
// Main Component
// ============================================================================

export function NotificationBell({
  className,
  variant = 'ghost',
  size = 'md',
}: NotificationBellProps) {
  const navigate = useNavigate();
  const user = useAuthStore((state) => state.user);

  // Store state
  const unreadCount = useNotificationStore((state) => state.unreadCounts.all);
  const updateUnreadCount = useNotificationStore((state) => state.updateUnreadCount);
  const updateUnreadCounts = useNotificationStore((state) => state.updateUnreadCounts);
  const setNotificationCenterOpen = useNotificationStore((state) => state.setNotificationCenterOpen);

  // Uncontrolled Sheet: no open/onOpenChange so Radix manages state and we avoid update loops
  const sheetCloseRef = useRef<HTMLButtonElement>(null);
  const [hasNewNotification, setHasNewNotification] = useState(false);

  // Notification ping (Web Audio, no external request)
  const playNotificationSoundRef = useRef(playNotificationPing);

  const closeSheet = useCallback(() => {
    sheetCloseRef.current?.click();
    setNotificationCenterOpen(false);
  }, [setNotificationCenterOpen]);

  // Fetch initial unread count
  useEffect(() => {
    const fetchUnreadCount = async () => {
      if (!user?.id) return;

      try {
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
      } catch (error) {
        console.error('Failed to fetch unread count:', error);
      }
    };

    fetchUnreadCount();
  }, [user?.id, updateUnreadCounts]);

  // Real-time notification handling (stable callback to avoid hook dependency churn)
  const onNewNotification = useCallback(
    (notification: Notification) => {
      const currentCount = useNotificationStore.getState().unreadCounts.all;
      updateUnreadCount('all', currentCount + 1);
      if (notification.category !== 'all') {
        const categoryCount = useNotificationStore.getState().unreadCounts[notification.category];
        updateUnreadCount(notification.category, categoryCount + 1);
      }
      setHasNewNotification(true);
      setTimeout(() => setHasNewNotification(false), 3000);
      try {
        playNotificationSoundRef.current();
      } catch {
        // Ignore (e.g. autoplay policy)
      }
    },
    [updateUnreadCount]
  );

  useNotificationRealtime({
    enabled: !!user?.id,
    onNewNotification,
  });

  // Handle notification click - close sheet then navigate
  const handleNotificationClick = (notification: Notification) => {
    closeSheet();
    if (notification.actionUrl) {
      navigate(notification.actionUrl);
    } else {
      navigate('/notifications');
    }
  };

  const handleViewAll = () => {
    closeSheet();
    navigate('/notifications');
  };

  const iconSize = getIconSize(size);
  const badgeSize = getBadgeSize(size);

  return (
    <Sheet>
      <SheetTrigger asChild>
        <Button
          variant={variant}
          size="icon"
          className={className}
          aria-label={`Notifications${unreadCount > 0 ? ` (${unreadCount} unread)` : ''}`}
        >
          <motion.div
            animate={hasNewNotification ? {
              rotate: [0, -15, 15, -15, 15, 0],
              scale: [1, 1.1, 1],
            } : {}}
            transition={{ duration: 0.5 }}
          >
            <Bell className={iconSize} />
          </motion.div>

          {/* Unread Count Badge */}
          <AnimatePresence>
            {unreadCount > 0 && (
              <motion.span
                initial={{ scale: 0, opacity: 0 }}
                animate={{ scale: 1, opacity: 1 }}
                exit={{ scale: 0, opacity: 0 }}
                className={`absolute -top-1 -right-1 flex items-center justify-center ${badgeSize} bg-error text-white font-bold rounded-full px-1`}
              >
                {unreadCount > 99 ? (
                  '99+'
                ) : (
                  <CountUp end={unreadCount} duration={0.35} preserveValue />
                )}
              </motion.span>
            )}
          </AnimatePresence>
        </Button>
      </SheetTrigger>

      <SheetContent
        className="w-full sm:max-w-lg bg-bg-secondary/95 backdrop-blur-xl border-l border-border-default p-0"
        side="right"
      >
        {/* Hidden close trigger for programmatic close (View All, notification click, etc.) */}
        <SheetClose ref={sheetCloseRef} className="sr-only" aria-hidden />
        <SheetHeader className="px-4 py-4 border-b border-border-subtle">
          <div className="flex items-center justify-between">
            <SheetTitle className="text-lg font-semibold text-text-primary">
              Notifications
            </SheetTitle>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleViewAll}
              className="text-brand-400 hover:text-brand-300"
            >
              View All
            </Button>
          </div>
        </SheetHeader>

        <ScrollArea className="h-[calc(100vh-80px)]">
          <NotificationCenter
            className="border-0 shadow-none"
            maxHeight="none"
            onNotificationClick={handleNotificationClick}
            onSettingsClick={() => {
              closeSheet();
              navigate('/settings/notifications');
            }}
          />
        </ScrollArea>
      </SheetContent>
    </Sheet>
  );
}

export default NotificationBell;
