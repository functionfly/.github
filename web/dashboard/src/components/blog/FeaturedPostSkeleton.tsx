'use client';

import { motion } from 'framer-motion';

export function FeaturedPostSkeleton() {
  return (
    <section className="pt-10 pb-20">
      <div className="container mx-auto px-4">
        <motion.div
          initial={{ opacity: 0, y: 24 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
          className="mb-16"
        >
          <div className="overflow-hidden rounded-2xl border border-border/50 bg-card shadow-xl">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-0">
              {/* Featured Image skeleton */}
              <div className="relative h-64 lg:h-full min-h-[320px] bg-muted/50 animate-pulse" />

              {/* Content skeleton */}
              <div className="p-8 lg:p-12 flex flex-col justify-center">
                {/* Featured badge skeleton */}
                <div className="flex items-center gap-2 mb-5">
                  <div className="h-4 w-20 bg-muted/50 rounded animate-pulse" />
                </div>

                {/* Title skeleton */}
                <div className="space-y-2 mb-4">
                  <div className="h-8 w-3/4 bg-muted/50 rounded animate-pulse" />
                  <div className="h-8 w-1/2 bg-muted/50 rounded animate-pulse" />
                </div>

                {/* Excerpt skeleton */}
                <div className="space-y-2 mb-6">
                  <div className="h-4 w-full bg-muted/50 rounded animate-pulse" />
                  <div className="h-4 w-full bg-muted/50 rounded animate-pulse" />
                  <div className="h-4 w-2/3 bg-muted/50 rounded animate-pulse" />
                </div>

                {/* Author skeleton */}
                <div className="flex items-center gap-3 mb-6">
                  <div className="w-10 h-10 rounded-full bg-muted/50 animate-pulse" />
                  <div className="space-y-1">
                    <div className="h-4 w-24 bg-muted/50 rounded animate-pulse" />
                    <div className="h-3 w-32 bg-muted/50 rounded animate-pulse" />
                  </div>
                </div>

                {/* Tags skeleton */}
                <div className="flex gap-2 mb-6">
                  <div className="h-6 w-16 bg-muted/50 rounded-full animate-pulse" />
                  <div className="h-6 w-20 bg-muted/50 rounded-full animate-pulse" />
                  <div className="h-6 w-14 bg-muted/50 rounded-full animate-pulse" />
                </div>

                {/* Button skeleton */}
                <div className="h-12 w-40 bg-muted/50 rounded-full animate-pulse" />
              </div>
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  );
}
