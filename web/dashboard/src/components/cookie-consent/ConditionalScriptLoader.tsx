'use client';

import { useEffect, useState } from 'react';
import { useCookieConsent } from '@/hooks/useCookieConsent';

interface ConditionalScriptLoaderProps {
  category: 'analytics' | 'marketing' | 'functionality';
  src?: string;
  children?: React.ReactNode;
  onLoad?: () => void;
  onError?: (error: Error) => void;
  id?: string;
}

/**
 * ConditionalScriptLoader Component
 *
 * Conditionally loads scripts based on cookie consent categories.
 * Can load external scripts via src prop or execute inline scripts via children.
 */
export function ConditionalScriptLoader({
  category,
  src,
  children,
  onLoad,
  onError,
  id,
}: ConditionalScriptLoaderProps) {
  const { canUseCategory, hasConsent } = useCookieConsent();
  const [scriptLoaded, setScriptLoaded] = useState(false);

  useEffect(() => {
    // Only load if user has given consent for this category
    if (!hasConsent || !canUseCategory(category)) {
      return;
    }

    // If we have a src, load external script
    if (src && !scriptLoaded) {
      const script = document.createElement('script');
      script.src = src;
      script.async = true;

      if (id) {
        script.id = id;
      }

      script.onload = () => {
        setScriptLoaded(true);
        onLoad?.();
      };

      script.onerror = (error) => {
        console.error(`Failed to load script: ${src}`, error);
        onError?.(new Error(`Failed to load script: ${src}`));
      };

      document.head.appendChild(script);

      return () => {
        // Cleanup function to remove script if component unmounts
        const existingScript = document.getElementById(id || '');
        if (existingScript) {
          document.head.removeChild(existingScript);
        }
      };
    }

    // If we have children (inline script), execute them
    if (children && !scriptLoaded && typeof children === 'string') {
      try {
        // eslint-disable-next-line no-eval
        eval(children);
        setScriptLoaded(true);
        onLoad?.();
      } catch (error) {
        console.error('Failed to execute inline script:', error);
        onError?.(error as Error);
      }
    }
  }, [hasConsent, canUseCategory(category), src, children, scriptLoaded, onLoad, onError, id]);

  // This component doesn't render anything visible
  return null;
}

// Utility function for loading Google Analytics
export function loadGoogleAnalytics(trackingId: string) {
  return (
    <ConditionalScriptLoader
      category="analytics"
      id="google-analytics"
      onLoad={() => {
        console.log('Google Analytics loaded successfully');
        // Initialize gtag if available
        if (window.gtag) {
          window.gtag('config', trackingId);
        }
      }}
    >
      {`
        (function() {
          var script = document.createElement('script');
          script.async = true;
          script.src = 'https://www.googletagmanager.com/gtag/js?id=${trackingId}';
          document.head.appendChild(script);

          window.dataLayer = window.dataLayer || [];
          function gtag(){dataLayer.push(arguments);}
          gtag('js', new Date());
          gtag('config', '${trackingId}');
        })();
      `}
    </ConditionalScriptLoader>
  );
}

// Utility function for loading Hotjar
export function loadHotjar(siteId: string) {
  return (
    <ConditionalScriptLoader
      category="analytics"
      id="hotjar-analytics"
      onLoad={() => {
        console.log('Hotjar loaded successfully');
      }}
    >
      {`
        (function(h,o,t,j,a,r){
            h.hj=h.hj||function(){(h.hj.q=h.hj.q||[]).push(arguments)};
            h._hjSettings={hjid:${siteId},hjsv:6};
            a=o.getElementsByTagName('head')[0];
            r=o.createElement('script');r.async=1;
            r.src=t+h._hjSettings.hjid+j+h._hjSettings.hjsv;
            a.appendChild(r);
        })(window,document,'https://static.hotjar.com/c/hotjar-','.js?sv=');
      `}
    </ConditionalScriptLoader>
  );
}