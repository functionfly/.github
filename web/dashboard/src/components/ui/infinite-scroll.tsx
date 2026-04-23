import * as React from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import { cn } from '@/lib/utils';

export interface InfiniteScrollProps extends React.HTMLAttributes<HTMLDivElement> {
  /** All items to render (flat list). */
  items: React.ReactNode[];
  /** Estimated row height in px. */
  estimateSize?: number;
  /** Number of items to render above/below viewport. */
  overscan?: number;
  /** Called when user scrolls near the top (for loading older items). */
  onLoadMore?: () => void;
  /** Threshold in px from top to trigger onLoadMore. */
  loadMoreThreshold?: number;
  /** Auto-scroll to bottom on new items. */
  autoScrollToBottom?: boolean;
}

export function InfiniteScroll({
  items,
  estimateSize = 60,
  overscan = 10,
  onLoadMore,
  loadMoreThreshold = 200,
  autoScrollToBottom = false,
  className,
  ...props
}: InfiniteScrollProps) {
  const parentRef = React.useRef<HTMLDivElement>(null);
  const prevCountRef = React.useRef(items.length);

  const virtualizer = useVirtualizer({
    count: items.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => estimateSize,
    overscan,
  });

  // Auto-scroll to bottom when new items are added
  React.useEffect(() => {
    if (autoScrollToBottom && items.length > prevCountRef.current) {
      const el = parentRef.current;
      if (el) {
        // Use requestAnimationFrame to ensure DOM has updated
        requestAnimationFrame(() => {
          el.scrollTop = el.scrollHeight;
        });
      }
    }
    prevCountRef.current = items.length;
  }, [items.length, autoScrollToBottom]);

  // Infinite scroll: load more when near top
  const handleScroll = React.useCallback(() => {
    const el = parentRef.current;
    if (!el || !onLoadMore) return;
    if (el.scrollTop < loadMoreThreshold) {
      onLoadMore();
    }
  }, [onLoadMore, loadMoreThreshold]);

  return (
    <div
      ref={parentRef}
      onScroll={handleScroll}
      className={cn('overflow-auto', className)}
      {...props}
    >
      <div
        style={{
          height: `${virtualizer.getTotalSize()}px`,
          width: '100%',
          position: 'relative',
        }}
      >
        {virtualizer.getVirtualItems().map((virtualRow) => (
          <div
            key={virtualRow.key}
            style={{
              position: 'absolute',
              top: 0,
              left: 0,
              width: '100%',
              height: `${virtualRow.size}px`,
              transform: `translateY(${virtualRow.start}px)`,
            }}
          >
            {items[virtualRow.index]}
          </div>
        ))}
      </div>
    </div>
  );
}
