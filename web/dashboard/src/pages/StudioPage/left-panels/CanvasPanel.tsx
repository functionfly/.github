import { useStudioMemory } from "@/hooks/useStudio";
import type { AgentMemory } from "@/types";
import {
  AgentLifecyclePanel,
  AgentMemoryViewer,
  type AgentData,
} from "@functionfly/ui-agent";
import {
  Badge,
  cn,
  GlassCard,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@functionfly/ui-core";
import {
  FunctionCanvas,
  type NodeData,
  type EdgeData,
  type NodeType,
} from "@functionfly/ui-graph";
import { Database, Plus, Settings } from "lucide-react";
import { useState } from "react";

interface CanvasPanelProps {
  canvasNodes: Array<{
    id: string;
    type: NodeType;
    label: string;
    position?: { x: number; y: number };
    status: "idle" | "running" | "completed" | "error" | "waiting";
    inputs?: unknown[];
    outputs?: unknown[];
  }>;
  canvasEdges: Array<{
    id: string;
    source: string;
    target: string;
    sourcePort?: string;
    targetPort?: string;
    status: "idle" | "active" | "error";
  }>;
  selectedAgent: AgentData | null;
  showGrid: boolean;
  showMinimap: boolean;
  onNodeSelect: (node: NodeData) => void;
  onNodeDoubleClick: (node: NodeData) => void;
  onCanvasClick: () => void;
  onAgentCreate: () => void;
  onAgentPause: (agentId: string) => void;
  onAgentResume: (agentId: string) => void;
  onAgentTerminate: (agentId: string) => void;
  onAgentRestart: (agentId: string) => void;
  onOpenDataSourceDialog: () => void;
}

type CanvasTab = "design" | "logic" | "data";

interface LogicFlowEditorProps {
  onBackToDesign: () => void;
}

function LogicFlowEditor({ onBackToDesign }: LogicFlowEditorProps) {
  const [nodes, setNodes] = useState([
    { id: "1", type: "condition", label: "Check Input", x: 100, y: 50 },
    { id: "2", type: "action", label: "Process Data", x: 300, y: 50 },
    { id: "3", type: "action", label: "Enrich Context", x: 500, y: 50 },
  ]);
  const [connections, setConnections] = useState([
    { from: "1", to: "2", condition: "true" },
    { from: "1", to: "3", condition: "false" },
  ]);
  const [selectedNode, setSelectedNode] = useState<string | null>(null);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h4 className="text-sm font-medium">Logic Flow Editor</h4>
          <p className="text-xs text-text-muted">Define conditional branching and control flow</p>
        </div>
        <button
          onClick={onBackToDesign}
          className="text-xs text-brand-400 hover:text-brand-300 flex items-center gap-1"
        >
          <Settings className="size-3" /> Back to Design
        </button>
      </div>

      <div className="border border-border-subtle rounded-lg p-4 bg-bg-primary min-h-[200px]">
        <div className="flex flex-wrap gap-2 mb-4">
          {nodes.map((node) => (
            <div
              key={node.id}
              className={cn(
                "px-3 py-2 rounded-lg border text-xs font-medium cursor-pointer transition-all",
                selectedNode === node.id
                  ? "border-brand-500 bg-brand-500/10"
                  : "border-border-subtle bg-bg-secondary hover:border-brand-500/50"
              )}
              onClick={() => setSelectedNode(node.id)}
            >
              <span className="text-text-muted mr-2">{node.type}</span>
              {node.label}
            </div>
          ))}
          <button
            className="px-3 py-2 rounded-lg border border-dashed border-border-subtle text-xs text-text-muted hover:border-brand-500/50 hover:text-brand-400 flex items-center gap-1"
            onClick={() => {
              const newId = String(nodes.length + 1);
              setNodes([
                ...nodes,
                { id: newId, type: "action", label: `Node ${newId}`, x: 0, y: 0 },
              ]);
            }}
          >
            <Plus className="size-3" /> Add Node
          </button>
        </div>

        {connections.length > 0 && (
          <div className="text-xs text-text-muted">
            <span className="font-medium">Connections:</span>{" "}
            {connections.map((c, i) => (
              <span key={i}>
                {c.from} → {c.to}
                {c.condition ? ` (${c.condition})` : ""}
                {i < connections.length - 1 ? ", " : ""}
              </span>
            ))}
          </div>
        )}
      </div>

      <div className="flex gap-2">
        <button className="flex-1 px-3 py-2 text-xs bg-brand-500 text-white rounded-lg hover:bg-brand-600 transition-colors">
          Validate Flow
        </button>
        <button className="flex-1 px-3 py-2 text-xs border border-border-subtle rounded-lg hover:bg-bg-hover transition-colors">
          Export Rules
        </button>
      </div>
    </div>
  );
}

interface DataFlowMapperProps {
  onConfigureSources: () => void;
}

function DataFlowMapper({ onConfigureSources }: DataFlowMapperProps) {
  const [mappings, setMappings] = useState([
    { source: "input.text", target: "parser.input", transform: "trim()" },
    { source: "context.user", target: "enricher.userContext", transform: "none" },
    { source: "parser.output", target: "generator.prompt", transform: "template" },
  ]);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h4 className="text-sm font-medium">Data Flow Mapper</h4>
          <p className="text-xs text-text-muted">Connect data sources to workflow inputs</p>
        </div>
        <button
          onClick={onConfigureSources}
          className="text-xs text-brand-400 hover:text-brand-300 flex items-center gap-1"
        >
          <Database className="size-3" /> Configure Sources
        </button>
      </div>

      <div className="space-y-2">
        {mappings.map((mapping, idx) => (
          <div
            key={idx}
            className="flex items-center gap-2 p-2 border border-border-subtle rounded-lg bg-bg-primary"
          >
            <div className="flex-1 min-w-0">
              <div className="text-xs font-mono text-brand-400 truncate">{mapping.source}</div>
            </div>
            <div className="text-text-muted">→</div>
            <div className="flex-1 min-w-0">
              <div className="text-xs font-mono text-success truncate">{mapping.target}</div>
            </div>
            <Badge variant="ghost" size="sm" className="text-[10px]">
              {mapping.transform}
            </Badge>
          </div>
        ))}
      </div>

      <button className="w-full px-3 py-2 text-xs border border-dashed border-border-subtle rounded-lg text-text-muted hover:border-brand-500/50 hover:text-brand-400 flex items-center justify-center gap-1">
        <Plus className="size-3" /> Add Mapping
      </button>
    </div>
  );
}

export function CanvasPanel({
  canvasNodes,
  canvasEdges,
  selectedAgent,
  showGrid,
  showMinimap,
  onNodeSelect,
  onNodeDoubleClick,
  onCanvasClick,
  onAgentCreate,
  onAgentPause,
  onAgentResume,
  onAgentTerminate,
  onAgentRestart,
  onOpenDataSourceDialog,
}: CanvasPanelProps) {
  const [canvasTab, setCanvasTab] = useState<CanvasTab>("design");
  const agentMemories = useStudioMemory(selectedAgent?.id || undefined);

  return (
    <div className="p-3 space-y-3">
      <GlassCard className="p-3">
        <div className="flex items-center gap-2 mb-2">
          <div className="w-4 h-4 rounded bg-brand-500/20 flex items-center justify-center">
            <div className="w-2 h-2 rounded bg-brand-500" />
          </div>
          <span className="text-sm font-medium">Workflow Graph</span>
          <Badge variant="success" size="sm" className="ml-auto">
            {canvasNodes.length} nodes
          </Badge>
        </div>
        <p className="text-xs text-text-muted mb-3">Visual representation of your workflow</p>
        <Tabs
          value={canvasTab}
          onValueChange={(v) => setCanvasTab(v as CanvasTab)}
          key={`canvas-tabs-${canvasTab}`}
        >
          <TabsList className="w-full mb-2">
            <TabsTrigger value="design" className="flex-1 text-[10px]">
              Design
            </TabsTrigger>
            <TabsTrigger value="logic" className="flex-1 text-[10px]">
              Logic
            </TabsTrigger>
            <TabsTrigger value="data" className="flex-1 text-[10px]">
              Data
            </TabsTrigger>
          </TabsList>
          <TabsContent value="design" className="mt-0">
            <FunctionCanvas
              nodes={canvasNodes}
              edges={canvasEdges}
              viewMode="design"
              onNodeSelect={onNodeSelect}
              onNodeDoubleClick={onNodeDoubleClick}
              onCanvasClick={onCanvasClick}
              showGrid={showGrid}
              showMinimap={showMinimap}
              className="h-[300px] border border-border-subtle rounded-lg"
            />
          </TabsContent>
          <TabsContent value="logic" className="mt-0">
            <LogicFlowEditor onBackToDesign={() => setCanvasTab("design")} />
          </TabsContent>
          <TabsContent value="data" className="mt-0">
            <DataFlowMapper onConfigureSources={onOpenDataSourceDialog} />
</TabsContent>
        </Tabs>
      </GlassCard>

      {selectedAgent && (
        <AgentLifecyclePanel
          agent={selectedAgent}
          onSpawn={onAgentCreate}
          onPause={() => onAgentPause(selectedAgent.id)}
          onResume={() => onAgentResume(selectedAgent.id)}
          onTerminate={() => onAgentTerminate(selectedAgent.id)}
          onRestart={() => onAgentRestart(selectedAgent.id)}
        />
      )}

      {agentMemories.memories && agentMemories.memories.length > 0 && (
        <div className="px-3 pb-3">
          <AgentMemoryViewer
            agentId={selectedAgent.id}
            memories={(agentMemories.memories || []).map((m) => ({
              id: m.id,
              type: m.memory_type as "working" | "longterm" | "context" | "episodic",
              content: m.content || "",
              importance: m.importance_score,
              lastAccessed: m.last_accessed_at || new Date().toISOString(),
              createdAt: m.created_at,
            }))}
            onMemoryAdd={() => {
              // Placeholder - component doesn't provide content/type on click
            }}
            onMemorySearch={(q) => agentMemories.searchMemories.mutate(q)}
            onMemoryDelete={(id) => agentMemories.deleteMemory.mutate(id)}
          />
        </div>
      )}
    </div>
  );
}