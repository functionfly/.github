import React, { useState, useCallback, useRef, useEffect } from 'react'
import { cn } from '@/lib/utils'

export type DockPosition = 'left' | 'right' | 'top' | 'bottom'

export interface DockPanel {
  id: string
  title: string
  content: React.ReactNode
  position: DockPosition
  isVisible: boolean
  isCollapsed: boolean
  size: number
  minSize?: number
  maxSize?: number
}

export interface DockLayoutProps {
  panels: DockPanel[]
  defaultPanelSizes?: Record<string, number>
  children?: React.ReactNode
  className?: string
  onPanelResize?: (panelId: string, size: number) => void
  onPanelToggle?: (panelId: string, isVisible: boolean) => void
}

/**
 * DockLayout - Dockable panel system with resizable dock areas
 * Supports multiple dock positions with smooth animations
 */
export function DockLayout({
  panels,
  defaultPanelSizes = {},
  children,
  className,
  onPanelResize,
  onPanelToggle,
}: DockLayoutProps) {
  const [panelSizes, setPanelSizes] = useState<Record<string, number>>({})
  const [dragState, setDragState] = useState<{
    isDragging: boolean
    panelId: string | null
    startX: number
    startSize: number
  }>({ isDragging: false, panelId: null, startX: 0, startSize: 0 })

  const containerRef = useRef<HTMLDivElement>(null)

  const handleMouseDown = useCallback((e: React.MouseEvent, panelId: string, position: DockPosition) => {
    e.preventDefault()
    const currentSize = panelSizes[panelId] ?? defaultPanelSizes[panelId] ?? 250
    setDragState({
      isDragging: true,
      panelId,
      startX: e.clientX,
      startSize: currentSize,
    })
  }, [panelSizes, defaultPanelSizes])

  const handleMouseMove = useCallback((e: MouseEvent) => {
    if (!dragState.isDragging || !dragState.panelId) return

    const delta = e.clientX - dragState.startX
    const newSize = Math.max(100, Math.min(800, dragState.startSize + delta))

    setPanelSizes(prev => ({
      ...prev,
      [dragState.panelId!]: newSize,
    }))
  }, [dragState])

  const handleMouseUp = useCallback(() => {
    if (dragState.isDragging && dragState.panelId) {
      onPanelResize?.(dragState.panelId, panelSizes[dragState.panelId])
    }
    setDragState({ isDragging: false, panelId: null, startX: 0, startSize: 0 })
  }, [dragState, panelSizes, onPanelResize])

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

  const visiblePanels = panels.filter(p => p.isVisible)
  const leftPanels = visiblePanels.filter(p => p.position === 'left')
  const rightPanels = visiblePanels.filter(p => p.position === 'right')
  const topPanels = visiblePanels.filter(p => p.position === 'top')
  const bottomPanels = visiblePanels.filter(p => p.position === 'bottom')

  const getPanelSize = (panel: DockPanel) => {
    return panelSizes[panel.id] ?? defaultPanelSizes[panel.id] ?? panel.size
  }

  return (
    <div ref={containerRef} className={cn('relative h-full w-full flex flex-col', className)}>
      {/* Top Panels */}
      {topPanels.length > 0 && (
        <div className="flex flex-col border-b border-aviation-border-panel">
          {topPanels.map(panel => (
            <div key={panel.id} className="p-4">
              {panel.content}
            </div>
          ))}
        </div>
      )}

      <div className="flex-1 flex flex-row overflow-hidden">
        {/* Left Panels */}
        {leftPanels.length > 0 && (
          <div className="flex flex-col border-r border-aviation-border-panel">
            {leftPanels.map(panel => (
              <div
                key={panel.id}
                className="relative"
                style={{ width: getPanelSize(panel) }}
              >
                {panel.content}
                <div
                  className="absolute top-0 right-0 w-1 h-full cursor-col-resize hover:bg-aviation-amber transition-colors"
                  onMouseDown={e => handleMouseDown(e, panel.id, 'left')}
                />
              </div>
            ))}
          </div>
        )}

        {/* Main Content */}
        <div className="flex-1 overflow-hidden">{children}</div>

        {/* Right Panels */}
        {rightPanels.length > 0 && (
          <div className="flex flex-col border-l border-aviation-border-panel">
            {rightPanels.map(panel => (
              <div
                key={panel.id}
                className="relative"
                style={{ width: getPanelSize(panel) }}
              >
                {panel.content}
                <div
                  className="absolute top-0 left-0 w-1 h-full cursor-col-resize hover:bg-aviation-amber transition-colors"
                  onMouseDown={e => handleMouseDown(e, panel.id, 'right')}
                />
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Bottom Panels */}
      {bottomPanels.length > 0 && (
        <div className="flex flex-col border-t border-aviation-border-panel">
          {bottomPanels.map(panel => (
            <div key={panel.id} className="p-4">
              {panel.content}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
