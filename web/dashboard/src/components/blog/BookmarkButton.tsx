'use client';

import { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Bookmark, Check } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';

interface BookmarkButtonProps {
  postId: string;
  postTitle: string;
  size?: 'sm' | 'md' | 'lg';
  showLabel?: boolean;
}

const STORAGE_KEY = 'bookmarked_posts';

interface BookmarkData {
  id: string;
  title: string;
  savedAt: string;
}

export function BookmarkButton({
  postId,
  postTitle,
  size = 'md',
  showLabel = false,
}: BookmarkButtonProps) {
  const [isBookmarked, setIsBookmarked] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  // Load bookmarks from localStorage
  useEffect(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      const bookmarks: BookmarkData[] = stored ? JSON.parse(stored) : [];
      setIsBookmarked(bookmarks.some(b => b.id === postId));
    } catch (err) {
      console.error('Failed to load bookmarks:', err);
    } finally {
      setIsLoading(false);
    }
  }, [postId]);

  const toggleBookmark = useCallback(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      let bookmarks: BookmarkData[] = stored ? JSON.parse(stored) : [];

      if (isBookmarked) {
        // Remove bookmark
        bookmarks = bookmarks.filter(b => b.id !== postId);
      } else {
        // Add bookmark
        bookmarks.push({
          id: postId,
          title: postTitle,
          savedAt: new Date().toISOString(),
        });
      }

      localStorage.setItem(STORAGE_KEY, JSON.stringify(bookmarks));
      setIsBookmarked(!isBookmarked);
    } catch (err) {
      console.error('Failed to toggle bookmark:', err);
    }
  }, [isBookmarked, postId, postTitle]);

  const sizeClasses = {
    sm: 'h-8 w-8',
    md: 'h-9 w-9',
    lg: 'h-10 w-10',
  };

  const iconSizes = {
    sm: 'h-4 w-4',
    md: 'h-4 w-4',
    lg: 'h-5 w-5',
  };

  if (isLoading) {
    return null;
  }

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className={`
            ${sizeClasses[size]} rounded-full transition-all duration-200
            ${isBookmarked
              ? 'bg-brand-500/10 text-brand-500 hover:bg-brand-500/20'
              : 'hover:bg-muted'
            }
          `}
          onClick={toggleBookmark}
          aria-label={isBookmarked ? 'Remove bookmark' : 'Add bookmark'}
        >
          <AnimatePresence mode="wait">
            {isBookmarked ? (
              <motion.div
                key="checked"
                initial={{ scale: 0.5, opacity: 0 }}
                animate={{ scale: 1, opacity: 1 }}
                exit={{ scale: 0.5, opacity: 0 }}
              >
                <Check className={iconSizes[size]} />
              </motion.div>
            ) : (
              <motion.div
                key="bookmark"
                initial={{ scale: 0.5, opacity: 0 }}
                animate={{ scale: 1, opacity: 1 }}
                exit={{ scale: 0.5, opacity: 0 }}
              >
                <Bookmark className={`${iconSizes[size]} ${isBookmarked ? 'fill-current' : ''}`} />
              </motion.div>
            )}
          </AnimatePresence>
        </Button>
      </TooltipTrigger>
      <TooltipContent>
        {isBookmarked ? 'Remove bookmark' : 'Save for later'}
      </TooltipContent>
    </Tooltip>
  );
}

// Hook for managing bookmarked posts
export function useBookmarks() {
  const [bookmarks, setBookmarks] = useState<BookmarkData[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      setBookmarks(stored ? JSON.parse(stored) : []);
    } catch (err) {
      console.error('Failed to load bookmarks:', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const isBookmarked = useCallback((postId: string) => {
    return bookmarks.some(b => b.id === postId);
  }, [bookmarks]);

  const removeBookmark = useCallback((postId: string) => {
    try {
      const updated = bookmarks.filter(b => b.id !== postId);
      localStorage.setItem(STORAGE_KEY, JSON.stringify(updated));
      setBookmarks(updated);
    } catch (err) {
      console.error('Failed to remove bookmark:', err);
    }
  }, [bookmarks]);

  const clearAll = useCallback(() => {
    localStorage.removeItem(STORAGE_KEY);
    setBookmarks([]);
  }, []);

  return {
    bookmarks,
    isLoading,
    isBookmarked,
    removeBookmark,
    clearAll,
    count: bookmarks.length,
  };
}
