import React from "react";
import {
  MemoryGraph,
  SemanticMemoryViewer,
  LongTermContextExplorer,
} from "@functionfly/ui-memory";
import { Brain, Network, Archive } from "lucide-react";

export function MemoryPanel() {
  return (
    <div className="p-3 space-y-4">
      <div className="border-b border-border-subtle pb-3">
        <h3 className="text-sm font-medium mb-1">Agent Memory</h3>
        <p className="text-xs text-text-muted">View and manage agent memory systems</p>
      </div>

      <div className="grid grid-cols-3 gap-2 mb-4">
        <div className="bg-bg-primary rounded-lg border border-border-subtle p-3 text-center">
          <Brain className="size-5 text-brand-400 mx-auto mb-1" />
          <div className="text-lg font-semibold">2.4K</div>
          <div className="text-[10px] text-text-muted">Memories</div>
        </div>
        <div className="bg-bg-primary rounded-lg border border-border-subtle p-3 text-center">
          <Network className="size-5 text-success mx-auto mb-1" />
          <div className="text-lg font-semibold">156</div>
          <div className="text-[10px] text-text-muted">Connections</div>
        </div>
        <div className="bg-bg-primary rounded-lg border border-border-subtle p-3 text-center">
          <Archive className="size-5 text-warning mx-auto mb-1" />
          <div className="text-lg font-semibold">89%</div>
          <div className="text-[10px] text-text-muted">Retention</div>
        </div>
      </div>

      <div className="h-64 overflow-hidden">
        <MemoryGraph
          className="h-full"
          nodes={[
            {
              id: "1",
              label: "Agent Core",
              type: "agent",
              timestamp: Date.now(),
              importance: 0.9,
              connections: ["2", "3"],
              content: "Core agent memory",
            },
            {
              id: "2",
              label: "Workflow Memory",
              type: "code",
              timestamp: Date.now() - 3600000,
              importance: 0.7,
              connections: ["3"],
              content: "Workflow execution patterns",
            },
            {
              id: "3",
              label: "User Context",
              type: "conversation",
              timestamp: Date.now() - 7200000,
              importance: 0.6,
              connections: [],
              content: "User preferences and context",
            },
          ]}
          edges={[
            { id: "e1", source: "1", target: "2", type: "references" },
            { id: "e2", source: "1", target: "3", type: "related_to" },
            { id: "e3", source: "2", target: "3", type: "associated_with" },
          ]}
          onNodeSelect={(node) => console.log("Selected node:", node)}
        />
      </div>

      <div className="border-t border-border-subtle pt-4 space-y-4">
        <h4 className="text-xs font-medium mb-2">Semantic Memory</h4>
        <SemanticMemoryViewer
          entries={[
            {
              id: "1",
              content: "User prefers dark mode interfaces",
              semanticType: "preference",
              confidence: 0.95,
              source: "user_profile",
              timestamp: Date.now() - 86400000,
              lastAccessed: Date.now() - 3600000,
              accessCount: 12,
              tags: ["ui", "preference"],
            },
            {
              id: "2",
              content: "API calls should be cached for 5 minutes",
              semanticType: "procedure",
              confidence: 0.88,
              source: "documentation",
              timestamp: Date.now() - 172800000,
              lastAccessed: Date.now() - 7200000,
              accessCount: 8,
              tags: ["api", "caching"],
            },
            {
              id: "3",
              content: "Database connection pool max size is 20",
              semanticType: "fact",
              confidence: 0.99,
              source: "config",
              timestamp: Date.now() - 259200000,
              tags: ["database", "config"],
            },
          ]}
          onEntrySelect={(entry) => console.log("Selected entry:", entry)}
          onSearch={(query) => console.log("Search:", query)}
          onEntryDelete={(id) => console.log("Delete entry:", id)}
        />

        <h4 className="text-xs font-medium mb-2">Long Term Context</h4>
        <LongTermContextExplorer
          chunks={[
            {
              id: "c1",
              content: "Project Alpha started in Q1 2024. The main goal is to build a scalable function runtime.",
              timestamp: Date.now() - 604800000,
              importance: 0.95,
              decayScore: 0.1,
              retentionPriority: "critical",
              retrievalCount: 45,
            },
            {
              id: "c2",
              content: "Team decided to use WebAssembly for sandboxing function execution.",
              timestamp: Date.now() - 518400000,
              importance: 0.85,
              decayScore: 0.2,
              retentionPriority: "high",
              retrievalCount: 30,
            },
            {
              id: "c3",
              content: "Performance benchmarks show 2ms average cold start time.",
              timestamp: Date.now() - 259200000,
              importance: 0.7,
              decayScore: 0.4,
              retentionPriority: "medium",
              retrievalCount: 12,
            },
          ]}
          onChunkSelect={(chunk) => console.log("Selected chunk:", chunk)}
          onMemoryReinforce={(id) => console.log("Reinforce:", id)}
        />
      </div>
    </div>
  );
}
