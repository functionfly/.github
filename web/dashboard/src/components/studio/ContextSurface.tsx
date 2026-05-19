import React, { useState, useEffect, useCallback } from 'react'
import { cn } from '@/lib/utils'
import { X, Pin, PinOff } from 'lucide-react'
import { Button } from '@/components/ui/button'

export interface ContextSurfaceProps {
  contextKey?: string
  children?: React.ReactNode
  className?: string
  defaultPinned?: boolean
}

type ContextPanel = {
  id: string
  title: string
  content: React.ReactNode
  isVisible: boolean
  isPinned: boolean
}

/**
 * ContextSurface - Contextual surface panel that appears based on context
 * Smartly shows relevant information based on current context
 */
export function ContextSurface({
  contextKey = 'default',
  children,
  className,
  defaultPinned = false,
}: ContextSurfaceProps) {
  const [panels, setPanels] = useState<ContextPanel[]>([])
  const [isVisible, setIsVisible] = useState(false)

  // Determine visibility based on context
  useEffect(() => {
    // Show for specific contexts
    const shouldShow = ['studio', 'editor', 'debug'].includes(contextKey)
    setIsVisible(shouldShow)
  }, [contextKey])

  const addPanel = useCallback((panel: ContextPanel) => {
    setPanels(prev => [...prev, panel])
  }, [])

  const removePanel = useCallback((id: string) => {
    setPanels(prev => prev.filter(p => p.id !== id))
  }, [])

  const togglePin = useCallback((id: string) => {
    setPanels(prev => prev.map(p =>
      p.id === id ? { ...p, isPinned: !p.isPinned } : p
    ))
  }, [])

  if (!isVisible && panels.length === 0) {
    return <>{children}</>
  }

  return (
    <div className={cn('w-full h-full flex', className)}>
      {/* Main Content */}
      <div className="flex-1">{children}</div>

      {/* Context Panels */}
      {panels.length > 0 && (
        <div className="w-80 border-l border-aviation-border-panel overflow-y-auto">
          {panels.map(panel => (
            <div key={panel.id} className="border-b border-aviation-border-panel">
              <div className="flex items-center justify-between p-3 bg-aviation-bg-instrument">
                <span className="text-sm font-medium">{panel.title}</span>
                <div className="flex items-center gap-1">
                  <Button
                    variant="ghost"
                    size="icon"
                    className="w-6 h-6"
                    onClick={() => togglePin(panel.id)}
                  >
                    {panel.isPinned ? (
                      <PinOff className="w-3 h-3" />
                    ) : (
                      <Pin className="w-3 h-3" />
                    )}
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="w-6 h-6"
                    onClick={() => removePanel(panel.id)}
                  >
                    <X className="w-3 h-3" />
                  </Button>
                </div>
              </div>
              <div className="p-3">{panel.content}</div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
