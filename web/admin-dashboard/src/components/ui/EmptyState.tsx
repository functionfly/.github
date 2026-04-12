import React from 'react';
import {
  Inbox,
  Search,
  FileX,
  AlertCircle,
  FolderOpen,
  CalendarX,
  FileQuestion,
  type LucideIcon,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from './Button';

interface EmptyStateProps {
  icon?: LucideIcon;
  title: string;
  description?: string;
  action?: {
    label: string;
    onClick: () => void;
  };
  secondaryAction?: {
    label: string;
    onClick: () => void;
  };
  className?: string;
  size?: 'sm' | 'md' | 'lg';
  variant?: 'default' | 'compact' | 'card';
}

export function EmptyState({
  icon: Icon = Inbox,
  title,
  description,
  action,
  secondaryAction,
  className,
  size = 'md',
  variant = 'default',
}: EmptyStateProps) {
  const sizeClasses = {
    sm: {
      wrapper: 'py-6',
      icon: 'w-10 h-10',
      iconWrapper: 'p-2',
      title: 'text-sm',
      description: 'text-xs',
    },
    md: {
      wrapper: 'py-12',
      icon: 'w-16 h-16',
      iconWrapper: 'p-4',
      title: 'text-lg',
      description: 'text-sm',
    },
    lg: {
      wrapper: 'py-16',
      icon: 'w-20 h-20',
      iconWrapper: 'p-6',
      title: 'text-xl',
      description: 'text-base',
    },
  };

  const variantClasses = {
    default: '',
    compact: 'bg-gray-50 dark:bg-gray-800/50 rounded-lg',
    card: 'bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm',
  };

  const s = sizeClasses[size];

  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center text-center',
        s.wrapper,
        variantClasses[variant],
        className
      )}
    >
      <div
        className={cn(
          'rounded-full bg-gray-100 dark:bg-gray-700/50 text-gray-400 dark:text-gray-500 mb-4',
          s.iconWrapper
        )}
      >
        <Icon className={s.icon} />
      </div>
      <h3 className={cn('font-semibold text-gray-900 dark:text-gray-100 mb-2', s.title)}>
        {title}
      </h3>
      {description && (
        <p className={cn('text-gray-500 dark:text-gray-400 max-w-md mb-6', s.description)}>
          {description}
        </p>
      )}
      {(action || secondaryAction) && (
        <div className="flex items-center gap-3">
          {action && (
            <Button onClick={action.onClick}>
              {action.label}
            </Button>
          )}
          {secondaryAction && (
            <Button variant="outline" onClick={secondaryAction.onClick}>
              {secondaryAction.label}
            </Button>
          )}
        </div>
      )}
    </div>
  );
}

// Preset empty states for common scenarios
export function EmptySearch({
  query,
  className,
}: {
  query?: string;
  className?: string;
}) {
  return (
    <EmptyState
      icon={Search}
      title={query ? `No results for "${query}"` : 'No results found'}
      description="Try adjusting your search or filters to find what you're looking for."
      variant="compact"
      className={className}
    />
  );
}

export function EmptyData({
  resourceName = 'items',
  className,
}: {
  resourceName?: string;
  className?: string;
}) {
  return (
    <EmptyState
      icon={FileX}
      title={`No ${resourceName} yet`}
      description={`Get started by creating your first ${resourceName.slice(0, -1) || 'item'}.`}
      variant="compact"
      className={className}
    />
  );
}

export function EmptyFolder({ className }: { className?: string }) {
  return (
    <EmptyState
      icon={FolderOpen}
      title="Folder is empty"
      description="This folder doesn't contain any items yet."
      variant="compact"
      className={className}
    />
  );
}

export function EmptyCalendar({ className }: { className?: string }) {
  return (
    <EmptyState
      icon={CalendarX}
      title="No events scheduled"
      description="There are no events scheduled for this period."
      variant="compact"
      className={className}
    />
  );
}

export function ErrorState({
  title = 'Something went wrong',
  description = 'An error occurred while loading the data. Please try again.',
  onRetry,
  className,
}: {
  title?: string;
  description?: string;
  onRetry?: () => void;
  className?: string;
}) {
  return (
    <EmptyState
      icon={AlertCircle}
      title={title}
      description={description}
      action={onRetry ? { label: 'Try Again', onClick: onRetry } : undefined}
      className={className}
    />
  );
}

export function NotFoundState({
  title = 'Page not found',
  description = "The page you're looking for doesn't exist or has been moved.",
  onGoBack,
  className,
}: {
  title?: string;
  description?: string;
  onGoBack?: () => void;
  className?: string;
}) {
  return (
    <EmptyState
      icon={FileQuestion}
      title={title}
      description={description}
      action={onGoBack ? { label: 'Go Back', onClick: onGoBack } : undefined}
      className={className}
    />
  );
}
