import { Award, Shield, Crown, Calendar, ExternalLink, Copy, Check } from 'lucide-react';
import { useState } from 'react';
import {
  Card,
  CornerBrace,
  TrustSeal,
  SealedButton,
  FrameButton,
  StatusPill,
} from '@/components/containment';
import type { CertCredential } from '@/api/certification';

const tierIcons: Record<string, React.ComponentType<{ className?: string }>> = {
  associate: Award,
  professional: Shield,
  architect: Crown,
};

interface CredentialCardProps {
  credential: CertCredential;
  compact?: boolean;
}

const statusMap: Record<string, 'live' | 'pending' | 'revoked'> = {
  active: 'live',
  expired: 'revoked',
  revoked: 'revoked',
};

export function CredentialCard({ credential, compact }: CredentialCardProps) {
  const [copied, setCopied] = useState(false);
  const tierSlug = credential.tier?.slug || 'associate';
  const Icon = tierIcons[tierSlug] || Award;

  const isActive = credential.status === 'active';
  const status = statusMap[credential.status] || 'pending';

  const copyNumber = async () => {
    await navigator.clipboard.writeText(credential.credential_number);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  if (compact) {
    return (
      <div className={`cred-card cred-card--compact ${!isActive ? 'cred-card--inactive' : ''}`}>
        <div className={`cred-card__compact-icon cred-card__compact-icon--${tierSlug}`}>
          <Icon className="cred-card__compact-icon-svg" />
        </div>
        <div className="cred-card__compact-info">
          <p className="cred-card__compact-name">{credential.tier?.name || 'Certification'}</p>
          <p className="cred-card__compact-number">{credential.credential_number}</p>
        </div>
        <StatusPill status={status} label={credential.status} />
      </div>
    );
  }

  return (
    <Card className={`cred-card cred-card--${tierSlug} ${!isActive ? 'cred-card--inactive' : ''}`}>
      <CornerBrace position="tl" />
      <CornerBrace position="br" />

      {/* Header */}
      <div className="cred-card__header">
        <div className="cred-card__info">
          <div className={`cred-card__icon cred-card__icon--${tierSlug}`}>
            <Icon className="cred-card__icon-svg" />
          </div>
          <div>
            <div className="cred-card__title-row">
              <h3 className="cred-card__title">{credential.tier?.name || 'Certification'}</h3>
              <TrustSeal size="sm" />
            </div>
            <p className="cred-card__number">{credential.credential_number}</p>
          </div>
        </div>
        <StatusPill status={status} label={credential.status} />
      </div>

      {/* Details */}
      <div className="cred-card__details">
        <div className="cred-card__detail">
          <Calendar className="cred-card__detail-icon" />
          <span>Issued: {new Date(credential.issued_at).toLocaleDateString()}</span>
        </div>
        <div className="cred-card__detail">
          <Calendar className="cred-card__detail-icon" />
          <span>Expires: {new Date(credential.expires_at).toLocaleDateString()}</span>
        </div>
      </div>

      {/* Actions */}
      <div className="cred-card__actions">
        <FrameButton
          onClick={copyNumber}
          size="sm"
          iconLeft={copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
        >
          {copied ? 'Copied' : 'Copy Number'}
        </FrameButton>
        {credential.verification_url && (
          <a
            href={credential.verification_url}
            target="_blank"
            rel="noopener noreferrer"
            className="cred-card__verify-link"
          >
            <FrameButton size="sm" iconLeft={<ExternalLink className="h-4 w-4" />}>
              Verify
            </FrameButton>
          </a>
        )}
      </div>
    </Card>
  );
}
