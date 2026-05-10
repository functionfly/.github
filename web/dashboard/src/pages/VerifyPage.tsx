import { useParams } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Award, Shield, Crown, CheckCircle2, XCircle, Loader2, ExternalLink } from 'lucide-react';
import { CredentialBadge } from '@/components/certification/CredentialBadge';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { useVerifyCredential } from '@/hooks/useCertification';

const tierIcons: Record<string, React.ComponentType<{ className?: string }>> = {
  associate: Award,
  professional: Shield,
  architect: Crown,
};

export function VerifyPage() {
  const { username } = useParams<{ username: string }>();
  const { data, isLoading, error } = useVerifyCredential(username || '');

  if (!username) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-bg-primary">
        <p className="text-text-muted">Username is required.</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-bg-primary">
      {/* Header */}
      <header className="border-b border-theme bg-card">
        <div className="max-w-4xl mx-auto px-4 py-4 flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-r from-brand-500 to-purple-500">
            <Award className="h-5 w-5 text-white" />
          </div>
          <div>
            <h1 className="text-lg font-bold text-text-primary">FunctionFly Certification</h1>
            <p className="text-xs text-text-muted">Credential Verification</p>
          </div>
        </div>
      </header>

      {/* Content */}
      <main className="max-w-4xl mx-auto px-4 py-12">
        {isLoading ? (
          <div className="flex items-center justify-center py-24">
            <Loader2 className="h-8 w-8 animate-spin text-brand-500" />
          </div>
        ) : error ? (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="text-center py-24"
          >
            <XCircle className="h-16 w-16 text-red-500 mx-auto mb-4" />
            <h2 className="text-2xl font-bold text-text-primary mb-2">User Not Found</h2>
            <p className="text-text-muted">No certified developer found with username "{username}".</p>
          </motion.div>
        ) : data ? (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
          >
            {/* User info */}
            <div className="text-center mb-8">
              <h2 className="text-3xl font-bold text-text-primary mb-2">
                {data.user.name || data.user.username || username}
              </h2>
              <p className="text-text-muted">@{data.user.username || username}</p>
            </div>

            {/* Credentials */}
            {data.credentials.length === 0 ? (
              <div className="text-center py-12">
                <Award className="h-12 w-12 text-text-muted mx-auto mb-4" />
                <p className="text-text-muted">No active certifications.</p>
              </div>
            ) : (
              <div className="space-y-4">
                <h3 className="text-lg font-bold text-text-primary text-center mb-6">
                  Verified Certifications ({data.credentials.length})
                </h3>

                {data.credentials.map((cred) => {
                  const tierSlug = cred.tier?.slug || 'associate';
                  const Icon = tierIcons[tierSlug] || Award;

                  return (
                    <motion.div
                      key={cred.credential_number}
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      className="glass-card rounded-xl border border-theme bg-card p-6 flex items-center gap-4"
                    >
                      <CredentialBadge
                        badge={{
                          tier_slug: tierSlug,
                          tier_name: cred.tier?.name || 'Certification',
                          tier_color: tierSlug === 'architect' ? 'gold' : tierSlug === 'professional' ? 'purple' : 'blue',
                          tier_icon: '',
                          credential_number: cred.credential_number,
                          issued_at: cred.issued_at,
                          expires_at: cred.expires_at,
                        }}
                        size="lg"
                      />
                      <div className="flex-1">
                        <h4 className="text-lg font-bold text-text-primary">
                          {cred.tier?.name || 'Certification'}
                        </h4>
                        <p className="text-sm text-text-muted font-mono">
                          {cred.credential_number}
                        </p>
                        <div className="flex items-center gap-3 mt-2">
                          <Badge variant="success">
                            <CheckCircle2 className="h-3 w-3 mr-1" />
                            Verified
                          </Badge>
                          <span className="text-xs text-text-muted">
                            Issued {new Date(cred.issued_at).toLocaleDateString()} ·
                            Expires {new Date(cred.expires_at).toLocaleDateString()}
                          </span>
                        </div>
                      </div>
                    </motion.div>
                  );
                })}
              </div>
            )}

            {/* Verification footer */}
            <div className="mt-12 text-center text-sm text-text-muted">
              <p>
                Verified at {new Date().toLocaleString()} ·
                Credentials are cryptographically signed and tamper-proof.
              </p>
            </div>
          </motion.div>
        ) : null}
      </main>
    </div>
  );
}
