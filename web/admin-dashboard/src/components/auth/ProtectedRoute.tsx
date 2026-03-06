/**
 * Protected Route Component
 * Ensures user is authenticated before accessing admin pages
 */

import { Navigate, Outlet } from 'react-router-dom';
import { useAdminAuth } from '@/hooks/useAdminAuth';
import { AdminSecurityGate } from '@/components/security/AdminSecurityGate';

export function ProtectedRoute() {
  const { isAuthenticated, isSessionValid } = useAdminAuth();

  if (!isAuthenticated || !isSessionValid()) {
    return <Navigate to="/auth/login" replace />;
  }

  return (
    <AdminSecurityGate>
      <Outlet />
    </AdminSecurityGate>
  );
}
