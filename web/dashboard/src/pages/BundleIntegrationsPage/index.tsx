import { useSearchParams, useNavigate, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import {
  Shield,
  CreditCard,
  Mail,
  BarChart3,
  CheckCircle,
  Loader2,
  ArrowRight,
  ChevronRight,
  Settings,
  Plug,
  Webhook,
  Rocket,
  type LucideIcon,
} from 'lucide-react';
import { apiClient } from '@/api/client';
import { appsApi } from '@/api/apps';
import { getBundleSubscription, getBundleCatalog, type BundleCatalogItem } from '@/api/billing';
import { usePageTitle } from '@/hooks';
import {
  PageGrid,
  Chamber,
  CornerBrace,
  SealedButton,
  FrameButton,
  StatusPill,
  AnnotationTag,
} from '@/components/containment';
import './styles.css';

const ICON_MAP: Record<string, LucideIcon> = {
  Shield,
  CreditCard,
  Mail,
  BarChart3,
  CheckCircle,
  Settings,
  Plug,
  Webhook,
};

function getIcon(name: string): LucideIcon {
  return ICON_MAP[name] || Shield;
}

const INTEGRATION_SECTION_MAP: Record<string, string> = {
  'Stripe Payments': 'stripe-payments',
  'OAuth Providers': 'oauth-providers',
  'Email Workflows': 'email-workflows',
  'Analytics': 'analytics',
};

interface BundleConfig {
  bundle_slug: string;
  tenant_id: string;
  auth: {
    oauth_providers: { provider: string; enabled: boolean }[];
    has_jwt_key: boolean;
    jwt_key_id: string;
    email_templates: number;
  };
  payments: {
    mode: string;
    default_currency: string;
    products: { name: string; active: boolean }[];
  };
  email: {
    transaction_templates: number;
    workflow_templates: number;
  };
  analytics: {
    dashboards: number;
    funnels: number;
    events: number;
  };
}

function useBundleConfig(appId: string) {
  return useQuery<BundleConfig>({
    queryKey: ['bundle-config', appId],
    queryFn: () => apiClient.get<BundleConfig>(`/v1/apps/${appId}/bundle/config`),
    enabled: !!appId,
  });
}

export default function BundleIntegrationsPage() {
  usePageTitle('Bundle Integrations');
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const bundleSlug = searchParams.get('bundle') || 'saas-starter';

  const { data: catalogData } = useQuery({
    queryKey: ['bundle-catalog'],
    queryFn: getBundleCatalog,
  });

  const bundle: BundleCatalogItem | undefined = catalogData?.bundles.find((b) => b.slug === bundleSlug);
  const Icon = bundle ? getIcon(bundle.icon) : Shield;

  const { data: subscription } = useQuery({
    queryKey: ['bundle-subscription', bundleSlug],
    queryFn: async () => {
      try {
        return await getBundleSubscription();
      } catch {
        return null;
      }
    },
    refetchInterval: (query) => {
      const status = query.state.data?.deploy_status;
      if (status === 'deploying' || status === 'pending') return 5000;
      return false;
    },
  });

  const { data: appsData, isLoading: appsLoading } = useQuery({
    queryKey: ['apps'],
    queryFn: async () => {
      const res = await appsApi.list();
      return res.apps;
    },
  });

  const apps = appsData ?? [];
  const bundleAppId = subscription?.default_app_id || apps[0]?.id;
  const bundleApp = apps.find((a) => a.id === bundleAppId);
  const bundleAppSlug = bundleApp?.slug || apps[0]?.slug;

  const { data: config, isLoading: configLoading } = useBundleConfig(bundleAppId || '');

  return (
    <div className="bp-page">
      <PageGrid />

      <Chamber>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="BUNDLE INTEGRATIONS" secondary={bundle?.name || bundleSlug} />
        <div className="sc-bundle-int__header">
          <div className="sc-bundle-int__header-icon">
            <Icon size={28} />
          </div>
          <div className="sc-bundle-int__header-content">
            <h1 className="sc-bundle-int__title">Configure your integrations</h1>
            <p className="sc-bundle-int__tagline">
              Connect your Stripe account, set up OAuth providers, and customize email templates.
            </p>
          </div>
        </div>

        <div className="sc-bundle-int__stepper">
          {(bundle?.provisioning_steps ?? []).map((step, i) => (
            <div
              key={step.label}
              className={`sc-bundle-int__step ${i === 1 ? 'sc-bundle-int__step--active' : i < 1 ? 'sc-bundle-int__step--done' : ''}`}
            >
              <div className="sc-bundle-int__step-num">
                {i < 1 ? <CheckCircle size={14} /> : i + 1}
              </div>
              <div className="sc-bundle-int__step-info">
                <span className="sc-bundle-int__step-label">{step.label}</span>
                <span className="sc-bundle-int__step-desc">{step.description}</span>
              </div>
              {i < (bundle?.provisioning_steps?.length ?? 0) - 1 && (
                <div className="sc-bundle-int__step-line" />
              )}
            </div>
          ))}
        </div>
      </Chamber>

      {appsLoading || configLoading ? (
        <Chamber nested className="sc-bundle-int__loading">
          <Loader2 size={20} className="sc-community-spinner" />
          <span className="sc-bundle-int__loading-text">Loading integrations...</span>
        </Chamber>
      ) : apps.length === 0 ? (
        <Chamber className="sc-bundle-int__no-app">
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <div className="sc-bundle-int__no-app-content">
            <div className="sc-bundle-int__no-app-icon">
              <Rocket size={32} />
            </div>
            <h3 className="sc-bundle-int__no-app-title">Deploy your bundle app first</h3>
            <p className="sc-bundle-int__no-app-desc">
              You need to deploy a backend app before you can configure integrations.
            </p>
            <SealedButton iconRight={<ArrowRight className="h-4 w-4" />} onClick={() => navigate(`/bundles/overview?bundle=${bundleSlug}`)}>
              Go to Bundle Overview
            </SealedButton>
          </div>
        </Chamber>
      ) : (
        <div className="sc-bundle-int__grid">
          {(bundle?.integrations ?? []).map((integration) => {
            const IntegrationIcon = getIcon(integration.icon);
            const section = INTEGRATION_SECTION_MAP[integration.title] || 'stripe-payments';
            return (
              <Chamber key={integration.title}>
                <CornerBrace position="tl" />
                <CornerBrace position="br" />
                <div className="sc-bundle-int__card">
                  <div className="sc-bundle-int__card-header">
                    <div className="sc-bundle-int__card-icon">
                      <IntegrationIcon size={20} />
                    </div>
                    <div className="sc-bundle-int__card-info">
                      <span className="sc-bundle-int__card-name">{integration.title}</span>
                      <StatusPill status="pending" label="Configure" />
                    </div>
                  </div>
                  <p className="sc-bundle-int__card-desc">{integration.description}</p>
                  <div className="sc-bundle-int__card-actions">
                    <SealedButton size="sm" iconRight={<ChevronRight size={14} />} onClick={() => navigate(`/bundles/integrations/${section}?bundle=${bundleSlug}`)}>
                      Configure
                    </SealedButton>
                  </div>
                </div>
              </Chamber>
            );
          })}
        </div>
      )}

      <Chamber nested className="sc-bundle-int__continue">
        <div className="sc-bundle-int__continue-inner">
          <div className="sc-bundle-int__continue-info">
            <CheckCircle size={18} className="sc-bundle-int__continue-icon--ok" />
            <div>
              <span className="sc-bundle-int__continue-title">Integrations ready to configure</span>
              <span className="sc-bundle-int__continue-desc">
                Set up each integration to complete your bundle setup
              </span>
            </div>
          </div>
          <div className="sc-bundle-int__continue-actions">
            <FrameButton
              size="sm"
              onClick={() => navigate(`/bundles/functions?bundle=${bundleSlug}`)}
            >
              Back to Functions
            </FrameButton>
            <SealedButton
              size="sm"
              iconRight={<ArrowRight size={14} />}
              onClick={() => navigate('/functions')}
            >
              Deploy to Production
            </SealedButton>
          </div>
        </div>
      </Chamber>
    </div>
  );
}
