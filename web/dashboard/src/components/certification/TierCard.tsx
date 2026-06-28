import { Award, Shield, Crown, Clock, BookOpen, Wrench, DollarSign } from 'lucide-react';
import {
  Card,
  CornerBrace,
  TrustSeal,
  SealedButton,
  FrameButton,
} from '@/components/containment';
import type { CertTier } from '@/api/certification';

const tierIcons: Record<string, React.ComponentType<{ className?: string }>> = {
  Award,
  Shield,
  Crown,
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
  const hasActive = !!activeExamId;
  const hasPending = !!pendingExamId;

  return (
    <Card className={`cert-tier-card cert-tier-card--${tier.color}`}>
      <CornerBrace position="tl" />
      <CornerBrace position="br" />

      {/* Tier header */}
      <div className="cert-tier-card__header">
        <div className={`cert-tier-card__icon cert-tier-card__icon--${tier.color}`}>
          <Icon className="cert-tier-card__icon-svg" />
        </div>
        <div className="cert-tier-card__title-group">
          <div className="cert-tier-card__title-row">
            <h3 className="cert-tier-card__title">{tier.name}</h3>
            <TrustSeal size="sm" />
          </div>
          <StatusPillInline tier={tier} />
        </div>
      </div>

      <p className="cert-tier-card__description">{tier.description}</p>

      {/* Stats */}
      <div className="cert-tier-card__stats">
        <div className="cert-tier-card__stat">
          <Clock className="cert-tier-card__stat-icon" />
          <span>{tier.time_limit_minutes} min</span>
        </div>
        <div className="cert-tier-card__stat">
          <BookOpen className="cert-tier-card__stat-icon" />
          <span>{tier.question_count} questions</span>
        </div>
        <div className="cert-tier-card__stat">
          <Wrench className="cert-tier-card__stat-icon" />
          <span>{tier.practical_count} challenges</span>
        </div>
        <div className="cert-tier-card__stat">
          <Award className="cert-tier-card__stat-icon" />
          <span>{tier.validity_months}mo validity</span>
        </div>
      </div>

      {/* Footer: price + pass threshold */}
      <div className="cert-tier-card__footer">
        <span className="cert-tier-card__badge">
          {tier.pass_threshold}% to pass
        </span>
        <div className="cert-tier-card__price">
          {tier.price_cents === 0 ? (
            <span className="cert-tier-card__price-free">Free</span>
          ) : (
            <>
              <DollarSign className="cert-tier-card__price-icon" />
              <span className="cert-tier-card__price-value">{(tier.price_cents / 100).toFixed(0)}</span>
            </>
          )}
        </div>
      </div>

      {/* CTA */}
      <div className="cert-tier-card__cta">
        {hasActive ? (
          <>
            <SealedButton onClick={onResume} iconLeft={<Clock className="h-4 w-4" />}>
              Resume Exam
            </SealedButton>
            <FrameButton onClick={onStart} disabled={isLoading} iconLeft={<Shield className="h-4 w-4" />}>
              Restart
            </FrameButton>
          </>
        ) : hasPending ? (
          <SealedButton
            onClick={paymentConfirmed ? onStart : onBuyNow}
            disabled={isLoading}
            loading={isLoading}
          >
            {paymentConfirmed ? 'Start Exam' : 'Complete Payment'}
          </SealedButton>
        ) : tier.slug === 'professional' || tier.slug === 'architect' ? (
          <SealedButton disabled>Coming Soon</SealedButton>
        ) : (
          <SealedButton
            onClick={onBuyNow || onStart}
            disabled={isLoading}
            loading={isLoading}
          >
            Buy Now
          </SealedButton>
        )}
      </div>
    </Card>
  );
}

function StatusPillInline({ tier }: { tier: CertTier }) {
  if (tier.is_coming_soon) {
    return (
      <span className="cert-tier-card__status cert-tier-card__status--pending">
        <span className="cert-tier-card__status-dot" aria-hidden="true" />
        Coming Soon
      </span>
    );
  }
  return (
    <span className="cert-tier-card__status cert-tier-card__status--live">
      <span className="cert-tier-card__status-dot" aria-hidden="true" />
      Available
    </span>
  );
}
