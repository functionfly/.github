/**
 * Dashboard Components Index
 * 
 * Unified exports for all dashboard components including:
 * - Studio Layout System (studio/)
 * - Registry & Marketplace (registry/)
 * - Graph Runtime (graph/)
 * - Observability (observability/)
 * - AI Command System (ai/)
 */

// Studio Layout System
export * from './studio'
export { useStudioStore } from '@/stores/studioStore'

// Registry & Marketplace Components
export * from './registry'
export { useRegistryStore, selectFilteredFunctions, selectSortedFunctions } from '@/stores/registryStore'

// Graph Runtime Components
export * from './graph'
export { useGraphRuntimeStore } from '@/stores/graphRuntimeStore'

// Observability Components
export * from './observability'
export { useObservabilityStore, selectCriticalMetrics, selectRecentLogs, selectActiveIncidents } from '@/stores/observabilityStore'

// AI Command System
export * from './ai'
export { useAICommandStore } from '@/stores/aiCommandStore'

// Visualization Components
export * from './visualization'
export { useVisualizationStore } from '@/stores/visualizationStore'

// Code Intelligence Components
export * from './codeintelligence'
export { useCodeIntelligenceStore } from '@/stores/codeIntelligenceStore'

// DevOps Components
export * from './devops'
export { useDevOpsStore } from '@/stores/devopsStore'

// Security Components
export * from './security'
export { useSecurityStore } from '@/stores/securityStore'

// Collaboration Components
export * from './collaboration'
export { useCollaborationStore, selectActivePresences, selectSpeakingParticipants, selectUnresolvedConflicts } from '@/stores/collaborationStore'

// Memory Components
export * from './memory'
export { useMemoryStore } from '@/stores/memoryStore'

// Simulation Components
export * from './simulation'
export { useSimulationStore } from '@/stores/simulationStore'

// Robotics Components
export * from './robotics'
export { useRoboticsStore } from '@/stores/roboticsStore'

// Marketplace Economy Components
export * from './marketplace-economy'
export { useMarketplaceEconomyStore } from '@/stores/marketplaceEconomyStore'

// Adaptive UX Components
export * from './adaptive-ux'
export { useAdaptiveUXStore } from '@/stores/adaptiveUXStore'

// Universal Runtime Components
export * from './universal-runtime'
export { useUniversalRuntimeStore } from '@/stores/universalRuntimeStore'

// Data Visualization Components
export * from './data-visualization'
export { useDataVisualizationStore } from '@/stores/dataVisualizationStore'

// Futuristic Signature Components
export * from './futuristic'
export { useFuturisticStore } from '@/stores/futuristicStore'
