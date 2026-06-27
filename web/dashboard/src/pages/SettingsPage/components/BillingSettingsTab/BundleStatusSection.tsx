import { Button } from '@/components/ui/button';
import { useProvisioningStatus } from '@/hooks/useProvisioning';
import { AlertCircle, ArrowRight, CheckCircle, ExternalLink, Loader2, Package } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

const BUNDLE_NAMES: Record<string, string> = {
  'saas-starter': 'SaaS Starter',
  marketplace: 'Marketplace',
  'ai-app': 'AI App',
};

const BUNDLE_PRICES: Record<string, string> = {
  'saas-starter': '$49/mo',
  marketplace: '$49/mo',
  'ai-app': '$39/mo',
};

export function BundleStatusSection() {
  const navigate = useNavigate();
  const { data: statusData, isLoading } = useProvisioningStatus();

  if (isLoading) {
    return (
      <div
        className="rounded-lg p-5 flex items-center justify-center"
        style={{
          background: 'var(--panel)',
          border: '1px solid var(--panel-edge)',
          boxShadow: 'var(--shadow-chamber)',
        }}
      >
        <Loader2 className="h-5 w-5 animate-spin" style={{ color: 'var(--text-dim)' }} />
      </div>
    );
  }

  // No bundle provisioned
  if (!statusData || !('bundle_slug' in statusData)) {
    return (
      <div
        className="rounded-lg p-5"
        style={{
          background: 'var(--panel)',
          border: '1px solid var(--panel-edge)',
          boxShadow: 'var(--shadow-chamber)',
        }}
      >
        <div className="mb-4">
          <h3
            className="font-display flex items-center gap-2 text-lg font-semibold"
            style={{ color: 'var(--text)' }}
          >
            <Package className="h-5 w-5" style={{ color: 'var(--accent)' }} />
            Backend-in-a-Box Bundle
          </h3>
          <p className="text-sm mt-1" style={{ color: 'var(--text-dim)' }}>
            Get a complete production backend in one click
          </p>
        </div>
        <div
          className="flex items-center justify-between rounded-lg p-4"
          style={{ background: 'var(--panel-raised)', border: '1px solid var(--panel-edge)' }}
        >
          <div>
            <p className="text-sm" style={{ color: 'var(--text-dim)' }}>
              No bundle active. Get Auth, Payments, Email, Analytics — provisioned in seconds.
            </p>
          </div>
          <Button
            size="sm"
            onClick={() => navigate('/bundles')}
            style={{
              background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
              color: 'var(--text-on-light)',
              boxShadow: 'var(--shadow-btn-primary-rest)',
            }}
          >
            View Bundles
            <ArrowRight className="ml-2 h-4 w-4" />
          </Button>
        </div>
      </div>
    );
  }

  // Bundle is active or provisioning
  const { bundle_slug, status, components, duration_ms } = statusData;
  const bundleName = BUNDLE_NAMES[bundle_slug] || bundle_slug;
  const bundlePrice = BUNDLE_PRICES[bundle_slug] || '';
  const isActive = status === 'active';
  const isProvisioning = status === 'provisioning';
  const hasFailures =
    status === 'failed' || Object.values(components || {}).some((c) => c.status === 'failed');
  const componentCount = components ? Object.keys(components).length : 0;
  const activeCount = components
    ? Object.values(components).filter((c) => c.status === 'active').length
    : 0;

  return (
    <div
      className="rounded-lg p-5"
      style={{
        background: 'var(--panel)',
        border: '1px solid var(--panel-edge)',
        boxShadow: 'var(--shadow-chamber)',
      }}
    >
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3
            className="font-display flex items-center gap-2 text-lg font-semibold"
            style={{ color: 'var(--text)' }}
          >
            <Package className="h-5 w-5" style={{ color: 'var(--accent)' }} />
            {bundleName} Bundle
          </h3>
          <p className="text-sm mt-1" style={{ color: 'var(--text-dim)' }}>
            {bundlePrice} — All infrastructure provisioned and isolated
          </p>
        </div>
        <div className="flex items-center gap-2">
          {isActive && (
            <span
              className="inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-medium"
              style={{ background: 'rgba(143, 255, 208, 0.06)', color: 'var(--status-ok)' }}
            >
              <CheckCircle className="h-3.5 w-3.5" />
              Active
            </span>
          )}
          {isProvisioning && (
            <span
              className="inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-medium"
              style={{ background: 'rgba(59, 130, 246, 0.1)', color: 'var(--accent)' }}
            >
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
              Provisioning
            </span>
          )}
          {hasFailures && (
            <span
              className="inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-medium"
              style={{ background: 'rgba(232, 196, 104, 0.06)', color: 'var(--status-pending)' }}
            >
              <AlertCircle className="h-3.5 w-3.5" />
              Partial
            </span>
          )}
        </div>
      </div>

      <div className="space-y-3">
        {/* Component status summary */}
        <div className="flex items-center gap-4 text-sm">
          <span style={{ color: 'var(--text-dim)' }}>
            {activeCount}/{componentCount} components active
          </span>
          {duration_ms > 0 && (
            <span style={{ color: 'var(--text-faint)' }}>
              Provisioned in {(duration_ms / 1000).toFixed(1)}s
            </span>
          )}
        </div>

        {/* Component chips */}
        <div className="flex flex-wrap gap-2">
          {components &&
            Object.entries(components).map(([key, comp]) => (
              <span
                key={key}
                className="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium"
                style={{
                  background:
                    comp.status === 'active'
                      ? 'rgba(143, 255, 208, 0.06)'
                      : comp.status === 'failed'
                        ? 'rgba(255, 107, 107, 0.06)'
                        : 'var(--panel-raised)',
                  color:
                    comp.status === 'active'
                      ? 'var(--status-ok)'
                      : comp.status === 'failed'
                        ? 'var(--status-revoked)'
                        : 'var(--text-dim)',
                }}
              >
                {comp.status === 'active' ? '✓' : comp.status === 'failed' ? '✗' : '○'}
                {key.replace(/_/g, ' ')}
              </span>
            ))}
        </div>

        {/* Actions */}
        <div className="flex items-center gap-3 pt-2">
          <Button
            size="sm"
            variant="outline"
            onClick={() => navigate(`/bundles/provisioning?bundle=${bundle_slug}`)}
            style={{ borderColor: 'var(--steel)', color: 'var(--text)' }}
          >
            View Details
          </Button>
          {hasFailures && (
            <Button
              size="sm"
              variant="outline"
              onClick={() => navigate(`/bundles/provisioning?bundle=${bundle_slug}&retry=true`)}
              style={{ borderColor: 'var(--steel)', color: 'var(--text)' }}
            >
              Retry Failed
            </Button>
          )}
          <Button
            size="sm"
            variant="ghost"
            onClick={() => navigate('/bundles')}
            style={{ color: 'var(--text-dim)' }}
          >
            <ExternalLink className="mr-1.5 h-3.5 w-3.5" />
            Change Bundle
          </Button>
        </div>
      </div>
    </div>
  );
}
