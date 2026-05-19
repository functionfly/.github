import React, { useState } from "react";
import { NeuralExecutionMap } from "@functionfly/ui-visualization";
import { RotateCcw, Maximize2, Minimize2 } from "lucide-react";

interface VisualizationNode {
  id: string;
  label: string;
  type: string;
  status: "pending" | "running" | "success" | "error";
}

interface VisualizationEdge {
  source: string;
  target: string;
  strength: number;
}

interface VisualizationPanelProps {
  nodes: VisualizationNode[];
  edges: VisualizationEdge[];
}

export function VisualizationPanel({ nodes, edges }: VisualizationPanelProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [showParticles, setShowParticles] = useState(true);

  return (
    <div className="h-full flex flex-col">
      <div className="flex items-center justify-between px-4 py-2 border-b border-border-subtle">
        <span className="text-sm font-medium">Neural Execution Map</span>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowParticles(!showParticles)}
            className={`px-2 py-1 text-[10px] rounded transition-colors ${
              showParticles
                ? "bg-brand-500/20 text-brand-400"
                : "bg-bg-primary hover:bg-bg-hover"
            }`}
          >
            Particles
          </button>
          <button
            onClick={() => setIsExpanded(!isExpanded)}
            className="p-1 rounded hover:bg-bg-hover transition-colors"
          >
            {isExpanded ? (
              <Minimize2 className="size-4 text-text-muted" />
            ) : (
              <Maximize2 className="size-4 text-text-muted" />
            )}
          </button>
          <button className="p-1 rounded hover:bg-bg-hover transition-colors">
            <RotateCcw className="size-4 text-text-muted" />
          </button>
        </div>
      </div>

      <div className="flex-1 relative">
        {nodes.length === 0 ? (
          <div className="flex items-center justify-center h-full text-text-muted text-sm">
            No nodes to visualize. Add nodes to your workflow graph.
          </div>
        ) : (
          <NeuralExecutionMap
            nodes={nodes}
            connections={edges}
          />
        )}
      </div>
    </div>
  );
}