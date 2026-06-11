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
      className={cn('certification-tier-card', tier.color)}
    >
      {/* Corner Accents */}
      <div className="corner-accent top-left" />
      <div className="corner-accent top-right" />
      <div className="corner-accent bottom-left" />
      <div className="corner-accent bottom-right" />

      {/* Gradient accent bar */}
      <div className={cn('absolute top-0 left-0 right-0 h-[3px] bg-gradient-to-r', gradient)} />

      {/* Icon */}
      <div className={cn('certification-tier-icon', gradient.replace('from-', '').replace('to-', ''))}>
        <Icon className="h-7 w-7" />
      </div>

      {/* Title */}
      <h3 className="certification-tier-title">{tier.name}</h3>
      <p className="certification-tier-description">{tier.description}</p>

      {/* Stats */}
      <div className="certification-tier-stats">
        <div className="certification-tier-stat">
          <Clock className="h-4 w-4" />
          <span>{tier.time_limit_minutes} min</span>
        </div>
        <div className="certification-tier-stat">
          <BookOpen className="h-4 w-4" />
          <span>{tier.question_count} questions</span>
        </div>
        <div className="certification-tier-stat">
          <Wrench className="h-4 w-4" />
          <span>{tier.practical_count} challenges</span>
        </div>
        <div className="certification-tier-stat">
          <Award className="h-4 w-4" />
          <span>{tier.validity_months}mo validity</span>
        </div>
      </div>

      {/* Price + Pass threshold */}
      <div className="certification-tier-footer">
        <span className="certification-tier-badge">
          {tier.pass_threshold}% to pass
        </span>
        <div className="certification-tier-price">
          {tier.price_cents === 0 ? (
            <span className="price-free">Free</span>
          ) : (
            <>
              <DollarSign className="h-4 w-4" />
              <span className="price-value">{(tier.price_cents / 100).toFixed(0)}</span>
            </>
          )}
        </div>
      </div>

      {/* CTA — pushed to bottom via mt-auto */}
      <div className="certification-tier-cta">
        {hasActive ? (
          <>
            <button
              onClick={onResume}
              className="btn-primary"
            >
              <Clock className="h-4 w-4" />
              Resume Exam
            </button>
            <button
              onClick={onStart}
              disabled={isLoading}
              className="btn-outline"
            >
              <Shield className="h-4 w-4" />
            </button>
          </>
        ) : hasPending ? (
          <button
            onClick={paymentConfirmed ? onStart : onBuyNow}
            disabled={isLoading}
            className="btn-primary"
          >
            {isLoading ? 'Redirecting...' : paymentConfirmed ? 'Start Exam' : 'Complete Payment'}
          </button>
        ) : tier.slug === 'professional' || tier.slug === 'architect' ? (
          <button
            disabled={true}
            className="btn-soon"
          >
            Coming Soon
          </button>
        ) : (
          <button
            onClick={onBuyNow || onStart}
            disabled={isLoading}
            className="btn-primary"
          >
            {isLoading ? 'Redirecting...' : 'Buy Now'}
          </button>
        )}
      </div>
    </motion.div>
  );
}
