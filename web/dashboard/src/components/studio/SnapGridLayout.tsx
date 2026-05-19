import React, { useState, useCallback, useRef, useEffect } from 'react'
import { cn } from '@/lib/utils'

export interface GridItem {
  id: string
  x: number
  y: number
  w: number
  h: number
  content: React.ReactNode
  isDraggable?: boolean
  isResizable?: boolean
}

export interface SnapGridLayoutProps {
  items: GridItem[]
  className?: string
  gridSize?: number
  gap?: number
  onItemMove?: (id: string, x: number, y: number) => void
  onItemResize?: (id: string, w: number, h: number) => void
}

/**
 * SnapGridLayout - Grid layout with snap-to-grid functionality
 * Items automatically align to grid and snap during drag/resize
 */
export function SnapGridLayout({
  items,
  className,
  gridSize = 20,
  gap = 8,
  onItemMove,
  onItemResize,
}: SnapGridLayoutProps) {
  const [dragState, setDragState] = useState<{
    id: string | null
    offsetX: number
    offsetY: number
  }>({ id: null, offsetX: 0, offsetY: 0 })
  const containerRef = useRef<HTMLDivElement>(null)

  const snapToGrid = useCallback((value: number) => {
    return Math.round(value / gridSize) * gridSize
  }, [gridSize])

  const handleItemMouseDown = useCallback((e: React.MouseEvent, id: string) => {
    const item = items.find(i => i.id === id)
    if (!item?.isDraggable) return

    e.preventDefault()
    e.stopPropagation()
    
    const rect = (e.target as HTMLElement).getBoundingClientRect()
    setDragState({
      id,
      offsetX: e.clientX - rect.left,
      offsetY: e.clientY - rect.top,
    })
  }, [items])

  const handleMouseMove = useCallback((e: MouseEvent) => {
    if (!dragState.id) return

    const containerRect = containerRef.current?.getBoundingClientRect()
    if (!containerRect) return

    const x = snapToGrid(e.clientX - containerRect.left - dragState.offsetX)
    const y = snapToGrid(e.clientY - containerRect.top - dragState.offsetY)

    onItemMove?.(dragState.id, x, y)
  }, [dragState, snapToGrid, onItemMove])

  const handleMouseUp = useCallback(() => {
    setDragState({ id: null, offsetX: 0, offsetY: 0 })
  }, [])

  useEffect(() => {
    if (dragState.id) {
      document.addEventListener('mousemove', handleMouseMove)
      document.addEventListener('mouseup', handleMouseUp)
      return () => {
        document.removeEventListener('mousemove', handleMouseMove)
        document.removeEventListener('mouseup', handleMouseUp)
      }
    }
  }, [dragState, handleMouseMove, handleMouseUp])

  return (
    <div
      ref={containerRef}
      className={cn('relative w-full h-full', className)}
      style={{
        backgroundImage: `
          linear-gradient(rgba(255,255,255,0.02) 1px, transparent 1px),
          linear-gradient(90deg, rgba(255,255,255,0.02) 1px, transparent 1px)
        `,
        backgroundSize: `${gridSize}px ${gridSize}px`,
      }}
    >
      {items.map(item => (
        <div
          key={item.id}
          className={cn(
            'absolute aviation-panel overflow-hidden',
            item.isDraggable && 'cursor-move'
          )}
          style={{
            left: item.x,
            top: item.y,
            width: item.w,
            height: item.h,
          }}
          onMouseDown={e => handleItemMouseDown(e, item.id)}
        >
          {item.content}
        </div>
      ))}
    </div>
  )
}
