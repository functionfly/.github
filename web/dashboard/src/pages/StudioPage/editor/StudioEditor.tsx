import React, { useCallback, useEffect } from "react";
import { SemanticCodeEditor } from "@functionfly/ui-editor";
import { Badge, Tooltip, cn } from "@functionfly/ui-core";
import {
  Play,
  Pause,
  Brain,
  LineChart,
  Save,
  Undo2,
  Redo2,
  Code,
  Download,
  Upload,
  Settings,
  ChevronDown,
} from "lucide-react";

interface StudioEditorProps {
  code: string;
  onChange: (code: string) => void;
  theme: "studio-dark" | "studio-light" | "monokai" | "github-dark";
  onThemeChange: (theme: "studio-dark" | "studio-light" | "monokai" | "github-dark") => void;
  isGhostModeActive: boolean;
  onToggleGhostMode: () => void;
  onRunWorkflow: () => void;
  onFormatCode: () => void;
  onSave: () => void;
  onUndo: () => void;
  onRedo: () => void;
  onCommandPalette: () => void;
}

const THEMES = [
  { value: "studio-dark", label: "Studio Dark" },
  { value: "studio-light", label: "Studio Light" },
  { value: "monokai", label: "Monokai" },
  { value: "github-dark", label: "GitHub Dark" },
] as const;

export function StudioEditor({
  code,
  onChange,
  theme,
  onThemeChange,
  isGhostModeActive,
  onToggleGhostMode,
  onRunWorkflow,
  onFormatCode,
  onSave,
  onUndo,
  onRedo,
  onCommandPalette,
}: StudioEditorProps) {
  const [isSaved, setIsSaved] = React.useState(true);
  const [lastSaved, setLastSaved] = React.useState<Date | null>(null);

  const handleSave = useCallback(() => {
    onSave();
    setIsSaved(true);
    setLastSaved(new Date());
  }, [onSave]);

  useEffect(() => {
    setIsSaved(false);
  }, [code]);

  return (
    <div className="h-full flex flex-col bg-bg-primary">
      <div className="flex items-center justify-between px-4 py-2 border-b border-border-subtle bg-bg-secondary">
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2">
            <Code className="size-4 text-brand-400" />
            <span className="text-sm font-medium">main.ts</span>
            <Badge
              variant={isSaved ? "success" : "warning"}
              size="sm"
              className="text-[10px]"
            >
              {isSaved ? "Saved" : "Unsaved"}
            </Badge>
          </div>
          <div className="flex items-center gap-2">
            <div className="flex items-center gap-1 text-[10px] text-text-muted">
              <span>UTF-8</span>
              <span>•</span>
              <span>TypeScript</span>
              <span>•</span>
              <span>LF</span>
            </div>
            <select
              value={theme}
              onChange={(e) => onThemeChange(e.target.value as typeof theme)}
              className="text-[10px] bg-bg-primary border border-border-subtle rounded px-1.5 py-0.5 text-text-muted hover:text-text-primary cursor-pointer hover:border-brand-500/50 transition-colors"
            >
              {THEMES.map((t) => (
                <option key={t.value} value={t.value}>
                  {t.label}
                </option>
              ))}
            </select>
          </div>
        </div>
        <div className="flex items-center gap-1">
          <Tooltip content="Undo (Cmd+Z)">
            <button
              onClick={onUndo}
              className="p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-bg-hover transition-colors"
            >
              <Undo2 className="size-4" />
            </button>
          </Tooltip>
          <Tooltip content="Redo (Cmd+Shift+Z)">
            <button
              onClick={onRedo}
              className="p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-bg-hover transition-colors"
            >
              <Redo2 className="size-4" />
            </button>
          </Tooltip>

          {isGhostModeActive && (
            <Tooltip content="Stop Ghost Mode">
              <button
                onClick={onToggleGhostMode}
                className="flex items-center gap-1.5 px-2 py-1 text-[11px] rounded-md bg-warning/20 text-warning hover:bg-warning/30 border border-warning/30 transition-colors"
              >
                <span>Ghost</span>
              </button>
            </Tooltip>
          )}

          <Tooltip content="AI Assist (Cmd+K)">
            <button
              onClick={onCommandPalette}
              className="flex items-center gap-1.5 px-2 py-1 text-[11px] bg-brand-500/20 text-brand-400 rounded-md hover:bg-brand-500/30 transition-colors"
            >
              <Brain className="size-3" />
              <span>AI</span>
            </button>
          </Tooltip>

          <Tooltip content="Format (Cmd+Shift+F)">
            <button
              onClick={onFormatCode}
              className="p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-bg-hover transition-colors"
            >
              <LineChart className="size-4" />
            </button>
          </Tooltip>

          <Tooltip content="Save (Cmd+S)">
            <button
              onClick={handleSave}
              className={cn(
                "p-1.5 rounded-md transition-colors",
                isSaved
                  ? "text-text-muted hover:text-text-primary hover:bg-bg-hover"
                  : "bg-success/20 text-success hover:bg-success/30"
              )}
            >
              <Save className="size-4" />
            </button>
          </Tooltip>

          <Tooltip content="Run (Cmd+Enter)">
            <button
              onClick={onRunWorkflow}
              className="p-1.5 rounded-md bg-success/20 text-success hover:bg-success/30 transition-colors"
            >
              <Play className="size-4" />
            </button>
          </Tooltip>
        </div>
      </div>

      <div className="flex-1 overflow-hidden">
        <SemanticCodeEditor
          value={code}
          onChange={onChange}
          language="typescript"
          theme={theme}
          readOnly={false}
          showLineNumbers={true}
          showMinimap={true}
          fontSize={13}
          className="h-full"
        />
      </div>
    </div>
  );
}