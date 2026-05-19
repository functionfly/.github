import React, { useState, useEffect, useCallback, useMemo } from 'react'
import { cn } from '@/lib/utils'
import { useWindowSize } from '@/hooks/useWindowSize'

export interface AdaptiveWorkspaceProps {
  children?: React.ReactNode
  className?: string
  breakpoints?: {
    mobile?: number
    tablet?: number
    desktop?: number
    ultrawide?: number
  }
  onModeChange?: (mode: WorkspaceMode) => void
}

type WorkspaceMode = 'compact' | 'standard' | 'expanded' | 'ultrawide'

/**
 * AdaptiveWorkspace - Responsive workspace that adapts to screen size
 * Automatically adjusts layout based on available viewport dimensions
 */
export function AdaptiveWorkspace({
  children,
  className,
  breakpoints = {
    mobile: 768,
    tablet: 1024,
    desktop: 1440,
    ultrawide: 1920,
  },
  onModeChange,
}: AdaptiveWorkspaceProps) {
  const [windowWidth, windowHeight] = useWindowSize()
  const [mode, setMode] = useState<WorkspaceMode>('standard')

  useEffect(() => {
    const determineMode = () => {
      if (windowWidth < breakpoints.mobile) return 'compact'
      if (windowWidth < breakpoints.tablet) return 'compact'
      if (windowWidth < breakpoints.desktop) return 'standard'
      if (windowWidth < breakpoints.ultrawide) return 'expanded'
      return 'ultrawide'
    }

    const newMode = determineMode()
    if (newMode !== mode) {
      setMode(newMode)
      onModeChange?.(newMode)
    }
  }, [windowWidth, mode, breakpoints, onModeChange])

  const getLayoutClasses = useCallback(() => {
    switch (mode) {
      case 'compact':
        return 'flex flex-col h-full'
      case 'ultrawide':
        return 'grid grid-cols-[280px_1fr_320px] h-full gap-4 p-4'
      default:
        return 'flex flex-row h-full'
    }
  }, [mode])

  return (
    <div
      className={cn(
        'w-full h-full transition-all duration-300',
        getLayoutClasses(),
        className
      )}
      data-workspace-mode={mode}
      data-window-size={`${windowWidth}x${windowHeight}`}
    >
      {children}
    </div>
  )
}

// Hook for getting current workspace mode
export function useWorkspaceMode() {
  const [windowWidth] = useWindowSize()

  return useMemo(() => {
    if (windowWidth < 768) return 'compact'
    if (windowWidth < 1024) return 'compact'
    if (windowWidth < 1440) return 'standard'
    if (windowWidth < 1920) return 'expanded'
    return 'ultrawide'
  }, [windowWidth])
}
