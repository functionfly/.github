/**
 * StudioGraphView
 * Full-screen graph runtime view integrated with studio
 */

import * as React from 'react'
import { cn } from '@functionfly/ui-core'
import { GraphRuntimePanel } from './GraphRuntimePanel'
import { useGraphRuntime } from '@/hooks/useGraphRuntime'
import { generateSampleGraph } from '@functionfly/ui-graph'
import type { NodeData, EdgeData } from '@functionfly/ui-graph'

export interface StudioGraphViewProps {
  className?: string
  showAllPanels?: boolean
}

/**
 * StudioGraphView - Complete graph runtime view for the Studio
 * Provides a workspace with graph runtime components
 */
export function StudioGraphView({
  className,
  showAllPanels = false,
}: StudioGraphViewProps) {
  const runtime = useGraphRuntime()
  
  // Initialize with sample graph on mount
  React.useEffect(() => {
    const { nodes, edges } = generateSampleGraph()
    runtime.setNodes(nodes as any[])
    runtime.setEdges(edges)
    
    // Set some default runtime metrics
    runtime.setRuntimeMetrics({
      totalExecutions: 847,
      successfulExecutions: 823,
      failedExecutions: 24,
      averageLatency: 234,
      tokensProcessed: 1250000,
      totalCost: 12.45,
      uptime: 99.97,
    })
    
    // Simulate some execution events
    const now = Date.now()
    runtime.setExecutionEvents([
      { id: 'e1', timestamp: now - 5000, type: 'node_start', nodeId: 'trigger-1', result: 'success' },
      { id: 'e2', timestamp: now - 4500, type: 'edge_traversal', edgeId: 'e1', result: 'success' },
      { id: 'e3', timestamp: now - 4000, type: 'node_start', nodeId: 'auth-agent', result: 'success' },
      { id: 'e4', timestamp: now - 3500, type: 'node_end', nodeId: 'auth-agent', result: 'success' },
      { id: 'e5', timestamp: now - 3000, type: 'node_start', nodeId: 'llm-agent', result: 'success' },
      { id: 'e6', timestamp: now - 2000, type: 'node_end', nodeId: 'llm-agent', result: 'success' },
      { id: 'e7', timestamp: now - 1500, type: 'edge_traversal', edgeId: 'e5', result: 'success' },
      { id: 'e8', timestamp: now - 1000, type: 'node_end', nodeId: 'output-1', result: 'success' },
    ])
    
    // Set replay duration based on events
    runtime.setReplayDuration(now - (now - 5000))
  }, [])

  return (
    <div className={cn("h-full w-full", className)}>
      <GraphRuntimePanel
        showViewport={true}
        showMinimap={true}
        showInspector={true}
        showReplayControls={true}
        showHeatmap={showAllPanels}
        showForkManager={showAllPanels}
        showTelemetry={showAllPanels}
        showDependencyExplorer={showAllPanels}
        showLayerView={showAllPanels}
        showBranchRenderer={showAllPanels}
        showLoopVisualizer={showAllPanels}
        showFailureMap={showAllPanels}
        showResourceOverlay={showAllPanels}
        showVersionHistory={showAllPanels}
        showTraceExplorer={showAllPanels}
        showDistributedMap={showAllPanels}
      />
    </div>
  )
}
