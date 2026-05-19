import { getBundles, registerFounderMode, createBundleCheckout, getBundleStats, type Bundle } from '@/api/billing';
import { PageLayout } from '@/components/layout/PageLayout';
import { PageHeader } from '@/components/layout/PageHeader';
import { Button } from '@/components/ui/button';
import { useAuthStore } from '@/stores/authStore';
import { motion } from 'framer-motion';
import { ArrowLeft, Rocket, ShoppingCart, Brain, Zap, Clock, TrendingUp, Shield, Lock, CheckCircle, Server, MessageCircle, Users, Package } from 'lucide-react';
import { useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import { Link, useNavigate } from 'react-router-dom';
import { BundleCard } from './components/BundleCard';
import { BundleDetailModal } from './components/BundleDetailModal';
import { FounderModeModal } from './components/FounderModeModal';
import { DeferredBillingSelector } from './components/DeferredBillingSelector';
import { DeployWizard } from './components/DeployWizard';
import { BundleComparisonTable } from './components/BundleComparisonTable';

type PricingMode = 'immediate' | 'deferred';

const iconMap = {
  rocket: Rocket,
  'shopping-cart': ShoppingCart,
  brain: Brain,
  zap: Zap,
};

const colorMap = {
  blue: 'from-blue-500 to-cyan-500',
  purple: 'from-purple-500 to-pink-500',
  orange: 'from-orange-500 to-amber-500',
};

const trustStats = [
  { value: '0', label: 'Decision fatigue. Everything\'s pre-configured.' },
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
  {
    question: "What's included in each bundle?",
    answer: "Each bundle includes pre-configured workflows, database schemas, integrations, and templates specific to that use case. For example, the SaaS Starter includes Auth, User DB, Stripe payments, Email workflows, and Analytics.",
  },
  {
    question: "How does 'Build Now, Pay Later' work?",
    answer: "You can start building immediately without entering a credit card. We'll track your usage and only start billing when you hit any of these triggers: 100 users, $1,000 MRR, or 3 months have passed. You'll get a 7-day grace period to add payment info.",
  },
  {
    question: "What happens when I hit the trigger?",
    answer: "You'll receive email notifications at 80% of your threshold and again when you hit 100%. Then you have a 7-day grace period to add a payment method before the bundle is suspended. Your data is never deleted.",
  },
  {
    question: "Can I switch bundles or cancel?",
    answer: "Yes! You can upgrade, downgrade, or cancel at any time. If you cancel during founder mode, you can keep building until your trigger date. If you've converted to paid, you'll have access until the end of your billing period.",
  },
  {
    question: "How is MRR calculated for the revenue trigger?",
    answer: "We track revenue from your integrated payment providers (Stripe, etc.). If you use external payment processors, you can manually verify your MRR in the dashboard.",
  },
];

export function BundlePricingPage() {
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

  useEffect(() => {
    loadBundles();
    loadBundleStats();
  }, []);

  const loadBundles = async () => {
    try {
      const response = await getBundles();
      setBundles(response.bundles || []);
    } catch (error) {
      console.error('Failed to load bundles:', error);
      toast.error('Failed to load pricing bundles');
    } finally {
      setLoading(false);
    }
  };

  const loadBundleStats = async () => {
    try {
      const stats = await getBundleStats();
      setBundleStats(stats);
    } catch (error) {
      console.error('Failed to load bundle stats:', error);
    } finally {
      setStatsLoading(false);
    }
  };

  const handleBundleSelect = (bundle: Bundle) => {
    setSelectedBundle(bundle);

    if (!user) {
      navigate(`/signup?returnTo=/bundles&plan=${bundle.slug}`);
      return;
    }

    setShowDeployWizard(true);
  };

  const handleViewDetails = (bundle: Bundle) => {
    setDetailBundle(bundle);
    setShowDetailModal(true);
  };

  const handleImmediateCheckout = async (bundle: Bundle) => {
    if (!user) {
      navigate(`/signup?returnTo=/bundles&plan=${bundle.slug}`);
      return;
    }

    setCheckoutLoading(true);
    try {
      const successUrl = `${window.location.origin}/dashboard?bundle=${bundle.slug}&success=true`;
      const cancelUrl = `${window.location.origin}/bundles`;

      const response = await createBundleCheckout(bundle.slug, successUrl, cancelUrl);

      if (response.url) {
        window.location.href = response.url;
      } else {
        toast.error('Failed to create checkout session');
      }
    } catch (error) {
      console.error('Checkout error:', error);
      toast.error('Failed to start checkout. Please try again.');
    } finally {
      setCheckoutLoading(false);
    }
  };

  const handleFounderModeSubmit = async (modeType: string, freeDays: number, mrrThreshold: number) => {
    if (!selectedBundle) return;

    if (!user) {
      navigate(`/signup?returnTo=/bundles&plan=${selectedBundle.slug}&mode=founder`);
      return;
    }

    setCheckoutLoading(true);
    try {
      toast.success('🎉 Activating Founder Mode!');
      setShowFounderModal(false);
      setShowDeployWizard(true);
    } catch (error) {
      console.error('Founder mode registration error:', error);
      toast.error('Failed to register founder mode. Please try again.');
    } finally {
      setCheckoutLoading(false);
    }
  };

  return (
    <PageLayout>
      <PageHeader
        title="Backend-in-a-Box Pricing"
        subtitle="Pre-configured bundles with viral pricing: SaaS Starter, Marketplace, and AI App packs. Start free until you hit 100 users or $1K MRR."
      />

      <div className="mt-8">
        {/* Back Link */}
        <Link
          to="/pricing"
          className="inline-flex items-center gap-2 text-text-secondary hover:text-text-primary mb-8 transition-colors"
        >
          <ArrowLeft className="w-4 h-4" />
          Back to all pricing
        </Link>

        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-center mb-12"
        >
          <h1 className="text-4xl sm:text-5xl font-bold text-text-primary mb-4">
            One Click → Full Backend
          </h1>

          <p className="text-xl text-text-secondary max-w-2xl mx-auto mb-6">
            Pre-configured bundles that include everything you need.
            No thinking required. Just build.
          </p>

          {/* Viral Pricing Badges */}
          <div className="flex flex-wrap justify-center gap-4 text-sm">
            <div className="flex items-center gap-2 px-4 py-2 bg-success/10 dark:bg-success/20 text-success rounded-full">
              <Clock className="w-4 h-4" />
              Free for 3 months (Founder Mode)
            </div>
            <div className="flex items-center gap-2 px-4 py-2 bg-info/10 dark:bg-info/20 text-info rounded-full">
              <TrendingUp className="w-4 h-4" />
              Or free until $1K MRR
            </div>
            <div className="flex items-center gap-2 px-4 py-2 bg-brand-500/10 dark:bg-brand-500/20 text-brand-400 rounded-full">
              <Zap className="w-4 h-4" />
              Or free until 100 users
            </div>
          </div>
        </motion.div>

        {/* Pricing Mode Selector */}
        <DeferredBillingSelector
          mode={pricingMode}
          onModeChange={setPricingMode}
        />

        {/* Live Usage Counter */}
        {!statsLoading && (bundleStats.active_founders > 0 || bundleStats.recent_deployments > 0) && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.35 }}
            className="flex justify-center gap-8 mb-8"
          >
            {bundleStats.active_founders > 0 && (
              <div className="flex items-center gap-2 text-sm text-text-secondary">
                <Users className="w-4 h-4 text-success" />
                <span className="font-semibold text-text-primary">{bundleStats.active_founders.toLocaleString()}</span>
                founders currently building
              </div>
            )}
            {bundleStats.recent_deployments > 0 && (
              <div className="flex items-center gap-2 text-sm text-text-secondary">
                <Package className="w-4 h-4 text-brand-400" />
                <span className="font-semibold text-text-primary">{bundleStats.recent_deployments.toLocaleString()}</span>
                deployments this week
              </div>
            )}
          </motion.div>
        )}

        {/* Bundle Cards */}
        {loading ? (
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            {[1, 2, 3].map((i) => (
              <BundleCardSkeleton key={i} />
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            {bundles.map((bundle, index) => (
              <BundleCard
                key={bundle.id}
                bundle={bundle}
                icon={iconMap[bundle.icon as keyof typeof iconMap] || Rocket}
                colorClass={colorMap[bundle.color as keyof typeof colorMap] || colorMap.blue}
                onSelect={handleBundleSelect}
                onViewDetails={handleViewDetails}
                delay={index * 0.15}
              />
            ))}
          </div>
        )}

        {/* Bundle Comparison Table */}
        {!loading && bundles.length > 0 && (
          <BundleComparisonTable bundles={bundles} iconMap={iconMap} colorMap={colorMap} />
        )}

        {/* Fallback if no bundles loaded */}
        {!loading && bundles.length === 0 && (
          <div className="text-center py-16 px-4">
            <div className="w-24 h-24 bg-bg-tertiary rounded-2xl flex items-center justify-center mx-auto mb-6">
              <Package className="w-12 h-12 text-text-muted" />
            </div>
            <h3 className="text-xl font-semibold text-text-primary mb-2">
              No bundles available right now
            </h3>
            <p className="text-text-secondary mb-6 max-w-md mx-auto">
              We&apos;re updating our offerings. Check back soon or reach out to our team for early access to upcoming bundles.
            </p>
            <a
              href="mailto:support@functionfly.com"
              className="inline-flex items-center gap-2 px-5 py-2.5 bg-white text-zinc-900 font-bold rounded-lg hover:bg-zinc-100 transition-colors shadow-lg"
            >
              <MessageCircle className="w-4 h-4" />
              Contact Support
            </a>
          </div>
        )}

        {/* Trust Badges / Security Signals */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.4 }}
          className="mt-12 py-6 border-y border-border/50"
        >
          <div className="flex flex-wrap justify-center items-center gap-6 lg:gap-12">
            {trustBadges.map((badge, index) => (
              <motion.div
                key={badge.label}
                initial={{ opacity: 0, scale: 0.9 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ delay: 0.5 + index * 0.1 }}
                className="flex items-center gap-3 px-4 py-2 rounded-lg bg-bg-secondary/50"
              >
                <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-brand-500/20 to-brand-400/10 flex items-center justify-center">
                  <badge.icon className="w-5 h-5 text-brand-400" />
                </div>
                <div className="text-left">
                  <div className="text-sm font-semibold text-text-primary">{badge.label}</div>
                  <div className="text-xs text-text-muted">{badge.sublabel}</div>
                </div>
              </motion.div>
            ))}
          </div>
        </motion.div>

        {/* Trust Section */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.5 }}
          className="mt-16 text-center"
        >
          <div className="bg-bg-secondary rounded-2xl p-8 shadow-lg">
            <h3 className="text-2xl font-bold text-text-primary mb-4">
              Why founders love this
            </h3>
            <div className="grid sm:grid-cols-3 gap-6">
              {trustStats.map((stat, index) => (
                <div key={index}>
                  <div className="text-3xl font-bold text-brand-400 mb-2">
                    {stat.value}
                  </div>
                  <p className="text-text-secondary">
                    {stat.label}
                  </p>
                </div>
              ))}
            </div>
          </div>
        </motion.div>

        {/* FAQ Section */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.6 }}
          className="mt-12 max-w-3xl mx-auto"
        >
          <h3 className="text-2xl font-bold text-text-primary text-center mb-6">
            Frequently Asked Questions
          </h3>
          <div className="space-y-4">
            {faqItems.map((item, index) => (
              <FAQItem
                key={index}
                question={item.question}
                answer={item.answer}
                index={index}
              />
            ))}
          </div>
        </motion.div>
      </div>

      {/* Bundle Detail Modal */}
      {detailBundle && (
        <BundleDetailModal
          bundle={detailBundle}
          icon={iconMap[detailBundle.icon as keyof typeof iconMap] || Rocket}
          colorClass={colorMap[detailBundle.color as keyof typeof colorMap] || colorMap.blue}
          isOpen={showDetailModal}
          onClose={() => setShowDetailModal(false)}
          onSelect={handleBundleSelect}
        />
      )}

      {/* Founder Mode Modal */}
      <FounderModeModal
        isOpen={showFounderModal}
        onClose={() => setShowFounderModal(false)}
        bundle={selectedBundle}
        onSubmit={handleFounderModeSubmit}
        loading={checkoutLoading}
      />

      {/* Deploy Wizard */}
      {selectedBundle && (
        <DeployWizard
          open={showDeployWizard}
          onOpenChange={setShowDeployWizard}
          bundle={selectedBundle}
          pricingMode={pricingMode}
          onDeployComplete={(appId, backendId) => {
            toast.success('Bundle deployed successfully!');
            setTimeout(() => navigate(`/dashboard/apps/${appId}`), 1500);
          }}
          onDeployError={(error) => {
            toast.error(error);
          }}
        />
      )}
    </PageLayout>
  );
}

function FAQItem({ question, answer, index }: { question: string; answer: string; index: number }) {
  const [isOpen, setIsOpen] = useState(false);
  const id = `faq-${index}`;

  return (
    <div className="bg-bg-secondary rounded-lg shadow-sm overflow-hidden">
      <button
        onClick={() => setIsOpen(!isOpen)}
        aria-expanded={isOpen}
        aria-controls={id}
        className="w-full px-6 py-4 flex items-center justify-between text-left hover:text-ff-flame transition-colors"
      >
        <span className="font-semibold text-text-primary">
          {question}
        </span>
        <span className="text-text-muted text-xl leading-none" aria-hidden="true">
          {isOpen ? '−' : '+'}
        </span>
      </button>
      {isOpen && (
        <div id={id} className="px-6 pb-4">
          <p className="text-text-secondary">{answer}</p>
        </div>
      )}
    </div>
  );
}

function BundleCardSkeleton() {
  return (
    <div className="bg-bg-secondary rounded-2xl shadow-lg border border-border overflow-hidden">
      {/* Header skeleton */}
      <div className="bg-gradient-to-r from-gray-400 to-gray-500 p-6 h-36" />
      {/* Content skeleton */}
      <div className="p-6 space-y-4">
        <div className="h-4 bg-bg-tertiary rounded w-3/4" />
        <div className="space-y-2">
          <div className="h-3 bg-bg-tertiary rounded w-1/2" />
          <div className="h-3 bg-bg-tertiary rounded w-2/3" />
          <div className="h-3 bg-bg-tertiary rounded w-1/2" />
        </div>
        <div className="pt-4 space-y-2">
          <div className="h-10 bg-bg-tertiary rounded" />
          <div className="h-10 bg-bg-tertiary rounded" />
        </div>
      </div>
    </div>
  );
}
