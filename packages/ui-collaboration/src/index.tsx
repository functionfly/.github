import React, { useState } from 'react';
import { cn } from '@functionfly/ui-core';
import { Users, Mic, MicOff, Phone, PhoneOff, MessageSquare, Send, Clock, Circle, CheckCircle2, XCircle, Eye, Bot, Play, Pause, SkipBack, Activity, MemoryStick, AlertTriangle, Check, X, Plus } from 'lucide-react';
import type { LivePresenceLayerProps, CollaboratorCursorProps, VoiceSessionPanelProps, SharedExecutionViewProps, CollaborativeGraphEditorProps, RealtimeAnnotationSystemProps, SessionReplayViewerProps, TeamActivityFeedProps, SharedMemoryBoardProps, ConflictResolutionPanelProps, AsyncReviewTimelineProps, CollaborativePromptEditorProps, LivePairProgrammingViewProps, AIHumanTaskAssignmentBoardProps, CollaboratorPresence, VoiceSession, ExecutionBookmark, GraphNode, GraphEdge, Annotation, ActivityItem, MemoryCard, ConflictMarker, ReviewSession, PromptSegment, PairProgrammingSession, TaskAssignment } from './types';

export const LivePresenceLayer: React.FC<LivePresenceLayerProps> = ({ presences, children }) => (
  <div className="live-presence-layer">
    {children}
    {(presences || []).map(p => (
    <div key={p.id} className="presence-indicator" style={{ borderColor: p.color }}>
      <div className="presence-avatar" style={{ backgroundColor: p.color }}>{p.userName.charAt(0)}</div>
    </div>
  ))}</div>
);

export const CollaboratorCursor: React.FC<CollaboratorCursorProps> = ({ presence }) => {
  if (!presence.cursor) return null;
  return <div className="collaborator-cursor" style={{ position: 'absolute', left: presence.cursor.column * 8, top: presence.cursor.line * 20 }}>
    <div style={{ width: 2, height: 20, backgroundColor: presence.color }} />
  </div>;
};

export const VoiceSessionPanel: React.FC<VoiceSessionPanelProps> = ({ session, onLeave }) => {
  if (!session) return null;
  return (
    <div className="voice-session-panel p-4 rounded-lg border">
      <div className="flex justify-between mb-4">
        <h3 className="font-semibold">Voice Session</h3>
        <button onClick={onLeave}><PhoneOff className="w-4 h-4" /></button>
      </div>
      {session.participants.map(p => (
        <div key={p.id} className="flex items-center gap-3 p-2">
          <div className="w-8 h-8 rounded-full bg-blue-500 text-white flex items-center justify-center">{p.userName.charAt(0)}</div>
          <div>{p.userName}</div>
        </div>
      ))}
    </div>
  );
};

export const SharedExecutionView: React.FC<SharedExecutionViewProps> = ({ executionId, bookmarks, onBookmarkJump }) => (
  <div className="p-4 rounded-lg border">
    <div className="flex gap-2 mb-4">
      <span>Execution: {executionId || 'None'}</span>
    </div>
    <div className="flex gap-2">
      {bookmarks.map(b => <button key={b.id} onClick={() => onBookmarkJump?.(b.id)} className="px-3 py-1 rounded bg-gray-100">{b.label}</button>)}
    </div>
  </div>
);

export const CollaborativeGraphEditor: React.FC<CollaborativeGraphEditorProps> = ({ nodes, edges }) => (
  <svg viewBox="0 0 400 300" className="w-full h-full">
    {edges.map(e => {
      const s = nodes.find(n => n.id === e.source), t = nodes.find(n => n.id === e.target);
      if (!s || !t) return null;
      return <line key={e.id} x1={s.x+40} y1={s.y+20} x2={t.x+40} y2={t.y+20} stroke="gray" strokeWidth={2} />;
    })}
    {nodes.map(n => (
      <g key={n.id} transform={`translate(${n.x}, ${n.y})`}>
        <rect width={80} height={40} rx={4} fill="#f3f4f6" stroke="gray" />
        <text x={40} y={25} textAnchor="middle">{n.label}</text>
      </g>
    ))}
  </svg>
);

export const RealtimeAnnotationSystem: React.FC<RealtimeAnnotationSystemProps> = ({ annotations, onResolveAnnotation }) => (
  <div className="p-4 rounded-lg border">
    <h3 className="font-semibold mb-2">Annotations ({annotations.length})</h3>
    {annotations.map(a => (
      <div key={a.id} className="p-2 border-l-4 bg-gray-50 mb-2">
        <div className="flex justify-between">
          <span className="font-medium">{a.author}</span>
          {!a.resolved && <button onClick={() => onResolveAnnotation?.(a.id)}><Check className="w-4 h-4" /></button>}
        </div>
        <p className="text-sm">{a.content}</p>
      </div>
    ))}
  </div>
);

export const SessionReplayViewer: React.FC<SessionReplayViewerProps> = ({ recording }) => (
  <div className="p-4 rounded-lg border">
    <h3 className="font-semibold">{recording.name}</h3>
    {recording.events.slice(0, 5).map(e => (
      <div key={e.id} className="text-sm p-1">{new Date(e.timestamp).toLocaleTimeString()} - {e.type}</div>
    ))}
  </div>
);

export const TeamActivityFeed: React.FC<TeamActivityFeedProps> = ({ activities }) => (
  <div className="p-4 rounded-lg border">
    <h3 className="font-semibold mb-2">Activity</h3>
    {activities.map(a => (
      <div key={a.id} className="flex items-center gap-2 p-2">
        <div className="w-8 h-8 rounded-full bg-blue-500 text-white flex items-center justify-center text-xs">{a.userName.charAt(0)}</div>
        <div>
          <div className="text-sm font-medium">{a.userName}</div>
          <div className="text-xs text-gray-500">{a.description}</div>
        </div>
      </div>
    ))}
  </div>
);

export const SharedMemoryBoard: React.FC<SharedMemoryBoardProps> = ({ cards, onCardCreate, onCardDelete }) => {
  const [title, setTitle] = useState('');
  return (
    <div className="p-4 rounded-lg border">
      <h3 className="font-semibold mb-2">Shared Memory</h3>
      <div className="flex gap-2 mb-4">
        <input value={title} onChange={e => setTitle(e.target.value)} placeholder="Title..." className="border px-2 py-1 rounded" />
        <button onClick={() => { if (title) { onCardCreate?.({ title, content: '', author: 'User', tags: [] }); setTitle(''); } }} className="bg-blue-500 text-white px-3 py-1 rounded"><Plus className="w-4 h-4" /></button>
      </div>
      <div className="grid grid-cols-2 gap-2">
        {cards.map(c => (
          <div key={c.id} className="p-2 border rounded">
            <div className="font-medium text-sm">{c.title}</div>
            <div className="text-xs text-gray-500 mb-1">{c.author}</div>
            <button onClick={() => onCardDelete?.(c.id)} className="text-red-500"><X className="w-3 h-3" /></button>
          </div>
        ))}
      </div>
    </div>
  );
};

export const ConflictResolutionPanel: React.FC<ConflictResolutionPanelProps> = ({ conflicts, onResolve }) => (
  <div className="p-4 rounded-lg border">
    <h3 className="font-semibold mb-2">Conflicts ({conflicts.length})</h3>
    {conflicts.map(c => (
      <div key={c.id} className="p-2 border rounded mb-2">
        <div className="font-medium text-sm">{c.type} conflict</div>
        <div className="text-xs font-mono mb-2">{c.filePath}</div>
        <div className="flex gap-2">
          <button onClick={() => onResolve?.({ conflictId: c.id, resolution: 'accept-ours' })} className="px-2 py-1 bg-blue-100 text-blue-600 rounded text-xs">Ours</button>
          <button onClick={() => onResolve?.({ conflictId: c.id, resolution: 'accept-theirs' })} className="px-2 py-1 bg-purple-100 text-purple-600 rounded text-xs">Theirs</button>
          <button onClick={() => onResolve?.({ conflictId: c.id, resolution: 'merge' })} className="px-2 py-1 bg-green-100 text-green-600 rounded text-xs">Merge</button>
        </div>
      </div>
    ))}
  </div>
);

export const AsyncReviewTimeline: React.FC<AsyncReviewTimelineProps> = ({ session, onCommentAdd, onCommentResolve }) => {
  const [text, setText] = useState('');
  return (
    <div className="p-4 rounded-lg border">
      <div className="flex justify-between mb-2">
        <h3 className="font-semibold">Code Review</h3>
        <span className="text-xs px-2 py-1 rounded bg-yellow-100 text-yellow-600">{session.status}</span>
      </div>
      {session.comments.map(c => (
        <div key={c.id} className={cn('p-2 border rounded mb-2', c.resolved && 'opacity-50')}>
          <div className="flex justify-between">
            <span className="font-medium text-sm">{c.author}</span>
            {!c.resolved && <button onClick={() => onCommentResolve?.(c.id)}><Check className="w-4 h-4" /></button>}
          </div>
          <p className="text-sm">{c.content}</p>
        </div>
      ))}
      <div className="flex gap-2 mt-2">
        <input value={text} onChange={e => setText(e.target.value)} placeholder="Comment..." className="flex-1 border px-2 py-1 rounded text-sm" />
        <button onClick={() => { if (text) { onCommentAdd?.({ author: 'User', content: text, timestamp: Date.now() }); setText(''); } }} className="bg-blue-500 text-white px-3 py-1 rounded"><Send className="w-4 h-4" /></button>
      </div>
    </div>
  );
};

export const CollaborativePromptEditor: React.FC<CollaborativePromptEditorProps> = ({ segments, onSegmentAdd }) => {
  const [input, setInput] = useState('');
  return (
    <div className="p-4 rounded-lg border">
      <h3 className="font-semibold mb-2">Prompt Editor</h3>
      <div className="space-y-2 mb-2">
        {segments.map(s => (
          <div key={s.id} className={cn('p-2 rounded text-sm', s.type === 'code' ? 'bg-gray-100 font-mono' : 'bg-white border')}>
            {s.author && <span className="text-xs text-gray-500">{s.author}: </span>}
            {s.content}
          </div>
        ))}
      </div>
      <div className="flex gap-2">
        <input value={input} onChange={e => setInput(e.target.value)} placeholder="Type..." className="flex-1 border px-2 py-1 rounded text-sm"
          onKeyDown={e => { if (e.key === 'Enter') { onSegmentAdd?.({ type: 'text', content: input, author: 'User' }); setInput(''); }}} />
        <button onClick={() => { if (input) { onSegmentAdd?.({ type: 'text', content: input, author: 'User' }); setInput(''); } }} className="bg-blue-500 text-white px-3 py-1 rounded"><Send className="w-4 h-4" /></button>
      </div>
    </div>
  );
};

export const LivePairProgrammingView: React.FC<LivePairProgrammingViewProps> = ({ session, onSessionEnd }) => (
  <div className="p-4 rounded-lg border">
    <div className="flex justify-between mb-4">
      <h3 className="font-semibold">Pair Programming</h3>
      <div className="flex gap-2">
        <span className={cn('text-xs px-2 py-1 rounded', session.isActive ? 'bg-green-100 text-green-600' : 'bg-gray-100 text-gray-600')}>
          {session.isActive ? 'Active' : 'Inactive'}
        </span>
        <button onClick={onSessionEnd} className="px-2 py-1 bg-red-100 text-red-600 rounded text-xs">End</button>
      </div>
    </div>
    <div className="flex items-center justify-center gap-4">
      <div className="text-center">
        <div className="w-12 h-12 rounded-full bg-blue-500 text-white flex items-center justify-center mx-auto mb-1">{session.driver.driverName.charAt(0)}</div>
        <div className="text-sm font-medium">{session.driver.driverName}</div>
        <div className="text-xs text-gray-500">Driver</div>
      </div>
      <div className="text-gray-400">↔</div>
      <div className="text-center">
        <div className="w-12 h-12 rounded-full bg-purple-500 text-white flex items-center justify-center mx-auto mb-1">{session.driver.navigatorName.charAt(0)}</div>
        <div className="text-sm font-medium">{session.driver.navigatorName}</div>
        <div className="text-xs text-gray-500">Navigator</div>
      </div>
    </div>
  </div>
);

export const AIHumanTaskAssignmentBoard: React.FC<AIHumanTaskAssignmentBoardProps> = ({ tasks, selectedTaskId, onTaskSelect, onAISuggestionAccept, onAISuggestionReject, onTaskCreate, className }) => {
  const [newTitle, setNewTitle] = useState('');
  const getPriorityColor = (p: TaskAssignment['priority']) => {
    switch (p) { case 'urgent': return 'border-red-500 bg-red-50'; case 'high': return 'border-orange-500 bg-orange-50'; case 'medium': return 'border-yellow-500 bg-yellow-50'; case 'low': return 'border-green-500 bg-green-50'; }
  };
  const getStatusIcon = (s: TaskAssignment['status']) => {
    switch (s) { case 'todo': return <Circle className="w-4 h-4" />; case 'in-progress': return <Clock className="w-4 h-4" />; case 'review': return <Eye className="w-4 h-4" />; case 'done': return <CheckCircle2 className="w-4 h-4 text-green-500" />; case 'blocked': return <XCircle className="w-4 h-4 text-red-500" />; }
  };
  return (
    <div className={cn('rounded-lg border bg-white overflow-hidden', className)}>
      <div className="p-3 border-b flex justify-between items-center">
        <div className="flex items-center gap-2"><Users className="w-4 h-4" /><span className="font-medium">AI-Human Task Board</span><span className="text-xs text-gray-500">{tasks.length}</span></div>
        <Bot className="w-4 h-4" />
      </div>
      <div className="max-h-96 overflow-y-auto">
        {tasks.map(t => (
          <div key={t.id} onClick={() => onTaskSelect?.(t)} className={cn('p-3 border-b cursor-pointer hover:bg-gray-50', selectedTaskId === t.id ? 'border-blue-500 bg-blue-50' : '', getPriorityColor(t.priority))}>
            <div className="flex items-start gap-2">
              <div className="mt-0.5">{getStatusIcon(t.status)}</div>
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  <span className="font-medium text-sm">{t.title}</span>
                  <span className={cn('text-[10px] px-1.5 py-0.5 rounded uppercase font-bold', t.priority === 'urgent' ? 'bg-red-500 text-white' : t.priority === 'high' ? 'bg-orange-500 text-white' : t.priority === 'medium' ? 'bg-yellow-500 text-black' : 'bg-green-500 text-white')}>{t.priority}</span>
                </div>
                <div className="flex items-center gap-1 mt-1">
                  {t.assignees.map(a => (
                    <div key={a.id} className={cn('w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-medium', a.type === 'ai' ? 'bg-purple-500 text-white' : 'bg-blue-500 text-white')} title={a.name}>{a.name.charAt(0)}</div>
                  ))}
                  {t.aiSuggestion && (
                    <div className="flex gap-1 ml-auto">
                      <button onClick={e => { e.stopPropagation(); onAISuggestionAccept?.(t.id); }} className="p-1 bg-green-100 text-green-600 rounded"><Check className="w-3 h-3" /></button>
                      <button onClick={e => { e.stopPropagation(); onAISuggestionReject?.(t.id); }} className="p-1 bg-red-100 text-red-600 rounded"><X className="w-3 h-3" /></button>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
      <div className="p-2 border-t flex gap-2">
        <input value={newTitle} onChange={e => setNewTitle(e.target.value)} placeholder="New task..." className="flex-1 border px-2 py-1 rounded text-sm"
          onKeyDown={e => { if (e.key === 'Enter') { onTaskCreate?.({ title: newTitle, priority: 'medium', status: 'todo', assignees: [] }); setNewTitle(''); }}} />
        <button onClick={() => { if (newTitle) { onTaskCreate?.({ title: newTitle, priority: 'medium', status: 'todo', assignees: [] }); setNewTitle(''); }}} className="bg-blue-500 text-white px-3 py-1 rounded"><Plus className="w-4 h-4" /></button>
      </div>
    </div>
  );
};

export const AIHumanTaskBoard = AIHumanTaskAssignmentBoard;
export const LivePresence = LivePresenceLayer;
