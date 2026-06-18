/**
 * useDevOps Hook
 * Main hook and specialized hooks for DevOps components
 */

import { useDevOpsStore } from '../stores/devopsStore'
import { useCodeIntelligenceStore } from '@/stores/codeIntelligenceStore'

// Main hook returning the full store
export const useDevOps = () => useDevOpsStore()

// Re-export all selectors and actions
export {
  useDevOpsStore,
  usePipeline,
  useEnvironments,
  useCloudRegions,
  useRuntimeTargets,
  useKubernetes,
  useEdgeLocations,
  useContainers,
  useVaults,
  useScalableResources,
  useTrafficBalancer,
  useRollbackManager,
  useBuildArtifacts,
  useClusterHealth,
  useColdStartAnalyzer,
  useServerlessExecutionMap,
  useDevOpsUI,
} from '../stores/devopsStore'
