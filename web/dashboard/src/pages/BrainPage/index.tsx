import { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { usePageTitle } from '@/hooks';
import {
  Brain,
  Search,
  Trash2,
  RefreshCw,
  Plus,
  Zap,
  Clock,
  Activity,
  BarChart3,
  Filter,
  ThumbsUp,
  ThumbsDown,
  ExternalLink,
  Loader2,
  AlertTriangle,
  Calendar,
  MessageSquare,
  GitBranch,
  Mail,
  FileText,
  Bell,
  Settings,
  ChevronDown,
  X,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import { Select } from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { EmptyState } from '@/components/ui/empty-state';
import { Progress } from '@/components/ui/progress';
import { MetaTags } from '@/components/seo/MetaTags';
import {
  brainApi,
  type BrainSignal,
  type BrainStats,
  type BrainComposer,
  type BrainTrigger,
} from '@/api/brain';

const connectorIcons: Record<string, React.ReactNode> = {
  github: <GitBranch className="w-3.5 h-3.5" />,
  notion: <FileText className="w-3.5 h-3.5" />,
  slack: <MessageSquare className="w-3.5 h-3.5" />,
  gmail: <Mail className="w-3.5 h-3.5" />,
  linear: <Zap className="w-3.5 h-3.5" />,
  flymind: <Brain className="w-3.5 h-3.5" />,
};

const connectorColors: Record<string, string> = {
  github: 'bg-gray-500/10 text-gray-400 border-gray-500/20',
  notion: 'bg-gray-500/10 text-gray-300 border-gray-500/20',
  slack: 'bg-purple-500/10 text-purple-400 border-purple-500/20',
  gmail: 'bg-red-500/10 text-red-400 border-red-500/20',
  linear: 'bg-blue-500/10 text-blue-400 border-blue-500/20',
  flymind: 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20',
};

const importanceColors: Record<number, string> = {
  1: 'bg-white/5 text-text-secondary border-white/10',
  2: 'bg-yellow-500/10 text-yellow-400 border-yellow-500/20',
  3: 'bg-red-500/10 text-red-400 border-red-500/20',
};

function formatRelativeTime(dateStr?: string): string {
  if (!dateStr) return 'Never';
  const date = new Date(dateStr);
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return 'Just now';
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

function StatsOverview({ stats }: { stats: BrainStats }) {
  const usagePercent = stats.memory_max > 0 ? (stats.memory_used / stats.memory_max) * 100 : 0;

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <Card className="border-white/[0.06] bg-white/[0.02]">
        <CardContent className="p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs text-text-secondary font-medium">Total Signals</span>
            <Activity className="w-4 h-4 text-indigo-400" />
          </div>
          <p className="text-2xl font-bold text-text-primary">
            {stats.total_signals.toLocaleString()}
          </p>
          <p className="text-xs text-text-secondary mt-1">Retention: {stats.retention_days} days</p>
        </CardContent>
      </Card>

      <Card className="border-white/[0.06] bg-white/[0.02]">
        <CardContent className="p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs text-text-secondary font-medium">Memory Used</span>
            <BarChart3 className="w-4 h-4 text-purple-400" />
          </div>
          <p className="text-2xl font-bold text-text-primary">
            {stats.memory_used}
            <span className="text-sm font-normal text-text-secondary">/{stats.memory_max}</span>
          </p>
          <Progress value={usagePercent} className="mt-2 h-1.5" />
        </CardContent>
      </Card>

      <Card className="border-white/[0.06] bg-white/[0.02]">
        <CardContent className="p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs text-text-secondary font-medium">Connector Types</span>
            <GitBranch className="w-4 h-4 text-green-400" />
          </div>
          <p className="text-2xl font-bold text-text-primary">
            {Object.keys(stats.signals_by_connector).length}
          </p>
          <div className="flex flex-wrap gap-1 mt-2">
            {Object.entries(stats.signals_by_connector).map(([slug, count]) => (
              <Badge
                key={slug}
                variant="outline"
                className={`text-[10px] ${connectorColors[slug] || 'bg-white/5 border-white/10'}`}
              >
                {slug}: {count}
              </Badge>
            ))}
          </div>
        </CardContent>
      </Card>

      <Card className="border-white/[0.06] bg-white/[0.02]">
        <CardContent className="p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs text-text-secondary font-medium">Signal Types</span>
            <Zap className="w-4 h-4 text-yellow-400" />
          </div>
          <p className="text-2xl font-bold text-text-primary">
            {Object.keys(stats.signals_by_type).length}
          </p>
          <p className="text-xs text-text-secondary mt-1">
            Newest: {formatRelativeTime(stats.newest_signal)}
          </p>
        </CardContent>
      </Card>
    </div>
  );
}

function SignalRow({
  signal,
  onDelete,
  onFeedback,
}: {
  signal: BrainSignal;
  onDelete: (id: string) => void;
  onFeedback: (id: string, helpful: boolean) => void;
}) {
  const [showActions, setShowActions] = useState(false);

  return (
    <motion.div
      initial={{ opacity: 0, x: -10 }}
      animate={{ opacity: 1, x: 0 }}
      exit={{ opacity: 0, x: 10 }}
      className="group"
    >
      <div
        className="flex items-start gap-3 p-3 rounded-lg hover:bg-white/[0.02] transition-colors cursor-pointer"
        onClick={() => setShowActions(!showActions)}
      >
        <div className="flex-shrink-0 mt-0.5">
          <Badge
            variant="outline"
            className={connectorColors[signal.connector_slug] || 'bg-white/5 border-white/10'}
          >
            <span className="flex items-center gap-1">
              {connectorIcons[signal.connector_slug] || <Zap className="w-3 h-3" />}
              {signal.connector_slug}
            </span>
          </Badge>
        </div>

        <div className="flex-1 min-w-0">
          <p className="text-sm text-text-primary leading-relaxed">{signal.fact}</p>
          <div className="flex items-center gap-2 mt-1.5">
            <span className="text-[10px] text-text-secondary">{signal.signal_type}</span>
            <span className="text-text-secondary/30">·</span>
            <span className="text-[10px] text-text-secondary">
              {formatRelativeTime(signal.created_at)}
            </span>
            <Badge
              variant="outline"
              className={`text-[10px] ${importanceColors[signal.importance] || importanceColors[1]}`}
            >
              P{signal.importance}
            </Badge>
            {signal.source_url && (
              <a
                href={signal.source_url}
                target="_blank"
                rel="noopener noreferrer"
                className="text-[10px] text-indigo-400 hover:text-indigo-300 flex items-center gap-0.5"
                onClick={(e) => e.stopPropagation()}
              >
                <ExternalLink className="w-2.5 h-2.5" />
                source
              </a>
            )}
          </div>
        </div>

        <div className="flex items-center gap-1 flex-shrink-0">
          <Button
            variant="ghost"
            size="sm"
            className="opacity-0 group-hover:opacity-100 h-7 w-7 p-0 text-green-400 hover:text-green-300 hover:bg-green-500/10"
            onClick={(e) => {
              e.stopPropagation();
              onFeedback(signal.id, true);
            }}
          >
            <ThumbsUp className="w-3.5 h-3.5" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className="opacity-0 group-hover:opacity-100 h-7 w-7 p-0 text-red-400 hover:text-red-300 hover:bg-red-500/10"
            onClick={(e) => {
              e.stopPropagation();
              onFeedback(signal.id, false);
            }}
          >
            <ThumbsDown className="w-3.5 h-3.5" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className="opacity-0 group-hover:opacity-100 h-7 w-7 p-0 text-red-400 hover:text-red-300 hover:bg-red-500/10"
            onClick={(e) => {
              e.stopPropagation();
              onDelete(signal.id);
            }}
          >
            <Trash2 className="w-3.5 h-3.5" />
          </Button>
        </div>
      </div>
    </motion.div>
  );
}

function ComposerCard({
  composer,
  onDelete,
}: {
  composer: BrainComposer;
  onDelete: (id: string) => void;
}) {
  return (
    <Card className="border-white/[0.06] bg-white/[0.02] hover:bg-white/[0.04] transition-colors">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Calendar className="w-4 h-4 text-indigo-400" />
            <CardTitle className="text-sm text-text-primary">{composer.name}</CardTitle>
          </div>
          <div className="flex items-center gap-2">
            <Badge
              variant="outline"
              className={
                composer.is_active
                  ? 'bg-green-500/10 text-green-400 border-green-500/20'
                  : 'bg-white/5 text-text-secondary border-white/10'
              }
            >
              {composer.is_active ? 'Active' : 'Paused'}
            </Badge>
            <Button
              variant="ghost"
              size="sm"
              className="h-7 w-7 p-0 text-red-400 hover:text-red-300 hover:bg-red-500/10"
              onClick={() => onDelete(composer.id)}
            >
              <Trash2 className="w-3.5 h-3.5" />
            </Button>
          </div>
        </div>
        <CardDescription className="text-xs">
          Schedule: {composer.schedule} · Format: {composer.output_format}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex flex-wrap gap-1.5">
          {composer.signal_filters?.map((f, i) => (
            <div key={i} className="flex items-center gap-1">
              {f.connector_slugs?.map((slug) => (
                <Badge
                  key={slug}
                  variant="outline"
                  className={`text-[10px] ${connectorColors[slug] || 'bg-white/5 border-white/10'}`}
                >
                  {slug}
                </Badge>
              ))}
            </div>
          ))}
        </div>
        {composer.last_run_at && (
          <p className="text-[10px] text-text-secondary mt-2">
            Last run: {formatRelativeTime(composer.last_run_at)}
          </p>
        )}
      </CardContent>
    </Card>
  );
}

function TriggerCard({
  trigger,
  onDelete,
}: {
  trigger: BrainTrigger;
  onDelete: (id: string) => void;
}) {
  return (
    <Card className="border-white/[0.06] bg-white/[0.02] hover:bg-white/[0.04] transition-colors">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Bell className="w-4 h-4 text-yellow-400" />
            <CardTitle className="text-sm text-text-primary">{trigger.name}</CardTitle>
          </div>
          <div className="flex items-center gap-2">
            <Badge
              variant="outline"
              className={
                trigger.is_active
                  ? 'bg-green-500/10 text-green-400 border-green-500/20'
                  : 'bg-white/5 text-text-secondary border-white/10'
              }
            >
              {trigger.is_active ? 'Active' : 'Paused'}
            </Badge>
            <Button
              variant="ghost"
              size="sm"
              className="h-7 w-7 p-0 text-red-400 hover:text-red-300 hover:bg-red-500/10"
              onClick={() => onDelete(trigger.id)}
            >
              <Trash2 className="w-3.5 h-3.5" />
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          <div className="flex flex-wrap gap-1.5">
            <span className="text-[10px] text-text-secondary">Types:</span>
            {trigger.signal_types.map((t) => (
              <Badge key={t} variant="outline" className="text-[10px] bg-white/5 border-white/10">
                {t}
              </Badge>
            ))}
          </div>
          <div className="flex items-center gap-2 text-[10px] text-text-secondary">
            <span>Min importance: P{trigger.min_importance}</span>
            <span>·</span>
            <span>Action: {trigger.action}</span>
            <span>·</span>
            <span>Schedule: {trigger.schedule}</span>
          </div>
          {trigger.last_fired_at && (
            <p className="text-[10px] text-text-secondary">
              Last fired: {formatRelativeTime(trigger.last_fired_at)}
            </p>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

type TabId = 'signals' | 'composers' | 'triggers';

export function BrainPage() {
  usePageTitle('Brain');
  const [activeTab, setActiveTab] = useState<TabId>('signals');
  const [stats, setStats] = useState<BrainStats | null>(null);
  const [signals, setSignals] = useState<BrainSignal[]>([]);
  const [composers, setComposers] = useState<BrainComposer[]>([]);
  const [triggers, setTriggers] = useState<BrainTrigger[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [connectorFilter, setConnectorFilter] = useState<string>('');
  const [purgeDialogOpen, setPurgeDialogOpen] = useState(false);
  const [purging, setPurging] = useState(false);

  const loadStats = useCallback(async () => {
    try {
      const s = await brainApi.getStats();
      setStats(s);
    } catch (err) {
      console.error('Failed to load brain stats:', err);
    }
  }, []);

  const loadSignals = useCallback(async () => {
    try {
      const params: Record<string, string | number> = { limit: 100 };
      if (connectorFilter) params.connector = connectorFilter;
      const res = await brainApi.listSignals(params);
      setSignals(res.signals || []);
    } catch (err) {
      console.error('Failed to load signals:', err);
    }
  }, [connectorFilter]);

  const loadComposers = useCallback(async () => {
    try {
      const c = await brainApi.listComposers();
      setComposers(c);
    } catch (err) {
      console.error('Failed to load composers:', err);
    }
  }, []);

  const loadTriggers = useCallback(async () => {
    try {
      const t = await brainApi.listTriggers();
      setTriggers(t);
    } catch (err) {
      console.error('Failed to load triggers:', err);
    }
  }, []);

  useEffect(() => {
    const loadAll = async () => {
      setLoading(true);
      await Promise.all([loadStats(), loadSignals(), loadComposers(), loadTriggers()]);
      setLoading(false);
    };
    loadAll();
  }, [loadStats, loadSignals, loadComposers, loadTriggers]);

  const handleDeleteSignal = async (id: string) => {
    try {
      await brainApi.deleteSignal(id);
      setSignals((prev) => prev.filter((s) => s.id !== id));
      loadStats();
    } catch (err) {
      console.error('Failed to delete signal:', err);
    }
  };

  const handleFeedback = async (signalId: string, helpful: boolean) => {
    try {
      await brainApi.submitFeedback({ signal_id: signalId, helpful });
    } catch (err) {
      console.error('Failed to submit feedback:', err);
    }
  };

  const handlePurge = async () => {
    setPurging(true);
    try {
      await brainApi.purgeSignals();
      setSignals([]);
      await loadStats();
      setPurgeDialogOpen(false);
    } catch (err) {
      console.error('Failed to purge signals:', err);
    } finally {
      setPurging(false);
    }
  };

  const handleDeleteComposer = async (id: string) => {
    try {
      await brainApi.deleteComposer(id);
      setComposers((prev) => prev.filter((c) => c.id !== id));
    } catch (err) {
      console.error('Failed to delete composer:', err);
    }
  };

  const handleDeleteTrigger = async (id: string) => {
    try {
      await brainApi.deleteTrigger(id);
      setTriggers((prev) => prev.filter((t) => t.id !== id));
    } catch (err) {
      console.error('Failed to delete trigger:', err);
    }
  };

  const filteredSignals = signals.filter((s) => {
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    return (
      s.fact.toLowerCase().includes(q) ||
      s.signal_type.toLowerCase().includes(q) ||
      s.connector_slug.toLowerCase().includes(q) ||
      s.entity_name?.toLowerCase().includes(q)
    );
  });

  const tabs: { id: TabId; label: string; icon: React.ReactNode; count?: number }[] = [
    {
      id: 'signals',
      label: 'Signals',
      icon: <Activity className="w-4 h-4" />,
      count: signals.length,
    },
    {
      id: 'composers',
      label: 'Composers',
      icon: <Calendar className="w-4 h-4" />,
      count: composers.length,
    },
    {
      id: 'triggers',
      label: 'Triggers',
      icon: <Bell className="w-4 h-4" />,
      count: triggers.length,
    },
  ];

  if (loading) {
    return (
      <div className="space-y-6">
        <div>
          <Skeleton className="h-8 w-48 mb-2" />
          <Skeleton className="h-4 w-96" />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {[1, 2, 3, 4].map((i) => (
            <Skeleton key={i} className="h-24 rounded-lg" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <MetaTags
        title="Brain | FunctionFly"
        description="Your AI Brain learns from connected accounts and gets smarter over time."
      />

      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-text-primary flex items-center gap-3">
            <Brain className="w-8 h-8 text-indigo-400" />
            Brain
          </h1>
          <p className="mt-1 text-text-secondary">
            Your personal AI memory. Signals from connected accounts are scored, stored, and
            injected into agent context.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              loadStats();
              loadSignals();
            }}
            className="border-white/10 hover:bg-white/5"
          >
            <RefreshCw className="w-3.5 h-3.5 mr-1.5" />
            Refresh
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setPurgeDialogOpen(true)}
            className="text-red-400 hover:text-red-300 hover:bg-red-500/10"
          >
            <Trash2 className="w-3.5 h-3.5 mr-1.5" />
            Purge All
          </Button>
        </div>
      </div>

      {/* Stats */}
      {stats && <StatsOverview stats={stats} />}

      {/* Tabs */}
      <div className="flex items-center gap-1 p-1 rounded-lg bg-white/[0.02] border border-white/[0.06] w-fit">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`flex items-center gap-2 px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
              activeTab === tab.id
                ? 'bg-white/[0.08] text-text-primary'
                : 'text-text-secondary hover:text-text-primary hover:bg-white/[0.04]'
            }`}
          >
            {tab.icon}
            {tab.label}
            {tab.count !== undefined && (
              <Badge
                variant="outline"
                className="ml-1 text-[10px] bg-white/5 border-white/10 px-1.5 py-0"
              >
                {tab.count}
              </Badge>
            )}
          </button>
        ))}
      </div>

      {/* Signals Tab */}
      {activeTab === 'signals' && (
        <div className="space-y-4">
          {/* Search and filters */}
          <div className="flex items-center gap-3">
            <div className="relative flex-1 max-w-md">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-secondary" />
              <Input
                placeholder="Search signals..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-9 bg-white/[0.03] border-white/[0.06] text-sm"
              />
              {searchQuery && (
                <button
                  onClick={() => setSearchQuery('')}
                  className="absolute right-3 top-1/2 -translate-y-1/2"
                >
                  <X className="w-3.5 h-3.5 text-text-secondary hover:text-text-primary" />
                </button>
              )}
            </div>
            <select
              value={connectorFilter}
              onChange={(e) => setConnectorFilter(e.target.value)}
              className="px-3 py-2 rounded-md bg-white/[0.03] border border-white/[0.06] text-sm text-text-primary"
            >
              <option value="">All connectors</option>
              {stats &&
                Object.keys(stats.signals_by_connector).map((slug) => (
                  <option key={slug} value={slug}>
                    {slug}
                  </option>
                ))}
            </select>
          </div>

          {/* Signal list */}
          {filteredSignals.length === 0 ? (
            <EmptyState
              variant="card"
              icon={<Activity className="w-8 h-8" />}
              title={searchQuery ? 'No matching signals' : 'No signals yet'}
              description={
                searchQuery
                  ? 'Try adjusting your search or filter.'
                  : "Link a connector to start building your Brain's memory."
              }
            />
          ) : (
            <Card className="border-white/[0.06] bg-white/[0.02]">
              <CardContent className="p-2">
                <AnimatePresence mode="popLayout">
                  {filteredSignals.map((signal) => (
                    <SignalRow
                      key={signal.id}
                      signal={signal}
                      onDelete={handleDeleteSignal}
                      onFeedback={handleFeedback}
                    />
                  ))}
                </AnimatePresence>
              </CardContent>
            </Card>
          )}
        </div>
      )}

      {/* Composers Tab */}
      {activeTab === 'composers' && (
        <div className="space-y-4">
          {composers.length === 0 ? (
            <EmptyState
              variant="card"
              icon={<Calendar className="w-8 h-8" />}
              title="No composers configured"
              description="Create a Brain Composer to get automated daily briefings from your connected accounts."
            />
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <AnimatePresence mode="popLayout">
                {composers.map((composer) => (
                  <ComposerCard
                    key={composer.id}
                    composer={composer}
                    onDelete={handleDeleteComposer}
                  />
                ))}
              </AnimatePresence>
            </div>
          )}
        </div>
      )}

      {/* Triggers Tab */}
      {activeTab === 'triggers' && (
        <div className="space-y-4">
          {triggers.length === 0 ? (
            <EmptyState
              variant="card"
              icon={<Bell className="w-8 h-8" />}
              title="No triggers configured"
              description="Create a Brain Trigger to run agents automatically when specific signal patterns are detected."
            />
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <AnimatePresence mode="popLayout">
                {triggers.map((trigger) => (
                  <TriggerCard key={trigger.id} trigger={trigger} onDelete={handleDeleteTrigger} />
                ))}
              </AnimatePresence>
            </div>
          )}
        </div>
      )}

      {/* Purge confirmation dialog */}
      <Dialog open={purgeDialogOpen} onOpenChange={setPurgeDialogOpen}>
        <DialogContent className="bg-slate-900 border-white/10">
          <DialogHeader>
            <DialogTitle className="text-text-primary flex items-center gap-2">
              <AlertTriangle className="w-5 h-5 text-red-400" />
              Purge All Brain Signals
            </DialogTitle>
          </DialogHeader>
          <div className="py-4">
            <p className="text-sm text-text-secondary">
              This will permanently delete all {stats?.total_signals || 0} signals from your Brain.
              This action cannot be undone. Your connectors will continue to generate new signals.
            </p>
          </div>
          <DialogFooter>
            <Button
              variant="ghost"
              onClick={() => setPurgeDialogOpen(false)}
              className="border-white/10"
            >
              Cancel
            </Button>
            <Button variant="destructive" onClick={handlePurge} disabled={purging}>
              {purging ? (
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
              ) : (
                <Trash2 className="w-4 h-4 mr-2" />
              )}
              Purge All Signals
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
