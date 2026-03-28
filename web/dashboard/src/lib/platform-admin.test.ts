import { describe, expect, it } from 'vitest';
import { decodeJwtRole, isPlatformAdminRole } from './platform-admin';

describe('isPlatformAdminRole', () => {
  it('accepts backend platform roles including admin', () => {
    expect(isPlatformAdminRole('admin')).toBe(true);
    expect(isPlatformAdminRole('super_admin')).toBe(true);
    expect(isPlatformAdminRole('support')).toBe(true);
    expect(isPlatformAdminRole('user')).toBe(false);
    expect(isPlatformAdminRole(undefined)).toBe(false);
  });
});

describe('decodeJwtRole', () => {
  it('reads role from JWT payload', () => {
    const payload = btoa(JSON.stringify({ role: 'admin', exp: 9999999999 }));
    expect(decodeJwtRole(`x.${payload}.y`)).toBe('admin');
  });

  it('returns undefined for invalid token', () => {
    expect(decodeJwtRole(null)).toBeUndefined();
    expect(decodeJwtRole('not-a-jwt')).toBeUndefined();
  });
});
