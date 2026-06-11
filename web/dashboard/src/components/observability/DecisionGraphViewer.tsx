'use client';

import { useEffect, useRef, useState } from 'react';
import { useAtlasEvents } from '@/hooks/useAtlasObservability';

interface DecisionGraphViewerProps {
  runId: string | null;
}

interface GraphNode {
  id: string;
  kind: string;
  sequence: number;
  x?: number;
  y?: number;
}

interface GraphEdge {
  from: string;
  to: string;
}

export default function DecisionGraphViewer({ runId }: DecisionGraphViewerProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [nodes, setNodes] = useState<GraphNode[]>([]);
  const [edges, setEdges] = useState<GraphEdge[]>([]);

  useEffect(() => {
    if (!runId) return;

    const fetchGraph = async () => {
      try {
        const response = await fetch(`/v1/agent-observability/runs/${runId}/graph`);
        if (response.ok) {
          const data = await response.json();
          setNodes(data.nodes || []);
          setEdges(data.edges || []);
        }
      } catch (error) {
        console.error('Failed to fetch graph:', error);
      }
    };

    fetchGraph();
  }, [runId]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    ctx.clearRect(0, 0, canvas.width, canvas.height);

    const nodeColors: Record<string, string> = {
      INPUT: '#3B82F6',
      DECISION: '#8B5CF6',
      ACTION: '#10B981',
      RESULT: '#14B8A6',
      ERROR: '#EF4444',
    };

    const centerX = canvas.width / 2;
    const centerY = canvas.height / 2;
    const radius = Math.min(canvas.width, canvas.height) / 3;

    nodes.forEach((node, index) => {
      const angle = (2 * Math.PI * index) / nodes.length - Math.PI / 2;
      const x = centerX + radius * Math.cos(angle);
      const y = centerY + radius * Math.sin(angle);

      ctx.beginPath();
      ctx.arc(x, y, 20, 0, 2 * Math.PI);
      ctx.fillStyle = nodeColors[node.kind] || '#6B7280';
      ctx.fill();
      ctx.strokeStyle = '#fff';
      ctx.lineWidth = 2;
      ctx.stroke();

      ctx.fillStyle = '#fff';
      ctx.font = '10px sans-serif';
      ctx.textAlign = 'center';
      ctx.fillText(String(node.sequence), x, y + 4);
    });

    edges.forEach((edge) => {
      const fromNode = nodes.find(n => n.id === edge.from);
      const toNode = nodes.find(n => n.id === edge.to);

      if (fromNode && toNode) {
        const fromIndex = nodes.indexOf(fromNode);
        const toIndex = nodes.indexOf(toNode);

        const fromAngle = (2 * Math.PI * fromIndex) / nodes.length - Math.PI / 2;
        const toAngle = (2 * Math.PI * toIndex) / nodes.length - Math.PI / 2;

        const fromX = centerX + radius * Math.cos(fromAngle);
        const fromY = centerY + radius * Math.sin(fromAngle);
        const toX = centerX + radius * Math.cos(toAngle);
        const toY = centerY + radius * Math.sin(toAngle);

        ctx.beginPath();
        ctx.moveTo(fromX, fromY);
        ctx.lineTo(toX, toY);
        ctx.strokeStyle = '#9CA3AF';
        ctx.lineWidth = 1;
        ctx.stroke();
      }
    });
  }, [nodes, edges]);

  if (!runId) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        Select a run to view the decision graph
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded-full bg-blue-500" />
            <span className="text-xs">Input</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded-full bg-purple-500" />
            <span className="text-xs">Decision</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded-full bg-green-500" />
            <span className="text-xs">Action</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded-full bg-teal-500" />
            <span className="text-xs">Result</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded-full bg-red-500" />
            <span className="text-xs">Error</span>
          </div>
        </div>
      </div>

      <canvas
        ref={canvasRef}
        width={600}
        height={400}
        className="w-full border rounded-lg bg-muted/20"
      />

      <div className="text-center text-xs text-muted-foreground">
        {nodes.length} nodes, {edges.length} edges
      </div>
    </div>
  );
}