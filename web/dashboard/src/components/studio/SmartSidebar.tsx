import React, { useState, useEffect, useCallback } from 'react'
import { cn } from '@/lib/utils'
import { ChevronLeft, ChevronRight, Settings, Home, BarChart3, Users } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useLocation } from 'react-router-dom'

export interface SidebarSection {
  id: string
  label: string
  icon: React.ReactNode
  path?: string
  badge?: number | string
  children?: SidebarSection[]
}

export interface SmartSidebarProps {
  sections?: SidebarSection[]
  className?: string
  defaultCollapsed?: boolean
}

/**
 * SmartSidebar - Context-aware sidebar with adaptive visibility
 * Expands/collapses based on context and user behavior
 */
export function SmartSidebar({
  sections = [],
  className,
  defaultCollapsed = false,
}: SmartSidebarProps) {
  const [isCollapsed, setIsCollapsed] = useState(defaultCollapsed)
  const [expandedSections, setExpandedSections] = useState<Set<string>>(new Set())
  const location = useLocation()

  const toggleSection = useCallback((sectionId: string) => {
    setExpandedSections(prev => {
      const next = new Set(prev)
      if (next.has(sectionId)) {
        next.delete(sectionId)
      } else {
        next.add(sectionId)
      }
      return next
    })
  }, [])

  const isActiveSection = useCallback((section: SidebarSection) => {
    if (section.path) {
      return location.pathname === section.path || location.pathname.startsWith(section.path)
    }
    return false
  }, [location])

  const defaultSections: SidebarSection[] = [
    {
      id: 'dashboard',
      label: 'Dashboard',
      icon: <Home className="w-5 h-5" />,
      path: '/',
    },
    {
      id: 'analytics',
      label: 'Analytics',
      icon: <BarChart3 className="w-5 h-5" />,
      path: '/analytics',
    },
    {
      id: 'team',
      label: 'Team',
      icon: <Users className="w-5 h-5" />,
      path: '/team',
    },
    {
      id: 'settings',
      label: 'Settings',
      icon: <Settings className="w-5 h-5" />,
      path: '/settings',
    },
  ]

  const sidebarSections = sections.length > 0 ? sections : defaultSections

  return (
    <div
      className={cn(
        'h-full aviation-sidebar transition-all duration-300',
        isCollapsed ? 'w-16' : 'w-64',
        className
      )}
    >
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-aviation-border-panel">
        {!isCollapsed && (
          <span className="text-lg font-bold text-aviation-text-primary">Studio</span>
        )}
        <Button
          variant="ghost"
          size="icon"
          className="w-8 h-8"
          onClick={() => setIsCollapsed(c => !c)}
        >
          {isCollapsed ? (
            <ChevronRight className="w-4 h-4" />
          ) : (
            <ChevronLeft className="w-4 h-4" />
          )}
        </Button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-2 space-y-1 overflow-y-auto">
        {sidebarSections.map(section => (
          <SidebarItem
            key={section.id}
            section={section}
            isCollapsed={isCollapsed}
            isActive={isActiveSection(section)}
            isExpanded={expandedSections.has(section.id)}
            onToggle={() => toggleSection(section.id)}
          />
        ))}
      </nav>
    </div>
  )
}

interface SidebarItemProps {
  section: SidebarSection
  isCollapsed: boolean
  isActive: boolean
  isExpanded: boolean
  onToggle: () => void
}

function SidebarItem({ section, isCollapsed, isActive, isExpanded, onToggle }: SidebarItemProps) {
  const hasChildren = section.children && section.children.length > 0

  return (
    <div>
      <button
        onClick={onToggle}
        className={cn(
          'w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors',
          isActive
            ? 'bg-aviation-amber-subtle text-aviation-amber'
            : 'text-aviation-text-secondary hover:bg-aviation-bg-instrument hover:text-aviation-text-primary'
        )}
      >
        <span className={cn('w-5 h-5', isCollapsed && 'mx-auto')}>{section.icon}</span>
        {!isCollapsed && (
          <>
            <span className="flex-1 text-left">{section.label}</span>
            {section.badge && (
              <span className="px-2 py-0.5 text-xs bg-aviation-amber rounded-full">
                {section.badge}
              </span>
            )}
          </>
        )}
      </button>

      {/* Sub-items */}
      {hasChildren && isExpanded && !isCollapsed && (
        <div className="ml-4 mt-1 space-y-1">
          {section.children!.map(child => (
            <a
              key={child.id}
              href={child.path}
              className="block px-3 py-1.5 text-sm text-aviation-text-muted rounded hover:bg-aviation-bg-instrument hover:text-aviation-text-primary"
            >
              {child.label}
            </a>
          ))}
        </div>
      )}
    </div>
  )
}
