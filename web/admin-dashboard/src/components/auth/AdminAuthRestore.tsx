/**
 * Restores admin session from sessionStorage on app load so a page refresh keeps
 * the user logged in.  Also picks up JWT tokens written by the main dashboard login.
 *
 * Routing responsibilities:
 *  - If the session is restored, ProtectedRoute (which reads from Zustand) will
 *    allow access to protected pages automatically — no explicit navigate() needed.
 *  - If the session cannot be restored, ProtectedRoute will redirect to /auth/login.
 *  - We only show a loading screen while the async restore is in flight so the
 *    router doesn't flash the wrong page before we know auth state.
 */

import { useEffect, useState } from 'react';
import { useAdminAuthStore } from '@/stores/adminAuthStore';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { bootstrapAdminSession } from '@/lib/api/adminAuth';
import { CACHE_KEYS } from '@/lib/constants';

export function AdminAuthRestore({ children }: { children: React.ReactNode }) {
  const [restoring, setRestoring] = useState(true);
  const initialize = useAdminAuthStore((s) => s.initialize);
  const login = useAdminAuthStore((s) => s.login);

  useEffect(() => {
    const restoreSession = async () => {
      // If the store already has a valid session (e.g. same-tab navigation)
      // there is nothing to restore.
      if (useAdminAuthStore.getState().isAuthenticated) {
        setRestoring(false);
        return;
      }

      try {
        // 1. Try to restore from sessionStorage (handled inside initialize())
        await initialize();

        // 2. If still not authenticated, try a JWT written by the main dashboard
        if (!useAdminAuthStore.getState().isAuthenticated) {
          const jwtToken =
            localStorage.getItem('ffly_jwt') || localStorage.getItem('ff-access-token');
          if (jwtToken) {
            try {
              const bootstrap = await bootstrapAdminSession(jwtToken);
              if (bootstrap.session && bootstrap.user) {
                try {
                  sessionStorage.setItem(CACHE_KEYS.ADMIN_ACCESS_TOKEN, jwtToken);
                } catch {
                  /* sessionStorage may be unavailable */
                }
                login(bootstrap.session, bootstrap.user);
              }
            } catch (error) {
              console.warn('Failed to bootstrap admin session from JWT:', error);
              localStorage.removeItem('ffly_jwt');
              localStorage.removeItem('ff-access-token');
            }
          }
        }
      } catch (error) {
        console.warn('Failed to restore admin session:', error);
      } finally {
        setRestoring(false);
      }
    };

    restoreSession();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (restoring) {
    return <LoadingScreen />;
  }
  return <>{children}</>;
}
