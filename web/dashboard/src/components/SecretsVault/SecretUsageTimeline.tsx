/**
 * SecretUsageTimeline - Timeline visualization of secret access events
 *
 * Displays a chronological list of secret access events with filtering,
 * expandable details, and pagination. Shows event types, actor information,
 * timestamps, and geographic data when available.
 *
 * @example
 * ```tsx
 * // Basic usage with events
 * <SecretUsageTimeline
 *   events={auditLogEvents}
 *   onLoadMore={() => fetchMoreEvents()}
 * />
 *
 * // With filtering controls
 * <SecretUsageTimeline
 *   events={events}
 *   filterableEventTypes={["read", "write", "rotate", "failed_access"]}
 *   onFilterChange={(filters) => setFilters(filters)}
 * />
 *
 * // Loading state
 * <SecretUsageTimeline isLoading />
 *
 * // Empty state
 * <SecretUsageTimeline events={[]} />
 * ```
 */

import { useState, useCallback, useMemo } from "react";
import {
  Clock,
  Eye,
  Edit3,
  RefreshCw,
  AlertTriangle,
  User,
  Bot,
  Key,
  Server,
  MapPin,
  Globe,
  ChevronDown,
  ChevronUp,
  Filter,
  Calendar,
  Check,
  X,
  Loader2,
  Activity,
  ShieldAlert,
  ShieldCheck,
} from "lucide-react";
import { formatDistanceToNow, format, parseISO, isValid } from "date-fns";
import { cn } from "@/lib/utils";
import type { AuditLogEntry, AuditAction, ActorType } from "@/types/vault";

import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

/** Extended audit event type with UI-specific metadata */
export type TimelineEventType = "read" | "write" | "rotate" | "failed_access" | "create" | "delete" | "revoke";

/** Actor information for display */
export interface TimelineActor {
  id: string;
  type: ActorType;
  name?: string;
  email?: string;
  avatarUrl?: string;
}

/** Geographic location information */
export interface TimelineLocation {
  country?: string;
  city?: string;
  region?: string;
  coordinates?: { lat: number; lng: number };
}

/** Extended timeline event with enriched data */
export interface TimelineEvent extends AuditLogEntry {
  actorName?: string;
  location?: TimelineLocation;
  eventType: TimelineEventType;
}

/** Date range filter */
export interface DateRangeFilter {
  from?: Date;
  to?: Date;
}

/** Active filters state */
export interface TimelineFilters {
  eventTypes?: TimelineEventType[];
  actorTypes?: ActorType[];
  dateRange?: DateRangeFilter;
  searchQuery?: string;
}

export interface SecretUsageTimelineProps {
  /** Timeline events to display */
  events?: TimelineEvent[];
  /** Whether data is loading */
  isLoading?: boolean;
  /** Whether more events can be loaded */
  hasMore?: boolean;
  /** Callback to load more events */
  onLoadMore?: () => void;
  /** Available filter options */
  filterableEventTypes?: TimelineEventType[];
  /** Callback when filters change */
  onFilterChange?: (filters: TimelineFilters) => void;
  /** Maximum number of events to show before pagination */
  pageSize?: number;
  /** Empty state message */
  emptyMessage?: string;
  /** Additional CSS classes */
  className?: string;
}

/** Event type configuration with icons and colors */
const eventTypeConfig: Record<TimelineEventType, {
  icon: typeof Eye;
  label: string;
  color: string;
  bgColor: string;
  variant: "default" | "secondary" | "destructive" | "outline";
}> = {
  read: {
    icon: Eye,
    label: "Read",
    color: "text-blue-500",
    bgColor: "bg-blue-500/10",
    variant: "default"
  },
  write: {
    icon: Edit3,
    label: "Write",
    color: "text-green-500",
    bgColor: "bg-green-500/10",
    variant: "default"
  },
  rotate: {
    icon: RefreshCw,
    label: "Rotation",
    color: "text-purple-500",
    bgColor: "bg-purple-500/10",
    variant: "secondary"
  },
  failed_access: {
    icon: AlertTriangle,
    label: "Failed Access",
    color: "text-[var(--status-revoked)]",
    bgColor: "rgba(255,107,107,0.06)",
    variant: "destructive"
  },
  create: {
    icon: ShieldCheck,
    label: "Created",
    color: "text-[var(--status-ok)]",
    bgColor: "rgba(143,255,208,0.06)",
    variant: "default"
  },
  delete: {
    icon: X,
    label: "Deleted",
    color: "text-gray-500",
    bgColor: "bg-gray-500/10",
    variant: "outline"
  },
  revoke: {
    icon: ShieldAlert,
    label: "Revoked",
    color: "text-orange-500",
    bgColor: "bg-orange-500/10",
    variant: "outline"
  },
};

/** Actor type configuration */
const actorTypeConfig: Record<ActorType, { icon: typeof User; label: string }> = {
  user: { icon: User, label: "User" },
  token: { icon: Key, label: "Access Token" },
  system: { icon: Bot, label: "System" },
  api_key: { icon: Server, label: "API Key" },
};

/** Format timestamp with relative and absolute time */
function formatTimestamp(dateString: string): { relative: string; absolute: string } {
  try {
    const date = parseISO(dateString);
    if (!isValid(date)) {
      return { relative: "Unknown", absolute: "Unknown" };
    }
    return {
      relative: formatDistanceToNow(date, { addSuffix: true }),
      absolute: format(date, "MMM d, yyyy HH:mm:ss"),
    };
  } catch {
    return { relative: "Unknown", absolute: "Unknown" };
  }
}

/** Parse user agent for device/browser info */
function parseUserAgent(userAgent?: string): { browser?: string; os?: string; device?: string } {
  if (!userAgent) return {};

  const info: { browser?: string; os?: string; device?: string } = {};

  // Simple UA parsing
  if (userAgent.includes("Chrome")) info.browser = "Chrome";
  else if (userAgent.includes("Firefox")) info.browser = "Firefox";
  else if (userAgent.includes("Safari")) info.browser = "Safari";
  else if (userAgent.includes("Edge")) info.browser = "Edge";

  if (userAgent.includes("Windows")) info.os = "Windows";
  else if (userAgent.includes("Mac")) info.os = "macOS";
  else if (userAgent.includes("Linux")) info.os = "Linux";
  else if (userAgent.includes("Android")) info.os = "Android";
  else if (userAgent.includes("iOS")) info.os = "iOS";

  return info;
}

/**
 * Skeleton loader for timeline items
 */
function TimelineSkeleton({ count = 5 }: { count?: number }) {
  return (
    <div className="space-y-4">
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="flex gap-4">
          <div className="flex flex-col items-center">
            <Skeleton className="h-8 w-8 rounded-full" />
            {i < count - 1 && <Skeleton className="w-px h-full mt-2" />}
          </div>
          <div className="flex-1 pb-6 space-y-2">
            <div className="flex items-center gap-2">
              <Skeleton className="h-4 w-24" />
              <Skeleton className="h-5 w-16 rounded-full" />
            </div>
            <Skeleton className="h-3 w-48" />
            <Skeleton className="h-16 w-full rounded-lg" />
          </div>
        </div>
      ))}
    </div>
  );
}

/**
 * Empty state component
 */
function TimelineEmpty({ message = "No events to display" }: { message?: string }) {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      <div className="h-16 w-16 rounded-full bg-(--color-bg-tertiary) flex items-center justify-center mb-4">
        <Activity className="h-8 w-8 text-(--color-text-muted)" />
      </div>
      <h4 className="text-base font-medium text-(--color-text-primary) mb-1">
        No Activity Found
      </h4>
      <p className="text-sm text-(--color-text-muted) max-w-xs">
        {message}. Secret access events will appear here once they occur.
      </p>
    </div>
  );
}

/**
 * Timeline event item component
 */
function TimelineEventItem({
  event,
  isLast,
  isExpanded,
  onToggle,
}: {
  event: TimelineEvent;
  isLast: boolean;
  isExpanded: boolean;
  onToggle: () => void;
}) {
  const config = eventTypeConfig[event.eventType];
  const Icon = config.icon;
  const actorConfig = actorTypeConfig[event.actor_type];
  const ActorIcon = actorConfig.icon;
  const timestamp = formatTimestamp(event.created_at);
  const userAgentInfo = parseUserAgent(event.user_agent);

  return (
    <div className="flex gap-4 group">
      {/* Timeline connector */}
      <div className="flex flex-col items-center">
        <div
          className={cn(
            "flex h-8 w-8 items-center justify-center rounded-full shrink-0",
            config.bgColor,
            !event.success && "rgba(255,107,107,0.06)"
          )}
        >
          <Icon className={cn("h-4 w-4", event.success ? config.color : "text-[var(--status-revoked)]")} />
        </div>
        {!isLast && (
          <div className="w-px flex-1 bg-(--border-subtle) mt-2 group-last:hidden" />
        )}
      </div>

      {/* Event content */}
      <div className={cn("flex-1 pb-6", isLast && "pb-0")}>
        {/* Header */}
        <div className="flex items-start justify-between gap-2 mb-2">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="font-medium text-(--color-text-primary)">
              {config.label}
            </span>
            <Badge
              variant={config.variant}
              className="text-xs"
            >
              {event.success ? "Success" : "Failed"}
            </Badge>
            {event.metadata?.impersonated && (
              <Badge variant="outline" className="text-xs">
                Impersonated
              </Badge>
            )}
          </div>
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <span className="text-xs text-(--color-text-muted) shrink-0">
                  {timestamp.relative}
                </span>
              </TooltipTrigger>
              <TooltipContent>
                <p className="text-xs">{timestamp.absolute}</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>

        {/* Actor info */}
        <div className="flex items-center gap-3 mb-3">
          <div className="flex items-center gap-1.5 text-sm text-(--color-text-secondary)">
            <ActorIcon className="h-3.5 w-3.5" />
            <span className="font-medium">
              {event.actorName || event.actor_id}
            </span>
            <span className="text-(--color-text-muted)">({actorConfig.label})</span>
          </div>

          {event.ip_address && (
            <div className="flex items-center gap-1.5 text-xs text-(--color-text-muted)">
              <Globe className="h-3 w-3" />
              <span>{event.ip_address}</span>
            </div>
          )}

          {event.location && (event.location.city || event.location.country) && (
            <div className="flex items-center gap-1.5 text-xs text-(--color-text-muted)">
              <MapPin className="h-3 w-3" />
              <span>
                {event.location.city && event.location.country
                  ? `${event.location.city}, ${event.location.country}`
                  : event.location.city || event.location.country}
              </span>
            </div>
          )}
        </div>

        {/* Expandable details */}
        <div
          className={cn(
            "rounded-lg border border-(--border-subtle) bg-(--color-bg-secondary)",
            "transition-all duration-200 overflow-hidden",
            isExpanded ? "p-3" : "p-0 h-0 border-transparent"
          )}
        >
          {isExpanded && (
            <div className="space-y-2 text-sm">
              {event.request_id && (
                <div className="flex items-center justify-between">
                  <span className="text-(--color-text-muted)">Request ID</span>
                  <code className="text-xs font-mono text-(--color-text-secondary)">
                    {event.request_id}
                  </code>
                </div>
              )}

              {(userAgentInfo.browser || userAgentInfo.os) && (
                <div className="flex items-center justify-between">
                  <span className="text-(--color-text-muted)">Device</span>
                  <span className="text-(--color-text-secondary)">
                    {[userAgentInfo.browser, userAgentInfo.os].filter(Boolean).join(" on ")}
                  </span>
                </div>
              )}

              {event.error_message && (
                <div className="flex items-start gap-2 p-2 rounded bg-error/10 text-error text-xs">
                  <AlertTriangle className="h-3.5 w-3.5 shrink-0 mt-0.5" />
                  <span>{event.error_message}</span>
                </div>
              )}

              {event.metadata && Object.keys(event.metadata).length > 0 && (
                <div className="pt-2 border-t border-(--border-subtle)">
                  <span className="text-(--color-text-muted) text-xs">Additional Metadata</span>
                  <pre className="mt-1 text-xs text-(--color-text-secondary) overflow-x-auto">
                    {JSON.stringify(event.metadata, null, 2)}
                  </pre>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Toggle details button */}
        <Button
          variant="ghost"
          size="sm"
          onClick={onToggle}
          className="h-7 px-2 mt-1 text-xs text-(--color-text-muted) hover:text-(--color-text-primary)"
        >
          {isExpanded ? (
            <>
              <ChevronUp className="h-3.5 w-3.5 mr-1" />
              Hide details
            </>
          ) : (
            <>
              <ChevronDown className="h-3.5 w-3.5 mr-1" />
              Show details
            </>
          )}
        </Button>
      </div>
    </div>
  );
}

/**
 * Filter controls component
 */
function TimelineFilters({
  availableTypes,
  activeFilters,
  onFilterChange,
}: {
  availableTypes: TimelineEventType[];
  activeFilters: TimelineFilters;
  onFilterChange: (filters: TimelineFilters) => void;
}) {
  const [localFilters, setLocalFilters] = useState<TimelineFilters>(activeFilters);
  const [isOpen, setIsOpen] = useState(false);

  const applyFilters = useCallback(() => {
    onFilterChange(localFilters);
    setIsOpen(false);
  }, [localFilters, onFilterChange]);

  const toggleEventType = useCallback((type: TimelineEventType) => {
    setLocalFilters((prev) => ({
      ...prev,
      eventTypes: prev.eventTypes?.includes(type)
        ? prev.eventTypes.filter((t) => t !== type)
        : [...(prev.eventTypes || []), type],
    }));
  }, []);

  const clearFilters = useCallback(() => {
    setLocalFilters({});
    onFilterChange({});
    setIsOpen(false);
  }, [onFilterChange]);

  const hasActiveFilters = Object.values(activeFilters).some(
    (v) => v !== undefined && (Array.isArray(v) ? v.length > 0 : true)
  );

  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className={cn(
            "gap-2",
            hasActiveFilters && "border-[rgba(143,255,208,0.3)] text-[var(--status-ok)]"
          )}
        >
          <Filter className="h-4 w-4" />
          Filters
          {hasActiveFilters && (
            <Badge variant="secondary" className="ml-1 h-5 px-1.5 text-xs">
              {Object.values(activeFilters).filter((v) =>
                Array.isArray(v) ? v.length > 0 : v !== undefined
              ).length}
            </Badge>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80 p-0" align="end">
        <div className="p-4 space-y-4">
          <div className="flex items-center justify-between">
            <h4 className="font-medium text-sm">Filter Events</h4>
            {hasActiveFilters && (
              <Button variant="ghost" size="sm" onClick={clearFilters} className="h-7 text-xs">
                Clear all
              </Button>
            )}
          </div>

          <Separator className="bg-(--border-subtle)" />

          {/* Event type filter */}
          <div className="space-y-2">
            <label className="text-xs font-medium text-(--color-text-muted)">
              Event Types
            </label>
            <div className="flex flex-wrap gap-2">
              {availableTypes.map((type) => {
                const config = eventTypeConfig[type];
                const isSelected = localFilters.eventTypes?.includes(type);
                return (
                  <button
                    key={type}
                    onClick={() => toggleEventType(type)}
                    className={cn(
                      "flex items-center gap-1.5 px-2 py-1 rounded-md text-xs font-medium transition-colors",
                      isSelected
                        ? cn(config.bgColor, config.color)
                        : "bg-(--color-bg-tertiary) text-(--color-text-muted) hover:bg-(--color-bg-secondary)"
                    )}
                  >
                    {isSelected && <Check className="h-3 w-3" />}
                    {config.label}
                  </button>
                );
              })}
            </div>
          </div>

          {/* Actor type filter */}
          <div className="space-y-2">
            <label className="text-xs font-medium text-(--color-text-muted)">
              Actor Types
            </label>
            <div className="flex flex-wrap gap-2">
              {(Object.keys(actorTypeConfig) as ActorType[]).map((type) => {
                const isSelected = localFilters.actorTypes?.includes(type);
                return (
                  <button
                    key={type}
                    onClick={() =>
                      setLocalFilters((prev) => ({
                        ...prev,
                        actorTypes: prev.actorTypes?.includes(type)
                          ? prev.actorTypes.filter((t) => t !== type)
                          : [...(prev.actorTypes || []), type],
                      }))
                    }
                    className={cn(
                      "px-2 py-1 rounded-md text-xs font-medium transition-colors",
                      isSelected
                        ? "rgba(143,255,208,0.06) text-[var(--status-ok)]"
                        : "bg-(--color-bg-tertiary) text-(--color-text-muted) hover:bg-(--color-bg-secondary)"
                    )}
                  >
                    {actorTypeConfig[type].label}
                  </button>
                );
              })}
            </div>
          </div>

          <Separator className="bg-(--border-subtle)" />

          <Button onClick={applyFilters} className="w-full">
            Apply Filters
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  );
}

/**
 * SecretUsageTimeline component
 *
 * Displays secret access events in a chronological timeline with filtering
 * and expandable details.
 */
export function SecretUsageTimeline({
  events = [],
  isLoading = false,
  hasMore = false,
  onLoadMore,
  filterableEventTypes = ["read", "write", "rotate", "failed_access", "create", "delete"],
  onFilterChange,
  pageSize = 20,
  emptyMessage = "No secret access events found",
  className,
}: SecretUsageTimelineProps) {
  const [expandedEvents, setExpandedEvents] = useState<Set<string>>(new Set());
  const [activeFilters, setActiveFilters] = useState<TimelineFilters>({});

  // Toggle event expansion
  const toggleEvent = useCallback((eventId: string) => {
    setExpandedEvents((prev) => {
      const next = new Set(prev);
      if (next.has(eventId)) {
        next.delete(eventId);
      } else {
        next.add(eventId);
      }
      return next;
    });
  }, []);

  // Apply filters
  const filteredEvents = useMemo(() => {
    return events.filter((event) => {
      if (activeFilters.eventTypes?.length && !activeFilters.eventTypes.includes(event.eventType)) {
        return false;
      }
      if (activeFilters.actorTypes?.length && !activeFilters.actorTypes.includes(event.actor_type)) {
        return false;
      }
      if (activeFilters.searchQuery) {
        const query = activeFilters.searchQuery.toLowerCase();
        const searchable = [
          event.actorName,
          event.actor_id,
          event.error_message,
          event.ip_address,
        ]
          .filter(Boolean)
          .join(" ")
          .toLowerCase();
        if (!searchable.includes(query)) return false;
      }
      return true;
    });
  }, [events, activeFilters]);

  // Handle filter changes
  const handleFilterChange = useCallback((filters: TimelineFilters) => {
    setActiveFilters(filters);
    onFilterChange?.(filters);
  }, [onFilterChange]);

  if (isLoading) {
    return (
      <Card className={cn("border-(--border-subtle)", className)}>
        <CardHeader>
          <CardTitle className="text-lg">Secret Usage Timeline</CardTitle>
          <CardDescription>Loading access events...</CardDescription>
        </CardHeader>
        <CardContent>
          <TimelineSkeleton />
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className={cn("border-(--border-subtle)", className)}>
      <CardHeader className="pb-4">
        <div className="flex items-start justify-between gap-4">
          <div>
            <CardTitle className="text-lg flex items-center gap-2">
              <Activity className="h-5 w-5 text-[var(--status-ok)]" />
              Secret Usage Timeline
            </CardTitle>
            <CardDescription>
              {filteredEvents.length} events
              {filteredEvents.length !== events.length && ` (filtered from ${events.length})`}
            </CardDescription>
          </div>

          <div className="flex items-center gap-2">
            {onFilterChange && (
              <TimelineFilters
                availableTypes={filterableEventTypes}
                activeFilters={activeFilters}
                onFilterChange={handleFilterChange}
              />
            )}
          </div>
        </div>

        {/* Search bar */}
        <div className="mt-4">
          <div className="relative">
            <Clock className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-(--color-text-muted)" />
            <Input
              placeholder="Search events by actor, IP, or error..."
              value={activeFilters.searchQuery || ""}
              onChange={(e) =>
                handleFilterChange({ ...activeFilters, searchQuery: e.target.value })
              }
              className="pl-10"
            />
          </div>
        </div>
      </CardHeader>

      <CardContent>
        {filteredEvents.length === 0 ? (
          <TimelineEmpty message={emptyMessage} />
        ) : (
          <ScrollArea className="h-[500px] pr-4">
            <div className="space-y-1">
              {filteredEvents.map((event, index) => (
                <TimelineEventItem
                  key={event.id}
                  event={event}
                  isLast={index === filteredEvents.length - 1}
                  isExpanded={expandedEvents.has(event.id)}
                  onToggle={() => toggleEvent(event.id)}
                />
              ))}
            </div>

            {/* Load more */}
            {hasMore && (
              <div className="flex justify-center pt-6 pb-2">
                <Button
                  variant="outline"
                  onClick={onLoadMore}
                  disabled={isLoading}
                  className="gap-2"
                >
                  {isLoading && <Loader2 className="h-4 w-4 animate-spin" />}
                  Load more events
                </Button>
              </div>
            )}
          </ScrollArea>
        )}
      </CardContent>
    </Card>
  );
}

export default SecretUsageTimeline;
