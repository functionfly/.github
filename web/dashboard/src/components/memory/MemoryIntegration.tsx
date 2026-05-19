/**
 * Memory Integration Component
 * Unified panel that wires all memory components together
 */

import React, { useState, useCallback, useMemo } from 'react'
import { cn } from '@functionfly/ui-core'
import {
  Brain,
  MemoryStick,
  Network,
  Clock,
  TrendingUp,
  TrendingDown,
  Sparkles,
  Search,
  Filter,
  ChevronRight,
  ChevronDown,
  Eye,
  Layers,
  GitMerge,
  Share2,
  Users,
  Bot,
  Zap,
  Activity,
  BarChart3,
  PieChart,
  LineChart,
  Hexagon,
  Circle,
  Box,
  Database,
  Archive,
  RefreshCw,
  Trash2,
  Plus,
  Minus,
  X,
  Check,
  AlertTriangle,
  Info,
  ArrowRight,
  ArrowUpRight,
  ArrowDownRight,
  Link,
  Unlink,
  GripVertical,
  MoreHorizontal,
  FolderOpen,
  FileText,
  MessageSquare,
  User,
  Users as UsersIcon,
  Copy,
  History,
  Target,
  Compass,
  Radar,
  Cpu,
  Binary,
  Grid3x3,
  List,
  LayoutGrid,
  Mic2,
  Volume2,
  Headphones,
  Mic,
  Download,
} from 'lucide-react'

// Import memory components
import {
  MemoryGraph,
  SemanticMemoryViewer,
  LongTermContextExplorer,
  MemoryRecallTimeline,
  KnowledgeClusterMap,
  MemoryDecayVisualizer,
  VectorEmbeddingExplorer,
  SharedAgentMemoryPanel,
  MemoryMergeTool,
  ConversationMemoryTree,
  MemoryAccessMonitor,
} from '@functionfly/ui-memory'

// Panel navigation items
const NAV_ITEMS = [
  { id: 'graph', label: 'Memory Graph', icon: Network },
  { id: 'semantic', label: 'Semantic Memory', icon: Brain },
  { id: 'context', label: 'Long Term Context', icon: Archive },
  { id: 'recall', label: 'Recall Timeline', icon: History },
  { id: 'clusters', label: 'Knowledge Clusters', icon: Hexagon },
  { id: 'decay', label: 'Memory Decay', icon: TrendingDown },
  { id: 'vectors', label: 'Vector Embeddings', icon: Binary },
  { id: 'agents', label: 'Shared Agents', icon: Users },
  { id: 'merge', label: 'Memory Merge', icon: GitMerge },
  { id: 'conversation', label: 'Conversation Tree', icon: MessageSquare },
  { id: 'access', label: 'Access Monitor', icon: Radar },
] as const

type PanelId = typeof NAV_ITEMS[number]['id']

// Mock data generators
const generateMockGraphNodes = () => [
  { id: 'node-1', type: 'concept' as const, label: 'User Authentication', timestamp: Date.now() - 86400000, importance: 0.9 },
  { id: 'node-2', type: 'entity' as const, label: 'User Entity', timestamp: Date.now() - 172800000, importance: 0.8 },
  { id: 'node-3', type: 'event' as const, label: 'Login Event', timestamp: Date.now() - 3600000, importance: 0.7 },
  { id: 'node-4', type: 'document' as const, label: 'Auth Docs', timestamp: Date.now() - 604800000, importance: 0.6 },
  { id: 'node-5', type: 'code' as const, label: 'Auth Module', timestamp: Date.now() - 259200000, importance: 0.85 },
  { id: 'node-6', type: 'concept' as const, label: 'Session Management', timestamp: Date.now() - 432000000, importance: 0.75 },
  { id: 'node-7', type: 'agent' as const, label: 'Auth Agent', timestamp: Date.now() - 86400000, importance: 0.9 },
]

const generateMockGraphEdges = () => [
  { id: 'edge-1', source: 'node-1', target: 'node-2', type: 'related_to' as const },
  { id: 'edge-2', source: 'node-2', target: 'node-3', type: 'references' as const },
  { id: 'edge-3', source: 'node-1', target: 'node-5', type: 'part_of' as const },
  { id: 'edge-4', source: 'node-5', target: 'node-6', type: 'derives_from' as const },
  { id: 'edge-5', source: 'node-7', target: 'node-1', type: 'associated_with' as const },
]

const generateMockSemanticEntries = () => [
  { id: 'sem-1', content: 'User prefers dark mode interface', semanticType: 'preference' as const, confidence: 0.92, timestamp: Date.now() - 86400000, tags: ['ui', 'theme'] },
  { id: 'sem-2', content: 'Authentication flow requires MFA for admin users', semanticType: 'procedure' as const, confidence: 0.88, timestamp: Date.now() - 172800000, tags: ['auth', 'security'] },
  { id: 'sem-3', content: 'User frequently accesses billing dashboard', semanticType: 'context' as const, confidence: 0.85, timestamp: Date.now() - 43200000, accessCount: 45 },
  { id: 'sem-4', content: 'API rate limits are per-tenant not per-user', semanticType: 'fact' as const, confidence: 0.95, timestamp: Date.now() - 604800000 },
  { id: 'sem-5', content: 'Session tokens should refresh every 30 minutes', semanticType: 'procedure' as const, confidence: 0.91, timestamp: Date.now() - 259200000, tags: ['auth', 'security'] },
  { id: 'sem-6', content: 'User belongs to enterprise tier with SSO enabled', semanticType: 'fact' as const, confidence: 0.98, timestamp: Date.now() - 86400000, tags: ['enterprise', 'sso'] },
  { id: 'sem-7', content: 'Notification preferences link to user settings', semanticType: 'relationship' as const, confidence: 0.78, timestamp: Date.now() - 172800000 },
]

const generateMockContextChunks = () => [
  { id: 'ctx-1', content: 'Initial user onboarding completed with basic profile setup and email verification', timestamp: Date.now() - 2592000000, importance: 0.9, retentionPriority: 'high' as const, retrievalCount: 12 },
  { id: 'ctx-2', content: 'Enterprise contract negotiation discussions with pricing tier adjustments', timestamp: Date.now() - 1209600000, importance: 0.95, retentionPriority: 'critical' as const, retrievalCount: 8 },
  { id: 'ctx-3', content: 'Quarterly product feedback session summary and feature priorities', timestamp: Date.now() - 604800000, importance: 0.7, retentionPriority: 'medium' as const, retrievalCount: 5 },
  { id: 'ctx-4', content: 'Technical debt cleanup sprint - removed deprecated API endpoints', timestamp: Date.now() - 432000000, importance: 0.5, decayScore: 0.65, retrievalCount: 2 },
  { id: 'ctx-5', content: 'User research interview notes for next-gen dashboard design', timestamp: Date.now() - 86400000, importance: 0.8, retentionPriority: 'high' as const, retrievalCount: 6 },
]

const generateMockRecallEvents = () => [
  { id: 'recall-1', timestamp: Date.now() - 3600000, type: 'retrieval' as const, memoryId: 'sem-1', memoryLabel: 'User Preferences', strength: 0.92 },
  { id: 'recall-2', timestamp: Date.now() - 7200000, type: 'reinforcement' as const, memoryId: 'sem-2', memoryLabel: 'Auth Procedure', strength: 0.88 },
  { id: 'recall-3', timestamp: Date.now() - 10800000, type: 'decay' as const, memoryId: 'ctx-4', memoryLabel: 'Tech Debt Sprint', strength: 0.45 },
  { id: 'recall-4', timestamp: Date.now() - 14400000, type: 'consolidation' as const, memoryId: 'sem-6', memoryLabel: 'Enterprise Facts', strength: 0.95 },
  { id: 'recall-5', timestamp: Date.now() - 18000000, type: 'retrieval' as const, memoryId: 'node-3', memoryLabel: 'Login Event', strength: 0.78 },
]

const generateMockClusters = () => [
  { id: 'cluster-1', name: 'Authentication', centralTopic: 'User Auth', members: ['Auth Module', 'Login', 'MFA', 'SSO'], importance: 0.95, coherence: 0.92, creationTimestamp: Date.now() - 2592000000 },
  { id: 'cluster-2', name: 'Billing', centralTopic: 'Payments', members: ['Invoice', 'Subscription', 'Pricing'], importance: 0.88, coherence: 0.89, creationTimestamp: Date.now() - 1728000000 },
  { id: 'cluster-3', name: 'Dashboard', centralTopic: 'Analytics', members: ['Metrics', 'Charts', 'Reports'], importance: 0.75, coherence: 0.85, creationTimestamp: Date.now() - 864000000 },
  { id: 'cluster-4', name: 'Notifications', centralTopic: 'Alerts', members: ['Email', 'Slack', 'Webhooks'], importance: 0.65, coherence: 0.78, creationTimestamp: Date.now() - 432000000 },
]

const generateMockDecayNodes = () => [
  { id: 'decay-1', label: 'Old Session Data', age: 180, strength: 0.25, decayRate: 0.05, type: 'episodic' as const, nextDecay: Date.now() + 86400000 },
  { id: 'decay-2', label: 'Cached API Responses', age: 90, strength: 0.45, decayRate: 0.03, type: 'semantic' as const, lastReinforced: Date.now() - 86400000 },
  { id: 'decay-3', label: 'User Preferences', age: 30, strength: 0.85, decayRate: 0.01, type: 'semantic' as const, lastReinforced: Date.now() - 3600000 },
  { id: 'decay-4', label: 'Login Handlers', age: 7, strength: 0.92, decayRate: 0.005, type: 'procedural' as const, lastReinforced: Date.now() - 1800000 },
  { id: 'decay-5', label: 'Temp File References', age: 60, strength: 0.35, decayRate: 0.08, type: 'episodic' as const, nextDecay: Date.now() + 172800000 },
]

const generateMockVectors = () => [
  { id: 'vec-1', values: [0.12, 0.45, 0.78, 0.23], label: 'UserPreference', dimension: 4, magnitude: 0.95, source: 'embeddings/v1' },
  { id: 'vec-2', values: [0.34, 0.67, 0.12, 0.89], label: 'AuthContext', dimension: 4, magnitude: 0.88, source: 'embeddings/v1' },
  { id: 'vec-3', values: [0.56, 0.23, 0.91, 0.45], label: 'BillingData', dimension: 4, magnitude: 0.92, source: 'embeddings/v1' },
  { id: 'vec-4', values: [0.78, 0.91, 0.34, 0.67], label: 'DashboardMetrics', dimension: 4, magnitude: 0.85, source: 'embeddings/v1' },
  { id: 'vec-5', values: [0.23, 0.78, 0.56, 0.12], label: 'NotificationPrefs', dimension: 4, magnitude: 0.79, source: 'embeddings/v1' },
  { id: 'vec-6', values: [0.45, 0.12, 0.89, 0.34], label: 'SessionState', dimension: 4, magnitude: 0.91, source: 'embeddings/v1' },
]

const generateMockAgents = () => [
  { agentId: 'agent-1', agentName: 'Auth Agent', role: 'Security', memoryCapacity: 1000, usedCapacity: 750, activeMemories: ['auth-1', 'auth-2'], sharedWith: ['agent-2'], lastActive: Date.now() - 300000 },
  { agentId: 'agent-2', agentName: 'Billing Agent', role: 'Finance', memoryCapacity: 800, usedCapacity: 620, activeMemories: ['bill-1'], sharedWith: ['agent-1', 'agent-3'], lastActive: Date.now() - 600000 },
  { agentId: 'agent-3', agentName: 'Dashboard Agent', role: 'Analytics', memoryCapacity: 1200, usedCapacity: 890, activeMemories: ['dash-1', 'dash-2'], sharedWith: ['agent-2'], lastActive: Date.now() - 120000 },
  { agentId: 'agent-4', agentName: 'Notification Agent', role: 'Communications', memoryCapacity: 600, usedCapacity: 340, activeMemories: ['notif-1'], sharedWith: [], lastActive: Date.now() - 1800000 },
]

const generateMockMergeCandidates = () => [
  {
    id: 'merge-1',
    sourceMemoryIds: ['sem-1', 'sem-3'],
    suggestedMerge: { content: 'User preferences for UI theme and notification settings consolidated', confidence: 0.88 },
    overlapScore: 0.75,
  },
  {
    id: 'merge-2',
    sourceMemoryIds: ['node-3', 'node-5'],
    suggestedMerge: { content: 'Authentication flow and session management merged', confidence: 0.82 },
    overlapScore: 0.68,
    conflicts: [{ field: 'timestamp', values: [Date.now() - 86400000, Date.now() - 172800000] }],
  },
]

const generateMockConversationNodes = () => [
  { id: 'conv-1', type: 'topic' as const, content: 'User authentication workflow discussion', timestamp: Date.now() - 3600000, children: ['conv-2', 'conv-3'] },
  { id: 'conv-2', type: 'turn' as const, content: 'User: I need to implement SSO for our enterprise clients', speaker: 'user' as const, timestamp: Date.now() - 3500000, parentId: 'conv-1' },
  { id: 'conv-3', type: 'turn' as const, content: 'Agent: I can help you set up SAML-based SSO with your existing identity provider', speaker: 'agent' as const, timestamp: Date.now() - 3400000, parentId: 'conv-1', sentiment: 'positive' as const },
  { id: 'conv-4', type: 'branch' as const, content: 'SAML Configuration Branch', timestamp: Date.now() - 3300000, parentId: 'conv-3', children: ['conv-5'] },
  { id: 'conv-5', type: 'turn' as const, content: 'Agent: You will need to configure the following endpoints...', speaker: 'agent' as const, timestamp: Date.now() - 3200000, parentId: 'conv-4', intent: 'provide instructions' },
]

const generateMockAccessEntries = () => [
  { id: 'acc-1', memoryId: 'sem-1', memoryLabel: 'User Preferences', accessType: 'read' as const, timestamp: Date.now() - 300000, agentName: 'Auth Agent', duration: 12, success: true, cacheHit: true },
  { id: 'acc-2', memoryId: 'node-3', memoryLabel: 'Login Event', accessType: 'write' as const, timestamp: Date.now() - 600000, agentName: 'Dashboard Agent', duration: 25, success: true, cacheHit: false },
  { id: 'acc-3', memoryId: 'sem-6', memoryLabel: 'Enterprise Facts', accessType: 'read' as const, timestamp: Date.now() - 900000, agentName: 'Billing Agent', duration: 8, success: true, cacheHit: true },
  { id: 'acc-4', memoryId: 'ctx-2', memoryLabel: 'Contract Discussion', accessType: 'read' as const, timestamp: Date.now() - 1200000, agentName: 'Auth Agent', duration: 15, success: false, cacheHit: false },
  { id: 'acc-5', memoryId: 'vec-1', memoryLabel: 'UserPreference', accessType: 'write' as const, timestamp: Date.now() - 1500000, agentName: 'Dashboard Agent', duration: 32, success: true, cacheHit: false },
]

interface MemoryIntegrationProps {
  className?: string
}

export const MemoryIntegration: React.FC<MemoryIntegrationProps> = ({ className }) => {
  const [activePanel, setActivePanel] = useState<PanelId>('graph')
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)

  // Mock data states
  const [mockGraphNodes] = useState(generateMockGraphNodes)
  const [mockGraphEdges] = useState(generateMockGraphEdges)
  const [mockSemanticEntries] = useState(generateMockSemanticEntries)
  const [mockContextChunks] = useState(generateMockContextChunks)
  const [mockRecallEvents] = useState(generateMockRecallEvents)
  const [mockClusters] = useState(generateMockClusters)
  const [mockDecayNodes] = useState(generateMockDecayNodes)
  const [mockVectors] = useState(generateMockVectors)
  const [mockAgents] = useState(generateMockAgents)
  const [mockMergeCandidates] = useState(generateMockMergeCandidates)
  const [mockConversationNodes] = useState(generateMockConversationNodes)
  const [mockAccessEntries] = useState(generateMockAccessEntries)

  return (
    <div className={cn('flex h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Navigation Sidebar */}
      <div className={cn(
        'flex flex-col border-r border-aviation-border-panel transition-all duration-300',
        sidebarCollapsed ? 'w-12' : 'w-56'
      )}>
        {/* Collapse Toggle */}
        <div className="flex items-center justify-end px-2 py-2 border-b border-aviation-border-panel">
          <button
            onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
            className="p-1.5 hover:bg-aviation-bg-instrument rounded transition-colors"
          >
            {sidebarCollapsed ? <ChevronRight className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
          </button>
        </div>

        {/* Navigation Items */}
        <nav className="flex-1 overflow-auto py-2">
          {NAV_ITEMS.map((item) => {
            const Icon = item.icon
            const isActive = activePanel === item.id
            return (
              <button
                key={item.id}
                onClick={() => setActivePanel(item.id)}
                className={cn(
                  'flex items-center gap-3 w-full px-3 py-2 text-left transition-colors',
                  isActive ? 'bg-aviation-cyan/20 text-aviation-cyan border-l-2 border-aviation-cyan' : 'text-aviation-text-muted hover:text-aviation-text-primary hover:bg-aviation-bg-secondary',
                  sidebarCollapsed && 'justify-center px-0'
                )}
                title={sidebarCollapsed ? item.label : undefined}
              >
                <Icon className="w-4 h-4 flex-shrink-0" />
                {!sidebarCollapsed && <span className="text-sm truncate">{item.label}</span>}
              </button>
            )
          })}
        </nav>

        {/* Status Indicator */}
        {!sidebarCollapsed && (
          <div className="px-3 py-2 border-t border-aviation-border-panel">
            <div className="flex items-center gap-2 text-xs text-aviation-text-muted">
              <div className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
              <span>Memory Active</span>
            </div>
          </div>
        )}
      </div>

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-aviation-border-panel bg-aviation-bg-secondary">
          <div className="flex items-center gap-2">
            <Brain className="w-5 h-5 text-aviation-cyan" />
            <span className="text-sm font-medium">AI Memory Systems</span>
          </div>
          <div className="flex items-center gap-2 text-xs text-aviation-text-muted">
            <MemoryStick className="w-4 h-4" />
            <span>{NAV_ITEMS.find(i => i.id === activePanel)?.label}</span>
          </div>
        </div>

        {/* Content Panel */}
        <div className="flex-1 overflow-hidden">
          {activePanel === 'graph' && (
            <MemoryGraph
              nodes={mockGraphNodes}
              edges={mockGraphEdges}
              layout="force"
              className="h-full"
            />
          )}

          {activePanel === 'semantic' && (
            <SemanticMemoryViewer
              entries={mockSemanticEntries}
              filterType="all"
              className="h-full"
            />
          )}

          {activePanel === 'context' && (
            <LongTermContextExplorer
              chunks={mockContextChunks}
              focusArea="important"
              className="h-full"
            />
          )}

          {activePanel === 'recall' && (
            <MemoryRecallTimeline
              events={mockRecallEvents}
              className="h-full"
            />
          )}

          {activePanel === 'clusters' && (
            <KnowledgeClusterMap
              clusters={mockClusters}
              className="h-full"
            />
          )}

          {activePanel === 'decay' && (
            <MemoryDecayVisualizer
              nodes={mockDecayNodes}
              showPredictions
              className="h-full"
            />
          )}

          {activePanel === 'vectors' && (
            <VectorEmbeddingExplorer
              vectors={mockVectors}
              metric="cosine"
              className="h-full"
            />
          )}

          {activePanel === 'agents' && (
            <SharedAgentMemoryPanel
              agents={mockAgents}
              sharedMemoryIds={['auth-1', 'bill-1']}
              className="h-full"
            />
          )}

          {activePanel === 'merge' && (
            <MemoryMergeTool
              candidates={mockMergeCandidates}
              className="h-full"
            />
          )}

          {activePanel === 'conversation' && (
            <ConversationMemoryTree
              nodes={mockConversationNodes}
              focusedNodeId="conv-1"
              className="h-full"
            />
          )}

          {activePanel === 'access' && (
            <MemoryAccessMonitor
              entries={mockAccessEntries}
              className="h-full"
            />
          )}
        </div>
      </div>
    </div>
  )
}

export default MemoryIntegration
