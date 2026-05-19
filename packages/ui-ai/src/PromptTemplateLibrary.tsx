/**
 * @functionfly/ui-ai
 * Prompt Template Library - Browse and manage prompt templates
 */

import * as React from "react";
import { cn } from "@functionfly/ui-core";
import { Badge } from "@functionfly/ui-core";
import { FileText, Search, Copy, Plus, Edit2, Trash2, FolderOpen, Tag } from "lucide-react";

export interface PromptTemplate {
  id: string;
  name: string;
  description: string;
  template: string;
  variables: string[];
  category: string;
  tags: string[];
  createdAt: number;
  updatedAt: number;
  usageCount: number;
}

export interface PromptTemplateLibraryProps {
  templates: PromptTemplate[];
  onTemplateSelect?: (template: PromptTemplate) => void;
  onTemplateCreate?: () => void;
  onTemplateEdit?: (template: PromptTemplate) => void;
  onTemplateDelete?: (id: string) => void;
  onTemplateDuplicate?: (template: PromptTemplate) => void;
  className?: string;
}

const categoryColors: Record<string, string> = {
  general: "bg-info/10 text-info",
  coding: "bg-success/10 text-success",
  analysis: "bg-warning/10 text-warning",
  creative: "bg-brand-500/10 text-brand-500",
  automation: "bg-purple-500/10 text-purple-500",
};

export function PromptTemplateLibrary({
  templates,
  onTemplateSelect,
  onTemplateCreate,
  onTemplateEdit,
  onTemplateDelete,
  onTemplateDuplicate,
  className,
}: PromptTemplateLibraryProps) {
  const [searchQuery, setSearchQuery] = React.useState("");
  const [selectedCategory, setSelectedCategory] = React.useState<string | null>(null);
  const [selectedTemplate, setSelectedTemplate] = React.useState<PromptTemplate | null>(null);

  const categories = React.useMemo(() => {
    const cats = new Set(templates.map(t => t.category));
    return Array.from(cats);
  }, [templates]);

  const filteredTemplates = React.useMemo(() => {
    return templates.filter(t => {
      const matchesSearch = !searchQuery || 
        t.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        t.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
        t.tags.some(tag => tag.toLowerCase().includes(searchQuery.toLowerCase()));
      const matchesCategory = !selectedCategory || t.category === selectedCategory;
      return matchesSearch && matchesCategory;
    });
  }, [templates, searchQuery, selectedCategory]);

  const handleSelect = (template: PromptTemplate) => {
    setSelectedTemplate(template);
    onTemplateSelect?.(template);
  };

  return (
    <div className={cn("flex h-full", className)}>
      {/* Template List */}
      <div className="w-80 border-r border-border-subtle flex flex-col">
        {/* Search & Filters */}
        <div className="p-3 border-b border-border-subtle space-y-2">
          <div className="flex items-center gap-2">
            <Search className="size-4 text-text-muted shrink-0" />
            <input
              type="text"
              value={searchQuery}
              onChange={e => setSearchQuery(e.target.value)}
              placeholder="Search templates..."
              className="flex-1 bg-transparent text-sm text-text-primary outline-none placeholder:text-text-muted"
            />
          </div>

          {/* Category Filters */}
          <div className="flex items-center gap-1 flex-wrap">
            <button
              onClick={() => setSelectedCategory(null)}
              className={cn(
                "px-2 py-1 text-[10px] rounded-full transition-colors",
                !selectedCategory
                  ? "bg-brand-500 text-white"
                  : "bg-bg-tertiary text-text-muted hover:text-text-primary"
              )}
            >
              All
            </button>
            {categories.map(cat => (
              <button
                key={cat}
                onClick={() => setSelectedCategory(cat === selectedCategory ? null : cat)}
                className={cn(
                  "px-2 py-1 text-[10px] rounded-full transition-colors",
                  selectedCategory === cat
                    ? "bg-brand-500 text-white"
                    : "bg-bg-tertiary text-text-muted hover:text-text-primary"
                )}
              >
                {cat}
              </button>
            ))}
          </div>
        </div>

        {/* Template List */}
        <div className="flex-1 overflow-y-auto">
          {filteredTemplates.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-text-muted">
              <FileText className="size-12 mb-3 opacity-50" />
              <p className="text-sm">No templates found</p>
            </div>
          ) : (
            <div className="p-2 space-y-1">
              {filteredTemplates.map(template => (
                <div
                  key={template.id}
                  onClick={() => handleSelect(template)}
                  className={cn(
                    "group p-3 rounded-lg cursor-pointer transition-colors",
                    selectedTemplate?.id === template.id
                      ? "bg-brand-500/10 border border-brand-500/20"
                      : "hover:bg-bg-hover border border-transparent"
                  )}
                >
                  <div className="flex items-start justify-between gap-2">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-medium text-text-primary">{template.name}</span>
                      </div>
                      <p className="text-xs text-text-muted mt-0.5 line-clamp-2">{template.description}</p>
                      <div className="flex items-center gap-2 mt-2">
                        <Badge
                          className={cn("text-[10px]", categoryColors[template.category] || "bg-bg-tertiary text-text-muted")}
                          variant="outline"
                          size="sm"
                        >
                          {template.category}
                        </Badge>
                        <span className="text-[10px] text-text-muted">{template.usageCount} uses</span>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Create Button */}
        <div className="p-3 border-t border-border-subtle">
          <button
            onClick={onTemplateCreate}
            className="w-full flex items-center justify-center gap-2 px-3 py-2 bg-brand-500 hover:bg-brand-600 text-white rounded-lg transition-colors text-sm"
          >
            <Plus className="size-4" />
            New Template
          </button>
        </div>
      </div>

      {/* Template Preview */}
      <div className="flex-1 flex flex-col">
        {selectedTemplate ? (
          <>
            {/* Header */}
            <div className="flex items-center justify-between px-4 py-3 border-b border-border-subtle">
              <div className="flex items-center gap-2">
                <FileText className="size-4 text-brand-500" />
                <span className="text-sm font-medium text-text-primary">{selectedTemplate.name}</span>
              </div>
              <div className="flex items-center gap-1">
                <button
                  onClick={() => onTemplateDuplicate?.(selectedTemplate)}
                  className="p-1.5 rounded hover:bg-bg-tertiary text-text-muted hover:text-text-primary"
                >
                  <Copy className="size-4" />
                </button>
                <button
                  onClick={() => onTemplateEdit?.(selectedTemplate)}
                  className="p-1.5 rounded hover:bg-bg-tertiary text-text-muted hover:text-text-primary"
                >
                  <Edit2 className="size-4" />
                </button>
                <button
                  onClick={() => onTemplateDelete?.(selectedTemplate.id)}
                  className="p-1.5 rounded hover:bg-bg-tertiary text-text-muted hover:text-error"
                >
                  <Trash2 className="size-4" />
                </button>
              </div>
            </div>

            {/* Content */}
            <div className="flex-1 overflow-y-auto p-4 space-y-4">
              <p className="text-sm text-text-secondary">{selectedTemplate.description}</p>

              {/* Template Preview */}
              <div className="space-y-2">
                <label className="text-[10px] font-medium text-text-muted uppercase tracking-wide">
                  Template
                </label>
                <pre className="p-4 bg-bg-tertiary/50 rounded-lg text-sm font-mono text-text-primary whitespace-pre-wrap border border-border-subtle">
                  {selectedTemplate.template}
                </pre>
              </div>

              {/* Variables */}
              <div className="space-y-2">
                <label className="text-[10px] font-medium text-text-muted uppercase tracking-wide">
                  Variables ({selectedTemplate.variables.length})
                </label>
                <div className="flex flex-wrap gap-2">
                  {selectedTemplate.variables.map(v => (
                    <code
                      key={v}
                      className="px-2 py-1 bg-bg-tertiary rounded text-xs font-mono text-brand-500"
                    >
                      {`{${v}}`}
                    </code>
                  ))}
                </div>
              </div>

              {/* Tags */}
              <div className="flex items-center gap-2 flex-wrap">
                <Tag className="size-3 text-text-muted" />
                {selectedTemplate.tags.map(tag => (
                  <Badge key={tag} variant="ghost" size="sm">{tag}</Badge>
                ))}
              </div>
            </div>
          </>
        ) : (
          <div className="flex-1 flex flex-col items-center justify-center text-text-muted">
            <FolderOpen className="size-12 mb-3 opacity-50" />
            <p className="text-sm">Select a template to preview</p>
          </div>
        )}
      </div>
    </div>
  );
}
