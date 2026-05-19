/**
 * @functionfly/ui-ai
 * Agent Directive Editor - Define agent behavior and directives
 */

import * as React from "react";
import { cn } from "@functionfly/ui-core";
import { Badge } from "@functionfly/ui-core";
import { Target, Plus, Trash2, GripVertical, Copy, Settings } from "lucide-react";

export interface Directive {
  id: string;
  type: "system" | "constraint" | "objective" | "behavior" | "output";
  content: string;
  priority: "critical" | "high" | "medium" | "low";
  isEnabled: boolean;
}

export interface AgentDirectiveEditorProps {
  directives: Directive[];
  onDirectiveAdd?: (directive: Omit<Directive, "id">) => void;
  onDirectiveUpdate?: (id: string, updates: Partial<Directive>) => void;
  onDirectiveDelete?: (id: string) => void;
  onDirectiveReorder?: (fromIndex: number, toIndex: number) => void;
  className?: string;
}

const priorityColors = {
  critical: "bg-error/10 text-error border-error/20",
  high: "bg-warning/10 text-warning border-warning/20",
  medium: "bg-brand-500/10 text-brand-500 border-brand-500/20",
  low: "bg-bg-tertiary text-text-muted border-transparent",
};

const typeIcons = {
  system: "S",
  constraint: "C",
  objective: "O",
  behavior: "B",
  output: "O",
};

export function AgentDirectiveEditor({
  directives,
  onDirectiveAdd,
  onDirectiveUpdate,
  onDirectiveDelete,
  className,
}: AgentDirectiveEditorProps) {
  const [editingId, setEditingId] = React.useState<string | null>(null);
  const [editContent, setEditContent] = React.useState("");

  const handleStartEdit = (directive: Directive) => {
    setEditingId(directive.id);
    setEditContent(directive.content);
  };

  const handleSaveEdit = (id: string) => {
    onDirectiveUpdate?.(id, { content: editContent });
    setEditingId(null);
    setEditContent("");
  };

  const handleAddDirective = (type: Directive["type"]) => {
    onDirectiveAdd?.({
      type,
      content: "",
      priority: "medium",
      isEnabled: true,
    });
  };

  return (
    <div className={cn("flex flex-col h-full", className)}>
      {/* Header */}
      <div className="flex items-center gap-2 px-4 py-3 border-b border-border-subtle">
        <Target className="size-4 text-brand-500" />
        <span className="text-sm font-medium text-text-primary">Agent Directives</span>
        <Badge variant="ghost" size="sm">{directives.length} rules</Badge>
      </div>

      {/* Directive List */}
      <div className="flex-1 overflow-y-auto p-3 space-y-2">
        {directives.map((directive, index) => (
          <div
            key={directive.id}
            className={cn(
              "group p-3 rounded-lg border transition-colors",
              directive.isEnabled
                ? "bg-bg-secondary border-border-subtle"
                : "bg-bg-tertiary/50 border-transparent opacity-60"
            )}
          >
            <div className="flex items-start gap-3">
              {/* Drag Handle */}
              <div className="mt-1 cursor-grab text-text-muted hover:text-text-primary">
                <GripVertical className="size-4" />
              </div>

              {/* Type Badge */}
              <div className={cn(
                "size-6 rounded flex items-center justify-center text-[10px] font-bold shrink-0",
                directive.isEnabled ? "bg-brand-500/10 text-brand-500" : "bg-bg-tertiary text-text-muted"
              )}>
                {typeIcons[directive.type]}
              </div>

              {/* Content */}
              <div className="flex-1 min-w-0">
                {editingId === directive.id ? (
                  <div className="space-y-2">
                    <textarea
                      value={editContent}
                      onChange={e => setEditContent(e.target.value)}
                      className="w-full min-h-[80px] p-2 text-sm bg-bg-primary border border-border-subtle rounded-lg text-text-primary resize-none focus:outline-none focus:border-brand-500"
                      autoFocus
                    />
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => handleSaveEdit(directive.id)}
                        className="px-3 py-1 text-xs bg-brand-500 text-white rounded hover:bg-brand-600 transition-colors"
                      >
                        Save
                      </button>
                      <button
                        onClick={() => setEditingId(null)}
                        className="px-3 py-1 text-xs bg-bg-tertiary text-text-secondary rounded hover:bg-bg-hover transition-colors"
                      >
                        Cancel
                      </button>
                    </div>
                  </div>
                ) : (
                  <>
                    <p className="text-sm text-text-primary whitespace-pre-wrap">
                      {directive.content || <span className="text-text-muted italic">No content</span>}
                    </p>
                    <div className="flex items-center gap-2 mt-2">
                      <Badge
                        className={cn("text-[10px]", priorityColors[directive.priority])}
                        variant="outline"
                        size="sm"
                      >
                        {directive.priority}
                      </Badge>
                      <Badge variant="ghost" size="sm">{directive.type}</Badge>
                    </div>
                  </>
                )}
              </div>

              {/* Actions */}
              <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                <button
                  onClick={() => handleStartEdit(directive)}
                  className="p-1.5 rounded hover:bg-bg-tertiary text-text-muted hover:text-text-primary"
                >
                  <Settings className="size-3" />
                </button>
                <button
                  onClick={() => onDirectiveUpdate?.(directive.id, { isEnabled: !directive.isEnabled })}
                  className="p-1.5 rounded hover:bg-bg-tertiary text-text-muted hover:text-text-primary"
                >
                  <Copy className="size-3" />
                </button>
                <button
                  onClick={() => onDirectiveDelete?.(directive.id)}
                  className="p-1.5 rounded hover:bg-bg-tertiary text-text-muted hover:text-error"
                >
                  <Trash2 className="size-3" />
                </button>
              </div>
            </div>
          </div>
        ))}

        {/* Add Directive Buttons */}
        <div className="flex flex-wrap gap-2 pt-2">
          {(["system", "constraint", "objective", "behavior", "output"] as const).map(type => (
            <button
              key={type}
              onClick={() => handleAddDirective(type)}
              className="flex items-center gap-1.5 px-3 py-1.5 text-xs bg-bg-tertiary hover:bg-bg-hover rounded-lg text-text-muted hover:text-text-primary transition-colors"
            >
              <Plus className="size-3" />
              {type}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
