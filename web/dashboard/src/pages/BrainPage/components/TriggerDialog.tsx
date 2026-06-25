/**
 * Brain Page - Trigger Dialog
 * Dialog for creating Brain Triggers
 */

import { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { Loader2 } from 'lucide-react';
import { brainApi, type BrainTrigger } from '@/api/brain';
import { toast } from 'sonner';

interface TriggerDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreated?: (trigger: BrainTrigger) => void;
  availableConnectors?: string[];
}

const SIGNAL_TYPES = [
  { value: 'commit', label: 'Commit' },
  { value: 'issue', label: 'Issue' },
  { value: 'pr', label: 'Pull Request' },
  { value: 'message', label: 'Message' },
  { value: 'email', label: 'Email' },
  { value: 'document', label: 'Document' },
];

const ACTION_TYPES = [
  { value: 'run_agent', label: 'Run Agent' },
  { value: 'send_notification', label: 'Send Notification' },
  { value: 'webhook', label: 'Webhook' },
];

const SCHEDULE_PRESETS = [
  { value: 'immediate', label: 'Immediate' },
  { value: '0 8 * * *', label: 'Daily at 8 AM' },
  { value: '0 9 * * *', label: 'Daily at 9 AM' },
  { value: '0 */4 * * *', label: 'Every 4 hours' },
];

export function TriggerDialog({
  open,
  onOpenChange,
  onCreated,
  availableConnectors = [],
}: TriggerDialogProps) {
  const [loading, setLoading] = useState(false);
  const [form, setForm] = useState({
    name: '',
    signal_types: [] as string[],
    connector_slugs: [] as string[],
    min_importance: 2,
    schedule: 'immediate',
    action: 'run_agent',
    is_active: true,
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.name.trim()) {
      toast.error('Please enter a name for the trigger');
      return;
    }
    if (form.signal_types.length === 0) {
      toast.error('Please select at least one signal type');
      return;
    }

    setLoading(true);
    try {
      const trigger = await brainApi.createTrigger({
        name: form.name,
        signal_types: form.signal_types,
        connector_slugs: form.connector_slugs,
        min_importance: form.min_importance,
        schedule: form.schedule,
        action: form.action,
        is_active: form.is_active,
      });
      toast.success('Trigger created');
      onCreated?.(trigger);
      onOpenChange(false);
      // Reset form
      setForm({
        name: '',
        signal_types: [],
        connector_slugs: [],
        min_importance: 2,
        schedule: 'immediate',
        action: 'run_agent',
        is_active: true,
      });
    } catch (err) {
      console.error('Failed to create trigger:', err);
      toast.error('Failed to create trigger');
    } finally {
      setLoading(false);
    }
  };

  const toggleArrayItem = (
    arr: string[],
    item: string,
    setter: (v: string[]) => void
  ) => {
    if (arr.includes(item)) {
      setter(arr.filter((i) => i !== item));
    } else {
      setter([...arr, item]);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="bg-slate-900 border-white/10 max-w-lg">
        <DialogHeader>
          <DialogTitle className="text-text-primary">Create Brain Trigger</DialogTitle>
          <DialogDescription className="text-text-secondary">
            Run agents automatically when specific signal patterns are detected.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name" className="text-text-primary">
              Name
            </Label>
            <Input
              id="name"
              value={form.name}
              onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
              placeholder="High priority issue alert"
              className="bg-white/[0.03] border-white/[0.06]"
            />
          </div>

          <div className="space-y-2">
            <Label className="text-text-primary">Signal Types</Label>
            <div className="flex flex-wrap gap-2">
              {SIGNAL_TYPES.map((type) => (
                <button
                  key={type.value}
                  type="button"
                  onClick={() =>
                    toggleArrayItem(form.signal_types, type.value, (v) =>
                      setForm((f) => ({ ...f, signal_types: v }))
                    )
                  }
                  className={`px-3 py-1.5 rounded-full text-xs font-medium transition-colors ${
                    form.signal_types.includes(type.value)
                      ? 'bg-indigo-500/20 text-indigo-300 border border-indigo-500/30'
                      : 'bg-white/[0.05] text-text-secondary border border-white/[0.1] hover:border-white/[0.2]'
                  }`}
                >
                  {type.label}
                </button>
              ))}
            </div>
          </div>

          {availableConnectors.length > 0 && (
            <div className="space-y-2">
              <Label className="text-text-primary">Connectors (optional)</Label>
              <div className="flex flex-wrap gap-2">
                {availableConnectors.map((slug) => (
                  <button
                    key={slug}
                    type="button"
                    onClick={() =>
                      toggleArrayItem(form.connector_slugs, slug, (v) =>
                        setForm((f) => ({ ...f, connector_slugs: v }))
                      )
                    }
                    className={`px-3 py-1.5 rounded-full text-xs font-medium transition-colors ${
                      form.connector_slugs.includes(slug)
                        ? 'bg-indigo-500/20 text-indigo-300 border border-indigo-500/30'
                        : 'bg-white/[0.05] text-text-secondary border border-white/[0.1] hover:border-white/[0.2]'
                    }`}
                  >
                    {slug}
                  </button>
                ))}
              </div>
            </div>
          )}

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="min_importance" className="text-text-primary">
                Min Importance
              </Label>
              <Select
                value={String(form.min_importance)}
                onValueChange={(v) =>
                  setForm((f) => ({ ...f, min_importance: parseInt(v) }))
                }
              >
                <SelectTrigger className="bg-white/[0.03] border-white/[0.06]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {[1, 2, 3].map((i) => (
                    <SelectItem key={i} value={String(i)}>
                      P{i} - {i === 1 ? 'Low' : i === 2 ? 'Medium' : 'High'}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="schedule" className="text-text-primary">
                Schedule
              </Label>
              <Select
                value={form.schedule}
                onValueChange={(v) => setForm((f) => ({ ...f, schedule: v }))}
              >
                <SelectTrigger className="bg-white/[0.03] border-white/[0.06]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {SCHEDULE_PRESETS.map((preset) => (
                    <SelectItem key={preset.value} value={preset.value}>
                      {preset.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="action" className="text-text-primary">
              Action
            </Label>
            <Select
              value={form.action}
              onValueChange={(v) => setForm((f) => ({ ...f, action: v }))}
            >
              <SelectTrigger className="bg-white/[0.03] border-white/[0.06]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {ACTION_TYPES.map((action) => (
                  <SelectItem key={action.value} value={action.value}>
                    {action.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center justify-between">
            <div>
              <Label htmlFor="is_active" className="text-text-primary">
                Active
              </Label>
              <p className="text-xs text-text-secondary">Enable this trigger</p>
            </div>
            <Switch
              id="is_active"
              checked={form.is_active}
              onCheckedChange={(v) => setForm((f) => ({ ...f, is_active: v }))}
            />
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="ghost"
              onClick={() => onOpenChange(false)}
              className="border-white/10"
            >
              Cancel
            </Button>
            <Button type="submit" disabled={loading}>
              {loading && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
              Create Trigger
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
