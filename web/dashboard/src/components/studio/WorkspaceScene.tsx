import React, { useState, useCallback, useMemo } from 'react'
import { cn } from '@/lib/utils'
import { Plus, Trash2, Copy, Clipboard } from 'lucide-react'
import { Button } from '@/components/ui/button'

export interface SceneNode {
  id: string
  type: string
  name: string
  position: { x: number; y: number }
  isSelected?: boolean
}

export interface WorkspaceSceneProps {
  nodes?: SceneNode[]
  children?: React.ReactNode
  className?: string
  gridSize?: number
  snapToGrid?: boolean
  onNodeSelect?: (nodeId: string | null) => void
  onNodeMove?: (nodeId: string, position: { x: number; y: number }) => void
  onAddNode?: (type: string) => void
}

/**
 * WorkspaceScene - Scene graph container for the workspace
 * Manages nodes and their relationships in a visual scene
 */
export function WorkspaceScene({
  nodes = [],
  children,
  className,
  gridSize = 20,
  snapToGrid = true,
  onNodeSelect,
  onNodeMove,
  onAddNode,
}: WorkspaceSceneProps) {
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)
  const [sceneOffset, setSceneOffset] = useState({ x: 0, y: 0 })

  const selectedNode = nodes.find(n => n.id === selectedNodeId)

  const handleNodeClick = useCallback((nodeId: string, e: React.MouseEvent) => {
    e.stopPropagation()
    setSelectedNodeId(nodeId)
    onNodeSelect?.(nodeId)
  }, [onNodeSelect])

  const handleSceneClick = useCallback(() => {
    setSelectedNodeId(null)
    onNodeSelect?.(null)
  }, [onNodeSelect])

  const handleAddNode = useCallback((type: string) => {
    onAddNode?.(type)
  }, [onAddNode])

  return (
    <div
      className={cn(
        'relative w-full h-full overflow-hidden',
        'bg-aviation-bg-primary',
        className
      )}
      onClick={handleSceneClick}
    >
      {/* Grid Background */}
      <div
        className="absolute inset-0 opacity-10"
        style={{
          backgroundImage: `
            linear-gradient(rgba(255,255,255,0.05) 1px, transparent 1px),
            linear-gradient(90deg, rgba(255,255,255,0.05) 1px, transparent 1px)
          `,
          backgroundSize: `${gridSize}px ${gridSize}px`,
          backgroundPosition: `${sceneOffset.x % gridSize}px ${sceneOffset.y % gridSize}px`,
        }}
      />

      {/* Nodes Layer */}
      <div className="absolute inset-0">
        {nodes.map(node => (
          <SceneNodeComponent
            key={node.id}
            node={node}
            isSelected={selectedNodeId === node.id}
            onClick={handleNodeClick}
          />
        ))}
      </div>

      {/* Children Content */}
      {children}

      {/* Node Toolbar */}
      {selectedNode && (
        <div className="absolute bottom-4 left-4 aviation-panel p-2 rounded-lg flex items-center gap-2">
          <span className="text-sm text-aviation-text-secondary">{selectedNode.name}</span>
          <div className="w-px h-4 bg-aviation-border-panel" />
          <Button variant="ghost" size="icon" className="w-7 h-7">
            <Copy className="w-3 h-3" />
          </Button>
          <Button variant="ghost" size="icon" className="w-7 h-7">
            <Trash2 className="w-3 h-3" />
          </Button>
        </div>
      )}

      {/* Add Node Button */}
      {onAddNode && (
        <Button
          variant="outline"
          size="sm"
          className="absolute bottom-4 right-4 aviation-panel gap-2"
          onClick={() => handleAddNode('default')}
        >
          <Plus className="w-4 h-4" />
          Add Node
        </Button>
      )}
    </div>
  )
}

interface SceneNodeComponentProps {
  node: SceneNode
  isSelected: boolean
  onClick: (nodeId: string, e: React.MouseEvent) => void
}

function SceneNodeComponent({ node, isSelected, onClick }: SceneNodeComponentProps) {
  return (
    <div
      className={cn(
        'absolute flex items-center justify-center rounded-lg border-2 cursor-pointer transition-all',
        isSelected
          ? 'border-aviation-amber bg-aviation-amber-subtle'
          : 'border-aviation-border-panel bg-aviation-bg-panel hover:border-aviation-cyan'
      )}
      style={{
        left: node.position.x,
        top: node.position.y,
        width: 120,
        height: 60,
      }}
      onClick={e => onClick(node.id, e)}
    >
      <span className="text-sm font-medium text-aviation-text-primary">{node.name}</span>
    </div>
  )
}
