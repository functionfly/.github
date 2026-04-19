import { cn } from '@/lib/utils';
import * as React from 'react';

const ScrollArea = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, children, ...props }, ref) => (
    <div ref={ref} className={cn('relative overflow-hidden', className)} {...props}>
      <div className="h-full w-full rounded-[inherit] overflow-auto scrollbar-thin scrollbar-thumb-border-default scrollbar-track-transparent hover:scrollbar-thumb-border-subtle">
        {children}
      </div>
    </div>
  )
);
ScrollArea.displayName = 'ScrollArea';

const ScrollBar = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement> & { orientation?: 'vertical' | 'horizontal' }
>(({ className, orientation = 'vertical', ...props }, ref) => (
  <div
    ref={ref}
    className={cn(
      'flex touch-none select-none transition-colors',
      orientation === 'vertical'
        ? 'h-full w-2 border-l border-l-transparent'
        : 'h-2 w-full border-t border-t-transparent',
      className
    )}
    {...props}
  />
));
ScrollBar.displayName = 'ScrollBar';

export { ScrollArea, ScrollBar };
