import { usePageTitle } from '@/hooks';
import { Chamber, PageGrid } from '@/components/containment';
import { Loader2, BookmarkIcon, ArrowLeft } from 'lucide-react';
import { communityApi } from '@/api/community';
import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { PostCard } from '@/components/community/PostCard';
import { useState } from 'react';

export function CommunityBookmarksPage() {
  usePageTitle('Saved Posts');
  const [offset, setOffset] = useState(0);
  const limit = 20;

  const { data, isLoading } = useQuery({
    queryKey: ['community-bookmarks', offset],
    queryFn: () => communityApi.listBookmarks(limit, offset),
  });

  const posts = data?.posts ?? [];
  const total = data?.total ?? 0;
  const hasMore = offset + limit < total;

  return (
    <div className="sc-community-page">
      <PageGrid />
      <div className="sc-thread-container">
        <Link to="/community" className="sc-thread-back">
          <ArrowLeft size={14} />
          <span>Back to community</span>
        </Link>

        <Chamber>
          <div style={{ padding: 'var(--space-5)' }}>
            <h1 style={{ fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 600, color: 'var(--text)', margin: 0 }}>
              <BookmarkIcon size={20} style={{ display: 'inline', verticalAlign: 'middle', marginRight: 8 }} />
              Saved Posts
            </h1>
            <p style={{ fontFamily: 'var(--font-body)', fontSize: 13, color: 'var(--text-dim)', marginTop: 'var(--space-2)' }}>
              {total} saved {total === 1 ? 'post' : 'posts'}
            </p>
          </div>
        </Chamber>

        {isLoading ? (
          <Chamber nested className="sc-community-loading">
            <Loader2 size={24} className="sc-community-spinner" />
            <span className="sc-community-loading-text">Loading saved posts...</span>
          </Chamber>
        ) : posts.length === 0 ? (
          <Chamber nested className="sc-community-empty">
            <BookmarkIcon size={40} className="sc-community-empty-icon" />
            <p className="sc-community-empty-title">No saved posts</p>
            <p className="sc-community-empty-description">Bookmark posts to find them later.</p>
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

export default CommunityBookmarksPage;
