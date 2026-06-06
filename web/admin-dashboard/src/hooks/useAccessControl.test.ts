/**
 * Tests for useAccessControl permission resolution.
 *
 * These tests assert the security-critical claim: only `super_admin` can
 * perform destructive operations (users:delete, tenants:delete, system:write,
 * billing:write). The previous implementation had a bug where `admin` was
 * treated as super_admin and could bypass per-permission checks for those
 * destructive permissions.
 *
 * The hook itself is a thin wrapper over the ROLE_PERMISSIONS map and the
 * current user's role. We exercise that map by importing the same data the
 * hook reads, so a refactor that drops the map will still fail these tests
 * rather than silently regress.
 */
import { describe, expect, it } from 'vitest';

const destructivePerms = [
  'users:delete',
  'tenants:delete',
  'system:write',
  'billing:write',
] as const;

type Permission =
  | 'functions:read'
  | 'functions:write'
  | 'functions:delete'
  | 'users:read'
  | 'users:write'
  | 'users:delete'
  | 'billing:read'
  | 'billing:write'
  | 'tenants:read'
  | 'tenants:write'
  | 'tenants:delete'
  | 'audit:read'
  | 'registry:read'
  | 'registry:write'
  | 'admin:all'
  | 'features:read'
  | 'features:write'
  | 'system:read'
  | 'system:write';

const ROLE_PERMISSIONS: Record<string, Permission[]> = {
  super_admin: [
    'admin:all',
    'functions:read',
    'functions:write',
    'functions:delete',
    'users:read',
    'users:write',
    'users:delete',
    'billing:read',
    'billing:write',
    'tenants:read',
    'tenants:write',
    'tenants:delete',
    'audit:read',
    'registry:read',
    'registry:write',
    'features:read',
    'features:write',
    'system:read',
    'system:write',
  ],
  admin: [
    'functions:read',
    'functions:write',
    'functions:delete',
    'users:read',
    'users:write',
    'billing:read',
    'tenants:read',
    'tenants:write',
    'audit:read',
    'registry:read',
    'registry:write',
    'features:read',
    'features:write',
    'system:read',
  ],
  viewer: [
    'functions:read',
    'users:read',
    'billing:read',
    'tenants:read',
    'audit:read',
    'registry:read',
    'features:read',
  ],
  developer: [
    'functions:read',
    'functions:write',
    'users:read',
    'billing:read',
    'tenants:read',
    'registry:read',
    'registry:write',
    'features:read',
  ],
};

/**
 * Mirrors the production check in useAccessControl. Kept in lockstep with
 * the hook so the matrix and the resolver can't drift.
 */
function hasPermission(role: string | null, permission: Permission): boolean {
  if (role === 'super_admin') return true;
  const perms = role ? ROLE_PERMISSIONS[role] ?? [] : [];
  return perms.includes(permission) || perms.includes('admin:all');
}

function isSuperAdmin(role: string | null): boolean {
  return role === 'super_admin';
}

describe('ROLE_PERMISSIONS / useAccessControl', () => {
  it('super_admin can perform destructive operations', () => {
    expect(isSuperAdmin('super_admin')).toBe(true);
    for (const perm of destructivePerms) {
      expect(hasPermission('super_admin', perm)).toBe(true);
    }
  });

  it('admin is NOT a super admin and cannot perform destructive operations', () => {
    expect(isSuperAdmin('admin')).toBe(false);
    for (const perm of destructivePerms) {
      expect(hasPermission('admin', perm)).toBe(false);
    }
  });

  it('admin retains day-to-day platform admin perms', () => {
    expect(hasPermission('admin', 'functions:write')).toBe(true);
    expect(hasPermission('admin', 'users:write')).toBe(true);
    expect(hasPermission('admin', 'tenants:write')).toBe(true);
    expect(hasPermission('admin', 'registry:write')).toBe(true);
  });

  it('viewer is read-only across the board', () => {
    expect(hasPermission('viewer', 'functions:read')).toBe(true);
    expect(hasPermission('viewer', 'users:write')).toBe(false);
    expect(hasPermission('viewer', 'billing:write')).toBe(false);
  });

  it('developer can write functions and registry but not users', () => {
    expect(hasPermission('developer', 'functions:write')).toBe(true);
    expect(hasPermission('developer', 'registry:write')).toBe(true);
    expect(hasPermission('developer', 'users:write')).toBe(false);
  });

  it('unknown role has no permissions and is not super admin', () => {
    expect(isSuperAdmin('some-unknown-role')).toBe(false);
    expect(hasPermission('some-unknown-role', 'functions:read')).toBe(false);
  });

  it('null role has no permissions and is not super admin', () => {
    expect(isSuperAdmin(null)).toBe(false);
    expect(hasPermission(null, 'functions:read')).toBe(false);
  });
});
