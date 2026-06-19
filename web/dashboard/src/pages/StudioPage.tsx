import type { CollabEvent } from "@/api/studioCollab";
import { useTeamMemories } from "@/hooks/use-team-memory";
import { useActiveEnvironment } from "@/hooks/useActiveEnvironment";
import { usePresence } from "@/hooks/usePresence";
import {
  useStudioAgents,
  useStudioGhost,
  useStudioMemory,
  useStudioSimulation,
  useStudioState,
  useStudioTelemetry,
  useStudioWorkflow,
  type AgentStatus,
} from "@/hooks/useStudio";
import { useStudioCollabActivity, useStudioCollabEvents } from "@/hooks/useStudioCollab";
import {
  useEndPairSession,
  useRecordActivity,
  useResolveAnnotation,
  useResolveComment,
  useUpdatePromptVersion
} from "@/hooks/useStudioCollabActions";
import { useStudioCollabData } from "@/hooks/useStudioCollabData";
import {
  useAssignTask,
  useCreateTask,
  useDeleteTask,
  useUpdateTask,
} from "@/hooks/useStudioTask";
import type { AgentData } from "@functionfly/ui-agent";
import { useTeamMembers, useTeams } from "@/hooks/useTeams";
import {
  StudioSettingsCenter
} from "@/pages/StudioPage/components";
import { DataSourceConfigDialog, type DataSource } from "@/pages/StudioPage/components/DataSourceConfigDialog";
import { PluginManager } from "@/pages/StudioPage/components/PluginManager";
import { useAuthStore } from "@/stores/authStore";
import {
  AICommandPalette,
  type AICommand,
} from "@functionfly/ui-ai";
import { LivePresence } from "@functionfly/ui-collaboration";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
  StudioShell,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
  Tooltip
} from "@functionfly/ui-core";
import {
  type NodeData,
  type NodeType
} from "@functionfly/ui-graph";
import {
  Activity,
  BarChart2,
  Bot,
  Brain,
  CheckSquare,
  Code,
  Cpu,
  Database,
  Eye,
  Gauge,
  GitBranch,
  Layers,
  Play,
  Puzzle,
  Redo2,
  Rocket,
  Save,
  Search,
  Share2,
  Undo2,
  Users,
  Wand2
} from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useLocation, useNavigate, useParams } from "react-router-dom";

// Import extracted panels
import {
  DevOpsPanel,
  ExecutionPanel,
  GhostPanel,
  MemoryPanel,
  SimulationPanel,
  TasksPanel
} from "@/pages/StudioPage/bottom-panels";
import { StudioEditor } from "@/pages/StudioPage/editor/StudioEditor";
import { useKeyboardShortcuts } from "@/pages/StudioPage/hooks";
import { AgentsPanel } from "@/pages/StudioPage/left-panels/AgentsPanel";
import { CanvasPanel } from "@/pages/StudioPage/left-panels/CanvasPanel";
import { MarketplacePanel } from "@/pages/StudioPage/left-panels/MarketplacePanel";
import { RuntimePanel } from "@/pages/StudioPage/left-panels/RuntimePanel";
import { SkillsPanel } from "@/pages/StudioPage/left-panels/SkillsPanel";
import { SwarmPanel } from "@/pages/StudioPage/left-panels/SwarmPanel";
import {
  CollabPanel,
  ProfilerPanel,
  TelemetryPanel,
  VisualizationPanel,
} from "@/pages/StudioPage/right-panels";
import { StatusBar } from "@/pages/StudioPage/StatusBar";

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
    status: agent.status === 'active' ? 'running' : agent.status === 'pending' ? 'idle' : agent.status === 'terminated' ? 'terminated' : agent.status === 'terminating' ? 'terminated' : agent.status as 'running' | 'idle' | 'paused' | 'terminated' | 'spawning' | 'error',
    memoryUsage: 0,
    memoryLimit: 512 * 1024 * 1024,
    executionBudget: 10.0,
    executionBudgetUsed: 0,
    permissions: [] as unknown as AgentData['permissions'],
    tools: [] as unknown as AgentData['tools'],
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

  const [codeVersion, setCodeVersion] = useState(1);

  const [lastSaved, setLastSaved] = useState<Date | null>(null);
  const [isConnected, setIsConnected] = useState(true);

  const { state: studioState, setSelectedPanel, toggleMinimap, toggleGrid } = useStudioState();
  const { collaborators, currentUser } = useStudioCollaborators();
  const { user } = useAuthStore();

  const { agents, isLoading: agentsLoading, refreshAgents, spawnAgent, pauseAgent, resumeAgent, terminateAgent } = useStudioAgents();
  const { graph, executions, isLoadingGraph, isLoadingExecutions, executeWorkflow, formatCode, saveCode, undoCode, redoCode, getVersionHistory } = useStudioWorkflow();
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
    type: n.type as NodeType,
    label: n.name,
    position: n.position,
    status: "idle" as const,
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
    status: "idle" as const,
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
  const {
    promptVersionsData,
    pairSessionsData,
    commentsData,
    annotationsData,
    graphEditsData,
    conflictsData,
  } = useStudioCollabData();

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
    setSelectedRuntime({ runtimeId, config: {}, priority: "speed" });
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
    } else if (graph) {
      await startSimulation.mutateAsync({
        graph,
        iterations: simulationConfig.iterations,
      });
    }
  }, [simulationRunning, activeSimulation, abortSimulation, startSimulation, graph, simulationConfig.iterations]);

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

  // ── Code Editor Handlers ────────────────────────────────────────────────

  const handleFormatCode = useCallback(async () => {
    try {
      const result = await formatCode.mutateAsync({
        code,
        language: 'typescript',
        file_path: 'main.ts',
      });
      setCode(result.formatted);
      setCodeVersion(result.version);
    } catch (error) {
      console.error("Failed to format code:", error);
    }
  }, [formatCode, code]);

  const handleSaveCode = useCallback(async () => {
    try {
      const result = await saveCode.mutateAsync({
        code,
        file_path: 'main.ts',
        metadata: { saved_at: new Date().toISOString() },
      });
      setCodeVersion(result.version);
      setLastSaved(new Date());
    } catch (error) {
      console.error("Failed to save code:", error);
    }
  }, [saveCode, code]);

  const handleUndo = useCallback(async () => {
    try {
      const result = await undoCode.mutateAsync({
        file_path: 'main.ts',
        current_version: codeVersion,
      });
      if (result.available && result.code) {
        setCode(result.code);
        setCodeVersion(result.version);
      }
    } catch (error) {
      console.error("Failed to undo:", error);
    }
  }, [undoCode, codeVersion]);

  const handleRedo = useCallback(async () => {
    try {
      const result = await redoCode.mutateAsync({
        file_path: 'main.ts',
        current_version: codeVersion,
      });
      if (result.available && result.code) {
        setCode(result.code);
        setCodeVersion(result.version);
      }
    } catch (error) {
      console.error("Failed to redo:", error);
    }
  }, [redoCode, codeVersion]);

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
            <ResizablePanel defaultSize={18} minSize={15} maxSize={35}>
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
                          selectedAgent={selectedAgent ? agentStatusToAgentData(selectedAgent) : null}
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
                    onFormatCode={handleFormatCode}
                    onSave={handleSaveCode}
                    onUndo={handleUndo}
                    onRedo={handleRedo}
                    onCommandPalette={() => setIsCommandPaletteOpen(true)}
                  />
                </ResizablePanel>

                <ResizableHandle />

                <ResizablePanel defaultSize={35} minSize={15}>
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
                          <TabsTrigger value="robotics" className="text-[11px] gap-1.5 relative" disabled>
                            <Bot className="size-3" /> Robotics
                            <span className="absolute -top-1 -right-1 text-[8px] bg-amber-500/20 text-amber-400 px-1 rounded">Soon</span>
                          </TabsTrigger>
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
                              result={activeSimulation ?? undefined}
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
                                ...builds[0],
                                goal: builds[0].goal || builds[0].name,
                                phase: builds[0].phase || (builds[0].status === 'ready' ? 'complete' : builds[0].status === 'building' ? 'building' : 'planning'),
                                progress: builds[0].progress ?? (builds[0].status === 'ready' ? 100 : 0),
                                started_at: builds[0].started_at || builds[0].createdAt,
                                updated_at: builds[0].updated_at || builds[0].updatedAt,
                              } : undefined}
                              tasks={builds.length > 0 && (builds[0].tasks ?? []).length > 0 ? builds[0].tasks : tasks}
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
                                createdAt: t.started_at || '',
                                updatedAt: t.updated_at || '',
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
                          <TabsContent value="robotics" className="h-full flex items-center justify-center">
                            <div className="text-center space-y-3">
                              <div className="inline-flex items-center justify-center w-12 h-12 rounded-full bg-amber-500/10">
                                <Bot className="size-6 text-amber-400" />
                              </div>
                              <div>
                                <p className="text-sm font-medium text-text-primary">Robotics Control</p>
                                <p className="text-xs text-text-muted">Coming in a future update</p>
                              </div>
                            </div>
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
            <ResizablePanel defaultSize={32} minSize={20} maxSize={45}>
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
                          isLoading={telemetryLoading}
                        />
                      </TabsContent>
                      <TabsContent value="visualization">
                        <VisualizationPanel
                          nodes={canvasNodes.map((n) => ({
                            id: n.id,
                            label: n.label,
                            type: n.type,
                            status: (n.status === 'idle' ? 'pending' : n.status) as 'pending' | 'running' | 'success' | 'error',
                          }))}
                          edges={canvasEdges.map((e) => ({ source: e.source, target: e.target, strength: 0.8 }))}
                          isLoading={isLoadingGraph}
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
                          isLoading={isLoadingExecutions}
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
                            annotationsData={annotationsData as any}
                            graphEditsData={graphEditsData}
                            conflictsData={conflictsData}
                            teamMemories={teamMemories?.memories || []}
                            executions={executions}
                            onResolveComment={(commentId, resolved) => resolveComment.mutate({ commentId, resolved })}
                            onResolveAnnotation={(annotationId, resolved) => resolveAnnotation.mutate({ annotationId, resolved })}
                            onEndPairSession={(sessionId) => endPairSession.mutate(sessionId)}
                            onUpdatePromptVersion={(prompt, changes) => updatePromptVersion.mutate({ prompt, changes })}
                            onRecordActivity={recordActivity.mutate}
                            isLoading={false}
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
          commands={[
            { id: "cmd-1", label: "Run Workflow", description: "Execute the current workflow", shortcut: "⌘↵", category: "execution", icon: <Play className="size-4" />, action: () => handleWorkflowExecute() },
            { id: "cmd-2", label: "Format Code", description: "Format the editor content", shortcut: "⇧⌘F", category: "editor", icon: <Code className="size-4" />, action: () => {} },
            { id: "cmd-3", label: "Create Agent", description: "Start a new agent", shortcut: "⌘A", category: "agent", icon: <Bot className="size-4" />, action: () => handleAgentSpawn() },
            { id: "cmd-4", label: "Open Marketplace", description: "Browse function marketplace", shortcut: "⌘M", category: "navigation", icon: <Layers className="size-4" />, action: () => setLeftPanelTab("marketplace") },
            { id: "cmd-5", label: "Start Simulation", description: "Run workflow simulation", shortcut: "⌘S", category: "simulation", icon: <Gauge className="size-4" />, action: () => {} },
            { id: "cmd-6", label: "Ghost Mode", description: "Toggle autonomous building", shortcut: "⌘G", category: "automation", icon: <Activity className="size-4" />, action: () => {} },
            { id: "cmd-7", label: "View Telemetry", description: "Open live telemetry panel", shortcut: "⌘T", category: "observability", icon: <Activity className="size-4" />, action: () => setRightPanelTab("telemetry") },
            { id: "cmd-8", label: "AI Assist", description: "Get AI-powered assistance", shortcut: "⌘K", category: "ai", icon: <Brain className="size-4" />, action: () => {} },
            { id: "cmd-9", label: "Save", description: "Save current file", shortcut: "⌘S", category: "editor", icon: <Save className="size-4" />, action: () => saveCode.mutateAsync({ code, file_path: 'main.ts' }) },
            { id: "cmd-10", label: "Undo", description: "Undo last action", shortcut: "⌘Z", category: "editor", icon: <Undo2 className="size-4" />, action: () => undoCode.mutateAsync({ file_path: 'main.ts', current_version: codeVersion }) },
            { id: "cmd-11", label: "Redo", description: "Redo last action", shortcut: "⇧⌘Z", category: "editor", icon: <Redo2 className="size-4" />, action: () => redoCode.mutateAsync({ file_path: 'main.ts', current_version: codeVersion }) },
          ]}
          promptTemplates={[]}
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

type RuntimeSelection = {
  runtimeId: string;
  config: Record<string, unknown>;
  priority: "cost" | "speed" | "reliability" | "privacy";
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