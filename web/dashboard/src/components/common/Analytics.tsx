import { useEffect, useState } from 'react';
import { Analytics as VercelAnalytics } from '@vercel/analytics/react';
import { loadGoogleAnalytics, loadHotjar } from '@/components/cookie-consent/ConditionalScriptLoader';
import { getAnalyticsSettings } from '@/api';
import type { AnalyticsSettings } from '@/types';

export function Analytics() {
  const [settings, setSettings] = useState<AnalyticsSettings | null>(null);
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    const loadAnalyticsSettings = async () => {
      try {
        // Only try to load analytics settings if user is authenticated
        const token = localStorage.getItem('ff-access-token');
        if (!token) {
          // User not authenticated, skip loading analytics settings
          setLoaded(true);
          return;
        }

        const analyticsSettings = await getAnalyticsSettings();
        setSettings(analyticsSettings);
      } catch (error) {
        console.error('Failed to load analytics settings:', error);
        // Don't set fallback settings for authenticated users - they'll get default behavior
      } finally {
        setLoaded(true);
      }
    };

    loadAnalyticsSettings();
  }, []);

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