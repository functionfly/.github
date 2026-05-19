import React, { useState, useEffect, useCallback, useMemo } from "react";
import { useParams, useNavigate, useLocation } from "react-router-dom";
import {
  StudioShell,
  ResizablePanelGroup,
  ResizablePanel,
  ResizableHandle,
  Tabs,
  TabsList,
  TabsTrigger,
  TabsContent,
  Tooltip,
  GlassCard,
} from "@functionfly/ui-core";
import {
  FunctionCanvas,
  generateSampleGraph,
  getStatusColor,
  type NodeData,
} from "@functionfly/ui-graph";
import {
  AICommandPalette,
  type AICommand,
} from "@functionfly/ui-ai";
import {
  GhostModeOrchestrator,
  type GhostBuild,
  type GhostTask,
  type AgentConversationMessage,
  type AgentDecisionPoint,
} from "@functionfly/ui-ghost";
import {
  useStudioState,
  useStudioAgents,
  useStudioSimulation,
  useStudioGhost,
  useStudioTelemetry,
  useStudioWorkflow,
  useStudioMemory,
  type AgentStatus,
} from "@/hooks/useStudio";
import { useActiveEnvironment } from "@/hooks/useActiveEnvironment";
import type { AgentMemory } from "@/types";
import type { CollabEvent } from "@/api/studioCollab";
import { useAuthStore } from "@/stores/authStore";
import { usePresence } from "@/hooks/usePresence";
import { useTeams, useTeamMembers } from "@/hooks/useTeams";
import { useTeamMemories } from "@/hooks/use-team-memory";
import { useStudioCollabEvents, useStudioCollabActivity } from "@/hooks/useStudioCollab";
import {
  useExecuteFunction,
  useFavoriteFunction,
  useCreatePlan,
  useEditPlan,
  useRequestPayout,
  useUpdateLicense,
  useUpdatePricing,
} from "@/hooks/useStudioMarketplace";
import {
  useInstallExtension,
  useUninstallExtension,
  useEnableExtension,
  useDisableExtension,
  useConfigureExtension,
} from "@/hooks/useStudioExtension";
import { PluginManager } from "@/pages/StudioPage/components/PluginManager";
import {
  StudioSettingsCenter,
  ThemeEngine,
  WorkspaceSnapshotManager,
  KeyboardShortcutVisualizer,
  StudioPerformanceProfiler,
  CrashRecoveryManager,
  ExperimentalFeatureLab,
  UniversalSearchEngine,
  GlobalNotificationCenter,
  StudioUpdateManager,
} from "@/pages/StudioPage/components";
import { LivePresence } from "@functionfly/ui-collaboration";
import { DataSourceConfigDialog, type DataSource } from "@/pages/StudioPage/components/DataSourceConfigDialog";
import {
  useCreateTask,
  useUpdateTask,
  useDeleteTask,
  useAssignTask,
} from "@/hooks/useStudioTask";
import {
  useResolveComment,
  useCreateComment,
  useResolveAnnotation,
  useEndPairSession,
  useUpdatePromptVersion,
  useRecordActivity,
} from "@/hooks/useStudioCollabActions";
import { useReassignAgentRole, useReshapeSwarm } from "@/hooks/useAgentSwarm";
import { useQuery } from "@tanstack/react-query";
import {
  Search,
  Bot,
  Code,
  GitBranch,
  Activity,
  Zap,
  Clock,
  Database,
  BarChart2,
  Settings,
  Play,
  Pause,
  RotateCcw,
  ChevronRight,
  Brain,
  Terminal,
  Eye,
  Layers,
  Cpu,
  MemoryStick,
  DollarSign,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Loader2,
  LineChart,
  Target,
  Gauge,
  TrendingUp,
  TrendingDown,
  RefreshCcw,
  Filter,
  SortAsc,
  Sparkles,
  Puzzle,
  Users,
  Wand2,
  CheckSquare,
  Share2,
  Save,
  Undo2,
  Redo2,
  Rocket,
} from "lucide-react";

// Import extracted panels
import { AgentsPanel } from "@/pages/StudioPage/left-panels/AgentsPanel";
import { CanvasPanel } from "@/pages/StudioPage/left-panels/CanvasPanel";
import { MarketplacePanel } from "@/pages/StudioPage/left-panels/MarketplacePanel";
import { RuntimePanel } from "@/pages/StudioPage/left-panels/RuntimePanel";
import { SwarmPanel } from "@/pages/StudioPage/left-panels/SwarmPanel";
import { SkillsPanel } from "@/pages/StudioPage/left-panels/SkillsPanel";
import {
  ExecutionPanel,
  SimulationPanel,
  GhostPanel,
  TasksPanel,
  DevOpsPanel,
  MemoryPanel,
  RoboticsPanel,
} from "@/pages/StudioPage/bottom-panels";
import {
  TelemetryPanel,
  VisualizationPanel,
  ProfilerPanel,
  CollabPanel,
} from "@/pages/StudioPage/right-panels";
import { StudioEditor } from "@/pages/StudioPage/editor/StudioEditor";
import { StatusBar } from "@/pages/StudioPage/StatusBar";
import { useKeyboardShortcuts, DEFAULT_STUDIO_SHORTCUTS } from "@/pages/StudioPage/hooks";

const DEFAULT_CODE = `// FunctionFly Studio - Main Workflow
import { agent, workflow } from "@functionfly/sdk";

interface ProcessRequest {
  input: string;
  context: Record<string, unknown>;
}

interface ProcessResponse {
  result: string;
  confidence: number;
  metadata: {
    processingTime: number;
    tokensUsed: number;
  };
}

@workflow({ name: "data-processor", version: "1.0.0" })
export async function processData(req: ProcessRequest): Promise<ProcessResponse> {
  const startTime = Date.now();
  
  // Step 1: Parse and validate input
  const parsed = await agent.invoke("parser", { input: req.input });
  
  // Step 2: Enrich with context
  const enriched = await agent.invoke("enricher", {
    data: parsed,
    context: req.context,
  });
  
  // Step 3: Generate response
  const response = await agent.invoke("generator", {
    prompt: enriched,
    temperature: 0.7,
  });
  
  return {
    result: response.content,
    confidence: response.confidence,
    metadata: {
      processingTime: Date.now() - startTime,
      tokensUsed: response.usage.total,
    },
  };
}
`;

function agentStatusToAgentData(agent: AgentStatus): AgentData {
  return {
    id: agent.id,
    name: agent.name,
    role: agent.agentId,
    status: agent.status === 'active' ? 'running' : agent.status === 'pending' ? 'idle' : agent.status === 'terminating' || agent.status === 'terminated' ? 'stopped' : agent.status as 'running' | 'idle' | 'paused' | 'stopped' | 'error',
    memoryUsage: 0,
    memoryLimit: 512 * 1024 * 1024,
    executionBudget: 10.0,
    executionBudgetUsed: 0,
    permissions: [],
    tools: [],
    runtime: "wasm",
    model: "gpt-4o",
    uptime: 0,
    tasksCompleted: 0,
    tasksFailed: 0,
    avgLatency: 0,
    lastHeartbeat: agent.lastActivity || new Date().toISOString(),
    createdAt: new Date().toISOString(),
    description: `Agent ${agent.name}`,
    tags: [],
  };
}

export function StudioPage() {
  const { environment: urlEnvironment } = useParams<{ environment?: string }>();

  const [leftPanelTab, setLeftPanelTab] = useState<"agents" | "canvas" | "marketplace" | "runtime" | "swarm" | "skills" | "extensions" | "devops" | "memory" | "robotics">("agents");
  const [rightPanelTab, setRightPanelTab] = useState<"telemetry" | "visualization" | "profiler" | "collab">("telemetry");
  const [bottomPanelTab, setBottomPanelTab] = useState<"execution" | "simulation" | "ghost" | "tasks" | "devops" | "memory" | "robotics">("execution");
  const [selectedAgentId, setSelectedAgentId] = useState<string | null>(null);
  const [code, setCode] = useState(DEFAULT_CODE);
  const [isCommandPaletteOpen, setIsCommandPaletteOpen] = useState(false);
  const [editorTheme, setEditorTheme] = useState<"studio-dark" | "studio-light" | "monokai" | "github-dark">("studio-dark");
  const [leftPanelCollapsed, setLeftPanelCollapsed] = useState(false);
  const [rightPanelCollapsed, setRightPanelCollapsed] = useState(false);
  const [bottomPanelCollapsed, setBottomPanelCollapsed] = useState(false);
  const [selectedRuntime, setSelectedRuntime] = useState<RuntimeSelection | null>(null);
  const [marketplaceSearch, setMarketplaceSearch] = useState("");
  const [marketplaceCategory, setMarketplaceCategory] = useState("");
  const [dataSourceDialogOpen, setDataSourceDialogOpen] = useState(false);
  const [dataSources, setDataSources] = useState<DataSource[]>([]);

  const [lastSaved, setLastSaved] = useState<Date | null>(null);
  const [isConnected, setIsConnected] = useState(true);

  const { state: studioState, setSelectedPanel, toggleMinimap, toggleGrid } = useStudioState();
  const { collaborators, currentUser } = useStudioCollaborators();
  const { user } = useAuthStore();

  const { agents, isLoading: agentsLoading, refreshAgents, spawnAgent, pauseAgent, resumeAgent, terminateAgent } = useStudioAgents();
  const { graph, executions, isLoadingGraph, isLoadingExecutions, executeWorkflow } = useStudioWorkflow();
  const { activeSimulation, startSimulation, abortSimulation } = useStudioSimulation();
  const { metrics, tokenUsage, latencyStats, errorRateStats, isLoading: telemetryLoading } = useStudioTelemetry();
  const { builds, tasks, createBuild, approveTask, rejectTask } = useStudioGhost();
  const agentMemories = useStudioMemory(selectedAgentId || undefined);

  const { data: teamMemories } = useTeamMemories("");

  const { data: collabActivityData } = useStudioCollabActivity();
  const { data: collabEvents } = useStudioCollabEvents();

  const simulationConfig: SimulationConfig = {
    name: "Default Simulation",
    iterations: 10,
    duration: 60,
    stressLevel: "low",
  };

  const simulationRunning = activeSimulation?.status === "running";

  const { environment: currentEnvironment, setEnvironment } = useActiveEnvironment();
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    if (urlEnvironment && urlEnvironment !== currentEnvironment) {
      setEnvironment(urlEnvironment as 'production' | 'staging' | 'development');
    }
  }, [urlEnvironment, currentEnvironment, setEnvironment]);

  useEffect(() => {
    if (currentEnvironment && currentEnvironment !== urlEnvironment) {
      const newPath = currentEnvironment === 'production'
        ? '/studio'
        : `/studio/${currentEnvironment}`;
      if (location.pathname !== newPath) {
        navigate(newPath, { replace: true });
      }
    }
  }, [currentEnvironment, urlEnvironment, navigate, location.pathname]);

  const canvasNodes = useMemo(() => (graph?.nodes || []).map((n) => ({
    id: n.id,
    type: n.type as NodeData["type"],
    label: n.name,
    position: n.position,
    status: "pending" as const,
    metadata: n.config || {},
    inputs: [],
    outputs: [],
  })), [graph]);

  const canvasEdges = useMemo(() => (graph?.edges || []).map((e, i) => ({
    id: `edge-${i}`,
    source: e.source,
    target: e.target,
    sourcePort: "output",
    targetPort: "input",
    status: "pending" as const,
  })), [graph]);

  const workflowNodes = graph?.nodes || [];
  const workflowEdges = graph?.edges || [];

  const selectedAgent = useMemo(
    () => agents.find((a) => a.id === selectedAgentId) || null,
    [agents, selectedAgentId]
  );

  const tokenUsageFormatted = {
    inputTokens: tokenUsage?.promptTokens ?? 0,
    outputTokens: tokenUsage?.completionTokens ?? 0,
    totalTokens: tokenUsage?.totalTokens ?? 0,
    costUSD: tokenUsage?.costUsd ?? 0,
    models: {} as Record<string, { calls: number; tokens: number; cost: number }>,
    timeRange: { start: "", end: "" },
  };

  const simulationTokenUsage = {
    totalTokens: tokenUsage?.totalTokens ?? 0,
    promptTokens: tokenUsage?.promptTokens ?? 0,
    completionTokens: tokenUsage?.completionTokens ?? 0,
    costUsd: tokenUsage?.costUsd ?? 0,
  };

  const telemetryMetricsFormatted = metrics.map((m) => ({
    id: `metric-${m.timestamp}`,
    label: "Request Latency",
    value: m.averageLatencyMs,
    unit: "ms",
    timestamp: new Date(m.timestamp).getTime(),
    trend: m.averageLatencyMs > 100 ? ("up" as const) : ("down" as const),
    delta: Math.floor(Math.random() * 20),
  }));

const collabActivity = (collabActivityData?.activities || []) as CollabEvent[];
   const promptVersionsData = [] as { id: string; metadata?: { prompt?: string; user_name?: string; user_color?: string; changes?: string }; created_at: string }[];
   const pairSessionsData = [] as { id: string; metadata?: { host_name?: string; host_color?: string; guest_name?: string; guest_color?: string; status?: string; current_file?: string; current_line?: number }; created_at: string }[];
   const commentsData = [] as { id: string; metadata?: { user_name?: string; user_color?: string; content?: string; line?: number; resolved?: boolean }; created_at: string }[];
   const annotationsData = [] as { id: string; metadata?: { user_name?: string; user_color?: string; target_id?: string; target_type?: string; content?: string; position?: { x: number; y: number }; resolved?: boolean }; created_at: string }[];
   const graphEditsData = [] as { id: string; created_by: string; metadata?: { user_name?: string; node_id?: string; field?: string; old_value?: string; new_value?: string }; created_at: string }[];
   const conflictsData = [] as { id: string; metadata?: { field?: string; current_user?: string; current_value?: string; incoming_user?: string; incoming_value?: string } }[];

  const createTask = useCreateTask();
  const updateTask = useUpdateTask();
  const deleteTask = useDeleteTask();
  const assignTask = useAssignTask();

  const resolveComment = useResolveComment();
  const resolveAnnotation = useResolveAnnotation();
  const endPairSession = useEndPairSession();
  const updatePromptVersion = useUpdatePromptVersion();
  const recordActivity = useRecordActivity();

  const handleAgentSelect = useCallback((agentId: string) => {
    setSelectedAgentId(agentId);
  }, []);

  const handleAgentSpawn = useCallback(async () => {
    try {
      await spawnAgent.mutateAsync({ name: `Agent-${Date.now()}` });
    } catch (error) {
      console.error("Failed to spawn agent:", error);
    }
  }, [spawnAgent]);

  const handleAgentTerminate = useCallback(async (agentId: string) => {
    try {
      await terminateAgent.mutateAsync(agentId);
      setSelectedAgentId(null);
    } catch (error) {
      console.error("Failed to terminate agent:", error);
    }
  }, [terminateAgent]);

  const handleAgentPause = useCallback(async (agentId: string) => {
    try {
      await pauseAgent.mutateAsync(agentId);
    } catch (error) {
      console.error("Failed to pause agent:", error);
    }
  }, [pauseAgent]);

  const handleAgentResume = useCallback(async (agentId: string) => {
    try {
      await resumeAgent.mutateAsync(agentId);
    } catch (error) {
      console.error("Failed to resume agent:", error);
    }
  }, [resumeAgent]);

  const handleAgentRestart = useCallback(async (agentId: string) => {
    await handleAgentTerminate(agentId);
    setTimeout(() => handleAgentSpawn(), 500);
  }, [handleAgentSpawn, handleAgentTerminate]);

  const handleRuntimeSelect = useCallback((runtimeId: string) => {
    setSelectedRuntime({ runtimeId, config: {}, priority: "normal" });
  }, []);

  const handleNodeSelect = useCallback((node: NodeData) => {
    console.log("Node selected:", node);
  }, []);

  const handleNodeDoubleClick = useCallback((node: NodeData) => {
    console.log("Node double clicked:", node);
  }, []);

  const handleCanvasClick = useCallback(() => {
    setSelectedAgentId(null);
  }, []);

  const handleCodeChange = useCallback((newCode: string) => {
    setCode(newCode);
  }, []);

  const handleSimulationConfig = useCallback((config: typeof simulationConfig) => {
    console.log("Simulation config changed:", config);
  }, []);

  const handleSimulationToggle = useCallback(async () => {
    if (simulationRunning && activeSimulation) {
      await abortSimulation.mutateAsync(activeSimulation.id);
    } else {
      await startSimulation.mutateAsync({
        ...simulationConfig,
        name: "Quick Simulation",
      });
    }
  }, [simulationRunning, activeSimulation, abortSimulation, startSimulation, simulationConfig]);

  const handleGhostModeToggle = useCallback(async () => {
    await createBuild.mutateAsync({ name: "New Build" });
  }, [createBuild]);

  const handleWorkflowExecute = useCallback(async () => {
    try {
      await executeWorkflow.mutateAsync({
        name: "default",
        nodes: workflowNodes,
        edges: workflowEdges,
      });
      setLastSaved(new Date());
    } catch (error) {
      console.error("Failed to execute workflow:", error);
    }
  }, [executeWorkflow, workflowNodes, workflowEdges]);

  const handleCommandExecute = useCallback((command: AICommand) => {
    console.log("Command executed:", command);
  }, []);

  const [timelineEvents, setTimelineEvents] = useState<Array<{ id: string; type: string; nodeLabel: string; result: "success" | "failure" | "partial"; timestamp: number; duration: number }>>([]);
  useKeyboardShortcuts([
    {
      key: "k",
      meta: true,
      action: () => setIsCommandPaletteOpen(true),
      description: "Open command palette",
    },
    {
      key: "s",
      meta: true,
      action: () => {
        setLastSaved(new Date());
        console.log("Saved");
      },
      description: "Save",
    },
    {
      key: "z",
      meta: true,
      action: () => console.log("Undo"),
      description: "Undo",
    },
    {
      key: "z",
      meta: true,
      shift: true,
      action: () => console.log("Redo"),
      description: "Redo",
    },
  ]);

  return (
    <div className="studio-root" style={{ fontSize: 'var(--studio-font-size)' }}>
      <StudioShell settingsContent={<div className="h-full bg-bg-primary"><StudioSettingsCenter /></div>}>
        {/* Collaboration bar */}
        <div className="flex items-center justify-between px-4 py-2 bg-bg-tertiary border-b border-border-subtle">
          <div className="flex items-center gap-4">
            <LivePresence presences={collaborators?.map(c => ({
              id: c.id,
              userId: c.id,
              userName: c.name,
              color: c.color,
              status: c.status === "online" ? "active" : "idle",
              lastActivity: Date.now(),
            })) || []} currentUserId="current" />
          </div>
          <div className="flex items-center gap-3">
            <Tooltip content="Share Studio Link">
              <button onClick={() => {
                navigator.clipboard.writeText(window.location.href);
                recordActivity.mutate({ action: 'shared studio link', target: window.location.href, icon: '🔗' });
              }} className="p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-bg-hover transition-colors">
                <Share2 className="size-4" />
              </button>
            </Tooltip>

            {/* User Avatar */}
            {user && (
              <div className="flex items-center gap-2 pl-2 border-l border-border-subtle">
                <div className="relative">
                  {user.avatar ? (
                    <img
                      src={user.avatar}
                      alt={user.username || user.name || user.email}
                      className="w-7 h-7 rounded-full object-cover ring-1 ring-border-subtle"
                    />
                  ) : (
                    <div className="w-7 h-7 rounded-full bg-brand-500/20 flex items-center justify-center ring-1 ring-brand-500/30">
                      <span className="text-[10px] font-semibold text-brand-400">
                        {(user.username || user.name || user.email || 'U')[0].toUpperCase()}
                      </span>
                    </div>
                  )}
                  <div className="absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 bg-green-500 rounded-full border-2 border-bg-tertiary" />
                </div>
                <span className="hidden md:inline text-xs font-medium text-text-secondary truncate max-w-[80px]">
                  {user.username || user.name || user.email?.split('@')[0]}
                </span>
              </div>
            )}
          </div>
        </div>

        <div className="flex-1 flex overflow-hidden">
          <ResizablePanelGroup direction="horizontal">
            {/* Left Panel */}
            <ResizablePanel defaultSize={18} minSize={15} maxSize={35} collapsible onCollapse={setLeftPanelCollapsed}>
              <div className="h-full flex flex-col bg-bg-secondary border-r border-border-subtle">
                {!leftPanelCollapsed && (
                  <>
                    <div className="p-3 border-b border-border-subtle">
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-xs font-semibold text-text-muted uppercase tracking-wider">Explorer</span>
                        <Tooltip content="Command Palette">
                          <button onClick={() => setIsCommandPaletteOpen(true)} className="p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-bg-hover transition-colors">
                            <Search className="size-4" />
                          </button>
                        </Tooltip>
                      </div>
                      <Tabs value={leftPanelTab} onValueChange={(v) => setLeftPanelTab(v as typeof leftPanelTab)}>
                        <TabsList className="w-full grid grid-cols-7 p-0.5 h-auto bg-bg-primary rounded-lg">
                          <TabsTrigger value="agents" className="flex-1 py-1.5 text-[10px]"><Bot className="size-3 mx-auto" /></TabsTrigger>
                          <TabsTrigger value="canvas" className="flex-1 py-1.5 text-[10px]"><GitBranch className="size-3 mx-auto" /></TabsTrigger>
                          <TabsTrigger value="marketplace" className="flex-1 py-1.5 text-[10px]"><Layers className="size-3 mx-auto" /></TabsTrigger>
                          <TabsTrigger value="runtime" className="flex-1 py-1.5 text-[10px]"><Cpu className="size-3 mx-auto" /></TabsTrigger>
                          <TabsTrigger value="swarm" className="flex-1 py-1.5 text-[10px]"><Users className="size-3 mx-auto" /></TabsTrigger>
                          <TabsTrigger value="skills" className="flex-1 py-1.5 text-[10px]"><Wand2 className="size-3 mx-auto" /></TabsTrigger>
                          <TabsTrigger value="extensions" className="flex-1 py-1.5 text-[10px]"><Puzzle className="size-3 mx-auto" /></TabsTrigger>
                        </TabsList>
                      </Tabs>
                    </div>

                    <div className="flex-1 overflow-y-auto">
                      {leftPanelTab === "agents" && (
                        <AgentsPanel
                          selectedAgentId={selectedAgentId}
                          onAgentSelect={handleAgentSelect}
                          onAgentCreate={handleAgentSpawn}
                          onAgentTerminate={handleAgentTerminate}
                          onAgentPause={handleAgentPause}
                          onAgentResume={handleAgentResume}
                          onAgentRestart={handleAgentRestart}
                        />
                      )}
                      {leftPanelTab === "canvas" && (
                        <CanvasPanel
                          canvasNodes={canvasNodes}
                          canvasEdges={canvasEdges}
                          selectedAgent={selectedAgent}
                          showGrid={studioState.gridEnabled}
                          showMinimap={studioState.showMinimap}
                          onNodeSelect={handleNodeSelect}
                          onNodeDoubleClick={handleNodeDoubleClick}
                          onCanvasClick={handleCanvasClick}
                          onAgentCreate={handleAgentSpawn}
                          onAgentPause={handleAgentPause}
                          onAgentResume={handleAgentResume}
                          onAgentTerminate={handleAgentTerminate}
                          onAgentRestart={handleAgentRestart}
                          onOpenDataSourceDialog={() => setDataSourceDialogOpen(true)}
                        />
                      )}
                      {leftPanelTab === "marketplace" && (
                        <MarketplacePanel
                          searchQuery={marketplaceSearch}
                          onSearchChange={setMarketplaceSearch}
                          categoryFilter={marketplaceCategory}
                          onCategoryChange={setMarketplaceCategory}
                          currentUserName="Studio User"
                        />
                      )}
                      {leftPanelTab === "runtime" && (
                        <RuntimePanel
                          selectedRuntime={selectedRuntime}
                          onSelect={handleRuntimeSelect}
                        />
                      )}
                      {leftPanelTab === "swarm" && (
                        <SwarmPanel />
                      )}
                      {leftPanelTab === "skills" && (
                        <SkillsPanel />
                      )}
                      {leftPanelTab === "extensions" && (
                        <div className="h-full"><PluginManager /></div>
                      )}
                    </div>
                  </>
                )}
              </div>
            </ResizablePanel>

            <ResizableHandle />

            {/* Editor Panel */}
            <ResizablePanel defaultSize={50} minSize={30}>
              <ResizablePanelGroup direction="vertical">
                <ResizablePanel defaultSize={65} minSize={30}>
                  <StudioEditor
                    code={code}
                    onChange={handleCodeChange}
                    theme={editorTheme}
                    onThemeChange={setEditorTheme}
                    isGhostModeActive={builds.length > 0 && builds[0]?.status === 'ready'}
                    onToggleGhostMode={handleGhostModeToggle}
                    onRunWorkflow={handleWorkflowExecute}
                    onFormatCode={() => recordActivity.mutate({ action: 'formatted code', target: 'code', icon: '✨' })}
                    onSave={() => console.log('Saved')}
                    onUndo={() => console.log('Undo')}
                    onRedo={() => console.log('Redo')}
                    onCommandPalette={() => setIsCommandPaletteOpen(true)}
                  />
                </ResizablePanel>

                <ResizableHandle />

                <ResizablePanel defaultSize={35} minSize={15} collapsible onCollapse={setBottomPanelCollapsed}>
                  <div className="h-full flex flex-col bg-bg-secondary border-t border-border-subtle">
                    {!bottomPanelCollapsed && (
                      <Tabs value={bottomPanelTab} onValueChange={(v) => setBottomPanelTab(v as typeof bottomPanelTab)} className="h-full flex flex-col">
                        <TabsList className="bg-bg-primary flex-wrap h-auto flex-shrink-0">
                          <TabsTrigger value="execution" className="text-[11px] gap-1.5"><Play className="size-3" /> Exec</TabsTrigger>
                          <TabsTrigger value="simulation" className="text-[11px] gap-1.5"><Gauge className="size-3" /> Sim</TabsTrigger>
                          <TabsTrigger value="ghost" className="text-[11px] gap-1.5"><Brain className="size-3" /> Ghost</TabsTrigger>
                          <TabsTrigger value="tasks" className="text-[11px] gap-1.5"><CheckSquare className="size-3" /> Tasks</TabsTrigger>
                          <TabsTrigger value="devops" className="text-[11px] gap-1.5"><Rocket className="size-3" /> DevOps</TabsTrigger>
                          <TabsTrigger value="memory" className="text-[11px] gap-1.5"><Database className="size-3" /> Memory</TabsTrigger>
                          <TabsTrigger value="robotics" className="text-[11px] gap-1.5"><Bot className="size-3" /> Robotics</TabsTrigger>
                        </TabsList>
                        <div className="flex-1 overflow-auto">
                          <TabsContent value="execution" className="h-full">
                            <ExecutionPanel
                              events={timelineEvents}
                              onRunFirstWorkflow={handleWorkflowExecute}
                            />
                          </TabsContent>
                          <TabsContent value="simulation" className="h-full">
                            <SimulationPanel
                              config={simulationConfig}
                              result={activeSimulation ? {
                                id: activeSimulation.id,
                                status: activeSimulation.status === 'completed' ? 'completed' : 'running',
                              } : undefined}
                              isRunning={simulationRunning}
                              tokenUsage={simulationTokenUsage}
                              onConfigChange={handleSimulationConfig}
                              onToggle={handleSimulationToggle}
                              onRefresh={refreshAgents}
                            />
                          </TabsContent>
                          <TabsContent value="ghost" className="h-full">
                            <GhostPanel
                              build={builds.length > 0 ? {
                                id: builds[0].id,
                                goal: builds[0].name,
                                phase: builds[0].status === 'ready' ? 'complete' : 'building',
                                progress: builds[0].status === 'ready' ? 100 : 0,
                                startedAt: builds[0].createdAt,
                                updatedAt: builds[0].updatedAt,
                              } : undefined}
                              tasks={tasks}
                              onCancelBuild={() => console.log('Cancel build')}
                              onCreateBuild={handleGhostModeToggle}
                            />
                          </TabsContent>
                          <TabsContent value="tasks" className="h-full">
                            <TasksPanel
                              tasks={tasks.map(t => ({
                                id: t.id,
                                title: t.title,
                                description: t.description || '',
                                status: t.status === 'in_progress' ? 'in-progress' : t.status === 'completed' ? 'done' : t.status === 'failed' ? 'blocked' : 'todo',
                                priority: 'medium' as const,
                                createdAt: t.createdAt,
                                updatedAt: t.updatedAt,
                              }))}
                              onTaskCreate={(task) => createTask.mutate({ title: task.title, description: task.description })}
                              onTaskUpdate={({ id, updates }) => updateTask.mutate({ taskId: id, updates: { title: updates.title, description: updates.description } })}
                              onTaskDelete={(id) => deleteTask.mutate(id)}
                              onTaskAssign={({ id, agentId }) => assignTask.mutate({ taskId: id, assigneeId: agentId || '' })}
                            />
                          </TabsContent>
                          <TabsContent value="devops" className="h-full">
                            <DevOpsPanel />
                          </TabsContent>
                          <TabsContent value="memory" className="h-full">
                            <MemoryPanel />
                          </TabsContent>
                          <TabsContent value="robotics" className="h-full">
                            <RoboticsPanel />
                          </TabsContent>
                        </div>
                      </Tabs>
                    )}
                  </div>
                </ResizablePanel>
              </ResizablePanelGroup>
            </ResizablePanel>

            <ResizableHandle />

            {/* Right Panel */}
            <ResizablePanel defaultSize={32} minSize={20} maxSize={45} collapsible onCollapse={setRightPanelCollapsed}>
              <div className="h-full flex flex-col bg-bg-secondary border-l border-border-subtle">
                {!rightPanelCollapsed && (
                  <Tabs value={rightPanelTab} onValueChange={(v) => setRightPanelTab(v as typeof rightPanelTab)} className="h-full">
                    <TabsList className="bg-bg-primary flex-wrap h-auto">
                      <TabsTrigger value="telemetry" className="text-[11px] gap-1.5"><Activity className="size-3" /> Telemetry</TabsTrigger>
                      <TabsTrigger value="visualization" className="text-[11px] gap-1.5"><Eye className="size-3" /> 3D View</TabsTrigger>
                      <TabsTrigger value="profiler" className="text-[11px] gap-1.5"><BarChart2 className="size-3" /> Profiler</TabsTrigger>
                      <TabsTrigger value="collab" className="text-[11px] gap-1.5"><Users className="size-3" /> Collab</TabsTrigger>
                    </TabsList>
                    <div className="flex-1 overflow-auto">
                      <TabsContent value="telemetry">
                        <TelemetryPanel
                          metrics={telemetryMetricsFormatted}
                          tokenUsage={tokenUsageFormatted}
                          onMetricClick={(m) => recordActivity.mutate({ action: 'viewed metric', target: m.id, icon: '📈' })}
                        />
                      </TabsContent>
                      <TabsContent value="visualization">
                        <VisualizationPanel
                          nodes={canvasNodes.map((n) => ({
                            id: n.id,
                            label: n.label,
                            type: n.type,
                            status: n.status,
                          }))}
                          edges={canvasEdges.map((e) => ({ source: e.source, target: e.target, strength: 0.8 }))}
                        />
                      </TabsContent>
                      <TabsContent value="profiler">
                        <ProfilerPanel
                          executions={executions.map((ex, i) => ({
                            id: ex.id || `ex-${i}`,
                            graphId: ex.graphId,
                            nodeResults: ex.nodeResults,
                            startedAt: ex.startedAt,
                            status: ex.status === 'completed' ? 'completed' : 'failed',
                          }))}
                        />
                      </TabsContent>
                      <TabsContent value="collab">
<CollabPanel
                            collaborators={collaborators}
                            currentUser={currentUser}
                            collabActivityData={collabActivity}
                            promptVersionsData={promptVersionsData}
                            pairSessionsData={pairSessionsData}
                            commentsData={commentsData}
                            annotationsData={annotationsData}
                            graphEditsData={graphEditsData}
                            conflictsData={conflictsData}
                            teamMemories={teamMemories?.memories || []}
                            executions={executions}
                            onResolveComment={(commentId, resolved) => resolveComment.mutate({ commentId, resolved })}
                            onResolveAnnotation={(annotationId, resolved) => resolveAnnotation.mutate({ annotationId, resolved })}
                            onEndPairSession={(sessionId) => endPairSession.mutate(sessionId)}
                            onUpdatePromptVersion={(prompt, changes) => updatePromptVersion.mutate({ prompt, changes })}
                            onRecordActivity={recordActivity.mutate}
                          />
                      </TabsContent>
                    </div>
                  </Tabs>
                )}
              </div>
            </ResizablePanel>
          </ResizablePanelGroup>
        </div>

        <StatusBar
          isConnected={isConnected}
          isSynced={true}
          runtimeName={selectedRuntime?.runtimeId || "wasm"}
          latencyMs={latencyStats.avg}
          lastSyncTime={lastSaved}
          activeAgents={agents.filter((a) => a.status === "active").length}
          activeExecutions={executions.filter((e) => e.status !== "completed").length}
        />

        <AICommandPalette
          isOpen={isCommandPaletteOpen}
          onClose={() => setIsCommandPaletteOpen(false)}
          onExecute={handleCommandExecute}
          commands={[
            { id: "cmd-1", name: "Run Workflow", description: "Execute the current workflow", shortcut: "⌘↵", category: "execution", icon: <Play className="size-4" /> },
            { id: "cmd-2", name: "Format Code", description: "Format the editor content", shortcut: "⇧⌘F", category: "editor", icon: <Code className="size-4" /> },
            { id: "cmd-3", name: "Create Agent", description: "Start a new agent", shortcut: "⌘A", category: "agent", icon: <Bot className="size-4" /> },
            { id: "cmd-4", name: "Open Marketplace", description: "Browse function marketplace", shortcut: "⌘M", category: "navigation", icon: <Layers className="size-4" /> },
            { id: "cmd-5", name: "Start Simulation", description: "Run workflow simulation", shortcut: "⌘S", category: "simulation", icon: <Gauge className="size-4" /> },
            { id: "cmd-6", name: "Ghost Mode", description: "Toggle autonomous building", shortcut: "⌘G", category: "automation", icon: <Activity className="size-4" /> },
            { id: "cmd-7", name: "View Telemetry", description: "Open live telemetry panel", shortcut: "⌘T", category: "observability", icon: <Activity className="size-4" /> },
            { id: "cmd-8", name: "AI Assist", description: "Get AI-powered assistance", shortcut: "⌘K", category: "ai", icon: <Brain className="size-4" /> },
            { id: "cmd-9", name: "Save", description: "Save current file", shortcut: "⌘S", category: "editor", icon: <Save className="size-4" /> },
            { id: "cmd-10", name: "Undo", description: "Undo last action", shortcut: "⌘Z", category: "editor", icon: <Undo2 className="size-4" /> },
            { id: "cmd-11", name: "Redo", description: "Redo last action", shortcut: "⇧⌘Z", category: "editor", icon: <Redo2 className="size-4" /> },
          ]}
        />

        <DataSourceConfigDialog
          open={dataSourceDialogOpen}
          onOpenChange={setDataSourceDialogOpen}
          onSave={(sources) => {
            setDataSources(sources);
            console.log('Data sources saved:', sources);
          }}
          sources={dataSources}
        />
      </StudioShell>
    </div>
  );
}

function useStudioCollaborators() {
  const { onlineUsers, myPresence } = usePresence();
  const { data: teamsData } = useTeams();
  const firstTeamId = teamsData?.teams?.[0]?.id ?? '';
  const { data: teamMembersData } = useTeamMembers(firstTeamId);

  const collaborators = useMemo(() => {
    if (!teamMembersData?.members) return [];
    return teamMembersData.members.map((m) => {
      const isOnline = onlineUsers.some((u) => u.userId === m.user_id);
      return {
        id: m.user_id,
        name: m.user?.name || m.user?.username || m.user?.email || 'Unknown',
        avatar: m.user?.avatar,
        color: isOnline ? '#10b981' : '#6b7280',
        status: isOnline ? ('online' as const) : ('offline' as const),
        lastActive: new Date().toISOString(),
        role: m.role,
      };
    });
  }, [teamMembersData?.members, onlineUsers]);

  const currentUser = useMemo(() => ({
    name: myPresence?.displayName || myPresence?.username || 'You',
    color: '#f97316',
  }), [myPresence]);

  return { collaborators, currentUser };
}

type AgentData = {
  id: string;
  name: string;
  role: string;
  status: 'running' | 'idle' | 'paused' | 'stopped' | 'error';
  memoryUsage: number;
  memoryLimit: number;
  executionBudget: number;
  executionBudgetUsed: number;
  permissions: string[];
  tools: string[];
  runtime: string;
  model: string;
  uptime: number;
  tasksCompleted: number;
  tasksFailed: number;
  avgLatency: number;
  lastHeartbeat: string;
  createdAt: string;
  description: string;
  tags: string[];
};

type RuntimeSelection = {
  runtimeId: string;
  config: Record<string, unknown>;
  priority: string;
};

interface SimulationConfig {
  name: string;
  iterations: number;
  duration: number;
  stressLevel?: "low" | "medium" | "high" | "extreme";
}

type TelemetryMetric = {
  id: string;
  label: string;
  value: number;
  unit: string;
  timestamp: number;
  trend: 'up' | 'down';
  delta: number;
};

type TokenUsage = {
  totalTokens: number;
  promptTokens: number;
  completionTokens: number;
  costUsd: number;
  models: Record<string, unknown>;
  timeRange: { start: string; end: string };
};