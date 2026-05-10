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
          'flex items-center gap-3 rounded-lg border p-3',
          isActive ? 'border-theme bg-card' : 'border-theme bg-card opacity-60'
        )}
      >
        <div className={cn('flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br', gradient)}>
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
        'glass-card glow rounded-xl border border-theme bg-card p-6',
        !isActive && 'opacity-70'
      )}
    >
      {/* Header */}
      <div className="flex items-start justify-between mb-4">
        <div className="flex items-center gap-3">
          <div className={cn('flex h-12 w-12 items-center justify-center rounded-xl bg-gradient-to-br', gradient)}>
            <Icon className="h-6 w-6 text-white" />
          </div>
          <div>
            <h3 className="text-lg font-bold text-text-primary">
              {credential.tier?.name || 'Certification'}
            </h3>
            <p className="text-sm text-text-muted font-mono">{credential.credential_number}</p>
          </div>
        </div>
        <Badge variant={isActive ? 'success' : isExpired ? 'secondary' : 'destructive'}>
          {credential.status}
        </Badge>
      </div>

      {/* Details */}
      <div className="grid grid-cols-2 gap-3 mb-4">
        <div className="flex items-center gap-2 text-sm text-text-secondary">
          <Calendar className="h-4 w-4 text-text-muted" />
          <span>Issued: {new Date(credential.issued_at).toLocaleDateString()}</span>
        </div>
        <div className="flex items-center gap-2 text-sm text-text-secondary">
          <Calendar className="h-4 w-4 text-text-muted" />
          <span>Expires: {new Date(credential.expires_at).toLocaleDateString()}</span>
        </div>
      </div>

      {/* Actions */}
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm" onClick={copyNumber}>
          {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
          {copied ? 'Copied' : 'Copy Number'}
        </Button>
        {credential.verification_url && (
          <Button variant="outline" size="sm" asChild>
            <a href={credential.verification_url} target="_blank" rel="noopener noreferrer">
              <ExternalLink className="h-4 w-4" />
              Verify
            </a>
          </Button>
        )}
      </div>
    </motion.div>
  );
}
