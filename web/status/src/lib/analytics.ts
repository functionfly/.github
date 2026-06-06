/**
 * Mixpanel Analytics Service for the Status Page
 *
 * Public-facing status pages — no user auth, no consent required.
 * Tracks: page views, incident lookups, subscribe actions.
 *
 * Env vars (Vite):
 *   VITE_MIXPANEL_TOKEN  – Mixpanel project token
 *   VITE_MIXPANEL_HOST  – Mixpanel API host (optional, for EU residency)
 *   VITE_MIXPANEL_DEBUG – Enable debug logging in dev
 */

import mixpanel from 'mixpanel-browser';

let isInitialized = false;
let pendingEvents: Array<{ name: string; payload?: Record<string, unknown> }> = [];

function getConfig() {
  const token = import.meta.env.VITE_MIXPANEL_TOKEN as string | undefined;
  const host = import.meta.env.VITE_MIXPANEL_HOST as string | undefined;
  const debug = import.meta.env.VITE_MIXPANEL_DEBUG === 'true' || import.meta.env.DEV;

  return { token, host, debug };
}

export function initAnalytics(): void {
  if (isInitialized) return;

  const { token, host, debug } = getConfig();
  if (!token) {
    if (import.meta.env.DEV) {
      // eslint-disable-next-line no-console
      console.warn('[status analytics] VITE_MIXPANEL_TOKEN not set — tracking disabled');
    }
    return;
  }

  mixpanel.init(token, {
    api_host: host ?? 'https://api.mixpanel.com',
    debug,
    autocapture: false,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    track_pageview: false as any,
    ignore_dnt: false,
  });

  isInitialized = true;

  for (const { name, payload } of pendingEvents) {
    mixpanel.track(name, payload ?? {});
  }
  pendingEvents = [];
}

export function trackEvent(
  name: string,
  payload?: Record<string, unknown>
): void {
  if (!isInitialized) {
    pendingEvents.push({ name, payload });
    return;
  }

  const cleanPayload = Object.fromEntries(
    Object.entries(payload ?? {}).filter(([, v]) => v !== undefined)
  ) as Record<string, unknown>;

  mixpanel.track(name, cleanPayload);
}

export function trackPageView(path: string): void {
  if (!isInitialized) {
    pendingEvents.push({ name: 'page_view', payload: { path } });
    return;
  }
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  mixpanel.track_pageview(path as any);
}
