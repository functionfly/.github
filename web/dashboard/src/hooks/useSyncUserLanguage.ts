import { useAuthStore } from '@/stores/authStore';
import { useEffect } from 'react';
import { useTranslation } from 'react-i18next';

/**
 * Syncs i18next language to the user's stored preference whenever:
 * 1. Auth state is first checked (initial load)
 * 2. The user logs in (auth becomes true)
 * 3. The user's language field changes
 *
 * Priority: user's backend language > localStorage persisted language > browser auto-detect
 */
export function useSyncUserLanguage() {
  const { i18n } = useTranslation();
  const user = useAuthStore((s) => s.user);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const authChecked = useAuthStore((s) => s.authChecked);

  useEffect(() => {
    if (!authChecked) return;

    // Priority 1: Use the user's stored backend language if available
    if (user?.language) {
      if (i18n.language !== user.language) {
        void i18n.changeLanguage(user.language);
      }
    }
  }, [authChecked, isAuthenticated, user?.language, i18n.language]);
}
