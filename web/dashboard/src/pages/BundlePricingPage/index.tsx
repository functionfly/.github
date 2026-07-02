import { getBundles, createBundleCheckout, getBundleStats, registerFounderMode, type Bundle } from '@/api/billing';
import { ArrowLeft, Rocket, ShoppingCart, Brain, Zap, Clock, TrendingUp, Shield, Lock, CheckCircle, Server, MessageCircle, Users, Package } from 'lucide-react';
import { useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import { Link, useNavigate } from 'react-router-dom';
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
} from '@/components/containment';
import { useAuthStore } from '@/stores/authStore';
import { usePageTitle } from '@/hooks';
import { BundleCard } from './components/BundleCard';
import { BundleDetailModal } from './components/BundleDetailModal';
import { FounderModeModal } from './components/FounderModeModal';
import { DeferredBillingSelector } from './components/DeferredBillingSelector';
import { DeployWizard } from './components/DeployWizard';
import { BundleComparisonTable } from './components/BundleComparisonTable';
import './styles.css';

type PricingMode = 'immediate' | 'deferred';

const iconMap = { rocket: Rocket, 'shopping-cart': ShoppingCart, brain: Brain, zap: Zap };
const colorMap = { blue: 'from-blue-500 to-cyan-500', purple: 'from-purple-500 to-pink-500', orange: 'from-orange-500 to-amber-500' };

const trustStats = [
  { value: '0', label: "Decision fatigue. Everything's pre-configured." },
  { value: '$0', label: 'To start. Pay only when you succeed.' },
  { value: '5min', label: 'To production-ready backend.' },
];

const trustBadges = [
  { icon: Shield, label: 'SOC 2 Type II', sublabel: 'Certified' },
  { icon: Lock, label: 'GDPR Compliant', sublabel: 'Data Protection' },
  { icon: Server, label: '99.9% Uptime', sublabel: 'SLA Guarantee' },
  { icon: CheckCircle, label: 'PCI DSS', sublabel: 'Level 1' },
];

const faqItems = [
  { question: "What's included in each bundle?", answer: "Each bundle includes pre-configured workflows, database schemas, integrations, and templates specific to that use case. For example, the SaaS Starter includes Auth, User DB, Stripe payments, Email workflows, and Analytics." },
  { question: "How does 'Build Now, Pay Later' work?", answer: "You can start building immediately without entering a credit card. We'll track your usage and only start billing when you hit any of these triggers: 100 users, $1,000 MRR, or 3 months have passed. You'll get a 7-day grace period to add payment info." },
  { question: "What happens when I hit the trigger?", answer: "You'll receive email notifications at 80% of your threshold and again when you hit 100%. Then you have a 7-day grace period to add a payment method before the bundle is suspended. Your data is never deleted." },
  { question: "Can I switch bundles or cancel?", answer: "Yes! You can upgrade, downgrade, or cancel at any time. If you cancel during founder mode, you can keep building until your trigger date. If you've converted to paid, you'll have access until the end of your billing period." },
  { question: "How is MRR calculated for the revenue trigger?", answer: "We track revenue from your integrated payment providers (Stripe, etc.). If you use external payment processors, you can manually verify your MRR in the dashboard." },
];

export function BundlePricingPage() {
  usePageTitle('Bundles');
  const navigate = useNavigate();
  const [bundles, setBundles] = useState<Bundle[]>([]);
  const [loading, setLoading] = useState(true);
  const [statsLoading, setStatsLoading] = useState(true);
  const [bundleStats, setBundleStats] = useState({ active_founders: 0, recent_deployments: 0 });
  const [selectedBundle, setSelectedBundle] = useState<Bundle | null>(null);
  const [detailBundle, setDetailBundle] = useState<Bundle | null>(null);
  const [showDetailModal, setShowDetailModal] = useState(false);
  const [pricingMode, setPricingMode] = useState<PricingMode>('immediate');
  const [showFounderModal, setShowFounderModal] = useState(false);
  const [showDeployWizard, setShowDeployWizard] = useState(false);
  const [checkoutLoading, setCheckoutLoading] = useState(false);
  const user = useAuthStore((s) => s.user);

  useEffect(() => { loadBundles(); loadBundleStats(); }, []);

  const loadBundles = async () => {
    try { const r = await getBundles(); setBundles(r.bundles || []); } catch (e) { console.error('Failed to load bundles:', e); toast.error('Failed to load pricing bundles'); } finally { setLoading(false); }
  };

  const loadBundleStats = async () => {
    try { setBundleStats(await getBundleStats()); } catch (e) { console.error('Failed to load bundle stats:', e); } finally { setStatsLoading(false); }
  };

  const handleBundleSelect = (bundle: Bundle) => {
    setSelectedBundle(bundle);
    if (!user) { navigate(`/signup?returnTo=/bundles&plan=${bundle.slug}&mode=founder`); return; }
    setShowFounderModal(true);
  };

  const handlePayNow = (bundle: Bundle) => {
    setSelectedBundle(bundle);
    handleImmediateCheckout(bundle);
  };

  const handleViewDetails = (bundle: Bundle) => { setDetailBundle(bundle); setShowDetailModal(true); };

  const handleImmediateCheckout = async (bundle: Bundle) => {
    if (!user) { navigate(`/signup?returnTo=/bundles&plan=${bundle.slug}`); return; }
    setCheckoutLoading(true);
    try {
      const successUrl = `${window.location.origin}/overview?bundle=${bundle.slug}&success=true`;
      const cancelUrl = `${window.location.origin}/bundles`;
      const response = await createBundleCheckout(bundle.slug, successUrl, cancelUrl);
      if (response.url) { window.location.href = response.url; } else { toast.error('Failed to create checkout session'); }
    } catch (e: any) {
      console.error('Checkout error:', e);
      const status = e?.response?.status;
      if (status === 503) {
        toast.error('Payments are not configured. Please use Founder Mode (free) or contact support.');
      } else {
        toast.error(e?.response?.data?.error || 'Failed to start checkout. Please try again.');
      }
    } finally { setCheckoutLoading(false); }
  };

  const handleFounderModeSubmit = async (modeType: string, freeDays: number, mrrThreshold: number) => {
    if (!selectedBundle) return;
    if (!user) { navigate(`/signup?returnTo=/bundles&plan=${selectedBundle.slug}&mode=founder`); return; }
    setCheckoutLoading(true);
    try {
      await registerFounderMode(selectedBundle.slug, {
        mode_type: modeType as 'time_based' | 'revenue_based' | 'hybrid',
        free_days: freeDays,
        mrr_threshold: mrrThreshold,
        success_url: `${window.location.origin}/overview?bundle=${selectedBundle.slug}&success=true`,
        cancel_url: `${window.location.origin}/bundles`,
      });
      toast.success('Founder Mode activated! Deploying your bundle...');
      setShowFounderModal(false);
      setShowDeployWizard(true);
    } catch (e: any) {
      console.error('Founder mode registration error:', e);
      const msg = e?.response?.data?.error || 'Failed to register founder mode. Please try again.';
      if (e?.response?.status === 409) {
        toast.success('Already registered! Opening deployment...');
        setShowFounderModal(false);
        setShowDeployWizard(true);
      } else {
        toast.error(msg);
      }
    } finally {
      setCheckoutLoading(false);
    }
  };

  return (
    <div className="bp-page">
      <PageGrid />

      {/* Back Link */}
      <Link to="/pricing" className="bp-back">
        <ArrowLeft className="bp-back__icon" />
        Back to all pricing
      </Link>

      {/* Hero */}
      <Chamber className="bp-hero" ribs>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="MODULE BP-01" secondary="Bundle Pricing" position="top-right" />

        <div className="bp-hero__center">
          <TrustSeal size="lg" />
          <h1 className="bp-hero__title">One Click → Full Backend</h1>
          <p className="bp-hero__subtitle">
            Pre-configured bundles that include everything you need. No thinking required. Just build.
          </p>

          <div className="bp-badges">
            <div className="bp-badge bp-badge--ok">
              <Clock className="bp-badge__icon" />
              Free for 3 months (Founder Mode)
            </div>
            <div className="bp-badge bp-badge--info">
              <TrendingUp className="bp-badge__icon" />
              Or free until $1K MRR
            </div>
            <div className="bp-badge bp-badge--accent">
              <Zap className="bp-badge__icon" />
              Or free until 100 users
            </div>
          </div>
        </div>

        {/* Billing Mode */}
        <div className="bp-billing-toggle">
          <DeferredBillingSelector mode={pricingMode} onModeChange={setPricingMode} />
        </div>

        {/* Live Stats */}
        {!statsLoading && (bundleStats.active_founders > 0 || bundleStats.recent_deployments > 0) && (
          <GaugeStrip>
            {bundleStats.active_founders > 0 && <Gauge isFirst data={{ value: bundleStats.active_founders.toLocaleString(), label: 'Founders Building' }} />}
            {bundleStats.recent_deployments > 0 && <Gauge data={{ value: bundleStats.recent_deployments.toLocaleString(), label: 'Deployments This Week' }} />}
          </GaugeStrip>
        )}
      </Chamber>

      {/* Bundle Cards */}
      {loading ? (
        <div className="bp-cards-grid">
          {[1, 2, 3].map((i) => <BundleCardSkeleton key={i} />)}
        </div>
      ) : (
        <div className="bp-cards-grid">
          {bundles.map((bundle, index) => (
            <BundleCard
              key={bundle.id}
              bundle={bundle}
              icon={iconMap[bundle.icon as keyof typeof iconMap] || Rocket}
              colorClass={colorMap[bundle.color as keyof typeof colorMap] || colorMap.blue}
              onSelect={handleBundleSelect}
              onPayNow={handlePayNow}
              onViewDetails={handleViewDetails}
              delay={index * 0.15}
              loading={checkoutLoading && selectedBundle?.id === bundle.id}
            />
          ))}
        </div>
      )}

      {/* Comparison Table */}
      {!loading && bundles.length > 0 && (
        <BundleComparisonTable bundles={bundles} iconMap={iconMap} colorMap={colorMap} />
      )}

      {/* Empty State */}
      {!loading && bundles.length === 0 && (
        <Chamber className="bp-empty">
          <CornerBrace position="tr" />
          <CornerBrace position="bl" />
          <div className="bp-empty__center">
            <Package className="bp-empty__icon" />
            <h3 className="bp-empty__title">No bundles available right now</h3>
            <p className="bp-empty__desc">We're updating our offerings. Check back soon or reach out to our team for early access to upcoming bundles.</p>
            <a href="mailto:support@functionfly.com">
              <SealedButton iconLeft={<MessageCircle className="h-4 w-4" />}>Contact Support</SealedButton>
            </a>
          </div>
        </Chamber>
      )}

      {/* Trust Badges */}
      <div className="bp-trust-row">
        {trustBadges.map((badge) => (
          <div key={badge.label} className="bp-trust-badge">
            <div className="bp-trust-badge__icon">
              <badge.icon className="bp-trust-badge__icon-svg" />
            </div>
            <div>
              <div className="bp-trust-badge__label">{badge.label}</div>
              <div className="bp-trust-badge__sublabel">{badge.sublabel}</div>
            </div>
          </div>
        ))}
      </div>

      {/* Trust Stats */}
      <Chamber className="bp-trust-stats">
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <h2 className="bp-section-title">Why founders love this</h2>
        <div className="bp-trust-stats-grid">
          {trustStats.map((stat, i) => (
            <div key={i} className="bp-trust-stat">
              <div className="bp-trust-stat__value">{stat.value}</div>
              <p className="bp-trust-stat__label">{stat.label}</p>
            </div>
          ))}
        </div>
      </Chamber>

      {/* FAQ */}
      <div className="bp-faq">
        <h2 className="bp-section-title bp-section-title--center">Frequently Asked Questions</h2>
        <div className="bp-faq-list">
          {faqItems.map((item, i) => <FAQItem key={i} question={item.question} answer={item.answer} index={i} />)}
        </div>
      </div>

      {/* Modals */}
      {detailBundle && (
        <BundleDetailModal bundle={detailBundle} icon={iconMap[detailBundle.icon as keyof typeof iconMap] || Rocket} colorClass={colorMap[detailBundle.color as keyof typeof colorMap] || colorMap.blue} isOpen={showDetailModal} onClose={() => setShowDetailModal(false)} onSelect={handleBundleSelect} onPayNow={handlePayNow} />
      )}
      <FounderModeModal isOpen={showFounderModal} onClose={() => setShowFounderModal(false)} bundle={selectedBundle} onSubmit={handleFounderModeSubmit} loading={checkoutLoading} />
      {selectedBundle && (
        <DeployWizard open={showDeployWizard} onOpenChange={setShowDeployWizard} bundle={selectedBundle} pricingMode={pricingMode} onDeployComplete={(appId, backendId) => { toast.success('Bundle deployed successfully!'); setTimeout(() => navigate(`/bundles/overview?bundle=${selectedBundle.slug}&deployed=true`), 1500); }} onDeployError={(error) => { toast.error(error); }} />
      )}
    </div>
  );
}

function FAQItem({ question, answer, index }: { question: string; answer: string; index: number }) {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <div className="bp-faq-item">
      <button onClick={() => setIsOpen(!isOpen)} aria-expanded={isOpen} className="bp-faq-trigger">
        <span>{question}</span>
        <span className={`bp-faq-icon ${isOpen ? 'bp-faq-icon--open' : ''}`}>{isOpen ? '−' : '+'}</span>
      </button>
      {isOpen && <div className="bp-faq-content"><p>{answer}</p></div>}
    </div>
  );
}

function BundleCardSkeleton() {
  return (
    <div className="bp-skeleton-card">
      <div className="bp-skeleton-card__header" />
      <div className="bp-skeleton-card__body">
        <div className="bp-skeleton bp-skeleton--title" />
        <div className="bp-skeleton bp-skeleton--line" />
        <div className="bp-skeleton bp-skeleton--line bp-skeleton--short" />
        <div className="bp-skeleton bp-skeleton--button" />
      </div>
    </div>
  );
}
