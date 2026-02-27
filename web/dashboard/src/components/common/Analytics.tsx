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
        const token = localStorage.getItem('sb-access-token');
        if (!token) {
          // User not authenticated, use fallback settings
          const fallbackSettings: AnalyticsSettings = {
            googleAnalytics: {
              measurementId: import.meta.env.VITE_GOOGLE_ANALYTICS_ID || 'G-XXXXXXXXXX',
              enabled: import.meta.env.VITE_GOOGLE_ANALYTICS_ID && import.meta.env.VITE_GOOGLE_ANALYTICS_ID !== 'G-XXXXXXXXXX',
            },
            hotjar: {
              siteId: import.meta.env.VITE_HOTJAR_SITE_ID || '0000000',
              enabled: import.meta.env.VITE_HOTJAR_SITE_ID && import.meta.env.VITE_HOTJAR_SITE_ID !== '0000000',
            },
            services: [],
          };
          setSettings(fallbackSettings);
          setLoaded(true);
          return;
        }

        const analyticsSettings = await getAnalyticsSettings();
        setSettings(analyticsSettings);
      } catch (error) {
        console.error('Failed to load analytics settings:', error);
        // Fallback to environment variables if API fails or user doesn't have permissions
        const fallbackSettings: AnalyticsSettings = {
          googleAnalytics: {
            measurementId: import.meta.env.VITE_GOOGLE_ANALYTICS_ID || 'G-XXXXXXXXXX',
            enabled: import.meta.env.VITE_GOOGLE_ANALYTICS_ID && import.meta.env.VITE_GOOGLE_ANALYTICS_ID !== 'G-XXXXXXXXXX',
          },
          hotjar: {
            siteId: import.meta.env.VITE_HOTJAR_SITE_ID || '0000000',
            enabled: import.meta.env.VITE_HOTJAR_SITE_ID && import.meta.env.VITE_HOTJAR_SITE_ID !== '0000000',
          },
          services: [],
        };
        setSettings(fallbackSettings);
      } finally {
        setLoaded(true);
      }
    };

    loadAnalyticsSettings();
  }, []);

  if (!loaded || !settings) {
    return <VercelAnalytics />;
  }

  return (
    <>
      <VercelAnalytics />
      {settings.googleAnalytics?.enabled &&
       settings.googleAnalytics.measurementId &&
       settings.googleAnalytics.measurementId !== 'G-XXXXXXXXXX' &&
       loadGoogleAnalytics(settings.googleAnalytics.measurementId)}
      {settings.hotjar?.enabled &&
       settings.hotjar.siteId &&
       settings.hotjar.siteId !== '0000000' &&
       loadHotjar(settings.hotjar.siteId)}
    </>
  );
}