import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { PLANS } from '@/lib/constants';
import { motion } from 'framer-motion';
import { Check, Loader2, Star } from 'lucide-react';
import { Link } from 'react-router-dom';
import { useCardGestures, useScrollAnimation } from '../hooks';

function cn(...classes: (string | boolean | undefined)[]) {
  return classes.filter(Boolean).join(' ');
}

type PlanId = keyof typeof PLANS;
type Plan = (typeof PLANS)[PlanId];

interface FunctionPlanCardProps {
  plan: Plan;
  index: number;
  onPlanSelect: (planId: string, priceId?: string) => void;
  disabled?: boolean;
  isLoading?: boolean;
  billingCycle?: 'monthly' | 'annual';
}

const FEATURE_TOOLTIPS: Record<string, string> = {
  '3 functions': 'Deploy up to 3 serverless functions',
  '5 functions': 'Deploy up to 5 serverless functions',
  '25 functions': 'Deploy up to 25 serverless functions',
  'Unlimited functions': 'Deploy unlimited serverless functions',
  '2 providers': 'Deploy to Vercel or Netlify',
  '3 providers': 'Deploy to Vercel, Netlify, or Fly.io',
  '5 providers': 'Deploy to all supported providers',
  'All providers': 'Deploy to all supported providers',
  '25K requests/month': '25,000 function invocations per month',
  '250K requests/month': '250,000 function invocations per month',
  '2.5M requests/month': '2.5 million function invocations per month',
  'Unlimited requests': 'Unlimited function invocations',
  '1 custom domain': 'Connect one custom domain',
  '5 custom domains': 'Connect up to 5 custom domains',
  'Unlimited custom domains': 'Connect unlimited custom domains',
  'Email support': 'Get help via email during business hours',
  'Priority support': '24/7 priority email and chat support',
  'Dedicated support': 'Dedicated account manager and phone support',
  'Basic analytics': 'View basic usage metrics and logs',
  'Advanced analytics': 'Detailed analytics with custom dashboards',
  'Custom analytics': 'White-labeled analytics with custom integrations',
  'Team collaboration': 'Invite team members to collaborate',
};

export function FunctionPlanCard({
  plan,
  index,
  onPlanSelect,
  disabled = false,
  isLoading = false,
  billingCycle = 'monthly',
}: FunctionPlanCardProps) {
  const { ref, inView } = useScrollAnimation(0.2, false);
  const gestures = useCardGestures(plan.name);

  const displayPrice = billingCycle === 'annual' && plan.priceAnnualCents
    ? Math.round(plan.priceAnnualCents / 12 / 100)
    : plan.price;
  const annualPrice = plan.priceAnnualCents ? plan.priceAnnualCents / 100 : 0;
  const monthlyEquivalentAnnual = typeof plan.price === 'number' && plan.price !== 0
    ? (annualPrice / 12).toFixed(0)
    : null;

  return (
    <motion.div
      ref={ref}
      {...gestures.bind()}
      initial={{ opacity: 0, y: 30, scale: 0.95 }}
      animate={inView ? { opacity: 1, y: 0, scale: 1 } : { opacity: 0, y: 30, scale: 0.95 }}
      transition={{
        duration: 0.6,
        delay: index * 0.1,
        ease: [0.25, 0.46, 0.45, 0.94],
      }}
      style={gestures.style}
      className="transition-shadow duration-300"
      onMouseEnter={() => setTimeout(() => {}, 0)}
      onMouseLeave={() => setTimeout(() => {}, 0)}
    >
      <Card
        className={cn(
          'pricing-plan-card h-full relative overflow-visible transition-all duration-300 group cursor-pointer',
          'bg-gradient-to-br from-white/5 to-white/10 backdrop-blur-sm',
          'border border-white/10 hover:border-white/20',
          'hover:shadow-2xl hover:shadow-[#6366f1]/10',
          plan.id === 'professional' &&
            'border-[#6366f1]/50 ring-1 ring-[#6366f1]/20 hover:ring-[#6366f1]/40',
          plan.comingSoon && 'opacity-60'
        )}
      >
        <div className="absolute inset-0 bg-gradient-to-br from-transparent via-white/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />

        {plan.id === 'professional' && !plan.comingSoon && (
          <div className="absolute -top-4 left-1/2 -translate-x-1/2 z-10">
            <motion.div
              initial={{ scale: 0, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              transition={{ delay: 0.5 + index * 0.1 }}
              className="relative"
            >
              <span className="px-4 py-2 rounded-full bg-gradient-to-r from-[#6366f1] to-[#8b5cf6] text-white text-sm font-semibold flex items-center gap-2 shadow-lg shadow-[#6366f1]/25">
                <Star className="w-4 h-4 fill-current animate-pulse" />
                Most Popular
              </span>
              <div className="absolute inset-0 bg-gradient-to-r from-[#6366f1] to-[#8b5cf6] rounded-full blur-lg opacity-50 -z-10" />
            </motion.div>
          </div>
        )}

        <CardContent className="p-8 relative z-10">
          <div className="mb-8">
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 + index * 0.1 }}
            >
              <h3 className="text-2xl font-bold text-white mb-3 group-hover:text-[#6366f1] transition-colors">
                {plan.name}
              </h3>
              <p className="text-text-secondary text-base mb-6 leading-relaxed">
                {plan.description}
              </p>
            </motion.div>

            <motion.div
              initial={{ opacity: 0, scale: 0.8 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ delay: 0.3 + index * 0.1 }}
              className="mb-4"
            >
              <div className="flex items-baseline gap-2">
                {plan.comingSoon ? (
                  <span className="text-3xl md:text-4xl font-bold bg-gradient-to-r from-white to-text-secondary bg-clip-text text-transparent">
                    Coming Soon
                  </span>
                ) : (
                  <>
                    <span className="text-5xl md:text-6xl font-bold bg-gradient-to-r from-white to-text-secondary bg-clip-text text-transparent">
                      {typeof displayPrice === 'string' ? displayPrice : `$${displayPrice}`}
                    </span>
                    {typeof displayPrice === 'number' && (
                      <span className="text-text-secondary text-lg">/month</span>
                    )}
                  </>
                )}
              </div>
              {billingCycle === 'annual' && plan.id !== 'free' && plan.id !== 'enterprise' && annualPrice > 0 && !plan.comingSoon && (
                <div className="flex items-center gap-2 mt-2">
                  <div className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
                  <p className="text-emerald-400 text-sm font-medium">
                    ${annualPrice.toFixed(0)}/year (save {Math.round((1 - annualPrice / ((plan.price as number) * 12)) * 100)}%)
                  </p>
                </div>
              )}
              {billingCycle === 'monthly' && plan.id !== 'free' && plan.id !== 'enterprise' && (
                <div className="flex items-center gap-2 mt-2">
                  <div className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
                  <p className="text-green-400 text-sm font-medium">
                    Billed monthly, cancel anytime
                  </p>
                </div>
              )}
              {billingCycle === 'annual' && plan.id !== 'free' && plan.id !== 'enterprise' && (
                <p className="text-text-muted text-xs mt-1">
                  Or ${monthlyEquivalentAnnual}/mo if paid monthly
                </p>
              )}
            </motion.div>
          </div>

          <motion.ul
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.4 + index * 0.1 }}
            className="space-y-4 mb-8"
          >
            {plan.features.map((feature, featureIndex) => {
              const tooltipId = `${plan.id}-feature-${featureIndex}`;
              return (
                <motion.li
                  key={feature}
                  initial={{ opacity: 0, x: -10 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: 0.5 + index * 0.1 + featureIndex * 0.05 }}
                  className="flex items-start gap-4 group"
                  data-tooltip-id={tooltipId}
                  data-tooltip-content={FEATURE_TOOLTIPS[feature] ?? ''}
                >
                  <div className="w-6 h-6 rounded-full bg-gradient-to-br from-emerald-500/20 to-green-500/20 border border-emerald-500/30 flex items-center justify-center mt-0.5 group-hover:scale-110 transition-transform duration-200">
                    <Check className="w-3.5 h-3.5 text-emerald-400" />
                  </div>
                  <span className="text-text-secondary group-hover:text-white transition-colors cursor-help text-base leading-relaxed">
                    {feature}
                  </span>
                </motion.li>
              );
            })}
          </motion.ul>

          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.6 + index * 0.1 }}
          >
            {plan.comingSoon ? (
              <Button
                variant="outline"
                size="lg"
                disabled={true}
                className="w-full py-4 text-base font-semibold transition-all duration-300 border-2 border-white/20 text-text-muted cursor-not-allowed"
              >
                Coming Soon
              </Button>
            ) : (
              <Link
                to={plan.id === 'enterprise' ? '/contact' : '/signup'}
                className="block group"
                onClick={() => !disabled && !isLoading && onPlanSelect(plan.id, billingCycle === 'annual' && plan.priceIdAnnual ? plan.priceIdAnnual : plan.priceId)}
              >
                <Button
                  variant={plan.id === 'free' ? 'outline' : 'default'}
                  size="lg"
                  disabled={disabled || isLoading}
                  className={cn(
                    'w-full py-4 text-base font-semibold transition-all duration-300 transform hover:scale-105',
                    plan.id === 'free' &&
                      'border-2 border-white/30 hover:border-white/50 hover:bg-white/10',
                    plan.id === 'professional' &&
                      'bg-gradient-to-r from-[#6366f1] to-[#8b5cf6] hover:from-[#6366f1]/90 hover:to-[#8b5cf6]/90 shadow-lg shadow-[#6366f1]/25 hover:shadow-[#6366f1]/40',
                    plan.id !== 'free' &&
                      plan.id !== 'professional' &&
                      'bg-gradient-to-r from-white/10 to-white/5 hover:from-white/20 hover:to-white/10 border border-white/20'
                  )}
                >
                  {isLoading ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Starting Checkout...
                    </>
                  ) : plan.id === 'enterprise' ? (
                    'Contact Sales'
                  ) : plan.id === 'free' ? (
                    'Start Free'
                  ) : (
                    'Start Free Trial'
                  )}
                </Button>
              </Link>
            )}
          </motion.div>
        </CardContent>
      </Card>
    </motion.div>
  );
}
