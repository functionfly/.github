import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { MessageSquare, Plus, Send, Loader2, Play, CheckCircle, Coins } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { conversationsApi, type Conversation, type ConversationMessage } from "@/api/conversations";
import {
  ExecutableMessage,
  RunInThreadPanel,
  ResolutionBanner,
  BountyAttachModal,
  FixModeLayout,
} from "@/components/conversations";
import { useAuthStore } from "@/stores/authStore";
import { formatDistanceToNow } from "date-fns";
import { toast } from "sonner";
import { cn } from "@/lib/utils";

export default function ConversationsPage() {
  const { id: conversationId } = useParams<{ id?: string }>();
  const queryClient = useQueryClient();
  const user = useAuthStore((s) => s.user);
  const [messageDraft, setMessageDraft] = useState("");
  const [showRunPanel, setShowRunPanel] = useState(false);
  const [bountyModalOpen, setBountyModalOpen] = useState(false);

  const { data: listData, isLoading: listLoading } = useQuery({
    queryKey: ["conversations"],
    queryFn: () => conversationsApi.listConversations({ limit: 50 }),
  });

  const { data: convData, isLoading: convLoading } = useQuery({
    queryKey: ["conversation", conversationId],
    queryFn: () => conversationsApi.getConversation(conversationId!),
    enabled: Boolean(conversationId),
  });

  const { data: messagesData, isLoading: messagesLoading } = useQuery({
    queryKey: ["conversation-messages", conversationId],
    queryFn: () =>
      conversationsApi.listMessages(conversationId!, { limit: 100 }),
    enabled: Boolean(conversationId),
  });

  const { data: bountiesData } = useQuery({
    queryKey: ["conversation-bounties", conversationId],
    queryFn: () => conversationsApi.listBounties(conversationId!),
    enabled: Boolean(conversationId),
  });

  const resolveMutation = useMutation({
    mutationFn: (messageId?: string) =>
      conversationsApi.resolveConversation(conversationId!, messageId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["conversation", conversationId] });
      queryClient.invalidateQueries({ queryKey: ["conversations"] });
      toast.success("Conversation resolved");
    },
    onError: (err: Error) => toast.error(err.message || "Failed to resolve"),
  });
  const claimBountyMutation = useMutation({
    mutationFn: (bountyId: string) =>
      conversationsApi.claimBounty(conversationId!, bountyId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["conversation-bounties", conversationId] });
      toast.success("Bounty claimed");
    },
    onError: (err: Error) => toast.error(err.message || "Failed to claim"),
  });

  const sendMessage = useMutation({
    mutationFn: (content: string) =>
      conversationsApi.createMessage(conversationId!, { content }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["conversation-messages", conversationId] });
      queryClient.invalidateQueries({ queryKey: ["conversations"] });
      setMessageDraft("");
    },
  });

  const conversations = listData?.conversations ?? [];
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
            onClick={() => {
              // TODO: open new conversation modal
            }}
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
                      "flex flex-col gap-0.5 rounded-lg px-3 py-2 text-left transition-colors",
                      conversationId === c.id
                        ? "bg-brand-500/15 border border-brand-500/30"
                        : "hover:bg-muted/60"
                    )}
                  >
                    <span className="text-xs text-muted-foreground capitalize">
                      {c.type.replace(/_/g, " ")}
                    </span>
                    <span className="text-sm font-medium truncate">
                      {c.participant_ids?.length
                        ? `${c.participant_ids.length} participant(s)`
                        : "Conversation"}
                    </span>
                    <span className="text-xs text-muted-foreground">
                      {formatDistanceToNow(new Date(c.updated_at), { addSuffix: true })}
                    </span>
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </ScrollArea>
      </aside>

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
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium capitalize">
                    {convData.type.replace(/_/g, " ")}
                  </span>
                  <span className="text-xs text-muted-foreground">
                    {convData.participant_ids?.length ?? 0} participant(s)
                  </span>
                </div>
                <div className="flex gap-1">
                  {!convData.resolved_at && (
                    <Button
                      size="sm"
                      variant="outline"
                      className="gap-1"
                      onClick={() => resolveMutation.mutate()}
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
              {convData?.type === "fix_mode" && (
                <div className="mb-4 p-3 rounded-lg border border-border bg-card">
                  <FixModeLayout
                    conversation={convData}
                    isResolved={Boolean(convData.resolved_at)}
                    onAcceptSolution={() => resolveMutation.mutate()}
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
                        {b.amount_cents ? ` · $${(b.amount_cents / 100).toFixed(2)}` : ""}
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
                    <ExecutableMessage
                      key={m.id}
                      message={m}
                      isOwn={isOwn(m.author_id)}
                    />
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
                title={showRunPanel ? "Hide Run panel" : "Run in thread"}
              >
                <Play className="h-4 w-4" />
              </Button>
              <Input
                placeholder="Type a message…"
                value={messageDraft}
                onChange={(e) => setMessageDraft(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter" && !e.shiftKey) {
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
