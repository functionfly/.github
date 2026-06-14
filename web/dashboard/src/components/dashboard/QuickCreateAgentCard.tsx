import { Card, CardContent } from '@/components/ui/card';
import { cn } from '@/lib/utils';
import { Bot, Plus, Sparkles } from 'lucide-react';

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
        'cursor-pointer transition-all duration-200',
        'border border-border-subtle bg-card hover:border-border-default hover:shadow-sm',
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
            'flex h-12 w-12 shrink-0 items-center justify-center rounded-xl border',
            isLocked
              ? 'bg-secondary border-border-subtle text-text-secondary'
              : 'bg-brand-500/10 border-brand-500/20 text-brand-500'
          )}
        >
          {isLocked ? <Sparkles className="h-5 w-5" /> : <Bot className="h-5 w-5" />}
        </div>
        <div className="min-w-0 flex-1">
          <h3 className="font-medium text-text-primary">
            {isLocked ? 'Upgrade to create agents' : title}
          </h3>
          {description && (
            <p className="mt-0.5 text-sm text-text-secondary">
              {isLocked ? 'Starter plans and above can create and deploy agents.' : description}
            </p>
          )}
        </div>
        <div
          className={cn(
            'flex shrink-0 items-center gap-1.5 text-sm font-medium',
            isLocked ? 'text-text-secondary' : 'text-brand-500'
          )}
        >
          {isLocked ? (
            <>
              <span>View plans</span>
              <Plus className="h-4 w-4" />
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
