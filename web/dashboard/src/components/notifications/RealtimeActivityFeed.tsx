/**
 * RealtimeActivityFeed Component
 *
 * Live stream of activity with WebSocket-powered real-time updates.
 * Features timeline-style display, activity grouping, filtering, and animations.
 */

'use client';

import React, { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { formatDistanceToNow } from 'date-fns';
import {
  Play,
  TrendingUp,
  CheckCircle,
  Shield,
  Rocket,
  Coins,
  WifiOff,
  Filter,
  Trash2,
  Activity,
  X,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { ActivityFeedItem } from '@/types/notifications';
import { useRealtimeSubscription } from '@/hooks/useRealtimeSubscription';
import { RealtimeEvent } from '@/hooks/types';

// ============================================================================
// Types & Interfaces
// ============================================================================

export type ActivityType =
  | 'execution'
  | 'reputation_gain'
  | 'issue_resolved'
  | 'trust_update'
  | 'deployment'
  | 'bounty';

interface RealtimeActivityFeedProps {
  className?: string;
  maxItems?: number;
  filter?: ActivityType[];
  showHeader?: boolean;
  compact?: boolean;
  onActivityClick?: (activity: ActivityFeedItem) => void;
}

interface ActivityEvent extends RealtimeEvent {
  activity: ActivityFeedItem;
}

interface ActivityConfig {
  icon: React.ReactNode;
  color: string;
  bgColor: string;
  label: string;
}

interface GroupedActivities {
  key: string;
  count: number;
  activities: ActivityFeedItem[];
  type: ActivityType;
  timestamp: string;
}

// ============================================================================
// Constants
// ============================================================================

const ACTIVITY_CONFIG: Record<ActivityType, ActivityConfig> = {
  execution: {
    icon: <Play className="h-4 w-4" />,
    color: 'text-blue-500',
    bgColor: 'bg-blue-500/10',
    label: 'Execution',
  },
  reputation_gain: {
    icon: <TrendingUp className="h-4 w-4" />,
    color: 'text-green-500',
    bgColor: 'bg-green-500/10',
    label: 'Reputation',
  },
  issue_resolved: {
    icon: <CheckCircle className="h-4 w-4" />,
    color: 'text-purple-500',
    bgColor: 'bg-purple-500/10',
    label: 'Issue Resolved',
  },
  trust_update: {
    icon: <Shield className="h-4 w-4" />,
    color: 'text-indigo-500',
    bgColor: 'bg-indigo-500/10',
    label: 'Trust Update',
  },
  deployment: {
    icon: <Rocket className="h-4 w-4" />,
    color: 'text-cyan-500',
    bgColor: 'bg-cyan-500/10',
    label: 'Deployment',
  },
  bounty: {
    icon: <Coins className="h-4 w-4" />,
    color: 'text-yellow-500',
    bgColor: 'bg-yellow-500/10',
    label: 'Bounty',
  },
};

const ACTIVITY_TYPE_OPTIONS: { value: ActivityType | 'all'; label: string }[] = [
  { value: 'all', label: 'All Activities' },
  { value: 'execution', label: 'Executions' },
  { value: 'reputation_gain', label: 'Reputation' },
  { value: 'issue_resolved', label: 'Issues Resolved' },
  { value: 'trust_update', label: 'Trust Updates' },
  { value: 'deployment', label: 'Deployments' },
  { value: 'bounty', label: 'Bounties' },
];

const ACTIVITY_TYPE_MAP: Record<string, ActivityType> = {
  executed: 'execution',
  execution: 'execution',
  reputation_gained: 'reputation_gain',
  reputation_gain: 'reputation_gain',
  issue_resolved: 'issue_resolved',
  resolved: 'issue_resolved',
  trust_change: 'trust_update',
  trust_update: 'trust_update',
  deployed: 'deployment',
  deployment: 'deployment',
  bounty_claimed: 'bounty',
  bounty: 'bounty',
};

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Determine activity type from action and context
 */
function getActivityType(activity: ActivityFeedItem): ActivityType {
  const action = activity.action.toLowerCase();
  const targetType = activity.target.type.toLowerCase();

  // Check context icon if available
  if (activity.context?.icon) {
    const iconType = ACTIVITY_TYPE_MAP[activity.context.icon.toLowerCase()];
    if (iconType) return iconType;
  }

  // Map based on action and target
  if (targetType === 'deployment' || action === 'deployed') return 'deployment';
  if (targetType === 'issue' && action === 'resolved') return 'issue_resolved';
  if (action === 'earned' || action === 'claimed') return 'reputation_gain';
  if (targetType === 'function' && action === 'created') return 'deployment';
  if (targetType === 'bounty' || action === 'claimed') return 'bounty';

  return 'execution';
}

/**
 * Format action text for display
 */
function formatAction(action: string): string {
  return action
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
}

/**
 * Group similar activities together
 */
function groupActivities(activities: ActivityFeedItem[]): (ActivityFeedItem | GroupedActivities)[] {
  const groups: Map<string, ActivityFeedItem[]> = new Map();

  activities.forEach((activity) => {
    const type = getActivityType(activity);
    const actorId = activity.actor.id;
    const action = activity.action;
    const targetType = activity.target.type;

    // Create a grouping key based on actor, action, and target type
    const key = `${actorId}:${action}:${targetType}:${type}`;

    if (!groups.has(key)) {
      groups.set(key, []);
    }
    groups.get(key)!.push(activity);
  });

  const result: (ActivityFeedItem | GroupedActivities)[] = [];

  groups.forEach((groupActivities, key) => {
    if (groupActivities.length === 1) {
      result.push(groupActivities[0]);
    } else {
      // Group activities that happened within 5 minutes of each other
      const sorted = groupActivities.sort(
        (a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
      );

      const type = getActivityType(sorted[0]);
      result.push({
        key,
        count: sorted.length,
        activities: sorted,
        type,
        timestamp: sorted[0].timestamp,
      });
    }
  });

  // Sort by timestamp (newest first)
  return result.sort((a, b) => {
    const timeA = 'activities' in a ? a.activities[0].timestamp : a.timestamp;
    const timeB = 'activities' in b ? b.activities[0].timestamp : b.timestamp;
    return new Date(timeB).getTime() - new Date(timeA).getTime();
  });
}

// ============================================================================
// Sub-Components
// ============================================================================

/**
 * Live indicator with pulse animation
 */
function LiveIndicator({ isConnected, error }: { isConnected: boolean; error?: Error | string | null }) {
  return (
    <div className="flex items-center gap-2">
      <div className="relative flex h-3 w-3">
        {isConnected ? (
          <>
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75" />
            <span className="relative inline-flex rounded-full h-3 w-3 bg-green-500" />
          </>
        ) : (
          <span className="relative inline-flex rounded-full h-3 w-3 bg-red-500">
            <WifiOff className="h-2 w-2 text-white absolute inset-0 m-auto" />
          </span>
        )}
      </div>
      <span
        className={cn(
          'text-xs font-medium',
          isConnected ? 'text-green-500' : 'text-red-500'
        )}
      >
        {isConnected ? 'Live' : error ? 'Connection Error' : 'Disconnected'}
      </span>
    </div>
  );
}

/**
 * Activity type filter dropdown
 */
function ActivityFilter({
  value,
  onChange,
}: {
  value: ActivityType | 'all';
  onChange: (value: ActivityType | 'all') => void;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" className="h-8 w-8" aria-label="Filter by activity type">
          <Filter className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-48">
        <DropdownMenuLabel>Filter by Type</DropdownMenuLabel>
        <DropdownMenuSeparator />
        {ACTIVITY_TYPE_OPTIONS.map((option) => (
          <DropdownMenuItem
            key={option.value}
            onClick={() => onChange(option.value as ActivityType | 'all')}
            className={cn(
              'flex items-center justify-between',
              value === option.value && 'bg-bg-secondary'
            )}
          >
            {option.label}
            {value === option.value && <CheckCircle className="h-4 w-4" />}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

/**
 * Empty state component
 */
function EmptyState() {
  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      className="flex flex-col items-center justify-center py-12 px-4 text-center"
    >
      <div className="w-16 h-16 rounded-full bg-bg-tertiary border border-border-subtle flex items-center justify-center mb-4">
        <Activity className="h-8 w-8 text-text-muted" />
      </div>
      <h3 className="text-lg font-medium text-text-primary mb-1">No activity yet</h3>
      <p className="text-sm text-text-muted max-w-xs">
        Activity will appear here in real-time as events occur.
      </p>
    </motion.div>
  );
}

/**
 * Individual activity item component
 */
interface ActivityItemProps {
  activity: ActivityFeedItem;
  compact?: boolean;
  onClick?: (activity: ActivityFeedItem) => void;
  isLast?: boolean;
}

function ActivityItem({ activity, compact, onClick, isLast }: ActivityItemProps) {
  const type = getActivityType(activity);
  const config = ACTIVITY_CONFIG[type];
  const timeAgo = formatDistanceToNow(new Date(activity.timestamp), { addSuffix: true });

  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: -20 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, x: -100 }}
      transition={{ duration: 0.3, ease: 'easeOut' }}
      onClick={() => onClick?.(activity)}
      className={cn(
        'relative flex items-start gap-3 p-3 rounded-lg transition-colors cursor-pointer',
        'hover:bg-bg-secondary/50',
        !isLast && 'border-b border-border-subtle'
      )}
    >
      {/* Timeline connector */}
      {!isLast && !compact && (
        <div className="absolute left-[1.6rem] top-10 bottom-0 w-px bg-border-subtle" />
      )}

      {/* Icon/Avatar */}
      <div
        className={cn(
          'relative flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center',
          config.bgColor,
          config.color
        )}
      >
        {activity.actor.avatarUrl ? (
          <Avatar className="w-8 h-8">
            <AvatarImage src={activity.actor.avatarUrl} alt={activity.actor.name} />
            <AvatarFallback className={cn(config.bgColor, config.color)}>
              {activity.actor.name.charAt(0).toUpperCase()}
            </AvatarFallback>
          </Avatar>
        ) : (
          config.icon
        )}
      </div>

      {/* Content */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 flex-wrap">
          <span className="font-medium text-text-primary text-sm">
            {activity.actor.name}
          </span>
          <span className="text-text-muted text-sm">{formatAction(activity.action)}</span>
          <span className="font-medium text-text-primary text-sm truncate">
            {activity.target.name}
          </span>
        </div>
        {activity.context?.description && !compact && (
          <p className="text-xs text-text-muted mt-0.5 line-clamp-2">
            {activity.context.description}
          </p>
        )}
        <span className="text-xs text-text-muted/70 mt-1 block">{timeAgo}</span>
      </div>

      {/* Type badge */}
      <Badge variant="secondary" className={cn('text-xs flex-shrink-0', config.color)}>
        {config.label}
      </Badge>
    </motion.div>
  );
}

/**
 * Grouped activity item component
 */
interface GroupedActivityItemProps {
  group: GroupedActivities;
  compact?: boolean;
  onClick?: (activity: ActivityFeedItem) => void;
  isLast?: boolean;
}

function GroupedActivityItem({ group, compact, onClick, isLast }: GroupedActivityItemProps) {
  const config = ACTIVITY_CONFIG[group.type];
  const firstActivity = group.activities[0];
  const timeAgo = formatDistanceToNow(new Date(group.timestamp), { addSuffix: true });
  const [isExpanded, setIsExpanded] = useState(false);

  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: -20 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, x: -100 }}
      transition={{ duration: 0.3, ease: 'easeOut' }}
      className={cn(
        'relative p-3 rounded-lg transition-colors',
        'hover:bg-bg-secondary/50',
        !isLast && 'border-b border-border-subtle'
      )}
    >
      {/* Timeline connector */}
      {!isLast && !compact && (
        <div className="absolute left-[1.6rem] top-10 bottom-0 w-px bg-border-subtle" />
      )}

      {/* Group header */}
      <div
        className="flex items-start gap-3 cursor-pointer"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        {/* Icon */}
        <div
          className={cn(
            'relative flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center',
            config.bgColor,
            config.color
          )}
        >
          {config.icon}
          {group.count > 1 && (
            <span className="absolute -top-1 -right-1 w-4 h-4 bg-brand-500 text-white text-[10px] rounded-full flex items-center justify-center font-medium">
              {group.count}
            </span>
          )}
        </div>

        {/* Content */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-medium text-text-primary text-sm">
              {firstActivity.actor.name}
            </span>
            <span className="text-text-muted text-sm">
              {formatAction(firstActivity.action)} {group.count} {firstActivity.target.type}s
            </span>
          </div>
          <span className="text-xs text-text-muted/70 mt-1 block">{timeAgo}</span>
        </div>

        {/* Expand indicator */}
        {group.count > 1 && (
          <Badge variant="secondary" className="text-xs flex-shrink-0">
            {isExpanded ? 'Collapse' : 'Expand'}
          </Badge>
        )}
      </div>

      {/* Expanded items */}
      <AnimatePresence>
        {isExpanded && group.count > 1 && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            className="mt-2 ml-11 space-y-1"
          >
            {group.activities.map((activity, idx) => (
              <div
                key={activity.id}
                className="text-sm text-text-muted py-1 px-2 rounded hover:bg-bg-tertiary cursor-pointer"
                onClick={() => onClick?.(activity)}
              >
                <span className="font-medium text-text-primary">{activity.target.name}</span>
                <span className="text-xs ml-2">
                  {formatDistanceToNow(new Date(activity.timestamp), { addSuffix: true })}
                </span>
              </div>
            ))}
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  );
}

// ============================================================================
// Main Component
// ============================================================================

export function RealtimeActivityFeed({
  className,
  maxItems = 50,
  filter,
  showHeader = true,
  compact = false,
  onActivityClick,
}: RealtimeActivityFeedProps) {
  const [activities, setActivities] = useState<ActivityFeedItem[]>([]);
  const [isPaused, setIsPaused] = useState(false);
  const [typeFilter, setTypeFilter] = useState<ActivityType | 'all'>('all');
  const [bufferedActivities, setBufferedActivities] = useState<ActivityFeedItem[]>([]);
  const activitiesRef = useRef<ActivityFeedItem[]>([]);

  // Update ref when activities change
  useEffect(() => {
    activitiesRef.current = activities;
  }, [activities]);

  // Handle new activity from WebSocket
  const handleNewActivity = useCallback((event: ActivityEvent) => {
    if (event.activity) {
      if (isPaused) {
        // Buffer activities when paused
        setBufferedActivities((prev) => [event.activity, ...prev]);
      } else {
        setActivities((prev) => {
          const newActivities = [event.activity, ...prev].slice(0, maxItems);
          return newActivities;
        });
      }
    }
  }, [isPaused, maxItems]);

  // Subscribe to real-time activity updates
  const { isConnected, error: wsError } = useRealtimeSubscription<ActivityEvent>(
    'activity',
    'new_activity',
    handleNewActivity
  );

  // Resume updates and flush buffer
  const handleResume = useCallback(() => {
    setIsPaused(false);
    if (bufferedActivities.length > 0) {
      setActivities((prev) => {
        const newActivities = [...bufferedActivities, ...prev].slice(0, maxItems);
        return newActivities;
      });
      setBufferedActivities([]);
    }
  }, [bufferedActivities, maxItems]);

  // Clear all activities
  const handleClear = useCallback(() => {
    setActivities([]);
    setBufferedActivities([]);
  }, []);

  // Filter activities
  const filteredActivities = useMemo(() => {
    let filtered = activities;

    // Apply type filter
    if (typeFilter !== 'all') {
      filtered = filtered.filter((activity) => getActivityType(activity) === typeFilter);
    }

    // Apply prop filter if provided
    if (filter && filter.length > 0) {
      filtered = filtered.filter((activity) => filter.includes(getActivityType(activity)));
    }

    return filtered;
  }, [activities, typeFilter, filter]);

  // Group activities
  const groupedActivities = useMemo(() => {
    return groupActivities(filteredActivities);
  }, [filteredActivities]);

  return (
    <TooltipProvider>
      <div
        className={cn(
          'flex flex-col bg-bg-primary border border-border-default rounded-xl overflow-hidden shadow-lg',
          className
        )}
        onMouseEnter={() => setIsPaused(true)}
        onMouseLeave={handleResume}
      >
        {/* Header */}
        {showHeader && (
          <div className="relative bg-bg-secondary/80 backdrop-blur-md border-b border-border-subtle">
            <div className="flex items-center justify-between px-4 py-3">
              <div className="flex items-center gap-3">
                <Activity className="h-5 w-5 text-text-primary" />
                <h2 className="text-lg font-semibold text-text-primary">Live Activity</h2>
                <LiveIndicator isConnected={isConnected} error={wsError} />
              </div>

              <div className="flex items-center gap-1">
                {/* Filter */}
                <ActivityFilter value={typeFilter} onChange={setTypeFilter} />

                {/* Clear button */}
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8"
                      onClick={handleClear}
                      disabled={activities.length === 0}
                      aria-label="Clear activity history"
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Clear history</TooltipContent>
                </Tooltip>
              </div>
            </div>

            {/* Buffered indicator */}
            {bufferedActivities.length > 0 && (
              <motion.div
                initial={{ opacity: 0, y: -10 }}
                animate={{ opacity: 1, y: 0 }}
                className="px-4 py-2 bg-brand-500/10 border-t border-brand-500/20"
              >
                <button
                  onClick={handleResume}
                  className="text-sm text-brand-500 hover:text-brand-400 font-medium"
                >
                  {bufferedActivities.length} new {bufferedActivities.length === 1 ? 'activity' : 'activities'} - Click to show
                </button>
              </motion.div>
            )}
          </div>
        )}

        {/* Activity list */}
        <ScrollArea className={cn('flex-1', compact ? 'max-h-[300px]' : 'max-h-[500px]')}>
          {groupedActivities.length === 0 ? (
            <EmptyState />
          ) : (
            <div className="divide-y divide-border-subtle">
              <AnimatePresence mode="popLayout">
                {groupedActivities.map((item, index) => {
                  const isLast = index === groupedActivities.length - 1;

                  if ('count' in item) {
                    // Grouped activity
                    return (
                      <GroupedActivityItem
                        key={item.key}
                        group={item}
                        compact={compact}
                        onClick={onActivityClick}
                        isLast={isLast}
                      />
                    );
                  }

                  // Single activity
                  return (
                    <ActivityItem
                      key={item.id}
                      activity={item}
                      compact={compact}
                      onClick={onActivityClick}
                      isLast={isLast}
                    />
                  );
                })}
              </AnimatePresence>
            </div>
          )}
        </ScrollArea>

        {/* Footer with count */}
        {showHeader && (
          <div className="px-4 py-2 bg-bg-secondary/50 border-t border-border-subtle">
            <div className="flex items-center justify-between text-xs text-text-muted">
              <span>
                {filteredActivities.length} {filteredActivities.length === 1 ? 'activity' : 'activities'}
              </span>
              {isPaused && <span className="text-brand-500">Paused</span>}
            </div>
          </div>
        )}
      </div>
    </TooltipProvider>
  );
}

export default RealtimeActivityFeed;
