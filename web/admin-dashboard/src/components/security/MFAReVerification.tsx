/**
 * MFA Re-verification Checker Component
 *
 * Prompts the user for a fresh TOTP code once the server-side
 * `mfa_verified_at` timestamp is older than SESSION.MFA_REVERIFY_INTERVAL.
 * The verification is server-validated (see adminAuthStore.verifyMFA) — the
 * prompt cannot be dismissed with a random 6-digit code.
 */

import { useState, useEffect } from 'react';
import { useAdminAuth } from '@/hooks/useAdminAuth';
import { Shield } from 'lucide-react';
import { SESSION } from '@/lib/constants';

export function MFAReVerificationChecker({
  children,
}: {
  children: React.ReactNode;
}) {
  const [showMFAPrompt, setShowMFAPrompt] = useState(false);
  const [mfaCode, setMfaCode] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { user, session, verifyMFA } = useAdminAuth();

  useEffect(() => {
    const checkMFAStatus = () => {
      if (!user || !user.mfa_enabled) return;

      const verifiedAt = user.mfa_verified_at || session?.mfa_verified_at;
      if (!verifiedAt) {
        // No prior MFA on file — never block here, the login flow handles
        // initial MFA. Treat as verified and don't prompt.
        return;
      }

      const now = Date.now();
      const verifiedMs = new Date(verifiedAt).getTime();
      const timeSinceVerification = now - verifiedMs;

      if (timeSinceVerification > SESSION.MFA_REVERIFY_INTERVAL) {
        setShowMFAPrompt(true);
      } else {
        setShowMFAPrompt(false);
      }
    };

    checkMFAStatus();
    const interval = setInterval(checkMFAStatus, 60_000);
    return () => clearInterval(interval);
  }, [user, session?.mfa_verified_at]);

  const handleVerifyMFA = async () => {
    if (mfaCode.length !== 6 || submitting) return;
    setSubmitting(true);
    setError(null);
    try {
      await verifyMFA(mfaCode);
      setShowMFAPrompt(false);
      setMfaCode('');
    } catch (err) {
      const message =
        err instanceof Error && err.message ? err.message : 'Invalid code. Please try again.';
      setError(message);
    } finally {
      setSubmitting(false);
    }
  };

  if (showMFAPrompt) {
    return (
      <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" role="dialog" aria-modal="true" aria-labelledby="mfa-reverify-title">
        <div className="bg-white p-8 rounded-lg shadow-xl max-w-md w-full">
          <div className="flex items-center gap-3 mb-6">
            <Shield className="w-8 h-8 text-blue-600" />
            <h2 id="mfa-reverify-title" className="text-2xl font-bold">MFA Re-verification Required</h2>
          </div>
          <p className="text-gray-600 mb-6">
            For security reasons, please re-verify your identity with MFA.
          </p>

          <div className="mb-4">
            <label htmlFor="mfa-reverify-code" className="block text-sm font-medium text-gray-700 mb-2">
              Enter 6-digit code
            </label>
            <input
              id="mfa-reverify-code"
              type="text"
              inputMode="numeric"
              autoComplete="one-time-code"
              maxLength={6}
              value={mfaCode}
              onChange={(e) => {
                setMfaCode(e.target.value.replace(/\D/g, ''));
                if (error) setError(null);
              }}
              onKeyDown={(e) => {
                if (e.key === 'Enter') handleVerifyMFA();
              }}
              autoFocus
              aria-invalid={Boolean(error)}
              aria-describedby={error ? 'mfa-reverify-error' : undefined}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-2xl tracking-widest text-center"
              placeholder="000000"
            />
          </div>

          {error && (
            <div
              id="mfa-reverify-error"
              role="alert"
              className="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 text-sm rounded"
            >
              {error}
            </div>
          )}

          <button
            onClick={handleVerifyMFA}
            disabled={mfaCode.length !== 6 || submitting}
            className="w-full py-2 px-4 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium"
          >
            {submitting ? 'Verifying…' : 'Verify'}
          </button>
        </div>
      </div>
    );
  }

  return <>{children}</>;
}
