'use client';

import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import { RefreshCw } from 'lucide-react';

interface AutoRefreshToggleProps {
  enabled: boolean;
  onChange: (enabled: boolean) => void;
  interval?: number;
  onIntervalChange?: (interval: number) => void;
}

const INTERVAL_OPTIONS = [
  { label: '5s', value: 5000 },
  { label: '10s', value: 10000 },
  { label: '30s', value: 30000 },
  { label: '1m', value: 60000 },
];

export default function AutoRefreshToggle({
  enabled,
  onChange,
  interval = 10000,
  onIntervalChange,
}: AutoRefreshToggleProps) {
  return (
    <div className="flex items-center gap-4">
      <div className="flex items-center gap-2">
        <Switch id="auto-refresh" checked={enabled} onCheckedChange={onChange} />
        <Label htmlFor="auto-refresh" className="text-sm cursor-pointer">
          Auto-refresh
        </Label>
      </div>

      {enabled && onIntervalChange && (
        <div className="flex items-center gap-2">
          <RefreshCw className="h-3 w-3 text-muted-foreground" />
          <select
            value={interval}
            onChange={(e) => onIntervalChange(Number(e.target.value))}
            className="text-xs border rounded px-2 py-1 bg-background"
          >
            {INTERVAL_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>
      )}
    </div>
  );
}
