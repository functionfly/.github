/**
 * Cloudflare Access places CF_Authorization on the admin hostname when Zero Trust is enabled.
 * When VITE_EXPECT_ZT_HEADERS=true we require that cookie in production builds.
 */
export function isZeroTrustSessionPresent(): boolean {
  if (import.meta.env.VITE_EXPECT_ZT_HEADERS !== 'true') {
    return true;
  }

  if (import.meta.env.DEV || import.meta.env.VITE_DEVELOPMENT === 'true') {
    return true;
  }

  if (typeof document === 'undefined') {
    return true;
  }

  return document.cookie.split(';').some((part) => part.trim().startsWith('CF_Authorization='));
}

export function zeroTrustBlockedReason(): string | null {
  if (isZeroTrustSessionPresent()) {
    return null;
  }
  return 'cloudflare_access_required';
}
