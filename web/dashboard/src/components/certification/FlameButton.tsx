import { motion } from 'framer-motion';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';

interface FlameButtonProps extends React.ComponentProps<typeof Button> {
  gradient?: string;
}

/**
 * Button with an animated flame/fire effect on hover.
 * Uses the FunctionFly brand orange (#f97316) with layered CSS animations.
 */
export function FlameButton({ gradient, className, children, ...props }: FlameButtonProps) {
  return (
    <div className="flame-button-wrapper group relative">
      {/* Flame particles — only visible on hover */}
      <div className="flame-particles">
        <div className="flame-particle flame-particle-1" />
        <div className="flame-particle flame-particle-2" />
        <div className="flame-particle flame-particle-3" />
        <div className="flame-particle flame-particle-4" />
        <div className="flame-particle flame-particle-5" />
      </div>

      {/* Fire glow backdrop */}
      <div className="flame-glow" />

      {/* The actual button */}
      <Button
        className={cn(
          'flame-button relative z-10 overflow-hidden',
          'transition-all duration-300',
          gradient && `bg-gradient-to-r ${gradient}`,
          !gradient && 'bg-gradient-to-r from-brand-500 via-orange-500 to-red-500',
          'text-white shadow-lg',
          'group-hover:shadow-brand-500/40 group-hover:shadow-xl',
          'group-hover:scale-[1.03]',
          className
        )}
        {...props}
      >
        {/* Shimmer overlay that sweeps across on hover */}
        <span className="flame-shimmer" />
        <span className="relative z-10 flex items-center gap-2">{children}</span>
      </Button>
    </div>
  );
}
