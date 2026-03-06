import { useEffect, useState } from 'react';
import { ShieldAlert } from 'lucide-react';
import { SECURITY } from '@/lib/constants';
import { checkAdminIPAccess } from '@/lib/security/adminSecurity';
import { useAdminAuthStore } from '@/stores/adminAuthStore';

export function AdminSecurityGate({ children }: { children: React.ReactNode }) {
  const [checking, setChecking] = useState(true);
  const [blockedReason, setBlockedReason] = useState<string | null>(null);
  const setIpAllowed = useAdminAuthStore((s) => s.setIpAllowed);

  useEffect(() => {
    let mounted = true;

    async function runChecks() {
      if (!SECURITY.ENABLE_IP_WHITELIST) {
        setIpAllowed(true, undefined);
        if (mounted) setChecking(false);
        return;
      }

      const result = await checkAdminIPAccess();
      setIpAllowed(result.allowed, result.reason);

      if (!mounted) return;

      if (!result.allowed) {
        setBlockedReason(result.reason || 'ip_not_whitelisted');
      }

      setChecking(false);
    }

    void runChecks();

    return () => {
      mounted = false;
    };
  }, [setIpAllowed]);

  if (checking) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="w-10 h-10 border-4 border-gray-200 border-t-blue-600 rounded-full animate-spin mx-auto" />
          <p className="mt-3 text-sm text-gray-600">Running admin security checks...</p>
        </div>
      </div>
    );
  }

  if (blockedReason) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <div className="max-w-md w-full bg-white border border-red-200 rounded-lg p-6">
          <div className="flex items-center gap-3">
            <ShieldAlert className="w-7 h-7 text-red-600" />
            <h1 className="text-lg font-semibold text-gray-900">Access blocked</h1>
          </div>
          <p className="mt-4 text-sm text-gray-700">
            This admin session was blocked by security policy.
          </p>
          <p className="mt-2 text-xs text-gray-500">Reason: {blockedReason}</p>
        </div>
      </div>
    );
  }

  return <>{children}</>;
}
