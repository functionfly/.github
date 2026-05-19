/**
 * Memory Store
 * Global state management for AI memory systems components
 */

import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'

// ============================================================================
// Types
// ============================================================================

export interface MemoryNode {
  id: string
  type: 'concept' | 'entity' | 'event' | 'document' | 'code' | 'conversation' | 'agent'
  label: string
  content?: string
  timestamp: number
  importance: number
  connections?: string[]
  metadata?: Record<string, unknown>
  embedding?: number[]
  parent?: string
  children?: string[]
}

export interface MemoryEdge {
  id: string
  source: string
  target: string
  type: 'references' | 'derives_from' | 'related_to' | 'part_of' | 'evolved_from' | 'associated_with'
  weight?: number
  timestamp?: number
}

export interface SemanticMemoryEntry {
  id: string
  content: string
  embedding: number[]
  semanticType: 'fact' | 'procedure' | 'preference' | 'context' | 'relationship'
  confidence: number
  source?: string
  timestamp: number
  lastAccessed?: number
  accessCount?: number
  tags?: string[]
  linkedMemories?: string[]
}

export interface ContextChunk {
  id: string
  content: string
  timestamp: number
  importance: number
  decayScore?: number
  retentionPriority?: 'critical' | 'high' | 'medium' | 'low'
  retrievalCount?: number
  associatedGoals?: string[]
  vector?: number[]
}

export interface MemoryRecallEvent {
  id: string
  timestamp: number
  type: 'retrieval' | 'reinforcement' | 'decay' | 'consolidation' | 'transfer'
  memoryId: string
  memoryLabel: string
  strength: number
  context?: string
}

export interface KnowledgeCluster {
  id: string
  name: string
  description?: string
  centralTopic: string
  members: string[]
  importance: number
  coherence?: number
  creationTimestamp: number
  lastModified?: number
  tags?: string[]
}

export interface DecayNode {
  id: string
  label: string
  age: number
  strength: number
  decayRate: number
  type: 'episodic' | 'semantic' | 'procedural'
  lastReinforced?: number
  nextDecay?: number
}

export interface EmbeddingVector {
  id: string
  values: number[]
  label: string
  dimension: number
  magnitude?: number
  timestamp?: number
  source?: string
  neighbors?: string[]
}

export interface AgentMemory {
  agentId: string
  agentName: string
  role?: string
  memoryCapacity?: number
  usedCapacity?: number
  activeMemories?: string[]
  sharedWith?: string[]
  lastActive?: number
}

export interface MemoryMergeCandidate {
  id: string
  sourceMemoryIds: string[]
  targetMemoryId?: string
  suggestedMerge?: {
    content: string
    confidence: number
  }
  overlapScore?: number
  conflicts?: Array<{
    field: string
    values: unknown[]
  }>
}

export interface ConversationNode {
  id: string
  type: 'turn' | 'topic' | 'segment' | 'branch'
  content: string
  speaker?: 'user' | 'agent' | 'system'
  timestamp: number
  parentId?: string
  children?: string[]
  summary?: string
  sentiment?: 'positive' | 'neutral' | 'negative'
  intent?: string
}

export interface MemoryAccessEntry {
  id: string
  memoryId: string
  memoryLabel: string
  accessType: 'read' | 'write' | 'delete' | 'share'
  timestamp: number
  agentId?: string
  agentName?: string
  duration?: number
  success: boolean
  cacheHit?: boolean
}

export interface MemoryAccessMonitorEntry {
  id: string
  memoryId: string
  memoryLabel: string
  accessType: 'read' | 'write' | 'delete' | 'share'
  timestamp: number
  agentId?: string
  agentName?: string
  duration?: number
  success: boolean
  cacheHit?: boolean
}

// ============================================================================
// State Interface
// ============================================================================

export interface MemoryState {
  // Memory Graph
  graphNodes: MemoryNode[]
  graphEdges: MemoryEdge[]
  selectedGraphNodeId: string | null
  highlightedGraphNodes: string[]
  graphLayout: 'force' | 'tree' | 'circular' | 'grid'

  // Semantic Memory
  semanticEntries: SemanticMemoryEntry[]
  selectedSemanticEntryId: string | null
  semanticSearchQuery: string
  semanticFilterType: 'all' | 'fact' | 'procedure' | 'preference' | 'context' | 'relationship'

  // Long Term Context
  contextChunks: ContextChunk[]
  selectedContextChunkId: string | null
  focusArea: 'recent' | 'important' | 'decaying' | 'goals'

  // Memory Recall Timeline
  recallEvents: MemoryRecallEvent[]
  selectedRecallEventId: string | null
  recallTimeRange: { start: number; end: number } | null

  // Knowledge Clusters
  clusters: KnowledgeCluster[]
  selectedClusterId: string | null
  highlightedClusters: string[]

  // Memory Decay
  decayNodes: DecayNode[]
  selectedDecayNodeId: string | null
  showDecayPredictions: boolean

  // Vector Embeddings
  vectors: EmbeddingVector[]
  selectedVectorId: string | null
  highlightedVectors: string[]
  vectorMetric: 'cosine' | 'euclidean' | 'dot'

  // Shared Agent Memory
  agents: AgentMemory[]
  selectedAgentId: string | null
  sharedMemoryIds: string[]

  // Memory Merge
  mergeCandidates: MemoryMergeCandidate[]
  selectedMergeCandidateId: string | null
  showMergeConflicts: boolean

  // Conversation Memory Tree
  conversationNodes: ConversationNode[]
  selectedConversationNodeId: string | null
  focusedConversationNodeId: string | null

  // Memory Access Monitor
  accessEntries: MemoryAccessEntry[]
  selectedAccessEntryId: string | null
  accessTimeRange: { start: number; end: number } | null
  accessTypeFilter: string | null
  accessAgentFilter: string | null

  // UI State
  activePanel: string
  sidebarCollapsed: boolean
}

// ============================================================================
// Store
// ============================================================================

interface MemoryActions {
  // Memory Graph Actions
  setGraphNodes: (nodes: MemoryNode[]) => void
  setGraphEdges: (edges: MemoryEdge[]) => void
  selectGraphNode: (nodeId: string | null) => void
  setHighlightedGraphNodes: (nodeIds: string[]) => void
  setGraphLayout: (layout: 'force' | 'tree' | 'circular' | 'grid') => void
  addGraphNode: (node: MemoryNode) => void
  updateGraphNode: (nodeId: string, updates: Partial<MemoryNode>) => void
  removeGraphNode: (nodeId: string) => void
  addGraphEdge: (edge: MemoryEdge) => void

  // Semantic Memory Actions
  setSemanticEntries: (entries: SemanticMemoryEntry[]) => void
  selectSemanticEntry: (entryId: string | null) => void
  setSemanticSearchQuery: (query: string) => void
  setSemanticFilterType: (type: 'all' | 'fact' | 'procedure' | 'preference' | 'context' | 'relationship') => void
  deleteSemanticEntry: (entryId: string) => void
  reinforceSemanticEntry: (entryId: string) => void

  // Long Term Context Actions
  setContextChunks: (chunks: ContextChunk[]) => void
  selectContextChunk: (chunkId: string | null) => void
  setFocusArea: (area: 'recent' | 'important' | 'decaying' | 'goals') => void
  reinforceContextChunk: (chunkId: string) => void

  // Memory Recall Actions
  setRecallEvents: (events: MemoryRecallEvent[]) => void
  selectRecallEvent: (eventId: string | null) => void
  setRecallTimeRange: (range: { start: number; end: number } | null) => void

  // Knowledge Cluster Actions
  setClusters: (clusters: KnowledgeCluster[]) => void
  selectCluster: (clusterId: string | null) => void
  setHighlightedClusters: (clusterIds: string[]) => void
  mergeClusters: (clusterIds: string[]) => void
  splitCluster: (clusterId: string) => void

  // Memory Decay Actions
  setDecayNodes: (nodes: DecayNode[]) => void
  selectDecayNode: (nodeId: string | null) => void
  setShowDecayPredictions: (show: boolean) => void
  reinforceDecayNode: (nodeId: string) => void

  // Vector Embedding Actions
  setVectors: (vectors: EmbeddingVector[]) => void
  selectVector: (vectorId: string | null) => void
  setHighlightedVectors: (vectorIds: string[]) => void
  setVectorMetric: (metric: 'cosine' | 'euclidean' | 'dot') => void
  analyzeClusters: () => void

  // Shared Agent Memory Actions
  setAgents: (agents: AgentMemory[]) => void
  selectAgent: (agentId: string | null) => void
  shareMemory: (memoryId: string, agentIds: string[]) => void
  retrieveSharedMemory: (memoryId: string) => void

  // Memory Merge Actions
  setMergeCandidates: (candidates: MemoryMergeCandidate[]) => void
  selectMergeCandidate: (candidateId: string | null) => void
  setShowMergeConflicts: (show: boolean) => void
  executeMerge: (candidateId: string) => void
  splitMemory: (memoryId: string) => void
  discardMergeCandidate: (candidateId: string) => void

  // Conversation Memory Tree Actions
  setConversationNodes: (nodes: ConversationNode[]) => void
  selectConversationNode: (nodeId: string | null) => void
  setFocusedConversationNode: (nodeId: string | null) => void
  branchConversation: (nodeId: string, newContent: string) => void
  summarizeConversationNode: (nodeId: string) => void

  // Memory Access Monitor Actions
  setAccessEntries: (entries: MemoryAccessEntry[]) => void
  selectAccessEntry: (entryId: string | null) => void
  setAccessTimeRange: (range: { start: number; end: number } | null) => void
  setAccessTypeFilter: (type: string | null) => void
  setAccessAgentFilter: (agentId: string | null) => void

  // UI Actions
  setActivePanel: (panel: string) => void
  toggleSidebar: () => void
}

export const useMemoryStore = create<MemoryState & MemoryActions>()(
  immer((set) => ({
    // ============================================================================
    // Initial State
    // ============================================================================

    // Memory Graph
    graphNodes: [],
    graphEdges: [],
    selectedGraphNodeId: null,
    highlightedGraphNodes: [],
    graphLayout: 'force',

    // Semantic Memory
    semanticEntries: [],
    selectedSemanticEntryId: null,
    semanticSearchQuery: '',
    semanticFilterType: 'all',

    // Long Term Context
    contextChunks: [],
    selectedContextChunkId: null,
    focusArea: 'important',

    // Memory Recall Timeline
    recallEvents: [],
    selectedRecallEventId: null,
    recallTimeRange: null,

    // Knowledge Clusters
    clusters: [],
    selectedClusterId: null,
    highlightedClusters: [],

    // Memory Decay
    decayNodes: [],
    selectedDecayNodeId: null,
    showDecayPredictions: false,

    // Vector Embeddings
    vectors: [],
    selectedVectorId: null,
    highlightedVectors: [],
    vectorMetric: 'cosine',

    // Shared Agent Memory
    agents: [],
    selectedAgentId: null,
    sharedMemoryIds: [],

    // Memory Merge
    mergeCandidates: [],
    selectedMergeCandidateId: null,
    showMergeConflicts: false,

    // Conversation Memory Tree
    conversationNodes: [],
    selectedConversationNodeId: null,
    focusedConversationNodeId: null,

    // Memory Access Monitor
    accessEntries: [],
    selectedAccessEntryId: null,
    accessTimeRange: null,
    accessTypeFilter: null,
    accessAgentFilter: null,

    // UI State
    activePanel: 'graph',
    sidebarCollapsed: false,

    // ============================================================================
    // Memory Graph Actions
    // ============================================================================

    setGraphNodes: (nodes) =>
      set((state) => {
        state.graphNodes = nodes
      }),

    setGraphEdges: (edges) =>
      set((state) => {
        state.graphEdges = edges
      }),

    selectGraphNode: (nodeId) =>
      set((state) => {
        state.selectedGraphNodeId = nodeId
      }),

    setHighlightedGraphNodes: (nodeIds) =>
      set((state) => {
        state.highlightedGraphNodes = nodeIds
      }),

    setGraphLayout: (layout) =>
      set((state) => {
        state.graphLayout = layout
      }),

    addGraphNode: (node) =>
      set((state) => {
        state.graphNodes.push(node)
      }),

    updateGraphNode: (nodeId, updates) =>
      set((state) => {
        const idx = state.graphNodes.findIndex((n) => n.id === nodeId)
        if (idx !== -1) {
          Object.assign(state.graphNodes[idx], updates)
        }
      }),

    removeGraphNode: (nodeId) =>
      set((state) => {
        state.graphNodes = state.graphNodes.filter((n) => n.id !== nodeId)
        state.graphEdges = state.graphEdges.filter(
          (e) => e.source !== nodeId && e.target !== nodeId
        )
      }),

    addGraphEdge: (edge) =>
      set((state) => {
        state.graphEdges.push(edge)
      }),

    // ============================================================================
    // Semantic Memory Actions
    // ============================================================================

    setSemanticEntries: (entries) =>
      set((state) => {
        state.semanticEntries = entries
      }),

    selectSemanticEntry: (entryId) =>
      set((state) => {
        state.selectedSemanticEntryId = entryId
      }),

    setSemanticSearchQuery: (query) =>
      set((state) => {
        state.semanticSearchQuery = query
      }),

    setSemanticFilterType: (type) =>
      set((state) => {
        state.semanticFilterType = type
      }),

    deleteSemanticEntry: (entryId) =>
      set((state) => {
        state.semanticEntries = state.semanticEntries.filter((e) => e.id !== entryId)
        if (state.selectedSemanticEntryId === entryId) {
          state.selectedSemanticEntryId = null
        }
      }),

    reinforceSemanticEntry: (entryId) =>
      set((state) => {
        const entry = state.semanticEntries.find((e) => e.id === entryId)
        if (entry) {
          entry.accessCount = (entry.accessCount || 0) + 1
          entry.lastAccessed = Date.now()
          entry.confidence = Math.min(1, entry.confidence + 0.05)
        }
      }),

    // ============================================================================
    // Long Term Context Actions
    // ============================================================================

    setContextChunks: (chunks) =>
      set((state) => {
        state.contextChunks = chunks
      }),

    selectContextChunk: (chunkId) =>
      set((state) => {
        state.selectedContextChunkId = chunkId
      }),

    setFocusArea: (area) =>
      set((state) => {
        state.focusArea = area
      }),

    reinforceContextChunk: (chunkId) =>
      set((state) => {
        const chunk = state.contextChunks.find((c) => c.id === chunkId)
        if (chunk) {
          chunk.retrievalCount = (chunk.retrievalCount || 0) + 1
          chunk.decayScore = Math.max(0, (chunk.decayScore || 0) - 0.1)
        }
      }),

    // ============================================================================
    // Memory Recall Actions
    // ============================================================================

    setRecallEvents: (events) =>
      set((state) => {
        state.recallEvents = events
      }),

    selectRecallEvent: (eventId) =>
      set((state) => {
        state.selectedRecallEventId = eventId
      }),

    setRecallTimeRange: (range) =>
      set((state) => {
        state.recallTimeRange = range
      }),

    // ============================================================================
    // Knowledge Cluster Actions
    // ============================================================================

    setClusters: (clusters) =>
      set((state) => {
        state.clusters = clusters
      }),

    selectCluster: (clusterId) =>
      set((state) => {
        state.selectedClusterId = clusterId
      }),

    setHighlightedClusters: (clusterIds) =>
      set((state) => {
        state.highlightedClusters = clusterIds
      }),

    mergeClusters: (clusterIds) =>
      set((state) => {
        // Implementation would merge clusters
        console.log('Merging clusters:', clusterIds)
      }),

    splitCluster: (clusterId) =>
      set((state) => {
        // Implementation would split cluster
        console.log('Splitting cluster:', clusterId)
      }),

    // ============================================================================
    // Memory Decay Actions
    // ============================================================================

    setDecayNodes: (nodes) =>
      set((state) => {
        state.decayNodes = nodes
      }),

    selectDecayNode: (nodeId) =>
      set((state) => {
        state.selectedDecayNodeId = nodeId
      }),

    setShowDecayPredictions: (show) =>
      set((state) => {
        state.showDecayPredictions = show
      }),

    reinforceDecayNode: (nodeId) =>
      set((state) => {
        const node = state.decayNodes.find((n) => n.id === nodeId)
        if (node) {
          node.strength = Math.min(1, node.strength + 0.1)
          node.lastReinforced = Date.now()
        }
      }),

    // ============================================================================
    // Vector Embedding Actions
    // ============================================================================

    setVectors: (vectors) =>
      set((state) => {
        state.vectors = vectors
      }),

    selectVector: (vectorId) =>
      set((state) => {
        state.selectedVectorId = vectorId
      }),

    setHighlightedVectors: (vectorIds) =>
      set((state) => {
        state.highlightedVectors = vectorIds
      }),

    setVectorMetric: (metric) =>
      set((state) => {
        state.vectorMetric = metric
      }),

    analyzeClusters: () =>
      set((state) => {
        // Implementation would run clustering analysis
        console.log('Analyzing vector clusters')
      }),

    // ============================================================================
    // Shared Agent Memory Actions
    // ============================================================================

    setAgents: (agents) =>
      set((state) => {
        state.agents = agents
      }),

    selectAgent: (agentId) =>
      set((state) => {
        state.selectedAgentId = agentId
      }),

    shareMemory: (memoryId, agentIds) =>
      set((state) => {
        // Implementation would share memory with agents
        console.log('Sharing memory:', memoryId, 'with agents:', agentIds)
      }),

    retrieveSharedMemory: (memoryId) =>
      set((state) => {
        // Implementation would retrieve shared memory
        console.log('Retrieving shared memory:', memoryId)
      }),

    // ============================================================================
    // Memory Merge Actions
    // ============================================================================

    setMergeCandidates: (candidates) =>
      set((state) => {
        state.mergeCandidates = candidates
      }),

    selectMergeCandidate: (candidateId) =>
      set((state) => {
        state.selectedMergeCandidateId = candidateId
      }),

    setShowMergeConflicts: (show) =>
      set((state) => {
        state.showMergeConflicts = show
      }),

    executeMerge: (candidateId) =>
      set((state) => {
        state.mergeCandidates = state.mergeCandidates.filter((c) => c.id !== candidateId)
      }),

    splitMemory: (memoryId) =>
      set((state) => {
        // Implementation would split memory
        console.log('Splitting memory:', memoryId)
      }),

    discardMergeCandidate: (candidateId) =>
      set((state) => {
        state.mergeCandidates = state.mergeCandidates.filter((c) => c.id !== candidateId)
      }),

    // ============================================================================
    // Conversation Memory Tree Actions
    // ============================================================================

    setConversationNodes: (nodes) =>
      set((state) => {
        state.conversationNodes = nodes
      }),

    selectConversationNode: (nodeId) =>
      set((state) => {
        state.selectedConversationNodeId = nodeId
      }),

    setFocusedConversationNode: (nodeId) =>
      set((state) => {
        state.focusedConversationNodeId = nodeId
      }),

    branchConversation: (nodeId, newContent) =>
      set((state) => {
        // Implementation would branch conversation
        console.log('Branching conversation:', nodeId, newContent)
      }),

    summarizeConversationNode: (nodeId) =>
      set((state) => {
        // Implementation would summarize conversation node
        console.log('Summarizing conversation node:', nodeId)
      }),

    // ============================================================================
    // Memory Access Monitor Actions
    // ============================================================================

    setAccessEntries: (entries) =>
      set((state) => {
        state.accessEntries = entries
      }),

    selectAccessEntry: (entryId) =>
      set((state) => {
        state.selectedAccessEntryId = entryId
      }),

    setAccessTimeRange: (range) =>
      set((state) => {
        state.accessTimeRange = range
      }),

    setAccessTypeFilter: (type) =>
      set((state) => {
        state.accessTypeFilter = type
      }),

    setAccessAgentFilter: (agentId) =>
      set((state) => {
        state.accessAgentFilter = agentId
      }),

    // ============================================================================
    // UI Actions
    // ============================================================================

    setActivePanel: (panel) =>
      set((state) => {
        state.activePanel = panel
      }),

    toggleSidebar: () =>
      set((state) => {
        state.sidebarCollapsed = !state.sidebarCollapsed
      }),
  }))
)

// ============================================================================
// Selectors
// ============================================================================

export const useMemoryGraph = () =>
  useMemoryStore((state) => ({
    nodes: state.graphNodes,
    edges: state.graphEdges,
    selectedNodeId: state.selectedGraphNodeId,
    highlightedNodes: state.highlightedGraphNodes,
    layout: state.graphLayout,
  }))

export const useSemanticMemory = () =>
  useMemoryStore((state) => ({
    entries: state.semanticEntries,
    selectedEntryId: state.selectedSemanticEntryId,
    searchQuery: state.semanticSearchQuery,
    filterType: state.semanticFilterType,
  }))

export const useLongTermContext = () =>
  useMemoryStore((state) => ({
    chunks: state.contextChunks,
    selectedChunkId: state.selectedContextChunkId,
    focusArea: state.focusArea,
  }))

export const useMemoryRecall = () =>
  useMemoryStore((state) => ({
    events: state.recallEvents,
    selectedEventId: state.selectedRecallEventId,
    timeRange: state.recallTimeRange,
  }))

export const useKnowledgeClusters = () =>
  useMemoryStore((state) => ({
    clusters: state.clusters,
    selectedClusterId: state.selectedClusterId,
    highlightedClusters: state.highlightedClusters,
  }))

export const useMemoryDecay = () =>
  useMemoryStore((state) => ({
    nodes: state.decayNodes,
    selectedNodeId: state.selectedDecayNodeId,
    showPredictions: state.showDecayPredictions,
  }))

export const useVectorEmbeddings = () =>
  useMemoryStore((state) => ({
    vectors: state.vectors,
    selectedVectorId: state.selectedVectorId,
    highlightedVectors: state.highlightedVectors,
    metric: state.vectorMetric,
  }))

export const useSharedAgentMemory = () =>
  useMemoryStore((state) => ({
    agents: state.agents,
    selectedAgentId: state.selectedAgentId,
    sharedMemoryIds: state.sharedMemoryIds,
  }))

export const useMemoryMerge = () =>
  useMemoryStore((state) => ({
    candidates: state.mergeCandidates,
    selectedCandidateId: state.selectedMergeCandidateId,
    showConflicts: state.showMergeConflicts,
  }))

export const useConversationTree = () =>
  useMemoryStore((state) => ({
    nodes: state.conversationNodes,
    selectedNodeId: state.selectedConversationNodeId,
    focusedNodeId: state.focusedConversationNodeId,
  }))

export const useMemoryAccessMonitor = () =>
  useMemoryStore((state) => ({
    entries: state.accessEntries,
    selectedEntryId: state.selectedAccessEntryId,
    timeRange: state.accessTimeRange,
    accessTypeFilter: state.accessTypeFilter,
    accessAgentFilter: state.accessAgentFilter,
  }))

export const useMemoryUI = () =>
  useMemoryStore((state) => ({
    activePanel: state.activePanel,
    sidebarCollapsed: state.sidebarCollapsed,
  }))
