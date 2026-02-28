'use client';

import { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Link } from 'react-router-dom';
import { ChevronLeft, ChevronRight, Sparkles, Clock, ArrowRight, Pause, Play } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { BlogPost } from '@/api/content';
import { calculateReadingTime, formatReadingTime, getRelativeTime, getAuthorAvatar } from '@/pages/BlogPage/utils';

interface FeaturedCarouselProps {
  posts: BlogPost[];
  autoPlayInterval?: number;
}

export function FeaturedCarousel({ posts, autoPlayInterval = 5000 }: FeaturedCarouselProps) {
  const [currentIndex, setCurrentIndex] = useState(0);
  const [isAutoPlaying, setIsAutoPlaying] = useState(true);
  const [direction, setDirection] = useState(0);

  const featuredPosts = posts.filter(post => post.is_published);

  const goToNext = useCallback(() => {
    if (featuredPosts.length <= 1) return;
    setDirection(1);
    setCurrentIndex((prev) => (prev + 1) % featuredPosts.length);
  }, [featuredPosts.length]);

  const goToPrev = useCallback(() => {
    if (featuredPosts.length <= 1) return;
    setDirection(-1);
    setCurrentIndex((prev) => (prev - 1 + featuredPosts.length) % featuredPosts.length);
  }, [featuredPosts.length]);

  const goToSlide = useCallback((index: number) => {
    setDirection(index > currentIndex ? 1 : -1);
    setCurrentIndex(index);
  }, [currentIndex]);

  // Auto-play functionality
  useEffect(() => {
    if (!isAutoPlaying || featuredPosts.length <= 1) return;

    const timer = setInterval(goToNext, autoPlayInterval);
    return () => clearInterval(timer);
  }, [isAutoPlaying, autoPlayInterval, goToNext, featuredPosts.length]);

  // Pause on hover
  const handleMouseEnter = () => setIsAutoPlaying(false);
  const handleMouseLeave = () => setIsAutoPlaying(true);

  if (featuredPosts.length === 0) {
    return null;
  }

  const currentPost = featuredPosts[currentIndex];

  const variants = {
    enter: (direction: number) => ({
      x: direction > 0 ? 300 : -300,
      opacity: 0,
    }),
    center: {
      x: 0,
      opacity: 1,
    },
    exit: (direction: number) => ({
      x: direction < 0 ? 300 : -300,
      opacity: 0,
    }),
  };

  return (
    <section 
      className="pt-10 pb-20"
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
    >
      <div className="container mx-auto px-4">
        <motion.div
          initial={{ opacity: 0, y: 24 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
          className="relative"
        >
          <Card className="overflow-hidden rounded-2xl border border-border/50 bg-linear-to-br from-brand-500/8 via-card to-brand-600/5 shadow-xl shadow-black/5 dark:shadow-none">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-0">
              {/* Featured Image */}
              <div className="relative h-64 lg:h-full min-h-[320px] overflow-hidden">
                <AnimatePresence mode="wait" custom={direction}>
                  <motion.div
                    key={currentIndex}
                    custom={direction}
                    variants={variants}
                    initial="enter"
                    animate="center"
                    exit="exit"
                    transition={{ duration: 0.4, ease: 'easeInOut' }}
                    className="absolute inset-0"
                  >
                    {currentPost.featured_image ? (
                      <img
                        src={currentPost.featured_image}
                        alt={currentPost.title}
                        className="w-full h-full object-cover"
                      />
                    ) : (
                      <div className="w-full h-full bg-linear-to-br from-brand-500/15 to-brand-600/10 flex items-center justify-center">
                        <span className="text-6xl">📝</span>
                      </div>
                    )}
                  </motion.div>
                </AnimatePresence>
                <div className="absolute inset-0 bg-linear-to-t from-black/30 via-transparent to-transparent pointer-events-none" />
              </div>

              {/* Content */}
              <CardContent className="p-8 lg:p-12 flex flex-col justify-center relative">
                <AnimatePresence mode="wait">
                  <motion.div
                    key={currentIndex}
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: -20 }}
                    transition={{ duration: 0.3 }}
                  >
                    {/* Featured badge */}
                    <div className="flex items-center gap-2 mb-5">
                      <Sparkles className="h-4 w-4 text-brand-500" />
                      <span className="text-xs font-semibold uppercase tracking-wider text-brand-600 dark:text-brand-400">
                        Featured
                      </span>
                    </div>

                    {/* Title */}
                    <h2 className="text-2xl sm:text-3xl lg:text-4xl font-bold mb-4 leading-[1.2] tracking-tight">
                      <Link
                        to={`/blog/${currentPost.slug}`}
                        className="text-foreground hover:text-brand-600 dark:hover:text-brand-400 transition-colors duration-200"
                      >
                        {currentPost.title}
                      </Link>
                    </h2>

                    {/* Excerpt */}
                    <p className="text-base text-muted-foreground mb-6 leading-[1.7]">
                      {currentPost.excerpt || currentPost.content.substring(0, 300) + '...'}
                    </p>

                    {/* Author and Meta */}
                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
                      <div className="flex items-center gap-3">
                        <div className="w-10 h-10 rounded-full bg-brand-500/15 flex items-center justify-center text-brand-600 dark:text-brand-400 font-semibold text-sm ring-2 ring-border/50">
                          {getAuthorAvatar(currentPost.author)}
                        </div>
                        <div>
                          <p className="font-medium text-foreground text-sm">{currentPost.author}</p>
                          <div className="flex items-center gap-3 text-xs text-muted-foreground">
                            <span>{getRelativeTime(currentPost.published_at || currentPost.created_at)}</span>
                            <span className="opacity-50">·</span>
                            <div className="flex items-center gap-1">
                              <Clock className="h-3 w-3" />
                              <span>{formatReadingTime(calculateReadingTime(currentPost.content))}</span>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>

                    {/* Tags */}
                    {currentPost.tags.length > 0 && (
                      <div className="flex flex-wrap gap-2 mb-6">
                        {currentPost.tags.slice(0, 3).map((tag) => (
                          <span
                            key={tag}
                            className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-medium bg-muted/60 text-muted-foreground border border-border/50"
                          >
                            {tag}
                          </span>
                        ))}
                      </div>
                    )}

                    {/* CTA */}
                    <div className="flex gap-3">
                      <Button asChild size="lg" className="rounded-full px-6 bg-brand-500 hover:bg-brand-600 shadow-lg shadow-brand-500/25 transition-all duration-200">
                        <Link to={`/blog/${currentPost.slug}`} className="inline-flex items-center gap-2">
                          Read Full Post
                          <ArrowRight className="h-4 w-4" />
                        </Link>
                      </Button>
                    </div>
                  </motion.div>
                </AnimatePresence>

                {/* Navigation Controls */}
                {featuredPosts.length > 1 && (
                  <div className="absolute bottom-6 right-6 flex items-center gap-2">
                    {/* Dot indicators */}
                    <div className="hidden sm:flex items-center gap-1.5 mr-4" role="tablist" aria-label="Featured posts navigation">
                      {featuredPosts.map((_, index) => (
                        <button
                          key={index}
                          onClick={() => goToSlide(index)}
                          className={`
                            w-2 h-2 rounded-full transition-all duration-300
                            ${index === currentIndex
                              ? 'w-6 bg-brand-500'
                              : 'bg-muted-foreground/30 hover:bg-muted-foreground/50'
                            }
                          `}
                          role="tab"
                          aria-selected={index === currentIndex}
                          aria-label={`Go to slide ${index + 1}`}
                        />
                      ))}
                    </div>

                    {/* Arrow buttons */}
                    <div className="flex items-center gap-1">
                      <Button
                        variant="outline"
                        size="icon"
                        className="h-9 w-9 rounded-full border-border/60 hover:bg-muted"
                        onClick={goToPrev}
                        aria-label="Previous post"
                      >
                        <ChevronLeft className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="outline"
                        size="icon"
                        className="h-9 w-9 rounded-full border-border/60 hover:bg-muted"
                        onClick={goToNext}
                        aria-label="Next post"
                      >
                        <ChevronRight className="h-4 w-4" />
                      </Button>
                    </div>

                    {/* Auto-play toggle */}
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-9 w-9 rounded-full"
                      onClick={() => setIsAutoPlaying(!isAutoPlaying)}
                      aria-label={isAutoPlaying ? 'Pause auto-play' : 'Play auto-play'}
                    >
                      {isAutoPlaying ? (
                        <Pause className="h-4 w-4" />
                      ) : (
                        <Play className="h-4 w-4" />
                      )}
                    </Button>
                  </div>
                )}
              </CardContent>
            </div>
          </Card>
        </motion.div>
      </div>
    </section>
  );
}
