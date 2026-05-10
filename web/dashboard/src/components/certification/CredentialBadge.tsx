import { motion } from 'framer-motion';
import { Award, Shield, Crown } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import type { PublicBadge } from '@/api/certification';

const tierIcons: Record<string, React.ComponentType<{ className?: string }>> = {
  associate: Award,
  professional: Shield,
  architect: Crown,
};

const tierColors: Record<string, string> = {
  blue: 'from-blue-500 to-cyan-500',
  purple: 'from-purple-500 to-pink-500',
  gold: 'from-amber-500 to-yellow-500',
};

interface CredentialBadgeProps {
  badge: PublicBadge;
  size?: 'sm' | 'md' | 'lg';
}

export function CredentialBadge({ badge, size = 'md' }: CredentialBadgeProps) {
  const Icon = tierIcons[badge.tier_slug] || Award;
  const gradient = tierColors[badge.tier_color] || tierColors.blue;

  const sizeClasses = {
    sm: 'h-8 w-8',
    md: 'h-10 w-10',
    lg: 'h-14 w-14',
  };

  const iconSizes = {
    sm: 'h-4 w-4',
    md: 'h-5 w-5',
    lg: 'h-7 w-7',
  };

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger>
          <motion.div
            whileHover={{ scale: 1.1 }}
            whileTap={{ scale: 0.95 }}
            className={cn(
              'flex items-center justify-center rounded-full bg-gradient-to-br shadow-lg',
              gradient,
              sizeClasses[size]
            )}
          >
            <Icon className={cn('text-white', iconSizes[size])} />
          </motion.div>
        </TooltipTrigger>
        <TooltipContent>
          <div className="text-center">
            <p className="font-medium">{badge.tier_name}</p>
            <p className="text-xs text-text-muted font-mono">{badge.credential_number}</p>
            <p className="text-xs text-text-muted">
              Valid until {new Date(badge.expires_at).toLocaleDateString()}
            </p>
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
