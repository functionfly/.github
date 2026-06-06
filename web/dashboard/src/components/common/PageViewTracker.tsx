/**
 * PageViewTracker — fires trackPageView on every React Router route change.
 * Drop into App.tsx inside <BrowserRouter>.
 *
 * Only tracks when:
 *   1. A user is authenticated ( Mixpanel identify was called)
 *   2. Analytics consent has been granted
 */
import { trackPageView } from '@/lib/analytics';
import { useAuthStore } from '@/stores/authStore';
import { useCookieConsentStore } from '@/stores/cookieConsentStore';
import { useEffect, useRef } from 'react';
import { useLocation } from 'react-router-dom';

export function PageViewTracker() {
  const location = useLocation();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const hasAnalyticsConsent = useCookieConsentStore((s) => s.categories.analytics);
  const prevPathRef = useRef<string | null>(null);

  useEffect(() => {
    if (!isAuthenticated || !hasAnalyticsConsent) return;

    const path = location.pathname + location.search;
    // Deduplicate: don't fire twice for same path
    if (path === prevPathRef.current) return;
    prevPathRef.current = path;

    trackPageView(path);
  }, [location, isAuthenticated, hasAnalyticsConsent]);

  return null;
}
