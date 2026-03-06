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
    <div className="min-h-screen bg-linear-to-br from-gray-900 to-gray-800 flex items-center justify-center px-4">
      <div className="w-full max-w-md">
        {/* Logo */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-blue-600 rounded-lg mb-4">
            <LogIn className="w-8 h-8 text-white" />
          </div>
          <h1 className="text-3xl font-bold text-white">FlyAdmin</h1>
          <p className="mt-2 text-gray-400">Admin Dashboard</p>
        </div>

        {/* Login Form */}
        <div className="bg-white rounded-lg shadow-xl p-8">
          <form onSubmit={handleLogin} className="space-y-6">
            {/* Email */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Email Address
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="admin@functionfly.com"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
                required
              />
            </div>

            {/* Password */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Password
              </label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
                required
              />
            </div>

            {/* Error Message */}
            {error && (
              <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
                <p className="text-sm text-red-700">{error}</p>
              </div>
            )}

            {/* Submit Button */}
            <button
              type="submit"
              disabled={isLoading}
              className="w-full py-2 px-4 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium"
            >
              {isLoading ? 'Signing in...' : 'Sign In'}
            </button>
          </form>

          {/* Footer */}
          <div className="mt-6 pt-6 border-t border-gray-200">
            <p className="text-center text-xs text-gray-500">
              Protected Admin Area • All access is logged
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
