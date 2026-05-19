/**
 * @functionfly/ui-editor
 * SemanticCodeEditor - Monaco-based code editor with AI assistance
 */

import * as React from "react";
import Editor, { type OnMount, type OnChange, loader } from "@monaco-editor/react";
import type * as Monaco from "monaco-editor";
import { Sparkles } from "lucide-react";
import { cn } from "@functionfly/ui-core";

// Configure Monaco worker paths
loader.config({ paths: { vs: "https://cdn.jsdelivr.net/npm/monaco-editor@0.45.0/min/vs" } });

// Runtime to Monaco language mapping
const RUNTIME_MONACO_LANG: Record<string, string> = {
  typescript: "typescript",
  javascript: "javascript",
  python: "python",
  rust: "rust",
  go: "go",
  ruby: "ruby",
  java: "java",
  kotlin: "kotlin",
  bun: "javascript",
  deno: "javascript",
  nodejs: "javascript",
  wasm: "rust",
  sar: "typescript",
};

// Theme configuration
const THEME_CONFIG = {
  dark: {
    base: "vs-dark" as const,
    inherit: true,
    rules: [
      { token: "comment", foreground: "6b7280", fontStyle: "italic" },
      { token: "keyword", foreground: "f97316" },
      { token: "string", foreground: "10b981" },
      { token: "number", foreground: "3b82f6" },
      { token: "function", foreground: "fbbf24" },
      { token: "variable", foreground: "e5e7eb" },
      { token: "type", foreground: "a78bfa" },
    ],
    colors: {
      "editor.background": "#0a0a0f",
      "editor.foreground": "#e5e7eb",
      "editor.lineHighlightBackground": "#1a1a2e",
      "editor.selectionBackground": "#f9731633",
      "editorLineNumber.foreground": "#4b5563",
      "editorLineNumber.activeForeground": "#f97316",
      "editorCursor.foreground": "#f97316",
"editor.inactiveSelectionBackground": "#f9731622",
    },
  },
  "studio-dark": {
    base: "vs-dark" as const,
    inherit: true,
    rules: [
      { token: "comment", foreground: "6b7280", fontStyle: "italic" },
      { token: "keyword", foreground: "f97316" },
      { token: "string", foreground: "10b981" },
      { token: "number", foreground: "3b82f6" },
      { token: "function", foreground: "fbbf24" },
      { token: "variable", foreground: "e5e7eb" },
      { token: "type", foreground: "a78bfa" },
    ],
    colors: {
      "editor.background": "#0f0f14",
      "editor.foreground": "#e5e7eb",
      "editor.lineHighlightBackground": "#1a1a24",
      "editor.selectionBackground": "#f9731633",
      "editorLineNumber.foreground": "#4b5563",
      "editorLineNumber.activeForeground": "#f97316",
      "editorCursor.foreground": "#f97316",
      "editor.inactiveSelectionBackground": "#f9731622",
      "editorWidget.background": "#1a1a24",
      "editorWidget.border": "#2d2d3a",
      "input.background": "#1a1a24",
      "input.border": "#2d2d3a",
    },
  },
};

export interface SemanticCodeEditorProps {
  value: string;
  onChange?: (value: string) => void;
  language?: string;
  runtime?: string;
  theme?: "dark" | "light" | "studio-dark" | "studio-light" | "monokai" | "github-dark";
  readOnly?: boolean;
  height?: string | number;
  showLineNumbers?: boolean;
  showMinimap?: boolean;
  showSyntaxHighlighting?: boolean;
  fontSize?: number;
  lineHeight?: number;
  padding?: { top?: number; bottom?: number };
  onEditorMount?: (editor: Monaco.editor.IStandaloneCodeEditor, monaco: typeof Monaco) => void;
  onCursorChange?: (position: { lineNumber: number; column: number }) => void;
  className?: string;
  aiAssisted?: boolean;
}

export function SemanticCodeEditor({
  value,
  onChange,
  language = "typescript",
  runtime,
  theme = "dark",
  readOnly = false,
  height = "100%",
  showLineNumbers = true,
  showMinimap = false,
  fontSize = 13,
  lineHeight = 20,
  padding,
  onEditorMount,
  onCursorChange,
  className,
  aiAssisted = false,
}: SemanticCodeEditorProps) {
  const editorRef = React.useRef<Monaco.editor.IStandaloneCodeEditor | null>(null);
  const monacoRef = React.useRef<typeof Monaco | null>(null);

  // Resolve language from runtime if provided
  const resolvedLanguage = React.useMemo(() => {
    if (runtime && RUNTIME_MONACO_LANG[runtime]) {
      return RUNTIME_MONACO_LANG[runtime];
    }
    return language;
  }, [runtime, language]);

  // Handle editor mount
  const handleEditorWillMount: (monaco: typeof Monaco) => void = (monaco) => {
    monacoRef.current = monaco;
    // Register all themes
    Object.entries(THEME_CONFIG).forEach(([key, themeData]) => {
      monaco.editor.defineTheme(`functionfly-${key}`, themeData);
    });
  };

  const handleEditorDidMount: OnMount = (editor, monaco) => {
    editorRef.current = editor;
    monacoRef.current = monaco;

    monaco.editor.setTheme(`functionfly-${theme}`);

    // Set editor options
    editor.updateOptions({
      readOnly,
      fontSize,
      lineHeight,
      minimap: { enabled: showMinimap },
      lineNumbers: showLineNumbers ? "on" : "off",
      scrollBeyondLastLine: false,
      automaticLayout: true,
      tabSize: 2,
      wordWrap: "on",
      padding: padding ?? { top: 12, bottom: 12 },
      suggest: {
        showKeywords: true,
        showSnippets: true,
        showFunctions: true,
        showVariables: true,
      },
      quickSuggestions: {
        other: true,
        comments: false,
        strings: false,
      },
      parameterHints: { enabled: true },
      formatOnPaste: true,
      formatOnType: true,
    });

    // Register AI completion provider if aiAssisted
    if (aiAssisted) {
      // AI-assisted completions would be registered here
      // This would integrate with FlyMind or other AI service
    }

    // Cursor change listener
    if (onCursorChange) {
      editor.onDidChangeCursorPosition((e) => {
        onCursorChange({
          lineNumber: e.position.lineNumber,
          column: e.position.column,
        });
      });
    }

    onEditorMount?.(editor, monaco);
  };

  // Handle change
  const handleChange: OnChange = (newValue) => {
    onChange?.(newValue ?? "");
  };

  // Update theme when it changes
  React.useEffect(() => {
    if (monacoRef.current) {
      const themeKey = theme in THEME_CONFIG ? theme : "studio-dark";
      monacoRef.current.editor.setTheme(`functionfly-${themeKey}`);
    }
  }, [theme]);

  return (
    <div className={cn("relative overflow-hidden rounded-lg", className)} style={{ height }}>
      {aiAssisted && (
        <div className="absolute top-2 right-2 z-10 flex items-center gap-1.5 px-2 py-1 bg-brand-500/20 border border-brand-500/30 rounded text-[10px] text-brand-400">
          <Sparkles className="size-3" />
          AI Assisted
        </div>
      )}
      <Editor
        height="100%"
        language={resolvedLanguage}
        value={value}
        theme={`functionfly-${theme}`}
        onChange={handleChange}
        beforeMount={handleEditorWillMount}
        onMount={handleEditorDidMount}
        loading={
          <div className="flex items-center justify-center h-full text-text-muted text-sm">
            Loading editor...
          </div>
        }
        options={{
          readOnly,
          fontSize,
          lineHeight,
          minimap: { enabled: showMinimap },
          lineNumbers: showLineNumbers ? "on" : "off",
          scrollBeyondLastLine: false,
          automaticLayout: true,
        }}
      />
    </div>
  );
}

export { THEME_CONFIG };
