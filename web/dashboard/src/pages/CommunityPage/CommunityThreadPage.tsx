'use client';

import {
  communityApi,
  displayAuthor,
  formatRelativeTime,
  type CommunityComment,
} from '@/api/community';
import { Navbar } from '@/components/common/Navbar';
import { CommentThread } from '@/components/community/CommentThread';
import { VoteControls } from '@/components/community/VoteControls';
import { MetaTags } from '@/components/seo/MetaTags';
import { Button } from '@/components/ui/button';
import { Footer } from '@/pages/LandingPage/components/Footer';
import { useAuthStore } from '@/stores/authStore';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { ArrowLeft, CheckCircle2, Loader2, MessageSquare } from 'lucide-react';
import { useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import './Community.css';

export function CommunityThreadPage() {
  const { postId } = useParams<{ postId: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const user = useAuthStore((s) => s.user);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  const [replyBody, setReplyBody] = useState('');
  const [replyToId, setReplyToId] = useState<string | null>(null);
  const [votingId, setVotingId] = useState<string | null>(null);

  const { data, isLoading, error } = useQuery({
    queryKey: ['community-post', postId],
    queryFn: () => communityApi.getPost(postId!),
    enabled: !!postId,
  });

  const post = data?.post;
  const comments = data?.comments ?? [];

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ['community-post', postId] });
    queryClient.invalidateQueries({ queryKey: ['community-posts'] });
  };

  const replyMutation = useMutation({
    mutationFn: (body: string) =>
      communityApi.createComment(postId!, {
        body,
        parent_id: replyToId || undefined,
      }),
    onSuccess: () => {
      setReplyBody('');
      setReplyToId(null);
      invalidate();
      toast.success('Reply posted');
    },
    onError: () => toast.error('Failed to post reply'),
  });

  const voteMutation = useMutation({
    mutationFn: communityApi.vote,
    onMutate: ({ target_id }) => setVotingId(target_id),
    onSettled: () => setVotingId(null),
    onSuccess: invalidate,
    onError: () => toast.error('Sign in to vote'),
  });

  const acceptMutation = useMutation({
    mutationFn: (commentId: string) => communityApi.acceptComment(postId!, commentId),
    onSuccess: () => {
      invalidate();
      toast.success('Marked as solved');
    },
    onError: () => toast.error('Could not accept answer'),
  });

  const requireAuth = (action: () => void) => {
    if (!isAuthenticated) {
      navigate(`/login?redirect=${encodeURIComponent(window.location.pathname)}`);
      return;
    }
    action();
  };

  const handlePostVote = (value: 1 | -1) => {
    requireAuth(() => voteMutation.mutate({ target_type: 'post', target_id: postId!, value }));
  };

  const handleCommentVote = (commentId: string, value: 1 | -1) => {
    requireAuth(() => voteMutation.mutate({ target_type: 'comment', target_id: commentId, value }));
  };

  if (isLoading) {
    return (
      <div className="community-page">
        <Navbar />
        <div className="community-empty py-24">
          <Loader2 className="h-8 w-8 animate-spin mx-auto" />
        </div>
      </div>
    );
  }

  if (error || !post) {
    return (
      <div className="community-page">
        <Navbar />
        <div className="community-empty py-24">
          <p className="font-medium mb-4">Thread not found</p>
          <Button asChild variant="outline">
            <Link to="/community">Back to community</Link>
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="community-page">
      <MetaTags title={`${post.title} | Community`} description={post.body.slice(0, 160)} />
      <Navbar />

      <div className="max-w-3xl mx-auto px-4 py-6">
        <Link
          to="/community"
          className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground mb-4"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to community
        </Link>

        <article className="community-thread-card">
          <div className="community-thread-header">
            <div className="flex gap-3">
              <VoteControls
                score={post.vote_score}
                userVote={post.user_vote}
                onVote={handlePostVote}
                disabled={votingId === post.id}
              />
              <div className="min-w-0 flex-1">
                <div className="community-post-meta mb-2">
                  <span className="community-badge community-badge-category">
                    {post.category?.name}
                  </span>
                  {post.status === 'solved' && (
                    <span className="community-badge community-badge-solved">
                      <CheckCircle2 className="h-3 w-3" />
                      Solved
                    </span>
                  )}
                  <span>{displayAuthor(post.author)}</span>
                  <span>·</span>
                  <span>{formatRelativeTime(post.created_at)}</span>
                </div>
                <h1 className="text-xl font-bold text-foreground leading-snug">{post.title}</h1>
                <div className="community-stats mt-2">
                  <span>{post.reply_count} replies</span>
                  <span>{post.view_count} views</span>
                </div>
              </div>
            </div>
          </div>
          <div className="community-thread-body">{post.body}</div>

          <section className="community-comments-section">
            <h2 className="text-sm font-semibold uppercase tracking-wide text-muted-foreground mb-4 flex items-center gap-2">
              <MessageSquare className="h-4 w-4" />
              {comments.length} {comments.length === 1 ? 'Reply' : 'Replies'}
            </h2>

            <CommentThread
              comments={comments as CommunityComment[]}
              currentUserId={user?.id}
              postAuthorId={post.author_id}
              onVote={handleCommentVote}
              onReply={(id) => {
                requireAuth(() => {
                  setReplyToId(id);
                  document.getElementById('community-reply')?.focus();
                });
              }}
              onAccept={(commentId) => requireAuth(() => acceptMutation.mutate(commentId))}
              votingId={votingId}
            />

            <div className="community-reply-box">
              {replyToId && (
                <p className="text-xs text-muted-foreground mb-2">
                  Replying to a comment ·{' '}
                  <button type="button" className="underline" onClick={() => setReplyToId(null)}>
                    cancel
                  </button>
                </p>
              )}
              <textarea
                id="community-reply"
                value={replyBody}
                onChange={(e) => setReplyBody(e.target.value)}
                placeholder={
                  isAuthenticated ? 'Share your answer or ask a follow-up…' : 'Sign in to reply'
                }
                disabled={!isAuthenticated}
              />
              <div className="flex justify-end mt-3">
                <Button
                  disabled={!replyBody.trim() || replyMutation.isPending}
                  onClick={() => requireAuth(() => replyMutation.mutate(replyBody.trim()))}
                >
                  {replyMutation.isPending ? 'Posting…' : 'Post reply'}
                </Button>
              </div>
            </div>
          </section>
        </article>
      </div>

      <Footer />
    </div>
  );
}

export default CommunityThreadPage;
