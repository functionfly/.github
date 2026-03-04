'use client';

import { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Search, X, Sparkles } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

interface SearchBarProps {
  onSearch: (query: string) => void;
  placeholder?: string;
  debounceMs?: number;
}

export function SearchBar({
  onSearch,
  placeholder = 'Search articles...',
  debounceMs = 300
}: SearchBarProps) {
  const [query, setQuery] = useState('');
  const [isFocused, setIsFocused] = useState(false);

  // Debounced search
  useEffect(() => {
    const timer = setTimeout(() => {
      if (query.trim()) {
        onSearch(query.trim());
      } else {
        onSearch('');
      }
    }, debounceMs);

    return () => clearTimeout(timer);
  }, [query, onSearch, debounceMs]);

  const handleClear = useCallback(() => {
    setQuery('');
    onSearch('');
  }, [onSearch]);

  return (
    <motion.div
      className="relative w-full max-w-xl"
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
    >
      <motion.div
        className={`
          relative flex items-center
          rounded-2xl bg-background/80 backdrop-blur-xl
          border-2 transition-all duration-300
          ${isFocused
            ? 'border-brand-500/50 shadow-[0_0_30px_rgba(99,102,241,0.2)]'
            : 'border-border/50 shadow-lg shadow-black/5'
          }
        `}
        animate={{
          scale: isFocused ? 1.02 : 1,
        }}
        transition={{ type: 'spring', stiffness: 400, damping: 25 }}
      >
        {/* Search icon with animation */}
        <motion.div
          className="absolute left-5"
          animate={{
            scale: isFocused ? 1.1 : 1,
            rotate: isFocused ? [0, -10, 10, 0] : 0,
          }}
          transition={{ duration: 0.3 }}
        >
          <Search className={`
            h-5 w-5 transition-colors duration-300
            ${isFocused ? 'text-brand-500' : 'text-muted-foreground'}
          `} />
        </motion.div>

        <Input
          type="search"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onFocus={() => setIsFocused(true)}
          onBlur={() => setIsFocused(false)}
          placeholder={placeholder}
          className="
            border-0 bg-transparent pl-14 pr-12 py-6
            focus-visible:ring-0 focus-visible:ring-offset-0
            rounded-2xl text-base placeholder:text-muted-foreground/70
          "
          aria-label="Search blog posts"
        />

        {/* Clear button with animation */}
        <AnimatePresence>
          {query && (
            <motion.div
              initial={{ opacity: 0, scale: 0.8, x: 10 }}
              animate={{ opacity: 1, scale: 1, x: 0 }}
              exit={{ opacity: 0, scale: 0.8, x: 10 }}
              className="absolute right-2"
            >
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={handleClear}
                className="
                  h-10 w-10 p-0 rounded-xl
                  hover:bg-destructive/10 hover:text-destructive
                  transition-colors duration-200
                "
                aria-label="Clear search"
              >
                <X className="h-5 w-5" />
              </Button>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Focus glow effect */}
        {isFocused && (
          <motion.div
            className="absolute -inset-[1px] rounded-2xl pointer-events-none"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            style={{
              background: 'linear-gradient(135deg, rgba(99, 102, 241, 0.3), rgba(139, 92, 246, 0.3))',
              zIndex: -1,
            }}
          />
        )}
      </motion.div>

      {/* Animated hint text */}
      <motion.p
        className="text-center text-sm text-muted-foreground mt-3"
        initial={{ opacity: 0 }}
        animate={{ opacity: isFocused ? 1 : 0 }}
        transition={{ duration: 0.2 }}
      >
        <span className="inline-flex items-center gap-1">
          <Sparkles className="h-3 w-3 text-brand-500" />
          Press Enter to search or wait for instant results
        </span>
      </motion.p>
    </motion.div>
  );
}

export default SearchBar;
