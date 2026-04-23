import { Skeleton } from './skeleton';
import { cn } from '@/lib/utils';

export interface SkeletonMessageProps {
  isOwn?: boolean;
}

export function SkeletonMessage({ isOwn = false }: SkeletonMessageProps) {
  return (
    <div
      className={cn(
        'flex flex-col gap-2 max-w-[85%]',
        isOwn ? 'self-end items-end' : 'self-start items-start',
      )}
    >
      {!isOwn && <Skeleton className="h-3 w-20 rounded" />}
      <div
        className={cn(
          'rounded-lg px-3 py-2 space-y-2',
          isOwn ? 'ml-auto' : '',
        )}
      >
        <Skeleton className={cn('h-4 rounded', isOwn ? 'w-48' : 'w-64')} />
        <Skeleton className="h-4 w-32 rounded" />
      </div>
      <Skeleton className="h-2.5 w-16 rounded" />
    </div>
  );
}

export interface SkeletonMessagesProps {
  count?: number;
}

export function SkeletonMessages({ count = 6 }: SkeletonMessagesProps) {
  return (
    <div className="space-y-4 p-4">
      {Array.from({ length: count }).map((_, i) => (
        <SkeletonMessage key={i} isOwn={i % 3 === 0} />
      ))}
    </div>
  );
}

export interface SkeletonConversationItemProps {}

export function SkeletonConversationItem(_props: SkeletonConversationItemProps) {
  return (
    <div className="flex flex-col gap-1 rounded-lg px-3 py-2">
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0 flex-1 flex flex-col gap-1.5">
          <Skeleton className="h-3 w-16 rounded" />
          <Skeleton className="h-4 w-32 rounded" />
          <Skeleton className="h-3 w-24 rounded" />
        </div>
      </div>
    </div>
  );
}

export interface SkeletonConversationListProps {
  count?: number;
}

export function SkeletonConversationList({ count = 6 }: SkeletonConversationListProps) {
  return (
    <div className="p-1 space-y-1">
      {Array.from({ length: count }).map((_, i) => (
        <SkeletonConversationItem key={i} />
      ))}
    </div>
  );
}
