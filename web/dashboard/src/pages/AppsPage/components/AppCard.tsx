import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { cn } from '@/lib/utils';
import type { App } from '@/types';
import {
  Building2,
  ChevronRight,
  Clock,
  ExternalLink,
  Globe,
  MoreVertical,
  Server,
  Settings,
  Trash2,
} from 'lucide-react';
import { Link } from 'react-router-dom';

// App icon color palette - deterministic based on app name
const APP_ICON_COLORS = [
  {
    bg: 'bg-violet-500/15 dark:bg-violet-500/20',
    icon: 'text-violet-500',
    border: 'border-violet-500/30',
  },
  { bg: 'bg-blue-500/15 dark:bg-blue-500/20', icon: 'text-blue-500', border: 'border-blue-500/30' },
  {
    bg: 'bg-emerald-500/15 dark:bg-emerald-500/20',
    icon: 'text-emerald-500',
    border: 'border-emerald-500/30',
  },
  {
    bg: 'bg-amber-500/15 dark:bg-amber-500/20',
    icon: 'text-amber-500',
    border: 'border-amber-500/30',
  },
  { bg: 'bg-rose-500/15 dark:bg-rose-500/20', icon: 'text-rose-500', border: 'border-rose-500/30' },
  { bg: 'bg-cyan-500/15 dark:bg-cyan-500/20', icon: 'text-cyan-500', border: 'border-cyan-500/30' },
  {
    bg: 'bg-indigo-500/15 dark:bg-indigo-500/20',
    icon: 'text-indigo-500',
    border: 'border-indigo-500/30',
  },
  { bg: 'bg-pink-500/15 dark:bg-pink-500/20', icon: 'text-pink-500', border: 'border-pink-500/30' },
];

function getAppColor(name: string) {
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash);
  }
  return APP_ICON_COLORS[Math.abs(hash) % APP_ICON_COLORS.length];
}

function formatRelativeTime(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSecs = Math.floor(diffMs / 1000);
  const diffMins = Math.floor(diffSecs / 60);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffDays > 30) {
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  }
  if (diffDays > 0) return `${diffDays}d ago`;
  if (diffHours > 0) return `${diffHours}h ago`;
  if (diffMins > 0) return `${diffMins}m ago`;
  return 'just now';
}

interface AppCardProps {
  app: App;
  onDelete?: (app: App) => void;
  index?: number;
}

export function AppCard({ app, onDelete, index = 0 }: AppCardProps) {
  const color = getAppColor(app.name);
  const animationDelay = `${index * 0.05}s`;
  const deployUrl = app.deployUrl || app.deploy_url || '';

  return (
    <div
      className="group relative rounded-xl border border-border/50 bg-card hover:border-brand-500/50 hover:shadow-lg hover:shadow-brand-500/5 transition-all duration-200 overflow-hidden"
      style={{ animationDelay, opacity: 1 }}
    >
      {/* Subtle gradient overlay on hover */}
      <div className="absolute inset-0 bg-gradient-to-br from-brand-500/0 to-brand-500/0 group-hover:from-brand-500/3 group-hover:to-transparent transition-all duration-300 pointer-events-none" />

      <div className="p-5">
        {/* Header row */}
        <div className="flex items-start justify-between gap-3 mb-4">
          <Link
            to={`/apps/${encodeURIComponent(app.slug)}`}
            className="flex items-center gap-3 flex-1 min-w-0"
            aria-label={`View ${app.name} details`}
          >
            {/* App Icon */}
            <div
              className={cn(
                'w-11 h-11 rounded-xl flex items-center justify-center flex-shrink-0 border',
                color.bg,
                color.border
              )}
            >
              <Building2 className={cn('w-5 h-5', color.icon)} />
            </div>

            {/* App Name & Slug */}
            <div className="flex-1 min-w-0">
              <h3 className="font-semibold text-foreground truncate group-hover:text-brand-500 transition-colors">
                {app.name}
              </h3>
              <p className="text-xs text-muted-foreground font-mono truncate mt-0.5">{app.slug}</p>
            </div>
          </Link>

          {/* Actions dropdown */}
          <DropdownMenu>
            <DropdownMenuTrigger
              className="opacity-0 group-hover:opacity-100 focus:opacity-100 transition-opacity p-1.5 rounded-lg hover:bg-muted text-muted-foreground hover:text-foreground"
              aria-label="App actions"
              onClick={(e) => e.stopPropagation()}
            >
              <MoreVertical className="w-4 h-4" />
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-48">
              <DropdownMenuItem asChild>
                <Link to={`/apps/${encodeURIComponent(app.slug)}`} className="flex items-center gap-2">
                  <ExternalLink className="w-4 h-4" />
                  View Details
                </Link>
              </DropdownMenuItem>
              {deployUrl && (
                <DropdownMenuItem asChild>
                  <a href={deployUrl} target="_blank" rel="noopener noreferrer" className="flex items-center gap-2">
                    <Globe className="w-4 h-4" />
                    Visit Live
                  </a>
                </DropdownMenuItem>
              )}
              <DropdownMenuItem asChild>
                <Link to={`/apps/${encodeURIComponent(app.slug)}`} className="flex items-center gap-2">
                  <Settings className="w-4 h-4" />
                  Settings
                </Link>
              </DropdownMenuItem>
              <DropdownMenuItem asChild>
                <Link to={`/apps/${encodeURIComponent(app.slug)}`} className="flex items-center gap-2">
                  <Server className="w-4 h-4" />
                  Backends
                </Link>
              </DropdownMenuItem>
              {onDelete && (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    className="text-destructive focus:text-destructive flex items-center gap-2"
                    onClick={(e) => {
                      e.stopPropagation();
                      onDelete(app);
                    }}
                  >
                    <Trash2 className="w-4 h-4" />
                    Delete App
                  </DropdownMenuItem>
                </>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        {/* Stats row */}
        <div className="flex items-center gap-3 text-xs text-muted-foreground">
          <div className="flex items-center gap-1.5">
            <div className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
            <span>Active</span>
          </div>
          <span className="text-border">·</span>
          <div className="flex items-center gap-1">
            <Clock className="w-3 h-3" />
            <span>{formatRelativeTime(app.createdAt)}</span>
          </div>
        </div>

        {/* Deploy URL */}
        {deployUrl && (
          <div className="mt-3 flex items-center gap-2">
            <Globe className="w-3.5 h-3.5 text-muted-foreground flex-shrink-0" />
            <span className="text-xs font-mono text-muted-foreground truncate flex-1 min-w-0">
              {deployUrl.replace(/^https?:\/\//, '')}
            </span>
            <a
              href={deployUrl}
              target="_blank"
              rel="noopener noreferrer"
              onClick={(e) => e.stopPropagation()}
              className="inline-flex items-center gap-1 px-2 py-1 text-xs font-medium rounded-md bg-brand-500/10 text-brand-500 hover:bg-brand-500/20 transition-colors flex-shrink-0"
              aria-label={`Open ${app.name} live`}
            >
              <ExternalLink className="w-3 h-3" />
              Open
            </a>
          </div>
        )}
      </div>

      {/* Footer link */}
      <Link
        to={`/apps/${encodeURIComponent(app.slug)}`}
        className="flex items-center justify-between px-5 py-3 border-t border-border/50 bg-muted/30 hover:bg-muted/60 transition-colors text-xs text-muted-foreground hover:text-foreground group/footer"
        aria-label={`Open ${app.name}`}
      >
        <span>View app details</span>
        <ChevronRight className="w-3.5 h-3.5 group-hover/footer:translate-x-0.5 transition-transform" />
      </Link>
    </div>
  );
}

// Skeleton card for loading state
export function AppCardSkeleton() {
  return (
    <div className="rounded-xl border border-border/50 bg-card overflow-hidden animate-pulse">
      <div className="p-5">
        <div className="flex items-start gap-3 mb-4">
          <div className="w-11 h-11 rounded-xl bg-muted flex-shrink-0" />
          <div className="flex-1 space-y-2">
            <div className="h-4 bg-muted rounded w-3/4" />
            <div className="h-3 bg-muted rounded w-1/2" />
          </div>
        </div>
        <div className="flex items-center gap-3">
          <div className="h-3 bg-muted rounded w-16" />
          <div className="h-3 bg-muted rounded w-20" />
        </div>
      </div>
      <div className="px-5 py-3 border-t border-border/50 bg-muted/30">
        <div className="h-3 bg-muted rounded w-24" />
      </div>
    </div>
  );
}
