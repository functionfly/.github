/**
 * Access Denied Page
 * Shown when a user tries to access a feature they don't have permission for
 */

import { Button } from '@/components/ui/Button';
import { useAccessControl, type Permission } from '@/hooks/useAccessControl';
import { useLocation, useNavigate } from 'react-router-dom';

interface AccessDeniedState {
  from?: { pathname?: string };
  permission?: Permission;
  featureName?: string;
}

const PERMISSION_LABELS: Record<string, string> = {
  'functions:read': 'View Functions',
  'functions:write': 'Create/Modify Functions',
  'functions:delete': 'Delete Functions',
  'users:read': 'View Users',
  'users:write': 'Create/Modify Users',
  'users:delete': 'Delete Users',
  'billing:read': 'View Billing',
  'billing:write': 'Modify Billing',
  'tenants:read': 'View Tenants',
  'tenants:write': 'Create/Modify Tenants',
  'tenants:delete': 'Delete Tenants',
  'audit:read': 'View Audit Logs',
  'registry:read': 'View Registry',
  'registry:write': 'Publish to Registry',
  'features:read': 'View Feature Flags',
  'features:write': 'Modify Feature Flags',
  'system:read': 'View System Settings',
  'system:write': 'Modify System Settings',
};

function formatPermission(permission: Permission): string {
  return PERMISSION_LABELS[permission] ?? permission.replace(':', ' ').replace('_', ' ');
}

interface AdminAccessDeniedPageProps {
  requiredPermission?: Permission;
  featureName?: string;
  onBackToDashboard?: boolean;
}

export function AdminAccessDeniedPage({
  requiredPermission,
  featureName,
  onBackToDashboard = true,
}: AdminAccessDeniedPageProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const { currentRole } = useAccessControl();

  // Get state from navigation (passed via ProtectedRoute redirect)
  const locationState = location.state as AccessDeniedState | null;
  const fromPath = locationState?.from?.pathname;

  // Use props first, then fall back to location state
  const feature =
    featureName ??
    locationState?.featureName ??
    (locationState?.permission ? formatPermission(locationState.permission) : null) ??
    requiredPermission ??
    'this feature';

  const permission = requiredPermission ?? locationState?.permission;
  void permission; // May be used for future enhanced messaging

  const handleGoBack = () => {
    if (fromPath) {
      navigate(fromPath, { replace: true });
    } else if (onBackToDashboard) {
      navigate('/', { replace: true });
    } else {
      navigate(-1);
    }
  };

  const handleContactSupport = () => {
    navigate('/support');
  };

  return (
    <div className="min-h-[calc(100vh-4rem)] flex items-center justify-center px-4">
      <div className="w-full max-w-md text-center">
        {/* Geometric lock icon */}
        <div className="mb-8 flex justify-center">
          <div className="relative">
            {/* Outer ring */}
            <div className="absolute inset-0 flex items-center justify-center">
              <div className="h-32 w-32 rounded-full border-2 border-red-200 opacity-60" />
            </div>
            {/* Middle ring */}
            <div className="absolute inset-0 flex items-center justify-center">
              <div className="h-24 w-24 rounded-full border-2 border-red-300 opacity-80" />
            </div>
            {/* Inner circle with lock */}
            <div className="relative flex h-20 w-20 items-center justify-center rounded-full bg-gradient-to-br from-red-50 to-red-100 shadow-sm">
              <svg
                className="h-10 w-10 text-red-500"
                fill="none"
                viewBox="0 0 24 24"
                strokeWidth={1.5}
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M12 9v3.75m0-3.75h.008v.008H12v-.008zM21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 12h.01M15 12h.01" />
              </svg>
            </div>
            {/* Decorative dots */}
            <div className="absolute -right-1 top-0 h-2 w-2 rounded-full bg-red-400" />
            <div className="absolute -bottom-0.5 -left-1 h-1.5 w-1.5 rounded-full bg-red-300" />
            <div className="absolute -right-0.5 -bottom-1 h-1 w-1 rounded-full bg-red-200" />
          </div>
        </div>

        {/* Error code */}
        <div className="mb-3">
          <span className="inline-block rounded-md bg-red-50 px-2 py-1 text-xs font-semibold tracking-wider text-red-600 uppercase">
            Error 403
          </span>
        </div>

        {/* Title */}
        <h1 className="mb-3 font-display text-2xl font-semibold text-gray-900">Access Denied</h1>

        {/* Description */}
        <p className="mb-2 text-sm text-gray-600">
          You don't have permission to access{' '}
          <span className="font-medium text-gray-900">{feature}</span>.
        </p>
        {currentRole && (
          <p className="mb-6 text-xs text-gray-500">
            Your current role: <span className="font-medium">{currentRole}</span>
          </p>
        )}

        {/* Decorative divider */}
        <div className="mb-8 flex items-center justify-center gap-2">
          <div className="h-px w-12 bg-gradient-to-r from-transparent to-gray-200" />
          <div className="h-1.5 w-1.5 rounded-full bg-red-300" />
          <div className="h-px w-12 bg-gradient-to-l from-transparent to-gray-200" />
        </div>

        {/* Help text */}
        <div className="mb-8 rounded-lg border border-gray-200 bg-gray-50 p-4">
          <p className="text-xs text-gray-600">
            If you believe this is an error, please contact your administrator or request access
            through the support portal.
          </p>
        </div>

        {/* Action buttons */}
        <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-center">
          <Button
            variant="outline"
            size="sm"
            onClick={handleGoBack}
            className="border-gray-300 text-gray-700 hover:bg-gray-50"
          >
            <svg
              className="mr-1.5 h-3.5 w-3.5"
              fill="none"
              viewBox="0 0 24 24"
              strokeWidth={2}
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M10.5 19.5L3 12m0 0l7.5-7.5M3 12h18"
              />
            </svg>
            {fromPath ? 'Go Back' : 'Back to Dashboard'}
          </Button>
          <Button
            variant="primary"
            size="sm"
            onClick={handleContactSupport}
            className="bg-gray-900 text-white hover:bg-gray-800"
          >
            Contact Support
            <svg
              className="ml-1.5 h-3.5 w-3.5"
              fill="none"
              viewBox="0 0 24 24"
              strokeWidth={2}
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M13.5 4.5L21 12m0 0l-7.5 7.5M21 12H3"
              />
            </svg>
          </Button>
        </div>

        {/* Subtle pattern overlay */}
        <div
          className="pointer-events-none absolute inset-0 z-[-1] opacity-[0.015]"
          style={{
            backgroundImage: `url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23000000' fill-opacity='1'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E")`,
          }}
        />
      </div>
    </div>
  );
}

export default AdminAccessDeniedPage;
