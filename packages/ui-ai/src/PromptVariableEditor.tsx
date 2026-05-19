/**
 * @functionfly/ui-ai
 * Prompt Variable Editor - Edit and manage prompt variables
 */

import * as React from "react";
import { cn } from "@functionfly/ui-core";
import { Badge } from "@functionfly/ui-core";
import { Variable, Plus, Trash2, Copy, GripVertical, ChevronDown, ChevronRight } from "lucide-react";

export interface Variable {
  id: string;
  name: string;
  value: string;
  type: "string" | "number" | "boolean" | "object" | "array";
  defaultValue?: string;
  description?: string;
  isRequired?: boolean;
  allowedValues?: string[];
}

export interface PromptVariableEditorProps {
  variables: Variable[];
  onVariableAdd?: (variable: Omit<Variable, "id">) => void;
  onVariableUpdate?: (id: string, updates: Partial<Variable>) => void;
  onVariableDelete?: (id: string) => void;
  onVariableReorder?: (fromIndex: number, toIndex: number) => void;
  className?: string;
}

const typeColors = {
  string: "bg-info/10 text-info",
  number: "bg-success/10 text-success",
  boolean: "bg-warning/10 text-warning",
  object: "bg-purple-500/10 text-purple-500",
  array: "bg-brand-500/10 text-brand-500",
};

const typeIcons: Record<string, string> = {
  string: "S",
  number: "#",
  boolean: "?",
  object: "{}",
  array: "[]",
};

export function PromptVariableEditor({
  variables,
  onVariableAdd,
  onVariableUpdate,
  onVariableDelete,
  className,
}: PromptVariableEditorProps) {
  const [expandedIds, setExpandedIds] = React.useState<Set<string>>(new Set());
  const [editingId, setEditingId] = React.useState<string | null>(null);

  const toggleExpand = (id: string) => {
    setExpandedIds(prev => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const handleStartEdit = (variable: Variable) => {
    setEditingId(variable.id);
  };

  const handleSaveEdit = (id: string, updates: Partial<Variable>) => {
    onVariableUpdate?.(id, updates);
    setEditingId(null);
  };

  const handleAddVariable = (type: Variable["type"] = "string") => {
    onVariableAdd?.({
      name: "",
      value: "",
      type,
      isRequired: false,
    });
  };

  return (
    <div className={cn("flex flex-col h-full", className)}>
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-border-subtle">
        <div className="flex items-center gap-2">
          <Variable className="size-4 text-brand-500" />
          <span className="text-sm font-medium text-text-primary">Variables</span>
          <Badge variant="brand" size="sm">{variables.length}</Badge>
        </div>
        <div className="flex items-center gap-1">
          <select
            onChange={e => handleAddVariable(e.target.value as Variable["type"])}
            className="px-2 py-1 text-xs bg-bg-tertiary border border-border-subtle rounded text-text-primary"
            value=""
          >
            <option value="" disabled>Add variable</option>
            <option value="string">String</option>
            <option value="number">Number</option>
            <option value="boolean">Boolean</option>
            <option value="object">Object</option>
            <option value="array">Array</option>
          </select>
        </div>
      </div>

      {/* Variable List */}
      <div className="flex-1 overflow-y-auto p-3 space-y-2">
        {variables.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 text-text-muted">
            <Variable className="size-12 mb-3 opacity-50" />
            <p className="text-sm">No variables defined</p>
            <p className="text-xs mt-1">Add variables to template your prompts</p>
          </div>
        ) : (
          variables.map(variable => {
            const isExpanded = expandedIds.has(variable.id);
            const isEditing = editingId === variable.id;

            return (
              <div
                key={variable.id}
                className={cn(
                  "rounded-lg border transition-colors",
                  variable.isRequired
                    ? "bg-warning-500/5 border-warning-500/20"
                    : "bg-bg-secondary border-border-subtle"
                )}
              >
                {/* Header */}
                <div
                  onClick={() => !isEditing && toggleExpand(variable.id)}
                  className="flex items-center gap-2 p-3 cursor-pointer hover:bg-bg-hover rounded-t-lg"
                >
                  <div className="cursor-grab text-text-muted hover:text-text-primary">
                    <GripVertical className="size-4" />
                  </div>

                  <span className={cn("px-1.5 py-0.5 rounded text-[10px] font-bold shrink-0", typeColors[variable.type])}>
                    {typeIcons[variable.type]}
                  </span>

                  <span className="flex-1 text-sm font-medium text-text-primary truncate">
                    {variable.name || <span className="text-text-muted italic">unnamed</span>}
                  </span>

                  {variable.isRequired && (
                    <Badge variant="error" size="sm">required</Badge>
                  )}

                  <span className="text-xs text-text-muted">
                    {variable.type}
                  </span>
                </div>

                {/* Expanded/Edit Content */}
                {isExpanded || isEditing ? (
                  <div className="px-3 pb-3 space-y-3">
                    {/* Name */}
                    <div className="space-y-1">
                      <label className="text-[10px] font-medium text-text-muted uppercase tracking-wide">Name</label>
                      {isEditing ? (
                        <input
                          type="text"
                          value={variable.name}
                          onChange={e => onVariableUpdate?.(variable.id, { name: e.target.value })}
                          className="w-full px-2 py-1.5 text-sm bg-bg-primary border border-border-subtle rounded text-text-primary"
                          placeholder="variable_name"
                        />
                      ) : (
                        <code className="text-sm text-brand-500">{`{${variable.name}}`}</code>
                      )}
                    </div>

                    {/* Value */}
                    <div className="space-y-1">
                      <label className="text-[10px] font-medium text-text-muted uppercase tracking-wide">Value</label>
                      {variable.type === "boolean" ? (
                        <select
                          value={variable.value}
                          onChange={e => onVariableUpdate?.(variable.id, { value: e.target.value })}
                          className="w-full px-2 py-1.5 text-sm bg-bg-primary border border-border-subtle rounded text-text-primary"
                        >
                          <option value="true">true</option>
                          <option value="false">false</option>
                        </select>
                      ) : variable.type === "number" ? (
                        <input
                          type="number"
                          value={variable.value}
                          onChange={e => onVariableUpdate?.(variable.id, { value: e.target.value })}
                          className="w-full px-2 py-1.5 text-sm bg-bg-primary border border-border-subtle rounded text-text-primary"
                        />
                      ) : (
                        <textarea
                          value={variable.value}
                          onChange={e => onVariableUpdate?.(variable.id, { value: e.target.value })}
                          rows={3}
                          className="w-full px-2 py-1.5 text-sm bg-bg-primary border border-border-subtle rounded text-text-primary resize-none font-mono"
                          placeholder="Enter value..."
                        />
                      )}
                    </div>

                    {/* Description */}
                    <div className="space-y-1">
                      <label className="text-[10px] font-medium text-text-muted uppercase tracking-wide">Description</label>
                      <input
                        type="text"
                        value={variable.description || ""}
                        onChange={e => onVariableUpdate?.(variable.id, { description: e.target.value })}
                        className="w-full px-2 py-1.5 text-sm bg-bg-primary border border-border-subtle rounded text-text-primary"
                        placeholder="What is this variable for?"
                      />
                    </div>

                    {/* Options */}
                    <div className="flex items-center gap-4">
                      <label className="flex items-center gap-2 cursor-pointer">
                        <input
                          type="checkbox"
                          checked={variable.isRequired || false}
                          onChange={e => onVariableUpdate?.(variable.id, { isRequired: e.target.checked })}
                          className="size-4 rounded border-border-subtle"
                        />
                        <span className="text-xs text-text-secondary">Required</span>
                      </label>
                    </div>

                    {/* Actions */}
                    <div className="flex items-center gap-2 pt-2">
                      {isEditing ? (
                        <>
                          <button
                            onClick={() => handleSaveEdit(variable.id, variable)}
                            className="px-3 py-1.5 text-xs bg-brand-500 text-white rounded hover:bg-brand-600 transition-colors"
                          >
                            Save
                          </button>
                          <button
                            onClick={() => setEditingId(null)}
                            className="px-3 py-1.5 text-xs bg-bg-tertiary text-text-secondary rounded hover:bg-bg-hover transition-colors"
                          >
                            Cancel
                          </button>
                        </>
                      ) : (
                        <>
                          <button
                            onClick={() => handleStartEdit(variable)}
                            className="px-3 py-1.5 text-xs bg-bg-tertiary text-text-secondary rounded hover:bg-bg-hover transition-colors"
                          >
                            Edit
                          </button>
                          <button
                            onClick={() => onVariableDelete?.(variable.id)}
                            className="px-3 py-1.5 text-xs bg-error/10 text-error rounded hover:bg-error/20 transition-colors ml-auto"
                          >
                            <Trash2 className="size-3" />
                          </button>
                        </>
                      )}
                    </div>
                  </div>
                ) : null}
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
