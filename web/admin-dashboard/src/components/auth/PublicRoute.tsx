/**
 * Public Route Component
 * Redirects to the dashboard when the user is already authenticated.
 * Use this to wrap login / auth pages so authenticated users are sent home.
 */

import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { adminApiClient } from '../../lib/api/adminClient';

export function PublicRoute() {
  const location = useLocation();
  const isAuthenticated = adminApiClient.isAuthenticated();

  if (isAuthenticated) {
    // If the previous navigation carried a "from" location, go back there.
    const from = (location.state as { from?: Location })?.from?.pathname ?? '/';
    return <Navigate to={from} replace />;
  }

  return <Outlet />;
}
