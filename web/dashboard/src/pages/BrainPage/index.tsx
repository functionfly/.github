import { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { usePageTitle } from '@/hooks';
import {
  Brain,
  Search,
  Trash2,
  RefreshCw,
  Zap,
  Clock,
  Activity,
  BarChart3,
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
  X,
} from 'lucide-react';
import {
  PageGrid,
  Chamber,
  CornerBrace,
  TrustSeal,
  SealedButton,
  FrameButton,
  StatusPill,
  GaugeStrip,
  Gauge,
  AnnotationTag,
  Card,
  Modal,
} from '@/components/containment';
import {
  brainApi,
  type BrainSignal,
  type BrainStats,
  type BrainComposer,
  type BrainTrigger,
} from '@/api/brain';

import './brain.css';

const connectorIcons: Record<string, React.ReactNode> = {
  github: <GitBranch className="brain-icon-xs" />,
  notion: <FileText className="brain-icon-xs" />,
  slack: <MessageSquare className="brain-icon-xs" />,
  gmail: <Mail className="brain-icon-xs" />,
  linear: <Zap className="brain-icon-xs" />,
  flymind: <Brain className="brain-icon-xs" />,
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

function connectorStatus(slug: string): 'live' | 'pending' | 'revoked' {
  if (['github', 'linear', 'flymind'].includes(slug)) return 'live';
  if (['slack', 'notion'].includes(slug)) return 'pending';
  return 'revoked';
}

function importanceStatus(importance: number): 'live' | 'pending' | 'revoked' {
  if (importance >= 3) return 'revoked';
  if (importance >= 2) return 'pending';
  return 'live';
}

function StatsOverview({ stats }: { stats: BrainStats }) {
  return (
    <GaugeStrip>
      <Gauge isFirst data={{ value: stats.total_signals.toLocaleString(), label: 'Total Signals' }} />
      <Gauge data={{ value: `${stats.memory_used}/${stats.memory_max}`, label: 'Memory Used' }} />
      <Gauge data={{ value: Object.keys(stats.signals_by_connector).length, label: 'Connector Types' }} />
      <Gauge data={{ value: Object.keys(stats.signals_by_type).length, label: 'Signal Types' }} />
    </GaugeStrip>
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
      className="brain-signal-row"
    >
      <div
        className="brain-signal-row__inner"
        onClick={() => setShowActions(!showActions)}
      >
        <div className="brain-signal-row__connector">
          <StatusPill status={connectorStatus(signal.connector_slug)} label={signal.connector_slug} />
        </div>

        <div className="brain-signal-row__content">
          <p className="brain-signal-row__fact">{signal.fact}</p>
          <div className="brain-signal-row__meta">
            <span className="brain-signal-row__type">{signal.signal_type}</span>
            <span className="brain-signal-row__dot">·</span>
            <span className="brain-signal-row__time">{formatRelativeTime(signal.created_at)}</span>
            <StatusPill status={importanceStatus(signal.importance)} label={`P${signal.importance}`} />
            {signal.source_url && (
              <a
                href={signal.source_url}
                target="_blank"
                rel="noopener noreferrer"
                className="brain-signal-row__source"
                onClick={(e) => e.stopPropagation()}
              >
                <ExternalLink className="brain-icon-xxs" />
                source
              </a>
            )}
          </div>
        </div>

        <div className="brain-signal-row__actions">
          <button
            className="brain-icon-btn brain-icon-btn--ok"
            onClick={(e) => { e.stopPropagation(); onFeedback(signal.id, true); }}
          >
            <ThumbsUp className="brain-icon-sm" />
          </button>
          <button
            className="brain-icon-btn brain-icon-btn--bad"
            onClick={(e) => { e.stopPropagation(); onFeedback(signal.id, false); }}
          >
            <ThumbsDown className="brain-icon-sm" />
          </button>
          <button
            className="brain-icon-btn brain-icon-btn--bad"
            onClick={(e) => { e.stopPropagation(); onDelete(signal.id); }}
          >
            <Trash2 className="brain-icon-sm" />
          </button>
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
    <Card className="brain-entity-card">
      <div className="brain-entity-card__header">
        <div className="brain-entity-card__title-row">
          <Calendar className="brain-icon-sm brain-icon-accent" />
          <h3 className="brain-entity-card__title">{composer.name}</h3>
        </div>
        <div className="brain-entity-card__header-actions">
          <StatusPill status={composer.is_active ? 'live' : 'pending'} label={composer.is_active ? 'Active' : 'Paused'} />
          <button className="brain-icon-btn brain-icon-btn--bad" onClick={() => onDelete(composer.id)}>
            <Trash2 className="brain-icon-sm" />
          </button>
        </div>
      </div>
      <p className="brain-entity-card__desc">
        Schedule: {composer.schedule} · Format: {composer.output_format}
      </p>
      <div className="brain-entity-card__tags">
        {composer.signal_filters?.map((f, i) => (
          <div key={i} className="brain-entity-card__tag-group">
            {f.connector_slugs?.map((slug) => (
              <StatusPill key={slug} status={connectorStatus(slug)} label={slug} />
            ))}
          </div>
        ))}
      </div>
      {composer.last_run_at && (
        <p className="brain-entity-card__footer">Last run: {formatRelativeTime(composer.last_run_at)}</p>
      )}
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
    <Card className="brain-entity-card">
      <div className="brain-entity-card__header">
        <div className="brain-entity-card__title-row">
          <Bell className="brain-icon-sm brain-icon-pending" />
          <h3 className="brain-entity-card__title">{trigger.name}</h3>
        </div>
        <div className="brain-entity-card__header-actions">
          <StatusPill status={trigger.is_active ? 'live' : 'pending'} label={trigger.is_active ? 'Active' : 'Paused'} />
          <button className="brain-icon-btn brain-icon-btn--bad" onClick={() => onDelete(trigger.id)}>
            <Trash2 className="brain-icon-sm" />
          </button>
        </div>
      </div>
      <div className="brain-entity-card__tags">
        <span className="brain-entity-card__tag-label">Types:</span>
        {trigger.signal_types.map((t) => (
          <span key={t} className="brain-entity-card__inline-tag">{t}</span>
        ))}
      </div>
      <div className="brain-entity-card__meta">
        <span>Min importance: P{trigger.min_importance}</span>
        <span>·</span>
        <span>Action: {trigger.action}</span>
        <span>·</span>
        <span>Schedule: {trigger.schedule}</span>
      </div>
      {trigger.last_fired_at && (
        <p className="brain-entity-card__footer">Last fired: {formatRelativeTime(trigger.last_fired_at)}</p>
      )}
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
    try { setStats(await brainApi.getStats()); } catch (err) { console.error('Failed to load brain stats:', err); }
  }, []);

  const loadSignals = useCallback(async () => {
    try {
      const params: Record<string, string | number> = { limit: 100 };
      if (connectorFilter) params.connector = connectorFilter;
      const res = await brainApi.listSignals(params);
      setSignals(res.signals || []);
    } catch (err) { console.error('Failed to load signals:', err); }
  }, [connectorFilter]);

  const loadComposers = useCallback(async () => {
    try { setComposers(await brainApi.listComposers()); } catch (err) { console.error('Failed to load composers:', err); }
  }, []);

  const loadTriggers = useCallback(async () => {
    try { setTriggers(await brainApi.listTriggers()); } catch (err) { console.error('Failed to load triggers:', err); }
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
    try { await brainApi.deleteSignal(id); setSignals((prev) => prev.filter((s) => s.id !== id)); loadStats(); } catch (err) { console.error('Failed to delete signal:', err); }
  };

  const handleFeedback = async (signalId: string, helpful: boolean) => {
    try { await brainApi.submitFeedback({ signal_id: signalId, helpful }); } catch (err) { console.error('Failed to submit feedback:', err); }
  };

  const handlePurge = async () => {
    setPurging(true);
    try { await brainApi.purgeSignals(); setSignals([]); await loadStats(); setPurgeDialogOpen(false); } catch (err) { console.error('Failed to purge signals:', err); } finally { setPurging(false); }
  };

  const handleDeleteComposer = async (id: string) => {
    try { await brainApi.deleteComposer(id); setComposers((prev) => prev.filter((c) => c.id !== id)); } catch (err) { console.error('Failed to delete composer:', err); }
  };

  const handleDeleteTrigger = async (id: string) => {
    try { await brainApi.deleteTrigger(id); setTriggers((prev) => prev.filter((t) => t.id !== id)); } catch (err) { console.error('Failed to delete trigger:', err); }
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
    { id: 'signals', label: 'Signals', icon: <Activity className="brain-icon-sm" />, count: signals.length },
    { id: 'composers', label: 'Composers', icon: <Calendar className="brain-icon-sm" />, count: composers.length },
    { id: 'triggers', label: 'Triggers', icon: <Bell className="brain-icon-sm" />, count: triggers.length },
  ];

  if (loading) {
    return (
      <div className="brain-page">
        <PageGrid />
        <div className="brain-loading">
          <Loader2 className="brain-loading__spinner" />
        </div>
      </div>
    );
  }

  return (
    <div className="brain-page">
      <PageGrid />

      {/* Hero */}
      <Chamber className="brain-hero" ribs>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="MODULE BR-01" secondary="AI Brain" position="top-right" />

        <div className="brain-hero__header">
          <div className="brain-hero__title-row">
            <TrustSeal size="lg" />
            <h1 className="brain-hero__title">Brain</h1>
          </div>
          <p className="brain-hero__subtitle">
            Your personal AI memory. Signals from connected accounts are scored, stored, and injected into agent context.
          </p>
          <div className="brain-hero__actions">
            <FrameButton
              size="sm"
              iconLeft={<RefreshCw className="brain-icon-sm" />}
              onClick={() => { loadStats(); loadSignals(); }}
            >
              Refresh
            </FrameButton>
            <button className="brain-purge-btn" onClick={() => setPurgeDialogOpen(true)}>
              <Trash2 className="brain-icon-sm" />
              Purge All
            </button>
          </div>
        </div>

        {stats && <StatsOverview stats={stats} />}
      </Chamber>

      {/* Tabs */}
      <div className="brain-tabs">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`brain-tab ${activeTab === tab.id ? 'brain-tab--active' : ''}`}
          >
            {tab.icon}
            {tab.label}
            {tab.count !== undefined && (
              <span className="brain-tab__count">{tab.count}</span>
            )}
          </button>
        ))}
      </div>

      {/* Signals Tab */}
      {activeTab === 'signals' && (
        <div className="brain-tab-content">
          <div className="brain-search-bar">
            <div className="brain-search-input-wrap">
              <Search className="brain-search-icon" />
              <input
                type="text"
                placeholder="Search signals..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="brain-search-input"
              />
              {searchQuery && (
                <button onClick={() => setSearchQuery('')} className="brain-search-clear">
                  <X className="brain-icon-xs" />
                </button>
              )}
            </div>
            <select
              value={connectorFilter}
              onChange={(e) => setConnectorFilter(e.target.value)}
              className="brain-filter-select"
            >
              <option value="">All connectors</option>
              {stats && Object.keys(stats.signals_by_connector).map((slug) => (
                <option key={slug} value={slug}>{slug}</option>
              ))}
            </select>
          </div>

          {filteredSignals.length === 0 ? (
            <Chamber className="brain-empty">
              <Activity className="brain-empty__icon" />
              <h3 className="brain-empty__title">{searchQuery ? 'No matching signals' : 'No signals yet'}</h3>
              <p className="brain-empty__desc">
                {searchQuery ? 'Try adjusting your search or filter.' : "Link a connector to start building your Brain's memory."}
              </p>
            </Chamber>
          ) : (
            <Chamber className="brain-signals-chamber">
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
            </Chamber>
          )}
        </div>
      )}

      {/* Composers Tab */}
      {activeTab === 'composers' && (
        <div className="brain-tab-content">
          {composers.length === 0 ? (
            <Chamber className="brain-empty">
              <Calendar className="brain-empty__icon" />
              <h3 className="brain-empty__title">No composers configured</h3>
              <p className="brain-empty__desc">Create a Brain Composer to get automated daily briefings from your connected accounts.</p>
            </Chamber>
          ) : (
            <div className="brain-entity-grid">
              <AnimatePresence mode="popLayout">
                {composers.map((composer) => (
                  <ComposerCard key={composer.id} composer={composer} onDelete={handleDeleteComposer} />
                ))}
              </AnimatePresence>
            </div>
          )}
        </div>
      )}

      {/* Triggers Tab */}
      {activeTab === 'triggers' && (
        <div className="brain-tab-content">
          {triggers.length === 0 ? (
            <Chamber className="brain-empty">
              <Bell className="brain-empty__icon" />
              <h3 className="brain-empty__title">No triggers configured</h3>
              <p className="brain-empty__desc">Create a Brain Trigger to run agents automatically when specific signal patterns are detected.</p>
            </Chamber>
          ) : (
            <div className="brain-entity-grid">
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
      {purgeDialogOpen && (
        <div className="brain-modal-overlay" onClick={() => setPurgeDialogOpen(false)}>
          <div className="brain-modal" onClick={(e) => e.stopPropagation()}>
            <div className="brain-modal__header">
              <AlertTriangle className="brain-icon-sm brain-icon-danger" />
              <h2 className="brain-modal__title">Purge All Brain Signals</h2>
            </div>
            <p className="brain-modal__body">
              This will permanently delete all {stats?.total_signals || 0} signals from your Brain.
              This action cannot be undone. Your connectors will continue to generate new signals.
            </p>
            <div className="brain-modal__actions">
              <FrameButton onClick={() => setPurgeDialogOpen(false)}>Cancel</FrameButton>
              <SealedButton onClick={handlePurge} disabled={purging} loading={purging}>
                Purge All Signals
              </SealedButton>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
