'use client';

import { useState, useCallback } from 'react';
import { X } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface FilterBarProps {
  availableTags: string[];
  selectedTags: string[];
  onTagSelect: (tag: string) => void;
  onClearAll: () => void;
}

export function FilterBar({
  availableTags,
  selectedTags,
  onTagSelect,
  onClearAll,
}: FilterBarProps) {
  const [activeFilter, setActiveFilter] = useState<string | null>(null);

  const handleTagClick = useCallback((tag: string) => {
    setActiveFilter(tag);
    onTagSelect(tag);
    // Reset after animation
    setTimeout(() => setActiveFilter(null), 300);
  }, [onTagSelect]);

  if (availableTags.length === 0) {
    return null;
  }

  return (
    <div className="w-full">
      {/* Active Filters Display */}
      {selectedTags.length > 0 && (
        <div className="flex flex-wrap items-center gap-2 mb-4">
          <span className="text-sm text-muted-foreground">Active filters:</span>
          {selectedTags.map((tag) => (
            <Button
              key={tag}
              variant="secondary"
              size="sm"
              onClick={() => onTagSelect(tag)}
              className="rounded-full gap-1.5 h-7 px-3 text-sm bg-brand-500/10 text-brand-600 dark:text-brand-400 border border-brand-500/20 hover:bg-brand-500/20"
            >
              {tag}
              <X className="h-3 w-3" />
            </Button>
          ))}
          <Button
            variant="ghost"
            size="sm"
            onClick={onClearAll}
            className="rounded-full h-7 px-3 text-sm text-muted-foreground hover:text-foreground"
          >
            Clear all
          </Button>
        </div>
      )}

      {/* Tag Pills */}
      <div className="flex flex-wrap gap-2" role="group" aria-label="Filter by tag">
        {availableTags.map((tag) => {
          const isSelected = selectedTags.includes(tag);
          const isAnimating = activeFilter === tag;

          return (
            <Button
              key={tag}
              variant={isSelected ? 'default' : 'outline'}
              size="sm"
              onClick={() => handleTagClick(tag)}
              className={`
                rounded-full h-8 px-4 text-sm font-medium transition-all duration-200
                ${isSelected 
                  ? 'bg-brand-500 hover:bg-brand-600 text-white shadow-md shadow-brand-500/25' 
                  : 'border-border/60 bg-muted/30 hover:bg-muted/60 hover:border-border'
                }
                ${isAnimating ? 'scale-95' : 'scale-100'}
              `}
              aria-pressed={isSelected}
            >
              {tag}
            </Button>
          );
        })}
      </div>
    </div>
  );
}
