'use client';

import { motion } from 'framer-motion';

interface BlogCardSkeletonProps {
  count?: number;
}

export function BlogCardSkeleton({ count = 6 }: BlogCardSkeletonProps) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 lg:gap-8">
      {Array.from({ length: count }).map((_, index) => (
        <motion.div
          key={index}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.3, delay: index * 0.05 }}
        >
          <div className="h-full overflow-hidden rounded-2xl border border-border/50 bg-card/80 shadow-lg">
            {/* Image skeleton */}
            <div className="h-52 w-full bg-muted/50 animate-pulse" />
            
            <div className="p-6">
              {/* Title skeleton */}
              <div className="h-6 w-3/4 bg-muted/50 rounded animate-pulse mb-3" />
              
              {/* Excerpt skeleton */}
              <div className="space-y-2 mb-4">
                <div className="h-4 w-full bg-muted/50 rounded animate-pulse" />
                <div className="h-4 w-5/6 bg-muted/50 rounded animate-pulse" />
                <div className="h-4 w-4/6 bg-muted/50 rounded animate-pulse" />
              </div>
              
              {/* Author skeleton */}
              <div className="flex items-center gap-3 mb-4">
                <div className="w-8 h-8 rounded-full bg-muted/50 animate-pulse" />
                <div className="space-y-1">
                  <div className="h-3 w-20 bg-muted/50 rounded animate-pulse" />
                  <div className="h-2 w-16 bg-muted/50 rounded animate-pulse" />
                </div>
              </div>
              
              {/* Tags skeleton */}
              <div className="flex gap-2">
                <div className="h-6 w-16 bg-muted/50 rounded-full animate-pulse" />
                <div className="h-6 w-20 bg-muted/50 rounded-full animate-pulse" />
              </div>
            </div>
          </div>
        </motion.div>
      ))}
    </div>
  );
}
