/**
 * @functionfly/ui-ai
 * Type definitions for AI components
 */

import * as React from "react";

// ============================================================================
// Core Types (from AICommandPalette.tsx)
// ============================================================================

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

// ============================================================================
// PromptHistory Types
// ============================================================================

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

// ============================================================================
// ConversationThread Types
// ============================================================================

export interface ThreadMessage {
  id: string;
  role: "user" | "assistant" | "system" | "tool";
  content: string;
  timestamp: number;
  agentName?: string;
  isEdited?: boolean;
  reactions?: Array<{ emoji: string; count: number }>;
  parentId?: string;
  childIds?: string[];
}

export interface ConversationThreadProps {
  messages: ThreadMessage[];
  currentMessageId?: string;
  onMessageSelect?: (message: ThreadMessage) => void;
  onMessageReply?: (messageId: string, content: string) => void;
  onMessageReact?: (messageId: string, emoji: string) => void;
  onThreadCollapse?: (messageId: string) => void;
  className?: string;
}

// ============================================================================
// ExecutionNarrator Types
// ============================================================================

export interface ExecutionStep {
  id: string;
  label: string;
  status: "pending" | "running" | "completed" | "failed" | "skipped";
  duration?: number;
  timestamp: number;
  description?: string;
  artifacts?: Array<{ name: string; type: string; url: string }>;
  error?: string;
  metadata?: Record<string, string>;
}

export interface ExecutionNarratorProps {
  steps: ExecutionStep[];
  currentStepId?: string;
  onStepClick?: (step: ExecutionStep) => void;
  className?: string;
  autoPlay?: boolean;
}

// ============================================================================
// AgentDirectiveEditor Types
// ============================================================================

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

// ============================================================================
// GoalPlanner Types
// ============================================================================

export interface Milestone {
  id: string;
  label: string;
  status: "pending" | "in_progress" | "completed" | "blocked";
  description?: string;
  tasks: Array<{
    id: string;
    label: string;
    completed: boolean;
  }>;
}

export interface Goal {
  id: string;
  title: string;
  description?: string;
  status: "active" | "completed" | "abandoned";
  priority: "critical" | "high" | "medium" | "low";
  progress: number;
  milestones: Milestone[];
  createdAt: number;
  deadline?: number;
}

export interface GoalPlannerProps {
  goals: Goal[];
  onGoalAdd?: (goal: Omit<Goal, "id" | "createdAt">) => void;
  onGoalUpdate?: (id: string, updates: Partial<Goal>) => void;
  onGoalDelete?: (id: string) => void;
  onMilestoneToggle?: (goalId: string, milestoneId: string) => void;
  onTaskToggle?: (goalId: string, milestoneId: string, taskId: string) => void;
  className?: string;
}

// ============================================================================
// IntentTranslator Types
// ============================================================================

export interface Intent {
  action: string;
  entities: Record<string, string>;
  confidence: number;
  original: string;
}

export interface IntentTranslatorProps {
  input: string;
  parsedIntent?: Intent;
  onIntentConfirm?: (intent: Intent) => void;
  onIntentEdit?: (intent: Intent) => void;
  isProcessing?: boolean;
  className?: string;
}

// ============================================================================
// PromptTemplateLibrary Types
// ============================================================================

export interface PromptTemplateLibraryProps {
  templates: PromptTemplate[];
  onTemplateSelect?: (template: PromptTemplate) => void;
  onTemplateCreate?: () => void;
  onTemplateEdit?: (template: PromptTemplate) => void;
  onTemplateDelete?: (id: string) => void;
  onTemplateDuplicate?: (template: PromptTemplate) => void;
  className?: string;
}

// ============================================================================
// PromptDiffViewer Types
// ============================================================================

export interface DiffLine {
  type: "added" | "removed" | "unchanged";
  content: string;
  lineNumber?: number;
}

export interface PromptDiffViewerProps {
  original: string;
  modified: string;
  title?: string;
  onRestore?: () => void;
  onApply?: () => void;
  className?: string;
}

// ============================================================================
// ContextInjector Types
// ============================================================================

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

// ============================================================================
// ToolInvocationFeed Types
// ============================================================================

export interface ToolInvocation {
  id: string;
  toolName: string;
  args: Record<string, any>;
  result?: any;
  status: "pending" | "running" | "completed" | "failed";
  startTime: number;
  endTime?: number;
  error?: string;
  duration?: number;
}

export interface ToolInvocationFeedProps {
  invocations: ToolInvocation[];
  onInvocationClick?: (invocation: ToolInvocation) => void;
  className?: string;
  autoScroll?: boolean;
}

// ============================================================================
// AIResponseInspector Types
// ============================================================================

export interface ResponseMetadata {
  model?: string;
  tokensUsed?: number;
  promptTokens?: number;
  completionTokens?: number;
  finishReason?: string;
  latency?: number;
  cost?: number;
}

export interface AIResponseInspectorProps {
  content: string;
  metadata?: ResponseMetadata;
  reasoning?: string;
  onContentCopy?: () => void;
  onContentExport?: () => void;
  className?: string;
}

// ============================================================================
// AIConfidenceMeter Types
// ============================================================================

export interface AIConfidenceMeterProps {
  confidence: number;
  thresholds?: {
    high?: number;
    medium?: number;
  };
  label?: string;
  showValue?: boolean;
  showHistory?: boolean;
  history?: number[];
  className?: string;
}

// ============================================================================
// PromptVariableEditor Types
// ============================================================================

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

// ============================================================================
// AgentConversationTimeline Types
// ============================================================================

export interface TimelineEvent {
  id: string;
  agentId: string;
  agentName: string;
  type: "message" | "action" | "decision" | "tool_call" | "error";
  content: string;
  timestamp: number;
  metadata?: Record<string, string>;
}

export interface AgentConversationTimelineProps {
  events: TimelineEvent[];
  onEventClick?: (event: TimelineEvent) => void;
  onAgentFilter?: (agentId: string | null) => void;
  className?: string;
}

// ============================================================================
// MultiAgentConversationView Types
// ============================================================================

export interface Agent {
  id: string;
  name: string;
  role: string;
  color: string;
  isActive: boolean;
  unreadCount: number;
}

export interface MultiAgentConversationViewProps {
  agents: Agent[];
  messages: AgentMessage[];
  currentAgentId?: string;
  onAgentSelect?: (agentId: string) => void;
  onMessageSend?: (agentId: string, content: string) => void;
  onAgentAdd?: () => void;
  onAgentSettings?: (agentId: string) => void;
  className?: string;
}
