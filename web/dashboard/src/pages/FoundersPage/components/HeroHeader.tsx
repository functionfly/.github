import { TrustSeal } from '@/components/ui/TrustSeal';
import type { FounderRank, FounderStatus } from '../hooks/useFounderConsole';

interface HeroHeaderProps {
  status: FounderStatus;
  rank: FounderRank | null;
}

export function HeroHeader({ status, rank }: HeroHeaderProps) {
  const founderNumber = status.founder_number;
  const percentile = rank?.percentile ?? 0;

  return (
    <div className="founders-hero">
      <div className="founders-hero__gradient" />
      <div className="founders-hero__content">
        <div className="founders-hero__row">
          <h1 className="founders-hero__title">Founders Chamber</h1>
          <TrustSeal size="lg" />
        </div>
        {founderNumber && (
          <div className="founders-hero__number">
            <span className="founders-hero__hash">#</span>
            {founderNumber.toLocaleString()}
          </div>
        )}
        {rank && (
          <p className="founders-hero__rank">
            Top {percentile.toFixed(0)}% of {status.total_founders.toLocaleString()} founders
          </p>
        )}
        <p className="founders-hero__subtitle">
          Your status is permanent and can never be revoked. Welcome to the inner circle.
        </p>
      </div>
    </div>
  );
}
