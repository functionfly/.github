/**
 * @functionfly/ui-code-intelligence
 * Code Intelligence Components - Types and Interfaces
 */

// ============================================================================
// Semantic Code Editor
// ============================================================================

export interface CodePosition {
  line: number;
  column: number;
  offset?: number;
}

export interface CodeRange {
  start: CodePosition;
  end: CodePosition;
}

export interface CodeToken {
  type: 'keyword' | 'string' | 'number' | 'comment' | 'operator' | 'function' | 'variable' | 'type' | 'punctuation' | 'import' | 'export';
  value: string;
  range: CodeRange;
  metadata?: Record<string, unknown>;
}

export interface SemanticSymbol {
  id: string;
  name: string;
  kind: 'function' | 'class' | 'interface' | 'type' | 'variable' | 'constant' | 'enum' | 'namespace' | 'method' | 'property';
  location: CodeRange;
  scope: string;
  signature?: string;
  documentation?: string;
  children?: SemanticSymbol[];
}

export interface CodeLens {
  id: string;
  range: CodeRange;
  command: {
    id: string;
    title: string;
    arguments?: unknown[];
  };
  isResolved: boolean;
}

export interface Diagnostic {
  id: string;
  severity: 'error' | 'warning' | 'information' | 'hint';
  message: string;
  range: CodeRange;
  code?: string | number;
  source?: string;
  relatedInformation?: Array<{
    location: CodeRange;
    message: string;
  }>;
}

export interface SemanticCodeEditorProps {
  value: string;
  language: string;
  onChange?: (value: string) => void;
  onSave?: (value: string) => void;
  readOnly?: boolean;
  showLineNumbers?: boolean;
  showSemanticHighlights?: boolean;
  showCodeLenses?: boolean;
  symbols?: SemanticSymbol[];
  tokens?: CodeToken[];
  codeLenses?: CodeLens[];
  diagnostics?: Diagnostic[];
  onCursorChange?: (position: CodePosition) => void;
  onSelectionChange?: (range: CodeRange | null) => void;
  onSymbolClick?: (symbol: SemanticSymbol) => void;
  className?: string;
}

// ============================================================================
// AST Explorer
// ============================================================================

export interface ASTNode {
  id: string;
  type: string;
  loc: CodeRange;
  range: CodeRange;
  children?: ASTNode[];
  value?: unknown;
  metadata?: Record<string, unknown>;
  parent?: string;
  depth?: number;
  path?: string[];
}

export interface ASTExplorerProps {
  ast: ASTNode | null;
  rootId?: string;
  selectedNodeId?: string | null;
  highlightedNodes?: string[];
  searchQuery?: string;
  onNodeSelect?: (node: ASTNode) => void;
  onNodeExpand?: (nodeId: string) => void;
  onNodeCollapse?: (nodeId: string) => void;
  onSearch?: (query: string) => void;
  showRange?: boolean;
  showMetadata?: boolean;
  className?: string;
}

// ============================================================================
// Dependency Heatmap
// ============================================================================

export interface DependencyNode {
  id: string;
  name: string;
  path: string;
  version?: string;
  type: 'package' | 'file' | 'module';
  size?: number;
  metrics?: {
    complexity?: number;
    coupling?: number;
    cohesion?: number;
    changes?: number;
    bugs?: number;
    maintenance?: number;
  };
}

export interface DependencyEdge {
  source: string;
  target: string;
  weight?: number;
  type: 'import' | 'inheritance' | 'composition' | 'dependency';
  line?: number;
}

export interface DependencyHeatmapProps {
  nodes: DependencyNode[];
  edges: DependencyEdge[];
  selectedNodeId?: string | null;
  highlightedNodes?: string[];
  colorMetric?: 'complexity' | 'coupling' | 'changes' | 'bugs' | 'maintenance';
  onNodeClick?: (node: DependencyNode) => void;
  onNodeHover?: (node: DependencyNode | null) => void;
  className?: string;
}

// ============================================================================
// Multi File Diff Viewer
// ============================================================================

export interface DiffFile {
  id: string;
  path: string;
  oldContent: string;
  newContent: string;
  language?: string;
  status?: 'added' | 'deleted' | 'modified' | 'renamed';
}

export interface DiffHunk {
  id: string;
  oldStart: number;
  oldCount: number;
  newStart: number;
  newCount: number;
  lines: DiffLine[];
}

export interface DiffLine {
  id: string;
  type: 'add' | 'delete' | 'context' | 'header';
  content: string;
  oldLineNumber?: number;
  newLineNumber?: number;
  comment?: string;
  author?: string;
  timestamp?: number;
}

export interface MultiFileDiffViewerProps {
  files: DiffFile[];
  selectedFileId?: string | null;
  hunks?: DiffHunk[];
  showLineNumbers?: boolean;
  showComments?: boolean;
  showSyntaxHighlighting?: boolean;
  diffContext?: number;
  onFileSelect?: (file: DiffFile) => void;
  onCommentAdd?: (lineId: string, comment: string) => void;
  onAcceptChange?: (fileId: string, changeId: string) => void;
  onRejectChange?: (fileId: string, changeId: string) => void;
  className?: string;
}

// ============================================================================
// Architecture Map
// ============================================================================

export interface ArchitectureNode {
  id: string;
  name: string;
  type: 'module' | 'package' | 'component' | 'service' | 'layer';
  path?: string;
  children?: ArchitectureNode[];
  metrics?: {
    components?: number;
    linesOfCode?: number;
    complexity?: number;
    stability?: number;
  };
  parent?: string;
  depth?: number;
}

export interface ArchitectureConnection {
  id: string;
  source: string;
  target: string;
  type: 'depends-on' | 'composes' | 'extends' | 'implements' | 'uses';
  weight?: number;
  bidirectional?: boolean;
}

export interface ArchitectureMapProps {
  nodes: ArchitectureNode[];
  connections: ArchitectureConnection[];
  selectedNodeId?: string | null;
  expandedNodes?: string[];
  showMetrics?: boolean;
  layout?: 'tree' | 'force' | 'circular';
  onNodeSelect?: (node: ArchitectureNode) => void;
  onNodeExpand?: (nodeId: string) => void;
  onNodeCollapse?: (nodeId: string) => void;
  className?: string;
}

// ============================================================================
// Smart Refactor Panel
// ============================================================================

export interface RefactorOpportunity {
  id: string;
  type: 'extract-method' | 'inline-method' | 'rename' | 'move' | 'extract-variable' | 'change-signature' | 'extract-interface' | 'introduce-parameter';
  title: string;
  description: string;
  location: CodeRange;
  original: string;
  preview: string;
  impact: 'low' | 'medium' | 'high';
  estimatedComplexity?: number;
  affectedFiles?: string[];
  automated?: boolean;
}

export interface RefactorPreview {
  fileId: string;
  originalContent: string;
  refactoredContent: string;
  hunks: DiffHunk[];
}

export interface SmartRefactorPanelProps {
  opportunities: RefactorOpportunity[];
  selectedOpportunityId?: string | null;
  previews?: RefactorPreview[];
  onOpportunitySelect?: (opportunity: RefactorOpportunity) => void;
  onPreviewGenerated?: (preview: RefactorPreview) => void;
  onApply?: (opportunityId: string) => void;
  onReject?: (opportunityId: string) => void;
  className?: string;
}

// ============================================================================
// Code Generation Preview
// ============================================================================

export interface GeneratedCode {
  id: string;
  language: string;
  code: string;
  title?: string;
  description?: string;
  context?: {
    originalCode?: string;
    language?: string;
    framework?: string;
    requirements?: string[];
  };
  metrics?: {
    complexity?: number;
    maintainability?: number;
    testability?: number;
    estimatedTokens?: number;
  };
  dependencies?: Array<{
    name: string;
    version: string;
  }>;
}

export interface CodeGenerationPreviewProps {
  generation: GeneratedCode | null;
  loading?: boolean;
  error?: string | null;
  onCopy?: (code: string) => void;
  onApply?: () => void;
  onDiscard?: () => void;
  onRegenerate?: () => void;
  onToggleExplanation?: () => void;
  showExplanation?: boolean;
  className?: string;
}

// ============================================================================
// Inline AI Assistant
// ============================================================================

export interface AIInlineSuggestion {
  id: string;
  type: 'completion' | 'refactor' | 'documentation' | 'test' | 'explanation';
  text: string;
  confidence: number;
  startPosition: CodePosition;
  endPosition: CodePosition;
  explanation?: string;
  ranking?: number;
}

export interface InlineAIAssistantProps {
  enabled?: boolean;
  suggestions?: AIInlineSuggestion[];
  currentSuggestion?: AIInlineSuggestion | null;
  onSuggestionAccept?: (suggestion: AIInlineSuggestion) => void;
  onSuggestionReject?: (suggestion: AIInlineSuggestion) => void;
  onSuggestionHover?: (suggestion: AIInlineSuggestion | null) => void;
  onExplain?: (suggestion: AIInlineSuggestion) => void;
  className?: string;
}

// ============================================================================
// Code Intent Explorer
// ============================================================================

export interface CodeIntent {
  id: string;
  type: 'feature' | 'bugfix' | 'refactor' | 'optimization' | 'documentation' | 'test' | 'security' | 'compliance';
  confidence: number;
  description: string;
  affectedCodeRanges: CodeRange[];
  affectedFiles: string[];
  reasoning: string;
  relatedIntents?: string[];
  extractedRequirements?: string[];
}

export interface CodeIntentExplorerProps {
  intents: CodeIntent[];
  selectedIntentId?: string | null;
  showReasoning?: boolean;
  onIntentSelect?: (intent: CodeIntent) => void;
  onIntentExpand?: (intentId: string) => void;
  onRequirementExtract?: (intentId: string, requirements: string[]) => void;
  className?: string;
}

// ============================================================================
// Semantic Search Panel
// ============================================================================

export interface SearchResult {
  id: string;
  filePath: string;
  lineNumber: number;
  lineContent: string;
  matchedText: string;
  context: string[];
  score: number;
  matchType: 'exact' | 'fuzzy' | 'semantic';
  symbols?: SemanticSymbol[];
}

export interface SemanticSearchPanelProps {
  query: string;
  results: SearchResult[];
  loading?: boolean;
  selectedResultId?: string | null;
  searchType?: 'text' | 'semantic' | 'symbol' | 'regex';
  filters?: {
    language?: string;
    fileType?: string;
    path?: string;
    dateRange?: { start: number; end: number };
  };
  onQueryChange?: (query: string) => void;
  onSearch?: () => void;
  onResultSelect?: (result: SearchResult) => void;
  onResultHover?: (result: SearchResult | null) => void;
  className?: string;
}

// ============================================================================
// Code Lineage Viewer
// ============================================================================

export interface LineageNode {
  id: string;
  type: 'commit' | 'change' | 'merge' | 'branch';
  name: string;
  author: string;
  timestamp: number;
  message?: string;
  parent?: string;
  children?: string[];
  metadata?: {
    filesChanged?: number;
    insertions?: number;
    deletions?: number;
    branch?: string;
    tags?: string[];
  };
}

export interface CodeLineageViewerProps {
  nodes: LineageNode[];
  selectedNodeId?: string | null;
  focusedFilePath?: string;
  maxDepth?: number;
  onNodeSelect?: (node: LineageNode) => void;
  onNodeExpand?: (nodeId: string) => void;
  className?: string;
}

// ============================================================================
// Code Risk Analyzer
// ============================================================================

export interface RiskIndicator {
  id: string;
  type: 'security' | 'performance' | 'maintainability' | 'testability' | 'complexity' | 'duplication';
  severity: 'critical' | 'high' | 'medium' | 'low' | 'info';
  message: string;
  location?: CodeRange;
  file?: string;
  code?: string;
  suggestion?: string;
  cwe?: string;
  estimates?: {
    effort?: number;
    impact?: number;
    priority?: number;
  };
}

export interface CodeRiskAnalyzerProps {
  risks: RiskIndicator[];
  selectedRiskId?: string | null;
  showMetrics?: boolean;
  onRiskSelect?: (risk: RiskIndicator) => void;
  onRiskHover?: (risk: RiskIndicator | null) => void;
  onFixApply?: (riskId: string) => void;
  onSuppress?: (riskId: string, reason: string) => void;
  className?: string;
}

// ============================================================================
// Import Graph Viewer
// ============================================================================

export interface ImportNode {
  id: string;
  name: string;
  type: 'default' | 'named' | 'namespace' | 'side-effect';
  source: string;
  isReExported?: boolean;
  line?: number;
}

export interface ImportEdge {
  source: string;
  target: string;
  type: 'import' | 're-export' | 'type-import';
}

export interface ImportGraphViewerProps {
  imports: ImportNode[];
  edges: ImportEdge[];
  selectedNodeId?: string | null;
  filePath?: string;
  onNodeClick?: (node: ImportNode) => void;
  onExpandImports?: (nodeId: string) => void;
  className?: string;
}

// ============================================================================
// Execution Aware Editor
// ============================================================================

export interface ExecutionPoint {
  id: string;
  timestamp: number;
  line: number;
  column: number;
  type: 'breakpoint' | 'current' | 'watch' | 'function-call' | 'return';
  callStack?: string[];
  variableState?: Record<string, unknown>;
  hitCount?: number;
  condition?: string;
}

export interface WatchExpression {
  id: string;
  expression: string;
  value?: unknown;
  type?: string;
  error?: string;
}

export interface ExecutionAwareEditorProps {
  code: string;
  language: string;
  executionPoints: ExecutionPoint[];
  currentExecutionPointId?: string | null;
  breakpoints?: string[];
  watchExpressions?: WatchExpression[];
  isRunning?: boolean;
  onExecutionPointSelect?: (point: ExecutionPoint) => void;
  onBreakpointToggle?: (line: number) => void;
  onWatchExpressionAdd?: (expression: string) => void;
  onWatchExpressionRemove?: (expressionId: string) => void;
  className?: string;
}

// ============================================================================
// AI Completion Inspector
// ============================================================================

export interface AICompletion {
  id: string;
  text: string;
  type: 'completion' | 'refactor' | 'explanation' | 'test';
  confidence: number;
  model?: string;
  timestamp: number;
  latency?: number;
  tokens?: number;
  context?: {
    cursorPosition?: CodePosition;
    selectedText?: string;
    filePath?: string;
    language?: string;
  };
  alternatives?: Array<{
    text: string;
    confidence: number;
    model?: string;
  }>;
}

export interface AICompletionInspectorProps {
  completions: AICompletion[];
  selectedCompletionId?: string | null;
  currentCompletion?: AICompletion | null;
  onCompletionSelect?: (completion: AICompletion) => void;
  onCompletionAccept?: (completionId: string) => void;
  onComparisonToggle?: (completionId: string) => void;
  className?: string;
}

// ============================================================================
// Refactor Simulation Viewer
// ============================================================================

export interface SimulationChange {
  fileId: string;
  filePath: string;
  changeType: 'add' | 'modify' | 'delete';
  before: string;
  after: string;
  hunks: DiffHunk[];
}

export interface RefactorSimulation {
  id: string;
  name: string;
  description: string;
  changes: SimulationChange[];
  impactAnalysis?: {
    estimatedTime?: number;
    riskLevel?: 'low' | 'medium' | 'high';
    affectedComponents?: string[];
    testCoverageImpact?: number;
  };
  validationResults?: Array<{
    type: 'compile' | 'test' | 'lint';
    passed: boolean;
    message?: string;
  }>;
}

export interface RefactorSimulationViewerProps {
  simulation: RefactorSimulation | null;
  onAccept?: () => void;
  onReject?: () => void;
  onStepForward?: () => void;
  onStepBackward?: () => void;
  className?: string;
}

// ============================================================================
// Architecture Constraint Panel
// ============================================================================

export interface ArchitectureConstraint {
  id: string;
  type: 'naming' | 'layering' | 'dependency' | 'visibility' | 'pattern';
  name: string;
  description: string;
  severity: 'error' | 'warning' | 'info';
  enforcement: 'strict' | 'advisory';
  violatedBy?: Array<{
    file: string;
    line?: number;
    details?: string;
  }>;
  fixSuggestion?: string;
}

export interface ArchitectureConstraintPanelProps {
  constraints: ArchitectureConstraint[];
  selectedConstraintId?: string | null;
  onConstraintSelect?: (constraint: ArchitectureConstraint) => void;
  onFixApply?: (constraintId: string) => void;
  onDismiss?: (constraintId: string) => void;
  className?: string;
}

// ============================================================================
// Code Ownership Map
// ============================================================================

export interface CodeOwner {
  id: string;
  name: string;
  email: string;
  avatar?: string;
  gitHubUsername?: string;
}

export interface FileOwnership {
  filePath: string;
  owners: CodeOwner[];
  lastModified?: number;
  lastModifiedBy?: string;
  reviewRequired?: boolean;
  autoAssignment?: boolean;
}

export interface CodeOwnershipMapProps {
  ownerships: FileOwnership[];
  selectedFilePath?: string | null;
  selectedOwnerId?: string | null;
  onFileSelect?: (ownership: FileOwnership) => void;
  onOwnerClick?: (owner: CodeOwner) => void;
  onAssign?: (filePath: string, ownerId: string) => void;
  className?: string;
}
