/**
 * Tests for the zero-trust (Cloudflare Access) gate.
 *
 * The gate is intentionally split into pure decision functions
 * (`cookieHasCfAuthorization`, `checkZeroTrust`) and thin wrappers
 * that read import.meta.env + document.cookie at runtime. We test the
 * pure functions exhaustively here so the production code's behavior is
 * pinned without depending on vitest's import.meta.env mocking.
 */
import { describe, expect, it } from 'vitest';

import {
  checkZeroTrust,
  cookieHasCfAuthorization,
  type ZeroTrustContext,
} from './zeroTrust';

const baseCtx = (overrides: Partial<ZeroTrustContext> = {}): ZeroTrustContext => ({
  expectHeaders: true,
  isDev: false,
  isDevelopmentEnv: false,
  cookie: '',
  ...overrides,
});

describe('cookieHasCfAuthorization', () => {
  it('returns true for a CF_Authorization cookie at the start of the blob', () => {
    expect(cookieHasCfAuthorization('CF_Authorization=jwt')).toBe(true);
  });

  it('returns true when the cookie is preceded by other cookies', () => {
    expect(cookieHasCfAuthorization('session=abc; CF_Authorization=jwt; theme=dark')).toBe(true);
  });

  it('returns true when the cookie has additional attributes', () => {
    expect(cookieHasCfAuthorization('CF_Authorization=jwt; Path=/; Secure')).toBe(true);
  });

  it('returns false when the cookie name is a prefix of another cookie', () => {
    expect(cookieHasCfAuthorization('foo_CF_Authorization=evil')).toBe(false);
    expect(cookieHasCfAuthorization('CF_Authorization_evil=evil')).toBe(false);
  });

  it('returns false for an empty / undefined / null blob', () => {
    expect(cookieHasCfAuthorization('')).toBe(false);
    expect(cookieHasCfAuthorization(undefined)).toBe(false);
    expect(cookieHasCfAuthorization(null)).toBe(false);
  });

  it('returns false when no cookies are present', () => {
    expect(cookieHasCfAuthorization('theme=dark; session=abc')).toBe(false);
  });

  it('trims whitespace around the cookie name match', () => {
    expect(cookieHasCfAuthorization('  CF_Authorization=jwt  ')).toBe(true);
  });
});

describe('checkZeroTrust', () => {
  it('returns true when enforcement is disabled', () => {
    expect(
      checkZeroTrust(
        baseCtx({ expectHeaders: false, cookie: 'session=abc' })
      )
    ).toBe(true);
  });

  it('returns true when CF cookie is present', () => {
    expect(
      checkZeroTrust(
        baseCtx({ cookie: 'CF_Authorization=some-jwt; Path=/' })
      )
    ).toBe(true);
  });

  it('returns false when enforcement is on and CF cookie is missing', () => {
    expect(checkZeroTrust(baseCtx({ cookie: 'session=abc; theme=dark' }))).toBe(false);
  });

  it('treats DEV as a bypass', () => {
    expect(
      checkZeroTruth_helper({ isDev: true, cookie: '' })
    ).toBe(true);
  });

  it('treats VITE_DEVELOPMENT=true as a bypass', () => {
    expect(
      checkZeroTrust(
        baseCtx({ isDevelopmentEnv: true, cookie: '' })
      )
    ).toBe(true);
  });

  it('treats cookie=undefined (non-DOM) as a bypass', () => {
    expect(checkZeroTrust(baseCtx({ cookie: undefined }))).toBe(true);
  });

  it('does not match cookies that merely contain the prefix', () => {
    expect(
      checkZeroTruth_helper({
        cookie: 'foo_CF_Authorization=evil; CF_Authorization_evil=evil',
      })
    ).toBe(false);
  });
});

// Helper that takes a partial and merges with enforcement-on defaults.
function checkZeroTruth_helper(overrides: Partial<ZeroTrustContext>): boolean {
  return checkZeroTrust(baseCtx(overrides));
}
