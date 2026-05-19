import React, { useState, useCallback } from 'react'
import { cn } from '@/lib/utils'

export interface ActivityItem {
  id: string
  label: string
  icon: React.ReactNode
  badge?: number | string
  isActive?: boolean
  isDisabled?: boolean
}

export interface ActivityRailProps {
  items: ActivityItem[]
  className?: string
  position?: 'left' | 'right'
  onItemClick?: (itemId: string) => void
  defaultActiveId?: string
}

/**
 * ActivityRail - Vertical navigation rail for primary actions
 * Compact vertical bar with icons for quick navigation
 */
export function ActivityRail({
  items,
  className,
  position = 'left',
  onItemClick,
  defaultActiveId,
}: ActivityRailProps) {
  const [activeId, setActiveId] = useState(defaultActiveId || items[0]?.id || '')

  const handleItemClick = useCallback((itemId: string) => {
    setActiveId(itemId)
    onItemClick?.(itemId)
  }, [onItemClick])

  return (
    <div
      className={cn(
        'flex flex-col items-center py-4 gap-2',
        'bg-aviation-bg-secondary border-r border-aviation-border-panel',
        position === 'right' && 'border-r-0 border-l',
        className
      )}
    >
      {items.map(item => (
        <button
          key={item.id}
          onClick={() => !item.isDisabled && handleItemClick(item.id)}
          disabled={item.isDisabled}
          className={cn(
            'relative flex items-center justify-center w-10 h-10 rounded-lg transition-colors',
            activeId === item.id
              ? 'bg-aviation-amber text-aviation-bg-primary'
              : 'text-aviation-text-secondary hover:bg-aviation-bg-instrument hover:text-aviation-text-primary',
            item.isDisabled && 'opacity-50 cursor-not-allowed'
          )}
          title={item.label}
        >
          <span className="w-5 h-5">{item.icon}</span>
          
          {/* Badge */}
          {item.badge && (
            <span className="absolute -top-1 -right-1 min-w-[16px] h-4 px-1 text-[10px] font-bold bg-aviation-red text-white rounded-full flex items-center justify-center">
              {typeof item.badge === 'number' && item.badge > 9 ? '9+' : item.badge}
            </span>
          )}
        </button>
      ))}
    </div>
  )
}
