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
import { Loader2 } from 'lucide-react';
import { useRef } from 'react';

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
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>New conversation</DialogTitle>
          <DialogDescription>
            Choose a type and add participant usernames (with or without @). You are included
            automatically for DMs.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="new-conv-type">Type</Label>
            <Select
              value={conversationType}
              onValueChange={(v) => onConversationTypeChange(v as ConversationType)}
            >
              <SelectTrigger id="new-conv-type" className="h-9 w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent position="popper" className="z-200">
                {CONVERSATION_TYPES.map(({ value, label }) => (
                  <SelectItem key={value} value={value}>
                    {label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label htmlFor="new-conv-participants">Participant usernames (comma-separated)</Label>
            <div className="relative">
              <Input
                ref={participantInputRef}
                id="new-conv-participants"
                placeholder="Start typing a username…"
                autoComplete="off"
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
                    className="absolute left-0 right-0 z-200 mt-1 max-h-48 overflow-auto rounded-md border bg-card py-1 shadow-md"
                    style={{
                      borderColor: 'var(--border-default)',
                      backgroundColor: 'var(--card)',
                    }}
                  >
                    {usernameSearchLoading ? (
                      <div
                        className="flex items-center gap-2 px-3 py-2 text-sm"
                        style={{ color: 'var(--text-muted)' }}
                      >
                        <Loader2 className="h-4 w-4 shrink-0 animate-spin opacity-70" />
                        Searching…
                      </div>
                    ) : usernameSuggestions.length === 0 ? (
                      <div className="px-3 py-2 text-sm" style={{ color: 'var(--text-muted)' }}>
                        No matching users
                      </div>
                    ) : (
                      <ul className="py-0.5" style={{ color: 'var(--card-foreground)' }}>
                        {usernameSuggestions.map((u) => (
                          <li key={u.id}>
                            <button
                              type="button"
                              className={cn(
                                'flex w-full flex-col gap-0.5 px-3 py-2 text-left text-sm',
                                'hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground focus:outline-none'
                              )}
                              style={{ color: 'var(--card-foreground)' }}
                              onMouseDown={(e) => e.preventDefault()}
                              onClick={() => onPickUsername(u.username)}
                            >
                              <span className="font-medium">@{u.username}</span>
                              {u.name ? (
                                <span className="text-xs" style={{ color: 'var(--text-muted)' }}>
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
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button
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
