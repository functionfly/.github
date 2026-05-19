/**
 * @functionfly/ui-ai
 * Context Injector - Manage context for AI interactions
 */

import * as React from "react";
import { cn } from "@functionfly/ui-core";
import { Badge } from "@functionfly/ui-core";
import { Database, Plus, Trash2, ChevronDown, ChevronRight, Copy, Search, FileText } from "lucide-react";

export interface ContextEntry {
  id: string;
  type: "file" | "variable" | "function" | "class" | "api" | "database" | "custom";
  name: string;
  content: string;
  metadata?: Record<string, string>;
  isActive: boolean;
  size?: number;
}

export interface ContextInjectorProps {
  entries: ContextEntry[];
  onEntryAdd?: (entry: Omit<ContextEntry, "id">) => void;
  onEntryRemove?: (id: string) => void;
  onEntryToggle?: (id: string) => void;
  onEntryUpdate?: (id: string, updates: Partial<ContextEntry>) => void;
  maxEntries?: number;
  className?: string;
}

const typeIcons: Record<string, string> = {
  file: "F",
  variable: "V",
  function: "fn",
  class: "C",
  api: "API",
  database: "DB",
  custom: "•",
};

const typeColors: Record<string, string> = {
  file: "bg-info/10 text-info",
  variable: "bg-success/10 text-success",
  function: "bg-warning/10 text-warning",
  class: "bg-purple-500/10 text-purple-500",
  api: "bg-brand-500/10 text-brand-500",
  database: "bg-orange-500/10 text-orange-500",
  custom: "bg-bg-tertiary text-text-muted",
};

export function ContextInjector({
  entries,
  onEntryAdd,
  onEntryRemove,
  onEntryToggle,
  maxEntries = 10,
  className,
}: ContextInjectorProps) {
  const [searchQuery, setSearchQuery] = React.useState("");
  const [expandedIds, setExpandedIds] = React.useState<Set<string>>(new Set());
  const [showAddMenu, setShowAddMenu] = React.useState(false);

  const toggleExpand = (id: string) => {
    setExpandedIds(prev => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const filteredEntries = React.useMemo(() => {
    if (!searchQuery) return entries;
    const q = searchQuery.toLowerCase();
    return entries.filter(e =>
      e.name.toLowerCase().includes(q) ||
      e.content.toLowerCase().includes(q)
    );
  }, [entries, searchQuery]);

  const activeEntries = filteredEntries.filter(e => e.isActive);
  const totalSize = activeEntries.reduce((acc, e) => acc + (e.size || 0), 0);

  return (
    <div className={cn("flex flex-col h-full", className)}>
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-border-subtle">
        <div className="flex items-center gap-2">
          <Database className="size-4 text-brand-500" />
          <span className="text-sm font-medium text-text-primary">Context</span>
          <Badge variant="brand" size="sm">{activeEntries.length}/{maxEntries}</Badge>
        </div>
        <div className="relative">
          <button
            onClick={() => setShowAddMenu(s => !s)}
            disabled={activeEntries.length >= maxEntries}
            className="p-1.5 rounded hover:bg-bg-tertiary text-text-muted hover:text-text-primary disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Plus className="size-4" />
          </button>
          
          {showAddMenu && (
            <div className="absolute right-0 top-full mt-1 w-40 bg-bg-secondary border border-border-subtle rounded-lg shadow-lg z-10">
              {(["file", "variable", "function", "api"] as const).map(type => (
                <button
                  key={type}
                  onClick={() => {
                    onEntryAdd?.({ type, name: "", content: "", isActive: true });
                    setShowAddMenu(false);
                  }}
                  className="w-full flex items-center gap-2 px-3 py-2 text-xs text-text-secondary hover:bg-bg-hover rounded-t-lg"
                >
                  <span className={cn("px-1.5 py-0.5 rounded text-[10px] font-bold", typeColors[type])}>
                    {typeIcons[type]}
                  </span>
                  Add {type}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Search */}
      <div className="px-3 py-2 border-b border-border-subtle">
        <div className="flex items-center gap-2 px-2 py-1.5 bg-bg-tertiary/50 rounded-lg">
          <Search className="size-3 text-text-muted" />
          <input
            type="text"
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            placeholder="Search context..."
            className="flex-1 bg-transparent text-xs text-text-primary outline-none placeholder:text-text-muted"
          />
        </div>
      </div>

      {/* Context List */}
      <div className="flex-1 overflow-y-auto">
        {filteredEntries.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 text-text-muted">
            <Database className="size-12 mb-3 opacity-50" />
            <p className="text-sm">No context entries</p>
            <p className="text-xs mt-1">Add context to enhance AI responses</p>
          </div>
        ) : (
          <div className="p-2 space-y-1">
            {filteredEntries.map(entry => {
              const isExpanded = expandedIds.has(entry.id);
              
              return (
                <div
                  key={entry.id}
                  className={cn(
                    "rounded-lg border transition-colors",
                    entry.isActive
                      ? "bg-bg-secondary border-border-subtle"
                      : "bg-bg-tertiary/50 border-transparent opacity-60"
                  )}
                >
                  {/* Entry Header */}
                  <div
                    className="flex items-center gap-2 p-2 cursor-pointer hover:bg-bg-hover rounded-t-lg"
                    onClick={() => toggleExpand(entry.id)}
                  >
                    <button className="text-text-muted">
                      {isExpanded ? <ChevronDown className="size-3" /> : <ChevronRight className="size-3" />}
                    </button>
                    
                    <span className={cn("px-1.5 py-0.5 rounded text-[10px] font-bold shrink-0", typeColors[entry.type])}>
                      {typeIcons[entry.type]}
                    </span>
                    
                    <span className="flex-1 text-sm text-text-primary truncate">{entry.name}</span>
                    
                    {entry.size && (
                      <span className="text-[10px] text-text-muted">{entry.size} chars</span>
                    )}
                    
                    <button
                      onClick={e => { e.stopPropagation(); onEntryToggle?.(entry.id); }}
                      className={cn(
                        "w-8 h-4 rounded-full transition-colors relative",
                        entry.isActive ? "bg-brand-500" : "bg-bg-tertiary"
                      )}
                    >
                      <div className={cn(
                        "absolute top-0.5 w-3 h-3 rounded-full bg-white transition-transform",
                        entry.isActive ? "left-4" : "left-0.5"
                      )} />
                    </button>
                  </div>

                  {/* Expanded Content */}
                  {isExpanded && (
                    <div className="px-3 pb-2">
                      <pre className="p-2 bg-bg-tertiary/50 rounded text-xs font-mono text-text-secondary whitespace-pre-wrap max-h-40 overflow-auto">
                        {entry.content}
                      </pre>
                      
                      {/* Metadata */}
                      {entry.metadata && (
                        <div className="flex flex-wrap gap-2 mt-2">
                          {Object.entries(entry.metadata).map(([key, value]) => (
                            <span key={key} className="text-[10px] text-text-muted">
                              <span className="font-medium">{key}:</span> {value}
                            </span>
                          ))}
                        </div>
                      )}
                      
                      {/* Actions */}
                      <div className="flex items-center gap-1 mt-2">
                        <button
                          onClick={() => navigator.clipboard.writeText(entry.content)}
                          className="p-1 rounded hover:bg-bg-tertiary text-text-muted hover:text-text-primary"
                        >
                          <Copy className="size-3" />
                        </button>
                        <button
                          onClick={() => onEntryRemove?.(entry.id)}
                          className="p-1 rounded hover:bg-bg-tertiary text-text-muted hover:text-error"
                        >
                          <Trash2 className="size-3" />
                        </button>
                      </div>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Footer Stats */}
      <div className="px-4 py-2 border-t border-border-subtle text-[10px] text-text-muted flex items-center justify-between">
        <span>{activeEntries.length} active entries</span>
        <span>~{totalSize.toLocaleString()} chars</span>
      </div>
    </div>
  );
}
