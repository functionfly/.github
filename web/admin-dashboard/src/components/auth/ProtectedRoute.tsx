/**
 * Protected Route Component
 * Wraps routes that require authentication
 */

import React from 'react';
import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { useAdminAuthStore } from '@/stores/adminAuthStore';

interface ProtectedRouteProps {
  children?: React.ReactNode;
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const location = useLocation();
  // Use the Zustand store as the single source of truth for auth state.
  // adminApiClient.isAuthenticated() only checks in-memory token which can
  // be out of sync with the store (e.g. after a hard refresh or 401 redirect).
  const isAuthenticated = useAdminAuthStore((s) => s.isAuthenticated);

  if (!isAuthenticated) {
    // Redirect to login page with return URL
    return <Navigate to="/auth/login" state={{ from: location }} replace />;
  }

  return <>{children ?? <Outlet />}</>;
}
