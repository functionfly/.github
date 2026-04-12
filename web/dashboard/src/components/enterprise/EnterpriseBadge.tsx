import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { usePlan } from '@/hooks/usePlan';
import { motion } from 'framer-motion';
import { Crown, Shield } from 'lucide-react';

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
                       bg-aviation-amber/10
                       border border-aviation-amber/30 cursor-pointer
                       hover:border-aviation-amber/50 hover:bg-aviation-amber/20 transition-colors"
          >
            <Crown className="w-3.5 h-3.5 text-aviation-amber" />
            <span className="text-xs font-semibold text-aviation-amber hidden sm:inline">
              Enterprise
            </span>
          </motion.div>
        </TooltipTrigger>
        <TooltipContent
          side="bottom"
          className="bg-aviation-bg-secondary border-aviation-border-panel"
        >
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <Shield className="w-4 h-4 text-aviation-amber" />
              <p className="font-semibold text-aviation-amber">Enterprise Plan</p>
            </div>
            <p className="text-xs text-aviation-text-secondary">
              99.99% SLA • Dedicated Support • Unlimited Everything
            </p>
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
