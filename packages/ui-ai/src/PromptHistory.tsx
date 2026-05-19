/**
 * @functionfly/ui-ai
 * Prompt History Component
 */

import * as React from "react";
import { cn } from "@functionfly/ui-core";
import { Badge } from "@functionfly/ui-core";
import { Clock, Search, Copy, Trash2, ArrowUpRight } from "lucide-react";

export interface PromptHistoryItem {
  id: string;
  prompt: string;
  timestamp: number;
  tokensUsed?: number;
  executionTime?: number;
  success: boolean;
  agentName?: string;
}

export interface PromptHistoryProps {
  items: PromptHistoryItem[];
  onSelect?: (item: PromptHistoryItem) => void;
  onDelete?: (id: string) => void;
  onCopy?: (item: PromptHistoryItem) => void;
  maxItems?: number;
  className?: string;
}

export function PromptHistory({
  items,
  onSelect,
  onDelete,
  onCopy,
  maxItems = 50,
  className,
}: PromptHistoryProps) {
  const [searchQuery, setSearchQuery] = React.useState("");
  const [selectedId, setSelectedId] = React.useState<string | null>(null);

  const filteredItems = React.useMemo(() => {
    const sorted = [...items].sort((a, b) => b.timestamp - a.timestamp);
    if (!searchQuery.trim()) return sorted.slice(0, maxItems);
    
    const q = searchQuery.toLowerCase();
    return sorted
      .filter(item => item.prompt.toLowerCase().includes(q))
      .slice(0, maxItems);
  }, [items, searchQuery, maxItems]);

  const handleSelect = (item: PromptHistoryItem) => {
    setSelectedId(item.id);
    onSelect?.(item);
  };

  return (
    <div className={cn("flex flex-col h-full", className)}>
      {/* Search */}
      <div className="flex items-center gap-2 px-3 py-2 border-b border-border-subtle">
        <Search className="size-4 text-text-muted shrink-0" />
        <input
          type="text"
          value={searchQuery}
          onChange={e => setSearchQuery(e.target.value)}
          placeholder="Search history..."
          className="flex-1 bg-transparent text-sm text-text-primary outline-none placeholder:text-text-muted"
        />
        {searchQuery && (
          <Badge variant="ghost" size="sm">
            {filteredItems.length} results
          </Badge>
        )}
      </div>

      {/* List */}
      <div className="flex-1 overflow-y-auto">
        {filteredItems.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 text-text-muted">
            <Clock className="size-12 mb-3 opacity-50" />
            <p className="text-sm">No prompts in history</p>
            <p className="text-xs mt-1">Your prompt history will appear here</p>
          </div>
        ) : (
          <div className="p-2 space-y-1">
            {filteredItems.map(item => (
              <div
                key={item.id}
                onClick={() => handleSelect(item)}
                className={cn(
                  "group relative p-3 rounded-lg cursor-pointer transition-colors",
                  selectedId === item.id
                    ? "bg-brand-500/10 border border-brand-500/30"
                    : "hover:bg-bg-hover border border-transparent"
                )}
              >
                {/* Timestamp */}
                <div className="flex items-center gap-2 mb-1">
                  <Clock className="size-3 text-text-muted" />
                  <span className="text-[10px] text-text-muted">
                    {new Date(item.timestamp).toLocaleString()}
                  </span>
                  {item.success ? (
                    <Badge variant="success" size="sm">Success</Badge>
                  ) : (
                    <Badge variant="error" size="sm">Failed</Badge>
                  )}
                </div>

                {/* Prompt Preview */}
                <p className="text-sm text-text-primary line-clamp-2 pr-16">
                  {item.prompt}
                </p>

                {/* Stats */}
                <div className="flex items-center gap-3 mt-2 text-[10px] text-text-muted">
                  {item.tokensUsed && (
                    <span>{item.tokensUsed} tokens</span>
                  )}
                  {item.executionTime && (
                    <span>{item.executionTime}ms</span>
                  )}
                  {item.agentName && (
                    <span>{item.agentName}</span>
                  )}
                </div>

                {/* Actions */}
                <div className="absolute top-2 right-2 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  <button
                    onClick={e => { e.stopPropagation(); onCopy?.(item); }}
                    className="p-1.5 rounded hover:bg-bg-tertiary text-text-muted hover:text-text-primary"
                  >
                    <Copy className="size-3" />
                  </button>
                  <button
                    onClick={e => { e.stopPropagation(); onDelete?.(item.id); }}
                    className="p-1.5 rounded hover:bg-bg-tertiary text-text-muted hover:text-error"
                  >
                    <Trash2 className="size-3" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Footer */}
      <div className="px-3 py-2 border-t border-border-subtle text-[10px] text-text-muted flex items-center justify-between">
        <span>{items.length} total prompts</span>
        <span>Click to reuse</span>
      </div>
    </div>
  );
}
