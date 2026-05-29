import {
  useStudioAccountPreferences,
  useStudioLastWorkspace,
} from '@/lib/studio-account-preferences';
import { useEffect, useRef } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

const DEFAULT_STUDIO_PATHS = new Set([
  '/studio',
  '/studio/development',
  '/studio/staging',
  '/studio/production',
]);

/**
 * Persists the current Studio route for restore-on-launch and restores it once on startup.
 */
export function useStudioWorkspacePersistence() {
  const location = useLocation();
  const navigate = useNavigate();
  const { preferences, isLoading: prefsLoading } = useStudioAccountPreferences();
  const {
    lastWorkspace,
    isLoading: workspaceLoading,
    saveLastWorkspace,
  } = useStudioLastWorkspace();
  const restoredRef = useRef(false);

  useEffect(() => {
    if (prefsLoading || workspaceLoading || restoredRef.current) return;
    if (!preferences.restoreLastWorkspace || !lastWorkspace?.route) return;
    if (!DEFAULT_STUDIO_PATHS.has(location.pathname)) return;

    restoredRef.current = true;
    navigate(lastWorkspace.route, { replace: true });
  }, [
    prefsLoading,
    workspaceLoading,
    preferences.restoreLastWorkspace,
    lastWorkspace,
    location.pathname,
    navigate,
  ]);

  useEffect(() => {
    if (!location.pathname.startsWith('/studio')) return;

    const route = `${location.pathname}${location.search}`;
    const timer = window.setTimeout(() => {
      saveLastWorkspace(route);
    }, 750);

    return () => window.clearTimeout(timer);
  }, [location.pathname, location.search, saveLastWorkspace]);
}
