/**
 * Restores admin session from sessionStorage on app load so refresh keeps the user logged in.
 */

import { useEffect, useState } from 'react';
import { useAdminAuthStore } from '@/stores/adminAuthStore';
import { adminApiClient } from '@/lib/api/adminClient';
import { bootstrapAdminSession } from '@/lib/api/adminAuth';
import { CACHE_KEYS } from '@/lib/constants';
import { LoadingScreen } from '@/components/common/LoadingScreen';

export function AdminAuthRestore({ children }: { children: React.ReactNode }) {
  const [restoring, setRestoring] = useState(true);
  const login = useAdminAuthStore((s) => s.login);
  const isAuthenticated = useAdminAuthStore((s) => s.isAuthenticated);

  useEffect(() => {
    const token = sessionStorage.getItem(CACHE_KEYS.ADMIN_ACCESS_TOKEN);
    if (!token) {
      setRestoring(false);
      return;
    }
    if (isAuthenticated) {
      setRestoring(false);
      return;
    }
    bootstrapAdminSession(token)
      .then((bootstrap) => {
        const accessToken = bootstrap.session.access_token ?? token;
        adminApiClient.setSessionToken(accessToken);
        login(bootstrap.session, bootstrap.user);
      })
      .catch(() => {
        sessionStorage.removeItem(CACHE_KEYS.ADMIN_ACCESS_TOKEN);
        adminApiClient.clearSessionToken();
      })
      .finally(() => setRestoring(false));
  }, [login, isAuthenticated]);

  if (restoring) {
    return <LoadingScreen />;
  }
  return <>{children}</>;
}
