'use client';

import { Button } from '@/components/ui/button';
import { AlertTriangle, Home, Loader2, RefreshCcw } from 'lucide-react';
import React, { Suspense } from 'react';

export type ErrorReportFn = (error: Error, errorInfo: React.ErrorInfo) => void;

interface Props {
  children: React.ReactNode;
  fallback?: React.ReactNode;
  /** Called in production when an error is caught; use to send to Sentry, LogRocket, etc. */
  onError?: ErrorReportFn;
}

interface State {
  hasError: boolean;
  error: Error | null;
  errorInfo: React.ErrorInfo | null;
}

/** Capture error to Sentry if available */
async function captureErrorToSentry(error: Error, errorInfo: React.ErrorInfo) {
  if (typeof window === 'undefined') return;
  try {
    const Sentry = await import('@sentry/react');
    Sentry.captureException(error, {
      contexts: {
        react: {
          componentStack: errorInfo.componentStack,
        },
      },
    });
  } catch {
    // Sentry not available or failed to load
  }
}

export class ErrorBoundary extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null, errorInfo: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error, errorInfo: null };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    // Log error to console in development
    if (import.meta.env.DEV) {
      console.error('Error caught by boundary:', error);
      console.error('Component stack:', errorInfo.componentStack);
    }

    this.setState({
      error,
      errorInfo,
    });

    // Capture to Sentry in production
    if (import.meta.env.PROD) {
      captureErrorToSentry(error, errorInfo);
      if (this.props.onError) {
        this.props.onError(error, errorInfo);
      }
    }
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: null, errorInfo: null });
  };

  render() {
    if (this.state.hasError) {
      // Custom fallback UI
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <div className="min-h-screen flex items-center justify-center px-4" style={{ backgroundColor: 'var(--bg-primary)' }}>
          <div className="max-w-md w-full text-center space-y-6" style={{ color: 'var(--text-primary)' }}>
            {/* Error Icon */}
            <div className="inline-flex items-center justify-center w-20 h-20 rounded-2xl mb-4" style={{ backgroundColor: 'rgba(239, 68, 68, 0.1)' }}>
              <AlertTriangle className="w-10 h-10" style={{ color: 'var(--color-error, #ef4444)' }} />
            </div>

            {/* Error Message */}
            <div className="space-y-2">
              <h1 className="text-2xl font-bold">Something went wrong</h1>
              <p style={{ color: 'var(--text-secondary)' }}>
                We apologize for the inconvenience. An unexpected error has occurred.
              </p>
            </div>

            {/* Error Details (only in development) */}
            {import.meta.env.DEV && this.state.error && (
              <div className="rounded-lg p-4 text-left" style={{ backgroundColor: 'rgba(239, 68, 68, 0.1)', borderColor: 'rgba(239, 68, 68, 0.2)' }}>
                <p className="font-mono text-sm mb-2" style={{ color: 'var(--color-error, #ef4444)' }}>{this.state.error.message}</p>
                {this.state.errorInfo && (
                  <pre className="font-mono text-xs overflow-auto max-h-32 whitespace-pre-wrap" style={{ color: 'var(--color-error, #ef4444)', opacity: 0.7 }}>
                    {this.state.errorInfo.componentStack}
                  </pre>
                )}
              </div>
            )}

            {/* Action Buttons - use <a> not <Link> because ErrorBoundary renders outside BrowserRouter */}
            <div className="flex flex-col sm:flex-row gap-3 justify-center">
              <button onClick={this.handleRetry} className="flex-1 flex items-center justify-center gap-2 h-10 px-4 rounded-lg font-medium text-white transition-all hover:brightness-110 active:scale-[0.98]" style={{ background: 'linear-gradient(135deg, #ef4444 0%, #dc2626 100%)' }}>
                <RefreshCcw className="w-4 h-4" />
                Try Again
              </button>
              <a href="/" className="flex-1 flex items-center justify-center gap-2 h-10 px-4 rounded-lg font-medium transition-all" style={{ backgroundColor: 'var(--bg-secondary)', border: '1px solid var(--border-default)', color: 'var(--text-primary)' }}>
                  <Home className="w-4 h-4" />
                  Go Home
                </a>
            </div>

            {/* Support Link */}
            <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
              If the problem persists, please{' '}
              <a href="/contact" style={{ color: 'var(--color-brand-500, #f97316)' }}>
                contact support
              </a>
            </p>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

// Hook for functional components to catch errors in children
export function useErrorBoundary() {
  const [error, setError] = React.useState<Error | null>(null);

  if (error) {
    throw error;
  }

  return setError;
}

/**
 * SectionErrorBoundary - A smaller error boundary for independent dashboard sections
 * Renders an inline error message instead of full-page crash
 */
interface SectionErrorBoundaryProps {
  children: React.ReactNode;
  /** Name of the section for error reporting */
  sectionName?: string;
  /** Custom fallback UI */
  fallback?: React.ReactNode;
  /** Called when error is caught */
  onError?: ErrorReportFn;
}

interface SectionErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
  errorInfo: React.ErrorInfo | null;
}

export class SectionErrorBoundary extends React.Component<SectionErrorBoundaryProps, SectionErrorBoundaryState> {
  constructor(props: SectionErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null, errorInfo: null };
  }

  static getDerivedStateFromError(error: Error): SectionErrorBoundaryState {
    return { hasError: true, error, errorInfo: null };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    if (import.meta.env.DEV) {
      console.error(`Error in section "${this.props.sectionName}":`, error);
    }

    this.setState({ error, errorInfo });

    if (import.meta.env.PROD) {
      captureErrorToSentry(error, errorInfo);
      if (this.props.onError) {
        this.props.onError(error, errorInfo);
      }
    }
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: null, errorInfo: null });
  };

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <div className="flex items-center justify-center p-6 rounded-lg border border-destructive/20 bg-destructive/5 min-h-[100px]">
          <div className="text-center space-y-3">
            <AlertTriangle className="h-6 w-6 mx-auto text-destructive" />
            <p className="text-sm font-medium text-foreground">
              {this.props.sectionName ? `${this.props.sectionName} couldn't load` : 'This section couldn\'t load'}
            </p>
            <div className="flex gap-2 justify-center">
              <button
                onClick={this.handleRetry}
                className="text-xs px-3 py-1.5 rounded-md bg-destructive/10 text-destructive hover:bg-destructive/20 transition-colors"
              >
                Try Again
              </button>
            </div>
            {import.meta.env.DEV && this.state.error && (
              <p className="text-xs text-destructive/70 font-mono mt-2">
                {this.state.error.message}
              </p>
            )}
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

/**
 * PageWrapper - Combines Suspense and ErrorBoundary for route-level code splitting
 * Use this to wrap lazy-loaded page components
 */
interface PageWrapperProps {
  children: React.ReactNode;
  /** Custom fallback for Suspense (loading state) */
  suspenseFallback?: React.ReactNode;
  /** Custom fallback for ErrorBoundary (error state) */
  errorFallback?: React.ReactNode;
  /** Called when error is caught */
  onError?: ErrorReportFn;
}

const defaultPageErrorFallback = (
  <div className="min-h-[40vh] flex items-center justify-center">
    <div className="text-center space-y-4">
      <p className="text-lg font-medium">This page encountered an error</p>
      <p className="text-sm text-muted-foreground">Try refreshing the page</p>
    </div>
  </div>
);

export function PageWrapper({
  children,
  suspenseFallback = (
    <div className="flex min-h-[40vh] items-center justify-center">
      <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
    </div>
  ),
  errorFallback = defaultPageErrorFallback,
  onError,
}: PageWrapperProps) {
  return (
    <ErrorBoundary fallback={errorFallback} onError={onError}>
      <Suspense fallback={suspenseFallback}>
        {children}
      </Suspense>
    </ErrorBoundary>
  );
}
