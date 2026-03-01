/**
 * Sentry is initialized only when VITE_SENTRY_DSN is set (e.g. in production env).
 * Set VITE_SENTRY_DSN in your build env to enable error tracking.
 */
export function initSentry() {
  const dsn = import.meta.env.VITE_SENTRY_DSN
  if (!dsn || typeof dsn !== 'string') return

  import('@sentry/react').then((Sentry) => {
    Sentry.init({
      dsn,
      environment: import.meta.env.MODE,
      integrations: [
        Sentry.browserTracingIntegration(),
        Sentry.replayIntegration({ maskAllText: true, blockAllMedia: true }),
      ],
      tracesSampleRate: 0.1,
      replaysOnErrorSampleRate: 1.0,
      replaysSessionSampleRate: 0.1,
    })
  })
}
