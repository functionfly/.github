/**
 * AI Command System Store
 * Global state management for AI components
 */

import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'

// ============================================================================
// Types
// ============================================================================

export interface AIMessage {
  id: string
  role: "user" | "assistant" | "system" | "tool"
  content: string
  timestamp: number
  agentId?: string
  agentName?: string
  executionTime?: number
  tokensUsed?: number
  toolCalls?: Array<{
    name: string
    args: Record<string, any>
    result?: any
  }>
}

export interface ToolInvocation {
  id: string
  toolName: string
  args: Record<string, any>
  result?: any
  status: "pending" | "running" | "completed" | "failed"
  startTime: number
  endTime?: number
  error?: string
  duration?: number
}

export interface PromptTemplate {
  id: string
  name: string
  template: string
  description: string
  variables: string[]
  category: string
  tags: string[]
  createdAt: number
  updatedAt: number
  usageCount: number
}

export interface Goal {
  id: string
  title: string
  description?: string
  status: "active" | "completed" | "abandoned"
  priority: "critical" | "high" | "medium" | "low"
  progress: number
  milestones: Array<{
    id: string
    label: string
    status: "pending" | "in_progress" | "completed" | "blocked"
    tasks: Array<{ id: string; label: string; completed: boolean }>
  }>
  createdAt: number
  deadline?: number
}

export interface Agent {
  id: string
  name: string
  role: string
  color: string
  isActive: boolean
  unreadCount: number
}

export interface ExecutionStep {
  id: string
  label: string
  status: "pending" | "running" | "completed" | "failed" | "skipped"
  duration?: number
  timestamp: number
  description?: string
  artifacts?: Array<{ name: string; type: string; url: string }>
  error?: string
}

export interface AICommand {
  id: string
  label: string
  description: string
  category: string
  icon?: React.ReactNode
  shortcut?: string
  action: () => void | Promise<void>
  keywords?: string[]
}

// ============================================================================
// Store Interface
// ============================================================================

interface AICommandState {
  // Command Palette
  isCommandPaletteOpen: boolean
  setCommandPaletteOpen: (open: boolean) => void
  
  // Messages
  messages: AIMessage[]
  addMessage: (message: AIMessage) => void
  clearMessages: () => void
  
  // Tool Invocations
  toolInvocations: ToolInvocation[]
  addToolInvocation: (invocation: ToolInvocation) => void
  updateToolInvocation: (id: string, updates: Partial<ToolInvocation>) => void
  clearToolInvocations: () => void
  
  // Templates
  templates: PromptTemplate[]
  selectedTemplateId: string | null
  setTemplates: (templates: PromptTemplate[]) => void
  setSelectedTemplate: (id: string | null) => void
  incrementTemplateUsage: (id: string) => void
  
  // Goals
  goals: Goal[]
  addGoal: (goal: Goal) => void
  updateGoal: (id: string, updates: Partial<Goal>) => void
  deleteGoal: (id: string) => void
  
  // Agents
  agents: Agent[]
  currentAgentId: string | null
  setAgents: (agents: Agent[]) => void
  setCurrentAgent: (id: string | null) => void
  updateAgent: (id: string, updates: Partial<Agent>) => void
  
  // Execution
  executionSteps: ExecutionStep[]
  currentStepId: string | null
  setExecutionSteps: (steps: ExecutionStep[]) => void
  addExecutionStep: (step: ExecutionStep) => void
  updateExecutionStep: (id: string, updates: Partial<ExecutionStep>) => void
  setCurrentStep: (id: string | null) => void
  
  // Commands
  commands: AICommand[]
  setCommands: (commands: AICommand[]) => void
  
  // Thinking state
  isThinking: boolean
  setIsThinking: (thinking: boolean) => void
  
  // Confidence
  lastConfidence: number
  confidenceHistory: number[]
  setConfidence: (confidence: number) => void
}

// ============================================================================
// Default Data
// ============================================================================

const defaultTemplates: PromptTemplate[] = [
  {
    id: 'coding-assist',
    name: 'Coding Assistant',
    template: 'You are a coding assistant. Help with the following task:\n\n{task}\n\nContext: {context}',
    description: 'General coding assistance with context awareness',
    variables: ['task', 'context'],
    category: 'coding',
    tags: ['assistant', 'coding'],
    createdAt: Date.now() - 86400000,
    updatedAt: Date.now(),
    usageCount: 42,
  },
  {
    id: 'code-review',
    name: 'Code Review',
    template: 'Review the following code and provide feedback:\n\n```{language}\n{code}\n```\n\nFocus on: {focus}',
    description: 'Review code for bugs, style, and best practices',
    variables: ['language', 'code', 'focus'],
    category: 'coding',
    tags: ['review', 'quality'],
    createdAt: Date.now() - 172800000,
    updatedAt: Date.now() - 86400000,
    usageCount: 18,
  },
  {
    id: 'debug-helper',
    name: 'Debug Helper',
    template: 'Help debug the following error:\n\n{error}\n\nStack trace:\n{stack_trace}\n\nCode context:\n{code}',
    description: 'Systematic debugging assistance',
    variables: ['error', 'stack_trace', 'code'],
    category: 'coding',
    tags: ['debug', 'error'],
    createdAt: Date.now() - 259200000,
    updatedAt: Date.now() - 172800000,
    usageCount: 25,
  },
]

const defaultAgents: Agent[] = [
  { id: 'primary', name: 'Primary Agent', role: 'Assistant', color: '#ff6b35', isActive: true, unreadCount: 0 },
  { id: 'coder', name: 'Code Agent', role: 'Coding Specialist', color: '#00d4ff', isActive: true, unreadCount: 0 },
  { id: 'reviewer', name: 'Review Agent', role: 'Code Reviewer', color: '#00ff9d', isActive: false, unreadCount: 2 },
]

const defaultGoals: Goal[] = [
  {
    id: 'debug-auth',
    title: 'Debug authentication flow',
    description: 'Investigate and fix the login issues',
    status: 'active',
    priority: 'high',
    progress: 60,
    milestones: [
      { id: 'm1', label: 'Reproduce issue', status: 'completed', tasks: [
        { id: 't1', label: 'Check logs', completed: true },
        { id: 't2', label: 'Verify config', completed: true },
      ]},
      { id: 'm2', label: 'Identify root cause', status: 'in_progress', tasks: [
        { id: 't3', label: 'Review code', completed: true },
        { id: 't4', label: 'Test hypothesis', completed: false },
      ]},
      { id: 'm3', label: 'Implement fix', status: 'pending', tasks: [
        { id: 't5', label: 'Write fix', completed: false },
        { id: 't6', label: 'Test fix', completed: false },
      ]},
    ],
    createdAt: Date.now() - 86400000,
  },
]

const defaultCommands: AICommand[] = [
  {
    id: 'new-prompt',
    label: 'New Prompt',
    description: 'Start a new prompt',
    category: 'General',
    keywords: ['new', 'create', 'prompt'],
    action: () => console.log('New prompt'),
  },
  {
    id: 'template-library',
    label: 'Open Template Library',
    description: 'Browse prompt templates',
    category: 'Templates',
    keywords: ['template', 'library', 'browse'],
    action: () => console.log('Open template library'),
  },
  {
    id: 'agent-chat',
    label: 'Chat with Agent',
    description: 'Open agent chat panel',
    category: 'Agents',
    keywords: ['chat', 'agent', 'talk'],
    action: () => console.log('Open agent chat'),
  },
]

// ============================================================================
// Store Implementation
// ============================================================================

export const useAICommandStore = create<AICommandState>()(
  immer((set) => ({
    // Command Palette
    isCommandPaletteOpen: false,
    setCommandPaletteOpen: (open) => set((state) => {
      state.isCommandPaletteOpen = open
    }),

    // Messages
    messages: [],
    addMessage: (message) => set((state) => {
      state.messages.push(message)
    }),
    clearMessages: () => set((state) => {
      state.messages = []
    }),

    // Tool Invocations
    toolInvocations: [],
    addToolInvocation: (invocation) => set((state) => {
      state.toolInvocations.push(invocation)
    }),
    updateToolInvocation: (id, updates) => set((state) => {
      const inv = state.toolInvocations.find(i => i.id === id)
      if (inv) Object.assign(inv, updates)
    }),
    clearToolInvocations: () => set((state) => {
      state.toolInvocations = []
    }),

    // Templates
    templates: defaultTemplates,
    selectedTemplateId: null,
    setTemplates: (templates) => set((state) => {
      state.templates = templates
    }),
    setSelectedTemplate: (id) => set((state) => {
      state.selectedTemplateId = id
    }),
    incrementTemplateUsage: (id) => set((state) => {
      const tmpl = state.templates.find(t => t.id === id)
      if (tmpl) tmpl.usageCount++
    }),

    // Goals
    goals: defaultGoals,
    addGoal: (goal) => set((state) => {
      state.goals.push(goal)
    }),
    updateGoal: (id, updates) => set((state) => {
      const goal = state.goals.find(g => g.id === id)
      if (goal) Object.assign(goal, updates)
    }),
    deleteGoal: (id) => set((state) => {
      state.goals = state.goals.filter(g => g.id !== id)
    }),

    // Agents
    agents: defaultAgents,
    currentAgentId: 'primary',
    setAgents: (agents) => set((state) => {
      state.agents = agents
    }),
    setCurrentAgent: (id) => set((state) => {
      state.currentAgentId = id
    }),
    updateAgent: (id, updates) => set((state) => {
      const agent = state.agents.find(a => a.id === id)
      if (agent) Object.assign(agent, updates)
    }),

    // Execution
    executionSteps: [],
    currentStepId: null,
    setExecutionSteps: (steps) => set((state) => {
      state.executionSteps = steps
    }),
    addExecutionStep: (step) => set((state) => {
      state.executionSteps.push(step)
    }),
    updateExecutionStep: (id, updates) => set((state) => {
      const step = state.executionSteps.find(s => s.id === id)
      if (step) Object.assign(step, updates)
    }),
    setCurrentStep: (id) => set((state) => {
      state.currentStepId = id
    }),

    // Commands
    commands: defaultCommands,
    setCommands: (commands) => set((state) => {
      state.commands = commands
    }),

    // Thinking state
    isThinking: false,
    setIsThinking: (thinking) => set((state) => {
      state.isThinking = thinking
    }),

    // Confidence
    lastConfidence: 0,
    confidenceHistory: [],
    setConfidence: (confidence) => set((state) => {
      state.lastConfidence = confidence
      state.confidenceHistory.push(confidence)
      if (state.confidenceHistory.length > 20) {
        state.confidenceHistory = state.confidenceHistory.slice(-20)
      }
    }),
  }))
)
