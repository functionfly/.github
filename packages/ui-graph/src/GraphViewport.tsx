/**
 * @functionfly/ui-graph
 * GraphViewport - viewport controls and navigation wrapper
 */

import * as React from "react";
import { cn } from "./utils";
import type { GraphViewportProps, CanvasViewport } from "./types";

const ZOOM_LEVELS = [0.2, 0.5, 0.75, 1, 1.25, 1.5, 2, 3, 4];

export function GraphViewport({
  viewport,
  zoomRange = { min: 0.2, max: 4 },
  showZoomIndicator = true,
  showGrid = true,
  gridSize = 64,
  onViewportChange,
  onZoomIn,
  onZoomOut,
  onFitView,
  onResetView,
  children,
  className,
}: GraphViewportProps) {
  const currentZoomPercent = Math.round(viewport.zoom * 100);

  return (
    <div className={cn("relative w-full h-full overflow-hidden bg-[#08080f]", className)}>
      {/* Grid background */}
      {showGrid && (
        <div
          className="absolute inset-0 pointer-events-none"
          style={{
            backgroundImage: `
              linear-gradient(rgba(255,255,255,0.03) 1px, transparent 1px),
              linear-gradient(90deg, rgba(255,255,255,0.03) 1px, transparent 1px)
            `,
            backgroundSize: `${gridSize * viewport.zoom}px ${gridSize * viewport.zoom}px`,
            backgroundPosition: `${viewport.x * viewport.zoom}px ${viewport.y * viewport.zoom}px`,
          }}
        />
      )}

      {/* Content layer */}
      <div
        className="absolute inset-0"
        style={{
          transform: `translate(${viewport.x * viewport.zoom}px, ${viewport.y * viewport.zoom}px) scale(${viewport.zoom})`,
        }}
      >
        {children}
      </div>

      {/* Controls overlay */}
      <div className="absolute top-4 left-4 z-20 flex flex-col gap-2">
        {/* Zoom controls */}
        <div className="flex flex-col gap-1 bg-[rgba(13,13,20,0.85)] backdrop-blur-sm rounded-lg border border-[rgba(255,255,255,0.08)] p-1">
          <button
            onClick={onZoomIn}
            className="size-8 flex items-center justify-center rounded-md text-[#9ca3af] hover:text-[#e8e8f0] hover:bg-[rgba(255,255,255,0.06)] transition-colors"
            title="Zoom in"
          >
            <span className="text-lg">+</span>
          </button>
          <button
            onClick={onResetView}
            className="size-8 flex items-center justify-center rounded-md text-[10px] text-[#6b7280] hover:text-[#e8e8f0] hover:bg-[rgba(255,255,255,0.06)] transition-colors font-mono"
            title="Reset zoom (100%)"
          >
            {currentZoomPercent}%
          </button>
          <button
            onClick={onZoomOut}
            className="size-8 flex items-center justify-center rounded-md text-[#9ca3af] hover:text-[#e8e8f0] hover:bg-[rgba(255,255,255,0.06)] transition-colors"
            title="Zoom out"
          >
            <span className="text-lg">−</span>
          </button>
        </div>

        {/* View controls */}
        <div className="flex flex-col gap-1 bg-[rgba(13,13,20,0.85)] backdrop-blur-sm rounded-lg border border-[rgba(255,255,255,0.08)] p-1">
          <button
            onClick={onFitView}
            className="size-8 flex items-center justify-center rounded-md text-[#9ca3af] hover:text-[#e8e8f0] hover:bg-[rgba(255,255,255,0.06)] transition-colors"
            title="Fit to view"
          >
            <span className="text-sm">⊡</span>
          </button>
          <button
            onClick={onResetView}
            className="size-8 flex items-center justify-center rounded-md text-[#9ca3af] hover:text-[#e8e8f0] hover:bg-[rgba(255,255,255,0.06)] transition-colors"
            title="Reset view"
          >
            <span className="text-sm">↺</span>
          </button>
        </div>
      </div>

      {/* Zoom indicator */}
      {showZoomIndicator && (
        <div className="absolute bottom-4 right-4 px-2 py-1 rounded bg-[rgba(13,13,20,0.85)] backdrop-blur-sm border border-[rgba(255,255,255,0.08)] text-[10px] text-[#6b7280] font-mono">
          {currentZoomPercent}%
        </div>
      )}
    </div>
  );
}
