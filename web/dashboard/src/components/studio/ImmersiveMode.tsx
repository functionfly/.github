import React, { useState, useEffect, useCallback } from 'react'
import { cn } from '@/lib/utils'
import { Maximize, Minimize, X } from 'lucide-react'
import { Button } from '@/components/ui/button'

export interface ImmersiveModeProps {
  isEnabled: boolean
  onToggle: () => void
  children?: React.ReactNode
  className?: string
  hideScrollbar?: boolean
}

/**
 * ImmersiveMode - Fullscreen distraction-free mode
 * Hides all UI chrome for focused work
 */
export function ImmersiveMode({
  isEnabled,
  onToggle,
  children,
  className,
  hideScrollbar = true,
}: ImmersiveModeProps) {
  const [showExitHint, setShowExitHint] = useState(false)

  useEffect(() => {
    if (isEnabled) {
      // Show hint after 1 second
      const timer = setTimeout(() => setShowExitHint(true), 1000)
      return () => clearTimeout(timer)
    } else {
      setShowExitHint(false)
    }
  }, [isEnabled])

  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    // Exit on Escape
    if (e.key === 'Escape' && isEnabled) {
      onToggle()
    }
  }, [isEnabled, onToggle])

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])

  if (!isEnabled) {
    return <>{children}</>
  }

  return (
    <div
      className={cn(
        'fixed inset-0 z-50 bg-aviation-bg-primary',
        hideScrollbar && 'overflow-hidden',
        className
      )}
    >
      {/* Content */}
      <div className="w-full h-full">{children}</div>

      {/* Exit Hint */}
      {showExitHint && (
        <div className="absolute bottom-4 right-4 aviation-panel px-3 py-2 rounded-lg animate-fade-in-up">
          <div className="flex items-center gap-2 text-sm text-aviation-text-secondary">
            <kbd className="px-1.5 py-0.5 bg-aviation-bg-instrument rounded">ESC</kbd>
            Exit Immersive Mode
          </div>
        </div>
      )}

      {/* Exit Button */}
      <Button
        variant="ghost"
        size="icon"
        className="absolute top-4 right-4 w-10 h-10 rounded-full bg-aviation-bg-panel/80 backdrop-blur"
        onClick={onToggle}
      >
        <X className="w-5 h-5" />
      </Button>
    </div>
  )
}
