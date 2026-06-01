import { motion } from 'framer-motion';
import { Award, Shield, Crown, Clock, BookOpen, Wrench, DollarSign } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import type { CertTier } from '@/api/certification';

const tierIcons: Record<string, React.ComponentType<{ className?: string }>> = {
  Award,
  Shield,
  Crown,
};

const tierGradients: Record<string, string> = {
  blue: 'from-blue-500 to-cyan-500',
  purple: 'from-purple-500 to-pink-500',
  gold: 'from-amber-500 to-yellow-500',
};

interface TierCardProps {
  tier: CertTier;
  onStart?: () => void;
  isLoading?: boolean;
  activeExamId?: string;
  onResume?: () => void;
  pendingExamId?: string;
  onBuyNow?: () => void;
  paymentConfirmed?: boolean;
}

export function TierCard({ tier, onStart, isLoading, activeExamId, onResume, pendingExamId, onBuyNow, paymentConfirmed }: TierCardProps) {
  const Icon = tierIcons[tier.icon] || Award;
  const gradient = tierGradients[tier.color] || tierGradients.blue;
  const hasActive = !!activeExamId;
  const hasPending = !!pendingExamId;

  // DEBUG: Log props received by TierCard
  console.log('[TierCard]', tier.slug, { activeExamId, pendingExamId, hasActive, hasPending, paymentConfirmed });

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      whileHover={{ y: -4, scale: 1.02 }}
      transition={{ duration: 0.3 }}
      className="glass-card glow hover-lift relative flex flex-col overflow-hidden rounded-xl border border-theme bg-card p-6"
    >
      {/* Gradient accent bar */}
      <div className={cn('absolute top-0 left-0 right-0 h-1 bg-gradient-to-r', gradient)} />

      {/* Icon */}
      <div className={cn('mb-4 flex h-14 w-14 items-center justify-center rounded-xl bg-gradient-to-br', gradient)}>
        <Icon className="h-7 w-7 text-white" />
      </div>

      {/* Title */}
      <h3 className="mb-2 text-xl font-bold text-text-primary">{tier.name}</h3>
      <p className="mb-4 text-sm text-text-muted line-clamp-3 min-h-[3.75rem]">{tier.description}</p>

      {/* Stats */}
      <div className="mb-4 grid grid-cols-2 gap-3">
        <div className="flex items-center gap-2 text-sm text-text-secondary">
          <Clock className="h-4 w-4 text-text-muted" />
          <span>{tier.time_limit_minutes} min</span>
        </div>
        <div className="flex items-center gap-2 text-sm text-text-secondary">
          <BookOpen className="h-4 w-4 text-text-muted" />
          <span>{tier.question_count} questions</span>
        </div>
        <div className="flex items-center gap-2 text-sm text-text-secondary">
          <Wrench className="h-4 w-4 text-text-muted" />
          <span>{tier.practical_count} challenges</span>
        </div>
        <div className="flex items-center gap-2 text-sm text-text-secondary">
          <Award className="h-4 w-4 text-text-muted" />
          <span>{tier.validity_months}mo validity</span>
        </div>
      </div>

      {/* Price + Pass threshold */}
      <div className="mb-4 flex items-center justify-between">
        <Badge variant="secondary">
          {tier.pass_threshold}% to pass
        </Badge>
        <div className="flex items-center gap-1 text-lg font-bold text-text-primary">
          {tier.price_cents === 0 ? (
            <span className="text-emerald-500">Free</span>
          ) : (
            <>
              <DollarSign className="h-4 w-4" />
              {(tier.price_cents / 100).toFixed(0)}
            </>
          )}
        </div>
      </div>

      {/* CTA — pushed to bottom via mt-auto */}
      <div className="mt-auto pt-2 flex gap-2">
        {hasActive ? (
          <>
            <Button
              onClick={onResume}
              className={cn(
                'flex-1 bg-gradient-to-r text-white',
                gradient,
                'hover:opacity-90'
              )}
            >
              <Clock className="h-4 w-4" />
              Resume Exam
            </Button>
            <Button
              variant="outline"
              onClick={onStart}
              disabled={isLoading}
              className="px-3"
            >
              <Shield className="h-4 w-4" />
            </Button>
          </>
        ) : hasPending ? (
          <Button
            onClick={paymentConfirmed ? onStart : onBuyNow}
            disabled={isLoading}
            className={cn(
              'w-full bg-gradient-to-r text-white',
              gradient,
              'hover:opacity-90 disabled:opacity-50'
            )}
          >
            {isLoading ? 'Redirecting...' : paymentConfirmed ? 'Start Exam' : 'Complete Payment'}
          </Button>
        ) : tier.slug === 'professional' || tier.slug === 'architect' ? (
          <Button
            disabled={true}
            className={cn(
              'w-full bg-gradient-to-r text-white opacity-60 cursor-not-allowed',
              gradient
            )}
          >
            Coming Soon
          </Button>
        ) : (
          <Button
            onClick={onBuyNow || onStart}
            disabled={isLoading}
            className={cn(
              'w-full bg-gradient-to-r text-white',
              gradient,
              'hover:opacity-90 disabled:opacity-50'
            )}
          >
            {isLoading ? 'Redirecting...' : 'Buy Now'}
          </Button>
        )}
      </div>
    </motion.div>
  );
}
