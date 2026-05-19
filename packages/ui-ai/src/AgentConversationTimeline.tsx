/**
 * @functionfly/ui-ai
 * Agent Conversation Timeline - Timeline view of agent conversations
 */

import * as React from "react";
import { cn } from "@functionfly/ui-core";
import { Badge } from "@functionfly/ui-core";
import { Clock, MessageSquare, User, Bot, ChevronRight, Search, Filter } from "lucide-react";

export interface TimelineEvent {
  id: string;
  agentId: string;
  agentName: string;
  type: "message" | "action" | "decision" | "tool_call" | "error";
  content: string;
  timestamp: number;
  metadata?: Record<string, string>;
}

export interface AgentConversationTimelineProps {
  events: TimelineEvent[];
  onEventClick?: (event: TimelineEvent) => void;
  onAgentFilter?: (agentId: string | null) => void;
  className?: string;
}

const eventTypeConfig = {
  message: { icon: MessageSquare, color: "text-info", bg: "bg-info/10" },
  action: { icon: User, color: "text-success", bg: "bg-success/10" },
  decision: { icon: Bot, color: "text-warning", bg: "bg-warning/10" },
  tool_call: { icon: Clock, color: "text-brand-500", bg: "bg-brand-500/10" },
  error: { icon: MessageSquare, color: "text-error", bg: "bg-error/10" },
};

export function AgentConversationTimeline({
  events,
  onEventClick,
  onAgentFilter,
  className,
}: AgentConversationTimelineProps) {
  const [selectedAgentId, setSelectedAgentId] = React.useState<string | null>(null);
  const [searchQuery, setSearchQuery] = React.useState("");

  const agents = React.useMemo(() => {
    const agentMap = new Map<string, { id: string; name: string; count: number }>();
    events.forEach(e => {
      const existing = agentMap.get(e.agentId);
      if (existing) {
        existing.count++;
      } else {
        agentMap.set(e.agentId, { id: e.agentId, name: e.agentName, count: 1 });
      }
    });
    return Array.from(agentMap.values());
  }, [events]);

  const filteredEvents = React.useMemo(() => {
    return events.filter(e => {
      const matchesAgent = !selectedAgentId || e.agentId === selectedAgentId;
      const matchesSearch = !searchQuery || e.content.toLowerCase().includes(searchQuery.toLowerCase());
      return matchesAgent && matchesSearch;
    });
  }, [events, selectedAgentId, searchQuery]);

  // Group events by time buckets
  const groupedEvents = React.useMemo(() => {
    const now = Date.now();
    const buckets: { label: string; events: TimelineEvent[] }[] = [
      { label: "Just now", events: [] },
      { label: "Minutes ago", events: [] },
      { label: "Hour ago", events: [] },
      { label: "Earlier", events: [] },
    ];

    filteredEvents.forEach(event => {
      const diff = now - event.timestamp;
      if (diff < 60000) {
        buckets[0].events.push(event);
      } else if (diff < 3600000) {
        buckets[1].events.push(event);
      } else if (diff < 86400000) {
        buckets[2].events.push(event);
      } else {
        buckets[3].events.push(event);
      }
    });

    return buckets.filter(b => b.events.length > 0);
  }, [filteredEvents]);

  return (
    <div className={cn("flex flex-col h-full", className)}>
      {/* Header */}
      <div className="flex items-center gap-2 px-4 py-3 border-b border-border-subtle">
        <Clock className="size-4 text-brand-500" />
        <span className="text-sm font-medium text-text-primary">Timeline</span>
        <Badge variant="ghost" size="sm">{events.length} events</Badge>
      </div>

      {/* Agent Filter */}
      <div className="px-3 py-2 border-b border-border-subtle">
        <div className="flex items-center gap-2 overflow-x-auto">
          <button
            onClick={() => { setSelectedAgentId(null); onAgentFilter?.(null); }}
            className={cn(
              "flex items-center gap-1 px-2 py-1 text-xs rounded-full whitespace-nowrap transition-colors",
              !selectedAgentId
                ? "bg-brand-500 text-white"
                : "bg-bg-tertiary text-text-muted hover:text-text-primary"
            )}
          >
            All
          </button>
          {agents.map(agent => (
            <button
              key={agent.id}
              onClick={() => { setSelectedAgentId(agent.id); onAgentFilter?.(agent.id); }}
              className={cn(
                "flex items-center gap-1 px-2 py-1 text-xs rounded-full whitespace-nowrap transition-colors",
                selectedAgentId === agent.id
                  ? "bg-brand-500 text-white"
                  : "bg-bg-tertiary text-text-muted hover:text-text-primary"
              )}
            >
              {agent.name}
              <span className="text-[10px] opacity-70">({agent.count})</span>
            </button>
          ))}
        </div>
      </div>

      {/* Search */}
      <div className="px-3 py-2 border-b border-border-subtle">
        <div className="flex items-center gap-2 px-2 py-1.5 bg-bg-tertiary/50 rounded-lg">
          <Search className="size-3 text-text-muted" />
          <input
            type="text"
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            placeholder="Search events..."
            className="flex-1 bg-transparent text-xs text-text-primary outline-none placeholder:text-text-muted"
          />
        </div>
      </div>

      {/* Timeline */}
      <div className="flex-1 overflow-y-auto">
        {groupedEvents.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 text-text-muted">
            <Clock className="size-12 mb-3 opacity-50" />
            <p className="text-sm">No events</p>
          </div>
        ) : (
          <div className="p-4 space-y-6">
            {groupedEvents.map((bucket, bucketIndex) => (
              <div key={bucket.label}>
                <div className="text-[10px] font-medium text-text-muted uppercase tracking-wide mb-3">
                  {bucket.label}
                </div>
                
                <div className="relative">
                  {/* Timeline line */}
                  <div className="absolute left-4 top-0 bottom-0 w-[1px] bg-border-subtle" />
                  
                  <div className="space-y-3">
                    {bucket.events.map(event => {
                      const config = eventTypeConfig[event.type];
                      const Icon = config.icon;
                      
                      return (
                        <div
                          key={event.id}
                          onClick={() => onEventClick?.(event)}
                          className="relative flex gap-4 pl-8 cursor-pointer group"
                        >
                          {/* Timeline dot */}
                          <div className={cn(
                            "absolute left-2.5 size-3 rounded-full border-2 z-10",
                            `bg-${event.type === "error" ? "error" : "brand-500"}`,
                            `border-${event.type === "error" ? "error" : "brand-500"}`
                          )} />
                          
                          {/* Event card */}
                          <div className="flex-1 p-3 rounded-lg bg-bg-secondary border border-border-subtle hover:border-brand-500/30 transition-colors">
                            <div className="flex items-center gap-2 mb-1">
                              <div className={cn("size-5 rounded flex items-center justify-center", config.bg)}>
                                <Icon className={cn("size-3", config.color)} />
                              </div>
                              <span className="text-xs font-medium text-text-primary">{event.agentName}</span>
                              <Badge variant="ghost" size="sm" className="text-[10px]">{event.type.replace("_", " ")}</Badge>
                              <span className="text-[10px] text-text-muted ml-auto">
                                {new Date(event.timestamp).toLocaleTimeString()}
                              </span>
                            </div>
                            <p className="text-sm text-text-secondary line-clamp-2">{event.content}</p>
                            {event.metadata && (
                              <div className="flex flex-wrap gap-2 mt-2">
                                {Object.entries(event.metadata).slice(0, 3).map(([key, value]) => (
                                  <span key={key} className="text-[10px] text-text-muted">
                                    <span className="font-medium">{key}:</span> {value}
                                  </span>
                                ))}
                              </div>
                            )}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
