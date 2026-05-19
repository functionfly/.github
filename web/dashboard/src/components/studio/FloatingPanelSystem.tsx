import React, { useState, useCallback, useRef, useEffect } from 'react'
import { cn } from '@/lib/utils'
import { X, Maximize2, Minimize2, Move } from 'lucide-react'
import { Button } from '@/components/ui/button'

export interface FloatingPanel {
  id: string
  title: string
  content: React.ReactNode
  x: number
  y: number
  width: number
  height: number
  zIndex: number
  isMaximized?: boolean
  isMinimized?: boolean
}

export interface FloatingPanelSystemProps {
  panels: FloatingPanel[]
  children?: React.ReactNode
  className?: string
  onClosePanel?: (panelId: string) => void
  onMovePanel?: (panelId: string, x: number, y: number) => void
  onResizePanel?: (panelId: string, width: number, height: number) => void
}

/**
 * FloatingPanelSystem - Manages floating/draggable panels
 * Supports drag, resize, maximize, and minimize operations
 */
export function FloatingPanelSystem({
  panels,
  children,
  className,
  onClosePanel,
  onMovePanel,
  onResizePanel,
}: FloatingPanelSystemProps) {
  const [activePanelId, setActivePanelId] = useState<string | null>(null)
  const [dragState, setDragState] = useState<{
    isDragging: boolean
    panelId: string | null
    offsetX: number
    offsetY: number
  }>({ isDragging: false, panelId: null, offsetX: 0, offsetY: 0 })

  const bringToFront = useCallback((panelId: string) => {
    setActivePanelId(panelId)
  }, [])

  const handleMouseDown = useCallback((e: React.MouseEvent, panelId: string) => {
    e.stopPropagation()
    const panel = panels.find(p => p.id === panelId)
    if (!panel || panel.isMaximized) return

    const rect = (e.target as HTMLElement).getBoundingClientRect()
    setDragState({
      isDragging: true,
      panelId,
      offsetX: e.clientX - rect.left,
      offsetY: e.clientY - rect.top,
    })
    bringToFront(panelId)
  }, [panels, bringToFront])

  const handleMouseMove = useCallback((e: MouseEvent) => {
    if (!dragState.isDragging || !dragState.panelId) return
    onMovePanel?.(dragState.panelId, e.clientX - dragState.offsetX, e.clientY - dragState.offsetY)
  }, [dragState, onMovePanel])

  const handleMouseUp = useCallback(() => {
    setDragState({ isDragging: false, panelId: null, offsetX: 0, offsetY: 0 })
  }, [])

  useEffect(() => {
    if (dragState.isDragging) {
      document.addEventListener('mousemove', handleMouseMove)
      document.addEventListener('mouseup', handleMouseUp)
      return () => {
        document.removeEventListener('mousemove', handleMouseMove)
        document.removeEventListener('mouseup', handleMouseUp)
      }
    }
  }, [dragState, handleMouseMove, handleMouseUp])

  if (panels.length === 0) {
    return <>{children}</>
  }

  return (
    <div className={cn('relative w-full h-full', className)}>
      {/* Main Content Layer */}
      <div className="w-full h-full">{children}</div>

      {/* Floating Panels Layer */}
      <div className="absolute inset-0 pointer-events-none">
        {panels.map(panel => (
          <FloatingPanel
            key={panel.id}
            panel={panel}
            isActive={activePanelId === panel.id}
            onMouseDown={handleMouseDown}
            onClose={() => onClosePanel?.(panel.id)}
            onActivate={() => bringToFront(panel.id)}
          />
        ))}
      </div>
    </div>
  )
}

interface FloatingPanelProps {
  panel: FloatingPanel
  isActive: boolean
  onMouseDown: (e: React.MouseEvent, panelId: string) => void
  onClose: () => void
  onActivate: () => void
}

function FloatingPanel({ panel, isActive, onMouseDown, onClose, onActivate }: FloatingPanelProps) {
  const [isMaximized, setIsMaximized] = useState(panel.isMaximized ?? false)
  const [isMinimized, setIsMinimized] = useState(panel.isMinimized ?? false)

  const handleToggleMaximize = useCallback(() => {
    setIsMaximized(max => !max)
  }, [])

  const handleToggleMinimize = useCallback(() => {
    setIsMinimized(min => !min)
  }, [])

  if (isMinimized) {
    return (
      <div
        className="absolute bottom-4 right-4 pointer-events-auto"
        style={{ zIndex: panel.zIndex }}
        onClick={onActivate}
      >
        <Button
          variant="outline"
          size="sm"
          className="aviation-panel gap-2"
        >
          <Move className="w-3 h-3" />
          {panel.title}
        </Button>
      </div>
    )
  }

  return (
    <div
      className={cn(
        'absolute aviation-panel overflow-hidden pointer-events-auto',
        isActive && 'ring-2 ring-aviation-amber'
      )}
      style={{
        left: isMaximized ? 0 : panel.x,
        top: isMaximized ? 0 : panel.y,
        width: isMaximized ? 'calc(100% - 16px)' : panel.width,
        height: isMaximized ? 'calc(100% - 16px)' : panel.height,
        zIndex: panel.zIndex,
      }}
      onClick={onActivate}
    >
      {/* Header */}
      <div
        className="flex items-center justify-between px-3 py-2 border-b border-aviation-border-panel cursor-move bg-aviation-bg-instrument"
        onMouseDown={e => onMouseDown(e, panel.id)}
      >
        <span className="text-sm font-medium text-aviation-text-primary">{panel.title}</span>
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon"
            className="w-6 h-6"
            onClick={handleToggleMinimize}
          >
            <Minimize2 className="w-3 h-3" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="w-6 h-6"
            onClick={handleToggleMaximize}
          >
            {isMaximized ? <Minimize2 className="w-3 h-3" /> : <Maximize2 className="w-3 h-3" />}
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="w-6 h-6 hover:text-aviation-red"
            onClick={onClose}
          >
            <X className="w-3 h-3" />
          </Button>
        </div>
      </div>

      {/* Content */}
      <div className="p-3 overflow-auto h-full">{panel.content}</div>
    </div>
  )
}
