'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { Link } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { BookOpen, Calendar, ArrowRight, User, Tag, Loader2, AlertTriangle, Sparkles, Clock, Bookmark, Search, TrendingUp, Grid3X3, List, Filter } from 'lucide-react';
import { contentApi, BlogPost } from '@/api/content';
import { Footer } from '@/pages/LandingPage/components/Footer';
import {
  calculateReadingTime,
  formatReadingTime,
  getRelativeTime,
  getFeaturedPost,
  getRemainingPosts,
  getAuthorAvatar
} from './utils';
import {
  SearchBar,
  FilterBar,
  FeaturedCarousel,
  BlogCardSkeleton,
  FeaturedPostSkeleton,
  BookmarkButton,
  BlogCategoriesSidebar,
  BlogPostCard
} from '@/components/blog';
import { useInfiniteScroll } from '@/hooks/useInfiniteScroll';
import {
  ShinyButton,
  ParticleBackground
} from "@/components/ui";
import { MetaTags } from '@/components/seo/MetaTags';
import { BlogPageStructuredData } from '@/components/seo/StructuredData';
import { useWebVitals } from '@/hooks/useWebVitals';
import { PublicAnalytics } from '@/components/common/PublicAnalytics';

const BlogPage = () => {
  // Monitor Core Web Vitals
  useWebVitals((metrics) => {
    // Optional: Send to your analytics service
    console.log('Web Vitals:', metrics);
  });

  const [blogPosts, setBlogPosts] = useState<BlogPost[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [hasMore, setHasMore] = useState(false);
  const [offset, setOffset] = useState(0);
  const [hasInitiallyFetched, setHasInitiallyFetched] = useState(false);

  // Search and filter state
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [availableTags, setAvailableTags] = useState<string[]>([]);
  const [categories, setCategories] = useState<{ id: string; title: string; slug: string; order?: number }[]>([]);

  // View mode toggle
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');

  // Enable carousel for multiple featured posts
  const [useCarousel, setUseCarousel] = useState(false);

  // Refs to prevent duplicate/out-of-order updates
  const fetchIdRef = useRef(0);
  const loadingMoreRef = useRef(false);
  const offsetRef = useRef(0);
  offsetRef.current = offset;

  // Infinite scroll
  const { sentinelRef, isNearBottom } = useInfiniteScroll({ rootMargin: '200px' });

  // Extract unique tags from all posts
  useEffect(() => {
    const tags = new Set<string>();
    blogPosts.forEach(post => {
      post.tags.forEach(tag => tags.add(tag));
    });
    setAvailableTags(Array.from(tags).sort());

    // Enable carousel if there are 2+ posts
    setUseCarousel(blogPosts.length >= 2);
  }, [blogPosts]);

  const fetchBlogPosts = useCallback(async (loadMore = false, search?: string, tags?: string[]) => {
    const id = ++fetchIdRef.current;
    const currentOffset = loadMore ? offsetRef.current : 0;

    try {
      if (!loadMore) {
        setLoading(true);
        setOffset(0);
        offsetRef.current = 0;
      }

      const result = await contentApi.getPublishedBlogPosts({
        limit: 10,
        offset: currentOffset,
        tags: tags && tags.length > 0 ? tags : undefined,
      });

      // Ignore result if a newer fetch started
      if (id !== fetchIdRef.current) return;

      // Ensure posts is always an array
      let posts = result.posts || [];

      // Client-side search filtering
      if (search && search.trim()) {
        const query = search.toLowerCase();
        posts = posts.filter(post =>
          post.title.toLowerCase().includes(query) ||
          post.content.toLowerCase().includes(query) ||
          post.excerpt?.toLowerCase().includes(query) ||
          post.author.toLowerCase().includes(query) ||
          post.tags.some(tag => tag.toLowerCase().includes(query))
        );
      }

      if (loadMore) {
        setBlogPosts(prev => [...prev, ...posts]);
        setOffset(prev => prev + posts.length);
        offsetRef.current = currentOffset + posts.length;
      } else {
        setBlogPosts(posts);
        setOffset(posts.length);
        offsetRef.current = posts.length;
      }

      setHasMore(posts.length === 10);
      setError(null);
    } catch (err) {
      if (id !== fetchIdRef.current) return;
      console.error('Failed to fetch blog posts:', err);
      setError('Failed to load blog posts');
    } finally {
      if (id === fetchIdRef.current) {
        setLoading(false);
        loadingMoreRef.current = false;
        if (!loadMore) setHasInitiallyFetched(true);
      }
    }
  }, []);

  // Initial fetch and categories
  useEffect(() => {
    fetchBlogPosts(false, searchQuery, selectedTags);
  }, [fetchBlogPosts]);

  // Fetch categories once on mount
  useEffect(() => {
    contentApi.getPublishedCategories().then(setCategories).catch(() => setCategories([]));
  }, []);

  // Handle search
  const handleSearch = useCallback((query: string) => {
    setSearchQuery(query);
    fetchBlogPosts(false, query, selectedTags);
  }, [selectedTags, fetchBlogPosts]);

  // Handle tag selection
  const handleTagSelect = useCallback((tag: string) => {
    setSelectedTags(prev => {
      const newTags = prev.includes(tag)
        ? prev.filter(t => t !== tag)
        : [...prev, tag];
      fetchBlogPosts(false, searchQuery, newTags);
      return newTags;
    });
  }, [searchQuery, fetchBlogPosts]);

  // Clear all filters
  const handleClearFilters = useCallback(() => {
    setSelectedTags([]);
    setSearchQuery('');
    fetchBlogPosts(false, '', []);
  }, [fetchBlogPosts]);

  // Infinite scroll load more
  useEffect(() => {
    if (!isNearBottom || !hasMore || loading || loadingMoreRef.current) return;
    loadingMoreRef.current = true;
    fetchBlogPosts(true, searchQuery, selectedTags);
  }, [isNearBottom, hasMore, loading, searchQuery, selectedTags, fetchBlogPosts]);

  const loadMore = () => {
    fetchBlogPosts(true, searchQuery, selectedTags);
  };

  const featuredPost = getFeaturedPost(blogPosts);
  const remainingPosts = getRemainingPosts(blogPosts, featuredPost);

  // Show carousel for multiple featured posts
  const showCarousel = useCarousel && blogPosts.length >= 2;

  return (
    <div className="min-h-screen bg-background">
      {/* SEO Meta Tags */}
      <MetaTags
        title="FunctionFly Blog - Serverless Functions & Edge Computing Insights"
        description="Stay updated with the latest in serverless functions, edge computing, AI agents, and cloud infrastructure. Expert insights, tutorials, and best practices from the FunctionFly team."
        keywords={["serverless blog", "edge computing", "cloud infrastructure", "AI agents", "function deployment", "serverless tutorials", "cloud best practices"]}
        url={`${window.location.origin}/blog`}
        type="website"
      />

      {/* Structured Data */}
      <BlogPageStructuredData posts={blogPosts.map(post => ({
        title: post.title,
        excerpt: post.excerpt,
        author: { name: post.author },
        publishedAt: post.published_at || post.created_at,
        slug: post.slug,
        tags: post.tags
      }))} />

      {/* Public Analytics (Hotjar for user behavior) */}
      <PublicAnalytics />

      {/* Animated background */}
      <div className="fixed inset-0 pointer-events-none overflow-hidden">
        <ParticleBackground
          particleCount={20}
          color="rgba(99, 102, 241, 0.08)"
          className="absolute inset-0"
        />
      </div>

      {/* Navigation - text-text-primary for theme-aware contrast (light/dark) */}
      <nav className="blog-page-nav fixed top-0 left-0 right-0 z-50 bg-background/80 backdrop-blur-xl border-b border-border/50">
        <div className="container mx-auto px-4 h-16 flex items-center justify-between">
          <Link to="/" className="flex items-center gap-2 text-text-primary">
            <motion.div
              className="w-8 h-8 rounded-lg bg-gradient-to-br from-brand-500 to-violet-600 flex items-center justify-center"
              whileHover={{ scale: 1.1, rotate: 5 }}
            >
              <Sparkles className="h-4 w-4 text-white" />
            </motion.div>
            <span className="font-bold text-lg">Blog</span>
          </Link>
          <Link to="/">
            <ShinyButton variant="ghost" size="sm" className="text-text-primary">
              Back to Home
            </ShinyButton>
          </Link>
        </div>
      </nav>

      <div className="pt-16 relative z-10">
        {/* Featured Section */}
        {loading && !hasInitiallyFetched ? (
          <section className="pt-8 pb-12">
            <div className="container mx-auto px-4">
              <FeaturedPostSkeleton />
            </div>
          </section>
        ) : showCarousel ? (
          <FeaturedCarousel posts={blogPosts} />
        ) : featuredPost ? (
          <section className="pt-8 pb-12">
            <div className="container mx-auto px-4">
              <FeaturedPostSingle post={featuredPost} />
            </div>
          </section>
        ) : null}

        {/* Blog Posts Grid */}
        <div className="container mx-auto px-4 pb-24">
          {/* Header Section */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            className="text-center mb-12"
          >
            <motion.div
              className="blog-section-pill inline-flex items-center gap-2 px-4 py-2 rounded-full bg-brand-500/10 text-brand-600 dark:text-brand-400 text-sm font-medium mb-4 border border-brand-500/20"
              whileHover={{ scale: 1.05 }}
            >
              <BookOpen className="h-4 w-4" />
              All Articles
            </motion.div>
            <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold tracking-tight mb-4">
              <span className="bg-clip-text text-transparent bg-gradient-to-r from-foreground to-muted-foreground">
                Latest Insights
              </span>
            </h2>
            <p className="text-muted-foreground text-lg max-w-2xl mx-auto">
              Discover tutorials, updates, and stories from our team
            </p>
          </motion.div>

          {/* Search and Filter Section */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.1 }}
            className="mb-12 space-y-8"
          >
            <SearchBar onSearch={handleSearch} placeholder="Search articles..." />
            <FilterBar
              availableTags={availableTags}
              selectedTags={selectedTags}
              onTagSelect={handleTagSelect}
              onClearAll={handleClearFilters}
            />
          </motion.div>

          {/* View Mode Toggle */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="flex items-center justify-end gap-2 mb-6"
          >
            <span className="text-sm text-muted-foreground mr-2">View:</span>
            <button
              onClick={() => setViewMode('grid')}
              className={`
                p-2 rounded-lg transition-all duration-200
                ${viewMode === 'grid'
                  ? 'bg-brand-500 text-white shadow-lg shadow-brand-500/25'
                  : 'bg-muted text-muted-foreground hover:text-foreground'
                }
              `}
            >
              <Grid3X3 className="h-4 w-4" />
            </button>
            <button
              onClick={() => setViewMode('list')}
              className={`
                p-2 rounded-lg transition-all duration-200
                ${viewMode === 'list'
                  ? 'bg-brand-500 text-white shadow-lg shadow-brand-500/25'
                  : 'bg-muted text-muted-foreground hover:text-foreground'
                }
              `}
            >
              <List className="h-4 w-4" />
            </button>
          </motion.div>

          {/* Content */}
          {loading && !hasInitiallyFetched ? (
            <div className={`
              grid gap-6 lg:gap-8
              ${viewMode === 'grid'
                ? 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3'
                : 'grid-cols-1 max-w-3xl mx-auto'
              }
            `}>
              <BlogCardSkeleton count={6} />
            </div>
          ) : error && blogPosts.length === 0 ? (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="text-center py-20"
            >
              <motion.div
                className="inline-flex p-6 rounded-2xl bg-destructive/10 mb-6"
                animate={{ x: [0, -5, 5, -5, 5, 0] }}
                transition={{ duration: 0.5 }}
              >
                <AlertTriangle className="h-16 w-16 text-destructive" />
              </motion.div>
              <h3 className="text-2xl font-bold mb-3">Something went wrong</h3>
              <p className="text-muted-foreground mb-6 max-w-md mx-auto">{error}</p>
              <ShinyButton onClick={() => fetchBlogPosts()} variant="outline" size="lg">
                Try Again
              </ShinyButton>
            </motion.div>
          ) : remainingPosts.length > 0 ? (
            <>
              <div className={`
                grid gap-6 lg:gap-8
                ${viewMode === 'grid'
                  ? 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3'
                  : 'grid-cols-1 max-w-3xl mx-auto'
                }
              `}>
                {remainingPosts.map((post: BlogPost, index: number) => (
                  <BlogPostCard
                    key={post.id}
                    post={post}
                    index={index}
                    variant={viewMode === 'list' ? 'compact' : 'default'}
                  />
                ))}
              </div>

              {/* Load More */}
              {hasMore && (
                <motion.div
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  className="text-center mt-14"
                >
                  <div ref={sentinelRef} className="h-4" />
                  {loading ? (
                    <div className="flex items-center justify-center gap-3">
                      <Loader2 className="h-6 w-6 animate-spin text-brand-500" />
                      <span className="text-muted-foreground">Loading more posts...</span>
                    </div>
                  ) : (
                    <ShinyButton
                      onClick={loadMore}
                      variant="secondary"
                      size="lg"
                      className="min-w-[220px]"
                    >
                      Load More Posts
                      <ArrowRight className="h-4 w-4 ml-2" />
                    </ShinyButton>
                  )}
                </motion.div>
              )}

              {/* End message */}
              {!hasMore && remainingPosts.length > 0 && (
                <motion.div
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  className="text-center mt-14"
                >
                  <p className="text-muted-foreground text-sm">
                    You've reached the end! 🎉
                  </p>
                </motion.div>
              )}
            </>
          ) : (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="text-center py-20"
            >
              <motion.div
                className="inline-flex p-6 rounded-2xl bg-muted/50 mb-6"
                animate={{ y: [0, -10, 0] }}
                transition={{ duration: 2, repeat: Infinity }}
              >
                <Search className="h-16 w-16 text-muted-foreground" />
              </motion.div>
              <h3 className="text-2xl font-bold mb-3">No posts found</h3>
              <p className="text-muted-foreground max-w-md mx-auto mb-6">
                {searchQuery || selectedTags.length > 0
                  ? 'Try adjusting your search or filters to find what you\'re looking for.'
                  : 'Check back soon for new insights and updates.'}
              </p>
              {(searchQuery || selectedTags.length > 0) && (
                <ShinyButton
                  variant="outline"
                  onClick={handleClearFilters}
                  size="lg"
                >
                  Clear all filters
                </ShinyButton>
              )}
            </motion.div>
          )}
        </div>
      </div>

      <Footer />
    </div>
  );
};

// Featured Post Single Component
function FeaturedPostSingle({ post }: { post: BlogPost }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 30 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.6 }}
      className="relative"
    >
      <div className="
        relative overflow-hidden rounded-3xl
        bg-white/5 dark:bg-black/20 backdrop-blur-xl
        border border-white/20 dark:border-white/10
        shadow-[0_25px_80px_-20px_rgba(99,102,241,0.25)]
      ">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-0">
          {/* Image */}
          <div className="relative h-72 lg:h-auto lg:min-h-[450px] overflow-hidden">
            {post.featured_image ? (
              <motion.img
                src={post.featured_image}
                alt={post.title}
                className="w-full h-full object-cover"
                whileHover={{ scale: 1.05 }}
                transition={{ duration: 0.6 }}
              />
            ) : (
              <div className="
                w-full h-full bg-gradient-to-br from-brand-500/20 via-violet-500/20 to-purple-500/20
                flex items-center justify-center
              ">
                <span className="text-8xl">📝</span>
              </div>
            )}
            <div className="absolute inset-0 bg-gradient-to-t from-black/70 via-black/30 to-transparent lg:bg-gradient-to-r lg:from-transparent lg:via-transparent lg:to-black/50" />

            <div className="absolute top-6 left-6">
              <span className="
                inline-flex items-center gap-2 px-4 py-2 rounded-full
                bg-gradient-to-r from-brand-500 to-violet-600
                text-white text-sm font-semibold shadow-lg shadow-brand-500/30
              ">
                <Sparkles className="h-4 w-4" />
                Featured
              </span>
            </div>
          </div>

          {/* Content */}
          <div className="p-8 lg:p-12 flex flex-col justify-center">
            <div className="flex flex-wrap gap-2 mb-4">
              {post.tags.slice(0, 3).map((tag) => (
                <span
                  key={tag}
                  className="blog-category-tag
                    inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full
                    text-xs font-semibold uppercase tracking-wider
                    bg-brand-500/10 text-brand-600 dark:text-brand-400
                    border border-brand-500/20
                  "
                >
                  {tag}
                </span>
              ))}
            </div>

            <h2 className="text-2xl sm:text-3xl lg:text-4xl font-bold mb-4 leading-tight tracking-tight">
              <Link
                to={`/blog/${post.slug}`}
                className="text-foreground hover:text-brand-600 dark:hover:text-brand-400 transition-colors duration-300"
              >
                {post.title}
              </Link>
            </h2>

            <p className="text-lg text-muted-foreground mb-6 leading-relaxed">
              {post.excerpt || post.content.substring(0, 280) + '...'}
            </p>

            <div className="flex items-center gap-4 py-4 border-y border-border/30 mb-6">
              <div className="
                w-14 h-14 rounded-full
                bg-gradient-to-br from-brand-500 to-violet-600
                flex items-center justify-center text-white font-bold text-lg
                shadow-lg shadow-brand-500/25
                ring-3 ring-white dark:ring-gray-800
              ">
                {getAuthorAvatar(post.author)}
              </div>
              <div>
                <p className="font-semibold text-foreground">{post.author}</p>
                <div className="flex items-center gap-3 text-sm text-muted-foreground">
                  <span>{getRelativeTime(post.published_at || post.created_at)}</span>
                  <span className="w-1 h-1 rounded-full bg-muted-foreground" />
                  <span className="flex items-center gap-1">
                    <Clock className="h-4 w-4" />
                    {formatReadingTime(calculateReadingTime(post.content))}
                  </span>
                </div>
              </div>
            </div>

            <Link
              to={`/blog/${post.slug}`}
              className="
                inline-flex items-center gap-3 px-8 py-4 rounded-xl
                bg-gradient-to-r from-brand-600 to-violet-600
                text-white font-semibold text-lg
                shadow-xl shadow-brand-500/25 hover:shadow-brand-500/40
                transition-all duration-300 group w-fit
              "
            >
              Read Full Article
              <ArrowRight className="h-5 w-5 group-hover:translate-x-1 transition-transform" />
            </Link>
          </div>
        </div>
      </div>
    </motion.div>
  );
}

export default BlogPage;
