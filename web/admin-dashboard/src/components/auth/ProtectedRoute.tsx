/**
 * Protected Route Component
 * Wraps routes that require authentication
 */

import React from 'react';
import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { adminApiClient } from '../../lib/api/adminClient';

interface ProtectedRouteProps {
  children?: React.ReactNode;
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const location = useLocation();
  const isAuthenticated = adminApiClient.isAuthenticated();

  if (!isAuthenticated) {
    // Redirect to login page with return URL
    return <Navigate to="/auth/login" state={{ from: location }} replace />;
  }

  return <>{children ?? <Outlet />}</>;
}
