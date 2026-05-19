/**
 * @functionfly/ui-code-intelligence
 * Code Intelligence Components - AI-powered code analysis and editing
 */

import React, { useState, useCallback, useMemo, useRef, useEffect } from 'react';
import { cn } from '@functionfly/ui-core';
import {
  Code,
  GitBranch,
  GitCommit,
  GitPullRequest,
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
  MemoryStick,
  Code2,
  Braces,
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
} from 'lucide-react';

// ============================================================================
// Semantic Code Editor
// ============================================================================

interface SemanticCodeEditorProps {
  value: string;
  language: string;
  onChange?: (value: string) => void;
  onSave?: (value: string) => void;
  readOnly?: boolean;
  showLineNumbers?: boolean;
  showSemanticHighlights?: boolean;
  showCodeLenses?: boolean;
  symbols?: Array<{
    id: string;
    name: string;
    kind: string;
    location: { start: { line: number; column: number }; end: { line: number; column: number } };
    scope: string;
    signature?: string;
    documentation?: string;
  }>;
  tokens?: Array<{
    type: string;
    value: string;
    range: { start: { line: number; column: number }; end: { line: number; column: number } };
  }>;
  codeLenses?: Array<{
    id: string;
    range: { start: { line: number; column: number }; end: { line: number; column: number } };
    command: { id: string; title: string };
  }>;
  diagnostics?: Array<{
    id: string;
    severity: string;
    message: string;
    range: { start: { line: number; column: number }; end: { line: number; column: number } };
  }>;
  onCursorChange?: (position: { line: number; column: number }) => void;
  onSelectionChange?: (range: { start: { line: number; column: number }; end: { line: number; column: number } } | null) => void;
  onSymbolClick?: (symbol: { id: string; name: string; kind: string }) => void;
  className?: string;
}

export const SemanticCodeEditor: React.FC<SemanticCodeEditorProps> = ({
  value,
  language,
  onChange,
  onSave,
  readOnly = false,
  showLineNumbers = true,
  showSemanticHighlights = true,
  showCodeLenses = true,
  symbols = [],
  tokens = [],
  codeLenses = [],
  diagnostics = [],
  onCursorChange,
  onSelectionChange,
  onSymbolClick,
  className,
}) => {
  const [activeLine, setActiveLine] = useState(1);
  const [activeColumn, setActiveColumn] = useState(1);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const lines = useMemo(() => value.split('\n'), [value]);

  const getTokenColor = (type: string) => {
    const colors: Record<string, string> = {
      keyword: 'text-aviation-amber',
      string: 'text-green-400',
      number: 'text-cyan-400',
      comment: 'text-gray-500',
      function: 'text-blue-400',
      variable: 'text-aviation-text-primary',
      type: 'text-purple-400',
      import: 'text-yellow-400',
    };
    return colors[type] || 'text-aviation-text-primary';
  };

  const handleCursorMove = useCallback(() => {
    if (textareaRef.current) {
      const cursor = textareaRef.current.selectionStart;
      const text = value.substring(0, cursor);
      const lines = text.split('\n');
      setActiveLine(lines.length);
      setActiveColumn(lines[lines.length - 1].length + 1);
      onCursorChange?.({ line: lines.length, column: lines[lines.length - 1].length + 1 });
    }
  }, [value, onCursorChange]);

  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if ((e.ctrlKey || e.metaKey) && e.key === 's') {
      e.preventDefault();
      onSave?.(value);
    }
  }, [value, onSave]);

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Code Lens Bar */}
      {showCodeLenses && codeLenses.length > 0 && (
        <div className="flex items-center gap-2 px-3 py-1 bg-aviation-bg-secondary border-b border-aviation-border-panel">
          {codeLenses.map((lens) => (
            <button
              key={lens.id}
              className="flex items-center gap-1 px-2 py-0.5 text-xs text-aviation-cyan hover:bg-aviation-bg-instrument rounded transition-colors"
              onClick={() => {}}
            >
              <Sparkles className="w-3 h-3" />
              {lens.command.title}
            </button>
          ))}
        </div>
      )}

      {/* Diagnostics Bar */}
      {diagnostics.length > 0 && (
        <div className="flex flex-col gap-0.5 px-3 py-1.5 bg-red-950/30 border-b border-red-900/50">
          {diagnostics.map((diag) => (
            <div
              key={diag.id}
              className={cn(
                'flex items-center gap-2 text-xs',
                diag.severity === 'error' ? 'text-red-400' :
                diag.severity === 'warning' ? 'text-amber-400' : 'text-blue-400'
              )}
            >
              {diag.severity === 'error' ? <XCircle className="w-3 h-3" /> :
               diag.severity === 'warning' ? <AlertTriangle className="w-3 h-3" /> :
               <Info className="w-3 h-3" />}
              <span>Line {diag.range.start.line}: {diag.message}</span>
            </div>
          ))}
        </div>
      )}

      {/* Editor Content */}
      <div className="flex-1 flex overflow-hidden">
        {/* Line Numbers */}
        {showLineNumbers && (
          <div className="flex-shrink-0 py-3 px-2 bg-aviation-bg-secondary text-right border-r border-aviation-border-panel select-none">
            {lines.map((_, idx) => {
              const lineNum = idx + 1;
              const hasDiagnostic = diagnostics.some(d => d.range.start.line === lineNum);
              return (
                <div
                  key={idx}
                  className={cn(
                    'px-1 text-xs leading-6 font-mono',
                    activeLine === lineNum ? 'text-aviation-cyan font-bold' : 'text-aviation-text-muted',
                    hasDiagnostic && 'text-red-400 bg-red-900/20'
                  )}
                >
                  {lineNum}
                </div>
              );
            })}
          </div>
        )}

        {/* Code Area */}
        <div className="flex-1 relative overflow-auto">
          <textarea
            ref={textareaRef}
            value={value}
            onChange={(e) => onChange?.(e.target.value)}
            onSelect={handleCursorMove}
            onKeyDown={handleKeyDown}
            onClick={handleCursorMove}
            readOnly={readOnly}
            className={cn(
              'w-full h-full p-3 bg-transparent text-aviation-text-primary font-mono text-sm leading-6',
              'resize-none outline-none border-0 focus:ring-0',
              'placeholder:text-aviation-text-muted',
              readOnly && 'opacity-75 cursor-not-allowed'
            )}
            placeholder={`// ${language} code editor`}
            spellCheck={false}
          />
        </div>
      </div>

      {/* Status Bar */}
      <div className="flex items-center justify-between px-3 py-1 bg-aviation-bg-secondary border-t border-aviation-border-panel text-xs text-aviation-text-muted">
        <div className="flex items-center gap-4">
          <span className="flex items-center gap-1">
            <FileCode className="w-3 h-3" />
            {language.toUpperCase()}
          </span>
          <span>Ln {activeLine}, Col {activeColumn}</span>
          {symbols.length > 0 && (
            <span className="flex items-center gap-1">
              <Box className="w-3 h-3" />
              {symbols.length} symbols
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          {diagnostics.length > 0 && (
            <span className={cn(
              'flex items-center gap-1',
              diagnostics.some(d => d.severity === 'error') ? 'text-red-400' : 'text-amber-400'
            )}>
              <AlertTriangle className="w-3 h-3" />
              {diagnostics.length} issues
            </span>
          )}
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// AST Explorer
// ============================================================================

interface ASTNodeData {
  id: string;
  type: string;
  loc: { start: { line: number; column: number }; end: { line: number; column: number } };
  children?: ASTNodeData[];
  value?: unknown;
  metadata?: Record<string, unknown>;
  parent?: string;
  depth?: number;
}

interface ASTExplorerProps {
  ast: ASTNodeData | null;
  rootId?: string;
  selectedNodeId?: string | null;
  highlightedNodes?: string[];
  searchQuery?: string;
  onNodeSelect?: (node: ASTNodeData) => void;
  onNodeExpand?: (nodeId: string) => void;
  onNodeCollapse?: (nodeId: string) => void;
  onSearch?: (query: string) => void;
  showRange?: boolean;
  showMetadata?: boolean;
  className?: string;
}

interface TreeNodeProps {
  node: ASTNodeData;
  selectedId: string | null;
  highlightedIds: string[];
  depth: number;
  expandedNodes: Set<string>;
  showRange: boolean;
  showMetadata: boolean;
  onToggle: (id: string) => void;
  onSelect: (node: ASTNodeData) => void;
}

const TreeNode: React.FC<TreeNodeProps> = ({
  node,
  selectedId,
  highlightedIds,
  depth,
  expandedNodes,
  showRange,
  showMetadata,
  onToggle,
  onSelect,
}) => {
  const hasChildren = node.children && node.children.length > 0;
  const isExpanded = expandedNodes.has(node.id);
  const isSelected = selectedId === node.id;
  const isHighlighted = highlightedIds.includes(node.id);
  const isMatch = node.type.toLowerCase().includes((''));

  return (
    <div className="select-none">
      <div
        className={cn(
          'flex items-center gap-1 px-2 py-1 hover:bg-aviation-bg-instrument cursor-pointer transition-colors text-xs',
          isSelected && 'bg-aviation-amber/20 text-aviation-amber',
          isHighlighted && !isSelected && 'bg-cyan-500/20',
          isMatch && 'text-green-400'
        )}
        style={{ paddingLeft: `${depth * 16 + 8}px` }}
        onClick={() => onSelect(node)}
      >
        {hasChildren ? (
          <button
            onClick={(e) => { e.stopPropagation(); onToggle(node.id); }}
            className="p-0.5 hover:bg-aviation-bg-tertiary rounded"
          >
            {isExpanded ? <ChevronDown className="w-3 h-3" /> : <ChevronRight className="w-3 h-3" />}
          </button>
        ) : (
          <Circle className="w-3 h-3 text-aviation-text-muted" />
        )}
        <span className={cn(
          'px-1.5 py-0.5 rounded text-xs font-mono',
          node.type === 'Identifier' ? 'bg-blue-500/20 text-blue-400' :
          node.type === 'Literal' ? 'bg-green-500/20 text-green-400' :
          node.type === 'FunctionDeclaration' ? 'bg-purple-500/20 text-purple-400' :
          node.type === 'VariableDeclaration' ? 'bg-amber-500/20 text-amber-400' :
          'bg-gray-500/20 text-gray-400'
        )}>
          {node.type}
        </span>
        {node.value !== undefined && (
          <span className="text-aviation-text-secondary truncate max-w-[200px]">
            {String(node.value).substring(0, 30)}
          </span>
        )}
        {showRange && (
          <span className="text-aviation-text-muted text-[10px]">
            L{node.loc.start.line}:{node.loc.start.column}
          </span>
        )}
      </div>
      {hasChildren && isExpanded && (
        <div>
          {node.children!.map((child) => (
            <TreeNode
              key={child.id}
              node={child}
              selectedId={selectedId}
              highlightedIds={highlightedIds}
              depth={depth + 1}
              expandedNodes={expandedNodes}
              showRange={showRange}
              showMetadata={showMetadata}
              onToggle={onToggle}
              onSelect={onSelect}
            />
          ))}
        </div>
      )}
    </div>
  );
};

export const ASTExplorer: React.FC<ASTExplorerProps> = ({
  ast,
  rootId,
  selectedNodeId,
  highlightedNodes = [],
  searchQuery = '',
  onNodeSelect,
  onNodeExpand,
  onNodeCollapse,
  onSearch,
  showRange = true,
  showMetadata = false,
  className,
}) => {
  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(new Set());
  const [internalSearch, setInternalSearch] = useState(searchQuery);

  const toggleNode = useCallback((id: string) => {
    setExpandedNodes((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
        onNodeCollapse?.(id);
      } else {
        next.add(id);
        onNodeExpand?.(id);
      }
      return next;
    });
  }, [onNodeExpand, onNodeCollapse]);

  const handleSelect = useCallback((node: ASTNodeData) => {
    onNodeSelect?.(node);
  }, [onNodeSelect]);

  if (!ast) {
    return (
      <div className={cn('flex flex-col h-full items-center justify-center bg-aviation-bg-panel rounded-lg border border-aviation-border-panel p-6', className)}>
        <Braces className="w-12 h-12 text-aviation-text-muted mb-3" />
        <p className="text-aviation-text-muted text-sm">No AST available</p>
        <p className="text-aviation-text-dim text-xs mt-1">Parse code to see its Abstract Syntax Tree</p>
      </div>
    );
  }

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Search Bar */}
      <div className="flex items-center gap-2 px-3 py-2 border-b border-aviation-border-panel">
        <div className="relative flex-1">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-4 h-4 text-aviation-text-muted" />
          <input
            type="text"
            value={internalSearch}
            onChange={(e) => setInternalSearch(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && onSearch?.(internalSearch)}
            placeholder="Search nodes by type..."
            className="w-full pl-8 pr-3 py-1.5 bg-aviation-bg-secondary border border-aviation-border-panel rounded text-sm text-aviation-text-primary placeholder:text-aviation-text-muted focus:outline-none focus:border-aviation-cyan"
          />
        </div>
        <span className="text-xs text-aviation-text-muted">{ast.type}</span>
      </div>

      {/* Tree View */}
      <div className="flex-1 overflow-auto py-2">
        <TreeNode
          node={ast}
          selectedId={selectedNodeId || ''}
          highlightedIds={highlightedNodes}
          depth={0}
          expandedNodes={expandedNodes}
          showRange={showRange}
          showMetadata={showMetadata}
          onToggle={toggleNode}
          onSelect={handleSelect}
        />
      </div>

      {/* Node Details */}
      {selectedNodeId && (
        <div className="px-3 py-2 border-t border-aviation-border-panel bg-aviation-bg-secondary">
          <div className="text-xs text-aviation-text-muted">
            Selected: <span className="text-aviation-cyan">{selectedNodeId}</span>
          </div>
        </div>
      )}
    </div>
  );
};

// ============================================================================
// Dependency Heatmap
// ============================================================================

interface DependencyNodeData {
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

interface DependencyEdgeData {
  source: string;
  target: string;
  weight?: number;
  type: 'import' | 'inheritance' | 'composition' | 'dependency';
}

interface DependencyHeatmapProps {
  nodes: DependencyNodeData[];
  edges: DependencyEdgeData[];
  selectedNodeId?: string | null;
  highlightedNodes?: string[];
  colorMetric?: 'complexity' | 'coupling' | 'changes' | 'bugs' | 'maintenance';
  onNodeClick?: (node: DependencyNodeData) => void;
  onNodeHover?: (node: DependencyNodeData | null) => void;
  className?: string;
}

export const DependencyHeatmap: React.FC<DependencyHeatmapProps> = ({
  nodes,
  edges,
  selectedNodeId,
  highlightedNodes = [],
  colorMetric = 'complexity',
  onNodeClick,
  onNodeHover,
  className,
}) => {
  const getNodeColor = useCallback((node: DependencyNodeData) => {
    const metric = node.metrics?.[colorMetric] || 0;
    if (metric > 0.7) return 'bg-red-500';
    if (metric > 0.4) return 'bg-amber-500';
    return 'bg-green-500';
  }, [colorMetric]);

  const getMetricValue = useCallback((node: DependencyNodeData) => {
    return node.metrics?.[colorMetric] || 0;
  }, [colorMetric]);

  const connectedNodes = useMemo(() => {
    const connected = new Set<string>();
    edges.forEach((edge) => {
      connected.add(edge.source);
      connected.add(edge.target);
    });
    return connected;
  }, [edges]);

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-aviation-border-panel">
        <div className="flex items-center gap-2">
          <Network className="w-4 h-4 text-aviation-cyan" />
          <span className="text-sm font-medium">Dependency Heatmap</span>
        </div>
        <div className="flex items-center gap-3 text-xs text-aviation-text-muted">
          <span className="flex items-center gap-1">
            <div className="w-2 h-2 rounded-full bg-green-500" /> Low
          </span>
          <span className="flex items-center gap-1">
            <div className="w-2 h-2 rounded-full bg-amber-500" /> Medium
          </span>
          <span className="flex items-center gap-1">
            <div className="w-2 h-2 rounded-full bg-red-500" /> High
          </span>
        </div>
      </div>

      {/* Heatmap Grid */}
      <div className="flex-1 overflow-auto p-4">
        <div className="grid grid-cols-[repeat(auto-fill,minmax(120px,1fr))] gap-3">
          {nodes.map((node) => {
            const isConnected = connectedNodes.has(node.id);
            const isSelected = selectedNodeId === node.id;
            const isHighlighted = highlightedNodes.includes(node.id);
            
            return (
              <div
                key={node.id}
                className={cn(
                  'relative flex flex-col p-3 rounded-lg border cursor-pointer transition-all',
                  isSelected ? 'border-aviation-amber bg-aviation-amber/10' :
                  isHighlighted ? 'border-cyan-500 bg-cyan-500/10' :
                  'border-aviation-border-panel bg-aviation-bg-secondary hover:border-aviation-cyan',
                  !isConnected && 'opacity-50'
                )}
                onClick={() => onNodeClick?.(node)}
                onMouseEnter={() => onNodeHover?.(node)}
                onMouseLeave={() => onNodeHover?.(null)}
              >
                {/* Color Indicator Bar */}
                <div className={cn('absolute top-0 left-0 right-0 h-1 rounded-t-lg', getNodeColor(node))} />
                
                <div className="flex items-center gap-2 mt-2">
                  {node.type === 'package' ? <Box className="w-3 h-3 text-aviation-cyan" /> :
                   node.type === 'module' ? <Layers className="w-3 h-3 text-purple-400" /> :
                   <FileCode className="w-3 h-3 text-gray-400" />}
                  <span className="text-xs font-medium truncate">{node.name}</span>
                </div>
                
                <div className="mt-2 text-[10px] text-aviation-text-muted truncate">
                  {node.path}
                </div>
                
                {/* Metric Value */}
                <div className="mt-auto pt-2 flex items-center justify-between">
                  <span className="text-[10px] text-aviation-text-muted capitalize">{colorMetric}</span>
                  <span className="text-xs font-mono text-aviation-text-primary">
                    {getMetricValue(node).toFixed(2)}
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Edge Legend */}
      <div className="px-4 py-2 border-t border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="flex items-center gap-4 text-xs text-aviation-text-muted">
          <span className="flex items-center gap-1">
            <div className="w-6 h-0.5 bg-aviation-cyan" /> Import
          </span>
          <span className="flex items-center gap-1">
            <div className="w-6 h-0.5 bg-purple-500" /> Inheritance
          </span>
          <span className="flex items-center gap-1">
            <div className="w-6 h-0.5 bg-amber-500" /> Composition
          </span>
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Multi File Diff Viewer
// ============================================================================

interface DiffFileData {
  id: string;
  path: string;
  oldContent: string;
  newContent: string;
  language?: string;
  status?: 'added' | 'deleted' | 'modified' | 'renamed';
}

interface DiffHunkData {
  id: string;
  oldStart: number;
  oldCount: number;
  newStart: number;
  newCount: number;
  lines: Array<{
    id: string;
    type: 'add' | 'delete' | 'context' | 'header';
    content: string;
    oldLineNumber?: number;
    newLineNumber?: number;
  }>;
}

interface MultiFileDiffViewerProps {
  files: DiffFileData[];
  selectedFileId?: string | null;
  hunks?: DiffHunkData[];
  showLineNumbers?: boolean;
  showComments?: boolean;
  showSyntaxHighlighting?: boolean;
  diffContext?: number;
  onFileSelect?: (file: DiffFileData) => void;
  onCommentAdd?: (lineId: string, comment: string) => void;
  onAcceptChange?: (fileId: string, changeId: string) => void;
  onRejectChange?: (fileId: string, changeId: string) => void;
  className?: string;
}

export const MultiFileDiffViewer: React.FC<MultiFileDiffViewerProps> = ({
  files,
  selectedFileId,
  hunks = [],
  showLineNumbers = true,
  showComments = false,
  showSyntaxHighlighting = true,
  diffContext = 3,
  onFileSelect,
  onCommentAdd,
  onAcceptChange,
  onRejectChange,
  className,
}) => {
  const selectedFile = files.find((f) => f.id === selectedFileId) || files[0];

  const getStatusColor = (status?: string) => {
    switch (status) {
      case 'added': return 'text-green-400 bg-green-500/20';
      case 'deleted': return 'text-red-400 bg-red-500/20';
      case 'modified': return 'text-amber-400 bg-amber-500/20';
      case 'renamed': return 'text-blue-400 bg-blue-500/20';
      default: return 'text-gray-400 bg-gray-500/20';
    }
  };

  const renderDiffLines = (file: DiffFileData) => {
    const oldLines = file.oldContent.split('\n');
    const newLines = file.newContent.split('\n');
    const maxLines = Math.max(oldLines.length, newLines.length);
    const lines: DiffHunkData['lines'] = [];

    for (let i = 0; i < maxLines; i++) {
      const oldLine = oldLines[i];
      const newLine = newLines[i];
      
      if (oldLine !== newLine) {
        if (oldLine !== undefined) {
          lines.push({
            id: `${file.id}-old-${i}`,
            type: 'delete',
            content: oldLine,
            oldLineNumber: i + 1,
          });
        }
        if (newLine !== undefined) {
          lines.push({
            id: `${file.id}-new-${i}`,
            type: 'add',
            content: newLine,
            newLineNumber: i + 1,
          });
        }
      } else if (oldLine !== undefined) {
        lines.push({
          id: `${file.id}-ctx-${i}`,
          type: 'context',
          content: oldLine,
          oldLineNumber: i + 1,
          newLineNumber: i + 1,
        });
      }
    }
    
    return lines;
  };

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* File Tabs */}
      <div className="flex items-center gap-1 px-2 py-1 border-b border-aviation-border-panel overflow-x-auto">
        {files.map((file) => {
          const isSelected = file.id === selectedFile?.id;
          return (
            <button
              key={file.id}
              onClick={() => onFileSelect?.(file)}
              className={cn(
                'flex items-center gap-2 px-3 py-1.5 rounded text-xs font-medium transition-colors whitespace-nowrap',
                isSelected ? 'bg-aviation-bg-instrument text-aviation-text-primary' : 'text-aviation-text-muted hover:text-aviation-text-primary'
              )}
            >
              <FileDiff className={cn('w-3 h-3', getStatusColor(file.status))} />
              <span className="truncate max-w-[150px]">{file.path.split('/').pop()}</span>
              {file.status && (
                <span className={cn('px-1 py-0.5 rounded text-[10px]', getStatusColor(file.status))}>
                  {file.status}
                </span>
              )}
            </button>
          );
        })}
      </div>

      {/* Diff Content */}
      {selectedFile && (
        <div className="flex-1 overflow-auto">
          {/* File Header */}
          <div className="sticky top-0 flex items-center gap-3 px-4 py-2 bg-aviation-bg-secondary border-b border-aviation-border-panel">
            <span className="text-sm font-medium truncate">{selectedFile.path}</span>
            <span className="text-xs text-aviation-text-muted">{selectedFile.language}</span>
          </div>

          {/* Diff Lines */}
          <div className="font-mono text-xs">
            {renderDiffLines(selectedFile).map((line) => (
              <div
                key={line.id}
                className={cn(
                  'flex items-start',
                  line.type === 'add' && 'bg-green-900/30',
                  line.type === 'delete' && 'bg-red-900/30',
                  line.type === 'context' && 'hover:bg-aviation-bg-secondary'
                )}
              >
                {showLineNumbers && (
                  <div className="flex items-center">
                    <span className="w-12 py-0.5 text-right pr-2 text-aviation-text-muted select-none border-r border-aviation-border-panel">
                      {line.oldLineNumber || ''}
                    </span>
                    <span className="w-12 py-0.5 text-right pr-2 text-aviation-text-muted select-none border-r border-aviation-border-panel">
                      {line.newLineNumber || ''}
                    </span>
                  </div>
                )}
                <div className={cn(
                  'w-6 py-0.5 text-center select-none',
                  line.type === 'add' && 'text-green-400' ,
                  line.type === 'delete' && 'text-red-400',
                  line.type === 'context' && 'text-aviation-text-muted'
                )}>
                  {line.type === 'add' ? '+' : line.type === 'delete' ? '-' : ' '}
                </div>
                <div className="flex-1 px-2 py-0.5 text-aviation-text-primary whitespace-pre-wrap break-all">
                  {line.content}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Summary Bar */}
      <div className="flex items-center justify-between px-4 py-2 border-t border-aviation-border-panel bg-aviation-bg-secondary text-xs">
        <div className="flex items-center gap-4 text-aviation-text-muted">
          <span className="flex items-center gap-1">
            <div className="w-3 h-3 rounded bg-red-500/30 border border-red-500/50" />
            {files.reduce((acc, f) => acc + f.oldContent.split('\n').length, 0)} deletions
          </span>
          <span className="flex items-center gap-1">
            <div className="w-3 h-3 rounded bg-green-500/30 border border-green-500/50" />
            {files.reduce((acc, f) => acc + f.newContent.split('\n').length, 0)} additions
          </span>
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Architecture Map
// ============================================================================

interface ArchitectureNodeData {
  id: string;
  name: string;
  type: 'module' | 'package' | 'component' | 'service' | 'layer';
  path?: string;
  children?: ArchitectureNodeData[];
  metrics?: {
    components?: number;
    linesOfCode?: number;
    complexity?: number;
    stability?: number;
  };
}

interface ArchitectureConnectionData {
  id: string;
  source: string;
  target: string;
  type: 'depends-on' | 'composes' | 'extends' | 'implements' | 'uses';
  weight?: number;
}

interface ArchitectureMapProps {
  nodes: ArchitectureNodeData[];
  connections: ArchitectureConnectionData[];
  selectedNodeId?: string | null;
  expandedNodes?: string[];
  showMetrics?: boolean;
  layout?: 'tree' | 'force' | 'circular';
  onNodeSelect?: (node: ArchitectureNodeData) => void;
  onNodeExpand?: (nodeId: string) => void;
  onNodeCollapse?: (nodeId: string) => void;
  className?: string;
}

interface ArchTreeNodeProps {
  node: ArchitectureNodeData;
  selectedId: string | null;
  expandedIds: string[];
  showMetrics: boolean;
  depth: number;
  onToggle: (id: string) => void;
  onSelect: (node: ArchitectureNodeData) => void;
}

const ArchTreeNode: React.FC<ArchTreeNodeProps> = ({
  node,
  selectedId,
  expandedIds,
  showMetrics,
  depth,
  onToggle,
  onSelect,
}) => {
  const hasChildren = node.children && node.children.length > 0;
  const isExpanded = expandedIds.includes(node.id);
  const isSelected = selectedId === node.id;

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'layer': return <LayersIcon className="w-4 h-4 text-purple-400" />;
      case 'module': return <Box className="w-4 h-4 text-cyan-400" />;
      case 'package': return <Package className="w-4 h-4 text-amber-400" />;
      case 'component': return <Component className="w-4 h-4 text-green-400" />;
      case 'service': return <Cpu className="w-4 h-4 text-red-400" />;
      default: return <FileCode className="w-4 h-4 text-gray-400" />;
    }
  };

  return (
    <div>
      <div
        className={cn(
          'flex items-center gap-2 px-3 py-2 hover:bg-aviation-bg-instrument cursor-pointer transition-colors',
          isSelected && 'bg-aviation-amber/20 border-l-2 border-aviation-amber'
        )}
        style={{ paddingLeft: `${depth * 20 + 12}px` }}
        onClick={() => onSelect(node)}
      >
        {hasChildren ? (
          <button onClick={(e) => { e.stopPropagation(); onToggle(node.id); }} className="p-0.5">
            {isExpanded ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}
          </button>
        ) : (
          <div className="w-5" />
        )}
        {getTypeIcon(node.type)}
        <div className="flex-1 min-w-0">
          <div className="text-sm font-medium truncate">{node.name}</div>
          {showMetrics && node.metrics && (
            <div className="flex items-center gap-3 text-[10px] text-aviation-text-muted mt-0.5">
              {node.metrics.components !== undefined && <span>{node.metrics.components} components</span>}
              {node.metrics.linesOfCode !== undefined && <span>{node.metrics.linesOfCode} LOC</span>}
              {node.metrics.complexity !== undefined && <span>CX: {node.metrics.complexity}</span>}
            </div>
          )}
        </div>
      </div>
      {hasChildren && isExpanded && (
        <div>
          {node.children!.map((child) => (
            <ArchTreeNode
              key={child.id}
              node={child}
              selectedId={selectedId}
              expandedIds={expandedIds}
              showMetrics={showMetrics}
              depth={depth + 1}
              onToggle={onToggle}
              onSelect={onSelect}
            />
          ))}
        </div>
      )}
    </div>
  );
};

const Package: React.FC<{ className?: string }> = ({ className }) => <Box className={className} />;
const Component: React.FC<{ className?: string }> = ({ className }) => <LayoutIcon className={className} />;

export const ArchitectureMap: React.FC<ArchitectureMapProps> = ({
  nodes,
  connections,
  selectedNodeId,
  expandedNodes = [],
  showMetrics = true,
  layout = 'tree',
  onNodeSelect,
  onNodeExpand,
  onNodeCollapse,
  className,
}) => {
  const [internalExpanded, setInternalExpanded] = useState<Set<string>>(new Set(expandedNodes));

  const toggleNode = useCallback((id: string) => {
    setInternalExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
        onNodeCollapse?.(id);
      } else {
        next.add(id);
        onNodeExpand?.(id);
      }
      return next;
    });
  }, [onNodeExpand, onNodeCollapse]);

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center gap-2">
          <LayoutIcon className="w-5 h-5 text-aviation-cyan" />
          <span className="text-sm font-medium">Architecture Map</span>
        </div>
        <div className="flex items-center gap-2">
          {(['tree', 'force', 'circular'] as const).map((l) => (
            <button
              key={l}
              onClick={() => {}}
              className={cn(
                'px-2 py-1 text-xs rounded transition-colors capitalize',
                layout === l ? 'bg-aviation-cyan text-black' : 'text-aviation-text-muted hover:bg-aviation-bg-instrument'
              )}
            >
              {l}
            </button>
          ))}
        </div>
      </div>

      {/* Tree View */}
      <div className="flex-1 overflow-auto">
        {nodes.map((node) => (
          <ArchTreeNode
            key={node.id}
            node={node}
            selectedId={selectedNodeId || ''}
            expandedIds={Array.from(internalExpanded)}
            showMetrics={showMetrics}
            depth={0}
            onToggle={toggleNode}
            onSelect={(n) => onNodeSelect?.(n)}
          />
        ))}
      </div>

      {/* Connection Legend */}
      <div className="px-4 py-2 border-t border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="flex items-center gap-4 text-xs text-aviation-text-muted">
          <span className="flex items-center gap-1"><ArrowRight className="w-3 h-3" /> depends-on</span>
          <span className="flex items-center gap-1"><LayersIcon className="w-3 h-3" /> composes</span>
          <span className="flex items-center gap-1"><GitFork className="w-3 h-3" /> extends</span>
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Smart Refactor Panel
// ============================================================================

interface RefactorOpportunityData {
  id: string;
  type: string;
  title: string;
  description: string;
  location: { start: { line: number }; end: { line: number } };
  original: string;
  preview: string;
  impact: 'low' | 'medium' | 'high';
  estimatedComplexity?: number;
  affectedFiles?: string[];
  automated?: boolean;
}

interface RefactorPreviewData {
  fileId: string;
  originalContent: string;
  refactoredContent: string;
}

interface SmartRefactorPanelProps {
  opportunities: RefactorOpportunityData[];
  selectedOpportunityId?: string | null;
  previews?: RefactorPreviewData[];
  onOpportunitySelect?: (opportunity: RefactorOpportunityData) => void;
  onPreviewGenerated?: (preview: RefactorPreviewData) => void;
  onApply?: (opportunityId: string) => void;
  onReject?: (opportunityId: string) => void;
  className?: string;
}

export const SmartRefactorPanel: React.FC<SmartRefactorPanelProps> = ({
  opportunities,
  selectedOpportunityId,
  previews = [],
  onOpportunitySelect,
  onPreviewGenerated,
  onApply,
  onReject,
  className,
}) => {
  const selectedOpportunity = opportunities.find((o) => o.id === selectedOpportunityId);

  const getImpactColor = (impact: string) => {
    switch (impact) {
      case 'high': return 'text-red-400 bg-red-500/20';
      case 'medium': return 'text-amber-400 bg-amber-500/20';
      case 'low': return 'text-green-400 bg-green-500/20';
      default: return 'text-gray-400 bg-gray-500/20';
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'extract-method': return <GitFork className="w-4 h-4 text-cyan-400" />;
      case 'rename': return <Type className="w-4 h-4 text-purple-400" />;
      case 'inline-method': return <ArrowRight className="w-4 h-4 text-amber-400" />;
      default: return <Hammer className="w-4 h-4 text-gray-400" />;
    }
  };

  return (
    <div className={cn('flex h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Opportunities List */}
      <div className="w-1/2 flex flex-col border-r border-aviation-border-panel">
        <div className="px-4 py-3 border-b border-aviation-border-panel">
          <div className="flex items-center gap-2">
            <Wand2 className="w-4 h-4 text-aviation-cyan" />
            <span className="text-sm font-medium">Refactor Opportunities</span>
            <span className="text-xs text-aviation-text-muted ml-auto">{opportunities.length}</span>
          </div>
        </div>
        <div className="flex-1 overflow-auto">
          {opportunities.map((opp) => {
            const isSelected = opp.id === selectedOpportunityId;
            return (
              <div
                key={opp.id}
                className={cn(
                  'px-4 py-3 border-b border-aviation-border-panel cursor-pointer transition-colors',
                  isSelected ? 'bg-aviation-amber/10' : 'hover:bg-aviation-bg-secondary'
                )}
                onClick={() => onOpportunitySelect?.(opp)}
              >
                <div className="flex items-start gap-3">
                  {getTypeIcon(opp.type)}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium">{opp.title}</span>
                      <span className={cn('px-1.5 py-0.5 rounded text-[10px]', getImpactColor(opp.impact))}>
                        {opp.impact}
                      </span>
                    </div>
                    <p className="text-xs text-aviation-text-muted mt-1 line-clamp-2">{opp.description}</p>
                    <div className="flex items-center gap-3 mt-2 text-[10px] text-aviation-text-dim">
                      <span>Line {opp.location.start.line}</span>
                      {opp.affectedFiles && <span>{opp.affectedFiles.length} files</span>}
                      {opp.automated && (
                        <span className="flex items-center gap-1 text-green-400">
                          <Sparkles className="w-3 h-3" /> automated
                        </span>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Preview Panel */}
      <div className="w-1/2 flex flex-col">
        <div className="px-4 py-3 border-b border-aviation-border-panel">
          <div className="flex items-center gap-2">
            <Eye className="w-4 h-4 text-aviation-cyan" />
            <span className="text-sm font-medium">Preview</span>
          </div>
        </div>
        {selectedOpportunity ? (
          <div className="flex-1 flex flex-col overflow-hidden">
            <div className="flex-1 overflow-auto p-4">
              <div className="text-xs text-aviation-text-muted mb-2">Before</div>
              <pre className="p-3 bg-red-950/30 rounded border border-red-900/50 text-xs font-mono text-aviation-text-primary overflow-x-auto">
                {selectedOpportunity.original}
              </pre>
              <div className="flex items-center justify-center my-3">
                <ArrowDownRight className="w-5 h-5 text-aviation-cyan" />
              </div>
              <div className="text-xs text-aviation-text-muted mb-2">After</div>
              <pre className="p-3 bg-green-950/30 rounded border border-green-900/50 text-xs font-mono text-aviation-text-primary overflow-x-auto">
                {selectedOpportunity.preview}
              </pre>
            </div>
            <div className="flex items-center justify-end gap-2 px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
              <button
                onClick={() => onReject?.(selectedOpportunity.id)}
                className="px-3 py-1.5 text-xs text-aviation-text-muted hover:text-aviation-text-primary transition-colors"
              >
                Reject
              </button>
              <button
                onClick={() => onApply?.(selectedOpportunity.id)}
                className="px-3 py-1.5 text-xs bg-aviation-cyan text-black rounded hover:bg-aviation-cyan/80 transition-colors flex items-center gap-1"
              >
                <Check className="w-3 h-3" /> Apply
              </button>
            </div>
          </div>
        ) : (
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center">
              <Lightbulb className="w-8 h-8 text-aviation-text-muted mx-auto mb-2" />
              <p className="text-sm text-aviation-text-muted">Select an opportunity to preview</p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

// ============================================================================
// Code Generation Preview
// ============================================================================

interface GeneratedCodeData {
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
  dependencies?: Array<{ name: string; version: string }>;
}

interface CodeGenerationPreviewProps {
  generation: GeneratedCodeData | null;
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

export const CodeGenerationPreview: React.FC<CodeGenerationPreviewProps> = ({
  generation,
  loading = false,
  error = null,
  onCopy,
  onApply,
  onDiscard,
  onRegenerate,
  onToggleExplanation,
  showExplanation = false,
  className,
}) => {
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(() => {
    if (generation?.code) {
      navigator.clipboard.writeText(generation.code);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
      onCopy?.(generation.code);
    }
  }, [generation, onCopy]);

  if (loading) {
    return (
      <div className={cn('flex flex-col h-full items-center justify-center bg-aviation-bg-panel rounded-lg border border-aviation-border-panel', className)}>
        <Sparkles className="w-8 h-8 text-aviation-cyan animate-pulse mb-3" />
        <p className="text-sm text-aviation-text-muted">Generating code...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className={cn('flex flex-col h-full items-center justify-center bg-aviation-bg-panel rounded-lg border border-red-900/50', className)}>
        <AlertOctagon className="w-8 h-8 text-red-400 mb-3" />
        <p className="text-sm text-red-400">{error}</p>
      </div>
    );
  }

  if (!generation) {
    return (
      <div className={cn('flex flex-col h-full items-center justify-center bg-aviation-bg-panel rounded-lg border border-aviation-border-panel', className)}>
        <Bot className="w-8 h-8 text-aviation-text-muted mb-3" />
        <p className="text-sm text-aviation-text-muted">No code generated yet</p>
        <p className="text-xs text-aviation-text-dim mt-1">AI will generate code based on your context</p>
      </div>
    );
  }

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Bot className="w-5 h-5 text-aviation-cyan" />
            <div>
              <div className="text-sm font-medium">{generation.title || 'Generated Code'}</div>
              <div className="text-xs text-aviation-text-muted flex items-center gap-2 mt-0.5">
                <span>{generation.language}</span>
                {generation.context?.framework && (
                  <>
                    <span>·</span>
                    <span>{generation.context.framework}</span>
                  </>
                )}
              </div>
            </div>
          </div>
          <div className="flex items-center gap-1">
            <button
              onClick={onToggleExplanation}
              className={cn(
                'p-2 rounded transition-colors',
                showExplanation ? 'bg-aviation-cyan/20 text-aviation-cyan' : 'text-aviation-text-muted hover:bg-aviation-bg-instrument'
              )}
            >
              <MessageCircle className="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>

      {/* Explanation */}
      {showExplanation && generation.description && (
        <div className="px-4 py-3 border-b border-aviation-border-panel bg-aviation-bg-secondary">
          <p className="text-xs text-aviation-text-primary leading-relaxed">{generation.description}</p>
        </div>
      )}

      {/* Metrics */}
      {generation.metrics && (
        <div className="flex items-center gap-4 px-4 py-2 border-b border-aviation-border-panel bg-aviation-bg-secondary">
          {generation.metrics.complexity !== undefined && (
            <div className="flex items-center gap-1 text-xs">
              <span className="text-aviation-text-muted">Complexity:</span>
              <span className={cn(
                generation.metrics.complexity > 0.7 ? 'text-red-400' :
                generation.metrics.complexity > 0.4 ? 'text-amber-400' : 'text-green-400'
              )}>
                {generation.metrics.complexity.toFixed(2)}
              </span>
            </div>
          )}
          {generation.metrics.maintainability !== undefined && (
            <div className="flex items-center gap-1 text-xs">
              <span className="text-aviation-text-muted">Maintainability:</span>
              <span className="text-green-400">{generation.metrics.maintainability.toFixed(2)}</span>
            </div>
          )}
          {generation.metrics.estimatedTokens !== undefined && (
            <div className="flex items-center gap-1 text-xs">
              <span className="text-aviation-text-muted">Tokens:</span>
              <span className="text-cyan-400">{generation.metrics.estimatedTokens}</span>
            </div>
          )}
        </div>
      )}

      {/* Code */}
      <div className="flex-1 overflow-auto">
        <pre className="p-4 text-sm font-mono text-aviation-text-primary leading-relaxed whitespace-pre-wrap">
          <code>{generation.code}</code>
        </pre>
      </div>

      {/* Dependencies */}
      {generation.dependencies && generation.dependencies.length > 0 && (
        <div className="px-4 py-2 border-t border-aviation-border-panel bg-aviation-bg-secondary">
          <div className="text-xs text-aviation-text-muted mb-1">Dependencies</div>
          <div className="flex flex-wrap gap-2">
            {generation.dependencies.map((dep) => (
              <span key={dep.name} className="px-2 py-0.5 bg-aviation-bg-instrument rounded text-xs">
                {dep.name} <span className="text-aviation-text-dim">@{dep.version}</span>
              </span>
            ))}
          </div>
        </div>
      )}

      {/* Actions */}
      <div className="flex items-center justify-between px-4 py-3 border-t border-aviation-border-panel">
        <button
          onClick={handleCopy}
          className="flex items-center gap-1.5 px-3 py-1.5 text-xs text-aviation-text-muted hover:text-aviation-text-primary transition-colors"
        >
          {copied ? <Check className="w-3 h-3 text-green-400" /> : <Copy className="w-3 h-3" />}
          {copied ? 'Copied!' : 'Copy'}
        </button>
        <div className="flex items-center gap-2">
          <button
            onClick={onDiscard}
            className="px-3 py-1.5 text-xs text-aviation-text-muted hover:text-red-400 transition-colors"
          >
            Discard
          </button>
          <button
            onClick={onRegenerate}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs text-aviation-text-muted hover:text-aviation-text-primary transition-colors"
          >
            <RefreshCw className="w-3 h-3" /> Regenerate
          </button>
          <button
            onClick={onApply}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs bg-aviation-cyan text-black rounded hover:bg-aviation-cyan/80 transition-colors"
          >
            <Check className="w-3 h-3" /> Apply
          </button>
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Inline AI Assistant
// ============================================================================

interface AIInlineSuggestionData {
  id: string;
  type: 'completion' | 'refactor' | 'documentation' | 'test' | 'explanation';
  text: string;
  confidence: number;
  startPosition: { line: number; column: number };
  endPosition: { line: number; column: number };
  explanation?: string;
}

interface InlineAIAssistantProps {
  enabled?: boolean;
  suggestions?: AIInlineSuggestionData[];
  currentSuggestion?: AIInlineSuggestionData | null;
  onSuggestionAccept?: (suggestion: AIInlineSuggestionData) => void;
  onSuggestionReject?: (suggestion: AIInlineSuggestionData) => void;
  onSuggestionHover?: (suggestion: AIInlineSuggestionData | null) => void;
  onExplain?: (suggestion: AIInlineSuggestionData) => void;
  className?: string;
}

export const InlineAIAssistant: React.FC<InlineAIAssistantProps> = ({
  enabled = true,
  suggestions = [],
  currentSuggestion,
  onSuggestionAccept,
  onSuggestionReject,
  onSuggestionHover,
  onExplain,
  className,
}) => {
  const [hoveredSuggestion, setHoveredSuggestion] = useState<AIInlineSuggestionData | null>(null);

  const getSuggestionIcon = (type: string) => {
    switch (type) {
      case 'completion': return <Sparkles className="w-3 h-3" />;
      case 'refactor': return <Hammer className="w-3 h-3" />;
      case 'documentation': return <FileText className="w-3 h-3" />;
      case 'test': return <CheckSquare className="w-3 h-3" />;
      case 'explanation': return <MessageSquare className="w-3 h-3" />;
      default: return <Lightbulb className="w-3 h-3" />;
    }
  };

  const getSuggestionColor = (type: string) => {
    switch (type) {
      case 'completion': return 'text-cyan-400 bg-cyan-500/20';
      case 'refactor': return 'text-purple-400 bg-purple-500/20';
      case 'documentation': return 'text-green-400 bg-green-500/20';
      case 'test': return 'text-amber-400 bg-amber-500/20';
      case 'explanation': return 'text-blue-400 bg-blue-500/20';
      default: return 'text-gray-400 bg-gray-500/20';
    }
  };

  if (!enabled) return null;

  return (
    <div className={cn('flex flex-col bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Suggestion Chips */}
      <div className="flex flex-wrap gap-2 p-3">
        {suggestions.map((suggestion) => {
          const isHovered = hoveredSuggestion?.id === suggestion.id;
          const isActive = currentSuggestion?.id === suggestion.id;
          
          return (
            <div
              key={suggestion.id}
              className={cn(
                'group relative flex items-center gap-2 px-3 py-1.5 rounded-lg border cursor-pointer transition-all',
                isActive ? 'border-aviation-cyan bg-aviation-cyan/20' :
                isHovered ? 'border-cyan-500/50 bg-cyan-500/10' :
                'border-aviation-border-panel bg-aviation-bg-secondary hover:border-aviation-text-muted'
              )}
              onMouseEnter={() => {
                setHoveredSuggestion(suggestion);
                onSuggestionHover?.(suggestion);
              }}
              onMouseLeave={() => {
                setHoveredSuggestion(null);
                onSuggestionHover?.(null);
              }}
            >
              <span className={cn('flex items-center gap-1 text-xs', getSuggestionColor(suggestion.type))}>
                {getSuggestionIcon(suggestion.type)}
              </span>
              <code className="text-xs text-aviation-text-primary font-mono max-w-[200px] truncate">
                {suggestion.text}
              </code>
              <div className="flex items-center gap-0.5">
                <span className="text-[10px] text-aviation-text-muted">
                  {Math.round(suggestion.confidence * 100)}%
                </span>
                <button
                  onClick={(e) => { e.stopPropagation(); onSuggestionAccept?.(suggestion); }}
                  className="p-0.5 hover:bg-green-500/30 rounded text-green-400 opacity-0 group-hover:opacity-100 transition-opacity"
                >
                  <Check className="w-3 h-3" />
                </button>
                <button
                  onClick={(e) => { e.stopPropagation(); onSuggestionReject?.(suggestion); }}
                  className="p-0.5 hover:bg-red-500/30 rounded text-red-400 opacity-0 group-hover:opacity-100 transition-opacity"
                >
                  <XCircle className="w-3 h-3" />
                </button>
              </div>
            </div>
          );
        })}
      </div>

      {/* Tooltip for Hovered Suggestion */}
      {hoveredSuggestion && (
        <div className="px-3 pb-3">
          <div className="p-3 bg-aviation-bg-instrument rounded border border-aviation-border-panel">
            <div className="flex items-start gap-2">
              <Info className="w-4 h-4 text-aviation-cyan mt-0.5" />
              <div className="flex-1">
                <div className="text-xs text-aviation-text-primary leading-relaxed">
                  {hoveredSuggestion.explanation || 'No explanation available'}
                </div>
                <button
                  onClick={() => onExplain?.(hoveredSuggestion)}
                  className="mt-2 text-xs text-aviation-cyan hover:underline"
                >
                  Get detailed explanation
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

// ============================================================================
// Code Intent Explorer
// ============================================================================

interface CodeIntentData {
  id: string;
  type: 'feature' | 'bugfix' | 'refactor' | 'optimization' | 'documentation' | 'test' | 'security' | 'compliance';
  confidence: number;
  description: string;
  affectedCodeRanges: Array<{ start: { line: number }; end: { line: number } }>;
  affectedFiles: string[];
  reasoning: string;
  relatedIntents?: string[];
  extractedRequirements?: string[];
}

interface CodeIntentExplorerProps {
  intents: CodeIntentData[];
  selectedIntentId?: string | null;
  showReasoning?: boolean;
  onIntentSelect?: (intent: CodeIntentData) => void;
  onIntentExpand?: (intentId: string) => void;
  onRequirementExtract?: (intentId: string, requirements: string[]) => void;
  className?: string;
}

export const CodeIntentExplorer: React.FC<CodeIntentExplorerProps> = ({
  intents,
  selectedIntentId,
  showReasoning = false,
  onIntentSelect,
  onIntentExpand,
  onRequirementExtract,
  className,
}) => {
  const [expandedIntents, setExpandedIntents] = useState<Set<string>>(new Set());

  const toggleExpand = useCallback((id: string) => {
    setExpandedIntents((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }, []);

  const getIntentIcon = (type: string) => {
    switch (type) {
      case 'feature': return <Sparkles className="w-4 h-4 text-green-400" />;
      case 'bugfix': return <Bug className="w-4 h-4 text-red-400" />;
      case 'refactor': return <Hammer className="w-4 h-4 text-purple-400" />;
      case 'optimization': return <Zap className="w-4 h-4 text-amber-400" />;
      case 'documentation': return <FileText className="w-4 h-4 text-blue-400" />;
      case 'test': return <CheckSquare className="w-4 h-4 text-cyan-400" />;
      case 'security': return <Shield className="w-4 h-4 text-red-500" />;
      case 'compliance': return <AlertTriangle className="w-4 h-4 text-orange-400" />;
      default: return <Lightbulb className="w-4 h-4 text-gray-400" />;
    }
  };

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center gap-2">
          <Target className="w-5 h-5 text-aviation-cyan" />
          <span className="text-sm font-medium">Code Intent Explorer</span>
          <span className="text-xs text-aviation-text-muted ml-auto">{intents.length} intents</span>
        </div>
      </div>

      {/* Intent List */}
      <div className="flex-1 overflow-auto">
        {intents.map((intent) => {
          const isSelected = intent.id === selectedIntentId;
          const isExpanded = expandedIntents.has(intent.id);
          
          return (
            <div
              key={intent.id}
              className={cn(
                'border-b border-aviation-border-panel',
                isSelected && 'bg-aviation-amber/10'
              )}
            >
              <div
                className="px-4 py-3 cursor-pointer hover:bg-aviation-bg-secondary transition-colors"
                onClick={() => onIntentSelect?.(intent)}
              >
                <div className="flex items-start gap-3">
                  {getIntentIcon(intent.type)}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium capitalize">{intent.type}</span>
                      <span className="text-xs text-aviation-text-muted">
                        {Math.round(intent.confidence * 100)}% confidence
                      </span>
                    </div>
                    <p className="text-xs text-aviation-text-primary mt-1">{intent.description}</p>
                    <div className="flex items-center gap-2 mt-2 text-[10px] text-aviation-text-muted">
                      <span>{intent.affectedFiles.length} files</span>
                      <span>·</span>
                      <span>{intent.affectedCodeRanges.length} ranges</span>
                    </div>
                  </div>
                  <button
                    onClick={(e) => { e.stopPropagation(); toggleExpand(intent.id); }}
                    className="p-1 hover:bg-aviation-bg-instrument rounded"
                  >
                    {isExpanded ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}
                  </button>
                </div>
              </div>

              {/* Expanded Content */}
              {isExpanded && (
                <div className="px-4 pb-3">
                  {showReasoning && intent.reasoning && (
                    <div className="p-3 bg-aviation-bg-secondary rounded mb-3">
                      <div className="text-xs text-aviation-text-muted mb-1">AI Reasoning</div>
                      <p className="text-xs text-aviation-text-primary leading-relaxed">{intent.reasoning}</p>
                    </div>
                  )}
                  {intent.extractedRequirements && intent.extractedRequirements.length > 0 && (
                    <div className="p-3 bg-aviation-bg-secondary rounded">
                      <div className="text-xs text-aviation-text-muted mb-1">Extracted Requirements</div>
                      <ul className="text-xs text-aviation-text-primary space-y-1">
                        {intent.extractedRequirements.map((req, idx) => (
                          <li key={idx} className="flex items-start gap-1">
                            <ChevronRight className="w-3 h-3 mt-0.5" />
                            {req}
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
};

// ============================================================================
// Semantic Search Panel
// ============================================================================

interface SearchResultData {
  id: string;
  filePath: string;
  lineNumber: number;
  lineContent: string;
  matchedText: string;
  context: string[];
  score: number;
  matchType: 'exact' | 'fuzzy' | 'semantic';
}

interface SemanticSearchPanelProps {
  query: string;
  results: SearchResultData[];
  loading?: boolean;
  selectedResultId?: string | null;
  searchType?: 'text' | 'semantic' | 'symbol' | 'regex';
  filters?: {
    language?: string;
    fileType?: string;
    path?: string;
  };
  onQueryChange?: (query: string) => void;
  onSearch?: () => void;
  onResultSelect?: (result: SearchResultData) => void;
  onResultHover?: (result: SearchResultData | null) => void;
  className?: string;
}

export const SemanticSearchPanel: React.FC<SemanticSearchPanelProps> = ({
  query,
  results,
  loading = false,
  selectedResultId,
  searchType = 'semantic',
  filters = {},
  onQueryChange,
  onSearch,
  onResultSelect,
  onResultHover,
  className,
}) => {
  const getMatchTypeIcon = (type: string) => {
    switch (type) {
      case 'semantic': return <Sparkles className="w-3 h-3 text-cyan-400" />;
      case 'fuzzy': return <Search className="w-3 h-3 text-amber-400" />;
      default: return <SearchCode className="w-3 h-3 text-green-400" />;
    }
  };

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Search Input */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center gap-2">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-aviation-text-muted" />
            <input
              type="text"
              value={query}
              onChange={(e) => onQueryChange?.(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && onSearch?.()}
              placeholder="Search code semantically..."
              className="w-full pl-10 pr-4 py-2 bg-aviation-bg-secondary border border-aviation-border-panel rounded text-sm text-aviation-text-primary placeholder:text-aviation-text-muted focus:outline-none focus:border-aviation-cyan"
            />
          </div>
          <button
            onClick={onSearch}
            disabled={loading}
            className="px-4 py-2 bg-aviation-cyan text-black rounded text-sm font-medium hover:bg-aviation-cyan/80 transition-colors disabled:opacity-50"
          >
            {loading ? 'Searching...' : 'Search'}
          </button>
        </div>

        {/* Search Type Tabs */}
        <div className="flex items-center gap-1 mt-3">
          {(['text', 'semantic', 'symbol', 'regex'] as const).map((type) => (
            <button
              key={type}
              onClick={() => {}}
              className={cn(
                'px-3 py-1 text-xs rounded capitalize transition-colors',
                searchType === type ? 'bg-aviation-cyan/20 text-aviation-cyan' : 'text-aviation-text-muted hover:bg-aviation-bg-instrument'
              )}
            >
              {type}
            </button>
          ))}
        </div>
      </div>

      {/* Results */}
      <div className="flex-1 overflow-auto">
        {results.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center p-6">
            <SearchCode className="w-8 h-8 text-aviation-text-muted mb-3" />
            <p className="text-sm text-aviation-text-muted">No results found</p>
            <p className="text-xs text-aviation-text-dim mt-1">Try a different query or search type</p>
          </div>
        ) : (
          results.map((result) => {
            const isSelected = result.id === selectedResultId;
            return (
              <div
                key={result.id}
                className={cn(
                  'px-4 py-3 border-b border-aviation-border-panel cursor-pointer transition-colors',
                  isSelected ? 'bg-aviation-cyan/10' : 'hover:bg-aviation-bg-secondary'
                )}
                onClick={() => onResultSelect?.(result)}
                onMouseEnter={() => onResultHover?.(result)}
                onMouseLeave={() => onResultHover?.(null)}
              >
                <div className="flex items-start gap-2">
                  <FileSearch className="w-4 h-4 text-aviation-text-muted mt-0.5" />
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-medium text-aviation-text-primary truncate">
                        {result.filePath}
                      </span>
                      <span className="text-xs text-aviation-text-muted">:{result.lineNumber}</span>
                      {getMatchTypeIcon(result.matchType)}
                      <span className="text-xs text-aviation-text-dim ml-auto">
                        {Math.round(result.score * 100)}%
                      </span>
                    </div>
                    <p className="text-xs text-aviation-text-muted mt-1 line-clamp-2">
                      <span className="bg-yellow-500/30 text-yellow-200">{result.matchedText}</span>
                    </p>
                    <p className="text-xs text-aviation-text-dim mt-1 font-mono truncate">
                      {result.lineContent}
                    </p>
                  </div>
                </div>
              </div>
            );
          })
        )}
      </div>

      {/* Status Bar */}
      <div className="px-4 py-2 border-t border-aviation-border-panel bg-aviation-bg-secondary text-xs text-aviation-text-muted">
        {results.length} results found
      </div>
    </div>
  );
};

// ============================================================================
// Code Lineage Viewer
// ============================================================================

interface LineageNodeData {
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

interface CodeLineageViewerProps {
  nodes: LineageNodeData[];
  selectedNodeId?: string | null;
  focusedFilePath?: string;
  maxDepth?: number;
  onNodeSelect?: (node: LineageNodeData) => void;
  onNodeExpand?: (nodeId: string) => void;
  className?: string;
}

export const CodeLineageViewer: React.FC<CodeLineageViewerProps> = ({
  nodes,
  selectedNodeId,
  focusedFilePath,
  maxDepth = 10,
  onNodeSelect,
  onNodeExpand,
  className,
}) => {
  const getNodeIcon = (type: string) => {
    switch (type) {
      case 'commit': return <GitCommit className="w-3 h-3" />;
      case 'merge': return <GitMerge className="w-3 h-3" />;
      case 'branch': return <GitBranch className="w-3 h-3" />;
      default: return <GitCommit className="w-3 h-3" />;
    }
  };

  const formatTimestamp = (timestamp: number) => {
    const date = new Date(timestamp);
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  const buildTree = (nodes: LineageNodeData[]) => {
    const nodeMap = new Map<string, LineageNodeData>();
    nodes.forEach((n) => nodeMap.set(n.id, n));
    return nodes.filter((n) => !n.parent || !nodeMap.has(n.parent));
  };

  const rootNodes = useMemo(() => buildTree(nodes), [nodes]);

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center gap-2">
          <History className="w-5 h-5 text-aviation-cyan" />
          <span className="text-sm font-medium">Code Lineage</span>
          {focusedFilePath && (
            <span className="text-xs text-aviation-text-muted ml-auto truncate max-w-[200px]">
              {focusedFilePath}
            </span>
          )}
        </div>
      </div>

      {/* Timeline */}
      <div className="flex-1 overflow-auto p-4">
        <div className="relative">
          {/* Timeline Line */}
          <div className="absolute left-4 top-0 bottom-0 w-px bg-aviation-border-panel" />

          <div className="space-y-4">
            {rootNodes.map((node) => (
              <div key={node.id} className="relative pl-10">
                {/* Timeline Node */}
                <div className={cn(
                  'absolute left-2 w-5 h-5 rounded-full border-2 flex items-center justify-center',
                  selectedNodeId === node.id ? 'border-aviation-cyan bg-aviation-cyan' : 'border-aviation-text-muted bg-aviation-bg-panel'
                )}>
                  <div className={cn('w-2 h-2 rounded-full', selectedNodeId === node.id ? 'bg-black' : 'bg-aviation-text-muted')} />
                </div>

                <div
                  className={cn(
                    'p-3 rounded-lg border cursor-pointer transition-colors',
                    selectedNodeId === node.id ? 'border-aviation-cyan bg-aviation-cyan/10' : 'border-aviation-border-panel hover:border-aviation-text-muted'
                  )}
                  onClick={() => onNodeSelect?.(node)}
                >
                  <div className="flex items-center gap-2">
                    <span className="text-aviation-text-muted">{getNodeIcon(node.type)}</span>
                    <span className="text-sm font-medium">{node.name}</span>
                    {node.metadata?.tags?.map((tag) => (
                      <span key={tag} className="px-1.5 py-0.5 bg-cyan-500/20 text-cyan-400 rounded text-[10px]">
                        {tag}
                      </span>
                    ))}
                  </div>
                  {node.message && (
                    <p className="text-xs text-aviation-text-muted mt-1 line-clamp-2">{node.message}</p>
                  )}
                  <div className="flex items-center gap-3 mt-2 text-[10px] text-aviation-text-dim">
                    <span className="flex items-center gap-1">
                      <User className="w-3 h-3" /> {node.author}
                    </span>
                    <span className="flex items-center gap-1">
                      <Clock className="w-3 h-3" /> {formatTimestamp(node.timestamp)}
                    </span>
                    {node.metadata?.filesChanged && (
                      <span>{node.metadata.filesChanged} files</span>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Code Risk Analyzer
// ============================================================================

interface RiskIndicatorData {
  id: string;
  type: 'security' | 'performance' | 'maintainability' | 'testability' | 'complexity' | 'duplication';
  severity: 'critical' | 'high' | 'medium' | 'low' | 'info';
  message: string;
  location?: { start: { line: number }; end: { line: number } };
  file?: string;
  code?: string;
  suggestion?: string;
  cwe?: string;
}

interface CodeRiskAnalyzerProps {
  risks: RiskIndicatorData[];
  selectedRiskId?: string | null;
  showMetrics?: boolean;
  onRiskSelect?: (risk: RiskIndicatorData) => void;
  onRiskHover?: (risk: RiskIndicatorData | null) => void;
  onFixApply?: (riskId: string) => void;
  onSuppress?: (riskId: string, reason: string) => void;
  className?: string;
}

export const CodeRiskAnalyzer: React.FC<CodeRiskAnalyzerProps> = ({
  risks,
  selectedRiskId,
  showMetrics = false,
  onRiskSelect,
  onRiskHover,
  onFixApply,
  onSuppress,
  className,
}) => {
  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'text-red-400 bg-red-500/20 border-red-500/50';
      case 'high': return 'text-orange-400 bg-orange-500/20 border-orange-500/50';
      case 'medium': return 'text-amber-400 bg-amber-500/20 border-amber-500/50';
      case 'low': return 'text-blue-400 bg-blue-500/20 border-blue-500/50';
      default: return 'text-gray-400 bg-gray-500/20 border-gray-500/50';
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'security': return <Shield className="w-4 h-4" />;
      case 'performance': return <Zap className="w-4 h-4" />;
      case 'maintainability': return <Wrench className="w-4 h-4" />;
      case 'testability': return <CheckSquare className="w-4 h-4" />;
      case 'complexity': return <Activity className="w-4 h-4" />;
      default: return <AlertTriangle className="w-4 h-4" />;
    }
  };

  const Wrench: React.FC<{ className?: string }> = ({ className }) => <Hammer className={className} />;

  const severityCounts = useMemo(() => {
    const counts = { critical: 0, high: 0, medium: 0, low: 0, info: 0 };
    risks.forEach((r) => counts[r.severity]++);
    return counts;
  }, [risks]);

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Header with Summary */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <AlertTriangle className="w-5 h-5 text-red-400" />
            <span className="text-sm font-medium">Code Risk Analyzer</span>
          </div>
          <div className="flex items-center gap-2">
            {severityCounts.critical > 0 && (
              <span className="px-2 py-0.5 bg-red-500/20 text-red-400 rounded text-xs font-medium">
                {severityCounts.critical} critical
              </span>
            )}
            {severityCounts.high > 0 && (
              <span className="px-2 py-0.5 bg-orange-500/20 text-orange-400 rounded text-xs font-medium">
                {severityCounts.high} high
              </span>
            )}
            <span className="text-xs text-aviation-text-muted">{risks.length} total</span>
          </div>
        </div>
      </div>

      {/* Risk List */}
      <div className="flex-1 overflow-auto">
        {risks.map((risk) => {
          const isSelected = risk.id === selectedRiskId;
          return (
            <div
              key={risk.id}
              className={cn(
                'px-4 py-3 border-b border-aviation-border-panel cursor-pointer transition-colors',
                isSelected ? 'bg-aviation-amber/10' : 'hover:bg-aviation-bg-secondary'
              )}
              onClick={() => onRiskSelect?.(risk)}
              onMouseEnter={() => onRiskHover?.(risk)}
              onMouseLeave={() => onRiskHover?.(null)}
            >
              <div className="flex items-start gap-3">
                <div className={cn('p-1.5 rounded border', getSeverityColor(risk.severity))}>
                  {getTypeIcon(risk.type)}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium">{risk.message}</span>
                    <span className={cn('px-1.5 py-0.5 rounded text-[10px] uppercase', getSeverityColor(risk.severity))}>
                      {risk.severity}
                    </span>
                  </div>
                  <div className="flex items-center gap-2 mt-1 text-xs text-aviation-text-muted">
                    {risk.file && <span>{risk.file}</span>}
                    {risk.location && <span>Line {risk.location.start.line}</span>}
                    {risk.cwe && <span className="text-cyan-400">CWE-{risk.cwe}</span>}
                  </div>
                  {risk.suggestion && isSelected && (
                    <div className="mt-2 p-2 bg-aviation-bg-secondary rounded">
                      <div className="text-xs text-aviation-text-muted mb-1">Suggestion</div>
                      <p className="text-xs text-aviation-text-primary">{risk.suggestion}</p>
                    </div>
                  )}
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* Fix Actions */}
      {selectedRiskId && (
        <div className="flex items-center justify-end gap-2 px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
          <button
            onClick={() => onSuppress?.(selectedRiskId, 'false positive')}
            className="px-3 py-1.5 text-xs text-aviation-text-muted hover:text-aviation-text-primary transition-colors"
          >
            Suppress
          </button>
          <button
            onClick={() => onFixApply?.(selectedRiskId)}
            className="px-3 py-1.5 text-xs bg-aviation-cyan text-black rounded hover:bg-aviation-cyan/80 transition-colors flex items-center gap-1"
          >
            <Wand2 className="w-3 h-3" /> Apply Fix
          </button>
        </div>
      )}
    </div>
  );
};

// ============================================================================
// Import Graph Viewer
// ============================================================================

interface ImportNodeData {
  id: string;
  name: string;
  type: 'default' | 'named' | 'namespace' | 'side-effect';
  source: string;
  isReExported?: boolean;
  line?: number;
}

interface ImportEdgeData {
  source: string;
  target: string;
  type: 'import' | 're-export' | 'type-import';
}

interface ImportGraphViewerProps {
  imports: ImportNodeData[];
  edges: ImportEdgeData[];
  selectedNodeId?: string | null;
  filePath?: string;
  onNodeClick?: (node: ImportNodeData) => void;
  onExpandImports?: (nodeId: string) => void;
  className?: string;
}

export const ImportGraphViewer: React.FC<ImportGraphViewerProps> = ({
  imports,
  edges,
  selectedNodeId,
  filePath,
  onNodeClick,
  onExpandImports,
  className,
}) => {
  const getTypeColor = (type: string) => {
    switch (type) {
      case 'default': return 'text-green-400';
      case 'named': return 'text-cyan-400';
      case 'namespace': return 'text-purple-400';
      case 'side-effect': return 'text-amber-400';
      default: return 'text-gray-400';
    }
  };

  const groupedImports = useMemo(() => {
    const groups: Record<string, ImportNodeData[]> = {};
    imports.forEach((imp) => {
      if (!groups[imp.source]) groups[imp.source] = [];
      groups[imp.source].push(imp);
    });
    return groups;
  }, [imports]);

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center gap-2">
          <GitFork className="w-5 h-5 text-aviation-cyan" />
          <span className="text-sm font-medium">Import Graph</span>
          {filePath && (
            <span className="text-xs text-aviation-text-muted ml-2 truncate">{filePath}</span>
          )}
        </div>
      </div>

      {/* Import Groups */}
      <div className="flex-1 overflow-auto p-4">
        {Object.entries(groupedImports).map(([source, imports]) => (
          <div key={source} className="mb-4">
            <div className="flex items-center gap-2 mb-2 px-2">
              <FileCode className="w-4 h-4 text-aviation-text-muted" />
              <span className="text-xs font-medium text-aviation-text-primary truncate">{source}</span>
              <span className="text-[10px] text-aviation-text-dim ml-auto">{imports.length} imports</span>
            </div>
            <div className="space-y-1">
              {imports.map((imp) => {
                const isSelected = imp.id === selectedNodeId;
                return (
                  <div
                    key={imp.id}
                    className={cn(
                      'flex items-center gap-2 px-3 py-2 rounded cursor-pointer transition-colors',
                      isSelected ? 'bg-aviation-cyan/20' : 'hover:bg-aviation-bg-secondary'
                    )}
                    onClick={() => onNodeClick?.(imp)}
                  >
                    <span className={cn('text-xs font-mono', getTypeColor(imp.type))}>
                      {imp.name}
                    </span>
                    <span className={cn('px-1 py-0.5 rounded text-[10px]', getTypeColor(imp.type), 'bg-current/10')}>
                      {imp.type}
                    </span>
                    {imp.isReExported && (
                      <span className="text-[10px] text-cyan-400">re-exported</span>
                    )}
                    {imp.line && (
                      <span className="text-[10px] text-aviation-text-dim ml-auto">L{imp.line}</span>
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        ))}
      </div>

      {/* Legend */}
      <div className="px-4 py-2 border-t border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="flex items-center gap-4 text-xs text-aviation-text-muted">
          <span className="flex items-center gap-1"><div className="w-2 h-2 rounded-full bg-green-400" /> default</span>
          <span className="flex items-center gap-1"><div className="w-2 h-2 rounded-full bg-cyan-400" /> named</span>
          <span className="flex items-center gap-1"><div className="w-2 h-2 rounded-full bg-purple-400" /> namespace</span>
          <span className="flex items-center gap-1"><div className="w-2 h-2 rounded-full bg-amber-400" /> side-effect</span>
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Execution Aware Editor
// ============================================================================

interface ExecutionPointData {
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

interface WatchExpressionData {
  id: string;
  expression: string;
  value?: unknown;
  type?: string;
  error?: string;
}

interface ExecutionAwareEditorProps {
  code: string;
  language: string;
  executionPoints: ExecutionPointData[];
  currentExecutionPointId?: string | null;
  breakpoints?: string[];
  watchExpressions?: WatchExpressionData[];
  isRunning?: boolean;
  onExecutionPointSelect?: (point: ExecutionPointData) => void;
  onBreakpointToggle?: (line: number) => void;
  onWatchExpressionAdd?: (expression: string) => void;
  onWatchExpressionRemove?: (expressionId: string) => void;
  className?: string;
}

export const ExecutionAwareEditor: React.FC<ExecutionAwareEditorProps> = ({
  code,
  language,
  executionPoints,
  currentExecutionPointId,
  breakpoints = [],
  watchExpressions = [],
  isRunning = false,
  onExecutionPointSelect,
  onBreakpointToggle,
  onWatchExpressionAdd,
  onWatchExpressionRemove,
  className,
}) => {
  const [newWatchExpr, setNewWatchExpr] = useState('');

  const lines = useMemo(() => code.split('\n'), [code]);
  const currentPoint = executionPoints.find((p) => p.id === currentExecutionPointId);

  const handleAddWatch = () => {
    if (newWatchExpr.trim()) {
      onWatchExpressionAdd?.(newWatchExpr.trim());
      setNewWatchExpr('');
    }
  };

  return (
    <div className={cn('flex h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Editor Panel */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Debug Bar */}
        <div className="flex items-center justify-between px-4 py-2 bg-aviation-bg-secondary border-b border-aviation-border-panel">
          <div className="flex items-center gap-2">
            {isRunning ? (
              <>
                <Pause className="w-4 h-4 text-amber-400" />
                <span className="text-xs text-amber-400">Running</span>
              </>
            ) : (
              <>
                <Play className="w-4 h-4 text-green-400" />
                <span className="text-xs text-aviation-text-muted">Stopped</span>
              </>
            )}
          </div>
          {currentPoint && (
            <span className="text-xs text-aviation-text-muted">
              Line {currentPoint.line}, Col {currentPoint.column}
            </span>
          )}
        </div>

        {/* Code with Execution Points */}
        <div className="flex-1 overflow-auto font-mono text-xs">
          <div className="flex">
            {/* Line Numbers with Breakpoints */}
            <div className="flex-shrink-0 py-3 px-2 bg-aviation-bg-secondary text-right border-r border-aviation-border-panel select-none">
              {lines.map((_, idx) => {
                const lineNum = idx + 1;
                const hasBreakpoint = breakpoints.includes(String(lineNum));
                const execPoint = executionPoints.find((p) => p.line === lineNum);
                const isCurrentLine = currentPoint?.line === lineNum;
                
                return (
                  <div
                    key={idx}
                    className={cn(
                      'flex items-center justify-end gap-1 px-1 leading-6',
                      isCurrentLine ? 'text-aviation-cyan font-bold' : 'text-aviation-text-muted'
                    )}
                  >
                    {hasBreakpoint && (
                      <button
                        onClick={() => onBreakpointToggle?.(lineNum)}
                        className="w-3 h-3 rounded-full bg-red-500 flex items-center justify-center hover:bg-red-400"
                      >
                        <div className="w-1.5 h-1.5 rounded-full bg-black" />
                      </button>
                    )}
                    <span>{lineNum}</span>
                    {execPoint && !isCurrentLine && (
                      <div className="w-2 h-2 rounded-full bg-amber-400" />
                    )}
                  </div>
                );
              })}
            </div>

            {/* Code Content */}
            <div className="flex-1 p-3">
              {lines.map((line, idx) => {
                const lineNum = idx + 1;
                const isCurrentLine = currentPoint?.line === lineNum;
                const execPoint = executionPoints.find((p) => p.line === lineNum);
                
                return (
                  <div
                    key={idx}
                    className={cn(
                      'leading-6 px-2 -mx-2 relative',
                      isCurrentLine && 'bg-cyan-500/20',
                      execPoint && 'border-l-2 border-amber-400'
                    )}
                  >
                    {execPoint && isCurrentLine && (
                      <div className="absolute left-0 top-0 bottom-0 w-1 bg-cyan-400" />
                    )}
                    <span className="text-aviation-text-primary">{line || ' '}</span>
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      </div>

      {/* Watch Panel */}
      <div className="w-64 flex flex-col border-l border-aviation-border-panel">
        <div className="px-3 py-2 border-b border-aviation-border-panel">
          <div className="flex items-center gap-2">
            <Eye className="w-4 h-4 text-aviation-cyan" />
            <span className="text-xs font-medium">Watch Expressions</span>
          </div>
        </div>

        {/* Add Watch */}
        <div className="px-3 py-2 border-b border-aviation-border-panel">
          <div className="flex gap-1">
            <input
              type="text"
              value={newWatchExpr}
              onChange={(e) => setNewWatchExpr(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleAddWatch()}
              placeholder="Add expression..."
              className="flex-1 px-2 py-1 bg-aviation-bg-secondary border border-aviation-border-panel rounded text-xs text-aviation-text-primary placeholder:text-aviation-text-muted focus:outline-none focus:border-aviation-cyan"
            />
            <button
              onClick={handleAddWatch}
              className="p-1 hover:bg-aviation-bg-instrument rounded"
            >
              <Plus className="w-3 h-3" />
            </button>
          </div>
        </div>

        {/* Watch List */}
        <div className="flex-1 overflow-auto">
          {watchExpressions.map((expr) => (
            <div
              key={expr.id}
              className="px-3 py-2 border-b border-aviation-border-panel hover:bg-aviation-bg-secondary"
            >
              <div className="flex items-center justify-between mb-1">
                <code className="text-xs text-aviation-text-primary">{expr.expression}</code>
                <button
                  onClick={() => onWatchExpressionRemove?.(expr.id)}
                  className="p-0.5 hover:bg-red-500/30 rounded"
                >
                  <Minus className="w-3 h-3" />
                </button>
              </div>
              {expr.error ? (
                <span className="text-xs text-red-400">{expr.error}</span>
              ) : (
                <span className="text-xs text-green-400">{String(expr.value)}</span>
              )}
            </div>
          ))}
        </div>

        {/* Call Stack */}
        {currentPoint?.callStack && currentPoint.callStack.length > 0 && (
          <div className="border-t border-aviation-border-panel">
            <div className="px-3 py-2">
              <div className="flex items-center gap-2 mb-2">
                <Terminal className="w-4 h-4 text-aviation-cyan" />
                <span className="text-xs font-medium">Call Stack</span>
              </div>
              <div className="space-y-1">
                {currentPoint.callStack.map((frame, idx) => (
                  <div key={idx} className="text-xs text-aviation-text-muted font-mono truncate">
                    {idx > 0 && <span className="mr-1">←</span>}
                    {frame}
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

// ============================================================================
// AI Completion Inspector
// ============================================================================

interface AICompletionData {
  id: string;
  text: string;
  type: 'completion' | 'refactor' | 'explanation' | 'test';
  confidence: number;
  model?: string;
  timestamp: number;
  latency?: number;
  tokens?: number;
  context?: {
    cursorPosition?: { line: number; column: number };
    selectedText?: string;
    filePath?: string;
    language?: string;
  };
  alternatives?: Array<{ text: string; confidence: number; model?: string }>;
}

interface AICompletionInspectorProps {
  completions: AICompletionData[];
  selectedCompletionId?: string | null;
  currentCompletion?: AICompletionData | null;
  onCompletionSelect?: (completion: AICompletionData) => void;
  onCompletionAccept?: (completionId: string) => void;
  onComparisonToggle?: (completionId: string) => void;
  className?: string;
}

export const AICompletionInspector: React.FC<AICompletionInspectorProps> = ({
  completions,
  selectedCompletionId,
  currentCompletion,
  onCompletionSelect,
  onCompletionAccept,
  onComparisonToggle,
  className,
}) => {
  const [showAlternatives, setShowAlternatives] = useState(false);

  return (
    <div className={cn('flex h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Completions List */}
      <div className="w-80 flex flex-col border-r border-aviation-border-panel">
        <div className="px-4 py-3 border-b border-aviation-border-panel">
          <div className="flex items-center gap-2">
            <Sparkles className="w-5 h-5 text-aviation-cyan" />
            <span className="text-sm font-medium">AI Completions</span>
            <span className="text-xs text-aviation-text-muted ml-auto">{completions.length}</span>
          </div>
        </div>
        <div className="flex-1 overflow-auto">
          {completions.map((completion) => {
            const isSelected = completion.id === selectedCompletionId;
            const isCurrent = completion.id === currentCompletion?.id;
            
            return (
              <div
                key={completion.id}
                className={cn(
                  'px-4 py-3 border-b border-aviation-border-panel cursor-pointer transition-colors',
                  isCurrent && 'border-l-2 border-l-green-400',
                  isSelected && 'bg-aviation-cyan/10'
                )}
                onClick={() => onCompletionSelect?.(completion)}
              >
                <div className="flex items-center justify-between mb-1">
                  <span className={cn(
                    'px-1.5 py-0.5 rounded text-[10px] capitalize',
                    completion.type === 'completion' ? 'bg-cyan-500/20 text-cyan-400' :
                    completion.type === 'refactor' ? 'bg-purple-500/20 text-purple-400' :
                    'bg-gray-500/20 text-gray-400'
                  )}>
                    {completion.type}
                  </span>
                  <span className="text-xs text-aviation-text-muted">
                    {Math.round(completion.confidence * 100)}%
                  </span>
                </div>
                <p className="text-xs text-aviation-text-primary font-mono line-clamp-3">
                  {completion.text}
                </p>
                <div className="flex items-center gap-2 mt-2 text-[10px] text-aviation-text-dim">
                  {completion.model && <span>{completion.model}</span>}
                  {completion.latency && <span>{completion.latency}ms</span>}
                  {completion.tokens && <span>{completion.tokens} tokens</span>}
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Completion Detail */}
      <div className="flex-1 flex flex-col">
        {currentCompletion ? (
          <>
            <div className="px-4 py-3 border-b border-aviation-border-panel">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Bot className="w-5 h-5 text-aviation-cyan" />
                  <span className="text-sm font-medium">Current Completion</span>
                </div>
                <button
                  onClick={() => onCompletionAccept?.(currentCompletion.id)}
                  className="px-3 py-1.5 bg-aviation-cyan text-black rounded text-xs font-medium hover:bg-aviation-cyan/80 transition-colors"
                >
                  Accept
                </button>
              </div>
            </div>
            <div className="flex-1 overflow-auto p-4">
              <pre className="text-sm font-mono text-aviation-text-primary whitespace-pre-wrap">
                {currentCompletion.text}
              </pre>
              
              {/* Context */}
              {currentCompletion.context && (
                <div className="mt-4 p-3 bg-aviation-bg-secondary rounded">
                  <div className="text-xs text-aviation-text-muted mb-2">Context</div>
                  <div className="text-xs text-aviation-text-primary space-y-1">
                    {currentCompletion.context.filePath && (
                      <div><span className="text-aviation-text-dim">File:</span> {currentCompletion.context.filePath}</div>
                    )}
                    {currentCompletion.context.language && (
                      <div><span className="text-aviation-text-dim">Language:</span> {currentCompletion.context.language}</div>
                    )}
                    {currentCompletion.context.cursorPosition && (
                      <div><span className="text-aviation-text-dim">Position:</span> Line {currentCompletion.context.cursorPosition.line}, Col {currentCompletion.context.cursorPosition.column}</div>
                    )}
                  </div>
                </div>
              )}

              {/* Alternatives */}
              {currentCompletion.alternatives && currentCompletion.alternatives.length > 0 && (
                <div className="mt-4">
                  <button
                    onClick={() => setShowAlternatives(!showAlternatives)}
                    className="flex items-center gap-2 text-xs text-aviation-text-muted hover:text-aviation-text-primary"
                  >
                    {showAlternatives ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}
                    {currentCompletion.alternatives.length} Alternatives
                  </button>
                  {showAlternatives && (
                    <div className="mt-2 space-y-2">
                      {currentCompletion.alternatives.map((alt, idx) => (
                        <div key={idx} className="p-3 bg-aviation-bg-secondary rounded border border-aviation-border-panel">
                          <div className="flex items-center justify-between mb-1">
                            <span className="text-xs text-aviation-text-muted">
                              Alternative {idx + 1}
                            </span>
                            <span className="text-xs text-aviation-text-muted">
                              {Math.round(alt.confidence * 100)}% - {alt.model}
                            </span>
                          </div>
                          <pre className="text-xs font-mono text-aviation-text-primary whitespace-pre-wrap">
                            {alt.text}
                          </pre>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              )}
            </div>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center">
              <Sparkles className="w-8 h-8 text-aviation-text-muted mx-auto mb-3" />
              <p className="text-sm text-aviation-text-muted">Select a completion to inspect</p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

// ============================================================================
// Refactor Simulation Viewer
// ============================================================================

interface SimulationChangeData {
  fileId: string;
  filePath: string;
  changeType: 'add' | 'modify' | 'delete';
  before: string;
  after: string;
}

interface RefactorSimulationData {
  id: string;
  name: string;
  description: string;
  changes: SimulationChangeData[];
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

interface RefactorSimulationViewerProps {
  simulation: RefactorSimulationData | null;
  onAccept?: () => void;
  onReject?: () => void;
  onStepForward?: () => void;
  onStepBackward?: () => void;
  className?: string;
}

export const RefactorSimulationViewer: React.FC<RefactorSimulationViewerProps> = ({
  simulation,
  onAccept,
  onReject,
  onStepForward,
  onStepBackward,
  className,
}) => {
  const [currentStep, setCurrentStep] = useState(0);

  if (!simulation) {
    return (
      <div className={cn('flex flex-col h-full items-center justify-center bg-aviation-bg-panel rounded-lg border border-aviation-border-panel', className)}>
        <GitFork className="w-8 h-8 text-aviation-text-muted mb-3" />
        <p className="text-sm text-aviation-text-muted">No refactor simulation loaded</p>
        <p className="text-xs text-aviation-text-dim mt-1">Select a refactor opportunity to preview</p>
      </div>
    );
  }

  const currentChange = simulation.changes[currentStep];
  const progress = ((currentStep + 1) / simulation.changes.length) * 100;

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-2">
            <GitFork className="w-5 h-5 text-aviation-cyan" />
            <div>
              <div className="text-sm font-medium">{simulation.name}</div>
              <div className="text-xs text-aviation-text-muted">{simulation.description}</div>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => { setCurrentStep(Math.max(0, currentStep - 1)); onStepBackward?.(); }}
              disabled={currentStep === 0}
              className="p-2 hover:bg-aviation-bg-instrument rounded disabled:opacity-50"
            >
              <SkipBack className="w-4 h-4" />
            </button>
            <span className="text-xs text-aviation-text-muted">
              {currentStep + 1} / {simulation.changes.length}
            </span>
            <button
              onClick={() => { setCurrentStep(Math.min(simulation.changes.length - 1, currentStep + 1)); onStepForward?.(); }}
              disabled={currentStep === simulation.changes.length - 1}
              className="p-2 hover:bg-aviation-bg-instrument rounded disabled:opacity-50"
            >
              <SkipForward className="w-4 h-4" />
            </button>
          </div>
        </div>
        
        {/* Progress Bar */}
        <div className="h-1 bg-aviation-bg-secondary rounded-full overflow-hidden">
          <div
            className="h-full bg-aviation-cyan transition-all duration-300"
            style={{ width: `${progress}%` }}
          />
        </div>
      </div>

      {/* Impact Analysis */}
      {simulation.impactAnalysis && (
        <div className="px-4 py-2 border-b border-aviation-border-panel bg-aviation-bg-secondary">
          <div className="flex items-center gap-4 text-xs">
            {simulation.impactAnalysis.estimatedTime && (
              <span><Clock className="w-3 h-3 inline mr-1" />{simulation.impactAnalysis.estimatedTime}min</span>
            )}
            {simulation.impactAnalysis.riskLevel && (
              <span className={cn(
                simulation.impactAnalysis.riskLevel === 'low' ? 'text-green-400' :
                simulation.impactAnalysis.riskLevel === 'medium' ? 'text-amber-400' : 'text-red-400'
              )}>
                Risk: {simulation.impactAnalysis.riskLevel}
              </span>
            )}
            {simulation.impactAnalysis.testCoverageImpact !== undefined && (
              <span>Coverage impact: {simulation.impactAnalysis.testCoverageImpact > 0 ? '+' : ''}{simulation.impactAnalysis.testCoverageImpact}%</span>
            )}
          </div>
        </div>
      )}

      {/* Validation Results */}
      {simulation.validationResults && simulation.validationResults.length > 0 && (
        <div className="px-4 py-2 border-b border-aviation-border-panel">
          <div className="flex items-center gap-3">
            {simulation.validationResults.map((result, idx) => (
              <div
                key={idx}
                className={cn(
                  'flex items-center gap-1 text-xs',
                  result.passed ? 'text-green-400' : 'text-red-400'
                )}
              >
                {result.passed ? <CheckCircle2 className="w-3 h-3" /> : <XCircle className="w-3 h-3" />}
                {result.type}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Change Preview */}
      {currentChange && (
        <div className="flex-1 overflow-auto p-4">
          <div className="flex items-center gap-2 mb-3">
            <FileCode className="w-4 h-4 text-aviation-text-muted" />
            <span className="text-sm font-medium">{currentChange.filePath}</span>
            <span className={cn(
              'px-1.5 py-0.5 rounded text-[10px]',
              currentChange.changeType === 'add' ? 'bg-green-500/20 text-green-400' :
              currentChange.changeType === 'delete' ? 'bg-red-500/20 text-red-400' :
              'bg-amber-500/20 text-amber-400'
            )}>
              {currentChange.changeType}
            </span>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <div className="text-xs text-aviation-text-muted mb-2">Before</div>
              <pre className="p-3 bg-red-950/30 rounded border border-red-900/50 text-xs font-mono text-aviation-text-primary overflow-x-auto">
                {currentChange.before}
              </pre>
            </div>
            <div>
              <div className="text-xs text-aviation-text-muted mb-2">After</div>
              <pre className="p-3 bg-green-950/30 rounded border border-green-900/50 text-xs font-mono text-aviation-text-primary overflow-x-auto">
                {currentChange.after}
              </pre>
            </div>
          </div>
        </div>
      )}

      {/* Actions */}
      <div className="flex items-center justify-between px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
        <button
          onClick={onReject}
          className="px-4 py-2 text-sm text-red-400 hover:text-red-300 transition-colors"
        >
          Cancel
        </button>
        <button
          onClick={onAccept}
          className="px-4 py-2 bg-aviation-cyan text-black rounded text-sm font-medium hover:bg-aviation-cyan/80 transition-colors"
        >
          Apply All Changes
        </button>
      </div>
    </div>
  );
};

// ============================================================================
// Architecture Constraint Panel
// ============================================================================

interface ArchitectureConstraintData {
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

interface ArchitectureConstraintPanelProps {
  constraints: ArchitectureConstraintData[];
  selectedConstraintId?: string | null;
  onConstraintSelect?: (constraint: ArchitectureConstraintData) => void;
  onFixApply?: (constraintId: string) => void;
  onDismiss?: (constraintId: string) => void;
  className?: string;
}

export const ArchitectureConstraintPanel: React.FC<ArchitectureConstraintPanelProps> = ({
  constraints,
  selectedConstraintId,
  onConstraintSelect,
  onFixApply,
  onDismiss,
  className,
}) => {
  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case 'error': return <XCircle className="w-4 h-4 text-red-400" />;
      case 'warning': return <AlertTriangle className="w-4 h-4 text-amber-400" />;
      default: return <Info className="w-4 h-4 text-blue-400" />;
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'naming': return <Type className="w-4 h-4" />;
      case 'layering': return <LayersIcon className="w-4 h-4" />;
      case 'dependency': return <ArrowRight className="w-4 h-4" />;
      case 'visibility': return <Eye className="w-4 h-4" />;
      default: return <Settings className="w-4 h-4" />;
    }
  };

  const violationCounts = useMemo(() => {
    return constraints.reduce((acc, c) => {
      acc[c.severity] = (acc[c.severity] || 0) + (c.violatedBy?.length || 0);
      return acc;
    }, {} as Record<string, number>);
  }, [constraints]);

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Shield className="w-5 h-5 text-aviation-cyan" />
            <span className="text-sm font-medium">Architecture Constraints</span>
          </div>
          <div className="flex items-center gap-2">
            {violationCounts.error > 0 && (
              <span className="px-2 py-0.5 bg-red-500/20 text-red-400 rounded text-xs">
                {violationCounts.error} errors
              </span>
            )}
            {violationCounts.warning > 0 && (
              <span className="px-2 py-0.5 bg-amber-500/20 text-amber-400 rounded text-xs">
                {violationCounts.warning} warnings
              </span>
            )}
          </div>
        </div>
      </div>

      {/* Constraint List */}
      <div className="flex-1 overflow-auto">
        {constraints.map((constraint) => {
          const isSelected = constraint.id === selectedConstraintId;
          const hasViolations = constraint.violatedBy && constraint.violatedBy.length > 0;
          
          return (
            <div
              key={constraint.id}
              className={cn(
                'px-4 py-3 border-b border-aviation-border-panel',
                isSelected && 'bg-aviation-amber/10'
              )}
            >
              <div
                className="flex items-start gap-3 cursor-pointer"
                onClick={() => onConstraintSelect?.(constraint)}
              >
                <div className="mt-0.5">{getSeverityIcon(constraint.severity)}</div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium">{constraint.name}</span>
                    <span className={cn(
                      'px-1.5 py-0.5 rounded text-[10px] capitalize',
                      constraint.enforcement === 'strict' ? 'bg-red-500/20 text-red-400' : 'bg-blue-500/20 text-blue-400'
                    )}>
                      {constraint.enforcement}
                    </span>
                  </div>
                  <p className="text-xs text-aviation-text-muted mt-1">{constraint.description}</p>
                  
                  {/* Violations */}
                  {hasViolations && isSelected && (
                    <div className="mt-3 space-y-2">
                      <div className="text-xs text-aviation-text-muted">Violations ({constraint.violatedBy!.length})</div>
                      {constraint.violatedBy!.map((v, idx) => (
                        <div key={idx} className="p-2 bg-aviation-bg-secondary rounded">
                          <div className="flex items-center gap-2 text-xs">
                            <FileCode className="w-3 h-3" />
                            <span className="text-aviation-text-primary">{v.file}</span>
                            {v.line && <span className="text-aviation-text-dim">:{v.line}</span>}
                          </div>
                          {v.details && (
                            <p className="text-xs text-aviation-text-muted mt-1">{v.details}</p>
                          )}
                        </div>
                      ))}
                      
                      {constraint.fixSuggestion && (
                        <div className="p-2 bg-green-500/10 border border-green-500/30 rounded">
                          <div className="text-xs text-green-400 mb-1">Fix Suggestion</div>
                          <p className="text-xs text-aviation-text-primary">{constraint.fixSuggestion}</p>
                        </div>
                      )}
                    </div>
                  )}
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* Actions */}
      {selectedConstraintId && (
        <div className="flex items-center justify-end gap-2 px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
          <button
            onClick={() => onDismiss?.(selectedConstraintId)}
            className="px-3 py-1.5 text-xs text-aviation-text-muted hover:text-aviation-text-primary transition-colors"
          >
            Dismiss
          </button>
          <button
            onClick={() => onFixApply?.(selectedConstraintId)}
            className="px-3 py-1.5 text-xs bg-aviation-cyan text-black rounded hover:bg-aviation-cyan/80 transition-colors flex items-center gap-1"
          >
            <Wand2 className="w-3 h-3" /> Apply Fix
          </button>
        </div>
      )}
    </div>
  );
};

// ============================================================================
// Code Ownership Map
// ============================================================================

interface CodeOwnerData {
  id: string;
  name: string;
  email: string;
  avatar?: string;
  gitHubUsername?: string;
}

interface FileOwnershipData {
  filePath: string;
  owners: CodeOwnerData[];
  lastModified?: number;
  lastModifiedBy?: string;
  reviewRequired?: boolean;
  autoAssignment?: boolean;
}

interface CodeOwnershipMapProps {
  ownerships: FileOwnershipData[];
  selectedFilePath?: string | null;
  selectedOwnerId?: string | null;
  onFileSelect?: (ownership: FileOwnershipData) => void;
  onOwnerClick?: (owner: CodeOwnerData) => void;
  onAssign?: (filePath: string, ownerId: string) => void;
  className?: string;
}

export const CodeOwnershipMap: React.FC<CodeOwnershipMapProps> = ({
  ownerships,
  selectedFilePath,
  selectedOwnerId,
  onFileSelect,
  onOwnerClick,
  onAssign,
  className,
}) => {
  const [showAssignModal, setShowAssignModal] = useState(false);
  const [selectedFile, setSelectedFile] = useState<FileOwnershipData | null>(null);

  const groupedByOwner = useMemo(() => {
    const groups: Record<string, FileOwnershipData[]> = {};
    ownerships.forEach((o) => {
      o.owners.forEach((owner) => {
        if (!groups[owner.id]) groups[owner.id] = [];
        groups[owner.id].push(o);
      });
    });
    return groups;
  }, [ownerships]);

  const handleAssign = (ownership: FileOwnershipData) => {
    setSelectedFile(ownership);
    setShowAssignModal(true);
  };

  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Users className="w-5 h-5 text-aviation-cyan" />
            <span className="text-sm font-medium">Code Ownership</span>
          </div>
          <div className="flex items-center gap-2 text-xs text-aviation-text-muted">
            <span>{ownerships.length} files</span>
            <span>·</span>
            <span>{Object.keys(groupedByOwner).length} owners</span>
          </div>
        </div>
      </div>

      {/* Ownership Grid */}
      <div className="flex-1 overflow-auto p-4">
        <div className="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-3">
          {ownerships.map((ownership) => {
            const isSelected = ownership.filePath === selectedFilePath;
            
            return (
              <div
                key={ownership.filePath}
                className={cn(
                  'p-3 rounded-lg border transition-colors',
                  isSelected ? 'border-aviation-cyan bg-aviation-cyan/10' : 'border-aviation-border-panel hover:border-aviation-text-muted'
                )}
                onClick={() => onFileSelect?.(ownership)}
              >
                <div className="flex items-center gap-2 mb-2">
                  <FileCode className="w-4 h-4 text-aviation-text-muted" />
                  <span className="text-xs font-medium truncate flex-1">{ownership.filePath}</span>
                </div>

                <div className="flex items-center gap-2">
                  {ownership.owners.slice(0, 3).map((owner) => (
                    <button
                      key={owner.id}
                      onClick={(e) => { e.stopPropagation(); onOwnerClick?.(owner); }}
                      className="flex items-center gap-1.5 px-2 py-1 bg-aviation-bg-secondary rounded-full hover:bg-aviation-bg-instrument transition-colors"
                    >
                      {owner.avatar ? (
                        <img src={owner.avatar} alt={owner.name} className="w-4 h-4 rounded-full" />
                      ) : (
                        <div className="w-4 h-4 rounded-full bg-aviation-cyan/30 flex items-center justify-center">
                          <User className="w-2.5 h-2.5 text-aviation-cyan" />
                        </div>
                      )}
                      <span className="text-xs">{owner.name}</span>
                    </button>
                  ))}
                  {ownership.owners.length > 3 && (
                    <span className="text-xs text-aviation-text-muted">+{ownership.owners.length - 3}</span>
                  )}
                </div>

                <div className="flex items-center justify-between mt-3 pt-2 border-t border-aviation-border-panel">
                  <div className="flex items-center gap-2 text-[10px] text-aviation-text-muted">
                    {ownership.reviewRequired && (
                      <span className="flex items-center gap-1 text-amber-400">
                        <GitPullRequest className="w-3 h-3" /> review
                      </span>
                    )}
                    {ownership.autoAssignment && (
                      <span className="flex items-center gap-1 text-cyan-400">
                        <Bot className="w-3 h-3" /> auto
                      </span>
                    )}
                  </div>
                  <button
                    onClick={(e) => { e.stopPropagation(); handleAssign(ownership); }}
                    className="text-xs text-aviation-cyan hover:underline"
                  >
                    Assign
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Owner Legend */}
      <div className="px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
        <div className="text-xs text-aviation-text-muted mb-2">Top Owners</div>
        <div className="flex flex-wrap gap-2">
          {Object.entries(groupedByOwner).slice(0, 5).map(([ownerId, files]) => (
            <div key={ownerId} className="flex items-center gap-2 px-2 py-1 bg-aviation-bg-instrument rounded">
              <div className="w-4 h-4 rounded-full bg-aviation-cyan/30 flex items-center justify-center">
                <User className="w-2.5 h-2.5 text-aviation-cyan" />
              </div>
              <span className="text-xs">{files[0]?.owners.find(o => o.id === ownerId)?.name}</span>
              <span className="text-[10px] text-aviation-text-dim">{files.length}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};
