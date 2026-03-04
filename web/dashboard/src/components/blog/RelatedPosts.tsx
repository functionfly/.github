'use client';

import { motion } from 'framer-motion';
import { Link } from 'react-router-dom';
import { ArrowRight, Clock } from 'lucide-react';
import { BlogPost } from '@/api/content';
import { calculateReadingTime, formatReadingTime, getRelativeTime, getAuthorAvatar } from '@/pages/BlogPage/utils';
import { SpotlightCard, ShinyButton, TextGradient } from '@/components/ui';

interface RelatedPostsProps {
  posts: BlogPost[];
  currentPostId?: string;
  maxPosts?: number;
}

export function RelatedPosts({ posts, currentPostId, maxPosts = 3 }: RelatedPostsProps) {
  // Filter out current post and limit to maxPosts
  const relatedPosts = posts
    .filter(post => post.id !== currentPostId)
    .slice(0, maxPosts);

  if (relatedPosts.length === 0) {
    return null;
  }

  return (
    <motion.section
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4 }}
      className="mt-16"
    >
      <TextGradient
        size="subheading"
        animate={true}
        className="text-2xl font-bold mb-8"
      >
        Related Posts
      </TextGradient>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {relatedPosts.map((post, index) => (
          <motion.div
            key={post.id}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3, delay: index * 0.1 }}
          >
            <SpotlightCard
              className="blog-post-card h-full overflow-hidden rounded-2xl border border-border/50 bg-card/80 hover:shadow-lg hover:border-border transition-all duration-300 group"
              spotlightColor="rgba(139, 92, 246, 0.1)"
              spotlightSize={220}
              hoverOnly={true}
            >
              {/* Featured Image */}
              <Link to={`/blog/${post.slug}`} className="block relative h-40 overflow-hidden">
                {post.featured_image ? (
                  <img
                    src={post.featured_image}
                    alt={post.title}
                    className="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
                  />
                ) : (
                  <div className="blog-related-card w-full h-full bg-gradient-to-br from-brand-500/10 to-brand-600/5 flex items-center justify-center">
                    <span className="text-4xl">📝</span>
                  </div>
                )}
              </Link>

              <div className="p-5">
                {/* Title */}
                <h3 className="text-base font-bold mb-2 line-clamp-2">
                  <Link
                    to={`/blog/${post.slug}`}
                    className="text-foreground hover:text-brand-600 dark:hover:text-brand-400 transition-colors"
                  >
                    {post.title}
                  </Link>
                </h3>

                {/* Meta */}
                <div className="flex items-center gap-2 mb-3">
                  <div className="w-6 h-6 rounded-full bg-brand-500/15 flex items-center justify-center text-brand-600 dark:text-brand-400 font-semibold text-xs">
                    {getAuthorAvatar(post.author)}
                  </div>
                  <span className="text-xs text-muted-foreground truncate">
                    {post.author}
                  </span>
                  <span className="text-xs text-muted-foreground">·</span>
                  <span className="text-xs text-muted-foreground">
                    {getRelativeTime(post.published_at || post.created_at)}
                  </span>
                </div>

                {/* Reading time */}
                <div className="flex items-center gap-1 text-xs text-muted-foreground mb-3">
                  <Clock className="h-3 w-3" />
                  <span>{formatReadingTime(calculateReadingTime(post.content))}</span>
                </div>

                {/* Tags */}
                {post.tags.length > 0 && (
                  <div className="flex flex-wrap gap-1 mb-3">
                    {post.tags.slice(0, 2).map((tag) => (
                      <span
                        key={tag}
                        className="inline-flex px-2 py-0.5 rounded-full text-xs font-medium bg-muted/50 text-muted-foreground"
                      >
                        {tag}
                      </span>
                    ))}
                  </div>
                )}

                {/* Read more */}
                <ShinyButton asChild size="sm" variant="ghost">
                  <Link
                    to={`/blog/${post.slug}`}
                    className="inline-flex items-center gap-1.5"
                  >
                    Read more
                    <ArrowRight className="h-3.5 w-3.5" />
                  </Link>
                </ShinyButton>
              </div>
            </SpotlightCard>
          </motion.div>
        ))}
      </div>
    </motion.section>
  );
}

export default RelatedPosts;
