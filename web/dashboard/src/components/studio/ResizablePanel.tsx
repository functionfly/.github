import React, { useState, useCallback, useRef, useEffect } from 'react'
import { cn } from '@/lib/utils'
import { GripVertical } from 'lucide-react'

export interface ResizablePanelProps {
  children?: React.ReactNode
  className?: string
  defaultSize?: number
  minSize?: number
  maxSize?: number
  direction?: 'horizontal' | 'vertical'
  onResize?: (size: number) => void
  showHandle?: boolean
}

/**
 * ResizablePanel - A resizable panel component with drag handle
 * Supports both horizontal and vertical resizing
 */
export function ResizablePanel({
  children,
  className,
  defaultSize = 300,
  minSize = 100,
  maxSize = 800,
  direction = 'horizontal',
  onResize,
  showHandle = true,
}: ResizablePanelProps) {
  const [size, setSize] = useState(defaultSize)
  const [isDragging, setIsDragging] = useState(false)
  const [startPos, setStartPos] = useState(0)
  const [startSize, setStartSize] = useState(0)
  
  const panelRef = useRef<HTMLDivElement>(null)

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault()
    setIsDragging(true)
    setStartPos(direction === 'horizontal' ? e.clientX : e.clientY)
    setStartSize(size)
  }, [direction, size])

  const handleMouseMove = useCallback((e: MouseEvent) => {
    if (!isDragging) return

    const currentPos = direction === 'horizontal' ? e.clientX : e.clientY
    const delta = currentPos - startPos
    const newSize = Math.max(minSize, Math.min(maxSize, startSize + delta))

    setSize(newSize)
    onResize?.(newSize)
  }, [isDragging, startPos, startSize, direction, minSize, maxSize, onResize])

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

  const style = direction === 'horizontal'
    ? { width: size }
    : { height: size }

  return (
    <div
      ref={panelRef}
      className={cn(
        'relative overflow-hidden',
        direction === 'horizontal' ? 'h-full' : 'w-full',
        className
      )}
      style={style}
    >
      {children}

      {/* Resize Handle */}
      {showHandle && (
        <div
          className={cn(
            'absolute bg-transparent hover:bg-aviation-amber/20 transition-colors cursor-resize',
            direction === 'horizontal'
              ? 'top-0 right-0 w-1 h-full cursor-col-resize'
              : 'bottom-0 left-0 h-1 w-full cursor-row-resize'
          )}
          onMouseDown={handleMouseDown}
        >
          <GripVertical
            className={cn(
              'absolute text-aviation-text-muted opacity-0 hover:opacity-100 transition-opacity',
              direction === 'horizontal'
                ? 'top-1/2 right-1 -translate-y-1/2 rotate-90'
                : 'left-1/2 bottom-1 -translate-x-1/2'
            )}
          />
        </div>
      )}
    </div>
  )
}
