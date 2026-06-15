import React, { useState, useEffect, useCallback } from 'react'
import { cn } from '@/lib/utils'
import { Menu, Bell, Command, Zap, Shield, Activity, Maximize2, Sparkles } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useNavigationStatus } from '@/hooks/useNavigationStatus'
import { useThemeStore } from '@/stores/themeStore'
import { useStudioContext } from './StudioShell'
import { useAICommandSystem } from '@/hooks/useAICommandSystem'
import { Badge } from '@/components/ui/badge'
import { useAuthStore } from '@/stores/authStore'

export interface StudioTopbarProps {
  className?: string
  onMenuClick?: () => void
}

/**
 * StudioTopbar - Studio-specific topbar with aviation styling
 * Contains navigation, status indicators, and quick actions
 * Includes AI Command Panel integration
 */
export function StudioTopbar({
  className,
  onMenuClick,
}: StudioTopbarProps) {
  const [notificationPulse, setNotificationPulse] = useState(false)
  const status = useNavigationStatus()
  const { isDarkMode } = useThemeStore()
  const ai = useAICommandSystem()
  const { user } = useAuthStore()
  
  // Get studio context for command palette
  let setCommandPaletteOpen = useCallback((_v: boolean) => {}, [])
  let setImmersiveMode = useCallback((_v: boolean) => {}, [])
  
  try {
    const ctx = useStudioContext()
    setCommandPaletteOpen = ctx.setCommandPaletteOpen
    setImmersiveMode = ctx.setImmersiveMode
  } catch {
    // Not inside StudioShell - context not available
  }

  // Calculate total notifications
  const totalNotifications = status.functions.pendingDeployments +
    (status.functions.hasIssues ? 1 : 0) +
    (status.providers.hasOffline ? 1 : 0)

  // Pulse animation when notifications change
  useEffect(() => {
    if (totalNotifications > 0) {
      setNotificationPulse(true)
      const timer = setTimeout(() => setNotificationPulse(false), 2000)
      return () => clearTimeout(timer)
    }
  }, [totalNotifications])

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        setCommandPaletteOpen(true)
      }
    }
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [setCommandPaletteOpen])

  const getStatusColor = () => {
    if (status.providers.hasOffline || status.functions.hasIssues) return 'bg-aviation-red'
    if (status.analytics.hasAlerts) return 'bg-aviation-amber'
    if (status.functions.pendingDeployments > 0) return 'bg-aviation-cyan'
    return 'bg-aviation-green'
  }

  return (
    <header
      className={cn(
        'h-14 bg-aviation-bg-primary/95 backdrop-blur-xl',
        'border-b border-aviation-border-panel',
        'flex items-center justify-between px-4',
        className
      )}
    >
      {/* Left Section */}
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="icon"
          className="lg:hidden text-aviation-text-secondary hover:text-aviation-text-primary"
          onClick={onMenuClick}
        >
          <Menu className="w-5 h-5" />
        </Button>

        <div className="flex items-center gap-3">
          <Zap className="w-5 h-5 text-aviation-amber" />
          <span className="font-bold text-aviation-text-primary">Studio</span>
        </div>

        {/* Status Indicator */}
        <div className="flex items-center gap-2">
          <div className={cn('w-2 h-2 rounded-full', getStatusColor())} />
          <span className="text-xs text-aviation-text-muted">Operational</span>
        </div>
      </div>

      {/* Center Section - AI Quick Access */}
      <div className="hidden md:flex items-center gap-2">
        {/* AI Thinking Indicator */}
        {ai.isThinking && (
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-brand-500/10 border border-brand-500/20">
            <div className="size-2 rounded-full bg-brand-500 animate-pulse" />
            <span className="text-xs text-brand-500">AI Processing...</span>
          </div>
        )}

        {/* Confidence Badge */}
        {ai.lastConfidence > 0 && (
          <div className="flex items-center gap-1.5 px-2 py-1 rounded-lg bg-bg-tertiary">
            <Sparkles className="size-3 text-brand-500" />
            <span className="text-xs font-mono text-text-secondary">
              {Math.round(ai.lastConfidence * 100)}%
            </span>
          </div>
        )}
      </div>

      {/* Right Section */}
      <div className="flex items-center gap-2">
        {/* AI Command Button */}
        <Button
          variant="ghost"
          size="icon"
          className="relative text-aviation-text-secondary hover:text-brand-500"
          onClick={() => ai.openCommandPalette()}
          title="AI Command Panel (Ctrl+K)"
        >
          <Sparkles className="w-4 h-4" />
        </Button>

        {/* Immersive Mode Toggle */}
        <Button
          variant="ghost"
          size="icon"
          className="text-aviation-text-secondary hover:text-aviation-text-primary"
          onClick={() => setImmersiveMode(true)}
          title="Immersive Mode"
        >
          <Maximize2 className="w-4 h-4" />
        </Button>

        {/* Command Palette Trigger */}
        <Button
          variant="ghost"
          size="sm"
          className="hidden md:flex items-center gap-2 text-aviation-text-muted"
          onClick={() => setCommandPaletteOpen(true)}
        >
          <Command className="w-4 h-4" />
          <span className="hidden lg:inline">Search</span>
          <kbd className="hidden xl:inline-flex items-center text-[10px] font-mono bg-aviation-bg-instrument px-1.5 py-0.5 rounded">
            ⌘K
          </kbd>
        </Button>

        {/* Notifications */}
        <Button
          variant="ghost"
          size="icon"
          className={cn('relative', notificationPulse && 'animate-pulse')}
        >
          <Bell className="w-5 h-5" />
          {totalNotifications > 0 && (
            <span className="absolute -top-1 -right-1 w-4 h-4 bg-aviation-red rounded-full text-[10px] flex items-center justify-center">
              {totalNotifications}
            </span>
          )}
        </Button>

        {/* User Avatar */}
        {user && (
          <div className="flex items-center gap-2 pl-2 border-l border-aviation-border-panel">
            <div className="relative">
              {user.avatar ? (
                <img
                  src={user.avatar}
                  alt={user.username || user.name || user.email}
                  className="w-8 h-8 rounded-full object-cover ring-2 ring-aviation-border-panel"
                />
              ) : (
                <div className="w-8 h-8 rounded-full bg-brand-500/20 flex items-center justify-center ring-2 ring-brand-500/30">
                  <span className="text-xs font-semibold text-brand-400">
                    {(user.username || user.name || user.email || 'U')[0].toUpperCase()}
                  </span>
                </div>
              )}
              <div className="absolute -bottom-0.5 -right-0.5 w-3 h-3 bg-aviation-green rounded-full border-2 border-aviation-bg-primary" />
            </div>
            <div className="hidden lg:flex flex-col">
              <span className="text-xs font-medium text-aviation-text-primary truncate max-w-[100px]">
                {user.username || user.name || user.email?.split('@')[0]}
              </span>
              <span className="text-[10px] text-aviation-text-muted truncate max-w-[100px]">
                {user.plan}
              </span>
            </div>
          </div>
        )}
      </div>
    </header>
  )
}
