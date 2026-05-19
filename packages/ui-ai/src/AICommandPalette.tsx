/**
 * @functionfly/ui-ai
 * AI Command Palette and prompt engineering components for FunctionFly Studio
 */

import * as React from "react";
import { cn } from "@functionfly/ui-core";
import { Badge } from "@functionfly/ui-core";
import { Sparkles, Send, ChevronDown, Search, Zap, Terminal } from "lucide-react";

// --- Types ---
export interface AICommand {
  id: string;
  label: string;
  description: string;
  category: string;
  icon?: React.ReactNode;
  shortcut?: string;
  action: () => void | Promise<void>;
  keywords?: string[];
}

export interface PromptTemplate {
  id: string;
  name: string;
  template: string;
  description: string;
  variables: string[];
  category: string;
}

export interface AICommandPaletteProps {
  commands: AICommand[];
  promptTemplates: PromptTemplate[];
  isOpen?: boolean;
  onClose?: () => void;
  onToggle?: () => void;
  placeholder?: string;
  maxResults?: number;
  className?: string;
}

export interface PromptComposerProps {
  template?: PromptTemplate;
  initialPrompt?: string;
  variables?: Record<string, string>;
  onChange?: (prompt: string) => void;
  onSubmit?: (prompt: string) => void;
  className?: string;
}

export interface AgentChatMessage {
  id: string;
  role: "user" | "assistant" | "system" | "tool";
  content: string;
  timestamp: number;
  agentId?: string;
  agentName?: string;
  executionTime?: number;
  tokensUsed?: number;
  toolCalls?: Array<{
    name: string;
    args: Record<string, any>;
    result?: any;
  }>;
}

export interface AgentChatPanelProps {
  messages: AgentChatMessage[];
  onSendMessage?: (message: string) => void;
  agentName?: string;
  isThinking?: boolean;
  className?: string;
}

export interface ReasoningStreamProps {
  steps: Array<{
    id: string;
    text: string;
    type: "observation" | "reasoning" | "decision" | "action" | "result";
    timestamp: number;
    confidence?: number;
  }>;
  currentStep?: number;
  className?: string;
}

// --- AI Command Palette ---
export function AICommandPalette({
  commands,
  promptTemplates,
  isOpen = false,
  onClose,
  onToggle,
  placeholder = "Type a command...",
  maxResults = 8,
  className,
}: AICommandPaletteProps) {
  const [query, setQuery] = React.useState("");
  const [selectedIndex, setSelectedIndex] = React.useState(0);

  const filteredCommands = React.useMemo(() => {
    if (!query.trim()) return commands.slice(0, maxResults);
    const q = query.toLowerCase();
    return commands
      .filter(
        (cmd) =>
          cmd.label.toLowerCase().includes(q) ||
          cmd.description.toLowerCase().includes(q) ||
          cmd.keywords?.some((k) => k.toLowerCase().includes(q))
      )
      .slice(0, maxResults);
  }, [commands, query, maxResults]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Escape") {
      onClose?.();
      return;
    }
    if (e.key === "ArrowDown") {
      setSelectedIndex((i) => Math.min(i + 1, filteredCommands.length - 1));
      e.preventDefault();
    }
    if (e.key === "ArrowUp") {
      setSelectedIndex((i) => Math.max(i - 1, 0));
      e.preventDefault();
    }
    if (e.key === "Enter" && filteredCommands[selectedIndex]) {
      filteredCommands[selectedIndex].action();
      setQuery("");
      onClose?.();
    }
  };

  if (!isOpen) return null;

  return (
    <div
      className={cn(
        "fixed inset-0 z-50 flex items-start justify-center pt-[20vh]",
        "bg-black/50 backdrop-blur-sm",
        className
      )}
      onClick={() => onClose?.()}
    >
      <div
        className={cn(
          "w-full max-w-2xl mx-4 bg-bg-secondary border border-border-default rounded-2xl shadow-2xl",
          "animate-in fade-in zoom-in-95 duration-150"
        )}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Search bar */}
        <div className="flex items-center gap-3 px-4 py-3 border-b border-border-subtle">
          <Search className="size-5 text-text-muted shrink-0" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            className="flex-1 bg-transparent text-text-primary text-sm outline-none placeholder:text-text-muted"
            autoFocus
          />
          {/* Keyboard shortcut hint */}
          <kbd className="hidden md:inline-flex items-center px-2 py-0.5 text-[10px] font-mono text-text-muted bg-bg-tertiary rounded border border-border-subtle">
            ESC
          </kbd>
        </div>

        {/* Results */}
        <div className="max-h-96 overflow-y-auto p-2">
          {filteredCommands.length > 0 ? (
            filteredCommands.map((cmd, i) => (
              <button
                key={cmd.id}
                className={cn(
                  "w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-left transition-colors duration-100",
                  i === selectedIndex
                    ? "bg-brand-500/10 border border-brand-500/30"
                    : "hover:bg-bg-hover"
                )}
                onClick={() => {
                  cmd.action();
                  setQuery("");
                  onClose?.();
                }}
              >
                {cmd.icon || <Sparkles className="size-5 text-brand-400 shrink-0" />}
                <div className="flex-1 min-w-0">
                  <div className="text-sm font-medium text-text-primary">{cmd.label}</div>
                  <div className="text-xs text-text-muted truncate">{cmd.description}</div>
                </div>
                {cmd.shortcut && (
                  <kbd className="text-[10px] font-mono text-text-muted bg-bg-tertiary px-1.5 py-0.5 rounded border border-border-subtle">
                    {cmd.shortcut}
                  </kbd>
                )}
              </button>
            ))
          ) : (
            <div className="flex flex-col items-center justify-center py-12 text-text-muted">
              <Zap className="size-12 mb-3 opacity-50" />
              <p className="text-sm">No commands found for "{query}"</p>
              <p className="text-xs mt-1">Try different keywords or press Escape to close</p>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between px-4 py-2 border-t border-border-subtle text-[10px] text-text-muted">
          <div className="flex items-center gap-3">
            <Badge variant="brand" size="sm">
              <Sparkles className="size-3 mr-1" />
              AI-Powered
            </Badge>
            <span>{commands.length} commands available</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="flex items-center gap-1">
              <ChevronDown className="size-3" />
              Navigate
            </span>
            <span className="flex items-center gap-1">
              <Send className="size-3" />
              Execute
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}

// --- Prompt Composer ---
export function PromptComposer({
  template,
  initialPrompt = "",
  variables = {},
  onChange,
  onSubmit,
  className,
}: PromptComposerProps) {
  const [prompt, setPrompt] = React.useState(initialPrompt);
  const [filledVars, setFilledVars] = React.useState<Record<string, string>>(variables);

  React.useEffect(() => {
    setPrompt(initialPrompt);
  }, [initialPrompt]);

  const handleVariableChange = (varName: string, value: string) => {
    setFilledVars((prev) => ({ ...prev, [varName]: value }));
    const newPrompt = prompt.replace(new RegExp(`\\{${varName}\\}`, "g"), value);
    setPrompt(newPrompt);
    onChange?.(newPrompt);
  };

  return (
    <div className={cn("space-y-3", className)}>
      {/* Variable inputs */}
      {template?.variables.length ? (
        <div className="grid grid-cols-2 gap-2">
          {template.variables.map((varName) => (
            <div key={varName} className="space-y-1">
              <label className="text-[11px] font-medium text-text-secondary capitalize">
                {varName.replace(/_/g, " ")}
              </label>
              <input
                type="text"
                value={filledVars[varName] || ""}
                onChange={(e) => handleVariableChange(varName, e.target.value)}
                placeholder={`Enter ${varName}...`}
                className="w-full px-3 py-1.5 text-sm bg-bg-primary border border-border-subtle rounded-lg text-text-primary focus:outline-none focus:border-brand-500 transition-colors"
              />
            </div>
          ))}
        </div>
      ) : null}

      {/* Template info */}
      {template && (
        <div className="flex items-center gap-2 text-xs text-text-muted">
          <Terminal className="size-3" />
          <span>Template: {template.name}</span>
          <Badge variant="outline" size="sm">
            {template.category}
          </Badge>
        </div>
      )}

      {/* Prompt preview */}
      <div className="relative">
        <textarea
          value={prompt}
          onChange={(e) => {
            setPrompt(e.target.value);
            onChange?.(e.target.value);
          }}
          className="w-full min-h-[120px] px-4 py-3 text-sm font-mono bg-bg-primary border border-border-subtle rounded-lg text-text-primary resize-none focus:outline-none focus:border-brand-500 transition-colors"
          placeholder="Compose your prompt..."
          spellCheck={false}
        />
        {/* Word/token count */}
        <div className="absolute bottom-2 right-2 text-[10px] text-text-muted">
          {prompt.split(/\s+/).filter(Boolean).length} words
        </div>
      </div>

      {/* Submit button */}
      {onSubmit && (
        <button
          onClick={() => onSubmit(prompt)}
          className="w-full px-4 py-2.5 bg-brand-500 hover:bg-brand-600 text-white font-medium rounded-lg transition-colors flex items-center justify-center gap-2 group"
        >
          <Sparkles className="size-4 group-hover:animate-pulse" />
          Execute Prompt
        </button>
      )}
    </div>
  );
}

// --- Agent Chat Panel ---
export function AgentChatPanel({
  messages,
  onSendMessage,
  agentName = "AI Agent",
  isThinking = false,
  className,
}: AgentChatPanelProps) {
  const [input, setInput] = React.useState("");
  const scrollRef = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    scrollRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, isThinking]);

  const handleSubmit = (e?: React.FormEvent) => {
    e?.preventDefault();
    if (input.trim() && onSendMessage) {
      onSendMessage(input.trim());
      setInput("");
    }
  };

  const renderMessage = (msg: AgentChatMessage) => {
    const isUser = msg.role === "user";
    const isTool = msg.role === "tool";

    return (
      <div
        key={msg.id}
        className={cn(
          "flex gap-3 py-2",
          isUser ? "flex-row-reverse" : "flex-row"
        )}
      >
        {/* Avatar */}
        <div
          className={cn(
            "size-8 rounded-lg flex items-center justify-center text-xs font-bold shrink-0",
            isUser
              ? "bg-brand-500 text-white"
              : "bg-bg-tertiary text-text-muted border border-border-subtle"
          )}
        >
          {isUser ? "U" : msg.agentName?.[0] || "A"}
        </div>

        {/* Content */}
        <div
          className={cn(
            "max-w-[80%] rounded-xl px-4 py-2.5",
            isUser
              ? "bg-brand-500/10 border border-brand-500/20 rounded-br-none"
              : "bg-bg-primary border border-border-subtle rounded-bl-none"
          )}
        >
          <div className="text-sm text-text-primary whitespace-pre-wrap">{msg.content}</div>
          <div className="flex items-center gap-2 mt-1">
            <span className="text-[10px] text-text-muted">
              {new Date(msg.timestamp).toLocaleTimeString()}
            </span>
            {msg.executionTime != null && (
              <span className="text-[10px] text-info">{msg.executionTime}ms</span>
            )}
            {msg.tokensUsed != null && (
              <span className="text-[10px] text-brand-500">{msg.tokensUsed} tokens</span>
            )}
          </div>
        </div>
      </div>
    );
  };

  const renderToolCall = (msg: AgentChatMessage) => {
    return msg.toolCalls?.map((tool, i) => (
      <div key={i} className="flex gap-3 py-1">
        <div className="size-8 rounded-lg flex items-center justify-center text-[10px] shrink-0 bg-bg-tertiary border border-border-subtle text-text-muted">
          TOOL
        </div>
        <div className="bg-bg-primary border border-border-subtle rounded-lg px-3 py-2 text-sm max-w-[80%]">
          <span className="font-mono text-brand-500">{tool.name}</span>
          <span className="text-text-muted ml-2">
            {JSON.stringify(tool.args)}
          </span>
          {tool.result && (
            <div className="mt-1 text-[11px] text-text-secondary">
              → {JSON.stringify(tool.result).slice(0, 200)}
            </div>
          )}
        </div>
      </div>
    ));
  };

  return (
    <div className={cn("flex flex-col h-full border border-border-subtle rounded-xl overflow-hidden", className)}>
      {/* Header */}
      <div className="flex items-center gap-2 px-4 py-3 border-b border-border-subtle bg-bg-secondary">
        <div className="size-2 rounded-full bg-brand-500 animate-pulse" />
        <span className="text-sm font-medium text-text-primary">{agentName}</span>
        {isThinking && (
          <Badge variant="brand" size="sm" dot pulse>
            Thinking...
          </Badge>
        )}
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {messages.map((msg) => (
          <div key={msg.id}>
            {renderMessage(msg)}
            {(msg.role === "assistant" || msg.role === "system") && msg.toolCalls && renderToolCall(msg)}
          </div>
        ))}

        {/* Thinking indicator */}
        {isThinking && (
          <div className="flex gap-3 py-2">
            <div className="size-8 rounded-lg flex items-center justify-center text-[10px] shrink-0 bg-bg-tertiary border border-border-subtle text-brand-500">
              🤖
            </div>
            <div className="bg-bg-primary border border-border-subtle rounded-xl rounded-bl-none px-4 py-3">
              <div className="flex gap-1.5">
                <span className="size-2 rounded-full bg-brand-500 animate-bounce" style={{ animationDelay: "0ms" }} />
                <span className="size-2 rounded-full bg-brand-500 animate-bounce" style={{ animationDelay: "150ms" }} />
                <span className="size-2 rounded-full bg-brand-500 animate-bounce" style={{ animationDelay: "300ms" }} />
              </div>
            </div>
          </div>
        )}

        <div ref={scrollRef} />
      </div>

      {/* Input */}
      {onSendMessage && (
        <form onSubmit={handleSubmit} className="flex items-center gap-2 p-3 border-t border-border-subtle bg-bg-secondary">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Message the agent..."
            className="flex-1 px-3 py-2 text-sm bg-bg-primary border border-border-subtle rounded-lg text-text-primary focus:outline-none focus:border-brand-500 transition-colors"
          />
          <button
            type="submit"
            disabled={!input.trim()}
            className="p-2.5 bg-brand-500 hover:bg-brand-600 text-white rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Send className="size-4" />
          </button>
        </form>
      )}
    </div>
  );
}

// --- Reasoning Stream ---
export function ReasoningStream({
  steps,
  currentStep,
  className,
}: ReasoningStreamProps) {
  return (
    <div className={cn("space-y-2", className)}>
      {steps.map((step, i) => {
        const isCurrent = currentStep != null ? i === currentStep : false;
        const isPast = currentStep != null ? i < currentStep : false;
        const typeColor = {
          observation: "#3b82f6",
          reasoning: "#8b5cf6",
          decision: "#f97316",
          action: "#10b981",
          result: "#6366f1",
        }[step.type];

        return (
          <div
            key={step.id}
            className={cn(
              "flex gap-3 items-start py-2 px-3 rounded-lg transition-all duration-300",
              isCurrent
                ? "bg-brand-500/5 border border-brand-500/20"
                : isPast
                ? "opacity-60"
                : "hover:bg-bg-hover"
            )}
          >
            {/* Timeline line */}
            <div className="flex flex-col items-center pt-1">
              <div
                className="size-2.5 rounded-full border-2 transition-all duration-300"
                style={{
                  backgroundColor: isCurrent ? typeColor : "transparent",
                  borderColor: isPast ? typeColor : "var(--border-subtle)",
                }}
              />
              {i < steps.length - 1 && (
                <div
                  className="w-[1px] flex-1 transition-all duration-300"
                  style={{
                    backgroundColor: isPast ? typeColor : "var(--border-subtle)",
                    opacity: isPast ? 1 : 0.3,
                  }}
                />
              )}
            </div>

            {/* Content */}
            <div className="flex-1 min-w-0 py-0.5">
              <div className="flex items-center gap-2 mb-0.5">
                <Badge
                  variant={isCurrent ? "brand" : "ghost"}
                  size="sm"
                >
                  {step.type}
                </Badge>
                {step.confidence != null && (
                  <span className="text-[10px] text-text-muted">
                    Confidence: {(step.confidence * 100).toFixed(0)}%
                  </span>
                )}
              </div>
              <p className="text-sm text-text-secondary">{step.text}</p>
            </div>

            {/* Timestamp */}
            <span className="text-[10px] text-text-muted shrink-0 ml-2">
              {new Date(step.timestamp).toLocaleTimeString()}
            </span>
          </div>
        );
      })}
    </div>
  );
}