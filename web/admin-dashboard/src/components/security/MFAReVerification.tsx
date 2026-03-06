/**
 * MFA Re-verification Checker Component
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
  const { user, session, verifyMFA } = useAdminAuth();

  useEffect(() => {
    const checkMFAStatus = () => {
      if (!user || !session?.mfa_verified_at) return;

      const now = Date.now();
      const mfaVerifiedAt = new Date(session.mfa_verified_at).getTime();
      const timeSinceVerification = now - mfaVerifiedAt;

      if (timeSinceVerification > SESSION.MFA_REVERIFY_INTERVAL) {
        setShowMFAPrompt(true);
      }
    };

    // Check on mount
    checkMFAStatus();

    // Check every minute
    const interval = setInterval(checkMFAStatus, 60 * 1000);

    return () => clearInterval(interval);
  }, [user, session?.mfa_verified_at]);

  const handleVerifyMFA = () => {
    if (mfaCode.length === 6) {
      verifyMFA();
      setShowMFAPrompt(false);
      setMfaCode('');
    }
  };

  if (showMFAPrompt) {
    return (
      <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
        <div className="bg-white p-8 rounded-lg shadow-xl max-w-md w-full">
          <div className="flex items-center gap-3 mb-6">
            <Shield className="w-8 h-8 text-blue-600" />
            <h2 className="text-2xl font-bold">MFA Re-verification Required</h2>
          </div>
          <p className="text-gray-600 mb-6">
            For security reasons, please re-verify your identity with MFA.
          </p>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Enter 6-digit code
            </label>
            <input
              type="text"
              maxLength={6}
              value={mfaCode}
              onChange={(e) => setMfaCode(e.target.value.replace(/\D/g, ''))}
              autoFocus
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-2xl tracking-widest text-center"
              placeholder="000000"
            />
          </div>

          <button
            onClick={handleVerifyMFA}
            disabled={mfaCode.length !== 6}
            className="w-full py-2 px-4 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium"
          >
            Verify
          </button>
        </div>
      </div>
    );
  }

  return <>{children}</>;
}
