'use client';

import { useState, useEffect, useCallback, useMemo } from 'react';
import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Navbar } from '@/components/common/Navbar';
import { BookOpen, Calendar, ArrowRight, User, Tag, Loader2, AlertTriangle, Sparkles, Clock, Bookmark } from 'lucide-react';
import { contentApi, BlogPost } from '@/api/content';
import {
  calculateReadingTime,
  formatReadingTime,
  getRelativeTime,
  getFeaturedPost,
  getRemainingPosts,
  getAuthorAvatar
} from './utils';
import { SearchBar, FilterBar, FeaturedCarousel, BlogCardSkeleton, FeaturedPostSkeleton, BookmarkButton } from '@/components/blog';
import { useInfiniteScroll } from '@/hooks/useInfiniteScroll';

const BlogPage = () => {
  const [blogPosts, setBlogPosts] = useState<BlogPost[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [hasMore, setHasMore] = useState(false);
  const [offset, setOffset] = useState(0);

  // Search and filter state
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [availableTags, setAvailableTags] = useState<string[]>([]);
  // Categories from admin (shown on blog page)
  const [categories, setCategories] = useState<{ id: string; title: string; slug: string; order?: number }[]>([]);

  // Enable carousel for multiple featured posts
  const [useCarousel, setUseCarousel] = useState(false);

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

  const fetchBlogPosts = async (loadMore = false, search?: string, tags?: string[]) => {
    try {
      if (!loadMore) {
        setLoading(true);
        setOffset(0);
      }

      const result = await contentApi.getPublishedBlogPosts({
        limit: 10,
        offset: loadMore ? offset : 0,
        tags: tags && tags.length > 0 ? tags : undefined,
      });

      // Ensure posts is always an array
      let posts = result.posts || [];

      // Client-side search filtering (if API doesn't support search)
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
      } else {
        setBlogPosts(posts);
        setOffset(posts.length);
      }

      setHasMore(posts.length === 10);
    } catch (err) {
      console.error('Failed to fetch blog posts:', err);
      setError('Failed to load blog posts');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchBlogPosts(false, searchQuery, selectedTags);
  }, []);

  // Fetch categories for filter display (admin-created categories)
  useEffect(() => {
    contentApi.getPublishedCategories().then(setCategories).catch(() => setCategories([]));
  }, []);

  // Handle search
  const handleSearch = useCallback((query: string) => {
    setSearchQuery(query);
    fetchBlogPosts(false, query, selectedTags);
  }, [selectedTags]);

  // Handle tag selection
  const handleTagSelect = useCallback((tag: string) => {
    setSelectedTags(prev => {
      const newTags = prev.includes(tag)
        ? prev.filter(t => t !== tag)
        : [...prev, tag];
      fetchBlogPosts(false, searchQuery, newTags);
      return newTags;
    });
  }, [searchQuery]);

  // Clear all filters
  const handleClearFilters = useCallback(() => {
    setSelectedTags([]);
    setSearchQuery('');
    fetchBlogPosts(false, '', []);
  }, []);

  // Infinite scroll load more
  useEffect(() => {
    if (isNearBottom && hasMore && !loading) {
      fetchBlogPosts(true, searchQuery, selectedTags);
    }
  }, [isNearBottom, hasMore, loading, searchQuery, selectedTags]);

  const loadMore = () => {
    fetchBlogPosts(true, searchQuery, selectedTags);
  };

  if (loading && blogPosts.length === 0) {
    return (
      <div className="min-h-screen bg-background">
        <Navbar variant="landing" />
        <div className="pt-16">
          <FeaturedPostSkeleton />
          <div className="container mx-auto px-4 pb-24">
            <div className="mb-14 text-center">
              <div className="h-10 w-64 mx-auto bg-muted/50 rounded animate-pulse mb-3" />
              <div className="h-5 w-96 mx-auto bg-muted/50 rounded animate-pulse" />
            </div>
            <BlogCardSkeleton count={6} />
          </div>
        </div>
      </div>
    );
  }

  if (error && blogPosts.length === 0) {
    return (
      <div className="min-h-screen bg-background">
        <Navbar variant="landing" />
        <div className="flex items-center justify-center min-h-[60vh] pt-24">
          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            className="text-center max-w-md mx-auto px-4"
          >
            <motion.div
              initial={{ scale: 0 }}
              animate={{ scale: 1 }}
              transition={{ delay: 0.2, type: "spring", stiffness: 200 }}
              className="inline-block mb-6 p-4 bg-destructive/10 rounded-full"
            >
              <AlertTriangle className="h-8 w-8 text-destructive" />
            </motion.div>
            <motion.h2
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.3 }}
              className="text-xl font-semibold mb-2"
            >
              Oops! Something went wrong
            </motion.h2>
            <motion.p
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.4 }}
              className="text-muted-foreground mb-6"
            >
              {error}
            </motion.p>
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.5 }}
            >
              <Button onClick={() => fetchBlogPosts()} variant="outline">
                Try Again
              </Button>
            </motion.div>
          </motion.div>
        </div>
      </div>
    );
  }

  const featuredPost = getFeaturedPost(blogPosts);
  const remainingPosts = getRemainingPosts(blogPosts, featuredPost);

  // Show carousel for multiple featured posts or single featured post
  const showCarousel = useCarousel && blogPosts.length >= 2;

  return (
    <div className="min-h-screen bg-background">
      <Navbar variant="landing" />
      <div className="pt-16">
      {/* Featured Section - Carousel or Single Post */}
      {showCarousel ? (
        <FeaturedCarousel posts={blogPosts} />
      ) : featuredPost ? (
        <section className="pt-10 pb-20">
          <div className="container mx-auto px-4">
            <motion.div
              initial={{ opacity: 0, y: 24 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, ease: [0.25, 0.46, 0.45, 0.94] }}
              className="mb-16"
            >
              <Card className="overflow-hidden rounded-2xl border border-border/50 bg-linear-to-br from-brand-500/8 via-card to-brand-600/5 shadow-xl shadow-black/5 dark:shadow-none hover:shadow-2xl hover:shadow-black/10 dark:hover:shadow-none transition-shadow duration-300">
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-0">
                  {/* Featured Image */}
                  <div className="relative h-64 lg:h-full min-h-[320px] overflow-hidden">
                    {featuredPost.featured_image ? (
                      <img
                        src={featuredPost.featured_image}
                        alt={featuredPost.title}
                        className="w-full h-full object-cover transition-transform duration-500 hover:scale-[1.02]"
                      />
                    ) : (
                      <div className="w-full h-full bg-linear-to-br from-brand-500/15 to-brand-600/10 flex items-center justify-center">
                        <BookOpen className="h-20 w-20 text-brand-500/50" />
                      </div>
                    )}
                    <div className="absolute inset-0 bg-linear-to-t from-black/30 via-transparent to-transparent pointer-events-none" />
                  </div>

                  {/* Content */}
                  <div className="p-8 lg:p-12 flex flex-col justify-center">
                    <div className="flex items-center gap-2 mb-5">
                      <Sparkles className="h-4 w-4 text-brand-500" />
                      <span className="text-xs font-semibold uppercase tracking-wider text-brand-600 dark:text-brand-400">
                        Featured
                      </span>
                    </div>

                    <h2 className="text-2xl sm:text-3xl lg:text-4xl font-bold mb-4 leading-[1.2] tracking-tight">
                      <Link
                        to={`/blog/${featuredPost.slug}`}
                        className="text-foreground hover:text-brand-600 dark:hover:text-brand-400 transition-colors duration-200"
                      >
                        {featuredPost.title}
                      </Link>
                    </h2>

                    <p className="text-base text-muted-foreground mb-6 leading-[1.7]">
                      {featuredPost.excerpt || featuredPost.content.substring(0, 300) + '...'}
                    </p>

                    {/* Author and Meta Info */}
                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
                      <div className="flex items-center gap-3">
                        <div className="w-10 h-10 rounded-full bg-brand-500/15 flex items-center justify-center text-brand-600 dark:text-brand-400 font-semibold text-sm ring-2 ring-border/50">
                          {getAuthorAvatar(featuredPost.author)}
                        </div>
                        <div>
                          <p className="font-medium text-foreground text-sm">{featuredPost.author}</p>
                          <div className="flex items-center gap-3 text-xs text-muted-foreground">
                            <span>{getRelativeTime(featuredPost.published_at || featuredPost.created_at)}</span>
                            <span className="opacity-50">·</span>
                            <div className="flex items-center gap-1">
                              <Clock className="h-3 w-3" />
                              <span>{formatReadingTime(calculateReadingTime(featuredPost.content))}</span>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>

                    {/* Tags */}
                    {featuredPost.tags.length > 0 && (
                      <div className="flex flex-wrap gap-2 mb-6">
                        {featuredPost.tags.slice(0, 3).map((tag) => (
                          <span
                            key={tag}
                            className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-medium bg-muted/60 text-muted-foreground border border-border/50"
                          >
                            <Tag className="h-3 w-3 opacity-70" />
                            {tag}
                          </span>
                        ))}
                      </div>
                    )}

                    <div className="flex gap-3">
                      <Button asChild size="lg" className="rounded-full px-6 bg-brand-500 hover:bg-brand-600 shadow-lg shadow-brand-500/25 transition-all duration-200">
                        <Link to={`/blog/${featuredPost.slug}`} className="inline-flex items-center gap-2">
                          Read Full Post
                          <ArrowRight className="h-4 w-4" />
                        </Link>
                      </Button>
                    </div>
                  </div>
                </div>
              </Card>
            </motion.div>
          </div>
        </section>
      ) : null}

      {/* Blog Posts Grid with Search and Filter */}
      <div className="container mx-auto px-4 pb-24">
        <motion.div
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.1 }}
          className="mb-14 text-center"
        >
          <h2 className="text-3xl sm:text-4xl font-bold tracking-tight text-foreground mb-3">
            Latest Insights
          </h2>
          <p className="text-muted-foreground text-lg max-w-xl mx-auto leading-relaxed">
            Discover the latest thoughts, tutorials, and updates from our team
          </p>
        </motion.div>

        {/* Search and Filter */}
        <motion.div
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: 0.15 }}
          className="mb-10 flex flex-col items-center gap-6"
        >
          <SearchBar onSearch={handleSearch} placeholder="Search articles..." />
          {/* Categories from admin */}
          {categories.length > 0 && (
            <div className="w-full max-w-3xl">
              <p className="text-sm text-muted-foreground mb-2 text-center">Categories</p>
              <div className="flex flex-wrap justify-center gap-2">
                {categories
                  .sort((a, b) => (a.order ?? 0) - (b.order ?? 0))
                  .map((cat) => (
                    <span
                      key={cat.id}
                      className="inline-flex px-3 py-1.5 rounded-full text-sm font-medium bg-muted/60 text-muted-foreground border border-border/50"
                    >
                      {cat.title}
                    </span>
                  ))}
              </div>
            </div>
          )}
          <FilterBar
            availableTags={availableTags}
            selectedTags={selectedTags}
            onTagSelect={handleTagSelect}
            onClearAll={handleClearFilters}
          />
        </motion.div>

        {remainingPosts.length > 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 lg:gap-8">
            {remainingPosts.map((post: BlogPost, index: number) => (
              <motion.div
                key={post.id}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.45, delay: index * 0.08, ease: [0.25, 0.46, 0.45, 0.94] }}
              >
                <Card className="h-full overflow-hidden rounded-2xl border border-border/50 bg-card/80 shadow-lg shadow-black/5 dark:shadow-none hover:shadow-xl hover:shadow-black/8 dark:hover:shadow-none hover:border-border transition-all duration-300 group">
                  {/* Featured Image */}
                  <Link to={`/blog/${post.slug}`} className="block relative h-52 overflow-hidden">
                    {post.featured_image ? (
                      <img
                        src={post.featured_image}
                        alt={post.title}
                        className="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
                      />
                    ) : (
                      <div className="w-full h-full bg-linear-to-br from-brand-500/10 to-brand-600/5 flex items-center justify-center">
                        <BookOpen className="h-14 w-14 text-brand-500/40" />
                      </div>
                    )}
                    <div className="absolute inset-0 bg-linear-to-t from-black/20 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none" />
                  </Link>

                  <CardContent className="p-6">
                    {/* Title */}
                    <h3 className="text-lg font-bold mb-3 leading-snug tracking-tight">
                      <Link
                        to={`/blog/${post.slug}`}
                        className="text-foreground hover:text-brand-600 dark:hover:text-brand-400 transition-colors duration-200 line-clamp-2"
                      >
                        {post.title}
                      </Link>
                    </h3>

                    {/* Excerpt */}
                    <p className="text-muted-foreground text-sm mb-4 line-clamp-3 leading-[1.6]">
                      {post.excerpt || post.content.substring(0, 150) + '...'}
                    </p>

                    {/* Author and Meta */}
                    <div className="flex items-center justify-between gap-3 mb-4">
                      <div className="flex items-center gap-2.5 min-w-0">
                        <div className="shrink-0 w-8 h-8 rounded-full bg-brand-500/15 flex items-center justify-center text-brand-600 dark:text-brand-400 font-semibold text-xs ring-1 ring-border/50">
                          {getAuthorAvatar(post.author)}
                        </div>
                        <div className="min-w-0 text-sm">
                          <p className="font-medium text-foreground truncate">{post.author}</p>
                          <p className="text-muted-foreground text-xs">{getRelativeTime(post.published_at || post.created_at)}</p>
                        </div>
                      </div>
                      <div className="flex items-center gap-1 text-xs text-muted-foreground shrink-0">
                        <Clock className="h-3 w-3" />
                        <span>{formatReadingTime(calculateReadingTime(post.content))}</span>
                      </div>
                    </div>

                    {/* Tags */}
                    {post.tags.length > 0 && (
                      <div className="flex flex-wrap gap-1.5 mb-4">
                        {post.tags.slice(0, 3).map((tag) => (
                          <span
                            key={tag}
                            className="inline-flex px-2 py-0.5 rounded-full text-xs font-medium bg-muted/50 text-muted-foreground border border-border/40"
                          >
                            {tag}
                          </span>
                        ))}
                      </div>
                    )}

                    {/* Actions Row */}
                    <div className="flex items-center justify-between">
                      {/* Read More Link */}
                      <Link
                        to={`/blog/${post.slug}`}
                        className="inline-flex items-center gap-1.5 text-sm font-medium text-brand-600 dark:text-brand-400 hover:text-brand-700 dark:hover:text-brand-300 transition-colors duration-200"
                      >
                        Read more
                        <ArrowRight className="h-3.5 w-3.5" />
                      </Link>

                      {/* Bookmark Button */}
                      <BookmarkButton
                        postId={post.id}
                        postTitle={post.title}
                        size="sm"
                      />
                    </div>
                  </CardContent>
                </Card>
              </motion.div>
            ))}
          </div>
        ) : (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="text-center py-20"
          >
            <div className="inline-flex p-5 rounded-2xl bg-muted/50 mb-6">
              <BookOpen className="h-14 w-14 text-muted-foreground" />
            </div>
            <h3 className="text-xl font-semibold mb-2">No posts found</h3>
            <p className="text-muted-foreground max-w-sm mx-auto">
              {searchQuery || selectedTags.length > 0
                ? 'Try adjusting your search or filters.'
                : 'Check back soon for new insights and updates.'}
            </p>
            {(searchQuery || selectedTags.length > 0) && (
              <Button
                variant="outline"
                onClick={handleClearFilters}
                className="mt-4 rounded-full"
              >
                Clear filters
              </Button>
            )}
          </motion.div>
        )}

        {/* Load More / Infinite Scroll */}
        {hasMore && remainingPosts.length > 0 && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.3 }}
            className="text-center mt-14"
          >
            {/* Sentinel element for infinite scroll */}
            <div ref={sentinelRef} className="h-4" />

            {loading ? (
              <div className="flex items-center justify-center gap-2">
                <Loader2 className="h-5 w-5 animate-spin text-brand-500" />
                <span className="text-muted-foreground">Loading more posts...</span>
              </div>
            ) : (
              <Button
                onClick={loadMore}
                variant="outline"
                size="lg"
                className="min-w-[220px] rounded-full border-border/60 bg-muted/30 hover:bg-muted/60 transition-colors duration-200"
              >
                Load More Posts
                <ArrowRight className="h-4 w-4 ml-2" />
              </Button>
            )}
          </motion.div>
        )}

        {/* End of content message */}
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

        {/* Back to Home */}
        <div className="text-center mt-16">
          <Link to="/">
            <Button variant="outline" size="lg" className="rounded-full px-8 border-border/60 bg-muted/20 hover:bg-muted/50 transition-colors duration-200">
              <ArrowRight className="h-4 w-4 mr-2 rotate-180" />
              Back to Home
            </Button>
          </Link>
        </div>
      </div>
      </div>
    </div>
  );
};

export default BlogPage;
