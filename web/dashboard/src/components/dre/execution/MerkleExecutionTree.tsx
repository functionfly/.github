import { useState } from "react";
import { ChevronRight, ChevronDown, Hash, FileInput, FileOutput, Settings, Package, GitBranch, HardDrive, FileText } from "lucide-react";
import { cn } from "@/lib/utils";
import { HashBlock } from "../primitives/HashBlock";
import { VerificationBadge } from "../primitives/VerificationBadge";

export type ComponentType = "input" | "output" | "environment" | "dependency" | "trace" | "resource" | "metadata";

export interface MerkleNodeData {
  type: ComponentType;
  hash: string;
  verified?: boolean;
  children?: MerkleNodeData[];
}

export interface MerkleExecutionTreeProps {
  /** The component hashes to display */
  hashes: Record<ComponentType, string>;
  /** Verification status for each component */
  verification?: Record<ComponentType, boolean>;
  /** Callback when a node is clicked */
  onNodeClick?: (type: ComponentType, hash: string) => void;
  /** Initial expanded nodes */
  defaultExpanded?: ComponentType[];
  /** Custom className */
  className?: string;
}

const componentConfig: Record<ComponentType, { label: string; icon: React.ElementType; description: string }> = {
  input: {
    label: "Input",
    icon: FileInput,
    description: "Function input parameters",
  },
  output: {
    label: "Output",
    icon: FileOutput,
    description: "Function output result",
  },
  environment: {
    label: "Environment",
    icon: Settings,
    description: "Environment variables and config",
  },
  dependency: {
    label: "Dependencies",
    icon: Package,
    description: "External dependencies and packages",
  },
  trace: {
    label: "Trace",
    icon: GitBranch,
    description: "Execution trace log",
  },
  resource: {
    label: "Resources",
    icon: HardDrive,
    description: "CPU, memory, and I/O usage",
  },
  metadata: {
    label: "Metadata",
    icon: FileText,
    description: "Execution metadata and timing",
  },
};

const componentOrder: ComponentType[] = [
  "input",
  "environment",
  "dependency",
  "trace",
  "resource",
  "output",
  "metadata",
];

interface TreeNodeProps {
  type: ComponentType;
  hash: string;
  verified?: boolean;
  onClick?: () => void;
  defaultExpanded?: boolean;
}

function TreeNode({ type, hash, verified, onClick, defaultExpanded = false }: TreeNodeProps) {
  const [expanded, setExpanded] = useState(defaultExpanded);
  const config = componentConfig[type];
  const Icon = config.icon;

  return (
    <div className="select-none">
      <div
        className={cn(
          "flex items-center gap-2 px-3 py-2 rounded-md cursor-pointer",
          "hover:bg-bg-secondary transition-colors group"
        )}
        onClick={() => setExpanded(!expanded)}
      >
        {expanded ? (
          <ChevronDown className="h-4 w-4 text-muted-foreground" />
        ) : (
          <ChevronRight className="h-4 w-4 text-muted-foreground" />
        )}
        <Icon className="h-4 w-4 text-muted-foreground group-hover:text-foreground transition-colors" />
        <span className="font-medium text-sm flex-1">{config.label}</span>
        <HashBlock
          hash={hash}
          truncate
          truncateChars={8}
          verified={verified}
          className="w-32"
        />
      </div>

      {expanded && (
        <div className="ml-8 pl-4 border-l border-border-subtle space-y-2 py-2">
          <div className="text-xs text-muted-foreground">
            {config.description}
          </div>
          {verified !== undefined && (
            <VerificationBadge
              status={verified ? "verified" : "pending"}
              size="sm"
            />
          )}
          <button
            className="text-xs text-blue-500 hover:underline"
            onClick={(e) => {
              e.stopPropagation();
              onClick?.();
            }}
          >
            View details →
          </button>
        </div>
      )}
    </div>
  );
}

export function MerkleExecutionTree({
  hashes,
  verification = {},
  onNodeClick,
  defaultExpanded = ["input", "output"],
  className,
}: MerkleExecutionTreeProps) {
  return (
    <div className={cn("space-y-2", className)}>
      <div className="flex items-center gap-2 mb-4">
        <Hash className="h-4 w-4 text-muted-foreground" />
        <span className="text-sm font-medium">Merkle Execution Tree</span>
      </div>

      <div className="space-y-1">
        {componentOrder.map((type) => (
          <TreeNode
            key={type}
            type={type}
            hash={hashes[type]}
            verified={verification[type]}
            defaultExpanded={defaultExpanded.includes(type)}
            onClick={() => onNodeClick?.(type, hashes[type])}
          />
        ))}
      </div>
    </div>
  );
}
