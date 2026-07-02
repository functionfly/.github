import { Button } from '@/components/ui/button';
import { useProvisioningStatus } from '@/hooks/useProvisioning';
import { Chamber, CornerBrace, FrameButton, SealedButton, StatusPill } from '@/components/sc';
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
      <Chamber>
        <div className="flex items-center justify-center p-5">
          <Loader2 className="h-5 w-5 animate-spin" style={{ color: 'var(--text-dim)' }} />
        </div>
      </Chamber>
    );
  }

  // No bundle provisioned
  if (!statusData || !('bundle_slug' in statusData)) {
    return (
      <Chamber>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
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
          <SealedButton size="sm" onClick={() => navigate('/bundles')} iconRight={<ArrowRight className="h-4 w-4" />}>
            View Bundles
          </SealedButton>
        </div>
      </Chamber>
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
    <Chamber>
      <CornerBrace position="tl" />
      <CornerBrace position="br" />
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
            <StatusPill status="live" label="Active" />
          )}
          {isProvisioning && (
            <StatusPill status="pending" label="Provisioning" />
          )}
          {hasFailures && (
            <StatusPill status="revoked" label="Partial" />
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
              Provisioned in {duration_ms < 1000 ? `${duration_ms}ms` : `${(duration_ms / 1000).toFixed(1)}s`}
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
          <FrameButton size="sm" onClick={() => navigate(`/bundles/provisioning?bundle=${bundle_slug}`)}>
            View Details
          </FrameButton>
          {hasFailures && (
            <FrameButton size="sm" onClick={() => navigate(`/bundles/provisioning?bundle=${bundle_slug}&retry=true`)}>
              Retry Failed
            </FrameButton>
          )}
          <FrameButton size="sm" onClick={() => navigate('/bundles')} iconLeft={<ExternalLink className="h-3.5 w-3.5" />}>
            Change Bundle
          </FrameButton>
        </div>
      </div>
    </Chamber>
  );
}
