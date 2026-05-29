import type { CommunityPost } from '@/api/community';
import { displayAuthor, formatRelativeTime } from '@/api/community';
import { VoteControls } from '@/components/community/VoteControls';
import { CheckCircle2, MessageSquare } from 'lucide-react';
import { Link } from 'react-router-dom';

interface PostCardProps {
  post: CommunityPost;
  onVote: (postId: string, value: 1 | -1) => void;
  voting?: boolean;
}

export function PostCard({ post, onVote, voting }: PostCardProps) {
  const preview = post.body.length > 180 ? `${post.body.slice(0, 180)}…` : post.body;

  return (
    <Link to={`/community/${post.id}`} className="community-post-row">
      <VoteControls
        score={post.vote_score}
        userVote={post.user_vote}
        onVote={(value) => onVote(post.id, value)}
        disabled={voting}
      />
      <div className="min-w-0">
        <div className="community-post-meta">
          <span className="community-badge community-badge-category">{post.category_name}</span>
          {post.status === 'solved' && (
            <span className="community-badge community-badge-solved">
              <CheckCircle2 className="h-3 w-3" />
              Solved
            </span>
          )}
          <span>{displayAuthor(post.author)}</span>
          <span>·</span>
          <span>{formatRelativeTime(post.last_activity_at || post.created_at)}</span>
        </div>
        <h3 className="community-post-title">{post.title}</h3>
        <p className="community-post-preview">{preview}</p>
        <div className="community-stats mt-2">
          <span className="inline-flex items-center gap-1">
            <MessageSquare className="h-3.5 w-3.5" />
            {post.reply_count} {post.reply_count === 1 ? 'reply' : 'replies'}
          </span>
          <span>{post.view_count} views</span>
        </div>
      </div>
    </Link>
  );
}
