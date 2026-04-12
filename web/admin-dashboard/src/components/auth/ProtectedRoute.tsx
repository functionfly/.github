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
  // Use the Zustand store as the single source of truth for auth state.
  // adminApiClient.isAuthenticated() only checks in-memory token which can
  // be out of sync with the store (e.g. after a hard refresh or 401 redirect).
  const isAuthenticated = useAdminAuthStore((s) => s.isAuthenticated);

  if (!isAuthenticated) {
    // Redirect to login page with return URL
    return <Navigate to="/auth/login" state={{ from: location }} replace />;
  }

  // If a specific permission is required, check it
  if (requiredPermission) {
    const { hasPermission, isSuperAdmin } = useAccessControl();

    if (!hasPermission(requiredPermission)) {
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
  }

  return <>{children ?? <Outlet />}</>;
}
