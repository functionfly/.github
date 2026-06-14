'use client';

import { useEffect, useState } from 'react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { X, Copy, Check, Clock } from 'lucide-react';

interface GraphNodeDetailProps {
  nodeId: string | null;
  runId: string | null;
  onClose: () => void;
}

interface GraphNodeDetail {
  event_id: string;
  kind: string;
  sequence: number;
  timestamp: string;
  payload?: any;
}

const kindColors: Record<string, string> = {
  INPUT: 'bg-blue-500',
  DECISION: 'bg-purple-500',
  ACTION: 'bg-green-500',
  RESULT: 'bg-teal-500',
  ERROR: 'bg-red-500',
};

export default function GraphNodeDetail({ nodeId, runId, onClose }: GraphNodeDetailProps) {
  const [node, setNode] = useState<GraphNodeDetail | null>(null);
  const [loading, setLoading] = useState(false);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!nodeId || !runId) {
      setNode(null);
      return;
    }

    const fetchNodeDetail = async () => {
      setLoading(true);
      try {
        const response = await fetch(`/v1/agent-observability/runs/${runId}/graph?event_id=${nodeId}`);
        if (response.ok) {
          const data = await response.json();
          const found = (data.nodes || []).find((n: any) => n.event_id === nodeId);
          if (found) {
            setNode(found);
          }
        }
      } catch (error) {
        console.error('Failed to fetch node detail:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchNodeDetail();
  }, [nodeId, runId]);

  const copyToClipboard = () => {
    if (node) {
      navigator.clipboard.writeText(JSON.stringify(node, null, 2));
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  if (!nodeId) return null;

  return (
    <Card className="w-full mt-4">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg flex items-center gap-2">
            Node Details
            {node && (
              <div className={`w-3 h-3 rounded-full ${kindColors[node.kind] || 'bg-gray-500'}`} />
            )}
            {node && (
              <Badge variant="outline">{node.kind}</Badge>
            )}
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
          <div className="text-center py-8 text-muted-foreground">Loading node details...</div>
        ) : node ? (
          <div className="space-y-4">
            <div className="flex items-center gap-3 text-sm">
              <div className="flex items-center gap-2 p-3 rounded-lg bg-muted/50">
                <Clock className="h-4 w-4 text-muted-foreground" />
                <div>
                  <p className="text-xs text-muted-foreground">Sequence</p>
                  <p className="font-medium">#{node.sequence}</p>
                </div>
              </div>
              <div className="flex items-center gap-2 p-3 rounded-lg bg-muted/50">
                <Clock className="h-4 w-4 text-muted-foreground" />
                <div>
                  <p className="text-xs text-muted-foreground">Timestamp</p>
                  <p className="font-medium text-xs">
                    {new Date(node.timestamp).toLocaleString()}
                  </p>
                </div>
              </div>
            </div>

            <div className="p-3 rounded-lg border">
              <p className="text-xs text-muted-foreground mb-1">Event ID</p>
              <p className="font-mono text-xs break-all">{node.event_id}</p>
            </div>

            {node.payload && (
              <div>
                <p className="text-sm font-medium mb-2">Payload</p>
                <pre className="p-3 rounded-lg bg-muted/50 text-xs overflow-x-auto max-h-[300px]">
                  {JSON.stringify(node.payload, null, 2)}
                </pre>
              </div>
            )}
          </div>
        ) : (
          <div className="text-center py-8 text-muted-foreground">Node not found</div>
        )}
      </CardContent>
    </Card>
  );
}
