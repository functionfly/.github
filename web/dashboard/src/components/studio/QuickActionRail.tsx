import React, { useState, useCallback } from 'react'
import { cn } from '@/lib/utils'
import { Zap, Plus, RefreshCw, Save, Download, Upload } from 'lucide-react'
import { Button } from '@/components/ui/button'

export interface QuickAction {
  id: string
  label: string
  icon: React.ReactNode
  onClick: () => void
  variant?: 'default' | 'primary' | 'danger'
  isEnabled?: boolean
}

export interface QuickActionRailProps {
  actions?: QuickAction[]
  className?: string
  position?: 'top' | 'bottom' | 'left' | 'right'
}

const defaultActions: QuickAction[] = [
  { id: 'new', label: 'New', icon: <Plus className="w-4 h-4" />, onClick: () => {} },
  { id: 'save', label: 'Save', icon: <Save className="w-4 h-4" />, onClick: () => {} },
  { id: 'deploy', label: 'Deploy', icon: <Zap className="w-4 h-4" />, onClick: () => {}, variant: 'primary' },
  { id: 'refresh', label: 'Refresh', icon: <RefreshCw className="w-4 h-4" />, onClick: () => {} },
]

/**
 * QuickActionRail - Quick action panel for common operations
 * Horizontally or vertically aligned action buttons
 */
export function QuickActionRail({
  actions = defaultActions,
  className,
  position = 'top',
}: QuickActionRailProps) {
  const [hoveredAction, setHoveredAction] = useState<string | null>(null)

  return (
    <div
      className={cn(
        'flex items-center gap-1 p-2 aviation-panel rounded-lg',
        position === 'top' && 'flex-row',
        position === 'bottom' && 'flex-row',
        position === 'left' && 'flex-col',
        position === 'right' && 'flex-col',
        className
      )}
    >
      {actions.map(action => (
        <Button
          key={action.id}
          variant={action.variant === 'primary' ? 'default' : 'ghost'}
          size="icon"
          className={cn(
            'w-8 h-8 relative',
            action.variant === 'primary' && 'bg-aviation-amber text-aviation-bg-primary',
            action.variant === 'danger' && 'hover:text-aviation-red'
          )}
          onClick={action.onClick}
          disabled={!action.isEnabled}
          onMouseEnter={() => setHoveredAction(action.id)}
          onMouseLeave={() => setHoveredAction(null)}
        >
          {action.icon}
          
          {/* Tooltip */}
          {hoveredAction === action.id && (
            <div className="absolute bottom-full mb-2 px-2 py-1 text-xs bg-aviation-bg-panel rounded shadow-lg whitespace-nowrap">
              {action.label}
            </div>
          )}
        </Button>
      ))}
    </div>
  )
}
