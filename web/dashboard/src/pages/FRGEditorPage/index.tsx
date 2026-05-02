/**
 * FRG Editor Page
 * Main page for the Function Runtime Graph editor
 * Integrates all FRG components with the 3-panel layout
 */

import { useCallback, useEffect, useState, useRef } from 'react';
import {
  ReactFlowProvider,
  Background,
  BackgroundVariant,
  Controls,
  MiniMap,
  Panel,
  useReactFlow,
  type Connection,
  type NodeChange,
  type EdgeChange,
  useStore,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';

import { GraphCanvas } from '@/components/frg/canvas/GraphCanvas';
import { FunctionNode } from '@/components/frg/nodes/FunctionNode';
import { CustomEdge } from '@/components/frg/edges/CustomEdge';
import { NodeInspector } from '@/components/frg/panels/NodeInspector';
import { FunctionLibrary } from '@/components/frg/panels/FunctionLibrary';
import { AIAssistantPanel } from '@/components/frg/panels/AIAssistantPanel';
import { TestRunnerPanel } from '@/components/frg/test/TestRunnerPanel';
import { VersionSelector } from '@/components/frg/panels/VersionSelector';
import { EvolutionPanel } from '@/components/frg/panels/EvolutionPanel';
import { ExecutionBar } from '@/components/frg/execution/ExecutionBar';
import { ExecutionOverlay } from '@/components/frg/execution/ExecutionOverlay';
import { CollapsibleExecutionPanel } from '@/components/frg/execution/CollapsibleExecutionPanel';
import { EmptyStateOverlay } from '@/components/frg/overlays/EmptyStateOverlay';
import { CanvasControls } from '@/components/frg/controls/CanvasControls';
import { KeyboardShortcutsHelp } from '@/components/frg/controls/KeyboardShortcutsHelp';
import { LiveCursors } from '@/components/frg/collaboration/LiveCursors';
import { CommentPins } from '@/components/frg/collaboration/CommentPins';
import { useFRGStore } from '@/stores/frgStore';
import { frgApi } from '@/api/frg';

import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { Separator } from '@/components/ui/separator';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';

import {
  Play,
  Square,
  Pause,
  StepForward,
  Bug,
  Save,
  Undo2,
  Redo2,
  Library,
  Sparkles,
  TestTube,
  GitBranch,
  PanelLeft,
  PanelRight,
  Settings,
  Share,
  MoreHorizontal,
  Loader2,
  Maximize2,
  Minimize2,
  Download,
  Upload,
  HelpCircle,
  Keyboard,
  Users,
  MessageSquare,
  Wand2,
} from 'lucide-react';
import { useParams, useNavigate } from 'react-router-dom';

// Node types mapping
const nodeTypes = {
  functionNode: FunctionNode,
};

// Edge types mapping
const edgeTypes = {
  custom: CustomEdge,
};

// Inner component that has access to React Flow context
function FRGEditorInner() {
  const reactFlow = useReactFlow();
  const store = useFRGStore();
  const { id } = useParams();
  const navigate = useNavigate();
  
  const {
    nodes,
    edges,
    selectedNodeId,
    editorMode,
    leftPanel,
    rightPanel,
    executionStatus,
    isDirty,
    canUndo,
    canRedo,
    setNodes,
    setEdges,
    setSelectedNode,
    setEditorMode,
    toggleLeftPanel,
    toggleRightPanel,
    onConnect,
    undo,
    redo,
    startExecution,
    pauseExecution,
    stopExecution,
    saveGraph,
    isLoading,
    isSaving,
    isExecuting,
    lastSavedAt,
    operationQueue,
    setViewport,
  } = store;

  const [isFullscreen, setIsFullscreen] = useState(false);
  const [graphName, setGraphName] = useState('Untitled Graph');
  const [showShortcuts, setShowShortcuts] = useState(false);
  const [showCollaboration, setShowCollaboration] = useState(false);
  const [presentationMode, setPresentationMode] = useState(false);
  const [isNewGraph, setIsNewGraph] = useState(!id && nodes.length === 0);
  const [autoSaveTimer, setAutoSaveTimer] = useState<NodeJS.Timeout | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  // Dirty-state navigation guard
  useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      if (store.isDirty) {
        e.preventDefault();
        e.returnValue = 'You have unsaved changes. Are you sure you want to leave?';
        return e.returnValue;
      }
    };
    window.addEventListener('beforeunload', handleBeforeUnload);
    return () => window.removeEventListener('beforeunload', handleBeforeUnload);
  }, [store.isDirty]);

  // Auto-save mechanism with debounce
  useEffect(() => {
    if (!store.isDirty || !store.autoSaveEnabled) return;

    if (autoSaveTimer) clearTimeout(autoSaveTimer);

    const timer = setTimeout(() => {
      if (store.isDirty && store.graphName && store.graphName !== 'Untitled Graph') {
        store.saveGraph();
      }
    }, store.autoSaveInterval);

    setAutoSaveTimer(timer);
    return () => {
      if (timer) clearTimeout(timer);
    };
  }, [store.isDirty, store.autoSaveEnabled, store.autoSaveInterval, store.graphName]);

  // Online/offline detection
  useEffect(() => {
    const handleOnline = () => store.setIsOffline(false);
    const handleOffline = () => store.setIsOffline(true);

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    // Set initial state
    if (!navigator.onLine) store.setIsOffline(true);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, []);

  // Check if this is a new graph
  useEffect(() => {
    setIsNewGraph(!id && nodes.length === 0);
  }, [id, nodes.length]);

  // Handle node changes
  const onNodesChange = useCallback((changes: NodeChange[]) => {
    // Apply changes using React Flow's built-in handler
    // store.onNodesChange(changes);
  }, [store]);

  const onEdgesChange = useCallback((changes: EdgeChange[]) => {
    // store.onEdgesChange(changes);
  }, [store]);

  // Handle node selection
  const onNodeClick = useCallback((_: React.MouseEvent, node: { id: string }) => {
    setSelectedNode(node.id);
  }, [setSelectedNode]);

  const onPaneClick = useCallback(() => {
    setSelectedNode(null);
  }, [setSelectedNode]);

  // Handle keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Undo/Redo
      if ((e.metaKey || e.ctrlKey) && e.key === 'z') {
        e.preventDefault();
        if (e.shiftKey) {
          redo();
        } else {
          undo();
        }
        return;
      }
      
      // Save
      if ((e.metaKey || e.ctrlKey) && e.key === 's') {
        e.preventDefault();
        // Save action
        return;
      }
      
      // Run graph
      if (e.key === 'F5' || ((e.metaKey || e.ctrlKey) && e.key === 'Enter')) {
        e.preventDefault();
        if (executionStatus === 'idle' || executionStatus === 'completed' || executionStatus === 'failed') {
          startExecution();
        }
        return;
      }
      
      // Fit view
      if (e.key === 'f' && !e.metaKey && !e.ctrlKey && !e.altKey) {
        e.preventDefault();
        reactFlow.fitView({ padding: 0.2 });
        return;
      }
      
      // Help
      if (e.key === '?') {
        e.preventDefault();
        setShowShortcuts(true);
        return;
      }
      
      // Delete selected
      if (e.key === 'Delete' && selectedNodeId) {
        e.preventDefault();
        // store.removeNode(selectedNodeId);
        return;
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [undo, redo, startExecution, executionStatus, selectedNodeId, reactFlow]);

  // Toggle fullscreen
  const toggleFullscreen = useCallback(() => {
    if (!document.fullscreenElement) {
      containerRef.current?.requestFullscreen();
      setIsFullscreen(true);
    } else {
      document.exitFullscreen();
      setIsFullscreen(false);
    }
  }, []);

  // Handle run/stop/pause/step
  const handleRun = useCallback(() => {
    if (executionStatus === 'idle' || executionStatus === 'completed' || executionStatus === 'failed') {
      startExecution();
    } else if (executionStatus === 'running') {
      pauseExecution();
    } else if (executionStatus === 'paused') {
      pauseExecution(); // resume
    }
  }, [executionStatus, startExecution, pauseExecution]);

  const handleStop = useCallback(() => {
    stopExecution();
  }, [stopExecution]);

  const handleSave = useCallback(async () => {
    if (!graphName.trim()) return;
    await saveGraph();
  }, [graphName, saveGraph]);

  // Handle template selection from empty state
  const handleTemplateSelect = useCallback((template: string) => {
    setIsNewGraph(false);
    // Template logic would populate the graph here
  }, []);

  // Handle AI prompt submission
  const handleAIPrompt = useCallback(async (prompt: string) => {
    setIsNewGraph(false);

    // Show loading state
    store.setIsLoading(true);

    // Call AI composition API
    try {
      const response = await frgApi.aiCompose({
        prompt,
        requirements: [],
      });

      console.log('[handleAIPrompt] Response:', JSON.stringify(response).slice(0, 500));

      // Handle both snake_case (from backend) and camelCase (if normalized)
      const graphData = response.graph as any;
      const nodeRefs = graphData?.nodeRefs || graphData?.node_refs;
      const edgesData = graphData?.edges || graphData?.edges;

      console.log('[handleAIPrompt] nodeRefs:', nodeRefs);
      console.log('[handleAIPrompt] Store nodes before:', store.nodes.length);

      if (nodeRefs && Array.isArray(nodeRefs) && nodeRefs.length > 0) {
        // Convert graph definition to React Flow nodes/edges
        const flowNodes = nodeRefs.map((ref: any) => ({
          id: ref.nodeId || ref.node_id,
          type: 'functionNode',
          position: { x: 100 + Math.random() * 400, y: 100 + Math.random() * 300 },
          data: {
            functionRef: ref,
            runtimeState: null,
            isSelected: false,
            isEditable: true,
          },
        }));

        const flowEdges = (edgesData || []).map((edge: any) => ({
          id: edge.id,
          source: edge.sourceNodeId || edge.source_node_id,
          target: edge.targetNodeId || edge.target_node_id,
          type: 'custom',
          data: {
            mapping: edge.mapping || { sourcePath: '*', targetPath: '*' },
          },
        }));

        console.log('[handleAIPrompt] Flow nodes:', flowNodes.length);
        console.log('[handleAIPrompt] Flow edges:', flowEdges.length);
        console.log('[handleAIPrompt] First node:', JSON.stringify(flowNodes[0]));

        store.setNodes(flowNodes);
        store.setEdges(flowEdges);

        console.log('[handleAIPrompt] Immediately after setNodes, store.nodes.length:', store.nodes.length);
        console.log('[handleAIPrompt] store.nodes:', JSON.stringify(store.nodes).slice(0, 200));
      } else if (response.suggestions && response.suggestions.length > 0) {
        console.log('AI suggestions returned:', response.suggestions);
      } else if (response.error) {
        console.error('AI composition error:', response.error);
      } else {
        console.log('AI composition returned empty result:', response);
      }
    } catch (err) {
      console.error('AI composition failed:', err);
    } finally {
      store.setIsLoading(false);
    }
  }, [store]);

  // Determine which right panel to show
  const showRightPanel = rightPanel !== null || selectedNodeId !== null;
  const activeRightPanel = selectedNodeId ? 'inspector' : rightPanel;

  return (
    <div 
      ref={containerRef}
      className={cn(
        "flex flex-col h-screen bg-[var(--bg-primary)] overflow-hidden",
        isFullscreen && "fixed inset-0 z-50",
        presentationMode && "presentation-mode"
      )}
    >
      {/* Top Bar */}
      {!presentationMode && (
        <header className="h-14 border-b border-[var(--border-subtle)] bg-[var(--bg-secondary)] flex items-center justify-between px-4 shrink-0">
          {/* Left: Logo & Graph Name */}
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-brand-500 to-purple-600 flex items-center justify-center">
                <Share className="w-4 h-4 text-white" />
              </div>
              <span className="font-semibold text-[var(--text-primary)]">FRG Editor</span>
            </div>
            <Separator orientation="vertical" className="h-6" />
            <div className="flex items-center gap-2">
              <Input
                value={graphName}
                onChange={(e) => setGraphName(e.target.value)}
                className="h-8 w-48 bg-transparent border-0 focus-visible:ring-0 px-0 text-[var(--text-primary)] font-medium"
                placeholder="Graph name..."
              />
              {isDirty && (
                <Badge variant="secondary" className="text-xs">Unsaved</Badge>
              )}
            </div>
          </div>

          {/* Center: Execution Controls */}
          <div className="flex items-center gap-2">
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={undo}
                    disabled={!canUndo}
                  >
                    <Undo2 className="w-4 h-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Undo (Ctrl+Z)</TooltipContent>
              </Tooltip>
            </TooltipProvider>
            
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={redo}
                    disabled={!canRedo}
                  >
                    <Redo2 className="w-4 h-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Redo (Ctrl+Shift+Z)</TooltipContent>
              </Tooltip>
            </TooltipProvider>

            <Separator orientation="vertical" className="h-6 mx-2" />

            <Button
              variant={editorMode === 'debug' ? 'default' : 'outline'}
              size="sm"
              onClick={() => setEditorMode(editorMode === 'debug' ? 'edit' : 'debug')}
            >
              <Bug className="w-4 h-4 mr-2" />
              Debug
            </Button>

            <Separator orientation="vertical" className="h-6 mx-2" />

            {/* Run Controls */}
            {executionStatus === 'running' ? (
              <>
                <Button variant="outline" size="sm" onClick={pauseExecution}>
                  <Pause className="w-4 h-4 mr-2" />
                  Pause
                </Button>
                <Button variant="destructive" size="sm" onClick={handleStop}>
                  <Square className="w-4 h-4 mr-2" />
                  Stop
                </Button>
              </>
            ) : executionStatus === 'paused' ? (
              <>
                <Button variant="default" size="sm" onClick={handleRun}>
                  <Play className="w-4 h-4 mr-2" />
                  Resume
                </Button>
              </>
            ) : (
              <Button 
                variant="default" 
                size="sm" 
                onClick={handleRun}
                className="bg-gradient-to-r from-green-500 to-emerald-500"
              >
                <Play className="w-4 h-4 mr-2" />
                Run
              </Button>
            )}

            {executionStatus === 'running' && (
              <div className="flex items-center gap-2 ml-2">
                <Loader2 className="w-4 h-4 animate-spin text-brand-500" />
                <span className="text-xs text-[var(--text-secondary)]">Running...</span>
              </div>
            )}
          </div>

          {/* Right: Actions */}
          <div className="flex items-center gap-2">
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => setShowCollaboration(!showCollaboration)}
                    className={cn(showCollaboration && "text-brand-500")}
                  >
                    <Users className="w-4 h-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Toggle collaboration cursors</TooltipContent>
              </Tooltip>
            </TooltipProvider>

            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => setShowShortcuts(true)}
                  >
                    <Keyboard className="w-4 h-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Keyboard shortcuts (?)</TooltipContent>
              </Tooltip>
            </TooltipProvider>

            <Button variant="outline" size="sm">
              <Upload className="w-4 h-4 mr-2" />
              Import
            </Button>
            <Button variant="outline" size="sm">
              <Download className="w-4 h-4 mr-2" />
              Export
            </Button>
            {store.isOffline && operationQueue.length > 0 && (
              <Badge variant="secondary" className="text-xs">
                {operationQueue.length} queued
              </Badge>
            )}
            <Button variant="default" size="sm" className={cn(isDirty && "animate-pulse")} onClick={handleSave} disabled={isLoading || isSaving}>
              <Save className="w-4 h-4 mr-2" />
              {isSaving ? 'Saving...' : isLoading ? 'Loading...' : 'Save'}
            </Button>
            {lastSavedAt && (
              <span className="text-xs text-[var(--text-muted)]">
                Saved {new Date(lastSavedAt).toLocaleTimeString()}
              </span>
            )}
            <Separator orientation="vertical" className="h-6 mx-1" />
            <Button variant="ghost" size="icon" onClick={toggleFullscreen}>
              {isFullscreen ? <Minimize2 className="w-4 h-4" /> : <Maximize2 className="w-4 h-4" />}
            </Button>
            <Button variant="ghost" size="icon">
              <Settings className="w-4 h-4" />
            </Button>
          </div>
        </header>
      )}

      {/* Main Content */}
      <div className="flex flex-1 overflow-hidden">
        {/* Left Sidebar */}
        {!presentationMode && leftPanel && (
          <aside className="w-72 border-r border-[var(--border-subtle)] bg-[var(--bg-secondary)] flex flex-col shrink-0 overflow-hidden">
            <Tabs value={leftPanel} className="w-full">
              <TabsList className="w-full grid grid-cols-4 h-10">
                <TabsTrigger 
                  value="library" 
                  onClick={() => toggleLeftPanel('library')}
                  className="text-xs"
                >
                  <Library className="w-3 h-3 mr-1" />
                  Library
                </TabsTrigger>
                <TabsTrigger 
                  value="ai" 
                  onClick={() => toggleLeftPanel('ai')}
                  className="text-xs"
                >
                  <Sparkles className="w-3 h-3 mr-1" />
                  AI
                </TabsTrigger>
                <TabsTrigger 
                  value="versions" 
                  onClick={() => toggleLeftPanel('versions')}
                  className="text-xs"
                >
                  <GitBranch className="w-3 h-3 mr-1" />
                  Versions
                </TabsTrigger>
                <TabsTrigger 
                  value="evolution" 
                  onClick={() => toggleLeftPanel('evolution')}
                  className="text-xs"
                >
                  <Wand2 className="w-3 h-3 mr-1" />
                  Evolve
                </TabsTrigger>
              </TabsList>
              
              <TabsContent value="library" className="m-0 flex-1 overflow-hidden">
                <FunctionLibrary />
              </TabsContent>
              <TabsContent value="ai" className="m-0 flex-1 overflow-hidden">
                <AIAssistantPanel />
              </TabsContent>
              <TabsContent value="versions" className="m-0 flex-1 overflow-hidden">
                <VersionSelector />
              </TabsContent>
              <TabsContent value="evolution" className="m-0 flex-1 overflow-hidden">
                <EvolutionPanel />
              </TabsContent>
            </Tabs>
          </aside>
        )}

        {/* Toggle Left Panel Button (when hidden) */}
        {!presentationMode && !leftPanel && (
          <div className="absolute left-4 top-[4.5rem] z-10">
            <Button
              variant="secondary"
              size="sm"
              onClick={() => toggleLeftPanel('library')}
              className="shadow-lg"
            >
              <PanelLeft className="w-4 h-4 mr-2" />
              Show Library
            </Button>
          </div>
        )}

        {/* Graph Canvas */}
        <main className="flex-1 relative overflow-hidden">
          <GraphCanvas
            nodeTypes={nodeTypes}
            edgeTypes={edgeTypes}
          />
          
          {/* Empty State Overlay */}
          {isNewGraph && nodes.length === 0 && (
            <EmptyStateOverlay
              onTemplateSelect={handleTemplateSelect}
              onAIPrompt={handleAIPrompt}
            />
          )}
          
          {/* Execution Overlay */}
          <ExecutionOverlay />
          
          {/* Collaboration Cursors */}
          {showCollaboration && <LiveCursors />}
          
          {/* Comment Pins */}
          {showCollaboration && <CommentPins />}
          
          {/* Canvas Controls */}
          <CanvasControls
            onPresentationModeToggle={() => setPresentationMode(!presentationMode)}
            presentationMode={presentationMode}
          />
          
          {/* React Flow Controls */}
          <Controls className="!bg-[var(--bg-secondary)] !border-[var(--border-subtle)]" />
          <MiniMap 
            className="!bg-[var(--bg-secondary)] !border-[var(--border-subtle)]"
            maskColor="rgba(0,0,0,0.2)"
            nodeColor={(node) => {
              if (node.selected) return '#8b5cf6';
              return '#6366f1';
            }}
          />
          <Background 
            variant={BackgroundVariant.Dots} 
            gap={20} 
            size={1}
            color="var(--border-subtle)"
          />
        </main>

        {/* Right Sidebar */}
        {!presentationMode && showRightPanel && (
          <aside className="w-80 border-l border-[var(--border-subtle)] bg-[var(--bg-secondary)] flex flex-col shrink-0 overflow-hidden">
            <Tabs value={activeRightPanel || 'inspector'} className="w-full h-full flex flex-col">
              <TabsList className="w-full grid grid-cols-3 h-10">
                <TabsTrigger 
                  value="inspector" 
                  className="text-xs"
                >
                  Inspector
                </TabsTrigger>
                <TabsTrigger 
                  value="test" 
                  onClick={() => toggleRightPanel('test')}
                  className="text-xs"
                >
                  <TestTube className="w-3 h-3 mr-1" />
                  Test
                </TabsTrigger>
                <TabsTrigger 
                  value="settings" 
                  className="text-xs"
                >
                  <Settings className="w-3 h-3" />
                </TabsTrigger>
              </TabsList>
              
              <TabsContent value="inspector" className="m-0 flex-1 overflow-hidden">
                <NodeInspector />
              </TabsContent>
              <TabsContent value="test" className="m-0 flex-1 overflow-hidden">
                <TestRunnerPanel />
              </TabsContent>
            </Tabs>
          </aside>
        )}

        {/* Toggle Right Panel Button (when hidden) */}
        {!presentationMode && !showRightPanel && (
          <div className="absolute right-4 top-[4.5rem] z-10">
            <Button
              variant="secondary"
              size="sm"
              onClick={() => toggleRightPanel('test')}
              className="shadow-lg"
            >
              Show Inspector
              <PanelRight className="w-4 h-4 ml-2" />
            </Button>
          </div>
        )}
      </div>

      {/* Bottom Execution Panel */}
      {!presentationMode && (
        <CollapsibleExecutionPanel 
          onRun={handleRun}
          onStop={handleStop}
        />
      )}

      {/* Keyboard Shortcuts Help Modal */}
      <KeyboardShortcutsHelp 
        open={showShortcuts} 
        onOpenChange={setShowShortcuts} 
      />
    </div>
  );
}

// Main page component with provider
export default function FRGEditorPage() {
  return (
    <ReactFlowProvider>
      <FRGEditorInner />
    </ReactFlowProvider>
  );
}
