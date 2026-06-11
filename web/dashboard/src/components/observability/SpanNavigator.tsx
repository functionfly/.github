'use client';

import { useEffect, useState } from 'react';
import { Badge } from '@/components/ui/badge';
import { ChevronRight, ChevronDown } from 'lucide-react';

interface SpanNavigatorProps {
  runId: string | null;
}

interface Span {
  span_id: string;
  parent_span_id?: string;
}

export default function SpanNavigator({ runId }: SpanNavigatorProps) {
  const [spans, setSpans] = useState<Span[]>([]);
  const [expanded, setExpanded] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!runId) return;

    const fetchSpans = async () => {
      setLoading(true);
      try {
        const response = await fetch(`/v1/agent-observability/runs/${runId}/spans`);
        if (response.ok) {
          const data = await response.json();
          setSpans(data.spans || []);
        }
      } catch (error) {
        console.error('Failed to fetch spans:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchSpans();
  }, [runId]);

  const toggleSpan = (spanId: string) => {
    const newExpanded = new Set(expanded);
    if (newExpanded.has(spanId)) {
      newExpanded.delete(spanId);
    } else {
      newExpanded.add(spanId);
    }
    setExpanded(newExpanded);
  };

  const getChildren = (parentId: string | undefined) => {
    return spans.filter((span) => span.parent_span_id === parentId);
  };

  const renderSpan = (span: Span, level: number = 0) => {
    const children = getChildren(span.span_id);
    const hasChildren = children.length > 0;
    const isExpanded = expanded.has(span.span_id);

    return (
      <div key={span.span_id}>
        <div
          className="flex items-center gap-2 p-2 rounded-lg hover:bg-muted/50 cursor-pointer"
          style={{ paddingLeft: `${level * 16 + 8}px` }}
          onClick={() => hasChildren && toggleSpan(span.span_id)}
        >
          {hasChildren ? (
            isExpanded ? (
              <ChevronDown className="h-4 w-4 text-muted-foreground" />
            ) : (
              <ChevronRight className="h-4 w-4 text-muted-foreground" />
            )
          ) : (
            <div className="w-4" />
          )}
          <Badge variant="outline" className="font-mono text-xs">
            {span.span_id.slice(0, 8)}
          </Badge>
          {span.parent_span_id && (
            <span className="text-xs text-muted-foreground">
              child of {span.parent_span_id.slice(0, 8)}
            </span>
          )}
        </div>
        {hasChildren && isExpanded && (
          <div>
            {children.map((child) => renderSpan(child, level + 1))}
          </div>
        )}
      </div>
    );
  };

  if (!runId) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        Select a run to view spans
      </div>
    );
  }

  if (loading) {
    return <div className="text-center py-8">Loading...</div>;
  }

  const rootSpans = getChildren(undefined);

  return (
    <div className="space-y-2">
      {rootSpans.length === 0 ? (
        <div className="text-center py-4 text-muted-foreground">
          No spans found
        </div>
      ) : (
        rootSpans.map((span) => renderSpan(span))
      )}
    </div>
  );
}