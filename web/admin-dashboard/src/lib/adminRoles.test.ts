import { describe, expect, it } from 'vitest';
import { isAdminRole } from '@/lib/adminRoles';

describe('isAdminRole', () => {
  it('accepts platform admin roles', () => {
    expect(isAdminRole('super_admin')).toBe(true);
    expect(isAdminRole('admin')).toBe(true);
    expect(isAdminRole('support')).toBe(true);
    expect(isAdminRole('billing_admin')).toBe(true);
    expect(isAdminRole('developer_admin')).toBe(true);
  });

  it('rejects non-admin roles', () => {
    expect(isAdminRole('read_only')).toBe(false);
    expect(isAdminRole('user')).toBe(false);
    expect(isAdminRole(null)).toBe(false);
  });
});
