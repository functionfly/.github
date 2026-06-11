import { conversationsApi } from '@/api/conversations';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { cn } from '@/lib/utils';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Check, Edit3, MoreHorizontal, Trash2, X, Loader2 } from 'lucide-react';
import { useState, useRef, useEffect } from 'react';
import { toast } from 'sonner';
import { ReactionPicker } from './ReactionPicker';
import type { ConversationMessage, ReactionSummary } from '@/api/conversations';
import { conversationKeys } from '@/hooks/useConversations';

export interface MessageActionsProps {
  messageId: string;
  conversationId: string;
  username: string;
  currentContent: string;
  isOwn: boolean;
  reactions?: ReactionSummary[];
  currentUserId?: string;
  onEditComplete?: () => void;
}

interface OptimisticEditState {
  originalContent: string;
  optimisticContent: string;
  isPending: boolean;
  error: string | null;
}

export function MessageActions({
  messageId,
  conversationId,
  username,
  currentContent,
  isOwn,
  reactions,
  currentUserId,
  onEditComplete,
}: MessageActionsProps) {
  const queryClient = useQueryClient();
  const [isEditing, setIsEditing] = useState(false);
  const [editDraft, setEditDraft] = useState(currentContent);
  const [optimisticEdit, setOptimisticEdit] = useState<OptimisticEditState | null>(null);
  const rollbackTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (rollbackTimeoutRef.current) {
        clearTimeout(rollbackTimeoutRef.current);
      }
    };
  }, []);

  const editMutation = useMutation({
    mutationFn: (content: string) =>
      conversationsApi.editMessage(username, conversationId, messageId, { content }),
    onMutate: async (newContent: string) => {
      await queryClient.cancelQueries({ queryKey: conversationKeys.messages(conversationId) });

      const previousMessages = queryClient.getQueryData(conversationKeys.messages(conversationId));

      setOptimisticEdit({
        originalContent: currentContent,
        optimisticContent: newContent,
        isPending: true,
        error: null,
      });

      queryClient.setQueryData(
        conversationKeys.messages(conversationId),
        (old: { messages: ConversationMessage[] } | undefined) => {
          if (!old) return old;
          return {
            messages: old.messages.map((m) =>
              m.id === messageId
                ? { ...m, content: newContent, edited_at: new Date().toISOString() }
                : m
            ),
          };
        }
      );

      return { previousMessages };
    },
    onSuccess: (updatedMessage) => {
      setOptimisticEdit(null);
      queryClient.setQueryData(
        conversationKeys.messages(conversationId),
        (old: { messages: ConversationMessage[] } | undefined) => {
          if (!old) return old;
          return {
            messages: old.messages.map((m) =>
              m.id === messageId ? updatedMessage : m
            ),
          };
        }
      );
      setIsEditing(false);
      onEditComplete?.();
      toast.success('Message updated');
    },
    onError: (err: Error, newContent, context) => {
      setOptimisticEdit({
        originalContent: currentContent,
        optimisticContent: newContent,
        isPending: false,
        error: err.message,
      });

      rollbackTimeoutRef.current = setTimeout(() => {
        setOptimisticEdit(null);
        queryClient.setQueryData(
          conversationKeys.messages(conversationId),
          context?.previousMessages
        );
        toast.error('Failed to edit message - changes rolled back');
      }, 3000);

      toast.error(err.message || 'Failed to edit message');
    },
    onSettled: () => {
      if (!optimisticEdit?.isPending) {
        setOptimisticEdit(null);
      }
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => conversationsApi.deleteMessage(username, conversationId, messageId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: conversationKeys.messages(conversationId) });
      queryClient.invalidateQueries({ queryKey: conversationKeys.lists() });
      toast.success('Message deleted');
    },
    onError: (err: Error) => toast.error(err.message || 'Failed to delete message'),
  });

  const startEdit = () => {
    setEditDraft(currentContent);
    setIsEditing(true);
  };

  const cancelEdit = () => {
    setIsEditing(false);
    setEditDraft(currentContent);
  };

  const saveEdit = () => {
    const trimmed = editDraft.trim();
    if (!trimmed || trimmed === currentContent) {
      cancelEdit();
      return;
    }
    editMutation.mutate(trimmed);
  };

  if (isEditing) {
    return (
      <div className="flex items-center gap-1.5 mt-1">
        <Input
          value={editDraft}
          onChange={(e) => setEditDraft(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
              e.preventDefault();
              saveEdit();
            }
            if (e.key === 'Escape') cancelEdit();
          }}
          className="h-7 text-sm"
          autoFocus
        />
        <Button
          size="icon"
          variant="ghost"
          className="h-7 w-7"
          onClick={saveEdit}
          disabled={editMutation.isPending || !editDraft.trim()}
        >
          {editMutation.isPending ? (
            <Loader2 className="h-3.5 w-3.5 animate-spin" />
          ) : (
            <Check className="h-3.5 w-3.5" />
          )}
        </Button>
        <Button size="icon" variant="ghost" className="h-7 w-7" onClick={cancelEdit}>
          <X className="h-3.5 w-3.5" />
        </Button>
      </div>
    );
  }

  if (optimisticEdit) {
    return (
      <div className="flex items-center gap-1.5 mt-1 text-xs text-muted-foreground">
        {optimisticEdit.isPending ? (
          <>
            <Loader2 className="h-3 w-3 animate-spin" />
            <span className="text-muted-foreground">Saving...</span>
          </>
        ) : optimisticEdit.error ? (
          <>
            <span className="text-destructive">Failed</span>
            <button
              onClick={() => {
                setOptimisticEdit(null);
                setEditDraft(optimisticEdit.originalContent);
                setIsEditing(true);
              }}
              className="text-xs underline hover:text-foreground"
            >
              Retry
            </button>
          </>
        ) : null}
      </div>
    );
  }

  if (!isOwn && !currentUserId) return null;

  return (
    <div className="flex items-center gap-1">
      {currentUserId && (
        <ReactionPicker
          messageId={messageId}
          conversationId={conversationId}
          currentUserId={currentUserId}
          reactions={reactions}
        />
      )}
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            className={cn(
              'h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity',
              'data-[state=open]:opacity-100'
            )}
          >
            <MoreHorizontal className="h-3.5 w-3.5" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-36">
          <DropdownMenuItem onClick={startEdit}>
            <Edit3 className="mr-2 h-3.5 w-3.5" />
            Edit
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={() => deleteMutation.mutate()}
            className="text-destructive focus:text-destructive"
          >
            <Trash2 className="mr-2 h-3.5 w-3.5" />
            Delete
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
