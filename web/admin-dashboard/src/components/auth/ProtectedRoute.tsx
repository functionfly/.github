/**
 * Protected Route Component
 * Wraps routes that require authentication
 */

import React from 'react';
import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { useAdminAuthStore } from '@/stores/adminAuthStore';
import {
  useAccessControl,
  type Permission,
} from '@/hooks/useAccessControl';

interface ProtectedRouteProps {
  children?: React.ReactNode;
  requiredPermission?: Permission;
  featureName?: string;
}

export function ProtectedRoute({
  children,
  requiredPermission,
  featureName,
}: ProtectedRouteProps) {
  const location = useLocation();
  const isAuthenticated = useAdminAuthStore((s) => s.isAuthenticated);
  const { hasPermission } = useAccessControl();

  if (!isAuthenticated) {
    return <Navigate to="/auth/login" state={{ from: location }} replace />;
  }

  if (requiredPermission && !hasPermission(requiredPermission)) {
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
