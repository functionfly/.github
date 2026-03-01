import { useState } from "react";
import { X, Server, DollarSign, AlertTriangle, CheckCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";
import { ReplayExecutionButton, ReplayMode } from "./ReplayExecutionButton";

export interface CapsuleDescriptor {
  version: string;
  runtime_version: string;
  memory_limit: number;
  instruction_limit: number;
  float_mode: string;
  determinism_flags: string[];
}

export interface ReplayModalProps {
  /** Whether the modal is open */
  open: boolean;
  /** Callback when open state changes */
  onOpenChange: (open: boolean) => void;
  /** Capsule descriptor */
  capsule?: CapsuleDescriptor;
  /** Available replay nodes */
  nodes?: { id: string; name: string; region: string; available: boolean }[];
  /** Estimated cost */
  costEstimate?: { amount: number; currency: string };
  /** Warning messages */
  warnings?: string[];
  /** Selected node ID */
  selectedNode?: string;
  /** Callback when node is selected */
  onNodeSelect?: (nodeId: string) => void;
  /** Callback when replay starts */
  onStartReplay: (mode: ReplayMode, nodeId: string) => void;
  /** Is loading */
  loading?: boolean;
}

export function ReplayModal({
  open,
  onOpenChange,
  capsule,
  nodes = [],
  costEstimate,
  warnings = [],
  selectedNode,
  onNodeSelect,
  onStartReplay,
  loading = false,
}: ReplayModalProps) {
  const [mode, setMode] = useState<ReplayMode>("strict");
  const [selectedNodeId, setSelectedNodeId] = useState(selectedNode || nodes[0]?.id || "");

  const handleStart = () => {
    onStartReplay(mode, selectedNodeId);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Replay Execution</DialogTitle>
          <DialogDescription>
            Re-run this execution in a deterministic sandbox to verify results.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          {/* Capsule Info */}
          {capsule && (
            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-sm">Capsule Descriptor</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Version</span>
                  <Badge variant="outline">{capsule.version}</Badge>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Runtime</span>
                  <span>{capsule.runtime_version}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Memory Limit</span>
                  <span>{capsule.memory_limit} MB</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Float Mode</span>
                  <Badge variant="secondary">{capsule.float_mode}</Badge>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Replay Mode */}
          <div className="space-y-3">
            <Label>Replay Mode</Label>
            <RadioGroup
              value={mode}
              onValueChange={(v) => setMode(v as ReplayMode)}
              className="flex gap-4"
            >
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="strict" id="strict" />
                <Label htmlFor="strict" className="cursor-pointer">Strict</Label>
              </div>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="lite" id="lite" />
                <Label htmlFor="lite" className="cursor-pointer">Lite</Label>
              </div>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="debug" id="debug" />
                <Label htmlFor="debug" className="cursor-pointer">Debug</Label>
              </div>
            </RadioGroup>
          </div>

          {/* Node Selection */}
          {nodes.length > 0 && (
            <div className="space-y-3">
              <Label>Select Replay Node</Label>
              <div className="grid gap-2">
                {nodes.map((node) => (
                  <button
                    key={node.id}
                    onClick={() => {
                      setSelectedNodeId(node.id);
                      onNodeSelect?.(node.id);
                    }}
                    className={cn(
                      "flex items-center justify-between p-3 rounded-lg border transition-colors",
                      selectedNodeId === node.id
                        ? "border-brand-500 bg-brand-500/10"
                        : "border-border-subtle hover:bg-bg-secondary",
                      !node.available && "opacity-50 cursor-not-allowed"
                    )}
                    disabled={!node.available}
                  >
                    <div className="flex items-center gap-3">
                      <Server className="h-4 w-4 text-muted-foreground" />
                      <div className="text-left">
                        <div className="font-medium text-sm">{node.name}</div>
                        <div className="text-xs text-muted-foreground">{node.region}</div>
                      </div>
                    </div>
                    {node.available ? (
                      <CheckCircle className="h-4 w-4 text-green-500" />
                    ) : (
                      <span className="text-xs text-muted-foreground">Unavailable</span>
                    )}
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Warnings */}
          {warnings.length > 0 && (
            <div className="space-y-2">
              {warnings.map((warning, i) => (
                <div
                  key={i}
                  className="flex items-start gap-2 p-3 bg-yellow-500/10 border border-yellow-500/20 rounded-md"
                >
                  <AlertTriangle className="h-4 w-4 text-yellow-500 shrink-0 mt-0.5" />
                  <p className="text-xs text-yellow-500">{warning}</p>
                </div>
              ))}
            </div>
          )}

          {/* Cost Estimate */}
          {costEstimate && (
            <div className="flex items-center justify-between p-3 bg-bg-secondary rounded-lg">
              <div className="flex items-center gap-2">
                <DollarSign className="h-4 w-4 text-muted-foreground" />
                <span className="text-sm">Estimated Cost</span>
              </div>
              <span className="font-medium">
                {costEstimate.currency} {costEstimate.amount.toFixed(4)}
              </span>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleStart} disabled={loading || !selectedNodeId}>
            Start Replay
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
