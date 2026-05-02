/**
 * FRG (Function Runtime Graph) TypeScript Types
 * Mirrors the Go backend types from internal/frg/models.go
 */

import { type Node, type Edge } from '@xyflow/react';

// Execution modes
export type ExecutionMode = 'sync' | 'async' | 'streaming' | 'event_driven';

// Edge types for data flow
export type EdgeType = 'sync' | 'async' | 'stream';

// Instance status
export type InstanceStatus = 
  | 'pending' 
  | 'running' 
  | 'streaming' 
  | 'paused' 
  | 'completed' 
  | 'failed';

// Node execution status
export type NodeExecutionStatus = 
  | 'idle'
  | 'pending' 
  | 'executing' 
  | 'waiting' 
  | 'completed' 
  | 'failed' 
  | 'retrying';

// Data mapping between nodes
export interface DataMapping {
  sourcePath?: string;
  targetPath?: string;
  transform?: 'map' | 'filter' | 'reduce' | 'flat' | 'custom';
  script?: string;
}

// Condition for conditional routing
export interface Condition {
  operator: 'eq' | 'ne' | 'gt' | 'lt' | 'contains' | 'regex' | 'exists';
  field: string;
  value: unknown;
}

// Retry policy
export interface RetryPolicy {
  maxAttempts: number;
  initialBackoffMs: number;
  maxBackoffMs: number;
  backoffFactor: number;
  retryableErrors?: string[];
}

// Graph edge definition
export interface GraphEdgeDefinition {
  id: string;
  sourceNodeId: string;
  targetNodeId: string;
  mapping: DataMapping;
  condition?: Condition;
  type: EdgeType;
  retryPolicy?: RetryPolicy;
  fallbackNodeId?: string;
  bufferSize?: number;
}

// Function reference in a graph node
export interface GraphNodeRef {
  nodeId: string;
  author: string;
  name: string;
  version: string;
  config: Record<string, unknown>;
  metadata: Record<string, unknown>;
}

// Graph definition (the blueprint)
export interface GraphDefinition {
  id: string;
  author: string;
  name: string;
  version: string;
  fullName: string;
  nodeRefs: GraphNodeRef[];
  edges: GraphEdgeDefinition[];
  executionMode: ExecutionMode;
  triggerConfig?: Record<string, unknown>;
  inputSchema?: Record<string, unknown>;
  outputSchema?: Record<string, unknown>;
  aiDescription?: string;
  compositionScore: number;
  trustScore: number;
  deterministic: boolean;
  forkedFromAuthor?: string;
  forkedFromName?: string;
  forkedFromVersion?: string;
  tenantId?: string;
  ownerUserId?: string;
  visibility: 'public' | 'private' | 'team';
  pricingType: 'free' | 'pay_per_use' | 'subscription';
  basePrice: number;
  revenueShare: number;
  createdAt: string;
  updatedAt: string;
  publishedAt?: string;
}

// Node state during execution
export interface NodeState {
  status: NodeExecutionStatus;
  output?: unknown;
  error?: string;
  attemptCount: number;
  durationMs: number;
  execCertId?: string;
}

// Graph instance (live execution)
export interface GraphInstance {
  id: string;
  definitionId: string;
  status: InstanceStatus;
  inputData?: unknown;
  outputData?: unknown;
  errorMessage?: string;
  frozenNodes?: GraphNodeRef[];
  frozenEdges?: GraphEdgeDefinition[];
  nodeStates?: Record<string, NodeState>;
  currentExecutionOrder?: string[];
  executionRootHash?: string;
  megRecordId?: string;
  eventStreamId?: string;
  lastEventAt?: string;
  stateNamespace?: string;
  totalDurationMs: number;
  totalComputeUnits: number;
  createdAt: string;
  startedAt?: string;
  completedAt?: string;
}

// Runtime node state for UI
export interface RuntimeNodeState {
  status: NodeExecutionStatus;
  output?: unknown;
  error?: string;
  attemptCount: number;
  durationMs: number;
  isActive: boolean;
  progress?: number;
  generatedCode?: string; // Source code from AI generation
}

// Runtime edge state for UI
export interface RuntimeEdgeState {
  status: 'idle' | 'active' | 'completed' | 'error';
  recordsTransferred: number;
  bytesTransferred: number;
  isDataFlowing: boolean;
  flowProgress: number;
}

// Graph event for live updates
export interface GraphEvent {
  id: string;
  instanceId: string;
  eventType: 'trigger' | 'complete' | 'error' | 'stream' | 'checkpoint' | 'retry' | 'node_start' | 'node_complete' | 'edge_data';
  nodeId?: string;
  payload?: unknown;
  contentType: string;
  sequenceNum: number;
  timestamp: string;
  inputHash?: string;
  outputHash?: string;
}

// Optimization suggestion
export interface GraphOptimizationSuggestion {
  id: string;
  definitionId: string;
  suggestionType: 'parallel' | 'cache' | 'runtime' | 'replacement' | 'structure';
  description: string;
  estimatedImpact: number;
  actionConfig: Record<string, unknown>;
  aiConfidence: number;
  generatedAt: string;
  dismissed: boolean;
  applied: boolean;
}

// Function catalog item for library
export interface FunctionCatalogItem {
  id: string;
  author: string;
  name: string;
  version: string;
  description: string;
  category: string;
  tags: string[];
  inputSchema: Record<string, unknown>;
  outputSchema: Record<string, unknown>;
  trustScore: number;
  usageCount: number;
  avgExecutionTimeMs: number;
  icon?: string;
  color?: string;
}

// React Flow node data
export interface FunctionNodeData {
  [key: string]: unknown;
  functionRef: GraphNodeRef;
  runtimeState?: RuntimeNodeState;
  isSelected: boolean;
  isEditable: boolean;
  onConfigChange?: (config: Record<string, unknown>) => void;
  onRun?: () => void;
  onTest?: () => void;
}

// React Flow edge data
export interface FlowEdgeData {
  [key: string]: unknown;
  mapping: DataMapping;
  condition?: Condition;
  retryPolicy?: RetryPolicy;
  runtimeState?: RuntimeEdgeState;
  isValid: boolean;
  validationError?: string;
}

// Extended React Flow types
export type FRGNode = Node<FunctionNodeData>;
export type FRGEdge = Edge<FlowEdgeData>;

// Graph execution result
export interface ExecutionResult {
  instanceId: string;
  status: InstanceStatus;
  output?: unknown;
  error?: string;
  nodeResults: Record<string, {
    status: NodeExecutionStatus;
    output?: unknown;
    error?: string;
    durationMs: number;
    certId?: string;
  }>;
  executionRootHash?: string;
  durationMs: number;
  computeUnits: number;
}

// AI suggestion for workflow building
export interface AISuggestion {
  id: string;
  type: 'add_node' | 'connect' | 'replace' | 'optimize' | 'fix';
  description: string;
  confidence: number;
  affectedNodes: string[];
  action: {
    type: string;
    payload: Record<string, unknown>;
  };
  explanation: string;
}

// Test case for function testing
export interface TestCase {
  id: string;
  name: string;
  input: Record<string, unknown>;
  expectedOutput?: Record<string, unknown>;
  actualOutput?: Record<string, unknown>;
  status: 'pending' | 'running' | 'passed' | 'failed';
  durationMs?: number;
  logs?: string[];
  error?: string;
}

// Version comparison
export interface VersionComparison {
  version: string;
  changes: {
    type: 'added' | 'removed' | 'modified';
    path: string;
    oldValue?: unknown;
    newValue?: unknown;
  }[];
  summary: string;
}

// Smart connection suggestion
export interface SmartConnection {
  sourceNodeId: string;
  suggestedTargetIds: string[];
  compatibilityScore: number;
  reason: string;
  autoConnectable: boolean;
}

// Live execution metrics
export interface LiveExecutionMetrics {
  instanceId: string;
  activeNodes: string[];
  dataFlowRate: number;
  throughputBps: number;
  latencyMs: number;
  errorRate: number;
  cpuPercent: number;
  memoryMb: number;
}
