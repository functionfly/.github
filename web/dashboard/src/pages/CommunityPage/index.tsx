import { useCallback, useEffect, useState } from 'react';
import { usePageTitle } from '@/hooks';
import {
  Chamber,
  CornerBrace,
  PageGrid,
  SealedButton,
  FrameButton,
  TrustSeal,
  AnnotationTag,
  GaugeStrip,
  Gauge,
  Card,
  Input,
} from '@/components/containment';
import {
  MessageSquare,
  TrendingUp,
  Users,
  Flame,
  Clock,
  Eye,
  Loader2,
  Rocket,
  Code,
  LayoutDashboard,
  Bot,
  Plug,
  Shield,
  CreditCard,
  Store,
  Bug,
  Sparkles,
  Lightbulb,
  MessageCircle,
  HelpCircle,
  Search,
  Bell,
  X,
  Send,
  BookmarkIcon,
} from 'lucide-react';
import {
  communityApi,
  displayAuthor,
  type CommunityCategory,
  type CommunitySort,
} from '@/api/community';
import { useAuthStore } from '@/stores/authStore';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link, useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import { PostCard } from '@/components/community/PostCard';
import { CommunityRules } from '@/components/community/CommunityRules';
import './CommunityPage.css';

const CATEGORY_ICON_MAP: Record<string, React.ReactNode> = {
  rocket: <Rocket size={14} />,
  code: <Code size={14} />,
  layout: <LayoutDashboard size={14} />,
  bot: <Bot size={14} />,
  plug: <Plug size={14} />,
  shield: <Shield size={14} />,
  'credit-card': <CreditCard size={14} />,
  store: <Store size={14} />,
  bug: <Bug size={14} />,
  sparkles: <Sparkles size={14} />,
  lightbulb: <Lightbulb size={14} />,
  'message-square': <MessageCircle size={14} />,
  'help-circle': <HelpCircle size={14} />,
};

function getCategoryIcon(icon: string): React.ReactNode {
  return CATEGORY_ICON_MAP[icon] ?? <HelpCircle size={14} />;
}

export function CommunityPage() {
  usePageTitle('Community');
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  const [activeSort, setActiveSort] = useState<CommunitySort>('hot');
  const [activeCategory, setActiveCategory] = useState<string | undefined>(undefined);
  const [activeTag, setActiveTag] = useState<string | undefined>(undefined);
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [showCompose, setShowCompose] = useState(false);
  const [newPostTitle, setNewPostTitle] = useState('');
  const [newPostContent, setNewPostContent] = useState('');
  const [newPostTags, setNewPostTags] = useState('');
  const [selectedCategorySlug, setSelectedCategorySlug] = useState('general');
  const [offset, setOffset] = useState(0);
  const limit = 20;

  // Debounce search
  useEffect(() => {
    const t = setTimeout(() => setDebouncedSearch(searchQuery), 300);
    return () => clearTimeout(t);
  }, [searchQuery]);

  // Reset offset on filter change
  useEffect(() => { setOffset(0); }, [activeSort, activeCategory, activeTag, debouncedSearch]);

  const { data: categoriesData } = useQuery({
    queryKey: ['community-categories'],
    queryFn: () => communityApi.listCategories(),
  });

  const { data: postsData, isLoading: postsLoading } = useQuery({
    queryKey: ['community-posts', activeSort, activeCategory, activeTag, debouncedSearch, offset],
    queryFn: () =>
      communityApi.listPosts({
        sort: activeSort,
        category: activeCategory,
        tag: activeTag,
        q: debouncedSearch || undefined,
        limit,
        offset,
      }),
    refetchInterval: 30000,
  });

  const categories = categoriesData?.categories ?? [];
  const posts = postsData?.posts ?? [];
  const total = postsData?.total ?? posts.length;
  const hasMore = offset + limit < total;

  const { data: unreadData } = useQuery({
    queryKey: ['community-unread-notifications'],
    queryFn: () => communityApi.unreadNotificationsCount(),
    enabled: isAuthenticated,
    refetchInterval: 60000,
  });
  const unreadCount = unreadData?.count ?? 0;

  const voteMutation = useMutation({
    mutationFn: communityApi.vote,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['community-posts'] });
    },
    onError: () => toast.error('Sign in to vote'),
  });

  const createPostMutation = useMutation({
    mutationFn: communityApi.createPost,
    onSuccess: (newPost) => {
      queryClient.invalidateQueries({ queryKey: ['community-posts'] });
      queryClient.invalidateQueries({ queryKey: ['community-categories'] });
      setNewPostTitle('');
      setNewPostContent('');
      setNewPostTags('');
      setShowCompose(false);
      toast.success('Post created');
      navigate(`/community/${newPost.slug || newPost.id}`);
    },
    onError: () => toast.error('Failed to create post'),
  });

  const requireAuth = useCallback((action: () => void) => {
    if (!isAuthenticated) {
      navigate(`/login?redirect=${encodeURIComponent(window.location.pathname)}`);
      return;
    }
    action();
  }, [isAuthenticated, navigate]);

  const handleVote = (postId: string, value: 1 | -1) => {
    requireAuth(() => voteMutation.mutate({ target_type: 'post', target_id: postId, value }));
  };

  const handleSubmitPost = () => {
    if (!newPostTitle.trim() || !newPostContent.trim()) return;
    requireAuth(() =>
      createPostMutation.mutate({
        category_slug: selectedCategorySlug,
        title: newPostTitle.trim(),
        body: newPostContent.trim(),
        tags: newPostTags.split(',').map((t) => t.trim()).filter(Boolean),
      })
    );
  };

  const handleTagClick = (tag: string) => {
    setActiveTag(activeTag === tag ? undefined : tag);
  };

  const totalPosts = total;

  return (
    <div className="sc-community-page">
      <PageGrid />

      {/* Hero */}
      <div className="sc-community-hero">
        <div className="sc-community-hero-inner">
          <Chamber ribs>
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <AnnotationTag primary="COMMUNITY · OPEN FORUM" secondary="v1.0" />
            <div className="sc-community-hero-content">
              <div className="sc-community-hero-badge">
                <TrustSeal size="sm" />
                <span>Community Forum</span>
              </div>
              <h1 className="sc-community-hero-title">FunctionFly Community</h1>
              <p className="sc-community-hero-description">
                Discussions, ideas, and knowledge sharing from the FunctionFly builder community.
              </p>
            </div>
            <GaugeStrip>
              <Gauge data={{ value: totalPosts, label: 'Posts' }} isFirst />
              <Gauge data={{ value: categories.length, label: 'Categories' }} />
              <Gauge data={{ value: 'Live', label: 'Status' }} />
            </GaugeStrip>
          </Chamber>
        </div>
      </div>

      {/* Main Layout */}
      <div className="sc-community-layout">
        {/* Sidebar */}
        <aside className="sc-community-sidebar">
          {/* Community Info */}
          <Card className="sc-community-sidebar-card">
            <div className="sc-community-sidebar-title">Community</div>
            <div className="sc-community-sidebar-info">
              <div className="sc-community-stat">
                <Users size={14} className="sc-community-stat-icon" />
                <span>Open Forum</span>
              </div>
              <div className="sc-community-stat">
                <Eye size={14} className="sc-community-stat-icon" />
                <span className="sc-community-stat-value">{totalPosts}</span>
                <span>posts</span>
              </div>
            </div>
            <div className="sc-community-sidebar-actions">
              <SealedButton
                size="sm"
                iconLeft={<Send size={14} />}
                onClick={() => requireAuth(() => setShowCompose(true))}
              >
                New Post
              </SealedButton>
              {isAuthenticated && (
                <div className="sc-community-sidebar-secondary-actions">
                  <Link to="/community/bookmarks" className="sc-community-sidebar-link">
                    <BookmarkIcon size={14} /> Saved
                  </Link>
                  <Link to="/community/notifications" className="sc-community-sidebar-link">
                    <Bell size={14} />
                    {unreadCount > 0 && <span className="sc-community-notif-badge">{unreadCount}</span>}
                    Alerts
                  </Link>
                </div>
              )}
            </div>
          </Card>

          {/* Categories */}
          {categories.length > 0 && (
            <Card className="sc-community-sidebar-card">
              <div className="sc-community-sidebar-title">Categories</div>
              <div className="sc-community-categories-nav">
                <button
                  className={`sc-community-category-btn ${!activeCategory ? 'active' : ''}`}
                  onClick={() => { setActiveCategory(undefined); setActiveTag(undefined); }}
                >
                  <span className="sc-community-category-icon"><LayoutDashboard size={14} /></span>
                  <span className="sc-community-category-name">All</span>
                </button>
                {categories.map((cat) => (
                  <button
                    key={cat.slug}
                    className={`sc-community-category-btn ${activeCategory === cat.slug ? 'active' : ''}`}
                    onClick={() => { setActiveCategory(cat.slug); setActiveTag(undefined); }}
                  >
                    <span className="sc-community-category-icon">{getCategoryIcon(cat.icon)}</span>
                    <span className="sc-community-category-name">{cat.name}</span>
                    {cat.post_count !== undefined && cat.post_count > 0 && (
                      <span className="sc-community-category-count">{cat.post_count}</span>
                    )}
                  </button>
                ))}
              </div>
            </Card>
          )}

          {/* Rules */}
          <CommunityRules />
        </aside>

        {/* Main Feed */}
        <main className="sc-community-feed">
          {/* Search Bar */}
          <div className="sc-community-search-bar">
            <Search size={16} className="sc-community-search-icon" />
            <input
              type="text"
              className="sc-community-search-input"
              placeholder="Search posts..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
            {searchQuery && (
              <button className="sc-community-search-clear" onClick={() => setSearchQuery('')} type="button">
                <X size={14} />
              </button>
            )}
          </div>

          {/* Compose Area */}
          {showCompose ? (
            <Chamber nested className="sc-community-compose">
              <div className="sc-community-compose-header">
                <h3 className="sc-community-compose-title">New Post</h3>
                <FrameButton size="sm" onClick={() => setShowCompose(false)}>Cancel</FrameButton>
              </div>
              <div className="sc-community-compose-fields">
                <Input
                  placeholder="Post title"
                  value={newPostTitle}
                  onChange={(e) => setNewPostTitle(e.target.value)}
                />
                <select
                  className="sc-community-select"
                  value={selectedCategorySlug}
                  onChange={(e) => setSelectedCategorySlug(e.target.value)}
                >
                  {categories.map((cat) => (
                    <option key={cat.slug} value={cat.slug}>{cat.name}</option>
                  ))}
                </select>
              </div>
              <textarea
                className="sc-community-textarea"
                placeholder="What's on your mind? (Markdown supported)"
                value={newPostContent}
                onChange={(e) => setNewPostContent(e.target.value)}
                rows={6}
              />
              <Input
                placeholder="Tags (comma-separated)"
                value={newPostTags}
                onChange={(e) => setNewPostTags(e.target.value)}
                style={{ marginTop: 'var(--space-2)' }}
              />
              <div className="sc-community-compose-footer">
                <SealedButton
                  size="sm"
                  iconLeft={<Send size={14} />}
                  loading={createPostMutation.isPending}
                  disabled={!newPostTitle.trim() || !newPostContent.trim()}
                  onClick={handleSubmitPost}
                >
                  Post
                </SealedButton>
              </div>
            </Chamber>
          ) : (
            <button
              className="sc-community-compose-prompt"
              onClick={() => requireAuth(() => setShowCompose(true))}
              type="button"
            >
              <div className="sc-community-compose-prompt-avatar">
                <Send size={16} />
              </div>
              <span className="sc-community-compose-prompt-text">What's on your mind?</span>
            </button>
          )}

          {/* Active Filters */}
          {(activeTag || activeCategory || debouncedSearch) && (
            <div className="sc-community-active-filters">
              {activeCategory && (
                <span className="sc-community-filter-pill">
                  Category: {categories.find((c) => c.slug === activeCategory)?.name}
                  <button onClick={() => setActiveCategory(undefined)} type="button"><X size={12} /></button>
                </span>
              )}
              {activeTag && (
                <span className="sc-community-filter-pill">
                  Tag: {activeTag}
                  <button onClick={() => setActiveTag(undefined)} type="button"><X size={12} /></button>
                </span>
              )}
              {debouncedSearch && (
                <span className="sc-community-filter-pill">
                  Search: "{debouncedSearch}"
                  <button onClick={() => setSearchQuery('')} type="button"><X size={12} /></button>
                </span>
              )}
            </div>
          )}

          {/* Sort Tabs */}
          <div className="sc-community-sort-tabs">
            <button
              className={`sc-community-sort-tab ${activeSort === 'hot' ? 'active' : ''}`}
              onClick={() => setActiveSort('hot')}
            >
              <Flame size={14} /> Hot
            </button>
            <button
              className={`sc-community-sort-tab ${activeSort === 'new' ? 'active' : ''}`}
              onClick={() => setActiveSort('new')}
            >
              <Clock size={14} /> New
            </button>
            <button
              className={`sc-community-sort-tab ${activeSort === 'top' ? 'active' : ''}`}
              onClick={() => setActiveSort('top')}
            >
              <TrendingUp size={14} /> Top
            </button>
          </div>

          {/* Posts */}
          {postsLoading ? (
            <Chamber nested className="sc-community-loading">
              <Loader2 size={24} className="sc-community-spinner" />
              <span className="sc-community-loading-text">Loading posts...</span>
            </Chamber>
          ) : posts.length === 0 ? (
            <Chamber nested className="sc-community-empty">
              <MessageSquare size={40} className="sc-community-empty-icon" />
              <p className="sc-community-empty-title">No posts yet</p>
              <p className="sc-community-empty-description">
                {debouncedSearch ? 'No posts match your search.' : 'Be the first to start a discussion!'}
              </p>
            </Chamber>
          ) : (
            <>
              <div className="sc-community-posts">
                {posts.map((post) => (
                  <PostCard
                    key={post.id}
                    post={post}
                    categories={categories}
                    onTagClick={handleTagClick}
                  />
                ))}
              </div>
              {/* Pagination */}
              <div className="sc-community-pagination">
                {offset > 0 && (
                  <FrameButton size="sm" onClick={() => setOffset(Math.max(0, offset - limit))}>
                    Previous
                  </FrameButton>
                )}
                <span className="sc-community-pagination-info">
                  {offset + 1}–{Math.min(offset + limit, total)} of {total}
                </span>
                {hasMore && (
                  <SealedButton size="sm" onClick={() => setOffset(offset + limit)}>
                    Load more
                  </SealedButton>
                )}
              </div>
            </>
          )}
        </main>
      </div>
    </div>
  );
}

export default CommunityPage;
