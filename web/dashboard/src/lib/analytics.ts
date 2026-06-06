/**
 * Lightweight analytics shim that routes to Mixpanel.
 *
 * Call sites do not need to change — import trackEvent / trackPageView from
 * this module or from '@/lib/analytics' (both are equivalent).
 *
 * Env vars (Vite):
 *   VITE_MIXPANEL_TOKEN    – Mixpanel project token
 *   VITE_MIXPANEL_HOST    – Mixpanel API host (optional, for EU residency)
 *   VITE_MIXPANEL_DEBUG   – Enable Mixpanel debug logging in dev
 */
export type AnalyticsPayload = Record<string, string | number | boolean | null | undefined>;

export { trackEvent, trackPageView } from './analytics/index';
