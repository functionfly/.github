/**
 * ThreadHeader - Thread detail header
 */

import { Link } from 'react-router-dom';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Bell,
  Share2,
  MoreHorizontal,
  Edit,
  Trash2,
  Flag,
  Bookmark,
  CheckCircle2,
} from 'lucide-react';
import { ReputationBadgeInline } from '../reputation/ReputationBadge';
import type { Thread, ThreadStatus } from '../types';

interface ThreadHeaderProps {
  thread: Thread;
  isSubscribed?: boolean;
  onSubscribe?: () => void;
  onShare?: () => void;
  onEdit?: () => void;
  onDelete?: () => void;
  className?: string;
}

const statusConfig: Record<ThreadStatus, { color: string; label: string }> = {
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
  if (diffMins < 60) return `${diffMins} minutes ago`;
  if (diffHours < 24) return `${diffHours} hours ago`;
  if (diffDays < 7) return `${diffDays} days ago`;
  return past.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}

export function ThreadHeader({
  thread,
  isSubscribed = false,
  onSubscribe,
  onShare,
  onEdit,
  onDelete,
  className,
}: ThreadHeaderProps) {
  const status = statusConfig[thread.status];
  const type = typeConfig[thread.type];

  return (
    <div className={cn('space-y-4', className)}>
      {/* Breadcrumb */}
      <nav className="flex items-center gap-2 text-sm text-slate-400">
        <Link to="/flywheel" className="hover:text-slate-300">
          Flywheel
        </Link>
        <span>/</span>
        {thread.category && (
          <>
            <Link
              to={`/flywheel/threads?category=${thread.category.slug}`}
              className="hover:text-slate-300"
            >
              {thread.category.name}
            </Link>
            <span>/</span>
          </>
        )}
        <span className="truncate text-slate-400">{thread.title}</span>
      </nav>

      {/* Title Section */}
      <div className="flex items-start justify-between gap-4">
        <h1 className="text-2xl font-bold text-white sm:text-3xl">{thread.title}</h1>

        {/* Action Buttons */}
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={onSubscribe}
            className={cn(
              'border-slate-700',
              isSubscribed && 'bg-indigo-500/10 text-indigo-400 border-indigo-500/30'
            )}
          >
            <Bell className="mr-2 h-4 w-4" />
            {isSubscribed ? 'Subscribed' : 'Subscribe'}
          </Button>

          <Button
            variant="outline"
            size="sm"
            onClick={onShare}
            className="border-slate-700"
          >
            <Share2 className="mr-2 h-4 w-4" />
            Share
          </Button>

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="icon" className="border-slate-700" aria-label="More options">
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="bg-slate-900 border-slate-800">
              <DropdownMenuItem
                onClick={onEdit}
                className="text-slate-300 focus:bg-slate-800 focus:text-slate-100"
              >
                <Edit className="mr-2 h-4 w-4" />
                Edit Thread
              </DropdownMenuItem>
              <DropdownMenuItem className="text-slate-300 focus:bg-slate-800 focus:text-slate-100">
                <Bookmark className="mr-2 h-4 w-4" />
                Bookmark
              </DropdownMenuItem>
              <DropdownMenuItem className="text-slate-300 focus:bg-slate-800 focus:text-slate-100">
                <Flag className="mr-2 h-4 w-4" />
                Report
              </DropdownMenuItem>
              <DropdownMenuSeparator className="bg-slate-800" />
              <DropdownMenuItem
                onClick={onDelete}
                className="text-red-400 focus:bg-slate-800 focus:text-red-300"
              >
                <Trash2 className="mr-2 h-4 w-4" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      {/* Author Info */}
      <div className="flex flex-wrap items-center gap-4">
        <div className="flex items-center gap-3">
          <Avatar className="h-12 w-12 border border-slate-800">
            <AvatarImage src={thread.author.avatarUrl} alt={thread.author.username} />
            <AvatarFallback className="bg-slate-800 text-slate-400">
              {thread.author.username.slice(0, 2).toUpperCase()}
            </AvatarFallback>
          </Avatar>

          <div>
            <Link
              to={`/flywheel/users/${thread.author.id}`}
              className="font-medium text-white hover:text-indigo-400"
            >
              @{thread.author.username}
            </Link>
            <div className="flex items-center gap-2 text-sm text-slate-400">
              <span>Posted {formatTimeAgo(thread.createdAt)}</span>
              {thread.createdAt !== thread.updatedAt && (
                <span>(edited)</span>
              )}
            </div>
          </div>
        </div>

        {/* Reputation */}
        {thread.author.reputation && (
          <div className="flex items-center gap-2">
            <ReputationBadgeInline
              score={thread.author.reputation.overallScore}
              type="overall"
              tier={thread.author.reputation.tier.level}
              size="sm"
            />
          </div>
        )}
      </div>

      {/* Metadata */}
      <div className="flex flex-wrap items-center gap-2">
        <Badge variant="outline" className={cn('text-sm', status.color)}>
          {status.label}
        </Badge>

        <Badge variant="outline" className={cn('text-sm', type.color)}>
          {type.label}
        </Badge>

        {thread.category && (
          <Badge
            variant="outline"
            className="border-slate-700 bg-slate-800/50 text-sm text-slate-400"
          >
            {thread.category.name}
          </Badge>
        )}

        {thread.hasVerifiedSolution && (
          <Badge
            variant="outline"
            className="border-indigo-500/30 bg-indigo-500/10 text-indigo-400"
          >
            <CheckCircle2 className="mr-1 h-3 w-3" />
            Verified Solution
          </Badge>
        )}
      </div>

      {/* Tags */}
      {thread.tags.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {thread.tags.map((tag) => (
            <Link
              key={tag}
              to={`/flywheel/threads?tags=${tag}`}
              className="rounded-md bg-slate-800 px-2.5 py-1 text-sm text-slate-400 transition-colors hover:bg-slate-700 hover:text-slate-300"
            >
              #{tag}
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
