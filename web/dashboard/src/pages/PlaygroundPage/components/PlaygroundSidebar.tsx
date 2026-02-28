import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { QRCodeSVG } from 'qrcode.react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  History,
  TreePine,
  Code2,
  Share2,
  Info,
  Trash2,
  Download,
  CheckCircle2,
  XCircle,
  Clock,
  Copy,
  Check,
  ExternalLink,
  Tag,
  DollarSign,
  Shield,
  Database,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { usePlaygroundStore, SidebarPanel, ExecutionHistoryItem } from '../store/playgroundStore';
import { usePlaygroundState } from '../hooks/usePlaygroundState';
import { SchemaExplorer } from './SchemaExplorer';
import { CodeSnippetGenerator } from './CodeSnippetGenerator';

interface PlaygroundSidebarProps {
  className?: string;
}

const PANELS: Array<{ id: SidebarPanel; icon: React.ReactNode; label: string }> = [
  { id: 'history', icon: <History className="w-4 h-4" />, label: 'History' },
  { id: 'schema', icon: <TreePine className="w-4 h-4" />, label: 'Schema' },
  { id: 'snippets', icon: <Code2 className="w-4 h-4" />, label: 'Snippets' },
  { id: 'share', icon: <Share2 className="w-4 h-4" />, label: 'Share' },
  { id: 'info', icon: <Info className="w-4 h-4" />, label: 'Info' },
];

// ─── History Panel ────────────────────────────────────────────────────────────

function HistoryItem({
  item,
  onLoad,
  onDelete,
}: {
  item: ExecutionHistoryItem;
  onLoad: () => void;
  onDelete: () => void;
}) {
  const date = new Date(item.timestamp);
  const inputPreview = JSON.stringify(item.input).slice(0, 60);

  return (
    <motion.div
      initial={{ opacity: 0, x: -8 }}
      animate={{ opacity: 1, x: 0 }}
      exit={{ opacity: 0, x: -8 }}
      className="group flex items-start gap-2 p-2 rounded-lg hover:bg-bg-tertiary cursor-pointer transition-colors"
      onClick={onLoad}
    >
      <div className="shrink-0 mt-0.5">
        {item.result.ok ? (
          <CheckCircle2 className="w-3.5 h-3.5 text-green-400" />
        ) : (
          <XCircle className="w-3.5 h-3.5 text-red-400" />
        )}
      </div>

      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-1.5">
          <span className="text-xs text-text-muted flex items-center gap-1">
            <Clock className="w-3 h-3" />
            {date.toLocaleTimeString()}
          </span>
          <span className="text-xs text-text-muted">{item.result.duration_ms}ms</span>
          {item.result.cached && (
            <Badge className="text-[9px] px-1 py-0 h-3.5 bg-amber-500/10 text-amber-400 border-amber-500/20">
              cached
            </Badge>
          )}
        </div>
        <p className="text-xs text-text-secondary font-mono truncate mt-0.5">
          {inputPreview}
          {inputPreview.length >= 60 ? '...' : ''}
        </p>
      </div>

      <button
        onClick={(e) => {
          e.stopPropagation();
          onDelete();
        }}
        className="opacity-0 group-hover:opacity-100 p-0.5 rounded hover:bg-bg-secondary transition-all shrink-0"
      >
        <Trash2 className="w-3 h-3 text-text-muted hover:text-red-400" />
      </button>
    </motion.div>
  );
}

function HistoryPanel() {
  const { executionHistory, loadFromHistory, removeFromHistory, clearHistory } =
    usePlaygroundStore();
  const [filter, setFilter] = useState<'all' | 'success' | 'error'>('all');

  const filtered = executionHistory.filter((item) => {
    if (filter === 'success') return item.result.ok;
    if (filter === 'error') return !item.result.ok;
    return true;
  });

  const handleExport = () => {
    const data = JSON.stringify(executionHistory, null, 2);
    const blob = new Blob([data], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'playground-history.json';
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className="flex flex-col h-full">
      {/* Toolbar */}
      <div className="flex items-center gap-2 px-3 py-2 border-b border-border-subtle shrink-0">
        <div className="flex gap-1">
          {(['all', 'success', 'error'] as const).map((f) => (
            <button
              key={f}
              onClick={() => setFilter(f)}
              className={cn(
                'px-2 py-0.5 text-[10px] rounded transition-colors capitalize',
                filter === f
                  ? 'bg-indigo-600 text-white'
                  : 'text-text-muted hover:text-text-secondary'
              )}
            >
              {f}
            </button>
          ))}
        </div>
        <div className="flex-1" />
        <button
          onClick={handleExport}
          className="p-1 rounded hover:bg-bg-tertiary text-text-muted hover:text-text-secondary transition-colors"
          title="Export history"
        >
          <Download className="w-3.5 h-3.5" />
        </button>
        <button
          onClick={clearHistory}
          className="p-1 rounded hover:bg-bg-tertiary text-text-muted hover:text-red-400 transition-colors"
          title="Clear history"
        >
          <Trash2 className="w-3.5 h-3.5" />
        </button>
      </div>

      {/* List */}
      <div className="flex-1 overflow-auto p-2">
        {filtered.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-8 text-center">
            <History className="w-8 h-8 text-text-muted mb-2" />
            <p className="text-xs text-text-muted">No history yet</p>
          </div>
        ) : (
          <AnimatePresence>
            {filtered.map((item) => (
              <HistoryItem
                key={item.id}
                item={item}
                onLoad={() => loadFromHistory(item)}
                onDelete={() => removeFromHistory(item.id)}
              />
            ))}
          </AnimatePresence>
        )}
      </div>
    </div>
  );
}

// ─── Schema Panel ─────────────────────────────────────────────────────────────

function SchemaPanel() {
  const { functionInfo } = usePlaygroundStore();
  const inputSchema = functionInfo?.manifest?.input?.schema as Record<string, unknown> | undefined;
  const outputSchema = functionInfo?.manifest?.output?.schema as Record<string, unknown> | undefined;
  const [activeSchema, setActiveSchema] = useState<'input' | 'output'>('input');

  return (
    <div className="flex flex-col h-full">
      <div className="flex gap-1 px-3 py-2 border-b border-border-subtle shrink-0">
        {(['input', 'output'] as const).map((s) => (
          <button
            key={s}
            onClick={() => setActiveSchema(s)}
            className={cn(
              'px-2 py-0.5 text-[10px] rounded transition-colors capitalize',
              activeSchema === s
                ? 'bg-indigo-600 text-white'
                : 'text-text-muted hover:text-text-secondary'
            )}
          >
            {s}
          </button>
        ))}
      </div>
      <div className="flex-1 overflow-auto p-2">
        {activeSchema === 'input' ? (
          inputSchema ? (
            <SchemaExplorer schema={inputSchema} />
          ) : (
            <p className="text-xs text-text-muted text-center py-8">No input schema</p>
          )
        ) : outputSchema ? (
          <SchemaExplorer schema={outputSchema} />
        ) : (
          <p className="text-xs text-text-muted text-center py-8">No output schema</p>
        )}
      </div>
    </div>
  );
}

// ─── Share Panel ──────────────────────────────────────────────────────────────

function SharePanel() {
  const { shareableUrl } = usePlaygroundState();
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(shareableUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // ignore
    }
  };

  return (
    <div className="p-3 space-y-4">
      <div>
        <p className="text-xs font-medium text-text-secondary mb-2">Shareable Link</p>
        <div className="flex gap-2">
          <input
            value={shareableUrl}
            readOnly
            className="flex-1 text-xs font-mono bg-bg-tertiary border border-border-subtle rounded px-2 py-1.5 text-text-secondary truncate"
          />
          <Button
            size="sm"
            variant="outline"
            onClick={handleCopy}
            className="h-7 gap-1.5 text-xs shrink-0"
          >
            {copied ? (
              <Check className="w-3 h-3 text-green-400" />
            ) : (
              <Copy className="w-3 h-3" />
            )}
            {copied ? 'Copied!' : 'Copy'}
          </Button>
        </div>
      </div>

      {/* QR Code */}
      <div>
        <p className="text-xs font-medium text-text-secondary mb-2">QR Code</p>
        <div className="flex justify-center p-3 bg-white rounded-lg">
          <QRCodeSVG value={shareableUrl} size={120} />
        </div>
      </div>

      {/* Open link */}
      <Button
        variant="outline"
        size="sm"
        className="w-full gap-2 text-xs"
        asChild
      >
        <a href={shareableUrl} target="_blank" rel="noopener noreferrer">
          <ExternalLink className="w-3.5 h-3.5" />
          Open in new tab
        </a>
      </Button>

      {/* Embed snippet */}
      <div>
        <p className="text-xs font-medium text-text-secondary mb-2">Embed</p>
        <pre className="text-[10px] font-mono bg-bg-tertiary border border-border-subtle rounded p-2 overflow-auto text-text-secondary">
          {`<iframe\n  src="${shareableUrl}"\n  width="100%"\n  height="600"\n  frameborder="0"\n/>`}
        </pre>
      </div>
    </div>
  );
}

// ─── Info Panel ───────────────────────────────────────────────────────────────

function InfoPanel() {
  const { functionInfo } = usePlaygroundStore();

  if (!functionInfo) return null;

  return (
    <div className="p-3 space-y-3">
      <div className="space-y-2">
        {[
          { label: 'Author', value: functionInfo.author },
          { label: 'Name', value: functionInfo.name },
          { label: 'Version', value: `v${functionInfo.version}` },
          { label: 'Runtime', value: functionInfo.runtime || 'Unknown' },
        ].map((item) => (
          <div key={item.label} className="flex items-center justify-between text-xs">
            <span className="text-text-muted">{item.label}</span>
            <span className="font-mono text-text-secondary">{item.value}</span>
          </div>
        ))}
      </div>

      {functionInfo.description && (
        <div className="pt-2 border-t border-border-subtle">
          <p className="text-xs text-text-muted mb-1">Description</p>
          <p className="text-xs text-text-secondary">{functionInfo.description}</p>
        </div>
      )}

      {functionInfo.tags && functionInfo.tags.length > 0 && (
        <div className="pt-2 border-t border-border-subtle">
          <p className="text-xs text-text-muted mb-2 flex items-center gap-1">
            <Tag className="w-3 h-3" />
            Tags
          </p>
          <div className="flex flex-wrap gap-1">
            {functionInfo.tags.map((tag) => (
              <Badge
                key={tag}
                variant="outline"
                className="text-[10px] px-1.5 py-0 h-4 border-border-subtle text-text-muted"
              >
                {tag}
              </Badge>
            ))}
          </div>
        </div>
      )}

      <div className="pt-2 border-t border-border-subtle space-y-2">
        {functionInfo.price_per_call !== undefined && (
          <div className="flex items-center justify-between text-xs">
            <span className="text-text-muted flex items-center gap-1">
              <DollarSign className="w-3 h-3" />
              Price per call
            </span>
            <span className="font-mono text-text-secondary">
              ${functionInfo.price_per_call.toFixed(6)}
            </span>
          </div>
        )}

        {functionInfo.reliability_score !== undefined && (
          <div className="flex items-center justify-between text-xs">
            <span className="text-text-muted flex items-center gap-1">
              <Shield className="w-3 h-3" />
              Reliability
            </span>
            <span
              className={cn(
                'font-mono',
                functionInfo.reliability_score >= 99
                  ? 'text-green-400'
                  : functionInfo.reliability_score >= 95
                  ? 'text-yellow-400'
                  : 'text-red-400'
              )}
            >
              {functionInfo.reliability_score}%
            </span>
          </div>
        )}

        {functionInfo.cache_ttl !== undefined && functionInfo.cache_ttl > 0 && (
          <div className="flex items-center justify-between text-xs">
            <span className="text-text-muted flex items-center gap-1">
              <Database className="w-3 h-3" />
              Cache TTL
            </span>
            <span className="font-mono text-text-secondary">{functionInfo.cache_ttl}s</span>
          </div>
        )}
      </div>

      <Button
        variant="outline"
        size="sm"
        className="w-full gap-2 text-xs mt-2"
        asChild
      >
        <a href={`/fx/${functionInfo.author}/${functionInfo.name}`}>
          <ExternalLink className="w-3.5 h-3.5" />
          View full docs
        </a>
      </Button>
    </div>
  );
}

// ─── Main Sidebar ─────────────────────────────────────────────────────────────

export function PlaygroundSidebar({ className }: PlaygroundSidebarProps) {
  const { activeSidebarPanel, setActiveSidebarPanel } = usePlaygroundStore();

  return (
    <div className={cn('flex h-full border-l border-border-subtle bg-bg-primary', className)}>
      {/* Icon nav */}
      <div className="flex flex-col items-center gap-1 py-2 px-1 border-r border-border-subtle bg-bg-secondary w-10 shrink-0">
        {PANELS.map((panel) => (
          <button
            key={panel.id}
            onClick={() => setActiveSidebarPanel(panel.id)}
            title={panel.label}
            className={cn(
              'w-8 h-8 flex items-center justify-center rounded-md transition-colors',
              activeSidebarPanel === panel.id
                ? 'bg-indigo-600 text-white'
                : 'text-text-muted hover:text-text-secondary hover:bg-bg-tertiary'
            )}
          >
            {panel.icon}
          </button>
        ))}
      </div>

      {/* Panel content */}
      <div className="flex-1 min-w-0 overflow-hidden">
        <div className="px-3 py-2 border-b border-border-subtle shrink-0">
          <p className="text-xs font-medium text-text-secondary">
            {PANELS.find((p) => p.id === activeSidebarPanel)?.label}
          </p>
        </div>

        <AnimatePresence mode="wait">
          <motion.div
            key={activeSidebarPanel}
            initial={{ opacity: 0, x: 8 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: -8 }}
            transition={{ duration: 0.15 }}
            className="h-[calc(100%-36px)] overflow-hidden"
          >
            {activeSidebarPanel === 'history' && <HistoryPanel />}
            {activeSidebarPanel === 'schema' && <SchemaPanel />}
            {activeSidebarPanel === 'snippets' && (
              <div className="p-3 overflow-auto h-full">
                <CodeSnippetGenerator />
              </div>
            )}
            {activeSidebarPanel === 'share' && <SharePanel />}
            {activeSidebarPanel === 'info' && <InfoPanel />}
          </motion.div>
        </AnimatePresence>
      </div>
    </div>
  );
}
