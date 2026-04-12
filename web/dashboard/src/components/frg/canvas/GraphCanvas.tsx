/**
 * GraphCanvas Component
 * Core React Flow canvas with node/edge rendering, zoom/pan, drag-drop
 */

import { useCallback, useRef, useEffect } from 'react';
import {
  ReactFlow,
  useNodesState,
  useEdgesState,
  useReactFlow,
  addEdge,
  applyNodeChanges,
  applyEdgeChanges,
  type NodeChange,
  type EdgeChange,
  type Connection,
  type Viewport,
  Background,
  BackgroundVariant,
  type NodeTypes,
  type EdgeTypes,
  useOnViewportChange,
  type Node,
  type Edge,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';

import { useFRGStore } from '@/stores/frgStore';
import type { FRGNode, FRGEdge } from '@/types/frg';

interface GraphCanvasProps {
  nodeTypes?: NodeTypes;
  edgeTypes?: EdgeTypes;
}

export function GraphCanvas({ nodeTypes, edgeTypes }: GraphCanvasProps) {
  const { getIntersectingNodes, screenToFlowPosition, setViewport } = useReactFlow();
  
  const store = useFRGStore();
  const {
    nodes: storeNodes,
    edges: storeEdges,
    selectedNodeId,
    editorMode,
    setNodes: setStoreNodes,
    setEdges: setStoreEdges,
    setSelectedNode,
    markDirty,
    addNode,
  } = store;

  // Convert store state to React Flow state
  const [nodes, setNodes, onNodesChange] = useNodesState(storeNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(storeEdges);

  // Sync with store
  useEffect(() => {
    setNodes(storeNodes);
  }, [storeNodes, setNodes]);

  useEffect(() => {
    setEdges(storeEdges);
  }, [storeEdges, setEdges]);

  // Handle node changes
  const handleNodesChange = useCallback((changes: NodeChange<FRGNode>[]) => {
    const nextNodes = applyNodeChanges(changes, nodes);
    setNodes(nextNodes as FRGNode[]);
    setStoreNodes(nextNodes as FRGNode[]);
    markDirty();
  }, [nodes, setNodes, setStoreNodes, markDirty]);

  // Handle edge changes
  const handleEdgesChange = useCallback((changes: EdgeChange<FRGEdge>[]) => {
    const nextEdges = applyEdgeChanges(changes, edges);
    setEdges(nextEdges as FRGEdge[]);
    setStoreEdges(nextEdges as FRGEdge[]);
    markDirty();
  }, [edges, setEdges, setStoreEdges, markDirty]);

  // Handle connections
  const handleConnect = useCallback((connection: Connection) => {
    const newEdge: FRGEdge = {
      id: `e-${connection.source}-${connection.target}-${Date.now()}`,
      source: connection.source ?? '',
      target: connection.target ?? '',
      sourceHandle: connection.sourceHandle,
      targetHandle: connection.targetHandle,
      type: 'custom',
      animated: false,
      data: {
        mapping: { sourcePath: '*', targetPath: '*' },
        isValid: true,
        runtimeState: {
          status: 'idle',
          recordsTransferred: 0,
          bytesTransferred: 0,
          isDataFlowing: false,
          flowProgress: 0,
        },
      },
    };
    const nextEdges = addEdge(newEdge, edges);
    setEdges(nextEdges as FRGEdge[]);
    setStoreEdges(nextEdges as FRGEdge[]);
    markDirty();
  }, [edges, setEdges, setStoreEdges, markDirty]);

  // Handle node click
  const handleNodeClick = useCallback((_: React.MouseEvent, node: Node) => {
    setSelectedNode(node.id);
  }, [setSelectedNode]);

  // Handle pane click (deselect)
  const handlePaneClick = useCallback(() => {
    setSelectedNode(null);
  }, [setSelectedNode]);

  // Handle drop from sidebar
  const handleDrop = useCallback(
    (event: React.DragEvent<HTMLDivElement>) => {
      event.preventDefault();

      const type = event.dataTransfer.getData('application/reactflow');
      const functionData = event.dataTransfer.getData('application/functionfly-function');

      if (!type || !functionData) return;

      try {
        const fn = JSON.parse(functionData);
        const position = screenToFlowPosition({
          x: event.clientX,
          y: event.clientY,
        });

        const nodeId = `node-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
        const newNode: FRGNode = {
          id: nodeId,
          type: 'functionNode',
          position,
          data: {
            functionRef: {
              nodeId: nodeId,
              author: fn.author ?? 'unknown',
              name: fn.name ?? 'unnamed',
              version: fn.version || 'latest',
              config: {},
              metadata: {},
            },
            isSelected: false,
            isEditable: true,
          },
        };

        addNode(newNode);
      } catch (error) {
        console.error('Error parsing function data:', error);
      }
    },
    [screenToFlowPosition, addNode]
  );

  // Handle drag over
  const handleDragOver = useCallback((event: React.DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
  }, []);

  // Viewport change
  useOnViewportChange({
    onEnd: (viewport: Viewport) => {
      store.setViewport(viewport);
    },
  });

  // Selection state update
  useEffect(() => {
    setNodes((nds) =>
      nds.map((node) => ({
        ...node,
        selected: node.id === selectedNodeId,
      }))
    );
  }, [selectedNodeId, setNodes]);

  // Enable/disable interactions based on editor mode
  const isInteractive = editorMode === 'edit' || editorMode === 'debug';

  return (
    <ReactFlow
      nodes={nodes}
      edges={edges}
      onNodesChange={handleNodesChange}
      onEdgesChange={handleEdgesChange}
      onConnect={handleConnect}
      onNodeClick={handleNodeClick}
      onPaneClick={handlePaneClick}
      onDrop={handleDrop}
      onDragOver={handleDragOver}
      nodeTypes={nodeTypes}
      edgeTypes={edgeTypes}
      fitView
      fitViewOptions={{ padding: 0.2 }}
      attributionPosition="bottom-left"
      deleteKeyCode={isInteractive ? 'Delete' : null}
      selectionKeyCode={isInteractive ? 'Shift' : null}
      multiSelectionKeyCode={isInteractive ? 'Meta' : null}
      nodesDraggable={isInteractive}
      nodesConnectable={isInteractive}
      elementsSelectable={isInteractive}
      zoomOnScroll={true}
      zoomOnPinch={true}
      panOnScroll={true}
      panOnDrag={true}
      selectionOnDrag={true}
      selectNodesOnDrag={false}
      style={{
        background: 'var(--bg-primary)',
      }}
    >
      <Background 
        variant={BackgroundVariant.Dots}
        gap={20}
        size={1}
        color="var(--border-subtle)"
        style={{ opacity: 0.5 }}
      />
    </ReactFlow>
  );
}
