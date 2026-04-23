import { useState, useRef } from "react";
import { useVirtualizer } from "@tanstack/react-virtual";
import {
  Search,
  Filter,
  Download,
  ChevronLeft,
  ChevronRight,
  Clock,
  Hash,
  FileJson,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { useStateFabricEventLogs } from "@/hooks/useStateFabric";
import type { EventLog } from "@/types";

interface EventLogViewerProps {
  fabricId: string;
  events: EventLog[];
  total: number;
}

const eventTypeColors: Record<string, string> = {
  create: "bg-green-500/10 text-green-400 border-green-500/20",
  update: "bg-blue-500/10 text-blue-400 border-blue-500/20",
  delete: "bg-red-500/10 text-red-400 border-red-500/20",
  snapshot: "bg-purple-500/10 text-purple-400 border-purple-500/20",
  sync: "bg-yellow-500/10 text-yellow-400 border-yellow-500/20",
};

const ITEM_HEIGHT = 56; // Height of each event row in pixels
const OVERSCAN = 5;

export function EventLogViewer({ fabricId, events, total }: EventLogViewerProps) {
  const [page, setPage] = useState(0);
  const [searchQuery, setSearchQuery] = useState("");
  const [eventTypeFilter, setEventTypeFilter] = useState<string>("");
  const [selectedEvent, setSelectedEvent] = useState<EventLog | null>(null);

  // Virtual scrolling container ref
  const parentRef = useRef<HTMLDivElement>(null);

  const pageSize = 20;
  const { data, isLoading } = useStateFabricEventLogs(fabricId, {
    limit: pageSize,
    offset: page * pageSize,
    eventType: eventTypeFilter || undefined,
  });

  const displayedEvents = data?.events || events;
  const displayedTotal = data?.total || total;
  const totalPages = Math.ceil(displayedTotal / pageSize);

  const filteredEvents = displayedEvents.filter((event) =>
    searchQuery
      ? event.correlationId?.toLowerCase().includes(searchQuery.toLowerCase()) ||
        event.eventType.toLowerCase().includes(searchQuery.toLowerCase())
      : true
  );

  // Virtualizer setup for handling large event lists
  const virtualizer = useVirtualizer({
    count: filteredEvents.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => ITEM_HEIGHT,
    overscan: OVERSCAN,
  });

  const virtualItems = virtualizer.getVirtualItems();

  const handleExport = () => {
    const dataStr = JSON.stringify(displayedEvents, null, 2);
    const blob = new Blob([dataStr], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `events-${fabricId}-${new Date().toISOString()}.json`;
    link.click();
  };

  return (
    <div className="space-y-4">
      {/* Toolbar */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
          <Input
            placeholder="Search events..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
        <select
          value={eventTypeFilter}
          onChange={(e) => setEventTypeFilter(e.target.value)}
          className="px-3 py-2 bg-bg-secondary border border-border-subtle rounded-lg text-text-primary text-sm"
        >
          <option value="">All Types</option>
          <option value="create">Create</option>
          <option value="update">Update</option>
          <option value="delete">Delete</option>
          <option value="snapshot">Snapshot</option>
          <option value="sync">Sync</option>
        </select>
        <Button variant="outline" onClick={handleExport}>
          <Download className="w-4 h-4 mr-2" />
          Export
        </Button>
      </div>

      {/* Event List */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center justify-between">
            <span>Event Log</span>
            <span className="text-sm text-text-muted font-normal">
              {displayedTotal.toLocaleString()} total events
            </span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="py-8 text-center text-text-muted">Loading events...</div>
          ) : filteredEvents.length === 0 ? (
            <div className="py-8 text-center text-text-muted">No events found</div>
          ) : (
            <div
              ref={parentRef}
              className="overflow-auto"
              style={{ height: `${Math.min(filteredEvents.length * ITEM_HEIGHT, 400)}px` }}
            >
              <div
                style={{
                  height: `${virtualizer.getTotalSize()}px`,
                  width: "100%",
                  position: "relative",
                }}
              >
                {virtualItems.map((virtualItem) => {
                  const event = filteredEvents[virtualItem.index];
                  return (
                    <Dialog key={event.id}>
                      <DialogTrigger asChild>
                        <div
                          className="absolute left-0 w-full flex items-center justify-between p-3 rounded-lg bg-bg-secondary/50 hover:bg-bg-secondary cursor-pointer transition-colors"
                          style={{
                            height: `${virtualItem.size}px`,
                            transform: `translateY(${virtualItem.start}px)`,
                          }}
                          onClick={() => setSelectedEvent(event)}
                        >
                          <div className="flex items-center gap-3">
                            <Badge
                              variant="outline"
                              className={eventTypeColors[event.eventType] || "bg-gray-500/10"}
                            >
                              {event.eventType}
                            </Badge>
                            <div className="flex items-center gap-2 text-sm text-text-muted">
                              <Hash className="w-3 h-3" />
                              <span className="font-mono">
                                #{event.sequenceNumber.toLocaleString()}
                              </span>
                            </div>
                            {event.correlationId && (
                              <span className="text-xs text-text-muted font-mono">
                                {event.correlationId.slice(0, 8)}...
                              </span>
                            )}
                          </div>
                          <div className="flex items-center gap-2 text-sm text-text-muted">
                            <Clock className="w-3 h-3" />
                            {new Date(event.timestamp).toLocaleString()}
                          </div>
                        </div>
                      </DialogTrigger>
                      <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
                        <DialogHeader>
                          <DialogTitle className="flex items-center gap-3">
                            <Badge
                              variant="outline"
                              className={eventTypeColors[event.eventType] || "bg-gray-500/10"}
                            >
                              {event.eventType}
                            </Badge>
                            <span className="font-mono text-sm">Event {event.sequenceNumber}</span>
                          </DialogTitle>
                        </DialogHeader>
                        <div className="space-y-4 mt-4">
                          <div className="grid grid-cols-2 gap-4 text-sm">
                            <div>
                              <span className="text-text-muted">Event ID:</span>
                              <p className="font-mono">{event.id}</p>
                            </div>
                            <div>
                              <span className="text-text-muted">Timestamp:</span>
                              <p>{new Date(event.timestamp).toLocaleString()}</p>
                            </div>
                            <div>
                              <span className="text-text-muted">Correlation ID:</span>
                              <p className="font-mono">{event.correlationId || "N/A"}</p>
                            </div>
                            <div>
                              <span className="text-text-muted">Store ID:</span>
                              <p className="font-mono">{event.storeId || "N/A"}</p>
                            </div>
                          </div>
                          <div>
                            <span className="text-text-muted text-sm">Payload:</span>
                            <pre className="mt-2 p-4 bg-bg-secondary rounded-lg overflow-x-auto text-xs font-mono">
                              {JSON.stringify(event.payload, null, 2)}
                            </pre>
                          </div>
                        </div>
                      </DialogContent>
                    </Dialog>
                  );
                })}
              </div>
            </div>
          )}

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between mt-4 pt-4 border-t border-border-subtle">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage((p) => Math.max(0, p - 1))}
                disabled={page === 0}
              >
                <ChevronLeft className="w-4 h-4 mr-2" />
                Previous
              </Button>
              <span className="text-sm text-text-muted">
                Page {page + 1} of {totalPages}
              </span>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
                disabled={page >= totalPages - 1}
              >
                Next
                <ChevronRight className="w-4 h-4 ml-2" />
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
