import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { FormError } from '@/components/ui/form-error';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { LoadingSpinner } from '@/components/ui/loading-spinner';
import { useLoginForm } from '@/hooks/useAuthForms';
import { ADMIN_DASHBOARD_URL, getApiBaseUrl } from '@/lib/constants';
import { isPlatformAdminRole, notifyAdminPanelAfterLogin } from '@/lib/platform-admin';
import { cn } from '@/lib/utils';
import { useAuthStore } from '@/stores/authStore';
import { AlertCircle, Eye, EyeOff, Info, LockKeyhole, Shield, User } from 'lucide-react';
import React, { useEffect, useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';

// New authentication libraries
import { useAutoAnimate } from '@formkit/auto-animate/react';
import { useGoogleReCaptcha } from 'react-google-recaptcha-v3';

// OAuth Provider type
interface OAuthProvider {
  id: string;
  name: string;
  clientId: string;
}

// Fetch available OAuth providers
async function fetchOAuthProviders(): Promise<OAuthProvider[]> {
  try {
    const base = getApiBaseUrl();
    const response = await fetch(`${base}/v1/auth/oauth/providers`);
    if (!response.ok) return [];
    const data = await response.json();
    return data.providers || [];
  } catch {
    return [];
  }
}

// Handle social login
const handleSocialLogin = async (provider: string) => {
  try {
    const base = getApiBaseUrl();
    const response = await fetch(`${base}/v1/auth/oauth/url?provider=${provider}`);
    if (!response.ok) {
      throw new Error(`Failed to get OAuth URL: ${response.statusText}`);
    }

    const data = await response.json();
    if (data.url) {
      window.location.href = data.url;
    }
  } catch (error) {
    console.error('Social login failed:', error);
  }
};

function getSafeRedirect(redirect: string | null): string | null {
  if (!redirect || typeof redirect !== 'string') return null;
  const decoded = decodeURIComponent(redirect.trim());
  if (decoded.startsWith('/') && !decoded.startsWith('//')) return decoded;
  return null;
}

interface LoginFormProps {
  authMode?: 'email' | 'social' | 'magic';
  setAuthMode?: (mode: 'email' | 'social' | 'magic') => void;
}

export function LoginForm({
  authMode: authModeProp,
  setAuthMode: setAuthModeProp,
}: LoginFormProps): React.JSX.Element {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const redirectTo = getSafeRedirect(searchParams.get('redirect'));
  const openAdminAfterLogin = searchParams.get('admin') === '1';
  const { login, isLoading, error, clearError } = useAuthStore();
  const [showPassword, setShowPassword] = useState(false);

  // Auth mode: use props from parent when provided (tabs rendered in AuthPage), otherwise local state
  const [authModeLocal, setAuthModeLocal] = useState<'email' | 'social'>('email');
  const authMode = authModeProp ?? authModeLocal;
  const setAuthMode = setAuthModeProp ?? setAuthModeLocal;
  const [recaptchaToken, setRecaptchaToken] = useState<string | null>(null);

  // OAuth providers
  const [oauthProviders, setOauthProviders] = useState<OAuthProvider[]>([]);
  const [isLoadingProviders, setIsLoadingProviders] = useState(true);

  // Rate limiting state
  const [rateLimited, setRateLimited] = useState(false);
  const [retryAfter, setRetryAfter] = useState(0);

  // Auto-animate refs
  const [formRef] = useAutoAnimate();
  const [socialAuthRef] = useAutoAnimate();

  // reCAPTCHA hook
  const { executeRecaptcha } = useGoogleReCaptcha();

  // Fetch OAuth providers on mount
  useEffect(() => {
    const loadProviders = async () => {
      const providers = await fetchOAuthProviders();
      setOauthProviders(providers);
      setIsLoadingProviders(false);
    };
    loadProviders();
  }, []);

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors, isValid, isSubmitting },
    clearErrors,
  } = useLoginForm();

  const watchedErrors = Object.keys(errors).length > 0;

  // Execute reCAPTCHA when form is valid (only in production)
  useEffect(() => {
    const executeCaptcha = async () => {
      if (executeRecaptcha && isValid && !watchedErrors && import.meta.env.PROD) {
        try {
          const token = await executeRecaptcha('login');
          setRecaptchaToken(token);
        } catch (error) {
          console.error('reCAPTCHA execution failed:', error);
        }
      }
    };
    executeCaptcha();
  }, [executeRecaptcha, isValid, watchedErrors]);

  const onSubmit = async (data: any) => {
    clearError();
    clearErrors();

    // Verify reCAPTCHA token (only in production)
    if (import.meta.env.PROD && !recaptchaToken) {
      toast.error('Please complete the security verification');
      return;
    }

    try {
      const loginData = {
        ...data,
        ...(import.meta.env.PROD && recaptchaToken ? { recaptchaToken } : {}),
      };
      console.log('[debug] loginData type:', typeof loginData, loginData.constructor.name, JSON.stringify(loginData));
      await login(loginData);

      // Check if MFA is required
      if (useAuthStore.getState().mfaRequired) {
        const redirectQuery = redirectTo ? `&redirect=${encodeURIComponent(redirectTo)}` : '';
        const adminQuery = openAdminAfterLogin ? '&admin=1' : '';
        navigate(
          `/auth/mfa-challenge?email=${encodeURIComponent(data.email)}${redirectQuery}${adminQuery}`,
          {
            replace: true,
          }
        );
        return;
      }

      const u = useAuthStore.getState().user;
      if (openAdminAfterLogin && isPlatformAdminRole(u?.role) && ADMIN_DASHBOARD_URL) {
        window.location.assign(ADMIN_DASHBOARD_URL);
        return;
      }
      notifyAdminPanelAfterLogin(u?.role);
      navigate(redirectTo ?? '/overview', { replace: true });
    } catch (err: unknown) {
      // Handle MFA required - redirect to MFA challenge
      const error = err as Error & {
        status?: number;
        retryAfter?: number;
        response?: { data?: { retryAfter?: number; retry_after?: number } };
      };
      if (error.message === 'MFA_REQUIRED') {
        const adminQuery = openAdminAfterLogin ? '&admin=1' : '';
        navigate(`/auth/mfa-challenge?email=${encodeURIComponent(data.email)}${adminQuery}`, {
          replace: true,
        });
        return;
      }
      // Only treat real HTTP 429 (or legacy axios shape) as rate limit — not wrong-password 401s.
      const status = error.status;
      const isRateLimited = status === 429;
      if (isRateLimited) {
        const retrySeconds = Math.min(
          3600,
          Math.max(
            1,
            error.retryAfter ??
              error.response?.data?.retry_after ??
              error.response?.data?.retryAfter ??
              60
          )
        );
        setRateLimited(true);
        setRetryAfter(retrySeconds);
        const countdown = setInterval(() => {
          setRetryAfter((prev: number) => {
            if (prev <= 1) {
              clearInterval(countdown);
              setRateLimited(false);
              return 0;
            }
            return prev - 1;
          });
        }, 1000);
      }
      // Error is shown via store (FormError); keep user on login page
    }
  };

  // Clear server/store error when user edits credentials so they can retry without stale message
  const handleClearErrorOnChange = () => {
    if (error) clearError();
  };

  return (
    <div className="space-y-6">
      {/* Auth Mode Selector - only when parent doesn't control it (no props) */}
      {setAuthModeProp == null && (
        <div className="flex justify-center space-x-2 mb-6">
          <Button
            type="button"
            size="sm"
            onClick={() => setAuthMode('email')}
            className={`auth-mode-btn transition-all duration-300 ${
              authMode === 'email'
                ? 'auth-mode-btn-active'
                : 'auth-mode-btn-inactive hover:bg-bg-hover'
            }`}
          >
            Login
          </Button>
          <Button
            type="button"
            size="sm"
            onClick={() => setAuthMode('social')}
            className={`auth-mode-btn transition-all duration-300 ${
              authMode === 'social'
                ? 'auth-mode-btn-active'
                : 'auth-mode-btn-inactive hover:bg-bg-hover'
            }`}
          >
            Social Login
          </Button>
        </div>
      )}

      <div ref={formRef}>
        {authMode === 'email' && (
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
            {error && (
              <FormError
                error={error}
                className="animate-in fade-in slide-in-from-top-2 duration-200"
              />
            )}

            {/* Rate Limited Warning */}
            {rateLimited && (
              <div className="flex items-center gap-2 p-3 bg-orange-50 border border-orange-200 rounded-lg text-orange-800">
                <AlertCircle className="w-5 h-5 shrink-0" />
                <div className="text-sm">
                  <p className="font-medium">Too many login attempts</p>
                  <p>
                    Please wait <span className="font-bold">{retryAfter}</span> seconds before
                    trying again.
                  </p>
                </div>
              </div>
            )}

            {/* reCAPTCHA Badge - only shown in production */}
            {import.meta.env.PROD && (
              <div className="flex items-center justify-center text-xs text-text-muted">
                <Shield className="w-3 h-3 mr-1" />
                Protected by reCAPTCHA
              </div>
            )}

            <div className="space-y-2">
              <Label
                htmlFor="email"
                className={cn(
                  'flex items-center gap-2',
                  errors.email && 'text-error',
                  !errors.email && watch('email') && 'text-success'
                )}
              >
                Username or Email <span className="text-error">*</span>
              </Label>
              <div className="relative">
                <div className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted">
                  <User className="w-4 h-4" />
                </div>
                <Input
                  id="email"
                  type="text"
                  placeholder="username or you@example.com"
                  className={cn(
                    'pl-10 focus-visible:ring-2 focus-visible:ring-brand-500 focus-visible:ring-offset-2 transition-all duration-200',
                    errors.email && 'auth-input-error',
                    !errors.email && watch('email') && 'border-success focus:border-success'
                  )}
                  {...register('email')}
                  onInput={handleClearErrorOnChange}
                  aria-invalid={errors.email ? 'true' : 'false'}
                  aria-describedby={errors.email ? 'email-error' : 'email-help'}
                />
              </div>
              {!errors.email && (
                <p id="email-help" className="text-xs text-text-muted flex items-center gap-1.5">
                  <Info className="w-3 h-3" />
                  Sign in with your username or email address
                </p>
              )}
              {errors.email && (
                <div id="email-error" className="auth-error-text">
                  <AlertCircle className="w-3 h-3" />
                  {typeof errors.email.message === 'string'
                    ? errors.email.message
                    : 'Invalid username or email'}
                </div>
              )}
            </div>

            <div className="space-y-2">
              <Label
                htmlFor="password"
                className={cn(
                  'flex items-center gap-2',
                  errors.password && 'text-error',
                  !errors.password && watch('password') && 'text-success'
                )}
              >
                Password <span className="text-error">*</span>
              </Label>
              <div className="relative">
                <div className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted">
                  <LockKeyhole className="w-4 h-4" />
                </div>
                <Input
                  id="password"
                  type={showPassword ? 'text' : 'password'}
                  placeholder="••••••••"
                  className={cn(
                    'pl-10 pr-10 focus-visible:ring-2 focus-visible:ring-brand-500 focus-visible:ring-offset-2 transition-all duration-200',
                    errors.password && 'auth-input-error',
                    !errors.password && watch('password') && 'border-success focus:border-success'
                  )}
                  {...register('password')}
                  onInput={handleClearErrorOnChange}
                  aria-invalid={errors.password ? 'true' : 'false'}
                  aria-describedby={errors.password ? 'password-error' : undefined}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="password-toggle absolute right-3 top-1/2 -translate-y-1/2 p-1 rounded text-text-muted hover:text-text-primary hover:bg-bg-hover focus:outline-none focus-visible:ring-2 focus-visible:ring-brand-500 focus-visible:ring-offset-2"
                  aria-label={showPassword ? 'Hide password' : 'Show password'}
                >
                  {showPassword ? (
                    <EyeOff className="w-4 h-4 transition-transform duration-200" />
                  ) : (
                    <Eye className="w-4 h-4 transition-transform duration-200" />
                  )}
                </button>
              </div>
              {errors.password && (
                <div id="password-error" className="auth-error-text">
                  <AlertCircle className="w-3 h-3" />
                  {typeof errors.password?.message === 'string'
                    ? errors.password.message
                    : 'Invalid password'}
                </div>
              )}
            </div>

            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <Checkbox
                  id="rememberMe"
                  {...register('rememberMe')}
                  className="focus-visible:ring-2 focus-visible:ring-brand-500 focus-visible:ring-offset-2"
                />
                <label
                  htmlFor="rememberMe"
                  className="text-sm text-text-secondary cursor-pointer hover:text-text-primary transition-colors"
                >
                  Remember me
                </label>
              </div>
              <span className="text-border-subtle">|</span>
              <Link
                to="/auth/reset-password"
                className="text-sm text-brand-500 hover:text-brand-400 hover:underline transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500 focus-visible:ring-offset-2 rounded-sm"
              >
                Forgot password?
              </Link>
            </div>

            {/* Magic Link Option */}
            <div className="flex items-center justify-center gap-2 pt-2">
              <span className="text-sm text-text-muted">Or</span>
              <Link
                to="/auth/magic-link"
                className="text-sm text-brand-500 hover:text-brand-400 hover:underline transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500 focus-visible:ring-offset-2 rounded-sm flex items-center gap-1"
              >
                <span className="inline-block">✨</span>
                Sign in with Magic Link
              </Link>
            </div>

            <Button
              type="submit"
              className="w-full focus-visible:ring-2 focus-visible:ring-brand-500 focus-visible:ring-offset-2 transition-all duration-200"
              disabled={isLoading || isSubmitting || !isValid || watchedErrors || rateLimited}
            >
              {isLoading || isSubmitting ? (
                <LoadingSpinner text="Signing in..." />
              ) : rateLimited ? (
                <>Wait {retryAfter}s</>
              ) : (
                'Sign In'
              )}
            </Button>
          </form>
        )}

        {authMode === 'social' && (
          <div ref={socialAuthRef} className="space-y-4">
            <div className="text-center mb-4">
              <h3 className="text-lg font-semibold">Sign in with Social</h3>
              <p className="text-sm text-text-muted">Choose your preferred social login method</p>
            </div>

            {/* Provider loading / empty state */}
            {isLoadingProviders && (
              <div className="flex items-center justify-center py-6">
                <LoadingSpinner text="Loading providers..." />
              </div>
            )}

            {!isLoadingProviders && oauthProviders.length === 0 && (
              <div className="rounded-lg border border-border-default bg-bg-secondary p-4 text-center">
                <p className="text-sm font-medium text-text-primary">Social login unavailable</p>
                <p className="mt-1 text-xs text-text-muted">
                  No OAuth providers are configured for this environment.
                </p>
              </div>
            )}

            {/* Social Auth Buttons */}
            <div className="space-y-3">
              <div className="relative">
                <div className="absolute inset-0 flex items-center">
                  <span className="w-full border-t" />
                </div>
                <div className="relative flex justify-center text-xs uppercase">
                  <span className="bg-background px-2 text-text-muted">Or continue with</span>
                </div>
              </div>

              {/* Static Social Login Buttons */}
              <div className="grid grid-cols-2 gap-3">
                <Button
                  type="button"
                  onClick={() => handleSocialLogin('github')}
                  disabled={isLoading || isLoadingProviders}
                  className="social-btn-github focus-visible:ring-2 focus-visible:ring-gray-600 focus-visible:ring-offset-2 transition-all duration-200"
                >
                  <svg className="w-4 h-4 mr-2" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
                  </svg>
                  GitHub
                </Button>
                <Button
                  type="button"
                  onClick={() => handleSocialLogin('google')}
                  disabled={isLoading || isLoadingProviders}
                  className="social-btn-google focus-visible:ring-2 focus-visible:ring-gray-400 focus-visible:ring-offset-2 transition-all duration-200"
                >
                  <svg className="w-4 h-4 mr-2" viewBox="0 0 24 24">
                    <path
                      fill="#4285F4"
                      d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                    />
                    <path
                      fill="#34A853"
                      d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                    />
                    <path
                      fill="#FBBC05"
                      d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
                    />
                    <path
                      fill="#EA4335"
                      d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                    />
                  </svg>
                  Google
                </Button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
