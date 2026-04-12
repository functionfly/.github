import { createCheckoutSession } from '@/api/billing';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { usePlan } from '@/hooks/usePlan';
import { PLANS } from '@/lib/constants';
import { formatLimit } from '@/lib/plan-limits';
import { cn } from '@/lib/utils';
import { useAuthStore } from '@/stores/authStore';
import { motion } from 'framer-motion';
import {
  ArrowRight,
  Check,
  ChevronDown,
  Crown,
  Loader2,
  Lock,
  Rocket,
  Sparkles,
  TrendingUp,
  Zap,
} from 'lucide-react';
import { useState } from 'react';
import toast from 'react-hot-toast';
import { useNavigate } from 'react-router-dom';

interface UpgradeBannerProps {
  /** Where the banner is displayed - affects styling */
  placement?: 'sidebar' | 'dashboard' | 'page';
  /** Optional className for custom styling */
  className?: string;
  /** Whether to show usage stats alongside the upgrade CTA */
  showUsage?: boolean;
  /** Current usage numbers for comparison (if available) */
  usage?: {
    functions?: number;
    providers?: number;
    requests?: number;
  };
  /** Optional feature name for feature gate banners */
  featureName?: string;
}

/**
 * Upgrade banner component for free/paid plan users
 * Shows prominently in sidebar, dashboard, or as a page banner
 * Only visible to users on free or lower-tier plans
 */
export function UpgradeBanner({
  placement = 'sidebar',
  className,
  showUsage = true,
  usage,
  featureName,
}: UpgradeBannerProps) {
  const { plan, isPaid, hasMinPlan, limits, planColor } = usePlan();
  const user = useAuthStore((state) => state.user);
  const navigate = useNavigate();

  const [isCheckoutLoading, setIsCheckoutLoading] = useState(false);
  const [showPlanModal, setShowPlanModal] = useState(false);
  const [isCollapsed, setIsCollapsed] = useState(false);

  // Don't show for enterprise users
  if (plan?.toLowerCase() === 'enterprise') {
    return null;
  }

  const isFree = !isPaid;
  const planLimits = limits;
  const nextTier = isFree ? 'Starter' : !hasMinPlan('professional') ? 'Professional' : 'Enterprise';
  const nextPlan = isFree ? PLANS.STARTER : !hasMinPlan('professional') ? PLANS.PROFESSIONAL : null;

  // Dynamic color based on next tier
  const tierColor = isFree ? 'indigo' : nextTier === 'Professional' ? 'violet' : 'amber';

  const handleUpgrade = () => {
    // For sidebar/page: show plan selection modal
    // For dashboard: also show plan selection for better UX
    if (placement === 'sidebar' || placement === 'page' || placement === 'dashboard') {
      setShowPlanModal(true);
    }
  };

  const handleDirectCheckout = async (planId: string, priceId?: string) => {
    if (!priceId || priceId.includes('placeholder')) {
      // For enterprise without price ID, go to contact page
      if (planId === 'enterprise') {
        navigate('/contact');
        return;
      }

      // For paid plans with placeholder price IDs, show error and don't proceed
      toast.error(
        'Billing is not fully configured. Please contact support to complete your upgrade.',
        {
          duration: 5000,
          style: {
            background: '#1a1a1a',
            color: '#fff',
            border: '1px solid #ef4444',
          },
        }
      );
      return;
    }

    setIsCheckoutLoading(true);
    try {
      const base = window.location.origin;
      const successUrl = user?.username
        ? `${base}/u/${user.username}/settings/billing?subscription=success`
        : `${base}/settings?tab=billing&subscription=success`;
      const cancelUrl = `${base}/dashboard?subscription=cancel`;

      const { url } = await createCheckoutSession(priceId, successUrl, cancelUrl);
      window.location.href = url;
    } catch (err) {
      setIsCheckoutLoading(false);
      console.error('Checkout error:', err);
      toast.error('Unable to start checkout. Please try again or contact support.', {
        duration: 5000,
        style: {
          background: '#1a1a1a',
          color: '#fff',
          border: '1px solid #ef4444',
        },
      });
    }
  };

  const handleCloseModal = () => {
    setShowPlanModal(false);
    setIsCheckoutLoading(false);
  };

  // Calculate usage percentages
  const functionLimit = planLimits?.functions ?? 0;
  const providerLimit = planLimits?.providers ?? 0;
  const requestLimit = planLimits?.requests ?? 0;

  const functionPercent =
    functionLimit > 0 && usage?.functions
      ? Math.min((usage.functions / functionLimit) * 100, 100)
      : 0;
  const providerPercent =
    providerLimit > 0 && usage?.providers
      ? Math.min((usage.providers / providerLimit) * 100, 100)
      : 0;
  const requestPercent =
    requestLimit > 0 && usage?.requests ? Math.min((usage.requests / requestLimit) * 100, 100) : 0;

  // Determine if user is approaching limits
  const isApproachingLimit = functionPercent >= 80 || providerPercent >= 80 || requestPercent >= 80;

  // Sidebar placement - collapsible compact premium card with aviation theme
  if (placement === 'sidebar') {
    return (
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        className={cn('px-3 pb-3', className)}
      >
        <div className={cn('aviation-upgrade-banner aviation-upgrade-collapsible', isCollapsed && 'collapsed')}>
          {/* Collapsible Header / Toggle */}
          <button
            onClick={() => setIsCollapsed(!isCollapsed)}
            className="aviation-upgrade-toggle"
            aria-label={isCollapsed ? 'Expand upgrade banner' : 'Collapse upgrade banner'}
          >
            {isCollapsed ? (
              <div className="aviation-upgrade-compact">
                <div className="aviation-upgrade-compact-icon">
                  {isFree ? <Rocket /> : <Crown />}
                </div>
                <div className="aviation-upgrade-compact-text">
                  <span className="aviation-upgrade-compact-title">
                    {isFree ? 'Unlock Pro' : `Upgrade to ${nextTier}`}
                  </span>
                  <span className="aviation-upgrade-compact-hint">Click to expand</span>
                </div>
              </div>
            ) : (
              <div className="aviation-upgrade-toggle-content">
                <div className="aviation-upgrade-icon" style={{ width: '32px', height: '32px' }}>
                  {isFree ? <Rocket /> : <Crown />}
                </div>
                <span className="aviation-upgrade-title">
                  {isFree ? 'Unlock Pro Features' : `Upgrade to ${nextTier}`}
                </span>
                {isApproachingLimit && <span className="aviation-upgrade-badge-indicator" />}
              </div>
            )}
            <ChevronDown className="aviation-upgrade-collapse-icon" />
          </button>

          {/* Expandable Content */}
          <div className="aviation-upgrade-content">
            <div className="space-y-4">
              {/* Description when expanded */}
              <p className="aviation-upgrade-desc">
                {isFree ? 'Get more power & flexibility' : 'Take your account to the next level'}
              </p>

              {/* Usage meters */}
              {showUsage && isFree && (
                <div className="space-y-3">
                  <div className="space-y-1.5">
                    <div className="flex justify-between items-center">
                      <span className="aviation-upgrade-meter-label">Functions</span>
                      <span className={cn(
                        'aviation-upgrade-meter-value',
                        functionPercent >= 95 ? 'text-aviation-red' : functionPercent >= 80 ? 'text-aviation-amber' : ''
                      )}>
                        <strong>{usage?.functions ?? 0}</strong>
                        <span className="text-aviation-text-muted font-normal"> / {functionLimit}</span>
                      </span>
                    </div>
                    <div className={cn('aviation-upgrade-meter', functionPercent >= 80 && 'aviation-upgrade-meter-warning')}>
                      <motion.div
                        initial={{ width: 0 }}
                        animate={{ width: `${functionPercent}%` }}
                        transition={{ duration: 0.5, ease: 'easeOut' }}
                        className="aviation-upgrade-meter-fill"
                      />
                    </div>
                  </div>
                  <div className="space-y-1.5">
                    <div className="flex justify-between items-center">
                      <span className="aviation-upgrade-meter-label">Providers</span>
                      <span className={cn(
                        'aviation-upgrade-meter-value',
                        providerPercent >= 95 ? 'text-aviation-red' : providerPercent >= 80 ? 'text-aviation-amber' : ''
                      )}>
                        <strong>{usage?.providers ?? 0}</strong>
                        <span className="text-aviation-text-muted font-normal"> / {providerLimit}</span>
                      </span>
                    </div>
                    <div className={cn('aviation-upgrade-meter', providerPercent >= 80 && 'aviation-upgrade-meter-warning')}>
                      <motion.div
                        initial={{ width: 0 }}
                        animate={{ width: `${providerPercent}%` }}
                        transition={{ duration: 0.5, ease: 'easeOut' }}
                        className="aviation-upgrade-meter-fill"
                      />
                    </div>
                  </div>
                </div>
              )}

              {/* Warning for approaching limits */}
              {isApproachingLimit && (
                <motion.div
                  initial={{ opacity: 0, scale: 0.95 }}
                  animate={{ opacity: 1, scale: 1 }}
                  className="aviation-upgrade-alert"
                >
                  <TrendingUp />
                  <span>Approaching limits</span>
                </motion.div>
              )}

              {/* CTA Button */}
              <button
                onClick={handleUpgrade}
                className="aviation-upgrade-cta"
              >
                {isFree ? 'Upgrade Now' : 'View Plans'}
                <ArrowRight />
              </button>

              {/* Pricing hint */}
              {isFree && nextPlan && (
                <p className="aviation-upgrade-price">
                  Starting at <strong>${nextPlan.price}/mo</strong>
                </p>
              )}
            </div>
          </div>
        </div>

        {/* Plan Selection Modal */}
        <PlanSelectionModal
          isOpen={showPlanModal}
          onClose={handleCloseModal}
          isFree={isFree}
          nextTier={nextTier}
          onSelectPlan={handleDirectCheckout}
          isLoading={isCheckoutLoading}
        />
      </motion.div>
    );
  }

  // Dashboard placement - prominent banner with glassmorphism
  if (placement === 'dashboard') {
    return (
      <motion.div
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        className={cn('mb-6', className)}
      >
        <Card
          className={cn(
            'relative overflow-hidden border-0 shadow-xl',
            tierColor === 'indigo' && 'ring-1 ring-indigo-500/20',
            tierColor === 'violet' && 'ring-1 ring-violet-500/20',
            tierColor === 'amber' && 'ring-1 ring-amber-500/20'
          )}
        >
          {/* Rich gradient background */}
          <div
            className={cn(
              'absolute inset-0',
              tierColor === 'indigo' &&
                'bg-linear-to-br from-indigo-600/15 via-purple-600/10 to-fuchsia-500/15',
              tierColor === 'violet' &&
                'bg-linear-to-br from-violet-600/15 via-purple-600/10 to-indigo-500/15',
              tierColor === 'amber' &&
                'bg-linear-to-br from-amber-600/15 via-orange-500/10 to-yellow-500/15'
            )}
          />

          {/* Animated glow effect */}
          <div
            className={cn(
              'absolute -top-24 -right-24 w-48 h-48 rounded-full blur-3xl opacity-40',
              tierColor === 'indigo' && 'bg-indigo-500',
              tierColor === 'violet' && 'bg-violet-500',
              tierColor === 'amber' && 'bg-amber-500'
            )}
          />

          {/* Subtle noise texture */}
          <div
            className="absolute inset-0 opacity-[0.015] mix-blend-overlay"
            style={{
              backgroundImage: `url("data:image/svg+xml,%3Csvg viewBox='0 0 256 256' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noiseFilter'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noiseFilter)'/%3E%3C/svg%3E")`,
            }}
          />

          <CardContent className="relative p-6 sm:p-8">
            <div className="flex flex-col lg:flex-row lg:items-center gap-6">
              {/* Left side - Icon and headline */}
              <div className="flex items-center gap-4">
                <div
                  className={cn(
                    'w-14 h-14 rounded-2xl flex items-center justify-center shadow-xl shrink-0',
                    tierColor === 'indigo' && 'bg-linear-to-br from-indigo-500 to-purple-600',
                    tierColor === 'violet' && 'bg-linear-to-br from-violet-500 to-purple-600',
                    tierColor === 'amber' && 'bg-linear-to-br from-amber-500 to-orange-500'
                  )}
                >
                  {isFree ? (
                    <Zap className="w-7 h-7 text-white" />
                  ) : (
                    <Crown className="w-7 h-7 text-white" />
                  )}
                </div>
                <div>
                  <h3 className="text-xl font-bold text-text-primary">
                    {isFree ? 'Ready to Scale?' : `Unlock ${nextTier} Power`}
                  </h3>
                  <p className="text-text-secondary mt-1 max-w-md">
                    {isFree
                      ? 'Upgrade to Starter and unlock more functions, providers, and priority support.'
                      : `Get advanced features, higher limits, and ${nextTier === 'Enterprise' ? 'dedicated' : 'priority'} support.`}
                  </p>
                </div>
              </div>

              {/* Right side - Pricing and CTA */}
              <div className="flex items-center gap-4 lg:ml-auto">
                {isFree && nextPlan && (
                  <div className="hidden sm:block text-right">
                    <p className="text-3xl font-bold text-text-primary">${nextPlan.price}</p>
                    <p className="text-sm text-text-secondary">per month</p>
                  </div>
                )}
                <Button
                  onClick={handleUpgrade}
                  size="lg"
                  className={cn(
                    'font-semibold shadow-lg shadow-black/20 px-6',
                    tierColor === 'indigo' &&
                      'bg-linear-to-r from-indigo-500 to-purple-600 hover:from-indigo-600 hover:to-purple-700 text-white',
                    tierColor === 'violet' &&
                      'bg-linear-to-r from-violet-500 to-purple-600 hover:from-violet-600 hover:to-purple-700 text-white',
                    tierColor === 'amber' &&
                      'bg-linear-to-r from-amber-500 to-orange-500 hover:from-amber-600 hover:to-orange-600 text-white'
                  )}
                >
                  Upgrade Now
                  <ArrowRight className="w-5 h-5 ml-2" />
                </Button>
              </div>
            </div>

            {/* Usage stats section */}
            {showUsage && isFree && (
              <div className="mt-8 pt-6 border-t border-border-subtle/50">
                <div className="flex items-center gap-2 mb-4">
                  <Sparkles className="w-4 h-4 text-text-muted" />
                  <p className="text-sm font-medium text-text-secondary">Your current usage</p>
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                  <UsageStat
                    label="Functions"
                    current={usage?.functions ?? 0}
                    limit={functionLimit}
                    percent={functionPercent}
                    color={tierColor}
                  />
                  <UsageStat
                    label="Providers"
                    current={usage?.providers ?? 0}
                    limit={providerLimit}
                    percent={providerPercent}
                    color={tierColor}
                  />
                  <UsageStat
                    label="Requests"
                    current={usage?.requests ?? 0}
                    limit={requestLimit}
                    percent={requestPercent}
                    color={tierColor}
                    format="number"
                  />
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Plan Selection Modal */}
        <PlanSelectionModal
          isOpen={showPlanModal}
          onClose={handleCloseModal}
          isFree={isFree}
          nextTier={nextTier}
          onSelectPlan={handleDirectCheckout}
          isLoading={isCheckoutLoading}
        />
      </motion.div>
    );
  }

  // Page placement - elegant inline banner for feature gates
  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.98 }}
      animate={{ opacity: 1, scale: 1 }}
      className={cn('my-4', className)}
    >
      <Card
        className={cn(
          'relative overflow-hidden border-dashed',
          tierColor === 'indigo' && 'border-indigo-500/30 bg-indigo-500/[0.03]',
          tierColor === 'violet' && 'border-violet-500/30 bg-violet-500/[0.03]',
          tierColor === 'amber' && 'border-amber-500/30 bg-amber-500/[0.03]'
        )}
      >
        {/* Subtle background */}
        <div
          className={cn(
            'absolute inset-0 opacity-50',
            tierColor === 'indigo' && 'bg-linear-to-r from-indigo-500/5 to-purple-500/5',
            tierColor === 'violet' && 'bg-linear-to-r from-violet-500/5 to-purple-500/5',
            tierColor === 'amber' && 'bg-linear-to-r from-amber-500/5 to-orange-500/5'
          )}
        />

        <CardContent className="relative p-5 flex flex-col sm:flex-row items-center gap-4">
          <div
            className={cn(
              'w-12 h-12 rounded-xl flex items-center justify-center shrink-0',
              tierColor === 'indigo' && 'bg-indigo-500/10',
              tierColor === 'violet' && 'bg-violet-500/10',
              tierColor === 'amber' && 'bg-amber-500/10'
            )}
          >
            <Lock
              className={cn(
                'w-6 h-6',
                tierColor === 'indigo' && 'text-indigo-400',
                tierColor === 'violet' && 'text-violet-400',
                tierColor === 'amber' && 'text-amber-400'
              )}
            />
          </div>
          <div className="flex-1 text-center sm:text-left">
            <p className="font-semibold text-text-primary">
              {featureName
                ? `${featureName} requires ${isFree ? 'Starter' : nextTier} plan`
                : `This feature requires ${isFree ? 'Starter' : nextTier} plan`}
            </p>
            <p className="text-sm text-text-secondary mt-0.5">
              Upgrade now to unlock this feature and more premium capabilities.
            </p>
          </div>
          <Button
            onClick={handleUpgrade}
            variant="outline"
            className={cn(
              'shrink-0 font-medium',
              tierColor === 'indigo' &&
                'border-indigo-500/30 hover:border-indigo-500/50 hover:bg-indigo-500/10 text-indigo-400',
              tierColor === 'violet' &&
                'border-violet-500/30 hover:border-violet-500/50 hover:bg-violet-500/10 text-violet-400',
              tierColor === 'amber' &&
                'border-amber-500/30 hover:border-amber-500/50 hover:bg-amber-500/10 text-amber-400'
            )}
          >
            Upgrade
            <ArrowRight className="w-4 h-4 ml-1.5" />
          </Button>
        </CardContent>
      </Card>

      {/* Plan Selection Modal */}
      <PlanSelectionModal
        isOpen={showPlanModal}
        onClose={handleCloseModal}
        isFree={isFree}
        nextTier={nextTier}
        onSelectPlan={handleDirectCheckout}
        isLoading={isCheckoutLoading}
      />
    </motion.div>
  );
}

/**
 * Plan Selection Modal Component - exported for reuse
 */
export interface PlanSelectionModalProps {
  isOpen: boolean;
  onClose: () => void;
  isFree: boolean;
  nextTier: string;
  onSelectPlan: (planId: string, priceId?: string) => void;
  isLoading: boolean;
  featureName?: string;
}

export function PlanSelectionModal({
  isOpen,
  onClose,
  isFree,
  nextTier,
  onSelectPlan,
  isLoading,
  featureName,
}: PlanSelectionModalProps) {
  const { plan } = usePlan();

  // Determine which plans to show
  const availablePlans = isFree
    ? [
        { id: 'starter', plan: PLANS.STARTER, recommended: true },
        { id: 'professional', plan: PLANS.PROFESSIONAL, recommended: false },
        { id: 'enterprise', plan: PLANS.ENTERPRISE, recommended: false },
      ]
    : plan?.toLowerCase() === 'starter'
      ? [
          { id: 'professional', plan: PLANS.PROFESSIONAL, recommended: true },
          { id: 'enterprise', plan: PLANS.ENTERPRISE, recommended: false },
        ]
      : [{ id: 'enterprise', plan: PLANS.ENTERPRISE, recommended: true }];

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[600px] p-0 overflow-hidden">
        <div className="bg-linear-to-br from-indigo-500/10 via-purple-500/10 to-fuchsia-500/10 p-6 pb-4">
          <DialogHeader>
            <DialogTitle className="text-2xl font-bold text-text-primary">
              {isFree
                ? featureName
                  ? `Upgrade to Use ${featureName}`
                  : 'Choose Your Plan'
                : `Upgrade to ${nextTier}`}
            </DialogTitle>
            <DialogDescription className="text-text-secondary">
              {isFree
                ? featureName
                  ? `Upgrade to Starter or higher to create and deploy ${featureName.toLowerCase()}. Select a plan that fits your needs.`
                  : 'Select a plan that fits your needs. Upgrade or downgrade anytime.'
                : 'Unlock more features and higher limits with your upgrade.'}
            </DialogDescription>
          </DialogHeader>
        </div>

        <div className="p-6 pt-4 space-y-4">
          {availablePlans.map(({ id, plan: planData, recommended }) => (
            <PlanCard
              key={id}
              id={id}
              plan={planData}
              isRecommended={recommended}
              isLoading={isLoading}
              onSelect={() => onSelectPlan(id, planData.priceId)}
            />
          ))}

          <p className="text-xs text-text-muted text-center pt-2">
            All plans include core features. Cancel anytime.
          </p>
        </div>
      </DialogContent>
    </Dialog>
  );
}

/**
 * Individual Plan Card for the Modal
 */
function PlanCard({
  id,
  plan,
  isRecommended,
  isLoading,
  onSelect,
}: {
  id: string;
  plan: {
    id: string;
    name: string;
    price: number | string;
    priceCents: number;
    priceId: string;
    description: string;
    features: readonly string[];
    limits: {
      functions: number;
      providers: number;
      requests: number;
      customDomains?: number;
      stateFabrics?: number;
      agents?: number;
      secrets?: number;
      tokensPerSecret?: number;
      sla?: string;
    };
  };
  isRecommended: boolean;
  isLoading: boolean;
  onSelect: () => void;
}) {
  const isEnterprise = id === 'enterprise';

  return (
    <motion.div
      whileHover={{ scale: 1.01 }}
      whileTap={{ scale: 0.99 }}
      className={cn(
        'relative rounded-xl border transition-all cursor-pointer overflow-hidden',
        isRecommended
          ? 'border-indigo-500/50 bg-indigo-500/5 ring-1 ring-indigo-500/20'
          : 'border-border-subtle hover:border-border-strong hover:bg-surface-elevated/50'
      )}
      onClick={onSelect}
    >
      {isRecommended && (
        <div className="absolute top-0 right-0 bg-indigo-500 text-white text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-bl-lg">
          Recommended
        </div>
      )}

      <div className="p-4 flex items-center gap-4">
        <div
          className={cn(
            'w-12 h-12 rounded-xl flex items-center justify-center shrink-0',
            isEnterprise
              ? 'bg-linear-to-br from-amber-500 to-orange-500'
              : isRecommended
                ? 'bg-linear-to-br from-indigo-500 to-purple-600'
                : 'bg-surface-elevated'
          )}
        >
          {isEnterprise ? (
            <Crown className="w-6 h-6 text-white" />
          ) : isRecommended ? (
            <Rocket className="w-6 h-6 text-white" />
          ) : (
            <Zap className="w-5 h-5 text-text-muted" />
          )}
        </div>

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <h4 className="font-semibold text-text-primary">{plan.name}</h4>
            {isEnterprise && (
              <span className="text-[10px] bg-amber-500/20 text-amber-400 px-1.5 py-0.5 rounded-full font-medium">
                Custom
              </span>
            )}
          </div>
          <p className="text-sm text-text-secondary truncate">{plan.description}</p>
          <div className="flex items-center gap-3 mt-1.5 text-xs text-text-muted">
            <span className="flex items-center gap-1">
              <Check className="w-3 h-3 text-green-400" />
              {plan.limits.functions === Infinity ? 'Unlimited' : plan.limits.functions} functions
            </span>
            <span className="flex items-center gap-1">
              <Check className="w-3 h-3 text-green-400" />
              {plan.limits.providers === Infinity ? 'Unlimited' : plan.limits.providers} providers
            </span>
          </div>
        </div>

        <div className="text-right shrink-0">
          <p className="text-2xl font-bold text-text-primary">
            {typeof plan.price === 'number' ? `$${plan.price}` : plan.price}
          </p>
          <p className="text-xs text-text-muted">{isEnterprise ? 'Contact us' : '/month'}</p>
        </div>
      </div>

      <div className="px-4 pb-4">
        <Button
          className={cn(
            'w-full',
            isEnterprise
              ? 'bg-linear-to-r from-amber-500 to-orange-500 hover:from-amber-600 hover:to-orange-600 text-white'
              : isRecommended
                ? 'bg-linear-to-r from-indigo-500 to-purple-600 hover:from-indigo-600 hover:to-purple-700 text-white'
                : 'bg-surface-elevated hover:bg-surface-elevated/80 text-text-primary border border-border-subtle'
          )}
          disabled={isLoading}
          onClick={(e) => {
            e.stopPropagation();
            onSelect();
          }}
        >
          {isLoading ? (
            <>
              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
              Loading...
            </>
          ) : isEnterprise ? (
            'Contact Sales'
          ) : (
            <>
              Select {plan.name}
              <ArrowRight className="w-4 h-4 ml-1" />
            </>
          )}
        </Button>
      </div>
    </motion.div>
  );
}

/**
 * Compact usage meter for sidebar
 */
function UsageMeter({
  label,
  current,
  limit,
  percent,
  color,
}: {
  label: string;
  current: number;
  limit: number;
  percent: number;
  color: string;
}) {
  const isWarning = percent >= 80;
  const isCritical = percent >= 95;

  return (
    <div className="space-y-1.5">
      <div className="flex justify-between text-xs">
        <span className="text-text-secondary font-medium">{label}</span>
        <span
          className={cn(
            'font-semibold tabular-nums',
            isCritical ? 'text-red-400' : isWarning ? 'text-amber-400' : 'text-text-primary'
          )}
        >
          {current} <span className="text-text-muted font-normal">/ {formatLimit(limit)}</span>
        </span>
      </div>
      <div className="h-1.5 w-full rounded-full bg-surface-elevated overflow-hidden">
        <motion.div
          initial={{ width: 0 }}
          animate={{ width: `${percent}%` }}
          transition={{ duration: 0.5, ease: 'easeOut' }}
          className={cn(
            'h-full rounded-full',
            isCritical
              ? 'bg-red-500'
              : isWarning
                ? 'bg-amber-500'
                : color === 'indigo'
                  ? 'bg-indigo-500'
                  : color === 'violet'
                    ? 'bg-violet-500'
                    : 'bg-amber-500'
          )}
        />
      </div>
    </div>
  );
}

/**
 * Full usage stat for dashboard with detailed info
 */
function UsageStat({
  label,
  current,
  limit,
  percent,
  color,
  format = 'compact',
}: {
  label: string;
  current: number;
  limit: number;
  percent: number;
  color: string;
  format?: 'compact' | 'number';
}) {
  const isUnlimited = limit === Infinity || limit >= 10000;
  const displayLimit = formatLimit(limit);
  const isWarning = percent >= 80;

  return (
    <div className="bg-surface-elevated/50 rounded-xl p-4 border border-border-subtle/50">
      <div className="flex justify-between items-center mb-3">
        <span className="text-sm font-medium text-text-secondary">{label}</span>
        <span
          className={cn(
            'text-sm font-bold tabular-nums',
            isWarning ? 'text-amber-400' : 'text-text-primary'
          )}
        >
          {format === 'number'
            ? `${current.toLocaleString()} / ${displayLimit}`
            : `${current} / ${displayLimit}`}
        </span>
      </div>
      <div className="h-2 w-full rounded-full bg-surface-elevated overflow-hidden">
        <motion.div
          initial={{ width: 0 }}
          animate={{ width: `${Math.max(percent, 3)}%` }}
          transition={{ duration: 0.6, ease: 'easeOut', delay: 0.1 }}
          className={cn(
            'h-full rounded-full',
            color === 'indigo' && 'bg-indigo-500',
            color === 'violet' && 'bg-violet-500',
            color === 'amber' && 'bg-amber-500'
          )}
        />
      </div>
      <p className="text-xs text-text-muted mt-2">
        {isUnlimited ? 'Unlimited on next tier' : `${Math.round(100 - percent)}% remaining`}
      </p>
    </div>
  );
}
