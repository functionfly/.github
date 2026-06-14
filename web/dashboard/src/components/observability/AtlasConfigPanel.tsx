'use client';

import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Slider } from '@/components/ui/slider';
import { Switch } from '@/components/ui/switch';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Settings, Save, RotateCcw } from 'lucide-react';

interface AtlasConfigPanelProps {
  config: {
    sampling_rate: number;
    trace_errors_only: boolean;
    sample_head_percent: number;
    sample_tail_count: number;
    retention_days: number;
  } | null;
  onUpdate: (updates: any) => Promise<any>;
  loading?: boolean;
}

export default function AtlasConfigPanel({ config, onUpdate, loading }: AtlasConfigPanelProps) {
  const { t } = useTranslation();
  const [localConfig, setLocalConfig] = useState(config);
  const [saving, setSaving] = useState(false);
  const [showPanel, setShowPanel] = useState(false);

  if (config && !localConfig) {
    setLocalConfig(config);
  }

  const handleSave = async () => {
    if (!localConfig) return;
    setSaving(true);
    try {
      await onUpdate(localConfig);
    } finally {
      setSaving(false);
    }
  };

  const handleReset = () => {
    setLocalConfig(config);
  };

  if (!showPanel) {
    return (
      <Button variant="outline" size="sm" onClick={() => setShowPanel(true)} className="gap-2">
        <Settings className="h-4 w-4" />
        Configure Atlas
      </Button>
    );
  }

  return (
    <Card className="w-full">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-lg flex items-center gap-2">
              <Settings className="h-4 w-4" />
              Atlas Configuration
            </CardTitle>
            <CardDescription>Configure observability sampling and retention settings</CardDescription>
          </div>
          <Button variant="ghost" size="sm" onClick={() => setShowPanel(false)}>
            Close
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        {loading ? (
          <div className="text-center py-4 text-muted-foreground">Loading configuration...</div>
        ) : localConfig ? (
          <>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <label className="text-sm font-medium">Sampling Rate</label>
                  <p className="text-xs text-muted-foreground">Percentage of runs to sample (0-100)</p>
                </div>
                <div className="flex items-center gap-3">
                  <Slider
                    value={[localConfig.sampling_rate * 100]}
                    onValueChange={([v]) => setLocalConfig({ ...localConfig, sampling_rate: v / 100 })}
                    min={0}
                    max={100}
                    step={1}
                    className="w-[150px]"
                  />
                  <Badge variant="outline" className="w-16 justify-center">
                    {Math.round(localConfig.sampling_rate * 100)}%
                  </Badge>
                </div>
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <label className="text-sm font-medium">Trace Errors Only</label>
                  <p className="text-xs text-muted-foreground">Only trace runs with errors</p>
                </div>
                <Switch
                  checked={localConfig.trace_errors_only}
                  onCheckedChange={(checked) => setLocalConfig({ ...localConfig, trace_errors_only: checked })}
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">Sample Head Percent</label>
                <p className="text-xs text-muted-foreground">Sample first N% of tokens per run</p>
                <div className="flex items-center gap-3">
                  <Slider
                    value={[localConfig.sample_head_percent]}
                    onValueChange={([v]) => setLocalConfig({ ...localConfig, sample_head_percent: v })}
                    min={0}
                    max={100}
                    step={1}
                    className="w-[150px]"
                  />
                  <Badge variant="outline" className="w-16 justify-center">
                    {localConfig.sample_head_percent}%
                  </Badge>
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">Sample Tail Count</label>
                <p className="text-xs text-muted-foreground">Number of final events to always capture</p>
                <Input
                  type="number"
                  value={localConfig.sample_tail_count}
                  onChange={(e) => setLocalConfig({ ...localConfig, sample_tail_count: parseInt(e.target.value) || 0 })}
                  className="w-32"
                  min={0}
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">Retention Days</label>
                <p className="text-xs text-muted-foreground">How long to keep observability data</p>
                <div className="flex items-center gap-2">
                  <Input
                    type="number"
                    value={localConfig.retention_days}
                    onChange={(e) => setLocalConfig({ ...localConfig, retention_days: parseInt(e.target.value) || 0 })}
                    className="w-32"
                    min={1}
                  />
                  <span className="text-sm text-muted-foreground">days</span>
                </div>
              </div>
            </div>

            <div className="flex items-center gap-3 pt-4 border-t">
              <Button onClick={handleSave} disabled={saving} className="gap-2">
                {saving ? (
                  <>Saving...</>
                ) : (
                  <>
                    <Save className="h-4 w-4" />
                    Save Changes
                  </>
                )}
              </Button>
              <Button variant="outline" onClick={handleReset} className="gap-2">
                <RotateCcw className="h-4 w-4" />
                Reset
              </Button>
            </div>
          </>
        ) : (
          <div className="text-center py-4 text-muted-foreground">No configuration available</div>
        )}
      </CardContent>
    </Card>
  );
}
