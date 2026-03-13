import { useEffect, useState } from 'react';
import { Analytics as VercelAnalytics } from '@vercel/analytics/react';
import { loadGoogleAnalytics } from '@/components/cookie-consent/ConditionalScriptLoader';
import { getAnalyticsSettings } from '@/api';
import type { AnalyticsSettings } from '@/types';
import { useAuthStore } from '@/stores/authStore';

const ADMIN_ROLES = ['admin', 'super_admin'];

export function Analytics() {
  const [settings, setSettings] = useState<AnalyticsSettings | null>(null);
  const [loaded, setLoaded] = useState(false);
  const user = useAuthStore((s) => s.user);

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

  if (!loaded) {
    return <VercelAnalytics />;
  }

  // Only load additional analytics for authenticated users with custom settings
  if (!settings) {
    return <VercelAnalytics />;
  }

  return (
    <>
      <VercelAnalytics />
      {settings.googleAnalytics?.enabled &&
       settings.googleAnalytics.measurementId &&
       settings.googleAnalytics.measurementId !== 'G-XXXXXXXXXX' &&
       loadGoogleAnalytics(settings.googleAnalytics.measurementId)}
      {/* Note: Hotjar is handled by PublicAnalytics for public pages */}
    </>
  );
}
