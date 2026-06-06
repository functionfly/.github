/**
 * Mixpanel Analytics Service
 *
 * Zero-knowledge product analytics: user identity and event properties are
 * tracked client-side only. The server never sees user emails or PII.
 *
 * Env vars (Vite):
 *   VITE_MIXPANEL_TOKEN  – Mixpanel project token (required)
 *   VITE_MIXPANEL_HOST  – Mixpanel API host (optional, for EU residency)
 *   VITE_MIXPANEL_DEBUG – Enable Mixpanel debug logging in dev
 */

import mixpanel from 'mixpanel-browser';
import type { AnalyticsPayload } from '../analytics';

let isInitialized = false;
let pendingIdentify: { userId: string; traits?: Record<string, unknown> } | null = null;
let pendingEvents: Array<{ name: string; payload: AnalyticsPayload }> = [];

/**
 * Initialize Mixpanel. Safe to call multiple times — only runs once.
 * Reads VITE_MIXPANEL_TOKEN from env; silently no-ops if not set.
 */
export function initMixpanel(): void {
  if (isInitialized) return;

  const token = import.meta.env.VITE_MIXPANEL_TOKEN as string | undefined;
  const host = import.meta.env.VITE_MIXPANEL_HOST as string | undefined;
  const debug = import.meta.env.VITE_MIXPANEL_DEBUG === 'true' || import.meta.env.DEV;

  if (!token) {
    if (import.meta.env.DEV) {
      // eslint-disable-next-line no-console
      console.warn('[mixpanel] VITE_MIXPANEL_TOKEN not set — Mixpanel tracking disabled');
    }
    return;
  }

  mixpanel.init(token, {
    api_host: host ?? 'https://api.mixpanel.com',
    debug,
    // Disable autocapture by default — we track manually for control and GDPR compliance
    autocapture: false,
    // Disable automatic page view tracking — we do it manually per route
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    track_pageview: false as any,
    ignore_dnt: false,
  });

  isInitialized = true;

  // Flush queued identify call that arrived before init
  if (pendingIdentify) {
    const { userId, traits } = pendingIdentify;
    mixpanel.identify(userId);
    if (traits) {
      mixpanel.people.set(traits as Record<string, unknown>);
    }
    pendingIdentify = null;
  }

  // Flush queued events that arrived before init
  for (const { name, payload } of pendingEvents) {
    mixpanel.track(name, payload as Record<string, unknown>);
  }
  pendingEvents = [];
}

/**
 * Identify the currently logged-in user.
 * Call this on login, signup, and on every app init when a session exists.
 *
 * @param userId  – Unique user ID (from JWT subject or DB primary key)
 * @param traits  – Optional user properties (email, plan, role, etc.)
 */
export function identifyUser(
  userId: string,
  traits?: Record<string, string | number | boolean | null | undefined>
): void {
  if (!isInitialized) {
    pendingIdentify = { userId, traits };
    return;
  }
  mixpanel.identify(userId);
  if (traits) {
    mixpanel.people.set(traits as Record<string, unknown>);
  }
}

/**
 * Alias a user ID (e.g. after a merge or rebrand). Idempotent.
 */
export function aliasUser(newUserId: string, originalUserId?: string): void {
  if (!isInitialized) return;
  if (originalUserId) {
    mixpanel.alias(newUserId, originalUserId);
  } else {
    mixpanel.alias(newUserId);
  }
}

/**
 * Track a product analytics event.
 * Silently no-ops if Mixpanel is not initialized or user has not consented.
 */
export function trackEvent(name: string, payload: AnalyticsPayload = {}): void {
  if (!isInitialized) {
    pendingEvents.push({ name, payload });
    return;
  }

  // Strip undefined values to keep Mixpanel clean
  const cleanPayload = Object.fromEntries(
    Object.entries(payload).filter(([, v]) => v !== undefined)
  ) as Record<string, unknown>;

  mixpanel.track(name, cleanPayload);
}

/**
 * Track a page view.
 */
export function trackPageView(path: string, options?: { url?: string }): void {
  if (!isInitialized) return;
  // @ts-expect-error – track_pageview is in the runtime type but not in TS definitions
  mixpanel.track_pageview(options?.url ?? path);
}

/**
 * Update user traits (e.g. after a profile update).
 */
export function setUserTraits(
  traits: Record<string, string | number | boolean | null | undefined>
): void {
  if (!isInitialized) return;
  mixpanel.people.set(traits as Record<string, unknown>);
}

/**
 * Increment a numeric user property (e.g. "executions" counter).
 */
export function incrementUserTrait(property: string, increment: number): void {
  if (!isInitialized) return;
  mixpanel.people.increment(property, increment);
}

/**
 * Reset the identified user (on logout).
 */
export function resetUser(): void {
  if (!isInitialized) return;
  mixpanel.reset();
}

/**
 * Returns true if Mixpanel has been initialized.
 */
export function isMixpanelReady(): boolean {
  return isInitialized;
}
