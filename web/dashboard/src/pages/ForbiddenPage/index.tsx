/**
 * Access Denied Page
 * Shown when a user tries to access a feature they don't have permission for
 */

import { Button } from '@/components/ui/button';
import { useAccessControl, type Permission } from '@/hooks/useAccessControl';
import { Shield, Home, ArrowLeft, Lock, AlertTriangle } from 'lucide-react';
import { Link, useLocation } from 'react-router-dom';

interface AccessDeniedState {
  from?: { pathname?: string };
  permission?: Permission;
  featureName?: string;
}

const PERMISSION_LABELS: Record<string, string> = {
  'functions:read': 'View Functions',
  'functions:write': 'Create/Modify Functions',
  'functions:delete': 'Delete Functions',
  'agents:read': 'View Agents',
  'agents:write': 'Create/Modify Agents',
  'agents:delete': 'Delete Agents',
  'wallet:read': 'View Wallet',
  'wallet:write': 'Manage Wallet',
  'secrets:read': 'View Secrets',
  'secrets:write': 'Create/Modify Secrets',
  'secrets:delete': 'Delete Secrets',
  'teams:read': 'View Teams',
  'teams:write': 'Create/Modify Teams',
  'teams:delete': 'Delete Teams',
  'billing:read': 'View Billing',
  'billing:write': 'Modify Billing',
  'analytics:read': 'View Analytics',
  'statefabric:read': 'View State Fabric',
  'statefabric:write': 'Modify State Fabric',
  'registry:read': 'View Registry',
  'registry:write': 'Publish to Registry',
  'enterprise:read': 'View Enterprise Features',
  'enterprise:write': 'Modify Enterprise Settings',
};

const PERMISSION_MESSAGES: Record<Permission, string> = {
  'functions:read': 'You need at least a Starter plan to view functions.',
  'functions:write': 'You need at least a Starter plan to create or modify functions.',
  'functions:delete': 'You need a Pro plan to delete functions.',
  'agents:read': 'You need at least a Starter plan to view agents.',
  'agents:write': 'You need at least a Starter plan to create or modify agents.',
  'agents:delete': 'You need a Pro plan to delete agents.',
  'wallet:read': 'You need at least a Starter plan to view your wallet.',
  'wallet:write': 'You need at least a Pro plan to manage your wallet.',
  'secrets:read': 'You need at least a Starter plan to view secrets.',
  'secrets:write': 'You need at least a Starter plan to create or modify secrets.',
  'secrets:delete': 'You need a Pro plan to delete secrets.',
  'teams:read': 'You need at least a Starter plan to view teams.',
  'teams:write': 'You need at least a Starter plan to create or modify teams.',
  'teams:delete': 'You need a Pro plan to delete teams.',
  'billing:read': 'You need at least a Starter plan to view billing.',
  'billing:write': 'You need at least a Starter plan to modify billing.',
  'analytics:read': 'You need at least a Starter plan to view analytics.',
  'statefabric:read': 'You need a Pro plan to view State Fabric.',
  'statefabric:write': 'You need a Pro plan to modify State Fabric.',
  'registry:read': 'You need at least a Starter plan to view the registry.',
  'registry:write': 'You need at least a Starter plan to publish to the registry.',
  'enterprise:read': 'You need an Enterprise plan to view enterprise features.',
  'enterprise:write': 'You need an Enterprise plan to modify enterprise settings.',
};

function formatPermission(permission: Permission): string {
  return PERMISSION_LABELS[permission] ?? permission.replace(':', ' ').replace('_', ' ');
}

function getUpgradeMessage(permission: Permission): string {
  return PERMISSION_MESSAGES[permission] ?? 'Upgrade your plan to access this feature.';
}

interface ForbiddenPageProps {
  requiredPermission?: Permission;
  featureName?: string;
}

export function ForbiddenPage({
  requiredPermission,
  featureName,
}: ForbiddenPageProps) {
  const location = useLocation();
  const { currentTier } = useAccessControl();

  // Get state from navigation (passed via ProtectedByPermissionRoute redirect)
  const locationState = location.state as AccessDeniedState | null;
  const fromPath = locationState?.from?.pathname;

  // Use props first, then fall back to location state
  const feature =
    featureName ??
    locationState?.featureName ??
    (locationState?.permission ? formatPermission(locationState.permission) : null) ??
    'this feature';

  const permission = requiredPermission ?? locationState?.permission;
  const upgradeMessage = permission ? getUpgradeMessage(permission) : null;

  const handleGoBack = () => {
    if (fromPath) {
      // Replace current history entry with the previous path
      window.history.replaceState(null, '', fromPath);
      window.history.back();
    } else {
      window.history.back();
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-bg-primary p-4">
      <div className="max-w-md w-full text-center">
        {/* Geometric lock icon */}
        <div className="mb-8">
          <div className="relative inline-block">
            {/* Outer ring */}
            <div className="absolute inset-0 flex items-center justify-center">
              <div className="h-32 w-32 rounded-full border-2 border-error/20" />
            </div>
            {/* Middle ring */}
            <div className="absolute inset-0 flex items-center justify-center">
              <div className="h-24 w-24 rounded-full border-2 border-error/30" />
            </div>
            {/* Inner circle with lock */}
            <div className="relative flex h-20 w-20 items-center justify-center rounded-full bg-error/10">
              <Shield className="h-10 w-10 text-error" />
            </div>
            {/* Decorative dots */}
            <div className="absolute -right-1 top-0 h-2 w-2 rounded-full bg-error/40" />
            <div className="absolute -bottom-0.5 -left-1 h-1.5 w-1.5 rounded-full bg-error/30" />
            <div className="absolute -right-0.5 -bottom-1 h-1 w-1 rounded-full bg-error/20" />
          </div>
        </div>

        {/* Error code badge */}
        <div className="mb-4">
          <span className="inline-block rounded-md bg-error/10 px-3 py-1 text-sm font-semibold tracking-wider text-error uppercase">
            Error 403
          </span>
        </div>

        {/* Title */}
        <h1 className="text-3xl font-bold text-text-primary mb-3">Access Denied</h1>

        {/* Feature being accessed */}
        <p className="text-lg text-text-secondary mb-2">
          You don't have permission to access{' '}
          <span className="font-semibold text-text-primary">{feature}</span>
        </p>

        {/* Current plan indicator */}
        <p className="text-sm text-text-muted mb-6">
          Your current plan: <span className="font-medium capitalize">{currentTier || 'Free'}</span>
        </p>

        {/* Upgrade message */}
        {upgradeMessage && (
          <div className="mb-8 rounded-lg border border-warning/20 bg-warning/10 p-4 text-left">
            <div className="flex items-start gap-3">
              <AlertTriangle className="h-5 w-5 text-warning mt-0.5 shrink-0" />
              <p className="text-sm text-text-secondary">{upgradeMessage}</p>
            </div>
          </div>
        )}

        {/* Action buttons */}
        <div className="space-y-3">
          <Button asChild className="w-full">
            <Link to="/dashboard">
              <Home className="w-4 h-4 mr-2" />
              Go to Dashboard
            </Link>
          </Button>
          <Button
            variant="outline"
            className="w-full"
            onClick={handleGoBack}
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Go Back
          </Button>
        </div>

        {/* Contact support hint */}
        <p className="mt-8 text-xs text-text-muted">
          If you believe this is an error, please contact support.
        </p>
      </div>
    </div>
  );
}

export default ForbiddenPage;
