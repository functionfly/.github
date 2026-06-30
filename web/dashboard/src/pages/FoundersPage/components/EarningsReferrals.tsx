import { Copy, DollarSign, Share2 } from 'lucide-react';
import { useState } from 'react';
import type { AffiliateCommission } from '@/api/billing';
import type { FounderReferralCode, FounderReferralStats } from '@/api/billing';

interface EarningsReferralsProps {
  referralCode: FounderReferralCode | null;
  referralStats: FounderReferralStats | null;
  commissions: AffiliateCommission[];
}

export function EarningsReferrals({
  referralCode,
  referralStats,
  commissions,
}: EarningsReferralsProps) {
  const [copied, setCopied] = useState(false);

  const totalEarned = commissions.reduce((sum, c) => sum + c.commission_cents, 0);
  const pendingCommission = commissions
    .filter((c) => c.status === 'pending')
    .reduce((sum, c) => sum + c.commission_cents, 0);
  const paidOut = commissions
    .filter((c) => c.status === 'paid')
    .reduce((sum, c) => sum + c.commission_cents, 0);
  const estimatedAnnual = totalEarned > 0 ? totalEarned * 12 : 0;

  const copyLink = async () => {
    if (referralCode?.share_url) {
      await navigator.clipboard.writeText(referralCode.share_url);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const shareToTwitter = () => {
    if (referralCode?.share_url) {
      const text = encodeURIComponent(
        `I'm a @functionfly Founder! Join me and get early access to the future of AI agents. ${referralCode.share_url}`
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

  return (
    <section className="founders-section">
      <div className="founders-section__title">
        <DollarSign size={14} />
        Earnings & Referrals
      </div>

      {referralStats && (
        <div className="gauge-strip">
          <div className="gauge-strip__item">
            <div className="gauge-strip__tick" />
            <div className="gauge-strip__value gauge-strip__value--accent">
              ${(totalEarned / 100).toFixed(2)}
            </div>
            <div className="gauge-strip__label">Total Earned</div>
          </div>
          <div className="gauge-strip__item">
            <div className="gauge-strip__value gauge-strip__value--warning">
              ${(pendingCommission / 100).toFixed(2)}
            </div>
            <div className="gauge-strip__label">Pending</div>
          </div>
          <div className="gauge-strip__item">
            <div className="gauge-strip__value">${(paidOut / 100).toFixed(2)}</div>
            <div className="gauge-strip__label">Paid Out</div>
          </div>
          <div className="gauge-strip__item">
            <div className="gauge-strip__value">${(estimatedAnnual / 100).toFixed(2)}</div>
            <div className="gauge-strip__label">Est. Annual</div>
          </div>
        </div>
      )}

      {referralStats && (
        <div className="gauge-strip" style={{ marginTop: '1rem' }}>
          <div className="gauge-strip__item">
            <div className="gauge-strip__tick" />
            <div className="gauge-strip__value">{referralStats.total_referrals ?? 0}</div>
            <div className="gauge-strip__label">Total Referrals</div>
          </div>
          <div className="gauge-strip__item">
            <div className="gauge-strip__value gauge-strip__value--accent">
              {referralStats.converted_count ?? 0}
            </div>
            <div className="gauge-strip__label">Converted</div>
          </div>
          <div className="gauge-strip__item">
            <div className="gauge-strip__value">
              {referralStats.conversion_rate?.toFixed(1) ?? '0.0'}%
            </div>
            <div className="gauge-strip__label">Conv. Rate</div>
          </div>
          <div className="gauge-strip__item">
            <div className="gauge-strip__value">
              ${(referralStats.average_commission_cents ?? 0).toFixed(2)}
            </div>
            <div className="gauge-strip__label">Avg Commission</div>
          </div>
        </div>
      )}

      {referralCode && (
        <div className="founders-chamber founders-chamber--medium" style={{ marginTop: '1rem' }}>
          <div className="referral-link-display">
            <code className="referral-link-display__code">{referralCode.share_url}</code>
            <button
              onClick={copyLink}
              className="sealed-button sealed-button--md"
              style={{ display: 'flex', alignItems: 'center', gap: 6 }}
            >
              <Copy size={14} />
              {copied ? 'Copied' : 'Copy'}
            </button>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginTop: 12 }}>
            <button onClick={shareToTwitter} className="frame-button frame-button--md" style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
              <Share2 size={14} />
              Share on X
            </button>
            <button onClick={shareToLinkedIn} className="frame-button frame-button--md" style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
              <Share2 size={14} />
              Share on LinkedIn
            </button>
          </div>
          <p style={{ color: 'var(--text-faint)', fontFamily: 'var(--font-mono)', fontSize: 11, marginTop: 12, letterSpacing: '0.04em', textTransform: 'uppercase' }}>
            Earn 10% lifetime commission on all referrals who subscribe.
          </p>
        </div>
      )}

      {commissions.length > 0 && (
        <div className="founders-chamber founders-chamber--large" style={{ marginTop: '1rem', overflow: 'auto' }}>
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
              {commissions.slice(0, 10).map((c) => (
                <tr key={c.id}>
                  <td>{new Date(c.created_at).toLocaleDateString()}</td>
                  <td style={{ fontFamily: 'var(--font-mono)', color: 'var(--text-faint)' }}>
                    {c.referral_id.slice(0, 8)}...
                  </td>
                  <td>${c.base_amount_usd.toFixed(2)}</td>
                  <td style={{ fontFamily: 'var(--font-mono)', fontWeight: 600, color: 'var(--text)' }}>
                    ${c.commission_usd.toFixed(2)}
                  </td>
                  <td>
                    <span className={`status-pill status-pill--${c.status === 'paid' ? 'live' : 'pending'}`}>
                      <span className="status-pill__dot" />
                      {c.status}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}
