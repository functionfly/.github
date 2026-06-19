/**
 * Code Intelligence Integration Component
 * Unified panel that wires all code intelligence components together
 */

import React, { useState, useCallback, useMemo } from 'react'
import { cn } from '@functionfly/ui-core'
import { useCodeIntelligenceStore } from '../../stores/codeIntelligenceStore'
import {
  Code,
  GitBranch,
  GitCommit,
  AlertTriangle,
  Shield,
  Zap,
  FileCode,
  Search,
  History,
  Activity,
  Network,
  Eye,
  Lightbulb,
  Sparkles,
  FileText,
  MessageSquare,
  ChevronRight,
  ChevronDown,
  FileDiff,
  Layers,
  Box,
  ArrowRight,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Info,
  Copy,
  Check,
  RefreshCw,
  Play,
  Pause,
  SkipForward,
  SkipBack,
  Target,
  Cpu,
  GitFork,
  User,
  Users,
  GitMerge,
  Bug,
  TrendingUp,
  Clock,
  Filter,
  SortAsc,
  EyeOff,
  SearchCode,
  FileSearch,
  TreePine,
  NetworkIcon,
  Layout as LayoutIcon,
  Hammer,
  Wand2,
  Bot,
  MessageCircle,
  Circle,
  Type,
  Hash,
  Text,
  List,
  Settings,
  Terminal,
  AlertOctagon,
  CheckSquare,
  XSquare,
  Plus,
  Minus,
  ArrowUpRight,
  ArrowDownRight,
  LayersIcon,
  PieChart,
  BarChart3,
  Radar,
  TerminalSquare,
  BugIcon,
  Crosshair,
  Map,
  Package,
  Component,
  CircleDot,
  Telescope,
  LineChart,
  Brain,
  Atom,
  Webhook,
  FileType,
  Braces,
  Scan,
  EyeIcon,
  Gauge,
  Microscope,
} from 'lucide-react'

// Panel navigation items
const NAV_ITEMS = [
  { id: 'editor', label: 'Code Editor', icon: Code },
  { id: 'ast', label: 'AST Explorer', icon: Braces },
  { id: 'dependencies', label: 'Dependencies', icon: GitFork },
  { id: 'diff', label: 'Diff Viewer', icon: FileDiff },
  { id: 'architecture', label: 'Architecture', icon: LayoutIcon },
  { id: 'refactor', label: 'Smart Refactor', icon: Wand2 },
  { id: 'generation', label: 'Code Generation', icon: Sparkles },
  { id: 'search', label: 'Semantic Search', icon: SearchCode },
  { id: 'lineage', label: 'Code Lineage', icon: History },
  { id: 'risk', label: 'Risk Analyzer', icon: Shield },
  { id: 'imports', label: 'Import Graph', icon: Network },
  { id: 'ownership', label: 'Ownership', icon: Users },
] as const

type PanelId = typeof NAV_ITEMS[number]['id']

// Mock data generators for demo purposes
const generateMockAST = () => ({
  id: 'root',
  type: 'Program',
  loc: { start: { line: 1, column: 1 }, end: { line: 50, column: 1 } },
  children: [
    {
      id: 'fn1',
      type: 'FunctionDeclaration',
      loc: { start: { line: 1, column: 1 }, end: { line: 20, column: 1 } },
      value: 'processData',
      children: [
        {
          id: 'param1',
          type: 'Identifier',
          loc: { start: { line: 1, column: 15 }, end: { line: 1, column: 20 } },
          value: 'input',
        },
        {
          id: 'block1',
          type: 'BlockStatement',
          loc: { start: { line: 2, column: 1 }, end: { line: 19, column: 1 } },
          children: [
            {
              id: 'ret1',
              type: 'ReturnStatement',
              loc: { start: { line: 18, column: 1 }, end: { line: 18, column: 30 } },
            },
          ],
        },
      ],
    },
    {
      id: 'cls1',
      type: 'ClassDeclaration',
      loc: { start: { line: 22, column: 1 }, end: { line: 45, column: 1 } },
      value: 'DataProcessor',
      children: [],
    },
  ],
})

const generateMockDependencies = () => {
  const nodes = [
    { id: 'pkg-react', name: 'react', path: 'node_modules/react', type: 'package' as const, metrics: { complexity: 0.3, coupling: 0.2, changes: 5 } },
    { id: 'pkg-lodash', name: 'lodash', path: 'node_modules/lodash', type: 'package' as const, metrics: { complexity: 0.5, coupling: 0.4, changes: 12 } },
    { id: 'file-utils', name: 'utils.ts', path: 'src/utils.ts', type: 'file' as const, metrics: { complexity: 0.6, coupling: 0.7, changes: 8 } },
    { id: 'file-api', name: 'api.ts', path: 'src/api.ts', type: 'file' as const, metrics: { complexity: 0.4, coupling: 0.5, changes: 3 } },
    { id: 'file-components', name: 'components', path: 'src/components', type: 'module' as const, metrics: { complexity: 0.8, coupling: 0.6, changes: 15 } },
  ]
  const edges = [
    { source: 'file-utils', target: 'pkg-lodash', type: 'import' as const },
    { source: 'file-api', target: 'file-utils', type: 'import' as const },
    { source: 'file-components', target: 'file-utils', type: 'import' as const },
    { source: 'file-components', target: 'pkg-react', type: 'import' as const },
  ]
  return { nodes, edges }
}

const generateMockArchitecture = () => {
  const nodes = [
    {
      id: 'layer-ui',
      name: 'UI Layer',
      type: 'layer' as const,
      metrics: { components: 15, linesOfCode: 2500, complexity: 0.6 },
      children: [
        { id: 'comp-button', name: 'Button', type: 'component' as const, metrics: { components: 1, linesOfCode: 150 } },
        { id: 'comp-input', name: 'Input', type: 'component' as const, metrics: { components: 1, linesOfCode: 200 } },
        { id: 'comp-modal', name: 'Modal', type: 'component' as const, metrics: { components: 1, linesOfCode: 300 } },
      ],
    },
    {
      id: 'layer-business',
      name: 'Business Logic',
      type: 'layer' as const,
      metrics: { components: 8, linesOfCode: 1800, complexity: 0.7 },
      children: [],
    },
    {
      id: 'layer-data',
      name: 'Data Layer',
      type: 'layer' as const,
      metrics: { components: 5, linesOfCode: 1200, complexity: 0.5 },
      children: [],
    },
  ]
  return { nodes, connections: [] }
}

const generateMockIntents = () => [
  {
    id: 'intent-1',
    type: 'feature' as const,
    confidence: 0.92,
    description: 'Add user authentication with JWT tokens',
    affectedCodeRanges: [{ start: { line: 1, column: 1 }, end: { line: 50, column: 1 } }],
    affectedFiles: ['src/auth/login.ts', 'src/auth/token.ts'],
    reasoning: 'The code analysis shows multiple login-related functions that could benefit from unified JWT-based authentication.',
    extractedRequirements: ['Support for refresh tokens', 'Token expiration handling', 'Secure storage'],
  },
  {
    id: 'intent-2',
    type: 'optimization' as const,
    confidence: 0.85,
    description: 'Optimize database queries for user lookup',
    affectedCodeRanges: [{ start: { line: 100, column: 1 }, end: { line: 150, column: 1 } }],
    affectedFiles: ['src/db/users.ts'],
    reasoning: 'Several queries are fetching full user records when only specific fields are needed.',
  },
  {
    id: 'intent-3',
    type: 'security' as const,
    confidence: 0.78,
    description: 'Add input validation to prevent SQL injection',
    affectedCodeRanges: [{ start: { line: 75, column: 1 }, end: { line: 90, column: 50 } }],
    affectedFiles: ['src/db/queries.ts'],
    reasoning: 'Raw SQL queries detected without parameterized statements.',
  },
]

const generateMockSearchResults = () => [
  {
    id: 'result-1',
    filePath: 'src/components/Button.tsx',
    lineNumber: 42,
    lineContent: 'const handleClick = (event: MouseEvent) => {',
    matchedText: 'handleClick',
    context: ['const handleClick = (event: MouseEvent) => {', '  onClick(event);', '};'],
    score: 0.95,
    matchType: 'semantic' as const,
  },
  {
    id: 'result-2',
    filePath: 'src/utils/handlers.ts',
    lineNumber: 18,
    lineContent: 'export function handleClick() {',
    matchedText: 'handleClick',
    context: ['export function handleClick() {', '  return true;', '}'],
    score: 0.88,
    matchType: 'fuzzy' as const,
  },
]

const generateMockLineage = () => [
  {
    id: 'commit-1',
    type: 'commit' as const,
    name: 'feat: add user authentication',
    author: 'Sarah Chen',
    timestamp: Date.now() - 86400000,
    message: 'Implement JWT-based authentication flow',
    metadata: { filesChanged: 5, insertions: 150, deletions: 20, branch: 'main', tags: ['v2.0'] },
  },
  {
    id: 'commit-2',
    type: 'commit' as const,
    name: 'fix: resolve race condition',
    author: 'Mike Johnson',
    timestamp: Date.now() - 172800000,
    message: 'Fix async state update race condition',
    metadata: { filesChanged: 2, insertions: 15, deletions: 10, branch: 'main' },
  },
  {
    id: 'merge-1',
    type: 'merge' as const,
    name: 'Merge pull request #42',
    author: 'Alex Rivera',
    timestamp: Date.now() - 259200000,
    message: 'Merge feature/login-redesign into main',
    metadata: { filesChanged: 12, insertions: 500, deletions: 200, branch: 'main' },
  },
]

const generateMockRisks = () => [
  {
    id: 'risk-1',
    type: 'security' as const,
    severity: 'critical' as const,
    message: 'Potential SQL injection vulnerability',
    location: { start: { line: 42, column: 1 }, end: { line: 42, column: 60 } },
    file: 'src/db/queries.ts',
    code: 'query = "SELECT * FROM users WHERE id = " + userId',
    suggestion: 'Use parameterized queries instead of string concatenation',
    cwe: '89',
  },
  {
    id: 'risk-2',
    type: 'performance' as const,
    severity: 'high' as const,
    message: 'Memory leak in event listener cleanup',
    location: { start: { line: 88, column: 1 }, end: { line: 95, column: 20 } },
    file: 'src/hooks/useEventListener.ts',
    suggestion: 'Ensure removeEventListener is called on cleanup',
  },
  {
    id: 'risk-3',
    type: 'maintainability' as const,
    severity: 'medium' as const,
    message: 'High cyclomatic complexity detected',
    location: { start: { line: 10, column: 1 }, end: { line: 50, column: 10 } },
    file: 'src/utils/processData.ts',
    code: 'function processData(input: any) { ... }',
    suggestion: 'Consider breaking into smaller, focused functions',
  },
]

const generateMockCompletions = () => [
  {
    id: 'comp-1',
    text: `const handleSubmit = async (formData: FormData) => {
  try {
    const response = await api.post('/submit', formData);
    setState({ status: 'success', data: response.data });
  } catch (error) {
    setState({ status: 'error', message: error.message });
  }
};`,
    type: 'completion' as const,
    confidence: 0.94,
    model: 'claude-3-5-sonnet',
    timestamp: Date.now(),
    latency: 450,
    tokens: 125,
    context: { cursorPosition: { line: 10, column: 20 }, filePath: 'src/components/Form.tsx', language: 'typescript' },
  },
  {
    id: 'comp-2',
    text: `const handleSubmit = async (formData: FormData) => {
  return api.submit(formData);
};`,
    type: 'completion' as const,
    confidence: 0.72,
    model: 'claude-3-haiku',
    timestamp: Date.now(),
    latency: 180,
    tokens: 45,
    context: { cursorPosition: { line: 10, column: 20 }, filePath: 'src/components/Form.tsx', language: 'typescript' },
    alternatives: [
      { text: 'Full async/await with error handling', confidence: 0.94, model: 'claude-3-5-sonnet' },
    ],
  },
]

const generateMockSimulation = () => ({
  id: 'sim-1',
  name: 'Extract Utility Functions',
  description: 'Move shared utility functions to a separate module',
  changes: [
    {
      fileId: 'file-1',
      filePath: 'src/utils/helpers.ts',
      changeType: 'modify' as const,
      before: 'function formatDate(date) { return date.toISOString(); }',
      after: 'export function formatDate(date: Date): string { return date.toISOString(); }',
      hunks: [],
    },
    {
      fileId: 'file-2',
      filePath: 'src/components/Card.tsx',
      changeType: 'modify' as const,
      before: 'function formatDate(date) { return date.toISOString(); }',
      after: "import { formatDate } from '../utils/helpers';",
      hunks: [],
    },
  ],
  impactAnalysis: {
    estimatedTime: 15,
    riskLevel: 'low' as const,
    affectedComponents: ['Card', 'Form', 'Modal'],
    testCoverageImpact: -2,
  },
  validationResults: [
    { type: 'compile' as const, passed: true },
    { type: 'test' as const, passed: true },
    { type: 'lint' as const, passed: true },
  ],
})

const generateMockConstraints = () => [
  {
    id: 'constraint-1',
    type: 'layering' as const,
    name: 'UI Cannot Access Data Directly',
    description: 'UI components must not import from data layer modules',
    severity: 'error' as const,
    enforcement: 'strict' as const,
    violatedBy: [
      { file: 'src/components/UserCard.tsx', line: 5, details: 'Directly imports from src/db/' },
    ],
    fixSuggestion: 'Move database access to a service layer and use dependency injection',
  },
  {
    id: 'constraint-2',
    type: 'naming' as const,
    name: 'Component Prefix Required',
    description: 'All React components must be prefixed with their type',
    severity: 'warning' as const,
    enforcement: 'advisory' as const,
    violatedBy: [
      { file: 'src/components/card.tsx', line: 1 },
    ],
  },
  {
    id: 'constraint-3',
    type: 'dependency' as const,
    name: 'No Circular Dependencies',
    description: 'Modules should not have circular import dependencies',
    severity: 'error' as const,
    enforcement: 'strict' as const,
    violatedBy: [
      { file: 'src/utils/a.ts', line: 3, details: 'Imports from src/utils/b.ts' },
      { file: 'src/utils/b.ts', line: 2, details: 'Imports from src/utils/a.ts' },
    ],
  },
]

const generateMockOwnerships = () => [
  {
    filePath: 'src/components/Button.tsx',
    owners: [
      { id: 'owner-1', name: 'Sarah Chen', email: 'sarah@company.com', gitHubUsername: 'sarahchen' },
    ],
    lastModified: Date.now() - 86400000,
    lastModifiedBy: 'sarahchen',
    reviewRequired: true,
    autoAssignment: false,
  },
  {
    filePath: 'src/utils/helpers.ts',
    owners: [
      { id: 'owner-2', name: 'Mike Johnson', email: 'mike@company.com', gitHubUsername: 'mikej' },
      { id: 'owner-3', name: 'Alex Rivera', email: 'alex@company.com', gitHubUsername: 'alexr' },
    ],
    lastModified: Date.now() - 172800000,
    reviewRequired: false,
    autoAssignment: true,
  },
  {
    filePath: 'src/api/client.ts',
    owners: [
      { id: 'owner-1', name: 'Sarah Chen', email: 'sarah@company.com' },
    ],
    lastModified: Date.now() - 604800000,
    reviewRequired: true,
    autoAssignment: false,
  },
]

// Import components from the code intelligence package
import {
  SemanticCodeEditor,
  ASTExplorer,
  DependencyHeatmap,
  MultiFileDiffViewer,
  ArchitectureMap,
  SmartRefactorPanel,
  CodeGenerationPreview,
  InlineAIAssistant,
  CodeIntentExplorer,
  SemanticSearchPanel,
  CodeLineageViewer,
  CodeRiskAnalyzer,
  ImportGraphViewer,
  ExecutionAwareEditor,
  AICompletionInspector,
  RefactorSimulationViewer,
  ArchitectureConstraintPanel,
  CodeOwnershipMap,
} from '@functionfly/ui-code-intelligence'

interface CodeIntelligenceIntegrationProps {
  className?: string
}

export const CodeIntelligenceIntegration: React.FC<CodeIntelligenceIntegrationProps> = ({
  className,
}) => {
  const [activePanel, setActivePanel] = useState<PanelId>('editor')
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)

  // Mock data states
  const [editorValue] = useState(`// Code Intelligence Demo
import React, { useState, useCallback } from 'react';
import { useAuth } from './hooks/useAuth';

interface User {
  id: string;
  name: string;
  email: string;
}

export const UserProfile: React.FC = () => {
  const { user, logout } = useAuth();
  const [isLoading, setIsLoading] = useState(false);

  const handleUpdate = useCallback(async (data: Partial<User>) => {
    setIsLoading(true);
    try {
      await api.updateUser(user.id, data);
    } catch (error) {
      console.error('Update failed:', error);
    } finally {
      setIsLoading(false);
    }
  }, [user.id]);

  return (
    <div className="profile">
      <h1>Welcome, {user.name}</h1>
      <button onClick={logout}>Logout</button>
    </div>
  );
};
`)

  const [diffFiles] = useState([
    {
      id: 'file-1',
      path: 'src/components/Button.tsx',
      oldContent: `const Button = ({ label, onClick }) => (
  <button onClick={onClick}>{label}</button>
);`,
      newContent: `const Button = ({ label, onClick, disabled = false }) => (
  <button onClick={onClick} disabled={disabled}>{label}</button>
);`,
      status: 'modified' as const,
      language: 'typescript',
    },
    {
      id: 'file-2',
      path: 'src/utils/helpers.ts',
      oldContent: '',
      newContent: `export const formatDate = (date: Date): string => date.toISOString();`,
      status: 'added' as const,
      language: 'typescript',
    },
  ])

  const [inlineSuggestions] = useState([
    {
      id: 'sug-1',
      type: 'completion' as const,
      text: 'const handleSubmit = async (e: Event) => { e.preventDefault(); }',
      confidence: 0.92,
      startPosition: { line: 15, column: 1 },
      endPosition: { line: 15, column: 50 },
      explanation: 'Completes the form submit handler based on common patterns',
    },
    {
      id: 'sug-2',
      type: 'refactor' as const,
      text: 'useCallback(() => { ... }, [dependency])',
      confidence: 0.88,
      startPosition: { line: 10, column: 5 },
      endPosition: { line: 10, column: 40 },
      explanation: 'Consider wrapping in useCallback for better performance',
    },
  ])

  const [watchExpressions] = useState([
    { id: 'watch-1', expression: 'user.name', value: '"Sarah Chen"' },
    { id: 'watch-2', expression: 'isLoading', value: 'false' },
  ])

  // Initialize mock data
  const [mockAST] = useState(generateMockAST)
  const [mockDependencies] = useState(generateMockDependencies)
  const [mockArchitecture] = useState(generateMockArchitecture)
  const [mockIntents] = useState(generateMockIntents)
  const [mockSearchResults] = useState(generateMockSearchResults)
  const [mockLineage] = useState(generateMockLineage)
  const [mockRisks] = useState(generateMockRisks)
  const [mockCompletions] = useState(generateMockCompletions)
  const [mockSimulation] = useState(generateMockSimulation)
  const [mockConstraints] = useState(generateMockConstraints)
  const [mockOwnerships] = useState(generateMockOwnerships)

  // Mock imports data
  const [mockImports] = useState([
    { id: 'imp-1', name: 'useState', type: 'named' as const, source: 'react', line: 1 },
    { id: 'imp-2', name: 'useCallback', type: 'named' as const, source: 'react', line: 2 },
    { id: 'imp-3', name: 'useAuth', type: 'named' as const, source: './hooks/useAuth', line: 3 },
  ])
  const [mockImportEdges] = useState([
    { source: 'imp-1', target: 'react', type: 'import' as const },
    { source: 'imp-2', target: 'react', type: 'import' as const },
    { source: 'imp-3', target: 'useAuth', type: 'import' as const },
  ])

  // Mock refactor opportunities
  const [refactorOpportunities] = useState([
    {
      id: 'ref-1',
      type: 'extract-method',
      title: 'Extract API call to separate function',
      description: 'The handleUpdate function contains API logic that could be extracted',
      location: { start: { line: 12, column: 1 }, end: { line: 20, column: 50 } },
      original: 'const handleUpdate = useCallback(async (data) => { ... }',
      preview: 'const handleUpdate = useCallback(async (data) => {\n  await updateUserProfile(user.id, data);\n}, [user.id]);',
      impact: 'medium' as const,
      automated: true,
      affectedFiles: ['UserProfile.tsx'],
    },
    {
      id: 'ref-2',
      type: 'rename',
      title: 'Rename isLoading to isSubmitting',
      description: 'More descriptive name for the loading state',
      location: { start: { line: 10, column: 1 }, end: { line: 10, column: 45 } },
      original: 'const [isLoading, setIsLoading] = useState(false);',
      preview: 'const [isSubmitting, setIsSubmitting] = useState(false);',
      impact: 'low' as const,
      automated: true,
    },
  ])

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
              <span>AI Ready</span>
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
            <span className="text-sm font-medium">Code Intelligence</span>
          </div>
          <div className="flex items-center gap-2 text-xs text-aviation-text-muted">
            <Microscope className="w-4 h-4" />
            <span>{NAV_ITEMS.find(i => i.id === activePanel)?.label}</span>
          </div>
        </div>

        {/* Content Panel */}
        <div className="flex-1 overflow-hidden">
          {activePanel === 'editor' && (
            <SemanticCodeEditor
              value={editorValue}
              language="typescript"
              showLineNumbers
              showSemanticHighlights
              showCodeLenses
              className="h-full"
            />
          )}

          {activePanel === 'ast' && (
            <ASTExplorer
              ast={mockAST}
              showRange
              showMetadata
              className="h-full"
            />
          )}

          {activePanel === 'dependencies' && (
            <DependencyHeatmap
              nodes={mockDependencies.nodes}
              edges={mockDependencies.edges}
              colorMetric="complexity"
              className="h-full"
            />
          )}

          {activePanel === 'diff' && (
            <MultiFileDiffViewer
              files={diffFiles}
              showLineNumbers
              className="h-full"
            />
          )}

          {activePanel === 'architecture' && (
            <ArchitectureMap
              nodes={mockArchitecture.nodes}
              connections={mockArchitecture.connections}
              showMetrics
              className="h-full"
            />
          )}

          {activePanel === 'refactor' && (
            <SmartRefactorPanel
              opportunities={refactorOpportunities}
              className="h-full"
            />
          )}

          {activePanel === 'generation' && (
            <CodeGenerationPreview
              generation={{
                id: 'gen-1',
                language: 'typescript',
                code: `interface UserProfileProps {
  user: User;
  onUpdate: (data: Partial<User>) => Promise<void>;
  isLoading?: boolean;
}

export const UserProfile: React.FC<UserProfileProps> = ({
  user,
  onUpdate,
  isLoading = false,
}) => {
  const handleSubmit = async (data: Partial<User>) => {
    await onUpdate(data);
  };

  return (
    <div className="profile">
      <h1>{user.name}</h1>
      <ProfileForm initialData={user} onSubmit={handleSubmit} />
    </div>
  );
};`,
                title: 'UserProfile Component',
                description: 'Generated a typed React component for user profile management with form handling',
                context: {
                  originalCode: '// Original code...',
                  framework: 'React',
                  requirements: ['Type safety', 'Form handling', 'Loading states'],
                },
                metrics: {
                  complexity: 0.35,
                  maintainability: 0.92,
                  estimatedTokens: 850,
                },
                dependencies: [
                  { name: 'react', version: '^18.2.0' },
                ],
              }}
              showExplanation
              onCopy={(code) => console.log('Copy:', code)}
              onApply={() => console.log('Apply')}
              onDiscard={() => console.log('Discard')}
              onRegenerate={() => console.log('Regenerate')}
              onToggleExplanation={() => {}}
              className="h-full"
            />
          )}

          {activePanel === 'search' && (
            <SemanticSearchPanel
              query="handleClick"
              results={mockSearchResults}
              searchType="semantic"
              className="h-full"
            />
          )}

          {activePanel === 'lineage' && (
            <CodeLineageViewer
              nodes={mockLineage}
              className="h-full"
            />
          )}

          {activePanel === 'risk' && (
            <CodeRiskAnalyzer
              risks={mockRisks}
              showMetrics
              className="h-full"
            />
          )}

          {activePanel === 'imports' && (
            <ImportGraphViewer
              imports={mockImports}
              edges={mockImportEdges}
              filePath="src/components/UserProfile.tsx"
              className="h-full"
            />
          )}

          {activePanel === 'ownership' && (
            <CodeOwnershipMap
              ownerships={mockOwnerships}
              className="h-full"
            />
          )}
        </div>
      </div>

      {/* Inline AI Assistant - always visible */}
      <div className="w-80 flex flex-col border-l border-aviation-border-panel">
        <div className="px-3 py-2 border-b border-aviation-border-panel">
          <div className="flex items-center gap-2">
            <Bot className="w-4 h-4 text-aviation-cyan" />
            <span className="text-xs font-medium">Inline AI</span>
          </div>
        </div>
        <div className="flex-1 overflow-auto p-3">
          <InlineAIAssistant
            suggestions={inlineSuggestions}
            enabled
            onSuggestionAccept={(s) => console.log('Accept:', s)}
            onSuggestionReject={(s) => console.log('Reject:', s)}
            className="w-full"
          />
        </div>

        {/* Additional panels */}
        <div className="flex flex-col border-t border-aviation-border-panel">
          <ExecutionAwareEditor
            code={editorValue}
            language="typescript"
            executionPoints={[]}
            breakpoints={['12', '15']}
            watchExpressions={watchExpressions}
            isRunning={false}
            className="h-64"
          />

          <AICompletionInspector
            completions={mockCompletions}
            currentCompletion={mockCompletions[0]}
            className="h-64"
          />

          <RefactorSimulationViewer
            simulation={mockSimulation}
            className="h-64"
          />

          <ArchitectureConstraintPanel
            constraints={mockConstraints}
            className="h-64"
          />
        </div>
      </div>
    </div>
  )
}

export default CodeIntelligenceIntegration
