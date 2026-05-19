/**
 * Visualization Store
 * Global state management for Neural & Cinematic Visualization components
 */

import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'

// ============================================================================
// Types
// ============================================================================

export interface NeuralNode {
  id: string
  label: string
  type: "agent" | "function" | "api" | "memory" | "database" | "input" | "output"
  position: [number, number, number]
  status?: "idle" | "active" | "error" | "processing"
  connections: string[]
  load?: number
  size?: number
  color?: string
}

export interface NeuralConnection {
  from: string
  to: string
  type?: "data" | "control" | "signal"
  strength?: number
  active?: boolean
}

export interface TokenParticle {
  id: string
  type: "input" | "output" | "hidden"
  position: { x: number; y: number; z?: number }
  velocity?: { x: number; y: number; z?: number }
  size: number
  color: string
  alpha?: number
  lifespan?: number
  createdAt: number
}

export interface MeshNode {
  id: string
  label: string
  position: { x: number; y: number }
  connections: string[]
  activation: number
  type: "input" | "hidden" | "output" | "attention"
}

export interface GalaxyCluster {
  id: string
  name: string
  position: [number, number, number]
  size: number
  nodes: Array<{
    id: string
    status: "healthy" | "degraded" | "down"
    connections: number
  }>
}

export interface ConstellationAgent {
  id: string
  name: string
  x: number
  y: number
  size: number
  connections: string[]
  status: "active" | "idle" | "error"
  role?: string
}

export interface PulseEvent {
  id: string
  timestamp: number
  source: string
  target: string
  type: "start" | "complete" | "error"
  duration?: number
}

export interface InfrastructureRegion {
  id: string
  name: string
  position: "us-east" | "us-west" | "eu-west" | "asia-pacific" | "south-america"
  load: number
  nodes: number
  status: "healthy" | "degraded" | "down"
}

export interface DataStream {
  id: string
  label: string
  flow: number
  packets: number
  bandwidth: number
  type: "request" | "response" | "event"
}

export interface SemanticCell {
  row: string
  col: string
  value: number
}

export interface ReasoningNode {
  id: string
  label: string
  type: "premise" | "inference" | "conclusion" | "assumption"
  children?: ReasoningNode[]
  confidence?: number
}

export interface NebulaNode {
  id: string
  name: string
  dependencies: string[]
  health: number
  critical: boolean
}

export interface TopologyNode {
  id: string
  label: string
  type: "gateway" | "service" | "database" | "cache" | "queue"
  status: "healthy" | "degraded" | "down"
  x?: number
  y?: number
}

export interface TopologyLink {
  from: string
  to: string
  bandwidth: number
}

export interface DensityPoint {
  x: number
  y: number
  density: number
}

export interface HologramLayer {
  id: string
  label: string
  items: Array<{
    id: string
    label: string
    value: number
    max: number
  }>
}

// ============================================================================
// Store Interface
// ============================================================================

interface VisualizationState {
  // Neural Execution Map
  neuralNodes: NeuralNode[]
  neuralConnections: NeuralConnection[]
  setNeuralData: (nodes: NeuralNode[], connections: NeuralConnection[]) => void
  updateNeuralNode: (id: string, updates: Partial<NeuralNode>) => void

  // Token Particles
  tokens: TokenParticle[]
  setTokens: (tokens: TokenParticle[]) => void
  addToken: (token: TokenParticle) => void
  removeToken: (id: string) => void

  // Cognitive Mesh
  meshNodes: MeshNode[]
  selectedMeshNodeId: string | null
  setMeshNodes: (nodes: MeshNode[]) => void
  setSelectedMeshNode: (id: string | null) => void

  // Galaxy clusters
  clusters: GalaxyCluster[]
  setClusters: (clusters: GalaxyCluster[]) => void

  // Agent constellation
  agents: ConstellationAgent[]
  selectedAgentId: string | null
  setAgents: (agents: ConstellationAgent[]) => void
  setSelectedAgent: (id: string | null) => void

  // Execution pulse
  pulseEvents: PulseEvent[]
  selectedPulseEventId: string | null
  addPulseEvent: (event: PulseEvent) => void
  setPulseEvents: (events: PulseEvent[]) => void
  setSelectedPulseEvent: (id: string | null) => void

  // Infrastructure
  regions: InfrastructureRegion[]
  selectedRegionId: string | null
  setRegions: (regions: InfrastructureRegion[]) => void
  setSelectedRegion: (id: string | null) => void

  // Data flow
  streams: DataStream[]
  setStreams: (streams: DataStream[]) => void

  // Semantic heatmap
  heatmapCells: SemanticCell[]
  setHeatmapCells: (cells: SemanticCell[]) => void

  // Reasoning tree
  reasoningRoot: ReasoningNode | null
  selectedReasoningNodeId: string | null
  setReasoningRoot: (root: ReasoningNode) => void
  setSelectedReasoningNode: (id: string | null) => void

  // Dependency nebula
  nebulaNodes: NebulaNode[]
  selectedNebulaNodeId: string | null
  setNebulaNodes: (nodes: NebulaNode[]) => void
  setSelectedNebulaNode: (id: string | null) => void

  // Topology
  topologyNodes: TopologyNode[]
  topologyLinks: TopologyLink[]
  selectedTopologyNodeId: string | null
  setTopologyData: (nodes: TopologyNode[], links: TopologyLink[]) => void
  setSelectedTopologyNode: (id: string | null) => void

  // Density field
  densityPoints: DensityPoint[]
  setDensityPoints: (points: DensityPoint[]) => void

  // Hologram layers
  hologramLayers: HologramLayer[]
  setHologramLayers: (layers: HologramLayer[]) => void

  // Active visualization
  activeVisualization: string
  setActiveVisualization: (id: string) => void

  // Reset
  reset: () => void
}

// ============================================================================
// Initial State
// ============================================================================

const initialState = {
  neuralNodes: [],
  neuralConnections: [],
  tokens: [],
  meshNodes: [],
  selectedMeshNodeId: null,
  clusters: [],
  agents: [],
  selectedAgentId: null,
  pulseEvents: [],
  selectedPulseEventId: null,
  regions: [],
  selectedRegionId: null,
  streams: [],
  heatmapCells: [],
  reasoningRoot: null,
  selectedReasoningNodeId: null,
  nebulaNodes: [],
  selectedNebulaNodeId: null,
  topologyNodes: [],
  topologyLinks: [],
  selectedTopologyNodeId: null,
  densityPoints: [],
  hologramLayers: [],
  activeVisualization: "neural-map",
}

// ============================================================================
// Store
// ============================================================================

export const useVisualizationStore = create<VisualizationState>()(
  immer((set) => ({
    ...initialState,

    setNeuralData: (nodes, connections) => set((state) => {
      state.neuralNodes = nodes
      state.neuralConnections = connections
    }),

    updateNeuralNode: (id, updates) => set((state) => {
      const node = state.neuralNodes.find((n) => n.id === id)
      if (node) Object.assign(node, updates)
    }),

    setTokens: (tokens) => set((state) => { state.tokens = tokens }),

    addToken: (token) => set((state) => {
      state.tokens.push(token)
    }),

    removeToken: (id) => set((state) => {
      state.tokens = state.tokens.filter((t) => t.id !== id)
    }),

    setMeshNodes: (nodes) => set((state) => { state.meshNodes = nodes }),

    setSelectedMeshNode: (id) => set((state) => { state.selectedMeshNodeId = id }),

    setClusters: (clusters) => set((state) => { state.clusters = clusters }),

    setAgents: (agents) => set((state) => { state.agents = agents }),

    setSelectedAgent: (id) => set((state) => { state.selectedAgentId = id }),

    addPulseEvent: (event) => set((state) => {
      state.pulseEvents.push(event)
      if (state.pulseEvents.length > 100) {
        state.pulseEvents = state.pulseEvents.slice(-100)
      }
    }),

    setPulseEvents: (events) => set((state) => { state.pulseEvents = events }),

    setSelectedPulseEvent: (id) => set((state) => { state.selectedPulseEventId = id }),

    setRegions: (regions) => set((state) => { state.regions = regions }),

    setSelectedRegion: (id) => set((state) => { state.selectedRegionId = id }),

    setStreams: (streams) => set((state) => { state.streams = streams }),

    setHeatmapCells: (cells) => set((state) => { state.heatmapCells = cells }),

    setReasoningRoot: (root) => set((state) => { state.reasoningRoot = root }),

    setSelectedReasoningNode: (id) => set((state) => { state.selectedReasoningNodeId = id }),

    setNebulaNodes: (nodes) => set((state) => { state.nebulaNodes = nodes }),

    setSelectedNebulaNode: (id) => set((state) => { state.selectedNebulaNodeId = id }),

    setTopologyData: (nodes, links) => set((state) => {
      state.topologyNodes = nodes
      state.topologyLinks = links
    }),

    setSelectedTopologyNode: (id) => set((state) => { state.selectedTopologyNodeId = id }),

    setDensityPoints: (points) => set((state) => { state.densityPoints = points }),

    setHologramLayers: (layers) => set((state) => { state.hologramLayers = layers }),

    setActiveVisualization: (id) => set((state) => { state.activeVisualization = id }),

    reset: () => set(() => ({ ...initialState })),
  }))
)

// ============================================================================
// Selectors
// ============================================================================

export const selectActiveNodes = (nodes: NeuralNode[]) =>
  nodes.filter((n) => n.status === "active")

export const selectCriticalNodes = (nodes: NebulaNode[]) =>
  nodes.filter((n) => n.critical)

export const selectHealthyRegions = (regions: InfrastructureRegion[]) =>
  regions.filter((r) => r.status === "healthy")
