/**
 * Access Control Hook
 * Checks user permissions for specific features/routes
 */

import React from 'react';
import { useAdminAuthStore } from '@/stores/adminAuthStore';

export type Permission =
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

export interface AccessDeniedReason {
  permission: Permission;
  message: string;
  requiredRole?: string;
}

interface UseAccessControlReturn {
  hasPermission: (permission: Permission) => boolean;
  hasAnyPermission: (permissions: Permission[]) => boolean;
  hasAllPermissions: (permissions: Permission[]) => boolean;
  checkAccess: (permission: Permission) => AccessDeniedReason | null;
  currentRole: string | null;
  isSuperAdmin: boolean;
}
const ROLE_PERMISSIONS: Record<string, Permission[]> = {
  super_admin: ['admin:all'],
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

const PERMISSION_MESSAGES: Record<Permission, string> = {
  'functions:read': 'View functions',
  'functions:write': 'Create or modify functions',
  'functions:delete': 'Delete functions',
  'users:read': 'View user accounts',
  'users:write': 'Create or modify users',
  'users:delete': 'Delete users',
  'billing:read': 'View billing information',
  'billing:write': 'Modify billing settings',
  'tenants:read': 'View tenants',
  'tenants:write': 'Create or modify tenants',
  'tenants:delete': 'Delete tenants',
  'audit:read': 'View audit logs',
  'registry:read': 'View function registry',
  'registry:write': 'Publish to registry',
  'admin:all': 'Full administrative access',
  'features:read': 'View feature flags',
  'features:write': 'Modify feature flags',
  'system:read': 'View system settings',
  'system:write': 'Modify system settings',
};

export function useAccessControl(): UseAccessControlReturn {
  const user = useAdminAuthStore((s) => s.user);

  const currentRole = user?.role?.toLowerCase() ?? null;
  const isSuperAdmin = currentRole === 'super_admin' || currentRole === 'admin';

  const permissions = currentRole ? ROLE_PERMISSIONS[currentRole] ?? [] : [];

  const hasPermission = (permission: Permission): boolean => {
    if (isSuperAdmin) return true;
    if (permissions.includes('admin:all')) return true;
    return permissions.includes(permission);
  };

  const hasAnyPermission = (perms: Permission[]): boolean => {
    if (isSuperAdmin) return true;
    return perms.some((p) => permissions.includes(p) || permissions.includes('admin:all'));
  };

  const hasAllPermissions = (perms: Permission[]): boolean => {
    if (isSuperAdmin) return true;
    return perms.every((p) => permissions.includes(p) || permissions.includes('admin:all'));
  };

  const checkAccess = (permission: Permission): AccessDeniedReason | null => {
    if (hasPermission(permission)) return null;

    return {
      permission,
      message: PERMISSION_MESSAGES[permission] ?? 'This feature',
      requiredRole: currentRole ?? 'authenticated',
    };
  };

  return {
    hasPermission,
    hasAnyPermission,
    hasAllPermissions,
    checkAccess,
    currentRole,
    isSuperAdmin,
  };
}

/**
 * HOC wrapper for protecting components that require specific permissions
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
type AnyComponent = React.ComponentType<any>;

export function withAccessControl(
  WrappedComponent: AnyComponent,
  requiredPermission: Permission
): AnyComponent {
  return function AccessControlledComponent(props: Record<string, unknown>) {
    const { hasPermission } = useAccessControl();

    if (!hasPermission(requiredPermission)) {
      return null;
    }

    return <WrappedComponent {...(props as object)} />;
  };
}
