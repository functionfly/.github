/**
 * Collaboration Store
 * Global state management for multiplayer & collaboration components
 */

import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'

// ============================================================================
// Types
// ============================================================================

export interface CodePosition {
  line: number
  column: number
  offset?: number
}

export interface CodeRange {
  start: CodePosition
  end: CodePosition
}

export interface CollaboratorPresence {
  id: string
  userId: string
  userName: string
  userAvatar?: string
  color: string
  cursor?: { line: number; column: number; filePath?: string }
  selection?: { start: CodePosition; end: CodePosition }
  status: 'active' | 'idle' | 'away'
  lastActivity: number
  followedBy?: string[]
}

export interface VoiceParticipant {
  id: string
  userId: string
  userName: string
  userAvatar?: string
  isMuted: boolean
  isDeafened: boolean
  isSpeaking: boolean
  isHandRaised: boolean
  joinedAt: number
}

export interface VoiceSession {
  id: string
  name: string
  participants: VoiceParticipant[]
  isActive: boolean
  startedAt?: number
  maxParticipants?: number
}

export interface ExecutionBookmark {
  id: string
  userId: string
  userName: string
  timestamp: number
  executionId: string
  stepIndex: number
  pausedAt?: number
}

export interface GraphNode {
  id: string
  type: string
  label: string
  position: { x: number; y: number }
  color?: string
  metadata?: Record<string, unknown>
  lockedBy?: string
}

export interface GraphEdge {
  id: string
  source: string
  target: string
  label?: string
  type?: string
  weight?: number
}

export interface GraphOperation {
  id: string
  type: 'add-node' | 'delete-node' | 'add-edge' | 'delete-edge' | 'move-node' | 'update-node'
  userId: string
  userName: string
  timestamp: number
  data: unknown
  applied: boolean
}

export interface Annotation {
  id: string
  type: 'comment' | 'suggestion' | 'highlight' | 'todo' | 'issue'
  content: string
  author: { id: string; name: string; avatar?: string }
  createdAt: number
  updatedAt?: number
  resolvedAt?: number
  resolvedBy?: string
  resolved?: boolean
  position: { filePath: string; startLine: number; endLine: number; startColumn?: number; endColumn?: number }
  replies?: Annotation[]
  reactions?: Array<{ emoji: string; userId: string }>
  mentions?: string[]
}

export interface SessionEvent {
  id: string
  type: 'cursor-move' | 'selection' | 'edit' | 'navigation' | 'panel-open' | 'panel-close'
  userId: string
  userName: string
  timestamp: number
  data: unknown
}

export interface SessionRecording {
  id: string
  name: string
  duration: number
  participants: CollaboratorPresence[]
  events: SessionEvent[]
  startedAt: number
  endedAt?: number
}

export interface ActivityItem {
  id: string
  type: 'code-edit' | 'comment' | 'review' | 'deployment' | 'function-run' | 'collaborator-join' | 'collaborator-leave' | 'file-create' | 'file-delete' | 'branch-create' | 'merge'
  userId: string
  userName: string
  userAvatar?: string
  timestamp: number
  description: string
  metadata?: {
    filePath?: string
    functionId?: string
    deploymentId?: string
    branchName?: string
    mergeRequestId?: string
  }
  collaborators?: Array<{ userId: string; userName: string }>
}

export interface MemoryCard {
  id: string
  content: string
  author: { id: string; name: string; avatar?: string }
  createdAt: number
  updatedAt?: number
  position: { x: number; y: number }
  color?: string
  tags?: string[]
  linkedCards?: string[]
  collaborators?: string[]
}

export interface ConflictMarker {
  id: string
  type: 'ours' | 'theirs' | 'base'
  content: string
  startLine: number
  endLine: number
}

export interface ConflictResolution {
  id: string
  filePath: string
  markers: ConflictMarker[]
  resolvedContent?: string
  resolvedAt?: number
  resolvedBy?: string
  status: 'unresolved' | 'resolved' | 'merged'
}

export interface ReviewComment {
  id: string
  author: { id: string; name: string; avatar?: string }
  content: string
  createdAt: number
  lineNumber?: number
  filePath?: string
  status?: 'pending' | 'resolved' | 'outdated'
}

export interface ReviewSession {
  id: string
  name: string
  description?: string
  status: 'open' | 'in-review' | 'approved' | 'changes-requested' | 'merged'
  author: { id: string; name: string; avatar?: string }
  createdAt: number
  updatedAt?: number
  participants: Array<{ userId: string; userName: string; role: 'reviewer' | 'author' }>
  comments: ReviewComment[]
  timeline: Array<{ id: string; event: string; userId: string; userName: string; timestamp: number; details?: string }>
}

export interface PromptSegment {
  id: string
  type: 'text' | 'variable' | 'function' | 'loop' | 'conditional' | 'ai-template'
  content: string
  metadata?: { variableName?: string; functionName?: string; outputType?: string }
  editedBy?: string
  lastEditAt?: number
}

export interface DriverNavigator {
  userId: string
  role: 'driver' | 'navigator'
  joinedAt: number
  cursorPosition?: { line: number; column: number; filePath: string }
}

export interface PairProgrammingSession {
  id: string
  drivers: DriverNavigator[]
  currentDriverId?: string
  currentNavigatorId?: string
  isActive: boolean
  startedAt: number
  history: Array<{ driverId: string; navigatorId: string; switchedAt: number }>
}

export interface TaskAssignee {
  userId: string
  userName: string
  avatar?: string
  type: 'human' | 'ai'
}

export interface TaskAssignment {
  id: string
  title: string
  description?: string
  priority: 'low' | 'medium' | 'high' | 'urgent'
  status: 'todo' | 'in-progress' | 'review' | 'done' | 'blocked'
  assignees: TaskAssignee[]
  suggestedAssignee?: { userId: string; userName: string; confidence: number; reasoning?: string }
  createdAt: number
  updatedAt?: number
  dueDate?: number
  labels?: string[]
  dependencies?: string[]
}

// ============================================================================
// Store Interface
// ============================================================================

interface CollaborationState {
  presences: CollaboratorPresence[]
  currentUserId: string | null
  voiceSession: VoiceSession | null
  isVoiceConnected: boolean
  sharedExecutionId: string | null
  sharedExecutionStep: number
  sharedExecutionPaused: boolean
  bookmarks: ExecutionBookmark[]
  graphNodes: GraphNode[]
  graphEdges: GraphEdge[]
  graphOperations: GraphOperation[]
  lockedNodeIds: string[]
  annotations: Annotation[]
  selectedAnnotationId: string | null
  currentRecording: SessionRecording | null
  replayTime: number
  replaySpeed: number
  isReplaying: boolean
  activities: ActivityItem[]
  hasMoreActivities: boolean
  isLoadingActivities: boolean
  memoryCards: MemoryCard[]
  conflicts: ConflictResolution[]
  selectedConflictId: string | null
  currentReview: ReviewSession | null
  promptSegments: PromptSegment[]
  pairSession: PairProgrammingSession | null
  tasks: TaskAssignment[]
  selectedTaskId: string | null

  // Actions (defined in the immer setup; exposed on the type for hook access)
  setPresences: (presences: CollaboratorPresence[]) => void
  addPresence: (presence: CollaboratorPresence) => void
  updatePresence: (userId: string, updates: Partial<CollaboratorPresence>) => void
  removePresence: (userId: string) => void
  setCurrentUserId: (userId: string | null) => void
  setVoiceSession: (session: VoiceSession | null) => void
  setVoiceConnected: (connected: boolean) => void
  updateVoiceParticipant: (participantId: string, updates: unknown) => void
  setSharedExecutionId: (id: string | null) => void
  setSharedExecutionStep: (step: number) => void
  setSharedExecutionPaused: (paused: boolean) => void
  addBookmark: (bookmark: ExecutionBookmark) => void
  removeBookmark: (bookmarkId: string) => void
  setGraphNodes: (nodes: GraphNode[]) => void
  setGraphEdges: (edges: GraphEdge[]) => void
  addGraphNode: (node: GraphNode) => void
  updateGraphNode: (nodeId: string, updates: Partial<GraphNode>) => void
  removeGraphNode: (nodeId: string) => void
  addGraphEdge: (edge: GraphEdge) => void
  removeGraphEdge: (edgeId: string) => void
  addGraphOperation: (operation: GraphOperation) => void
  lockNode: (nodeId: string) => void
  unlockNode: (nodeId: string) => void
  setAnnotations: (annotations: Annotation[]) => void
  addAnnotation: (annotation: Annotation) => void
  updateAnnotation: (id: string, updates: unknown) => void
  resolveAnnotation: (id: string) => void
  deleteAnnotation: (id: string) => void
  setSelectedAnnotationId: (id: string | null) => void
  setCurrentRecording: (recording: SessionRecording | null) => void
  setReplayTime: (time: number) => void
  setReplaySpeed: (speed: number) => void
  setIsReplaying: (replaying: boolean) => void
  setActivities: (activities: ActivityItem[]) => void
  addActivity: (activity: ActivityItem) => void
  setHasMoreActivities: (hasMore: boolean) => void
  setIsLoadingActivities: (loading: boolean) => void
  setMemoryCards: (cards: MemoryCard[]) => void
  addMemoryCard: (card: MemoryCard) => void
  updateMemoryCard: (cardId: string, updates: Partial<MemoryCard>) => void
  moveMemoryCard: (cardId: string, position: { x: number; y: number }) => void
  deleteMemoryCard: (cardId: string) => void
  linkMemoryCards: (cardId: string, targetCardId: string) => void
  setConflicts: (conflicts: ConflictResolution[]) => void
  selectConflict: (id: string | null) => void
  resolveConflict: (id: string, resolution: unknown) => void
  dismissConflict: (id: string) => void
  setCurrentReview: (review: ReviewSession | null) => void
  addReviewComment: (comment: ReviewComment) => void
  resolveReviewComment: (commentId: string) => void
  updateReviewStatus: (status: ReviewSession['status']) => void
  setPromptSegments: (segments: PromptSegment[]) => void
  addPromptSegment: (segment: PromptSegment) => void
  updatePromptSegment: (segmentId: string, content: string) => void
  deletePromptSegment: (segmentId: string) => void
  setPairSession: (session: PairProgrammingSession | null) => void
  switchPairRoles: () => void
  handoverDriver: (newDriverId: string) => void
  endPairSession: () => void
  setTasks: (tasks: TaskAssignment[]) => void
  addTask: (task: TaskAssignment) => void
  updateTask: (taskId: string, updates: Partial<TaskAssignment>) => void
  deleteTask: (taskId: string) => void
  selectTask: (taskId: string | null) => void
  assignTask: (taskId: string, assignee: TaskAssignee) => void
  acceptAISuggestion: (taskId: string) => void
  rejectAISuggestion: (taskId: string) => void
}

// ============================================================================
// Store
// ============================================================================

export const useCollaborationStore = create<CollaborationState>()(
  immer((set) => ({
    presences: [],
    currentUserId: null,
    voiceSession: null,
    isVoiceConnected: false,
    sharedExecutionId: null,
    sharedExecutionStep: 0,
    sharedExecutionPaused: false,
    bookmarks: [],
    graphNodes: [],
    graphEdges: [],
    graphOperations: [],
    lockedNodeIds: [],
    annotations: [],
    selectedAnnotationId: null,
    currentRecording: null,
    replayTime: 0,
    replaySpeed: 1,
    isReplaying: false,
    activities: [],
    hasMoreActivities: false,
    isLoadingActivities: false,
    memoryCards: [],
    conflicts: [],
    selectedConflictId: null,
    currentReview: null,
    promptSegments: [],
    pairSession: null,
    tasks: [],
    selectedTaskId: null,

    setPresences: (presences) =>
      set((state) => { state.presences = presences }),
    addPresence: (presence) =>
      set((state) => { state.presences.push(presence) }),
    updatePresence: (userId, updates) =>
      set((state) => {
        const presence = state.presences.find((p) => p.userId === userId)
        if (presence) Object.assign(presence, updates)
      }),
    removePresence: (userId) =>
      set((state) => { state.presences = state.presences.filter((p) => p.userId !== userId) }),
    setCurrentUserId: (userId) =>
      set((state) => { state.currentUserId = userId }),

    setVoiceSession: (session) =>
      set((state) => { state.voiceSession = session }),
    setVoiceConnected: (connected) =>
      set((state) => { state.isVoiceConnected = connected }),
    updateVoiceParticipant: (participantId, updates) =>
      set((state) => {
        if (state.voiceSession) {
          const participant = state.voiceSession.participants.find((p) => p.id === participantId)
          if (participant) Object.assign(participant, updates)
        }
      }),

    setSharedExecutionId: (id) =>
      set((state) => { state.sharedExecutionId = id }),
    setSharedExecutionStep: (step) =>
      set((state) => { state.sharedExecutionStep = step }),
    setSharedExecutionPaused: (paused) =>
      set((state) => { state.sharedExecutionPaused = paused }),
    addBookmark: (bookmark) =>
      set((state) => { state.bookmarks.push(bookmark) }),
    removeBookmark: (bookmarkId) =>
      set((state) => { state.bookmarks = state.bookmarks.filter((b) => b.id !== bookmarkId) }),

    setGraphNodes: (nodes) =>
      set((state) => { state.graphNodes = nodes }),
    setGraphEdges: (edges) =>
      set((state) => { state.graphEdges = edges }),
    addGraphNode: (node) =>
      set((state) => { state.graphNodes.push(node) }),
    updateGraphNode: (nodeId, updates) =>
      set((state) => {
        const node = state.graphNodes.find((n) => n.id === nodeId)
        if (node) Object.assign(node, updates)
      }),
    removeGraphNode: (nodeId) =>
      set((state) => {
        state.graphNodes = state.graphNodes.filter((n) => n.id !== nodeId)
        state.graphEdges = state.graphEdges.filter((e) => e.source !== nodeId && e.target !== nodeId)
      }),
    addGraphEdge: (edge) =>
      set((state) => { state.graphEdges.push(edge) }),
    removeGraphEdge: (edgeId) =>
      set((state) => { state.graphEdges = state.graphEdges.filter((e) => e.id !== edgeId) }),
    addGraphOperation: (operation) =>
      set((state) => { state.graphOperations.push(operation) }),
    lockNode: (nodeId) =>
      set((state) => { if (!state.lockedNodeIds.includes(nodeId)) state.lockedNodeIds.push(nodeId) }),
    unlockNode: (nodeId) =>
      set((state) => { state.lockedNodeIds = state.lockedNodeIds.filter((id) => id !== nodeId) }),

    setAnnotations: (annotations) =>
      set((state) => { state.annotations = annotations }),
    addAnnotation: (annotation) =>
      set((state) => { state.annotations.push(annotation) }),
    updateAnnotation: (id, updates) =>
      set((state) => {
        const annotation = state.annotations.find((a) => a.id === id)
        if (annotation) Object.assign(annotation, updates)
      }),
    resolveAnnotation: (id) =>
      set((state) => {
        const annotation = state.annotations.find((a) => a.id === id)
        if (annotation) {
          annotation.resolved = true
          annotation.resolvedAt = Date.now()
          annotation.resolvedBy = state.currentUserId || undefined
        }
      }),
    deleteAnnotation: (id) =>
      set((state) => { state.annotations = state.annotations.filter((a) => a.id !== id) }),
    setSelectedAnnotationId: (id) =>
      set((state) => { state.selectedAnnotationId = id }),

    setCurrentRecording: (recording) =>
      set((state) => { state.currentRecording = recording }),
    setReplayTime: (time) =>
      set((state) => { state.replayTime = time }),
    setReplaySpeed: (speed) =>
      set((state) => { state.replaySpeed = speed }),
    setIsReplaying: (replaying) =>
      set((state) => { state.isReplaying = replaying }),

    setActivities: (activities) =>
      set((state) => { state.activities = activities }),
    addActivity: (activity) =>
      set((state) => { state.activities.unshift(activity) }),
    setHasMoreActivities: (hasMore) =>
      set((state) => { state.hasMoreActivities = hasMore }),
    setIsLoadingActivities: (loading) =>
      set((state) => { state.isLoadingActivities = loading }),

    setMemoryCards: (cards) =>
      set((state) => { state.memoryCards = cards }),
    addMemoryCard: (card) =>
      set((state) => { state.memoryCards.push(card) }),
    updateMemoryCard: (cardId, updates) =>
      set((state) => {
        const card = state.memoryCards.find((c) => c.id === cardId)
        if (card) Object.assign(card, updates)
      }),
    moveMemoryCard: (cardId, position) =>
      set((state) => {
        const card = state.memoryCards.find((c) => c.id === cardId)
        if (card) card.position = position
      }),
    deleteMemoryCard: (cardId) =>
      set((state) => { state.memoryCards = state.memoryCards.filter((c) => c.id !== cardId) }),
    linkMemoryCards: (cardId, targetCardId) =>
      set((state) => {
        const card = state.memoryCards.find((c) => c.id === cardId)
        if (card) {
          if (!card.linkedCards) card.linkedCards = []
          card.linkedCards.push(targetCardId)
        }
      }),

    setConflicts: (conflicts) =>
      set((state) => { state.conflicts = conflicts }),
    selectConflict: (id) =>
      set((state) => { state.selectedConflictId = id }),
    resolveConflict: (id, resolution) =>
      set((state) => {
        const conflict = state.conflicts.find((c) => c.id === id)
        if (conflict) {
          conflict.status = resolution === 'merged' ? 'merged' : 'resolved'
          conflict.resolvedAt = Date.now()
          conflict.resolvedBy = state.currentUserId || undefined
        }
      }),
    dismissConflict: (id) =>
      set((state) => {
        state.conflicts = state.conflicts.filter((c) => c.id !== id)
        if (state.selectedConflictId === id) state.selectedConflictId = null
      }),

    setCurrentReview: (review) =>
      set((state) => { state.currentReview = review }),
    addReviewComment: (comment) =>
      set((state) => {
        if (state.currentReview) {
          state.currentReview.comments.push(comment)
          state.currentReview.updatedAt = Date.now()
        }
      }),
    resolveReviewComment: (commentId) =>
      set((state) => {
        if (state.currentReview) {
          const comment = state.currentReview.comments.find((c) => c.id === commentId)
          if (comment) comment.status = 'resolved'
        }
      }),
    updateReviewStatus: (status) =>
      set((state) => {
        if (state.currentReview) {
          state.currentReview.status = status
          state.currentReview.updatedAt = Date.now()
        }
      }),

    setPromptSegments: (segments) =>
      set((state) => { state.promptSegments = segments }),
    addPromptSegment: (segment) =>
      set((state) => { state.promptSegments.push(segment) }),
    updatePromptSegment: (segmentId, content) =>
      set((state) => {
        const segment = state.promptSegments.find((s) => s.id === segmentId)
        if (segment) {
          segment.content = content
          segment.editedBy = state.currentUserId || undefined
          segment.lastEditAt = Date.now()
        }
      }),
    deletePromptSegment: (segmentId) =>
      set((state) => { state.promptSegments = state.promptSegments.filter((s) => s.id !== segmentId) }),

    setPairSession: (session) =>
      set((state) => { state.pairSession = session }),
    switchPairRoles: () =>
      set((state) => {
        if (state.pairSession && state.currentUserId) {
          const currentDriver = state.pairSession.currentDriverId
          const currentNavigator = state.pairSession.currentNavigatorId
          if (currentDriver && currentNavigator) {
            state.pairSession.currentDriverId = currentNavigator
            state.pairSession.currentNavigatorId = currentDriver
            state.pairSession.history.push({ driverId: currentNavigator, navigatorId: currentDriver, switchedAt: Date.now() })
          }
        }
      }),
    handoverDriver: (newDriverId) =>
      set((state) => {
        if (state.pairSession && state.currentUserId) {
          const oldDriverId = state.pairSession.currentDriverId
          if (oldDriverId && newDriverId !== oldDriverId) {
            state.pairSession.currentNavigatorId = oldDriverId
            state.pairSession.currentDriverId = newDriverId
            state.pairSession.history.push({ driverId: newDriverId, navigatorId: oldDriverId, switchedAt: Date.now() })
          }
        }
      }),
    endPairSession: () =>
      set((state) => { if (state.pairSession) state.pairSession.isActive = false }),

    setTasks: (tasks) =>
      set((state) => { state.tasks = tasks }),
    addTask: (task) =>
      set((state) => { state.tasks.push(task) }),
    updateTask: (taskId, updates) =>
      set((state) => {
        const task = state.tasks.find((t) => t.id === taskId)
        if (task) {
          Object.assign(task, updates)
          task.updatedAt = Date.now()
        }
      }),
    deleteTask: (taskId) =>
      set((state) => { state.tasks = state.tasks.filter((t) => t.id !== taskId) }),
    selectTask: (taskId) =>
      set((state) => { state.selectedTaskId = taskId }),
    assignTask: (taskId, assignee) =>
      set((state) => {
        const task = state.tasks.find((t) => t.id === taskId)
        if (task) {
          const existingIndex = task.assignees.findIndex((a) => a.userId === assignee.userId)
          if (existingIndex === -1) task.assignees.push(assignee)
        }
      }),
    acceptAISuggestion: (taskId) =>
      set((state) => {
        const task = state.tasks.find((t) => t.id === taskId)
        if (task?.suggestedAssignee) {
          task.assignees.push(task.suggestedAssignee as TaskAssignee)
          task.suggestedAssignee = undefined
        }
      }),
    rejectAISuggestion: (taskId) =>
      set((state) => {
        const task = state.tasks.find((t) => t.id === taskId)
        if (task) task.suggestedAssignee = undefined
      }),
  }))
)

export const selectActivePresences = (state: CollaborationState) =>
  state.presences.filter((p) => p.status === 'active')
export const selectSpeakingParticipants = (state: CollaborationState) =>
  state.voiceSession?.participants.filter((p) => p.isSpeaking) || []
export const selectUnresolvedConflicts = (state: CollaborationState) =>
  state.conflicts.filter((c) => c.status === 'unresolved')
