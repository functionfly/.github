import { conversationsApi, type ConversationType } from '@/api/conversations';
import { usersApi } from '@/api/users';
import {
  BountyAttachModal,
  ExecutableMessage,
  FixModeLayout,
  ResolutionBanner,
  RunInThreadPanel,
} from '@/components/conversations';
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
import { ScrollArea } from '@/components/ui/scroll-area';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { useDebounce } from '@/hooks/useInfiniteScroll';
import { cn } from '@/lib/utils';
import { useAuthStore } from '@/stores/authStore';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { formatDistanceToNow } from 'date-fns';
import { CheckCircle, Coins, Loader2, MessageSquare, Play, Plus, Send } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';

const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

function normalizeParticipantHandle(s: string): string {
  const t = s.trim();
  if (t.startsWith('@')) return t.slice(1).trim();
  return t;
}

function participantSegmentAtCaret(full: string, caret: number): string {
  const left = full.slice(0, Math.min(Math.max(caret, 0), full.length));
  const lastComma = left.lastIndexOf(',');
  const raw = lastComma === -1 ? left : left.slice(lastComma + 1);
  return normalizeParticipantHandle(raw);
}

function applyParticipantUsernamePick(
  full: string,
  caret: number,
  username: string
): { value: string; caret: number } {
  const left = full.slice(0, caret);
  const lastComma = left.lastIndexOf(',');
  const prefix = lastComma === -1 ? '' : left.slice(0, lastComma + 1);
  const suffix = full.slice(caret);
  let head: string;
  if (!prefix) {
    head = `${username}, `;
  } else {
    const needsSpace = !/,\s*$/.test(prefix);
    head = (needsSpace ? `${prefix} ` : prefix) + `${username}, `;
  }
  return { value: head + suffix, caret: head.length };
}

const CONVERSATION_TYPES: { value: ConversationType; label: string }[] = [
  { value: 'dm', label: 'Direct message' },
  { value: 'function_thread', label: 'Function thread' },
  { value: 'issue_thread', label: 'Issue thread' },
  { value: 'fix_mode', label: 'Fix mode' },
  { value: 'bounty_thread', label: 'Bounty thread' },
  { value: 'org_thread', label: 'Org thread' },
  { value: 'security_disclosure', label: 'Security disclosure' },
];

function formatParticipantLine(
  participantIds: string[] | undefined,
  currentUserId: string | undefined,
  displayFor: (id: string) => string
): string {
  if (!participantIds?.length) return 'Conversation';
  const self = currentUserId?.toLowerCase();
  const others = participantIds.filter((id) => id.toLowerCase() !== self);
  const idsToShow = others.length > 0 ? others : participantIds;
  const labels = idsToShow.map(displayFor);
  if (labels.every((l) => l === '…')) {
    return `${participantIds.length} participant(s)`;
  }
  if (labels.length <= 4) return labels.join(' · ');
  return `${labels.slice(0, 4).join(' · ')} +${labels.length - 4}`;
}

export default function ConversationsPage() {
  const { id: conversationId } = useParams<{ id?: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const user = useAuthStore((s) => s.user);
  const [messageDraft, setMessageDraft] = useState('');
  const [showRunPanel, setShowRunPanel] = useState(false);
  const [bountyModalOpen, setBountyModalOpen] = useState(false);
  const [newConversationModalOpen, setNewConversationModalOpen] = useState(false);
  const [newConvType, setNewConvType] = useState<ConversationType>('dm');
  const [newConvParticipantUsernames, setNewConvParticipantUsernames] = useState('');
  const participantInputRef = useRef<HTMLInputElement>(null);
  const [participantCaret, setParticipantCaret] = useState(0);
  const [participantSuggestOpen, setParticipantSuggestOpen] = useState(false);

  const participantSegment = participantSegmentAtCaret(
    newConvParticipantUsernames,
    participantCaret
  );
  const debouncedParticipantSegment = useDebounce(participantSegment, 250);

  const { data: usernameSearchData, isFetching: usernameSearchLoading } = useQuery({
    queryKey: ['users-search', debouncedParticipantSegment],
    queryFn: () => usersApi.searchUsersByUsername(debouncedParticipantSegment),
    enabled:
      newConversationModalOpen &&
      debouncedParticipantSegment.length >= 2 &&
      !UUID_RE.test(debouncedParticipantSegment),
    staleTime: 30_000,
  });

  const usernameSuggestions = useMemo(() => {
    const list = usernameSearchData?.users ?? [];
    const selfId = user?.id?.toLowerCase();
    if (!selfId) return list;
    return list.filter((u) => u.id.toLowerCase() !== selfId);
  }, [usernameSearchData?.users, user?.id]);

  const { data: listData, isLoading: listLoading } = useQuery({
    queryKey: ['conversations'],
    queryFn: () => conversationsApi.listConversations({ limit: 50 }),
  });

  const conversations = listData?.conversations ?? [];

  const participantIdsForLookup = useMemo(() => {
    const set = new Set<string>();
    for (const c of conversations) {
      for (const id of c.participant_ids ?? []) {
        if (id) set.add(id);
      }
    }
    return Array.from(set).sort();
  }, [conversations]);

  const { data: convData } = useQuery({
    queryKey: ['conversation', conversationId],
    queryFn: () => conversationsApi.getConversation(conversationId!),
    enabled: Boolean(conversationId),
  });

  const participantIdsForLookupWithOpen = useMemo(() => {
    const set = new Set(participantIdsForLookup);
    for (const id of convData?.participant_ids ?? []) {
      if (id) set.add(id);
    }
    return Array.from(set).sort();
  }, [participantIdsForLookup, convData?.participant_ids]);

  const { data: participantsLookup } = useQuery({
    queryKey: ['users-lookup-by-ids', participantIdsForLookupWithOpen.join(',')],
    queryFn: () => usersApi.lookupUsersByIds(participantIdsForLookupWithOpen),
    enabled: participantIdsForLookupWithOpen.length > 0,
    staleTime: 60_000,
  });

  const displayForParticipantId = useMemo(() => {
    const map = new Map<string, string>();
    for (const u of participantsLookup?.users ?? []) {
      const key = u.id.toLowerCase();
      const label =
        u.username?.trim() !== ''
          ? `@${u.username}`
          : u.name?.trim() !== ''
            ? u.name
            : u.id.slice(0, 8);
      map.set(key, label);
    }
    return (id: string) => map.get(id.toLowerCase()) ?? '…';
  }, [participantsLookup]);

  const { data: messagesData, isLoading: messagesLoading } = useQuery({
    queryKey: ['conversation-messages', conversationId],
    queryFn: () => conversationsApi.listMessages(conversationId!, { limit: 100 }),
    enabled: Boolean(conversationId),
  });

  useEffect(() => {
    if (!conversationId || messagesLoading) return;
    let cancelled = false;
    void (async () => {
      try {
        await conversationsApi.markConversationRead(conversationId);
        if (!cancelled) {
          queryClient.invalidateQueries({ queryKey: ['conversations'] });
        }
      } catch {
        // Non-fatal: list still shows stale unread until next refresh
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [conversationId, messagesLoading, queryClient]);

  const { data: bountiesData } = useQuery({
    queryKey: ['conversation-bounties', conversationId],
    queryFn: () => conversationsApi.listBounties(conversationId!),
    enabled: Boolean(conversationId),
  });

  const resolveMutation = useMutation({
    mutationFn: (messageId?: string) =>
      conversationsApi.resolveConversation(conversationId!, messageId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['conversation', conversationId] });
      queryClient.invalidateQueries({ queryKey: ['conversations'] });
      toast.success('Conversation resolved');
    },
    onError: (err: Error) => toast.error(err.message || 'Failed to resolve'),
  });
  const claimBountyMutation = useMutation({
    mutationFn: (bountyId: string) => conversationsApi.claimBounty(conversationId!, bountyId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['conversation-bounties', conversationId] });
      toast.success('Bounty claimed');
    },
    onError: (err: Error) => toast.error(err.message || 'Failed to claim'),
  });

  const sendMessage = useMutation({
    mutationFn: (content: string) => conversationsApi.createMessage(conversationId!, { content }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['conversation-messages', conversationId] });
      queryClient.invalidateQueries({ queryKey: ['conversations'] });
      setMessageDraft('');
    },
  });

  const createConversationMutation = useMutation({
    mutationFn: async () => {
      const handles = newConvParticipantUsernames
        .split(',')
        .map((s) => normalizeParticipantHandle(s))
        .filter(Boolean);

      const resolved: string[] = [];
      const seen = new Set<string>();

      for (const h of handles) {
        let id: string;
        if (UUID_RE.test(h)) {
          id = h;
        } else {
          try {
            const profile = await usersApi.getPublicProfile(h);
            id = profile.id;
          } catch {
            throw new Error(`User not found: @${h}`);
          }
        }
        const key = id.toLowerCase();
        if (seen.has(key)) continue;
        seen.add(key);
        resolved.push(id);
      }

      const selfLower = user?.id?.toLowerCase();
      const others = resolved.filter((id) => id.toLowerCase() !== selfLower);
      const participant_ids = user?.id ? [user.id, ...others] : others;
      if (participant_ids.length === 0) {
        throw new Error('At least one participant is required');
      }
      return conversationsApi.createConversation({
        type: newConvType,
        participant_ids,
      });
    },
    onSuccess: (conv) => {
      queryClient.invalidateQueries({ queryKey: ['conversations'] });
      setNewConversationModalOpen(false);
      setNewConvParticipantUsernames('');
      navigate(`/conversations/${conv.id}`);
    },
    onError: (err: Error) => toast.error(err.message || 'Failed to create conversation'),
  });

  const messages = messagesData?.messages ?? [];
  const isOwn = (authorId: string) => authorId === user?.id;

  const handleSend = () => {
    const t = messageDraft.trim();
    if (!t || !conversationId) return;
    sendMessage.mutate(t);
  };

  return (
    <div className="flex h-[calc(100vh-4rem)] border-t border-border">
      {/* Sidebar: conversation list */}
      <aside className="w-72 border-r border-border bg-muted/20 flex flex-col">
        <div className="p-3 border-b border-border flex items-center justify-between">
          <h2 className="font-semibold text-sm">Messages</h2>
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={() => setNewConversationModalOpen(true)}
            title="New conversation"
          >
            <Plus className="h-4 w-4" />
          </Button>
        </div>
        <ScrollArea className="flex-1">
          {listLoading ? (
            <div className="p-4 flex items-center justify-center">
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
            </div>
          ) : conversations.length === 0 ? (
            <div className="p-4 text-center text-sm text-muted-foreground">
              <MessageSquare className="h-8 w-8 mx-auto mb-2 opacity-50" />
              <p>No conversations yet.</p>
              <p className="text-xs mt-1">Start one from a function or profile.</p>
            </div>
          ) : (
            <ul className="p-1">
              {conversations.map((c) => (
                <li key={c.id}>
                  <Link
                    to={`/conversations/${c.id}`}
                    className={cn(
                      'flex flex-col gap-0.5 rounded-lg px-3 py-2 text-left transition-colors',
                      conversationId === c.id
                        ? 'bg-brand-500/15 border border-brand-500/30'
                        : 'hover:bg-muted/60'
                    )}
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div className="min-w-0 flex-1 flex flex-col gap-0.5">
                        <span className="text-xs text-muted-foreground capitalize">
                          {c.type.replace(/_/g, ' ')}
                        </span>
                        <span className="text-sm font-medium truncate">
                          {formatParticipantLine(
                            c.participant_ids,
                            user?.id,
                            displayForParticipantId
                          )}
                        </span>
                        <span className="text-xs text-muted-foreground">
                          {formatDistanceToNow(new Date(c.updated_at), { addSuffix: true })}
                        </span>
                      </div>
                      {(c.unread_count ?? 0) > 0 && (
                        <span
                          className="mt-0.5 shrink-0 flex h-5 min-w-5 items-center justify-center rounded-full bg-brand-500 px-1.5 text-[10px] font-semibold text-white tabular-nums"
                          aria-label={`${c.unread_count} unread`}
                        >
                          {(c.unread_count ?? 0) > 99 ? '99+' : c.unread_count}
                        </span>
                      )}
                    </div>
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </ScrollArea>
      </aside>

      {/* New conversation modal */}
      <Dialog open={newConversationModalOpen} onOpenChange={setNewConversationModalOpen}>
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
                value={newConvType}
                onValueChange={(v) => setNewConvType(v as ConversationType)}
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
                  value={newConvParticipantUsernames}
                  onChange={(e) => {
                    setNewConvParticipantUsernames(e.target.value);
                    setParticipantCaret(e.target.selectionStart ?? e.target.value.length);
                  }}
                  onClick={(e) =>
                    setParticipantCaret(
                      e.currentTarget.selectionStart ?? e.currentTarget.value.length
                    )
                  }
                  onKeyUp={(e) =>
                    setParticipantCaret(
                      (e.target as HTMLInputElement).selectionStart ??
                        (e.target as HTMLInputElement).value.length
                    )
                  }
                  onSelect={(e) =>
                    setParticipantCaret(
                      (e.target as HTMLInputElement).selectionStart ??
                        (e.target as HTMLInputElement).value.length
                    )
                  }
                  onFocus={() => setParticipantSuggestOpen(true)}
                  onBlur={() => {
                    window.setTimeout(() => setParticipantSuggestOpen(false), 200);
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
                                onClick={() => {
                                  const { value, caret } = applyParticipantUsernamePick(
                                    newConvParticipantUsernames,
                                    participantCaret,
                                    u.username
                                  );
                                  setNewConvParticipantUsernames(value);
                                  setParticipantCaret(caret);
                                  requestAnimationFrame(() => {
                                    const el = participantInputRef.current;
                                    if (el) {
                                      el.focus();
                                      el.setSelectionRange(caret, caret);
                                    }
                                  });
                                }}
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
              <Button variant="outline" onClick={() => setNewConversationModalOpen(false)}>
                Cancel
              </Button>
              <Button
                onClick={() => createConversationMutation.mutate()}
                disabled={
                  createConversationMutation.isPending ||
                  (!user?.id && !newConvParticipantUsernames.trim())
                }
              >
                {createConversationMutation.isPending ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  'Create'
                )}
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* Main: thread or empty state */}
      <main className="flex-1 flex flex-col min-w-0">
        {!conversationId ? (
          <div className="flex-1 flex items-center justify-center text-muted-foreground">
            {conversations.length > 0 ? (
              <p className="text-sm">Select a conversation</p>
            ) : (
              <p className="text-sm">Your executable conversations will appear here.</p>
            )}
          </div>
        ) : (
          <>
            {convData && (
              <div className="border-b border-border px-4 py-2 flex items-center justify-between gap-2">
                <div className="flex flex-col min-w-0 gap-0.5">
                  <span className="text-sm font-medium truncate">
                    {formatParticipantLine(
                      convData.participant_ids,
                      user?.id,
                      displayForParticipantId
                    )}
                  </span>
                  <span className="text-xs text-muted-foreground capitalize">
                    {convData.type.replace(/_/g, ' ')}
                  </span>
                </div>
                <div className="flex gap-1">
                  {!convData.resolved_at && (
                    <Button
                      size="sm"
                      variant="outline"
                      className="gap-1"
                      onClick={() => resolveMutation.mutate(undefined)}
                      disabled={resolveMutation.isPending}
                    >
                      <CheckCircle className="h-3.5 w-3.5" />
                      Resolve
                    </Button>
                  )}
                  <Button
                    size="sm"
                    variant="ghost"
                    className="gap-1"
                    onClick={() => setBountyModalOpen(true)}
                  >
                    <Coins className="h-3.5 w-3.5" />
                    Bounty
                  </Button>
                </div>
              </div>
            )}
            {conversationId && (
              <BountyAttachModal
                conversationId={conversationId}
                open={bountyModalOpen}
                onOpenChange={setBountyModalOpen}
              />
            )}
            <ScrollArea className="flex-1 p-4">
              {convData?.type === 'fix_mode' && (
                <div className="mb-4 p-3 rounded-lg border border-border bg-card">
                  <FixModeLayout
                    conversation={convData}
                    isResolved={Boolean(convData.resolved_at)}
                    onAcceptSolution={() => resolveMutation.mutate(undefined)}
                  />
                </div>
              )}
              {convData?.resolved_at && (
                <div className="mb-4">
                  <ResolutionBanner
                    resolvedAt={convData.resolved_at}
                    message="Conversation resolved"
                  />
                </div>
              )}
              {bountiesData?.bounties && bountiesData.bounties.length > 0 && (
                <div className="mb-4 flex flex-wrap gap-2">
                  {bountiesData.bounties.map((b) => (
                    <div
                      key={b.id}
                      className="flex items-center gap-2 rounded-md border border-amber-500/30 bg-amber-500/10 px-2 py-1 text-sm"
                    >
                      <Coins className="h-4 w-4 text-amber-600" />
                      <span>
                        +{b.amount_reputation} rep
                        {b.amount_cents ? ` · $${(b.amount_cents / 100).toFixed(2)}` : ''}
                      </span>
                      {b.claimed_by ? (
                        <span className="text-xs text-muted-foreground">Claimed</span>
                      ) : (
                        <Button
                          size="sm"
                          variant="outline"
                          className="h-6 text-xs"
                          onClick={() => claimBountyMutation.mutate(b.id)}
                          disabled={claimBountyMutation.isPending || !convData?.resolved_at}
                        >
                          Claim
                        </Button>
                      )}
                    </div>
                  ))}
                </div>
              )}
              {messagesLoading ? (
                <div className="flex justify-center py-8">
                  <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                </div>
              ) : (
                <div className="space-y-4">
                  {messages.map((m) => (
                    <ExecutableMessage key={m.id} message={m} isOwn={isOwn(m.author_id)} />
                  ))}
                </div>
              )}
            </ScrollArea>
            {showRunPanel && conversationId && (
              <div className="border-t border-border p-3 bg-muted/20">
                <RunInThreadPanel
                  conversationId={conversationId}
                  onSnippetAdded={() => setShowRunPanel(false)}
                />
              </div>
            )}
            <div className="border-t border-border p-3 flex gap-2">
              <Button
                variant="ghost"
                size="icon"
                className="shrink-0"
                onClick={() => setShowRunPanel((v) => !v)}
                title={showRunPanel ? 'Hide Run panel' : 'Run in thread'}
              >
                <Play className="h-4 w-4" />
              </Button>
              <Input
                placeholder="Type a message…"
                value={messageDraft}
                onChange={(e) => setMessageDraft(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && !e.shiftKey) {
                    e.preventDefault();
                    handleSend();
                  }
                }}
                className="flex-1"
              />
              <Button
                size="icon"
                onClick={handleSend}
                disabled={!messageDraft.trim() || sendMessage.isPending}
              >
                {sendMessage.isPending ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Send className="h-4 w-4" />
                )}
              </Button>
            </div>
          </>
        )}
      </main>
    </div>
  );
}
