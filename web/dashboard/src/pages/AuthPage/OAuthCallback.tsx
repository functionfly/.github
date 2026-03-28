import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { LoadingSpinner } from '@/components/ui/loading-spinner';
import { notifyAdminPanelAfterLogin } from '@/lib/platform-admin';
import { useAuthStore } from '@/stores/authStore';
import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';

export function OAuthCallback() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { initialize } = useAuthStore();
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isProcessing, setIsProcessing] = useState(true);

  useEffect(() => {
    const handleOAuthCallback = async () => {
      try {
        // Extract OAuth parameters from URL
        const token = searchParams.get('token');
        const refreshToken = searchParams.get('refresh_token');
        const error = searchParams.get('error');
        const errorDescription = searchParams.get('error_description');
        const newUser = searchParams.get('new_user') === 'true';

        if (error) {
          console.error('OAuth error:', error, errorDescription);

          // Provide user-friendly error messages based on error type
          let userFriendlyMessage = 'Authentication failed';
          switch (error) {
            case 'invalid_provider':
              userFriendlyMessage =
                'This social login provider is not configured. Please try a different method.';
              break;
            case 'token_exchange_failed':
              userFriendlyMessage = 'Failed to connect with the social provider. Please try again.';
              break;
            case 'user_info_failed':
              userFriendlyMessage =
                'Could not retrieve your account information. Please try again.';
              break;
            case 'missing_email':
              userFriendlyMessage =
                'Your social account must have a verified email address. Please update your account or try a different login method.';
              break;
            case 'account_link_failed':
              userFriendlyMessage =
                'Could not link your social account. Please contact support if this continues.';
              break;
            case 'user_creation_failed':
              userFriendlyMessage =
                'Could not create your account. Please try again or contact support.';
              break;
            default:
              userFriendlyMessage =
                errorDescription || 'An unexpected error occurred during authentication.';
          }

          // Avoid leaving stale tokens from a previous attempt.
          localStorage.removeItem('ff-access-token');
          localStorage.removeItem('ff-refresh-token');

          setErrorMessage(userFriendlyMessage);
          setIsProcessing(false);
          return;
        }

        if (!token) {
          console.error('No token received from OAuth callback');
          localStorage.removeItem('ff-access-token');
          localStorage.removeItem('ff-refresh-token');
          setErrorMessage('Authentication failed - no token received. Please try again.');
          setIsProcessing(false);
          return;
        }

        // Avoid leaving stale tokens if initialize() fails.
        localStorage.removeItem('ff-access-token');
        localStorage.removeItem('ff-refresh-token');

        // Store the JWT token and refresh token — initialize() will validate them and fetch real user data
        localStorage.setItem('ff-access-token', token);
        if (refreshToken) {
          localStorage.setItem('ff-refresh-token', refreshToken);
        }

        // Re-initialize auth store: validates the token against the backend
        // and populates real user data (no placeholder values stored)
        await initialize();

        notifyAdminPanelAfterLogin(useAuthStore.getState().user?.role);

        // Navigate to dashboard or onboarding based on whether it's a new user
        if (newUser) {
          navigate('/onboarding', { replace: true });
        } else {
          navigate('/overview', { replace: true });
        }
      } catch (error) {
        console.error('OAuth callback processing failed:', error);
        localStorage.removeItem('ff-access-token');
        localStorage.removeItem('ff-refresh-token');
        setErrorMessage('Authentication failed. Please try again.');
        setIsProcessing(false);
      }
    };

    setIsProcessing(true);
    setErrorMessage(null);
    handleOAuthCallback();
  }, [navigate, searchParams, initialize]);

  return (
    <div className="min-h-screen bg-bg-primary flex items-center justify-center px-4">
      <div className="w-full max-w-md">
        {isProcessing ? (
          <div className="text-center">
            <LoadingSpinner text="Completing authentication..." />
            <p className="mt-4 text-text-secondary">
              Please wait while we complete your sign in...
            </p>
          </div>
        ) : (
          <Card className="border-white/8 bg-bg-secondary">
            <CardContent className="p-6">
              <div className="text-center">
                <p className="text-sm font-medium text-text-primary">
                  {errorMessage ?? 'Authentication failed'}
                </p>
                <div className="mt-6 flex gap-2">
                  <Button
                    className="flex-1"
                    variant="outline"
                    onClick={() => navigate('/login', { replace: true })}
                  >
                    Back to Login
                  </Button>
                  <Button
                    className="flex-1"
                    onClick={() => {
                      // Quick retry: re-run the callback logic for the same URL.
                      setIsProcessing(true);
                      setErrorMessage(null);
                      window.location.reload();
                    }}
                  >
                    Retry
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}
