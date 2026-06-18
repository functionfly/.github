/**
 * UI Package Adapters
 * Transform dashboard store types to match UI package component expectations
 */

import type { CollaboratorPresence, VoiceSession, ExecutionBookmark, GraphNode, GraphEdge, Annotation, ActivityItem, MemoryCard, ReviewSession, ReviewComment, PromptSegment, PairProgrammingSession, TaskAssignment, TaskAssignee } from '../stores/collaborationStore';

/**
 * Target types for UI package compatibility (derived from ui-collaboration types)
 * Note: These types are defined locally because the UI package doesn't export them as standalone types
 */

export interface UICursorState {
  line: number;
  column: number;
  filePath: string;
}

export interface UICollaboratorPresence {
  id: string;
  userId: string;
  userName: string;
  color: string;
  status: 'active' | 'idle' | 'away';
  lastActivity: number;
  cursor?: UICursorState;
}

export interface UIVoiceSession {
  id: string;
  name: string;
  participants: Array<{
    id: string;
    userId: string;
    userName: string;
    userAvatar?: string;
    isMuted: boolean;
    isDeafened: boolean;
    isSpeaking: boolean;
    isHandRaised: boolean;
    joinedAt: number;
  }>;
  isActive: boolean;
  startedAt?: number;
  maxParticipants?: number;
}

export interface UIExecutionBookmark {
  id: string;
  userId: string;
  userName: string;
  timestamp: number;
  executionId: string;
  stepIndex: number;
  pausedAt?: number;
}

export interface UIGraphNode {
  id: string;
  type: string;
  label: string;
  position: { x: number; y: number };
  color?: string;
  metadata?: Record<string, unknown>;
  lockedBy?: string;
}

export interface UIGraphEdge {
  id: string;
  source: string;
  target: string;
  label?: string;
  type?: string;
  weight?: number;
}

export interface UIGraphOperation {
  type: 'add' | 'update' | 'delete';
  nodeId?: string;
  edgeId?: string;
  data?: UIGraphNode | UIGraphEdge;
}

export interface UIAnnotation {
  id: string;
  type: 'comment' | 'highlight' | 'note';
  content: string;
  author: string;
  timestamp: number;
  position: { line: number; column: number };
  resolved?: boolean;
}

export interface UIActivityItem {
  id: string;
  type: string;
  userId: string;
  userName: string;
  timestamp: number;
  description: string;
  metadata?: Record<string, unknown>;
}

export interface UIMemoryCard {
  id: string;
  title: string;
  content: string;
  author: string;
  createdAt: number;
  tags: string[];
}

export interface UIConflictMarker {
  id: string;
  type: 'edit' | 'delete' | 'structural';
  filePath: string;
  position: { start: number; end: number };
  original?: string;
  incoming?: string;
}

export interface UIConflictResolution {
  conflictId: string;
  resolution: 'accept-ours' | 'accept-theirs' | 'merge';
  merged?: string;
}

export interface UIReviewComment {
  id: string;
  author: string;
  content: string;
  timestamp: number;
  line?: number;
  resolved?: boolean;
}

export interface UIReviewSession {
  id: string;
  status: 'open' | 'in-progress' | 'completed';
  comments: UIReviewComment[];
}

export interface UIPromptSegment {
  id: string;
  type: 'text' | 'code' | 'result';
  content: string;
  timestamp: number;
  author?: string;
}

export interface UIPairProgrammingSession {
  id: string;
  driver: {
    driverId: string;
    driverName: string;
    navigatorId: string;
    navigatorName: string;
    startedAt: number;
  };
  isActive: boolean;
  startedAt: number;
}

export interface UIDriverNavigator {
  driverId: string;
  driverName: string;
  navigatorId: string;
  navigatorName: string;
  startedAt: number;
}

export interface UITaskAssignee {
  id: string;
  name: string;
  avatar?: string;
  type: 'human' | 'ai';
}

export interface UITaskAssignment {
  id: string;
  title: string;
  status: 'todo' | 'in-progress' | 'review' | 'done' | 'blocked';
  priority: 'urgent' | 'high' | 'medium' | 'low';
  assignees: UITaskAssignee[];
  createdAt: number;
  aiSuggestion?: {
    assignee: UITaskAssignee;
    confidence: number;
    reason: string;
  };
}

export interface UISessionRecording {
  id: string;
  name: string;
  events: UISessionEvent[];
  duration: number;
  createdAt: number;
}

export interface UISessionEvent {
  id: string;
  type: string;
  userId: string;
  timestamp: number;
  data: Record<string, unknown>;
}

/**
 * Adapt CollaboratorPresence from store to UI component format
 */
export function adaptCollaboratorPresence(
  presence: CollaboratorPresence
): UICollaboratorPresence {
  return {
    id: presence.id,
    userId: presence.userId,
    userName: presence.userName,
    color: presence.color,
    status: presence.status,
    lastActivity: presence.lastActivity,
    cursor: presence.cursor
      ? { line: presence.cursor.line, column: presence.cursor.column, filePath: presence.cursor.filePath || '' }
      : undefined,
  };
}

/**
 * Adapt CollaboratorPresence array
 */
export function adaptCollaboratorPresenceList(
  presences: CollaboratorPresence[]
): UICollaboratorPresence[] {
  return presences.map(adaptCollaboratorPresence);
}

/**
 * Adapt VoiceSession from store to UI component format
 */
export function adaptVoiceSession(session: VoiceSession | null): UIVoiceSession | null {
  if (!session) return null;
  return {
    id: session.id,
    name: session.name || 'Voice Session',
    participants: session.participants.map(p => ({
      id: p.id,
      userId: p.userId,
      userName: p.userName,
      userAvatar: p.userAvatar,
      isMuted: p.isMuted,
      isDeafened: p.isDeafened,
      isSpeaking: p.isSpeaking,
      isHandRaised: p.isHandRaised,
      joinedAt: p.joinedAt,
    })),
    isActive: session.isActive,
    startedAt: session.startedAt,
    maxParticipants: session.maxParticipants,
  };
}

/**
 * Adapt ExecutionBookmark from store to UI component format
 */
export function adaptExecutionBookmark(bookmark: ExecutionBookmark): UIExecutionBookmark {
  return {
    id: bookmark.id,
    userId: bookmark.userId,
    userName: bookmark.userName,
    timestamp: bookmark.timestamp,
    executionId: bookmark.executionId,
    stepIndex: bookmark.stepIndex,
    pausedAt: bookmark.pausedAt,
  };
}

/**
 * Adapt GraphNode from store to UI component format
 */
export function adaptGraphNode(node: GraphNode): UIGraphNode {
  return {
    id: node.id,
    type: node.type,
    label: node.label,
    position: node.position || { x: 0, y: 0 },
    color: node.color,
    metadata: node.metadata,
    lockedBy: node.lockedBy,
  };
}

/**
 * Adapt GraphEdge from store to UI component format
 */
export function adaptGraphEdge(edge: GraphEdge): UIGraphEdge {
  return {
    id: edge.id,
    source: edge.source,
    target: edge.target,
    label: edge.label,
    type: edge.type,
    weight: edge.weight,
  };
}

/**
 * Adapt Annotation from store to UI component format
 */
export function adaptAnnotation(annotation: Annotation): UIAnnotation {
  const authorName = typeof annotation.author === 'string'
    ? annotation.author
    : annotation.author.name;
  return {
    id: annotation.id,
    type: 'comment' as const,
    content: annotation.content,
    author: authorName,
    timestamp: annotation.createdAt,
    position: { line: 1, column: 1 },
    resolved: annotation.resolved,
  };
}

/**
 * Adapt ActivityItem from store to UI component format
 */
export function adaptActivityItem(item: ActivityItem): UIActivityItem {
  return {
    id: item.id,
    type: item.type,
    userId: item.userId,
    userName: item.userName,
    timestamp: item.timestamp || Date.now(),
    description: item.description,
    metadata: item.metadata,
  };
}

/**
 * Adapt MemoryCard from store to UI component format
 */
export function adaptMemoryCard(card: MemoryCard): UIMemoryCard {
  const authorName = typeof card.author === 'string' ? card.author : card.author.name;
  return {
    id: card.id,
    title: (card as any).title || card.content.substring(0, 30),
    content: card.content,
    author: authorName,
    createdAt: card.createdAt,
    tags: card.tags || [],
  };
}

/**
 * Adapt ReviewComment from store to UI component format
 */
export function adaptReviewComment(comment: ReviewComment): UIReviewComment {
  const authorName = typeof comment.author === 'string' ? comment.author : comment.author.name;
  return {
    id: comment.id,
    author: authorName,
    timestamp: comment.createdAt,
    content: comment.content,
  };
}

/**
 * Adapt ReviewSession from store to UI component format
 */
export function adaptReviewSession(review: ReviewSession | null): UIReviewSession | null {
  if (!review) return null;
  const authorName = typeof review.author === 'string' ? review.author : review.author.name;
  return {
    id: review.id,
    status: review.status as 'open' | 'in-progress' | 'completed',
    comments: review.comments.map(adaptReviewComment),
  };
}

/**
 * Adapt PromptSegment from store to UI component format
 */
export function adaptPromptSegment(segment: PromptSegment): UIPromptSegment {
  return {
    id: segment.id,
    type: 'text',
    content: segment.content,
    timestamp: segment.lastEditAt ?? Date.now(),
    author: segment.editedBy,
  };
}

/**
 * Adapt PairProgrammingSession from store to UI component format
 */
export function adaptPairProgrammingSession(
  session: PairProgrammingSession | null
): UIPairProgrammingSession | null {
  if (!session) return null;
  return {
    id: session.id,
    driver: {
      driverId: session.currentDriverId || '',
      driverName: session.drivers?.find(d => d.role === 'driver')?.userId || 'Driver',
      navigatorId: session.currentNavigatorId || '',
      navigatorName: session.drivers?.find(d => d.role === 'navigator')?.userId || 'Navigator',
      startedAt: session.startedAt,
    },
    isActive: session.isActive,
    startedAt: session.startedAt,
  };
}

/**
 * Adapt TaskAssignee from store to UI component format
 */
export function adaptTaskAssignee(assignee: TaskAssignee): UITaskAssignee {
  return {
    id: assignee.userId,
    name: assignee.userName,
    avatar: assignee.avatar,
    type: assignee.type,
  };
}

/**
 * Adapt TaskAssignment from store to UI component format
 */
export function adaptTaskAssignment(task: TaskAssignment): UITaskAssignment {
  return {
    id: task.id,
    title: task.title,
    status: task.status,
    priority: task.priority,
    assignees: task.assignees.map(adaptTaskAssignee),
    createdAt: task.createdAt,
    aiSuggestion: task.suggestedAssignee
      ? {
          assignee: {
            id: task.suggestedAssignee.userId,
            name: task.suggestedAssignee.userName,
            avatar: undefined,
            type: 'ai' as const,
          },
          confidence: task.suggestedAssignee.confidence,
          reason: task.suggestedAssignee.reasoning || '',
        }
      : undefined,
  };
}

/**
 * Adapt TaskAssignment array
 */
export function adaptTaskAssignmentList(tasks: TaskAssignment[]): UITaskAssignment[] {
  return tasks.map(adaptTaskAssignment);
}

// ============================================================================
// Additional adapters for CollabPanel types
// ============================================================================

export interface PromptVersion {
  id: string;
  metadata?: {
    prompt?: string;
    user_name?: string;
    user_color?: string;
    changes?: string;
  };
  created_at: string;
}

export interface PairSession {
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

export interface Comment {
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

export interface GraphEdit {
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

export interface Execution {
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

export interface CollabEvent {
  id: string;
  event_type?: string;
  created_by?: string;
  metadata?: {
    user_name?: string;
    user_color?: string;
    is_ai?: boolean;
    action?: string;
    target?: string;
    icon?: string;
  };
  created_at: string;
}

/**
 * Adapt PromptVersion to UIPromptSegment for CollaborativePromptEditor
 */
export function adaptPromptVersion(version: PromptVersion): UIPromptSegment {
  return {
    id: version.id,
    type: 'text',
    content: version.metadata?.prompt || '',
    timestamp: new Date(version.created_at).getTime(),
    author: version.metadata?.user_name,
  };
}

/**
 * Adapt PromptVersion array for CollaborativePromptEditor
 */
export function adaptPromptVersionList(versions: PromptVersion[]): UIPromptSegment[] {
  return versions.map(adaptPromptVersion);
}

/**
 * Adapt PairSession to PairProgrammingSession for LivePairProgrammingView
 */
export function adaptPairSession(session: PairSession | null): UIPairProgrammingSession | null {
  if (!session) return null;
  const hostName = session.metadata?.host_name || 'Host';
  const guestName = session.metadata?.guest_name;
  return {
    id: session.id,
    driver: {
      driverId: `host-${session.id}`,
      driverName: hostName,
      navigatorId: guestName ? `guest-${session.id}` : '',
      navigatorName: guestName || '',
      startedAt: new Date(session.created_at).getTime(),
    },
    isActive: session.metadata?.status === 'active',
    startedAt: new Date(session.created_at).getTime(),
  };
}

/**
 * Adapt Comment to ReviewComment for AsyncReviewTimeline
 */
export function adaptComment(comment: Comment): UIReviewComment {
  return {
    id: comment.id,
    author: comment.metadata?.user_name || 'Unknown',
    content: comment.metadata?.content || '',
    timestamp: new Date(comment.created_at).getTime(),
    line: comment.metadata?.line,
    resolved: comment.metadata?.resolved,
  };
}

/**
 * Adapt Comments array to ReviewSession for AsyncReviewTimeline
 */
export function adaptCommentListToReviewSession(comments: Comment[]): UIReviewSession {
  return {
    id: 'review-session',
    status: 'open',
    comments: comments.map(adaptComment),
  };
}

/**
 * Adapt GraphEdit to GraphOperation for CollaborativeGraphEditor
 */
export function adaptGraphEdit(edit: GraphEdit): UIGraphOperation {
  return {
    type: 'update',
    nodeId: edit.metadata?.node_id,
    data: {
      id: edit.id,
      type: 'update-node',
      label: edit.metadata?.field || 'property',
    } as UIGraphNode,
  };
}

/**
 * Adapt Execution to SessionRecording for SessionReplayViewer
 */
export function adaptExecutionToRecording(executions: Execution[]): UISessionRecording {
  const events: UISessionEvent[] = executions.slice(0, 10).map((ex) => ({
    id: ex.id || crypto.randomUUID(),
    type: ex.status,
    userId: ex.graphId || 'system',
    timestamp: new Date(ex.startedAt || Date.now()).getTime(),
    data: { graphId: ex.graphId, status: ex.status, nodeResults: ex.nodeResults },
  }));
  return {
    id: executions[0]?.id || 'session-1',
    name: 'Execution Replay',
    events,
    duration: 60000,
    createdAt: Date.now(),
  };
}

/**
 * Adapt Execution to ExecutionBookmark array for SharedExecutionView
 */
export function adaptExecutionToBookmarks(executions: Execution[]): UIExecutionBookmark[] {
  return executions.flatMap((ex, execIndex) =>
    (ex.nodeResults || []).map((nr, nodeIndex) => ({
      id: `${ex.id}-bookmark-${nodeIndex}`,
      userId: 'system',
      userName: 'System',
      executionId: ex.id || `exec-${execIndex}`,
      stepIndex: nodeIndex,
      timestamp: new Date(ex.startedAt || Date.now()).getTime(),
      pausedAt: undefined,
    }))
  );
}

/**
 * Adapt CollabEvent to ActivityItem for TeamActivityFeed
 */
export function adaptCollabEvent(event: CollabEvent): UIActivityItem {
  return {
    id: event.id,
    type: event.event_type || 'action',
    userId: event.created_by || 'unknown',
    userName: event.metadata?.user_name || event.created_by || 'Unknown',
    timestamp: new Date(event.created_at).getTime(),
    description: event.metadata?.action || event.event_type || 'performed action',
    metadata: {
      target: event.metadata?.target,
      icon: event.metadata?.icon,
      isAI: event.metadata?.is_ai,
    },
  };
}

/**
 * Adapt CollabEvent array for TeamActivityFeed
 */
export function adaptCollabEventList(events: CollabEvent[]): UIActivityItem[] {
  return events.map(adaptCollabEvent);
}

/**
 * Adapt Execution to TaskAssignment for AIHumanTaskAssignmentBoard
 */
export function adaptExecutionToTask(execution: Execution, index: number): UITaskAssignment {
  return {
    id: execution.id || `task-${index}`,
    title: execution.graphId || `Execution ${index + 1}`,
    status: execution.status === 'completed' ? 'done' : execution.status === 'failed' ? 'blocked' : 'in-progress',
    priority: 'medium',
    assignees: [{ id: 'agent-1', name: 'Agent', type: 'ai' }],
    createdAt: new Date(execution.startedAt || Date.now()).getTime(),
  };
}

/**
 * Adapt Execution array to TaskAssignment array for AIHumanTaskAssignmentBoard
 */
export function adaptExecutionListToTasks(executions: Execution[]): UITaskAssignment[] {
  return executions.slice(0, 5).map(adaptExecutionToTask);
}
