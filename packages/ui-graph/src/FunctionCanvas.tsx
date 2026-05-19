/**
 * @functionfly/ui-graph
 * FunctionCanvas - the main visual graph editor
 */

import * as React from "react";
import { cn } from "./utils";
import { ExecutionNode } from "./ExecutionNode";
import { ExecutionEdge } from "./ExecutionEdge";
import {
  CanvasProps,
  CanvasViewport,
  ViewMode,
  CanvasZoomLevel,
  CanvasState,
  GraphContextType,
  NodeData,
  EdgeData,
} from "./types";
import { useGraphInteraction } from "./hooks/useGraphInteraction";
import { generateSampleGraph, GRAPH_THEME, VIEW_MODE_CONFIG, getCanvasBounds } from "./utils";

// Default canvas state
const createDefaultState = (): CanvasState => ({
  nodes: [],
  edges: [],
  selectedNodes: [],
  selectedEdges: [],
  viewport: { x: 0, y: 0, zoom: 1, rotation: 0 },
  viewMode: "design",
  isPanning: false,
  isConnecting: false,
  hoveredNode: null,
  hoveredEdge: null,
});

const GraphContext = React.createContext<GraphContextType | null>(null);
export { GraphContext };

export function FunctionCanvas({
  nodes: propNodes,
  edges: propEdges,
  viewMode: propViewMode = "design",
  zoomLevel = "normal",
  viewport: propViewport,
  onNodeSelect,
  onNodeDoubleClick,
  onEdgeSelect,
  onCanvasClick,
  onNodeDrag,
  onNodeAdd,
  onEdgeAdd,
  onEdgeRemove,
  children,
  className,
  readOnly = false,
  showGrid = true,
  showMinimap = false,
  enablePan = true,
  enableZoom = true,
  enableConnect = true,
  enableDrag = true,
  animateTokenFlow = true,
  tokenFlowSpeed = 1,
}: CanvasProps) {
  const canvasRef = React.useRef<HTMLDivElement>(null);
  const svgRef = React.useRef<SVGSVGElement>(null);

  const [state, setState] = React.useState<CanvasState>(() => ({
    ...createDefaultState(),
    nodes: propNodes || [],
    edges: propEdges || [],
  }));

  const [viewMode, setViewMode] = React.useState<ViewMode>(propViewMode);
  const [viewport, setViewport] = React.useState<CanvasViewport>(propViewport || { x: 0, y: 0, zoom: 1 });
  const [internalZoomLevel, setInternalZoomLevel] = React.useState<CanvasZoomLevel>(zoomLevel);

  // Sync props
  React.useEffect(() => {
    if (propNodes) setState((s) => ({ ...s, nodes: propNodes }));
  }, [propNodes]);

  React.useEffect(() => {
    if (propEdges) setState((s) => ({ ...s, edges: propEdges }));
  }, [propEdges]);

  React.useEffect(() => {
    setViewMode(propViewMode);
  }, [propViewMode]);

  // Graph context value
  const contextValue: GraphContextType = React.useMemo(
    () => ({
      nodes: state.nodes,
      edges: state.edges,
      viewport,
      selectedNodes: state.selectedNodes,
      selectedEdges: state.selectedEdges,
      viewMode,
      setViewMode,
      selectNode: (id, append = false) => {
        setState((s) => ({
          ...s,
          selectedNodes: append ? [...s.selectedNodes, id] : [id],
          selectedEdges: [],
        }));
        const node = state.nodes.find((n) => n.id === id);
        if (node) onNodeSelect?.(node);
      },
      selectEdge: (id) => {
        setState((s) => ({ ...s, selectedEdges: [id], selectedNodes: [] }));
        const edge = state.edges.find((e) => e.id === id);
        if (edge) onEdgeSelect?.(edge);
      },
      deselectAll: () => {
        setState((s) => ({ ...s, selectedNodes: [], selectedEdges: [] }));
      },
      updateNode: (id, data) => {
        setState((s) => ({
          ...s,
          nodes: s.nodes.map((n) => (n.id === id ? { ...n, ...data } : n)),
        }));
      },
      updateEdge: (id, data) => {
        setState((s) => ({
          ...s,
          edges: s.edges.map((e) => (e.id === id ? { ...e, ...data } : e)),
        }));
      },
      addNode: (type, position) => {
        const newNode: any = {
          id: `node-${Date.now()}`,
          type,
          label: `${type} ${state.nodes.length + 1}`,
          position,
          status: "idle" as const,
        };
        setState((s) => ({ ...s, nodes: [...s.nodes, newNode] }));
        onNodeAdd?.(type, position);
        return newNode.id;
      },
      removeNode: (id) => {
        setState((s) => ({
          ...s,
          nodes: s.nodes.filter((n) => n.id !== id),
          edges: s.edges.filter((e) => e.source !== id && e.target !== id),
          selectedNodes: s.selectedNodes.filter((n) => n !== id),
        }));
      },
      connect: (from, fromPort, to, toPort) => {
        const newEdge: EdgeData = {
          id: `edge-${Date.now()}`,
          source: from,
          target: to,
          sourcePort: fromPort,
          targetPort: toPort,
          status: "idle",
        };
        setState((s) => ({ ...s, edges: [...s.edges, newEdge] }));
        onEdgeAdd?.(from, to);
      },
      disconnect: (edgeId) => {
        setState((s) => ({
          ...s,
          edges: s.edges.filter((e) => e.id !== edgeId),
          selectedEdges: s.selectedEdges.filter((e) => e !== edgeId),
        }));
        onEdgeRemove?.(edgeId);
      },
      setViewport,
      zoomIn: () => setViewport((v) => ({ ...v, zoom: Math.min(v.zoom * 1.2, 4) })),
      zoomOut: () => setViewport((v) => ({ ...v, zoom: Math.max(v.zoom / 1.2, 0.2) })),
      fitView: () => {
        const bounds = getCanvasBounds(state.nodes);
        if (canvasRef.current) {
          const canvasWidth = canvasRef.current.clientWidth;
          const canvasHeight = canvasRef.current.clientHeight;
          const scaleX = canvasWidth / bounds.width;
          const scaleY = canvasHeight / bounds.height;
          const zoom = Math.min(scaleX, scaleY, 1);
          setViewport({
            x: -(bounds.minX + bounds.width / 2) * zoom + canvasWidth / 2,
            y: -(bounds.minY + bounds.height / 2) * zoom + canvasHeight / 2,
            zoom,
          });
        }
      },
      resetView: () => setViewport({ x: 0, y: 0, zoom: 1 }),
      forkWorkflow: () => {
        const forked = JSON.parse(JSON.stringify(state));
        forked.nodes = forked.nodes.map((n: NodeData) => ({ ...n, id: `forked-${n.id}-${Date.now()}` }));
        forked.edges = forked.edges.map((e: EdgeData) => ({
          ...e,
          id: `forked-${e.id}-${Date.now()}`,
        }));
        return forked;
      },
    }),
    [state, viewport, viewMode, onNodeSelect, onEdgeSelect, onNodeAdd, onEdgeAdd, onEdgeRemove]
  );

  // Graph interaction hook for pan/zoom/drag
  const { handleWheel, handleMouseDown, handleMouseMove, handleMouseUp } = useGraphInteraction({
    enablePan,
    enableZoom,
    enableDrag,
    enableConnect,
    readOnly,
    viewport,
    onViewportChange: setViewport,
    onNodeDrag: (nodeId, pos) => {
      setState((s) => ({
        ...s,
        nodes: s.nodes.map((n) => (n.id === nodeId ? { ...n, position: pos } : n)),
      }));
      onNodeDrag?.(nodeId, pos);
    },
    onConnectStart: (nodeId, portId, type) => {
      setState((s) => ({ ...s, isConnecting: true, connectionStart: { nodeId, portId, type } }));
    },
    onConnectEnd: (nodeId, portId) => {
      setState((s) => {
        if (s.connectionStart) {
          // Determine from/to based on which port type started the connection
          const fromType = s.connectionStart.type as "input" | "output";
          const toType = fromType === "output" ? "input" : "output";
          const from = fromType === "output" ? s.connectionStart : { nodeId, portId };
          const to = fromType === "input" ? s.connectionStart : { nodeId, portId };
          if (fromType === "output") {
            s.edges.push({
              id: `edge-${Date.now()}`,
              source: from.nodeId,
              target: to.nodeId,
              sourcePort: from.portId,
              targetPort: to.portId,
              status: "idle",
            });
          }
        }
        return { ...s, isConnecting: false, connectionStart: undefined };
      });
    },
    onClick: (pos) => {
      onCanvasClick?.(pos);
    },
  });

  // Token flow animation
  const [tokenProgress, setTokenProgress] = React.useState(0);
  React.useEffect(() => {
    if (!animateTokenFlow) return;
    let frameId: number;
    const animate = () => {
      setTokenProgress((p) => (p + 0.005 * tokenFlowSpeed) % 1);
      frameId = requestAnimationFrame(animate);
    };
    frameId = requestAnimationFrame(animate);
    return () => cancelAnimationFrame(frameId);
  }, [animateTokenFlow, tokenFlowSpeed]);

  const viewModeConfig = VIEW_MODE_CONFIG[viewMode];

  return (
    <GraphContext.Provider value={contextValue}>
      <div
        ref={canvasRef}
        className={cn(
          "relative w-full h-full overflow-hidden select-none",
          "bg-[#08080f]",
          className
        )}
        style={{
          backgroundImage:
            showGrid
              ? `
            linear-gradient(${GRAPH_THEME.gridLine} 1px, transparent 1px),
            linear-gradient(90deg, ${GRAPH_THEME.gridLine} 1px, transparent 1px)
          `
              : "none",
          backgroundSize: `${64 * viewport.zoom}px ${64 * viewport.zoom}px`,
        }}
        onWheel={handleWheel}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseUp}
      >
        {/* Major grid overlay */}
        {showGrid && (
          <div
            className="absolute inset-0 pointer-events-none"
            style={{
              backgroundImage: `
              linear-gradient(${GRAPH_THEME.gridLineMajor} 1px, transparent 1px),
              linear-gradient(90deg, ${GRAPH_THEME.gridLineMajor} 1px, transparent 1px)
            `,
              backgroundSize: `${512 * viewport.zoom}px ${512 * viewport.zoom}px`,
              transform: `translate(${(viewport.x * viewport.zoom) % (512 * viewport.zoom)}px, ${(viewport.y * viewport.zoom) % (512 * viewport.zoom)}px)`,
            }}
          />
        )}

        {/* View mode indicator */}
        <div className="absolute top-4 left-4 z-20">
          <div
            className={cn(
              "flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-semibold",
              "bg-bg-glass backdrop-blur-md border border-border-subtle",
              "text-text-muted"
            )}
          >
            <span>{viewModeConfig.icon}</span>
            <span>{viewModeConfig.label}</span>
          </div>
        </div>

        {/* Zoom controls */}
        {enableZoom && (
          <div className="absolute bottom-4 left-4 z-20 flex flex-col gap-1">
            <button
              className="w-8 h-8 flex items-center justify-center rounded-lg bg-bg-glass backdrop-blur-md border border-border-subtle text-text-secondary hover:text-text-primary hover:bg-bg-hover transition-all"
              onClick={() =>
                setViewport((v) => ({ ...v, zoom: Math.min(v.zoom * 1.2, 4) }))
              }
            >
              +
            </button>
            <button
              className="w-8 h-8 flex items-center justify-center rounded-lg bg-bg-glass backdrop-blur-md border border-border-subtle text-text-secondary hover:text-text-primary hover:bg-bg-hover transition-all text-xs font-bold"
              onClick={() => setViewport({ x: 0, y: 0, zoom: 1 })}
              title="Reset zoom"
            >
              ⟳
            </button>
            <button
              className="w-8 h-8 flex items-center justify-center rounded-lg bg-bg-glass backdrop-blur-md border border-border-subtle text-text-secondary hover:text-text-primary hover:bg-bg-hover transition-all"
              onClick={() =>
                setViewport((v) => ({ ...v, zoom: Math.max(v.zoom / 1.2, 0.2) }))
              }
            >
              −
            </button>
          </div>
        )}

        {/* SVG Canvas */}
        <svg
          ref={svgRef}
          className="absolute inset-0 w-full h-full pointer-events-none"
          style={{
            transform: `translate(${viewport.x * viewport.zoom}px, ${viewport.y * viewport.zoom}px) scale(${viewport.zoom})`,
          }}
        >
          <defs>
            {/* Glow filter */}
            <filter id="node-glow" x="-50%" y="-50%" width="200%" height="200%">
              <feGaussianBlur in="SourceGraphic" stdDeviation="3" />
              <feFlood floodColor="#f97316" floodOpacity="0.3" />
              <feComposite in2="blur" operator="in" />
              <feMerge>
                <feMergeNode />
                <feMergeNode in="SourceGraphic" />
              </feMerge>
            </filter>
          </defs>

          {/* Edges */}
          {state.edges.map((edge) => (
            <ExecutionEdge
              key={edge.id}
              {...edge}
              animated={viewMode === "execute"}
              tokenFlow={animateTokenFlow && viewMode === "execute"}
              tokenPosition={tokenProgress}
              isSelected={state.selectedEdges.includes(edge.id)}
              onClick={() => contextValue.selectEdge(edge.id)}
              onSelect={() => contextValue.selectEdge(edge.id)}
            />
          ))}
        </svg>

        {/* Nodes */}
        <div
          className="absolute inset-0"
          style={{
            transform: `translate(${viewport.x * viewport.zoom}px, ${viewport.y * viewport.zoom}px) scale(${viewport.zoom})`,
          }}
        >
          {state.nodes.map((node) => (
            <ExecutionNode
              key={node.id}
              {...node}
              data-node-id={node.id}
              isSelected={state.selectedNodes.includes(node.id)}
              isHovered={state.hoveredNode === node.id}
              onSelect={(id) => {
                const found = state.nodes.find((n) => n.id === id);
                if (found) onNodeSelect?.(found);
                contextValue.selectNode(id);
              }}
              onDragStart={(_id, _pos) => setState((s) => ({ ...s, isPanning: false }))}
              onDrag={(id, pos) => contextValue.updateNode(id, { position: pos })}
              onDragEnd={(id, pos) => {
                const snapped = {
                  x: Math.round(pos.x / 20) * 20,
                  y: Math.round(pos.y / 20) * 20,
                };
                onNodeDrag?.(id, snapped);
              }}
              onPortMouseDown={(nodeId, portId, type) => {
                if (type === "output" && enableConnect) {
                  contextValue.connect(nodeId, portId, "temp-target", "temp-port");
                }
              }}
              onRemove={enableDrag && !readOnly ? (id) => contextValue.removeNode(id) : undefined}
            />
          ))}
        </div>

        {/* Connection line preview */}
        {state.isConnecting && state.connectionStart && (
          <ConnectionPreview
            startNodeId={state.connectionStart.nodeId}
            startPortId={state.connectionStart.portId}
            startType={state.connectionStart.type}
            nodes={state.nodes}
            viewport={viewport}
          />
        )}

        {/* Minimap */}
        {showMinimap && (
          <Minimap
            nodes={state.nodes}
            viewport={viewport}
            canvasWidth={canvasRef.current?.clientWidth || 800}
            canvasHeight={canvasRef.current?.clientHeight || 600}
            onNavigate={(v) => setViewport(v)}
          />
        )}

        {children}
      </div>
    </GraphContext.Provider>
  );
}

// Connection preview component for drawing temporary connection lines
interface ConnectionPreviewProps {
  startNodeId: string;
  startPortId: string;
  startType: "input" | "output";
  nodes: NodeData[];
  viewport: CanvasViewport;
}

function ConnectionPreview({ startNodeId, startPortId, startType, nodes, viewport }: ConnectionPreviewProps) {
  const [mousePos, setMousePos] = React.useState({ x: 0, y: 0 });
  const svgRef = React.useRef<SVGSVGElement>(null);

  const startNode = nodes.find((n) => n.id === startNodeId);
  const startPos = startNode?.position ?? { x: 0, y: 0 };

  React.useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      const svg = svgRef.current;
      if (!svg) return;
      const rect = svg.getBoundingClientRect();
      const x = (e.clientX - rect.left - viewport.x * viewport.zoom) / viewport.zoom;
      const y = (e.clientY - rect.top - viewport.y * viewport.zoom) / viewport.zoom;
      setMousePos({ x, y });
    };
    window.addEventListener("mousemove", handleMouseMove);
    return () => window.removeEventListener("mousemove", handleMouseMove);
  }, [viewport]);

  const sourceX = startPos.x + (startType === "output" ? 200 : 0);
  const sourceY = startPos.y + 60; // middle of node

  return (
    <svg ref={svgRef} className="absolute inset-0 pointer-events-none w-full h-full">
      <line
        x1={sourceX}
        y1={sourceY}
        x2={mousePos.x}
        y2={mousePos.y}
        stroke="#f97316"
        strokeWidth={2}
        strokeDasharray="5,5"
        strokeOpacity={0.6}
      />
    </svg>
  );
}

// Minimap sub-component
interface MinimapProps {
  nodes: NodeData[];
  viewport: CanvasViewport;
  canvasWidth: number;
  canvasHeight: number;
  onNavigate?: (viewport: CanvasViewport) => void;
}

function Minimap({ nodes, viewport, canvasWidth, canvasHeight, onNavigate }: MinimapProps) {
  const [bounds, setBounds] = React.useState({ x: 0, y: 0, w: 800, h: 600 });
  const minimapW = 200;
  const minimapH = 150;

  React.useEffect(() => {
    if (nodes.length > 0) {
      let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;
      for (const n of nodes) {
        if (n.position) {
          minX = Math.min(minX, n.position.x);
          minY = Math.min(minY, n.position.y);
          maxX = Math.max(maxX, n.position.x + 200);
          maxY = Math.max(maxY, n.position.y + 120);
        }
      }
      const pad = 100;
      setBounds({ x: minX - pad, y: minY - pad, w: maxX - minX + pad * 2, h: maxY - minY + pad * 2 });
    }
  }, [nodes]);

  const scaleX = minimapW / bounds.w;
  const scaleY = minimapH / bounds.h;

  const handleClick = (e: React.MouseEvent) => {
    const rect = (e.target as HTMLElement).getBoundingClientRect();
    const mx = e.clientX - rect.left;
    const my = e.clientY - rect.top;
    const canvasX = mx / scaleX + bounds.x;
    const canvasY = my / scaleY + bounds.y;
    const screenW = canvasWidth / viewport.zoom;
    const screenH = canvasHeight / viewport.zoom;
    onNavigate?.({
      x: -(canvasX - screenW / 2) * viewport.zoom + canvasWidth / 2,
      y: -(canvasY - screenH / 2) * viewport.zoom + canvasHeight / 2,
      zoom: viewport.zoom,
    });
  };

  return (
    <div
      className="absolute bottom-4 right-4 z-20 rounded-lg overflow-hidden"
      style={{
        width: minimapW,
        height: minimapH,
        background: GRAPH_THEME.minimapBg,
        border: `1px solid ${GRAPH_THEME.minimaperBorder}`,
      }}
    >
      <div onClick={handleClick} style={{ width: "100%", height: "100%", position: "relative" }}>
        {/* Grid */}
        <svg width="100%" height="100%" style={{ position: "absolute", inset: 0 }}>
          {nodes.map((node) =>
            node.position ? (
              <rect
                key={node.id}
                x={(node.position.x - bounds.x) * scaleX}
                y={(node.position.y - bounds.y) * scaleY}
                width={200 * scaleX}
                height={120 * scaleY}
                fill={GRAPH_THEME.nodeBg}
                stroke={GRAPH_THEME.nodeBorder}
                strokeWidth={0.5}
                rx={2}
              />
            ) : null
          )}
        </svg>
        {/* Viewport rectangle */}
        <div
          className="absolute border pointer-events-none"
          style={{
            left: ((-viewport.x / viewport.zoom - bounds.x) * scaleX),
            top: ((-viewport.y / viewport.zoom - bounds.y) * scaleY),
            width: (canvasWidth / viewport.zoom) * scaleX,
            height: (canvasHeight / viewport.zoom) * scaleY,
            borderColor: GRAPH_THEME.minimapViewport,
            backgroundColor: GRAPH_THEME.minimapViewport,
          }}
        />
      </div>
    </div>
  );
}