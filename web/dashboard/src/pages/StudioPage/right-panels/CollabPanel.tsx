import type { CollabEvent } from "@/api/studioCollab";
import {
  AIHumanTaskAssignmentBoard,
  AsyncReviewTimeline,
  CollaborativeGraphEditor,
  CollaborativePromptEditor,
  ConflictResolutionPanel,
  LivePairProgrammingView,
  RealtimeAnnotationSystem,
  SessionReplayViewer,
  SharedExecutionView,
  SharedMemoryBoard,
  TeamActivityFeed,
} from "@functionfly/ui-collaboration";
import { AlertTriangle, Users } from "lucide-react";
import { useMemo } from "react";
import { CollabPanelSkeleton } from "../components/StudioPanelsSkeleton";
import {
  adaptMemoryCard,
  adaptCollabEventList,
  adaptPromptVersionList,
  adaptPairSession,
  adaptCommentListToReviewSession,
  adaptGraphEdit,
  adaptExecutionToRecording,
  adaptExecutionToBookmarks,
  adaptExecutionListToTasks,
} from "@/adapters";
import type {
  MemoryCard as UIMemoryCard,
  ActivityItem,
  PromptSegment,
  PairProgrammingSession,
  ReviewSession,
  GraphNode,
  GraphEdge,
  Annotation,
  ExecutionBookmark,
  ConflictMarker,
  TaskAssignment as UITaskAssignment,
} from "@functionfly/ui-collaboration/types";

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

interface AnnotationData {
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
  collaborators: Array<{ id: string; name: string; color: string }>;
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
  isLoading?: boolean;
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
  isLoading,
}: CollabPanelProps) {
  const hasConflicts = conflictsData && conflictsData.length > 0;

  const adaptedRecording = useMemo(() => adaptExecutionToRecording(executions), [executions]);

  const adaptedMemoryCards = useMemo<UIMemoryCard[]>(() =>
    teamMemories.map(m => ({
      id: m.id,
      title: m.summary || m.memory_type || "Memory",
      content: m.summary || "",
      author: m.created_by || "Team",
      createdAt: new Date(m.created_at).getTime(),
      tags: [],
    })),
    [teamMemories]
  );

  const adaptedActivities = useMemo<ActivityItem[]>(() =>
    adaptCollabEventList(collabActivityData || []),
    [collabActivityData]
  );

  const adaptedPromptSegments = useMemo<PromptSegment[]>(() =>
    adaptPromptVersionList(promptVersionsData || []),
    [promptVersionsData]
  );

  const adaptedPairSession = useMemo<PairProgrammingSession | null>(() =>
    adaptPairSession((pairSessionsData || [])[0] || null),
    [pairSessionsData]
  );

  const adaptedReviewSession = useMemo<ReviewSession>(() =>
    adaptCommentListToReviewSession(commentsData || []),
    [commentsData]
  );

  const adaptedAnnotations: Annotation[] = annotationsData || [];

  const adaptedBookmarks = useMemo<ExecutionBookmark[]>(() =>
    adaptExecutionToBookmarks(executions) as unknown as ExecutionBookmark[],
    [executions]
  );

  const adaptedTasks = useMemo<UITaskAssignment[]>(() =>
    adaptExecutionListToTasks(executions),
    [executions]
  );

  const conflictMarkers = useMemo<ConflictMarker[]>(() =>
    (conflictsData || []).map(ev => ({
      id: ev.id,
      type: 'edit' as const,
      filePath: 'unknown',
      position: { start: 0, end: 0 },
      original: ev.metadata?.current_value,
      incoming: ev.metadata?.incoming_value,
    })),
    [conflictsData]
  );

  if (isLoading) {
    return <CollabPanelSkeleton />;
  }

  return (
    <div className="p-3 space-y-4">
      {hasConflicts && (
        <div className="bg-error/10 border border-error/20 rounded-lg p-3">
          <div className="flex items-center gap-2 mb-2">
            <AlertTriangle className="size-4 text-error" />
            <span className="text-xs font-medium text-error">Conflicts Detected</span>
          </div>
          <ConflictResolutionPanel
            conflicts={conflictMarkers}
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
        recording={adaptedRecording}
        onEventClick={undefined}
      />

      <SharedMemoryBoard
        cards={adaptedMemoryCards}
        onCardCreate={undefined}
        onCardUpdate={undefined}
        onCardDelete={undefined}
      />

      <TeamActivityFeed
        activities={adaptedActivities}
        onActivityClick={undefined}
      />

      <CollaborativeGraphEditor
        nodes={[]}
        edges={[]}
        onNodesChange={undefined}
        onEdgesChange={undefined}
        onOperation={undefined}
        editable={false}
      />

      <CollaborativePromptEditor
        segments={adaptedPromptSegments}
        onSegmentUpdate={(id, content) => onUpdatePromptVersion?.(content, "Edit")}
      />

      <AIHumanTaskAssignmentBoard
        tasks={adaptedTasks}
      />

      <LivePairProgrammingView
        session={adaptedPairSession}
        onSessionEnd={() => {
          const ev = (pairSessionsData || [])[0];
          if (ev?.id) onEndPairSession?.(ev.id);
        }}
      />

      <AsyncReviewTimeline
        session={adaptedReviewSession}
        onCommentResolve={(id) => onResolveComment?.(id, true)}
      />

      <RealtimeAnnotationSystem
        annotations={adaptedAnnotations}
        onResolveAnnotation={(id) => onResolveAnnotation?.(id, true)}
      />

      <SharedExecutionView
        executionId={executions[0]?.id || null}
        bookmarks={adaptedBookmarks}
        onBookmarkJump={(id) => onRecordActivity?.({ action: "clicked bookmark", target: id, icon: "👆" })}
      />
    </div>
  );
}
