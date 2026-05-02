import { getBundles, registerFounderMode, createBundleCheckout, type Bundle } from '@/api/billing';
import { MetaTags } from '@/components/seo/MetaTags';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { useAuthStore } from '@/stores/authStore';
import { motion } from 'framer-motion';
import { ArrowLeft, Rocket, ShoppingCart, Brain, Zap, Sparkles, Clock, TrendingUp } from 'lucide-react';
import { useEffect, useState } from 'react';
import Confetti from 'react-confetti';
import toast from 'react-hot-toast';
import { Link, useNavigate } from 'react-router-dom';
import { useWindowSize } from 'react-use';
import { BundleCard } from './components/BundleCard';
import { FounderModeModal } from './components/FounderModeModal';
import { DeferredBillingSelector } from './components/DeferredBillingSelector';
import { DeployWizard } from './components/DeployWizard';

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

export function BundlePricingPage() {
  const { width, height } = useWindowSize();
  const navigate = useNavigate();
  const [bundles, setBundles] = useState<Bundle[]>([]);
  const [loading, setLoading] = useState(true);
  const [showConfetti, setShowConfetti] = useState(false);
  const [selectedBundle, setSelectedBundle] = useState<Bundle | null>(null);
  const [pricingMode, setPricingMode] = useState<PricingMode>('immediate');
  const [showFounderModal, setShowFounderModal] = useState(false);
  const [showDeployWizard, setShowDeployWizard] = useState(false);
  const [checkoutLoading, setCheckoutLoading] = useState(false);
  const user = useAuthStore((s) => s.user);

  useEffect(() => {
    loadBundles();
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

  const handleBundleSelect = (bundle: Bundle) => {
    setSelectedBundle(bundle);

    if (!user) {
      navigate(`/signup?returnTo=/pricing/bundles&plan=${bundle.slug}`);
      return;
    }

    setShowDeployWizard(true);
  };

  const handleImmediateCheckout = async (bundle: Bundle) => {
    if (!user) {
      // Redirect to signup with return path
      navigate(`/signup?returnTo=/pricing/bundles&plan=${bundle.slug}`);
      return;
    }

    setCheckoutLoading(true);
    try {
      const successUrl = `${window.location.origin}/dashboard?bundle=${bundle.slug}&success=true`;
      const cancelUrl = `${window.location.origin}/pricing/bundles`;

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
      navigate(`/signup?returnTo=/pricing/bundles&plan=${selectedBundle.slug}&mode=founder`);
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
    <div className="min-h-screen bg-bg-primary relative">
      <MetaTags
        title="Backend-in-a-Box Pricing | FunctionFly"
        description="Pre-configured backend bundles with viral pricing: SaaS Starter, Marketplace, and AI App packs. Start free until you hit 100 users or $1K MRR."
      />

      {showConfetti && <Confetti width={width} height={height} recycle={false} numberOfPieces={200} />}

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
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
          <div className="inline-flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-brand-500 to-brand-400 text-white rounded-full text-sm font-semibold mb-6">
            <Sparkles className="w-4 h-4" />
            Backend-in-a-Box Pricing
          </div>

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

        {/* Bundle Cards */}
        {loading ? (
          <div className="grid md:grid-cols-3 gap-8">
            {[1, 2, 3].map((i) => (
              <div
                key={i}
                className="h-96 bg-bg-secondary rounded-2xl animate-pulse"
              />
            ))}
          </div>
        ) : (
          <div className="grid md:grid-cols-3 gap-8">
            {bundles.map((bundle, index) => (
              <BundleCard
                key={bundle.id}
                bundle={bundle}
                icon={iconMap[bundle.icon as keyof typeof iconMap] || Rocket}
                colorClass={colorMap[bundle.color as keyof typeof colorMap] || colorMap.blue}
                onSelect={handleBundleSelect}
                delay={index * 0.1}
              />
            ))}
          </div>
        )}

        {/* Fallback if no bundles loaded */}
        {!loading && bundles.length === 0 && (
          <div className="text-center py-12">
            <p className="text-text-muted">
              No bundles available. Please check back later.
            </p>
          </div>
        )}

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
              <div>
                <div className="text-3xl font-bold text-brand-400 mb-2">
                  0
                </div>
                <p className="text-text-secondary">
                  Decision fatigue. Everything's pre-configured.
                </p>
              </div>
              <div>
                <div className="text-3xl font-bold text-brand-400 mb-2">
                  $0
                </div>
                <p className="text-text-secondary">
                  To start. Pay only when you succeed.
                </p>
              </div>
              <div>
                <div className="text-3xl font-bold text-brand-400 mb-2">
                  5min
                </div>
                <p className="text-text-secondary">
                  To production-ready backend.
                </p>
              </div>
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
            <FAQItem
              question="What's included in each bundle?"
              answer="Each bundle includes pre-configured workflows, database schemas, integrations, and templates specific to that use case. For example, the SaaS Starter includes Auth, User DB, Stripe payments, Email workflows, and Analytics."
            />
            <FAQItem
              question="How does 'Build Now, Pay Later' work?"
              answer="You can start building immediately without entering a credit card. We'll track your usage and only start billing when you hit any of these triggers: 100 users, $1,000 MRR, or 3 months have passed. You'll get a 7-day grace period to add payment info."
            />
            <FAQItem
              question="What happens when I hit the trigger?"
              answer="You'll receive email notifications at 80% of your threshold and again when you hit 100%. Then you have a 7-day grace period to add a payment method before the bundle is suspended. Your data is never deleted."
            />
            <FAQItem
              question="Can I switch bundles or cancel?"
              answer="Yes! You can upgrade, downgrade, or cancel at any time. If you cancel during founder mode, you can keep building until your trigger date. If you've converted to paid, you'll have access until the end of your billing period."
            />
            <FAQItem
              question="How is MRR calculated for the revenue trigger?"
              answer="We track revenue from your integrated payment providers (Stripe, etc.). If you use external payment processors, you can manually verify your MRR in the dashboard."
            />
          </div>
        </motion.div>
      </div>

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
    </div>
  );
}

function FAQItem({ question, answer }: { question: string; answer: string }) {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <div className="bg-bg-secondary rounded-lg shadow-sm overflow-hidden">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full px-6 py-4 flex items-center justify-between text-left hover:text-ff-flame transition-colors"
      >
        <span className="font-semibold text-text-primary">
          {question}
        </span>
        <span className="text-text-muted">
          {isOpen ? '−' : '+'}
        </span>
      </button>
      {isOpen && (
        <div className="px-6 pb-4">
          <p className="text-text-secondary">{answer}</p>
        </div>
      )}
    </div>
  );
}
