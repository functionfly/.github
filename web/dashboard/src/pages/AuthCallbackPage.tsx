/**
 * Auth Callback Page
 *
 * Handles callbacks from:
 * 1. OAuth provider flow via orchestrator: /auth/oauth/callback?token=xxx&refresh_token=yyy
 * 2. Standalone auth site (@web/auth): /auth/callback#token=xxx&refresh_token=yyy
 *
 * Security: Tokens are stored encrypted via TokenVault to prevent XSS exfiltration.
 */

import { buildAuthSiteLoginUrl } from '@/lib/auth-integration';
import { logger } from '@/lib/logger';
import { tokenVault } from '@/utils/token-vault';
import { useAuthStore } from '@/stores/authStore';
import { AlertCircle, CheckCircle, Loader2 } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';

/**
 * Extract token from URL — checks query params first, then fragment, then sessionStorage.
 * URLSearchParams.get() already decodes values — do NOT call decodeURIComponent on them.
 */
function getToken(key: string): string | null {
  // 1. Query params (primary — standard OAuth pattern)
  const url = new URL(window.location.href);
  const fromQuery = url.searchParams.get(key);
  if (fromQuery) return fromQuery;

  // 2. Fragment (fallback — older auth site builds)
  const hash = window.location.hash;
  if (hash && hash.length > 1) {
    const hashParams = new URLSearchParams(hash.substring(1));
    const fromHash = hashParams.get(key);
    if (fromHash) return fromHash;
  }

  // 3. sessionStorage (legacy — same-origin flows only)
  if (key === 'token') return sessionStorage.getItem('ff_token');
  if (key === 'refresh_token') return sessionStorage.getItem('ff_refresh_token');
  return null;
}

/**
 * Extract redirect path from URL — checks fragment first (from auth site), then query params.
 * Fragment takes precedence because auth site passes redirect through the fragment.
 */
function getRedirect(): string {
  // 1. Fragment (from auth site callback.astro redirect with tokens in fragment)
  const hash = window.location.hash;
  if (hash && hash.length > 1) {
    const hashParams = new URLSearchParams(hash.substring(1));
    const fromHash = hashParams.get('redirect');
    if (fromHash) return fromHash;
  }

  // 2. Query params (standard OAuth pattern, fallback)
  const url = new URL(window.location.href);
  const fromQuery = url.searchParams.get('redirect');
  if (fromQuery) return fromQuery;

  return '/overview';
}

export function AuthCallbackPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const initialize = useAuthStore((state) => state.initialize);

  const [status, setStatus] = useState<'processing' | 'success' | 'error'>('processing');
  const [errorMessage, setErrorMessage] = useState<string>('');

  // Extract params from URL (for render-time use)
  const errorParam = searchParams.get('error');
  const errorDescription = searchParams.get('error_description');
  const isNewUser = (getToken('new_user') || '') === 'true';

  useEffect(() => {
    const processAuthCallback = async () => {
      // Extract tokens fresh from URL inside the effect to avoid stale values on re-renders
      const tokenParam = getToken('token');
      const refreshTokenParam = getToken('refresh_token');
      const redirectPath = getRedirect();

      try {
        if (errorParam) {
          logger.error('Auth callback error:', {
            error: errorParam,
            description: errorDescription,
          });
          setStatus('error');
          setErrorMessage(errorDescription || errorParam || 'Authentication failed');
          return;
        }

        if (!tokenParam) {
          logger.error('No access token found in URL fragment or query params');
          setStatus('error');
          setErrorMessage('No authentication token found. Please try logging in again.');
          return;
        }

        // Store tokens in encrypted storage via TokenVault
        await tokenVault.initialize();
        await tokenVault.setAccessToken(tokenParam);
        if (refreshTokenParam) {
          await tokenVault.setRefreshToken(refreshTokenParam);
        }
        logger.info('Got tokens from auth site redirect and stored in TokenVault');

        await initialize();
        const authState = useAuthStore.getState();

        if (authState.isAuthenticated) {
          setStatus('success');

          if (isNewUser) {
            logger.info('New user signup detected, redirecting to onboarding');
            navigate('/onboarding', { replace: true });
            return;
          }

          const destination = redirectPath.startsWith('/') ? redirectPath : `/${redirectPath}`;
          logger.info('Auth successful, redirecting to:', destination);
          navigate(destination, { replace: true });
        } else {
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
  }, [
    errorParam,
    errorDescription,
    initialize,
    navigate,
  ]);

  // Handle "Try Again" click
  const handleTryAgain = () => {
    // Extract just the path from the redirect URL to avoid double-encoding issues
    const redirectValue = getRedirect();
    let redirectPath = redirectValue;
    try {
      // If redirect is a full URL, extract just the path
      if (redirectValue.startsWith('http://') || redirectValue.startsWith('https://')) {
        const url = new URL(redirectValue);
        redirectPath = url.pathname + url.search;
      }
    } catch {
      // Use as-is if URL parsing fails
    }
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
              {isNewUser ? 'Setting up your new account...' : 'Please wait while we sign you in...'}
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
            <h1 className="text-xl font-semibold text-text-primary mb-2">Authentication failed</h1>
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
