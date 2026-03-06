/**
 * Session Monitor Hook
 * Monitors and enforces session security
 */

import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { useAdminAuthStore } from '@/stores/adminAuthStore';
import { SESSION } from '@/lib/constants';

export function useSessionMonitor() {
  const checkSession = useAdminAuthStore((state) => state.checkSession);
  const updateActivity = useAdminAuthStore((state) => state.updateActivity);
  const location = useLocation();

  useEffect(() => {
    // Check session validity every minute
    const sessionCheckInterval = setInterval(() => {
      checkSession();
    }, SESSION.CHECK_INTERVAL);

    return () => clearInterval(sessionCheckInterval);
  }, [checkSession]);

  useEffect(() => {
    // Update activity on user interaction
    const handleActivity = () => {
      updateActivity();
    };

    window.addEventListener('mousedown', handleActivity);
    window.addEventListener('keydown', handleActivity);
    window.addEventListener('scroll', handleActivity);
    window.addEventListener('touchstart', handleActivity);

    return () => {
      window.removeEventListener('mousedown', handleActivity);
      window.removeEventListener('keydown', handleActivity);
      window.removeEventListener('scroll', handleActivity);
      window.removeEventListener('touchstart', handleActivity);
    };
  }, [updateActivity]);

  useEffect(() => {
    // Reset idle timer on route change
    updateActivity();
  }, [location, updateActivity]);

  return { checkSession, updateActivity };
}
