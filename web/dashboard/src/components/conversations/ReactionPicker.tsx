import { useAddReaction, useRemoveReaction } from '@/hooks/useConversations';
import { cn } from '@/lib/utils';
import { useState } from 'react';

const REACTIONS = ['👍', '👎', '❤️', '😂', '🎉', '👀'];

export interface ReactionPickerProps {
  messageId: string;
  conversationId: string;
  currentUserId: string;
  reactions?: { reaction: string; count: number; user_ids: string[] }[];
  className?: string;
}

export function ReactionPicker({
  messageId,
  conversationId,
  currentUserId,
  reactions = [],
  className,
}: ReactionPickerProps) {
  const [isOpen, setIsOpen] = useState(false);
  const addReaction = useAddReaction();
  const removeReaction = useRemoveReaction();

  const hasUserReacted = (reaction: string) => {
    return reactions.some(r => r.reaction === reaction && r.user_ids.includes(currentUserId));
  };

  const handleReactionClick = (reaction: string) => {
    if (hasUserReacted(reaction)) {
      removeReaction.mutate({ conversationId, messageId, reaction });
    } else {
      addReaction.mutate({ conversationId, messageId, reaction });
    }
  };

  return (
    <div className={cn('relative inline-block', className)}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          'p-1 rounded hover:bg-accent transition-colors',
          isOpen && 'bg-accent'
        )}
        type="button"
      >
        <span className="text-sm">+</span>
      </button>
      {isOpen && (
        <>
          <div
            className="fixed inset-0 z-10"
            onClick={() => setIsOpen(false)}
          />
          <div className="absolute bottom-full mb-1 left-0 z-20 bg-popover border rounded-lg shadow-lg p-1 flex gap-0.5">
            {REACTIONS.map((reaction) => (
              <button
                key={reaction}
                onClick={() => handleReactionClick(reaction)}
                className={cn(
                  'p-1.5 rounded hover:bg-accent transition-colors text-lg',
                  hasUserReacted(reaction) && 'bg-accent'
                )}
                type="button"
              >
                {reaction}
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  );
}