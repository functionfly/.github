import type { NotificationCount } from '@/api/notifications';
import type { UnreadCounts } from '@/stores/notificationStore';

/** Map GET /v1/notifications/unread-count into Zustand unread shape. */
export function unreadPartialFromServerCount(
  counts: NotificationCount | null | undefined
): Partial<UnreadCounts> {
  if (!counts) return {};
  const byCategory = counts.byCategory || {};
  const unreadTotal = typeof counts.unread === 'number' ? counts.unread : counts.total || 0;
  return {
    all: unreadTotal,
    trust: byCategory.trust || 0,
    revenue: byCategory.revenue || 0,
    issues: byCategory.issues || 0,
    messages: byCategory.messages || 0,
    security: byCategory.security || 0,
  };
}

/** Backend category → store slice (API uses `team`; dashboard store uses `messages`). */
export function unreadStoreKeyFromEventCategory(category: string): keyof UnreadCounts | null {
  if (category === 'team') return 'messages';
  if (
    category === 'trust' ||
    category === 'revenue' ||
    category === 'issues' ||
    category === 'messages' ||
    category === 'security'
  ) {
    return category;
  }
  return null;
}
