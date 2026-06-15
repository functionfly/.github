/**
 * Code Intelligence Store
 * Global state management for code intelligence components
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

export interface SemanticSymbol {
  id: string
  name: string
  kind: 'function' | 'class' | 'interface' | 'type' | 'variable' | 'constant' | 'enum' | 'namespace' | 'method' | 'property'
  location: CodeRange
  scope: string
  signature?: string
  documentation?: string
}

export interface Diagnostic {
  id: string
  severity: 'error' | 'warning' | 'information' | 'hint'
  message: string
  range: CodeRange
  code?: string | number
  source?: string
}

export interface ASTNode {
  id: string
  type: string
  loc: CodeRange
  children?: ASTNode[]
  value?: unknown
  metadata?: Record<string, unknown>
}

export interface DependencyNode {
  id: string
  name: string
  path: string
  version?: string
  type: 'package' | 'file' | 'module'
  size?: number
  metrics?: {
    complexity?: number
    coupling?: number
    changes?: number
    bugs?: number
    maintenance?: number
  }
}

export interface DependencyEdge {
  source: string
  target: string
  weight?: number
  type: 'import' | 'inheritance' | 'composition' | 'dependency'
}

export interface DiffFile {
  id: string
  path: string
  oldContent: string
  newContent: string
  language?: string
  status?: 'added' | 'deleted' | 'modified' | 'renamed'
}

export interface ArchitectureNode {
  id: string
  name: string
  type: 'module' | 'package' | 'component' | 'service' | 'layer'
  path?: string
  children?: ArchitectureNode[]
  metrics?: {
    components?: number
    linesOfCode?: number
    complexity?: number
    stability?: number
  }
}

export interface ArchitectureConnection {
  id: string
  source: string
  target: string
  type: 'depends-on' | 'composes' | 'extends' | 'implements' | 'uses'
  weight?: number
}

export interface RefactorOpportunity {
  id: string
  type: string
  title: string
  description: string
  location: CodeRange
  original: string
  preview: string
  impact: 'low' | 'medium' | 'high'
  estimatedComplexity?: number
  affectedFiles?: string[]
  automated?: boolean
}

export interface GeneratedCode {
  id: string
  language: string
  code: string
  title?: string
  description?: string
  context?: {
    originalCode?: string
    language?: string
    framework?: string
    requirements?: string[]
  }
  metrics?: {
    complexity?: number
    maintainability?: number
    testability?: number
    estimatedTokens?: number
  }
  dependencies?: Array<{ name: string; version: string }>
}

export interface AIInlineSuggestion {
  id: string
  type: 'completion' | 'refactor' | 'documentation' | 'test' | 'explanation'
  text: string
  confidence: number
  startPosition: CodePosition
  endPosition: CodePosition
  explanation?: string
}

export interface CodeIntent {
  id: string
  type: 'feature' | 'bugfix' | 'refactor' | 'optimization' | 'documentation' | 'test' | 'security' | 'compliance'
  confidence: number
  description: string
  affectedCodeRanges: CodeRange[]
  affectedFiles: string[]
  reasoning: string
  relatedIntents?: string[]
  extractedRequirements?: string[]
}

export interface SearchResult {
  id: string
  filePath: string
  lineNumber: number
  lineContent: string
  matchedText: string
  context: string[]
  score: number
  matchType: 'exact' | 'fuzzy' | 'semantic'
}

export interface LineageNode {
  id: string
  type: 'commit' | 'change' | 'merge' | 'branch'
  name: string
  author: string
  timestamp: number
  message?: string
  parent?: string
  children?: string[]
  metadata?: {
    filesChanged?: number
    insertions?: number
    deletions?: number
    branch?: string
    tags?: string[]
  }
}

export interface RiskIndicator {
  id: string
  type: 'security' | 'performance' | 'maintainability' | 'testability' | 'complexity' | 'duplication'
  severity: 'critical' | 'high' | 'medium' | 'low' | 'info'
  message: string
  location?: CodeRange
  file?: string
  code?: string
  suggestion?: string
  cwe?: string
}

export interface ImportNode {
  id: string
  name: string
  type: 'default' | 'named' | 'namespace' | 'side-effect'
  source: string
  isReExported?: boolean
  line?: number
}

export interface ImportEdge {
  source: string
  target: string
  type: 'import' | 're-export' | 'type-import'
}

export interface ExecutionPoint {
  id: string
  timestamp: number
  line: number
  column: number
  type: 'breakpoint' | 'current' | 'watch' | 'function-call' | 'return'
  callStack?: string[]
  variableState?: Record<string, unknown>
  hitCount?: number
  condition?: string
}

export interface WatchExpression {
  id: string
  expression: string
  value?: unknown
  type?: string
  error?: string
}

export interface AICompletion {
  id: string
  text: string
  type: 'completion' | 'refactor' | 'explanation' | 'test'
  confidence: number
  model?: string
  timestamp: number
  latency?: number
  tokens?: number
  context?: {
    cursorPosition?: CodePosition
    selectedText?: string
    filePath?: string
    language?: string
  }
  alternatives?: Array<{ text: string; confidence: number; model?: string }>
}

export interface SimulationChange {
  fileId: string
  filePath: string
  changeType: 'add' | 'modify' | 'delete'
  before: string
  after: string
}

export interface RefactorSimulation {
  id: string
  name: string
  description: string
  changes: SimulationChange[]
  impactAnalysis?: {
    estimatedTime?: number
    riskLevel?: 'low' | 'medium' | 'high'
    affectedComponents?: string[]
    testCoverageImpact?: number
  }
  validationResults?: Array<{
    type: 'compile' | 'test' | 'lint'
    passed: boolean
    message?: string
  }>
}

export interface ArchitectureConstraint {
  id: string
  type: 'naming' | 'layering' | 'dependency' | 'visibility' | 'pattern'
  name: string
  description: string
  severity: 'error' | 'warning' | 'info'
  enforcement: 'strict' | 'advisory'
  violatedBy?: Array<{
    file: string
    line?: number
    details?: string
  }>
  fixSuggestion?: string
}

export interface CodeOwner {
  id: string
  name: string
  email: string
  avatar?: string
  gitHubUsername?: string
}

export interface FileOwnership {
  filePath: string
  owners: CodeOwner[]
  lastModified?: number
  lastModifiedBy?: string
  reviewRequired?: boolean
  autoAssignment?: boolean
}

export interface CodeEditorState {
  value: string
  language: string
  filePath?: string
  cursorPosition: CodePosition
  selection: CodeRange | null
  symbols: SemanticSymbol[]
  tokens: Array<{
    type: string
    value: string
    range: CodeRange
  }>
  diagnostics: Diagnostic[]
}

export interface CodeIntelligenceState {
  // Editor state
  editor: CodeEditorState

  // AST
  ast: ASTNode | null
  selectedASTNodeId: string | null
  expandedASTNodes: string[]
  astSearchQuery: string

  // Dependencies
  dependencies: {
    nodes: DependencyNode[]
    edges: DependencyEdge[]
  }
  selectedDependencyNodeId: string | null
  dependencyColorMetric: 'complexity' | 'coupling' | 'changes' | 'bugs' | 'maintenance'

  // Diff
  diffFiles: DiffFile[]
  selectedDiffFileId: string | null

  // Architecture
  architecture: {
    nodes: ArchitectureNode[]
    connections: ArchitectureConnection[]
  }
  selectedArchitectureNodeId: string | null
  expandedArchitectureNodes: string[]
  architectureLayout: 'tree' | 'force' | 'circular'
  showArchitectureMetrics: boolean

  // Refactor
  refactorOpportunities: RefactorOpportunity[]
  selectedRefactorOpportunityId: string | null

  // Code Generation
  generatedCode: GeneratedCode | null
  isGenerating: boolean
  generationError: string | null
  showGenerationExplanation: boolean

  // Inline AI
  inlineSuggestions: AIInlineSuggestion[]
  currentInlineSuggestion: AIInlineSuggestion | null
  inlineAIEnabled: boolean

  // Intent Explorer
  intents: CodeIntent[]
  selectedIntentId: string | null
  showIntentReasoning: boolean

  // Semantic Search
  searchQuery: string
  searchResults: SearchResult[]
  selectedSearchResultId: string | null
  searchType: 'text' | 'semantic' | 'symbol' | 'regex'
  isSearching: boolean

  // Lineage
  lineageNodes: LineageNode[]
  selectedLineageNodeId: string | null
  focusedFilePath: string | null

  // Risk Analysis
  riskIndicators: RiskIndicator[]
  selectedRiskId: string | null
  showRiskMetrics: boolean

  // Import Graph
  imports: ImportNode[]
  importEdges: ImportEdge[]
  selectedImportNodeId: string | null
  selectedImportFilePath: string | null

  // Execution Aware Editor
  executionPoints: ExecutionPoint[]
  currentExecutionPointId: string | null
  breakpoints: string[]
  watchExpressions: WatchExpression[]
  isRunning: boolean

  // AI Completion Inspector
  completions: AICompletion[]
  selectedCompletionId: string | null
  currentCompletion: AICompletion | null

  // Refactor Simulation
  simulation: RefactorSimulation | null
  simulationStep: number

  // Architecture Constraints
  constraints: ArchitectureConstraint[]
  selectedConstraintId: string | null

  // Code Ownership
  ownerships: FileOwnership[]
  selectedFilePath: string | null
  selectedOwnerId: string | null

  // UI State
  activePanel: 'editor' | 'ast' | 'dependencies' | 'diff' | 'architecture' | 'refactor' | 'generation' | 'search' | 'lineage' | 'risk' | 'imports' | 'ownership'
  sidebarCollapsed: boolean

  // Actions (defined below in the immer setup; exposed on the type for hook access)
  setEditorValue: (value: string) => void
  setEditorLanguage: (language: string) => void
  setEditorFilePath: (filePath: string) => void
  setEditorCursorPosition: (position: { line: number; column: number }) => void
  setEditorSelection: (selection: { start: { line: number; column: number }; end: { line: number; column: number } } | null) => void
  setEditorSymbols: (symbols: unknown[]) => void
  setEditorDiagnostics: (diagnostics: unknown[]) => void
  setAST: (ast: unknown) => void
  selectASTNode: (nodeId: string | null) => void
  toggleASTNode: (nodeId: string) => void
  setASTSearchQuery: (query: string) => void
  setDependencies: (nodes: unknown[], edges: unknown[]) => void
  selectDependencyNode: (nodeId: string | null) => void
  setDependencyColorMetric: (metric: 'complexity' | 'coupling' | 'changes' | 'bugs' | 'maintenance') => void
  setDiffFiles: (files: unknown[]) => void
  selectDiffFile: (fileId: string | null) => void
  setArchitecture: (nodes: unknown[], connections: unknown[]) => void
  selectArchitectureNode: (nodeId: string | null) => void
  toggleArchitectureNode: (nodeId: string) => void
  setArchitectureLayout: (layout: 'tree' | 'force' | 'circular') => void
  setShowArchitectureMetrics: (show: boolean) => void
  setRefactorOpportunities: (opportunities: unknown[]) => void
  selectRefactorOpportunity: (opportunityId: string | null) => void
  setGeneratedCode: (code: unknown) => void
  setIsGenerating: (isGenerating: boolean) => void
  setGenerationError: (error: string | null) => void
  toggleGenerationExplanation: () => void
  clearGeneratedCode: () => void
  setInlineSuggestions: (suggestions: unknown[]) => void
  setCurrentInlineSuggestion: (suggestion: unknown) => void
  toggleInlineAI: () => void
  setIntents: (intents: unknown[]) => void
  selectIntent: (intentId: string | null) => void
  toggleIntentReasoning: () => void
  setSearchQuery: (query: string) => void
  setSearchResults: (results: unknown[]) => void
  selectSearchResult: (resultId: string | null) => void
  setSearchType: (type: 'text' | 'semantic' | 'symbol' | 'regex') => void
  setIsSearching: (isSearching: boolean) => void
  clearSearchResults: () => void
  setLineageNodes: (nodes: unknown[]) => void
  selectLineageNode: (nodeId: string | null) => void
  setFocusedFilePath: (filePath: string | null) => void
  setRiskIndicators: (risks: unknown[]) => void
  selectRisk: (riskId: string | null) => void
  setShowRiskMetrics: (show: boolean) => void
  setImports: (imports: unknown[], edges: unknown[]) => void
  selectImportNode: (nodeId: string | null) => void
  setSelectedImportFilePath: (filePath: string | null) => void
  setExecutionPoints: (points: unknown[]) => void
  selectExecutionPoint: (pointId: string | null) => void
  toggleBreakpoint: (line: number) => void
  addWatchExpression: (expression: string) => void
  removeWatchExpression: (expressionId: string) => void
  setIsRunning: (isRunning: boolean) => void
  setCompletions: (completions: unknown[]) => void
  selectCompletion: (completionId: string | null) => void
  setSimulation: (simulation: unknown) => void
  stepSimulationForward: () => void
  stepSimulationBackward: () => void
  clearSimulation: () => void
  setConstraints: (constraints: unknown[]) => void
  selectConstraint: (constraintId: string | null) => void
  setOwnerships: (ownerships: unknown[]) => void
  selectFileOwnership: (filePath: string | null) => void
  selectOwner: (ownerId: string | null) => void
  setActivePanel: (panel: 'editor' | 'ast' | 'dependencies' | 'diff' | 'architecture' | 'refactor' | 'generation' | 'search' | 'lineage' | 'risk' | 'imports' | 'ownership') => void
  toggleSidebar: () => void
}

// ============================================================================
// Store
// ============================================================================

export const useCodeIntelligenceStore = create<CodeIntelligenceState>()(
  immer((set) => ({
    // Editor state
    editor: {
      value: '',
      language: 'typescript',
      filePath: undefined,
      cursorPosition: { line: 1, column: 1 },
      selection: null,
      symbols: [],
      tokens: [],
      diagnostics: [],
    },

    // AST
    ast: null,
    selectedASTNodeId: null,
    expandedASTNodes: [],
    astSearchQuery: '',

    // Dependencies
    dependencies: {
      nodes: [],
      edges: [],
    },
    selectedDependencyNodeId: null,
    dependencyColorMetric: 'complexity',

    // Diff
    diffFiles: [],
    selectedDiffFileId: null,

    // Architecture
    architecture: {
      nodes: [],
      connections: [],
    },
    selectedArchitectureNodeId: null,
    expandedArchitectureNodes: [],
    architectureLayout: 'tree',
    showArchitectureMetrics: true,

    // Refactor
    refactorOpportunities: [],
    selectedRefactorOpportunityId: null,

    // Code Generation
    generatedCode: null,
    isGenerating: false,
    generationError: null,
    showGenerationExplanation: false,

    // Inline AI
    inlineSuggestions: [],
    currentInlineSuggestion: null,
    inlineAIEnabled: true,

    // Intent Explorer
    intents: [],
    selectedIntentId: null,
    showIntentReasoning: false,

    // Semantic Search
    searchQuery: '',
    searchResults: [],
    selectedSearchResultId: null,
    searchType: 'semantic',
    isSearching: false,

    // Lineage
    lineageNodes: [],
    selectedLineageNodeId: null,
    focusedFilePath: null,

    // Risk Analysis
    riskIndicators: [],
    selectedRiskId: null,
    showRiskMetrics: false,

    // Import Graph
    imports: [],
    importEdges: [],
    selectedImportNodeId: null,
    selectedImportFilePath: null,

    // Execution Aware Editor
    executionPoints: [],
    currentExecutionPointId: null,
    breakpoints: [],
    watchExpressions: [],
    isRunning: false,

    // AI Completion Inspector
    completions: [],
    selectedCompletionId: null,
    currentCompletion: null,

    // Refactor Simulation
    simulation: null,
    simulationStep: 0,

    // Architecture Constraints
    constraints: [],
    selectedConstraintId: null,

    // Code Ownership
    ownerships: [],
    selectedFilePath: null,
    selectedOwnerId: null,

    // UI State
    activePanel: 'editor',
    sidebarCollapsed: false,

    // ============================================================================
    // Editor Actions
    // ============================================================================

    setEditorValue: (value) =>
      set((state) => {
        state.editor.value = value
      }),

    setEditorLanguage: (language) =>
      set((state) => {
        state.editor.language = language
      }),

    setEditorFilePath: (filePath) =>
      set((state) => {
        state.editor.filePath = filePath
      }),

    setEditorCursorPosition: (position) =>
      set((state) => {
        state.editor.cursorPosition = position
      }),

    setEditorSelection: (selection) =>
      set((state) => {
        state.editor.selection = selection
      }),

    setEditorSymbols: (symbols) =>
      set((state) => {
        state.editor.symbols = symbols
      }),

    setEditorDiagnostics: (diagnostics) =>
      set((state) => {
        state.editor.diagnostics = diagnostics
      }),

    // ============================================================================
    // AST Actions
    // ============================================================================

    setAST: (ast) =>
      set((state) => {
        state.ast = ast
      }),

    selectASTNode: (nodeId) =>
      set((state) => {
        state.selectedASTNodeId = nodeId
      }),

    toggleASTNode: (nodeId) =>
      set((state) => {
        const idx = state.expandedASTNodes.indexOf(nodeId)
        if (idx === -1) {
          state.expandedASTNodes.push(nodeId)
        } else {
          state.expandedASTNodes.splice(idx, 1)
        }
      }),

    setASTSearchQuery: (query) =>
      set((state) => {
        state.astSearchQuery = query
      }),

    // ============================================================================
    // Dependencies Actions
    // ============================================================================

    setDependencies: (nodes, edges) =>
      set((state) => {
        state.dependencies.nodes = nodes
        state.dependencies.edges = edges
      }),

    selectDependencyNode: (nodeId) =>
      set((state) => {
        state.selectedDependencyNodeId = nodeId
      }),

    setDependencyColorMetric: (metric) =>
      set((state) => {
        state.dependencyColorMetric = metric
      }),

    // ============================================================================
    // Diff Actions
    // ============================================================================

    setDiffFiles: (files) =>
      set((state) => {
        state.diffFiles = files
      }),

    selectDiffFile: (fileId) =>
      set((state) => {
        state.selectedDiffFileId = fileId
      }),

    // ============================================================================
    // Architecture Actions
    // ============================================================================

    setArchitecture: (nodes, connections) =>
      set((state) => {
        state.architecture.nodes = nodes
        state.architecture.connections = connections
      }),

    selectArchitectureNode: (nodeId) =>
      set((state) => {
        state.selectedArchitectureNodeId = nodeId
      }),

    toggleArchitectureNode: (nodeId) =>
      set((state) => {
        const idx = state.expandedArchitectureNodes.indexOf(nodeId)
        if (idx === -1) {
          state.expandedArchitectureNodes.push(nodeId)
        } else {
          state.expandedArchitectureNodes.splice(idx, 1)
        }
      }),

    setArchitectureLayout: (layout) =>
      set((state) => {
        state.architectureLayout = layout
      }),

    setShowArchitectureMetrics: (show) =>
      set((state) => {
        state.showArchitectureMetrics = show
      }),

    // ============================================================================
    // Refactor Actions
    // ============================================================================

    setRefactorOpportunities: (opportunities) =>
      set((state) => {
        state.refactorOpportunities = opportunities
      }),

    selectRefactorOpportunity: (opportunityId) =>
      set((state) => {
        state.selectedRefactorOpportunityId = opportunityId
      }),

    // ============================================================================
    // Code Generation Actions
    // ============================================================================

    setGeneratedCode: (code) =>
      set((state) => {
        state.generatedCode = code
      }),

    setIsGenerating: (isGenerating) =>
      set((state) => {
        state.isGenerating = isGenerating
      }),

    setGenerationError: (error) =>
      set((state) => {
        state.generationError = error
      }),

    toggleGenerationExplanation: () =>
      set((state) => {
        state.showGenerationExplanation = !state.showGenerationExplanation
      }),

    clearGeneratedCode: () =>
      set((state) => {
        state.generatedCode = null
        state.generationError = null
      }),

    // ============================================================================
    // Inline AI Actions
    // ============================================================================

    setInlineSuggestions: (suggestions) =>
      set((state) => {
        state.inlineSuggestions = suggestions
      }),

    setCurrentInlineSuggestion: (suggestion) =>
      set((state) => {
        state.currentInlineSuggestion = suggestion
      }),

    toggleInlineAI: () =>
      set((state) => {
        state.inlineAIEnabled = !state.inlineAIEnabled
      }),

    // ============================================================================
    // Intent Explorer Actions
    // ============================================================================

    setIntents: (intents) =>
      set((state) => {
        state.intents = intents
      }),

    selectIntent: (intentId) =>
      set((state) => {
        state.selectedIntentId = intentId
      }),

    toggleIntentReasoning: () =>
      set((state) => {
        state.showIntentReasoning = !state.showIntentReasoning
      }),

    // ============================================================================
    // Semantic Search Actions
    // ============================================================================

    setSearchQuery: (query) =>
      set((state) => {
        state.searchQuery = query
      }),

    setSearchResults: (results) =>
      set((state) => {
        state.searchResults = results
      }),

    selectSearchResult: (resultId) =>
      set((state) => {
        state.selectedSearchResultId = resultId
      }),

    setSearchType: (type) =>
      set((state) => {
        state.searchType = type
      }),

    setIsSearching: (isSearching) =>
      set((state) => {
        state.isSearching = isSearching
      }),

    clearSearchResults: () =>
      set((state) => {
        state.searchResults = []
        state.selectedSearchResultId = null
      }),

    // ============================================================================
    // Lineage Actions
    // ============================================================================

    setLineageNodes: (nodes) =>
      set((state) => {
        state.lineageNodes = nodes
      }),

    selectLineageNode: (nodeId) =>
      set((state) => {
        state.selectedLineageNodeId = nodeId
      }),

    setFocusedFilePath: (filePath) =>
      set((state) => {
        state.focusedFilePath = filePath
      }),

    // ============================================================================
    // Risk Analysis Actions
    // ============================================================================

    setRiskIndicators: (risks) =>
      set((state) => {
        state.riskIndicators = risks
      }),

    selectRisk: (riskId) =>
      set((state) => {
        state.selectedRiskId = riskId
      }),

    setShowRiskMetrics: (show) =>
      set((state) => {
        state.showRiskMetrics = show
      }),

    // ============================================================================
    // Import Graph Actions
    // ============================================================================

    setImports: (imports, edges) =>
      set((state) => {
        state.imports = imports
        state.importEdges = edges
      }),

    selectImportNode: (nodeId) =>
      set((state) => {
        state.selectedImportNodeId = nodeId
      }),

    setSelectedImportFilePath: (filePath) =>
      set((state) => {
        state.selectedImportFilePath = filePath
      }),

    // ============================================================================
    // Execution Aware Editor Actions
    // ============================================================================

    setExecutionPoints: (points) =>
      set((state) => {
        state.executionPoints = points
      }),

    selectExecutionPoint: (pointId) =>
      set((state) => {
        state.currentExecutionPointId = pointId
      }),

    toggleBreakpoint: (line) =>
      set((state) => {
        const idx = state.breakpoints.indexOf(line)
        if (idx === -1) {
          state.breakpoints.push(line)
        } else {
          state.breakpoints.splice(idx, 1)
        }
      }),

    addWatchExpression: (expression) =>
      set((state) => {
        state.watchExpressions.push({
          id: `watch-${Date.now()}`,
          expression,
        })
      }),

    removeWatchExpression: (expressionId) =>
      set((state) => {
        const idx = state.watchExpressions.findIndex((e) => e.id === expressionId)
        if (idx !== -1) {
          state.watchExpressions.splice(idx, 1)
        }
      }),

    setIsRunning: (isRunning) =>
      set((state) => {
        state.isRunning = isRunning
      }),

    // ============================================================================
    // AI Completion Inspector Actions
    // ============================================================================

    setCompletions: (completions) =>
      set((state) => {
        state.completions = completions
      }),

    selectCompletion: (completionId) =>
      set((state) => {
        state.selectedCompletionId = completionId
        state.currentCompletion = state.completions.find((c) => c.id === completionId) || null
      }),

    // ============================================================================
    // Refactor Simulation Actions
    // ============================================================================

    setSimulation: (simulation) =>
      set((state) => {
        state.simulation = simulation
        state.simulationStep = 0
      }),

    stepSimulationForward: () =>
      set((state) => {
        if (state.simulation && state.simulationStep < state.simulation.changes.length - 1) {
          state.simulationStep++
        }
      }),

    stepSimulationBackward: () =>
      set((state) => {
        if (state.simulationStep > 0) {
          state.simulationStep--
        }
      }),

    clearSimulation: () =>
      set((state) => {
        state.simulation = null
        state.simulationStep = 0
      }),

    // ============================================================================
    // Architecture Constraints Actions
    // ============================================================================

    setConstraints: (constraints) =>
      set((state) => {
        state.constraints = constraints
      }),

    selectConstraint: (constraintId) =>
      set((state) => {
        state.selectedConstraintId = constraintId
      }),

    // ============================================================================
    // Code Ownership Actions
    // ============================================================================

    setOwnerships: (ownerships) =>
      set((state) => {
        state.ownerships = ownerships
      }),

    selectFileOwnership: (filePath) =>
      set((state) => {
        state.selectedFilePath = filePath
      }),

    selectOwner: (ownerId) =>
      set((state) => {
        state.selectedOwnerId = ownerId
      }),

    // ============================================================================
    // UI Actions
    // ============================================================================

    setActivePanel: (panel) =>
      set((state) => {
        state.activePanel = panel
      }),

    toggleSidebar: () =>
      set((state) => {
        state.sidebarCollapsed = !state.sidebarCollapsed
      }),
  }))
)

// ============================================================================
// Selectors
// ============================================================================

export const useEditor = () => useCodeIntelligenceStore((state) => state.editor)
export const useAST = () => useCodeIntelligenceStore((state) => ({
  ast: state.ast,
  selectedNodeId: state.selectedASTNodeId,
  expandedNodes: state.expandedASTNodes,
  searchQuery: state.astSearchQuery,
}))
export const useDependencies = () => useCodeIntelligenceStore((state) => state.dependencies)
export const useDiff = () => useCodeIntelligenceStore((state) => ({
  files: state.diffFiles,
  selectedFileId: state.selectedDiffFileId,
}))
export const useArchitecture = () => useCodeIntelligenceStore((state) => state.architecture)
export const useRefactor = () => useCodeIntelligenceStore((state) => ({
  opportunities: state.refactorOpportunities,
  selectedId: state.selectedRefactorOpportunityId,
}))
export const useGeneratedCode = () => useCodeIntelligenceStore((state) => ({
  code: state.generatedCode,
  isGenerating: state.isGenerating,
  error: state.generationError,
  showExplanation: state.showGenerationExplanation,
}))
export const useInlineAI = () => useCodeIntelligenceStore((state) => ({
  suggestions: state.inlineSuggestions,
  current: state.currentInlineSuggestion,
  enabled: state.inlineAIEnabled,
}))
export const useIntents = () => useCodeIntelligenceStore((state) => ({
  intents: state.intents,
  selectedId: state.selectedIntentId,
  showReasoning: state.showIntentReasoning,
}))
export const useSearch = () => useCodeIntelligenceStore((state) => ({
  query: state.searchQuery,
  results: state.searchResults,
  selectedId: state.selectedSearchResultId,
  searchType: state.searchType,
  isSearching: state.isSearching,
}))
export const useLineage = () => useCodeIntelligenceStore((state) => ({
  nodes: state.lineageNodes,
  selectedId: state.selectedLineageNodeId,
  focusedFilePath: state.focusedFilePath,
}))
export const useRisk = () => useCodeIntelligenceStore((state) => ({
  risks: state.riskIndicators,
  selectedId: state.selectedRiskId,
  showMetrics: state.showRiskMetrics,
}))
export const useImports = () => useCodeIntelligenceStore((state) => ({
  imports: state.imports,
  edges: state.importEdges,
  selectedNodeId: state.selectedImportNodeId,
  filePath: state.selectedImportFilePath,
}))
export const useExecutionEditor = () => useCodeIntelligenceStore((state) => ({
  executionPoints: state.executionPoints,
  currentPointId: state.currentExecutionPointId,
  breakpoints: state.breakpoints,
  watchExpressions: state.watchExpressions,
  isRunning: state.isRunning,
}))
export const useCompletions = () => useCodeIntelligenceStore((state) => ({
  completions: state.completions,
  selectedId: state.selectedCompletionId,
  current: state.currentCompletion,
}))
export const useSimulation = () => useCodeIntelligenceStore((state) => ({
  simulation: state.simulation,
  step: state.simulationStep,
}))
export const useConstraints = () => useCodeIntelligenceStore((state) => ({
  constraints: state.constraints,
  selectedId: state.selectedConstraintId,
}))
export const useOwnership = () => useCodeIntelligenceStore((state) => ({
  ownerships: state.ownerships,
  selectedFilePath: state.selectedFilePath,
  selectedOwnerId: state.selectedOwnerId,
}))
export const useCodeIntelligenceUI = () => useCodeIntelligenceStore((state) => ({
  activePanel: state.activePanel,
  sidebarCollapsed: state.sidebarCollapsed,
}))
