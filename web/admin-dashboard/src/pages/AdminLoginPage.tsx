/**
 * Admin Login Page
 * Placeholder for authentication
 */

import { bootstrapAdminSession, loginAdmin } from '@/lib/api/adminAuth';
import { adminApiClient } from '@/lib/api/adminClient';
import { trackSecurityEvent } from '@/lib/monitoring/securityEvents';
import { generateDeviceFingerprint } from '@/lib/security/deviceFingerprint';
import { useAdminAuthStore } from '@/stores/adminAuthStore';
import { AlertTriangle, Check, LogIn, Shield } from 'lucide-react';
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

      const fingerprint = await generateDeviceFingerprint();
      const auth = await loginAdmin(email, password);

      if (!auth.token) {
        throw new Error('Login succeeded but no access token was returned');
      }

      adminApiClient.setSessionToken(auth.token);
      adminApiClient.setDeviceFingerprint(fingerprint);

      const bootstrap = await bootstrapAdminSession(auth.token);

      // Check if this is a suspicious login (new device/IP)
      const lastLoginInfo: LastLoginDisplay | undefined = lastLogin?.suspicious
        ? { ...lastLogin, suspicious: false }
        : undefined;

      login(bootstrap.session, bootstrap.user, lastLoginInfo);
      setDeviceFingerprint(fingerprint);
      setLastLoginInfo(lastLoginInfo || null);

      // Store token in regular localStorage for cross-app compatibility
      try {
        localStorage.setItem('ffly_jwt', auth.token);
      } catch {
        /* localStorage may be unavailable */
      }

      navigate('/');
    } catch (err) {
      trackSecurityEvent('login_failed', { email });
      const message = err instanceof Error ? err.message : 'Login failed. Please try again.';
      setError(message);
    } finally {
      setIsLoading(false);
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
        <svg
          width="40"
          height="40"
          viewBox="0 0 32 32"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <rect width="32" height="32" rx="6" fill="#0F172A" />
          <path d="M16 5L27 16L16 27L5 16L16 5Z" fill="#6366F1" />
          <path d="M16 9.5L22.5 16L16 22.5L9.5 16L16 9.5Z" fill="white" />
          <path d="M16 12.5L19.5 16L16 19.5L12.5 16L16 12.5Z" fill="#6366F1" />
        </svg>
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
        </form>

        <p className="mt-6 pt-6 border-t border-gray-100 dark:border-slate-700 text-center text-xs text-gray-500 dark:text-slate-400">
          Protected admin area · All access is logged
        </p>
      </div>
    </div>
  );
}
