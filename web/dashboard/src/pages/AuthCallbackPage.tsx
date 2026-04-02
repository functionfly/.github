/**
 * Auth Callback Page
 *
 * Handles the callback from the standalone auth site (@web/auth).
 * The auth site stores tokens in sessionStorage and redirects here.
 * This component extracts the tokens, migrates them to localStorage,
 * and initializes the auth session.
 */

import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useAuthStore } from '@/stores/authStore';
import {
  migrateTokensFromSessionStorage,
  buildAuthSiteLoginUrl,
} from '@/lib/auth-integration';
import { Loader2, AlertCircle, CheckCircle } from 'lucide-react';
import { logger } from '@/lib/logger';

/**
 * Auth callback page - receives users returning from auth.functionfly.com
 *
 * Query params:
 * - redirect: Final destination path after successful auth (e.g., "/functions")
 * - error: Error message if auth failed
 * - new_user: "true" if this was a new signup
 */
export function AuthCallbackPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const initialize = useAuthStore((state) => state.initialize);

  const [status, setStatus] = useState<'processing' | 'success' | 'error'>('processing');
  const [errorMessage, setErrorMessage] = useState<string>('');

  // Get query parameters
  const redirectPath = searchParams.get('redirect') || '/overview';
  const errorParam = searchParams.get('error');
  const errorDescription = searchParams.get('error_description');
  const isNewUser = searchParams.get('new_user') === 'true';

  useEffect(() => {
    const processAuthCallback = async () => {
      try {
        // Handle errors from auth site
        if (errorParam) {
          logger.error('Auth callback error:', { error: errorParam, description: errorDescription });
          setStatus('error');
          setErrorMessage(errorDescription || errorParam || 'Authentication failed');
          return;
        }

        // Migrate tokens from sessionStorage (set by auth site) to localStorage
        const { accessToken, source } = migrateTokensFromSessionStorage();

        if (!accessToken) {
          logger.error('No access token found in sessionStorage or localStorage');
          setStatus('error');
          setErrorMessage('No authentication token found. Please try logging in again.');
          return;
        }

        logger.info(`Migrated tokens from ${source}, initializing session...`);

        // Initialize the auth store with the new tokens
        await initialize();

        // Check if auth was successful
        const authState = useAuthStore.getState();

        if (authState.isAuthenticated) {
          setStatus('success');

          // For new users, redirect to onboarding
          if (isNewUser) {
            logger.info('New user signup detected, redirecting to onboarding');
            navigate('/onboarding', { replace: true });
            return;
          }

          // Navigate to the intended destination
          // Ensure the path is valid (starts with /)
          const destination = redirectPath.startsWith('/') ? redirectPath : `/${redirectPath}`;
          logger.info('Auth successful, redirecting to:', destination);
          navigate(destination, { replace: true });
        } else {
          // Auth initialization failed
          setStatus('error');
          setErrorMessage('Failed to initialize session. Please try logging in again.');
        }
      } catch (error) {
        logger.error('Error processing auth callback:', error);
        setStatus('error');
        setErrorMessage(error instanceof Error ? error.message : 'An unexpected error occurred');
      }
    };

    processAuthCallback();
  }, [errorParam, errorDescription, redirectPath, isNewUser, initialize, navigate]);

  // Handle "Try Again" click
  const handleTryAgain = () => {
    window.location.href = buildAuthSiteLoginUrl(redirectPath);
  };

  // Handle "Go to Login" click
  const handleGoToLogin = () => {
    window.location.href = buildAuthSiteLoginUrl();
  };

  return (
    <div className="min-h-screen bg-bg-primary flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-bg-secondary border border-border-subtle rounded-xl p-8 text-center">
        {status === 'processing' && (
          <>
            <div className="flex justify-center mb-4">
              <Loader2 className="h-10 w-10 text-brand-500 animate-spin" />
            </div>
            <h1 className="text-xl font-semibold text-text-primary mb-2">
              {isNewUser ? 'Welcome to FunctionFly!' : 'Completing sign in...'}
            </h1>
            <p className="text-text-secondary">
              {isNewUser
                ? 'Setting up your new account...'
                : 'Please wait while we sign you in...'}
            </p>
          </>
        )}

        {status === 'success' && (
          <>
            <div className="flex justify-center mb-4">
              <CheckCircle className="h-10 w-10 text-emerald-500" />
            </div>
            <h1 className="text-xl font-semibold text-text-primary mb-2">
              {isNewUser ? 'Account created!' : 'Signed in!'}
            </h1>
            <p className="text-text-secondary">Redirecting to your dashboard...</p>
          </>
        )}

        {status === 'error' && (
          <>
            <div className="flex justify-center mb-4">
              <AlertCircle className="h-10 w-10 text-error" />
            </div>
            <h1 className="text-xl font-semibold text-text-primary mb-2">
              Authentication failed
            </h1>
            <p className="text-text-secondary mb-6">{errorMessage}</p>
            <div className="flex flex-col gap-3">
              <button
                onClick={handleTryAgain}
                className="w-full px-4 py-2 bg-brand-600 hover:bg-brand-500 text-white font-medium rounded-lg transition-colors"
              >
                Try Again
              </button>
              <button
                onClick={handleGoToLogin}
                className="w-full px-4 py-2 bg-transparent border border-border-subtle hover:bg-bg-tertiary text-text-primary font-medium rounded-lg transition-colors"
              >
                Go to Login
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
