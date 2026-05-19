/**
 * Collaboration Page
 * Multiplayer & Collaboration workspace
 */

import React, { useState, useCallback, useMemo, useEffect } from 'react'
import { cn } from '@functionfly/ui-core'
import { useCollaborationStore } from '../../stores/collaborationStore'
import {
  Users,
  Phone,
  Play,
  GitBranch,
  MessageSquare,
  Activity,
  Lightbulb,
  AlertTriangle,
  Eye,
  Zap,
  CheckSquare,
  Settings,
  Maximize2,
  Minimize2,
  X,
} from 'lucide-react'

import {
  LivePresenceLayer,
  VoiceSessionPanel,
  SharedExecutionView,
  CollaborativeGraphEditor,
  RealtimeAnnotationSystem,
  SessionReplayViewer,
  TeamActivityFeed,
  SharedMemoryBoard,
  ConflictResolutionPanel,
  AsyncReviewTimeline,
  CollaborativePromptEditor,
  LivePairProgrammingView,
  AIHumanTaskAssignmentBoard,
} from '@functionfly/ui-collaboration'

export const CollaborationPage: React.FC = () => {
  const store = useCollaborationStore()
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [showSettings, setShowSettings] = useState(false)
  const [activePanel, setActivePanel] = useState<'presence' | 'voice' | 'execution' | 'graph' | 'annotations' | 'activity' | 'memory' | 'conflicts' | 'review' | 'prompt' | 'pair' | 'tasks'>('presence')

  useEffect(() => {
    if (store.presences.length === 0) {
      store.setPresences([
        { id: 'presence-1', userId: 'user-1', userName: 'Alice Chen', color: '#3B82F6', status: 'active', lastActivity: Date.now(), cursor: { line: 42, column: 15, filePath: 'src/components/Collaboration.tsx' } },
        { id: 'presence-2', userId: 'user-2', userName: 'Bob Smith', color: '#10B981', status: 'active', lastActivity: Date.now() - 30000, cursor: { line: 87, column: 3, filePath: 'src/stores/collaborationStore.ts' } },
        { id: 'presence-3', userId: 'user-3', userName: 'Carol Davis', color: '#8B5CF6', status: 'idle', lastActivity: Date.now() - 120000 },
      ])
      store.setCurrentUserId('user-current')
    }
    if (store.activities.length === 0) {
      store.setActivities([
        { id: 'activity-1', type: 'code-edit', userId: 'user-1', userName: 'Alice Chen', timestamp: Date.now() - 60000, description: 'Updated collaboration store with new actions', metadata: { filePath: 'src/stores/collaborationStore.ts' } },
        { id: 'activity-2', type: 'comment', userId: 'user-2', userName: 'Bob Smith', timestamp: Date.now() - 180000, description: 'Added review comment on PR #234', metadata: { mergeRequestId: '234' } },
        { id: 'activity-3', type: 'collaborator-join', userId: 'user-3', userName: 'Carol Davis', timestamp: Date.now() - 300000, description: 'Joined the collaboration session' },
      ])
    }
    if (store.tasks.length === 0) {
      store.setTasks([
        { id: 'task-1', title: 'Implement real-time presence', description: 'Add WebSocket support for live presence updates', priority: 'high', status: 'in-progress', assignees: [{ userId: 'user-1', userName: 'Alice Chen', type: 'human' }], suggestedAssignee: { userId: 'ai-1', userName: 'AI Assistant', confidence: 0.85, reasoning: 'Based on expertise in real-time systems' }, createdAt: Date.now() - 86400000, labels: ['feature', 'collaboration'] },
        { id: 'task-2', title: 'Design conflict resolution UI', description: 'Create intuitive UI for handling merge conflicts', priority: 'medium', status: 'todo', assignees: [], suggestedAssignee: { userId: 'ai-2', userName: 'AI Assistant', confidence: 0.72 }, createdAt: Date.now() - 172800000, labels: ['design', 'ux'] },
        { id: 'task-3', title: 'Add voice session controls', description: 'Implement mute, deafen, and speaking indicators', priority: 'low', status: 'done', assignees: [{ userId: 'user-2', userName: 'Bob Smith', type: 'human' }], createdAt: Date.now() - 259200000, labels: ['feature'] },
      ])
    }
    if (store.annotations.length === 0) {
      store.setAnnotations([
        { id: 'annotation-1', type: 'comment', content: 'We should use Zustand for state management here for better performance', author: { id: 'user-1', name: 'Alice Chen' }, createdAt: Date.now() - 3600000, position: { filePath: 'src/stores/collaborationStore.ts', startLine: 45, endLine: 48 }, reactions: [{ emoji: '👍', userId: 'user-2' }] },
        { id: 'annotation-2', type: 'suggestion', content: 'Consider adding a debounce to reduce re-renders', author: { id: 'user-2', name: 'Bob Smith' }, createdAt: Date.now() - 7200000, position: { filePath: 'src/stores/collaborationStore.ts', startLine: 120, endLine: 125 } },
      ])
    }
  }, [])

  const panelTabs = useMemo(() => [
    { id: 'presence' as const, label: 'Presence', icon: Users, count: store.presences.length },
    { id: 'voice' as const, label: 'Voice', icon: Phone, count: store.voiceSession?.participants.length },
    { id: 'execution' as const, label: 'Execution', icon: Play, count: store.sharedExecutionId ? 1 : 0 },
    { id: 'graph' as const, label: 'Graph', icon: GitBranch, count: store.graphNodes.length },
    { id: 'annotations' as const, label: 'Annotations', icon: MessageSquare, count: store.annotations.filter(a => !a.resolved).length },
    { id: 'activity' as const, label: 'Activity', icon: Activity, count: store.activities.length },
    { id: 'memory' as const, label: 'Memory', icon: Lightbulb, count: store.memoryCards.length },
    { id: 'conflicts' as const, label: 'Conflicts', icon: AlertTriangle, count: store.conflicts.filter(c => c.status === 'unresolved').length },
    { id: 'review' as const, label: 'Review', icon: Eye, count: store.currentReview ? 1 : 0 },
    { id: 'prompt' as const, label: 'Prompt', icon: MessageSquare, count: store.promptSegments.length },
    { id: 'pair' as const, label: 'Pair', icon: Users, count: store.pairSession?.isActive ? 1 : 0 },
    { id: 'tasks' as const, label: 'Tasks', icon: CheckSquare, count: store.tasks.filter(t => t.status !== 'done').length },
  ], [store.presences.length, store.voiceSession?.participants.length, store.sharedExecutionId, store.graphNodes.length, store.annotations, store.activities.length, store.memoryCards.length, store.conflicts, store.currentReview, store.promptSegments.length, store.pairSession, store.tasks])

  const renderPanel = () => {
    switch (activePanel) {
      case 'presence':
        return (
          <div className="space-y-4">
            <LivePresenceLayer presences={store.presences} currentUserId={store.currentUserId || undefined} showAvatars showCursors showStatus onUserClick={(p) => console.log('User clicked:', p)} onFollowUser={(id) => console.log('Follow:', id)} />
            <div className="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] p-4">
              <div className="text-sm font-medium mb-3">Active Collaborators</div>
              <div className="space-y-2">
                {store.presences.map((presence) => (
                  <div key={presence.id} className="flex items-center justify-between p-2 rounded hover:bg-[var(--color-bg-secondary)]">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-full flex items-center justify-center text-white text-sm font-medium" style={{ backgroundColor: presence.color }}>{presence.userName.charAt(0).toUpperCase()}</div>
                      <div>
                        <div className="text-sm font-medium">{presence.userName}</div>
                        {presence.cursor && <div className="text-xs text-[var(--color-text-tertiary)]">Line {presence.cursor.line}, Col {presence.cursor.column}</div>}
                      </div>
                    </div>
                    <span className={cn('w-2 h-2 rounded-full', presence.status === 'active' ? 'bg-green-500' : presence.status === 'idle' ? 'bg-yellow-500' : 'bg-gray-400')} />
                  </div>
                ))}
              </div>
            </div>
          </div>
        )
      case 'voice':
        return <VoiceSessionPanel session={store.voiceSession} currentUserId={store.currentUserId || undefined} onParticipantMute={(id) => store.updateVoiceParticipant(id, { isMuted: !store.voiceSession?.participants.find(p => p.id === id)?.isMuted })} onParticipantDeafen={(id) => store.updateVoiceParticipant(id, { isDeafened: !store.voiceSession?.participants.find(p => p.id === id)?.isDeafened })} onParticipantKick={(id) => console.log('Kick:', id)} onRaiseHand={() => console.log('Raise hand')} onLeaveSession={() => { store.setVoiceSession(null); store.setVoiceConnected(false) }} />
      case 'execution':
        return <SharedExecutionView executionId={store.sharedExecutionId || 'no-execution'} bookmarks={store.bookmarks} currentStep={store.sharedExecutionStep} isPaused={store.sharedExecutionPaused} participants={store.presences} onBookmarkCreate={(step) => store.addBookmark({ id: `bookmark-${Date.now()}`, userId: store.currentUserId || '', userName: 'You', timestamp: Date.now(), executionId: store.sharedExecutionId || '', stepIndex: step })} onBookmarkNavigate={(id) => { const b = store.bookmarks.find(b => b.id === id); if (b) store.setSharedExecutionStep(b.stepIndex) }} onStepChange={(step) => store.setSharedExecutionStep(step)} onPlayPause={() => store.setSharedExecutionPaused(!store.sharedExecutionPaused)} />
      case 'graph':
        return <CollaborativeGraphEditor nodes={store.graphNodes} edges={store.graphEdges} operations={store.graphOperations} currentUserId={store.currentUserId || undefined} lockedNodeIds={store.lockedNodeIds} onNodeClick={(node) => console.log('Node:', node)} onNodeMove={(id, pos) => store.updateGraphNode(id, { position: pos })} onEdgeCreate={(src, tgt) => store.addGraphEdge({ id: `edge-${Date.now()}`, source: src, target: tgt })} onNodeLock={(id) => store.lockNode(id)} onNodeUnlock={(id) => store.unlockNode(id)} />
      case 'annotations':
        return <RealtimeAnnotationSystem annotations={store.annotations} currentUserId={store.currentUserId || undefined} onAnnotationCreate={(a) => store.addAnnotation({ ...a, id: `annotation-${Date.now()}`, createdAt: Date.now() })} onAnnotationUpdate={(id, content) => store.updateAnnotation(id, { content, updatedAt: Date.now() })} onAnnotationResolve={(id) => store.resolveAnnotation(id)} onAnnotationDelete={(id) => store.deleteAnnotation(id)} onReplyCreate={(parentId, content) => { const parent = store.annotations.find(a => a.id === parentId); if (parent) store.updateAnnotation(parentId, { replies: [...(parent.replies || []), { id: `reply-${Date.now()}`, type: 'comment', content, author: { id: store.currentUserId || '', name: 'You' }, createdAt: Date.now(), position: parent.position }] }) }} onReactionAdd={(id, emoji) => { const a = store.annotations.find(a => a.id === id); if (a) store.updateAnnotation(id, { reactions: [...(a.reactions || []), { emoji, userId: store.currentUserId || '' }] }) }} />
      case 'activity':
        return <TeamActivityFeed activities={store.activities} isLoading={store.isLoadingActivities} hasMore={store.hasMoreActivities} onLoadMore={() => console.log('Load more')} onActivityClick={(a) => console.log('Activity:', a)} />
      case 'memory':
        return <SharedMemoryBoard cards={store.memoryCards} currentUserId={store.currentUserId || undefined} onCardCreate={(content, pos) => store.addMemoryCard({ id: `card-${Date.now()}`, content, author: { id: store.currentUserId || '', name: 'You' }, createdAt: Date.now(), position: pos })} onCardUpdate={(id, content) => store.updateMemoryCard(id, { content, updatedAt: Date.now() })} onCardMove={(id, pos) => store.moveMemoryCard(id, pos)} onCardDelete={(id) => store.deleteMemoryCard(id)} onCardLink={(id, targetId) => store.linkMemoryCards(id, targetId)} onCardCollaboratorAdd={(id, userId) => { const card = store.memoryCards.find(c => c.id === id); if (card) store.updateMemoryCard(id, { collaborators: [...(card.collaborators || []), userId] }) }} />
      case 'conflicts':
        return <ConflictResolutionPanel conflicts={store.conflicts} selectedConflictId={store.selectedConflictId} onConflictSelect={(c) => store.selectConflict(c.id)} onResolutionApply={(id, res) => store.resolveConflict(id, res)} onManualResolution={(id, content) => store.resolveConflict(id, 'merged')} onConflictDismiss={(id) => store.dismissConflict(id)} />
      case 'review':
        return <AsyncReviewTimeline review={store.currentReview} currentUserId={store.currentUserId || undefined} onCommentAdd={(content, line, path) => store.addReviewComment({ id: `comment-${Date.now()}`, author: { id: store.currentUserId || '', name: 'You' }, content, createdAt: Date.now(), lineNumber: line, filePath: path, status: 'pending' })} onCommentResolve={(id) => store.resolveReviewComment(id)} onCommentReply={(id, content) => console.log('Reply:', id, content)} onStatusChange={(status) => store.updateReviewStatus(status)} />
      case 'prompt':
        return <CollaborativePromptEditor segments={store.promptSegments} currentUserId={store.currentUserId || undefined} onSegmentUpdate={(id, content) => store.updatePromptSegment(id, content)} onSegmentAdd={(type, content, afterId) => store.addPromptSegment({ id: `segment-${Date.now()}`, type, content, editedBy: store.currentUserId || undefined, lastEditAt: Date.now() })} onSegmentDelete={(id) => store.deletePromptSegment(id)} onVariableInsert={(name, value) => store.addPromptSegment({ id: `segment-${Date.now()}`, type: 'variable', content: `{{${name}: "${value}"}}`, metadata: { variableName: name }, editedBy: store.currentUserId || undefined, lastEditAt: Date.now() })} onExecute={() => console.log('Execute')} />
      case 'pair':
        return <LivePairProgrammingView session={store.pairSession} currentUserId={store.currentUserId || undefined} codeContent="// Code content appears here" onRoleSwitch={() => store.switchPairRoles()} onDriverHandOver={(id) => store.handoverDriver(id)} onSessionEnd={() => store.endPairSession()} />
      case 'tasks':
        return <AIHumanTaskAssignmentBoard tasks={store.tasks} selectedTaskId={store.selectedTaskId || null} onTaskSelect={(t) => store.selectTask(t.id)} onTaskUpdate={(id, updates) => store.updateTask(id, updates)} onTaskAssign={(id, assignee) => store.assignTask(id, assignee)} onAISuggestionAccept={(id) => store.acceptAISuggestion(id)} onAISuggestionReject={(id) => store.rejectAISuggestion(id)} onTaskCreate={(task) => store.addTask({ ...task, id: `task-${Date.now()}`, createdAt: Date.now() })} />
      default:
        return null
    }
  }

  return (
    <div className={cn('h-full flex flex-col bg-[var(--color-bg-primary)]', isFullscreen && 'fixed inset-0 z-50')}>
      <div className="flex items-center justify-between px-6 py-4 border-b border-[var(--color-border)]">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <Users className="w-5 h-5 text-[var(--color-aviation-accent)]" />
            <h1 className="text-xl font-semibold">Collaboration</h1>
          </div>
          <div className="flex items-center gap-2 text-sm text-[var(--color-text-secondary)]">
            <span className="flex items-center gap-1"><span className="w-2 h-2 rounded-full bg-green-500" />{store.presences.filter(p => p.status === 'active').length} active</span>
            {store.voiceSession?.isActive && <span className="flex items-center gap-1 px-2 py-0.5 rounded bg-green-500/20 text-green-400"><Phone className="w-3 h-3" />Voice active</span>}
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={() => setShowSettings(!showSettings)} className="p-2 rounded hover:bg-[var(--color-bg-tertiary)]"><Settings className="w-4 h-4" /></button>
          <button onClick={() => setIsFullscreen(!isFullscreen)} className="p-2 rounded hover:bg-[var(--color-bg-tertiary)]">{isFullscreen ? <Minimize2 className="w-4 h-4" /> : <Maximize2 className="w-4 h-4" />}</button>
        </div>
      </div>
      <div className="flex-1 overflow-auto p-6">
        <div className="max-w-7xl mx-auto space-y-6">
          <div className="grid grid-cols-4 gap-4">
            <div className="p-4 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-secondary)]">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-blue-500/20"><Users className="w-5 h-5 text-blue-400" /></div>
                <div><div className="text-2xl font-semibold">{store.presences.length}</div><div className="text-xs text-[var(--color-text-secondary)]">Collaborators</div></div>
              </div>
            </div>
            <div className="p-4 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-secondary)]">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-purple-500/20"><MessageSquare className="w-5 h-5 text-purple-400" /></div>
                <div><div className="text-2xl font-semibold">{store.annotations.filter(a => !a.resolved).length}</div><div className="text-xs text-[var(--color-text-secondary)]">Open Annotations</div></div>
              </div>
            </div>
            <div className="p-4 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-secondary)]">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-orange-500/20"><AlertTriangle className="w-5 h-5 text-orange-400" /></div>
                <div><div className="text-2xl font-semibold">{store.conflicts.filter(c => c.status === 'unresolved').length}</div><div className="text-xs text-[var(--color-text-secondary)]">Unresolved Conflicts</div></div>
              </div>
            </div>
            <div className="p-4 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-secondary)]">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-green-500/20"><CheckSquare className="w-5 h-5 text-green-400" /></div>
                <div><div className="text-2xl font-semibold">{store.tasks.filter(t => t.status !== 'done').length}</div><div className="text-xs text-[var(--color-text-secondary)]">Active Tasks</div></div>
              </div>
            </div>
          </div>
          <div className="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-secondary)] overflow-hidden">
            <div className="p-4 border-b border-[var(--color-border)]">
              <div className="flex items-center gap-2 mb-4">
                <Users className="w-5 h-5 text-[var(--color-aviation-accent)]" />
                <h2 className="font-medium">Collaboration Workspace</h2>
                <LivePresenceLayer presences={store.presences} currentUserId={store.currentUserId || undefined} showAvatars showCursors={false} showStatus={false} />
              </div>
              <div className="flex items-center gap-1 overflow-x-auto pb-2">
                {panelTabs.map((tab) => (
                  <button key={tab.id} onClick={() => setActivePanel(tab.id)} className={cn('flex items-center gap-1.5 px-3 py-1.5 rounded text-xs font-medium whitespace-nowrap transition-colors', activePanel === tab.id ? 'bg-[var(--color-aviation-accent)] text-white' : 'bg-[var(--color-bg-tertiary)] text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]/80')}>
                    <tab.icon className="w-3.5 h-3.5" />
                    {tab.label}
                    {typeof tab.count === 'number' && tab.count > 0 && <span className={cn('px-1.5 py-0.5 rounded-full text-[10px]', activePanel === tab.id ? 'bg-white/20' : 'bg-[var(--color-aviation-accent)]/20')}>{tab.count}</span>}
                  </button>
                ))}
              </div>
            </div>
            <div className="p-4 max-h-[600px] overflow-y-auto">{renderPanel()}</div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default CollaborationPage
