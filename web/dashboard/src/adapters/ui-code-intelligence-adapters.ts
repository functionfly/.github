/**
 * Code Intelligence Adapters
 * Transform dashboard data formats to match UI package component expectations
 */

/**
 * Target types for UI package compatibility (derived from ui-code-intelligence types.d.ts)
 */

interface UICodePosition {
  line: number;
  column: number;
  offset?: number;
}

interface UICodeRange {
  start: UICodePosition;
  end: UICodePosition;
}

interface UIDependencyNode {
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

interface UIDependencyEdge {
  source: string;
  target: string;
  weight?: number;
  type: 'import' | 'inheritance' | 'composition' | 'dependency';
  line?: number;
}

interface UIDiffLine {
  id: string;
  type: 'add' | 'delete' | 'context' | 'header';
  content: string;
  oldLineNumber?: number;
  newLineNumber?: number;
  comment?: string;
  author?: string;
  timestamp?: number;
}

interface UIDiffHunk {
  id: string;
  oldStart: number;
  oldCount: number;
  newStart: number;
  newCount: number;
  lines: UIDiffLine[];
}

interface UIDiffFile {
  id: string;
  path: string;
  oldContent: string;
  newContent: string;
  language?: string;
  status?: 'added' | 'deleted' | 'modified' | 'renamed';
}

interface UIArchitectureNode {
  id: string;
  name: string;
  type: 'module' | 'package' | 'component' | 'service' | 'layer';
  path?: string;
  children?: UIArchitectureNode[];
  metrics?: {
    components?: number;
    linesOfCode?: number;
    complexity?: number;
    stability?: number;
  };
  parent?: string;
  depth?: number;
}

interface UIArchitectureConnection {
  id: string;
  source: string;
  target: string;
  type: 'depends-on' | 'composes' | 'extends' | 'implements' | 'uses';
  weight?: number;
  bidirectional?: boolean;
}

interface UIRefactorOpportunity {
  id: string;
  type: 'extract-method' | 'inline-method' | 'rename' | 'move' | 'extract-variable' | 'change-signature' | 'extract-interface' | 'introduce-parameter';
  title: string;
  description: string;
  location: UICodeRange;
  original: string;
  preview: string;
  impact: 'low' | 'medium' | 'high';
  estimatedComplexity?: number;
  affectedFiles?: string[];
  automated?: boolean;
}

interface UIRefactorPreview {
  fileId: string;
  originalContent: string;
  refactoredContent: string;
  hunks: UIDiffHunk[];
}

interface UICodeIntent {
  id: string;
  type: 'feature' | 'bugfix' | 'refactor' | 'optimization' | 'documentation' | 'test' | 'security' | 'compliance';
  confidence: number;
  description: string;
  affectedCodeRanges: UICodeRange[];
  affectedFiles: string[];
  reasoning: string;
  relatedIntents?: string[];
  extractedRequirements?: string[];
}

interface UISearchResult {
  id: string;
  filePath: string;
  lineNumber: number;
  lineContent: string;
  matchedText: string;
  context: string[];
  score: number;
  matchType: 'exact' | 'fuzzy' | 'semantic';
  symbols?: Array<{
    id: string;
    name: string;
    kind: string;
    location: UICodeRange;
    scope: string;
    signature?: string;
    documentation?: string;
    children?: unknown[];
  }>;
}

interface UILineageNode {
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

interface UIRiskIndicator {
  id: string;
  type: 'security' | 'performance' | 'maintainability' | 'testability' | 'complexity' | 'duplication';
  severity: 'critical' | 'high' | 'medium' | 'low' | 'info';
  message: string;
  location?: UICodeRange;
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

interface UIImportNode {
  id: string;
  name: string;
  type: 'default' | 'named' | 'namespace' | 'side-effect';
  source: string;
  isReExported?: boolean;
  line?: number;
}

interface UIImportEdge {
  source: string;
  target: string;
  type: 'import' | 're-export' | 'type-import';
}

interface UIWatchExpression {
  id: string;
  expression: string;
  value?: unknown;
  type?: string;
  error?: string;
}

interface UIAICompletion {
  id: string;
  text: string;
  type: 'completion' | 'refactor' | 'explanation' | 'test';
  confidence: number;
  model?: string;
  timestamp: number;
  latency?: number;
  tokens?: number;
  context?: {
    cursorPosition?: UICodePosition;
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

interface UISimulationChange {
  fileId: string;
  filePath: string;
  changeType: 'add' | 'modify' | 'delete';
  before: string;
  after: string;
  hunks: UIDiffHunk[];
}

interface UIRefactorSimulation {
  id: string;
  name: string;
  description: string;
  changes: UISimulationChange[];
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

interface UIArchitectureConstraint {
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

interface UICodeOwner {
  id: string;
  name: string;
  email: string;
  avatar?: string;
  gitHubUsername?: string;
}

interface UIFileOwnership {
  filePath: string;
  owners: UICodeOwner[];
  lastModified?: number;
  lastModifiedBy?: string;
  reviewRequired?: boolean;
  autoAssignment?: boolean;
}

export interface PositionInput {
  line: number;
  column?: number;
}

export interface RangeInput {
  start: PositionInput;
  end: PositionInput;
}

export interface DependencyNodeInput {
  id: string;
  name: string;
  path: string;
  version?: string;
  type: 'package' | 'file' | 'module';
  size?: number;
  metrics?: {
    complexity?: number;
    coupling?: number;
    changes?: number;
    bugs?: number;
    maintenance?: number;
  };
}

export interface DependencyEdgeInput {
  source: string;
  target: string;
  type?: 'import' | 'inheritance' | 'composition' | 'dependency';
  weight?: number;
}

export interface DiffFileInput {
  id: string;
  path: string;
  oldContent: string;
  newContent: string;
  language?: string;
  status?: 'added' | 'deleted' | 'modified' | 'renamed';
}

export interface ArchitectureNodeInput {
  id: string;
  name: string;
  type: 'module' | 'package' | 'component' | 'service' | 'layer';
  path?: string;
  children?: ArchitectureNodeInput[];
  metrics?: {
    components?: number;
    linesOfCode?: number;
    complexity?: number;
    stability?: number;
  };
}

export interface RefactorOpportunityInput {
  id: string;
  type: 'extract-method' | 'inline-method' | 'rename' | 'move' | 'extract-variable' | 'change-signature' | 'extract-interface' | 'introduce-parameter' | string;
  title: string;
  description: string;
  location: RangeInput;
  original: string;
  preview: string;
  impact: 'low' | 'medium' | 'high';
  estimatedComplexity?: number;
  affectedFiles?: string[];
  automated?: boolean;
}

export interface SearchResultInput {
  id: string;
  filePath: string;
  lineNumber: number;
  lineContent: string;
  matchedText: string;
  context: string[];
  score: number;
  matchType: 'exact' | 'fuzzy' | 'semantic';
  symbols?: Array<{
    id: string;
    name: string;
    kind: string;
    location: RangeInput;
    scope: string;
    signature?: string;
    documentation?: string;
  }>;
}

export interface LineageNodeInput {
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

export interface RiskIndicatorInput {
  id: string;
  type: 'security' | 'performance' | 'maintainability' | 'testability' | 'complexity' | 'duplication';
  severity: 'critical' | 'high' | 'medium' | 'low' | 'info';
  message: string;
  location?: RangeInput;
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

export interface ImportNodeInput {
  id: string;
  name: string;
  type: 'default' | 'named' | 'namespace' | 'side-effect';
  source: string;
  isReExported?: boolean;
  line?: number;
}

export interface ImportEdgeInput {
  source: string;
  target: string;
  type: 'import' | 're-export' | 'type-import';
}

export interface WatchExpressionInput {
  id: string;
  expression: string;
  value?: unknown;
  type?: string;
  error?: string;
}

export interface AICompletionInput {
  id: string;
  text: string;
  type: 'completion' | 'refactor' | 'explanation' | 'test';
  confidence: number;
  model?: string;
  timestamp: number;
  latency?: number;
  tokens?: number;
  context?: {
    cursorPosition?: PositionInput;
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

export interface SimulationChangeInput {
  fileId: string;
  filePath: string;
  changeType: 'add' | 'modify' | 'delete';
  before: string;
  after: string;
  hunks?: UIDiffHunk[];
}

export interface RefactorSimulationInput {
  id: string;
  name: string;
  description: string;
  changes: SimulationChangeInput[];
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

export interface ArchitectureConstraintInput {
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

export interface FileOwnershipInput {
  filePath: string;
  owners: Array<{
    id: string;
    name: string;
    email: string;
    avatar?: string;
    gitHubUsername?: string;
  }>;
  lastModified?: number;
  lastModifiedBy?: string;
  reviewRequired?: boolean;
  autoAssignment?: boolean;
}

/**
 * Adapt CodePosition
 */
export function adaptCodePosition(pos: PositionInput): UICodePosition {
  return {
    line: pos.line,
    column: pos.column ?? 1,
    offset: undefined,
  };
}

/**
 * Adapt CodeRange
 */
export function adaptCodeRange(range: RangeInput): UICodeRange {
  return {
    start: adaptCodePosition(range.start),
    end: adaptCodePosition(range.end),
  };
}

/**
 * Adapt DependencyNode
 */
export function adaptDependencyNode(node: DependencyNodeInput): UIDependencyNode {
  return {
    id: node.id,
    name: node.name,
    path: node.path,
    version: node.version,
    type: node.type,
    size: node.size,
    metrics: node.metrics
      ? {
          ...node.metrics,
          cohesion: undefined,
        }
      : undefined,
  };
}

/**
 * Adapt DependencyEdge
 */
export function adaptDependencyEdge(edge: DependencyEdgeInput): UIDependencyEdge {
  return {
    source: edge.source,
    target: edge.target,
    weight: edge.weight,
    type: edge.type || 'dependency',
    line: undefined,
  };
}

/**
 * Adapt RefactorOpportunity
 */
export function adaptRefactorOpportunity(opp: RefactorOpportunityInput): UIRefactorOpportunity {
  return {
    id: opp.id,
    type: opp.type as UIRefactorOpportunity['type'],
    title: opp.title,
    description: opp.description,
    location: adaptCodeRange(opp.location),
    original: opp.original,
    preview: opp.preview,
    impact: opp.impact,
    estimatedComplexity: opp.estimatedComplexity,
    affectedFiles: opp.affectedFiles,
    automated: opp.automated,
  };
}

/**
 * Adapt SearchResult
 */
export function adaptSearchResult(result: SearchResultInput): UISearchResult {
  return {
    id: result.id,
    filePath: result.filePath,
    lineNumber: result.lineNumber,
    lineContent: result.lineContent,
    matchedText: result.matchedText,
    context: result.context,
    score: result.score,
    matchType: result.matchType,
    symbols: result.symbols?.map(s => ({
      id: s.id,
      name: s.name,
      kind: s.kind as any,
      location: adaptCodeRange(s.location),
      scope: s.scope,
      signature: s.signature,
      documentation: s.documentation,
      children: undefined,
    })),
  };
}

/**
 * Adapt RiskIndicator
 */
export function adaptRiskIndicator(risk: RiskIndicatorInput): UIRiskIndicator {
  return {
    id: risk.id,
    type: risk.type,
    severity: risk.severity,
    message: risk.message,
    location: risk.location ? adaptCodeRange(risk.location) : undefined,
    file: risk.file,
    code: risk.code,
    suggestion: risk.suggestion,
    cwe: risk.cwe,
    estimates: risk.estimates,
  };
}

/**
 * Adapt ImportNode
 */
export function adaptImportNode(node: ImportNodeInput): UIImportNode {
  return {
    id: node.id,
    name: node.name,
    type: node.type,
    source: node.source,
    isReExported: node.isReExported,
    line: node.line,
  };
}

/**
 * Adapt ImportEdge
 */
export function adaptImportEdge(edge: ImportEdgeInput): UIImportEdge {
  return {
    source: edge.source,
    target: edge.target,
    type: edge.type,
  };
}

/**
 * Adapt WatchExpression
 */
export function adaptWatchExpression(expr: WatchExpressionInput): UIWatchExpression {
  return {
    id: expr.id,
    expression: expr.expression,
    value: expr.value,
    type: expr.type,
    error: expr.error,
  };
}

/**
 * Adapt AICompletion
 */
export function adaptAICompletion(comp: AICompletionInput): UIAICompletion {
  return {
    id: comp.id,
    text: comp.text,
    type: comp.type,
    confidence: comp.confidence,
    model: comp.model,
    timestamp: comp.timestamp,
    latency: comp.latency,
    tokens: comp.tokens,
    context: comp.context
      ? {
          cursorPosition: comp.context.cursorPosition
            ? adaptCodePosition(comp.context.cursorPosition)
            : undefined,
          selectedText: comp.context.selectedText,
          filePath: comp.context.filePath,
          language: comp.context.language,
        }
      : undefined,
    alternatives: comp.alternatives,
  };
}

/**
 * Adapt SimulationChange
 */
export function adaptSimulationChange(change: SimulationChangeInput): UISimulationChange {
  return {
    fileId: change.fileId,
    filePath: change.filePath,
    changeType: change.changeType,
    before: change.before,
    after: change.after,
    hunks: change.hunks || [],
  };
}

/**
 * Adapt RefactorSimulation
 */
export function adaptRefactorSimulation(sim: RefactorSimulationInput): UIRefactorSimulation {
  return {
    id: sim.id,
    name: sim.name,
    description: sim.description,
    changes: sim.changes.map(adaptSimulationChange),
    impactAnalysis: sim.impactAnalysis,
    validationResults: sim.validationResults,
  };
}

/**
 * Adapt FileOwnership
 */
export function adaptFileOwnership(ownership: FileOwnershipInput): UIFileOwnership {
  return {
    filePath: ownership.filePath,
    owners: ownership.owners.map(o => ({
      id: o.id,
      name: o.name,
      email: o.email,
      avatar: o.avatar,
      gitHubUsername: o.gitHubUsername,
    })),
    lastModified: ownership.lastModified,
    lastModifiedBy: ownership.lastModifiedBy,
    reviewRequired: ownership.reviewRequired,
    autoAssignment: ownership.autoAssignment,
  };
}
