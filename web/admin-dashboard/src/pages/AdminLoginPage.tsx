/**
 * Admin Login Page
 * Placeholder for authentication
 */

import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAdminAuthStore } from '@/stores/adminAuthStore';
import { LogIn } from 'lucide-react';
import { generateDeviceFingerprint } from '@/lib/security/deviceFingerprint';
import { adminApiClient } from '@/lib/api/adminClient';
import { bootstrapAdminSession, loginAdmin } from '@/lib/api/adminAuth';

export function AdminLoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const navigate = useNavigate();
  const { login, setDeviceFingerprint } = useAdminAuthStore();

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

      login(bootstrap.session, bootstrap.user);
      setDeviceFingerprint(fingerprint);
      navigate('/');
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Login failed. Please try again.';
      setError(message);
    } finally {
      setIsLoading(false);
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
            <label htmlFor="admin-password" className="block text-sm font-medium text-gray-700 mb-1.5">
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
