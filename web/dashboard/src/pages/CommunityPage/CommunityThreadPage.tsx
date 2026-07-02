'use client';

import {
  communityApi,
  displayAuthor,
  formatRelativeTime,
  profileUrl,
  type CommunityComment,
} from '@/api/community';
import { CommentThread } from '@/components/community/CommentThread';
import { VoteControls } from '@/components/community/VoteControls';
import { MetaTags } from '@/components/seo/MetaTags';
import { MarkdownRenderer } from '@/pages/FunctionPage/MarkdownRenderer';
import {
  Chamber,
  CornerBrace,
  PageGrid,
  SealedButton,
  FrameButton,
  StatusPill,
  AnnotationTag,
} from '@/components/containment';
import { useAuthStore } from '@/stores/authStore';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  ArrowLeft,
  CheckCircle2,
  Loader2,
  MessageSquare,
  Send,
  Eye,
  Clock,
} from 'lucide-react';
import { useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import './Community.css';

export function CommunityThreadPage() {
  const { postId: slugOrId } = useParams<{ postId: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const user = useAuthStore((s) => s.user);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  const [replyBody, setReplyBody] = useState('');
  const [replyToId, setReplyToId] = useState<string | null>(null);
  const [votingId, setVotingId] = useState<string | null>(null);

  const { data, isLoading, error } = useQuery({
    queryKey: ['community-post', slugOrId],
    queryFn: () => communityApi.getPost(slugOrId!),
    enabled: !!slugOrId,
    refetchInterval: 30000,
  });

  const post = data?.post;
  const comments = data?.comments ?? [];
  const postId = post?.id;

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ['community-post', slugOrId] });
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
    onMutate: async (variables) => {
      setVotingId(variables.target_id);
      await queryClient.cancelQueries({ queryKey: ['community-post', slugOrId] });
      const key = ['community-post', slugOrId];
      const previous = queryClient.getQueryData<any>(key);
      queryClient.setQueryData(key, (old: any) => {
        if (!old) return old;
        if (variables.target_type === 'post') {
          return {
            ...old,
            post: {
              ...old.post,
              vote_score: old.post.vote_score - ((old.post.user_vote as number) || 0) + variables.value,
              user_vote: variables.value,
            },
          };
        }
        return {
          ...old,
          comments: old.comments?.map((c: any) =>
            c.id === variables.target_id
              ? { ...c, vote_score: c.vote_score - ((c.user_vote as number) || 0) + variables.value, user_vote: variables.value }
              : c
          ),
        };
      });
      return { previous };
    },
    onError: (_err, _vars, context) => {
      if (context?.previous) {
        queryClient.setQueryData(['community-post', slugOrId], context.previous);
      }
      toast.error('Sign in to vote');
    },
    onSettled: () => {
      setVotingId(null);
      invalidate();
    },
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
      <div className="sc-community-page">
        <PageGrid />
        <Chamber nested className="sc-community-loading">
          <Loader2 size={24} className="sc-community-spinner" />
          <span className="sc-community-loading-text">Loading thread...</span>
        </Chamber>
      </div>
    );
  }

  if (error || !post) {
    return (
      <div className="sc-community-page">
        <PageGrid />
        <Chamber nested className="sc-community-empty">
          <MessageSquare size={40} className="sc-community-empty-icon" />
          <p className="sc-community-empty-title">Thread not found</p>
          <p className="sc-community-empty-description">
            This thread may have been removed or the link is incorrect.
          </p>
          <FrameButton
            size="sm"
            iconLeft={<ArrowLeft size={14} />}
            onClick={() => navigate('/community')}
            className="sc-community-mt-4"
          >
            Back to community
          </FrameButton>
        </Chamber>
      </div>
    );
  }

  return (
    <div className="sc-community-page">
      <PageGrid />
      <MetaTags title={`${post.title} | Community`} description={post.body.slice(0, 160)} />

      <div className="sc-thread-container">
        {/* Back link */}
        <Link to="/community" className="sc-thread-back">
          <ArrowLeft size={14} />
          <span>Back to community</span>
        </Link>

        {/* Thread Card */}
        <Chamber>
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <AnnotationTag
            primary={post.category?.name?.toUpperCase() || 'THREAD'}
            secondary={`${post.reply_count} replies`}
          />

          {/* Thread Header */}
          <div className="sc-thread-header">
            <div className="sc-thread-header-main">
              <VoteControls
                score={post.vote_score}
                userVote={post.user_vote}
                onVote={handlePostVote}
                disabled={votingId === post.id}
              />
              <div className="sc-thread-header-content">
                <div className="sc-thread-meta">
                  {post.category?.name && (
                    <span className="sc-thread-category">{post.category.name}</span>
                  )}
                  {post.status === 'solved' && <StatusPill status="live" label="Solved" />}
                  {post.is_pinned && <StatusPill status="pending" label="Pinned" />}
                </div>
                <h1 className="sc-thread-title">{post.title}</h1>
                <div className="sc-thread-info">
                  <div className="sc-thread-author">
                    <Link to={profileUrl(post.author)} className="sc-community-avatar-link">
                      {post.author?.avatar_url ? (
                        <img src={post.author.avatar_url} alt="" width={22} height={22} className="sc-community-avatar-img" style={{ borderRadius: '50%' }} />
                      ) : (
                        <div className="sc-community-avatar">
                          {(post.author?.name || post.author?.username || '?')[0].toUpperCase()}
                        </div>
                      )}
                    </Link>
                    <Link to={profileUrl(post.author)} className="sc-community-post-author">
                      {displayAuthor(post.author)}
                    </Link>
                  </div>
                  <span className="sc-community-meta-sep">·</span>
                  <span className="sc-thread-stat">
                    <Clock size={12} /> {formatRelativeTime(post.created_at)}
                  </span>
                  <span className="sc-community-meta-sep">·</span>
                  <span className="sc-thread-stat">
                    <MessageSquare size={12} /> {post.reply_count} replies
                  </span>
                  <span className="sc-community-meta-sep">·</span>
                  <span className="sc-thread-stat">
                    <Eye size={12} /> {post.view_count} views
                  </span>
                </div>
              </div>
            </div>
          </div>

          {/* Thread Body — Markdown */}
          <div className="sc-thread-body">
            <MarkdownRenderer content={post.body} />
          </div>

          {/* Tags */}
          {post.tags && post.tags.length > 0 && (
            <div className="sc-community-tags" style={{ marginBottom: 'var(--space-5)' }}>
              {post.tags.map((tag) => (
                <Link key={tag} to={`/community?tag=${encodeURIComponent(tag)}`} className="sc-community-tag">
                  {tag}
                </Link>
              ))}
            </div>
          )}

          {/* Accepted Answer Banner */}
          {post.status === 'solved' && (
            <div className="sc-thread-solved-banner">
              <CheckCircle2 size={16} />
              <span>This thread has been marked as solved</span>
            </div>
          )}

          {/* Comments Section */}
          <section className="sc-thread-comments">
            <h2 className="sc-thread-comments-header">
              <MessageSquare size={14} />
              {comments.length} {comments.length === 1 ? 'Reply' : 'Replies'}
            </h2>

            <CommentThread
              comments={comments as CommunityComment[]}
              currentUserId={user?.id}
              postAuthorId={post.author_id}
              postSlug={post.slug}
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

            {/* Reply Box */}
            <div className="sc-thread-reply-box">
              {replyToId && (
                <p className="sc-thread-reply-context">
                  Replying to a comment ·{' '}
                  <button type="button" className="sc-thread-reply-cancel" onClick={() => setReplyToId(null)}>
                    cancel
                  </button>
                </p>
              )}
              <textarea
                id="community-reply"
                className="sc-community-textarea"
                value={replyBody}
                onChange={(e) => setReplyBody(e.target.value)}
                placeholder={
                  isAuthenticated ? 'Share your answer or ask a follow-up… (Markdown supported)' : 'Sign in to reply'
                }
                disabled={!isAuthenticated}
                rows={4}
              />
              <div className="sc-thread-reply-footer">
                <SealedButton
                  size="sm"
                  iconLeft={<Send size={14} />}
                  loading={replyMutation.isPending}
                  disabled={!replyBody.trim()}
                  onClick={() => requireAuth(() => replyMutation.mutate(replyBody.trim()))}
                >
                  Post reply
                </SealedButton>
              </div>
            </div>
          </section>
        </Chamber>
      </div>
    </div>
  );
}

export default CommunityThreadPage;
