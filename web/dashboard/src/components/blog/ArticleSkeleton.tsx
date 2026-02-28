'use client';

import { motion } from 'framer-motion';

export function ArticleSkeleton() {
  return (
    <div className="min-h-screen bg-background">
      {/* Header skeleton */}
      <header className="pt-20 pb-12">
        <div className="container mx-auto px-4">
          <motion.div
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            className="max-w-3xl mx-auto text-center"
          >
            {/* Title skeleton */}
            <div className="space-y-3 mb-6">
              <div className="h-10 w-3/4 mx-auto bg-muted/50 rounded animate-pulse" />
              <div className="h-10 w-1/2 mx-auto bg-muted/50 rounded animate-pulse" />
            </div>

            {/* Meta skeleton */}
            <div className="flex flex-wrap items-center justify-center gap-6">
              <div className="flex items-center gap-2">
                <div className="w-8 h-8 rounded-full bg-muted/50 animate-pulse" />
                <div className="h-4 w-24 bg-muted/50 rounded animate-pulse" />
              </div>
              <div className="h-4 w-20 bg-muted/50 rounded animate-pulse" />
              <div className="h-4 w-16 bg-muted/50 rounded animate-pulse" />
            </div>
          </motion.div>
        </div>
      </header>

      <div className="container mx-auto px-4 pb-20">
        <div className="max-w-3xl mx-auto">
          {/* Featured image skeleton */}
          <motion.div
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            className="mb-10"
          >
            <div className="aspect-[21/9] bg-muted/50 rounded-2xl animate-pulse" />
          </motion.div>

          {/* Tags skeleton */}
          <motion.div
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            className="flex justify-center gap-2 mb-10"
          >
            <div className="h-7 w-16 bg-muted/50 rounded-full animate-pulse" />
            <div className="h-7 w-20 bg-muted/50 rounded-full animate-pulse" />
            <div className="h-7 w-14 bg-muted/50 rounded-full animate-pulse" />
          </motion.div>

          {/* Article content skeleton */}
          <motion.article
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            className="rounded-2xl border border-border/50 bg-card/80 p-8 sm:p-10 md:p-12"
          >
            <div className="space-y-4">
              {/* Paragraphs */}
              {Array.from({ length: 8 }).map((_, i) => (
                <div 
                  key={i} 
                  className="space-y-2"
                  style={{ opacity: 1 - (i * 0.08) }}
                >
                  <div className="h-4 w-full bg-muted/50 rounded animate-pulse" />
                  <div className="h-4 w-full bg-muted/50 rounded animate-pulse" />
                  <div className="h-4 w-3/4 bg-muted/50 rounded animate-pulse" />
                </div>
              ))}

              {/* Code block */}
              <div className="h-40 bg-muted/50 rounded-xl animate-pulse my-6" />

              {/* More paragraphs */}
              {Array.from({ length: 4 }).map((_, i) => (
                <div 
                  key={`para-${i}`} 
                  className="space-y-2"
                >
                  <div className="h-4 w-full bg-muted/50 rounded animate-pulse" />
                  <div className="h-4 w-full bg-muted/50 rounded animate-pulse" />
                  <div className="h-4 w-2/3 bg-muted/50 rounded animate-pulse" />
                </div>
              ))}
            </div>
          </motion.article>

          {/* Back button skeleton */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.4 }}
            className="mt-12 text-center"
          >
            <div className="h-12 w-40 mx-auto bg-muted/50 rounded-full animate-pulse" />
          </motion.div>
        </div>
      </div>
    </div>
  );
}
