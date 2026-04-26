'use client';

import { Button } from '@/components/ui/button';
import { AlertTriangle, Home, RefreshCcw } from 'lucide-react';
import React from 'react';

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
        <div className="min-h-screen flex items-center justify-center bg-bg-primary px-4">
          <div className="max-w-md w-full text-center space-y-6">
            {/* Error Icon */}
            <div className="inline-flex items-center justify-center w-20 h-20 rounded-2xl bg-red-500/10 mb-4">
              <AlertTriangle className="w-10 h-10 text-red-500" />
            </div>

            {/* Error Message */}
            <div className="space-y-2">
              <h1 className="text-2xl font-bold text-text-primary">Something went wrong</h1>
              <p className="text-text-secondary">
                We apologize for the inconvenience. An unexpected error has occurred.
              </p>
            </div>

            {/* Error Details (only in development) */}
            {import.meta.env.DEV && this.state.error && (
              <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4 text-left">
                <p className="text-red-400 font-mono text-sm mb-2">{this.state.error.message}</p>
                {this.state.errorInfo && (
                  <pre className="text-red-400/70 font-mono text-xs overflow-auto max-h-32 whitespace-pre-wrap">
                    {this.state.errorInfo.componentStack}
                  </pre>
                )}
              </div>
            )}

            {/* Action Buttons - use <a> not <Link> because ErrorBoundary renders outside BrowserRouter */}
            <div className="flex flex-col sm:flex-row gap-3 justify-center">
              <Button onClick={this.handleRetry} className="gap-2 flex-1">
                <RefreshCcw className="w-4 h-4" />
                Try Again
              </Button>
              <a href="/" className="flex-1">
                <Button variant="outline" className="gap-2 w-full">
                  <Home className="w-4 h-4" />
                  Go Home
                </Button>
              </a>
            </div>

            {/* Support Link */}
            <p className="text-sm text-text-muted">
              If the problem persists, please{' '}
              <a href="/contact" className="text-brand-500 hover:underline">
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
