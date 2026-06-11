'use client';

import { useAtlasRuns } from '@/hooks/useAtlasObservability';
import { useTranslation } from 'react-i18next';
import { Badge } from '@/components/ui/badge';
import { Clock, DollarSign, Bot } from 'lucide-react';

interface AgentRunTimelineProps {
  events: any[];
}

const EVENT_COLORS = {
  INPUT: 'bg-blue-500',
  DECISION: 'bg-purple-500',
  ACTION: 'bg-green-500',
  RESULT: 'bg-teal-500',
  ERROR: 'bg-red-500',
};

const EVENT_LABELS = {
  INPUT: 'Input',
  DECISION: 'Decision',
  ACTION: 'Action',
  RESULT: 'Result',
  ERROR: 'Error',
};

export default function AgentRunTimeline({ events }: AgentRunTimelineProps) {
  const { t } = useTranslation();

  if (!events || events.length === 0) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        No events to display
      </div>
    );
  }

  return (
    <div className="relative">
      <div className="absolute left-4 top-0 bottom-0 w-0.5 bg-border" />

      <div className="space-y-4">
        {events.map((event, index) => (
          <div key={event.event_id || index} className="relative pl-10">
            <div
              className={`absolute left-2.5 top-2 w-3 h-3 rounded-full ${
                EVENT_COLORS[event.kind as keyof typeof EVENT_COLORS] || 'bg-gray-500'
              }`}
            />

            <div className="p-4 rounded-lg border bg-card">
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <Badge variant="outline">
                    {EVENT_LABELS[event.kind as keyof typeof EVENT_LABELS] || event.kind}
                  </Badge>
                  <span className="text-xs text-muted-foreground font-mono">
                    seq: {event.sequence}
                  </span>
                </div>
                <span className="text-xs text-muted-foreground">
                  {new Date(event.timestamp).toLocaleTimeString()}
                </span>
              </div>

              <div className="text-sm font-mono text-muted-foreground mb-2">
                {event.system_id}
              </div>

              {event.payload && (
                <div className="mt-2 p-2 rounded bg-muted/50 text-xs font-mono overflow-x-auto">
                  {typeof event.payload === 'string'
                    ? event.payload
                    : JSON.stringify(event.payload, null, 2).slice(0, 200)}
                </div>
              )}

              {event.payload?.cost && (
                <div className="mt-2 flex items-center gap-4 text-xs">
                  <span className="flex items-center gap-1">
                    <DollarSign className="h-3 w-3" />
                    ${event.payload.cost.cost_usd?.toFixed(6) || '0'}
                  </span>
                  <span>
                    {event.payload.cost.input_tokens || 0} in / {event.payload.cost.output_tokens || 0} out
                  </span>
                </div>
              )}

              {event.kind === 'DECISION' && event.payload?.reasoning && (
                <div className="mt-2 text-sm text-foreground">
                  {event.payload.reasoning.slice(0, 150)}
                  {event.payload.reasoning.length > 150 && '...'}
                </div>
              )}

              {event.kind === 'ACTION' && event.payload?.tool_name && (
                <div className="mt-2 text-sm">
                  <span className="font-medium">Tool:</span> {event.payload.tool_name}
                </div>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}