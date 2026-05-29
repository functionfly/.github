/**
 * GraphSettingsPanel — visibility, execution mode, and publish controls.
 */

import { Globe, Loader2, Lock, Rocket, Users, Zap } from 'lucide-react';
import { useCallback, useState } from 'react';

import { toast } from '@/components/ui';
import { useFrgConfirmDialog } from '@/components/frg/FrgConfirmDialog';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { useFRGStore } from '@/stores/frgStore';
import type { ExecutionMode, GraphDefinition } from '@/types/frg';

const VISIBILITY_OPTIONS: Array<{
  value: GraphDefinition['visibility'];
  label: string;
  description: string;
  icon: typeof Lock;
}> = [
  { value: 'private', label: 'Private', description: 'Only your account can access', icon: Lock },
  { value: 'team', label: 'Team', description: 'Shared with your tenant members', icon: Users },
  {
    value: 'public',
    label: 'Public',
    description: 'Discoverable in the graph registry',
    icon: Globe,
  },
];

const EXECUTION_MODES: Array<{ value: ExecutionMode; label: string; description: string }> = [
  { value: 'sync', label: 'Sync', description: 'Run to completion and return output' },
  { value: 'async', label: 'Async', description: 'Start execution and poll for results' },
  { value: 'streaming', label: 'Streaming', description: 'Stream node output as it arrives' },
  {
    value: 'event_driven',
    label: 'Event driven',
    description: 'Triggered by webhooks or schedules',
  },
];

export function GraphSettingsPanel() {
  const store = useFRGStore();
  const {
    definition,
    graphAuthor,
    graphName,
    isDirty,
    isSaving,
    isLoading,
    setGraphSettings,
    saveGraph,
    publishGraph,
  } = store;

  const [isPublishing, setIsPublishing] = useState(false);
  const { requestConfirm, dialog: confirmDialog } = useFrgConfirmDialog();

  const isPublished = !!definition?.publishedAt;
  const isPersisted = !!definition?.id;
  const visibility = definition?.visibility ?? 'private';
  const executionMode = definition?.executionMode ?? 'sync';

  const handlePublish = useCallback(async () => {
    if (!isPersisted) {
      toast({ title: 'Save the graph before publishing', variant: 'destructive' });
      return;
    }
    if (isDirty) {
      toast({ title: 'Save your changes before publishing', variant: 'destructive' });
      return;
    }
    if (isPublished) return;

    const confirmed = await requestConfirm({
      title: 'Publish graph',
      description: `Publish "${graphName}"? Published graphs cannot be edited — create a remix to iterate.`,
      confirmLabel: 'Publish',
    });
    if (!confirmed) return;

    setIsPublishing(true);
    try {
      await publishGraph();
    } finally {
      setIsPublishing(false);
    }
  }, [isPersisted, isDirty, isPublished, graphName, publishGraph, requestConfirm]);

  const handleSaveSettings = useCallback(async () => {
    await saveGraph();
  }, [saveGraph]);

  return (
    <div className="flex h-full flex-col">
      <div className="border-b border-[var(--border-subtle)] px-4 py-3">
        <div className="flex items-center justify-between gap-2">
          <div>
            <h3 className="text-sm font-semibold text-[var(--text-primary)]">Graph Settings</h3>
            <p className="text-xs text-[var(--text-secondary)] mt-0.5">
              {graphAuthor && graphName ? `${graphAuthor}/${graphName}` : 'Unsaved draft'}
            </p>
          </div>
          <Badge variant={isPublished ? 'default' : 'secondary'} className="text-xs">
            {isPublished ? 'Published' : 'Draft'}
          </Badge>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-5">
        {isPublished && definition?.publishedAt && (
          <div className="rounded-lg border border-green-500/20 bg-green-500/5 p-3 text-xs text-[var(--text-secondary)]">
            Published {new Date(definition.publishedAt).toLocaleString()}. Structure is locked;
            remix to create an editable copy.
          </div>
        )}

        <div className="space-y-2">
          <Label className="text-xs text-[var(--text-secondary)]">Visibility</Label>
          <Select
            value={visibility}
            onValueChange={(value) =>
              setGraphSettings({ visibility: value as GraphDefinition['visibility'] })
            }
            disabled={isPublished}
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {VISIBILITY_OPTIONS.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  <div className="flex items-center gap-2">
                    <option.icon className="w-3.5 h-3.5" />
                    <span>{option.label}</span>
                  </div>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <p className="text-[11px] text-[var(--text-muted)]">
            {VISIBILITY_OPTIONS.find((o) => o.value === visibility)?.description}
          </p>
        </div>

        <div className="space-y-2">
          <Label className="text-xs text-[var(--text-secondary)]">Execution mode</Label>
          <Select
            value={executionMode}
            onValueChange={(value) => setGraphSettings({ executionMode: value as ExecutionMode })}
            disabled={isPublished}
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {EXECUTION_MODES.map((mode) => (
                <SelectItem key={mode.value} value={mode.value}>
                  {mode.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <p className="text-[11px] text-[var(--text-muted)]">
            {EXECUTION_MODES.find((m) => m.value === executionMode)?.description}
          </p>
        </div>

        {!isPublished && isDirty && (
          <Button
            variant="outline"
            size="sm"
            className="w-full"
            onClick={handleSaveSettings}
            disabled={isSaving || isLoading}
          >
            {isSaving ? (
              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            ) : (
              <Zap className="w-4 h-4 mr-2" />
            )}
            Save settings
          </Button>
        )}
      </div>

      <div className="border-t border-[var(--border-subtle)] p-4 space-y-3">
        <Separator />
        <Button
          className="w-full bg-gradient-to-r from-brand-500 to-purple-500"
          onClick={handlePublish}
          disabled={isPublished || !isPersisted || isDirty || isPublishing || isSaving}
        >
          {isPublishing ? (
            <Loader2 className="w-4 h-4 mr-2 animate-spin" />
          ) : (
            <Rocket className="w-4 h-4 mr-2" />
          )}
          {isPublished ? 'Already published' : 'Publish graph'}
        </Button>
        <p className="text-[11px] text-[var(--text-muted)] text-center">
          Publishing locks the graph and enables execution endpoints and webhooks.
        </p>
      </div>
      {confirmDialog}
    </div>
  );
}

export default GraphSettingsPanel;
