import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import type { Pipeline, PipelineStep } from '@/types';
import {
  Play,
  Pause,
  AlertCircle,
  CheckCircle,
  ArrowRight,
  Clock,
  Database,
  Filter,
  Wrench,
  Sparkles,
  Layers,
  GitCommit,
} from 'lucide-react';

interface PipelineReadOnlyViewerProps {
  pipelines: Pipeline[];
  fabricName?: string;
}

const stepTypeIcons: Record<string, React.ReactNode> = {
  transform: <Wrench className="h-4 w-4" />,
  filter: <Filter className="h-4 w-4" />,
  aggregate: <Layers className="h-4 w-4" />,
  enrich: <Sparkles className="h-4 w-4" />,
  custom: <GitCommit className="h-4 w-4" />,
};

const stepTypeColors: Record<string, string> = {
  transform: 'bg-blue-500/10 text-blue-600 border-blue-500/20',
  filter: 'bg-purple-500/10 text-purple-600 border-purple-500/20',
  aggregate: 'bg-green-500/10 text-green-600 border-green-500/20',
  enrich: 'bg-amber-500/10 text-amber-600 border-amber-500/20',
  custom: 'bg-gray-500/10 text-gray-600 border-gray-500/20',
};

const statusColors: Record<string, string> = {
  active: 'bg-green-500/10 text-green-600 border-green-500/20',
  paused: 'bg-yellow-500/10 text-yellow-600 border-yellow-500/20',
  error: 'bg-red-500/10 text-red-600 border-red-500/20',
  draft: 'bg-gray-500/10 text-gray-600 border-gray-500/20',
};

function StepCard({
  step,
  index,
  totalSteps,
  isLast,
}: {
  step: PipelineStep;
  index: number;
  totalSteps: number;
  isLast: boolean;
}) {
  const [isExpanded, setIsExpanded] = useState(false);

  return (
    <div className="relative flex items-start gap-4">
      {/* Step number indicator */}
      <div className="flex flex-col items-center">
        <div
          className={`flex h-8 w-8 items-center justify-center rounded-full border-2 text-xs font-semibold ${
            step.enabled
              ? 'border-brand-500 text-brand-500'
              : 'border-gray-400 text-gray-400'
          }`}
        >
          {index + 1}
        </div>
        {!isLast && (
          <div className="h-full min-h-[40px] w-px bg-border-subtle my-2" />
        )}
      </div>

      {/* Step content */}
      <div className="flex-1 pb-4">
        <Card
          className={`cursor-pointer transition-all hover:shadow-md ${
            !step.enabled ? 'opacity-60' : ''
          } ${isExpanded ? 'ring-1 ring-brand-500' : ''}`}
          onClick={() => setIsExpanded(!isExpanded)}
        >
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <div className={`p-1.5 rounded-md ${stepTypeColors[step.type] ?? stepTypeColors.custom}`}>
                  {stepTypeIcons[step.type] ?? stepTypeIcons.custom}
                </div>
                <span className="font-medium">{step.name}</span>
              </div>
              <div className="flex items-center gap-2">
                {step.enabled ? (
                  <CheckCircle className="h-4 w-4 text-green-500" />
                ) : (
                  <AlertCircle className="h-4 w-4 text-yellow-500" />
                )}
                <Badge variant="outline" className={stepTypeColors[step.type] ?? stepTypeColors.custom}>
                  {step.type}
                </Badge>
              </div>
            </div>
          </CardHeader>
          <CardContent className="pt-0">
            <p className="text-sm text-muted-foreground">{step.description || 'No description'}</p>

            {isExpanded && (
              <div className="mt-4 space-y-3 border-t pt-3">
                {step.condition && (
                  <div>
                    <span className="text-xs font-medium text-muted-foreground uppercase">Condition</span>
                    <code className="mt-1 block rounded bg-bg-tertiary px-2 py-1 text-xs font-mono">
                      {step.condition}
                    </code>
                  </div>
                )}
                {step.timeout && (
                  <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <Clock className="h-4 w-4" />
                    <span>Timeout: {step.timeout}ms</span>
                  </div>
                )}
                {step.retry_count !== undefined && step.retry_count > 0 && (
                  <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <AlertCircle className="h-4 w-4" />
                    <span>Retries: {step.retry_count}</span>
                  </div>
                )}
                {step.config && Object.keys(step.config).length > 0 && (
                  <div>
                    <span className="text-xs font-medium text-muted-foreground uppercase">Config</span>
                    <pre className="mt-1 rounded bg-bg-tertiary p-2 text-xs overflow-x-auto">
                      {JSON.stringify(step.config, null, 2)}
                    </pre>
                  </div>
                )}
              </div>
            )}

            {!isExpanded && step.condition && (
              <p className="mt-2 text-xs text-muted-foreground">
                Condition: <code>{step.condition}</code>
              </p>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export function PipelineReadOnlyViewer({ pipelines, fabricName }: PipelineReadOnlyViewerProps) {
  const [selectedPipelineId, setSelectedPipelineId] = useState<string | null>(
    pipelines.length > 0 ? pipelines[0].id : null
  );

  const selectedPipeline = pipelines.find((p) => p.id === selectedPipelineId);

  if (pipelines.length === 0) {
    return (
      <Card className="p-8 text-center">
        <Database className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
        <p className="text-muted-foreground">No pipelines configured</p>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      {/* Pipeline Selector */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span className="text-sm font-medium">Viewing:</span>
          <Select
            value={selectedPipelineId ?? undefined}
            onValueChange={setSelectedPipelineId}
          >
            <SelectTrigger className="w-[280px]">
              <SelectValue placeholder="Select a pipeline" />
            </SelectTrigger>
            <SelectContent>
              {pipelines.map((pipeline) => (
                <SelectItem key={pipeline.id} value={pipeline.id}>
                  <div className="flex items-center gap-2">
                    <span>{pipeline.name}</span>
                    <Badge variant="outline" className={statusColors[pipeline.status]}>
                      {pipeline.status}
                    </Badge>
                  </div>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {selectedPipeline && (
          <div className="flex items-center gap-4 text-sm text-muted-foreground">
            <div className="flex items-center gap-2">
              <span className="text-xs">Throughput:</span>
              <span className="font-medium">{selectedPipeline.throughput?.toFixed(1) || 0}/sec</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-xs">Error Rate:</span>
              <span className="font-medium">
                {((selectedPipeline.errorRate || 0) * 100).toFixed(2)}%
              </span>
            </div>
          </div>
        )}
      </div>

      {/* Pipeline Details */}
      {selectedPipeline ? (
        <div className="grid gap-6 lg:grid-cols-[1fr,400px]">
          {/* Left: Step Flow */}
          <div>
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle>{selectedPipeline.name}</CardTitle>
                    <p className="text-sm text-muted-foreground mt-1">
                      {selectedPipeline.description || 'No description'}
                    </p>
                  </div>
                  <Badge variant="outline" className={statusColors[selectedPipeline.status]}>
                    {selectedPipeline.status === 'active' && <Play className="h-3 w-3 mr-1" />}
                    {selectedPipeline.status === 'paused' && <Pause className="h-3 w-3 mr-1" />}
                    {selectedPipeline.status}
                  </Badge>
                </div>
              </CardHeader>
              <CardContent>
                {selectedPipeline.steps && selectedPipeline.steps.length > 0 ? (
                  <div className="py-2">
                    {[...selectedPipeline.steps]
                      .sort((a, b) => a.order - b.order)
                      .map((step, index, arr) => (
                        <StepCard
                          key={step.id}
                          step={step}
                          index={index}
                          totalSteps={arr.length}
                          isLast={index === arr.length - 1}
                        />
                      ))}
                  </div>
                ) : (
                  <div className="text-center py-8 text-muted-foreground">
                    <p>No steps configured for this pipeline</p>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          {/* Right: Stats & Info */}
          <div className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Pipeline Info</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-xs text-muted-foreground">Total Steps</p>
                    <p className="font-semibold">{selectedPipeline.steps?.length || 0}</p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Active Steps</p>
                    <p className="font-semibold">
                      {selectedPipeline.steps?.filter((s) => s.enabled).length || 0}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Last Executed</p>
                    <p className="font-semibold text-xs">
                      {selectedPipeline.lastExecutedAt
                        ? new Date(selectedPipeline.lastExecutedAt).toLocaleDateString()
                        : 'Never'}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Created</p>
                    <p className="font-semibold text-xs">
                      {new Date(selectedPipeline.createdAt).toLocaleDateString()}
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Step Type Distribution */}
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Step Types</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex flex-wrap gap-2">
                  {Array.from(
                    new Set(selectedPipeline.steps?.map((s) => s.type))
                  ).map((type) => {
                    const count = selectedPipeline.steps?.filter((s) => s.type === type).length ?? 0;
                    return (
                      <Badge
                        key={type}
                        variant="outline"
                        className={stepTypeColors[type] ?? stepTypeColors.custom}
                      >
                        {stepTypeIcons[type]}
                        <span className="ml-1 capitalize">{type}</span>
                        <span className="ml-1">({count})</span>
                      </Badge>
                    );
                  })}
                </div>
              </CardContent>
            </Card>

            {/* Fabric Reference */}
            {fabricName && (
              <Card>
                <CardHeader>
                  <CardTitle className="text-base">State Fabric</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-muted-foreground">{fabricName}</p>
                </CardContent>
              </Card>
            )}
          </div>
        </div>
      ) : (
        <Card className="p-8 text-center">
          <p className="text-muted-foreground">Select a pipeline to view details</p>
        </Card>
      )}
    </div>
  );
}
