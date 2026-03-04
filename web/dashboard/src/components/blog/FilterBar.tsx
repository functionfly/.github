'use client';

import { useState, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { X, Filter, Hash } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface FilterBarProps {
  availableTags: string[];
  selectedTags: string[];
  onTagSelect: (tag: string) => void;
  onClearAll: () => void;
}

// Vibrant color schemes for tags
const tagColors: Record<string, { bg: string; text: string; border: string; glow: string }> = {
  'tutorial': {
    bg: 'from-blue-500/20 to-cyan-500/10',
    text: 'text-blue-600 dark:text-blue-400',
    border: 'border-blue-500/30',
    glow: 'shadow-blue-500/20',
  },
  'news': {
    bg: 'from-rose-500/20 to-pink-500/10',
    text: 'text-rose-600 dark:text-rose-400',
    border: 'border-rose-500/30',
    glow: 'shadow-rose-500/20',
  },
  'update': {
    bg: 'from-emerald-500/20 to-green-500/10',
    text: 'text-emerald-600 dark:text-emerald-400',
    border: 'border-emerald-500/30',
    glow: 'shadow-emerald-500/20',
  },
  'feature': {
    bg: 'from-violet-500/20 to-purple-500/10',
    text: 'text-violet-600 dark:text-violet-400',
    border: 'border-violet-500/30',
    glow: 'shadow-violet-500/20',
  },
  'guide': {
    bg: 'from-amber-500/20 to-orange-500/10',
    text: 'text-amber-600 dark:text-amber-400',
    border: 'border-amber-500/30',
    glow: 'shadow-amber-500/20',
  },
  'announcement': {
    bg: 'from-brand-500/20 to-indigo-500/10',
    text: 'text-brand-600 dark:text-brand-400',
    border: 'border-brand-500/30',
    glow: 'shadow-brand-500/20',
  },
};

const getTagColors = (tag: string) => {
  const normalizedTag = tag.toLowerCase();
  for (const [key, colors] of Object.entries(tagColors)) {
    if (normalizedTag.includes(key)) {
      return colors;
    }
  }
  // Default colors
  return {
    bg: 'from-muted/80 to-muted/40',
    text: 'text-muted-foreground',
    border: 'border-border/50',
    glow: 'shadow-black/5',
  };
};

export function FilterBar({
  availableTags,
  selectedTags,
  onTagSelect,
  onClearAll,
}: FilterBarProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  const handleTagClick = useCallback((tag: string) => {
    onTagSelect(tag);
  }, [onTagSelect]);

  if (availableTags.length === 0) {
    return null;
  }

  // Show first 6 tags by default, expand to show all
  const displayTags = isExpanded ? availableTags : availableTags.slice(0, 6);
  const hasMoreTags = availableTags.length > 6;

  return (
    <motion.div
      className="w-full"
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, delay: 0.1 }}
    >
      {/* Header with icon */}
      <div className="flex items-center justify-center gap-2 mb-4">
        <motion.div
          className="flex items-center gap-2 text-muted-foreground"
          whileHover={{ scale: 1.05 }}
        >
          <Filter className="h-4 w-4" />
          <span className="text-sm font-medium">Filter by topic</span>
        </motion.div>
      </div>

      {/* Active Filters Display */}
      <AnimatePresence mode="wait">
        {selectedTags.length > 0 && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            className="flex flex-wrap items-center justify-center gap-2 mb-5"
          >
            <span className="text-sm text-muted-foreground">Active:</span>
            {selectedTags.map((tag) => {
              const colors = getTagColors(tag);
              return (
                <motion.button
                  key={tag}
                  onClick={() => onTagSelect(tag)}
                  className={`
                    inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full
                    bg-gradient-to-r ${colors.bg} ${colors.text}
                    border ${colors.border}
                    text-sm font-medium
                    hover:shadow-lg ${colors.glow}
                    transition-all duration-200
                  `}
                  initial={{ opacity: 0, scale: 0.8 }}
                  animate={{ opacity: 1, scale: 1 }}
                  exit={{ opacity: 0, scale: 0.8 }}
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  layout
                >
                  {tag}
                  <X className="h-3.5 w-3.5" />
                </motion.button>
              );
            })}
            <motion.button
              onClick={onClearAll}
              className="
                text-sm text-muted-foreground hover:text-foreground
                underline underline-offset-4
                transition-colors duration-200
              "
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
            >
              Clear all
            </motion.button>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Tag Pills */}
      <div
        className="flex flex-wrap justify-center gap-2.5"
        role="group"
        aria-label="Filter by tag"
      >
        {displayTags.map((tag, index) => {
          const isSelected = selectedTags.includes(tag);
          const colors = getTagColors(tag);

          return (
            <motion.button
              key={tag}
              onClick={() => handleTagClick(tag)}
              className={`
                relative overflow-hidden rounded-full px-4 py-2 text-sm font-medium
                transition-all duration-300
                ${isSelected
                  ? `bg-gradient-to-r ${colors.bg} ${colors.text} border-2 ${colors.border} shadow-lg ${colors.glow}`
                  : `bg-muted/40 text-muted-foreground border border-border/50 hover:border-brand-500/30 hover:text-brand-600 dark:hover:text-brand-400`
                }
              `}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.05 }}
              whileHover={{
                scale: 1.08,
                y: -2,
              }}
              whileTap={{ scale: 0.95 }}
              aria-pressed={isSelected}
            >
              {/* Shimmer effect on hover */}
              <motion.div
                className="absolute inset-0 bg-gradient-to-r from-transparent via-white/20 to-transparent"
                initial={{ x: '-100%' }}
                whileHover={{ x: '100%' }}
                transition={{ duration: 0.6 }}
              />

              <span className="relative flex items-center gap-1.5">
                <Hash className="h-3.5 w-3.5 opacity-60" />
                {tag}
              </span>
            </motion.button>
          );
        })}

        {/* Show more/less button */}
        {hasMoreTags && (
          <motion.button
            onClick={() => setIsExpanded(!isExpanded)}
            className="
              rounded-full px-4 py-2 text-sm font-medium
              bg-gradient-to-r from-brand-500/10 to-violet-500/10
              text-brand-600 dark:text-brand-400
              border border-brand-500/20
              hover:shadow-lg hover:shadow-brand-500/20
              transition-all duration-300
            "
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
          >
            {isExpanded ? 'Show less' : `+${availableTags.length - 6} more`}
          </motion.button>
        )}
      </div>
    </motion.div>
  );
}

export default FilterBar;
