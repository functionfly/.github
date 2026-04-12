/**
 * Protected Route with Permission Check
 * Redirects to /forbidden when user lacks required permission
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
  const { hasPermission, isEnterprise } = useAccessControl();

  // Enterprise users have all permissions
  if (isEnterprise) {
    return <>{children ?? <Outlet />}</>;
  }

  // Check if user has the required permission
  if (!hasPermission(requiredPermission)) {
    // Redirect to forbidden page with context
    return (
      <Navigate
        to="/forbidden"
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
