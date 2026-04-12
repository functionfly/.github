import { useState, useMemo } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import type { Snapshot } from '@/types';
import {
  GitCompare,
  ArrowRight,
  Plus,
  Minus,
  AlertCircle,
  FileJson,
  Clock,
  User,
  Hash,
  ChevronDown,
  ChevronRight,
  Copy,
  Check,
} from 'lucide-react';

interface StateDiffViewerProps {
  snapshots: Snapshot[];
  currentState?: Record<string, unknown>;
}

type DiffType = 'added' | 'removed' | 'modified' | 'unchanged';

interface DiffEntry {
  key: string;
  path: string;
  type: DiffType;
  oldValue?: unknown;
  newValue?: unknown;
}

interface DiffResult {
  entries: DiffEntry[];
  stats: {
    added: number;
    removed: number;
    modified: number;
    unchanged: number;
  };
}

function computeDiff(oldObj: Record<string, unknown>, newObj: Record<string, unknown>): DiffResult {
  const entries: DiffEntry[] = [];
  const stats = { added: 0, removed: 0, modified: 0, unchanged: 0 };

  const allKeys = new Set([...Object.keys(oldObj), ...Object.keys(newObj)]);

  for (const key of allKeys) {
    const oldVal = oldObj[key];
    const newVal = newObj[key];

    if (!(key in oldObj)) {
      entries.push({ key, path: key, type: 'added', newValue: newVal });
      stats.added++;
    } else if (!(key in newObj)) {
      entries.push({ key, path: key, type: 'removed', oldValue: oldVal });
      stats.removed++;
    } else if (JSON.stringify(oldVal) !== JSON.stringify(newVal)) {
      entries.push({
        key,
        path: key,
        type: 'modified',
        oldValue: oldVal,
        newValue: newVal,
      });
      stats.modified++;
    } else {
      entries.push({ key, path: key, type: 'unchanged', oldValue: oldVal, newValue: newVal });
      stats.unchanged++;
    }
  }

  return { entries, stats };
}

function formatValue(value: unknown): string {
  if (value === null) return 'null';
  if (value === undefined) return 'undefined';
  if (typeof value === 'string') return `"${value}"`;
  if (typeof value === 'number' || typeof value === 'boolean') return String(value);
  if (Array.isArray(value)) return `[${value.length} items]`;
  if (typeof value === 'object') {
    const keys = Object.keys(value as Record<string, unknown>);
    return `{${keys.length} keys}`;
  }
  return String(value);
}

function ValuePreview({ value }: { value: unknown }) {
  const [isExpanded, setIsExpanded] = useState(false);

  if (value === null || value === undefined) {
    return <span className="text-muted-foreground italic">{String(value)}</span>;
  }

  if (typeof value !== 'object') {
    return <code className="text-xs bg-bg-tertiary px-1 py-0.5 rounded">{formatValue(value)}</code>;
  }

  return (
    <div className="space-y-1">
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
      >
        {isExpanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
        {formatValue(value)}
      </button>
      {isExpanded && (
        <pre className="text-xs bg-bg-tertiary p-2 rounded overflow-x-auto max-h-40">
          {JSON.stringify(value, null, 2)}
        </pre>
      )}
    </div>
  );
}

function DiffEntryRow({
  entry,
  viewMode,
}: {
  entry: DiffEntry;
  viewMode: 'split' | 'unified';
}) {
  const typeColors: Record<DiffType, { bg: string; border: string; icon: React.ReactNode }> = {
    added: {
      bg: 'bg-green-500/10',
      border: 'border-green-500/30',
      icon: <Plus className="h-4 w-4 text-green-600" />,
    },
    removed: {
      bg: 'bg-red-500/10',
      border: 'border-red-500/30',
      icon: <Minus className="h-4 w-4 text-red-600" />,
    },
    modified: {
      bg: 'bg-yellow-500/10',
      border: 'border-yellow-500/30',
      icon: <AlertCircle className="h-4 w-4 text-yellow-600" />,
    },
    unchanged: {
      bg: 'bg-muted',
      border: 'border-border-subtle',
      icon: <div className="h-4 w-4" />,
    },
  };

  const colors = typeColors[entry.type];

  if (entry.type === 'unchanged' && viewMode === 'unified') {
    return null;
  }

  return (
    <div
      className={`flex items-start gap-3 p-3 rounded-lg border ${colors.bg} ${colors.border} ${
        entry.type === 'unchanged' ? 'opacity-50' : ''
      }`}
    >
      <div className="mt-0.5">{colors.icon}</div>
      <div className="flex-1 min-w-0 space-y-1">
        <code className="text-sm font-semibold">{entry.key}</code>

        {viewMode === 'split' ? (
          <div className="grid grid-cols-2 gap-4">
            {entry.oldValue !== undefined && (
              <div>
                <span className="text-xs text-muted-foreground">Before:</span>
                <ValuePreview value={entry.oldValue} />
              </div>
            )}
            {entry.newValue !== undefined && (
              <div>
                <span className="text-xs text-muted-foreground">After:</span>
                <ValuePreview value={entry.newValue} />
              </div>
            )}
          </div>
        ) : (
          <div>
            {entry.type === 'added' && <ValuePreview value={entry.newValue} />}
            {entry.type === 'removed' && <ValuePreview value={entry.oldValue} />}
            {entry.type === 'modified' && (
              <div className="space-y-2">
                <div className="bg-red-500/5 p-2 rounded">
                  <span className="text-xs text-red-600">-</span>
                  <ValuePreview value={entry.oldValue} />
                </div>
                <div className="bg-green-500/5 p-2 rounded">
                  <span className="text-xs text-green-600">+</span>
                  <ValuePreview value={entry.newValue} />
                </div>
              </div>
            )}
          </div>
        )}
      </div>
      <Badge variant="outline" className="capitalize shrink-0">
        {entry.type}
      </Badge>
    </div>
  );
}

export function StateDiffViewer({ snapshots, currentState }: StateDiffViewerProps) {
  const [leftSnapshotId, setLeftSnapshotId] = useState<string | null>(null);
  const [rightSnapshotId, setRightSnapshotId] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<'split' | 'unified'>('split');
  const [showUnchanged, setShowUnchanged] = useState(false);
  const [copied, setCopied] = useState(false);

  // Sort snapshots by date (newest first)
  const sortedSnapshots = useMemo(() => {
    return [...snapshots].sort(
      (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
    );
  }, [snapshots]);

  // Default to comparing most recent with current state
  useMemo(() => {
    if (sortedSnapshots.length > 0 && !leftSnapshotId && !rightSnapshotId) {
      setRightSnapshotId(sortedSnapshots[0].id);
      if (sortedSnapshots.length > 1) {
        setLeftSnapshotId(sortedSnapshots[1].id);
      }
    }
  }, [sortedSnapshots, leftSnapshotId, rightSnapshotId]);

  const leftSnapshot = sortedSnapshots.find((s) => s.id === leftSnapshotId);
  const rightSnapshot = sortedSnapshots.find((s) => s.id === rightSnapshotId);

  // Compute diff
  const diffResult = useMemo<DiffResult | null>(() => {
    if (!rightSnapshot) return null;

    const leftData = leftSnapshot?.state || currentState || {};
    const rightData = rightSnapshot?.state || {};

    return computeDiff(leftData, rightData);
  }, [leftSnapshot, rightSnapshot, currentState]);

  const filteredEntries = useMemo(() => {
    if (!diffResult) return [];
    return showUnchanged
      ? diffResult.entries
      : diffResult.entries.filter((e) => e.type !== 'unchanged');
  }, [diffResult, showUnchanged]);

  const handleCopyDiff = () => {
    if (!diffResult) return;
    const diffJson = JSON.stringify(
      {
        from: leftSnapshot?.name || 'Current State',
        to: rightSnapshot?.name || 'Current State',
        timestamp: new Date().toISOString(),
        changes: diffResult.entries.filter((e) => e.type !== 'unchanged'),
      },
      null,
      2
    );
    navigator.clipboard.writeText(diffJson);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="space-y-6">
      {/* Header with Controls */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <GitCompare className="h-5 w-5" />
              <CardTitle>State Version Comparison</CardTitle>
            </div>
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <Switch
                  id="showUnchanged"
                  checked={showUnchanged}
                  onCheckedChange={setShowUnchanged}
                />
                <Label htmlFor="showUnchanged" className="text-sm cursor-pointer">
                  Show unchanged
                </Label>
              </div>
              <div className="flex border rounded-md overflow-hidden">
                <Button
                  variant={viewMode === 'split' ? 'default' : 'ghost'}
                  size="sm"
                  onClick={() => setViewMode('split')}
                  className="rounded-none"
                >
                  Split
                </Button>
                <Button
                  variant={viewMode === 'unified' ? 'default' : 'ghost'}
                  size="sm"
                  onClick={() => setViewMode('unified')}
                  className="rounded-none"
                >
                  Unified
                </Button>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={handleCopyDiff}
                disabled={!diffResult}
              >
                {copied ? (
                  <>
                    <Check className="h-4 w-4 mr-2" />
                    Copied
                  </>
                ) : (
                  <>
                    <Copy className="h-4 w-4 mr-2" />
                    Copy
                  </>
                )}
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Version Selectors */}
          <div className="grid grid-cols-[1fr,auto,1fr] gap-4 items-center">
            <div className="space-y-2">
              <Label className="text-xs text-muted-foreground uppercase">From (Earlier)</Label>
              <Select value={leftSnapshotId ?? 'current'} onValueChange={setLeftSnapshotId}>
                <SelectTrigger>
                  <SelectValue placeholder="Select version" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="current">Current State</SelectItem>
                  {sortedSnapshots.map((snapshot) => (
                    <SelectItem key={snapshot.id} value={snapshot.id}>
                      <div className="flex items-center gap-2">
                        <span>{snapshot.name}</span>
                        <span className="text-xs text-muted-foreground">
                          {new Date(snapshot.createdAt).toLocaleDateString()}
                        </span>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <ArrowRight className="h-5 w-5 text-muted-foreground mt-6" />

            <div className="space-y-2">
              <Label className="text-xs text-muted-foreground uppercase">To (Later)</Label>
              <Select value={rightSnapshotId ?? 'current'} onValueChange={setRightSnapshotId}>
                <SelectTrigger>
                  <SelectValue placeholder="Select version" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="current">Current State</SelectItem>
                  {sortedSnapshots.map((snapshot) => (
                    <SelectItem key={snapshot.id} value={snapshot.id}>
                      <div className="flex items-center gap-2">
                        <span>{snapshot.name}</span>
                        <span className="text-xs text-muted-foreground">
                          {new Date(snapshot.createdAt).toLocaleDateString()}
                        </span>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* Stats */}
          {diffResult && (
            <div className="flex flex-wrap gap-3">
              <Badge variant="outline" className="bg-green-500/10 text-green-600 border-green-500/30">
                <Plus className="h-3 w-3 mr-1" />
                {diffResult.stats.added} added
              </Badge>
              <Badge variant="outline" className="bg-red-500/10 text-red-600 border-red-500/30">
                <Minus className="h-3 w-3 mr-1" />
                {diffResult.stats.removed} removed
              </Badge>
              <Badge variant="outline" className="bg-yellow-500/10 text-yellow-600 border-yellow-500/30">
                <AlertCircle className="h-3 w-3 mr-1" />
                {diffResult.stats.modified} modified
              </Badge>
              {showUnchanged && (
                <Badge variant="outline">
                  {diffResult.stats.unchanged} unchanged
                </Badge>
              )}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Selected Versions Info */}
      <div className="grid grid-cols-2 gap-4">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base flex items-center gap-2">
              <FileJson className="h-4 w-4" />
              From
            </CardTitle>
          </CardHeader>
          <CardContent>
            {leftSnapshot ? (
              <div className="space-y-2">
                <p className="font-medium">{leftSnapshot.name}</p>
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <Clock className="h-3 w-3" />
                  {new Date(leftSnapshot.createdAt).toLocaleString()}
                </div>
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <Hash className="h-3 w-3" />
                  {leftSnapshot.eventCount.toLocaleString()} events
                </div>
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <User className="h-3 w-3" />
                  {leftSnapshot.createdBy || 'System'}
                </div>
              </div>
            ) : (
              <p className="text-muted-foreground">Current State</p>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base flex items-center gap-2">
              <FileJson className="h-4 w-4" />
              To
            </CardTitle>
          </CardHeader>
          <CardContent>
            {rightSnapshot ? (
              <div className="space-y-2">
                <p className="font-medium">{rightSnapshot.name}</p>
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <Clock className="h-3 w-3" />
                  {new Date(rightSnapshot.createdAt).toLocaleString()}
                </div>
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <Hash className="h-3 w-3" />
                  {rightSnapshot.eventCount.toLocaleString()} events
                </div>
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <User className="h-3 w-3" />
                  {rightSnapshot.createdBy || 'System'}
                </div>
              </div>
            ) : (
              <p className="text-muted-foreground">Current State</p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Diff Results */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Changes</CardTitle>
        </CardHeader>
        <CardContent>
          {diffResult ? (
            <div className="space-y-2 max-h-[600px] overflow-y-auto">
              {filteredEntries.length === 0 ? (
                <div className="text-center py-12 text-muted-foreground">
                  <p>No changes detected</p>
                  {!showUnchanged && (
                    <p className="text-sm mt-2">
                      Enable &quot;Show unchanged&quot; to see all fields
                    </p>
                  )}
                </div>
              ) : (
                filteredEntries.map((entry) => (
                  <DiffEntryRow key={entry.key} entry={entry} viewMode={viewMode} />
                ))
              )}
            </div>
          ) : (
            <div className="text-center py-12 text-muted-foreground">
              <p>Select two versions to compare</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
