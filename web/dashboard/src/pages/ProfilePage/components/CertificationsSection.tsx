import { motion } from 'framer-motion';
import { BadgeCheck, Sparkles, ExternalLink } from 'lucide-react';
import { Card } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { CertificationBadgeInner } from '@/components/profile/CertificationBadge';
import type { PublicBadge } from '@/api/certification';

export interface CertificationsSectionProps {
  badges: PublicBadge[];
  isLoading?: boolean;
  username?: string;
}

export function CertificationsSection({ badges, isLoading, username }: CertificationsSectionProps) {
  const hasBadges = badges && badges.length > 0;

  return (
    <Card className="border-border-subtle overflow-hidden">
      <div className="border-b border-border-subtle/70 bg-gradient-to-r from-slate-900/40 via-slate-900/20 to-transparent px-5 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-brand-500/20 to-brand-600/10 text-brand-400 ring-1 ring-brand-500/20">
              <BadgeCheck className="h-4 w-4" strokeWidth={2.5} />
            </div>
            <div>
              <h3 className="text-sm font-semibold text-text-primary">Certifications</h3>
              {hasBadges && (
                <p className="text-xs text-text-muted">
                  {badges.length} earned credential{badges.length === 1 ? '' : 's'}
                </p>
              )}
            </div>
          </div>
          {hasBadges && badges.length > 4 && (
            <span className="flex items-center gap-1 rounded-full border border-border-subtle bg-background-muted px-2.5 py-1 text-[10px] font-semibold uppercase tracking-wide text-text-muted">
              <Sparkles className="h-3 w-3 text-brand-400" strokeWidth={2.5} />
              Verified
            </span>
          )}
        </div>
      </div>

      <div className="px-5 py-5">
        {isLoading ? (
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4">
            {Array.from({ length: 4 }).map((_, idx) => (
              <div key={idx} className="flex flex-col items-center gap-3">
                <Skeleton className="h-20 w-20 rounded-full" />
                <Skeleton className="h-3 w-24" />
                <Skeleton className="h-2.5 w-16" />
              </div>
            ))}
          </div>
        ) : !hasBadges ? (
          <div className="flex flex-col items-center justify-center py-10 text-center">
            <div className="flex h-14 w-14 items-center justify-center rounded-full bg-muted/60 text-text-muted">
              <BadgeCheck className="h-6 w-6" strokeWidth={1.5} />
            </div>
            <p className="mt-3 text-sm text-text-muted">No certifications yet</p>
            <p className="text-xs text-text-muted/80">
              {username ? (
                <a href="/certification" className="inline-flex items-center gap-1 text-brand-400 hover:text-brand-300">
                  View available exams <ExternalLink className="h-3 w-3" />
                </a>
              ) : (
                'Complete an exam to earn your first certification'
              )}
            </p>
          </div>
        ) : (
          <motion.div
            initial="hidden"
            animate="visible"
            variants={{
              hidden: { opacity: 0 },
              visible: {
                opacity: 1,
                transition: { staggerChildren: 0.12 },
              },
            }}
            className="grid grid-cols-2 gap-6 sm:grid-cols-3 md:grid-cols-4"
          >
            {badges.map((badge, idx) => (
              <motion.div
                key={badge.credential_number + idx}
                variants={{
                  hidden: { opacity: 0, y: 18, scale: 0.92 },
                  visible: {
                    opacity: 1,
                    y: 0,
                    scale: 1,
                    transition: { type: 'spring', stiffness: 180, damping: 18 },
                  },
                }}
                className="flex flex-col items-center"
              >
                <CertificationBadgeInner badge={badge} size="md" />
              </motion.div>
            ))}
          </motion.div>
        )}
      </div>
    </Card>
  );
}
