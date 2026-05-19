import React, { useState, useRef, useCallback, useEffect } from 'react'
import { cn } from '@/lib/utils'
import { ZoomIn, ZoomOut, RotateCw } from 'lucide-react'
import { Button } from '@/components/ui/button'

export interface WorkspaceViewportProps {
  children?: React.ReactNode
  className?: string
  initialZoom?: number
  minZoom?: number
  maxZoom?: number
  showControls?: boolean
  onZoomChange?: (zoom: number) => void
}

/**
 * WorkspaceViewport - Main viewport container with zoom/pan support
 * Provides canvas-like viewing experience
 */
export function WorkspaceViewport({
  children,
  className,
  initialZoom = 1,
  minZoom = 0.1,
  maxZoom = 5,
  showControls = true,
  onZoomChange,
}: WorkspaceViewportProps) {
  const [zoom, setZoom] = useState(initialZoom)
  const [pan, setPan] = useState({ x: 0, y: 0 })
  const [isDragging, setIsDragging] = useState(false)
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 })
  
  const containerRef = useRef<HTMLDivElement>(null)

  const handleZoomIn = useCallback(() => {
    const newZoom = Math.min(maxZoom, zoom + 0.1)
    setZoom(newZoom)
    onZoomChange?.(newZoom)
  }, [zoom, maxZoom, onZoomChange])

  const handleZoomOut = useCallback(() => {
    const newZoom = Math.max(minZoom, zoom - 0.1)
    setZoom(newZoom)
    onZoomChange?.(newZoom)
  }, [zoom, minZoom, onZoomChange])

  const handleReset = useCallback(() => {
    setZoom(1)
    setPan({ x: 0, y: 0 })
    onZoomChange?.(1)
  }, [onZoomChange])

  const handleWheel = useCallback((e: React.WheelEvent) => {
    if (e.ctrlKey || e.metaKey) {
      e.preventDefault()
      const delta = e.deltaY > 0 ? -0.1 : 0.1
      const newZoom = Math.min(maxZoom, Math.max(minZoom, zoom + delta))
      setZoom(newZoom)
      onZoomChange?.(newZoom)
    }
  }, [zoom, minZoom, maxZoom, onZoomChange])

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    if (e.button === 0 && !e.ctrlKey) {
      setIsDragging(true)
      setDragStart({ x: e.clientX - pan.x, y: e.clientY - pan.y })
    }
  }, [pan])

  const handleMouseMove = useCallback((e: MouseEvent) => {
    if (isDragging) {
      setPan({
        x: e.clientX - dragStart.x,
        y: e.clientY - dragStart.y,
      })
    }
  }, [isDragging, dragStart])

  const handleMouseUp = useCallback(() => {
    setIsDragging(false)
  }, [])

  useEffect(() => {
    if (isDragging) {
      document.addEventListener('mousemove', handleMouseMove)
      document.addEventListener('mouseup', handleMouseUp)
      return () => {
        document.removeEventListener('mousemove', handleMouseMove)
        document.removeEventListener('mouseup', handleMouseUp)
      }
    }
  }, [isDragging, handleMouseMove, handleMouseUp])

  const transformStyle = {
    transform: `translate(${pan.x}px, ${pan.y}px) scale(${zoom})`,
    transformOrigin: '0 0',
  }

  return (
    <div
      ref={containerRef}
      className={cn('relative w-full h-full overflow-hidden', className)}
      onWheel={handleWheel}
      onMouseDown={handleMouseDown}
    >
      {/* Controls */}
      {showControls && (
        <div className="absolute top-4 right-4 z-10 flex items-center gap-1 aviation-panel p-1 rounded-lg">
          <Button variant="ghost" size="icon" className="w-8 h-8" onClick={handleZoomOut}>
            <ZoomOut className="w-4 h-4" />
          </Button>
          <span className="text-xs font-mono text-aviation-text-secondary w-12 text-center">
            {Math.round(zoom * 100)}%
          </span>
          <Button variant="ghost" size="icon" className="w-8 h-8" onClick={handleZoomIn}>
            <ZoomIn className="w-4 h-4" />
          </Button>
          <div className="w-px h-4 bg-aviation-border-panel" />
          <Button variant="ghost" size="icon" className="w-8 h-8" onClick={handleReset}>
            <RotateCw className="w-4 h-4" />
          </Button>
        </div>
      )}

      {/* Viewport Content */}
      <div className="absolute inset-0" style={transformStyle}>
        {children}
      </div>
    </div>
  )
}
