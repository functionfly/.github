/**
 * NodeInspector Panel (Right Sidebar)
 * Shows: Input fields, Config options, Runtime settings, Environment variables
 */

import { useCallback, useState, useMemo, useEffect } from 'react';
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
  Code,
  FileJson,
  X,
  Loader2,
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { registryApi } from '@/api/registry';
import { useFRGStore } from '@/stores/frgStore';
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

interface ConfigField {
  key: string;
  type: 'string' | 'number' | 'boolean' | 'json' | 'secret';
  value: unknown;
  description?: string;
  required?: boolean;
  defaultValue?: unknown;
  min?: number;
  max?: number;
  validation?: (value: unknown) => string | null;
}

export function NodeInspector() {
  const store = useFRGStore();
  const selectedNodeId = store.selectedNodeId;
  const nodes = store.nodes;
  const nodeRuntimeStates = store.nodeRuntimeStates;
  const { removeNode, updateNode, setSelectedNode, runNode } = store;

  const [activeTab, setActiveTab] = useState('config');
  const [configValues, setConfigValues] = useState<Record<string, unknown>>({});
  const [envVars, setEnvVars] = useState<Record<string, string>>({});
  const [isRunning, setIsRunning] = useState(false);
  const [sourceCode, setSourceCode] = useState<string | null>(null);
  const [loadingSource, setLoadingSource] = useState(false);
  const [inputSchema, setInputSchema] = useState<Record<string, unknown> | null>(null);
  const [loadingSchema, setLoadingSchema] = useState(false);

  const loadSourceCode = useCallback(async (author: string, name: string, version?: string) => {
    setLoadingSource(true);
    try {
      const code = await registryApi.getFunctionVersionSource(author, name, version);
      setSourceCode(code);
    } catch (err) {
      console.error('[NodeInspector] Failed to load source code:', err);
      setSourceCode(null);
    } finally {
      setLoadingSource(false);
    }
  }, []);

  const selectedNode = useMemo(() => {
    return nodes.find(n => n.id === selectedNodeId) ?? null;
  }, [nodes, selectedNodeId]);

  useEffect(() => {
    if (!selectedNode?.data?.functionRef) {
      setInputSchema(null);
      return;
    }
    const fn = selectedNode.data.functionRef;
    setLoadingSchema(true);
    registryApi.getFunction(fn.author, fn.name)
      .then((response) => {
        const latestVersion = response.versions?.[0];
        if (latestVersion?.manifest?.input_schema) {
          setInputSchema(latestVersion.manifest.input_schema as Record<string, unknown>);
        } else {
          setInputSchema(null);
        }
      })
      .catch(() => setInputSchema(null))
      .finally(() => setLoadingSchema(false));
  }, [selectedNode]);

  const handleClose = useCallback(() => {
    setSelectedNode(null);
  }, [setSelectedNode]);

  const handleConfigChange = useCallback((key: string, value: unknown) => {
    setConfigValues(prev => ({ ...prev, [key]: value }));
  }, []);

  const handleEnvVarChange = useCallback((key: string, value: string) => {
    setEnvVars(prev => ({ ...prev, [key]: value }));
  }, []);

  const configFields: ConfigField[] = [
    {
      key: 'timeout',
      type: 'number',
      value: 30000,
      description: 'Execution timeout in milliseconds',
      required: true,
      min: 1000,
      max: 300000,
      validation: (v) => {
        const n = v as number;
        if (isNaN(n)) return 'Timeout must be a number';
        if (n < 1000) return 'Timeout must be at least 1000ms';
        if (n > 300000) return 'Timeout cannot exceed 300000ms (5 minutes)';
        return null;
      },
    },
    {
      key: 'retryAttempts',
      type: 'number',
      value: 3,
      description: 'Number of retry attempts on failure',
      defaultValue: 3,
      min: 0,
      max: 10,
      validation: (v) => {
        const n = v as number;
        if (isNaN(n)) return 'Retry attempts must be a number';
        if (n < 0) return 'Retry attempts cannot be negative';
        if (n > 10) return 'Retry attempts cannot exceed 10';
        return null;
      },
    },
    { key: 'cacheEnabled', type: 'boolean', value: false, description: 'Enable result caching', defaultValue: false },
    { key: 'apiKey', type: 'secret', value: '', description: 'API key for external service', required: false },
  ];

  const [configErrors, setConfigErrors] = useState<Record<string, string>>({});

  const validateConfig = useCallback(() => {
    const errors: Record<string, string> = {};
    for (const field of configFields) {
      if (field.validation) {
        const currentValue = configValues[field.key] ?? field.value;
        const error = field.validation(currentValue);
        if (error) {
          errors[field.key] = error;
        }
      }
    }
    setConfigErrors(errors);
    return Object.keys(errors).length === 0;
  }, [configValues, configFields]);

  const handleConfigChangeWithValidation = useCallback((key: string, value: unknown) => {
    setConfigValues(prev => ({ ...prev, [key]: value }));
    const field = configFields.find(f => f.key === key);
    if (field?.validation) {
      const error = field.validation(value);
      setConfigErrors(prev => ({ ...prev, [key]: error || '' }));
    }
  }, [configFields]);

  const handleSave = useCallback(() => {
    if (!selectedNode?.data?.functionRef) return;
    if (!validateConfig()) return;
    const fn = selectedNode.data.functionRef;
    updateNode(selectedNode.id, {
      data: {
        ...selectedNode.data,
        functionRef: {
          ...fn,
          config: configValues,
          metadata: {
            ...fn.metadata,
            envVars,
          },
        },
      },
    });
  }, [selectedNode, configValues, envVars, updateNode, validateConfig]);

  const handleDelete = useCallback(() => {
    if (!selectedNode) return;
    removeNode(selectedNode.id);
  }, [selectedNode, removeNode]);

  const handleReset = useCallback(() => {
    if (!selectedNode?.data?.functionRef) return;
    const fn = selectedNode.data.functionRef;
    setConfigValues(fn.config || {});
    setEnvVars((fn.metadata?.envVars as Record<string, string>) || {});
  }, [selectedNode]);

  const handleRun = useCallback(async () => {
    if (!selectedNode) return;
    console.log('[NodeInspector] Running node:', selectedNode.id, selectedNode.data.functionRef);
    setIsRunning(true);
    try {
      await runNode(selectedNode.id);
    } finally {
      setIsRunning(false);
    }
  }, [selectedNode, runNode]);

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

  const SECRET_PATTERNS = [
    /(?:api[_-]?key|apikey|api[_-]?secret)\s*[:=]\s*["']?([a-zA-Z0-9_\-]{16,})["']?/gi,
    /(?:password|passwd|pwd|passphrase)\s*[:=]\s*["']?([^"'\s]{4,})["']?/gi,
    /(?:secret|token|auth(?:orization)?)\s*[:=]\s*["']?([a-zA-Z0-9_\-\.]{16,})["']?/gi,
    /(?:bearer|basic)\s+([a-zA-Z0-9_\-\.]+)/gi,
    /(?:private[_-]?key|priv[_-]?key)\s*[:=]\s*["']?([a-zA-Z0-9_\-\.+=/]{20,})["']?/gi,
    /(?:access[_-]?token|refresh[_-]?token|client[_-]?secret)\s*[:=]\s*["']?([a-zA-Z0-9_\-\._]{16,})["']?/gi,
    /sk-[a-zA-Z0-9]{20,}/g,
    /pk_[a-zA-Z0-9]{20,}/g,
    /eyJ[a-zA-Z0-9_-]+\.eyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+/g,
    /[a-f0-9]{32,}/gi,
    /[A-Za-z0-9]{40,}/g,
  ];

  function redactSecrets(code: string): string {
    let redacted = code;
    for (const pattern of SECRET_PATTERNS) {
      redacted = redacted.replace(pattern, (match) => {
        if (match.length > 16) {
          return match.slice(0, 6) + '****' + match.slice(-4);
        }
        return '****';
      });
    }
    return redacted;
  }

  const { functionRef } = selectedNode.data;
  const runtimeState = nodeRuntimeStates[selectedNode.id];
  if (!functionRef) {
    return (
      <div className="flex flex-col items-center justify-center h-full p-6 text-center">
        <div className="w-16 h-16 rounded-full bg-[var(--bg-tertiary)] flex items-center justify-center mb-4">
          <AlertCircle className="w-8 h-8 text-[var(--text-muted)]" />
        </div>
        <h3 className="text-sm font-medium text-[var(--text-secondary)] mb-1">
          Invalid Node Data
        </h3>
        <p className="text-xs text-[var(--text-muted)]">
          This node does not have valid function data
        </p>
      </div>
    );
  }

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
            {field.min !== undefined && field.max !== undefined && (
              <p className="text-[10px] text-[var(--text-muted)]">Range: {field.min} - {field.max}</p>
            )}
            <Input
              type="number"
              value={value as number}
              onChange={(e) => handleConfigChangeWithValidation(field.key, parseInt(e.target.value))}
              className={`h-8 ${configErrors[field.key] ? 'border-red-500' : ''}`}
            />
            {configErrors[field.key] && (
              <p className="text-[10px] text-red-500 flex items-center gap-1">
                <AlertCircle className="w-3 h-3" />
                {configErrors[field.key]}
              </p>
            )}
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
      <div className="p-4 border-b border-[var(--border-subtle)]">
        <div className="flex items-start justify-between">
          <div>
            <div className="flex items-center gap-2">
              <h3 className="font-semibold text-[var(--text-primary)]">
                {functionRef.name}
              </h3>
              <span className="text-xs text-[var(--text-secondary)]">
                {functionRef.author}/{functionRef.version}
              </span>
            </div>
            <div className="flex gap-1 mt-2">
              <Button variant="ghost" size="icon" className="h-8 w-8" onClick={handleRun} disabled={isRunning}>
                {isRunning ? <Loader2 className="w-4 h-4 animate-spin" /> : <Play className="w-4 h-4" />}
              </Button>
              <Button variant="ghost" size="icon" className="h-8 w-8" onClick={handleSave}>
                <Save className="w-4 h-4" />
              </Button>
              <Button variant="ghost" size="icon" className="h-8 w-8 text-red-500" onClick={handleDelete}>
                <Trash2 className="w-4 h-4" />
              </Button>
            </div>
          </div>
          <div className="flex items-center gap-2">
            {runtimeState && (
              <Badge
                variant={runtimeState.status === 'failed' ? 'destructive' : runtimeState.status === 'completed' ? 'default' : 'secondary'}
                className="text-[10px]"
              >
                {runtimeState.status === 'completed' && <CheckCircle className="w-3 h-3 mr-1" />}
                {runtimeState.status === 'failed' && <AlertCircle className="w-3 h-3 mr-1" />}
                {runtimeState.status}
              </Badge>
            )}
            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={handleClose}>
              <X className="w-4 h-4" />
            </Button>
          </div>
        </div>
      </div>

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
          <div className="p-4">
            <Button
              variant="outline"
              size="sm"
              className="w-full mb-2"
              onClick={() => functionRef && loadSourceCode(functionRef.author, functionRef.name, functionRef.version)}
              disabled={loadingSource || !functionRef}
            >
              {loadingSource ? (
                <>
                  <Loader2 className="w-4 h-4 mr-1 animate-spin" />
                  Loading Source...
                </>
              ) : (
                <>
                  <Code className="w-4 h-4 mr-1" />
                  View Source Code
                </>
              )}
            </Button>
            {sourceCode && (
              <div className="bg-[var(--bg-tertiary)] rounded-lg p-3 max-h-[300px] overflow-auto mb-4">
                <pre className="text-xs font-mono text-[var(--text-secondary)]">
                  {redactSecrets(sourceCode)}
                </pre>
              </div>
            )}
          </div>

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
                    <div className="relative z-[300]">
                      <Select
                        value={envVars.NODE_ENV || 'development'}
                        onValueChange={(v) => handleEnvVarChange('NODE_ENV', v)}
                      >
                        <SelectTrigger className="h-8">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent style={{ zIndex: 300 }}>
                          <SelectItem value="development">development</SelectItem>
                          <SelectItem value="staging">staging</SelectItem>
                          <SelectItem value="production">production</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
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
              {loadingSchema ? (
                <div className="bg-[var(--bg-tertiary)] rounded-lg p-3 flex items-center gap-2">
                  <Loader2 className="w-4 h-4 animate-spin text-[var(--text-muted)]" />
                  <span className="text-xs text-[var(--text-muted)]">Loading schema...</span>
                </div>
              ) : inputSchema ? (
                <div className="bg-[var(--bg-tertiary)] rounded-lg p-3">
                  <pre className="text-xs font-mono text-[var(--text-secondary)]">
                    {JSON.stringify(inputSchema, null, 2)}
                  </pre>
                </div>
              ) : (
                <div className="bg-[var(--bg-tertiary)] rounded-lg p-3">
                  <p className="text-xs text-[var(--text-muted)]">No input schema available for this function.</p>
                  <p className="text-xs text-[var(--text-muted)] mt-1">The function may not have defined an input schema or schema information is not available.</p>
                </div>
              )}
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
                <span className="text-[var(--text-secondary)]">Status</span>
                <Badge variant={runtimeState?.status === 'failed' ? 'destructive' : runtimeState?.status === 'completed' ? 'default' : 'secondary'}>
                  {runtimeState?.status || 'idle'}
                </Badge>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-[var(--text-secondary)]">Duration</span>
                <span className="text-[var(--text-primary)]">{runtimeState?.durationMs ? `${runtimeState.durationMs}ms` : '—'}</span>
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
                  {runtimeState?.output
                    ? (typeof runtimeState.output === 'string' && runtimeState.output.includes('=')
                        ? redactSecrets(runtimeState.output)
                        : JSON.stringify(runtimeState.output, null, 2))
                    : runtimeState?.error
                    ? `Error: ${runtimeState.error}`
                    : 'No output yet'}
                </pre>
              </div>
            </div>

            {runtimeState?.generatedCode && (
              <div className="space-y-2">
                <Label className="text-xs flex items-center gap-1">
                  <Code className="w-3 h-3" />
                  Generated Code
                </Label>
                <div className="bg-[var(--bg-tertiary)] rounded-lg p-3 max-h-[300px] overflow-auto">
                  <pre className="text-xs font-mono text-[var(--text-secondary)]">
                    {redactSecrets(runtimeState.generatedCode)}
                  </pre>
                </div>
              </div>
            )}
          </TabsContent>
        </ScrollArea>
      </Tabs>
    </div>
  );
}

export default NodeInspector;