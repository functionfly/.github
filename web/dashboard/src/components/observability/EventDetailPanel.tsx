'use client';

import { useState } from 'react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ChevronLeft, ChevronRight, Copy, Check } from 'lucide-react';

interface EventDetailPanelProps {
  events: any[];
}

export default function EventDetailPanel({ events }: EventDetailPanelProps) {
  const [selectedIndex, setSelectedIndex] = useState<number | null>(null);
  const [copied, setCopied] = useState(false);

  const selectedEvent = selectedIndex !== null ? events[selectedIndex] : null;

  const copyToClipboard = () => {
    if (selectedEvent) {
      navigator.clipboard.writeText(JSON.stringify(selectedEvent, null, 2));
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  if (!events || events.length === 0) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        No events to display
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setSelectedIndex(selectedIndex !== null ? selectedIndex - 1 : events.length - 1)}
            disabled={selectedIndex === null && events.length === 0}
          >
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <span className="text-sm">
            {selectedIndex !== null ? `${selectedIndex + 1} of ${events.length}` : `${events.length} events`}
          </span>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setSelectedIndex(selectedIndex !== null ? selectedIndex + 1 : 0)}
            disabled={selectedIndex !== null && selectedIndex >= events.length - 1}
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>

        {selectedEvent && (
          <Button variant="outline" size="sm" onClick={copyToClipboard}>
            {copied ? <Check className="h-4 w-4 mr-1" /> : <Copy className="h-4 w-4 mr-1" />}
            {copied ? 'Copied!' : 'Copy'}
          </Button>
        )}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div className="space-y-2 max-h-[400px] overflow-y-auto">
          {events.map((event, index) => (
            <div
              key={event.event_id || index}
              className={`p-3 rounded-lg border cursor-pointer transition-colors ${
                selectedIndex === index
                  ? 'border-purple-500 bg-purple-50 dark:bg-purple-900/20'
                  : 'border-border hover:border-purple-300'
              }`}
              onClick={() => setSelectedIndex(index)}
            >
              <div className="flex items-center justify-between">
                <Badge variant="outline">{event.kind}</Badge>
                <span className="text-xs text-muted-foreground font-mono">
                  #{event.sequence}
                </span>
              </div>
              <div className="mt-1 text-xs text-muted-foreground truncate">
                {event.system_id}
              </div>
            </div>
          ))}
        </div>

        <div className="border rounded-lg p-4">
          {selectedEvent ? (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="font-medium">Event Details</h3>
                <Badge variant="outline">{selectedEvent.kind}</Badge>
              </div>

              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Event ID:</span>
                  <span className="font-mono text-xs">{selectedEvent.event_id}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Sequence:</span>
                  <span className="font-mono">{selectedEvent.sequence}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Timestamp:</span>
                  <span className="text-xs">
                    {new Date(selectedEvent.timestamp).toLocaleString()}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">System ID:</span>
                  <span className="font-mono text-xs">{selectedEvent.system_id}</span>
                </div>
              </div>

              <div>
                <h4 className="text-sm font-medium mb-2">Payload</h4>
                <pre className="p-3 rounded bg-muted/50 text-xs overflow-x-auto max-h-[300px]">
                  {JSON.stringify(selectedEvent.payload, null, 2)}
                </pre>
              </div>
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              Select an event to view details
            </div>
          )}
        </div>
      </div>
    </div>
  );
}