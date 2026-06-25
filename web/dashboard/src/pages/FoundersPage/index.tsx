import React, { useEffect, useState } from 'react';
import { usePageTitle } from '@/hooks';
import { API_URLS } from '../../lib/api-urls';
import {
  getMyAffiliateCommissions,
  getMyReferralCode,
  getMyReferralStats,
  AffiliateCommission,
  FounderReferralCode,
  FounderReferralStats,
} from '../../api/billing';
import { apiClient } from '../../api/client';
import { Chamber } from '../../components/ui/Chamber';
import { TrustSeal } from '../../components/ui/TrustSeal';
import { GaugeStrip } from '../../components/ui/GaugeStrip';
import { StatusPill, LiveStatus, ClaimedBadge, AvailableBadge } from '../../components/ui/FoundersStatusPill';
import {
  Shield,
  Vote,
  Wallet,
  Zap,
  Copy,
  AlertCircle,
} from 'lucide-react';

import '@/styles/sc-founders.css';

interface FounderStatus {
  is_founder: boolean;
  founder_number: number | null;
  total_founders: number;
  max_founders: number;
  benefits: {
    permanent_badge: boolean;
    voting_rights: boolean;
    lifetime_commissions: boolean;
    early_access: boolean;
  };
}

interface Vote {
  id: string;
  title: string;
  description: string;
  vote_type: string;
  status: string;
  options: { id: string; label: string }[];
  has_voted: boolean;
  my_vote?: string;
}

interface EarlyAccessFeature {
  slug: string;
  name: string;
  description: string;
  is_claimed: boolean;
  launched_at?: string;
}

const benefitIcons = {
  permanent_badge: Shield,
  voting_rights: Vote,
  lifetime_commissions: Wallet,
  early_access: Zap,
};

const benefitLabels = {
  permanent_badge: 'Permanent Badge',
  voting_rights: 'Voting Rights',
  lifetime_commissions: 'Lifetime Commissions',
  early_access: 'Early Access',
};

const benefitDescriptions = {
  permanent_badge: 'Your gold founder badge is displayed permanently on your profile.',
  voting_rights: 'Vote on feature roadmap priorities and help shape FunctionFly\'s future.',
  lifetime_commissions: 'Earn commissions on referrals for as long as they remain active — no expiration.',
  early_access: 'Get early access to new features before they\'re available to everyone.',
};

export default function FoundersPage() {
  usePageTitle('Founders');

  const [status, setStatus] = useState<FounderStatus | null>(null);
  const [votes, setVotes] = useState<Vote[]>([]);
  const [features, setFeatures] = useState<EarlyAccessFeature[]>([]);
  const [referralCode, setReferralCode] = useState<FounderReferralCode | null>(null);
  const [referralStats, setReferralStats] = useState<FounderReferralStats | null>(null);
  const [commissions, setCommissions] = useState<AffiliateCommission[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [
          statusRes,
          votesRes,
          featuresRes,
          referralCodeRes,
          referralStatsRes,
          commissionsRes,
        ] = await Promise.all([
          apiClient.get(API_URLS.founders.status),
          apiClient.get(API_URLS.founders.votes),
          apiClient.get(API_URLS.founders.earlyAccess),
          getMyReferralCode(),
          getMyReferralStats(),
          getMyAffiliateCommissions(),
        ]);

        if (statusRes) {
          setStatus(statusRes);
        }

        if (votesRes) {
          setVotes(votesRes.votes || []);
        }

        if (featuresRes) {
          setFeatures(featuresRes.features || []);
        }

        if (referralCodeRes) {
          setReferralCode(referralCodeRes);
        }
        if (referralStatsRes) {
          setReferralStats(referralStatsRes);
        }
        if (commissionsRes) {
          setCommissions(commissionsRes.commissions || []);
        }
      } catch {
        setError('Failed to load founders data');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  const copyReferralLink = async () => {
    if (referralCode?.share_url) {
      await navigator.clipboard.writeText(referralCode.share_url);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const shareToTwitter = () => {
    if (referralCode?.share_url) {
      const text = encodeURIComponent(
        `I just joined @functionfly as a Founder! Join me and get early access to the future of AI agents. Use my link: ${referralCode.share_url}`
      );
      window.open(`https://twitter.com/intent/tweet?text=${text}`, '_blank');
    }
  };

  const shareToLinkedIn = () => {
    if (referralCode?.share_url) {
      const url = encodeURIComponent(referralCode.share_url);
      window.open(`https://www.linkedin.com/sharing/share-offsite/?url=${url}`, '_blank');
    }
  };

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
          <p className="text-mono text-[10px] uppercase tracking-widest text-[var(--text-muted)] mt-2">
            {status.total_founders.toLocaleString()} of {status.max_founders.toLocaleString()} slots claimed
          </p>
        </Chamber>
      </div>
    );
  }

  const totalEarned = commissions.reduce((sum, c) => sum + c.commission_cents, 0);
  const pendingCommission = commissions
    .filter((c) => c.status === 'pending')
    .reduce((sum, c) => sum + c.commission_cents, 0);
  const paidOut = commissions
    .filter((c) => c.status === 'paid')
    .reduce((sum, c) => sum + c.commission_cents, 0);
  const estimatedAnnual = totalEarned > 0 ? totalEarned * 12 : 0;

  const earningsGauges = [
    { value: `$${(totalEarned / 100).toFixed(2)}`, label: 'Total Earned', variant: 'accent' as const, tick: true },
    { value: `$${(pendingCommission / 100).toFixed(2)}`, label: 'Pending', variant: 'warning' as const },
    { value: `$${(paidOut / 100).toFixed(2)}`, label: 'Paid Out' },
    { value: `$${(estimatedAnnual / 100).toFixed(2)}`, label: 'Est. Annual' },
  ];

  const referralGauges = [
    { value: referralStats?.total_referrals ?? 0, label: 'Total Referrals', tick: true },
    { value: referralStats?.converted_count ?? 0, label: 'Converted', variant: 'accent' as const },
    { value: `${referralStats?.conversion_rate?.toFixed(1) ?? '0.0'}%`, label: 'Conv. Rate' },
    { value: `$${referralStats?.average_commission_cents?.toFixed(2) ?? '0.00'}`, label: 'Avg Commission' },
  ];

  return (
    <div className="founders-page">
      <div className="flex items-center gap-4 mb-2">
        <h1 className="founders-page__title">Founders Program</h1>
        <TrustSeal size="md" />
      </div>
      <p className="founders-page__subtitle">
        Congratulations — you're a permanent FunctionFly founder. Your status can never be revoked.
      </p>

      {/* Benefits Grid */}
      <section className="founders-page__section">
        <div className="founders-page__grid founders-page__grid--2col">
          {Object.entries(status.benefits).map(([key, enabled]) => {
            if (!enabled) return null;
            const Icon = benefitIcons[key as keyof typeof benefitIcons];
            return (
              <Chamber key={key} size="md" corners={['tl', 'br']}>
                <div className="benefit-card">
                  <div className="benefit-card__icon">
                    <Icon size={20} />
                  </div>
                  <h3 className="benefit-card__title">
                    {benefitLabels[key as keyof typeof benefitLabels]}
                  </h3>
                  <p className="benefit-card__description">
                    {benefitDescriptions[key as keyof typeof benefitDescriptions]}
                  </p>
                </div>
              </Chamber>
            );
          })}
        </div>
      </section>

      {/* Earnings */}
      {referralStats && referralCode && (
        <section className="founders-page__section">
          <div className="founders-page__section-title">Earnings</div>
          <GaugeStrip items={earningsGauges} />
        </section>
      )}

      {/* Referral Stats */}
      {referralStats && (
        <section className="founders-page__section">
          <div className="founders-page__section-title">Referral Performance</div>
          <GaugeStrip items={referralGauges} />
        </section>
      )}

      {/* Referral Link */}
      {referralCode && (
        <section className="founders-page__section">
          <div className="founders-page__section-title">Your Referral Link</div>
          <Chamber size="md" corners={['tl', 'br']}>
            <div className="referral-link-display">
              <code className="referral-link-display__code">
                {referralCode.share_url}
              </code>
              <button
                onClick={copyReferralLink}
                className="sealed-button-primary sealed-btn-md flex items-center gap-2"
              >
                <Copy size={14} />
                {copied ? 'Copied' : 'Copy'}
              </button>
            </div>
            <div className="flex items-center gap-3 mt-4">
              <button
                onClick={shareToTwitter}
                className="frame-button-secondary frame-btn-md flex items-center gap-2"
              >
                <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M14.234 10.162 22.977 0h-2.072l-7.591 8.824L7.251 0H.258l9.168 13.343L.258 24H2.33l8.016-9.318L16.749 24h6.993zm-2.837 3.299-.929-1.329L3.076 1.56h3.182l5.965 8.532.929 1.329 7.754 11.09h-3.182z"/>
                </svg>
                Share on X
              </button>
              <button
                onClick={shareToLinkedIn}
                className="frame-button-secondary frame-btn-md flex items-center gap-2"
              >
                <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433c-1.144 0-2.063-.926-2.063-2.065 0-1.138.92-2.063 2.063-2.063 1.14 0 2.064.925 2.064 2.063 0 1.139-.925 2.065-2.064 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z"/>
                </svg>
                Share on LinkedIn
              </button>
            </div>
            <p className="text-[var(--text-muted)] text-[11px] font-mono uppercase tracking-wider mt-4">
              Earn 10% lifetime commission on all referrals who subscribe to FunctionFly.
            </p>
          </Chamber>
        </section>
      )}

      {/* Commission History */}
      {commissions.length > 0 && (
        <section className="founders-page__section">
          <div className="founders-page__section-title">Commission History</div>
          <Chamber size="lg" corners={['tl', 'br']}>
            <div className="overflow-x-auto">
              <table className="founders-table">
                <thead>
                  <tr>
                    <th>Date</th>
                    <th>Referral</th>
                    <th>Plan</th>
                    <th>Commission</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {commissions.slice(0, 10).map((commission) => (
                    <tr key={commission.id}>
                      <td>{new Date(commission.created_at).toLocaleDateString()}</td>
                      <td className="font-mono text-[var(--text-muted)]">
                        {commission.referral_id.slice(0, 8)}...
                      </td>
                      <td>${commission.base_amount_usd.toFixed(2)}</td>
                      <td className="font-mono font-semibold text-[var(--text-primary)]">
                        ${commission.commission_usd.toFixed(2)}
                      </td>
                      <td>
                        <StatusPill
                          status={commission.status === 'paid' ? 'live' : 'pending'}
                          label={commission.status}
                        />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </Chamber>
        </section>
      )}

      {/* Active Votes */}
      {votes.length > 0 && (
        <section className="founders-page__section">
          <div className="founders-page__section-title">Active Votes</div>
          <div className="founders-page__grid">
            {votes.map((vote) => (
              <Chamber key={vote.id} size="md" corners={['tl', 'br']}>
                <div className="vote-card">
                  <div className="vote-card__header">
                    <h3 className="vote-card__title">{vote.title}</h3>
                    <LiveStatus label={vote.status} />
                  </div>
                  {vote.description && (
                    <p className="vote-card__description">{vote.description}</p>
                  )}
                  {vote.has_voted ? (
                    <div className="vote-card__voted">
                      <Vote size={14} />
                      You voted for: {vote.options.find((o) => o.id === vote.my_vote)?.label}
                    </div>
                  ) : (
                    <p className="text-[var(--accent)] text-[12px] font-mono mt-3">
                      Cast your vote to help prioritize
                    </p>
                  )}
                </div>
              </Chamber>
            ))}
          </div>
        </section>
      )}

      {/* Early Access Features */}
      {features.length > 0 && (
        <section className="founders-page__section">
          <div className="founders-page__section-title">Early Access Features</div>
          <div className="founders-page__grid">
            {features.map((feature) => (
              <Chamber key={feature.slug} size="md" corners={['tl', 'br']}>
                <div className="feature-card">
                  <div className="feature-card__info">
                    <h3 className="feature-card__name">{feature.name}</h3>
                    {feature.description && (
                      <p className="feature-card__description">{feature.description}</p>
                    )}
                  </div>
                  {feature.is_claimed ? (
                    <ClaimedBadge />
                  ) : (
                    <AvailableBadge />
                  )}
                </div>
              </Chamber>
            ))}
          </div>
        </section>
      )}

      <div className="founders-page__footer">
        <p>
          <span className="text-[var(--status-ok)]">{status.total_founders.toLocaleString()}</span>
          {status.total_founders === 1 ? ' founder has' : ' founders have'} joined —{' '}
          <span>{(status.max_founders - status.total_founders).toLocaleString()}</span>
          {' '}slots remaining
        </p>
      </div>
    </div>
  );
}
