import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
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
import { Textarea } from '@/components/ui/textarea';
import type { Pipeline, PipelineStep } from '@/types';
import {
  AlertCircle,
  ArrowDown,
  ArrowUp,
  CheckCircle,
  Play,
  Plus,
  Settings,
  Trash2,
} from 'lucide-react';
import { useState } from 'react';

interface PipelineStepEditorProps {
  fabricId: string;
  pipeline: Pipeline;
  onUpdate: (steps: PipelineStep[]) => Promise<void>;
}

const stepTypeLabels: Record<string, string> = {
  transform: 'Transform',
  filter: 'Filter',
  aggregate: 'Aggregate',
  enrich: 'Enrich',
  custom: 'Custom',
};

const stepTypeDescriptions: Record<string, string> = {
  transform: 'Transform the input data using functions or mappings',
  filter: 'Filter events based on conditions',
  aggregate: 'Aggregate multiple events into summary data',
  enrich: 'Enrich events with additional data from external sources',
  custom: 'Run custom code or logic',
};

const stepTypeIcons: Record<string, string> = {
  transform: '🔧',
  filter: '🔍',
  aggregate: '📊',
  enrich: '✨',
  custom: '⚙️',
};

const defaultStepConfig: Record<string, Record<string, any>> = {
  transform: {
    mapping: {},
    function: '',
  },
  filter: {
    conditions: [],
    expression: '',
  },
  aggregate: {
    windowSize: 100,
    function: 'sum',
  },
  enrich: {
    sources: [],
    mapping: {},
  },
  custom: {
    code: '',
    language: 'javascript',
  },
};

export function PipelineStepEditor({ pipeline, onUpdate }: PipelineStepEditorProps) {
  const [steps, setSteps] = useState<PipelineStep[]>(pipeline.steps || []);
  const [isAddStepOpen, setIsAddStepOpen] = useState(false);
  const [editingStep, setEditingStep] = useState<PipelineStep | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [newStep, setNewStep] = useState<Partial<PipelineStep>>({
    name: '',
    type: 'transform',
    config: defaultStepConfig.transform,
    enabled: true,
    timeoutMs: 30000,
    retryCount: 3,
  });

  const handleAddStep = async () => {
    if (!newStep.name?.trim()) return;

    setIsSubmitting(true);
    try {
      const step: PipelineStep = {
        id: `step-${Date.now()}`,
        name: newStep.name,
        type: (newStep.type as PipelineStep['type']) || 'transform',
        config: newStep.config || {},
        order: steps.length,
        enabled: newStep.enabled ?? true,
        timeoutMs: newStep.timeoutMs ?? 30000,
        retryCount: newStep.retryCount ?? 3,
      };

      const updatedSteps = [...steps, step];
      setSteps(updatedSteps);
      await onUpdate(updatedSteps);

      setIsAddStepOpen(false);
      setNewStep({
        name: '',
        type: 'transform',
        config: defaultStepConfig.transform,
        enabled: true,
        timeoutMs: 30000,
        retryCount: 3,
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleUpdateStep = async () => {
    if (!editingStep) return;

    setIsSubmitting(true);
    try {
      const updatedSteps = steps.map((s) => (s.id === editingStep.id ? editingStep : s));
      setSteps(updatedSteps);
      await onUpdate(updatedSteps);

      setEditingStep(null);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDeleteStep = async (stepId: string) => {
    if (!confirm('Are you sure you want to delete this step?')) return;

    setIsSubmitting(true);
    try {
      const updatedSteps = steps
        .filter((s) => s.id !== stepId)
        .map((s, idx) => ({ ...s, order: idx }));
      setSteps(updatedSteps);
      await onUpdate(updatedSteps);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleMoveStep = async (stepId: string, direction: 'up' | 'down') => {
    const idx = steps.findIndex((s) => s.id === stepId);
    if (idx === -1) return;

    if (direction === 'up' && idx === 0) return;
    if (direction === 'down' && idx === steps.length - 1) return;

    const newIdx = direction === 'up' ? idx - 1 : idx + 1;
    const updatedSteps = [...steps];
    [updatedSteps[idx], updatedSteps[newIdx]] = [updatedSteps[newIdx], updatedSteps[idx]];

    const reorderedSteps = updatedSteps.map((s, i) => ({ ...s, order: i }));
    setSteps(reorderedSteps);
    await onUpdate(reorderedSteps);
  };

  const handleToggleStep = async (stepId: string) => {
    const updatedSteps = steps.map((s) => (s.id === stepId ? { ...s, enabled: !s.enabled } : s));
    setSteps(updatedSteps);
    await onUpdate(updatedSteps);
  };

  const renderStepConfig = (
    step: Partial<PipelineStep>,
    onChange: (updates: Partial<PipelineStep>) => void
  ) => {
    switch (step.type) {
      case 'transform':
        return (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>Transform Function</Label>
              <Input
                value={step.config?.function || ''}
                onChange={(e) => onChange({ config: { ...step.config, function: e.target.value } })}
                placeholder="e.g., JSON.stringify, uppercase, etc."
              />
            </div>
            <div className="space-y-2">
              <Label>Mapping (JSON)</Label>
              <Textarea
                value={JSON.stringify(step.config?.mapping || {}, null, 2)}
                onChange={(e) => {
                  try {
                    onChange({ config: { ...step.config, mapping: JSON.parse(e.target.value) } });
                  } catch {}
                }}
                placeholder='{"key": "value"}'
                rows={4}
                className="font-mono text-sm"
              />
            </div>
          </div>
        );

      case 'filter':
        return (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>Filter Expression</Label>
              <Input
                value={step.config?.expression || ''}
                onChange={(e) =>
                  onChange({ config: { ...step.config, expression: e.target.value } })
                }
                placeholder="e.g., value > 100"
              />
            </div>
            <div className="space-y-2">
              <Label>Conditions (JSON)</Label>
              <Textarea
                value={JSON.stringify(step.config?.conditions || [], null, 2)}
                onChange={(e) => {
                  try {
                    onChange({
                      config: { ...step.config, conditions: JSON.parse(e.target.value) },
                    });
                  } catch {}
                }}
                placeholder='[{"field": "status", "operator": "eq", "value": "active"}]'
                rows={4}
                className="font-mono text-sm"
              />
            </div>
          </div>
        );

      case 'aggregate':
        return (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>Aggregation Function</Label>
              <Select
                value={step.config?.function || 'sum'}
                onValueChange={(v) => onChange({ config: { ...step.config, function: v } })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="sum">Sum</SelectItem>
                  <SelectItem value="avg">Average</SelectItem>
                  <SelectItem value="min">Min</SelectItem>
                  <SelectItem value="max">Max</SelectItem>
                  <SelectItem value="count">Count</SelectItem>
                  <SelectItem value="collect">Collect</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>Window Size (events)</Label>
              <Input
                type="number"
                value={step.config?.windowSize || 100}
                onChange={(e) =>
                  onChange({ config: { ...step.config, windowSize: parseInt(e.target.value) } })
                }
                min={1}
                max={10000}
              />
            </div>
          </div>
        );

      case 'enrich':
        return (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>Data Sources (comma-separated URLs)</Label>
              <Input
                value={(step.config?.sources || []).join(', ')}
                onChange={(e) =>
                  onChange({
                    config: {
                      ...step.config,
                      sources: e.target.value.split(',').map((s) => s.trim()),
                    },
                  })
                }
                placeholder="https://api.example.com/data, https://api2.example.com/users"
              />
            </div>
            <div className="space-y-2">
              <Label>Enrichment Mapping (JSON)</Label>
              <Textarea
                value={JSON.stringify(step.config?.mapping || {}, null, 2)}
                onChange={(e) => {
                  try {
                    onChange({ config: { ...step.config, mapping: JSON.parse(e.target.value) } });
                  } catch {}
                }}
                placeholder='{"userId": "user_info.id"}'
                rows={4}
                className="font-mono text-sm"
              />
            </div>
          </div>
        );

      case 'custom':
        return (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>Language</Label>
              <Select
                value={step.config?.language || 'javascript'}
                onValueChange={(v) => onChange({ config: { ...step.config, language: v } })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="javascript">JavaScript</SelectItem>
                  <SelectItem value="python">Python</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>Custom Code</Label>
              <Textarea
                value={step.config?.code || ''}
                onChange={(e) => onChange({ config: { ...step.config, code: e.target.value } })}
                placeholder="// Your custom processing code here"
                rows={8}
                className="font-mono text-sm"
              />
            </div>
          </div>
        );

      default:
        return (
          <div className="space-y-2">
            <Label>Configuration (JSON)</Label>
            <Textarea
              value={JSON.stringify(step.config || {}, null, 2)}
              onChange={(e) => {
                try {
                  onChange({ config: JSON.parse(e.target.value) });
                } catch {}
              }}
              placeholder='{"key": "value"}'
              rows={6}
              className="font-mono text-sm"
            />
          </div>
        );
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold">Pipeline Steps</h3>
          <p className="text-sm text-text-muted">Define the processing steps for your pipeline</p>
        </div>
        <Button onClick={() => setIsAddStepOpen(true)}>
          <Plus className="w-4 h-4 mr-2" />
          Add Step
        </Button>
      </div>

      {steps.length === 0 ? (
        <Card className="p-8 text-center">
          <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-bg-tertiary flex items-center justify-center">
            <Play className="w-8 h-8 text-text-muted" />
          </div>
          <p className="text-text-muted mb-4">No steps configured yet</p>
          <p className="text-sm text-text-muted mb-4">
            Add steps to define how data flows through your pipeline
          </p>
          <Button onClick={() => setIsAddStepOpen(true)}>
            <Plus className="w-4 h-4 mr-2" />
            Add Your First Step
          </Button>
        </Card>
      ) : (
        <div className="space-y-3">
          {steps
            .sort((a, b) => a.order - b.order)
            .map((step, index) => (
              <Card key={step.id} className={!step.enabled ? 'opacity-60' : ''}>
                <CardContent className="pt-4">
                  <div className="flex items-start gap-3">
                    <div className="flex flex-col gap-1 mt-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-6 w-6"
                        onClick={() => handleMoveStep(step.id, 'up')}
                        disabled={index === 0}
                      >
                        <ArrowUp className="w-3 h-3" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-6 w-6"
                        onClick={() => handleMoveStep(step.id, 'down')}
                        disabled={index === steps.length - 1}
                      >
                        <ArrowDown className="w-3 h-3" />
                      </Button>
                    </div>

                    <div className="w-10 h-10 rounded-lg bg-bg-secondary flex items-center justify-center shrink-0">
                      <span className="text-xl">{stepTypeIcons[step.type]}</span>
                    </div>

                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="font-medium">{step.name}</span>
                        <Badge variant="outline" className="text-xs">
                          {stepTypeLabels[step.type]}
                        </Badge>
                        {step.enabled ? (
                          <CheckCircle className="w-4 h-4 text-green-400" />
                        ) : (
                          <AlertCircle className="w-4 h-4 text-yellow-400" />
                        )}
                      </div>
                      <p className="text-sm text-text-muted">{stepTypeDescriptions[step.type]}</p>
                      <div className="flex items-center gap-4 mt-2 text-xs text-text-muted">
                        <span>Timeout: {step.timeoutMs}ms</span>
                        <span>Retries: {step.retryCount}</span>
                      </div>
                    </div>

                    <div className="flex items-center gap-1">
                      <Switch
                        checked={step.enabled}
                        onCheckedChange={() => handleToggleStep(step.id)}
                      />
                      <Button variant="ghost" size="icon" onClick={() => setEditingStep(step)}>
                        <Settings className="w-4 h-4" />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => handleDeleteStep(step.id)}>
                        <Trash2 className="w-4 h-4 text-red-400" />
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
        </div>
      )}

      {/* Add Step Dialog */}
      <Dialog open={isAddStepOpen} onOpenChange={setIsAddStepOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Add Pipeline Step</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-4 max-h-[60vh] overflow-y-auto">
            <div className="space-y-2">
              <Label>Step Name</Label>
              <Input
                value={newStep.name || ''}
                onChange={(e) => setNewStep({ ...newStep, name: e.target.value })}
                placeholder="e.g., Validate Input, Transform Data"
              />
            </div>

            <div className="space-y-2">
              <Label>Step Type</Label>
              <Select
                value={newStep.type}
                onValueChange={(v) =>
                  setNewStep({
                    ...newStep,
                    type: v as PipelineStep['type'],
                    config: defaultStepConfig[v] || {},
                  })
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {Object.entries(stepTypeLabels).map(([value, label]) => (
                    <SelectItem key={value} value={value}>
                      <div className="flex items-center gap-2">
                        <span>{stepTypeIcons[value]}</span>
                        <span>{label}</span>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <p className="text-xs text-text-muted">
                {stepTypeDescriptions[newStep.type || 'transform']}
              </p>
            </div>

            {renderStepConfig(newStep, setNewStep)}

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Timeout (ms)</Label>
                <Input
                  type="number"
                  value={newStep.timeoutMs || 30000}
                  onChange={(e) => setNewStep({ ...newStep, timeoutMs: parseInt(e.target.value) })}
                  min={1000}
                  max={300000}
                />
              </div>
              <div className="space-y-2">
                <Label>Retry Count</Label>
                <Input
                  type="number"
                  value={newStep.retryCount || 3}
                  onChange={(e) => setNewStep({ ...newStep, retryCount: parseInt(e.target.value) })}
                  min={0}
                  max={10}
                />
              </div>
            </div>

            <div className="flex items-center justify-between">
              <Label>Enabled</Label>
              <Switch
                checked={newStep.enabled ?? true}
                onCheckedChange={(v) => setNewStep({ ...newStep, enabled: v })}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsAddStepOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleAddStep} disabled={!newStep.name?.trim() || isSubmitting}>
              Add Step
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit Step Dialog */}
      <Dialog open={!!editingStep} onOpenChange={(open) => !open && setEditingStep(null)}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Edit Step: {editingStep?.name}</DialogTitle>
          </DialogHeader>
          {editingStep && (
            <div className="space-y-4 py-4 max-h-[60vh] overflow-y-auto">
              <div className="space-y-2">
                <Label>Step Name</Label>
                <Input
                  value={editingStep.name}
                  onChange={(e) => setEditingStep({ ...editingStep, name: e.target.value })}
                />
              </div>

              <div className="space-y-2">
                <Label>Step Type</Label>
                <Select
                  value={editingStep.type}
                  onValueChange={(v) =>
                    setEditingStep({
                      ...editingStep,
                      type: v as PipelineStep['type'],
                    })
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {Object.entries(stepTypeLabels).map(([value, label]) => (
                      <SelectItem key={value} value={value}>
                        <div className="flex items-center gap-2">
                          <span>{stepTypeIcons[value]}</span>
                          <span>{label}</span>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {renderStepConfig(editingStep, (updates) =>
                setEditingStep((prev) => (prev ? { ...prev, ...updates } : null))
              )}

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>Timeout (ms)</Label>
                  <Input
                    type="number"
                    value={editingStep.timeoutMs}
                    onChange={(e) =>
                      setEditingStep({ ...editingStep, timeoutMs: parseInt(e.target.value) })
                    }
                    min={1000}
                    max={300000}
                  />
                </div>
                <div className="space-y-2">
                  <Label>Retry Count</Label>
                  <Input
                    type="number"
                    value={editingStep.retryCount}
                    onChange={(e) =>
                      setEditingStep({ ...editingStep, retryCount: parseInt(e.target.value) })
                    }
                    min={0}
                    max={10}
                  />
                </div>
              </div>

              <div className="flex items-center justify-between">
                <Label>Enabled</Label>
                <Switch
                  checked={editingStep.enabled}
                  onCheckedChange={(v) => setEditingStep({ ...editingStep, enabled: v })}
                />
              </div>
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditingStep(null)}>
              Cancel
            </Button>
            <Button onClick={handleUpdateStep} disabled={isSubmitting}>
              Save Changes
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
