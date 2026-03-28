import { useEffect } from 'react';
import { Metric, onCLS, onFCP, onINP, onLCP, onTTFB } from 'web-vitals';

interface WebVitalsMetrics {
  CLS?: number;
  INP?: number;
  FCP?: number;
  LCP?: number;
  TTFB?: number;
}

export function useWebVitals(onMetrics?: (metrics: WebVitalsMetrics) => void) {
  useEffect(() => {
    const metrics: WebVitalsMetrics = {};

    // Cumulative Layout Shift (CLS)
    onCLS((metric: Metric) => {
      metrics.CLS = metric.value;
      if (import.meta.env.DEV) {
        console.log('CLS:', metric);
      }
      onMetrics?.(metrics);

      // Send to Google Analytics 4
      if (typeof window !== 'undefined' && (window as any).gtag) {
        (window as any).gtag('event', 'web_vitals', {
          event_category: 'Web Vitals',
          event_label: 'CLS',
          value: Math.round(metric.value * 1000),
          custom_map: { metric_value: metric.value },
        });
      }
    });

    // Interaction to Next Paint (INP) - replaced FID
    onINP((metric: Metric) => {
      metrics.INP = metric.value;
      if (import.meta.env.DEV) {
        console.log('INP:', metric);
      }
      onMetrics?.(metrics);

      // Send to Google Analytics 4
      if (typeof window !== 'undefined' && (window as any).gtag) {
        (window as any).gtag('event', 'web_vitals', {
          event_category: 'Web Vitals',
          event_label: 'INP',
          value: Math.round(metric.value),
          custom_map: { metric_value: metric.value },
        });
      }
    });

    // First Contentful Paint (FCP)
    onFCP((metric: Metric) => {
      metrics.FCP = metric.value;
      if (import.meta.env.DEV) {
        console.log('FCP:', metric);
      }
      onMetrics?.(metrics);

      // Send to Google Analytics 4
      if (typeof window !== 'undefined' && (window as any).gtag) {
        (window as any).gtag('event', 'web_vitals', {
          event_category: 'Web Vitals',
          event_label: 'FCP',
          value: Math.round(metric.value),
          custom_map: { metric_value: metric.value },
        });
      }
    });

    // Largest Contentful Paint (LCP)
    onLCP((metric: Metric) => {
      metrics.LCP = metric.value;
      if (import.meta.env.DEV) {
        console.log('LCP:', metric);
      }
      onMetrics?.(metrics);

      // Send to Google Analytics 4
      if (typeof window !== 'undefined' && (window as any).gtag) {
        (window as any).gtag('event', 'web_vitals', {
          event_category: 'Web Vitals',
          event_label: 'LCP',
          value: Math.round(metric.value),
          custom_map: { metric_value: metric.value },
        });
      }
    });

    // Time to First Byte (TTFB)
    onTTFB((metric: Metric) => {
      metrics.TTFB = metric.value;
      if (import.meta.env.DEV) {
        console.log('TTFB:', metric);
      }
      onMetrics?.(metrics);

      // Send to Google Analytics 4
      if (typeof window !== 'undefined' && (window as any).gtag) {
        (window as any).gtag('event', 'web_vitals', {
          event_category: 'Web Vitals',
          event_label: 'TTFB',
          value: Math.round(metric.value),
          custom_map: { metric_value: metric.value },
        });
      }
    });
  }, [onMetrics]);
}

// Utility function for custom analytics integration
// This function can be used if you need to send metrics to a custom analytics service
export function reportWebVitalsToCustomAnalytics(metric: any) {
  fetch('/api/analytics/web-vitals', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(metric),
  }).catch((error) => {
    console.error('Failed to send web vitals to custom analytics:', error);
  });
}

/** Send aggregated Web Vitals to your analytics endpoint (e.g. in production). */
export function reportWebVitalsBatch(metrics: WebVitalsMetrics, context?: { page?: string }) {
  const payload = { ...metrics, ...(context?.page && { page: context.page }) };
  fetch('/api/analytics/web-vitals', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  }).catch(() => {
    // Silently ignore (endpoint may not be implemented yet)
  });
}
