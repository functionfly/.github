'use client';

import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Navbar } from '@/components/common/Navbar';
import { BookOpen, Calendar, ArrowRight, User, Tag, Loader2, AlertTriangle, Sparkles, Clock } from 'lucide-react';
import { contentApi, BlogPost } from '@/api/content';
import {
  calculateReadingTime,
  formatReadingTime,
  getRelativeTime,
  getFeaturedPost,
  getRemainingPosts,
  getAuthorAvatar
} from './utils';

const BlogPage = () => {
  const [blogPosts, setBlogPosts] = useState<BlogPost[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [hasMore, setHasMore] = useState(false);
  const [offset, setOffset] = useState(0);

  const fetchBlogPosts = async (loadMore = false) => {
    try {
      if (!loadMore) {
        setLoading(true);
        setOffset(0);
      }

      const result = await contentApi.getPublishedBlogPosts({
        limit: 10,
        offset: loadMore ? offset : 0,
      });

      // Ensure posts is always an array
      const posts = result.posts || [];

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
    fetchBlogPosts();
  }, []);

  const loadMore = () => {
    fetchBlogPosts(true);
  };

  if (loading && blogPosts.length === 0) {
    return (
      <div className="min-h-screen bg-background">
        <Navbar variant="landing" />
        <div className="flex items-center justify-center min-h-[60vh] pt-24">
          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            className="text-center"
          >
            <motion.div
              animate={{ rotate: 360 }}
              transition={{ duration: 2, repeat: Infinity, ease: "linear" }}
              className="inline-block mb-6"
            >
              <BookOpen className="h-12 w-12 text-primary" />
            </motion.div>
            <motion.p
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
              className="text-lg text-muted-foreground"
            >
              Loading amazing insights...
            </motion.p>
          </motion.div>
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

  return (
    <div className="min-h-screen bg-background">
      <Navbar variant="landing" />

      {/* Hero Section - Featured Post */}
      {featuredPost && (
        <section className="pt-24 pb-16">
          <div className="container mx-auto px-4">
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6 }}
              className="mb-12"
            >
              <Card className="overflow-hidden bg-linear-to-r from-brand-500/10 to-brand-600/10 border-brand-500/20">
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-0">
                  {/* Featured Image */}
                  <div className="relative h-64 lg:h-full min-h-[300px]">
                    {featuredPost.featured_image ? (
                      <img
                        src={featuredPost.featured_image}
                        alt={featuredPost.title}
                        className="w-full h-full object-cover"
                      />
                    ) : (
                      <div className="w-full h-full bg-linear-to-br from-brand-500/20 to-brand-600/20 flex items-center justify-center">
                        <BookOpen className="h-16 w-16 text-brand-500/60" />
                      </div>
                    )}
                    <div className="absolute inset-0 bg-linear-to-t from-black/20 to-transparent" />
                  </div>

                  {/* Content */}
                  <div className="p-8 lg:p-12 flex flex-col justify-center">
                    <div className="flex items-center gap-2 mb-4">
                      <Sparkles className="h-5 w-5 text-brand-500" />
                      <Badge variant="secondary" className="bg-brand-500/10 text-brand-600 border-brand-500/20">
                        Featured Post
                      </Badge>
                    </div>

                    <h2 className="text-3xl lg:text-4xl font-bold mb-4 leading-tight">
                      <Link
                        to={`/blog/${featuredPost.slug}`}
                        className="hover:text-brand-500 transition-colors"
                      >
                        {featuredPost.title}
                      </Link>
                    </h2>

                    <p className="text-lg text-muted-foreground mb-6 leading-relaxed">
                      {featuredPost.excerpt || featuredPost.content.substring(0, 300) + '...'}
                    </p>

                    {/* Author and Meta Info */}
                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
                      <div className="flex items-center gap-3">
                        <div className="w-10 h-10 rounded-full bg-brand-500/20 flex items-center justify-center text-brand-600 font-semibold">
                          {getAuthorAvatar(featuredPost.author)}
                        </div>
                        <div>
                          <p className="font-medium text-foreground">{featuredPost.author}</p>
                          <div className="flex items-center gap-4 text-sm text-muted-foreground">
                            <span>{getRelativeTime(featuredPost.published_at || featuredPost.created_at)}</span>
                            <span>•</span>
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
                          <Badge key={tag} variant="outline" className="text-xs">
                            <Tag className="h-3 w-3 mr-1" />
                            {tag}
                          </Badge>
                        ))}
                      </div>
                    )}

                    <div className="flex gap-3">
                      <Button asChild className="bg-brand-500 hover:bg-brand-600">
                        <Link to={`/blog/${featuredPost.slug}`}>
                          Read Full Post
                          <ArrowRight className="h-4 w-4 ml-2" />
                        </Link>
                      </Button>
                    </div>
                  </div>
                </div>
              </Card>
            </motion.div>
          </div>
        </section>
      )}

      {/* Blog Posts Grid */}
      <div className="container mx-auto px-4 pb-16">
        <div className="mb-12 text-center">
          <h1 className="text-4xl font-bold mb-4">Latest Insights</h1>
          <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
            Discover the latest thoughts, tutorials, and updates from our team
          </p>
        </div>

        {remainingPosts.length > 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {remainingPosts.map((post: BlogPost, index: number) => (
              <motion.div
                key={post.id}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5, delay: index * 0.1 }}
              >
                <Card className="h-full overflow-hidden hover:shadow-lg transition-all duration-300 group border-border/50 hover:border-brand-500/20">
                  {/* Featured Image */}
                  <div className="relative h-48 overflow-hidden">
                    {post.featured_image ? (
                      <img
                        src={post.featured_image}
                        alt={post.title}
                        className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                      />
                    ) : (
                      <div className="w-full h-full bg-linear-to-br from-brand-500/10 to-brand-600/10 flex items-center justify-center">
                        <BookOpen className="h-12 w-12 text-brand-500/40" />
                      </div>
                    )}
                    <div className="absolute inset-0 bg-linear-to-t from-black/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
                  </div>

                  <CardContent className="p-6">
                    {/* Title */}
                    <h3 className="text-xl font-bold mb-3 leading-tight">
                      <Link
                        to={`/blog/${post.slug}`}
                        className="hover:text-brand-500 transition-colors line-clamp-2"
                      >
                        {post.title}
                      </Link>
                    </h3>

                    {/* Excerpt */}
                    <p className="text-muted-foreground mb-4 line-clamp-3 leading-relaxed">
                      {post.excerpt || post.content.substring(0, 150) + '...'}
                    </p>

                    {/* Author and Meta */}
                    <div className="flex items-center justify-between mb-4">
                      <div className="flex items-center gap-2">
                        <div className="w-8 h-8 rounded-full bg-brand-500/20 flex items-center justify-center text-brand-600 font-semibold text-sm">
                          {getAuthorAvatar(post.author)}
                        </div>
                        <div className="text-sm">
                          <p className="font-medium text-foreground">{post.author}</p>
                          <p className="text-muted-foreground">{getRelativeTime(post.published_at || post.created_at)}</p>
                        </div>
                      </div>
                      <div className="flex items-center gap-1 text-sm text-muted-foreground">
                        <Clock className="h-3 w-3" />
                        <span>{formatReadingTime(calculateReadingTime(post.content))}</span>
                      </div>
                    </div>

                    {/* Tags */}
                    {post.tags.length > 0 && (
                      <div className="flex flex-wrap gap-1 mb-4">
                        {post.tags.slice(0, 3).map((tag) => (
                          <Badge key={tag} variant="secondary" className="text-xs px-2 py-1">
                            {tag}
                          </Badge>
                        ))}
                      </div>
                    )}

                    {/* Read More Link */}
                    <Button variant="ghost" size="sm" asChild className="p-0 h-auto text-brand-600 hover:text-brand-700 hover:bg-transparent">
                      <Link to={`/blog/${post.slug}`}>
                        Read more
                        <ArrowRight className="h-3 w-3 ml-1" />
                      </Link>
                    </Button>
                  </CardContent>
                </Card>
              </motion.div>
            ))}
          </div>
        ) : (
          <div className="text-center py-16">
            <BookOpen className="h-16 w-16 mx-auto text-muted-foreground mb-4" />
            <h3 className="text-xl font-semibold mb-2">No posts yet</h3>
            <p className="text-muted-foreground">Check back soon for new insights and updates.</p>
          </div>
        )}

        {/* Load More Button */}
        {hasMore && remainingPosts.length > 0 && (
          <div className="text-center mt-12">
            <Button
              onClick={loadMore}
              variant="outline"
              size="lg"
              disabled={loading}
              className="min-w-[200px]"
            >
              {loading ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Loading...
                </>
              ) : (
                <>
                  Load More Posts
                  <ArrowRight className="h-4 w-4 ml-2" />
                </>
              )}
            </Button>
          </div>
        )}

        {/* Back to Home */}
        <div className="text-center mt-16">
          <Link to="/">
            <Button variant="outline" size="lg">
              <ArrowRight className="h-4 w-4 mr-2 rotate-180" />
              Back to Home
            </Button>
          </Link>
        </div>
      </div>
    </div>
  );
};

export default BlogPage;