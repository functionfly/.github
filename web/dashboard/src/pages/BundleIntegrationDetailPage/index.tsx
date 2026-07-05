import { useParams, useSearchParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import {
  Shield,
  CreditCard,
  Mail,
  BarChart3,
  CheckCircle,
  XCircle,
  Loader2,
  ChevronRight,
  Settings,
  Rocket,
} from 'lucide-react';
import { appsApi } from '@/api/apps';
import { getBundleSubscription, getBundleCatalog } from '@/api/billing';
import { usePageTitle } from '@/hooks';
import {
  PageGrid,
  Chamber,
  CornerBrace,
  SealedButton,
  FrameButton,
  AnnotationTag,
} from '@/components/containment';
import {
  useBundleConfig,
  AuthCard,
  PaymentsCard,
  EmailCard,
  AnalyticsCard,
} from './components';
import './styles.css';

const INTEGRATION_META: Record<string, {
  title: string;
  description: string;
  icon: typeof Shield;
  Card: typeof AuthCard;
  sectionTitle: string;
}> = {
  'stripe-payments': {
    title: 'Stripe Payments',
    description: 'Subscriptions, invoices, webhooks, and billing management',
    icon: CreditCard,
    Card: PaymentsCard,
    sectionTitle: 'Stripe Payments Configuration',
  },
  'oauth-providers': {
    title: 'OAuth Providers',
    description: 'Google, GitHub, and social login authentication',
    icon: Shield,
    Card: AuthCard,
    sectionTitle: 'OAuth Provider Configuration',
  },
  'email-workflows': {
    title: 'Email Workflows',
    description: 'Transactional emails, welcome sequences, and dunning flows',
    icon: Mail,
    Card: EmailCard,
    sectionTitle: 'Email Workflow Configuration',
  },
  analytics: {
    title: 'Analytics',
    description: 'Event tracking, funnels, cohorts, and dashboards',
    icon: BarChart3,
    Card: AnalyticsCard,
    sectionTitle: 'Analytics Configuration',
  },
};

export default function BundleIntegrationDetailPage() {
  const { type } = useParams<{ type: string }>();
  const [searchParams] = useSearchParams();
  const bundleSlug = searchParams.get('bundle') || 'saas-starter';

  const meta = INTEGRATION_META[type] || INTEGRATION_META['stripe-payments'];
  const { title, description, icon: Icon, Card, sectionTitle } = meta;

  usePageTitle(title);

  const { data: catalogData } = useQuery({
    queryKey: ['bundle-catalog'],
    queryFn: getBundleCatalog,
  });

  const bundle = catalogData?.bundles.find((b) => b.slug === bundleSlug);

  const { data: subscription } = useQuery({
    queryKey: ['bundle-subscription', bundleSlug],
    queryFn: async () => {
      try {
        return await getBundleSubscription();
      } catch {
        return null;
      }
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
  const bundleAppSlug = apps.find((a) => a.id === bundleAppId)?.slug || apps[0]?.slug;

  const { data: config, isLoading: configLoading, error } = useBundleConfig(bundleAppId || '');

  if (appsLoading || configLoading) {
    return (
      <div className="bp-page">
        <PageGrid />
        <Chamber nested className="sc-bundle-int-detail__loading">
          <Loader2 size={20} className="sc-community-spinner" />
          <span className="sc-bundle-int-detail__loading-text">Loading {title}...</span>
        </Chamber>
      </div>
    );
  }

  if (error || !config) {
    return (
      <div className="bp-page">
        <PageGrid />
        <Chamber nested className="sc-bundle-int-detail__loading">
          <XCircle size={24} style={{ color: 'var(--status-revoked)' }} />
          <span className="sc-bundle-int-detail__loading-text">Failed to load configuration</span>
        </Chamber>
      </div>
    );
  }

  if (apps.length === 0) {
    return (
      <div className="bp-page">
        <PageGrid />
        <Chamber className="sc-bundle-int-detail__no-app">
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <div className="sc-bundle-int-detail__no-app-content">
            <div className="sc-bundle-int-detail__no-app-icon">
              <Rocket size={32} />
            </div>
            <h3 className="sc-bundle-int-detail__no-app-title">Deploy your bundle app first</h3>
            <p className="sc-bundle-int-detail__no-app-desc">
              You need to deploy a backend app before you can configure integrations.
            </p>
            <SealedButton iconRight={<ChevronRight className="h-4 w-4" />} onClick={() => window.location.href = `/bundles/overview?bundle=${bundleSlug}`}>
              Go to Bundle Overview
            </SealedButton>
          </div>
        </Chamber>
      </div>
    );
  }

  return (
    <div className="bp-page">
      <PageGrid />

      {/* Header */}
      <Chamber>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary={sectionTitle} secondary={bundle?.name || bundleSlug} />
        <div className="sc-bundle-int-detail__header">
          <div className="sc-bundle-int-detail__header-icon">
            <Icon size={28} />
          </div>
          <div className="sc-bundle-int-detail__header-content">
            <h1 className="sc-bundle-int-detail__title">{title}</h1>
            <p className="sc-bundle-int-detail__tagline">{description}</p>
          </div>
        </div>

        <div className="sc-bundle-int-detail__breadcrumb">
          <Link to={`/bundles/integrations?bundle=${bundleSlug}`} className="sc-bundle-int-detail__breadcrumb-link">
            Integrations
          </Link>
          <ChevronRight size={14} style={{ color: 'var(--text-faint)' }} />
          <span className="sc-bundle-int-detail__breadcrumb-current">{title}</span>
        </div>
      </Chamber>

      {/* Config Card */}
      <div className="sc-bundle-int-detail__card-wrapper">
        {type === 'stripe-payments' && <PaymentsCard config={config.payments} />}
        {type === 'oauth-providers' && <AuthCard config={config.auth} appId={bundleAppId} />}
        {type === 'email-workflows' && <EmailCard config={config.email} appId={bundleAppId} />}
        {type === 'analytics' && <AnalyticsCard config={config.analytics} />}
      </div>

      {/* Footer Nav */}
      <Chamber nested className="sc-bundle-int-detail__footer">
        <div className="sc-bundle-int-detail__footer-inner">
          <FrameButton size="sm" iconRight={<ChevronRight className="w-3.5 h-3.5" />} onClick={() => window.location.href = `/bundles/integrations?bundle=${bundleSlug}`}>
            Back to Integrations
          </FrameButton>
          <SealedButton size="sm" iconRight={<ChevronRight className="w-3.5 h-3.5" />} onClick={() => window.location.href = `/bundles/functions?bundle=${bundleSlug}`}>
            Next: Functions
          </SealedButton>
        </div>
      </Chamber>
    </div>
  );
}
