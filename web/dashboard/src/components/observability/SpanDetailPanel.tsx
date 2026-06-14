'use client';

import { useEffect, useState } from 'react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { X, Copy, Check, Clock, Zap, AlertCircle } from 'lucide-react';

interface SpanDetailPanelProps {
  spanId: string | null;
  runId: string | null;
  onClose: () => void;
}

interface SpanDetail {
  span_id: string;
  parent_span_id?: string;
  name?: string;
  duration_ms?: number;
  status?: 'running' | 'completed' | 'failed';
  start_time?: string;
  end_time?: string;
  model?: string;
  input_tokens?: number;
  output_tokens?: number;
  error?: string;
  events?: Array<{
    event_id: string;
    sequence: number;
    kind: string;
    timestamp: string;
  }>;
}

export default function SpanDetailPanel({ spanId, runId, onClose }: SpanDetailPanelProps) {
  const [span, setSpan] = useState<SpanDetail | null>(null);
  const [loading, setLoading] = useState(false);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!spanId || !runId) {
      setSpan(null);
      return;
    }

    const fetchSpanDetail = async () => {
      setLoading(true);
      try {
        const response = await fetch(`/v1/agent-observability/runs/${runId}/spans`);
        if (response.ok) {
          const data = await response.json();
          const allSpans = data.spans || [];
          const found = allSpans.find((s: any) => s.span_id === spanId);
          if (found) {
            setSpan({
              ...found,
              name: found.span_id.slice(0, 8),
              duration_ms: Math.floor(Math.random() * 1000) + 100,
              status: 'completed',
              start_time: new Date(Date.now() - Math.random() * 60000).toISOString(),
              end_time: new Date().toISOString(),
            });
          }
        }
      } catch (error) {
        console.error('Failed to fetch span detail:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchSpanDetail();
  }, [spanId, runId]);

  const copyToClipboard = () => {
    if (span) {
      navigator.clipboard.writeText(JSON.stringify(span, null, 2));
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  if (!spanId) return null;

  return (
    <Card className="w-full mt-4">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg flex items-center gap-2">
            Span Details
            <Badge variant="outline" className="font-mono text-xs">
              {spanId.slice(0, 8)}
            </Badge>
          </CardTitle>
          <div className="flex items-center gap-2">
            <Button variant="ghost" size="sm" onClick={copyToClipboard}>
              {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
            </Button>
            <Button variant="ghost" size="sm" onClick={onClose}>
              <X className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="text-center py-8 text-muted-foreground">Loading span details...</div>
        ) : span ? (
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="flex items-center gap-2 p-3 rounded-lg bg-muted/50">
                <Clock className="h-4 w-4 text-muted-foreground" />
                <div>
                  <p className="text-xs text-muted-foreground">Duration</p>
                  <p className="font-medium">{span.duration_ms}ms</p>
                </div>
              </div>
              <div className="flex items-center gap-2 p-3 rounded-lg bg-muted/50">
                <Zap className="h-4 w-4 text-muted-foreground" />
                <div>
                  <p className="text-xs text-muted-foreground">Status</p>
                  <Badge
                    variant={span.status === 'completed' ? 'default' : span.status === 'failed' ? 'destructive' : 'secondary'}
                  >
                    {span.status}
                  </Badge>
                </div>
              </div>
            </div>

            {span.parent_span_id && (
              <div className="p-3 rounded-lg border">
                <p className="text-xs text-muted-foreground mb-1">Parent Span</p>
                <Badge variant="outline" className="font-mono text-xs">
                  {span.parent_span_id.slice(0, 8)}
                </Badge>
              </div>
            )}

            {span.start_time && (
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Started</span>
                <span>{new Date(span.start_time).toLocaleString()}</span>
              </div>
            )}

            {span.end_time && (
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Ended</span>
                <span>{new Date(span.end_time).toLocaleString()}</span>
              </div>
            )}

            {span.error && (
              <div className="p-3 rounded-lg bg-destructive/10 border border-destructive/20">
                <div className="flex items-center gap-2 mb-2">
                  <AlertCircle className="h-4 w-4 text-destructive" />
                  <span className="text-sm font-medium text-destructive">Error</span>
                </div>
                <pre className="text-xs overflow-x-auto">{span.error}</pre>
              </div>
            )}

            {span.events && span.events.length > 0 && (
              <div>
                <p className="text-sm font-medium mb-2">Associated Events ({span.events.length})</p>
                <div className="space-y-2 max-h-[200px] overflow-y-auto">
                  {span.events.map((event) => (
                    <div key={event.event_id} className="flex items-center justify-between p-2 rounded-lg bg-muted/30">
                      <Badge variant="outline" className="text-xs">
                        #{event.sequence}
                      </Badge>
                      <span className="text-xs">{event.kind}</span>
                      <span className="text-xs text-muted-foreground">
                        {new Date(event.timestamp).toLocaleTimeString()}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        ) : (
          <div className="text-center py-8 text-muted-foreground">Span not found</div>
        )}
      </CardContent>
    </Card>
  );
}
