/**
 * VisualizationIntegration
 * Main container that wires all Neural & Cinematic Visualization components together
 */

import * as React from "react"
import {
  NeuralExecutionMap,
  TokenParticleSystem,
  InferenceFlowField,
  CognitiveMesh,
  RuntimeGalaxyView,
  AgentConstellation,
  ExecutionPulseMap,
  GlobalInfrastructureMap,
  DataFlowRiver,
  SemanticHeatmap,
  AIReasoningTree,
  LiveDependencyNebula,
  RealtimeTopologyGraph,
  ExecutionDensityField,
  InfrastructureHologram,
  type NeuralNode,
  type NeuralConnection,
  type TokenParticle,
  type MeshNode,
  type GalaxyCluster,
  type ConstellationAgent,
  type PulseEvent,
  type InfrastructureRegion,
  type DataStream,
  type SemanticCell,
  type ReasoningNode,
  type NebulaNode,
  type TopologyNode,
  type TopologyLink,
  type DensityPoint,
  type HologramLayer,
} from "@functionfly/ui-visualization"
import { useVisualizationStore } from "@/stores/visualizationStore"
import { cn } from "@functionfly/ui-core"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@functionfly/ui-core"

// ============================================================================
// Types
// ============================================================================

interface VisualizationIntegrationProps {
  className?: string
}

// ============================================================================
// Sample Data Helpers
// ============================================================================

const sampleNeuralNodes: NeuralNode[] = [
  { id: "n1", label: "Orchestrator", type: "agent", position: [0, 0, 0], status: "active", connections: ["n2", "n3"], load: 0.8 },
  { id: "n2", label: "DataProcessor", type: "function", position: [-2, 1, 0], status: "active", connections: ["n4"], load: 0.6 },
  { id: "n3", label: "MLInference", type: "function", position: [2, 1, 0], status: "processing", connections: ["n5"], load: 0.9 },
  { id: "n4", label: "PostgreSQL", type: "database", position: [-3, 2, 0], status: "idle", connections: [], load: 0.2 },
  { id: "n5", label: "RedisCache", type: "memory", position: [3, 2, 0], status: "active", connections: [], load: 0.5 },
  { id: "n6", label: "UserInput", type: "input", position: [0, -2, 0], status: "active", connections: ["n1"], load: 1 },
  { id: "n7", label: "Response", type: "output", position: [0, 3, 0], status: "idle", connections: [], load: 0 },
]

const sampleNeuralConnections: NeuralConnection[] = [
  { from: "n6", to: "n1", type: "data", strength: 1 },
  { from: "n1", to: "n2", type: "control", strength: 0.8 },
  { from: "n1", to: "n3", type: "control", strength: 0.9 },
  { from: "n2", to: "n4", type: "data", strength: 0.5 },
  { from: "n3", to: "n5", type: "data", strength: 0.7 },
  { from: "n1", to: "n7", type: "data", strength: 0.6 },
]

const sampleTokens: TokenParticle[] = Array.from({ length: 30 }, (_, i) => ({
  id: `tok-${i}`,
  type: ["input", "output", "hidden"][Math.floor(Math.random() * 3)] as "input" | "output" | "hidden",
  position: { x: Math.random() * 600, y: Math.random() * 400 },
  velocity: { x: Math.random() * 2 - 1, y: Math.random() * 2 - 1 },
  size: 3 + Math.random() * 4,
  color: "",
  alpha: 0.8,
  createdAt: Date.now(),
}))

const sampleMeshNodes: MeshNode[] = [
  { id: "m1", label: "Input Layer", position: { x: 100, y: 200 }, connections: ["m2", "m3"], activation: 0.9, type: "input" },
  { id: "m2", label: "Hidden 1", position: { x: 200, y: 120 }, connections: ["m4"], activation: 0.7, type: "hidden" },
  { id: "m3", label: "Hidden 2", position: { x: 200, y: 280 }, connections: ["m4"], activation: 0.6, type: "hidden" },
  { id: "m4", label: "Attention", position: { x: 300, y: 200 }, connections: ["m5", "m6"], activation: 0.8, type: "attention" },
  { id: "m5", label: "Output", position: { x: 400, y: 150 }, connections: [], activation: 0.5, type: "output" },
  { id: "m6", label: "Output 2", position: { x: 400, y: 250 }, connections: [], activation: 0.4, type: "output" },
]

const sampleClusters: GalaxyCluster[] = [
  { id: "c1", name: "US-East", position: [-100, 0, 0], size: 80, nodes: [{ id: "n1", status: "healthy", connections: 12 }, { id: "n2", status: "healthy", connections: 8 }] },
  { id: "c2", name: "EU-West", position: [50, 50, 50], size: 60, nodes: [{ id: "n3", status: "degraded", connections: 6 }, { id: "n4", status: "healthy", connections: 10 }] },
  { id: "c3", name: "Asia-Pacific", position: [100, -50, 100], size: 70, nodes: [{ id: "n5", status: "healthy", connections: 15 }, { id: "n6", status: "down", connections: 0 }] },
]

const sampleAgents: ConstellationAgent[] = [
  { id: "a1", name: "Orchestrator", x: 300, y: 200, size: 12, connections: ["a2", "a3"], status: "active", role: "Coordinator" },
  { id: "a2", name: "DataAgent", x: 200, y: 150, size: 8, connections: ["a4"], status: "active", role: "Data Processing" },
  { id: "a3", name: "MLAgent", x: 400, y: 150, size: 8, connections: ["a4", "a5"], status: "active", role: "Inference" },
  { id: "a4", name: "Validator", x: 300, y: 100, size: 6, connections: [], status: "idle", role: "Validation" },
  { id: "a5", name: "OutputAgent", x: 450, y: 220, size: 7, connections: [], status: "error", role: "Output Formatting" },
]

const samplePulseEvents: PulseEvent[] = [
  { id: "p1", timestamp: Date.now() - 5000, source: "input", target: "processor", type: "start" },
  { id: "p2", timestamp: Date.now() - 4000, source: "processor", target: "ml", type: "complete", duration: 1000 },
  { id: "p3", timestamp: Date.now() - 3000, source: "ml", target: "cache", type: "start" },
  { id: "p4", timestamp: Date.now() - 2000, source: "cache", target: "output", type: "complete", duration: 500 },
  { id: "p5", timestamp: Date.now() - 1000, source: "output", target: "response", type: "error" },
]

const sampleRegions: InfrastructureRegion[] = [
  { id: "r1", name: "US East", position: "us-east", load: 0.7, nodes: 156, status: "healthy" },
  { id: "r2", name: "US West", position: "us-west", load: 0.4, nodes: 89, status: "healthy" },
  { id: "r3", name: "EU West", position: "eu-west", load: 0.85, nodes: 234, status: "degraded" },
  { id: "r4", name: "Asia Pacific", position: "asia-pacific", load: 0.3, nodes: 67, status: "healthy" },
]

const sampleStreams: DataStream[] = [
  { id: "s1", label: "HTTP Requests", flow: 1, packets: 1200, bandwidth: 45.2, type: "request" },
  { id: "s2", label: "HTTP Responses", flow: -1, packets: 1150, bandwidth: 38.7, type: "response" },
  { id: "s3", label: "WebSocket Events", flow: 0.3, packets: 3400, bandwidth: 12.5, type: "event" },
  { id: "s4", label: "gRPC", flow: 0.5, packets: 890, bandwidth: 22.3, type: "request" },
]

const sampleHeatmapCells: SemanticCell[] = [
  { row: "transform", col: "etl", value: 0.8 }, { row: "transform", col: "stream", value: 0.6 },
  { row: "ml", col: "inference", value: 0.9 }, { row: "ml", col: "training", value: 0.4 },
  { row: "automation", col: "webhook", value: 0.7 }, { row: "automation", col: "cron", value: 0.5 },
  { row: "integration", col: "api", value: 0.85 }, { row: "integration", col: "queue", value: 0.6 },
]

const sampleReasoningTree: ReasoningNode = {
  id: "root",
  label: "User Query Analysis",
  type: "premise",
  confidence: 1,
  children: [
    {
      id: "inf1",
      label: "Intent Detection",
      type: "inference",
      confidence: 0.95,
      children: [
        { id: "inf1-1", label: "Search Intent", type: "inference", confidence: 0.88 },
        { id: "inf1-2", label: "Action Intent", type: "conclusion", confidence: 0.92 },
      ],
    },
    {
      id: "inf2",
      label: "Entity Extraction",
      type: "inference",
      confidence: 0.78,
      children: [
        { id: "inf2-1", label: "Date Entities", type: "assumption", confidence: 0.65 },
        { id: "inf2-2", label: "Name Entities", type: "assumption", confidence: 0.71 },
      ],
    },
    {
      id: "conc1",
      label: "Final Response Plan",
      type: "conclusion",
      confidence: 0.89,
    },
  ],
}

const sampleNebulaNodes: NebulaNode[] = [
  { id: "neb1", name: "Orchestrator", dependencies: ["neb2", "neb3"], health: 0.95, critical: true },
  { id: "neb2", name: "DataProcessor", dependencies: ["neb4"], health: 0.8, critical: false },
  { id: "neb3", name: "MLInference", dependencies: ["neb5"], health: 0.7, critical: true },
  { id: "neb4", name: "PostgreSQL", dependencies: [], health: 0.6, critical: false },
  { id: "neb5", name: "RedisCache", dependencies: [], health: 0.9, critical: false },
]

const sampleTopologyNodes: TopologyNode[] = [
  { id: "t1", label: "API Gateway", type: "gateway", status: "healthy", x: 300, y: 50 },
  { id: "t2", label: "Auth Service", type: "service", status: "healthy", x: 200, y: 150 },
  { id: "t3", label: "User Service", type: "service", status: "healthy", x: 400, y: 150 },
  { id: "t4", label: "User DB", type: "database", status: "healthy", x: 200, y: 250 },
  { id: "t5", label: "Redis", type: "cache", status: "degraded", x: 400, y: 250 },
  { id: "t6", label: "Task Queue", type: "queue", status: "healthy", x: 300, y: 200 },
]

const sampleTopologyLinks: TopologyLink[] = [
  { from: "t1", to: "t2", bandwidth: 100 },
  { from: "t1", to: "t3", bandwidth: 100 },
  { from: "t2", to: "t4", bandwidth: 50 },
  { from: "t3", to: "t5", bandwidth: 50 },
  { from: "t2", to: "t6", bandwidth: 30 },
  { from: "t3", to: "t6", bandwidth: 30 },
]

const sampleDensityPoints: DensityPoint[] = Array.from({ length: 50 }, () => ({
  x: Math.random() * 600,
  y: Math.random() * 400,
  density: Math.random(),
}))

const sampleHologramLayers: HologramLayer[] = [
  {
    id: "layer1",
    label: "compute",
    items: [
      { id: "i1", label: "CPU", value: 78, max: 100 },
      { id: "i2", label: "Memory", value: 65, max: 100 },
      { id: "i3", label: "GPU", value: 92, max: 100 },
    ],
  },
  {
    id: "layer2",
    label: "network",
    items: [
      { id: "i4", label: "Ingress", value: 45, max: 100 },
      { id: "i5", label: "Egress", value: 38, max: 100 },
      { id: "i6", label: "Packets", value: 890, max: 1000 },
    ],
  },
  {
    id: "layer3",
    label: "storage",
    items: [
      { id: "i7", label: "Read", value: 120, max: 200 },
      { id: "i8", label: "Write", value: 85, max: 200 },
      { id: "i9", label: "IOPS", value: 4500, max: 5000 },
    ],
  },
]

// ============================================================================
// Component
// ============================================================================

export function VisualizationIntegration({ className }: VisualizationIntegrationProps) {
  const {
    activeVisualization,
    setActiveVisualization,
  } = useVisualizationStore()

  return (
    <div className={cn("space-y-6", className)}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold text-text-primary flex items-center gap-2">
          <span className="text-2xl">🌌</span>
          Neural Visualization
        </h3>
      </div>

      {/* Main tabs */}
      <Tabs value={activeVisualization} onValueChange={setActiveVisualization}>
        <TabsList>
          <TabsTrigger value="neural-map">Neural Map</TabsTrigger>
          <TabsTrigger value="particles">Particles</TabsTrigger>
          <TabsTrigger value="flow">Flow Fields</TabsTrigger>
          <TabsTrigger value="cognitive">Cognitive Mesh</TabsTrigger>
          <TabsTrigger value="galaxy">Galaxy</TabsTrigger>
          <TabsTrigger value="constellation">Agents</TabsTrigger>
          <TabsTrigger value="pulse">Pulse</TabsTrigger>
          <TabsTrigger value="infrastructure">Infrastructure</TabsTrigger>
          <TabsTrigger value="nebula">Dependency</TabsTrigger>
          <TabsTrigger value="topology">Topology</TabsTrigger>
          <TabsTrigger value="hologram">Hologram</TabsTrigger>
        </TabsList>

        {/* Neural Map */}
        <TabsContent value="neural-map">
          <NeuralExecutionMap
            nodes={sampleNeuralNodes}
            connections={sampleNeuralConnections}
            showLabels={true}
            showConnections={true}
            autoRotate={true}
            style="cosmic"
          />
        </TabsContent>

        {/* Token Particles */}
        <TabsContent value="particles">
          <TokenParticleSystem
            tokens={sampleTokens}
            width={600}
            height={400}
            showTrails={true}
            colorScheme="brand"
          />
        </TabsContent>

        {/* Flow Field */}
        <TabsContent value="flow">
          <InferenceFlowField
            width={600}
            height={400}
            density={20}
            autoStream={true}
            flowColor="#ff6b35"
            showArrows={true}
          />
        </TabsContent>

        {/* Cognitive Mesh */}
        <TabsContent value="cognitive">
          <CognitiveMesh
            nodes={sampleMeshNodes}
            width={600}
            height={400}
            animated={true}
          />
        </TabsContent>

        {/* Galaxy View */}
        <TabsContent value="galaxy">
          <RuntimeGalaxyView
            clusters={sampleClusters}
            showLabels={true}
            autoRotate={true}
          />
        </TabsContent>

        {/* Agent Constellation */}
        <TabsContent value="constellation">
          <AgentConstellation
            agents={sampleAgents}
            showRoles={true}
          />
        </TabsContent>

        {/* Execution Pulse */}
        <TabsContent value="pulse">
          <ExecutionPulseMap
            events={samplePulseEvents}
            width={600}
            height={400}
          />
        </TabsContent>

        {/* Infrastructure */}
        <TabsContent value="infrastructure">
          <div className="space-y-4">
            <GlobalInfrastructureMap regions={sampleRegions} />
            <DataFlowRiver streams={sampleStreams} width={600} height={200} />
          </div>
        </TabsContent>

        {/* Dependency Nebula */}
        <TabsContent value="nebula">
          <div className="space-y-4">
            <LiveDependencyNebula nodes={sampleNebulaNodes} />
            <SemanticHeatmap cells={sampleHeatmapCells} width={500} height={350} />
          </div>
        </TabsContent>

        {/* Topology */}
        <TabsContent value="topology">
          <div className="space-y-4">
            <RealtimeTopologyGraph
              nodes={sampleTopologyNodes}
              links={sampleTopologyLinks}
            />
            <ExecutionDensityField
              points={sampleDensityPoints}
              width={600}
              height={300}
              colorScheme="brand"
            />
          </div>
        </TabsContent>

        {/* Hologram */}
        <TabsContent value="hologram">
          <div className="space-y-4">
            <InfrastructureHologram layers={sampleHologramLayers} width={600} height={350} />
            <AIReasoningTree root={sampleReasoningTree} direction="horizontal" />
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}

// ============================================================================
// Export all visualization components
// ============================================================================

export {
  NeuralExecutionMap,
  TokenParticleSystem,
  InferenceFlowField,
  CognitiveMesh,
  RuntimeGalaxyView,
  AgentConstellation,
  ExecutionPulseMap,
  GlobalInfrastructureMap,
  DataFlowRiver,
  SemanticHeatmap,
  AIReasoningTree,
  LiveDependencyNebula,
  RealtimeTopologyGraph,
  ExecutionDensityField,
  InfrastructureHologram,
} from "@functionfly/ui-visualization"

export type {
  NeuralNode,
  NeuralConnection,
  TokenParticle,
  MeshNode,
  GalaxyCluster,
  ConstellationAgent,
  PulseEvent,
  InfrastructureRegion,
  DataStream,
  SemanticCell,
  ReasoningNode,
  NebulaNode,
  TopologyNode,
  TopologyLink,
  DensityPoint,
  HologramLayer,
} from "@functionfly/ui-visualization"
