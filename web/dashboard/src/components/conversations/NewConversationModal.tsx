import type { ConversationType } from '@/api/conversations';
import { CONVERSATION_TYPES, UUID_RE } from '@/components/conversations/constants';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { cn } from '@/lib/utils';
import { Loader2, MessageSquarePlus } from 'lucide-react';

interface UsernameSuggestion {
  id: string;
  username: string;
  name?: string;
}

export interface NewConversationModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  conversationType: ConversationType;
  onConversationTypeChange: (type: ConversationType) => void;
  participantUsernames: string;
  onParticipantUsernamesChange: (value: string) => void;
  participantCaret: number;
  onParticipantCaretChange: (caret: number) => void;
  participantSuggestOpen: boolean;
  onParticipantSuggestOpenChange: (open: boolean) => void;
  participantSegment: string;
  usernameSuggestions: UsernameSuggestion[];
  usernameSearchLoading: boolean;
  onCreate: () => void;
  createPending: boolean;
  canCreate: boolean;
  onPickUsername: (username: string) => void;
  participantInputRef: React.RefObject<HTMLInputElement | null>;
}

export function NewConversationModal({
  open,
  onOpenChange,
  conversationType,
  onConversationTypeChange,
  participantUsernames,
  onParticipantUsernamesChange,
  participantCaret,
  onParticipantCaretChange,
  participantSuggestOpen,
  onParticipantSuggestOpenChange,
  participantSegment,
  usernameSuggestions,
  usernameSearchLoading,
  onCreate,
  createPending,
  canCreate,
  onPickUsername,
  participantInputRef,
}: NewConversationModalProps) {

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="conv-aviation-dialog sm:max-w-md">
        <div className="conv-aviation-dialog-accent" />
        <DialogHeader className="conv-aviation-dialog-header">
          <div className="conv-aviation-dialog-icon">
            <MessageSquarePlus className="h-5 w-5" />
          </div>
          <DialogTitle className="conv-aviation-dialog-title">New conversation</DialogTitle>
          <DialogDescription className="conv-aviation-dialog-description">
            Choose a type and add participant usernames (with or without @). You are included
            automatically for DMs.
          </DialogDescription>
        </DialogHeader>
        <div className="conv-aviation-dialog-body space-y-4 py-2">
          <div className="conv-aviation-form-group space-y-2">
            <Label htmlFor="new-conv-type" className="conv-aviation-form-label">Type</Label>
            <Select
              value={conversationType}
              onValueChange={(v) => onConversationTypeChange(v as ConversationType)}
            >
              <SelectTrigger id="new-conv-type" className="conv-aviation-select-trigger h-9 w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent position="popper" className="conv-aviation-select-content z-[200]">
                {CONVERSATION_TYPES.map(({ value, label }) => (
                  <SelectItem key={value} value={value} className="conv-aviation-select-item">
                    {label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="conv-aviation-form-group space-y-2">
            <Label htmlFor="new-conv-participants" className="conv-aviation-form-label">Participant usernames (comma-separated)</Label>
            <div className="relative">
              <Input
                ref={participantInputRef}
                id="new-conv-participants"
                placeholder="Start typing a username…"
                autoComplete="off"
                className="conv-aviation-input"
                value={participantUsernames}
                onChange={(e) => {
                  onParticipantUsernamesChange(e.target.value);
                  onParticipantCaretChange(e.target.selectionStart ?? e.target.value.length);
                }}
                onClick={(e) =>
                  onParticipantCaretChange(
                    e.currentTarget.selectionStart ?? e.currentTarget.value.length
                  )
                }
                onKeyUp={(e) =>
                  onParticipantCaretChange(
                    (e.target as HTMLInputElement).selectionStart ??
                      (e.target as HTMLInputElement).value.length
                  )
                }
                onSelect={(e) =>
                  onParticipantCaretChange(
                    (e.target as HTMLInputElement).selectionStart ??
                      (e.target as HTMLInputElement).value.length
                  )
                }
                onFocus={() => onParticipantSuggestOpenChange(true)}
                onBlur={() => {
                  window.setTimeout(() => onParticipantSuggestOpenChange(false), 200);
                }}
              />
              {participantSuggestOpen &&
                participantSegment.length >= 2 &&
                !UUID_RE.test(participantSegment) && (
                  <div
                    className="conv-aviation-suggestions absolute left-0 right-0 z-[200] mt-1 max-h-48 overflow-auto rounded-md border bg-card py-1 shadow-md"
                  >
                    {usernameSearchLoading ? (
                      <div
                        className="conv-aviation-suggestions-loading flex items-center gap-2 px-3 py-2 text-sm"
                      >
                        <Loader2 className="h-4 w-4 shrink-0 animate-spin opacity-70" />
                        Searching…
                      </div>
                    ) : usernameSuggestions.length === 0 ? (
                      <div className="conv-aviation-suggestions-empty px-3 py-2 text-sm">
                        No matching users
                      </div>
                    ) : (
                      <ul className="conv-aviation-suggestions-list py-0.5">
                        {usernameSuggestions.map((u) => (
                          <li key={u.id}>
                            <button
                              type="button"
                              className={cn(
                                'conv-aviation-suggestions-item flex w-full flex-col gap-0.5 px-3 py-2 text-left text-sm',
                                'hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground focus:outline-none'
                              )}
                              onMouseDown={(e) => e.preventDefault()}
                              onClick={() => onPickUsername(u.username)}
                            >
                              <span className="conv-aviation-suggestions-username font-medium">@{u.username}</span>
                              {u.name ? (
                                <span className="conv-aviation-suggestions-name text-xs">
                                  {u.name}
                                </span>
                              ) : null}
                            </button>
                          </li>
                        ))}
                      </ul>
                    )}
                  </div>
                )}
            </div>
          </div>
          <div className="conv-aviation-dialog-actions flex justify-end gap-2">
            <Button variant="outline" className="conv-aviation-btn-outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button
              className="conv-aviation-btn-primary"
              onClick={onCreate}
              disabled={createPending || !canCreate}
            >
              {createPending ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                'Create'
              )}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
