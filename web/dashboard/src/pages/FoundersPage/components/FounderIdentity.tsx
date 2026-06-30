import { Award } from 'lucide-react';
import type { FounderRank, FounderStatus } from '../hooks/useFounderConsole';

interface FounderIdentityProps {
  status: FounderStatus;
  rank: FounderRank | null;
}

const TIER_STANDARD = { label: 'STANDARD', color: 'standard', min: 0, next: 3, nextLabel: 'Pro' };
const TIER_PRO = { label: 'PRO', color: 'pro', min: 3, next: 10, nextLabel: 'Elite' };
const TIER_ELITE = { label: 'ELITE', color: 'elite', min: 10, next: null, nextLabel: '' };

function getTier(referralCount: number) {
  if (referralCount >= 10) return TIER_ELITE;
  if (referralCount >= 3) return TIER_PRO;
  return TIER_STANDARD;
}

export function FounderIdentity({ status, rank }: FounderIdentityProps) {
  const tier = getTier(0);
  const founderNumber = status.founder_number;
  const percentile = rank?.percentile ?? 0;

  return (
    <section className="founders-section">
      <div className="founders-section__title">
        <Award size={14} />
        Founder Identity
      </div>

      <div className="founder-identity">
        <div className="founder-identity__main">
          {founderNumber && (
            <div className="founder-identity__number">
              <span className="founder-identity__hash">#</span>
              {founderNumber.toLocaleString()}
            </div>
          )}
          <span className={`founder-identity__tier founder-identity__tier--${tier.color}`}>
            {tier.label}
          </span>
        </div>

        <div className="founder-identity__stats">
          <div className="founder-identity__stat">
            <span className="founder-identity__stat-value">
              {rank?.rank ?? '—'}
            </span>
            <span className="founder-identity__stat-label">Rank</span>
          </div>
          <div className="founder-identity__stat">
            <span className="founder-identity__stat-value">
              {percentile > 0 ? `Top ${percentile.toFixed(0)}%` : '—'}
            </span>
            <span className="founder-identity__stat-label">Percentile</span>
          </div>
          <div className="founder-identity__stat">
            <span className="founder-identity__stat-value">
              {status.total_founders.toLocaleString()}
            </span>
            <span className="founder-identity__stat-label">Total Founders</span>
          </div>
        </div>

        <div className="founder-identity__tier-progress">
          <div className="founder-identity__tier-labels">
            <span className="founder-identity__tier-label founder-identity__tier-label--active">Standard</span>
            <span className={`founder-identity__tier-label ${tier.min >= 3 ? 'founder-identity__tier-label--active' : ''}`}>Pro</span>
            <span className={`founder-identity__tier-label ${tier.min >= 10 ? 'founder-identity__tier-label--active' : ''}`}>Elite</span>
          </div>
          <div className="founder-identity__progress-bar">
            <div className="founder-identity__progress-fill" style={{ width: '5%' }} />
            <div className="founder-identity__progress-marker founder-identity__progress-marker--pro" style={{ left: '30%' }} />
            <div className="founder-identity__progress-marker founder-identity__progress-marker--elite" style={{ left: '100%' }} />
          </div>
          <p className="founder-identity__tier-hint">
            Earn referral commissions to unlock higher tiers
          </p>
        </div>
      </div>
    </section>
  );
}
