import { useState } from "react";
import { Plus, Play, Pause, Settings, Trash2, ArrowRight, CheckCircle, AlertCircle } from "lucide-react";
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
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  useCreatePipeline,
  useDeletePipeline,
  useUpdatePipeline,
} from "@/hooks/useStateFabric";
import type { Pipeline, PipelineStep } from "@/types";

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

  const createPipeline = useCreatePipeline(fabricId);
  const deletePipeline = useDeletePipeline(fabricId);
  const updatePipeline = useUpdatePipeline(fabricId);

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

  const handleDelete = async (pipelineId: string) => {
    if (confirm("Are you sure you want to delete this pipeline?")) {
      await deletePipeline.mutateAsync(pipelineId);
    }
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
                      variant="ghost"
                      size="icon"
                      onClick={() => handleToggleStatus(pipeline)}
                    >
                      {pipeline.status === "active" ? (
                        <Pause className="w-4 h-4" />
                      ) : (
                        <Play className="w-4 h-4" />
                      )}
                    </Button>
                    <Button variant="ghost" size="icon" aria-label="Pipeline settings">
                      <Settings className="w-4 h-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleDelete(pipeline.id)}
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
    </div>
  );
}
