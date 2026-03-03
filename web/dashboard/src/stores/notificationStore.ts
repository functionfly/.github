/**
 * Notification Store
 *
 * Zustand store for managing notification state with persistence.
 * Handles unread counts, dismissed alerts, filters, and last read timestamps.
 */

import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { NotificationCategory } from '@/types/notifications';

// ============================================================================
// Types & Interfaces
// ============================================================================

export interface NotificationFilter {
  category: NotificationCategory;
  priority: 'all' | 'low' | 'medium' | 'high' | 'critical';
  status: 'all' | 'unread' | 'read' | 'archived';
}

export interface UnreadCounts {
  all: number;
  trust: number;
  revenue: number;
  issues: number;
  messages: number;
  security: number;
}

interface NotificationState {
  // State
  unreadCounts: UnreadCounts;
  dismissedAlertIds: string[];
  filter: NotificationFilter;
  lastReadTimestamp: string | null;
  isNotificationCenterOpen: boolean;

  // Actions
  markAsRead: (category?: NotificationCategory) => void;
  dismissAlert: (alertId: string) => void;
  setFilter: (filter: Partial<NotificationFilter>) => void;
  clearAll: () => void;
  updateUnreadCount: (category: keyof UnreadCounts, count: number) => void;
  updateUnreadCounts: (counts: Partial<UnreadCounts>) => void;
  setNotificationCenterOpen: (isOpen: boolean) => void;
  resetDismissedAlerts: () => void;
  updateLastReadTimestamp: () => void;
}

// ============================================================================
// Default State
// ============================================================================

const defaultUnreadCounts: UnreadCounts = {
  all: 0,
  trust: 0,
  revenue: 0,
  issues: 0,
  messages: 0,
  security: 0,
};

const defaultFilter: NotificationFilter = {
  category: 'all',
  priority: 'all',
  status: 'unread',
};

// ============================================================================
// Store Creation
// ============================================================================

export const useNotificationStore = create<NotificationState>()(
  persist(
    (set, get) => ({
      // Initial State
      unreadCounts: { ...defaultUnreadCounts },
      dismissedAlertIds: [],
      filter: { ...defaultFilter },
      lastReadTimestamp: null,
      isNotificationCenterOpen: false,

      // Mark notifications as read (optionally by category)
      markAsRead: (category) => {
        const { unreadCounts } = get();

        if (category && category !== 'all') {
          // Mark specific category as read
          set({
            unreadCounts: {
              ...unreadCounts,
              [category]: 0,
              all: Math.max(0, unreadCounts.all - unreadCounts[category]),
            },
          });
        } else {
          // Mark all as read
          set({
            unreadCounts: { ...defaultUnreadCounts },
            lastReadTimestamp: new Date().toISOString(),
          });
        }
      },

      // Dismiss a trust alert (persists dismissal)
      dismissAlert: (alertId) => {
        const { dismissedAlertIds } = get();
        if (!dismissedAlertIds.includes(alertId)) {
          set({
            dismissedAlertIds: [...dismissedAlertIds, alertId],
          });
        }
      },

      // Update filter settings
      setFilter: (filter) => {
        set({
          filter: { ...get().filter, ...filter },
        });
      },

      // Clear all notifications and reset state
      clearAll: () => {
        set({
          unreadCounts: { ...defaultUnreadCounts },
          lastReadTimestamp: new Date().toISOString(),
        });
      },

      // Update unread count for a specific category
      updateUnreadCount: (category, count) => {
        const { unreadCounts } = get();
        const diff = count - unreadCounts[category];

        set({
          unreadCounts: {
            ...unreadCounts,
            [category]: count,
            all: Math.max(0, unreadCounts.all + diff),
          },
        });
      },

      // Batch update unread counts
      updateUnreadCounts: (counts) => {
        const { unreadCounts } = get();
        const newCounts = { ...unreadCounts, ...counts };

        // Recalculate total
        newCounts.all =
          newCounts.trust +
          newCounts.revenue +
          newCounts.issues +
          newCounts.messages +
          newCounts.security;

        set({ unreadCounts: newCounts });
      },

      // Toggle notification center visibility (only update when changed to avoid re-render loops)
      setNotificationCenterOpen: (isOpen) => {
        if (get().isNotificationCenterOpen !== isOpen) {
          set({ isNotificationCenterOpen: isOpen });
        }
      },

      // Reset dismissed alerts (for testing or user preference)
      resetDismissedAlerts: () => {
        set({ dismissedAlertIds: [] });
      },

      // Update last read timestamp
      updateLastReadTimestamp: () => {
        set({ lastReadTimestamp: new Date().toISOString() });
      },
    }),
    {
      name: 'notification-storage',
      partialize: (state) => ({
        dismissedAlertIds: state.dismissedAlertIds,
        filter: state.filter,
        lastReadTimestamp: state.lastReadTimestamp,
        // Don't persist unreadCounts - those should be fetched fresh
        // Don't persist isNotificationCenterOpen - should default to closed
      }),
    }
  )
);

// ============================================================================
// Selectors (for computed values)
// ============================================================================

export const selectTotalUnreadCount = (state: NotificationState): number =>
  state.unreadCounts.all;

export const selectHasUnreadNotifications = (state: NotificationState): boolean =>
  state.unreadCounts.all > 0;

export const selectIsAlertDismissed = (alertId: string) => (state: NotificationState): boolean =>
  state.dismissedAlertIds.includes(alertId);

export const selectUnreadCountByCategory = (category: keyof UnreadCounts) =>
  (state: NotificationState): number => state.unreadCounts[category];
