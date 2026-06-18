/**
 * @functionfly/ui-collaboration
 * Multiplayer & Collaboration Components - Types
 */

import React from 'react';

// ============================================================================
// Shared Types
// ============================================================================

export interface CursorState {
  line: number;
  column: number;
  filePath: string;
}

export interface CollaboratorPresence {
  id: string;
  userId: string;
  userName: string;
  color: string;
  status: 'active' | 'idle' | 'away';
  lastActivity: number;
  cursor?: CursorState;
}

export interface VoiceParticipant {
  id: string;
  userId: string;
  userName: string;
  isMuted: boolean;
  isSpeaking: boolean;
  joinedAt: number;
}

export interface VoiceSession {
  id: string;
  participants: VoiceParticipant[];
  startedAt: number;
  isActive: boolean;
}

export interface ExecutionBookmark {
  id: string;
  executionId: string;
  label: string;
  timestamp: number;
  metadata?: Record<string, unknown>;
}

export interface GraphNode {
  id: string;
  type: string;
  label: string;
  x: number;
  y: number;
  metadata?: Record<string, unknown>;
}

export interface GraphEdge {
  id: string;
  source: string;
  target: string;
  label?: string;
  metadata?: Record<string, unknown>;
}

export interface GraphOperation {
  type: 'add' | 'update' | 'delete';
  nodeId?: string;
  edgeId?: string;
  data?: GraphNode | GraphEdge;
}

export interface Annotation {
  id: string;
  type: 'comment' | 'highlight' | 'note' | 'suggestion' | 'todo' | 'issue';
  content: string;
  author: string;
  timestamp?: number;
  createdAt?: number;
  updatedAt?: number;
  resolvedAt?: number;
  resolvedBy?: string;
  resolved?: boolean;
  position: { line: number; column: number } | { filePath: string; startLine: number; endLine: number; startColumn?: number; endColumn?: number };
  replies?: Annotation[];
  reactions?: Array<{ emoji: string; userId: string }>;
  mentions?: string[];
}

export interface SessionEvent {
  id: string;
  type: string;
  userId: string;
  timestamp: number;
  data: Record<string, unknown>;
}

export interface SessionRecording {
  id: string;
  name: string;
  events: SessionEvent[];
  duration: number;
  createdAt: number;
}

export interface ActivityItem {
  id: string;
  type: string;
  userId: string;
  userName: string;
  timestamp: number;
  description: string;
  metadata?: Record<string, unknown>;
}

export interface MemoryCard {
  id: string;
  title: string;
  content: string;
  author: string;
  createdAt: number;
  tags: string[];
}

export interface ConflictMarker {
  id: string;
  type: 'edit' | 'delete' | 'structural';
  filePath: string;
  position: { start: number; end: number };
  original?: string;
  incoming?: string;
}

export interface ConflictResolution {
  conflictId: string;
  resolution: 'accept-ours' | 'accept-theirs' | 'merge';
  merged?: string;
}

export interface ReviewComment {
  id: string;
  author: string;
  content: string;
  timestamp?: number;
  createdAt?: number;
  lineNumber?: number;
  filePath?: string;
  resolved?: boolean;
}

export interface ReviewSession {
  id: string;
  name?: string;
  description?: string;
  status: 'open' | 'in-progress' | 'completed' | 'in-review' | 'approved' | 'changes-requested' | 'merged';
  author?: { id: string; name: string; avatar?: string };
  createdAt?: number;
  updatedAt?: number;
  participants?: Array<{ userId: string; userName: string; role: 'reviewer' | 'author' }>;
  comments: ReviewComment[];
  timeline?: Array<{ id: string; event: string; userId: string; userName: string; timestamp: number; details?: string }>;
}

export interface PromptSegment {
  id: string;
  type: 'text' | 'code' | 'result' | 'variable' | 'function' | 'loop' | 'conditional' | 'ai-template';
  content: string;
  timestamp?: number;
  createdAt?: number;
  author?: string;
  editedBy?: string;
  lastEditAt?: number;
  metadata?: { variableName?: string; functionName?: string; outputType?: string };
}

export interface DriverNavigator {
  driverId: string;
  driverName: string;
  navigatorId: string;
  navigatorName: string;
  startedAt: number;
}

export interface PairProgrammingSession {
  id: string;
  driver: DriverNavigator;
  isActive: boolean;
  startedAt: number;
}

export interface CodePosition {
  line: number;
  column: number;
}

export interface CodeRange {
  start: CodePosition;
  end: CodePosition;
}

export interface TaskAssignee {
  id: string;
  name: string;
  avatar?: string;
  type: 'human' | 'ai';
}

export interface TaskAssignment {
  id: string;
  title: string;
  status: 'todo' | 'in-progress' | 'review' | 'done' | 'blocked';
  priority: 'urgent' | 'high' | 'medium' | 'low';
  assignees: TaskAssignee[];
  createdAt: number;
  aiSuggestion?: {
    assignee: TaskAssignee;
    confidence: number;
    reason: string;
  };
}

// ============================================================================
// Component Props
// ============================================================================

export interface LivePresenceLayerProps {
  presences: CollaboratorPresence[];
  currentUserId: string;
  children?: React.ReactNode;
}

export interface CollaboratorCursorProps {
  presence: CollaboratorPresence;
  children?: React.ReactNode;
}

export interface VoiceSessionPanelProps {
  session: VoiceSession | null;
  onParticipantToggleMute?: (participantId: string) => void;
  onParticipantToggleDeafen?: (participantId: string) => void;
  onLeave?: () => void;
}

export interface SharedExecutionViewProps {
  executionId: string | null;
  bookmarks: ExecutionBookmark[];
  onBookmarkCreate?: (label: string) => void;
  onBookmarkJump?: (bookmarkId: string) => void;
}

export interface CollaborativeGraphEditorProps {
  nodes: GraphNode[];
  edges: GraphEdge[];
  onNodesChange?: (nodes: GraphNode[]) => void;
  onEdgesChange?: (edges: GraphEdge[]) => void;
  onOperation?: (op: GraphOperation) => void;
  editable?: boolean;
}

export interface RealtimeAnnotationSystemProps {
  annotations: Annotation[];
  onAddAnnotation?: (annotation: Omit<Annotation, 'id'>) => void;
  onResolveAnnotation?: (annotationId: string) => void;
}

export interface SessionReplayViewerProps {
  recording: SessionRecording;
  onEventClick?: (event: SessionEvent) => void;
}

export interface TeamActivityFeedProps {
  activities: ActivityItem[];
  onActivityClick?: (activity: ActivityItem) => void;
}

export interface SharedMemoryBoardProps {
  cards: MemoryCard[];
  onCardCreate?: (card: Omit<MemoryCard, 'id' | 'createdAt'>) => void;
  onCardUpdate?: (cardId: string, updates: Partial<MemoryCard>) => void;
  onCardDelete?: (cardId: string) => void;
}

export interface ConflictResolutionPanelProps {
  conflicts: ConflictMarker[];
  onResolve?: (resolution: ConflictResolution) => void;
}

export interface AsyncReviewTimelineProps {
  session: ReviewSession;
  onCommentAdd?: (comment: Omit<ReviewComment, 'id'>) => void;
  onCommentResolve?: (commentId: string) => void;
}

export interface CollaborativePromptEditorProps {
  segments: PromptSegment[];
  onSegmentAdd?: (segment: Omit<PromptSegment, 'id' | 'timestamp'>) => void;
  onSegmentUpdate?: (segmentId: string, content: string) => void;
}

export interface LivePairProgrammingViewProps {
  session: PairProgrammingSession;
  onSessionEnd?: () => void;
}

export interface AIHumanTaskAssignmentBoardProps {
  tasks: TaskAssignment[];
  selectedTaskId?: string | null;
  onTaskSelect?: (task: TaskAssignment) => void;
  onTaskUpdate?: (taskId: string, updates: Partial<TaskAssignment>) => void;
  onTaskAssign?: (taskId: string, assignee: TaskAssignee) => void;
  onAISuggestionAccept?: (taskId: string) => void;
  onAISuggestionReject?: (taskId: string) => void;
  onTaskCreate?: (task: Omit<TaskAssignment, 'id' | 'createdAt'>) => void;
  className?: string;
}
