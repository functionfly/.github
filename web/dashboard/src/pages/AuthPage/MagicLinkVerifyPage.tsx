import { Button } from '@/components/ui/button';
import { LoadingSpinner } from '@/components/ui/loading-spinner';
import { auth } from '@/lib/auth';
import { CheckCircle2, AlertCircle, Sparkles, ArrowRight } from 'lucide-react';
import React, { useEffect, useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';

export function MagicLinkVerifyPage(): React.JSX.Element {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const token = searchParams.get('token');

  const [isLoading, setIsLoading] = useState(true);
  const [isSuccess, setIsSuccess] = useState(false);
  const [isNewUser, setIsNewUser] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!token) {
      setError('Invalid magic link. No token provided.');
      setIsLoading(false);
      return;
    }

    const verifyMagicLink = async () => {
      try {
        const result = await auth.verifyMagicLink(token);
        setIsSuccess(true);
        setIsNewUser(!!result.new_user);

        toast.success(result.new_user
          ? 'Welcome to FunctionFly! Your account has been created.'
          : 'Successfully signed in!'
        );

        // Small delay to show success state before redirecting
        setTimeout(() => {
          navigate('/overview', { replace: true });
        }, 1500);
      } catch (err) {
        const message = err instanceof Error ? err.message : 'Failed to verify magic link';
        setError(message);
        toast.error(message);
      } finally {
        setIsLoading(false);
      }
    };

    verifyMagicLink();
  }, [token, navigate]);

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-bg-secondary p-4">
        <div className="w-full max-w-md text-center space-y-6">
          <div className="flex justify-center">
            <div className="rounded-full bg-brand-100 p-4 dark:bg-brand-900/30 animate-pulse">
              <Sparkles className="h-8 w-8 text-brand-600 dark:text-brand-400" />
            </div>
          </div>
          <div className="space-y-2">
            <h1 className="text-xl font-semibold">Verifying your magic link...</h1>
            <p className="text-sm text-text-muted">
              Just a moment while we sign you in securely.
            </p>
          </div>
          <div className="flex justify-center">
            <LoadingSpinner size="lg" text="" />
          </div>
        </div>
      </div>
    );
  }

  if (isSuccess) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-bg-secondary p-4">
        <div className="w-full max-w-md text-center space-y-6">
          <div className="flex justify-center">
            <div className="rounded-full bg-green-100 p-4 dark:bg-green-900/30">
              <CheckCircle2 className="h-8 w-8 text-green-600 dark:text-green-400" />
            </div>
          </div>
          <div className="space-y-2">
            <h1 className="text-xl font-semibold">
              {isNewUser ? 'Welcome to FunctionFly!' : 'Successfully signed in!'}
            </h1>
            <p className="text-sm text-text-muted">
              {isNewUser
                ? 'Your account has been created and you\'re all set.'
                : 'Redirecting you to the dashboard...'}
            </p>
          </div>
          <Button
            onClick={() => navigate('/overview', { replace: true })}
            className="mt-4"
          >
            Go to Dashboard
            <ArrowRight className="ml-2 h-4 w-4" />
          </Button>
        </div>
      </div>
    );
  }

  // Error state
  return (
    <div className="min-h-screen flex items-center justify-center bg-bg-secondary p-4">
      <div className="w-full max-w-md text-center space-y-6">
        <div className="flex justify-center">
          <div className="rounded-full bg-red-100 p-4 dark:bg-red-900/30">
            <AlertCircle className="h-8 w-8 text-red-600 dark:text-red-400" />
          </div>
        </div>
        <div className="space-y-2">
          <h1 className="text-xl font-semibold">Couldn&apos;t sign in</h1>
          <p className="text-sm text-text-muted">
            {error || 'This magic link has expired or already been used.'}
          </p>
        </div>
        <div className="space-y-3">
          <Link to="/auth/login">
            <Button variant="outline" className="w-full">
              Try signing in another way
            </Button>
          </Link>
          <Link to="/auth/magic-link">
            <Button variant="ghost" className="w-full">
              Request a new magic link
            </Button>
          </Link>
        </div>
      </div>
    </div>
  );
}
