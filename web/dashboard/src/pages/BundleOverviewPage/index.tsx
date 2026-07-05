import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import {
  CheckCircle,
  Rocket,
  Shield,
  CreditCard,
  Mail,
  BarChart3,
  Store,
  MessageSquare,
  Bell,
  Brain,
  Cpu,
  Database,
  Sparkles,
  Settings,
  ExternalLink,
  ArrowRight,
  Zap,
  Clock,
  TrendingUp,
  Server,
  Code,
  Loader2,
  AlertCircle,
  Cloud,
  Webhook,
  type LucideIcon,
} from 'lucide-react';
import { appsApi } from '@/api/apps';
import { getBundleSubscription, getBundleCatalog, type BundleCatalogItem } from '@/api/billing';
import { usePageTitle } from '@/hooks';
import {
  PageGrid,
  Chamber,
  CornerBrace,
  TrustSeal,
  SealedButton,
  FrameButton,
  StatusPill,
  GaugeStrip,
  Gauge,
  AnnotationTag,
  Card,
} from '@/components/containment';
import './styles.css';

const ICON_MAP: Record<string, LucideIcon> = {
  Rocket,
  Shield,
  CreditCard,
  Mail,
  BarChart3,
  Store,
  MessageSquare,
  Bell,
  Brain,
  Cpu,
  Database,
  Sparkles,
  Settings,
  Webhook: Webhook,
  CheckCircle,
  ExternalLink,
  ArrowRight,
  Zap,
  Clock,
  TrendingUp,
  Server,
  Code,
  Loader2,
  AlertCircle,
  Cloud,
};

function getIcon(name: string): LucideIcon {
  return ICON_MAP[name] || Shield;
}

export default function BundleOverviewPage() {
  usePageTitle('Bundle Overview');
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const bundleSlug = searchParams.get('bundle') || 'saas-starter';
  const justDeployed = searchParams.get('deployed') === 'true';
  const [showConfetti, setShowConfetti] = useState(justDeployed);

  const { data: catalogData } = useQuery({
    queryKey: ['bundle-catalog'],
    queryFn: getBundleCatalog,
  });

  const bundle: BundleCatalogItem | undefined = catalogData?.bundles.find((b) => b.slug === bundleSlug);
  const Icon = bundle ? getIcon(bundle.icon) : Rocket;

  const { data: appsData, isLoading: appsLoading } = useQuery({
    queryKey: ['apps'],
    queryFn: async () => {
      const res = await appsApi.list();
      return res.apps;
    },
  });

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

  const apps = appsData ?? [];
  const bundleAppId = subscription?.default_app_id;
  const bundleApps = bundleAppId ? apps.filter((a) => a.id === bundleAppId) : apps;
  const latestApp = apps[0];

  useEffect(() => {
    if (justDeployed) {
      const timer = setTimeout(() => setShowConfetti(false), 5000);
      return () => clearTimeout(timer);
    }
  }, [justDeployed]);

  const deployStatus = subscription?.deploy_status;

  return (
    <div className="bp-page">
      <PageGrid />

      {/* Success Banner */}
      {justDeployed && (
        <Chamber className="sc-bundle-overview__success">
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <div className="sc-bundle-overview__success-inner">
            <div className="sc-bundle-overview__success-icon">
              <CheckCircle size={24} />
            </div>
            <div className="sc-bundle-overview__success-content">
              <h2 className="sc-bundle-overview__success-title">
                {bundle.name} deployed successfully!
              </h2>
              <p className="sc-bundle-overview__success-desc">
                Your backend is live and ready. All functions, auth, and integrations have been configured.
              </p>
              <div className="sc-bundle-overview__success-actions">
                {latestApp && (
                  <SealedButton
                    size="sm"
                    iconLeft={<ExternalLink size={14} />}
                    onClick={() => navigate(`/apps/${latestApp.id}`)}
                  >
                    Open App
                  </SealedButton>
                )}
                <FrameButton
                  size="sm"
                  iconLeft={<Code size={14} />}
                  onClick={() => navigate('/functions')}
                >
                  View Functions
                </FrameButton>
              </div>
            </div>
          </div>
        </Chamber>
      )}

      {/* Deploy Status Banner */}
      {deployStatus && deployStatus !== 'deployed' && (
        <Chamber nested className={`sc-bundle-overview__status sc-bundle-overview__status--${deployStatus}`}>
          <div className="sc-bundle-overview__status-inner">
            {deployStatus === 'deploying' && <Loader2 size={18} className="sc-community-spinner" />}
            {deployStatus === 'awaiting_provider' && <Cloud size={18} />}
            {deployStatus === 'failed' && <AlertCircle size={18} />}
            {deployStatus === 'pending' && <Clock size={18} />}
            <div className="sc-bundle-overview__status-content">
              <p className="sc-bundle-overview__status-text">
                {deployStatus === 'deploying' && 'Deploying to your provider...'}
                {deployStatus === 'awaiting_provider' && 'Connect a provider to deploy'}
                {deployStatus === 'failed' && `Deployment failed: ${subscription?.deploy_error || 'Unknown error'}`}
                {deployStatus === 'pending' && 'Deployment pending...'}
              </p>
              {deployStatus === 'awaiting_provider' && (
                <p className="sc-bundle-overview__status-hint">
                  Your bundle is ready. Connect Cloudflare Workers, Vercel, or another provider to go live.
                </p>
              )}
            </div>
            {deployStatus === 'awaiting_provider' && (
              <SealedButton size="sm" iconLeft={<Cloud size={14} />} onClick={() => navigate('/providers')}>
                Connect Provider
              </SealedButton>
            )}
            {deployStatus === 'failed' && (
              <FrameButton size="sm">Retry</FrameButton>
            )}
          </div>
        </Chamber>
      )}

      {/* Header */}
      <Chamber>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="BUNDLE · ACTIVE" secondary={bundle.price_usd} />
        <div className="sc-bundle-overview__header">
          <div className="sc-bundle-overview__header-icon">
            <Icon size={28} />
          </div>
          <div className="sc-bundle-overview__header-content">
            <h1 className="sc-bundle-overview__title">{bundle.name}</h1>
            <p className="sc-bundle-overview__tagline">{bundle.tagline}</p>
          </div>
          <div className="sc-bundle-overview__header-status">
            <StatusPill status="live" label="Active" />
            <span className="sc-bundle-overview__price">{bundle.price_usd}</span>
          </div>
        </div>
        <GaugeStrip>
          <Gauge data={{ value: bundle.features.length, label: 'Features' }} isFirst />
          <Gauge data={{ value: bundleApps.length, label: 'Apps' }} />
          <Gauge data={{ value: 'Live', label: 'Status' }} />
        </GaugeStrip>
      </Chamber>

      {/* Quick Actions */}
      <div className="sc-bundle-overview__actions-grid">
        {[
          { label: 'View Functions', href: `/bundles/functions?bundle=${bundleSlug}`, icon: Code },
          { label: 'App Settings', href: `/bundles/integrations?bundle=${bundleSlug}`, icon: Settings },
          { label: 'Billing', href: '/billing', icon: CreditCard },
        ].map((action) => (
          <Link key={action.label} to={action.href} className="sc-bundle-overview__action-card">
            <Card>
              <div className="sc-bundle-overview__action-inner">
                <div className="sc-bundle-overview__action-icon">
                  <action.icon size={18} />
                </div>
                <span className="sc-bundle-overview__action-label">{action.label}</span>
                <ArrowRight size={14} className="sc-bundle-overview__action-arrow" />
              </div>
            </Card>
          </Link>
        ))}
      </div>

      {/* Features Grid */}
      <Chamber>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <h2 className="sc-bundle-overview__section-title">What's Included</h2>
        <div className="sc-bundle-overview__features-grid">
          {bundle.features.map((f) => (
            <div key={f.title} className="sc-bundle-overview__feature">
              <div className="sc-bundle-overview__feature-icon">
                {(() => {
                  const FeatureIcon = getIcon(f.icon);
                  return <FeatureIcon size={16} />;
                })()}
              </div>
              <div>
                <div className="sc-bundle-overview__feature-title">{f.title}</div>
                <div className="sc-bundle-overview__feature-desc">{f.description}</div>
              </div>
            </div>
          ))}
        </div>
      </Chamber>

      {/* Apps */}
      {appsLoading ? (
        <Chamber nested className="sc-bundle-overview__loading">
          <Loader2 size={20} className="sc-community-spinner" />
          <span className="sc-bundle-overview__loading-text">Loading your apps...</span>
        </Chamber>
      ) : bundleApps.length > 0 ? (
        <Chamber>
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <h2 className="sc-bundle-overview__section-title">Your Apps</h2>
          <div className="sc-bundle-overview__apps-list">
            {bundleApps.map((app) => (
              <Link key={app.id} to={`/apps/${app.id}`} className="sc-bundle-overview__app-row">
                <div className="sc-bundle-overview__app-icon">
                  <Server size={18} />
                </div>
                <div className="sc-bundle-overview__app-info">
                  <span className="sc-bundle-overview__app-name">{app.name}</span>
                  <span className="sc-bundle-overview__app-id">{app.id}</span>
                </div>
                <StatusPill status="live" label="Active" />
                <ArrowRight size={14} className="sc-bundle-overview__action-arrow" />
              </Link>
            ))}
          </div>
        </Chamber>
      ) : null}

      {/* Next Steps */}
      <Chamber>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <h2 className="sc-bundle-overview__section-title">Next Steps</h2>
        <div className="sc-bundle-overview__steps">
          <NextStep
            num={1}
            title="Explore your functions"
            desc="Your bundle came with pre-configured function templates. Customize them for your use case."
            href={`/bundles/functions?bundle=${bundleSlug}`}
          />
          <NextStep
            num={2}
            title="Configure integrations"
            desc="Set up your Stripe keys, OAuth providers, and email templates."
            href={`/bundles/integrations?bundle=${bundleSlug}`}
          />
          <NextStep
            num={3}
            title="Deploy to production"
            desc="When you're ready, deploy your functions to the edge."
            href="/functions"
          />
        </div>
      </Chamber>
    </div>
  );
}

function NextStep({ num, title, desc, href }: { num: number; title: string; desc: string; href: string }) {
  return (
    <Link to={href} className="sc-bundle-overview__step">
      <div className="sc-bundle-overview__step-num">{num}</div>
      <div className="sc-bundle-overview__step-content">
        <div className="sc-bundle-overview__step-title">{title}</div>
        <div className="sc-bundle-overview__step-desc">{desc}</div>
      </div>
      <ArrowRight size={14} className="sc-bundle-overview__action-arrow" />
    </Link>
  );
}
