/**
 * Restores the admin session on app load so a page refresh keeps the user
 * logged in within the same tab.
 *
 * The token lives in sessionStorage (cleared on tab close) — the dashboard's
 * `localStorage` JWT is NOT read here. A previous version of this component
 * cross-read `localStorage.ffly_jwt` / `localStorage['ff-access-token']`,
 * which made the admin JWT stealable by any XSS on the marketing site, the
 * main dashboard, or any third-party script with a single tab. That was
 * removed; opening the admin dashboard in a new tab now requires a fresh
 * admin login (the right tradeoff for the surface area it protects).
 *
 * Routing responsibilities:
 *  - If the session is restored, ProtectedRoute (which reads from Zustand) will
 *    allow access to protected pages automatically — no explicit navigate() needed.
 *  - If the session cannot be restored, ProtectedRoute will redirect to /auth/login.
 *  - We only show a loading screen while the async restore is in flight so the
 *    router doesn't flash the wrong page before we know auth state.
 */

import { LoadingScreen } from '@/components/common/LoadingScreen';
import { logger } from '@/lib/monitoring/logger';
import { useAdminAuthStore } from '@/stores/adminAuthStore';
import { useEffect, useState } from 'react';

export function AdminAuthRestore({ children }: { children: React.ReactNode }) {
  const [restoring, setRestoring] = useState(true);
  const initialize = useAdminAuthStore((s) => s.initialize);

  useEffect(() => {
    const restoreSession = async () => {
      // If the store already has a valid session (e.g. same-tab navigation)
      // there is nothing to restore.
      if (useAdminAuthStore.getState().isAuthenticated) {
        setRestoring(false);
        return;
      }

      try {
        await initialize();
      } catch (error) {
        logger.warn('Failed to restore admin session', { error });
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
