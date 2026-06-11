import { useQuery } from '@tanstack/react-query';
import { Shield } from 'lucide-react';
import { useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';
import {
  createTrustAPICheckout,
  formatCurrency,
  formatNumber,
  getTrustAPIErrorMessage,
  getTrustAPITierPricing,
  getTrustAPIBillingStatus,
  getTrustAPIUsageReport,
  getTrustAPIInvoices,
  type TrustAPITierPricing,
  type TrustAPIBillingStatus,
  type TrustAPIUsageReport,
  type TrustAPIInvoice,
} from '@/api/trustapi';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { CollapsibleSection } from '@/components/ui/collapsible-section';
import { Progress } from '@/components/ui/progress';
import { Zap } from 'lucide-react';
import { formatDate } from '../../settings-utils';

export interface TrustAPISettingsTabProps {
  returnUrl: string;
}

interface TierOption {
  id: string;
  name: string;
  price: number;
  includedRequests: number;
  isCurrent: boolean;
  isUpgrade: boolean;
  isDowngrade: boolean;
}

export function TrustAPISettingsTab({ returnUrl }: TrustAPISettingsTabProps) {
  const [searchParams, setSearchParams] = useSearchParams();

  const {
    data: tiersData,
    isLoading: tiersLoading,
    error: tiersError,
  } = useQuery({
    queryKey: ['trustapi', 'tiers'],
    queryFn: getTrustAPITierPricing,
    retry: false,
  });

  const {
    data: billingStatus,
    isLoading: billingLoading,
    error: billingError,
  } = useQuery({
    queryKey: ['trustapi', 'billing'],
    queryFn: async () => {
      try {
        return await getTrustAPIBillingStatus();
      } catch (e: unknown) {
        const status = (e as { response?: { status?: number } })?.response?.status;
        if (status === 404) return null;
        throw e;
      }
    },
    retry: false,
  });

  const {
    data: usageData,
    isLoading: usageLoading,
  } = useQuery({
    queryKey: ['trustapi', 'usage'],
    queryFn: async () => {
      try {
        return await getTrustAPIUsageReport();
      } catch {
        return null;
      }
    },
    retry: false,
  });

  const {
    data: invoicesData,
    isLoading: invoicesLoading,
  } = useQuery({
    queryKey: ['trustapi', 'invoices'],
    queryFn: () => getTrustAPIInvoices(10, 0),
    retry: false,
  });

  const tiers = tiersData?.tiers ?? [];
  const billing = billingStatus as TrustAPIBillingStatus | null | undefined;
  const invoices = (invoicesData as { invoices: TrustAPIInvoice[] })?.invoices ?? [];

  const tierOptions = buildTierOptions(tiers, billing);

  const handleCheckout = async (tier: string) => {
    try {
      const returnUrlObj = new URL(returnUrl);
      returnUrlObj.searchParams.set('trustApiCheckout', tier);
      const { url } = await createTrustAPICheckout(
        tier,
        returnUrlObj.toString(),
        returnUrl
      );
      window.location.href = url;
    } catch (e: unknown) {
      toast.error(getTrustAPIErrorMessage(e, 'Failed to create checkout session'));
    }
  };

  const usagePercent = billing
    ? Math.min((billing.current_usage / billing.included_requests) * 100, 100)
    : 0;
  const isOverLimit = billing && billing.current_usage > billing.included_requests;

  return (
    <div className="space-y-6">
      <Card className="ff-card-velocity">
        <CardHeader>
          <CardTitle className="font-display flex items-center gap-2">
            <Shield className="h-5 w-5 text-brand-500" />
            Trust API
          </CardTitle>
          <CardDescription className="text-text-secondary">
            Manage your Trust API partner account, billing, and API credentials
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {billingLoading || tiersLoading ? (
            <div className="flex items-center justify-center p-4">
              <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-brand-500" />
            </div>
          ) : billingError && !billing ? (
            <div className="p-4 rounded-lg bg-amber-500/10 border border-amber-500/20">
              <p className="text-amber-500 text-sm">
                {getTrustAPIErrorMessage(billingError, 'Failed to load billing status')}
              </p>
            </div>
          ) : billing ? (
            <div className="space-y-6">
              <div className="flex items-center justify-between p-4 rounded-lg bg-linear-to-r from-brand-500/10 to-brand-600/10 border border-border-default">
                <div>
                  <h3 className="font-semibold font-display text-text-primary capitalize">
                    {billing.tier} Plan
                  </h3>
                  <div className="flex items-center gap-2 mt-1">
                    <Badge
                      variant={billing.billing_status === 'active' ? 'success' : 'secondary'}
                      className={billing.billing_status === 'active' ? 'ff-badge-success' : ''}
                    >
                      {billing.billing_status}
                    </Badge>
                    {billing.is_founder_mode && (
                      <Badge variant="outline" className="border-purple-500/50 text-purple-400">
                        Founder Mode
                      </Badge>
                    )}
                  </div>
                </div>
                <Badge variant="success" className="ff-badge-primary font-semibold px-3 py-1">
                  Current
                </Badge>
              </div>

              <div className="grid grid-cols-2 gap-4 p-4 rounded-lg bg-bg-secondary border border-border-default">
                <div className="flex items-center gap-3">
                  <div>
                    <p className="text-sm text-text-muted">Monthly Price</p>
                    <p className="text-text-primary font-medium">
                      {formatCurrency(billing.monthly_price_usd)}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  <div>
                    <p className="text-sm text-text-muted">Billing Period</p>
                    <p className="text-text-primary font-medium">
                      {formatDate(billing.billing_period_start)} - {formatDate(billing.billing_period_end)}
                    </p>
                  </div>
                </div>
              </div>

              <CollapsibleSection
                title="Usage This Period"
                icon={<Zap className="w-4 h-4" />}
                defaultOpen={true}
                variant="default"
              >
                <div className="space-y-3">
                  <div className="flex items-center justify-between text-sm">
                    <div className="flex items-center gap-2 text-text-muted">
                      <Zap className="w-4 h-4" />
                      <span>API Requests</span>
                    </div>
                    <span
                      className={`font-medium font-mono tabular-nums ${
                        isOverLimit ? 'text-red-400' : 'text-text-secondary'
                      }`}
                    >
                      {formatNumber(billing.current_usage)} / {formatNumber(billing.included_requests)}
                    </span>
                  </div>
                  <div className="relative">
                    <div className="h-2 rounded-full overflow-hidden bg-bg-tertiary">
                      <div
                        className={`h-full rounded-full transition-all duration-500 ${
                          isOverLimit
                            ? 'bg-gradient-to-r from-red-500 to-red-400'
                            : usagePercent > 80
                              ? 'bg-gradient-to-r from-amber-500 to-orange-400'
                              : 'bg-gradient-to-r from-brand-500 to-brand-400'
                        }`}
                        style={{ width: `${usagePercent}%` }}
                      />
                    </div>
                    {usagePercent > 80 && (
                      <div
                        className={`absolute -right-0.5 -top-0.5 w-2 h-2 rounded-full ${
                          isOverLimit ? 'bg-red-500 animate-pulse' : 'bg-amber-500'
                        }`}
                      />
                    )}
                  </div>
                  {isOverLimit && (
                    <p className="text-xs text-red-400">
                      Overage: {formatNumber(billing.overage_requests)} requests (
                      {formatCurrency(billing.overage_charge_usd)})
                    </p>
                  )}
                </div>
              </CollapsibleSection>
            </div>
          ) : (
            <div className="p-6 text-center">
              <Shield className="h-12 w-12 text-text-muted mx-auto mb-3" />
              <h3 className="font-semibold text-text-primary mb-2">No Partner Account</h3>
              <p className="text-sm text-text-muted mb-4">
                You don't have a Trust API partner account yet. Register to get API access for your
                agents.
              </p>
              <Button
                onClick={() => (window.location.href = '/trust-api/register')}
                className="bg-brand-500 hover:bg-brand-500/90"
              >
                Register as Partner
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {billing && (
        <Card className="ff-card-velocity">
          <CardHeader>
            <CardTitle className="font-display">Change Plan</CardTitle>
            <CardDescription className="text-text-secondary">
              Upgrade or downgrade your Trust API tier
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
              {tierOptions.map((tier) => (
                <button
                  key={tier.id}
                  onClick={() => !tier.isCurrent && handleCheckout(tier.id)}
                  disabled={tier.isCurrent}
                  className={`p-4 rounded-lg border text-left transition-colors ${
                    tier.isCurrent
                      ? 'border-brand-500 bg-brand-500/10'
                      : 'border-border-default bg-bg-secondary hover:border-border-strong'
                  }`}
                >
                  <div className="flex items-center justify-between mb-2">
                    <span className="font-medium">{tier.name}</span>
                    {tier.isCurrent && <Badge variant="success">Current</Badge>}
                  </div>
                  <div className="text-lg font-semibold text-text-primary mb-1">
                    {tier.price === 0 ? 'Free' : formatCurrency(tier.price)}
                    {tier.price > 0 && <span className="text-sm text-text-muted">/mo</span>}
                  </div>
                  <div className="text-xs text-text-muted">
                    {formatNumber(tier.includedRequests)} requests/mo
                  </div>
                  {!tier.isCurrent && (
                    <div className="mt-2 text-xs">
                      {tier.isUpgrade && <span className="text-green-400">Upgrade</span>}
                      {tier.isDowngrade && <span className="text-amber-400">Downgrade</span>}
                    </div>
                  )}
                </button>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {billing && invoices.length > 0 && (
        <Card className="ff-card-velocity">
          <CardHeader>
            <CardTitle className="font-display">Invoices</CardTitle>
            <CardDescription className="text-text-secondary">
              Your Trust API billing history
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {invoicesLoading ? (
                <div className="space-y-2">
                  {[1, 2, 3].map((i) => (
                    <div key={i} className="h-12 rounded-md bg-bg-hover animate-pulse" />
                  ))}
                </div>
              ) : (
                invoices.slice(0, 5).map((invoice) => (
                  <div
                    key={invoice.id}
                    className="flex items-center justify-between p-3 rounded-lg bg-bg-secondary border border-border-default"
                  >
                    <div>
                      <p className="text-sm font-medium text-text-primary">
                        {formatDate(invoice.invoice_date)}
                      </p>
                      <p className="text-xs text-text-muted">
                        {invoice.stripe_invoice_id ? `Invoice ${invoice.stripe_invoice_id}` : 'Trust API'}
                      </p>
                    </div>
                    <div className="flex items-center gap-3">
                      <Badge
                        variant={invoice.status === 'paid' ? 'success' : 'secondary'}
                        className={invoice.status === 'paid' ? 'ff-badge-success' : ''}
                      >
                        {invoice.status}
                      </Badge>
                      <span className="font-medium text-text-primary">
                        {formatCurrency(invoice.amount / 100)}
                      </span>
                      {invoice.hosted_invoice_url && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => window.open(invoice.hosted_invoice_url!, '_blank')}
                        >
                          View
                        </Button>
                      )}
                    </div>
                  </div>
                ))
              )}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

function buildTierOptions(
  tiers: TrustAPITierPricing[],
  billing: TrustAPIBillingStatus | null | undefined
): TierOption[] {
  const currentTier = billing?.tier?.toLowerCase() ?? 'developer';

  return tiers.map((tier) => {
    const tierId = tier.tier.toLowerCase();
    const currentTierIndex = ['developer', 'payg', 'startup', 'business', 'enterprise'].indexOf(currentTier);
    const thisTierIndex = ['developer', 'payg', 'startup', 'business', 'enterprise'].indexOf(tierId);

    return {
      id: tierId,
      name: tier.tier.charAt(0).toUpperCase() + tier.tier.slice(1),
      price: tier.monthly_price_usd,
      includedRequests: tier.monthly_request_limit,
      isCurrent: tierId === currentTier,
      isUpgrade: thisTierIndex > currentTierIndex,
      isDowngrade: thisTierIndex < currentTierIndex && thisTierIndex > 0,
    };
  });
}