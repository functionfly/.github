import { createCheckoutSession, getCheckoutErrorMessage } from '@/api/billing';
import { MetaTags } from '@/components/seo/MetaTags';
import { PricingPageStructuredData } from '@/components/seo/StructuredData';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { PLANS, STATE_FABRIC_PLANS } from '@/lib/constants';
import { Footer } from '@/pages/LandingPage/components';
import { useAuthStore } from '@/stores/authStore';
import { motion } from 'framer-motion';
import { ArrowLeft, Bot, CreditCard, Database, Mail, Shield, Zap } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import Confetti from 'react-confetti';
import toast, { Toaster } from 'react-hot-toast';
import { useTranslation } from 'react-i18next';
import { Link, useSearchParams } from 'react-router-dom';
import { Tooltip } from 'react-tooltip';
import { useWindowSize } from 'react-use';
import { AgentPricingSection } from './components/AgentPricingSection';
import { CTASection } from './components/CTASection';
import { FAQSection } from './components/FAQSection';
import { FunctionPlanCard } from './components/FunctionPlanCard';
import { FunctionsComparisonTable } from './components/FunctionsComparisonTable';
import { StateFabricPricingSection } from './components/StateFabricPricingSection';
import { BundleCTACard } from './components/BundleCTACard';

type PricingTab = 'functions' | 'state-fabric' | 'agents';

function cn(...classes: (string | boolean | undefined)[]) {
  return classes.filter(Boolean).join(' ');
}

export function PricingPage() {
  const { t } = useTranslation();
  const { width, height } = useWindowSize();
  const TABS: { id: PricingTab; label: string; icon: typeof Zap }[] = [
    { id: 'functions', label: t("pricing:tabs.functions"), icon: Zap },
    { id: 'state-fabric', label: t("pricing:tabs.stateFabric"), icon: Database },
    { id: 'agents', label: t("pricing:tabs.agents"), icon: Bot },
  ];
  const [showConfetti, setShowConfetti] = useState(false);
  const [activeTab, setActiveTab] = useState<PricingTab>('functions');
  const [searchParams] = useSearchParams();
  const user = useAuthStore((s) => s.user);
  const [checkoutError, setCheckoutError] = useState<{
    message: string;
    planId: string;
    priceId?: string;
  } | null>(null);
  const [isRetrying, setIsRetrying] = useState(false);
  const [checkoutInitiating, setCheckoutInitiating] = useState(false);

  // Auto-trigger checkout from marketing site URL params
  // e.g. /pricing?tab=functions&plan=starter or /pricing?tab=state-fabric&plan=pro
  const autoCheckoutTriggered = useRef(false);
  useEffect(() => {
    if (autoCheckoutTriggered.current) return;
    const tabParam = searchParams.get('tab') as PricingTab | null;
    const planParam = searchParams.get('plan');
    if (!tabParam || !planParam) return;
    autoCheckoutTriggered.current = true;

    // Set the correct tab
    if (tabParam === 'functions' || tabParam === 'state-fabric' || tabParam === 'agents') {
      setActiveTab(tabParam);
    }

    // Small delay to let the tab render before triggering checkout
    const timer = setTimeout(() => {
      const planMap: Record<string, { planId: string; priceId?: string }> = {
        // Functions plans
        starter: { planId: 'starter', priceId: PLANS.STARTER.priceId },
        professional: { planId: 'professional', priceId: PLANS.PROFESSIONAL.priceId },
        // State Fabric plans
        sf_starter: { planId: 'sf_starter', priceId: STATE_FABRIC_PLANS.STARTER.priceId },
        sf_pro: { planId: 'sf_pro', priceId: STATE_FABRIC_PLANS.PRO.priceId },
        sf_business: { planId: 'sf_business', priceId: STATE_FABRIC_PLANS.BUSINESS.priceId },
      };
      const entry = planMap[planParam];
      if (entry) {
        handlePlanSelect(entry.planId, entry.priceId);
      }
    }, 350);

    return () => clearTimeout(timer);
  }, [searchParams]);

  const handlePlanSelect = async (planId: string, priceId?: string) => {
    // For enterprise, navigate to contact page
    if (planId === 'enterprise') {
      window.location.href = '/contact';
      return;
    }

    // For free plan, just show success and redirect to signup
    if (planId === 'free') {
      setShowConfetti(true);
      setTimeout(() => setShowConfetti(false), 3000);
      toast.success(t("pricing:toasts.freePlanSuccess"), {
        duration: 4000,
        style: {
          background: '#1a1a1a',
          color: '#fff',
          border: '1px solid #6366f1',
        },
        icon: '🚀',
      });
      return;
    }

    // For paid plans, create checkout session
    if (!priceId) {
      toast.error(t("pricing:toasts.invalidPlan"));
      return;
    }

    setCheckoutInitiating(true);
    setShowConfetti(true);
    setTimeout(() => setShowConfetti(false), 3000);

    try {
      const base = window.location.origin;
      const successUrl = user?.username
        ? `${base}/u/${user.username}/settings/billing?subscription=success`
        : `${base}/settings?tab=billing&subscription=success`;
      const cancelUrl = `${base}/pricing?subscription=cancel`;

      const { url } = await createCheckoutSession(priceId, successUrl, cancelUrl);

      // Redirect to Stripe Checkout
      window.location.href = url;
    } catch (error) {
      setCheckoutError({
        message: getCheckoutErrorMessage(error),
        planId,
        priceId,
      });
    } finally {
      setCheckoutInitiating(false);
    }
  };

  const handleRetryCheckout = async () => {
    if (!checkoutError?.priceId) return;

    setIsRetrying(true);
    try {
      const base = window.location.origin;
      const successUrl = user?.username
        ? `${base}/u/${user.username}/settings/billing?subscription=success`
        : `${base}/settings?tab=billing&subscription=success`;
      const cancelUrl = `${base}/pricing?subscription=cancel`;

      const { url } = await createCheckoutSession(checkoutError.priceId, successUrl, cancelUrl);
      window.location.href = url;
    } catch (error) {
      setCheckoutError({
        message: getCheckoutErrorMessage(error),
        planId: checkoutError.planId,
        priceId: checkoutError.priceId,
      });
    } finally {
      setIsRetrying(false);
    }
  };

  const handleContactSales = () => {
    setCheckoutError(null);
    window.location.href = '/contact';
  };

  return (
    <>
      {showConfetti && (
        <Confetti
          width={width}
          height={height}
          recycle={false}
          numberOfPieces={200}
          gravity={0.3}
          colors={['#6366f1', '#8b5cf6', '#06b6d4', '#10b981', '#f59e0b']}
        />
      )}

      {/* SEO Meta Tags */}
      <MetaTags
        title="Pricing | FunctionFly - Serverless Function Deployment Pricing"
        description="Simple, transparent pricing for serverless function deployment. Start free, scale as you grow. 14-day free trial on all paid plans. No hidden fees, no surprise charges."
        keywords={[
          'serverless pricing',
          'function deployment pricing',
          'serverless plans',
          'function as a service pricing',
          'cloud function pricing',
          'serverless hosting pricing',
        ]}
        url="/pricing"
      />

      {/* Structured Data */}
      <PricingPageStructuredData />

      <div className="pricing-page min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 relative overflow-hidden">
        {/* Background decorative elements */}
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_20%_30%,rgba(99,102,241,0.08),transparent_50%)] pointer-events-none" />
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_80%_70%,rgba(139,92,246,0.08),transparent_50%)] pointer-events-none" />
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_50%,rgba(6,182,212,0.05),transparent_50%)] pointer-events-none" />
        <div className="absolute inset-0 bg-[linear-gradient(45deg,transparent_25%,rgba(255,255,255,0.005)_25%,rgba(255,255,255,0.005)_50%,transparent_50%,transparent_75%,rgba(255,255,255,0.005)_75%)] bg-[length:20px_20px] opacity-30 pointer-events-none" />

        {/* Navigation Bar */}
        <nav className="border-b border-white/10 bg-black/30 backdrop-blur-md sticky top-0 z-50 relative overflow-hidden">
          {/* Background gradient overlay */}
          <div className="absolute inset-0 bg-gradient-to-r from-[#6366f1]/5 via-transparent to-[#8b5cf6]/5" />
          <div className="relative max-w-7xl mx-auto px-4 lg:px-6">
            <div className="flex items-center justify-between h-16">
              <Link
                to="/"
                className="flex items-center gap-2 text-white hover:text-[#6366f1] transition-all duration-300 group"
              >
                <div className="p-1 rounded-lg bg-white/5 group-hover:bg-[#6366f1]/10 transition-colors">
                  <ArrowLeft className="w-4 h-4" />
                </div>
                <span className="font-medium">{t("pricing:hero.backToHome")}</span>
              </Link>
              <div className="flex items-center gap-3">
                <div className="w-2 h-2 rounded-full bg-gradient-to-r from-[#6366f1] to-[#8b5cf6] animate-pulse" />
                <h1 className="text-xl font-bold bg-gradient-to-r from-white to-text-secondary bg-clip-text text-transparent">
                  {t("pricing:hero.pricing")}
                </h1>
              </div>
              <div className="w-24" /> {/* Spacer for centering */}
            </div>
          </div>
        </nav>

        <div className="relative z-10 max-w-7xl mx-auto px-4 lg:px-6 py-8">
          {/* Page Header */}
          <div className="text-center py-20 md:py-24">
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5 }}
            >
              <div className="relative inline-block mb-8">
                <div className="w-20 h-20 mx-auto rounded-3xl bg-gradient-to-br from-[#6366f1]/30 via-[#8b5cf6]/20 to-[#06b6d4]/20 border border-[#6366f1]/30 flex items-center justify-center backdrop-blur-sm">
                  <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-[#6366f1] to-[#8b5cf6] flex items-center justify-center">
                    <CreditCard className="w-8 h-8 text-white" />
                  </div>
                </div>
                <div className="absolute -inset-4 bg-gradient-to-r from-[#6366f1]/20 via-[#8b5cf6]/10 to-[#06b6d4]/20 rounded-full blur-xl -z-10" />
              </div>

              <h1 className="text-5xl md:text-6xl lg:text-7xl font-bold text-white mb-6 leading-tight">
                  <span className="bg-gradient-to-r from-white via-white to-text-secondary bg-clip-text text-transparent">
                    {t("pricing:hero.headingLine1")}
                  </span>
                  <br />
                  <span className="bg-gradient-to-r from-[#6366f1] via-[#8b5cf6] to-[#06b6d4] bg-clip-text text-transparent">
                    {t("pricing:hero.headingLine2")}
                  </span>
              </h1>

              <p className="text-text-secondary max-w-2xl mx-auto text-lg md:text-xl mb-6 leading-relaxed font-light">
                {t("pricing:hero.subtitle")}
              </p>

              <div className="flex flex-wrap items-center justify-center gap-x-4 gap-y-2 text-sm">
                <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-green-500/10 border border-green-500/20">
                  <Shield className="w-3.5 h-3.5 text-green-400" />
                  <span className="font-medium text-green-400">{t("pricing:hero.freeTrialBadge")}</span>
                </div>
                <span className="text-text-secondary">
                  {t("pricing:hero.trustBadges")}
                </span>
              </div>
            </motion.div>
          </div>

          {/* Backend-in-a-Box Bundles Promo */}
          <BundleCTACard />

          {/* Product tabs */}
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4, delay: 0.2 }}
            className="flex flex-wrap justify-center gap-2 mb-12"
          >
            {TABS.map((tab) => {
              const Icon = tab.icon;
              const isActive = activeTab === tab.id;
              return (
                <button
                  key={tab.id}
                  type="button"
                  onClick={() => setActiveTab(tab.id)}
                  className={cn(
                    'inline-flex items-center gap-2 px-5 py-2.5 rounded-xl text-sm font-medium transition-all duration-200',
                    isActive
                      ? 'bg-gradient-to-r from-[#6366f1] to-[#8b5cf6] text-white shadow-lg shadow-[#6366f1]/25'
                      : 'bg-white/5 border border-white/10 text-text-secondary hover:border-white/20 hover:text-white'
                  )}
                >
                  <Icon className="w-4 h-4" />
                  {tab.label}
                </button>
              );
            })}
          </motion.div>

          {/* Functions tab */}
          {activeTab === 'functions' && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.4 }}
              className="relative mb-16"
            >
              <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/5 to-transparent blur-3xl -mx-8 rounded-3xl" />
              <div className="relative grid md:grid-cols-2 lg:grid-cols-4 gap-6 max-w-7xl mx-auto pt-6">
                {Object.values(PLANS).map((plan, index) => (
                  <FunctionPlanCard
                    key={plan.id}
                    plan={plan}
                    index={index}
                    onPlanSelect={handlePlanSelect}
                    disabled={checkoutInitiating}
                    isLoading={checkoutInitiating}
                  />
                ))}
              </div>
              <FunctionsComparisonTable />
            </motion.div>
          )}

          {/* State Fabric tab */}
          {activeTab === 'state-fabric' && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.4 }}
              className="mb-16"
            >
              <StateFabricPricingSection compact onPlanSelect={handlePlanSelect} />
            </motion.div>
          )}

          {/* Agents tab */}
          {activeTab === 'agents' && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.4 }}
              className="mb-16"
            >
              <AgentPricingSection compact onPlanSelect={handlePlanSelect} />
            </motion.div>
          )}

          {/* Condensed value strip */}
          <motion.section
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true, amount: 0.2 }}
            transition={{ duration: 0.5 }}
            className="pricing-value-strip grid grid-cols-1 md:grid-cols-3 gap-6 mb-16 py-8 border-y border-white/10"
          >
            <div className="flex items-center gap-4 text-center md:text-left">
              <div className="w-12 h-12 rounded-xl bg-[#6366f1]/20 border border-[#6366f1]/30 flex items-center justify-center shrink-0">
                <Zap className="w-6 h-6 text-[#6366f1]" />
              </div>
              <div>
                <h3 className="font-semibold text-white text-base md:text-lg">{t("pricing:valueStrip.fastFailover")}</h3>
                <p className="text-text-secondary text-base mt-1">
                  {t("pricing:valueStrip.fastFailoverDesc")}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-4 text-center md:text-left">
              <div className="w-12 h-12 rounded-xl bg-emerald-500/20 border border-emerald-500/30 flex items-center justify-center shrink-0">
                <Shield className="w-6 h-6 text-emerald-400" />
              </div>
              <div>
                <h3 className="font-semibold text-white text-base md:text-lg">
                  {t("pricing:valueStrip.enterpriseReliability")}
                </h3>
                <p className="text-text-secondary text-base mt-1">
                  {t("pricing:valueStrip.enterpriseReliabilityDesc")}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-4 text-center md:text-left">
              <div className="w-12 h-12 rounded-xl bg-amber-500/20 border border-amber-500/30 flex items-center justify-center shrink-0">
                <CreditCard className="w-6 h-6 text-amber-400" />
              </div>
              <div>
                <h3 className="font-semibold text-white text-base md:text-lg">{t("pricing:valueStrip.developerFirst")}</h3>
                <p className="text-text-secondary text-base mt-1">
                  {t("pricing:valueStrip.developerFirstDesc")}
                </p>
              </div>
            </div>
          </motion.section>

          <FAQSection />
          <CTASection onPlanSelect={handlePlanSelect} />
        </div>

        <Footer />

        {/* Global Tooltip */}
        <Tooltip
          place="top"
          className="!bg-black !text-white !border !border-white/20 !rounded-lg !text-sm !max-w-xs"
          clickable={false}
          noArrow={false}
          offset={10}
        />

        {/* Toast Notifications */}
        <Toaster
          position="bottom-right"
          toastOptions={{
            style: {
              background: '#1a1a1a',
              color: '#fff',
              border: '1px solid #6366f1',
            },
          }}
        />

        {/* Checkout Error Dialog */}
        <Dialog open={!!checkoutError} onOpenChange={() => setCheckoutError(null)}>
          <DialogContent className="sm:max-w-[500px]">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2">
                <CreditCard className="w-5 h-5 text-red-500" />
                {t("pricing:checkoutDialog.title")}
              </DialogTitle>
              <DialogDescription>
                {t("pricing:checkoutDialog.description")}
              </DialogDescription>
            </DialogHeader>
            <div className="py-4 space-y-4">
              <div className="p-4 rounded-lg bg-red-500/10 border border-red-500/20">
                <p className="text-red-400 text-sm">{checkoutError?.message}</p>
              </div>
              <div className="space-y-2">
                <p className="text-sm text-text-muted">{t("pricing:checkoutDialog.prompt")}</p>
                <ul className="text-sm text-text-secondary space-y-1 ml-4 list-disc">
                  <li>{t("pricing:checkoutDialog.tryAgainOption")}</li>
                  <li>{t("pricing:checkoutDialog.contactSalesOption")}</li>
                  <li>{t("pricing:checkoutDialog.tryLaterOption")}</li>
                </ul>
              </div>
            </div>
            <DialogFooter className="flex gap-2 sm:gap-0">
              <Button variant="outline" onClick={() => setCheckoutError(null)}>
                {t("pricing:checkoutDialog.close")}
              </Button>
              <Button variant="outline" onClick={handleContactSales} className="gap-2">
                <Mail className="w-4 h-4" />
                {t("pricing:checkoutDialog.contactSales")}
              </Button>
              <Button
                onClick={handleRetryCheckout}
                disabled={isRetrying}
                className="bg-[#6366f1] hover:bg-[#6366f1]/90"
              >
                {isRetrying ? t("pricing:checkoutDialog.retrying") : t("pricing:checkoutDialog.tryAgain")}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </>
  );
}
