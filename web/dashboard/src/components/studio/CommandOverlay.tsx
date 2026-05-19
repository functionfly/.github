import React, { useState, useEffect, useCallback, useRef } from 'react'
import { cn } from '@/lib/utils'
import { Command, Search, ArrowRight, X } from 'lucide-react'
import { Button } from '@/components/ui/button'

export interface CommandItem {
  id: string
  title: string
  description?: string
  icon?: React.ReactNode
  keywords?: string[]
  onSelect: () => void
  group?: string
}

export interface CommandOverlayProps {
  isOpen: boolean
  onClose: () => void
  commands?: CommandItem[]
  placeholder?: string
  className?: string
}

/**
 * CommandOverlay - Modal command palette overlay
 * Provides quick access to actions via keyboard
 */
export function CommandOverlay({
  isOpen,
  onClose,
  commands = [],
  placeholder = 'Type a command...',
  className,
}: CommandOverlayProps) {
  const [query, setQuery] = useState('')
  const [selectedIndex, setSelectedIndex] = useState(0)
  const inputRef = useRef<HTMLInputElement>(null)

  const filteredCommands = commands.filter(cmd =>
    cmd.title.toLowerCase().includes(query.toLowerCase()) ||
    cmd.keywords?.some(k => k.toLowerCase().includes(query.toLowerCase())) ||
    cmd.description?.toLowerCase().includes(query.toLowerCase())
  )

  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if (!isOpen) return

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault()
        setSelectedIndex(i => Math.min(i + 1, filteredCommands.length - 1))
        break
      case 'ArrowUp':
        e.preventDefault()
        setSelectedIndex(i => Math.max(i - 1, 0))
        break
      case 'Enter':
        e.preventDefault()
        if (filteredCommands[selectedIndex]) {
          filteredCommands[selectedIndex].onSelect()
          onClose()
        }
        break
      case 'Escape':
        e.preventDefault()
        onClose()
        break
    }
  }, [isOpen, filteredCommands, selectedIndex, onClose])

  useEffect(() => {
    if (isOpen) {
      setQuery('')
      setSelectedIndex(0)
      setTimeout(() => inputRef.current?.focus(), 100)
    }
  }, [isOpen])

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-[20vh] pointer-events-none">
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={onClose}
      />
      
      <div
        className={cn(
          'relative w-full max-w-lg mx-4 aviation-panel rounded-lg shadow-2xl pointer-events-auto',
          className
        )}
        onClick={e => e.stopPropagation()}
      >
        {/* Search Input */}
        <div className="flex items-center gap-3 px-4 py-3 border-b border-aviation-border-panel">
          <Search className="w-5 h-5 text-aviation-text-muted" />
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={e => setQuery(e.target.value)}
            placeholder={placeholder}
            className="flex-1 bg-transparent text-aviation-text-primary placeholder:text-aviation-text-dim outline-none"
          />
          <div className="flex items-center gap-1 text-xs text-aviation-text-dim">
            <kbd className="px-1.5 py-0.5 bg-aviation-bg-instrument rounded">ESC</kbd>
          </div>
        </div>

        {/* Command List */}
        <div className="max-h-96 overflow-y-auto py-2">
          {filteredCommands.length === 0 ? (
            <div className="px-4 py-8 text-center text-aviation-text-muted">
              No results found
            </div>
          ) : (
            Object.entries(
              filteredCommands.reduce((acc, cmd) => {
                const group = cmd.group || 'General'
                if (!acc[group]) acc[group] = []
                acc[group].push(cmd)
                return acc
              }, {} as Record<string, CommandItem[]>)
            ).map(([group, cmds]) => (
              <div key={group}>
                <div className="px-4 py-1 text-xs font-medium text-aviation-text-muted uppercase tracking-wider">
                  {group}
                </div>
                {cmds.map((cmd, idx) => {
                  const globalIdx = filteredCommands.indexOf(cmd)
                  return (
                    <button
                      key={cmd.id}
                      onClick={() => {
                        cmd.onSelect()
                        onClose()
                      }}
                      onMouseEnter={() => setSelectedIndex(globalIdx)}
                      className={cn(
                        'w-full flex items-center gap-3 px-4 py-2 text-left hover:bg-aviation-bg-instrument transition-colors',
                        selectedIndex === globalIdx && 'bg-aviation-amber-subtle'
                      )}
                    >
                      {cmd.icon && <span className="text-aviation-text-secondary">{cmd.icon}</span>}
                      <div className="flex-1 min-w-0">
                        <div className="text-sm text-aviation-text-primary">{cmd.title}</div>
                        {cmd.description && (
                          <div className="text-xs text-aviation-text-muted truncate">{cmd.description}</div>
                        )}
                      </div>
                      {selectedIndex === globalIdx && (
                        <ArrowRight className="w-4 h-4 text-aviation-amber" />
                      )}
                    </button>
                  )
                })}
              </div>
            ))
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between px-4 py-2 border-t border-aviation-border-panel text-xs text-aviation-text-dim">
          <div className="flex items-center gap-4">
            <span className="flex items-center gap-1">
              <kbd className="px-1.5 py-0.5 bg-aviation-bg-instrument rounded">↑↓</kbd>
              Navigate
            </span>
            <span className="flex items-center gap-1">
              <kbd className="px-1.5 py-0.5 bg-aviation-bg-instrument rounded">↵</kbd>
              Select
            </span>
          </div>
          <Button variant="ghost" size="icon" className="w-6 h-6" onClick={onClose}>
            <X className="w-3 h-3" />
          </Button>
        </div>
      </div>
    </div>
  )
}
