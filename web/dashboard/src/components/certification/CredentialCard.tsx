import { motion } from 'framer-motion';
import { Award, Shield, Crown, Calendar, ExternalLink, Copy, Check } from 'lucide-react';
import { useState } from 'react';
import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import type { CertCredential } from '@/api/certification';

const tierIcons: Record<string, React.ComponentType<{ className?: string }>> = {
  associate: Award,
  professional: Shield,
  architect: Crown,
};

const tierGradients: Record<string, string> = {
  blue: 'from-blue-500 to-cyan-500',
  purple: 'from-purple-500 to-pink-500',
  gold: 'from-amber-500 to-yellow-500',
};

interface CredentialCardProps {
  credential: CertCredential;
  compact?: boolean;
}

export function CredentialCard({ credential, compact }: CredentialCardProps) {
  const [copied, setCopied] = useState(false);
  const tierSlug = credential.tier?.slug || 'associate';
  const Icon = tierIcons[tierSlug] || Award;
  const gradient = tierGradients[credential.tier?.slug === 'architect' ? 'gold' : credential.tier?.slug === 'professional' ? 'purple' : 'blue'] || tierGradients.blue;

  const isExpired = credential.status === 'expired';
  const isRevoked = credential.status === 'revoked';
  const isActive = credential.status === 'active';

  const copyNumber = async () => {
    await navigator.clipboard.writeText(credential.credential_number);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  if (compact) {
    return (
      <motion.div
        whileHover={{ scale: 1.02 }}
        className={cn(
          'credentials-card compact',
          isActive ? '' : 'inactive'
        )}
      >
        <div className={cn('flex h-10 w-10 items-center justify-center rounded-lg', gradient.replace('from-', 'bg-gradient-to-br from-'))}>
          <Icon className="h-5 w-5 text-white" />
        </div>
        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium text-text-primary truncate">
            {credential.tier?.name || 'Certification'}
          </p>
          <p className="text-xs text-text-muted font-mono">{credential.credential_number}</p>
        </div>
        <Badge variant={isActive ? 'success' : isExpired ? 'secondary' : 'destructive'}>
          {credential.status}
        </Badge>
      </motion.div>
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className={cn(
        'credentials-card',
        !isActive && 'inactive',
        credential.tier?.slug === 'architect' ? 'gold' : credential.tier?.slug === 'professional' ? 'purple' : 'blue'
      )}
    >
      {/* Header */}
      <div className="credentials-card-header">
        <div className="credentials-card-info">
          <div className="credentials-card-icon">
            <Icon className="h-6 w-6" />
          </div>
          <div>
            <h3 className="credentials-card-title">
              {credential.tier?.name || 'Certification'}
            </h3>
            <p className="credentials-card-number">{credential.credential_number}</p>
          </div>
        </div>
        <span className={`credentials-status-badge ${credential.status}`}>
          {credential.status}
        </span>
      </div>

      {/* Details */}
      <div className="credentials-card-details">
        <div className="credentials-card-detail">
          <Calendar className="h-4 w-4" />
          <span>Issued: {new Date(credential.issued_at).toLocaleDateString()}</span>
        </div>
        <div className="credentials-card-detail">
          <Calendar className="h-4 w-4" />
          <span>Expires: {new Date(credential.expires_at).toLocaleDateString()}</span>
        </div>
      </div>

      {/* Actions */}
      <div className="credentials-card-actions">
        <button
          onClick={copyNumber}
          className="btn-outline"
        >
          {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
          {copied ? 'Copied' : 'Copy Number'}
        </button>
        {credential.verification_url && (
          <a
            href={credential.verification_url}
            target="_blank"
            rel="noopener noreferrer"
            className="btn-outline"
          >
            <ExternalLink className="h-4 w-4" />
            Verify
          </a>
        )}
      </div>
    </motion.div>
  );
}
