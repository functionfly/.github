import React, { useState } from 'react';
import { Slider } from '@/components/ui/slider';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { RuntimeSelector, RuntimeType } from './RuntimeSelector';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  AlertCircle,
  CheckCircle2,
  Clock,
  Database,
  Globe,
  Lock,
  RefreshCw,
  Save,
  Settings,
  Shield,
  Zap,
} from 'lucide-react';

interface RuntimeSettings {
  runtime: RuntimeType;
  memoryMb: number;
  timeoutMs: number;
  networkEnabled: boolean;
  cachingEnabled: boolean;
  maxConcurrent: number;
}

interface RuntimeSettingsPanelProps {
  settings: RuntimeSettings;
  onChange: (settings: RuntimeSettings) => void;
  onSave?: () => void;
  isLoading?: boolean;
}

const MEMORY_OPTIONS = [64, 128, 256, 512, 1024, 2048];
const TIMEOUT_OPTIONS = [1000, 5000, 10000, 30000, 60000, 120000];

export function RuntimeSettingsPanel({
  settings,
  onChange,
  onSave,
  isLoading = false,
}: RuntimeSettingsPanelProps) {
  const [localSettings, setLocalSettings] = useState<RuntimeSettings>(settings);

  const updateSetting = <K extends keyof RuntimeSettings>(
    key: K,
    value: RuntimeSettings[K]
  ) => {
    const newSettings = { ...localSettings, [key]: value };
    setLocalSettings(newSettings);
    onChange(newSettings);
  };

  const handleSave = () => {
    onChange(localSettings);
    onSave?.();
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">Runtime Settings</h2>
          <p className="text-muted-foreground">
            Configure execution environment for your function
          </p>
        </div>
        {onSave && (
          <Button onClick={handleSave} disabled={isLoading}>
            {isLoading ? (
              <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Save className="mr-2 h-4 w-4" />
            )}
            Save Settings
          </Button>
        )}
      </div>

      <Tabs defaultValue="basic" className="space-y-4">
        <TabsList>
          <TabsTrigger value="basic">Basic</TabsTrigger>
          <TabsTrigger value="resources">Resources</TabsTrigger>
          <TabsTrigger value="advanced">Advanced</TabsTrigger>
        </TabsList>

        {/* Basic Settings */}
        <TabsContent value="basic" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Settings className="h-5 w-5" />
                Basic Configuration
              </CardTitle>
              <CardDescription>
                Essential runtime settings
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <RuntimeSelector
                value={localSettings.runtime}
                onChange={(runtime) => updateSetting('runtime', runtime)}
                showDescription={true}
                size="lg"
              />

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="timeout">Timeout (ms)</Label>
                  <Select
                    value={localSettings.timeoutMs.toString()}
                    onValueChange={(v) => updateSetting('timeoutMs', parseInt(v))}
                  >
                    <SelectTrigger id="timeout">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {TIMEOUT_OPTIONS.map((ms) => (
                        <SelectItem key={ms} value={ms.toString()}>
                          {ms < 1000 ? `${ms}ms` : `${ms / 1000}s`}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <p className="text-xs text-muted-foreground">
                    Maximum execution time before timeout
                  </p>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="memory">Memory (MB)</Label>
                  <Select
                    value={localSettings.memoryMb.toString()}
                    onValueChange={(v) => updateSetting('memoryMb', parseInt(v))}
                  >
                    <SelectTrigger id="memory">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {MEMORY_OPTIONS.map((mb) => (
                        <SelectItem key={mb} value={mb.toString()}>
                          {mb} MB
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <p className="text-xs text-muted-foreground">
                    Maximum memory allocation
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Resources Settings */}
        <TabsContent value="resources" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Zap className="h-5 w-5" />
                Resource Limits
              </CardTitle>
              <CardDescription>
                Control resource consumption
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div className="space-y-1">
                    <Label>Memory Allocation</Label>
                    <p className="text-sm text-muted-foreground">
                      {localSettings.memoryMb} MB allocated
                    </p>
                  </div>
                  <Badge variant="outline">
                    <Database className="mr-1 h-3 w-3" />
                    {localSettings.memoryMb} MB
                  </Badge>
                </div>
                <Slider
                  value={[localSettings.memoryMb]}
                  min={64}
                  max={2048}
                  step={64}
                  onValueChange={([v]) => updateSetting('memoryMb', v)}
                  className="py-4"
                />
                <div className="flex justify-between text-xs text-muted-foreground">
                  <span>64 MB</span>
                  <span>2 GB</span>
                </div>
              </div>

              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div className="space-y-1">
                    <Label>Execution Timeout</Label>
                    <p className="text-sm text-muted-foreground">
                      {localSettings.timeoutMs < 1000
                        ? `${localSettings.timeoutMs}ms`
                        : `${localSettings.timeoutMs / 1000}s`}
                    </p>
                  </div>
                  <Badge variant="outline">
                    <Clock className="mr-1 h-3 w-3" />
                    {localSettings.timeoutMs < 1000
                      ? `${localSettings.timeoutMs}ms`
                      : `${localSettings.timeoutMs / 1000}s`}
                  </Badge>
                </div>
                <Slider
                  value={[localSettings.timeoutMs]}
                  min={1000}
                  max={120000}
                  step={1000}
                  onValueChange={([v]) => updateSetting('timeoutMs', v)}
                  className="py-4"
                />
                <div className="flex justify-between text-xs text-muted-foreground">
                  <span>1s</span>
                  <span>2min</span>
                </div>
              </div>

              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div className="space-y-1">
                    <Label>Max Concurrent Executions</Label>
                    <p className="text-sm text-muted-foreground">
                      Functions can run up to {localSettings.maxConcurrent} times simultaneously
                    </p>
                  </div>
                  <Badge variant="outline">
                    {localSettings.maxConcurrent}x
                  </Badge>
                </div>
                <Slider
                  value={[localSettings.maxConcurrent]}
                  min={1}
                  max={100}
                  step={1}
                  onValueChange={([v]) => updateSetting('maxConcurrent', v)}
                  className="py-4"
                />
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Advanced Settings */}
        <TabsContent value="advanced" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Shield className="h-5 w-5" />
                Security & Network
              </CardTitle>
              <CardDescription>
                Configure security and network access
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="flex items-center justify-between">
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <Globe className="h-4 w-4" />
                    <Label>Network Access</Label>
                  </div>
                  <p className="text-sm text-muted-foreground">
                    Allow function to make external HTTP requests
                  </p>
                </div>
                <Switch
                  checked={localSettings.networkEnabled}
                  onCheckedChange={(v) => updateSetting('networkEnabled', v)}
                />
              </div>

              {localSettings.networkEnabled && (
                <Alert>
                  <Globe className="h-4 w-4" />
                  <AlertTitle>Network Access Enabled</AlertTitle>
                  <AlertDescription>
                    Your function can make external HTTP requests. This may incur additional costs
                    and should be used carefully.
                  </AlertDescription>
                </Alert>
              )}

              <div className="flex items-center justify-between">
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <Database className="h-4 w-4" />
                    <Label>Code Caching</Label>
                  </div>
                  <p className="text-sm text-muted-foreground">
                    Cache compiled code for faster cold starts
                  </p>
                </div>
                <Switch
                  checked={localSettings.cachingEnabled}
                  onCheckedChange={(v) => updateSetting('cachingEnabled', v)}
                />
              </div>

              {!localSettings.cachingEnabled && (
                <Alert variant="warning">
                  <AlertCircle className="h-4 w-4" />
                  <AlertTitle>Caching Disabled</AlertTitle>
                  <AlertDescription>
                    Disabling caching may result in slower cold starts.
                  </AlertDescription>
                </Alert>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Environment Variables</CardTitle>
              <CardDescription>
                Configure environment variables available to your function
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="rounded-md border p-4 text-center text-sm text-muted-foreground">
                <p>Environment variable configuration coming soon</p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Summary */}
      <Card>
        <CardHeader>
          <CardTitle>Configuration Summary</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="space-y-1">
              <p className="text-xs text-muted-foreground">Runtime</p>
              <p className="font-medium">{localSettings.runtime}</p>
            </div>
            <div className="space-y-1">
              <p className="text-xs text-muted-foreground">Memory</p>
              <p className="font-medium">{localSettings.memoryMb} MB</p>
            </div>
            <div className="space-y-1">
              <p className="text-xs text-muted-foreground">Timeout</p>
              <p className="font-medium">
                {localSettings.timeoutMs < 1000
                  ? `${localSettings.timeoutMs}ms`
                  : `${localSettings.timeoutMs / 1000}s`}
              </p>
            </div>
            <div className="space-y-1">
              <p className="text-xs text-muted-foreground">Network</p>
              <p className="font-medium">
                {localSettings.networkEnabled ? 'Enabled' : 'Disabled'}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

export default RuntimeSettingsPanel;
