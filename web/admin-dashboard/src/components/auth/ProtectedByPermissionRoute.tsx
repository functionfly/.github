/**
 * Access Controlled Route
 * Wraps routes that require specific permissions
 * Shows access denied page if user lacks required permission
 */

import { useAccessControl, type Permission } from '@/hooks/useAccessControl';
import React from 'react';
import { Navigate, Outlet, useLocation } from 'react-router-dom';

interface ProtectedByPermissionRouteProps {
  children?: React.ReactNode;
  requiredPermission: Permission;
  featureName?: string;
}

export function ProtectedByPermissionRoute({
  children,
  requiredPermission,
  featureName,
}: ProtectedByPermissionRouteProps) {
  const location = useLocation();
  const { hasPermission, isSuperAdmin } = useAccessControl();

  // Super admins always have access
  if (isSuperAdmin) {
    return <>{children ?? <Outlet />}</>;
  }

  // Check if user has the required permission
  if (!hasPermission(requiredPermission)) {
    // Redirect to access denied page with context
    return (
      <Navigate
        to="/access-denied"
        state={{
          from: location,
          permission: requiredPermission,
          featureName,
        }}
        replace
      />
    );
  }

  return <>{children ?? <Outlet />}</>;
}
