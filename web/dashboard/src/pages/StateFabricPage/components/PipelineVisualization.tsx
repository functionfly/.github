import { useState, useEffect, useCallback } from "react";
import { Plus, Play, Pause, Trash2, ArrowRight, CheckCircle, AlertCircle, Zap, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Progress } from "@/components/ui/progress";
import {
  useCreatePipeline,
  useDeletePipeline,
  useUpdatePipeline,
  useExecutePipeline,
} from "@/hooks/useStateFabric";
import { stateFabricApi } from "@/api/stateFabric";
import type { Pipeline, PipelineStep } from "@/types";
import { PipelineStepEditor } from "./PipelineStepEditor";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";

type StepExecutionStatus = "pending" | "running" | "completed" | "error";

interface StepExecution {
  stepId: string;
  status: StepExecutionStatus;
  output?: any;
  error?: string;
  durationMs?: number;
}

interface PipelineVisualizationProps {
  fabricId: string;
  pipelines: Pipeline[];
}

const stepTypeIcons: Record<string, string> = {
  transform: "🔧",
  filter: "🔍",
  aggregate: "📊",
  enrich: "✨",
  custom: "⚙️",
};

const statusColors: Record<string, string> = {
  active: "bg-green-500/10 text-green-400 border-green-500/20",
  paused: "bg-yellow-500/10 text-yellow-400 border-yellow-500/20",
  error: "bg-red-500/10 text-red-400 border-red-500/20",
  draft: "bg-gray-500/10 text-gray-400 border-gray-500/20",
};

export function PipelineVisualization({ fabricId, pipelines }: PipelineVisualizationProps) {
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newPipelineName, setNewPipelineName] = useState("");
  const [newPipelineDescription, setNewPipelineDescription] = useState("");
  const [editingPipeline, setEditingPipeline] = useState<Pipeline | null>(null);
  const [runningPipeline, setRunningPipeline] = useState<Pipeline | null>(null);
  const [executionInput, setExecutionInput] = useState("{}");
  const [executionResult, setExecutionResult] = useState<any>(null);

  // Delete confirmation dialog state
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [pipelineToDelete, setPipelineToDelete] = useState<string | null>(null);

  // Live execution tracking state
  const [stepExecutions, setStepExecutions] = useState<StepExecution[]>([]);
  const [executionProgress, setExecutionProgress] = useState(0);
  const [isExecuting, setIsExecuting] = useState(false);

  const createPipeline = useCreatePipeline(fabricId);
  const deletePipeline = useDeletePipeline(fabricId);
  const updatePipeline = useUpdatePipeline(fabricId);
  const executePipeline = useExecutePipeline(fabricId, runningPipeline?.id || "");

  const handleCreate = async () => {
    if (!newPipelineName.trim()) return;
    await createPipeline.mutateAsync({
      name: newPipelineName,
      description: newPipelineDescription,
      steps: [],
    });
    setIsCreateOpen(false);
    setNewPipelineName("");
    setNewPipelineDescription("");
  };

  const handleToggleStatus = async (pipeline: Pipeline) => {
    await updatePipeline.mutateAsync({
      pipelineId: pipeline.id,
      data: {
        status: pipeline.status === "active" ? "paused" : "active",
      },
    });
  };

  const openDeleteDialog = (pipelineId: string) => {
    setPipelineToDelete(pipelineId);
    setDeleteDialogOpen(true);
  };

  const handleDelete = async () => {
    if (!pipelineToDelete) return;
    await deletePipeline.mutateAsync(pipelineToDelete);
    setDeleteDialogOpen(false);
    setPipelineToDelete(null);
  };

  const handleCancelDelete = () => {
    setDeleteDialogOpen(false);
    setPipelineToDelete(null);
  };

  const handleRunPipeline = async () => {
    if (!runningPipeline) return;

    setIsExecuting(true);
    setExecutionProgress(0);
    setExecutionResult(null);

    const sortedSteps = runningPipeline.steps.sort((a, b) => a.order - b.order);

    // Initialize step executions for live tracking
    const initialExecutions = sortedSteps.map((step) => ({
      stepId: step.id,
      status: "pending" as StepExecutionStatus,
    }));
    setStepExecutions(initialExecutions);

    try {
      let inputData: Record<string, any> = {};
      try {
        inputData = JSON.parse(executionInput);
      } catch {
        inputData = { raw: executionInput };
      }

      // Mark all steps as running
      setStepExecutions((prev) =>
        prev.map((exec) => ({ ...exec, status: "running" as StepExecutionStatus }))
      );
      setExecutionProgress(10);

      // Execute pipeline via API
      const { executionId } = await executePipeline.mutateAsync(inputData);
      setExecutionProgress(20);

      // Poll for execution status until completed or failed
      const maxPolls = 60;
      const pollInterval = 500;
      let pollCount = 0;

      const pollStatus = async (): Promise<void> => {
        if (pollCount >= maxPolls) {
          throw new Error("Execution timed out");
        }

        pollCount++;
        const status = await stateFabricApi.getPipelineExecution(fabricId, runningPipeline.id, executionId);

        // Update progress based on execution status
        if (status.status === "completed") {
          setExecutionProgress(100);
          // Mark all steps as completed with real durations
          setStepExecutions((prev) =>
            prev.map((exec) => ({
              ...exec,
              status: "completed" as StepExecutionStatus,
              durationMs: status.steps?.find((s) => s.id === exec.stepId)?.durationMs,
            }))
          );
          setExecutionResult({ executionId, status: "completed" });
          return;
        }

        if (status.status === "failed") {
          const failedStep = status.steps?.find((s) => s.status === "error");
          setStepExecutions((prev) =>
            prev.map((exec) => ({
              ...exec,
              status: failedStep?.id === exec.stepId ? "error" : "completed",
              error: failedStep?.id === exec.stepId ? failedStep.error : undefined,
            }))
          );
          setExecutionResult({ executionId, status: "failed", error: failedStep?.error });
          return;
        }

        // Update step statuses from API response
        if (status.steps) {
          setStepExecutions((prev) =>
            prev.map((exec) => {
              const stepStatus = status.steps?.find((s) => s.id === exec.stepId);
              return {
                ...exec,
                status: (stepStatus?.status as StepExecutionStatus) || exec.status,
                durationMs: stepStatus?.durationMs,
                error: stepStatus?.error,
              };
            })
          );
        }

        // Calculate progress (20-90% range for polling)
        const estimatedProgress = 20 + Math.min((pollCount / maxPolls) * 70, 70);
        setExecutionProgress(estimatedProgress);

        // Continue polling
        await new Promise((resolve) => setTimeout(resolve, pollInterval));
        return pollStatus();
      };

      await pollStatus();
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "Execution failed";
      setExecutionResult({ error: errorMessage });

      // Mark all running steps as error
      setStepExecutions((prev) =>
        prev.map((exec) =>
          exec.status === "running"
            ? { ...exec, status: "error" as StepExecutionStatus, error: errorMessage }
            : exec
        )
      );
    } finally {
      setIsExecuting(false);
    }
  };

  const getStepStatusIcon = (status: StepExecutionStatus) => {
    switch (status) {
      case "pending":
        return <div className="w-3 h-3 rounded-full bg-gray-400" />;
      case "running":
        return <Loader2 className="w-3 h-3 text-blue-400 animate-spin" />;
      case "completed":
        return <CheckCircle className="w-3 h-3 text-green-400" />;
      case "error":
        return <AlertCircle className="w-3 h-3 text-red-400" />;
      default:
        return null;
    }
  };

  const openRunDialog = (pipeline: Pipeline) => {
    setRunningPipeline(pipeline);
    setExecutionInput("{}");
    setExecutionResult(null);
    setStepExecutions([]);
    setExecutionProgress(0);
    setIsExecuting(false);
  };

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold text-text-primary">Pipelines</h3>
        <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
          <DialogTrigger asChild>
            <Button size="sm">
              <Plus className="w-4 h-4 mr-2" />
              Create Pipeline
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New Pipeline</DialogTitle>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="name">Name</Label>
                <Input
                  id="name"
                  value={newPipelineName}
                  onChange={(e) => setNewPipelineName(e.target.value)}
                  placeholder="Enter pipeline name"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="description">Description</Label>
                <Input
                  id="description"
                  value={newPipelineDescription}
                  onChange={(e) => setNewPipelineDescription(e.target.value)}
                  placeholder="Enter pipeline description"
                />
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setIsCreateOpen(false)}>
                Cancel
              </Button>
              <Button onClick={handleCreate} disabled={!newPipelineName.trim()}>
                Create
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {/* Pipeline List */}
      {pipelines.length === 0 ? (
        <Card className="p-8 text-center">
          <p className="text-text-muted mb-4">No pipelines configured yet</p>
          <Button onClick={() => setIsCreateOpen(true)}>
            <Plus className="w-4 h-4 mr-2" />
            Create Your First Pipeline
          </Button>
        </Card>
      ) : (
        <div className="grid gap-4">
          {pipelines.map((pipeline) => (
            <Card key={pipeline.id}>
              <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <CardTitle className="text-lg">{pipeline.name}</CardTitle>
                    <Badge variant="outline" className={statusColors[pipeline.status]}>
                      {pipeline.status}
                    </Badge>
                  </div>
                  <div className="flex items-center gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => openRunDialog(pipeline)}
                      disabled={pipeline.status !== "active"}
                    >
                      <Zap className="w-4 h-4 mr-1" />
                      Run
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setEditingPipeline(pipeline)}
                    >
                      <Plus className="w-4 h-4 mr-1" />
                      Steps
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleToggleStatus(pipeline)}
                      aria-label={pipeline.status === 'active' ? 'Pause pipeline' : 'Start pipeline'}
                    >
                      {pipeline.status === "active" ? (
                        <Pause className="w-4 h-4" />
                      ) : (
                        <Play className="w-4 h-4" />
                      )}
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => openDeleteDialog(pipeline.id)}
                      aria-label="Delete pipeline"
                    >
                      <Trash2 className="w-4 h-4 text-red-400" />
                    </Button>
                  </div>
                </div>
                <p className="text-sm text-text-muted">{pipeline.description}</p>
              </CardHeader>
              <CardContent>
                {/* Pipeline Steps Visualization */}
                {pipeline.steps && pipeline.steps.length > 0 ? (
                  <div className="flex items-center gap-2 overflow-x-auto pb-2">
                    {pipeline.steps
                      .sort((a, b) => a.order - b.order)
                      .map((step, index) => (
                        <div key={step.id} className="flex items-center gap-2 shrink-0">
                          <div
                            className={`flex items-center gap-2 px-3 py-2 rounded-lg border ${
                              step.enabled
                                ? "bg-bg-secondary border-border-subtle"
                                : "bg-bg-secondary/50 border-border-subtle/50 opacity-50"
                            }`}
                          >
                            <span>{stepTypeIcons[step.type] || "⚙️"}</span>
                            <span className="text-sm font-medium">{step.name}</span>
                            {step.enabled ? (
                              <CheckCircle className="w-3 h-3 text-green-400" />
                            ) : (
                              <AlertCircle className="w-3 h-3 text-yellow-400" />
                            )}
                          </div>
                          {index < pipeline.steps.length - 1 && (
                            <ArrowRight className="w-4 h-4 text-text-muted shrink-0" />
                          )}
                        </div>
                      ))}
                  </div>
                ) : (
                  <p className="text-sm text-text-muted">No steps configured</p>
                )}

                {/* Metrics */}
                <div className="grid grid-cols-3 gap-4 mt-4 pt-4 border-t border-border-subtle">
                  <div>
                    <p className="text-xs text-text-muted">Throughput</p>
                    <p className="font-medium">{pipeline.throughput?.toFixed(1) || 0} /sec</p>
                  </div>
                  <div>
                    <p className="text-xs text-text-muted">Error Rate</p>
                    <p className="font-medium">{((pipeline.errorRate || 0) * 100).toFixed(2)}%</p>
                  </div>
                  <div>
                    <p className="text-xs text-text-muted">Last Executed</p>
                    <p className="font-medium">
                      {pipeline.lastExecutedAt
                        ? new Date(pipeline.lastExecutedAt).toLocaleDateString()
                        : "Never"}
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Step Editor Dialog */}
      <Dialog open={!!editingPipeline} onOpenChange={(open) => !open && setEditingPipeline(null)}>
        <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Edit Steps: {editingPipeline?.name}</DialogTitle>
          </DialogHeader>
          {editingPipeline && (
            <PipelineStepEditor
              fabricId={fabricId}
              pipeline={editingPipeline}
              onUpdate={async (steps) => {
                await updatePipeline.mutateAsync({
                  pipelineId: editingPipeline.id,
                  data: { steps },
                });
                setEditingPipeline({ ...editingPipeline, steps });
              }}
            />
          )}
        </DialogContent>
      </Dialog>

      {/* Pipeline Execution Dialog */}
      <Dialog open={!!runningPipeline} onOpenChange={(open) => !open && setRunningPipeline(null)}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Run Pipeline: {runningPipeline?.name}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>Input Data (JSON)</Label>
              <Textarea
                value={executionInput}
                onChange={(e) => setExecutionInput(e.target.value)}
                placeholder='{"key": "value"}'
                rows={6}
                className="font-mono text-sm"
                disabled={isExecuting}
              />
              <p className="text-xs text-text-muted">
                Enter the input data for the pipeline as JSON
              </p>
            </div>

            {/* Live Execution Tracking */}
            {(isExecuting || stepExecutions.length > 0) && runningPipeline && (
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <Label>Execution Progress</Label>
                  <span className="text-xs text-text-muted">
                    {Math.round(executionProgress)}%
                  </span>
                </div>
                <Progress value={executionProgress} className="h-2" />

                {/* Step Execution Status */}
                <div className="space-y-2 mt-4">
                  <Label className="text-xs text-text-muted">Step Status</Label>
                  <div className="flex items-center gap-2 overflow-x-auto pb-2">
                    {runningPipeline.steps
                      .sort((a, b) => a.order - b.order)
                      .map((step, index) => {
                        const execution = stepExecutions[index];
                        return (
                          <div key={step.id} className="flex items-center gap-2 shrink-0">
                            <div
                              className={`flex items-center gap-2 px-3 py-2 rounded-lg border ${
                                step.enabled
                                  ? "bg-bg-secondary border-border-subtle"
                                  : "bg-bg-secondary/50 border-border-subtle/50 opacity-50"
                              } ${
                                execution?.status === "running"
                                  ? "ring-2 ring-blue-500/50"
                                  : ""
                              }`}
                            >
                              <span>{stepTypeIcons[step.type] || "⚙️"}</span>
                              <span className="text-sm font-medium">{step.name}</span>
                              {execution ? (
                                getStepStatusIcon(execution.status)
                              ) : step.enabled ? (
                                <CheckCircle className="w-3 h-3 text-green-400" />
                              ) : (
                                <AlertCircle className="w-3 h-3 text-yellow-400" />
                              )}
                            </div>
                            {index < runningPipeline.steps.length - 1 && (
                              <ArrowRight className="w-4 h-4 text-text-muted shrink-0" />
                            )}
                          </div>
                        );
                      })}
                  </div>
                </div>

                {/* Step Execution Details */}
                {stepExecutions.some((e) => e.status !== "pending") && (
                  <div className="space-y-2 mt-2">
                    {stepExecutions
                      .filter((e) => e.status !== "pending")
                      .map((execution, idx) => {
                        const step = runningPipeline.steps.find((s) => s.id === execution.stepId);
                        return (
                          <div
                            key={execution.stepId}
                            className={`flex items-center justify-between px-3 py-2 rounded text-sm ${
                              execution.status === "error"
                                ? "bg-red-500/10 text-red-400"
                                : execution.status === "completed"
                                ? "bg-green-500/10 text-green-400"
                                : "bg-blue-500/10 text-blue-400"
                            }`}
                          >
                            <div className="flex items-center gap-2">
                              {getStepStatusIcon(execution.status)}
                              <span>{step?.name || execution.stepId}</span>
                            </div>
                            {execution.durationMs && (
                              <span className="text-xs text-text-muted">
                                {execution.durationMs}ms
                              </span>
                            )}
                          </div>
                        );
                      })}
                  </div>
                )}
              </div>
            )}

            {executionResult && !isExecuting && (
              <div className="space-y-2">
                <Label>Result</Label>
                <div className={`p-4 rounded-lg ${executionResult.error ? "bg-red-500/10 border border-red-500/20" : "bg-green-500/10 border border-green-500/20"}`}>
                  <pre className="text-sm font-mono overflow-x-auto">
                    {JSON.stringify(executionResult, null, 2)}
                  </pre>
                </div>
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRunningPipeline(null)}>
              Close
            </Button>
            <Button
              onClick={handleRunPipeline}
              disabled={isExecuting || !executionInput.trim()}
            >
              {isExecuting ? (
                <>
                  <LoadingSpinner size="sm" className="mr-2" />
                  Running...
                </>
              ) : (
                <>
                  <Zap className="w-4 h-4 mr-2" />
                  Execute
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Pipeline</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete this pipeline? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={handleCancelDelete}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              className="bg-red-500 hover:bg-red-600"
            >
              {deletePipeline.isPending ? "Deleting..." : "Delete"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
