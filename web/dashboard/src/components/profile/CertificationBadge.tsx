import { motion, useMotionValue, useSpring, useTransform } from 'framer-motion';
import { Award, Shield, Crown, Sparkles } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import type { PublicBadge } from '@/api/certification';

const tierConfig: Record<
  string,
  { icon: typeof Award; gradient: string; glow: string; border: string; label: string }
> = {
  associate: {
    icon: Award,
    gradient: 'from-sky-400 via-blue-500 to-indigo-500',
    glow: 'shadow-blue-500/60',
    border: 'border-blue-400/60',
    label: 'Associate',
  },
  professional: {
    icon: Shield,
    gradient: 'from-violet-500 via-purple-500 to-fuchsia-500',
    glow: 'shadow-purple-500/60',
    border: 'border-purple-400/60',
    label: 'Professional',
  },
  architect: {
    icon: Crown,
    gradient: 'from-amber-400 via-orange-500 to-yellow-500',
    glow: 'shadow-amber-500/60',
    border: 'border-amber-400/60',
    label: 'Architect',
  },
};

const defaultTier = {
  icon: Award,
  gradient: 'from-slate-400 via-slate-500 to-zinc-500',
  glow: 'shadow-slate-500/60',
  border: 'border-slate-400/60',
  label: 'Certified',
};

function CertificationBadgeInner({ badge, size = 'md' }: { badge: PublicBadge; size?: 'sm' | 'md' | 'lg' }) {
  const config = tierConfig[badge.tier_slug] || defaultTier;
  const Icon = config.icon;

  const mouseX = useMotionValue(0.5);
  const mouseY = useMotionValue(0.5);
  const springX = useSpring(mouseX, { stiffness: 220, damping: 18 });
  const springY = useSpring(mouseY, { stiffness: 220, damping: 18 });

  const rotateX = useTransform(springY, [0, 1], [14, -14]);
  const rotateY = useTransform(springX, [0, 1], [-14, 14]);
  const sheenOpacity = useTransform(springX, [0, 0.5, 1], [0.08, 0.38, 0.08]);
  const sheenX = useTransform(springX, [0, 1], [-60, 60]);

  const sizeClasses = {
    sm: { container: 'h-14 w-14', icon: 'h-6 w-6', inner: 'p-2', text: 'text-[10px]' },
    md: { container: 'h-20 w-20', icon: 'h-8 w-8', inner: 'p-2.5', text: 'text-xs' },
    lg: { container: 'h-28 w-28', icon: 'h-11 w-11', inner: 'p-3', text: 'text-sm' },
  }[size];

  const issuedAt = badge.issued_at ? new Date(badge.issued_at) : null;
  const expiresAt = badge.expires_at ? new Date(badge.expires_at) : null;
  const now = new Date();
  const isExpired = expiresAt ? expiresAt < now : false;
  const isExpiringSoon =
    expiresAt &&
    !isExpired &&
    (expiresAt.getTime() - now.getTime()) / (1000 * 60 * 60 * 24) < 90;
  const remainingText = isExpired
    ? 'Expired'
    : expiresAt
      ? `Expires ${expiresAt.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}`
      : 'Active';

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <motion.button
            whileHover={{ scale: 1.08, y: -4 }}
            whileTap={{ scale: 0.94 }}
            className="relative inline-flex flex-col items-center group"
            onMouseMove={(e) => {
              const rect = e.currentTarget.getBoundingClientRect();
              mouseX.set((e.clientX - rect.left) / rect.width);
              mouseY.set((e.clientY - rect.top) / rect.height);
            }}
          >
            <motion.div
              style={{ rotateX, rotateY, transformStyle: 'preserve-3d' }}
              className={cn(
                'relative flex items-center justify-center rounded-[28%_72%_28%_72%/_52%_28%_72%_48%]',
                'border bg-gradient-to-br shadow-xl transition-shadow duration-500',
                'group-hover:shadow-2xl',
                config.gradient,
                config.glow,
                sizeClasses.container,
                isExpired && 'opacity-60',
              )}
            >
              <motion.div
                style={{ x: sheenX, opacity: sheenOpacity }}
                className="absolute inset-0 rounded-[inherit] bg-gradient-to-b from-white/60 via-white/10 to-transparent mix-blend-overlay pointer-events-none"
              />

              <motion.div
                className={cn('relative z-10 flex flex-col items-center justify-center gap-0.5', sizeClasses.inner)}
                animate={{
                  y: [0, -1.8, 0],
                }}
                transition={{
                  duration: 2.4,
                  ease: 'easeInOut',
                  repeat: Infinity,
                  repeatType: 'mirror',
                }}
              >
                <Sparkles className="absolute -top-2 -right-2 h-3 w-3 text-white/70" strokeWidth={2.5} />
                <Icon className={cn('text-white drop-shadow-md', sizeClasses.icon)} />
                <span className={cn('font-bold text-white drop-shadow-md leading-none mt-0.5', sizeClasses.text)}>
                  {config.label}
                </span>
              </motion.div>
            </motion.div>

            <div className="mt-2 flex flex-col items-center gap-0.5">
              <span className="font-mono text-[10px] text-text-muted">{badge.credential_number}</span>
              <span
                className={cn(
                  sizeClasses.text,
                  'font-medium',
                  isExpired
                    ? 'text-red-400'
                    : isExpiringSoon
                      ? 'text-amber-400'
                      : 'text-emerald-400',
                )}
              >
                {remainingText}
              </span>
            </div>
          </motion.button>
        </TooltipTrigger>

        <TooltipContent side="right" className="min-w-[220px] border border-border-subtle bg-card/95 p-4">
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <Icon className={cn('h-4 w-4', size === 'lg' ? 'h-5 w-5' : '')} />
              <p className="text-sm font-semibold">{badge.tier_name}</p>
            </div>
            <div className="space-y-1">
              <div className="flex items-center justify-between text-xs">
                <span className="text-text-muted">Credential</span>
                <span className="font-mono">{badge.credential_number}</span>
              </div>
              <div className="flex items-center justify-between text-xs">
                <span className="text-text-muted">Issued</span>
                <span>{issuedAt?.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}</span>
              </div>
              <div className="flex items-center justify-between text-xs">
                <span className="text-text-muted">Expires</span>
                <span>{expiresAt?.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}</span>
              </div>
            </div>
            <div className="h-px w-full bg-border-subtle" />
            <a
              href={`/verify/${badge.credential_number}`}
              className="block text-center text-xs font-medium text-brand-400 hover:text-brand-300"
            >
              Verify credential
            </a>
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

export { CertificationBadgeInner };
