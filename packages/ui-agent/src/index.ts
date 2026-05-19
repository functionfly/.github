/**
 * @functionfly/ui-agent
 * Index and exports
 */

// Re-export types from AgentDock
export type {
  AgentData,
  AgentPermission,
  AgentTool,
  AgentCardProps,
  AgentDockProps,
  AgentLifecyclePanelProps,
  AgentMemoryViewerProps,
  AgentPermissionEditorProps,
  AgentToolchainEditorProps,
  AgentBudgetMeterProps,
  SwarmAgent,
  SwarmCoordinatorProps,
} from "./AgentDock";

// Additional types from types.ts (only those defined there)
export type {
  Task,
  AutonomousTaskBoardProps,
} from "./types";

// Agent components
export {
  AgentCard,
  AgentDock,
  AgentLifecyclePanel,
  AgentMemoryViewer,
  AgentPermissionEditor,
  AgentToolchainEditor,
  AgentBudgetMeter,
  SwarmCoordinator,
  getStatusConfig,
  getRiskColor,
} from "./AgentDock";

// Additional agent components
export { AgentSkillGraph, AgentDependencyMap } from "./AgentSkillGraph";
export { AutonomousTaskBoard } from "./AutonomousTaskBoard";

// Consensus and conflict resolution
export type {
  ConsensusVote,
  ConsensusDecision,
  ConflictEntry,
  AgentConsensusViewerProps,
  AgentConflictResolverProps,
} from "./types";
export { AgentConsensusViewer, AgentConflictResolver } from "./AgentConsensusViewer";

// Runtime Inspector
export type { AgentRuntimeInspectorProps } from "./AgentDock";
export { AgentRuntimeInspector } from "./AgentDock";
