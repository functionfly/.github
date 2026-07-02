import { Link, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
  Rocket,
  Store,
  Brain,
  Shield,
  CreditCard,
  Mail,
  BarChart3,
  Database,
  MessageSquare,
  Bell,
  Cpu,
  Sparkles,
  MessageCircle,
  CheckCircle,
  Clock,
  ArrowRight,
  ExternalLink,
  Zap,
  Settings,
  AlertTriangle,
  Loader2,
} from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { usePageTitle, useProvisioningStatus } from '@/hooks';
import { getFounderModeStatus, getDeferredBillingStatus, type FounderModeRegistration } from '@/api/billing';
import { appsApi } from '@/api/apps';
import {
  Chamber,
  CornerBrace,
  PageGrid,
  SealedButton,
  FrameButton,
  StatusPill,
  GaugeStrip,
  Gauge,
} from '@/components/containment';
import '@/styles/sc-my-bundles.css';

// ─── Bundle metadata ─────────────────────────────────────────────────────────

const BUNDLES: Record<string, {
  name: string;
  slug: string;
  price: string;
  icon: typeof Rocket;
  gradient: string;
  features: { icon: typeof Shield; label: string }[];
}> = {
  'saas-starter': {
    name: 'SaaS Starter',
    slug: 'saas-starter',
    price: '$29/mo',
    icon: Rocket,
    gradient: 'from-blue-500 to-cyan-500',
    features: [
      { icon: Shield, label: 'Authentication' },
      { icon: CreditCard, label: 'Payments' },
      { icon: Mail, label: 'Email Workflows' },
      { icon: BarChart3, label: 'Analytics' },
    ],
  },
  marketplace: {
    name: 'Marketplace',
    slug: 'marketplace',
    price: '$49/mo',
    icon: Store,
    gradient: 'from-emerald-500 to-teal-500',
    features: [
      { icon: Store, label: 'Listings' },
      { icon: CreditCard, label: 'Stripe Connect' },
      { icon: MessageSquare, label: 'Messaging' },
      { icon: Bell, label: 'Notifications' },
    ],
  },
  'ai-app': {
    name: 'AI App',
    slug: 'ai-app',
    price: '$39/mo',
    icon: Brain,
    gradient: 'from-violet-500 to-purple-500',
    features: [
      { icon: Cpu, label: 'Vector DB' },
      { icon: Database, label: 'Embeddings' },
      { icon: MessageCircle, label: 'Chat Workflows' },
      { icon: Sparkles, label: 'Memory System' },
    ],
  },
};

// ─── Page Component ──────────────────────────────────────────────────────────

export default function MyBundlesPage() {
  usePageTitle('My Bundles');
  const navigate = useNavigate();

  const { data: provisioningData, isLoading: provisioningLoading } = useProvisioningStatus();

  const { data: founderData } = useQuery({
    queryKey: ['founder-mode-status'],
    queryFn: getFounderModeStatus,
    retry: false,
  });

  const { data: deferredData } = useQuery({
    queryKey: ['deferred-billing-status'],
    queryFn: getDeferredBillingStatus,
    retry: false,
  });

  const { data: appsData } = useQuery({
    queryKey: ['apps'],
    queryFn: async () => { const res = await appsApi.list(); return res.apps; },
  });

  const apps = appsData ?? [];
  const founderModes = founderData?.founder_modes || [];
  const activeFounder = founderModes.find((f: FounderModeRegistration) => f.status === 'active');

  const isProvisioned = provisioningData && 'components' in provisioningData && provisioningData.status === 'active';
  const bundleSlug = isProvisioned ? provisioningData.bundle_slug : null;
  const bundle = bundleSlug ? (BUNDLES[bundleSlug] || BUNDLES['saas-starter']) : null;
  const founderMode = bundleSlug ? founderModes.find((f: FounderModeRegistration) => f.bundle_slug === bundleSlug) : null;
  const isPaid = founderMode?.status === 'converted';
  const isFounder = founderMode?.status === 'active';
  const isGrace = founderMode?.status === 'grace_period';
  const daysLeft = founderMode?.days_remaining || 0;

  return (
    <div className="sc-my-bundles">
      <PageGrid />

      {/* Header */}
      <div className="sc-my-bundles__header">
        <div>
          <h1 className="sc-my-bundles__title">
            <span className="sc-my-bundles__title-icon">
              <Zap className="h-5 w-5" />
            </span>
            My Bundles
          </h1>
          <p className="sc-my-bundles__subtitle">Your deployed backend bundles and billing status</p>
        </div>
        <SealedButton iconRight={<ArrowRight className="h-4 w-4" />} onClick={() => navigate('/bundles')}>
          Browse Bundles
        </SealedButton>
      </div>

      {/* Active Bundle Card */}
      {provisioningLoading ? (
        <div className="sc-my-bundles__loading">
          <Loader2 className="sc-my-bundles__spinner" />
          Loading your bundles...
        </div>
      ) : isProvisioned && bundle ? (
        <motion.div
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3 }}
        >
          <Chamber className="sc-my-bundles__card">
            <CornerBrace position="tl" />
            <CornerBrace position="br" />

            {/* Card Header */}
            <div className="sc-my-bundles__card-header">
              <div className={`sc-my-bundles__card-icon bg-gradient-to-r ${bundle.gradient}`}>
                <bundle.icon className="h-5 w-5 text-white" />
              </div>
              <div className="sc-my-bundles__card-info">
                <h3 className="sc-my-bundles__card-name">{bundle.name}</h3>
                <span className="sc-my-bundles__card-price">{bundle.price}</span>
              </div>
              <StatusPill
                status={isPaid ? 'live' : isGrace ? 'pending' : 'live'}
                label={isPaid ? 'Paid' : isGrace ? 'Grace Period' : 'Active'}
              />
            </div>

            {/* Status Details */}
            <div className="sc-my-bundles__card-status">
              {isFounder && (
                <div className="sc-my-bundles__status-row">
                  <Clock className="h-4 w-4" style={{ color: 'var(--status-pending)' }} />
                  <span>Free for <strong>{daysLeft} days</strong> remaining</span>
                </div>
              )}
              {isGrace && (
                <div className="sc-my-bundles__status-row sc-my-bundles__status-row--warn">
                  <AlertTriangle className="h-4 w-4" style={{ color: 'var(--status-pending)' }} />
                  <span>Grace period — add payment to continue</span>
                </div>
              )}
              {(isPaid || (!founderMode)) && (
                <div className="sc-my-bundles__status-row sc-my-bundles__status-row--ok">
                  <CheckCircle className="h-4 w-4" style={{ color: 'var(--status-ok)' }} />
                  <span>All components provisioned</span>
                </div>
              )}
            </div>

            {/* Features */}
            <div className="sc-my-bundles__card-features">
              {bundle.features.map((f) => (
                <div key={f.label} className="sc-my-bundles__feature">
                  <f.icon className="h-3.5 w-3.5" />
                  <span>{f.label}</span>
                </div>
              ))}
            </div>

            {/* Actions */}
            <div className="sc-my-bundles__card-actions">
              <FrameButton size="sm" iconLeft={<Settings />} onClick={() => navigate(apps.length > 0 ? `/apps/${apps[0].slug}/bundle` : '/apps')}>
                Manage
              </FrameButton>
              {!isPaid && founderMode && (
                <SealedButton size="sm" iconLeft={<CreditCard />} onClick={() => navigate('/billing')}>
                  Convert to Paid
                </SealedButton>
              )}
            </div>
          </Chamber>
        </motion.div>
      ) : (
        /* Empty State */
        <Chamber className="sc-my-bundles__empty">
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <div className="sc-my-bundles__empty-content">
            <Rocket className="sc-my-bundles__empty-icon" />
            <h3 className="sc-my-bundles__empty-title">No bundles yet</h3>
            <p className="sc-my-bundles__empty-desc">
              Deploy a pre-configured backend bundle to get started in minutes.
              Choose from SaaS Starter, Marketplace, or AI App.
            </p>
            <SealedButton iconRight={<ArrowRight className="h-4 w-4" />} onClick={() => navigate('/bundles')}>
              Browse Bundles
            </SealedButton>
          </div>
        </Chamber>
      )}

      {/* Deployed Apps */}
      {apps.length > 0 && (
        <Chamber nested className="sc-my-bundles__apps">
          <div className="sc-my-bundles__section-header">
            <h3 className="sc-my-bundles__section-title">Deployed Apps</h3>
            <FrameButton size="sm" iconRight={<ArrowRight className="h-3.5 w-3.5" />} onClick={() => navigate('/apps')}>
              View All
            </FrameButton>
          </div>
          <div className="sc-my-bundles__apps-list">
            {apps.slice(0, 5).map((app: { id: string; name: string }) => (
              <Link key={app.id} to={`/apps/${app.id}`} className="sc-my-bundles__app-row">
                <div className="sc-my-bundles__app-icon">
                  <ExternalLink className="h-4 w-4" />
                </div>
                <div className="sc-my-bundles__app-body">
                  <span className="sc-my-bundles__app-name">{app.name}</span>
                  <span className="sc-my-bundles__app-id">{app.id}</span>
                </div>
                <StatusPill status="live" label="Active" />
                <ArrowRight className="h-4 w-4 text-[var(--text-faint)]" />
              </Link>
            ))}
          </div>
        </Chamber>
      )}

      {/* Founder Mode Stats */}
      {activeFounder && deferredData && (
        <Chamber nested className="sc-my-bundles__stats">
          <div className="sc-my-bundles__section-header">
            <h3 className="sc-my-bundles__section-title">Founder Mode Progress</h3>
            <StatusPill status="pending" label={deferredData.status || 'Building'} />
          </div>
          <GaugeStrip>
            {deferredData.current_progress?.user_count !== undefined && (
              <Gauge isFirst data={{ value: String(deferredData.current_progress.user_count), label: 'Users' }} />
            )}
            {deferredData.current_progress?.mrr_cents !== undefined && (
              <Gauge data={{ value: `$${Number(deferredData.current_progress.mrr_cents) / 100}`, label: 'MRR' }} />
            )}
            {deferredData.estimated_days_left !== undefined && (
              <Gauge data={{ value: String(deferredData.estimated_days_left), label: 'Days Left' }} />
            )}
          </GaugeStrip>
        </Chamber>
      )}
    </div>
  );
}
