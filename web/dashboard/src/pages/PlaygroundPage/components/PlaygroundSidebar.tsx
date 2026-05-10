import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { QRCodeSVG } from 'qrcode.react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
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
  Variable,
  Plus,
  X,
  Edit3,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { cn } from '@/lib/utils';
import { usePlaygroundStore, SidebarPanel, ExecutionHistoryItem, PlaygroundVariable } from '../store/playgroundStore';
import { usePlaygroundState } from '../hooks/usePlaygroundState';
import { SchemaExplorer } from './SchemaExplorer';
import { CodeSnippetGenerator } from './CodeSnippetGenerator';

interface PlaygroundSidebarProps {
  className?: string;
}

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
          <CheckCircle2 className="w-3.5 h-3.5 text-green-500 dark:text-green-400" />
        ) : (
          <XCircle className="w-3.5 h-3.5 text-red-500 dark:text-red-400" />
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
  const { t } = useTranslation();
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
                  ? 'bg-indigo-600 text-white dark:text-white'
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
              title={t('playground.exportHistory')}
            >
              <Download className="w-3.5 h-3.5" />
            </button>
            <button
              onClick={clearHistory}
              className="p-1 rounded hover:bg-bg-tertiary text-text-muted hover:text-red-400 transition-colors"
              title={t('playground.clearHistory')}
        >
          <Trash2 className="w-3.5 h-3.5" />
        </button>
      </div>

      {/* List */}
      <div className="flex-1 overflow-auto p-2">
        {filtered.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-8 text-center">
            <History className="w-8 h-8 text-text-muted mb-2" />
            <p className="text-xs text-text-muted">{t('playground.noHistoryYet')}</p>
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
  const { t } = useTranslation();
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
                ? 'bg-indigo-600 text-white dark:text-white'
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
            <p className="text-xs text-text-muted text-center py-8">{t('playground.noInputSchemaLabel')}</p>
          )
        ) : outputSchema ? (
          <SchemaExplorer schema={outputSchema} />
        ) : (
          <p className="text-xs text-text-muted text-center py-8">{t('playground.noOutputSchema')}</p>
        )}
      </div>
    </div>
  );
}

// ─── Variables Panel ─────────────────────────────────────────────────────────

interface VariableItemProps {
  variable: PlaygroundVariable;
  onUpdate: (id: string, updates: Partial<PlaygroundVariable>) => void;
  onDelete: (id: string) => void;
}

function VariableItem({ variable, onUpdate, onDelete }: VariableItemProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [editName, setEditName] = useState(variable.name);
  const [editValue, setEditValue] = useState(variable.value);

  const handleSave = () => {
    if (editName.trim() && editValue.trim()) {
      onUpdate(variable.id, { name: editName.trim(), value: editValue.trim() });
      setIsEditing(false);
    }
  };

  const handleCancel = () => {
    setEditName(variable.name);
    setEditValue(variable.value);
    setIsEditing(false);
  };

  if (isEditing) {
    return (
      <div className="p-2 bg-bg-tertiary rounded-lg space-y-2">
        <Input
          value={editName}
          onChange={(e) => setEditName(e.target.value)}
          placeholder="name"
          className="h-7 text-xs font-mono"
        />
        <Input
          value={editValue}
          onChange={(e) => setEditValue(e.target.value)}
          placeholder="value"
          className="h-7 text-xs font-mono"
        />
        <div className="flex gap-1">
          <Button size="sm" variant="ghost" onClick={handleSave} className="h-6 text-xs gap-1 px-1.5">
            <Check className="w-3 h-3" />
            Save
          </Button>
          <Button size="sm" variant="ghost" onClick={handleCancel} className="h-6 text-xs gap-1 px-1.5">
            <X className="w-3 h-3" />
            Cancel
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="group flex items-start gap-2 p-2 rounded-lg hover:bg-bg-tertiary transition-colors">
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-1.5">
          <span className="text-xs text-purple-500 dark:text-purple-300 font-mono">
            {`{{${variable.name}}}`}
          </span>
        </div>
        <p className="text-xs text-text-muted font-mono truncate mt-0.5" title={variable.value}>
          {variable.value.length > 40 ? variable.value.slice(0, 40) + '...' : variable.value}
        </p>
        {variable.description && (
          <p className="text-[10px] text-text-muted mt-0.5">{variable.description}</p>
        )}
      </div>
      <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
        <button
          onClick={() => setIsEditing(true)}
          className="p-0.5 rounded hover:bg-bg-secondary transition-all"
          title="Edit variable"
        >
          <Edit3 className="w-3 h-3 text-text-muted hover:text-text-secondary" />
        </button>
        <button
          onClick={() => onDelete(variable.id)}
          className="p-0.5 rounded hover:bg-bg-secondary transition-all"
          title="Delete variable"
        >
          <Trash2 className="w-3 h-3 text-text-muted hover:text-red-400" />
        </button>
      </div>
    </div>
  );
}

function VariablesPanel() {
  const { t } = useTranslation();
  const { variables, addVariable, updateVariable, removeVariable, clearVariables } = usePlaygroundStore();
  const [showAddForm, setShowAddForm] = useState(false);
  const [newName, setNewName] = useState('');
  const [newValue, setNewValue] = useState('');
  const [newDescription, setNewDescription] = useState('');

  const handleAdd = () => {
    if (newName.trim() && newValue.trim()) {
      addVariable({
        id: crypto.randomUUID(),
        name: newName.trim(),
        value: newValue.trim(),
        description: newDescription.trim() || undefined,
      });
      setNewName('');
      setNewValue('');
      setNewDescription('');
      setShowAddForm(false);
    }
  };

  return (
    <div className="flex flex-col h-full">
      {/* Toolbar */}
      <div className="flex items-center gap-2 px-3 py-2 border-b border-border-subtle shrink-0">
        <div className="flex-1 text-xs text-text-muted">
          {t('playground.variablesTooltip')}
        </div>
        <button
          onClick={() => setShowAddForm(true)}
          className="p-1 rounded hover:bg-bg-tertiary text-text-muted hover:text-text-secondary transition-colors"
          title={t('playground.addVariable')}
        >
          <Plus className="w-3.5 h-3.5" />
        </button>
        {variables.length > 0 && (
          <button
            onClick={clearVariables}
            className="p-1 rounded hover:bg-bg-tertiary text-text-muted hover:text-red-400 transition-colors"
            title={t('playground.clearVariables')}
          >
            <Trash2 className="w-3.5 h-3.5" />
          </button>
        )}
      </div>

      {/* Add form */}
      {showAddForm && (
        <div className="p-2 border-b border-border-subtle bg-bg-tertiary/50">
          <div className="space-y-2">
            <Input
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              placeholder="variable_name"
              className="h-7 text-xs font-mono"
            />
            <Input
              value={newValue}
              onChange={(e) => setNewValue(e.target.value)}
              placeholder="value"
              className="h-7 text-xs font-mono"
            />
            <Input
              value={newDescription}
              onChange={(e) => setNewDescription(e.target.value)}
              placeholder="description (optional)"
              className="h-7 text-xs"
            />
            <div className="flex gap-1">
              <Button size="sm" variant="ghost" onClick={handleAdd} className="h-6 text-xs gap-1 px-1.5">
                <Check className="w-3 h-3" />
                Add
              </Button>
              <Button size="sm" variant="ghost" onClick={() => setShowAddForm(false)} className="h-6 text-xs gap-1 px-1.5">
                <X className="w-3 h-3" />
                Cancel
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* List */}
      <div className="flex-1 overflow-auto p-2">
        {variables.length === 0 && !showAddForm ? (
          <div className="flex flex-col items-center justify-center py-8 text-center">
            <Variable className="w-8 h-8 text-text-muted mb-2" />
            <p className="text-xs text-text-muted">{t('playground.noVariables')}</p>
            <p className="text-[10px] text-text-muted mt-1">{t('playground.variablesHint')}</p>
          </div>
        ) : (
          <AnimatePresence>
            {variables.map((variable) => (
              <VariableItem
                key={variable.id}
                variable={variable}
                onUpdate={updateVariable}
                onDelete={removeVariable}
              />
            ))}
          </AnimatePresence>
        )}
      </div>
    </div>
  );
}

// ─── Share Panel ──────────────────────────────────────────────────────────────

function SharePanel() {
  const { t } = useTranslation();
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
        <p className="text-xs font-medium text-text-secondary mb-2">{t('playground.shareableLink')}</p>
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
            {copied ? t('playground.copied') : t('playground.copy')}
          </Button>
        </div>
      </div>

      {/* QR Code */}
      <div>
        <p className="text-xs font-medium text-text-secondary mb-2">{t('playground.qrCode')}</p>
        <div className="flex justify-center p-3 bg-bg-tertiary rounded-lg">
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
          {t('playground.openInNewTab')}
        </a>
      </Button>

      {/* Embed snippet */}
      <div>
        <p className="text-xs font-medium text-text-secondary mb-2">{t('playground.embed')}</p>
        <pre className="text-[10px] font-mono bg-bg-tertiary border border-border-subtle rounded p-2 overflow-auto text-text-secondary">
          {`<iframe\n  src="${shareableUrl}"\n  width="100%"\n  height="600"\n  frameborder="0"\n/>`}
        </pre>
      </div>
    </div>
  );
}

// ─── Info Panel ───────────────────────────────────────────────────────────────

function InfoPanel() {
  const { t } = useTranslation();
  const { functionInfo } = usePlaygroundStore();

  if (!functionInfo) return null;

  return (
    <div className="p-3 space-y-3">
      <div className="space-y-2">
        {[
          { label: t('playground.author'), value: functionInfo.author },
          { label: t('playground.name'), value: functionInfo.name },
          { label: t('playground.version'), value: `v${functionInfo.version}` },
          { label: t('playground.runtime'), value: functionInfo.runtime || 'Unknown' },
        ].map((item) => (
          <div key={item.label} className="flex items-center justify-between text-xs">
            <span className="text-text-muted">{item.label}</span>
            <span className="font-mono text-text-secondary">{item.value}</span>
          </div>
        ))}
      </div>

      {functionInfo.description && (
        <div className="pt-2 border-t border-border-subtle">
          <p className="text-xs text-text-muted mb-1">{t('playground.description')}</p>
          <p className="text-xs text-text-secondary">{functionInfo.description}</p>
        </div>
      )}

      {functionInfo.tags && functionInfo.tags.length > 0 && (
        <div className="pt-2 border-t border-border-subtle">
          <p className="text-xs text-text-muted mb-2 flex items-center gap-1">
            <Tag className="w-3 h-3" />
            {t('playground.tags')}
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
              {t('playground.pricePerCall')}
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
              {t('playground.reliability')}
            </span>
            <span
              className={cn(
                'font-mono',
                functionInfo.reliability_score >= 99
                  ? 'text-green-500 dark:text-green-400'
                  : functionInfo.reliability_score >= 95
                  ? 'text-yellow-500 dark:text-yellow-400'
                  : 'text-red-500 dark:text-red-400'
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
              {t('playground.cacheTtl')}
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
          {t('playground.viewFullDocs')}
        </a>
      </Button>
    </div>
  );
}

// ─── Main Sidebar ─────────────────────────────────────────────────────────────

export function PlaygroundSidebar({ className }: PlaygroundSidebarProps) {
  const { t } = useTranslation();
  const { activeSidebarPanel, setActiveSidebarPanel } = usePlaygroundStore();

  const PANELS: Array<{ id: SidebarPanel; icon: React.ReactNode; label: string }> = [
    { id: 'history', icon: <History className="w-4 h-4" />, label: t('playground.history') },
    { id: 'variables', icon: <Variable className="w-4 h-4" />, label: t('playground.variables') },
    { id: 'schema', icon: <TreePine className="w-4 h-4" />, label: t('playground.schema') },
    { id: 'snippets', icon: <Code2 className="w-4 h-4" />, label: t('playground.snippets') },
    { id: 'share', icon: <Share2 className="w-4 h-4" />, label: t('playground.share') },
    { id: 'info', icon: <Info className="w-4 h-4" />, label: t('playground.info') },
  ];

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
                ? 'bg-indigo-600 text-white dark:text-white'
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
            {activeSidebarPanel === 'variables' && <VariablesPanel />}
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
