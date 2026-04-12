/**
 * NodeInspector Panel (Right Sidebar)
 * Shows: Input fields, Config options, Runtime settings, Environment variables
 */

import { useCallback, useState } from 'react';
import { 
  Settings, 
  Play, 
  Save, 
  RotateCcw, 
  Trash2, 
  Terminal,
  Database,
  Globe,
  Lock,
  AlertCircle,
  CheckCircle,
  ChevronDown,
  ChevronRight,
  Code,
  FileJson,
} from 'lucide-react';

import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { Switch } from '@/components/ui/switch';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

import { useFRGStore, selectSelectedNode } from '@/stores/frgStore';

interface ConfigField {
  key: string;
  type: 'string' | 'number' | 'boolean' | 'json' | 'secret';
  value: unknown;
  description?: string;
  required?: boolean;
  defaultValue?: unknown;
}

export function NodeInspector() {
  const store = useFRGStore();
  const selectedNode = selectSelectedNode(store);
  const { removeNode, updateNode } = store;

  const [activeTab, setActiveTab] = useState('config');
  const [configValues, setConfigValues] = useState<Record<string, unknown>>({});
  const [envVars, setEnvVars] = useState<Record<string, string>>({});

  if (!selectedNode) {
    return (
      <div className="flex flex-col items-center justify-center h-full p-6 text-center">
        <div className="w-16 h-16 rounded-full bg-[var(--bg-tertiary)] flex items-center justify-center mb-4">
          <Settings className="w-8 h-8 text-[var(--text-muted)]" />
        </div>
        <h3 className="text-sm font-medium text-[var(--text-secondary)] mb-1">
          No Node Selected
        </h3>
        <p className="text-xs text-[var(--text-muted)]">
          Click on a node in the canvas to inspect and configure it
        </p>
      </div>
    );
  }

  const { functionRef, runtimeState } = selectedNode.data;

  const handleConfigChange = useCallback((key: string, value: unknown) => {
    setConfigValues(prev => ({ ...prev, [key]: value }));
  }, []);

  const handleEnvVarChange = useCallback((key: string, value: string) => {
    setEnvVars(prev => ({ ...prev, [key]: value }));
  }, []);

  const handleSave = useCallback(() => {
    updateNode(selectedNode.id, {
      data: {
        ...selectedNode.data,
        functionRef: {
          ...functionRef,
          config: configValues,
          metadata: {
            ...functionRef.metadata,
            envVars,
          },
        },
      },
    });
  }, [selectedNode, configValues, envVars, updateNode, functionRef]);

  const handleDelete = useCallback(() => {
    removeNode(selectedNode.id);
  }, [removeNode, selectedNode.id]);

  const handleReset = useCallback(() => {
    setConfigValues(functionRef.config || {});
    setEnvVars((functionRef.metadata?.envVars as Record<string, string>) || {});
  }, [functionRef]);

  const handleRun = useCallback(() => {
    // Run this node
  }, []);

  // Mock config fields based on function
  const configFields: ConfigField[] = [
    { key: 'timeout', type: 'number', value: 30000, description: 'Execution timeout in milliseconds', required: true },
    { key: 'retryAttempts', type: 'number', value: 3, description: 'Number of retry attempts on failure', defaultValue: 3 },
    { key: 'cacheEnabled', type: 'boolean', value: false, description: 'Enable result caching', defaultValue: false },
    { key: 'apiKey', type: 'secret', value: '', description: 'API key for external service', required: false },
  ];

  const renderConfigField = (field: ConfigField) => {
    const value = configValues[field.key] ?? field.value;

    switch (field.type) {
      case 'boolean':
        return (
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label className="text-xs">{field.key}</Label>
              {field.description && (
                <p className="text-[10px] text-[var(--text-muted)]">{field.description}</p>
              )}
            </div>
            <Switch
              checked={value as boolean}
              onCheckedChange={(checked) => handleConfigChange(field.key, checked)}
            />
          </div>
        );
      case 'number':
        return (
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label className="text-xs">{field.key}</Label>
              {field.required && <span className="text-[10px] text-red-500">Required</span>}
            </div>
            {field.description && (
              <p className="text-[10px] text-[var(--text-muted)]">{field.description}</p>
            )}
            <Input
              type="number"
              value={value as number}
              onChange={(e) => handleConfigChange(field.key, parseInt(e.target.value))}
              className="h-8"
            />
          </div>
        );
      case 'secret':
        return (
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label className="text-xs flex items-center gap-1">
                <Lock className="w-3 h-3" />
                {field.key}
              </Label>
              {field.required && <span className="text-[10px] text-red-500">Required</span>}
            </div>
            {field.description && (
              <p className="text-[10px] text-[var(--text-muted)]">{field.description}</p>
            )}
            <Input
              type="password"
              value={value as string}
              onChange={(e) => handleConfigChange(field.key, e.target.value)}
              className="h-8 font-mono"
              placeholder="••••••••"
            />
          </div>
        );
      case 'json':
        return (
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label className="text-xs">{field.key}</Label>
              {field.required && <span className="text-[10px] text-red-500">Required</span>}
            </div>
            <Textarea
              value={JSON.stringify(value, null, 2)}
              onChange={(e) => handleConfigChange(field.key, e.target.value)}
              className="min-h-[100px] font-mono text-xs"
            />
          </div>
        );
      default:
        return (
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label className="text-xs">{field.key}</Label>
              {field.required && <span className="text-[10px] text-red-500">Required</span>}
            </div>
            {field.description && (
              <p className="text-[10px] text-[var(--text-muted)]">{field.description}</p>
            )}
            <Input
              value={value as string}
              onChange={(e) => handleConfigChange(field.key, e.target.value)}
              className="h-8"
            />
          </div>
        );
    }
  };

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-4 border-b border-[var(--border-subtle)]">
        <div className="flex items-start justify-between">
          <div>
            <h3 className="font-semibold text-[var(--text-primary)]">
              {functionRef.name}
            </h3>
            <p className="text-xs text-[var(--text-secondary)]">
              {functionRef.author}/{functionRef.version}
            </p>
          </div>
          <div className="flex gap-1">
            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={handleRun}>
              <Play className="w-4 h-4" />
            </Button>
            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={handleSave}>
              <Save className="w-4 h-4" />
            </Button>
            <Button variant="ghost" size="icon" className="h-8 w-8 text-red-500" onClick={handleDelete}>
              <Trash2 className="w-4 h-4" />
            </Button>
          </div>
        </div>

        {/* Status */}
        {runtimeState && (
          <div className="flex items-center gap-2 mt-3">
            <Badge 
              variant={runtimeState.status === 'failed' ? 'destructive' : runtimeState.status === 'completed' ? 'default' : 'secondary'}
              className="text-[10px]"
            >
              {runtimeState.status === 'completed' && <CheckCircle className="w-3 h-3 mr-1" />}
              {runtimeState.status === 'failed' && <AlertCircle className="w-3 h-3 mr-1" />}
              {runtimeState.status}
            </Badge>
            {runtimeState.durationMs > 0 && (
              <span className="text-xs text-[var(--text-secondary)]">
                {runtimeState.durationMs}ms
              </span>
            )}
          </div>
        )}
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col min-h-0">
        <TabsList className="w-full rounded-none border-b border-[var(--border-subtle)] bg-transparent p-0 h-10">
          <TabsTrigger 
            value="config" 
            className="flex-1 rounded-none data-[state=active]:bg-transparent data-[state=active]:border-b-2 data-[state=active]:border-brand-500"
          >
            <Settings className="w-3 h-3 mr-1" />
            Config
          </TabsTrigger>
          <TabsTrigger 
            value="inputs"
            className="flex-1 rounded-none data-[state=active]:bg-transparent data-[state=active]:border-b-2 data-[state=active]:border-brand-500"
          >
            <FileJson className="w-3 h-3 mr-1" />
            Inputs
          </TabsTrigger>
          <TabsTrigger 
            value="runtime"
            className="flex-1 rounded-none data-[state=active]:bg-transparent data-[state=active]:border-b-2 data-[state=active]:border-brand-500"
          >
            <Terminal className="w-3 h-3 mr-1" />
            Runtime
          </TabsTrigger>
        </TabsList>

        <ScrollArea className="flex-1">
          <TabsContent value="config" className="m-0 p-4 space-y-4">
            <Accordion type="multiple" defaultValue={['settings']} className="space-y-2">
              <AccordionItem value="settings" className="border-0">
                <AccordionTrigger className="text-xs py-2 hover:no-underline">
                  <div className="flex items-center gap-2">
                    <Settings className="w-4 h-4" />
                    Settings
                  </div>
                </AccordionTrigger>
                <AccordionContent className="space-y-3 pt-2">
                  {configFields.map((field) => (
                    <div key={field.key}>
                      {renderConfigField(field)}
                    </div>
                  ))}
                </AccordionContent>
              </AccordionItem>

              <AccordionItem value="environment" className="border-0">
                <AccordionTrigger className="text-xs py-2 hover:no-underline">
                  <div className="flex items-center gap-2">
                    <Globe className="w-4 h-4" />
                    Environment Variables
                  </div>
                </AccordionTrigger>
                <AccordionContent className="space-y-3 pt-2">
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <Label className="text-xs">NODE_ENV</Label>
                    </div>
                    <Select 
                      value={envVars.NODE_ENV || 'development'}
                      onValueChange={(v) => handleEnvVarChange('NODE_ENV', v)}
                    >
                      <SelectTrigger className="h-8">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="development">development</SelectItem>
                        <SelectItem value="staging">staging</SelectItem>
                        <SelectItem value="production">production</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </AccordionContent>
              </AccordionItem>
            </Accordion>

            <div className="flex gap-2 pt-2">
              <Button variant="outline" size="sm" className="flex-1" onClick={handleReset}>
                <RotateCcw className="w-4 h-4 mr-1" />
                Reset
              </Button>
              <Button size="sm" className="flex-1" onClick={handleSave}>
                <Save className="w-4 h-4 mr-1" />
                Save
              </Button>
            </div>
          </TabsContent>

          <TabsContent value="inputs" className="m-0 p-4 space-y-4">
            <div className="space-y-2">
              <Label className="text-xs">Input Schema</Label>
              <div className="bg-[var(--bg-tertiary)] rounded-lg p-3">
                <pre className="text-xs font-mono text-[var(--text-secondary)]">
                  {JSON.stringify({
                    type: 'object',
                    properties: {
                      data: { type: 'array', description: 'Input data array' },
                      options: { type: 'object', description: 'Processing options' },
                    }
                  }, null, 2)}
                </pre>
              </div>
            </div>

            <div className="space-y-2">
              <Label className="text-xs">Test Input</Label>
              <Textarea
                placeholder="Enter JSON test input..."
                className="min-h-[150px] font-mono text-xs"
                defaultValue={JSON.stringify({ data: [] }, null, 2)}
              />
            </div>
          </TabsContent>

          <TabsContent value="runtime" className="m-0 p-4 space-y-4">
            <div className="space-y-3">
              <div className="flex items-center justify-between text-sm">
                <span className="text-[var(--text-secondary)]">Execution Mode</span>
                <Badge variant="secondary">Sync</Badge>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-[var(--text-secondary)]">Timeout</span>
                <span className="text-[var(--text-primary)]">30s</span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-[var(--text-secondary)]">Max Retries</span>
                <span className="text-[var(--text-primary)]">3</span>
              </div>
            </div>

            <Separator />

            <div className="space-y-2">
              <Label className="text-xs flex items-center gap-1">
                <Database className="w-3 h-3" />
                Last Execution Output
              </Label>
              <div className="bg-[var(--bg-tertiary)] rounded-lg p-3 max-h-[200px] overflow-auto">
                <pre className="text-xs font-mono text-[var(--text-secondary)]">
                  {JSON.stringify({ result: 'success', data: {} }, null, 2)}
                </pre>
              </div>
            </div>

            <div className="space-y-2">
              <Label className="text-xs flex items-center gap-1">
                <Code className="w-3 h-3" />
                Logs
              </Label>
              <div className="bg-[var(--bg-tertiary)] rounded-lg p-3 font-mono text-[10px] text-[var(--text-muted)]">
                <div className="text-green-500">[INFO] Node initialized</div>
                <div className="text-blue-500">[DEBUG] Processing input...</div>
                <div className="text-green-500">[INFO] Execution completed in 42ms</div>
              </div>
            </div>
          </TabsContent>
        </ScrollArea>
      </Tabs>
    </div>
  );
}

export default NodeInspector;
