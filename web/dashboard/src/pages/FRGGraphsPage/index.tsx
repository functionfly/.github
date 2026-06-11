/**
 * FRGGraphsPage - Function Runtime Graphs Gallery/Landing Page
 * Comprehensive listing page for all graphs at /frg route
 */

import './styles.css';

import { useState, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Plus,
  Upload,
  Search,
  Filter,
  MoreHorizontal,
  Play,
  Copy,
  Trash2,
  Edit3,
  Clock,
  Activity,
  Zap,
  GitBranch,
  Webhook,
  Database,
  FileJson,
  Sparkles,
  ChevronRight,
  Calendar,
  Tag,
  AlertCircle,
  CheckCircle2,
  Loader2,
  X,
} from 'lucide-react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { cn } from '@/lib/utils';
import { frgApi } from '@/api/frg';
import type { GraphDefinition } from '@/types/frg';

const quickStartTemplates = [
  {
    id: 'ai-pipeline',
    name: 'AI Pipeline',
    description: 'Input → GPT-4 → Output processing chain',
    icon: Sparkles,
    color: 'from-purple-500 to-pink-500',
    nodeCount: 3,
  },
  {
    id: 'webhook-handler',
    name: 'Webhook Handler',
    description: 'API endpoint with validation and response',
    icon: Webhook,
    color: 'from-blue-500 to-cyan-500',
    nodeCount: 4,
  },
  {
    id: 'data-processing',
    name: 'Data Processing',
    description: 'Upload → Transform → Store workflow',
    icon: Database,
    color: 'from-green-500 to-emerald-500',
    nodeCount: 5,
  },
  {
    id: 'etl-workflow',
    name: 'ETL Workflow',
    description: 'Extract, Transform, Load pipeline',
    icon: FileJson,
    color: 'from-orange-500 to-amber-500',
    nodeCount: 6,
  },
];

function GraphThumbnail({ gradient = true }: { gradient?: boolean }) {
  return (
    <div className="relative w-full h-full overflow-hidden rounded-lg">
      {gradient && (
        <div className="absolute inset-0 bg-gradient-to-br from-brand-500/20 via-purple-500/20 to-pink-500/20" />
      )}
      <svg className="absolute inset-0 w-full h-full opacity-30 frg-graph-thumbnail-grid" viewBox="0 0 100 60">
        <defs>
          <pattern id="frg-grid" width="10" height="10" patternUnits="userSpaceOnUse">
            <path d="M 10 0 L 0 0 0 10" fill="none" stroke="currentColor" strokeWidth="0.5" />
          </pattern>
        </defs>
        <rect width="100" height="60" fill="url(#frg-grid)" />
        <circle cx="20" cy="30" r="4" fill="currentColor" className="text-brand-500 frg-graph-thumbnail-node" />
        <circle cx="50" cy="20" r="4" fill="currentColor" className="text-purple-500 frg-graph-thumbnail-node" />
        <circle cx="50" cy="40" r="4" fill="currentColor" className="text-purple-500 frg-graph-thumbnail-node" />
        <circle cx="80" cy="30" r="4" fill="currentColor" className="text-pink-500 frg-graph-thumbnail-node" />
        <line x1="24" y1="30" x2="46" y2="20" stroke="currentColor" strokeWidth="1" className="text-brand-500/50" />
        <line x1="24" y1="30" x2="46" y2="40" stroke="currentColor" strokeWidth="1" className="text-brand-500/50" />
        <line x1="54" y1="20" x2="76" y2="30" stroke="currentColor" strokeWidth="1" className="text-purple-500/50" />
        <line x1="54" y1="40" x2="76" y2="30" stroke="currentColor" strokeWidth="1" className="text-purple-500/50" />
      </svg>
    </div>
  );
}

function GraphCard({
  graph,
  onEdit,
  onRun,
  onDuplicate,
  onDelete,
}: {
  graph: GraphDefinition;
  onEdit: (graph: GraphDefinition) => void;
  onRun: (graph: GraphDefinition) => void;
  onDuplicate: (graph: GraphDefinition) => void;
  onDelete: (graph: GraphDefinition) => void;
}) {
  const isPublished = !!graph.publishedAt;

  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, scale: 0.95 }}
      whileHover={{ y: -4 }}
      className="group"
    >
      <div className="frg-graph-card">
        <div className="frg-graph-thumbnail">
          <GraphThumbnail />
          <div className="absolute top-2 right-2">
            <span className={cn("frg-graph-badge", isPublished ? "frg-graph-badge-published" : "frg-graph-badge-draft")}>
              <span className="frg-graph-badge-dot" />
              {isPublished ? 'Published' : 'Draft'}
            </span>
          </div>
        </div>

        <div className="frg-graph-content">
          <div className="frg-graph-header">
            <div className="min-w-0 flex-1">
              <h3 className="frg-graph-title">
                {graph.name}
              </h3>
              <p className="frg-graph-meta">
                v{graph.version} • {graph.author}
              </p>
            </div>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon" className="frg-more-button h-8 w-8 -mr-2">
                  <MoreHorizontal className="w-4 h-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="frg-dropdown-content">
                <DropdownMenuItem onClick={() => onEdit(graph)} className="frg-dropdown-item">
                  <Edit3 className="frg-dropdown-icon mr-2" />
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => onRun(graph)} className="frg-dropdown-item">
                  <Play className="frg-dropdown-icon mr-2" />
                  Run
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => onDuplicate(graph)} className="frg-dropdown-item">
                  <Copy className="frg-dropdown-icon mr-2" />
                  Duplicate
                </DropdownMenuItem>
                <DropdownMenuSeparator className="frg-dropdown-separator" />
                <DropdownMenuItem
                  onClick={() => onDelete(graph)}
                  className="frg-dropdown-item frg-dropdown-item-danger"
                >
                  <Trash2 className="frg-dropdown-icon mr-2" />
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>

          <div className="frg-graph-tags">
            {graph.visibility && (
              <span className="frg-graph-tag">
                {graph.visibility}
              </span>
            )}
            {graph.executionMode && (
              <span className="frg-graph-tag">
                {graph.executionMode}
              </span>
            )}
          </div>

          <div className="frg-graph-stats">
            <div className="frg-graph-stat">
              <Clock className="frg-graph-stat-icon" />
              {new Date(graph.updatedAt).toLocaleDateString()}
            </div>
            <div className="frg-graph-stat">
              <Activity className="frg-graph-stat-icon" />
              {graph.nodeRefs?.length || 0} nodes
            </div>
          </div>

          <div className="frg-graph-actions">
            <button
              className="frg-graph-action-button"
              onClick={() => onEdit(graph)}
            >
              <Edit3 className="frg-graph-action-icon" />
              Edit
            </button>
            <button
              className="frg-graph-action-button"
              onClick={() => onRun(graph)}
            >
              <Play className="frg-graph-action-icon" />
              Run
            </button>
            <button
              className="frg-graph-action-button"
              onClick={() => onDuplicate(graph)}
            >
              <Copy className="frg-graph-action-icon" />
            </button>
          </div>
        </div>
      </div>
    </motion.div>
  );
}

function EmptyState({ onCreate }: { onCreate: () => void }) {
  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      className="frg-empty-state"
    >
      <div className="frg-empty-icon-wrapper">
        <GitBranch className="frg-empty-icon" />
      </div>
      <h3 className="frg-empty-title">
        No graphs yet
      </h3>
      <p className="frg-empty-description">
        Create your first Function Runtime Graph to start building powerful workflows
      </p>
      <Button onClick={onCreate} className="frg-create-button">
        <Plus className="w-4 h-4 mr-2" />
        Create Your First Graph
      </Button>
    </motion.div>
  );
}

function QuickStartTemplate({
  template,
  onSelect,
}: {
  template: typeof quickStartTemplates[0];
  onSelect: (id: string) => void;
}) {
  const Icon = template.icon;

  return (
    <motion.div
      whileHover={{ scale: 1.02 }}
      whileTap={{ scale: 0.98 }}
      onClick={() => onSelect(template.id)}
      className="frg-template-card"
    >
      <div className="frg-template-content">
        <div className={cn("frg-template-icon-wrapper bg-gradient-to-br", template.color)}>
          <Icon className="frg-template-icon" />
        </div>
        <div className="frg-template-info">
          <h4 className="frg-template-name">
            {template.name}
          </h4>
          <p className="frg-template-description">
            {template.description}
          </p>
          <div className="frg-template-nodes">
            <Zap className="frg-template-nodes-icon" />
            {template.nodeCount} nodes
          </div>
        </div>
        <ChevronRight className="frg-template-arrow" />
      </div>
    </motion.div>
  );
}

function ErrorState({ message, onRetry }: { message: string; onRetry: () => void }) {
  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      className="frg-error-state"
    >
      <div className="frg-error-icon-wrapper">
        <AlertCircle className="frg-error-icon" />
      </div>
      <h3 className="frg-error-title">
        Failed to load graphs
      </h3>
      <p className="frg-error-description">
        {message}
      </p>
      <Button onClick={onRetry} variant="outline" className="frg-retry-button">
        Try Again
      </Button>
    </motion.div>
  );
}

export function FRGGraphsPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [visibilityFilter, setVisibilityFilter] = useState<string>('all');
  const [executionModeFilter, setExecutionModeFilter] = useState<string>('all');

  // Fetch graphs from API
  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ['frg', 'graphs', visibilityFilter, executionModeFilter],
    queryFn: () => frgApi.listGraphs({
      visibility: visibilityFilter !== 'all' ? visibilityFilter : undefined,
      executionMode: executionModeFilter !== 'all' ? executionModeFilter : undefined,
    }),
  });

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: ({ author, name }: { author: string; name: string }) =>
      frgApi.deleteGraph(author, name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['frg', 'graphs'] });
    },
  });

  // Remix mutation
  const remixMutation = useMutation({
    mutationFn: ({ author, name, newName }: { author: string; name: string; newName: string }) =>
      frgApi.remixGraph(author, name, newName),
    onSuccess: (newGraph) => {
      queryClient.invalidateQueries({ queryKey: ['frg', 'graphs'] });
      navigate(`/frg/${newGraph.author}/${newGraph.name}`);
    },
  });

  const graphs = data?.graphs || [];

  const filteredGraphs = useMemo(() => {
    return graphs.filter((graph) => {
      const matchesSearch =
        graph.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        graph.author.toLowerCase().includes(searchQuery.toLowerCase());
      return matchesSearch;
    });
  }, [graphs, searchQuery]);

  const handleCreateGraph = () => {
    navigate('/frg/new');
  };

  const handleImportGraph = () => {
    console.log('Import graph');
  };

  const handleEdit = (graph: GraphDefinition) => {
    navigate(`/frg/${graph.author}/${graph.name}`);
  };

  const handleRun = async (graph: GraphDefinition) => {
    try {
      const result = await frgApi.executeGraph(graph.author, graph.name);
      if (result.instanceId) {
        navigate(`/frg/${graph.author}/${graph.name}?instance=${result.instanceId}`);
      }
    } catch (err) {
      console.error('Failed to run graph:', err);
    }
  };

  const handleDuplicate = (graph: GraphDefinition) => {
    remixMutation.mutate({
      author: graph.author,
      name: graph.name,
      newName: `${graph.name}-copy`,
    });
  };

  const handleDelete = (graph: GraphDefinition) => {
    if (confirm(`Delete "${graph.name}"? This cannot be undone.`)) {
      deleteMutation.mutate({ author: graph.author, name: graph.name });
    }
  };

  const handleTemplateSelect = (templateId: string) => {
    navigate(`/frg/new?template=${templateId}`);
  };

  return (
    <div className="frg-page">
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        className="frg-header"
      >
        <div>
          <h1 className="frg-title">
            Function Runtime Graphs
          </h1>
          <p className="frg-subtitle">
            Build, deploy, and monitor powerful function workflows
          </p>
        </div>
        <div className="frg-toolbar-actions">
          <Button variant="outline" onClick={handleImportGraph} className="frg-import-button">
            <Upload className="w-4 h-4 mr-2" />
            Import
          </Button>
          <Button
            onClick={handleCreateGraph}
            className="frg-create-button"
          >
            <Plus className="w-4 h-4 mr-2" />
            Create Graph
          </Button>
        </div>
      </motion.div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.1 }}
        className="frg-toolbar"
      >
        <div className="frg-search-container">
          <Search className="frg-search-icon" />
          <Input
            placeholder="Search graphs by name or author..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="frg-search-input"
          />
        </div>
        <div className="flex items-center gap-2">
          <Select value={visibilityFilter} onValueChange={setVisibilityFilter}>
            <SelectTrigger className="frg-filter-select">
              <Filter className="w-4 h-4 mr-2" />
              <SelectValue placeholder="Visibility" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Visibility</SelectItem>
              <SelectItem value="public">Public</SelectItem>
              <SelectItem value="private">Private</SelectItem>
              <SelectItem value="team">Team</SelectItem>
            </SelectContent>
          </Select>
          <Select value={executionModeFilter} onValueChange={setExecutionModeFilter}>
            <SelectTrigger className="frg-filter-select">
              <SelectValue placeholder="Mode" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Modes</SelectItem>
              <SelectItem value="sync">Sync</SelectItem>
              <SelectItem value="async">Async</SelectItem>
              <SelectItem value="streaming">Streaming</SelectItem>
              <SelectItem value="event_driven">Event Driven</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </motion.div>

      <div className="frg-content-grid">
        <div>
          <AnimatePresence mode="wait">
            {isLoading ? (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="frg-loading-state"
              >
                <Loader2 className="frg-loading-spinner" />
              </motion.div>
            ) : isError ? (
              <ErrorState
                message={error instanceof Error ? error.message : 'Unknown error'}
                onRetry={() => refetch()}
              />
            ) : filteredGraphs.length === 0 ? (
              <EmptyState onCreate={handleCreateGraph} />
            ) : (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="frg-graphs-grid"
              >
                {filteredGraphs.map((graph) => (
                  <GraphCard
                    key={`${graph.author}/${graph.name}@${graph.version}`}
                    graph={graph}
                    onEdit={handleEdit}
                    onRun={handleRun}
                    onDuplicate={handleDuplicate}
                    onDelete={handleDelete}
                  />
                ))}
              </motion.div>
            )}
          </AnimatePresence>
        </div>

        <div className="frg-templates-sidebar">
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: 0.2 }}
          >
            <div className="frg-templates-card">
              <div className="frg-templates-header">
                <Zap className="frg-templates-icon" />
                <span className="frg-templates-title">Quick Start Templates</span>
              </div>
              <div className="frg-templates-list">
                {quickStartTemplates.map((template) => (
                  <QuickStartTemplate
                    key={template.id}
                    template={template}
                    onSelect={handleTemplateSelect}
                  />
                ))}
              </div>
            </div>
          </motion.div>
        </div>
      </div>
    </div>
  );
}

export default FRGGraphsPage;