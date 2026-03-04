import { useEffect } from 'react';
import { loadHotjar } from '@/components/cookie-consent/ConditionalScriptLoader';

/**
 * PublicAnalytics Component
 *
 * Handles analytics for public pages (non-authenticated users).
 * Includes Hotjar for user behavior analysis and session recording.
 */
export function PublicAnalytics() {
  useEffect(() => {
    // Load Hotjar for public pages if site ID is configured
    const hotjarSiteId = import.meta.env.VITE_HOTJAR_SITE_ID;

    if (hotjarSiteId && hotjarSiteId !== '0000000') {
      // Load Hotjar immediately for public pages
      const script = document.createElement('script');
      script.innerHTML = `
        (function(h,o,t,j,a,r){
            h.hj=h.hj||function(){(h.hj.q=h.hj.q||[]).push(arguments)};
            h._hjSettings={hjid:${hotjarSiteId},hjsv:6};
            a=o.getElementsByTagName('head')[0];
            r=o.createElement('script');r.async=1;
            r.src=t+h._hjSettings.hjid+j+h._hjSettings.hjsv;
            a.appendChild(r);
        })(window,document,'https://static.hotjar.com/c/hotjar-','.js?sv=');
      `;
      document.head.appendChild(script);

      console.log('Hotjar loaded for public page');
    }
  }, []);

  // This component doesn't render anything visible
  return null;
}