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
  Archive,
} from 'lucide-react';

import { DashboardLayout } from '@/components/layout/DashboardLayout';
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
  Tabs,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs';
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
import { Separator } from '@/components/ui/separator';
import { ScrollArea } from '@/components/ui/scroll-area';
import { cn } from '@/lib/utils';

// Types
interface Graph {
  id: string;
  name: string;
  version: string;
  status: 'active' | 'draft' | 'archived';
  lastModified: string;
  executionCount: number;
  thumbnail?: string;
  tags: string[];
  author: string;
}

interface ExecutionActivity {
  id: string;
  graphName: string;
  status: 'success' | 'error' | 'running';
  timestamp: string;
  duration: number;
}

// Mock data for demonstration
const mockGraphs: Graph[] = [
  {
    id: '1',
    name: 'AI Pipeline Processor',
    version: '2.1.0',
    status: 'active',
    lastModified: '2026-04-08T14:30:00Z',
    executionCount: 12543,
    tags: ['ai', 'ml', 'production'],
    author: 'functionfly',
  },
  {
    id: '2',
    name: 'Webhook Handler Chain',
    version: '1.5.2',
    status: 'active',
    lastModified: '2026-04-09T09:15:00Z',
    executionCount: 8921,
    tags: ['webhook', 'api', 'production'],
    author: 'acme-corp',
  },
  {
    id: '3',
    name: 'Data Transformation Flow',
    version: '0.9.0',
    status: 'draft',
    lastModified: '2026-04-10T16:45:00Z',
    executionCount: 0,
    tags: ['etl', 'data', 'wip'],
    author: 'data-team',
  },
  {
    id: '4',
    name: 'Scheduled Report Generator',
    version: '3.0.1',
    status: 'archived',
    lastModified: '2026-03-15T10:00:00Z',
    executionCount: 45678,
    tags: ['scheduled', 'reports', 'deprecated'],
    author: 'analytics',
  },
  {
    id: '5',
    name: 'Real-time Stream Processor',
    version: '1.2.0',
    status: 'active',
    lastModified: '2026-04-07T11:20:00Z',
    executionCount: 234567,
    tags: ['streaming', 'realtime', 'kafka'],
    author: 'streaming-team',
  },
  {
    id: '6',
    name: 'ML Model Inference Pipeline',
    version: '2.0.0',
    status: 'draft',
    lastModified: '2026-04-10T08:00:00Z',
    executionCount: 123,
    tags: ['ml', 'inference', 'testing'],
    author: 'ml-team',
  },
];

const recentActivities: ExecutionActivity[] = [
  {
    id: '1',
    graphName: 'AI Pipeline Processor',
    status: 'success',
    timestamp: '2026-04-10T19:25:00Z',
    duration: 234,
  },
  {
    id: '2',
    graphName: 'Webhook Handler Chain',
    status: 'running',
    timestamp: '2026-04-10T19:24:00Z',
    duration: 0,
  },
  {
    id: '3',
    graphName: 'Real-time Stream Processor',
    status: 'success',
    timestamp: '2026-04-10T19:22:00Z',
    duration: 12,
  },
  {
    id: '4',
    graphName: 'AI Pipeline Processor',
    status: 'error',
    timestamp: '2026-04-10T19:20:00Z',
    duration: 45,
  },
  {
    id: '5',
    graphName: 'Webhook Handler Chain',
    status: 'success',
    timestamp: '2026-04-10T19:18:00Z',
    duration: 89,
  },
];

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

const statusConfig = {
  active: {
    label: 'Active',
    color: 'bg-green-500/10 text-green-500 border-green-500/20',
    dotColor: 'bg-green-500',
  },
  draft: {
    label: 'Draft',
    color: 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20',
    dotColor: 'bg-yellow-500',
  },
  archived: {
    label: 'Archived',
    color: 'bg-gray-500/10 text-gray-400 border-gray-500/20',
    dotColor: 'bg-gray-400',
  },
};

function GraphThumbnail({ gradient = true }: { gradient?: boolean }) {
  return (
    <div className="relative w-full h-full overflow-hidden rounded-lg">
      {gradient ? (
        <div className="absolute inset-0 bg-gradient-to-br from-brand-500/20 via-purple-500/20 to-pink-500/20" />
      ) : null}
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
  graph: Graph;
  onEdit: (id: string) => void;
  onRun: (id: string) => void;
  onDuplicate: (id: string) => void;
  onDelete: (id: string) => void;
}) {
  const status = statusConfig[graph.status];

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
        {/* Thumbnail */}
        <div className="relative h-32 bg-[var(--bg-tertiary)]">
          <GraphThumbnail />
          <div className="absolute top-2 right-2">
            <Badge variant="outline" className={cn("text-xs", status.color)}>
              <span className={cn("w-1.5 h-1.5 rounded-full mr-1.5", status.dotColor)} />
              {status.label}
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
                <DropdownMenuItem onClick={() => onEdit(graph.id)}>
                  <Edit3 className="w-4 h-4 mr-2" />
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => onRun(graph.id)}>
                  <Play className="w-4 h-4 mr-2" />
                  Run
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => onDuplicate(graph.id)}>
                  <Copy className="w-4 h-4 mr-2" />
                  Duplicate
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  onClick={() => onDelete(graph.id)}
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
          {/* Tags */}
          <div className="flex flex-wrap gap-1 mb-3">
            {graph.tags.slice(0, 3).map((tag) => (
              <Badge key={tag} variant="secondary" className="text-[10px] px-1.5 py-0">
                {tag}
              </Badge>
            ))}
          </div>

          {/* Meta info */}
          <div className="flex items-center justify-between text-xs text-[var(--text-muted)]">
            <div className="flex items-center gap-1">
              <Clock className="w-3 h-3" />
              {new Date(graph.lastModified).toLocaleDateString()}
            </div>
            <div className="flex items-center gap-1">
              <Activity className="w-3 h-3" />
              {graph.executionCount.toLocaleString()}
            </div>
          </div>

          {/* Quick actions bar */}
          <div className="flex items-center gap-1 mt-3 pt-3 border-t border-[var(--border-subtle)] opacity-0 group-hover:opacity-100 transition-opacity">
            <Button
              variant="ghost"
              size="sm"
              className="h-7 text-xs"
              onClick={() => onEdit(graph.id)}
            >
              <Edit3 className="w-3 h-3 mr-1" />
              Edit
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className="h-7 text-xs"
              onClick={() => onRun(graph.id)}
            >
              <Play className="w-3 h-3 mr-1" />
              Run
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7 ml-auto"
              onClick={() => onDuplicate(graph.id)}
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

function RecentActivityItem({ activity }: { activity: ExecutionActivity }) {
  const statusIcon =
    activity.status === 'success' ? (
      <CheckCircle2 className="w-4 h-4 text-green-500" />
    ) : activity.status === 'error' ? (
      <AlertCircle className="w-4 h-4 text-red-500" />
    ) : (
      <div className="w-4 h-4 rounded-full border-2 border-blue-500 border-t-transparent animate-spin" />
    );

  return (
    <div className="flex items-center gap-3 py-2 px-3 rounded-lg hover:bg-[var(--bg-tertiary)] transition-colors">
      {statusIcon}
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-[var(--text-primary)] truncate">
          {activity.graphName}
        </p>
        <p className="text-xs text-[var(--text-secondary)]">
          {new Date(activity.timestamp).toLocaleTimeString()} •{' '}
          {activity.duration > 0 ? `${activity.duration}ms` : 'Running...'}
        </p>
      </div>
    </div>
  );
}

export function FRGGraphsPage() {
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [dateFilter, setDateFilter] = useState<string>('all');
  const [selectedTag, setSelectedTag] = useState<string | null>(null);

  const filteredGraphs = useMemo(() => {
    return mockGraphs.filter((graph) => {
      const matchesSearch =
        graph.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        graph.tags.some((tag) => tag.toLowerCase().includes(searchQuery.toLowerCase()));
      const matchesStatus = statusFilter === 'all' || graph.status === statusFilter;
      const matchesTag = !selectedTag || graph.tags.includes(selectedTag);
      return matchesSearch && matchesStatus && matchesTag;
    });
  }, [searchQuery, statusFilter, selectedTag]);

  const allTags = useMemo(() => {
    const tags = new Set<string>();
    mockGraphs.forEach((graph) => graph.tags.forEach((tag) => tags.add(tag)));
    return Array.from(tags);
  }, []);

  const handleCreateGraph = () => {
    navigate('/frg/new');
  };

  const handleImportGraph = () => {
    // TODO: Implement import
    console.log('Import graph');
  };

  const handleEdit = (id: string) => {
    navigate(`/frg/${id}`);
  };

  const handleRun = (id: string) => {
    console.log('Run graph', id);
  };

  const handleDuplicate = (id: string) => {
    console.log('Duplicate graph', id);
  };

  const handleDelete = (id: string) => {
    console.log('Delete graph', id);
  };

  const handleTemplateSelect = (templateId: string) => {
    navigate(`/frg/new?template=${templateId}`);
  };

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Header */}
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

        {/* Search and Filters */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="flex flex-col lg:flex-row gap-4"
        >
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
            <Input
              placeholder="Search graphs by name or tag..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>
          <div className="flex items-center gap-2">
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-[140px]">
                <Filter className="w-4 h-4 mr-2" />
                <SelectValue placeholder="Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Status</SelectItem>
                <SelectItem value="active">Active</SelectItem>
                <SelectItem value="draft">Draft</SelectItem>
                <SelectItem value="archived">Archived</SelectItem>
              </SelectContent>
            </Select>
            <Select value={dateFilter} onValueChange={setDateFilter}>
              <SelectTrigger className="w-[140px]">
                <Calendar className="w-4 h-4 mr-2" />
                <SelectValue placeholder="Date" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Time</SelectItem>
                <SelectItem value="today">Today</SelectItem>
                <SelectItem value="week">This Week</SelectItem>
                <SelectItem value="month">This Month</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </motion.div>

        {/* Tags filter */}
        {allTags.length > 0 && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.15 }}
            className="flex items-center gap-2 flex-wrap"
          >
            <Tag className="w-4 h-4 text-[var(--text-muted)]" />
            <Button
              variant={selectedTag === null ? 'secondary' : 'ghost'}
              size="sm"
              className="h-7 text-xs"
              onClick={() => setSelectedTag(null)}
            >
              All
            </Button>
            {allTags.map((tag) => (
              <Button
                key={tag}
                variant={selectedTag === tag ? 'secondary' : 'ghost'}
                size="sm"
                className="h-7 text-xs"
                onClick={() => setSelectedTag(tag === selectedTag ? null : tag)}
              >
                {tag}
              </Button>
            ))}
          </motion.div>
        )}

        {/* Main Content */}
        <div className="grid grid-cols-1 xl:grid-cols-4 gap-6">
          {/* Graphs Grid */}
          <div className="xl:col-span-3">
            <AnimatePresence mode="wait">
              {filteredGraphs.length === 0 ? (
                <EmptyState onCreate={handleCreateGraph} />
              ) : (
                <motion.div
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4"
                >
                  {filteredGraphs.map((graph) => (
                    <GraphCard
                      key={graph.id}
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

          {/* Sidebar */}
          <div className="space-y-6">
            {/* Quick Start Templates */}
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

            {/* Recent Activity */}
            <motion.div
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: 0.3 }}
            >
              <Card className="border-[var(--border-subtle)] bg-[var(--bg-secondary)]">
                <CardHeader className="p-4 pb-2">
                  <CardTitle className="text-sm font-semibold flex items-center gap-2">
                    <Activity className="w-4 h-4 text-brand-500" />
                    Recent Activity
                  </CardTitle>
                </CardHeader>
                <CardContent className="p-4 pt-0">
                  <ScrollArea className="h-48">
                    <div className="space-y-1">
                      {recentActivities.map((activity) => (
                        <RecentActivityItem key={activity.id} activity={activity} />
                      ))}
                    </div>
                  </ScrollArea>
                </CardContent>
              </Card>
            </motion.div>
          </div>
        </div>
      </div>
    </DashboardLayout>
  );
}

export default FRGGraphsPage;
