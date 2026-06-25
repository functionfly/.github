/**
 * Brain Page - Composer Dialog
 * Dialog for creating/editing Brain Composers
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
import { brainApi, type BrainComposer, type SignalFilter } from '@/api/brain';
import { toast } from 'sonner';

interface ComposerDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreated?: (composer: BrainComposer) => void;
  initialData?: Partial<BrainComposer>;
}

const SCHEDULE_PRESETS = [
  { value: '0 8 * * *', label: 'Daily at 8 AM' },
  { value: '0 9 * * *', label: 'Daily at 9 AM' },
  { value: '0 18 * * *', label: 'Daily at 6 PM' },
  { value: '0 8 * * 1-5', label: 'Weekdays at 8 AM' },
  { value: '0 */4 * * *', label: 'Every 4 hours' },
];

const OUTPUT_FORMATS = [
  { value: 'briefing', label: 'Briefing' },
  { value: 'digest', label: 'Digest' },
  { value: 'summary', label: 'Summary' },
];

export function ComposerDialog({
  open,
  onOpenChange,
  onCreated,
  initialData,
}: ComposerDialogProps) {
  const [loading, setLoading] = useState(false);
  const [form, setForm] = useState({
    name: initialData?.name || '',
    schedule: initialData?.schedule || '0 8 * * *',
    output_format: initialData?.output_format || 'briefing',
    is_active: initialData?.is_active ?? true,
    signal_filters: initialData?.signal_filters || [
      {
        connector_slugs: [],
        signal_types: [],
        importance_min: 1,
        time_window: '24h',
      },
    ] as SignalFilter[],
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.name.trim()) {
      toast.error('Please enter a name for the composer');
      return;
    }

    setLoading(true);
    try {
      const composer = await brainApi.createComposer({
        name: form.name,
        schedule: form.schedule,
        output_format: form.output_format,
        is_active: form.is_active,
        signal_filters: form.signal_filters,
        actions: [],
      });
      toast.success('Composer created');
      onCreated?.(composer);
      onOpenChange(false);
    } catch (err) {
      console.error('Failed to create composer:', err);
      toast.error('Failed to create composer');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="bg-slate-900 border-white/10">
        <DialogHeader>
          <DialogTitle className="text-text-primary">Create Brain Composer</DialogTitle>
          <DialogDescription className="text-text-secondary">
            Set up automated daily briefings from your connected accounts.
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
              placeholder="Daily briefing"
              className="bg-white/[0.03] border-white/[0.06]"
            />
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

          <div className="space-y-2">
            <Label htmlFor="output_format" className="text-text-primary">
              Output Format
            </Label>
            <Select
              value={form.output_format}
              onValueChange={(v) => setForm((f) => ({ ...f, output_format: v }))}
            >
              <SelectTrigger className="bg-white/[0.03] border-white/[0.06]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {OUTPUT_FORMATS.map((format) => (
                  <SelectItem key={format.value} value={format.value}>
                    {format.label}
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
              <p className="text-xs text-text-secondary">
                Run this composer on schedule
              </p>
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
              Create Composer
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
