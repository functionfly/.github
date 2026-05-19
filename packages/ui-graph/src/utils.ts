/**
 * @functionfly/ui-graph
 * Utility functions for graph/canvas rendering and layout
 */

import { type NodeData, type EdgeData, type CanvasViewport, type ViewMode, type NodeType } from "./types";
import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

// Node type configurations
export const NODE_CONFIG: Record<NodeType, { color: string; icon: string; label: string }> = {
  function: { color: "#f97316", icon: "⚡", label: "Function" },
  agent: { color: "#8b5cf6", icon: "🤖", label: "Agent" },
  api: { color: "#3b82f6", icon: "🔌", label: "API" },
  memory: { color: "#10b981", icon: "🧠", label: "Memory" },
  database: { color: "#ef4444", icon: "🗄️", label: "Database" },
  robot: { color: "#f59e0b", icon: "🦾", label: "Robot" },
  browser: { color: "#06b6d4", icon: "🌐", label: "Browser" },
  gpu: { color: "#ec4899", icon: "🖥️", label: "GPU" },
  workflow: { color: "#a855f7", icon: "🔀", label: "Workflow" },
  trigger: { color: "#6366f1", icon: "🎯", label: "Trigger" },
  condition: { color: "#14b8a6", icon: "🔀", label: "Condition" },
  output: { color: "#64748b", icon: "📤", label: "Output" },
};

// Generate grid positions for auto-layout
export function autoLayoutNodes(
  nodes: NodeData[],
  opts?: {
    columns?: number;
    cellWidth?: number;
    cellHeight?: number;
    gapX?: number;
    gapY?: number;
    startX?: number;
    startY?: number;
  }
): NodeData[] {
  const {
    columns = 4,
    cellWidth = 200,
    cellHeight = 150,
    gapX = 40,
    gapY = 60,
    startX = 100,
    startY = 100,
  } = opts || {};

  return nodes.map((node, index) => {
    const row = Math.floor(index / columns);
    const col = index % columns;
    return {
      ...node,
      position: {
        x: startX + col * (cellWidth + gapX),
        y: startY + row * (cellHeight + gapY),
      },
    };
  });
}

// Calculate canvas bounds from nodes
export function getCanvasBounds(
  nodes: NodeData[],
  padding = 100
): { minX: number; minY: number; maxX: number; maxY: number; width: number; height: number } {
  if (nodes.length === 0) {
    return { minX: 0, minY: 0, maxX: 800, maxY: 600, width: 800, height: 600 };
  }

  let minX = Infinity,
    minY = Infinity,
    maxX = -Infinity,
    maxY = -Infinity;

  for (const node of nodes) {
    if (!node.position) continue;
    minX = Math.min(minX, node.position.x);
    minY = Math.min(minY, node.position.y);
    maxX = Math.max(maxX, node.position.x + 200); // node width estimate
    maxY = Math.max(maxY, node.position.y + 120); // node height estimate
  }

  return {
    minX: minX - padding,
    minY: minY - padding,
    maxX: maxX + padding,
    maxY: maxY + padding,
    width: maxX - minX + padding * 2,
    height: maxY - minY + padding * 2,
  };
}

// Format node label for display
export function formatNodeLabel(label: string, maxLength = 20): string {
  if (label.length <= maxLength) return label;
  return label.slice(0, maxLength - 3) + "...";
}

// Calculate edge path (bezier curve)
export function getEdgePath(
  sourceX: number,
  sourceY: number,
  targetX: number,
  targetY: number,
  offsetX = 100
): string {
  const midX = (sourceX + targetX) / 2;
  const cp1x = sourceX + offsetX;
  const cp2x = targetX - offsetX;
  return `M ${sourceX} ${sourceY} C ${cp1x} ${sourceY}, ${cp2x} ${targetY}, ${targetX} ${targetY}`;
}

// Status color mapping
export function getStatusColor(status: string): string {
  const colors: Record<string, string> = {
    idle: "#6b7280",
    running: "#3b82f6",
    completed: "#10b981",
    error: "#ef4444",
    waiting: "#f59e0b",
  };
  return colors[status] || colors.idle;
}

// View mode config
export const VIEW_MODE_CONFIG: Record<ViewMode, { label: string; icon: string; color: string }> = {
  design: { label: "Design", icon: "✏️", color: "#f97316" },
  execute: { label: "Execute", icon: "▶️", color: "#10b981" },
  debug: { label: "Debug", icon: "🐛", color: "#ef4444" },
  monitor: { label: "Monitor", icon: "📊", color: "#3b82f6" },
  simulate: { label: "Simulate", icon: "🔮", color: "#8b5cf6" },
};

// Generate sample graph data for demo
export function generateSampleGraph(): { nodes: NodeData[]; edges: EdgeData[] } {
  const nodes: NodeData[] = [
    {
      id: "trigger-1",
      type: "trigger",
      label: "HTTP Trigger",
      description: "REST API endpoint /api/v1/process",
      status: "idle",
      position: { x: 100, y: 200 },
      outputs: [{ id: "out", type: "output", label: "request", dataType: "http_request" }],
    },
    {
      id: "auth-agent",
      type: "agent",
      label: "Auth Agent",
      description: "Validate and authenticate requests",
      status: "idle",
      position: { x: 350, y: 100 },
      inputs: [{ id: "in", type: "input", label: "request", dataType: "http_request" }],
      outputs: [{ id: "out", type: "output", label: "user", dataType: "user_context" }],
    },
    {
      id: "processor-fn",
      type: "function",
      label: "Data Processor",
      description: "Transform and enrich data",
      status: "idle",
      position: { x: 350, y: 300 },
      inputs: [{ id: "in", type: "input", label: "payload", dataType: "json" }],
      outputs: [{ id: "out", type: "output", label: "result", dataType: "json" }],
    },
    {
      id: "vector-memory",
      type: "memory",
      label: "Vector Memory",
      description: "Semantic search and context retrieval",
      status: "idle",
      position: { x: 600, y: 100 },
      inputs: [{ id: "in", type: "input", label: "query", dataType: "text" }],
      outputs: [{ id: "out", type: "output", label: "context", dataType: "embeddings" }],
    },
    {
      id: "llm-agent",
      type: "agent",
      label: "LLM Agent",
      description: "Core reasoning and generation",
      status: "idle",
      position: { x: 600, y: 300 },
      inputs: [
        { id: "ctx", type: "input", label: "context", dataType: "embeddings" },
        { id: "data", type: "input", label: "data", dataType: "json" },
      ],
      outputs: [{ id: "out", type: "output", label: "response", dataType: "text" }],
    },
    {
      id: "output-1",
      type: "output",
      label: "Response",
      description: "Final API response",
      status: "idle",
      position: { x: 850, y: 200 },
      inputs: [{ id: "in", type: "input", label: "result", dataType: "json" }],
    },
  ];

  const edges: EdgeData[] = [
    { id: "e1", source: "trigger-1", target: "auth-agent", sourcePort: "out", targetPort: "in", status: "idle" },
    { id: "e2", source: "trigger-1", target: "processor-fn", sourcePort: "out", targetPort: "in", status: "idle" },
    { id: "e3", source: "auth-agent", target: "vector-memory", sourcePort: "out", targetPort: "in", status: "idle" },
    { id: "e4", source: "vector-memory", target: "llm-agent", sourcePort: "out", targetPort: "ctx", status: "idle" },
    { id: "e5", source: "processor-fn", target: "llm-agent", sourcePort: "out", targetPort: "data", status: "idle" },
    { id: "e6", source: "llm-agent", target: "output-1", sourcePort: "out", targetPort: "in", status: "idle" },
  ];

  return { nodes, edges };
}

// Color utilities for graph rendering
export const GRAPH_THEME = {
  background: "#08080f",
  gridLine: "rgba(255, 255, 255, 0.03)",
  gridLineMajor: "rgba(255, 255, 255, 0.06)",
  nodeBorder: "rgba(255, 255, 255, 0.12)",
  nodeBg: "rgba(26, 26, 37, 0.95)",
  nodeBgHover: "rgba(37, 37, 53, 0.95)",
  nodeBgSelected: "rgba(51, 51, 77, 0.95)",
  nodeShadow: "rgba(0, 0, 0, 0.4)",
  portSize: 12,
  portBorderRadius: 4,
  edgeColor: "rgba(255, 255, 255, 0.15)",
  edgeColorActive: "#f97316",
  edgeColorError: "#ef4444",
  edgeColorSuccess: "#10b981",
  labelColor: "#a0a0b0",
  sublabelColor: "#6b6b7b",
  minimapBg: "rgba(10, 10, 15, 0.8)",
  minimapViewport: "rgba(249, 115, 22, 0.15)",
  minimaperBorder: "rgba(249, 115, 22, 0.3)",
} as const;