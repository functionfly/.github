import { useEffect } from 'react';
import { onCLS, onINP, onFCP, onLCP, onTTFB, Metric } from 'web-vitals';

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
      console.log('CLS:', metric);
      onMetrics?.(metrics);

      // Send to analytics (replace with your analytics service)
      if (typeof window !== 'undefined' && (window as any).gtag) {
        (window as any).gtag('event', 'web_vitals', {
          event_category: 'Web Vitals',
          event_label: 'CLS',
          value: Math.round(metric.value * 1000),
          custom_map: { metric_value: metric.value }
        });
      }
    });

    // Interaction to Next Paint (INP) - replaced FID
    onINP((metric: Metric) => {
      metrics.INP = metric.value;
      console.log('INP:', metric);
      onMetrics?.(metrics);
    });

    // First Contentful Paint (FCP)
    onFCP((metric: Metric) => {
      metrics.FCP = metric.value;
      console.log('FCP:', metric);
      onMetrics?.(metrics);
    });

    // Largest Contentful Paint (LCP)
    onLCP((metric: Metric) => {
      metrics.LCP = metric.value;
      console.log('LCP:', metric);
      onMetrics?.(metrics);

      // Send to analytics
      if (typeof window !== 'undefined' && (window as any).gtag) {
        (window as any).gtag('event', 'web_vitals', {
          event_category: 'Web Vitals',
          event_label: 'LCP',
          value: Math.round(metric.value),
          custom_map: { metric_value: metric.value }
        });
      }
    });

    // Time to First Byte (TTFB)
    onTTFB((metric: Metric) => {
      metrics.TTFB = metric.value;
      console.log('TTFB:', metric);
      onMetrics?.(metrics);
    });
  }, [onMetrics]);
}

// Utility function to send metrics to your analytics service
export function reportWebVitals(metric: any) {
  // Example: Send to Google Analytics 4
  if (typeof window !== 'undefined' && (window as any).gtag) {
    (window as any).gtag('event', 'web_vitals', {
      event_category: 'Web Vitals',
      event_label: metric.name,
      value: Math.round(metric.name === 'CLS' ? metric.value * 1000 : metric.value),
      custom_map: {
        metric_id: metric.id,
        metric_value: metric.value,
        metric_delta: metric.delta
      }
    });
  }

  // Example: Send to custom analytics endpoint
  /*
  fetch('/api/analytics/web-vitals', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(metric)
  });
  */
}