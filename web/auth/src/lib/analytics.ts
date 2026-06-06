/**
 * Analytics tracking for auth funnel events
 * Supports Mixpanel, Plausible, PostHog, or any analytics provider
 */

type AuthEvent =
  | "signup_start"
  | "signup_step_1_complete" // Account info
  | "signup_step_2_complete" // Profile info
  | "signup_step_3_complete" // Invite/terms
  | "signup_complete"
  | "signup_error"
  | "login_start"
  | "login_success"
  | "login_error"
  | "magic_link_requested"
  | "magic_link_sent"
  | "password_reset_start"
  | "password_reset_sent"
  | "password_reset_complete"
  | "email_verification_start"
  | "email_verification_success"
  | "email_verification_error"
  | "resend_verification"
  | "oauth_start"
  | "oauth_callback"
  | "oauth_error";

interface EventProperties {
  method?: string;
  error?: string;
  provider?: string;
  step?: number;
  has_invite?: boolean;
  duration?: number;
  [key: string]: string | number | boolean | undefined;
}

// Type definitions for analytics globals
interface PlausibleWindow extends Window {
  plausible?: (
    eventName: string,
    options?: { props?: Record<string, unknown> }
  ) => void;
}

interface PostHogWindow extends Window {
  posthog?: {
    capture: (eventName: string, properties?: Record<string, unknown>) => void;
  };
}

/**
 * Track an auth funnel event to all configured providers
 */
export function trackAuth(
  event: AuthEvent,
  properties: EventProperties = {}
): void {
  const eventName = `auth_${event}`;

  // Mixpanel
  if (typeof window !== "undefined" && (window as MixpanelWindow).mixpanel) {
    (window as MixpanelWindow).mixpanel.track(eventName, properties);
  }

  // Plausible
  if (typeof window !== "undefined" && (window as PlausibleWindow).plausible) {
    (window as PlausibleWindow).plausible!(eventName, {
      props: properties,
    });
  }

  // PostHog
  if (typeof window !== "undefined" && (window as PostHogWindow).posthog) {
    (window as PostHogWindow).posthog!.capture(eventName, properties);
  }

  // Console logging in development
  if (
    typeof process !== "undefined" &&
    process.env?.NODE_ENV === "development"
  ) {
    // eslint-disable-next-line no-console
    console.log(`[Analytics] ${eventName}`, properties);
  }
}

/**
 * Track form errors with field details
 */
export function trackFormError(
  formName: string,
  errors: Record<string, string>
): void {
  const errorFields = Object.keys(errors);

  trackAuth("signup_error", {
    form: formName,
    fields: errorFields.join(","),
    error_count: errorFields.length,
  });
}

/**
 * Track timing metrics
 */
export function trackTiming(
  category: string,
  variable: string,
  value: number
): void {
  if (typeof window !== "undefined" && (window as PlausibleWindow).plausible) {
    (window as PlausibleWindow).plausible!("timing", {
      props: { category, variable, value },
    });
  }
}

/**
 * Create a timer for tracking duration
 */
export function createTimer(variable: string): () => void {
  const start = performance.now();
  return () => {
    const duration = Math.round(performance.now() - start);
    trackTiming("auth", variable, duration);
  };
}

/**
 * Initialize analytics (call once on app load)
 */
export function initAnalytics(): void {
  // Plausible auto-tracks page views, no initialization needed
  // PostHog would be initialized here if needed
  // Mixpanel initialization is handled by the script snippet in BaseLayout.astro

  // Track initial page load
  if (typeof window !== "undefined") {
    const path = window.location.pathname;
    if (path.includes("/signup")) {
      trackAuth("signup_start");
    } else if (path.includes("/login")) {
      trackAuth("login_start");
    } else if (path.includes("/forgot-password")) {
      trackAuth("password_reset_start");
    } else if (path.includes("/verify-email")) {
      trackAuth("email_verification_start");
    }
  }
}

// Mixpanel global type
interface MixpanelWindow extends Window {
  mixpanel?: {
    track: (name: string, props?: Record<string, unknown>) => void;
  };
}
