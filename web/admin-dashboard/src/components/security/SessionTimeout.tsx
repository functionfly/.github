/**
 * Session Timeout Warning Component
 * Renders in bottom-right; Extend Session calls backend to issue a new JWT and updates the store.
 */

import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAdminAuthStore } from '@/stores/adminAuthStore';
import { Clock } from 'lucide-react';

const WARNING_TIME = 5 * 60 * 1000; // 5 minutes before expiry

export function SessionTimeoutWarning() {
  const [showWarning, setShowWarning] = useState(false);
  const [timeRemaining, setTimeRemaining] = useState(0);
  const [extending, setExtending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { session, logout, extendSession } = useAdminAuthStore();
  const navigate = useNavigate();

  useEffect(() => {
    if (!session) return;

    const checkTimeout = () => {
      const now = Date.now();
      const sessionExpiresAt = new Date(session.expires_at).getTime();
      const remaining = sessionExpiresAt - now;

      if (remaining <= 0) {
        logout();
        navigate('/auth/login?reason=session_expired');
        return;
      }

      if (remaining <= WARNING_TIME) {
        setShowWarning(true);
        setTimeRemaining(remaining);
      } else {
        setShowWarning(false);
      }
    };

    checkTimeout();
    const interval = setInterval(checkTimeout, 1000);

    return () => clearInterval(interval);
  }, [session, logout, navigate]);

  const handleExtendSession = async () => {
    setError(null);
    setExtending(true);
    try {
      await extendSession();
      setShowWarning(false);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to extend session');
    } finally {
      setExtending(false);
    }
  };

  if (!showWarning) return null;

  const minutes = Math.floor(timeRemaining / 60000);
  const seconds = Math.floor((timeRemaining % 60000) / 1000);

  return (
    <div className="fixed bottom-4 right-4 bg-yellow-50 border border-yellow-200 rounded-lg p-4 shadow-lg z-50 max-w-sm">
      <div className="flex items-start gap-3">
        <Clock className="w-6 h-6 text-yellow-600 shrink-0" />
        <div className="flex-1">
          <h3 className="font-semibold text-yellow-900 mb-1">
            Session Expiring Soon
          </h3>
          <p className="text-sm text-yellow-800 mb-3">
            Your session will expire in {minutes}:{seconds.toString().padStart(2, '0')}.
          </p>
          {error && (
            <p className="text-sm text-red-600 mb-2">{error}</p>
          )}
          <button
            onClick={handleExtendSession}
            disabled={extending}
            className="w-full py-2 px-4 bg-yellow-600 text-white rounded hover:bg-yellow-700 transition-colors text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {extending ? 'Extending…' : 'Extend Session'}
          </button>
        </div>
      </div>
    </div>
  );
}
