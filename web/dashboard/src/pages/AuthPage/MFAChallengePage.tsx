import { OTPInput } from '@/components/auth/OTPInput';
import { Button } from '@/components/ui/button';
import { LoadingSpinner } from '@/components/ui/loading-spinner';
import { trackEvent } from '@/lib/analytics';
import { ADMIN_DASHBOARD_URL } from '@/lib/constants';
import {
    decodeJwtRole,
    isPlatformAdminRole,
    notifyAdminPanelAfterLogin,
} from '@/lib/platform-admin';
import { useAuthStore } from '@/stores/authStore';
import { ArrowLeft, Shield, Zap } from 'lucide-react';
import { useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';

interface MFAChallengePageProps {
  onVerify?: (code: string) => Promise<void>;
}

function getSafeRedirect(redirect: string | null): string | null {
  if (!redirect || typeof redirect !== 'string') return null;
  const decoded = decodeURIComponent(redirect.trim());
  if (decoded.startsWith('/') && !decoded.startsWith('//')) return decoded;
  return null;
}

export function MFAChallengePage({ onVerify }: MFAChallengePageProps) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { verifyMFA, isLoading, error, clearError, user } = useAuthStore();

  const [mfaCode, setMfaCode] = useState('');
  const [attempts, setAttempts] = useState(0);
  const [maxAttempts] = useState(5);
  const [isVerifying, setIsVerifying] = useState(false);
  const [resendNotice, setResendNotice] = useState<string | null>(null);

  const email = searchParams.get('email') || user?.email || '';
  const redirectTo = getSafeRedirect(searchParams.get('redirect'));
  const openAdminAfterLogin = searchParams.get('admin') === '1';
  const isRateLimited = attempts >= maxAttempts;

  const handleVerify = async (code: string) => {
    if (isRateLimited || isVerifying) return;

    setIsVerifying(true);
    setResendNotice(null);
    clearError();

    try {
      if (onVerify) {
        await onVerify(code);
      } else {
        await verifyMFA(code);
      }
      trackEvent('auth_mfa_verified');
      const role = decodeJwtRole(localStorage.getItem('ff-access-token'));
      if (
        openAdminAfterLogin &&
        isPlatformAdminRole(role) &&
        ADMIN_DASHBOARD_URL
      ) {
        window.location.assign(ADMIN_DASHBOARD_URL);
        return;
      }
      notifyAdminPanelAfterLogin(role);
      navigate(redirectTo ?? '/overview', { replace: true });
    } catch (err) {
      trackEvent('auth_mfa_failed');
      setAttempts((prev) => prev + 1);
      setMfaCode('');
    } finally {
      setIsVerifying(false);
    }
  };

  const handleResend = async () => {
    // TOTP doesn't have a "resend" concept. Codes rotate periodically, so the best UX
    // is to tell the user to enter a fresh code from their authenticator app.
    if (isRateLimited) return;
    setResendNotice(
      'Authenticator codes rotate every ~30 seconds. Please enter the current code from your app.'
    );
  };

  return (
    <div className="min-h-screen bg-bg-primary flex flex-col">
      {/* Simple Header */}
      <header className="border-b border-border-subtle bg-bg-secondary">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <Link to="/" className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-lg bg-linear-to-br from-[#6366f1] to-[#8b5cf6] flex items-center justify-center">
                <Zap className="w-5 h-5 text-white" fill="currentColor" />
              </div>
              <span className="text-xl font-bold gradient-text">FunctionFly</span>
            </Link>
            <Link to="/login">
              <Button variant="ghost" size="sm">
                <ArrowLeft className="w-4 h-4 mr-2" />
                Back to Login
              </Button>
            </Link>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="flex-1 flex items-center justify-center px-4 py-12">
        <div className="w-full max-w-md">
          {/* Icon */}
          <div className="flex justify-center mb-8">
            <div className="w-16 h-16 rounded-full bg-[#6366f1]/10 flex items-center justify-center">
              <Shield className="w-8 h-8 text-[#6366f1]" />
            </div>
          </div>

          {/* Error Message */}
          {error && (
            <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          {/* Rate Limited Message */}
          {isRateLimited && (
            <div className="mb-6 p-4 bg-orange-50 border border-orange-200 rounded-lg">
              <p className="text-sm text-orange-800">
                Too many failed attempts. Please try again later or contact support.
              </p>
            </div>
          )}

          {/* OTP Input */}
          {!isRateLimited && (
            <OTPInput
              length={6}
              onComplete={handleVerify}
              onResend={handleResend}
              isLoading={isVerifying}
              title="Two-Factor Authentication"
              description={`Enter the 6-digit code from your authenticator app`}
              error={error || undefined}
            />
          )}

          {resendNotice && !isRateLimited && (
            <div className="mt-4 rounded-lg border border-border-default bg-bg-secondary p-3 text-sm text-text-secondary">
              {resendNotice}
            </div>
          )}

          {/* Attempts Counter */}
          {!isRateLimited && attempts > 0 && (
            <p className="text-center text-sm text-text-muted mt-4">
              Attempts remaining: {maxAttempts - attempts}
            </p>
          )}

          {/* Loading State */}
          {isVerifying && (
            <div className="flex justify-center py-4">
              <LoadingSpinner text="Verifying..." />
            </div>
          )}

          {/* Help Text */}
          <div className="mt-8 text-center text-sm text-text-muted">
            <p>
              Having trouble?{' '}
              <a href="mailto:support@functionfly.com" className="text-brand-500 hover:underline">
                Contact Support
              </a>
            </p>
          </div>
        </div>
      </main>
    </div>
  );
}
