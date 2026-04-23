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
import { Check, Edit3, MoreHorizontal, Trash2, X } from 'lucide-react';
import { useState } from 'react';
import { toast } from 'sonner';

export interface MessageActionsProps {
  messageId: string;
  conversationId: string;
  username: string;
  currentContent: string;
  isOwn: boolean;
  onEditComplete?: () => void;
}

export function MessageActions({
  messageId,
  conversationId,
  username,
  currentContent,
  isOwn,
  onEditComplete,
}: MessageActionsProps) {
  const queryClient = useQueryClient();
  const [isEditing, setIsEditing] = useState(false);
  const [editDraft, setEditDraft] = useState(currentContent);

  const editMutation = useMutation({
    mutationFn: (content: string) =>
      conversationsApi.editMessage(username, conversationId, messageId, { content }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['conversation-messages', username, conversationId] });
      setIsEditing(false);
      onEditComplete?.();
      toast.success('Message updated');
    },
    onError: (err: Error) => toast.error(err.message || 'Failed to edit message'),
  });

  const deleteMutation = useMutation({
    mutationFn: () => conversationsApi.deleteMessage(username, conversationId, messageId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['conversation-messages', username, conversationId] });
      queryClient.invalidateQueries({ queryKey: ['conversations', username] });
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
          <Check className="h-3.5 w-3.5" />
        </Button>
        <Button size="icon" variant="ghost" className="h-7 w-7" onClick={cancelEdit}>
          <X className="h-3.5 w-3.5" />
        </Button>
      </div>
    );
  }

  if (!isOwn) return null;

  return (
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
  );
}
