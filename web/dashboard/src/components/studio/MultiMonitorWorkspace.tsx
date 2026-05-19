import React, { useState, useCallback, useEffect } from 'react'
import { cn } from '@/lib/utils'
import { Monitor, MonitorOff } from 'lucide-react'

export interface Monitor {
  id: string
  name: string
  isActive: boolean
  bounds: {
    x: number
    y: number
    width: number
    height: number
  }
}

export interface MultiMonitorWorkspaceProps {
  monitors?: Monitor[]
  children?: React.ReactNode
  className?: string
  onMonitorChange?: (monitorId: string | null) => void
}

/**
 * MultiMonitorWorkspace - Supports multi-monitor setups
 * Detects and manages content across multiple displays
 */
export function MultiMonitorWorkspace({
  monitors = [],
  children,
  className,
  onMonitorChange,
}: MultiMonitorWorkspaceProps) {
  const [detectedMonitors, setDetectedMonitors] = useState<Monitor[]>(monitors)
  const [activeMonitorId, setActiveMonitorId] = useState<string | null>(null)

  // Detect screens (simulated - in real app would use Screen API)
  useEffect(() => {
    if (monitors.length === 0) {
      // Simulate detection
      const primary: Monitor = {
        id: 'primary',
        name: 'Primary Display',
        isActive: true,
        bounds: { x: 0, y: 0, width: window.screen.width, height: window.screen.height },
      }
      setDetectedMonitors([primary])
      setActiveMonitorId('primary')
    }
  }, [monitors])

  const handleMonitorSelect = useCallback((monitorId: string) => {
    setActiveMonitorId(monitorId)
    onMonitorChange?.(monitorId)
  }, [onMonitorChange])

  const activeMonitor = detectedMonitors.find(m => m.id === activeMonitorId)

  return (
    <div className={cn('relative w-full h-full', className)}>
      {/* Monitor Selector - Top Bar */}
      {detectedMonitors.length > 1 && (
        <div className="absolute top-2 right-2 z-50 flex items-center gap-2 aviation-panel p-1 rounded-lg">
          <Monitor className="w-4 h-4 text-aviation-text-muted" />
          {detectedMonitors.map(monitor => (
            <button
              key={monitor.id}
              onClick={() => handleMonitorSelect(monitor.id)}
              className={cn(
                'px-2 py-1 text-xs rounded',
                activeMonitorId === monitor.id
                  ? 'bg-aviation-amber text-aviation-bg-primary'
                  : 'hover:bg-aviation-bg-instrument text-aviation-text-secondary'
              )}
            >
              {monitor.name}
            </button>
          ))}
        </div>
      )}

      {/* Content Area */}
      <div className="w-full h-full" style={activeMonitor ? {
        transform: `translate(${activeMonitor.bounds.x}px, ${activeMonitor.bounds.y}px)`,
        width: activeMonitor.bounds.width,
        height: activeMonitor.bounds.height,
      } : {}}>
        {children}
      </div>
    </div>
  )
}

// Hook to detect screen count
export function useScreenDetection() {
  const [screenCount, setScreenCount] = useState(1)

  useEffect(() => {
    // Screen enumeration API (limited browser support)
    if ('getScreenDetails' in window) {
      // @ts-ignore
      window.getScreenDetails().then((screens) => {
        setScreenCount(screens.length)
      }).catch(() => {
        // Fallback
        setScreenCount(window.screen.isExtended ? 2 : 1)
      })
    } else {
      setScreenCount(window.screen.isExtended ? 2 : 1)
    }
  }, [])

  return screenCount
}
