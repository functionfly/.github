import React, { useState, useCallback, useEffect } from 'react'
import { cn } from '@/lib/utils'
import { Maximize, Minimize, X, Copy, Save } from 'lucide-react'
import { Button } from '@/components/ui/button'

export interface WindowState {
  id: string
  title: string
  content: React.ReactNode
  x: number
  y: number
  width: number
  height: number
  isMaximized: boolean
  isMinimized: boolean
  zIndex: number
}

export interface WindowManagerProps {
  windows: WindowState[]
  className?: string
  onWindowClose?: (id: string) => void
  onWindowFocus?: (id: string) => void
  onWindowMinimize?: (id: string) => void
  onWindowMaximize?: (id: string) => void
}

/**
 * WindowManager - Manages multiple windows in the workspace
 * Handles focus, minimize, maximize, and close operations
 */
export function WindowManager({
  windows,
  className,
  onWindowClose,
  onWindowFocus,
  onWindowMinimize,
  onWindowMaximize,
}: WindowManagerProps) {
  const [windowsState, setWindowsState] = useState<Record<string, WindowState>>({})

  useEffect(() => {
    const stateMap = windows.reduce((acc, w) => {
      acc[w.id] = w
      return acc
    }, {} as Record<string, WindowState>)
    setWindowsState(stateMap)
  }, [windows])

  const bringToFront = useCallback((id: string) => {
    onWindowFocus?.(id)
  }, [onWindowFocus])

  const handleClose = useCallback((id: string) => {
    onWindowClose?.(id)
  }, [onWindowClose])

  const handleMinimize = useCallback((id: string) => {
    onWindowMinimize?.(id)
  }, [onWindowMinimize])

  const handleMaximize = useCallback((id: string) => {
    onWindowMaximize?.(id)
  }, [onWindowMaximize])

  const visibleWindows = windows.filter(w => !w.isMinimized).sort((a, b) => b.zIndex - a.zIndex)

  if (visibleWindows.length === 0) {
    return null
  }

  return (
    <div className={cn('absolute inset-0 pointer-events-none', className)}>
      {visibleWindows.map(window => (
        <WindowComponent
          key={window.id}
          window={window}
          onClose={handleClose}
          onFocus={bringToFront}
          onMinimize={handleMinimize}
          onMaximize={handleMaximize}
        />
      ))}
    </div>
  )
}

interface WindowComponentProps {
  window: WindowState
  onClose: (id: string) => void
  onFocus: (id: string) => void
  onMinimize: (id: string) => void
  onMaximize: (id: string) => void
}

function WindowComponent({
  window,
  onClose,
  onFocus,
  onMinimize,
  onMaximize,
}: WindowComponentProps) {
  return (
    <div
      className={cn(
        'absolute aviation-panel rounded-lg overflow-hidden pointer-events-auto shadow-xl',
        window.isMaximized && 'inset-4 w-auto h-auto'
      )}
      style={{
        left: window.isMaximized ? undefined : window.x,
        top: window.isMaximized ? undefined : window.y,
        width: window.isMaximized ? undefined : window.width,
        height: window.isMaximized ? undefined : window.height,
        zIndex: window.zIndex,
      }}
      onClick={() => onFocus(window.id)}
    >
      {/* Title Bar */}
      <div className="flex items-center justify-between px-3 py-2 bg-aviation-bg-instrument border-b border-aviation-border-panel cursor-move">
        <span className="text-sm font-medium">{window.title}</span>
        <div className="flex items-center gap-1">
          <Button variant="ghost" size="icon" className="w-6 h-6" onClick={() => onMinimize(window.id)}>
            <Minimize className="w-3 h-3" />
          </Button>
          <Button variant="ghost" size="icon" className="w-6 h-6" onClick={() => onMaximize(window.id)}>
            {window.isMaximized ? <Minimize className="w-3 h-3" /> : <Maximize className="w-3 h-3" />}
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="w-6 h-6 hover:text-aviation-red"
            onClick={() => onClose(window.id)}
          >
            <X className="w-3 h-3" />
          </Button>
        </div>
      </div>

      {/* Content */}
      <div className="p-3 overflow-auto" style={{ height: 'calc(100% - 40px)' }}>
        {window.content}
      </div>
    </div>
  )
}
