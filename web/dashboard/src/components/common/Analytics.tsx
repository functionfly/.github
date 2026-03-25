import { getAnalyticsSettings } from '@/api';
import { loadGoogleAnalytics } from '@/components/cookie-consent/ConditionalScriptLoader';
import { COMING_SOON_ONLY } from '@/lib/constants';
import { useAuthStore } from '@/stores/authStore';
import type { AnalyticsSettings } from '@/types';
import { Analytics as VercelAnalytics } from '@vercel/analytics/react';
import { SpeedInsights } from '@vercel/speed-insights/react';
import { useEffect, useState } from 'react';
import { useLocation } from 'react-router-dom';

const ADMIN_ROLES = ['admin', 'super_admin'];

/**
 * Only load Vercel Analytics and Speed Insights when deployed on Vercel (avoids 404/MIME errors on Cloudflare Pages).
 * Speed Insights on the coming-soon/launch page is rendered by LaunchPage to avoid duplicate scripts.
 */
const useVercelInsights = () => import.meta.env.VITE_VERCEL_ANALYTICS === 'true';

/** When true, only LaunchPage is shown so it owns Speed Insights; we skip it here. */
const LAUNCH_PAGE_PATHS = ['/coming-soon', '/launch'];

export function Analytics() {
  const [settings, setSettings] = useState<AnalyticsSettings | null>(null);
  const [loaded, setLoaded] = useState(false);
  const user = useAuthStore((s) => s.user);
  const enableVercelInsights = useVercelInsights();
  const location = useLocation();
  const isLaunchPage =
    COMING_SOON_ONLY ||
    LAUNCH_PAGE_PATHS.some((p) => location.pathname === p || location.pathname.startsWith(p + '/'));

  useEffect(() => {
    const loadAnalyticsSettings = async () => {
      try {
        const token = localStorage.getItem('ff-access-token');
        if (!token) {
          setLoaded(true);
          return;
        }
        // Admin analytics settings (Google Analytics, etc.) are only available to admins
        const isAdmin = user?.role && ADMIN_ROLES.includes(user.role);
        if (!isAdmin) {
          setLoaded(true);
          return;
        }

        const analyticsSettings = await getAnalyticsSettings();
        setSettings(analyticsSettings);
      } catch (error) {
        // 403 expected for non-admins; only log for admins
        if (user?.role && ADMIN_ROLES.includes(user.role)) {
          console.error('Failed to load analytics settings:', error);
        }
      } finally {
        setLoaded(true);
      }
    };

    loadAnalyticsSettings();
  }, [user?.role]);

  const vercelNode = enableVercelInsights ? (
    <>
      <VercelAnalytics />
      {!isLaunchPage && <SpeedInsights />}
    </>
  ) : null;

  if (!loaded) {
    return <>{vercelNode}</>;
  }

  // Only load additional analytics for authenticated users with custom settings
  if (!settings) {
    return <>{vercelNode}</>;
  }

  return (
    <>
      {vercelNode}
      {settings.googleAnalytics?.enabled &&
        settings.googleAnalytics.measurementId &&
        settings.googleAnalytics.measurementId !== 'G-XXXXXXXXXX' &&
        loadGoogleAnalytics(settings.googleAnalytics.measurementId)}
      {/* Note: Hotjar is handled by PublicAnalytics for public pages */}
    </>
  );
}
