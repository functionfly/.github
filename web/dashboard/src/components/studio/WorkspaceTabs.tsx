import React, { useState, useCallback } from 'react'
import { cn } from '@/lib/utils'
import { X, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'

export interface WorkspaceTab {
  id: string
  label: string
  icon?: React.ReactNode
  content: React.ReactNode
  isClosable?: boolean
  isActive?: boolean
}

export interface WorkspaceTabsProps {
  tabs: WorkspaceTab[]
  className?: string
  defaultActiveTab?: string
  onTabChange?: (tabId: string) => void
  onTabClose?: (tabId: string) => void
  onTabAdd?: () => void
  showAddButton?: boolean
}

/**
 * WorkspaceTabs - Tabbed workspace interface
 * Manages multiple workspaces with tab switching
 */
export function WorkspaceTabs({
  tabs,
  className,
  defaultActiveTab,
  onTabChange,
  onTabClose,
  onTabAdd,
  showAddButton = false,
}: WorkspaceTabsProps) {
  const [activeTabId, setActiveTabId] = useState(defaultActiveTab || tabs[0]?.id || '')

  const handleTabClick = useCallback((tabId: string) => {
    setActiveTabId(tabId)
    onTabChange?.(tabId)
  }, [onTabChange])

  const handleCloseTab = useCallback((tabId: string, e: React.MouseEvent) => {
    e.stopPropagation()
    onTabClose?.(tabId)
  }, [onTabClose])

  const activeTab = tabs.find(t => t.id === activeTabId)

  return (
    <div className={cn('flex flex-col h-full w-full', className)}>
      {/* Tab Bar */}
      <div className="flex items-center gap-1 p-2 border-b border-aviation-border-panel overflow-x-auto">
        {tabs.map(tab => (
          <button
            key={tab.id}
            onClick={() => handleTabClick(tab.id)}
            className={cn(
              'flex items-center gap-2 px-3 py-1.5 text-sm rounded-t-lg transition-colors',
              activeTabId === tab.id
                ? 'bg-aviation-bg-panel text-aviation-text-primary border-b-2 border-aviation-amber'
                : 'text-aviation-text-secondary hover:bg-aviation-bg-instrument'
            )}
          >
            {tab.icon && <span className="w-4 h-4">{tab.icon}</span>}
            <span>{tab.label}</span>
            {tab.isClosable !== false && (
              <button
                onClick={e => handleCloseTab(tab.id, e)}
                className="ml-1 rounded hover:bg-aviation-bg-panel"
              >
                <X className="w-3 h-3" />
              </button>
            )}
          </button>
        ))}

        {/* Add Tab Button */}
        {showAddButton && (
          <Button
            variant="ghost"
            size="icon"
            className="w-7 h-7"
            onClick={onTabAdd}
          >
            <Plus className="w-4 h-4" />
          </Button>
        )}
      </div>

      {/* Tab Content */}
      <div className="flex-1 overflow-hidden">
        {activeTab?.content}
      </div>
    </div>
  )
}
