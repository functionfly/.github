import { Crown } from 'lucide-react';
import { motion } from 'framer-motion';
import { usePlan } from '@/hooks/usePlan';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';

/**
 * Enterprise badge component for the navbar
 * Displays a gold crown badge with tooltip showing enterprise benefits
 */
export function EnterpriseBadge() {
  const { isEnterprise } = usePlan();

  if (!isEnterprise) return null;

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <motion.div
            initial={{ opacity: 0, scale: 0.8 }}
            animate={{ opacity: 1, scale: 1 }}
            className="flex items-center gap-1.5 px-2.5 py-1 rounded-full
                       bg-gradient-to-r from-amber-500/20 to-yellow-500/20
                       border border-amber-500/30 cursor-pointer
                       hover:border-amber-500/50 transition-colors"
          >
            <Crown className="w-3.5 h-3.5 text-amber-400" />
            <span className="text-xs font-medium text-amber-400">
              Enterprise
            </span>
          </motion.div>
        </TooltipTrigger>
        <TooltipContent
          side="bottom"
          className="bg-bg-tertiary border-white/10"
        >
          <div className="space-y-1">
            <p className="font-medium text-amber-400">Enterprise Plan</p>
            <p className="text-xs text-text-secondary">
              99.99% SLA • Dedicated Support • Unlimited Everything
            </p>
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
