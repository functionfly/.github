'use client';

import { communityApi, type CommunityPost, type CommunitySort } from '@/api/community';
import { Navbar } from '@/components/common/Navbar';
import { CreatePostDialog } from '@/components/community/CreatePostDialog';
import { PostCard } from '@/components/community/PostCard';
import { MetaTags } from '@/components/seo/MetaTags';
import { Button } from '@/components/ui/button';
import { Footer } from '@/pages/LandingPage/components/Footer';
import { useAuthStore } from '@/stores/authStore';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Loader2, MessageSquare, Plus, Search, Users } from 'lucide-react';
import { useMemo, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';
import './Community.css';

const SORT_OPTIONS: { id: CommunitySort; label: string }[] = [
  { id: 'hot', label: 'Hot' },
  { id: 'new', label: 'New' },
  { id: 'top', label: 'Top' },
];

export function CommunityPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchParams, setSearchParams] = useSearchParams();
  const user = useAuthStore((s) => s.user);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  const category = searchParams.get('category') || '';
  const sort = (searchParams.get('sort') as CommunitySort) || 'hot';
  const q = searchParams.get('q') || '';
  const [searchInput, setSearchInput] = useState(q);
  const [createOpen, setCreateOpen] = useState(false);
  const [votingId, setVotingId] = useState<string | null>(null);

  const { data: categoriesData } = useQuery({
    queryKey: ['community-categories'],
    queryFn: () => communityApi.listCategories(),
    staleTime: 60_000,
  });

  const { data: postsData, isLoading } = useQuery({
    queryKey: ['community-posts', category, sort, q],
    queryFn: () =>
      communityApi.listPosts({
        category: category || undefined,
        sort,
        q: q || undefined,
        limit: 50,
      }),
  });

  const categories = categoriesData?.categories ?? [];
  const posts = postsData?.posts ?? [];

  const createMutation = useMutation({
    mutationFn: communityApi.createPost,
    onSuccess: (post) => {
      queryClient.invalidateQueries({ queryKey: ['community-posts'] });
      toast.success('Thread posted');
      navigate(`/community/${post.id}`);
    },
    onError: () => toast.error('Failed to create thread'),
  });

  const voteMutation = useMutation({
    mutationFn: communityApi.vote,
    onMutate: ({ target_id }) => setVotingId(target_id),
    onSettled: () => setVotingId(null),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['community-posts'] }),
    onError: () => toast.error('Sign in to vote'),
  });

  const handleVote = (postId: string, value: 1 | -1) => {
    if (!isAuthenticated) {
      navigate(`/login?redirect=${encodeURIComponent('/community')}`);
      return;
    }
    voteMutation.mutate({ target_type: 'post', target_id: postId, value });
  };

  const handleCreateClick = () => {
    if (!isAuthenticated) {
      navigate(`/login?redirect=${encodeURIComponent('/community')}`);
      return;
    }
    setCreateOpen(true);
  };

  const updateParams = (updates: Record<string, string | null>) => {
    const next = new URLSearchParams(searchParams);
    for (const [key, val] of Object.entries(updates)) {
      if (val) next.set(key, val);
      else next.delete(key);
    }
    setSearchParams(next, { replace: true });
  };

  const activeCategoryName = useMemo(() => {
    if (!category) return 'All topics';
    return categories.find((c) => c.slug === category)?.name ?? category;
  }, [category, categories]);

  return (
    <div className="community-page">
      <MetaTags
        title="Community Forum | FunctionFly"
        description="Get help from the FunctionFly community. Ask questions, share solutions, and learn from other builders."
      />
      <Navbar />

      <section className="community-hero">
        <div className="max-w-6xl mx-auto px-4 lg:px-6">
          <div className="community-hero-badge">
            <Users className="h-4 w-4" />
            Community
          </div>
          <h1 className="community-hero-title">Community Help Forum</h1>
          <p className="community-hero-description">
            Ask questions, share what worked, and help other builders on FunctionFly. Search
            existing threads before posting — someone may have already solved it.
          </p>
        </div>
      </section>

      <div className="community-layout">
        <aside className="community-sidebar">
          <div className="community-sidebar-card">
            <div className="community-sidebar-title">Topics</div>
            <button
              type="button"
              className={`community-category-btn ${!category ? 'active' : ''}`}
              onClick={() => updateParams({ category: null })}
            >
              <span className="community-category-name">All topics</span>
              <span className="community-category-desc">Browse every community thread</span>
            </button>
            {categories.map((cat) => (
              <button
                key={cat.id}
                type="button"
                className={`community-category-btn ${category === cat.slug ? 'active' : ''}`}
                onClick={() => updateParams({ category: cat.slug })}
              >
                <span className="community-category-name">{cat.name}</span>
                <span className="community-category-desc">{cat.description}</span>
              </button>
            ))}
          </div>
          <Button className="w-full" onClick={handleCreateClick}>
            <Plus className="h-4 w-4 mr-2" />
            New thread
          </Button>
        </aside>

        <main className="community-feed-card">
          <div className="community-toolbar">
            <div className="community-sort-tabs">
              {SORT_OPTIONS.map((opt) => (
                <button
                  key={opt.id}
                  type="button"
                  className={`community-sort-tab ${sort === opt.id ? 'active' : ''}`}
                  onClick={() => updateParams({ sort: opt.id })}
                >
                  {opt.label}
                </button>
              ))}
            </div>
            <form
              className="community-search"
              onSubmit={(e) => {
                e.preventDefault();
                updateParams({ q: searchInput.trim() || null });
              }}
            >
              <Search className="h-4 w-4 text-muted-foreground" />
              <input
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
                placeholder={`Search in ${activeCategoryName.toLowerCase()}…`}
              />
            </form>
          </div>

          {isLoading ? (
            <div className="community-empty">
              <Loader2 className="h-6 w-6 animate-spin mx-auto mb-2" />
              Loading threads…
            </div>
          ) : posts.length === 0 ? (
            <div className="community-empty">
              <MessageSquare className="h-10 w-10 mx-auto mb-3 opacity-40" />
              <p className="font-medium mb-1">No threads yet</p>
              <p className="text-sm mb-4">Start the conversation — ask your first question.</p>
              <Button onClick={handleCreateClick}>Create thread</Button>
            </div>
          ) : (
            posts.map((post: CommunityPost) => (
              <PostCard
                key={post.id}
                post={post}
                onVote={handleVote}
                voting={votingId === post.id}
              />
            ))
          )}
        </main>
      </div>

      <CreatePostDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        categories={categories}
        defaultCategory={category || undefined}
        onSubmit={async (data) => {
          await createMutation.mutateAsync(data);
        }}
      />

      <Footer />
    </div>
  );
}

export default CommunityPage;
