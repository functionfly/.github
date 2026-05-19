/**
 * @functionfly/ui-memory
 * Memory Systems Components - AI memory and knowledge management
 */

// ============================================================================
// Memory Graph
// ============================================================================

export interface MemoryNode {
  id: string;
  type: 'concept' | 'entity' | 'event' | 'document' | 'code' | 'conversation' | 'agent';
  label: string;
  content?: string;
  timestamp: number;
  importance: number;
  connections?: string[];
  metadata?: Record<string, unknown>;
  embedding?: number[];
  parent?: string;
  children?: string[];
}

export interface MemoryEdge {
  id: string;
  source: string;
  target: string;
  type: 'references' | 'derives_from' | 'related_to' | 'part_of' | 'evolved_from' | 'associated_with';
  weight?: number;
  timestamp?: number;
}

export interface MemoryGraphProps {
  nodes: MemoryNode[];
  edges: MemoryEdge[];
  selectedNodeId?: string | null;
  highlightedNodes?: string[];
  onNodeSelect?: (node: MemoryNode) => void;
  onNodeHover?: (node: MemoryNode | null) => void;
  onEdgeClick?: (edge: MemoryEdge) => void;
  layout?: 'force' | 'tree' | 'circular' | 'grid';
  className?: string;
}

// ============================================================================
// Semantic Memory Viewer
// ============================================================================

export interface SemanticMemoryEntry {
  id: string;
  content: string;
  embedding: number[];
  semanticType: 'fact' | 'procedure' | 'preference' | 'context' | 'relationship';
  confidence: number;
  source?: string;
  timestamp: number;
  lastAccessed?: number;
  accessCount?: number;
  tags?: string[];
  linkedMemories?: string[];
}

export interface SemanticMemoryViewerProps {
  entries: SemanticMemoryEntry[];
  selectedEntryId?: string | null;
  searchQuery?: string;
  filterType?: SemanticMemoryEntry['semanticType'] | 'all';
  onEntrySelect?: (entry: SemanticMemoryEntry) => void;
  onSearch?: (query: string) => void;
  onEntryDelete?: (entryId: string) => void;
  className?: string;
}

// ============================================================================
// Long Term Context Explorer
// ============================================================================

export interface ContextChunk {
  id: string;
  content: string;
  timestamp: number;
  importance: number;
  decayScore?: number;
  retentionPriority?: 'critical' | 'high' | 'medium' | 'low';
  retrievalCount?: number;
  associatedGoals?: string[];
  vector?: number[];
}

export interface LongTermContextExplorerProps {
  chunks: ContextChunk[];
  selectedChunkId?: string | null;
  focusArea?: 'recent' | 'important' | 'decaying' | 'goals';
  onChunkSelect?: (chunk: ContextChunk) => void;
  onChunkExpand?: (chunkId: string) => void;
  onMemoryReinforce?: (chunkId: string) => void;
  className?: string;
}

// ============================================================================
// Memory Recall Timeline
// ============================================================================

export interface MemoryRecallEvent {
  id: string;
  timestamp: number;
  type: 'retrieval' | 'reinforcement' | 'decay' | 'consolidation' | 'transfer';
  memoryId: string;
  memoryLabel: string;
  strength: number;
  context?: string;
}

export interface MemoryRecallTimelineProps {
  events: MemoryRecallEvent[];
  selectedEventId?: string | null;
  timeRange?: { start: number; end: number };
  onEventSelect?: (event: MemoryRecallEvent) => void;
  onEventHover?: (event: MemoryRecallEvent | null) => void;
  className?: string;
}

// ============================================================================
// Knowledge Cluster Map
// ============================================================================

export interface KnowledgeCluster {
  id: string;
  name: string;
  description?: string;
  centralTopic: string;
  members: string[];
  importance: number;
  coherence?: number;
  creationTimestamp: number;
  lastModified?: number;
  tags?: string[];
}

export interface KnowledgeClusterMapProps {
  clusters: KnowledgeCluster[];
  selectedClusterId?: string | null;
  highlightedClusters?: string[];
  onClusterSelect?: (cluster: KnowledgeCluster) => void;
  onClusterHover?: (cluster: KnowledgeCluster | null) => void;
  onClusterMerge?: (clusterIds: string[]) => void;
  onClusterSplit?: (clusterId: string) => void;
  className?: string;
}

// ============================================================================
// Memory Decay Visualizer
// ============================================================================

export interface DecayNode {
  id: string;
  label: string;
  age: number;
  strength: number;
  decayRate: number;
  type: 'episodic' | 'semantic' | 'procedural';
  lastReinforced?: number;
  nextDecay?: number;
}

export interface MemoryDecayVisualizerProps {
  nodes: DecayNode[];
  selectedNodeId?: string | null;
  showPredictions?: boolean;
  onNodeSelect?: (node: DecayNode) => void;
  onReinforce?: (nodeId: string) => void;
  className?: string;
}

// ============================================================================
// Vector Embedding Explorer
// ============================================================================

export interface EmbeddingVector {
  id: string;
  values: number[];
  label: string;
  dimension: number;
  magnitude?: number;
  timestamp?: number;
  source?: string;
  neighbors?: string[];
}

export interface VectorEmbeddingExplorerProps {
  vectors: EmbeddingVector[];
  selectedVectorId?: string | null;
  highlightedVectors?: string[];
  metric?: 'cosine' | 'euclidean' | 'dot';
  onVectorSelect?: (vector: EmbeddingVector) => void;
  onClusterAnalyze?: () => void;
  className?: string;
}

// ============================================================================
// Shared Agent Memory Panel
// ============================================================================

export interface AgentMemory {
  agentId: string;
  agentName: string;
  role?: string;
  memoryCapacity?: number;
  usedCapacity?: number;
  activeMemories?: string[];
  sharedWith?: string[];
  lastActive?: number;
}

export interface SharedAgentMemoryPanelProps {
  agents: AgentMemory[];
  selectedAgentId?: string | null;
  sharedMemoryIds?: string[];
  onAgentSelect?: (agent: AgentMemory) => void;
  onShareMemory?: (memoryId: string, agentIds: string[]) => void;
  onRetrieveShared?: (memoryId: string) => void;
  className?: string;
}

// ============================================================================
// Memory Merge Tool
// ============================================================================

export interface MemoryMergeCandidate {
  id: string;
  sourceMemoryIds: string[];
  targetMemoryId?: string;
  suggestedMerge?: {
    content: string;
    confidence: number;
  };
  overlapScore?: number;
  conflicts?: Array<{
    field: string;
    values: unknown[];
  }>;
}

export interface MemoryMergeToolProps {
  candidates: MemoryMergeCandidate[];
  selectedCandidateId?: string | null;
  onCandidateSelect?: (candidate: MemoryMergeCandidate) => void;
  onMerge?: (candidateId: string) => void;
  onSplit?: (memoryId: string) => void;
  onDiscard?: (candidateId: string) => void;
  className?: string;
}

// ============================================================================
// Conversation Memory Tree
// ============================================================================

export interface ConversationNode {
  id: string;
  type: 'turn' | 'topic' | 'segment' | 'branch';
  content: string;
  speaker?: 'user' | 'agent' | 'system';
  timestamp: number;
  parentId?: string;
  children?: string[];
  summary?: string;
  sentiment?: 'positive' | 'neutral' | 'negative';
  intent?: string;
}

export interface ConversationMemoryTreeProps {
  nodes: ConversationNode[];
  selectedNodeId?: string | null;
  focusedNodeId?: string;
  onNodeSelect?: (node: ConversationNode) => void;
  onBranch?: (nodeId: string, newContent: string) => void;
  onSummarize?: (nodeId: string) => void;
  className?: string;
}

// ============================================================================
// Memory Access Monitor
// ============================================================================

export interface MemoryAccessEntry {
  id: string;
  memoryId: string;
  memoryLabel: string;
  accessType: 'read' | 'write' | 'delete' | 'share';
  timestamp: number;
  agentId?: string;
  agentName?: string;
  duration?: number;
  success: boolean;
  cacheHit?: boolean;
}

export interface MemoryAccessMonitorProps {
  entries: MemoryAccessEntry[];
  selectedEntryId?: string | null;
  timeRange?: { start: number; end: number };
  onEntrySelect?: (entry: MemoryAccessEntry) => void;
  onFilterChange?: (filters: { accessType?: string; agentId?: string }) => void;
  className?: string;
}
