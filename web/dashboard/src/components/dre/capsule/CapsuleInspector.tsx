import { Box, Cpu, MemoryStick, Hash, Binary, Shield } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { DeterminismBadge } from "../primitives/DeterminismBadge";
import { CollapsibleSection } from "../primitives/CollapsibleSection";
import { cn } from "@/lib/utils";

export interface CapsuleDescriptor {
  runtime_version: string;
  memory_limit: number;
  instruction_limit: number;
  rng_seed: string;
  float_mode: "ieee754" | "deterministic";
  determinism_flags: string[];
}

export interface CapsuleInspectorProps {
  /** Capsule descriptor */
  capsule: CapsuleDescriptor;
  /** Custom className */
  className?: string;
}

export function CapsuleInspector({
  capsule,
  className,
}: CapsuleInspectorProps) {
  return (
    <Card className={className}>
      <CardHeader className="pb-3">
        <CardTitle className="text-base flex items-center gap-2">
          <Box className="h-4 w-4" />
          Capsule Inspector
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Runtime Info */}
        <div className="grid grid-cols-2 gap-4">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-bg-secondary rounded-lg">
              <Cpu className="h-4 w-4 text-muted-foreground" />
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Runtime</p>
              <p className="font-medium">{capsule.runtime_version}</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <div className="p-2 bg-bg-secondary rounded-lg">
              <MemoryStick className="h-4 w-4 text-muted-foreground" />
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Memory Limit</p>
              <p className="font-medium">{capsule.memory_limit} MB</p>
            </div>
          </div>
        </div>

        {/* Instruction Limit */}
        <div className="flex items-center gap-3 p-3 bg-bg-secondary rounded-lg">
          <Hash className="h-4 w-4 text-muted-foreground" />
          <div>
            <p className="text-xs text-muted-foreground">Instruction Limit</p>
            <p className="font-medium font-mono">
              {capsule.instruction_limit.toLocaleString()}
            </p>
          </div>
        </div>

        {/* Float Mode */}
        <div className="flex items-center gap-3 p-3 bg-bg-secondary rounded-lg">
          <Binary className="h-4 w-4 text-muted-foreground" />
          <div className="flex-1">
            <p className="text-xs text-muted-foreground">Float Mode</p>
            <Badge variant="secondary" className="mt-1">
              {capsule.float_mode === "deterministic" ? "Deterministic" : "IEEE754"}
            </Badge>
          </div>
        </div>

        {/* RNG Seed */}
        <CollapsibleSection
          title="RNG Seed"
          icon={<Hash className="h-4 w-4" />}
          defaultOpen={false}
        >
          <code className="block font-mono text-xs bg-bg-secondary p-2 rounded break-all">
            {capsule.rng_seed}
          </code>
        </CollapsibleSection>

        {/* Determinism Flags */}
        {capsule.determinism_flags && capsule.determinism_flags.length > 0 && (
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Shield className="h-4 w-4 text-muted-foreground" />
              <span className="text-sm font-medium">Determinism Flags</span>
            </div>
            <div className="flex flex-wrap gap-2">
              {capsule.determinism_flags.map((flag, i) => (
                <Badge key={i} variant="outline" className="font-mono text-xs">
                  {flag}
                </Badge>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
