'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { motion, useScroll, useTransform } from 'framer-motion';
import { Button } from '@/components/ui/button';
import { TooltipProvider } from '@/components/ui/tooltip';
import { Navbar } from '@/components/common/Navbar';
import { Calendar, ArrowLeft, Tag, Loader2, AlertTriangle, Clock } from 'lucide-react';
import { contentApi, BlogPost } from '@/api/content';
import { Footer } from '@/pages/LandingPage/components/Footer';
import { calculateReadingTime, formatReadingTime, getAuthorAvatar } from '@/pages/BlogPage/utils';
import { BlogPostBody } from './BlogPostBody';
import { ArticleSkeleton, ShareButtons, NewsletterSignup, AuthorBio, RelatedPosts, TableOfContents, BookmarkButton } from '@/components/blog';
import { MetaTags } from '@/components/seo/MetaTags';
import { BlogPostStructuredData, BreadcrumbStructuredData } from '@/components/seo/StructuredData';
import { useWebVitals } from '@/hooks/useWebVitals';
import { PublicAnalytics } from '@/components/common/PublicAnalytics';
import './blog-post.css';

const BlogPostPage = () => {
  // Monitor Core Web Vitals
  useWebVitals((metrics) => {
    // Optional: Send to your analytics service
    console.log('Web Vitals:', metrics);
  });

  const { slug } = useParams<{ slug: string }>();
  const [post, setPost] = useState<BlogPost | null>(null);
  const [relatedPosts, setRelatedPosts] = useState<BlogPost[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchBlogPost = async () => {
      if (!slug) return;

      try {
        setLoading(true);
        const blogPost = await contentApi.getPublishedBlogPostBySlug(slug);
        setPost(blogPost);

        // Fetch related posts based on tags
        if (blogPost.tags && blogPost.tags.length > 0) {
          const result = await contentApi.getPublishedBlogPosts({
            limit: 10,
            tags: blogPost.tags,
          });
          // Filter out current post
          setRelatedPosts(result.posts.filter(p => p.id !== blogPost.id));
        }
      } catch (err) {
        console.error('Failed to fetch blog post:', err);
        setError('Failed to load blog post');
      } finally {
        setLoading(false);
      }
    };

    fetchBlogPost();
  }, [slug]);

  const { scrollYProgress } = useScroll();
  const progressWidth = useTransform(scrollYProgress, [0, 0.92], [0, 1]);

  if (loading) {
    return (
      <div className="min-h-screen bg-background">
        <Navbar variant="landing" />
        <ArticleSkeleton />
      </div>
    );
  }

  if (error || !post) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center px-4">
        <motion.div
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-center max-w-sm"
        >
          <div className="inline-flex p-4 rounded-2xl bg-destructive/10 mb-6">
            <AlertTriangle className="h-10 w-10 text-destructive" />
          </div>
          <h2 className="text-xl font-semibold mb-2">Something went wrong</h2>
          <p className="text-muted-foreground text-sm mb-8">{error || 'Blog post not found'}</p>
          <Button asChild variant="outline" size="lg" className="rounded-full px-6">
            <Link to="/blog">
              <ArrowLeft className="h-4 w-4 mr-2" />
              Back to Blog
            </Link>
          </Button>
        </motion.div>
      </div>
    );
  }

  const readingMinutes = calculateReadingTime(post.content);

  // Get current URL for sharing
  const shareUrl = typeof window !== 'undefined' ? window.location.href : '';

  return (
    <TooltipProvider>
      <div className="min-h-screen bg-background">
        {/* SEO Meta Tags - only render when post is loaded */}
        {post && (
          <MetaTags
            title={`${post.title} | FunctionFly Blog`}
            description={post.excerpt}
            keywords={post.tags}
            image={post.featured_image}
            url={shareUrl}
            type="article"
            author={post.author}
          />
        )}

        {/* Structured Data - only render when post is loaded */}
        {post && (
          <>
            <BlogPostStructuredData
              post={{
                title: post.title,
                excerpt: post.excerpt,
                content: post.content,
                author: {
                  name: post.author as string,
                  bio: undefined,
                  avatar: undefined
                },
                publishedAt: post.published_at || post.created_at,
                updatedAt: post.updated_at,
                tags: post.tags,
                slug: post.slug,
                readingTime: calculateReadingTime(post.content),
                image: post.featured_image
              }}
              url={shareUrl}
            />
            <BreadcrumbStructuredData
              items={[
                { name: 'Home', url: window.location.origin },
                { name: 'Blog', url: `${window.location.origin}/blog` },
                { name: post.title, url: shareUrl }
              ]}
            />
          </>
        )}

        {/* Public Analytics (Hotjar for user behavior) */}
        <PublicAnalytics />

      <Navbar variant="landing" />

      {/* Reading progress bar */}
      <motion.div
        className="fixed top-0 left-0 right-0 z-50 h-0.5 bg-brand-500 origin-left"
        style={{ scaleX: progressWidth }}
        initial={false}
      />

      {/* Hero - centered, soft 2027 style */}
      <header className="pt-20 pb-12 md:pt-24 md:pb-16">
        <div className="container mx-auto px-4">
          <motion.div
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, ease: [0.25, 0.46, 0.45, 0.94] }}
            className="max-w-3xl mx-auto text-center"
          >
            {/* Back to Blog Button */}
            <motion.div
              initial={{ opacity: 0, y: -8 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.4, delay: 0.1 }}
              className="mb-6"
            >
              <Button
                asChild
                variant="ghost"
                size="sm"
                className="rounded-full gap-1.5 text-muted-foreground hover:text-foreground"
              >
                <Link to="/blog">
                  <ArrowLeft className="h-4 w-4" />
                  Back to Blog
                </Link>
              </Button>
            </motion.div>

            {/* Logo - links to blog home */}
            <motion.div
              initial={{ opacity: 0, y: -8 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.4, delay: 0.15 }}
              className="mb-6"
            >
              <Link to="/blog" className="inline-block">
                <img
                  src="/logo/logo-icon.svg"
                  alt="FunctionFly"
                  className="h-12 w-auto"
                />
              </Link>
            </motion.div>

            <h1 className="text-3xl sm:text-4xl md:text-5xl font-bold tracking-tight leading-[1.15] text-foreground mb-6">
              {post.title}
            </h1>
            <div className="flex flex-wrap items-center justify-center gap-x-6 gap-y-2 text-sm text-muted-foreground">
              <span className="inline-flex items-center gap-2">
                <span className="flex items-center justify-center w-8 h-8 rounded-full bg-brand-500/10 text-brand-600 dark:text-brand-400 font-medium text-xs">
                  {getAuthorAvatar(post.author)}
                </span>
                <span className="font-medium text-foreground/90">{post.author}</span>
              </span>
              <span className="inline-flex items-center gap-1.5">
                <Calendar className="h-4 w-4 opacity-70" />
                {new Date(post.published_at || post.created_at).toLocaleDateString(undefined, {
                  year: 'numeric',
                  month: 'short',
                  day: 'numeric',
                })}
              </span>
              <span className="inline-flex items-center gap-1.5">
                <Clock className="h-4 w-4 opacity-70" />
                {formatReadingTime(readingMinutes)}
              </span>
            </div>
          </motion.div>
        </div>
      </header>

      <div className="container mx-auto px-4 pb-20">
        <div className="max-w-3xl mx-auto">
          {/* Featured image - soft radius, subtle shadow */}
          {post.featured_image && (
            <motion.div
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.1 }}
              className="mb-10 overflow-hidden rounded-2xl bg-muted/30 shadow-lg shadow-black/5 dark:shadow-none ring-1 ring-border/50"
            >
              <img
                src={post.featured_image}
                alt={post.title}
                className="w-full aspect-[21/9] object-cover"
              />
            </motion.div>
          )}

          {/* Tags - soft pills */}
          {post.tags.length > 0 && (
            <motion.div
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.4, delay: 0.15 }}
              className="flex flex-wrap justify-center gap-2 mb-10"
            >
              {post.tags.map((tag: string) => (
                <span
                  key={tag}
                  className="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-full text-xs font-medium bg-muted/60 text-muted-foreground border border-border/50 hover:bg-muted hover:border-border/80 transition-colors duration-200"
                >
                  <Tag className="h-3 w-3 opacity-70" />
                  {tag}
                </span>
              ))}
            </motion.div>
          )}

          {/* Share Buttons */}
          <motion.div
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4, delay: 0.2 }}
            className="flex justify-center mb-10"
          >
            <ShareButtons
              url={shareUrl}
              title={post.title}
              description={post.excerpt || post.content.substring(0, 150)}
              size="md"
            />
          </motion.div>

          {/* Article content - card with refined prose */}
          <motion.article
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.25 }}
            className="rounded-2xl border border-border/50 bg-white dark:bg-slate-900 shadow-xl shadow-black/5 dark:shadow-none overflow-hidden"
          >
            <div className="p-8 sm:p-10 md:p-12">
              <BlogPostBody
                html={post.content}
                className="blog-post-prose prose prose-slate prose-lg dark:prose-invert max-w-none text-foreground prose-headings:font-semibold prose-headings:tracking-tight prose-headings:text-foreground prose-p:leading-[1.75] prose-p:text-foreground prose-strong:text-foreground prose-a:text-brand-600 dark:prose-a:text-brand-400 prose-a:no-underline hover:prose-a:underline prose-img:rounded-xl prose-pre:rounded-xl prose-pre:bg-transparent prose-blockquote:border-l-brand-500/50 prose-blockquote:bg-muted/30 prose-blockquote:py-1 prose-blockquote:px-4 prose-blockquote:rounded-r-lg prose-blockquote:text-foreground prose-code:text-foreground prose-code:bg-transparent prose-ul:text-foreground prose-ol:text-foreground prose-li:text-foreground"
              />
            </div>
          </motion.article>

          {/* Author Bio Card */}
          <motion.div
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4, delay: 0.35 }}
            className="mt-10"
          >
            <AuthorBio
              name={post.author}
              bio="Writer and content creator sharing insights on technology and development."
              postsUrl={`/blog?author=${encodeURIComponent(post.author)}`}
            />
          </motion.div>

          {/* Newsletter Subscription */}
          <motion.div
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4, delay: 0.4 }}
            className="mt-10"
          >
            <NewsletterSignup />
          </motion.div>

          {/* Related Posts */}
          <RelatedPosts
            posts={relatedPosts}
            currentPostId={post.id}
            maxPosts={3}
          />

          {/* Back to Blog - soft pill CTA */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.5 }}
            className="mt-12 text-center"
          >
            <Button
              asChild
              variant="outline"
              size="lg"
              className="rounded-full px-8 py-6 text-sm font-medium border-border/60 bg-muted/30 hover:bg-muted/60 hover:border-border transition-all duration-200"
            >
              <Link to="/blog" className="inline-flex items-center gap-2">
                <ArrowLeft className="h-4 w-4" />
                Back to Blog
              </Link>
            </Button>
          </motion.div>
        </div>

        {/* Table of Contents (Desktop) */}
        <TableOfContents content={post.content} />
      </div>

      <Footer />
    </div>
    </TooltipProvider>
  );
};

export default BlogPostPage;
