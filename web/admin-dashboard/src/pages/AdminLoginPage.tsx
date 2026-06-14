/**
 * Admin Login Page
 * Placeholder for authentication
 */

import { bootstrapAdminSession, loginAdmin } from '@/lib/api/adminAuth';
import { adminApiClient } from '@/lib/api/adminClient';
import { trackSecurityEvent } from '@/lib/monitoring/securityEvents';
import { generateDeviceFingerprint } from '@/lib/security/deviceFingerprint';
import { useAdminAuthStore } from '@/stores/adminAuthStore';
import { AlertTriangle, Check, KeyRound, LogIn, Shield, X } from 'lucide-react';
import { useEffect, useState } from 'react';
import { Navigate, useNavigate } from 'react-router-dom';

interface LastLoginDisplay {
  ip_address: string;
  device_name: string;
  timestamp: string;
  suspicious: boolean;
}

export function AdminLoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [lastLogin, setLastLogin] = useState<LastLoginDisplay | null>(null);
  const [suspiciousConfirmed, setSuspiciousConfirmed] = useState(false);

  // Forgot-password inline form state
  const [forgotMode, setForgotMode] = useState(false);
  const [forgotEmail, setForgotEmail] = useState('');
  const [forgotStatus, setForgotStatus] = useState<'idle' | 'sending' | 'sent' | 'error'>('idle');
  const [forgotMessage, setForgotMessage] = useState('');

  // Optional MFA step (after password is accepted)
  const [mfaStep, setMfaStep] = useState<{ token: string; email: string } | null>(null);
  const [mfaCode, setMfaCode] = useState('');
  const [mfaError, setMfaError] = useState('');
  const [mfaLoading, setMfaLoading] = useState(false);

  const navigate = useNavigate();
  const { login, setDeviceFingerprint, setLastLoginInfo } = useAdminAuthStore();
  const isAuthenticated = useAdminAuthStore((s) => s.isAuthenticated);

  // If already authenticated, redirect to dashboard immediately
  if (isAuthenticated) {
    return <Navigate to="/" replace />;
  }

  // Fetch last login info on mount (if user has a session cookie)
  useEffect(() => {
    async function fetchLastLogin() {
      try {
        // Use getNoAuth so a 401 (not authenticated) doesn't trigger
        // a redirect to login, which would cause a redirect loop
        const resp = await adminApiClient.getNoAuth<{ data?: LastLoginDisplay }>(
          '/auth/last-login'
        );
        if (resp?.data) {
          setLastLogin(resp.data);
        }
      } catch {
        // Ignore - endpoint may not exist yet
      }
    }
    fetchLastLogin();
  }, []);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      if (!email || !password) {
        setError('Please enter email and password');
        return;
      }

      const auth = await loginAdmin(email, password);

      // Some backends return a "mfa_required" sentinel with a short-lived
      // challenge token. When that happens, prompt for the TOTP code here
      // before exchanging it for a real session.
      const anyAuth = auth as typeof auth & {
        mfa_required?: boolean;
        challenge_token?: string;
      };
      if (anyAuth?.mfa_required || (anyAuth as any)?.status === 'mfa_required') {
        setMfaStep({
          token: anyAuth.challenge_token ?? (anyAuth as any).token ?? '',
          email,
        });
        return;
      }

      if (!auth.token) {
        throw new Error('Login succeeded but no access token was returned');
      }

      await completeLogin(auth.token, email);
    } catch (err) {
      trackSecurityEvent('login_failed', { email });
      const message = err instanceof Error ? err.message : 'Login failed. Please try again.';
      setError(message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleMfaSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!mfaStep) return;
    setMfaError('');
    setMfaLoading(true);
    try {
      const res = await adminApiClient.postNoAuth<{ token: string }>('/auth/mfa/challenge', {
        challenge_token: mfaStep.token,
        code: mfaCode,
      });
      if (!res?.token) {
        throw new Error('MFA verification did not return a session token');
      }
      await completeLogin(res.token, mfaStep.email);
    } catch (err) {
      trackSecurityEvent('mfa_failed', { email: mfaStep.email });
      const message = err instanceof Error ? err.message : 'Invalid code. Please try again.';
      setMfaError(message);
    } finally {
      setMfaLoading(false);
    }
  };

  const completeLogin = async (token: string, loginEmail: string) => {
    const fingerprint = await generateDeviceFingerprint();
    adminApiClient.setSessionToken(token);
    adminApiClient.setDeviceFingerprint(fingerprint);

    const bootstrap = await bootstrapAdminSession(token);

    // Check if this is a suspicious login (new device/IP)
    const lastLoginInfo: LastLoginDisplay | undefined = lastLogin?.suspicious
      ? { ...lastLogin, suspicious: false }
      : undefined;

    login(bootstrap.session, bootstrap.user, lastLoginInfo);
    setDeviceFingerprint(fingerprint);
    setLastLoginInfo(lastLoginInfo || null);

    // Store token in regular localStorage for cross-app compatibility
    try {
      localStorage.setItem('ffly_jwt', token);
    } catch {
      /* localStorage may be unavailable */
    }

    trackSecurityEvent('login_success', { email: loginEmail });
    navigate('/');
  };

  const handleForgotPassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setForgotStatus('sending');
    setForgotMessage('');
    try {
      await adminApiClient.postNoAuth('/auth/forgot-password', {
        email: forgotEmail,
      });
      setForgotStatus('sent');
      setForgotMessage(
        'If an account exists for that email, a password reset link has been sent. Please check your inbox.'
      );
    } catch (err) {
      setForgotStatus('error');
      const message = err instanceof Error ? err.message : 'Unable to send reset email right now.';
      setForgotMessage(message);
    }
  };

  const handleSuspiciousLogin = (wasMe: boolean) => {
    if (wasMe) {
      // User confirmed this was them - clear the flag
      setLastLogin((prev) => (prev ? { ...prev, suspicious: false } : null));
      setSuspiciousConfirmed(true);
    } else {
      // User says this wasn't them - report and suggest password change
      setSuspiciousConfirmed(true);
      // Report to backend
      adminApiClient
        .post('/security/report-suspicious-login', {
          reported_at: new Date().toISOString(),
          last_login: lastLogin,
        })
        .catch(() => {});
      setError('Suspicious activity has been reported. Please change your password immediately.');
    }
  };

  return (
    <div className="min-h-screen flex flex-col items-center justify-center px-4 py-8" style={{ backgroundColor: 'var(--color-bg)' }}>
      {/* Logo */}
      <div className="flex items-center justify-center gap-3 mb-8">
        <img
          src="/favicon.svg"
          alt=""
          width="40"
          height="40"
        />
        <span className="text-xl font-bold text-gray-900 dark:text-white">FunctionFly™</span>
      </div>

      {/* Single card: title + form */}
      <div className="w-full max-w-md bg-white dark:bg-slate-800 rounded-2xl border border-gray-200 dark:border-slate-700 shadow-lg shadow-gray-200/50 dark:shadow-slate-900/50 p-8 sm:p-10">
        <div className="mb-8">
          <h1 className="text-2xl font-semibold tracking-tight text-gray-900 dark:text-white">Sign in</h1>
          <p className="mt-2 text-sm text-gray-500 dark:text-slate-400">
            Use your admin credentials to access the dashboard.
          </p>
        </div>

        {/* Last successful login info */}
        {lastLogin && !suspiciousConfirmed && (
          <div
            className={`mb-6 p-4 rounded-lg border ${
              lastLogin.suspicious ? 'bg-amber-50 dark:bg-amber-900/30 border-amber-200 dark:border-amber-700' : 'bg-green-50 dark:bg-green-900/30 border-green-200 dark:border-green-700'
            }`}
          >
            <div className="flex items-start gap-3">
              {lastLogin.suspicious ? (
                <AlertTriangle className="w-5 h-5 text-amber-600 shrink-0 mt-0.5" />
              ) : (
                <Shield className="w-5 h-5 text-green-600 shrink-0 mt-0.5" />
              )}
              <div className="flex-1 min-w-0">
                <p
                  className={`text-sm font-medium ${
                    lastLogin.suspicious ? 'text-amber-800 dark:text-amber-300' : 'text-green-800 dark:text-green-300'
                  }`}
                >
                  {lastLogin.suspicious ? 'Unusual login detected' : 'Last successful login'}
                </p>
                <p className="text-xs text-gray-600 dark:text-slate-400 mt-1">
                  {lastLogin.device_name} · {lastLogin.ip_address}
                </p>
                <p className="text-xs text-gray-500 dark:text-slate-500 mt-0.5">
                  {new Date(lastLogin.timestamp).toLocaleString()}
                </p>

                {lastLogin.suspicious && (
                  <div className="mt-3 pt-3 border-t border-amber-200 dark:border-amber-700">
                    <p className="text-xs text-amber-700 dark:text-amber-400 mb-2">
                      Was this you? If not, your account may be compromised.
                    </p>
                    <div className="flex gap-2">
                      <button
                        type="button"
                        onClick={() => handleSuspiciousLogin(true)}
                        className="flex-1 text-xs py-1.5 px-3 bg-green-100 dark:bg-green-900/50 text-green-700 dark:text-green-300 rounded hover:bg-green-200 dark:hover:bg-green-800 transition-colors"
                      >
                        <Check className="w-3 h-3 inline mr-1" />
                        This was me
                      </button>
                      <button
                        type="button"
                        onClick={() => handleSuspiciousLogin(false)}
                        className="flex-1 text-xs py-1.5 px-3 bg-red-100 dark:bg-red-900/50 text-red-700 dark:text-red-300 rounded hover:bg-red-200 dark:hover:bg-red-800 transition-colors"
                      >
                        <AlertTriangle className="w-3 h-3 inline mr-1" />
                        Not me
                      </button>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>
        )}

        <form onSubmit={handleLogin} className="space-y-5">
          <div>
            <label htmlFor="admin-email" className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1.5">
              Email address
            </label>
            <input
              id="admin-email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="admin@functionfly.com"
              autoComplete="email"
              className="w-full px-4 py-2.5 bg-white dark:bg-slate-900 rounded-lg ring-1 ring-inset ring-gray-300 dark:ring-slate-600 focus:ring-2 focus:ring-blue-500 transition-all text-gray-900 dark:text-white placeholder:text-gray-400 dark:placeholder:text-slate-500 outline-none"
              required
            />
          </div>

          <div>
            <label
              htmlFor="admin-password"
              className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1.5"
            >
              Password
            </label>
            <input
              id="admin-password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              autoComplete="current-password"
              className="w-full px-4 py-2.5 bg-white dark:bg-slate-900 rounded-lg ring-1 ring-inset ring-gray-300 dark:ring-slate-600 focus:ring-2 focus:ring-blue-500 transition-all text-gray-900 dark:text-white placeholder:text-gray-400 dark:placeholder:text-slate-500 outline-none"
              required
            />
          </div>

          {error && (
            <div className="p-3 bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded-lg">
              <p className="text-sm text-red-700 dark:text-red-300">{error}</p>
            </div>
          )}

          <button
            type="submit"
            disabled={isLoading}
            className="w-full py-3 px-4 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium inline-flex items-center justify-center gap-2"
          >
            <LogIn className="w-4 h-4 shrink-0" />
            {isLoading ? 'Signing in...' : 'Sign In'}
          </button>

          <div className="flex items-center justify-between text-xs">
            <button
              type="button"
              onClick={() => {
                setForgotMode((v) => !v);
                setForgotEmail(email);
                setForgotStatus('idle');
                setForgotMessage('');
              }}
              className="text-blue-600 dark:text-blue-400 hover:underline inline-flex items-center gap-1"
            >
              <KeyRound className="w-3 h-3" />
              {forgotMode ? 'Back to sign in' : 'Forgot password?'}
            </button>
            <span className="text-gray-400 dark:text-slate-500">
              Need an account? Contact your super admin.
            </span>
          </div>
        </form>

        {forgotMode && (
          <form
            onSubmit={handleForgotPassword}
            className="mt-6 pt-6 border-t border-gray-100 dark:border-slate-700 space-y-4"
          >
            <div>
              <h2 className="text-sm font-semibold text-gray-900 dark:text-white">
                Reset your password
              </h2>
              <p className="text-xs text-gray-500 dark:text-slate-400 mt-1">
                Enter the email associated with your account. We'll send a reset link.
              </p>
            </div>
            <div>
              <label
                htmlFor="admin-forgot-email"
                className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1.5"
              >
                Email address
              </label>
              <input
                id="admin-forgot-email"
                type="email"
                value={forgotEmail}
                onChange={(e) => setForgotEmail(e.target.value)}
                placeholder="admin@functionfly.com"
                autoComplete="email"
                className="w-full px-4 py-2.5 bg-white dark:bg-slate-900 rounded-lg ring-1 ring-inset ring-gray-300 dark:ring-slate-600 focus:ring-2 focus:ring-blue-500 transition-all text-gray-900 dark:text-white placeholder:text-gray-400 dark:placeholder:text-slate-500 outline-none"
                required
              />
            </div>
            {forgotMessage && (
              <div
                className={`p-3 rounded-lg border text-sm ${
                  forgotStatus === 'sent'
                    ? 'bg-green-50 dark:bg-green-900/30 border-green-200 dark:border-green-800 text-green-700 dark:text-green-300'
                    : 'bg-red-50 dark:bg-red-900/30 border-red-200 dark:border-red-800 text-red-700 dark:text-red-300'
                }`}
              >
                {forgotMessage}
              </div>
            )}
            <button
              type="submit"
              disabled={forgotStatus === 'sending'}
              className="w-full py-2.5 px-4 bg-slate-700 hover:bg-slate-800 text-white rounded-lg disabled:opacity-50 disabled:cursor-not-allowed transition-colors text-sm font-medium"
            >
              {forgotStatus === 'sending' ? 'Sending…' : 'Send reset link'}
            </button>
          </form>
        )}

        <p className="mt-6 pt-6 border-t border-gray-100 dark:border-slate-700 text-center text-xs text-gray-500 dark:text-slate-400">
          Protected admin area · All access is logged
        </p>
      </div>

      {/* MFA challenge modal — appears after password is accepted but
          a TOTP code is required before a real session is issued. */}
      {mfaStep && (
        <div
          role="dialog"
          aria-modal="true"
          aria-labelledby="mfa-modal-title"
          className="fixed inset-0 z-50 bg-black/50 flex items-center justify-center px-4"
        >
          <div className="w-full max-w-sm bg-white dark:bg-slate-800 rounded-2xl border border-gray-200 dark:border-slate-700 shadow-xl p-6 relative">
            <button
              type="button"
              aria-label="Close MFA prompt"
              onClick={() => {
                setMfaStep(null);
                setMfaCode('');
                setMfaError('');
              }}
              className="absolute top-3 right-3 text-gray-400 hover:text-gray-600 dark:hover:text-slate-200"
            >
              <X className="w-4 h-4" />
            </button>
            <Shield className="w-8 h-8 text-blue-600 mb-3" />
            <h2 id="mfa-modal-title" className="text-lg font-semibold text-gray-900 dark:text-white">
              Two-factor required
            </h2>
            <p className="text-sm text-gray-500 dark:text-slate-400 mt-1">
              Enter the 6-digit code from your authenticator app for {mfaStep.email}.
            </p>
            <form onSubmit={handleMfaSubmit} className="mt-5 space-y-4">
              <input
                type="text"
                inputMode="numeric"
                pattern="[0-9]*"
                maxLength={6}
                autoFocus
                value={mfaCode}
                onChange={(e) => setMfaCode(e.target.value.replace(/\D/g, ''))}
                className="w-full text-center tracking-[0.5em] font-mono text-lg py-3 rounded-lg ring-1 ring-inset ring-gray-300 dark:ring-slate-600 focus:ring-2 focus:ring-blue-500 bg-white dark:bg-slate-900 text-gray-900 dark:text-white"
                aria-label="MFA code"
              />
              {mfaError && (
                <div className="p-2.5 bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded-lg text-sm text-red-700 dark:text-red-300">
                  {mfaError}
                </div>
              )}
              <button
                type="submit"
                disabled={mfaCode.length !== 6 || mfaLoading}
                className="w-full py-2.5 px-4 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors text-sm font-medium"
              >
                {mfaLoading ? 'Verifying…' : 'Verify'}
              </button>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
