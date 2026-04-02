/**
 * Client-side guard for Stripe-hosted redirects (onboarding, checkout).
 * Mitigates open redirects if an API response were ever tampered with or mis-issued.
 * Server-side validation remains authoritative; this is defense in depth.
 */

const ALLOWED_STRIPE_HOSTS = new Set([
  'connect.stripe.com',
  'checkout.stripe.com',
  'billing.stripe.com',
  'invoice.stripe.com',
]);

export class StripeRedirectError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'StripeRedirectError';
  }
}

/**
 * Throws if `url` is not an absolute HTTPS URL to a known Stripe host.
 */
export function assertStripeHostedRedirectUrl(url: string): void {
  let parsed: URL;
  try {
    parsed = new URL(url);
  } catch {
    throw new StripeRedirectError('Invalid URL');
  }

  if (parsed.protocol !== 'https:') {
    throw new StripeRedirectError('URL must use HTTPS');
  }

  const host = parsed.hostname.toLowerCase();
  if (!ALLOWED_STRIPE_HOSTS.has(host)) {
    throw new StripeRedirectError('URL is not an allowed Stripe host');
  }
}

/**
 * Navigates only after allowlist check. Returns whether navigation was started.
 */
export function navigateToStripeHostedUrl(url: string): boolean {
  try {
    assertStripeHostedRedirectUrl(url);
    window.location.assign(url);
    return true;
  } catch {
    return false;
  }
}
