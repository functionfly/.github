import type { CommunityComment } from '@/api/community';
import { communityApi, displayAuthor, formatRelativeTime, profileUrl } from '@/api/community';
import { VoteControls } from '@/components/community/VoteControls';
import { MarkdownRenderer } from '@/pages/FunctionPage/MarkdownRenderer';
import { CheckCircle2, Pencil, Trash2, X, Send } from 'lucide-react';
import { useMemo, useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { Link } from 'react-router-dom';

interface CommentThreadProps {
  comments: CommunityComment[];
  currentUserId?: string;
  postAuthorId?: string;
  postSlug?: string;
  onVote: (commentId: string, value: 1 | -1) => void;
  onReply: (commentId: string) => void;
  onAccept?: (commentId: string) => void;
  votingId?: string | null;
}

interface CommentNode extends CommunityComment {
  children: CommentNode[];
}

function buildTree(comments: CommunityComment[]): CommentNode[] {
  const byId = new Map<string, CommentNode>();
  const roots: CommentNode[] = [];

  for (const c of comments) {
    byId.set(c.id, { ...c, children: [] });
  }

  for (const c of comments) {
    const node = byId.get(c.id)!;
    if (c.parent_id && byId.has(c.parent_id)) {
      byId.get(c.parent_id)!.children.push(node);
    } else {
      roots.push(node);
    }
  }

  return roots;
}

function CommentItem({
  comment,
  depth,
  currentUserId,
  postAuthorId,
  postSlug,
  onVote,
  onReply,
  onAccept,
  votingId,
}: {
  comment: CommentNode;
  depth: number;
  currentUserId?: string;
  postAuthorId?: string;
  postSlug?: string;
  onVote: (commentId: string, value: 1 | -1) => void;
  onReply: (commentId: string) => void;
  onAccept?: (commentId: string) => void;
  votingId?: string | null;
}) {
  const queryClient = useQueryClient();
  const [isEditing, setIsEditing] = useState(false);
  const [editBody, setEditBody] = useState(comment.body);
  const canAccept = postAuthorId === currentUserId && !comment.is_accepted && onAccept;
  const isOwner = currentUserId === comment.author_id;

  const editMutation = useMutation({
    mutationFn: () => communityApi.updateComment(comment.id, editBody.trim()),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['community-post'] });
      setIsEditing(false);
      toast.success('Comment updated');
    },
    onError: () => toast.error('Failed to update comment'),
  });

  const deleteMutation = useMutation({
    mutationFn: () => communityApi.deleteComment(comment.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['community-post'] });
      toast.success('Comment deleted');
    },
    onError: () => toast.error('Failed to delete comment'),
  });

  return (
    <div className={depth > 0 ? 'community-comment-nested' : undefined}>
      <article className="community-comment">
        <VoteControls
          score={comment.vote_score}
          userVote={comment.user_vote}
          onVote={(value) => onVote(comment.id, value)}
          disabled={votingId === comment.id}
          compact
        />
        <div>
          <div className="community-post-meta">
            <Link to={profileUrl(comment.author)} className="sc-community-avatar-link" style={{ display: 'inline-flex', alignItems: 'center', gap: 'var(--space-2)' }}>
              {comment.author?.avatar_url ? (
                <img src={comment.author.avatar_url} alt="" width={18} height={18} className="sc-community-avatar-img" style={{ width: 18, height: 18, borderRadius: '50%' }} />
              ) : (
                <span className="sc-community-avatar" style={{ width: 18, height: 18, fontSize: 9 }}>
                  {(comment.author?.name || comment.author?.username || '?')[0].toUpperCase()}
                </span>
              )}
              <span>{displayAuthor(comment.author)}</span>
            </Link>
            <span>·</span>
            <span>{formatRelativeTime(comment.created_at)}</span>
            {comment.is_accepted && (
              <span className="community-accepted">
                <CheckCircle2 className="h-3.5 w-3.5" />
                Accepted answer
              </span>
            )}
          </div>
          {isEditing ? (
            <div style={{ marginTop: 'var(--space-2)' }}>
              <textarea
                className="sc-community-textarea"
                value={editBody}
                onChange={(e) => setEditBody(e.target.value)}
                rows={4}
              />
              <div style={{ display: 'flex', gap: 'var(--space-2)', justifyContent: 'flex-end', marginTop: 'var(--space-2)' }}>
                <button type="button" className="community-comment-action" onClick={() => { setIsEditing(false); setEditBody(comment.body); }}>
                  <X size={12} /> Cancel
                </button>
                <button type="button" className="community-comment-action" onClick={() => editMutation.mutate()} style={{ color: 'var(--status-ok)' }}>
                  <Send size={12} /> Save
                </button>
              </div>
            </div>
          ) : (
            <div className="community-comment-body">
              <MarkdownRenderer content={comment.body} />
            </div>
          )}
          <div className="community-comment-actions">
            <button type="button" className="community-comment-action" onClick={() => onReply(comment.id)}>
              Reply
            </button>
            {canAccept && (
              <button type="button" className="community-comment-action" onClick={() => onAccept!(comment.id)}>
                Mark as solution
              </button>
            )}
            {isOwner && !isEditing && (
              <>
                <button type="button" className="community-comment-action" onClick={() => setIsEditing(true)}>
                  <Pencil size={12} /> Edit
                </button>
                <button
                  type="button"
                  className="community-comment-action"
                  style={{ color: 'var(--status-revoked)' }}
                  onClick={() => { if (confirm('Delete this comment?')) deleteMutation.mutate(); }}
                >
                  <Trash2 size={12} /> Delete
                </button>
              </>
            )}
          </div>
        </div>
      </article>
      {comment.children.map((child) => (
        <CommentItem
          key={child.id}
          comment={child}
          depth={depth + 1}
          currentUserId={currentUserId}
          postAuthorId={postAuthorId}
          postSlug={postSlug}
          onVote={onVote}
          onReply={onReply}
          onAccept={onAccept}
          votingId={votingId}
        />
      ))}
    </div>
  );
}

export function CommentThread({
  comments,
  currentUserId,
  postAuthorId,
  postSlug,
  onVote,
  onReply,
  onAccept,
  votingId,
}: CommentThreadProps) {
  const tree = useMemo(() => buildTree(comments), [comments]);

  if (tree.length === 0) {
    return (
      <div className="community-empty py-8">
        No replies yet. Be the first to help!
      </div>
    );
  }

  return (
    <div>
      {tree.map((comment) => (
        <CommentItem
          key={comment.id}
          comment={comment}
          depth={0}
          currentUserId={currentUserId}
          postAuthorId={postAuthorId}
          postSlug={postSlug}
          onVote={onVote}
          onReply={onReply}
          onAccept={onAccept}
          votingId={votingId}
        />
      ))}
    </div>
  );
}
