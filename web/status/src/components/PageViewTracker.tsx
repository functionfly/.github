/**
 * PageViewTracker — fires trackPageView on every React Router route change.
 * Drop into status App.tsx inside <BrowserRouter>.
 */
import { trackPageView } from '@/lib/analytics';
import { useEffect, useRef } from 'react';
import { useLocation } from 'react-router-dom';

export function PageViewTracker() {
  const location = useLocation();
  const prevPathRef = useRef<string | null>(null);

  useEffect(() => {
    const path = location.pathname + location.search;
    if (path === prevPathRef.current) return;
    prevPathRef.current = path;
    trackPageView(path);
  }, [location]);

  return null;
}
