import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { Switch } from '@/components/ui/switch';
import { Clock, Globe, Webhook } from 'lucide-react';
import { useMemo } from 'react';
import { FieldError, InfoTip, SectionCard } from '../components/editor-ui';
import { HTTP_METHODS } from '../constants';
import type { HttpMethod } from '../types';
import type { FunctionEditorModel } from '../useFunctionEditor';

type Props = { editor: FunctionEditorModel };

const COMMON_TIMEZONES = [
  'UTC',
  'America/New_York',
  'America/Chicago',
  'America/Denver',
  'America/Los_Angeles',
  'Europe/London',
  'Europe/Paris',
  'Europe/Berlin',
  'Asia/Tokyo',
  'Asia/Shanghai',
  'Asia/Kolkata',
  'Australia/Sydney',
];

function parseCronHuman(cron: string): string {
  const parts = cron.trim().split(/\s+/);
  if (parts.length !== 5) return 'Invalid cron expression';
  const [min, hour, dom, month, dow] = parts;

  if (min === '*' && hour === '*' && dom === '*' && month === '*' && dow === '*') {
    return 'Every minute';
  }
  if (min !== '*' && hour === '*' && dom === '*' && month === '*' && dow === '*') {
    return `Every hour at minute ${min}`;
  }
  if (min !== '*' && hour !== '*' && dom === '*' && month === '*' && dow === '*') {
    const h = parseInt(hour, 10);
    const m = parseInt(min, 10);
    if (!isNaN(h) && !isNaN(m)) {
      const ampm = h >= 12 ? 'PM' : 'AM';
      const h12 = h % 12 || 12;
      return `Daily at ${h12}:${String(m).padStart(2, '0')} ${ampm} UTC`;
    }
  }
  if (min === '0' && hour === '0' && dom === '*' && month === '*' && dow !== '*') {
    const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
    const dayNum = parseInt(dow, 10);
    if (!isNaN(dayNum) && dayNum >= 0 && dayNum <= 6) {
      return `Every ${days[dayNum]} at midnight UTC`;
    }
  }
  if (min === '0' && hour === '0' && dom === '1' && month === '*' && dow === '*') {
    return 'First day of every month at midnight UTC';
  }
  return `Custom: ${cron}`;
}

export function TriggersSection({ editor }: Props) {
  const { httpTrigger, setHttpTrigger, scheduleTrigger, setScheduleTrigger, errors, markDirty } =
    editor;

  const cronHuman = useMemo(
    () => (scheduleTrigger.enabled ? parseCronHuman(scheduleTrigger.cron) : ''),
    [scheduleTrigger.enabled, scheduleTrigger.cron]
  );

  return (
    <SectionCard
      icon={<Webhook className="w-4 h-4" />}
      title="Triggers"
      step={6}
      description="Define how your function is invoked"
    >
      {/* HTTP Trigger */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Globe className={`w-4 h-4 ${httpTrigger.enabled ? 'text-[#FF6B35]' : 'text-text-muted'}`} />
            <span className="text-sm font-medium text-text-primary">HTTP Trigger</span>
          </div>
          <Switch
            checked={httpTrigger.enabled}
            onCheckedChange={(c) => {
              setHttpTrigger((t) => ({ ...t, enabled: c }));
              markDirty();
            }}
            aria-label="Enable HTTP trigger"
          />
        </div>
        {httpTrigger.enabled && (
          <div className="grid grid-cols-[120px_1fr] gap-3 pl-6">
            <div>
              <Label className="text-xs text-text-secondary mb-1 block">Method</Label>
              <Select
                value={httpTrigger.method}
                onValueChange={(v) => {
                  setHttpTrigger((t) => ({ ...t, method: v as HttpMethod }));
                  markDirty();
                }}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {HTTP_METHODS.map((m) => (
                    <SelectItem key={m} value={m}>
                      {m}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label htmlFor="http-path" className="text-xs text-text-secondary mb-1 block">
                Path
              </Label>
              <Input
                id="http-path"
                placeholder="/api/my-function"
                value={httpTrigger.path}
                onChange={(e) => {
                  setHttpTrigger((t) => ({ ...t, path: e.target.value }));
                  markDirty();
                }}
                className="input font-mono text-sm"
              />
              <FieldError message={errors.httpPath} />
            </div>
          </div>
        )}
      </div>

      <Separator className="opacity-30" />

      {/* Schedule Trigger */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Clock className={`w-4 h-4 ${scheduleTrigger.enabled ? 'text-[#FF6B35]' : 'text-text-muted'}`} />
            <span className="text-sm font-medium text-text-primary">Schedule (Cron)</span>
          </div>
          <Switch
            checked={scheduleTrigger.enabled}
            onCheckedChange={(c) => {
              setScheduleTrigger((t) => ({ ...t, enabled: c }));
              markDirty();
            }}
            aria-label="Enable schedule trigger"
          />
        </div>
        {scheduleTrigger.enabled && (
          <div className="space-y-3 pl-6">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <Label htmlFor="cron" className="text-xs text-text-secondary mb-1 block">
                  Cron Expression
                  <InfoTip content="Standard 5-field cron: minute hour day month weekday" />
                </Label>
                <Input
                  id="cron"
                  placeholder="0 * * * *"
                  value={scheduleTrigger.cron}
                  onChange={(e) => {
                    setScheduleTrigger((t) => ({ ...t, cron: e.target.value }));
                    markDirty();
                  }}
                  className="input font-mono text-sm"
                />
              </div>
              <div>
                <Label htmlFor="tz" className="text-xs text-text-secondary mb-1 block">
                  Timezone
                </Label>
                <Select
                  value={scheduleTrigger.timezone}
                  onValueChange={(v) => {
                    setScheduleTrigger((t) => ({ ...t, timezone: v }));
                    markDirty();
                  }}
                >
                  <SelectTrigger id="tz">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {COMMON_TIMEZONES.map((tz) => (
                      <SelectItem key={tz} value={tz}>
                        {tz}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
            {cronHuman && (
              <p className="text-xs text-[#FF6B35] flex items-center gap-1.5">
                <Clock className="w-3 h-3 shrink-0" />
                {cronHuman}
              </p>
            )}
          </div>
        )}
      </div>
    </SectionCard>
  );
}
