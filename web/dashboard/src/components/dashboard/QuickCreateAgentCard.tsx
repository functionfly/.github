import { Card, CardContent } from '@/components/ui/card';
import { cn } from '@/lib/utils';
import { Bot, Lock, Plus, Sparkles } from 'lucide-react';

export interface QuickCreateAgentCardProps {
  title?: string;
  description?: string;
  /** Callback when card is clicked (e.g. navigate to create flow) */
  onCreateClick?: () => void;
  /** Primary action label */
  actionLabel?: string;
  className?: string;
  /** If true, shows upgrade prompt instead of create action (for free users) */
  isLocked?: boolean;
  /** Callback when locked card is clicked to upgrade */
  onUpgradeClick?: () => void;
}

export function QuickCreateAgentCard({
  title = 'Create agent',
  description = 'Deploy a new agent in seconds.',
  onCreateClick,
  actionLabel = 'Create agent',
  className,
  isLocked = false,
  onUpgradeClick,
}: QuickCreateAgentCardProps) {
  const handleClick = isLocked ? onUpgradeClick : onCreateClick;

  return (
    <Card
      className={cn(
        'border-theme bg-card cursor-pointer transition-all duration-200',
        'border-dashed',
        isLocked
          ? 'border-amber-500/30 bg-amber-500/[0.03] hover:border-amber-500/50 hover:bg-amber-500/[0.05]'
          : 'hover:border-[var(--color-brand-500)]/40 hover:bg-bg-hover/50',
        className
      )}
      onClick={handleClick}
      role={handleClick ? 'button' : undefined}
      tabIndex={handleClick ? 0 : undefined}
      onKeyDown={
        handleClick
          ? (e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                handleClick();
              }
            }
          : undefined
      }
    >
      <CardContent className="flex items-center gap-4 p-5">
        <div
          className={cn(
            'flex h-12 w-12 shrink-0 items-center justify-center rounded-xl',
            isLocked ? 'bg-amber-500/15 text-amber-400' : 'bg-brand-500/15 text-brand-400'
          )}
        >
          {isLocked ? <Lock className="h-6 w-6" /> : <Bot className="h-6 w-6" />}
        </div>
        <div className="min-w-0 flex-1">
          <h3 className="font-semibold text-text-primary">
            {isLocked ? 'Unlock Agent Creation' : title}
          </h3>
          {description && (
            <p className="mt-0.5 text-sm text-text-secondary">
              {isLocked ? 'Upgrade to Starter to create and deploy agents.' : description}
            </p>
          )}
        </div>
        <div
          className={cn(
            'flex shrink-0 items-center gap-1.5 text-sm font-medium',
            isLocked ? 'text-amber-400' : 'text-brand-400'
          )}
        >
          {isLocked ? (
            <>
              <Sparkles className="h-4 w-4" />
              <span>Upgrade</span>
            </>
          ) : (
            <>
              <Plus className="h-4 w-4" />
              <span>{actionLabel}</span>
            </>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
