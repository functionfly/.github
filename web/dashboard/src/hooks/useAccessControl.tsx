/**
 * Access Control Hook
 * Checks user permissions for specific features/routes
 */

import { useAuthStore } from '@/stores/authStore';

export type Permission =
  | 'functions:read'
  | 'functions:write'
  | 'functions:delete'
  | 'agents:read'
  | 'agents:write'
  | 'agents:delete'
  | 'wallet:read'
  | 'wallet:write'
  | 'secrets:read'
  | 'secrets:write'
  | 'secrets:delete'
  | 'teams:read'
  | 'teams:write'
  | 'teams:delete'
  | 'billing:read'
  | 'billing:write'
  | 'analytics:read'
  | 'statefabric:read'
  | 'statefabric:write'
  | 'registry:read'
  | 'registry:write'
  | 'enterprise:read'
  | 'enterprise:write';

export interface AccessDeniedReason {
  permission: Permission;
  message: string;
  requiredTier?: string;
}

interface UseAccessControlReturn {
  hasPermission: (permission: Permission) => boolean;
  hasAnyPermission: (permissions: Permission[]) => boolean;
  hasAllPermissions: (permissions: Permission[]) => boolean;
  checkAccess: (permission: Permission) => AccessDeniedReason | null;
  currentTier: string;
  isPro: boolean;
  isEnterprise: boolean;
}

const PERMISSION_MESSAGES: Record<Permission, string> = {
  'functions:read': 'View functions',
  'functions:write': 'Create or modify functions',
  'functions:delete': 'Delete functions',
  'agents:read': 'View agents',
  'agents:write': 'Create or modify agents',
  'agents:delete': 'Delete agents',
  'wallet:read': 'View wallet',
  'wallet:write': 'Manage wallet',
  'secrets:read': 'View secrets',
  'secrets:write': 'Create or modify secrets',
  'secrets:delete': 'Delete secrets',
  'teams:read': 'View teams',
  'teams:write': 'Create or modify teams',
  'teams:delete': 'Delete teams',
  'billing:read': 'View billing',
  'billing:write': 'Modify billing',
  'analytics:read': 'View analytics',
  'statefabric:read': 'View State Fabric',
  'statefabric:write': 'Modify State Fabric',
  'registry:read': 'View registry',
  'registry:write': 'Publish to registry',
  'enterprise:read': 'View enterprise features',
  'enterprise:write': 'Modify enterprise settings',
};

// Tier-based permissions (simplified model)
const TIER_PERMISSIONS: Record<string, Permission[]> = {
  free: [
    'functions:read',
    'agents:read',
    'wallet:read',
    'secrets:read',
    'teams:read',
    'billing:read',
    'registry:read',
  ],
  starter: [
    'functions:read',
    'functions:write',
    'agents:read',
    'agents:write',
    'wallet:read',
    'secrets:read',
    'secrets:write',
    'teams:read',
    'teams:write',
    'billing:read',
    'analytics:read',
    'registry:read',
    'registry:write',
  ],
  pro: [
    'functions:read',
    'functions:write',
    'functions:delete',
    'agents:read',
    'agents:write',
    'agents:delete',
    'wallet:read',
    'wallet:write',
    'secrets:read',
    'secrets:write',
    'secrets:delete',
    'teams:read',
    'teams:write',
    'teams:delete',
    'billing:read',
    'billing:write',
    'analytics:read',
    'statefabric:read',
    'statefabric:write',
    'registry:read',
    'registry:write',
  ],
  enterprise: [
    'functions:read',
    'functions:write',
    'functions:delete',
    'agents:read',
    'agents:write',
    'agents:delete',
    'wallet:read',
    'wallet:write',
    'secrets:read',
    'secrets:write',
    'secrets:delete',
    'teams:read',
    'teams:write',
    'teams:delete',
    'billing:read',
    'billing:write',
    'analytics:read',
    'statefabric:read',
    'statefabric:write',
    'registry:read',
    'registry:write',
    'enterprise:read',
    'enterprise:write',
  ],
};

export function useAccessControl(): UseAccessControlReturn {
  const user = useAuthStore((s) => s.user);

  const currentTier = user?.plan?.toLowerCase() ?? 'free';
  const isPro = currentTier === 'pro' || currentTier === 'enterprise';
  const isEnterprise = currentTier === 'enterprise';

  const permissions = TIER_PERMISSIONS[currentTier] ?? TIER_PERMISSIONS.free ?? [];

  const hasPermission = (permission: Permission): boolean => {
    // Enterprise has all permissions
    if (isEnterprise) return true;
    // Pro has all permissions except enterprise:write
    if (isPro) {
      if (permission === 'enterprise:write') return false;
      return true;
    }
    return permissions.includes(permission);
  };

  const hasAnyPermission = (perms: Permission[]): boolean => {
    if (isEnterprise) return true;
    if (isPro) return !perms.every((p) => p === 'enterprise:write');
    return perms.some((p) => permissions.includes(p));
  };

  const hasAllPermissions = (perms: Permission[]): boolean => {
    if (isEnterprise) return true;
    if (isPro) return !perms.some((p) => p === 'enterprise:write');
    return perms.every((p) => permissions.includes(p));
  };

  const checkAccess = (permission: Permission): AccessDeniedReason | null => {
    if (hasPermission(permission)) return null;

    return {
      permission,
      message: PERMISSION_MESSAGES[permission] ?? 'This feature',
      requiredTier: currentTier,
    };
  };

  return {
    hasPermission,
    hasAnyPermission,
    hasAllPermissions,
    checkAccess,
    currentTier,
    isPro,
    isEnterprise,
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
