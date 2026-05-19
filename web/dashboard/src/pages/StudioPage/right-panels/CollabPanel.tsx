import React, { useMemo } from "react";
import {
  SessionReplayViewer,
  SharedMemoryBoard,
  TeamActivityFeed,
  CollaborativeGraphEditor,
  ConflictResolutionPanel,
  CollaborativePromptEditor,
  AIHumanTaskBoard,
  LivePairProgrammingView,
  AsyncReviewTimeline,
  RealtimeAnnotationSystem,
  SharedExecutionView,
} from "@functionfly/ui-collaboration";
import type { Collaborator } from "@functionfly/ui-collaboration";
import type { CollabEvent } from "@/api/studioCollab";
import { Users, MessageSquare, GitBranch, AlertTriangle } from "lucide-react";

interface PromptVersion {
  id: string;
  metadata?: {
    prompt?: string;
    user_name?: string;
    user_color?: string;
    changes?: string;
  };
  created_at: string;
}

interface PairSession {
  id: string;
  metadata?: {
    host_name?: string;
    host_color?: string;
    guest_name?: string;
    guest_color?: string;
    status?: string;
    current_file?: string;
    current_line?: number;
  };
  created_at: string;
}

interface Comment {
  id: string;
  metadata?: {
    user_name?: string;
    user_color?: string;
    content?: string;
    line?: number;
    resolved?: boolean;
  };
  created_at: string;
}

interface Annotation {
  id: string;
  metadata?: {
    user_name?: string;
    user_color?: string;
    target_id?: string;
    target_type?: string;
    content?: string;
    position?: { x: number; y: number };
    resolved?: boolean;
  };
  created_at: string;
}

interface GraphEdit {
  id: string;
  created_by: string;
  metadata?: {
    user_name?: string;
    node_id?: string;
    field?: string;
    old_value?: string;
    new_value?: string;
  };
  created_at: string;
}

interface Conflict {
  id: string;
  metadata?: {
    field?: string;
    current_user?: string;
    current_value?: string;
    incoming_user?: string;
    incoming_value?: string;
  };
}

interface TeamMemory {
  id: string;
  summary?: string;
  memory_type?: string;
  created_by?: string;
  created_at: string;
}

interface Execution {
  id?: string;
  status: string;
  graphId?: string;
  startedAt?: string;
  nodeResults?: Array<{
    nodeId: string;
    status: string;
    error?: string;
  }>;
}

interface CollabPanelProps {
  collaborators: Collaborator[];
  currentUser: { name: string; color: string };
  collabActivityData: CollabEvent[];
  promptVersionsData: PromptVersion[];
  pairSessionsData: PairSession[];
  commentsData: Comment[];
  annotationsData: Annotation[];
  graphEditsData: GraphEdit[];
  conflictsData: Conflict[];
  teamMemories: TeamMemory[];
  executions: Execution[];
  onResolveComment?: (commentId: string, resolved: boolean) => void;
  onResolveAnnotation?: (annotationId: string, resolved: boolean) => void;
  onEndPairSession?: (sessionId: string) => void;
  onUpdatePromptVersion?: (prompt: string, changes: string) => void;
  onRecordActivity?: (activity: { action: string; target: string; icon: string }) => void;
}

export function CollabPanel({
  collaborators,
  currentUser,
  collabActivityData,
  promptVersionsData,
  pairSessionsData,
  commentsData,
  annotationsData,
  graphEditsData,
  conflictsData,
  teamMemories,
  executions,
  onResolveComment,
  onResolveAnnotation,
  onEndPairSession,
  onUpdatePromptVersion,
  onRecordActivity,
}: CollabPanelProps) {
  const hasConflicts = conflictsData && conflictsData.length > 0;

  return (
    <div className="p-3 space-y-4">
      {hasConflicts && (
        <div className="bg-error/10 border border-error/20 rounded-lg p-3">
          <div className="flex items-center gap-2 mb-2">
            <AlertTriangle className="size-4 text-error" />
            <span className="text-xs font-medium text-error">Conflicts Detected</span>
          </div>
          <ConflictResolutionPanel
            conflicts={conflictsData!.map((ev) => ({
              id: ev.id,
              field: ev.metadata?.field || "unknown",
              current: {
                user: ev.metadata?.current_user || "You",
                value: ev.metadata?.current_value || "",
              },
              incoming: {
                user: ev.metadata?.incoming_user || "Collaborator",
                value: ev.metadata?.incoming_value || "",
              },
            }))}
          />
        </div>
      )}

      <div>
        <h4 className="text-xs font-medium mb-2 flex items-center gap-2">
          <Users className="size-3 text-brand-400" />
          Active Collaborators ({collaborators.length})
        </h4>
        <div className="flex flex-wrap gap-1">
          {collaborators.map((c) => (
            <div
              key={c.id}
              className="px-2 py-1 bg-bg-primary rounded-full border border-border-subtle text-[10px]"
            >
              <span
                className="inline-block w-2 h-2 rounded-full mr-1"
                style={{ backgroundColor: c.color }}
              />
              {c.name}
            </div>
          ))}
        </div>
      </div>

      <SessionReplayViewer
        recording={{
          id: executions[0]?.id || "session-1",
          name: "Execution Replay",
          events: executions.slice(0, 10).map((ex) => ({
            id: ex.id,
            type: ex.status,
            userId: ex.graphId || "system",
            timestamp: new Date(ex.startedAt || Date.now()).getTime(),
            data: { graphId: ex.graphId, status: ex.status, nodeResults: ex.nodeResults },
          })),
          duration: 60000,
          createdAt: Date.now(),
        }}
        className="bg-bg-primary rounded-lg border border-border-subtle p-3"
      />

      <SharedMemoryBoard
        cards={teamMemories.map((m) => ({
          id: m.id,
          title: m.summary || m.memory_type || "Memory",
          content: m.summary || "",
          author: m.created_by || "Team",
          createdAt: new Date(m.created_at).getTime(),
          tags: [],
        }))}
        className="bg-bg-primary rounded-lg border border-border-subtle p-3"
      />

      <TeamActivityFeed
        activities={useMemo(() => {
          const collabActivities = (collabActivityData || []).map((ev) => ({
            id: ev.id,
            user: {
              name: ev.metadata?.user_name || ev.created_by || "Unknown",
              color: ev.metadata?.user_color || "#6b7280",
              isAI: !!ev.metadata?.is_ai,
            },
            action: ev.metadata?.action || ev.event_type || "performed action",
            target: ev.metadata?.target,
            timestamp: ev.created_at,
            icon: ev.metadata?.icon || "🔹",
          }));
          return collabActivities.sort(
            (a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
          );
        }, [collabActivityData])}
        className="bg-bg-primary rounded-lg border border-border-subtle p-3"
      />

      <CollaborativeGraphEditor
        edits={(graphEditsData || []).map((ev) => ({
          id: ev.id,
          userId: ev.created_by,
          userName: ev.metadata?.user_name || "Unknown",
          nodeId: ev.metadata?.node_id || "",
          field: ev.metadata?.field || "property",
          oldValue: ev.metadata?.old_value || "",
          newValue: ev.metadata?.new_value || "",
          timestamp: ev.created_at,
        }))}
        className="bg-bg-primary rounded-lg border border-border-subtle p-3"
      />

      <CollaborativePromptEditor
        versions={(promptVersionsData || []).map((ev, idx) => ({
          id: ev.id,
          author: {
            name: ev.metadata?.user_name || "Unknown",
            color: ev.metadata?.user_color || "#6b7280",
          },
          version: idx + 1,
          prompt: ev.metadata?.prompt || "",
          timestamp: ev.created_at,
          changes: ev.metadata?.changes || "Update",
        }))}
        currentPrompt={(promptVersionsData || [])[0]?.metadata?.prompt || ""}
        onPromptChange={(p) => onUpdatePromptVersion?.(p, "Edit")}
      />

      <AIHumanTaskBoard
        tasks={executions.slice(0, 5).map((ex, i) => ({
          id: ex.id || `task-${i}`,
          title: ex.graphId || `Execution ${i + 1}`,
          assignedTo: {
            name: "Agent",
            isAI: true,
          },
          status:
            ex.status === "completed"
              ? "done"
              : ex.status === "failed"
                ? "blocked"
                : "in-progress",
          priority: "medium" as const,
        }))}
      />

      <LivePairProgrammingView
        session={{
          id: (pairSessionsData || [])[0]?.id || "none",
          host: {
            name:
              (pairSessionsData || [])[0]?.metadata?.host_name ||
              currentUser.name,
            color:
              (pairSessionsData || [])[0]?.metadata?.host_color ||
              currentUser.color,
          },
          guest: (pairSessionsData || [])[0]?.metadata?.guest_name
            ? {
                name: (pairSessionsData || [])[0]!.metadata!.guest_name!,
                color:
                  (pairSessionsData || [])[0]!.metadata?.guest_color || "#10b981",
              }
            : undefined,
          status:
            ((pairSessionsData || [])[0]?.metadata?.status) as
              | "active"
              | "ended"
              | "paused" || "ended",
          startedAt:
            (pairSessionsData || [])[0]?.created_at ||
            new Date().toISOString(),
          currentFile:
            (pairSessionsData || [])[0]?.metadata?.current_file,
          currentLine: (pairSessionsData || [])[0]?.metadata?.current_line,
        }}
        onEndSession={() => {
          const ev = (pairSessionsData || [])[0];
          if (ev?.id) onEndPairSession?.(ev.id);
        }}
      />

      <AsyncReviewTimeline
        comments={(commentsData || []).map((ev) => ({
          id: ev.id,
          author: {
            name: ev.metadata?.user_name || "Unknown",
            color: ev.metadata?.user_color || "#6b7280",
          },
          content: ev.metadata?.content || "",
          timestamp: ev.created_at,
          line: ev.metadata?.line,
          resolved: !!ev.metadata?.resolved,
        }))}
        onResolve={(id) => onResolveComment?.(id, true)}
      />

      <RealtimeAnnotationSystem
        annotations={(annotationsData || []).map((ev) => ({
          id: ev.id,
          author: {
            name: ev.metadata?.user_name || "Unknown",
            color: ev.metadata?.user_color || "#6b7280",
          },
          targetId: ev.metadata?.target_id || "",
          targetType: ev.metadata?.target_type || "canvas",
          content: ev.metadata?.content || "",
          position: ev.metadata?.position || { x: 0, y: 0 },
          timestamp: ev.created_at,
          resolved: !!ev.metadata?.resolved,
        }))}
        onResolve={(id) => onResolveAnnotation?.(id, true)}
      />

      <SharedExecutionView
        executionId={executions[0]?.id || "none"}
        participants={collaborators.map((c) => ({
          id: c.id,
          name: c.name,
          color: c.color,
          currentStep: 0,
        }))}
        steps={(executions[0]?.nodeResults || []).map((nr, idx) => ({
          step: idx,
          action: nr.nodeId,
          agent: "Agent",
          timestamp: executions[0]?.startedAt || new Date().toISOString(),
          result: nr.status === "success" ? "Completed" : nr.error,
        }))}
        onStepClick={(step) =>
          onRecordActivity?.({ action: "clicked step", target: `step-${step}`, icon: "👆" })
        }
      />
    </div>
  );
}