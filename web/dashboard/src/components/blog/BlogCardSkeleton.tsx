'use client';

import { motion } from 'framer-motion';

interface BlogCardSkeletonProps {
  count?: number;
  variant?: 'default' | 'compact' | 'featured';
}

export function BlogCardSkeleton({ count = 3, variant = 'default' }: BlogCardSkeletonProps) {
  const isCompact = variant === 'compact';
  const isFeatured = variant === 'featured';

  return (
    <>
      {Array.from({ length: count }).map((_, index) => (
        <motion.div
          key={index}
          className={`
            overflow-hidden rounded-2xl
            border border-border/30 bg-card/50
            ${isFeatured ? 'col-span-full' : ''}
          `}
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3, delay: index * 0.1 }}
        >
          {/* Image Skeleton */}
          <div className={`
            relative overflow-hidden
            ${isCompact ? 'h-36' : isFeatured ? 'h-80' : 'h-52'}
            bg-muted/50
          `}>
            <ShimmerEffect />
          </div>

          <div className={`${isCompact ? 'p-4' : 'p-6'}`}>
            {/* Title Skeleton */}
            <div className={`
              h-6 rounded-lg bg-muted/70 mb-3 relative overflow-hidden
              ${isFeatured ? 'w-3/4' : 'w-full'}
            `}>
              <ShimmerEffect />
            </div>
            {!isCompact && (
              <div className="h-6 rounded-lg bg-muted/70 w-2/3 mb-4 relative overflow-hidden">
                <ShimmerEffect />
              </div>
            )}

            {/* Excerpt Skeleton */}
            {!isCompact && (
              <>
                <div className="h-4 rounded bg-muted/50 mb-2 relative overflow-hidden">
                  <ShimmerEffect />
                </div>
                <div className="h-4 rounded bg-muted/50 mb-2 w-5/6 relative overflow-hidden">
                  <ShimmerEffect />
                </div>
                <div className="h-4 rounded bg-muted/50 mb-4 w-4/6 relative overflow-hidden">
                  <ShimmerEffect />
                </div>
              </>
            )}

            {/* Author Skeleton */}
            <div className="flex items-center gap-3 mb-4">
              <div className="w-10 h-10 rounded-full bg-muted/70 relative overflow-hidden">
                <ShimmerEffect />
              </div>
              <div className="flex-1">
                <div className="h-4 rounded bg-muted/70 w-24 mb-2 relative overflow-hidden">
                  <ShimmerEffect />
                </div>
                <div className="h-3 rounded bg-muted/50 w-16 relative overflow-hidden">
                  <ShimmerEffect />
                </div>
              </div>
            </div>

            {/* Tags Skeleton */}
            <div className="flex gap-2">
              <div className="h-6 rounded-full bg-muted/50 w-16 relative overflow-hidden">
                <ShimmerEffect />
              </div>
              <div className="h-6 rounded-full bg-muted/50 w-20 relative overflow-hidden">
                <ShimmerEffect />
              </div>
              <div className="h-6 rounded-full bg-muted/50 w-14 relative overflow-hidden">
                <ShimmerEffect />
              </div>
            </div>
          </div>
        </motion.div>
      ))}
    </>
  );
}

// Shimmer effect component
function ShimmerEffect() {
  return (
    <motion.div
      className="absolute inset-0 bg-gradient-to-r from-transparent via-white/20 to-transparent"
      animate={{
        x: ['-100%', '100%'],
      }}
      transition={{
        duration: 1.5,
        repeat: Infinity,
        ease: 'linear',
      }}
    />
  );
}

export default BlogCardSkeleton;
