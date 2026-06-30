import { conversationsApi, type ConversationType } from '@/api/conversations';
import { usersApi } from '@/api/users';
import {
  BountyAttachModal,
  ExecutableMessage,
  FixModeLayout,
  MessageSearch,
  ResolutionBanner,
  RunInThreadPanel,
  TypingIndicator,
} from '@/components/conversations';
import { ConversationHeader } from '@/components/conversations/ConversationHeader';
import { ConversationSidebar } from '@/components/conversations/ConversationSidebar';
import { NewConversationModal } from '@/components/conversations/NewConversationModal';
import {
  applyParticipantUsernamePick,
  normalizeParticipantHandle,
  participantSegmentAtCaret,
  UUID_RE,
} from '@/components/conversations/constants';
import { Button } from '@/components/ui/button';
import { ChatInput } from '@/components/ui/chat-input';
import { EmptyChat } from '@/components/ui/empty-chat';
import { ScrollArea } from '@/components/ui/scroll-area';
import { SkeletonMessages } from '@/components/ui/skeleton-chat';
import { useRealtimeMessages } from '@/hooks/useConversations';
import { useDebounce } from '@/hooks/useInfiniteScroll';
import { useAuthStore } from '@/stores/authStore';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Coins, Play, MessageSquare, Plus } from 'lucide-react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import { PageGrid } from '@/components/containment';
import { usePageTitle } from '@/hooks';
import '@/styles/conversations-aviation.css';

export default function ConversationsPage() {
  usePageTitle('Conversations');
  const { username, id: conversationId } = useParams<{ username?: string; id?: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const user = useAuthStore((s) => s.user);
  const currentUsername = username || user?.username;
  const [messageDraft, setMessageDraft] = useState('');
  const [showRunPanel, setShowRunPanel] = useState(false);
  const [searchOpen, setSearchOpen] = useState(false);
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
    queryKey: ['conversations', currentUsername],
    queryFn: () => conversationsApi.listConversations(currentUsername!, { limit: 50 }),
    enabled: Boolean(currentUsername),
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
    queryKey: ['conversation', currentUsername, conversationId],
    queryFn: () => conversationsApi.getConversation(currentUsername!, conversationId!),
    enabled: Boolean(currentUsername) && Boolean(conversationId),
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

  const {
    data: messagesData,
    isLoading: messagesLoading,
    typingUsers,
    sendTyping,
  } = useRealtimeMessages(
    Boolean(currentUsername) && Boolean(conversationId) ? conversationId! : '',
    { limit: 100 },
  );

  useEffect(() => {
    if (!currentUsername || !conversationId || messagesLoading) return;
    let cancelled = false;
    void (async () => {
      try {
        await conversationsApi.markConversationRead(currentUsername, conversationId);
        if (!cancelled) {
          queryClient.invalidateQueries({ queryKey: ['conversations', currentUsername] });
        }
      } catch {
        // Non-fatal: list still shows stale unread until next refresh
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [currentUsername, conversationId, messagesLoading, queryClient]);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setSearchOpen(true);
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, []);

  const { data: bountiesData } = useQuery({
    queryKey: ['conversation-bounties', currentUsername, conversationId],
    queryFn: () => conversationsApi.listBounties(currentUsername!, conversationId!),
    enabled: Boolean(currentUsername) && Boolean(conversationId),
  });

  const resolveMutation = useMutation({
    mutationFn: (messageId?: string) =>
      conversationsApi.resolveConversation(currentUsername!, conversationId!, messageId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['conversation', currentUsername, conversationId] });
      queryClient.invalidateQueries({ queryKey: ['conversations', currentUsername] });
      toast.success('Conversation resolved');
    },
    onError: (err: Error) => toast.error(err.message || 'Failed to resolve'),
  });
  const claimBountyMutation = useMutation({
    mutationFn: (bountyId: string) => conversationsApi.claimBounty(currentUsername!, conversationId!, bountyId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['conversation-bounties', currentUsername, conversationId] });
      toast.success('Bounty claimed');
    },
    onError: (err: Error) => toast.error(err.message || 'Failed to claim'),
  });

  const sendMessage = useMutation({
    mutationFn: (content: string) => conversationsApi.createMessage(currentUsername!, conversationId!, { content }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['conversation-messages', currentUsername, conversationId] });
      queryClient.invalidateQueries({ queryKey: ['conversations', currentUsername] });
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
      return conversationsApi.createConversation(currentUsername!, {
        type: newConvType,
        participant_ids,
      });
    },
    onSuccess: (conv) => {
      queryClient.invalidateQueries({ queryKey: ['conversations', currentUsername] });
      setNewConversationModalOpen(false);
      setNewConvParticipantUsernames('');
      navigate(`/u/${currentUsername}/conversations/${conv.id}`);
    },
    onError: (err: Error) => toast.error(err.message || 'Failed to create conversation'),
  });

  const messages = messagesData?.messages ?? [];
  const isOwn = (authorId: string) => authorId === user?.id;
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom when new messages arrive
  const scrollToBottom = useCallback(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, []);

  useEffect(() => {
    if (!messagesLoading && messages.length > 0) {
      scrollToBottom();
    }
  }, [messages.length, messagesLoading, scrollToBottom]);

  const handleSend = () => {
    const t = messageDraft.trim();
    if (!t || !conversationId) return;
    sendMessage.mutate(t);
  };

  const handlePickUsername = (pickedUsername: string) => {
    const { value, caret } = applyParticipantUsernamePick(
      newConvParticipantUsernames,
      participantCaret,
      pickedUsername
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
  };

  return (
    <div className="conv-aviation-page flex h-full" style={{ background: 'var(--bg)' }}>
      <PageGrid />
      {/* Conversations sidebar */}
      <div className="conv-aviation-sidebar w-72 shrink-0 flex flex-col" style={{ borderRight: '1px solid var(--panel-edge)' }}>
        <ConversationSidebar
          conversations={conversations}
          loading={listLoading}
          currentUsername={currentUsername ?? ''}
          activeConversationId={conversationId}
          currentUserId={user?.id}
          displayForParticipantId={displayForParticipantId}
          onNewConversation={() => setNewConversationModalOpen(true)}
        />
      </div>

      <NewConversationModal
        open={newConversationModalOpen}
        onOpenChange={setNewConversationModalOpen}
        conversationType={newConvType}
        onConversationTypeChange={setNewConvType}
        participantUsernames={newConvParticipantUsernames}
        onParticipantUsernamesChange={setNewConvParticipantUsernames}
        participantCaret={participantCaret}
        onParticipantCaretChange={setParticipantCaret}
        participantSuggestOpen={participantSuggestOpen}
        onParticipantSuggestOpenChange={setParticipantSuggestOpen}
        participantSegment={participantSegment}
        usernameSuggestions={usernameSuggestions}
        usernameSearchLoading={usernameSearchLoading}
        onCreate={() => createConversationMutation.mutate()}
        createPending={createConversationMutation.isPending}
        canCreate={!(!user?.id && !newConvParticipantUsernames.trim())}
        onPickUsername={handlePickUsername}
        participantInputRef={participantInputRef}
      />

      <main className="conv-aviation-main flex-1 flex flex-col min-w-0 relative">
        {!conversationId ? (
          <div className="conv-aviation-empty-state">
            {conversations.length > 0 ? (
              <>
                <div className="conv-aviation-empty-icon-wrap">
                  <MessageSquare />
                </div>
                <h2 className="conv-aviation-empty-title">Select a conversation</h2>
                <p className="conv-aviation-empty-description">
                  Choose a thread from the sidebar to start collaborating.
                </p>
              </>
            ) : (
              <>
                <div className="conv-aviation-empty-icon-wrap">
                  <MessageSquare />
                </div>
                <h2 className="conv-aviation-empty-title">Your conversations</h2>
                <p className="conv-aviation-empty-description">
                  Executable threads for functions, fixes, bounties, and DMs — all in one place.
                </p>
                <div className="conv-aviation-empty-cta">
                  <Button
                    onClick={() => setNewConversationModalOpen(true)}
                    className="conv-aviation-btn-primary"
                  >
                    <Plus className="h-4 w-4 mr-1" />
                    Start a conversation
                  </Button>
                </div>
              </>
            )}
          </div>
        ) : (
          <>
            {convData && (
              <ConversationHeader
                conversation={convData}
                currentUserId={user?.id}
                displayForParticipantId={displayForParticipantId}
                onSearch={() => setSearchOpen(true)}
                onResolve={() => resolveMutation.mutate(undefined)}
                onBounty={() => setBountyModalOpen(true)}
                resolvePending={resolveMutation.isPending}
              />
            )}
            {conversationId && currentUsername && (
              <BountyAttachModal
                username={currentUsername}
                conversationId={conversationId}
                open={bountyModalOpen}
                onOpenChange={setBountyModalOpen}
              />
            )}
            {currentUsername && (
              <MessageSearch
                username={currentUsername}
                conversationId={conversationId}
                open={searchOpen}
                onOpenChange={setSearchOpen}
              />
            )}
            <ScrollArea className="conv-aviation-message-list flex-1 p-4">
              {convData?.type === 'fix_mode' && (
                <div className="conv-aviation-fix-timeline mb-4 p-3 rounded-[var(--radius)]" style={{ border: '1px solid var(--panel-edge)', background: 'var(--panel-raised)' }}>
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
                      className="conv-aviation-bounty-badge flex items-center gap-2 rounded-md border px-2 py-1 text-sm"
                    >
                      <Coins className="h-4 w-4" />
                      <span>
                        +{b.amount_reputation} rep
                        {b.amount_cents ? ` · $${(b.amount_cents / 100).toFixed(2)}` : ''}
                      </span>
                      {b.claimed_by ? (
                        <span className="text-xs" style={{ color: 'var(--text-faint)' }}>Claimed</span>
                      ) : (
                        <Button
                          size="sm"
                          variant="outline"
                          className="conv-aviation-btn h-6 text-xs"
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
                <SkeletonMessages count={6} />
              ) : messages.length === 0 ? (
                <EmptyChat conversationType={convData?.type} />
              ) : (
                <div className="space-y-4">
                  {messages.map((m) => (
                    <ExecutableMessage
                      key={m.id}
                      message={m}
                      isOwn={isOwn(m.author_id)}
                      username={currentUsername}
                      currentUserId={user?.id}
                    />
                  ))}
                  <div ref={messagesEndRef} />
                </div>
              )}
            </ScrollArea>
            <TypingIndicator
              typingUsers={typingUsers}
              displayForParticipantId={displayForParticipantId}
            />
            {showRunPanel && conversationId && currentUsername && (
              <div className="conv-aviation-run-panel p-3" style={{ borderTop: '1px solid var(--panel-edge)', background: 'var(--panel-raised)' }}>
                <RunInThreadPanel
                  username={currentUsername}
                  conversationId={conversationId}
                  onSnippetAdded={() => setShowRunPanel(false)}
                />
              </div>
            )}
            <div className="conv-aviation-composer flex items-center gap-1 px-2 py-1" style={{ borderTop: '1px solid var(--panel-edge)' }}>
              <Button
                variant="ghost"
                size="sm"
                className="conv-aviation-btn h-7 px-2 text-xs"
                onClick={() => setShowRunPanel((v) => !v)}
              >
                <Play className="h-3.5 w-3.5 mr-1" />
                Run in thread
              </Button>
            </div>
            <ChatInput
              value={messageDraft}
              onChange={setMessageDraft}
              onSend={handleSend}
              onTyping={sendTyping}
              pending={sendMessage.isPending}
            />
          </>
        )}
      </main>
    </div>
  );
}
