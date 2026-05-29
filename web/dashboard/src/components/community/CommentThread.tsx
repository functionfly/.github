import type { CommunityComment } from '@/api/community';
import { displayAuthor, formatRelativeTime } from '@/api/community';
import { VoteControls } from '@/components/community/VoteControls';
import { CheckCircle2 } from 'lucide-react';
import { useMemo } from 'react';

interface CommentThreadProps {
  comments: CommunityComment[];
  currentUserId?: string;
  postAuthorId?: string;
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
  onVote,
  onReply,
  onAccept,
  votingId,
}: {
  comment: CommentNode;
  depth: number;
  currentUserId?: string;
  postAuthorId?: string;
  onVote: (commentId: string, value: 1 | -1) => void;
  onReply: (commentId: string) => void;
  onAccept?: (commentId: string) => void;
  votingId?: string | null;
}) {
  const canAccept = postAuthorId === currentUserId && !comment.is_accepted && onAccept;

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
            <span>{displayAuthor(comment.author)}</span>
            <span>·</span>
            <span>{formatRelativeTime(comment.created_at)}</span>
            {comment.is_accepted && (
              <span className="community-accepted">
                <CheckCircle2 className="h-3.5 w-3.5" />
                Accepted answer
              </span>
            )}
          </div>
          <div className="community-comment-body">{comment.body}</div>
          <div className="community-comment-actions">
            <button type="button" className="community-comment-action" onClick={() => onReply(comment.id)}>
              Reply
            </button>
            {canAccept && (
              <button type="button" className="community-comment-action" onClick={() => onAccept!(comment.id)}>
                Mark as solution
              </button>
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
          onVote={onVote}
          onReply={onReply}
          onAccept={onAccept}
          votingId={votingId}
        />
      ))}
    </div>
  );
}
