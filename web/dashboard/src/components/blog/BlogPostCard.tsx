'use client';

import { useState, useRef } from 'react';
import { motion, useMotionValue, useSpring, useTransform } from 'framer-motion';
import { Link } from 'react-router-dom';
import { Clock, ArrowRight, Sparkles, Eye } from 'lucide-react';
import { BlogPost } from '@/api/content';
import { calculateReadingTime, formatReadingTime, getRelativeTime, getAuthorAvatar } from '@/pages/BlogPage/utils';
import { ShinyButton, TooltipProvider } from '@/components/ui';
import { BookmarkButton } from './BookmarkButton';

interface BlogPostCardProps {
  post: BlogPost;
  index?: number;
  variant?: 'default' | 'compact' | 'featured';
  showExcerpt?: boolean;
  className?: string;
}

export function BlogPostCard({
  post,
  index = 0,
  variant = 'default',
  showExcerpt = true,
  className = ''
}: BlogPostCardProps) {
  const [isHovered, setIsHovered] = useState(false);
  const cardRef = useRef<HTMLDivElement>(null);

  const isCompact = variant === 'compact';
  const isFeatured = variant === 'featured';

  const imageHeight = isCompact ? 'h-36' : isFeatured ? 'h-64' : 'h-52';
  const padding = isCompact ? 'p-4' : 'p-6';
  const titleSize = isCompact ? 'text-base' : isFeatured ? 'text-2xl' : 'text-lg';

  // 3D tilt effect
  const x = useMotionValue(0);
  const y = useMotionValue(0);

  const mouseXSpring = useSpring(x, { stiffness: 500, damping: 100 });
  const mouseYSpring = useSpring(y, { stiffness: 500, damping: 100 });

  const rotateX = useTransform(mouseYSpring, [-0.5, 0.5], ['8deg', '-8deg']);
  const rotateY = useTransform(mouseXSpring, [-0.5, 0.5], ['-8deg', '8deg']);
  const glowX = useTransform(mouseXSpring, [-0.5, 0.5], ['0%', '100%']);
  const glowY = useTransform(mouseYSpring, [-0.5, 0.5], ['0%', '100%']);

  const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!cardRef.current) return;
    const rect = cardRef.current.getBoundingClientRect();
    const width = rect.width;
    const height = rect.height;
    const mouseX = e.clientX - rect.left;
    const mouseY = e.clientY - rect.top;
    const xPct = mouseX / width - 0.5;
    const yPct = mouseY / height - 0.5;
    x.set(xPct);
    y.set(yPct);
  };

  const handleMouseLeave = () => {
    setIsHovered(false);
    x.set(0);
    y.set(0);
  };

  // Generate a gradient based on the post id for visual variety
  const gradients = [
    'from-violet-500/20 via-purple-500/10 to-indigo-500/20',
    'from-blue-500/20 via-cyan-500/10 to-teal-500/20',
    'from-rose-500/20 via-pink-500/10 to-fuchsia-500/20',
    'from-amber-500/20 via-orange-500/10 to-red-500/20',
    'from-emerald-500/20 via-green-500/10 to-lime-500/20',
    'from-cyan-500/20 via-sky-500/10 to-blue-500/20',
  ];
  const gradientIndex = parseInt(post.id.slice(-1), 16) % gradients.length;
  const cardGradient = gradients[gradientIndex];

  return (
    <motion.div
      ref={cardRef}
      initial={{ opacity: 0, y: 30 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{
        duration: 0.5,
        delay: index * 0.1,
        ease: [0.25, 0.46, 0.45, 0.94]
      }}
      style={{
        rotateX: isHovered ? rotateX : 0,
        rotateY: isHovered ? rotateY : 0,
        transformStyle: 'preserve-3d',
      }}
      onMouseMove={handleMouseMove}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={handleMouseLeave}
      className={`${className} perspective-1000`}
    >
      <div className={`
        blog-post-card h-full overflow-hidden rounded-2xl
        border border-border/40 bg-card/60 backdrop-blur-xl
        shadow-[0_8px_30px_rgb(0,0,0,0.08)] dark:shadow-[0_8px_30px_rgb(0,0,0,0.2)]
        hover:shadow-[0_20px_50px_rgb(99,102,241,0.15)] dark:hover:shadow-[0_20px_50px_rgb(99,102,241,0.2)]
        transition-all duration-500 ease-out
        group relative
      `}>
        {/* Animated gradient border */}
        <motion.div
          className="absolute -inset-[1px] rounded-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none"
          style={{
            background: `radial-gradient(circle at ${glowX} ${glowY}, rgba(99, 102, 241, 0.4), transparent 50%)`,
          }}
        />

        {/* Gradient overlay on hover */}
        <div className={`
          absolute inset-0 bg-gradient-to-br ${cardGradient}
          opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none rounded-2xl
        `} />

        {/* Featured Image */}
        <Link
          to={`/blog/${post.slug}`}
          className={`block relative ${imageHeight} overflow-hidden`}
        >
          {post.featured_image ? (
            <motion.img
              src={post.featured_image}
              alt={post.title}
              className="w-full h-full object-cover"
              animate={{
                scale: isHovered ? 1.08 : 1,
              }}
              transition={{ duration: 0.6, ease: [0.25, 0.46, 0.45, 0.94] }}
            />
          ) : (
            <div className={`
              w-full h-full bg-gradient-to-br ${cardGradient}
              flex items-center justify-center
            `}>
              <motion.span
                className="text-5xl"
                animate={{
                  scale: isHovered ? 1.2 : 1,
                  rotate: isHovered ? 10 : 0,
                }}
                transition={{ duration: 0.4 }}
              >
                📝
              </motion.span>
            </div>
          )}

          {/* Image overlay with animated gradient */}
          <div className="absolute inset-0 bg-gradient-to-t from-black/60 via-black/20 to-transparent opacity-60 group-hover:opacity-80 transition-opacity duration-500" />

          {/* Floating tags on image - blog-card-top-tags for light-mode contrast override */}
          <motion.div
            className="blog-card-top-tags absolute top-4 left-4 flex flex-wrap gap-2"
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.1 + 0.2 }}
          >
            {post.tags.slice(0, 2).map((tag) => (
              <span
                key={tag}
                className="
                  inline-flex items-center gap-1 px-3 py-1 rounded-full text-xs font-semibold
                  bg-white/20 backdrop-blur-md text-white border border-white/30
                  shadow-lg
                "
              >
                <Sparkles className="h-3 w-3" />
                {tag}
              </span>
            ))}
          </motion.div>

          {/* View indicator on hover */}
          <motion.div
            className="absolute inset-0 flex items-center justify-center"
            initial={{ opacity: 0 }}
            animate={{ opacity: isHovered ? 1 : 0 }}
            transition={{ duration: 0.3 }}
          >
            <div className="
              flex items-center gap-2 px-4 py-2 rounded-full
              bg-white/20 backdrop-blur-md text-white font-medium
              border border-white/30 shadow-xl
            ">
              <Eye className="h-4 w-4" />
              Read Article
            </div>
          </motion.div>
        </Link>

        <div className={`${padding} relative z-10`}>
          {/* Title with enhanced typography */}
          <h3 className={`${titleSize} font-bold mb-3 leading-snug tracking-tight`}>
            <Link
              to={`/blog/${post.slug}`}
              className="
                text-foreground hover:text-brand-600 dark:hover:text-brand-400
                transition-colors duration-300 line-clamp-2
                group-hover:text-transparent group-hover:bg-clip-text group-hover:bg-gradient-to-r
                group-hover:from-brand-600 group-hover:to-violet-600
                dark:group-hover:from-brand-400 dark:group-hover:to-violet-400
              "
            >
              {post.title}
            </Link>
          </h3>

          {/* Excerpt with fade effect */}
          {showExcerpt && (
            <p className="text-muted-foreground text-sm mb-4 line-clamp-3 leading-relaxed">
              {post.excerpt || post.content.substring(0, 150) + '...'}
            </p>
          )}

          {/* Author and Meta with enhanced styling */}
          <div className="flex items-center justify-between gap-3 mb-4">
            <div className="flex items-center gap-3 min-w-0">
              <motion.div
                className="
                  shrink-0 w-10 h-10 rounded-full
                  bg-gradient-to-br from-brand-500 to-violet-600
                  flex items-center justify-center text-white font-bold text-sm
                  shadow-lg shadow-brand-500/25
                  ring-2 ring-white dark:ring-gray-800
                "
                whileHover={{ scale: 1.1, rotate: 5 }}
                transition={{ type: 'spring', stiffness: 400, damping: 17 }}
              >
                {getAuthorAvatar(post.author)}
              </motion.div>
              <div className="min-w-0">
                <p className="font-semibold text-foreground text-sm truncate">{post.author}</p>
                <p className="text-muted-foreground text-xs flex items-center gap-1">
                  {getRelativeTime(post.published_at || post.created_at)}
                </p>
              </div>
            </div>
            <div className="
              flex items-center gap-1.5 text-xs text-muted-foreground
              px-2.5 py-1.5 rounded-full bg-muted/50
            ">
              <Clock className="h-3.5 w-3.5" />
              <span className="font-medium">{formatReadingTime(calculateReadingTime(post.content))}</span>
            </div>
          </div>

          {/* Tags with enhanced styling */}
          {post.tags.length > 0 && (
            <div className="flex flex-wrap gap-1.5 mb-4">
              {post.tags.slice(0, isCompact ? 2 : 3).map((tag, i) => (
                <motion.span
                  key={tag}
                  className="
                    inline-flex px-2.5 py-1 rounded-full text-xs font-medium
                    bg-gradient-to-r from-muted/80 to-muted/50
                    text-muted-foreground border border-border/50
                    hover:border-brand-500/30 hover:text-brand-600 dark:hover:text-brand-400
                    transition-colors duration-200 cursor-pointer
                  "
                  whileHover={{ scale: 1.05, y: -2 }}
                  transition={{ type: 'spring', stiffness: 400, damping: 17 }}
                >
                  {tag}
                </motion.span>
              ))}
            </div>
          )}

          {/* Actions Row with enhanced buttons */}
          <div className="flex items-center justify-between pt-2 border-t border-border/30">
            {/* Read More Button */}
            <ShinyButton asChild size={isCompact ? 'sm' : 'default'} variant="ghost">
              <Link
                to={`/blog/${post.slug}`}
                className="inline-flex items-center gap-1.5 group/btn"
              >
                <span>Read more</span>
                <motion.span
                  animate={{ x: isHovered ? 4 : 0 }}
                  transition={{ duration: 0.3 }}
                >
                  <ArrowRight className="h-3.5 w-3.5" />
                </motion.span>
              </Link>
            </ShinyButton>

            {/* Bookmark Button */}
            {!isCompact && (
              <TooltipProvider>
                <BookmarkButton
                  postId={post.id}
                  postTitle={post.title}
                  size="sm"
                />
              </TooltipProvider>
            )}
          </div>
        </div>
      </div>
    </motion.div>
  );
}

export default BlogPostCard;
