import * as React from 'react';
import { cn } from '@/lib/utils';

export interface Reaction {
  emoji: string;
  count: number;
  users: string[];
}

export interface ReactionBarProps {
  reactions: Reaction[];
  onReact?: (emoji: string) => void;
  onRemoveReact?: (emoji: string) => void;
  className?: string;
}

export function ReactionBar({
  reactions,
  onReact,
  onRemoveReact,
  className,
}: ReactionBarProps) {
  if (reactions.length === 0) return null;

  return (
    <div className={cn('flex flex-wrap gap-1', className)}>
      {reactions.map((r) => (
        <button
          key={r.emoji}
          onClick={() => {
            // Toggle: if user already reacted, remove; otherwise add
            if (r.users.length > 0 && onRemoveReact) {
              onRemoveReact(r.emoji);
            } else if (onReact) {
              onReact(r.emoji);
            }
          }}
          className={cn(
            'inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs transition-colors',
            'border-border/60 bg-muted/40 hover:bg-muted/70',
            r.count > 0 && 'border-brand-500/30 bg-brand-500/10',
          )}
          title={r.users.join(', ')}
        >
          <span>{r.emoji}</span>
          <span className="tabular-nums text-muted-foreground">{r.count}</span>
        </button>
      ))}
    </div>
  );
}
