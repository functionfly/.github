/**
 * @functionfly/ui-memory
 * Memory Systems Components - AI-powered memory and knowledge management
 */

import React, { useState, useCallback, useMemo, useRef, useEffect } from 'react';
import { cn } from '@functionfly/ui-core';
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
  Code,
  Heart,
  Download,
  GitFork,
  Edit,
} from 'lucide-react';

// ============================================================================
// Memory Graph
// ============================================================================

interface MemoryGraphProps {
  nodes: Array<{
    id: string;
    type: 'concept' | 'entity' | 'event' | 'document' | 'code' | 'conversation' | 'agent';
    label: string;
    content?: string;
    timestamp: number;
    importance: number;
    connections?: string[];
    metadata?: Record<string, unknown>;
  }>;
  edges: Array<{
    id: string;
    source: string;
    target: string;
    type: 'references' | 'derives_from' | 'related_to' | 'part_of' | 'evolved_from' | 'associated_with';
    weight?: number;
  }>;
  selectedNodeId?: string | null;
  highlightedNodes?: string[];
  onNodeSelect?: (node: { id: string; type: string; label: string }) => void;
  onNodeHover?: (node: { id: string; type: string; label: string } | null) => void;
  onEdgeClick?: (edge: { id: string; source: string; target: string; type: string }) => void;
  layout?: 'force' | 'tree' | 'circular' | 'grid';
  className?: string;
}

export const MemoryGraph: React.FC<MemoryGraphProps> = ({
  nodes,
  edges,
  selectedNodeId = null,
  highlightedNodes = [],
  onNodeSelect,
  onNodeHover,
  onEdgeClick,
  layout = 'force',
  className,
}) => {
  const [hoveredNode, setHoveredNode] = useState<string | null>(null);
  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(new Set());
  
  const getNodeColor = (type: string) => {
    const colors: Record<string, string> = {
      concept: 'bg-aviation-cyan/20 border-aviation-cyan',
      entity: 'bg-aviation-amber/20 border-aviation-amber',
      event: 'bg-purple-500/20 border-purple-500',
      document: 'bg-blue-500/20 border-blue-500',
      code: 'bg-green-500/20 border-green-500',
      conversation: 'bg-pink-500/20 border-pink-500',
      agent: 'bg-orange-500/20 border-orange-500',
      context: 'bg-violet-500/20 border-violet-500',
      core: 'bg-aviation-cyan/20 border-aviation-cyan',
      workflow: 'bg-aviation-amber/20 border-aviation-amber',
    };
    return colors[type] || 'bg-gray-500/20 border-gray-500';
  };

  const getNodeSvgColor = (type: string) => {
    const colors: Record<string, { fill: string; stroke: string }> = {
      concept: { fill: 'rgba(6, 182, 212, 0.25)', stroke: '#06b6d4' },
      entity: { fill: 'rgba(245, 158, 11, 0.25)', stroke: '#f59e0b' },
      event: { fill: 'rgba(168, 85, 247, 0.25)', stroke: '#a855f7' },
      document: { fill: 'rgba(59, 130, 246, 0.25)', stroke: '#3b82f6' },
      code: { fill: 'rgba(34, 197, 94, 0.25)', stroke: '#22c55e' },
      conversation: { fill: 'rgba(236, 72, 153, 0.25)', stroke: '#ec4899' },
      agent: { fill: 'rgba(249, 115, 22, 0.25)', stroke: '#f97316' },
      context: { fill: 'rgba(139, 92, 246, 0.25)', stroke: '#8b5cf6' },
      core: { fill: 'rgba(6, 182, 212, 0.25)', stroke: '#06b6d4' },
      workflow: { fill: 'rgba(245, 158, 11, 0.25)', stroke: '#f59e0b' },
    };
    return colors[type] || { fill: 'rgba(107, 114, 128, 0.25)', stroke: '#9ca3af' };
  };

  const getEdgeSvgColor = (type: string) => {
    const colors: Record<string, string> = {
      references: '#06b6d4',
      derives_from: '#f59e0b',
      related_to: '#6b7280',
      part_of: '#a855f7',
      evolved_from: '#3b82f6',
      associated_with: '#22c55e',
    };
    return colors[type] || '#9ca3af';
  };

  const getEdgeColor = (type: string) => {
    const colors: Record<string, string> = {
      references: 'stroke-aviation-cyan',
      derives_from: 'stroke-aviation-amber',
      related_to: 'stroke-gray-500',
      part_of: 'stroke-purple-500',
      evolved_from: 'stroke-blue-500',
      associated_with: 'stroke-green-500',
    };
    return colors[type] || '#9ca3af';
  };

  const nodePositions = useMemo(() => {
    const positions: Record<string, { x: number; y: number }> = {};
    const centerX = 200;
    const centerY = 150;
    
    if (layout === 'circular') {
      nodes.forEach((node, index) => {
        const angle = (2 * Math.PI * index) / nodes.length;
        const radius = 120;
        positions[node.id] = {
          x: centerX + radius * Math.cos(angle),
          y: centerY + radius * Math.sin(angle),
        };
      });
    } else if (layout === 'grid') {
      const cols = Math.ceil(Math.sqrt(nodes.length));
      nodes.forEach((node, index) => {
        const row = Math.floor(index / cols);
        const col = index % cols;
        positions[node.id] = {
          x: 40 + col * 100,
          y: 40 + row * 80,
        };
      });
    } else {
      // Force-like layout
      nodes.forEach((node, index) => {
        const seed = index * 137.5;
        const radius = 80 + (node.importance * 40);
        positions[node.id] = {
          x: centerX + radius * Math.cos(seed * Math.PI / 180),
          y: centerY + radius * Math.sin(seed * Math.PI / 180),
        };
      });
    }
    return positions;
  }, [nodes, layout]);
  
  return (
    <div className={cn('relative bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className="absolute top-3 left-3 flex items-center gap-2 z-10">
        <div className="flex items-center gap-1.5 px-2 py-1 bg-aviation-bg-instrument rounded border border-aviation-border-panel">
          <Brain className="w-4 h-4 text-aviation-cyan" />
          <span className="text-xs text-aviation-text-primary font-medium">Memory Graph</span>
        </div>
        <div className="flex items-center gap-1 px-2 py-1 bg-aviation-bg-secondary rounded border border-aviation-border-panel">
          <span className="text-xs text-aviation-text-muted">{nodes.length} nodes</span>
        </div>
      </div>
      
      <svg className="w-full h-full" viewBox="0 0 400 300">
        {/* Edges */}
        {edges.map((edge) => {
          const sourcePos = nodePositions[edge.source];
          const targetPos = nodePositions[edge.target];
          if (!sourcePos || !targetPos) return null;
          
          return (
            <g key={edge.id} onClick={() => onEdgeClick?.(edge)} className="cursor-pointer">
              <line
                x1={sourcePos.x}
                y1={sourcePos.y}
                x2={targetPos.x}
                y2={targetPos.y}
                stroke={getEdgeSvgColor(edge.type)}
                strokeWidth={edge.weight || 1}
                opacity={0.5}
                className="transition-opacity"
              />
              <circle
                cx={(sourcePos.x + targetPos.x) / 2}
                cy={(sourcePos.y + targetPos.y) / 2}
                r={4}
                fill="var(--color-aviation-bg-panel, #1a1a28)"
                stroke={getEdgeSvgColor(edge.type)}
                strokeWidth={1}
              />
            </g>
          );
        })}
        
        {/* Nodes */}
        {nodes.map((node) => {
          const pos = nodePositions[node.id];
          if (!pos) return null;
          
          const isSelected = selectedNodeId === node.id;
          const isHighlighted = highlightedNodes.includes(node.id);
          const isHovered = hoveredNode === node.id;
          
          return (
            <g
              key={node.id}
              onClick={() => onNodeSelect?.(node)}
              onMouseEnter={() => {
                setHoveredNode(node.id);
                onNodeHover?.(node);
              }}
              onMouseLeave={() => {
                setHoveredNode(null);
                onNodeHover?.(null);
              }}
              className="cursor-pointer"
            >
              <circle
                cx={pos.x}
                cy={pos.y}
                r={isSelected ? 20 : isHovered ? 18 : 16}
                fill={getNodeSvgColor(node.type).fill}
                stroke={isSelected ? '#06b6d4' : getNodeSvgColor(node.type).stroke}
                strokeWidth={isSelected ? 2 : 1}
                className={cn(
                  'transition-all',
                  isHighlighted && 'opacity-100',
                  !isHighlighted && highlightedNodes.length > 0 && 'opacity-40'
                )}
                fillOpacity={0.9}
              />
              <text
                x={pos.x}
                y={pos.y + 30}
                textAnchor="middle"
                fill="var(--color-aviation-text-primary, #e8e8f0)"
                fontSize={10}
                fontFamily="var(--font-sans, ui-sans-serif, system-ui, sans-serif)"
              >
                {node.label.length > 12 ? node.label.slice(0, 12) + '...' : node.label}
              </text>
              {node.importance > 0.7 && (
                <circle cx={pos.x + 12} cy={pos.y - 12} r={4} fill="#fbbf24" />
              )}
            </g>
          );
        })}
      </svg>
      
      {/* Legend */}
      <div className="absolute bottom-3 left-3 flex items-center gap-3 px-3 py-2 bg-aviation-bg-secondary/80 rounded border border-aviation-border-panel">
        {['concept', 'entity', 'event'].map((type) => (
          <div key={type} className="flex items-center gap-1.5">
            <div className={cn('w-3 h-3 rounded-sm border', getNodeColor(type))} />
            <span className="text-[10px] text-aviation-text-muted capitalize">{type}</span>
          </div>
        ))}
      </div>
    </div>
  );
};

// ============================================================================
// Semantic Memory Viewer
// ============================================================================

interface SemanticMemoryViewerProps {
  entries: Array<{
    id: string;
    content: string;
    semanticType: 'fact' | 'procedure' | 'preference' | 'context' | 'relationship';
    confidence: number;
    source?: string;
    timestamp: number;
    lastAccessed?: number;
    accessCount?: number;
    tags?: string[];
  }>;
  selectedEntryId?: string | null;
  searchQuery?: string;
  filterType?: 'fact' | 'procedure' | 'preference' | 'context' | 'relationship' | 'all';
  onEntrySelect?: (entry: { id: string; content: string; semanticType: string }) => void;
  onSearch?: (query: string) => void;
  onEntryDelete?: (entryId: string) => void;
  className?: string;
}

export const SemanticMemoryViewer: React.FC<SemanticMemoryViewerProps> = ({
  entries,
  selectedEntryId = null,
  searchQuery = '',
  filterType = 'all',
  onEntrySelect,
  onSearch,
  onEntryDelete,
  className,
}) => {
  const [localSearch, setLocalSearch] = useState(searchQuery);
  const [selectedType, setSelectedType] = useState(filterType);
  
  const filteredEntries = useMemo(() => {
    return entries.filter((entry) => {
      const matchesType = selectedType === 'all' || entry.semanticType === selectedType;
      const matchesSearch = !localSearch || 
        entry.content.toLowerCase().includes(localSearch.toLowerCase()) ||
        entry.tags?.some(tag => tag.toLowerCase().includes(localSearch.toLowerCase()));
      return matchesType && matchesSearch;
    });
  }, [entries, localSearch, selectedType]);
  
  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'fact': return <Hexagon className="w-4 h-4" />;
      case 'procedure': return <List className="w-4 h-4" />;
      case 'preference': return <Heart className="w-4 h-4" />;
      case 'context': return <Compass className="w-4 h-4" />;
      case 'relationship': return <Link className="w-4 h-4" />;
      default: return <Database className="w-4 h-4" />;
    }
  };
  
  const getTypeColor = (type: string) => {
    const colors: Record<string, string> = {
      fact: 'text-aviation-cyan',
      procedure: 'text-aviation-amber',
      preference: 'text-pink-400',
      context: 'text-purple-400',
      relationship: 'text-blue-400',
    };
    return colors[type] || 'text-gray-400';
  };
  
  const formatTimestamp = (ts: number) => {
    const date = new Date(ts);
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  };
  
  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2">
            <Brain className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Semantic Memory</h3>
          </div>
          <span className="text-xs text-aviation-text-muted">{filteredEntries.length} entries</span>
        </div>
        
        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-aviation-text-muted" />
          <input
            type="text"
            value={localSearch}
            onChange={(e) => {
              setLocalSearch(e.target.value);
              onSearch?.(e.target.value);
            }}
            placeholder="Search memories..."
            className="w-full pl-9 pr-3 py-2 bg-aviation-bg-instrument border border-aviation-border-panel rounded text-sm text-aviation-text-primary placeholder:text-aviation-text-dim focus:outline-none focus:border-aviation-cyan"
          />
        </div>
        
        {/* Filter Tabs */}
        <div className="flex items-center gap-1 mt-3">
          {['all', 'fact', 'procedure', 'preference', 'context', 'relationship'].map((type) => (
            <button
              key={type}
              onClick={() => setSelectedType(type as typeof selectedType)}
              className={cn(
                'px-2 py-1 text-xs rounded transition-colors',
                selectedType === type
                  ? 'bg-aviation-cyan/20 text-aviation-cyan'
                  : 'text-aviation-text-muted hover:text-aviation-text-primary hover:bg-aviation-bg-instrument'
              )}
            >
              {type === 'all' ? 'All' : type.charAt(0).toUpperCase() + type.slice(1)}
            </button>
          ))}
        </div>
      </div>
      
      {/* Entries List */}
      <div className="flex-1 overflow-y-auto">
        {filteredEntries.map((entry) => (
          <div
            key={entry.id}
            onClick={() => onEntrySelect?.(entry)}
            className={cn(
              'px-4 py-3 border-b border-aviation-border-panel cursor-pointer transition-colors',
              selectedEntryId === entry.id ? 'bg-aviation-bg-instrument' : 'hover:bg-aviation-bg-secondary'
            )}
          >
            <div className="flex items-start justify-between mb-2">
              <div className="flex items-center gap-2">
                <span className={cn('flex items-center justify-center', getTypeColor(entry.semanticType))}>
                  {getTypeIcon(entry.semanticType)}
                </span>
                <span className="text-xs text-aviation-text-muted uppercase">{entry.semanticType}</span>
              </div>
              <div className="flex items-center gap-2">
                <div className={cn(
                  'flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px]',
                  entry.confidence > 0.8 ? 'bg-green-500/20 text-green-400' :
                  entry.confidence > 0.5 ? 'bg-amber-500/20 text-amber-400' :
                  'bg-gray-500/20 text-gray-400'
                )}>
                  <span>{Math.round(entry.confidence * 100)}%</span>
                </div>
                <button
                  onClick={(e) => { e.stopPropagation(); onEntryDelete?.(entry.id); }}
                  className="p-1 hover:bg-aviation-bg-panel rounded opacity-0 group-hover:opacity-100"
                >
                  <Trash2 className="w-3 h-3 text-aviation-text-muted" />
                </button>
              </div>
            </div>
            
            <p className="text-sm text-aviation-text-primary mb-2 line-clamp-2">{entry.content}</p>
            
            <div className="flex items-center gap-3 text-[10px] text-aviation-text-dim">
              {entry.source && <span>{entry.source}</span>}
              <span>{formatTimestamp(entry.timestamp)}</span>
              {entry.accessCount !== undefined && (
                <span className="flex items-center gap-1">
                  <Eye className="w-3 h-3" /> {entry.accessCount}
                </span>
              )}
            </div>
            
            {entry.tags && entry.tags.length > 0 && (
              <div className="flex items-center gap-1 mt-2">
                {entry.tags.slice(0, 3).map((tag) => (
                  <span key={tag} className="px-1.5 py-0.5 bg-aviation-bg-secondary rounded text-[10px] text-aviation-text-muted">
                    {tag}
                  </span>
                ))}
                {entry.tags.length > 3 && (
                  <span className="text-[10px] text-aviation-text-dim">+{entry.tags.length - 3}</span>
                )}
              </div>
            )}
          </div>
        ))}
        
        {filteredEntries.length === 0 && (
          <div className="flex flex-col items-center justify-center py-12 text-aviation-text-muted">
            <Brain className="w-8 h-8 mb-2 opacity-50" />
            <p className="text-sm">No memories found</p>
          </div>
        )}
      </div>
    </div>
  );
};

// ============================================================================
// Long Term Context Explorer
// ============================================================================

interface LongTermContextExplorerProps {
  chunks: Array<{
    id: string;
    content: string;
    timestamp: number;
    importance: number;
    decayScore?: number;
    retentionPriority?: 'critical' | 'high' | 'medium' | 'low';
    retrievalCount?: number;
  }>;
  selectedChunkId?: string | null;
  focusArea?: 'recent' | 'important' | 'decaying' | 'goals';
  onChunkSelect?: (chunk: { id: string; content: string }) => void;
  onChunkExpand?: (chunkId: string) => void;
  onMemoryReinforce?: (chunkId: string) => void;
  className?: string;
}

export const LongTermContextExplorer: React.FC<LongTermContextExplorerProps> = ({
  chunks,
  selectedChunkId = null,
  focusArea = 'important',
  onChunkSelect,
  onChunkExpand,
  onMemoryReinforce,
  className,
}) => {
  const [expandedId, setExpandedId] = useState<string | null>(null);
  
  const sortedChunks = useMemo(() => {
    const sorted = [...chunks];
    switch (focusArea) {
      case 'recent':
        return sorted.sort((a, b) => b.timestamp - a.timestamp);
      case 'important':
        return sorted.sort((a, b) => b.importance - a.importance);
      case 'decaying':
        return sorted.sort((a, b) => (b.decayScore || 0) - (a.decayScore || 0));
      case 'goals':
        return sorted.sort((a, b) => (b.retrievalCount || 0) - (a.retrievalCount || 0));
      default:
        return sorted;
    }
  }, [chunks, focusArea]);
  
  const getPriorityColor = (priority?: string) => {
    const colors: Record<string, string> = {
      critical: 'bg-red-500/20 border-red-500/50',
      high: 'bg-amber-500/20 border-amber-500/50',
      medium: 'bg-blue-500/20 border-blue-500/50',
      low: 'bg-gray-500/20 border-gray-500/50',
    };
    return colors[priority || 'low'] || colors.low;
  };
  
  const formatAge = (timestamp: number) => {
    const now = Date.now();
    const diff = now - timestamp;
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));
    const hours = Math.floor(diff / (1000 * 60 * 60));
    const minutes = Math.floor(diff / (1000 * 60));
    
    if (days > 0) return `${days}d ago`;
    if (hours > 0) return `${hours}h ago`;
    return `${minutes}m ago`;
  };
  
  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2">
            <Archive className="w-5 h-5 text-aviation-amber" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Long Term Context</h3>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-xs text-aviation-text-muted">{chunks.length} chunks</span>
          </div>
        </div>
        
        {/* Focus Area Tabs */}
        <div className="flex items-center gap-1">
          {(['important', 'recent', 'decaying', 'goals'] as const).map((area) => (
            <button
              key={area}
              onClick={() => {}}
              className={cn(
                'px-3 py-1.5 text-xs rounded transition-colors flex items-center gap-1.5',
                focusArea === area
                  ? 'bg-aviation-amber/20 text-aviation-amber border border-aviation-amber/30'
                  : 'text-aviation-text-muted hover:text-aviation-text-primary hover:bg-aviation-bg-instrument'
              )}
            >
              {area === 'important' && <TrendingUp className="w-3 h-3" />}
              {area === 'recent' && <Clock className="w-3 h-3" />}
              {area === 'decaying' && <TrendingDown className="w-3 h-3" />}
              {area === 'goals' && <Target className="w-3 h-3" />}
              {area.charAt(0).toUpperCase() + area.slice(1)}
            </button>
          ))}
        </div>
      </div>
      
      {/* Timeline */}
      <div className="flex-1 overflow-y-auto p-4">
        <div className="relative">
          {/* Timeline Line */}
          <div className="absolute left-4 top-0 bottom-0 w-px bg-aviation-border-panel" />
          
          {/* Chunks */}
          <div className="space-y-3">
            {sortedChunks.map((chunk) => {
              const isExpanded = expandedId === chunk.id;
              const isSelected = selectedChunkId === chunk.id;
              
              return (
                <div key={chunk.id} className="relative pl-10">
                  {/* Timeline Dot */}
                  <div className={cn(
                    'absolute left-2.5 top-4 w-3 h-3 rounded-full border-2 transition-colors',
                    isSelected ? 'bg-aviation-amber border-aviation-amber' : 'bg-aviation-bg-panel border-aviation-text-muted'
                  )} />
                  
                  {/* Card */}
                  <div
                    onClick={() => onChunkSelect?.(chunk)}
                    className={cn(
                      'p-3 bg-aviation-bg-secondary rounded-lg border transition-all cursor-pointer',
                      isSelected ? 'border-aviation-amber/50' : 'border-aviation-border-panel hover:border-aviation-text-muted'
                    )}
                  >
                    <div className="flex items-start justify-between mb-2">
                      <div className="flex items-center gap-2">
                        {chunk.retentionPriority && (
                          <span className={cn(
                            'px-1.5 py-0.5 text-[10px] rounded border uppercase',
                            getPriorityColor(chunk.retentionPriority)
                          )}>
                            {chunk.retentionPriority}
                          </span>
                        )}
                        <span className="text-xs text-aviation-text-dim">{formatAge(chunk.timestamp)}</span>
                      </div>
                      <div className="flex items-center gap-2">
                        {chunk.decayScore !== undefined && (
                          <span className={cn(
                            'flex items-center gap-1 text-[10px]',
                            chunk.decayScore > 0.7 ? 'text-red-400' :
                            chunk.decayScore > 0.4 ? 'text-amber-400' :
                            'text-aviation-text-muted'
                          )}>
                            <TrendingDown className="w-3 h-3" />
                            {Math.round(chunk.decayScore * 100)}%
                          </span>
                        )}
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            setExpandedId(isExpanded ? null : chunk.id);
                            onChunkExpand?.(chunk.id);
                          }}
                          className="p-1 hover:bg-aviation-bg-panel rounded"
                        >
                          {isExpanded ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}
                        </button>
                      </div>
                    </div>
                    
                    <p className={cn(
                      'text-sm text-aviation-text-primary transition-all',
                      isExpanded ? '' : 'line-clamp-2'
                    )}>
                      {chunk.content}
                    </p>
                    
                    <div className="flex items-center justify-between mt-2 pt-2 border-t border-aviation-border-panel">
                      <div className="flex items-center gap-3">
                        <span className="text-[10px] text-aviation-text-dim flex items-center gap-1">
                          <Eye className="w-3 h-3" />
                          {chunk.retrievalCount || 0} retrievals
                        </span>
                        <span className="text-[10px] text-aviation-text-dim flex items-center gap-1">
                          <Activity className="w-3 h-3" />
                          {Math.round(chunk.importance * 100)}% importance
                        </span>
                      </div>
                      <button
                        onClick={(e) => { e.stopPropagation(); onMemoryReinforce?.(chunk.id); }}
                        className="flex items-center gap-1 px-2 py-1 text-xs text-aviation-cyan hover:bg-aviation-cyan/10 rounded transition-colors"
                      >
                        <RefreshCw className="w-3 h-3" />
                        Reinforce
                      </button>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Memory Recall Timeline
// ============================================================================

interface MemoryRecallEvent {
  id: string;
  timestamp: number;
  type: 'retrieval' | 'reinforcement' | 'decay' | 'consolidation' | 'transfer';
  memoryId: string;
  memoryLabel: string;
  strength: number;
  context?: string;
}

interface MemoryRecallTimelineProps {
  events: MemoryRecallEvent[];
  selectedEventId?: string | null;
  timeRange?: { start: number; end: number };
  onEventSelect?: (event: MemoryRecallEvent) => void;
  onEventHover?: (event: MemoryRecallEvent | null) => void;
  className?: string;
}

export const MemoryRecallTimeline: React.FC<MemoryRecallTimelineProps> = ({
  events,
  selectedEventId = null,
  timeRange,
  onEventSelect,
  onEventHover,
  className,
}) => {
  const [hoveredEvent, setHoveredEvent] = useState<string | null>(null);
  
  const filteredEvents = useMemo(() => {
    if (!timeRange) return events;
    return events.filter(e => e.timestamp >= timeRange.start && e.timestamp <= timeRange.end);
  }, [events, timeRange]);
  
  const getEventIcon = (type: string) => {
    switch (type) {
      case 'retrieval': return <Eye className="w-4 h-4" />;
      case 'reinforcement': return <RefreshCw className="w-4 h-4" />;
      case 'decay': return <TrendingDown className="w-4 h-4" />;
      case 'consolidation': return <GitMerge className="w-4 h-4" />;
      case 'transfer': return <Share2 className="w-4 h-4" />;
      default: return <Activity className="w-4 h-4" />;
    }
  };
  
  const getEventColor = (type: string) => {
    const colors: Record<string, string> = {
      retrieval: 'text-aviation-cyan',
      reinforcement: 'text-green-400',
      decay: 'text-red-400',
      consolidation: 'text-purple-400',
      transfer: 'text-aviation-amber',
    };
    return colors[type] || 'text-gray-400';
  };
  
  const formatTime = (timestamp: number) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  };
  
  const groupedByHour = useMemo(() => {
    const groups: Record<string, MemoryRecallEvent[]> = {};
    filteredEvents.forEach(event => {
      const hour = new Date(event.timestamp).toLocaleString('en-US', { hour: 'numeric', hour12: true });
      if (!groups[hour]) groups[hour] = [];
      groups[hour].push(event);
    });
    return groups;
  }, [filteredEvents]);
  
  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Activity className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Memory Recall</h3>
          </div>
          <span className="text-xs text-aviation-text-muted">{filteredEvents.length} events</span>
        </div>
      </div>
      
      {/* Timeline */}
      <div className="flex-1 overflow-y-auto p-4">
        <div className="space-y-6">
          {Object.entries(groupedByHour).map(([hour, hourEvents]) => (
            <div key={hour}>
              <div className="flex items-center gap-2 mb-3">
                <Clock className="w-4 h-4 text-aviation-text-dim" />
                <span className="text-xs font-medium text-aviation-text-muted">{hour}</span>
                <div className="flex-1 h-px bg-aviation-border-panel" />
              </div>
              
              <div className="space-y-2 ml-4">
                {hourEvents.map((event) => {
                  const isSelected = selectedEventId === event.id;
                  const isHovered = hoveredEvent === event.id;
                  
                  return (
                    <div
                      key={event.id}
                      onClick={() => onEventSelect?.(event)}
                      onMouseEnter={() => {
                        setHoveredEvent(event.id);
                        onEventHover?.(event);
                      }}
                      onMouseLeave={() => {
                        setHoveredEvent(null);
                        onEventHover?.(null);
                      }}
                      className={cn(
                        'p-3 rounded-lg border transition-all cursor-pointer',
                        isSelected ? 'bg-aviation-bg-instrument border-aviation-cyan/50' :
                        isHovered ? 'bg-aviation-bg-secondary border-aviation-border-panel' :
                        'bg-aviation-bg-secondary/50 border-transparent hover:border-aviation-border-panel'
                      )}
                    >
                      <div className="flex items-center justify-between mb-2">
                        <div className="flex items-center gap-2">
                          <span className={getEventColor(event.type)}>
                            {getEventIcon(event.type)}
                          </span>
                          <span className="text-xs text-aviation-text-muted uppercase">{event.type}</span>
                          <span className="text-xs text-aviation-text-dim">{formatTime(event.timestamp)}</span>
                        </div>
                        <div className={cn(
                          'w-2 h-2 rounded-full',
                          event.strength > 0.7 ? 'bg-green-400' :
                          event.strength > 0.4 ? 'bg-amber-400' :
                          'bg-red-400'
                        )} />
                      </div>
                      
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-aviation-text-primary font-medium">{event.memoryLabel}</span>
                        <span className="text-xs text-aviation-text-dim">{Math.round(event.strength * 100)}%</span>
                      </div>
                      
                      {event.context && (
                        <p className="mt-2 text-xs text-aviation-text-muted">{event.context}</p>
                      )}
                    </div>
                  );
                })}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Knowledge Cluster Map
// ============================================================================

interface KnowledgeCluster {
  id: string;
  name: string;
  description?: string;
  centralTopic: string;
  members: string[];
  importance: number;
  coherence?: number;
  creationTimestamp: number;
}

interface KnowledgeClusterMapProps {
  clusters: KnowledgeCluster[];
  selectedClusterId?: string | null;
  highlightedClusters?: string[];
  onClusterSelect?: (cluster: KnowledgeCluster) => void;
  onClusterHover?: (cluster: KnowledgeCluster | null) => void;
  onClusterMerge?: (clusterIds: string[]) => void;
  onClusterSplit?: (clusterId: string) => void;
  className?: string;
}

export const KnowledgeClusterMap: React.FC<KnowledgeClusterMapProps> = ({
  clusters,
  selectedClusterId = null,
  highlightedClusters = [],
  onClusterSelect,
  onClusterHover,
  onClusterMerge,
  onClusterSplit,
  className,
}) => {
  const [hoveredCluster, setHoveredCluster] = useState<string | null>(null);
  
  const clusterPositions = useMemo(() => {
    const positions: Record<string, { x: number; y: number }> = {};
    const centerX = 250;
    const centerY = 200;
    
    clusters.forEach((cluster, index) => {
      const angle = (2 * Math.PI * index) / clusters.length;
      const radius = 100 + (cluster.importance * 60);
      positions[cluster.id] = {
        x: centerX + radius * Math.cos(angle),
        y: centerY + radius * Math.sin(angle),
      };
    });
    return positions;
  }, [clusters]);
  
  const selectedCluster = clusters.find(c => c.id === selectedClusterId);
  
  return (
    <div className={cn('relative h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Header */}
      <div className="absolute top-3 left-3 flex items-center gap-2 z-10">
        <div className="flex items-center gap-1.5 px-2 py-1 bg-aviation-bg-instrument rounded border border-aviation-border-panel">
          <Hexagon className="w-4 h-4 text-purple-400" />
          <span className="text-xs text-aviation-text-primary font-medium">Knowledge Clusters</span>
        </div>
      </div>
      
      <svg className="w-full h-full" viewBox="0 0 500 400">
        {/* Connection lines between nearby clusters */}
        {clusters.map((cluster, i) => {
          const pos = clusterPositions[cluster.id];
          if (!pos) return null;
          
          // Find related clusters (simplified - just closest 2)
          const sortedByDist = clusters
            .filter(c => c.id !== cluster.id)
            .map(c => ({
              cluster: c,
              dist: Math.hypot(
                (clusterPositions[c.id]?.x || 0) - pos.x,
                (clusterPositions[c.id]?.y || 0) - pos.y
              ),
            }))
            .sort((a, b) => a.dist - b.dist)
            .slice(0, 2);
          
          return sortedByDist.map(({ cluster: targetCluster }) => {
            const targetPos = clusterPositions[targetCluster.id];
            if (!targetPos) return null;
            
            return (
              <line
                key={`${cluster.id}-${targetCluster.id}`}
                x1={pos.x}
                y1={pos.y}
                x2={targetPos.x}
                y2={targetPos.y}
                stroke="var(--color-aviation-border-panel, rgba(255,255,255,0.08))"
                strokeWidth={1}
                strokeDasharray="4 4"
                opacity={0.5}
              />
            );
          });
        })}
        
        {/* Clusters */}
        {clusters.map((cluster) => {
          const pos = clusterPositions[cluster.id];
          if (!pos) return null;
          
          const isSelected = selectedClusterId === cluster.id;
          const isHighlighted = highlightedClusters.includes(cluster.id);
          const isHovered = hoveredCluster === cluster.id;
          const size = 40 + (cluster.importance * 30);
          
          return (
            <g
              key={cluster.id}
              onClick={() => onClusterSelect?.(cluster)}
              onMouseEnter={() => {
                setHoveredCluster(cluster.id);
                onClusterHover?.(cluster);
              }}
              onMouseLeave={() => {
                setHoveredCluster(null);
                onClusterHover?.(null);
              }}
              className="cursor-pointer"
            >
              {/* Outer ring */}
              <circle
                cx={pos.x}
                cy={pos.y}
                r={size + 8}
                fill="none"
                stroke={isSelected ? '#06b6d4' : isHovered ? '#a855f7' : 'var(--color-aviation-border-panel, rgba(255,255,255,0.08))'}
                strokeWidth={isSelected ? 2 : isHovered ? 1.5 : 1}
                className="transition-all"
                opacity={isHighlighted ? 1 : highlightedClusters.length > 0 ? 0.3 : 0.8}
              />

              {/* Main circle */}
              <circle
                cx={pos.x}
                cy={pos.y}
                r={size}
                fill="rgba(168, 85, 247, 0.2)"
                stroke="#a855f7"
                strokeWidth={isSelected ? 2 : 1}
                className="transition-all"
              />

              {/* Inner decoration */}
              <circle
                cx={pos.x}
                cy={pos.y}
                r={size * 0.6}
                fill="rgba(168, 85, 247, 0.1)"
              />
              
              {/* Label */}
              <text
                x={pos.x}
                y={pos.y}
                textAnchor="middle"
                dominantBaseline="middle"
                fill="var(--color-aviation-text-primary, #e8e8f0)"
                fontSize={12}
                fontWeight={500}
                fontFamily="var(--font-sans, ui-sans-serif, system-ui, sans-serif)"
              >
                {cluster.name.length > 10 ? cluster.name.slice(0, 10) + '...' : cluster.name}
              </text>

              {/* Member count */}
              <text
                x={pos.x}
                y={pos.y + size + 16}
                textAnchor="middle"
                fill="var(--color-aviation-text-dim, #4b5563)"
                fontSize={10}
                fontFamily="var(--font-sans, ui-sans-serif, system-ui, sans-serif)"
              >
                {cluster.members.length} members
              </text>
            </g>
          );
        })}
      </svg>
      
      {/* Selected Cluster Details */}
      {selectedCluster && (
        <div className="absolute bottom-3 left-3 right-3 p-4 bg-aviation-bg-secondary/90 rounded-lg border border-aviation-border-panel backdrop-blur-sm">
          <div className="flex items-start justify-between">
            <div>
              <h4 className="text-sm font-medium text-aviation-text-primary">{selectedCluster.name}</h4>
              {selectedCluster.description && (
                <p className="text-xs text-aviation-text-muted mt-1">{selectedCluster.description}</p>
              )}
              <div className="flex items-center gap-3 mt-2 text-[10px] text-aviation-text-dim">
                <span>{selectedCluster.members.length} members</span>
                {selectedCluster.coherence !== undefined && (
                  <span>{Math.round(selectedCluster.coherence * 100)}% coherence</span>
                )}
              </div>
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={() => onClusterMerge?.([selectedCluster.id])}
                className="p-1.5 hover:bg-aviation-bg-panel rounded"
                title="Merge cluster"
              >
                <GitMerge className="w-4 h-4 text-aviation-text-muted" />
              </button>
              <button
                onClick={() => onClusterSplit?.(selectedCluster.id)}
                className="p-1.5 hover:bg-aviation-bg-panel rounded"
                title="Split cluster"
              >
                <Share2 className="w-4 h-4 text-aviation-text-muted" />
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

// ============================================================================
// Memory Decay Visualizer
// ============================================================================

interface DecayNode {
  id: string;
  label: string;
  age: number;
  strength: number;
  decayRate: number;
  type: 'episodic' | 'semantic' | 'procedural';
  lastReinforced?: number;
  nextDecay?: number;
}

interface MemoryDecayVisualizerProps {
  nodes: DecayNode[];
  selectedNodeId?: string | null;
  showPredictions?: boolean;
  onNodeSelect?: (node: DecayNode) => void;
  onReinforce?: (nodeId: string) => void;
  className?: string;
}

export const MemoryDecayVisualizer: React.FC<MemoryDecayVisualizerProps> = ({
  nodes,
  selectedNodeId = null,
  showPredictions = false,
  onNodeSelect,
  onReinforce,
  className,
}) => {
  const [hoveredNode, setHoveredNode] = useState<string | null>(null);
  
  const getTypeColor = (type: string) => {
    const colors: Record<string, string> = {
      episodic: 'bg-blue-500/20 border-blue-500',
      semantic: 'bg-green-500/20 border-green-500',
      procedural: 'bg-amber-500/20 border-amber-500',
    };
    return colors[type] || 'bg-gray-500/20 border-gray-500';
  };
  
  const getStrengthColor = (strength: number) => {
    if (strength > 0.7) return 'bg-green-500';
    if (strength > 0.4) return 'bg-amber-500';
    return 'bg-red-500';
  };
  
  const sortedNodes = useMemo(() => {
    return [...nodes].sort((a, b) => b.age - a.age);
  }, [nodes]);
  
  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <TrendingDown className="w-5 h-5 text-red-400" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Memory Decay</h3>
          </div>
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2 text-[10px]">
              <span className="text-aviation-text-dim">Legend:</span>
              <span className="flex items-center gap-1">
                <div className="w-2 h-2 rounded bg-blue-500" />
                <span className="text-aviation-text-muted">Episodic</span>
              </span>
              <span className="flex items-center gap-1">
                <div className="w-2 h-2 rounded bg-green-500" />
                <span className="text-aviation-text-muted">Semantic</span>
              </span>
              <span className="flex items-center gap-1">
                <div className="w-2 h-2 rounded bg-amber-500" />
                <span className="text-aviation-text-muted">Procedural</span>
              </span>
            </div>
          </div>
        </div>
      </div>
      
      {/* Decay Bars */}
      <div className="flex-1 overflow-y-auto p-4">
        <div className="space-y-3">
          {sortedNodes.map((node) => {
            const isSelected = selectedNodeId === node.id;
            const isHovered = hoveredNode === node.id;
            
            return (
              <div
                key={node.id}
                onClick={() => onNodeSelect?.(node)}
                onMouseEnter={() => setHoveredNode(node.id)}
                onMouseLeave={() => setHoveredNode(null)}
                className={cn(
                  'p-3 rounded-lg border transition-all cursor-pointer',
                  isSelected ? 'bg-aviation-bg-instrument border-aviation-cyan/50' :
                  isHovered ? 'bg-aviation-bg-secondary border-aviation-border-panel' :
                  'bg-aviation-bg-secondary/50 border-transparent'
                )}
              >
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <div className={cn('w-2 h-2 rounded', getStrengthColor(node.strength))} />
                    <span className="text-sm text-aviation-text-primary font-medium">{node.label}</span>
                    <span className={cn(
                      'px-1.5 py-0.5 text-[10px] rounded border uppercase',
                      getTypeColor(node.type)
                    )}>
                      {node.type}
                    </span>
                  </div>
                  <div className="flex items-center gap-3 text-[10px] text-aviation-text-dim">
                    <span>Age: {node.age}d</span>
                    <span>Decay: {Math.round(node.decayRate * 100)}%/day</span>
                    {node.nextDecay && (
                      <span>Next: {new Date(node.nextDecay).toLocaleDateString()}</span>
                    )}
                  </div>
                </div>
                
                {/* Strength Bar */}
                <div className="relative h-2 bg-aviation-bg-instrument rounded overflow-hidden mb-2">
                  <div
                    className={cn('h-full transition-all', getStrengthColor(node.strength))}
                    style={{ width: `${node.strength * 100}%` }}
                  />
                  {showPredictions && node.nextDecay && (
                    <div
                      className="absolute top-0 h-full bg-red-500/30"
                      style={{ left: `${node.strength * 100}%`, width: '20%' }}
                    />
                  )}
                </div>
                
                <div className="flex items-center justify-between">
                  <span className="text-xs text-aviation-text-muted">
                    Strength: {Math.round(node.strength * 100)}%
                  </span>
                  {isHovered && (
                    <button
                      onClick={(e) => { e.stopPropagation(); onReinforce?.(node.id); }}
                      className="flex items-center gap-1 px-2 py-1 text-xs text-green-400 hover:bg-green-500/10 rounded transition-colors"
                    >
                      <RefreshCw className="w-3 h-3" />
                      Reinforce
                    </button>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// Vector Embedding Explorer
// ============================================================================

interface EmbeddingVector {
  id: string;
  values: number[];
  label: string;
  dimension: number;
  magnitude?: number;
  timestamp?: number;
  source?: string;
  neighbors?: string[];
}

interface VectorEmbeddingExplorerProps {
  vectors: EmbeddingVector[];
  selectedVectorId?: string | null;
  highlightedVectors?: string[];
  metric?: 'cosine' | 'euclidean' | 'dot';
  onVectorSelect?: (vector: EmbeddingVector) => void;
  onClusterAnalyze?: () => void;
  className?: string;
}

export const VectorEmbeddingExplorer: React.FC<VectorEmbeddingExplorerProps> = ({
  vectors,
  selectedVectorId = null,
  highlightedVectors = [],
  metric = 'cosine',
  onVectorSelect,
  onClusterAnalyze,
  className,
}) => {
  const [searchQuery, setSearchQuery] = useState('');
  
  const filteredVectors = useMemo(() => {
    if (!searchQuery) return vectors;
    return vectors.filter(v => 
      v.label.toLowerCase().includes(searchQuery.toLowerCase()) ||
      v.source?.toLowerCase().includes(searchQuery.toLowerCase())
    );
  }, [vectors, searchQuery]);
  
  const selectedVector = vectors.find(v => v.id === selectedVectorId);
  
  // Simple 2D projection using first 2 dimensions for visualization
  const vectorPositions = useMemo(() => {
    return filteredVectors.map((v, i) => ({
      id: v.id,
      x: 50 + (v.values[0] || 0) * 40 + (i % 3) * 10,
      y: 50 + (v.values[1] || 0) * 40 + Math.floor(i / 3) * 10,
    }));
  }, [filteredVectors]);
  
  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2">
            <Binary className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Vector Embeddings</h3>
          </div>
          <button
            onClick={onClusterAnalyze}
            className="flex items-center gap-1.5 px-3 py-1.5 bg-aviation-cyan/10 text-aviation-cyan text-xs rounded hover:bg-aviation-cyan/20 transition-colors"
          >
            <Grid3x3 className="w-3 h-3" />
            Cluster Analysis
          </button>
        </div>
        
        {/* Search and Metric */}
        <div className="flex items-center gap-3">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-aviation-text-muted" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search vectors..."
              className="w-full pl-9 pr-3 py-2 bg-aviation-bg-instrument border border-aviation-border-panel rounded text-sm text-aviation-text-primary placeholder:text-aviation-text-dim focus:outline-none focus:border-aviation-cyan"
            />
          </div>
          <select
            value={metric}
            onChange={(e) => {}}
            className="px-3 py-2 bg-aviation-bg-instrument border border-aviation-border-panel rounded text-sm text-aviation-text-primary focus:outline-none"
          >
            <option value="cosine">Cosine</option>
            <option value="euclidean">Euclidean</option>
            <option value="dot">Dot Product</option>
          </select>
        </div>
      </div>
      
      {/* Visualization Area */}
      <div className="flex-1 p-4">
        <div className="relative h-full bg-aviation-bg-secondary rounded-lg overflow-hidden">
          {/* Grid */}
          <div className="absolute inset-0 opacity-20">
            <svg className="w-full h-full">
              {[...Array(10)].map((_, i) => (
                <g key={i}>
                  <line x1={`${i * 10}%`} y1="0" x2={`${i * 10}%`} y2="100%" stroke="var(--color-aviation-border-panel, rgba(255,255,255,0.08))" strokeWidth={0.5} />
                  <line x1="0" y1={`${i * 10}%`} x2="100%" y2={`${i * 10}%`} stroke="var(--color-aviation-border-panel, rgba(255,255,255,0.08))" strokeWidth={0.5} />
                </g>
              ))}
            </svg>
          </div>
          
          {/* Vectors */}
          {vectorPositions.map((pos, i) => {
            const vector = filteredVectors[i];
            const isSelected = selectedVectorId === vector.id;
            const isHighlighted = highlightedVectors.includes(vector.id);
            
            return (
              <div
                key={vector.id}
                onClick={() => onVectorSelect?.(vector)}
                className={cn(
                  'absolute transition-all cursor-pointer',
                  isSelected ? 'z-20' : 'z-10'
                )}
                style={{ left: `${pos.x}%`, top: `${pos.y}%` }}
              >
                <div className={cn(
                  'w-8 h-8 rounded-lg border-2 flex items-center justify-center transition-all',
                  isSelected ? 'bg-aviation-cyan border-aviation-cyan scale-125' :
                  isHighlighted ? 'bg-aviation-cyan/50 border-aviation-cyan/50' :
                  'bg-aviation-bg-panel border-aviation-text-muted hover:border-aviation-cyan'
                )}>
                  <span className="text-[10px] font-bold text-aviation-text-primary">
                    {vector.dimension > 0 ? vector.values.slice(0, 2).join(',') : '?'}
                  </span>
                </div>
                <span className="absolute top-full left-1/2 -translate-x-1/2 mt-1 text-[10px] text-aviation-text-muted whitespace-nowrap">
                  {vector.label.length > 8 ? vector.label.slice(0, 8) : vector.label}
                </span>
              </div>
            );
          })}
        </div>
      </div>
      
      {/* Selected Vector Details */}
      {selectedVector && (
        <div className="px-4 py-3 border-t border-aviation-border-panel">
          <div className="flex items-start justify-between">
            <div>
              <h4 className="text-sm font-medium text-aviation-text-primary">{selectedVector.label}</h4>
              <div className="flex items-center gap-3 mt-1 text-[10px] text-aviation-text-dim">
                <span>{selectedVector.dimension} dims</span>
                {selectedVector.magnitude !== undefined && (
                  <span>Mag: {selectedVector.magnitude.toFixed(3)}</span>
                )}
                {selectedVector.source && <span>{selectedVector.source}</span>}
              </div>
            </div>
            <div className="text-[10px] font-mono text-aviation-text-muted">
              [{selectedVector.values.slice(0, 5).map(v => v.toFixed(2)).join(', ')}...]
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

// ============================================================================
// Shared Agent Memory Panel
// ============================================================================

interface AgentMemory {
  agentId: string;
  agentName: string;
  role?: string;
  memoryCapacity?: number;
  usedCapacity?: number;
  activeMemories?: string[];
  sharedWith?: string[];
  lastActive?: number;
}

interface SharedAgentMemoryPanelProps {
  agents: AgentMemory[];
  selectedAgentId?: string | null;
  sharedMemoryIds?: string[];
  onAgentSelect?: (agent: AgentMemory) => void;
  onShareMemory?: (memoryId: string, agentIds: string[]) => void;
  onRetrieveShared?: (memoryId: string) => void;
  className?: string;
}

export const SharedAgentMemoryPanel: React.FC<SharedAgentMemoryPanelProps> = ({
  agents,
  selectedAgentId = null,
  sharedMemoryIds = [],
  onAgentSelect,
  onShareMemory,
  onRetrieveShared,
  className,
}) => {
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  
  const selectedAgent = agents.find(a => a.agentId === selectedAgentId);
  
  const getCapacityColor = (used?: number, total?: number) => {
    if (!used || !total) return 'bg-aviation-text-muted';
    const ratio = used / total;
    if (ratio > 0.9) return 'bg-red-500';
    if (ratio > 0.7) return 'bg-amber-500';
    return 'bg-green-500';
  };
  
  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2">
            <Users className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Shared Agent Memory</h3>
          </div>
          <div className="flex items-center gap-1">
            <button
              onClick={() => setViewMode('grid')}
              className={cn(
                'p-1.5 rounded transition-colors',
                viewMode === 'grid' ? 'bg-aviation-cyan/20 text-aviation-cyan' : 'text-aviation-text-muted hover:bg-aviation-bg-instrument'
              )}
            >
              <LayoutGrid className="w-4 h-4" />
            </button>
            <button
              onClick={() => setViewMode('list')}
              className={cn(
                'p-1.5 rounded transition-colors',
                viewMode === 'list' ? 'bg-aviation-cyan/20 text-aviation-cyan' : 'text-aviation-text-muted hover:bg-aviation-bg-instrument'
              )}
            >
              <List className="w-4 h-4" />
            </button>
          </div>
        </div>
        <p className="text-xs text-aviation-text-muted">
          {agents.length} agents connected • {sharedMemoryIds.length} shared memories
        </p>
      </div>
      
      {/* Agent Grid/List */}
      <div className="flex-1 overflow-y-auto p-4">
        {viewMode === 'grid' ? (
          <div className="grid grid-cols-2 gap-3">
            {agents.map((agent) => {
              const isSelected = selectedAgentId === agent.agentId;
              const capacityRatio = agent.memoryCapacity 
                ? (agent.usedCapacity || 0) / agent.memoryCapacity 
                : 0;
              
              return (
                <div
                  key={agent.agentId}
                  onClick={() => onAgentSelect?.(agent)}
                  className={cn(
                    'p-3 rounded-lg border transition-all cursor-pointer',
                    isSelected ? 'bg-aviation-bg-instrument border-aviation-cyan/50' :
                    'bg-aviation-bg-secondary border-aviation-border-panel hover:border-aviation-text-muted'
                  )}
                >
                  <div className="flex items-center gap-2 mb-3">
                    <div className="w-8 h-8 rounded-full bg-aviation-cyan/20 flex items-center justify-center">
                      <Bot className="w-4 h-4 text-aviation-cyan" />
                    </div>
                    <div>
                      <h4 className="text-sm font-medium text-aviation-text-primary">{agent.agentName}</h4>
                      {agent.role && <span className="text-[10px] text-aviation-text-dim">{agent.role}</span>}
                    </div>
                  </div>
                  
                  {/* Capacity Bar */}
                  <div className="mb-2">
                    <div className="flex items-center justify-between text-[10px] text-aviation-text-muted mb-1">
                      <span>Memory</span>
                      <span>{agent.usedCapacity || 0}/{agent.memoryCapacity || '?'}</span>
                    </div>
                    <div className="h-1.5 bg-aviation-bg-instrument rounded overflow-hidden">
                      <div
                        className={cn('h-full transition-all', getCapacityColor(agent.usedCapacity, agent.memoryCapacity))}
                        style={{ width: `${capacityRatio * 100}%` }}
                      />
                    </div>
                  </div>
                  
                  <div className="flex items-center justify-between text-[10px] text-aviation-text-dim">
                    <span className="flex items-center gap-1">
                      <Eye className="w-3 h-3" />
                      {agent.activeMemories?.length || 0} active
                    </span>
                    <span className="flex items-center gap-1">
                      <Share2 className="w-3 h-3" />
                      {agent.sharedWith?.length || 0} shared
                    </span>
                  </div>
                </div>
              );
            })}
          </div>
        ) : (
          <div className="space-y-2">
            {agents.map((agent) => {
              const isSelected = selectedAgentId === agent.agentId;
              
              return (
                <div
                  key={agent.agentId}
                  onClick={() => onAgentSelect?.(agent)}
                  className={cn(
                    'flex items-center gap-3 p-3 rounded-lg border transition-all cursor-pointer',
                    isSelected ? 'bg-aviation-bg-instrument border-aviation-cyan/50' :
                    'bg-aviation-bg-secondary border-aviation-border-panel hover:border-aviation-text-muted'
                  )}
                >
                  <div className="w-8 h-8 rounded-full bg-aviation-cyan/20 flex items-center justify-center">
                    <Bot className="w-4 h-4 text-aviation-cyan" />
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium text-aviation-text-primary">{agent.agentName}</span>
                      {agent.role && <span className="text-[10px] text-aviation-text-dim">{agent.role}</span>}
                    </div>
                    <div className="flex items-center gap-4 mt-1 text-[10px] text-aviation-text-dim">
                      <span>{agent.activeMemories?.length || 0} active</span>
                      <span>{agent.sharedWith?.length || 0} shared</span>
                      {agent.lastActive && <span>Active {new Date(agent.lastActive).toLocaleTimeString()}</span>}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
      
      {/* Selected Agent Actions */}
      {selectedAgent && (
        <div className="px-4 py-3 border-t border-aviation-border-panel bg-aviation-bg-secondary">
          <div className="flex items-center justify-between">
            <div>
              <h4 className="text-sm font-medium text-aviation-text-primary">{selectedAgent.agentName}</h4>
              <p className="text-xs text-aviation-text-muted">Select memories to share</p>
            </div>
            <div className="flex items-center gap-2">
              <button className="flex items-center gap-1.5 px-3 py-1.5 bg-aviation-cyan/10 text-aviation-cyan text-xs rounded hover:bg-aviation-cyan/20 transition-colors">
                <Share2 className="w-3 h-3" />
                Share Memory
              </button>
              <button className="flex items-center gap-1.5 px-3 py-1.5 bg-aviation-bg-instrument text-aviation-text-primary text-xs rounded hover:bg-aviation-bg-panel transition-colors">
                <Download className="w-3 h-3" />
                Retrieve
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

// ============================================================================
// Memory Merge Tool
// ============================================================================

interface MemoryMergeCandidate {
  id: string;
  sourceMemoryIds: string[];
  targetMemoryId?: string;
  suggestedMerge?: {
    content: string;
    confidence: number;
  };
  overlapScore?: number;
  conflicts?: Array<{
    field: string;
    values: unknown[];
  }>;
}

interface MemoryMergeToolProps {
  candidates: MemoryMergeCandidate[];
  selectedCandidateId?: string | null;
  onCandidateSelect?: (candidate: MemoryMergeCandidate) => void;
  onMerge?: (candidateId: string) => void;
  onSplit?: (memoryId: string) => void;
  onDiscard?: (candidateId: string) => void;
  className?: string;
}

export const MemoryMergeTool: React.FC<MemoryMergeToolProps> = ({
  candidates,
  selectedCandidateId = null,
  onCandidateSelect,
  onMerge,
  onSplit,
  onDiscard,
  className,
}) => {
  const [showConflicts, setShowConflicts] = useState(false);
  
  const selectedCandidate = candidates.find(c => c.id === selectedCandidateId);
  
  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2">
            <GitMerge className="w-5 h-5 text-aviation-amber" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Memory Merge</h3>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-xs text-aviation-text-muted">{candidates.length} candidates</span>
            <button
              onClick={() => setShowConflicts(!showConflicts)}
              className={cn(
                'flex items-center gap-1.5 px-2 py-1 text-xs rounded transition-colors',
                showConflicts ? 'bg-red-500/20 text-red-400' : 'text-aviation-text-muted hover:bg-aviation-bg-instrument'
              )}
            >
              <AlertTriangle className="w-3 h-3" />
              {showConflicts ? 'Hide' : 'Show'} Conflicts
            </button>
          </div>
        </div>
      </div>
      
      {/* Candidates List */}
      <div className="flex-1 overflow-y-auto">
        {candidates.map((candidate) => {
          const isSelected = selectedCandidateId === candidate.id;
          const hasConflicts = candidate.conflicts && candidate.conflicts.length > 0;
          
          return (
            <div
              key={candidate.id}
              onClick={() => onCandidateSelect?.(candidate)}
              className={cn(
                'p-4 border-b border-aviation-border-panel cursor-pointer transition-colors',
                isSelected ? 'bg-aviation-bg-instrument' : 'hover:bg-aviation-bg-secondary'
              )}
            >
              <div className="flex items-start justify-between mb-2">
                <div className="flex items-center gap-2">
                  <div className={cn(
                    'w-6 h-6 rounded-full flex items-center justify-center',
                    hasConflicts ? 'bg-amber-500/20 text-amber-400' : 'bg-aviation-cyan/20 text-aviation-cyan'
                  )}>
                    <GitMerge className="w-3.5 h-3.5" />
                  </div>
                  <div>
                    <span className="text-sm font-medium text-aviation-text-primary">
                      {candidate.sourceMemoryIds.length} memories to merge
                    </span>
                    {candidate.overlapScore !== undefined && (
                      <span className="ml-2 text-xs text-aviation-text-dim">
                        {Math.round(candidate.overlapScore * 100)}% overlap
                      </span>
                    )}
                  </div>
                </div>
                {hasConflicts && (
                  <span className="flex items-center gap-1 px-1.5 py-0.5 bg-red-500/20 text-red-400 text-[10px] rounded">
                    <AlertTriangle className="w-3 h-3" />
                    {candidate.conflicts?.length} conflicts
                  </span>
                )}
              </div>
              
              {candidate.suggestedMerge && (
                <div className="mt-2 p-2 bg-aviation-bg-secondary rounded">
                  <p className="text-xs text-aviation-text-primary">{candidate.suggestedMerge.content}</p>
                  <div className="flex items-center justify-between mt-2">
                    <span className="text-[10px] text-aviation-text-dim">
                      Confidence: {Math.round(candidate.suggestedMerge.confidence * 100)}%
                    </span>
                    <div className="flex items-center gap-1">
                      <button
                        onClick={(e) => { e.stopPropagation(); onMerge?.(candidate.id); }}
                        className="px-2 py-1 text-xs text-aviation-cyan hover:bg-aviation-cyan/10 rounded transition-colors"
                      >
                        Merge
                      </button>
                      <button
                        onClick={(e) => { e.stopPropagation(); onDiscard?.(candidate.id); }}
                        className="px-2 py-1 text-xs text-aviation-text-muted hover:bg-aviation-bg-panel rounded transition-colors"
                      >
                        Discard
                      </button>
                    </div>
                  </div>
                </div>
              )}
            </div>
          );
        })}
        
        {candidates.length === 0 && (
          <div className="flex flex-col items-center justify-center py-12 text-aviation-text-muted">
            <GitMerge className="w-8 h-8 mb-2 opacity-50" />
            <p className="text-sm">No merge candidates</p>
            <p className="text-xs text-aviation-text-dim mt-1">Memories are automatically grouped for potential merging</p>
          </div>
        )}
      </div>
      
      {/* Selected Candidate Details */}
      {selectedCandidate && showConflicts && selectedCandidate.conflicts && (
        <div className="px-4 py-3 border-t border-aviation-border-panel bg-red-500/5">
          <h4 className="text-sm font-medium text-red-400 mb-2 flex items-center gap-1.5">
            <AlertTriangle className="w-4 h-4" />
            Merge Conflicts
          </h4>
          <div className="space-y-2">
            {selectedCandidate.conflicts.map((conflict, i) => (
              <div key={i} className="p-2 bg-aviation-bg-panel rounded border border-red-500/20">
                <span className="text-xs font-medium text-aviation-text-primary">{conflict.field}</span>
                <div className="flex items-center gap-2 mt-1">
                  {conflict.values.map((value, j) => (
                    <span key={j} className="px-2 py-0.5 bg-aviation-bg-secondary rounded text-[10px] text-aviation-text-muted">
                      {String(value)}
                    </span>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

// ============================================================================
// Conversation Memory Tree
// ============================================================================

interface ConversationNode {
  id: string;
  type: 'turn' | 'topic' | 'segment' | 'branch';
  content: string;
  speaker?: 'user' | 'agent' | 'system';
  timestamp: number;
  parentId?: string;
  children?: string[];
  summary?: string;
  sentiment?: 'positive' | 'neutral' | 'negative';
  intent?: string;
}

interface ConversationMemoryTreeProps {
  nodes: ConversationNode[];
  selectedNodeId?: string | null;
  focusedNodeId?: string;
  onNodeSelect?: (node: ConversationNode) => void;
  onBranch?: (nodeId: string, newContent: string) => void;
  onSummarize?: (nodeId: string) => void;
  className?: string;
}

export const ConversationMemoryTree: React.FC<ConversationMemoryTreeProps> = ({
  nodes,
  selectedNodeId = null,
  focusedNodeId,
  onNodeSelect,
  onBranch,
  onSummarize,
  className,
}) => {
  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(new Set([focusedNodeId || '']));
  
  const toggleExpand = (nodeId: string) => {
    const newExpanded = new Set(expandedNodes);
    if (newExpanded.has(nodeId)) {
      newExpanded.delete(nodeId);
    } else {
      newExpanded.add(nodeId);
    }
    setExpandedNodes(newExpanded);
  };
  
  const rootNodes = nodes.filter(n => !n.parentId);
  
  const renderNode = (node: ConversationNode, depth: number = 0) => {
    const isExpanded = expandedNodes.has(node.id);
    const hasChildren = node.children && node.children.length > 0;
    const isSelected = selectedNodeId === node.id;
    const isFocused = focusedNodeId === node.id;
    
    return (
      <div key={node.id}>
        <div
          onClick={() => onNodeSelect?.(node)}
          className={cn(
            'flex items-center gap-2 p-2 rounded cursor-pointer transition-colors',
            isSelected && 'bg-aviation-bg-instrument',
            isFocused && 'bg-aviation-cyan/10 border-l-2 border-aviation-cyan',
            !isSelected && !isFocused && 'hover:bg-aviation-bg-secondary'
          )}
          style={{ paddingLeft: `${depth * 20 + 8}px` }}
        >
          {/* Expand/Collapse */}
          {hasChildren ? (
            <button
              onClick={(e) => { e.stopPropagation(); toggleExpand(node.id); }}
              className="p-0.5 hover:bg-aviation-bg-panel rounded"
            >
              {isExpanded ? (
                <ChevronDown className="w-4 h-4 text-aviation-text-muted" />
              ) : (
                <ChevronRight className="w-4 h-4 text-aviation-text-muted" />
              )}
            </button>
          ) : (
            <div className="w-5" />
          )}
          
          {/* Speaker Icon */}
          <div className={cn(
            'w-5 h-5 rounded-full flex items-center justify-center',
            node.speaker === 'user' ? 'bg-aviation-cyan/20' :
            node.speaker === 'agent' ? 'bg-purple-500/20' :
            'bg-aviation-text-dim/20'
          )}>
            {node.speaker === 'user' ? (
              <User className="w-3 h-3 text-aviation-cyan" />
            ) : node.speaker === 'agent' ? (
              <Bot className="w-3 h-3 text-purple-400" />
            ) : (
              <Activity className="w-3 h-3 text-aviation-text-dim" />
            )}
          </div>
          
          {/* Content */}
          <div className="flex-1 min-w-0">
            <p className={cn(
              'text-sm text-aviation-text-primary truncate',
              node.type === 'topic' && 'font-medium'
            )}>
              {node.content}
            </p>
            <div className="flex items-center gap-2 mt-0.5">
              <span className="text-[10px] text-aviation-text-dim uppercase">{node.type}</span>
              {node.sentiment && (
                <span className={cn(
                  'text-[10px]',
                  node.sentiment === 'positive' && 'text-green-400',
                  node.sentiment === 'negative' && 'text-red-400',
                  node.sentiment === 'neutral' && 'text-aviation-text-dim'
                )}>
                  {node.sentiment}
                </span>
              )}
              {node.intent && (
                <span className="text-[10px] text-aviation-text-dim">• {node.intent}</span>
              )}
            </div>
          </div>
          
          {/* Actions */}
          <div className="flex items-center gap-1">
            <button
              onClick={(e) => { e.stopPropagation(); onBranch?.(node.id, ''); }}
              className="p-1 hover:bg-aviation-bg-panel rounded opacity-0 group-hover:opacity-100"
              title="Branch conversation"
            >
              <GitFork className="w-3.5 h-3.5 text-aviation-text-muted" />
            </button>
            <button
              onClick={(e) => { e.stopPropagation(); onSummarize?.(node.id); }}
              className="p-1 hover:bg-aviation-bg-panel rounded opacity-0 group-hover:opacity-100"
              title="Summarize"
            >
              <FileText className="w-3.5 h-3.5 text-aviation-text-muted" />
            </button>
          </div>
        </div>
        
        {/* Children */}
        {isExpanded && hasChildren && (
          <div>
            {node.children?.map(childId => {
              const child = nodes.find(n => n.id === childId);
              return child ? renderNode(child, depth + 1) : null;
            })}
          </div>
        )}
      </div>
    );
  };
  
  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <MessageSquare className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Conversation Tree</h3>
          </div>
          <span className="text-xs text-aviation-text-muted">{nodes.length} nodes</span>
        </div>
      </div>
      
      {/* Tree */}
      <div className="flex-1 overflow-y-auto py-2">
        {rootNodes.map(node => renderNode(node))}
        
        {rootNodes.length === 0 && (
          <div className="flex flex-col items-center justify-center py-12 text-aviation-text-muted">
            <MessageSquare className="w-8 h-8 mb-2 opacity-50" />
            <p className="text-sm">No conversation history</p>
          </div>
        )}
      </div>
    </div>
  );
};

// ============================================================================
// Memory Access Monitor
// ============================================================================

interface MemoryAccessEntry {
  id: string;
  memoryId: string;
  memoryLabel: string;
  accessType: 'read' | 'write' | 'delete' | 'share';
  timestamp: number;
  agentId?: string;
  agentName?: string;
  duration?: number;
  success: boolean;
  cacheHit?: boolean;
}

interface MemoryAccessMonitorProps {
  entries: MemoryAccessEntry[];
  selectedEntryId?: string | null;
  timeRange?: { start: number; end: number };
  onEntrySelect?: (entry: MemoryAccessEntry) => void;
  onFilterChange?: (filters: { accessType?: string; agentId?: string }) => void;
  className?: string;
}

export const MemoryAccessMonitor: React.FC<MemoryAccessMonitorProps> = ({
  entries,
  selectedEntryId = null,
  timeRange,
  onEntrySelect,
  onFilterChange,
  className,
}) => {
  const [accessTypeFilter, setAccessTypeFilter] = useState<string | null>(null);
  const [agentFilter, setAgentFilter] = useState<string | null>(null);
  
  const filteredEntries = useMemo(() => {
    return entries.filter(entry => {
      if (accessTypeFilter && entry.accessType !== accessTypeFilter) return false;
      if (agentFilter && entry.agentId !== agentFilter) return false;
      if (timeRange && (entry.timestamp < timeRange.start || entry.timestamp > timeRange.end)) return false;
      return true;
    });
  }, [entries, accessTypeFilter, agentFilter, timeRange]);
  
  const stats = useMemo(() => {
    const total = filteredEntries.length;
    const successful = filteredEntries.filter(e => e.success).length;
    const cacheHits = filteredEntries.filter(e => e.cacheHit).length;
    const avgDuration = filteredEntries.reduce((sum, e) => sum + (e.duration || 0), 0) / (total || 1);
    
    return { total, successful, cacheHits, avgDuration };
  }, [filteredEntries]);
  
  const getAccessTypeIcon = (type: string) => {
    switch (type) {
      case 'read': return <Eye className="w-4 h-4" />;
      case 'write': return <Edit className="w-4 h-4" />;
      case 'delete': return <Trash2 className="w-4 h-4" />;
      case 'share': return <Share2 className="w-4 h-4" />;
      default: return <Activity className="w-4 h-4" />;
    }
  };
  
  const getAccessTypeColor = (type: string) => {
    const colors: Record<string, string> = {
      read: 'text-aviation-cyan',
      write: 'text-green-400',
      delete: 'text-red-400',
      share: 'text-aviation-amber',
    };
    return colors[type] || 'text-gray-400';
  };
  
  const formatTime = (timestamp: number) => {
    return new Date(timestamp).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  };
  
  return (
    <div className={cn('flex flex-col h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      {/* Header */}
      <div className="px-4 py-3 border-b border-aviation-border-panel">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2">
            <Radar className="w-5 h-5 text-aviation-cyan" />
            <h3 className="text-sm font-medium text-aviation-text-primary">Memory Access</h3>
          </div>
          <span className="text-xs text-aviation-text-muted">{filteredEntries.length} access events</span>
        </div>
        
        {/* Stats */}
        <div className="grid grid-cols-4 gap-3">
          <div className="p-2 bg-aviation-bg-secondary rounded">
            <div className="text-lg font-bold text-aviation-text-primary">{stats.total}</div>
            <div className="text-[10px] text-aviation-text-dim">Total</div>
          </div>
          <div className="p-2 bg-aviation-bg-secondary rounded">
            <div className="text-lg font-bold text-green-400">{stats.successful}</div>
            <div className="text-[10px] text-aviation-text-dim">Successful</div>
          </div>
          <div className="p-2 bg-aviation-bg-secondary rounded">
            <div className="text-lg font-bold text-aviation-cyan">{stats.cacheHits}</div>
            <div className="text-[10px] text-aviation-text-dim">Cache Hits</div>
          </div>
          <div className="p-2 bg-aviation-bg-secondary rounded">
            <div className="text-lg font-bold text-aviation-text-primary">{stats.avgDuration.toFixed(0)}ms</div>
            <div className="text-[10px] text-aviation-text-dim">Avg Duration</div>
          </div>
        </div>
        
        {/* Filters */}
        <div className="flex items-center gap-2 mt-3">
          <span className="text-xs text-aviation-text-dim">Filter:</span>
          {['read', 'write', 'delete', 'share'].map((type) => (
            <button
              key={type}
              onClick={() => {
                setAccessTypeFilter(accessTypeFilter === type ? null : type);
                onFilterChange?.({ accessType: accessTypeFilter === type ? undefined : type, agentId: agentFilter || undefined });
              }}
              className={cn(
                'flex items-center gap-1 px-2 py-1 text-xs rounded transition-colors',
                accessTypeFilter === type
                  ? 'bg-aviation-cyan/20 text-aviation-cyan'
                  : 'text-aviation-text-muted hover:bg-aviation-bg-instrument'
              )}
            >
              {getAccessTypeIcon(type)}
              {type.charAt(0).toUpperCase() + type.slice(1)}
            </button>
          ))}
        </div>
      </div>
      
      {/* Entries List */}
      <div className="flex-1 overflow-y-auto">
        {filteredEntries.map((entry) => {
          const isSelected = selectedEntryId === entry.id;
          
          return (
            <div
              key={entry.id}
              onClick={() => onEntrySelect?.(entry)}
              className={cn(
                'px-4 py-3 border-b border-aviation-border-panel cursor-pointer transition-colors',
                isSelected ? 'bg-aviation-bg-instrument' : 'hover:bg-aviation-bg-secondary'
              )}
            >
              <div className="flex items-center justify-between mb-1">
                <div className="flex items-center gap-2">
                  <span className={getAccessTypeColor(entry.accessType)}>
                    {getAccessTypeIcon(entry.accessType)}
                  </span>
                  <span className="text-sm font-medium text-aviation-text-primary">{entry.memoryLabel}</span>
                </div>
                <div className="flex items-center gap-2">
                  {entry.cacheHit ? (
                    <span className="px-1.5 py-0.5 bg-aviation-cyan/20 text-aviation-cyan text-[10px] rounded">Cache</span>
                  ) : null}
                  <span className={cn(
                    'w-2 h-2 rounded-full',
                    entry.success ? 'bg-green-400' : 'bg-red-400'
                  )} />
                </div>
              </div>
              
              <div className="flex items-center gap-3 text-[10px] text-aviation-text-dim">
                <span>{formatTime(entry.timestamp)}</span>
                {entry.agentName && <span>{entry.agentName}</span>}
                {entry.duration !== undefined && <span>{entry.duration}ms</span>}
              </div>
            </div>
          );
        })}
        
        {filteredEntries.length === 0 && (
          <div className="flex flex-col items-center justify-center py-12 text-aviation-text-muted">
            <Radar className="w-8 h-8 mb-2 opacity-50" />
            <p className="text-sm">No access events</p>
          </div>
        )}
      </div>
    </div>
  );
};
