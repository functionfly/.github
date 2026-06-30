import { usePageTitle } from '@/hooks';
import { AlertCircle } from 'lucide-react';
import { Chamber } from '@/components/ui/Chamber';
import { TrustSeal } from '@/components/ui/TrustSeal';
import { useFounderConsole } from './hooks/useFounderConsole';
import { usePlatformStream } from './hooks/usePlatformStream';
import { useKeyboardNav } from './hooks/useKeyboardNav';
import { HeroHeader } from './components/HeroHeader';
import { ActivityTicker } from './components/ActivityTicker';
import { GovernanceVotes } from './components/GovernanceVotes';
import { PlatformState } from './components/PlatformState';
import { FounderIdentity } from './components/FounderIdentity';
import { EarningsReferrals } from './components/EarningsReferrals';
import { EarlyAccess } from './components/EarlyAccess';
import { CommunityStats } from './components/CommunityStats';

import '@/styles/sc-founders.css';

export default function FoundersPage() {
  usePageTitle('Founders');
  useKeyboardNav();

  const {
    status,
    rank,
    votes,
    features,
    referralCode,
    referralStats,
    commissions,
    leaderboard,
    loading,
    error,
    castVote,
    claimFeature,
  } = useFounderConsole();

  const platformMetrics = usePlatformStream();

  if (loading) {
    return (
      <div className="founders-page">
        <div className="founders-loading">
          <div className="founders-loading__spinner" />
        </div>
      </div>
    );
  }

  if (error || !status) {
    return (
      <div className="founders-page">
        <div className="founders-error">
          <div className="founders-error__icon">
            <AlertCircle size={24} />
          </div>
          <p className="founders-error__message">{error || 'Failed to load founders data'}</p>
        </div>
      </div>
    );
  }

  if (!status.is_founder) {
    return (
      <div className="founders-page">
        <Chamber size="lg" corners={['tl', 'br']} className="founders-empty">
          <TrustSeal size="lg" />
          <h1 className="founders-empty__title">Founders Program</h1>
          <p className="founders-empty__description">
            The first 10,000 FunctionFly members received permanent founder status.
          </p>
          <p className="text-mono text-[10px] uppercase tracking-widest text-[var(--text-faint)] mt-2">
            {status.total_founders.toLocaleString()} of {status.max_founders.toLocaleString()} slots
            claimed
          </p>
        </Chamber>
      </div>
    );
  }

  return (
    <div className="founders-page">
      <HeroHeader status={status} rank={rank} />
      <ActivityTicker status={status} votes={votes} />

      <GovernanceVotes
        votes={votes}
        castVote={castVote}
        totalFounders={status.total_founders}
      />

      <PlatformState metrics={platformMetrics} />

      <FounderIdentity status={status} rank={rank} />

      <EarningsReferrals
        referralCode={referralCode}
        referralStats={referralStats}
        commissions={commissions}
      />

      <EarlyAccess features={features} claimFeature={claimFeature} />

      <CommunityStats status={status} leaderboard={leaderboard} />

      <div className="founders-page__footer">
        <p>
          <span className="text-[var(--status-ok)]">{status.total_founders.toLocaleString()}</span>
          {status.total_founders === 1 ? ' founder has' : ' founders have'} joined —{' '}
          <span>{(status.max_founders - status.total_founders).toLocaleString()}</span> slots
          remaining
        </p>
      </div>
    </div>
  );
}
