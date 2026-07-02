import { usePageTitle } from '@/hooks';
import {
  Chamber,
  CornerBrace,
  PageGrid,
  Card,
  AnnotationTag,
  GaugeStrip,
  Gauge,
} from '@/components/containment';
import { Loader2, MessageSquare, ArrowLeft, Users } from 'lucide-react';
import { communityApi, displayAuthor, type CommunityPost } from '@/api/community';
import { useQuery } from '@tanstack/react-query';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { PostCard } from '@/components/community/PostCard';
import { useState } from 'react';

export function CommunityUserProfilePage() {
  const { userId } = useParams<{ userId: string }>();
  const navigate = useNavigate();
  usePageTitle('Community Profile');

  const [offset, setOffset] = useState(0);
  const limit = 20;

  const { data, isLoading } = useQuery({
    queryKey: ['community-user-posts', userId, offset],
    queryFn: () => communityApi.listPostsByAuthor(userId!, limit, offset),
    enabled: !!userId,
  });

  const posts = data?.posts ?? [];
  const total = data?.total ?? 0;
  const hasMore = offset + limit < total;
  const authorName = posts[0]?.author ? displayAuthor(posts[0].author) : 'User';

  if (isLoading) {
    return (
      <div className="sc-community-page">
        <PageGrid />
        <Chamber nested className="sc-community-loading">
          <Loader2 size={24} className="sc-community-spinner" />
          <span className="sc-community-loading-text">Loading profile...</span>
        </Chamber>
      </div>
    );
  }

  return (
    <div className="sc-community-page">
      <PageGrid />
      <div className="sc-thread-container">
        <Link to="/community" className="sc-thread-back">
          <ArrowLeft size={14} />
          <span>Back to community</span>
        </Link>

        <Chamber>
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <AnnotationTag primary="COMMUNITY PROFILE" secondary={`${total} posts`} />
          <div style={{ padding: 'var(--space-4) 0' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)', marginBottom: 'var(--space-4)' }}>
              <div className="sc-community-avatar" style={{ width: 48, height: 48, fontSize: 20 }}>
                {authorName[0]?.toUpperCase() || '?'}
              </div>
              <div>
                <h1 style={{ fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 600, color: 'var(--text)', margin: 0 }}>
                  {authorName}
                </h1>
                <p style={{ fontFamily: 'var(--font-body)', fontSize: 13, color: 'var(--text-dim)', margin: 0 }}>
                  {total} community {total === 1 ? 'post' : 'posts'}
                </p>
              </div>
            </div>
          </div>
        </Chamber>

        {posts.length === 0 ? (
          <Chamber nested className="sc-community-empty">
            <Users size={40} className="sc-community-empty-icon" />
            <p className="sc-community-empty-title">No posts yet</p>
            <p className="sc-community-empty-description">This user hasn't posted anything.</p>
          </Chamber>
        ) : (
          <>
            <div className="sc-community-posts">
              {posts.map((post) => (
                <PostCard key={post.id} post={post} compact />
              ))}
            </div>
            <div className="sc-community-pagination">
              {offset > 0 && (
                <button className="sc-community-sort-tab" onClick={() => setOffset(Math.max(0, offset - limit))}>
                  Previous
                </button>
              )}
              <span className="sc-community-pagination-info">
                {offset + 1}–{Math.min(offset + limit, total)} of {total}
              </span>
              {hasMore && (
                <button className="sc-community-sort-tab active" onClick={() => setOffset(offset + limit)}>
                  Load more
                </button>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  );
}

export default CommunityUserProfilePage;
