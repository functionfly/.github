import React, { Component, ErrorInfo, ReactNode } from 'react';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
}

interface State {
  hasError: boolean;
  error?: Error;
  validationErrors: string[];
}

export class ValidationErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      hasError: false,
      validationErrors: []
    };
  }

  static getDerivedStateFromError(error: Error): State {
    // Check if this is a validation-related error
    const isValidationError = error.message.includes('validation') ||
                             error.message.includes('schema') ||
                             error.name === 'ZodError';

    if (isValidationError) {
      return {
        hasError: true,
        error,
        validationErrors: [error.message]
      };
    }

    // Re-throw non-validation errors
    throw error;
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // Only handle validation errors here
    if (error.message.includes('validation') ||
        error.message.includes('schema') ||
        error.name === 'ZodError') {

      console.error('Validation Error Boundary caught an error:', error, errorInfo);

      this.props.onError?.(error, errorInfo);
    } else {
      // Re-throw non-validation errors
      throw error;
    }
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <div className="rounded-lg border border-red-200 bg-red-50 p-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">
                Validation Error
              </h3>
              <div className="mt-2 text-sm text-red-700">
                <ul role="list" className="list-disc pl-5 space-y-1">
                  {this.state.validationErrors.map((error, index) => (
                    <li key={index}>{error}</li>
                  ))}
                </ul>
              </div>
              <div className="mt-4">
                <div className="-mx-2 -my-1.5 flex">
                  <button
                    type="button"
                    className="bg-red-50 px-2 py-1.5 rounded-md text-sm font-medium text-red-800 hover:bg-red-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-red-50 focus:ring-red-600"
                    onClick={() => this.setState({ hasError: false, validationErrors: [] })}
                  >
                    Try Again
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

// Hook for handling validation errors in functional components
export function useValidationErrorHandler() {
  return (error: Error) => {
    // Log validation errors for monitoring
    if (error.message.includes('validation') ||
        error.message.includes('schema') ||
        error.name === 'ZodError') {
      console.error('Validation error handled:', error);

      // In a real app, you might want to send this to an error reporting service
      // reportError(error, { type: 'validation' });
    }
  };
}

// Higher-order component for validation error handling
export function withValidationErrorHandler<P extends object>(
  Component: React.ComponentType<P>,
  fallback?: ReactNode
) {
  const WrappedComponent = (props: P) => (
    <ValidationErrorBoundary fallback={fallback}>
      <Component {...props} />
    </ValidationErrorBoundary>
  );

  WrappedComponent.displayName = `withValidationErrorHandler(${Component.displayName || Component.name})`;

  return WrappedComponent;
}