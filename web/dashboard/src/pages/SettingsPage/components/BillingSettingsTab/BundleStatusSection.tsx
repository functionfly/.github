import { useNavigate } from 'react-router-dom';
import { Package, CheckCircle, AlertCircle, Loader2, ArrowRight, ExternalLink } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { useProvisioningStatus } from '@/hooks/useProvisioning';

const BUNDLE_NAMES: Record<string, string> = {
  'saas-starter': 'SaaS Starter',
  'marketplace': 'Marketplace',
  'ai-app': 'AI App',
};

const BUNDLE_PRICES: Record<string, string> = {
  'saas-starter': '$49/mo',
  'marketplace': '$49/mo',
  'ai-app': '$39/mo',
};

export function BundleStatusSection() {
  const navigate = useNavigate();
  const { data: statusData, isLoading } = useProvisioningStatus();

  if (isLoading) {
    return (
      <Card className="ff-card-velocity">
        <CardContent className="flex items-center justify-center p-6">
          <Loader2 className="h-5 w-5 animate-spin text-zinc-400" />
        </CardContent>
      </Card>
    );
  }

  // No bundle provisioned
  if (!statusData || !('bundle_slug' in statusData)) {
    return (
      <Card className="ff-card-velocity">
        <CardHeader>
          <CardTitle className="font-display flex items-center gap-2">
            <Package className="h-5 w-5" />
            Backend-in-a-Box Bundle
          </CardTitle>
          <CardDescription className="text-text-secondary">
            Get a complete production backend in one click
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between rounded-lg border border-zinc-200 bg-zinc-50 p-4 dark:border-zinc-800 dark:bg-zinc-900">
            <div>
              <p className="text-sm text-zinc-600 dark:text-zinc-400">
                No bundle active. Get Auth, Payments, Email, Analytics — provisioned in seconds.
              </p>
            </div>
            <Button
              size="sm"
              onClick={() => navigate('/bundles')}
            >
              View Bundles
              <ArrowRight className="ml-2 h-4 w-4" />
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  // Bundle is active or provisioning
  const { bundle_slug, status, components, duration_ms } = statusData;
  const bundleName = BUNDLE_NAMES[bundle_slug] || bundle_slug;
  const bundlePrice = BUNDLE_PRICES[bundle_slug] || '';
  const isActive = status === 'active';
  const isProvisioning = status === 'provisioning';
  const hasFailures = status === 'failed' || Object.values(components || {}).some(c => c.status === 'failed');
  const componentCount = components ? Object.keys(components).length : 0;
  const activeCount = components ? Object.values(components).filter(c => c.status === 'active').length : 0;

  return (
    <Card className="ff-card-velocity">
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="font-display flex items-center gap-2">
              <Package className="h-5 w-5" />
              {bundleName} Bundle
            </CardTitle>
            <CardDescription className="text-text-secondary">
              {bundlePrice} — All infrastructure provisioned and isolated
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            {isActive && (
              <span className="inline-flex items-center gap-1.5 rounded-full bg-emerald-50 px-3 py-1 text-xs font-medium text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-400">
                <CheckCircle className="h-3.5 w-3.5" />
                Active
              </span>
            )}
            {isProvisioning && (
              <span className="inline-flex items-center gap-1.5 rounded-full bg-blue-50 px-3 py-1 text-xs font-medium text-blue-700 dark:bg-blue-500/10 dark:text-blue-400">
                <Loader2 className="h-3.5 w-3.5 animate-spin" />
                Provisioning
              </span>
            )}
            {hasFailures && (
              <span className="inline-flex items-center gap-1.5 rounded-full bg-amber-50 px-3 py-1 text-xs font-medium text-amber-700 dark:bg-amber-500/10 dark:text-amber-400">
                <AlertCircle className="h-3.5 w-3.5" />
                Partial
              </span>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {/* Component status summary */}
          <div className="flex items-center gap-4 text-sm">
            <span className="text-zinc-600 dark:text-zinc-400">
              {activeCount}/{componentCount} components active
            </span>
            {duration_ms > 0 && (
              <span className="text-zinc-500 dark:text-zinc-500">
                Provisioned in {(duration_ms / 1000).toFixed(1)}s
              </span>
            )}
          </div>

          {/* Component chips */}
          <div className="flex flex-wrap gap-2">
            {components && Object.entries(components).map(([key, comp]) => (
              <span
                key={key}
                className={`inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium ${
                  comp.status === 'active'
                    ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-400'
                    : comp.status === 'failed'
                    ? 'bg-red-50 text-red-700 dark:bg-red-500/10 dark:text-red-400'
                    : 'bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-400'
                }`}
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
            >
              View Details
            </Button>
            {hasFailures && (
              <Button
                size="sm"
                variant="outline"
                onClick={() => navigate(`/bundles/provisioning?bundle=${bundle_slug}&retry=true`)}
              >
                Retry Failed
              </Button>
            )}
            <Button
              size="sm"
              variant="ghost"
              onClick={() => navigate('/bundles')}
            >
              <ExternalLink className="mr-1.5 h-3.5 w-3.5" />
              Change Bundle
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
