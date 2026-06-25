import { motion } from 'framer-motion';
import { Check, CreditCard, Sparkles, Zap } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { HelpTooltip } from '@/components/ui/help-tooltip';
import { useOnboardingStore, type PlanTier } from '@/stores/onboardingStore';
import { PLANS } from '@/lib/constants';

const PLAN_ORDER: PlanTier[] = ['free', 'starter', 'professional', 'enterprise'];

export function PlanSelectionStep() {
  const { selectedPlan, setSelectedPlan, billingCycle } = useOnboardingStore();

  const currentPlan = selectedPlan || 'free';

  const getPlanFeatures = (planId: string): string[] => {
    const plan = PLANS[planId.toUpperCase() as keyof typeof PLANS];
    return plan?.features ? [...plan.features] : [];
  };

  return (
    <div className="space-y-6">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-center space-y-4"
      >
        <div className="onboarding-step-icon w-16 h-16 rounded-2xl flex items-center justify-center mx-auto">
          <Sparkles className="w-8 h-8 text-aviation-amber" />
        </div>
        <p className="text-lg text-aviation-text-secondary font-mono max-w-xl mx-auto">
          Choose the plan that best fits your needs. You can upgrade or downgrade anytime.
        </p>
      </motion.div>

      <div className="grid gap-4">
        {PLAN_ORDER.map((planId, index) => {
          const plan = PLANS[planId.toUpperCase() as keyof typeof PLANS];
          if (!plan) return null;

          const isSelected = currentPlan === planId;
          const isPopular = planId === 'professional';

          return (
            <motion.div
              key={planId}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.1 }}
            >
              <Card
                className={`onboarding-step-card p-4 cursor-pointer transition-all relative overflow-hidden ${
                  isSelected
                    ? 'border-aviation-amber shadow-lg shadow-aviation-amber/20'
                    : 'hover:border-aviation-amber/50'
                }`}
                onClick={() => setSelectedPlan(planId, billingCycle)}
              >
                {isPopular && (
                  <div className="absolute top-0 right-0 bg-aviation-amber text-aviation-bg-primary text-xs font-mono font-bold px-3 py-1 rounded-bl-lg">
                    POPULAR
                  </div>
                )}

                <div className="flex items-start gap-4">
                  <div
                    className={`w-6 h-6 rounded-full border-2 flex items-center justify-center flex-shrink-0 mt-1 ${
                      isSelected
                        ? 'border-aviation-amber bg-aviation-amber'
                        : 'border-aviation-border-panel'
                    }`}
                  >
                    {isSelected && <Check className="w-4 h-4 text-aviation-bg-primary" />}
                  </div>

                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <h3 className="font-mono font-bold text-aviation-text-primary text-lg">
                        {plan.name}
                      </h3>
                      {plan.price === 0 && (
                        <span className="text-xs font-mono bg-aviation-green/20 text-aviation-green px-2 py-0.5 rounded">
                          FREE
                        </span>
                      )}
                      {isSelected && (
                        <span className="text-xs font-mono bg-aviation-amber/20 text-aviation-amber px-2 py-0.5 rounded">
                          SELECTED
                        </span>
                      )}
                    </div>

                    <div className="flex items-baseline gap-1 mb-3">
                      <span className="text-2xl font-mono font-bold text-aviation-text-primary">
                        ${plan.price}
                      </span>
                      <span className="text-aviation-text-muted font-mono">/month</span>
                      {plan.price > 0 && (
                        <span className="text-xs text-aviation-text-muted font-mono ml-2">
                          or ${Math.round(plan.price * 10)}/year (2 months free)
                        </span>
                      )}
                    </div>

                    <p className="text-sm text-aviation-text-secondary font-mono mb-3">
                      {plan.description}
                    </p>

                    <div className="flex flex-wrap gap-2">
                      {getPlanFeatures(planId).slice(0, 5).map((feature, i) => (
                        <div
                          key={i}
                          className="text-xs font-mono bg-aviation-bg-tertiary text-aviation-text-secondary px-2 py-1 rounded"
                        >
                          {feature}
                        </div>
                      ))}
                      {getPlanFeatures(planId).length > 5 && (
                        <div className="text-xs font-mono bg-aviation-bg-tertiary text-aviation-text-muted px-2 py-1 rounded">
                          +{getPlanFeatures(planId).length - 5} more
                        </div>
                      )}
                    </div>
                  </div>

                  <div className="flex-shrink-0">
                    {planId !== 'free' && (
                      <Button
                        variant={isSelected ? 'default' : 'outline'}
                        size="sm"
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedPlan(planId, billingCycle);
                        }}
                        className={`font-mono ${
                          isSelected
                            ? 'bg-aviation-amber text-aviation-bg-primary hover:bg-aviation-amber/90'
                            : ''
                        }`}
                      >
                        {isSelected ? 'Selected' : 'Select'}
                      </Button>
                    )}
                  </div>
                </div>
              </Card>
            </motion.div>
          );
        })}
      </div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.5 }}
        className="flex items-center justify-center gap-2 text-sm text-aviation-text-muted font-mono"
      >
        <CreditCard className="w-4 h-4" />
        <span>All plans include a 14-day free trial. No credit card required for free tier.</span>
      </motion.div>

      {currentPlan === 'free' && (
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.6 }}
          className="bg-aviation-amber/10 border border-aviation-amber/30 rounded-lg p-4"
        >
          <div className="flex items-start gap-3">
            <Zap className="w-5 h-5 text-aviation-amber flex-shrink-0 mt-0.5" />
            <div>
              <h4 className="font-mono font-semibold text-aviation-amber mb-1">
                Ready to Scale?
              </h4>
              <p className="text-sm text-aviation-text-secondary font-mono">
                Upgrade anytime to unlock more providers, custom domains, and advanced features.
                Your free tier includes 3 functions and 2 providers.
              </p>
            </div>
          </div>
        </motion.div>
      )}
    </div>
  );
}
