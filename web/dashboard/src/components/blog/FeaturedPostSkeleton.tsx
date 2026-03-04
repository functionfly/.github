'use client';

import { motion } from 'framer-motion';

export function FeaturedPostSkeleton() {
  return (
    <motion.div
      className="
        relative overflow-hidden rounded-3xl
        border border-border/30 bg-card/50
        shadow-xl
      "
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
    >
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-0">
        {/* Image Skeleton */}
        <div className="relative h-72 lg:h-auto lg:min-h-[450px] overflow-hidden bg-muted/50">
          <ShimmerEffect />

          {/* Featured badge placeholder */}
          <div className="absolute top-6 left-6">
            <div className="w-28 h-8 rounded-full bg-white/20 backdrop-blur-md relative overflow-hidden">
              <ShimmerEffect />
            </div>
          </div>
        </div>

        {/* Content Skeleton */}
        <div className="p-8 lg:p-12 space-y-6">
          {/* Tags */}
          <div className="flex gap-2">
            <div className="w-20 h-7 rounded-full bg-muted/70 relative overflow-hidden">
              <ShimmerEffect />
            </div>
            <div className="w-24 h-7 rounded-full bg-muted/70 relative overflow-hidden">
              <ShimmerEffect />
            </div>
          </div>

          {/* Title */}
          <div className="space-y-3">
            <div className="h-10 rounded-lg bg-muted/70 w-full relative overflow-hidden">
              <ShimmerEffect />
            </div>
            <div className="h-10 rounded-lg bg-muted/70 w-4/5 relative overflow-hidden">
              <ShimmerEffect />
            </div>
          </div>

          {/* Excerpt */}
          <div className="space-y-3">
            <div className="h-5 rounded bg-muted/50 w-full relative overflow-hidden">
              <ShimmerEffect />
            </div>
            <div className="h-5 rounded bg-muted/50 w-full relative overflow-hidden">
              <ShimmerEffect />
            </div>
            <div className="h-5 rounded bg-muted/50 w-3/4 relative overflow-hidden">
              <ShimmerEffect />
            </div>
          </div>

          {/* Author */}
          <div className="flex items-center gap-4 py-4 border-y border-border/30">
            <div className="w-14 h-14 rounded-full bg-muted/70 relative overflow-hidden">
              <ShimmerEffect />
            </div>
            <div className="space-y-2">
              <div className="h-5 rounded bg-muted/70 w-32 relative overflow-hidden">
                <ShimmerEffect />
              </div>
              <div className="h-4 rounded bg-muted/50 w-48 relative overflow-hidden">
                <ShimmerEffect />
              </div>
            </div>
          </div>

          {/* CTA Button */}
          <div className="w-48 h-14 rounded-xl bg-muted/70 relative overflow-hidden">
            <ShimmerEffect />
          </div>
        </div>
      </div>

      {/* Navigation dots */}
      <div className="absolute bottom-6 left-1/2 -translate-x-1/2 flex items-center gap-2">
        <div className="w-8 h-2 rounded-full bg-muted/70 relative overflow-hidden">
          <ShimmerEffect />
        </div>
        <div className="w-2 h-2 rounded-full bg-muted/50 relative overflow-hidden">
          <ShimmerEffect />
        </div>
        <div className="w-2 h-2 rounded-full bg-muted/50 relative overflow-hidden">
          <ShimmerEffect />
        </div>
      </div>
    </motion.div>
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

export default FeaturedPostSkeleton;
