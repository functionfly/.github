/**
 * FRGGraphsPage - Function Runtime Graphs Gallery/Landing Page
 * Comprehensive listing page for all graphs at /frg route
 */

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
import { Badge } from '@/components/ui/badge';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from '@/components/ui/card';
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
import { ScrollArea } from '@/components/ui/scroll-area';
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
      <svg className="absolute inset-0 w-full h-full opacity-30" viewBox="0 0 100 60">
        <defs>
          <pattern id="grid" width="10" height="10" patternUnits="userSpaceOnUse">
            <path d="M 10 0 L 0 0 0 10" fill="none" stroke="currentColor" strokeWidth="0.5" />
          </pattern>
        </defs>
        <rect width="100" height="60" fill="url(#grid)" />
        <circle cx="20" cy="30" r="4" fill="currentColor" className="text-brand-500" />
        <circle cx="50" cy="20" r="4" fill="currentColor" className="text-purple-500" />
        <circle cx="50" cy="40" r="4" fill="currentColor" className="text-purple-500" />
        <circle cx="80" cy="30" r="4" fill="currentColor" className="text-pink-500" />
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
      <Card className="overflow-hidden border-[var(--border-subtle)] bg-[var(--bg-secondary)] hover:border-[var(--border-focus)] transition-all duration-200">
        <div className="relative h-32 bg-[var(--bg-tertiary)]">
          <GraphThumbnail />
          <div className="absolute top-2 right-2">
            <Badge variant="outline" className={cn("text-xs", isPublished ? "bg-green-500/10 text-green-500 border-green-500/20" : "bg-yellow-500/10 text-yellow-500 border-yellow-500/20")}>
              <span className={cn("w-1.5 h-1.5 rounded-full mr-1.5", isPublished ? "bg-green-500" : "bg-yellow-500")} />
              {isPublished ? 'Published' : 'Draft'}
            </Badge>
          </div>
        </div>

        <CardHeader className="p-4 pb-2">
          <div className="flex items-start justify-between">
            <div className="min-w-0 flex-1">
              <CardTitle className="text-sm font-semibold text-[var(--text-primary)] truncate">
                {graph.name}
              </CardTitle>
              <CardDescription className="text-xs text-[var(--text-secondary)] mt-0.5">
                v{graph.version} • {graph.author}
              </CardDescription>
            </div>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon" className="h-8 w-8 -mr-2 opacity-0 group-hover:opacity-100 transition-opacity">
                  <MoreHorizontal className="w-4 h-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-40">
                <DropdownMenuItem onClick={() => onEdit(graph)}>
                  <Edit3 className="w-4 h-4 mr-2" />
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => onRun(graph)}>
                  <Play className="w-4 h-4 mr-2" />
                  Run
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => onDuplicate(graph)}>
                  <Copy className="w-4 h-4 mr-2" />
                  Duplicate
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  onClick={() => onDelete(graph)}
                  className="text-red-500 focus:text-red-500"
                >
                  <Trash2 className="w-4 h-4 mr-2" />
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </CardHeader>

        <CardContent className="p-4 pt-0">
          <div className="flex flex-wrap gap-1 mb-3">
            {graph.visibility && (
              <Badge variant="secondary" className="text-[10px] px-1.5 py-0">
                {graph.visibility}
              </Badge>
            )}
            {graph.executionMode && (
              <Badge variant="secondary" className="text-[10px] px-1.5 py-0">
                {graph.executionMode}
              </Badge>
            )}
          </div>

          <div className="flex items-center justify-between text-xs text-[var(--text-muted)]">
            <div className="flex items-center gap-1">
              <Clock className="w-3 h-3" />
              {new Date(graph.updatedAt).toLocaleDateString()}
            </div>
            <div className="flex items-center gap-1">
              <Activity className="w-3 h-3" />
              {graph.nodeRefs?.length || 0} nodes
            </div>
          </div>

          <div className="flex items-center gap-1 mt-3 pt-3 border-t border-[var(--border-subtle)] opacity-0 group-hover:opacity-100 transition-opacity">
            <Button
              variant="ghost"
              size="sm"
              className="h-7 text-xs"
              onClick={() => onEdit(graph)}
            >
              <Edit3 className="w-3 h-3 mr-1" />
              Edit
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className="h-7 text-xs"
              onClick={() => onRun(graph)}
            >
              <Play className="w-3 h-3 mr-1" />
              Run
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7 ml-auto"
              onClick={() => onDuplicate(graph)}
            >
              <Copy className="w-3 h-3" />
            </Button>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  );
}

function EmptyState({ onCreate }: { onCreate: () => void }) {
  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      className="flex flex-col items-center justify-center py-20 px-4"
    >
      <div className="w-24 h-24 rounded-2xl bg-gradient-to-br from-brand-500/20 to-purple-500/20 flex items-center justify-center mb-6">
        <GitBranch className="w-10 h-10 text-brand-500" />
      </div>
      <h3 className="text-xl font-semibold text-[var(--text-primary)] mb-2">
        No graphs yet
      </h3>
      <p className="text-sm text-[var(--text-secondary)] text-center max-w-sm mb-6">
        Create your first Function Runtime Graph to start building powerful workflows
      </p>
      <Button onClick={onCreate} className="bg-gradient-to-r from-brand-500 to-purple-500">
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
      className="cursor-pointer"
    >
      <Card className="border-[var(--border-subtle)] bg-[var(--bg-secondary)] hover:border-[var(--border-focus)] transition-all duration-200">
        <CardContent className="p-4">
          <div className="flex items-start gap-3">
            <div className={cn("w-10 h-10 rounded-lg bg-gradient-to-br flex items-center justify-center shrink-0", template.color)}>
              <Icon className="w-5 h-5 text-white" />
            </div>
            <div className="flex-1 min-w-0">
              <h4 className="text-sm font-medium text-[var(--text-primary)] truncate">
                {template.name}
              </h4>
              <p className="text-xs text-[var(--text-secondary)] mt-0.5 line-clamp-2">
                {template.description}
              </p>
              <div className="flex items-center gap-1 mt-2 text-xs text-[var(--text-muted)]">
                <Zap className="w-3 h-3" />
                {template.nodeCount} nodes
              </div>
            </div>
            <ChevronRight className="w-4 h-4 text-[var(--text-muted)] shrink-0" />
          </div>
        </CardContent>
      </Card>
    </motion.div>
  );
}

function ErrorState({ message, onRetry }: { message: string; onRetry: () => void }) {
  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      className="flex flex-col items-center justify-center py-20 px-4"
    >
      <div className="w-24 h-24 rounded-2xl bg-gradient-to-br from-red-500/20 to-red-500/10 flex items-center justify-center mb-6">
        <AlertCircle className="w-10 h-10 text-red-500" />
      </div>
      <h3 className="text-xl font-semibold text-[var(--text-primary)] mb-2">
        Failed to load graphs
      </h3>
      <p className="text-sm text-[var(--text-secondary)] text-center max-w-sm mb-6">
        {message}
      </p>
      <Button onClick={onRetry} variant="outline">
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
    <div className="space-y-6">
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4"
      >
        <div>
          <h1 className="text-3xl font-bold text-[var(--text-primary)]">
            Function Runtime Graphs
          </h1>
          <p className="text-sm text-[var(--text-secondary)] mt-1">
            Build, deploy, and monitor powerful function workflows
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" onClick={handleImportGraph}>
            <Upload className="w-4 h-4 mr-2" />
            Import
          </Button>
          <Button
            onClick={handleCreateGraph}
            className="bg-gradient-to-r from-brand-500 to-purple-500"
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
        className="flex flex-col lg:flex-row gap-4"
      >
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
          <Input
            placeholder="Search graphs by name or author..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
        <div className="flex items-center gap-2">
          <Select value={visibilityFilter} onValueChange={setVisibilityFilter}>
            <SelectTrigger className="w-[140px]">
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
            <SelectTrigger className="w-[140px]">
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

      <div className="grid grid-cols-1 xl:grid-cols-4 gap-6">
        <div className="xl:col-span-3">
          <AnimatePresence mode="wait">
            {isLoading ? (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="flex items-center justify-center py-20"
              >
                <Loader2 className="w-8 h-8 animate-spin text-brand-500" />
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
                className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4"
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

        <div className="space-y-6">
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: 0.2 }}
          >
            <Card className="border-[var(--border-subtle)] bg-[var(--bg-secondary)]">
              <CardHeader className="p-4 pb-2">
                <CardTitle className="text-sm font-semibold flex items-center gap-2">
                  <Zap className="w-4 h-4 text-brand-500" />
                  Quick Start Templates
                </CardTitle>
              </CardHeader>
              <CardContent className="p-4 pt-0 space-y-3">
                {quickStartTemplates.map((template) => (
                  <QuickStartTemplate
                    key={template.id}
                    template={template}
                    onSelect={handleTemplateSelect}
                  />
                ))}
              </CardContent>
            </Card>
          </motion.div>
        </div>
      </div>
    </div>
  );
}

export default FRGGraphsPage;