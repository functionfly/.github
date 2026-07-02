import { useState } from 'react';
import {
  MessageSquare,
  Share2,
  Bookmark,
  BookmarkCheck,
  MoreHorizontal,
  ArrowUp,
  ArrowDown,
  Eye,
  CheckCircle2,
  Pencil,
  Trash2,
  X,
  Send,
} from 'lucide-react';
import { Link, useNavigate } from 'react-router-dom';
import {
  communityApi,
  displayAuthor,
  formatRelativeTime,
  profileUrl,
  type CommunityPost,
  type CommunityCategory,
} from '@/api/community';
import { useAuthStore } from '@/stores/authStore';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { MarkdownRenderer } from '@/pages/FunctionPage/MarkdownRenderer';
import {
  Chamber,
  Card,
  SealedButton,
  FrameButton,
  StatusPill,
  Input,
} from '@/components/containment';

function formatViewCount(n: number): string {
  if (n >= 1000) return `${(n / 1000).toFixed(1)}K`;
  return String(n);
}

interface PostCardProps {
  post: CommunityPost;
  categories?: CommunityCategory[];
  onTagClick?: (tag: string) => void;
  compact?: boolean;
}

export function PostCard({ post, categories, onTagClick, compact }: PostCardProps) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const currentUser = useAuthStore((s) => s.user);

  const [isEditing, setIsEditing] = useState(false);
  const [editTitle, setEditTitle] = useState(post.title);
  const [editBody, setEditBody] = useState(post.body);
  const [editTags, setEditTags] = useState(post.tags?.join(', ') || '');
  const [showMenu, setShowMenu] = useState(false);
  const [bookmarked, setBookmarked] = useState(false);

  const userVote = post.user_vote ?? 0;
  const authorInitial = (post.author?.name || post.author?.username || '?')[0].toUpperCase();
  const isOwner = currentUser?.id === post.author_id;
  const hasAcceptedAnswer = !!post.accepted_comment_id;

  const requireAuth = (action: () => void) => {
    if (!isAuthenticated) {
      navigate(`/login?redirect=${encodeURIComponent(window.location.pathname)}`);
      return;
    }
    action();
  };

  const voteMutation = useMutation({
    mutationFn: communityApi.vote,
    onMutate: async (variables) => {
      await queryClient.cancelQueries({ queryKey: ['community-posts'] });
      const previous = queryClient.getQueriesData({ queryKey: ['community-posts'] });
      queryClient.setQueriesData({ queryKey: ['community-posts'] }, (old: any) => {
        if (!old?.posts) return old;
        return {
          ...old,
          posts: old.posts.map((p: any) =>
            p.id === post.id
              ? {
                  ...p,
                  vote_score: p.vote_score - (userVote || 0) + variables.value,
                  user_vote: variables.value,
                }
              : p
          ),
        };
      });
      return { previous };
    },
    onError: (_err, _vars, context) => {
      if (context?.previous) {
        context.previous.forEach(([key, data]) => queryClient.setQueryData(key, data));
      }
      toast.error('Sign in to vote');
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['community-posts'] });
      queryClient.invalidateQueries({ queryKey: ['community-post'] });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => communityApi.deletePost(post.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['community-posts'] });
      toast.success('Post deleted');
    },
    onError: () => toast.error('Failed to delete post'),
  });

  const editMutation = useMutation({
    mutationFn: () =>
      communityApi.updatePost(post.id, {
        title: editTitle.trim(),
        body: editBody.trim(),
        tags: editTags.split(',').map((t) => t.trim()).filter(Boolean),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['community-posts'] });
      setIsEditing(false);
      toast.success('Post updated');
    },
    onError: () => toast.error('Failed to update post'),
  });

  const bookmarkMutation = useMutation({
    mutationFn: () =>
      bookmarked ? communityApi.unbookmarkPost(post.id) : communityApi.bookmarkPost(post.id),
    onSuccess: () => {
      setBookmarked(!bookmarked);
      toast.success(bookmarked ? 'Bookmark removed' : 'Post saved');
    },
    onError: () => toast.error('Failed to update bookmark'),
  });

  const handleVote = (value: 1 | -1) => {
    requireAuth(() => voteMutation.mutate({ target_type: 'post', target_id: post.id, value }));
  };

  const handleShare = async () => {
    const url = `${window.location.origin}/community/${post.slug || post.id}`;
    try {
      await navigator.clipboard.writeText(url);
      toast.success('Link copied to clipboard');
    } catch {
      toast.error('Failed to copy link');
    }
  };

  if (isEditing) {
    return (
      <Card className="sc-community-post-card">
        <div className="sc-community-post-content" style={{ padding: 'var(--space-5)' }}>
          <div className="sc-community-compose-fields">
            <Input
              placeholder="Post title"
              value={editTitle}
              onChange={(e) => setEditTitle(e.target.value)}
            />
            <Input
              placeholder="Tags (comma-separated)"
              value={editTags}
              onChange={(e) => setEditTags(e.target.value)}
              style={{ marginTop: 'var(--space-2)' }}
            />
          </div>
          <textarea
            className="sc-community-textarea"
            value={editBody}
            onChange={(e) => setEditBody(e.target.value)}
            rows={6}
            style={{ marginTop: 'var(--space-2)' }}
          />
          <div style={{ display: 'flex', gap: 'var(--space-2)', justifyContent: 'flex-end', marginTop: 'var(--space-3)' }}>
            <FrameButton size="sm" onClick={() => { setIsEditing(false); setEditTitle(post.title); setEditBody(post.body); }}>
              <X size={14} /> Cancel
            </FrameButton>
            <SealedButton size="sm" iconLeft={<Send size={14} />} loading={editMutation.isPending} onClick={() => editMutation.mutate()}>
              Save
            </SealedButton>
          </div>
        </div>
      </Card>
    );
  }

  return (
    <Card className="sc-community-post-card">
      <div className="sc-community-post-layout">
        {/* Vote Column */}
        <div className="sc-community-post-votes">
          <button
            className={`sc-community-vote-btn ${userVote === 1 ? 'active' : ''}`}
            onClick={() => handleVote(1)}
            aria-label="Upvote"
          >
            <ArrowUp size={16} />
          </button>
          <span className="sc-community-vote-count">{post.vote_score}</span>
          <button
            className={`sc-community-vote-btn ${userVote === -1 ? 'active' : ''}`}
            onClick={() => handleVote(-1)}
            aria-label="Downvote"
          >
            <ArrowDown size={16} />
          </button>
        </div>

        {/* Content Column */}
        <div className="sc-community-post-content">
          {/* Post Header */}
          <div className="sc-community-post-header">
            <Link to={profileUrl(post.author)} className="sc-community-avatar-link">
              {post.author?.avatar_url ? (
                <img src={post.author.avatar_url} alt="" width={22} height={22} className="sc-community-avatar-img" style={{ borderRadius: '50%' }} />
              ) : (
                <div className="sc-community-avatar">{authorInitial}</div>
              )}
            </Link>
            <Link to={profileUrl(post.author)} className="sc-community-post-author">
              {displayAuthor(post.author)}
            </Link>
            {post.category_name && (
              <>
                <span className="sc-community-meta-sep">·</span>
                <span className="sc-community-post-category">{post.category_name}</span>
              </>
            )}
            <span className="sc-community-meta-sep">·</span>
            <span className="sc-community-meta">{formatRelativeTime(post.created_at)}</span>
            {post.status === 'solved' && <StatusPill status="live" label="Solved" />}
            {hasAcceptedAnswer && !compact && (
              <StatusPill status="live" label="Has Answer" />
            )}
            {post.is_pinned && <StatusPill status="pending" label="Pinned" />}
          </div>

          {/* Post Title */}
          <Link to={`/community/${post.slug || post.id}`} className="sc-community-title-link">
            <h3 className="sc-community-title">{post.title}</h3>
          </Link>

          {/* Post Body — markdown for full view, plain preview for compact */}
          {compact ? (
            <p className="sc-community-post-body">{post.body}</p>
          ) : (
            <div className="sc-community-post-body-preview">
              <MarkdownRenderer content={post.body.length > 400 ? `${post.body.slice(0, 400)}…` : post.body} />
            </div>
          )}

          {/* Tags */}
          {post.tags && post.tags.length > 0 && (
            <div className="sc-community-tags">
              {post.tags.map((tag) => (
                <button
                  key={tag}
                  className="sc-community-tag"
                  onClick={() => onTagClick?.(tag)}
                  type="button"
                >
                  {tag}
                </button>
              ))}
            </div>
          )}

          {/* Post Actions */}
          <div className="sc-community-post-actions">
            <Link
              to={`/community/${post.slug || post.id}`}
              className="sc-community-action-btn"
            >
              <MessageSquare size={14} />
              <span>{post.reply_count} Comments</span>
            </Link>
            <button className="sc-community-action-btn" type="button">
              <Eye size={14} />
              <span>{formatViewCount(post.view_count)}</span>
            </button>
            <button className="sc-community-action-btn" type="button" onClick={handleShare}>
              <Share2 size={14} />
              <span>Share</span>
            </button>
            <button
              className="sc-community-action-btn"
              type="button"
              onClick={() => requireAuth(() => bookmarkMutation.mutate())}
            >
              {bookmarked ? <BookmarkCheck size={14} /> : <Bookmark size={14} />}
              <span>{bookmarked ? 'Saved' : 'Save'}</span>
            </button>

            {/* Owner actions */}
            {isOwner && (
              <div className="sc-community-owner-actions">
                <button
                  className="sc-community-action-btn"
                  type="button"
                  onClick={() => setIsEditing(true)}
                >
                  <Pencil size={14} />
                  <span>Edit</span>
                </button>
                <button
                  className="sc-community-action-btn sc-community-action-btn--danger"
                  type="button"
                  onClick={() => {
                    if (confirm('Delete this post?')) deleteMutation.mutate();
                  }}
                >
                  <Trash2 size={14} />
                  <span>Delete</span>
                </button>
              </div>
            )}

            <button
              className="sc-community-action-btn sc-community-action-btn--icon"
              type="button"
              onClick={() => setShowMenu(!showMenu)}
            >
              <MoreHorizontal size={14} />
            </button>
          </div>
        </div>
      </div>
    </Card>
  );
}
