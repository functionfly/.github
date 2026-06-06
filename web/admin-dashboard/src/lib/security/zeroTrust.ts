/**
 * Cloudflare Access places CF_Authorization on the admin hostname when Zero Trust
 * is enabled. When VITE_EXPECT_ZT_HEADERS=true we require that cookie in
 * production builds. In dev/non-prod the check is a no-op so local development
 * is unaffected.
 *
 * Note: `document.cookie` is HttpOnly for the CF_Authorization cookie set by
 * Cloudflare Access in typical deployments, so this client-side check is
 * a "soft" gate. The hard enforcement happens at the Cloudflare edge: if the
 * cookie is missing, Cloudflare intercepts the request before it reaches
 * this bundle. This guard exists to give the user a clear UI message in the
 * rare case where the app shell loads directly (e.g. during a Cloudflare
 * policy rollout) instead of bouncing through the access page.
 */

/**
 * Pure parser: does the cookie blob contain a CF_Authorization entry that
 * begins at a cookie boundary (i.e. not just somewhere inside a longer
 * cookie name)? Exported for testing and reuse.
 */
export function cookieHasCfAuthorization(cookieBlob: string | undefined | null): boolean {
  if (!cookieBlob) return false;
  return cookieBlob
    .split(';')
    .some((part) => part.trim().startsWith('CF_Authorization='));
}

export interface ZeroTrustContext {
  /** Value of the VITE_EXPECT_ZT_HEADERS env at build time. */
  expectHeaders: boolean;
  /** True if we're in a Vite dev build (skips enforcement). */
  isDev: boolean;
  /** Value of the VITE_DEVELOPMENT env at build time. */
  isDevelopmentEnv: boolean;
  /** Current document.cookie string. Pass undefined in non-DOM contexts. */
  cookie: string | undefined;
}

/**
 * Pure decision function. Given a context describing the current env and
 * the cookie string, returns true if a valid Cloudflare Access session is
 * present (or no enforcement is required).
 */
export function checkZeroTrust(ctx: ZeroTrustContext): boolean {
  if (!ctx.expectHeaders) return true;
  if (ctx.isDev || ctx.isDevelopmentEnv) return true;
  if (ctx.cookie === undefined) return true;
  return cookieHasCfAuthorization(ctx.cookie);
}

/** Read the runtime Zero Trust context. Used by the public wrappers. */
function readContext(): ZeroTrustContext {
  const env = (import.meta as unknown as { env: Record<string, unknown> }).env;
  return {
    expectHeaders: env.VITE_EXPECT_ZT_HEADERS === 'true',
    isDev: Boolean(env.DEV),
    isDevelopmentEnv: env.VITE_DEVELOPMENT === 'true',
    cookie: typeof document === 'undefined' ? undefined : document.cookie,
  };
}

export function isZeroTrustEnabled(): boolean {
  return readContext().expectHeaders;
}

export function isZeroTrustSessionPresent(): boolean {
  return checkZeroTrust(readContext());
}

export function zeroTrustBlockedReason(): string | null {
  return isZeroTrustSessionPresent() ? null : 'cloudflare_access_required';
}
