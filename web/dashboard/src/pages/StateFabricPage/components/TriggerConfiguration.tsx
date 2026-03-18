import { useState } from "react";
import { Plus, Zap, Trash2, Play, Pause, Code } from "lucide-react";
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { useStateFabricTriggers, useCreateTrigger, useDeleteTrigger } from "@/hooks/useStateFabric";
import type { StateFabricTrigger } from "@/types";

interface TriggerConfigurationProps {
  fabricId: string;
}

const triggerTypeLabels: Record<string, string> = {
  on_create: "On Create",
  on_update: "On Update",
  on_delete: "On Delete",
  on_read: "On Read",
  on_condition: "On Condition",
  scheduled: "Scheduled",
};

const triggerTypeDescriptions: Record<string, string> = {
  on_create: "Triggered when a new key is created",
  on_update: "Triggered when a key is updated",
  on_delete: "Triggered when a key is deleted",
  on_read: "Triggered when a key is read",
  on_condition: "Triggered when a condition is met",
  scheduled: "Triggered on a schedule",
};

const statusColors: Record<string, string> = {
  active: "bg-green-500/10 text-green-400",
  inactive: "bg-gray-500/10 text-gray-400",
  error: "bg-red-500/10 text-red-400",
};

export function TriggerConfiguration({ fabricId }: TriggerConfigurationProps) {
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  
  // Form state
  const [triggerType, setTriggerType] = useState("on_update");
  const [keyPattern, setKeyPattern] = useState("*");
  const [targetFunction, setTargetFunction] = useState("");
  const [includePrevious, setIncludePrevious] = useState(true);
  const [includeNew, setIncludeNew] = useState(true);
  const [maxInvocations, setMaxInvocations] = useState("60");
  const [isActive, setIsActive] = useState(true);

  const { data: triggersData, isLoading, refetch } = useStateFabricTriggers(fabricId);
  const createTrigger = useCreateTrigger(fabricId);
  const deleteTrigger = useDeleteTrigger(fabricId);

  const triggers = triggersData?.triggers || [];

  const handleCreate = async () => {
    if (!targetFunction.trim()) return;
    await createTrigger.mutateAsync({
      triggerType,
      keyPattern: keyPattern || undefined,
      targetFunction,
      includePrevious,
      includeNew,
      maxInvocationsPerMinute: parseInt(maxInvocations) || 60,
      isActive,
    });
    setIsCreateOpen(false);
    resetForm();
    refetch();
  };

  const resetForm = () => {
    setTriggerType("on_update");
    setKeyPattern("*");
    setTargetFunction("");
    setIncludePrevious(true);
    setIncludeNew(true);
    setMaxInvocations("60");
    setIsActive(true);
  };

  const handleDelete = async (triggerId: string) => {
    if (confirm("Are you sure you want to delete this trigger?")) {
      await deleteTrigger.mutateAsync(triggerId);
      refetch();
    }
  };

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold text-text-primary">Triggers</h3>
          <p className="text-sm text-text-muted">
            Automate actions when state changes occur
          </p>
        </div>
        <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
          <DialogTrigger asChild>
            <Button size="sm">
              <Plus className="w-4 h-4 mr-2" />
              Add Trigger
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New Trigger</DialogTitle>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="triggerType">Trigger Type</Label>
                <Select value={triggerType} onValueChange={setTriggerType}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="on_create">On Create</SelectItem>
                    <SelectItem value="on_update">On Update</SelectItem>
                    <SelectItem value="on_delete">On Delete</SelectItem>
                    <SelectItem value="on_read">On Read</SelectItem>
                    <SelectItem value="on_condition">On Condition</SelectItem>
                    <SelectItem value="scheduled">Scheduled</SelectItem>
                  </SelectContent>
                </Select>
                <p className="text-xs text-text-muted">
                  {triggerTypeDescriptions[triggerType]}
                </p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="keyPattern">Key Pattern (glob)</Label>
                <Input
                  id="keyPattern"
                  value={keyPattern}
                  onChange={(e) => setKeyPattern(e.target.value)}
                  placeholder="* or user:*"
                />
                <p className="text-xs text-text-muted">
                  Use * to match all keys, or user:* to match keys starting with "user:"
                </p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="targetFunction">Target Function</Label>
                <Input
                  id="targetFunction"
                  value={targetFunction}
                  onChange={(e) => setTargetFunction(e.target.value)}
                  placeholder="my-function-name"
                />
                <p className="text-xs text-text-muted">
                  The function to invoke when the trigger fires
                </p>
              </div>
              <div className="flex items-center justify-between">
                <div className="space-y-2">
                  <Label htmlFor="includePrevious">Include Previous Value</Label>
                  <p className="text-xs text-text-muted">
                    Pass the old value to the function
                  </p>
                </div>
                <Switch
                  id="includePrevious"
                  checked={includePrevious}
                  onCheckedChange={setIncludePrevious}
                />
              </div>
              <div className="flex items-center justify-between">
                <div className="space-y-2">
                  <Label htmlFor="includeNew">Include New Value</Label>
                  <p className="text-xs text-text-muted">
                    Pass the new value to the function
                  </p>
                </div>
                <Switch
                  id="includeNew"
                  checked={includeNew}
                  onCheckedChange={setIncludeNew}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="maxInvocations">Max Invocations per Minute</Label>
                <Input
                  id="maxInvocations"
                  type="number"
                  value={maxInvocations}
                  onChange={(e) => setMaxInvocations(e.target.value)}
                  placeholder="60"
                />
              </div>
              <div className="flex items-center justify-between">
                <div className="space-y-2">
                  <Label htmlFor="isActive">Active</Label>
                  <p className="text-xs text-text-muted">
                    Enable or disable this trigger
                  </p>
                </div>
                <Switch
                  id="isActive"
                  checked={isActive}
                  onCheckedChange={setIsActive}
                />
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setIsCreateOpen(false)}>
                Cancel
              </Button>
              <Button onClick={handleCreate} disabled={!targetFunction.trim()}>
                Create Trigger
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {/* Trigger List */}
      {isLoading ? (
        <Card className="p-8 text-center">
          <p className="text-text-muted">Loading triggers...</p>
        </Card>
      ) : triggers.length === 0 ? (
        <Card className="p-8 text-center">
          <Zap className="w-12 h-12 mx-auto mb-4 text-text-muted" />
          <p className="text-text-muted mb-4">No triggers configured yet</p>
          <p className="text-sm text-text-muted mb-4">
            Triggers allow you to execute functions automatically when state changes
          </p>
          <Button onClick={() => setIsCreateOpen(true)}>
            <Plus className="w-4 h-4 mr-2" />
            Add Your First Trigger
          </Button>
        </Card>
      ) : (
        <div className="grid gap-4">
          {triggers.map((trigger) => (
            <Card key={trigger.id}>
              <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-lg bg-brand-500/10 flex items-center justify-center">
                      <Zap className="w-5 h-5 text-brand-400" />
                    </div>
                    <div>
                      <CardTitle className="text-lg">{trigger.targetFunction}</CardTitle>
                      <div className="flex items-center gap-2 mt-1">
                        <Badge variant="secondary" className="text-xs">
                          {triggerTypeLabels[trigger.triggerType] || trigger.triggerType}
                        </Badge>
                        <Badge className={`text-xs ${statusColors[trigger.isActive ? "active" : "inactive"]}`}>
                          {trigger.isActive ? "Active" : "Inactive"}
                        </Badge>
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Button variant="ghost" size="icon" aria-label="Edit trigger code">
                      <Code className="w-4 h-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleDelete(trigger.id)}
                      aria-label="Delete trigger"
                    >
                      <Trash2 className="w-4 h-4 text-red-400" />
                    </Button>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                {/* Key Pattern */}
                {trigger.keyPattern && (
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-text-muted">Key Pattern:</span>
                    <code className="px-2 py-1 bg-bg-secondary rounded text-sm font-mono">
                      {trigger.keyPattern}
                    </code>
                  </div>
                )}

                {/* Configuration */}
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4 pt-4 border-t border-border-subtle">
                  <div className="flex items-center gap-2">
                    {trigger.includePrevious ? (
                      <Play className="w-4 h-4 text-green-400" />
                    ) : (
                      <Pause className="w-4 h-4 text-gray-400" />
                    )}
                    <div>
                      <p className="text-xs text-text-muted">Previous</p>
                      <p className="font-medium text-sm">
                        {trigger.includePrevious ? "Included" : "Excluded"}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {trigger.includeNew ? (
                      <Play className="w-4 h-4 text-green-400" />
                    ) : (
                      <Pause className="w-4 h-4 text-gray-400" />
                    )}
                    <div>
                      <p className="text-xs text-text-muted">New Value</p>
                      <p className="font-medium text-sm">
                        {trigger.includeNew ? "Included" : "Excluded"}
                      </p>
                    </div>
                  </div>
                  <div>
                    <p className="text-xs text-text-muted">Rate Limit</p>
                    <p className="font-medium text-sm">
                      {trigger.maxInvocationsPerMinute}/min
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-text-muted">Created</p>
                    <p className="font-medium text-sm">
                      {new Date(trigger.createdAt).toLocaleDateString()}
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
