/**
 * Admin Auth Hook
 * Provides authentication utilities for admin components
 */

import { useAdminAuthStore } from '@/stores/adminAuthStore';
import { useCallback } from 'react';

export function useAdminAuth() {
  const { user, session, isAuthenticated, mfaVerified, lastActivity } =
    useAdminAuthStore();
  const { login, logout, verifyMFA, updateActivity, checkSession } =
    useAdminAuthStore();

  const handleLogout = useCallback(() => {
    logout();
  }, [logout]);

  const handleVerifyMFA = useCallback(
    async (code: string) => {
      return verifyMFA(code);
    },
    [verifyMFA]
  );

  const handleUpdateActivity = useCallback(() => {
    updateActivity();
  }, [updateActivity]);

  const handleCheckSession = useCallback(() => {
    return checkSession();
  }, [checkSession]);

  const isSessionValid = useCallback(() => {
    return checkSession();
  }, [checkSession]);

  return {
    user,
    session,
    isAuthenticated,
    mfaVerified,
    lastActivity,
    login,
    logout: handleLogout,
    verifyMFA: handleVerifyMFA,
    updateActivity: handleUpdateActivity,
    checkSession: handleCheckSession,
    isSessionValid,
  };
}
