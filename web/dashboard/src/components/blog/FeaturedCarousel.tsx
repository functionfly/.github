'use client';

import { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Link } from 'react-router-dom';
import { ChevronLeft, ChevronRight, Sparkles, Clock, ArrowRight, Pause, Play, TrendingUp } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { BlogPost } from '@/api/content';
import { calculateReadingTime, formatReadingTime, getRelativeTime, getAuthorAvatar } from '@/pages/BlogPage/utils';
import { ParticleBackground } from '@/components/ui';

interface FeaturedCarouselProps {
  posts: BlogPost[];
  autoPlayInterval?: number;
}

export function FeaturedCarousel({ posts, autoPlayInterval = 6000 }: FeaturedCarouselProps) {
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

  const slideVariants = {
    enter: (direction: number) => ({
      x: direction > 0 ? '100%' : '-100%',
      opacity: 0,
      scale: 0.95,
    }),
    center: {
      x: 0,
      opacity: 1,
      scale: 1,
    },
    exit: (direction: number) => ({
      x: direction < 0 ? '100%' : '-100%',
      opacity: 0,
      scale: 0.95,
    }),
  };

  const contentVariants = {
    enter: { opacity: 0, y: 30 },
    center: { opacity: 1, y: 0 },
    exit: { opacity: 0, y: -30 },
  };

  return (
    <section
      className="relative pt-8 pb-20 overflow-hidden"
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
    >
      {/* Animated Background */}
      <div className="absolute inset-0">
        {/* Gradient orbs */}
        <motion.div
          className="absolute top-0 left-1/4 w-[600px] h-[600px] rounded-full"
          style={{
            background: 'radial-gradient(circle, rgba(99, 102, 241, 0.15) 0%, transparent 70%)',
          }}
          animate={{
            x: [0, 50, 0],
            y: [0, 30, 0],
            scale: [1, 1.1, 1],
          }}
          transition={{
            duration: 10,
            repeat: Infinity,
            ease: 'easeInOut',
          }}
        />
        <motion.div
          className="absolute bottom-0 right-1/4 w-[500px] h-[500px] rounded-full"
          style={{
            background: 'radial-gradient(circle, rgba(139, 92, 246, 0.12) 0%, transparent 70%)',
          }}
          animate={{
            x: [0, -40, 0],
            y: [0, -20, 0],
            scale: [1, 1.15, 1],
          }}
          transition={{
            duration: 12,
            repeat: Infinity,
            ease: 'easeInOut',
            delay: 2,
          }}
        />

        {/* Particle effect */}
        <ParticleBackground
          particleCount={30}
          color="rgba(99, 102, 241, 0.2)"
          className="absolute inset-0"
        />
      </div>

      <div className="container mx-auto px-4 relative z-10">
        {/* Section Header */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-center mb-12"
        >
          <motion.div
            className="blog-section-pill inline-flex items-center gap-2 px-4 py-2 rounded-full bg-brand-500/10 text-brand-600 dark:text-brand-400 text-sm font-medium mb-4 border border-brand-500/20"
            whileHover={{ scale: 1.05 }}
          >
            <TrendingUp className="h-4 w-4" />
            Featured Stories
          </motion.div>
          <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold tracking-tight">
            <span className="bg-clip-text text-transparent bg-gradient-to-r from-foreground via-foreground to-muted-foreground">
              Discover Our Latest
            </span>
            <br />
            <span className="bg-clip-text text-transparent bg-gradient-to-r from-brand-600 via-violet-600 to-purple-600">
              Insights & Updates
            </span>
          </h2>
        </motion.div>

        {/* Main Carousel Card */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.2 }}
          className="relative max-w-6xl mx-auto"
        >
          {/* Glassmorphism Card */}
          <div className="
            relative overflow-hidden rounded-3xl
            bg-white/5 dark:bg-black/20 backdrop-blur-xl
            border border-white/20 dark:border-white/10
            shadow-[0_25px_80px_-20px_rgba(99,102,241,0.25)]
          ">
            {/* Animated border gradient */}
            <div className="absolute inset-0 rounded-3xl p-[1px] bg-gradient-to-br from-brand-500/30 via-transparent to-violet-500/30 pointer-events-none" />

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-0">
              {/* Image Section */}
              <div className="relative h-72 lg:h-auto lg:min-h-[450px] overflow-hidden">
                <AnimatePresence mode="wait" custom={direction}>
                  <motion.div
                    key={currentIndex}
                    custom={direction}
                    variants={slideVariants}
                    initial="enter"
                    animate="center"
                    exit="exit"
                    transition={{ duration: 0.5, ease: [0.25, 0.46, 0.45, 0.94] }}
                    className="absolute inset-0"
                  >
                    {currentPost.featured_image ? (
                      <img
                        src={currentPost.featured_image}
                        alt={currentPost.title}
                        className="w-full h-full object-cover"
                      />
                    ) : (
                      <div className="
                        w-full h-full bg-gradient-to-br from-brand-500/20 via-violet-500/20 to-purple-500/20
                        flex items-center justify-center
                      ">
                        <motion.span
                          className="text-8xl"
                          animate={{ scale: [1, 1.1, 1], rotate: [0, 5, -5, 0] }}
                          transition={{ duration: 4, repeat: Infinity }}
                        >
                          📝
                        </motion.span>
                      </div>
                    )}
                    {/* Image overlay with gradient */}
                    <div className="absolute inset-0 bg-gradient-to-t from-black/70 via-black/30 to-transparent lg:bg-gradient-to-r lg:from-transparent lg:via-transparent lg:to-black/50" />
                  </motion.div>
                </AnimatePresence>

                {/* Featured badge on image */}
                <motion.div
                  className="absolute top-6 left-6"
                  initial={{ opacity: 0, scale: 0.8 }}
                  animate={{ opacity: 1, scale: 1 }}
                  transition={{ delay: 0.3 }}
                >
                  <span className="
                    inline-flex items-center gap-2 px-4 py-2 rounded-full
                    bg-gradient-to-r from-brand-500 to-violet-600
                    text-white text-sm font-semibold shadow-lg shadow-brand-500/30
                  ">
                    <Sparkles className="h-4 w-4" />
                    Featured
                  </span>
                </motion.div>
              </div>

              {/* Content Section */}
              <div className="relative p-8 lg:p-12 flex flex-col justify-center bg-gradient-to-br from-background/80 to-background/40">
                <AnimatePresence mode="wait">
                  <motion.div
                    key={currentIndex}
                    variants={contentVariants}
                    initial="enter"
                    animate="center"
                    exit="exit"
                    transition={{ duration: 0.4, ease: [0.25, 0.46, 0.45, 0.94] }}
                    className="space-y-6"
                  >
                    {/* Category Tags */}
                    <div className="flex flex-wrap gap-2">
                      {currentPost.tags.slice(0, 3).map((tag, i) => (
                        <motion.span
                          key={tag}
                          className="blog-category-tag
                            inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full
                            text-xs font-semibold uppercase tracking-wider
                            bg-brand-500/10 text-brand-600 dark:text-brand-400
                            border border-brand-500/20
                          "
                          initial={{ opacity: 0, x: -20 }}
                          animate={{ opacity: 1, x: 0 }}
                          transition={{ delay: i * 0.1 }}
                        >
                          {tag}
                        </motion.span>
                      ))}
                    </div>

                    {/* Title */}
                    <h2 className="text-2xl sm:text-3xl lg:text-4xl font-bold leading-tight tracking-tight">
                      <Link
                        to={`/blog/${currentPost.slug}`}
                        className="
                          text-foreground hover:text-brand-600 dark:hover:text-brand-400
                          transition-colors duration-300
                        "
                      >
                        {currentPost.title}
                      </Link>
                    </h2>

                    {/* Excerpt */}
                    <p className="text-lg text-muted-foreground leading-relaxed">
                      {currentPost.excerpt || currentPost.content.substring(0, 280) + '...'}
                    </p>

                    {/* Author Info */}
                    <div className="flex items-center gap-4 py-4 border-y border-border/30">
                      <motion.div
                        className="
                          w-14 h-14 rounded-full
                          bg-gradient-to-br from-brand-500 to-violet-600
                          flex items-center justify-center text-white font-bold text-lg
                          shadow-lg shadow-brand-500/25
                          ring-3 ring-white dark:ring-gray-800
                        "
                        whileHover={{ scale: 1.1, rotate: 5 }}
                      >
                        {getAuthorAvatar(currentPost.author)}
                      </motion.div>
                      <div>
                        <p className="font-semibold text-foreground">{currentPost.author}</p>
                        <div className="flex items-center gap-3 text-sm text-muted-foreground">
                          <span>{getRelativeTime(currentPost.published_at || currentPost.created_at)}</span>
                          <span className="w-1 h-1 rounded-full bg-muted-foreground" />
                          <span className="flex items-center gap-1">
                            <Clock className="h-4 w-4" />
                            {formatReadingTime(calculateReadingTime(currentPost.content))}
                          </span>
                        </div>
                      </div>
                    </div>

                    {/* CTA Button */}
                    <motion.div
                      whileHover={{ scale: 1.02 }}
                      whileTap={{ scale: 0.98 }}
                    >
                      <Link
                        to={`/blog/${currentPost.slug}`}
                        className="
                          inline-flex items-center gap-3 px-8 py-4 rounded-xl
                          bg-gradient-to-r from-brand-600 to-violet-600
                          text-white font-semibold text-lg
                          shadow-xl shadow-brand-500/25 hover:shadow-brand-500/40
                          transition-all duration-300 group
                        "
                      >
                        Read Full Article
                        <ArrowRight className="h-5 w-5 group-hover:translate-x-1 transition-transform" />
                      </Link>
                    </motion.div>
                  </motion.div>
                </AnimatePresence>
              </div>
            </div>

            {/* Navigation Controls */}
            {featuredPosts.length > 1 && (
              <>
                {/* Prev/Next Buttons */}
                <div className="absolute top-1/2 -translate-y-1/2 left-4 right-4 flex justify-between pointer-events-none">
                  <motion.button
                    onClick={goToPrev}
                    className="blog-carousel-arrows
                      pointer-events-auto w-12 h-12 rounded-full
                      bg-white/90 dark:bg-black/80 backdrop-blur-md
                      border border-border/50 shadow-lg
                      flex items-center justify-center
                      text-foreground hover:text-brand-600
                      transition-colors duration-200
                    "
                    whileHover={{ scale: 1.1, x: -2 }}
                    whileTap={{ scale: 0.95 }}
                  >
                    <ChevronLeft className="h-6 w-6" />
                  </motion.button>
                  <motion.button
                    onClick={goToNext}
                    className="blog-carousel-arrows
                      pointer-events-auto w-12 h-12 rounded-full
                      bg-white/90 dark:bg-black/80 backdrop-blur-md
                      border border-border/50 shadow-lg
                      flex items-center justify-center
                      text-foreground hover:text-brand-600
                      transition-colors duration-200
                    "
                    whileHover={{ scale: 1.1, x: 2 }}
                    whileTap={{ scale: 0.95 }}
                  >
                    <ChevronRight className="h-6 w-6" />
                  </motion.button>
                </div>

                {/* Bottom Controls */}
                <div className="absolute bottom-6 left-1/2 -translate-x-1/2 flex items-center gap-4">
                  {/* Dots */}
                  <div className="flex items-center gap-2">
                    {featuredPosts.map((_, index) => (
                      <motion.button
                        key={index}
                        onClick={() => goToSlide(index)}
                        className={`
                          h-2 rounded-full transition-all duration-300
                          ${index === currentIndex
                            ? 'w-8 bg-brand-500'
                            : 'w-2 bg-muted-foreground/30 hover:bg-muted-foreground/50'
                          }
                        `}
                        whileHover={{ scale: 1.2 }}
                        whileTap={{ scale: 0.9 }}
                      />
                    ))}
                  </div>

                  {/* Play/Pause */}
                  <motion.button
                    onClick={() => setIsAutoPlaying(!isAutoPlaying)}
                    className="
                      w-8 h-8 rounded-full
                      bg-muted/80 hover:bg-muted
                      flex items-center justify-center
                      text-muted-foreground hover:text-foreground
                      transition-colors duration-200
                    "
                    whileHover={{ scale: 1.1 }}
                    whileTap={{ scale: 0.95 }}
                  >
                    {isAutoPlaying ? (
                      <Pause className="h-4 w-4" />
                    ) : (
                      <Play className="h-4 w-4" />
                    )}
                  </motion.button>
                </div>
              </>
            )}
          </div>
        </motion.div>
      </div>
    </section>
  );
}

export default FeaturedCarousel;
