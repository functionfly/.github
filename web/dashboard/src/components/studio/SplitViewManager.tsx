import React, { useState, useCallback, useRef, useEffect } from 'react'
import { cn } from '@/lib/utils'
import { GripVertical } from 'lucide-react'

export interface SplitView {
  id: string
  content: React.ReactNode
  size?: number
}

export interface SplitViewManagerProps {
  views: SplitView[]
  className?: string
  direction?: 'horizontal' | 'vertical'
  defaultSizes?: number[]
}

/**
 * SplitViewManager - Split view management with draggable dividers
 * Supports both horizontal and vertical splits
 */
export function SplitViewManager({
  views,
  className,
  direction = 'horizontal',
  defaultSizes,
}: SplitViewManagerProps) {
  const viewCount = views.length
  const [sizes, setSizes] = useState<number[]>(() => {
    if (defaultSizes?.length === viewCount) {
      return defaultSizes
    }
    const equalSize = 100 / viewCount
    return Array(viewCount).fill(equalSize)
  })

  const [dragIndex, setDragIndex] = useState<number | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)

  const handleMouseDown = useCallback((index: number) => {
    setDragIndex(index)
  }, [])

  const handleMouseMove = useCallback((e: MouseEvent) => {
    if (dragIndex === null || !containerRef.current) return

    const rect = containerRef.current.getBoundingClientRect()
    const totalSize = direction === 'horizontal' ? rect.width : rect.height
    const position = direction === 'horizontal' ? e.clientX - rect.left : e.clientY - rect.top

    const newSizes = [...sizes]
    const dragPosition = (position / totalSize) * 100

    // Update the two views around the divider
    const leftPercent = Math.max(5, Math.min(95, dragPosition))
    const rightPercent = 100 - leftPercent

    if (dragIndex > 0) {
      newSizes[dragIndex - 1] = leftPercent
    }
    if (dragIndex < viewCount - 1) {
      newSizes[dragIndex] = rightPercent
    }

    setSizes(newSizes)
  }, [dragIndex, sizes, direction, viewCount])

  const handleMouseUp = useCallback(() => {
    setDragIndex(null)
  }, [])

  useEffect(() => {
    if (dragIndex !== null) {
      document.addEventListener('mousemove', handleMouseMove)
      document.addEventListener('mouseup', handleMouseUp)
      return () => {
        document.removeEventListener('mousemove', handleMouseMove)
        document.removeEventListener('mouseup', handleMouseUp)
      }
    }
  }, [dragIndex, handleMouseMove, handleMouseUp])

  const containerStyle = {
    display: 'flex',
    flexDirection: direction === 'horizontal' ? 'row' as const : 'column' as const,
    width: '100%',
    height: '100%',
  }

  const viewStyle = (size: number) => ({
    [direction === 'horizontal' ? 'width' : 'height']: `${size}%`,
    overflow: 'hidden',
  })

  return (
    <div ref={containerRef} style={containerStyle} className={className}>
      {views.map((view, index) => (
        <React.Fragment key={view.id}>
          <div style={viewStyle(sizes[index])}>{view.content}</div>
          
          {/* Divider */}
          {index < viewCount - 1 && (
            <div
              className={cn(
                'bg-aviation-border-panel hover:bg-aviation-amber transition-colors',
                direction === 'horizontal'
                  ? 'w-1 cursor-col-resize'
                  : 'h-1 cursor-row-resize'
              )}
              onMouseDown={() => handleMouseDown(index)}
            >
              <GripVertical
                className={cn(
                  'text-aviation-text-muted opacity-0 hover:opacity-100 transition-opacity',
                  direction === 'horizontal' ? 'rotate-90 mx-auto' : 'mx-auto'
                )}
              />
            </div>
          )}
        </React.Fragment>
      ))}
    </div>
  )
}
