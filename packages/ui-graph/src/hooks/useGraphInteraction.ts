/**
 * @functionfly/ui-graph
 * Graph interaction hook for pan/zoom/drag/connect
 */

import * as React from "react";
import type { CanvasViewport } from "../types";

export interface UseGraphInteractionOptions {
  enablePan: boolean;
  enableZoom: boolean;
  enableDrag: boolean;
  enableConnect: boolean;
  readOnly: boolean;
  viewport: CanvasViewport;
  onViewportChange: (v: CanvasViewport) => void;
  onNodeDrag?: (nodeId: string, pos: { x: number; y: number }) => void;
  onConnectStart?: (nodeId: string, portId: string, type: "input" | "output") => void;
  onConnectEnd?: (nodeId: string, portId: string) => void;
  onClick?: (pos: { x: number; y: number }) => void;
}

export function useGraphInteraction(opts: UseGraphInteractionOptions) {
  const {
    enablePan,
    enableZoom,
    enableDrag,
    readOnly,
    viewport,
    onViewportChange,
    onNodeDrag,
    onClick,
  } = opts;

  const isPanning = React.useRef(false);
  const panStart = React.useRef({ x: 0, y: 0, vpX: 0, vpY: 0 });

  const handleWheel = React.useCallback(
    (e: React.WheelEvent) => {
      if (!enableZoom || readOnly) return;
      e.preventDefault();
      const delta = e.deltaY > 0 ? 0.9 : 1.1;
      onViewportChange({
        ...viewport,
        zoom: Math.min(Math.max(viewport.zoom * delta, 0.2), 4),
      });
    },
    [enableZoom, readOnly, viewport, onViewportChange]
  );

  const handleMouseDown = React.useCallback(
    (e: React.MouseEvent) => {
      // Middle mouse or space+click for pan, or left click when enablePan is true
      if (e.button === 1 || enablePan || (e.button === 0 && enablePan)) {
        if (readOnly) return;
        isPanning.current = true;
        panStart.current = {
          x: e.clientX,
          y: e.clientY,
          vpX: viewport.x,
          vpY: viewport.y,
        };
        e.preventDefault();
      }
    },
    [readOnly, enablePan, viewport]
  );

  const handleMouseMove = React.useCallback(
    (e: MouseEvent) => {
      if (isPanning.current && enablePan) {
        onViewportChange({
          ...viewport,
          x: panStart.current.vpX + (e.clientX - panStart.current.x),
          y: panStart.current.vpY + (e.clientY - panStart.current.y),
        });
      }
    },
    [enablePan, viewport, onViewportChange]
  );

  const handleMouseUp = React.useCallback(() => {
    if (isPanning.current) {
      isPanning.current = false;
    }
  }, []);

  // Attach global mouse move/up for panning
  React.useEffect(() => {
    if (!enablePan) return;

    const onMove = (e: MouseEvent) => handleMouseMove(e);
    const onUp = () => handleMouseUp();

    window.addEventListener("mousemove", onMove);
    window.addEventListener("mouseup", onUp);

    return () => {
      window.removeEventListener("mousemove", onMove);
      window.removeEventListener("mouseup", onUp);
    };
  }, [enablePan, handleMouseMove, handleMouseUp]);

  return {
    handleWheel,
    handleMouseDown,
    handleMouseMove: () => {},
    handleMouseUp,
  };
}

// Snap to grid helper
export function snapToGrid(value: number, gridSize: number = 20): number {
  return Math.round(value / gridSize) * gridSize;
}

// Calculate distance between two points
export function distance(
  p1: { x: number; y: number },
  p2: { x: number; y: number }
): number {
  return Math.sqrt(Math.pow(p2.x - p1.x, 2) + Math.pow(p2.y - p1.y, 2));
}

// Get center point between two nodes
export function getNodeCenter(
  node: { position?: { x: number; y: number }; width?: number; height?: number }
): { x: number; y: number } {
  const pos = node.position ?? { x: 0, y: 0 };
  const w = node.width ?? 180;
  const h = node.height ?? 80;
  return {
    x: pos.x + w / 2,
    y: pos.y + h / 2,
  };
}

// Check if a point is inside a node bounds
export function isPointInNode(
  point: { x: number; y: number },
  node: { position?: { x: number; y: number }; width?: number; height?: number }
): boolean {
  const pos = node.position ?? { x: 0, y: 0 };
  const w = node.width ?? 180;
  const h = node.height ?? 80;
  return (
    point.x >= pos.x &&
    point.x <= pos.x + w &&
    point.y >= pos.y &&
    point.y <= pos.y + h
  );
}