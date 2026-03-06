/**
 * ThreadCard - Thread preview card for lists
 */

import { Link } from 'react-router-dom';
import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import {
  MessageCircle,
  Eye,
  Clock,
  CheckCircle2,
  Check,
  Lock,
} from 'lucide-react';
import { ReputationBadgeInline } from '../reputation/ReputationBadge';
import type { Thread } from '../types';

interface ThreadCardProps {
  thread: Thread;
  className?: string;
}

const statusConfig = {
  open: { color: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20', label: 'Open' },
  in_progress: { color: 'bg-blue-500/10 text-blue-400 border-blue-500/20', label: 'In Progress' },
  resolved: { color: 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20', label: 'Resolved' },
  closed: { color: 'bg-slate-500/10 text-slate-400 border-slate-500/20', label: 'Closed' },
  archived: { color: 'bg-slate-500/10 text-slate-400 border-slate-500/20', label: 'Archived' },
};

const typeConfig = {
  problem: { color: 'bg-amber-500/10 text-amber-400', label: 'Problem' },
  discussion: { color: 'bg-violet-500/10 text-violet-400', label: 'Discussion' },
  challenge: { color: 'bg-pink-500/10 text-pink-400', label: 'Challenge' },
};

function formatTimeAgo(date: string): string {
  const now = new Date();
  const past = new Date(date);
  const diffMs = now.getTime() - past.getTime();
  const diffSecs = Math.floor(diffMs / 1000);
  const diffMins = Math.floor(diffSecs / 60);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffSecs < 60) return 'just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;
  return past.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
}

export function ThreadCard({ thread, className }: ThreadCardProps) {
  const status = statusConfig[thread.status];
  const type = typeConfig[thread.type];

  return (
    <Link
      to={`/flywheel/threads/${thread.id}`}
      className={cn(
        'group block rounded-xl border border-slate-800 bg-slate-900 p-5 transition-all duration-200 hover:-translate-y-0.5 hover:border-slate-700 hover:shadow-lg',
        className
      )}
    >
      {/* Header */}
      <div className="flex items-start gap-3">
        <Avatar className="h-10 w-10 border border-slate-800">
          <AvatarImage src={thread.author.avatarUrl} alt={thread.author.username} />
          <AvatarFallback className="bg-slate-800 text-slate-400">
            {thread.author.username.slice(0, 2).toUpperCase()}
          </AvatarFallback>
        </Avatar>

        <div className="min-w-0 flex-1">
          <h3 className="line-clamp-1 text-lg font-semibold text-white group-hover:text-indigo-400">
            {thread.title}
          </h3>

          <div className="mt-1 flex flex-wrap items-center gap-2">
            <span className="text-sm text-slate-400">
              @{thread.author.username}
            </span>

            {thread.author.reputation && (
              <ReputationBadgeInline
                score={thread.author.reputation.overallScore}
                type="overall"
                tier={thread.author.reputation.tier.level}
              />
            )}

            <span className="text-slate-600">•</span>
            <span className="text-sm text-slate-400">
              {formatTimeAgo(thread.createdAt)}
            </span>
          </div>
        </div>
      </div>

      {/* Metadata Row */}
      <div className="mt-3 flex flex-wrap items-center gap-2">
        <Badge variant="outline" className={cn('text-xs', status.color)}>
          {status.label}
        </Badge>

        <Badge variant="outline" className={cn('text-xs', type.color)}>
          {type.label}
        </Badge>

        {thread.category && (
          <Badge
            variant="outline"
            className="border-slate-700 bg-slate-800/50 text-xs text-slate-400"
          >
            {thread.category.name}
          </Badge>
        )}
      </div>

      {/* Preview */}
      {thread.problemData?.description && (
        <p className="mt-3 line-clamp-2 text-sm text-slate-400">
          {thread.problemData.description}
        </p>
      )}

      {/* Tags */}
      {thread.tags.length > 0 && (
        <div className="mt-3 flex flex-wrap gap-1.5">
          {thread.tags.slice(0, 4).map((tag) => (
            <span
              key={tag}
              className="rounded-md bg-slate-800 px-2 py-0.5 text-xs text-slate-400"
            >
              {tag}
            </span>
          ))}
          {thread.tags.length > 4 && (
            <span className="text-xs text-slate-400">
              +{thread.tags.length - 4}
            </span>
          )}
        </div>
      )}

      {/* Footer Stats */}
      <div className="mt-4 flex items-center gap-4 border-t border-slate-800 pt-3">
        <div className="flex items-center gap-1.5 text-sm text-slate-400">
          <MessageCircle className="h-4 w-4" />
          <span>{thread.replyCount}</span>
        </div>

        <div className="flex items-center gap-1.5 text-sm text-slate-400">
          <Eye className="h-4 w-4" />
          <span>{thread.viewCount.toLocaleString()}</span>
        </div>

        <div className="flex items-center gap-1.5 text-sm text-slate-400">
          <Clock className="h-4 w-4" />
          <span>{formatTimeAgo(thread.updatedAt)}</span>
        </div>

        {/* Solution Indicators */}
        <div className="ml-auto flex items-center gap-2">
          {thread.hasVerifiedSolution && (
            <div className="flex items-center gap-1 text-xs font-medium text-indigo-400">
              <CheckCircle2 className="h-4 w-4" />
              <span className="hidden sm:inline">Verified</span>
            </div>
          )}

          {thread.hasAcceptedSolution && (
            <div className="flex items-center gap-1 text-xs font-medium text-emerald-400">
              <Check className="h-4 w-4" />
              <span className="hidden sm:inline">Solved</span>
            </div>
          )}
        </div>
      </div>
    </Link>
  );
}

/**
 * ThreadCardSkeleton - Loading state for thread cards
 */
export function ThreadCardSkeleton() {
  return (
    <div className="animate-pulse rounded-xl border border-slate-800 bg-slate-900 p-5">
      <div className="flex items-start gap-3">
        <div className="h-10 w-10 rounded-full bg-slate-800" />
        <div className="flex-1 space-y-2">
          <div className="h-5 w-3/4 rounded bg-slate-800" />
          <div className="h-4 w-1/2 rounded bg-slate-800" />
        </div>
      </div>
      <div className="mt-3 flex gap-2">
        <div className="h-5 w-16 rounded bg-slate-800" />
        <div className="h-5 w-20 rounded bg-slate-800" />
      </div>
      <div className="mt-3 space-y-2">
        <div className="h-4 w-full rounded bg-slate-800" />
        <div className="h-4 w-5/6 rounded bg-slate-800" />
      </div>
      <div className="mt-4 flex gap-4 border-t border-slate-800 pt-3">
        <div className="h-4 w-16 rounded bg-slate-800" />
        <div className="h-4 w-16 rounded bg-slate-800" />
        <div className="h-4 w-16 rounded bg-slate-800" />
      </div>
    </div>
  );
}
