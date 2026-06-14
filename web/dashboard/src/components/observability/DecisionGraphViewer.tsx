'use client';

import { useEffect, useRef, useState, useCallback } from 'react';
import { useAtlasEvents } from '@/hooks/useAtlasObservability';

interface DecisionGraphViewerProps {
  runId: string | null;
  onNodeClick?: (nodeId: string) => void;
}

interface GraphNode {
  id: string;
  event_id: string;
  kind: string;
  sequence: number;
  x?: number;
  y?: number;
}

interface GraphEdge {
  from: string;
  to: string;
}

export default function DecisionGraphViewer({ runId, onNodeClick }: DecisionGraphViewerProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [nodes, setNodes] = useState<GraphNode[]>([]);
  const [edges, setEdges] = useState<GraphEdge[]>([]);
  const [hoveredNode, setHoveredNode] = useState<GraphNode | null>(null);
  const [selectedNode, setSelectedNode] = useState<GraphNode | null>(null);

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

  const getNodePosition = useCallback((node: GraphNode, index: number, total: number, centerX: number, centerY: number, radius: number) => {
    const angle = (2 * Math.PI * index) / total - Math.PI / 2;
    return {
      x: centerX + radius * Math.cos(angle),
      y: centerY + radius * Math.sin(angle),
    };
  }, []);

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

    const nodePositions: Map<string, { x: number; y: number }> = new Map();

    nodes.forEach((node, index) => {
      const pos = getNodePosition(node, index, nodes.length, centerX, centerY, radius);
      nodePositions.set(node.event_id, pos);

      const isHovered = hoveredNode?.event_id === node.event_id;
      const isSelected = selectedNode?.event_id === node.event_id;
      const nodeRadius = isHovered || isSelected ? 24 : 20;

      ctx.beginPath();
      ctx.arc(pos.x, pos.y, nodeRadius, 0, 2 * Math.PI);
      ctx.fillStyle = nodeColors[node.kind] || '#6B7280';
      ctx.fill();

      if (isHovered || isSelected) {
        ctx.strokeStyle = '#fff';
        ctx.lineWidth = 3;
        ctx.stroke();
      } else {
        ctx.strokeStyle = '#fff';
        ctx.lineWidth = 2;
        ctx.stroke();
      }

      ctx.fillStyle = '#fff';
      ctx.font = '10px sans-serif';
      ctx.textAlign = 'center';
      ctx.fillText(String(node.sequence), pos.x, pos.y + 4);
    });

    edges.forEach((edge) => {
      const fromPos = nodePositions.get(edge.from);
      const toPos = nodePositions.get(edge.to);

      if (fromPos && toPos) {
        ctx.beginPath();
        ctx.moveTo(fromPos.x, fromPos.y);
        ctx.lineTo(toPos.x, toPos.y);
        ctx.strokeStyle = '#9CA3AF';
        ctx.lineWidth = 1;
        ctx.stroke();
      }
    });

    if (hoveredNode) {
      const pos = nodePositions.get(hoveredNode.event_id);
      if (pos) {
        ctx.fillStyle = 'rgba(0, 0, 0, 0.8)';
        ctx.fillRect(pos.x + 30, pos.y - 10, 80, 24);
        ctx.fillStyle = '#fff';
        ctx.font = '10px sans-serif';
        ctx.textAlign = 'left';
        ctx.fillText(hoveredNode.kind, pos.x + 35, pos.y + 4);
      }
    }
  }, [nodes, edges, hoveredNode, selectedNode, getNodePosition]);

  const handleCanvasClick = useCallback((event: React.MouseEvent<HTMLCanvasElement>) => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const rect = canvas.getBoundingClientRect();
    const x = event.clientX - rect.left;
    const y = event.clientY - rect.top;

    const centerX = canvas.width / 2;
    const centerY = canvas.height / 2;
    const radius = Math.min(canvas.width, canvas.height) / 3;

    let clickedNode: GraphNode | null = null;

    nodes.forEach((node, index) => {
      const pos = getNodePosition(node, index, nodes.length, centerX, centerY, radius);
      const distance = Math.sqrt((x - pos.x) ** 2 + (y - pos.y) ** 2);
      if (distance <= 24) {
        clickedNode = node;
      }
    });

    if (clickedNode) {
      setSelectedNode(clickedNode);
      onNodeClick?.(clickedNode.event_id);
    }
  }, [nodes, getNodePosition, onNodeClick]);

  const handleCanvasMouseMove = useCallback((event: React.MouseEvent<HTMLCanvasElement>) => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const rect = canvas.getBoundingClientRect();
    const x = event.clientX - rect.left;
    const y = event.clientY - rect.top;

    const centerX = canvas.width / 2;
    const centerY = canvas.height / 2;
    const radius = Math.min(canvas.width, canvas.height) / 3;

    let hovered: GraphNode | null = null;

    nodes.forEach((node, index) => {
      const pos = getNodePosition(node, index, nodes.length, centerX, centerY, radius);
      const distance = Math.sqrt((x - pos.x) ** 2 + (y - pos.y) ** 2);
      if (distance <= 24) {
        hovered = node;
      }
    });

    setHoveredNode(hovered);
  }, [nodes, getNodePosition]);

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
        {selectedNode && (
          <span className="text-xs text-muted-foreground">
            Selected: {selectedNode.kind} #{selectedNode.sequence}
          </span>
        )}
      </div>

      <canvas
        ref={canvasRef}
        width={600}
        height={400}
        className="w-full border rounded-lg bg-muted/20 cursor-pointer"
        onClick={handleCanvasClick}
        onMouseMove={handleCanvasMouseMove}
      />

      <div className="text-center text-xs text-muted-foreground">
        {nodes.length} nodes, {edges.length} edges. Click a node to view details.
      </div>
    </div>
  );
}
