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
import { useNavigate } from 'react-router-dom';

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

  // Fetch last login info on mount (if user has a session cookie)
  useEffect(() => {
    async function fetchLastLogin() {
      try {
        // Try to get last login info from the API
        const resp = await adminApiClient.get<LastLoginDisplay>('/auth/last-login');
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
    <div className="w-full max-w-md">
      {/* Single card: title + form */}
      <div className="bg-white rounded-2xl border border-gray-200 shadow-lg shadow-gray-200/50 p-8 sm:p-10">
        <div className="mb-8">
          <h1 className="text-2xl font-semibold tracking-tight text-gray-900">Sign in</h1>
          <p className="mt-2 text-sm text-gray-500">
            Use your admin credentials to access the dashboard.
          </p>
        </div>

        {/* Last successful login info */}
        {lastLogin && !suspiciousConfirmed && (
          <div
            className={`mb-6 p-4 rounded-lg border ${
              lastLogin.suspicious ? 'bg-amber-50 border-amber-200' : 'bg-green-50 border-green-200'
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
                    lastLogin.suspicious ? 'text-amber-800' : 'text-green-800'
                  }`}
                >
                  {lastLogin.suspicious ? 'Unusual login detected' : 'Last successful login'}
                </p>
                <p className="text-xs text-gray-600 mt-1">
                  {lastLogin.device_name} · {lastLogin.ip_address}
                </p>
                <p className="text-xs text-gray-500 mt-0.5">
                  {new Date(lastLogin.timestamp).toLocaleString()}
                </p>

                {lastLogin.suspicious && (
                  <div className="mt-3 pt-3 border-t border-amber-200">
                    <p className="text-xs text-amber-700 mb-2">
                      Was this you? If not, your account may be compromised.
                    </p>
                    <div className="flex gap-2">
                      <button
                        type="button"
                        onClick={() => handleSuspiciousLogin(true)}
                        className="flex-1 text-xs py-1.5 px-3 bg-green-100 text-green-700 rounded hover:bg-green-200 transition-colors"
                      >
                        <Check className="w-3 h-3 inline mr-1" />
                        This was me
                      </button>
                      <button
                        type="button"
                        onClick={() => handleSuspiciousLogin(false)}
                        className="flex-1 text-xs py-1.5 px-3 bg-red-100 text-red-700 rounded hover:bg-red-200 transition-colors"
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
            <label htmlFor="admin-email" className="block text-sm font-medium text-gray-700 mb-1.5">
              Email address
            </label>
            <input
              id="admin-email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="admin@functionfly.com"
              autoComplete="email"
              className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all text-gray-900 placeholder:text-gray-400"
              required
            />
          </div>

          <div>
            <label
              htmlFor="admin-password"
              className="block text-sm font-medium text-gray-700 mb-1.5"
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
              className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all text-gray-900 placeholder:text-gray-400"
              required
            />
          </div>

          {error && (
            <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
              <p className="text-sm text-red-700">{error}</p>
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

        <p className="mt-6 pt-6 border-t border-gray-100 text-center text-xs text-gray-500">
          Protected admin area · All access is logged
        </p>
      </div>
    </div>
  );
}
