import { ChevronDown, ChevronUp } from 'lucide-react';
import { cn } from '@/lib/utils';

interface VoteControlsProps {
  score: number;
  userVote?: number | null;
  onVote: (value: 1 | -1) => void;
  disabled?: boolean;
  compact?: boolean;
}

export function VoteControls({ score, userVote, onVote, disabled, compact }: VoteControlsProps) {
  return (
    <div className={cn('community-vote-stack', compact && 'scale-90')}>
      <button
        type="button"
        className={cn('community-vote-btn', userVote === 1 && 'active-up')}
        onClick={(e) => {
          e.preventDefault();
          e.stopPropagation();
          onVote(1);
        }}
        disabled={disabled}
        aria-label="Upvote"
      >
        <ChevronUp className="h-4 w-4" />
      </button>
      <span className="community-vote-score">{score}</span>
      <button
        type="button"
        className={cn('community-vote-btn', userVote === -1 && 'active-down')}
        onClick={(e) => {
          e.preventDefault();
          e.stopPropagation();
          onVote(-1);
        }}
        disabled={disabled}
        aria-label="Downvote"
      >
        <ChevronDown className="h-4 w-4" />
      </button>
    </div>
  );
}
