/**
 * AI Command System Hook
 * Provides unified access to AI components and their state
 */

import { useCallback, useMemo } from 'react'
import { useAICommandStore } from '@/stores/aiCommandStore'
import type {
  AIMessage,
  ToolInvocation,
  PromptTemplate,
  Goal,
  Agent,
  ExecutionStep,
  AICommand,
} from '@/stores/aiCommandStore'

// ============================================================================
// useAICommandSystem
// ============================================================================

export function useAICommandSystem() {
  const store = useAICommandStore()

  // Command Palette
  const openCommandPalette = useCallback(() => {
    store.setCommandPaletteOpen(true)
  }, [store])

  const closeCommandPalette = useCallback(() => {
    store.setCommandPaletteOpen(false)
  }, [store])

  const toggleCommandPalette = useCallback(() => {
    store.setCommandPaletteOpen(!store.isCommandPaletteOpen)
  }, [store])

  // Messages
  const addMessage = useCallback((message: Omit<AIMessage, 'id' | 'timestamp'>) => {
    store.addMessage({
      ...message,
      id: `msg-${Date.now()}`,
      timestamp: Date.now(),
    })
  }, [store])

  const sendMessage = useCallback((content: string) => {
    addMessage({
      role: 'user',
      content,
    })
    store.setIsThinking(true)
    
    // Simulate AI response after delay
    setTimeout(() => {
      const assistantMessage: AIMessage = {
        id: `ai-${Date.now()}`,
        timestamp: Date.now(),
        role: 'assistant',
        content: 'I understand. Let me help you with that.',
        agentId: store.currentAgentId || 'primary',
        agentName: store.agents.find(a => a.id === store.currentAgentId)?.name || 'Assistant',
      }
      store.addMessage(assistantMessage)
      store.setIsThinking(false)
    }, 1500)
  }, [addMessage, store])

  // Tool Invocations
  const addToolInvocation = useCallback((tool: Omit<ToolInvocation, 'id' | 'startTime'>) => {
    store.addToolInvocation({
      ...tool,
      id: `tool-${Date.now()}`,
      startTime: Date.now(),
    })
  }, [store])

  const completeToolInvocation = useCallback((id: string, result: any) => {
    store.updateToolInvocation(id, {
      status: 'completed',
      result,
      endTime: Date.now(),
      duration: Date.now() - (store.toolInvocations.find(t => t.id === id)?.startTime || 0),
    })
  }, [store])

  const failToolInvocation = useCallback((id: string, error: string) => {
    store.updateToolInvocation(id, {
      status: 'failed',
      error,
      endTime: Date.now(),
      duration: Date.now() - (store.toolInvocations.find(t => t.id === id)?.startTime || 0),
    })
  }, [store])

  // Templates
  const selectTemplate = useCallback((id: string | null) => {
    store.setSelectedTemplate(id)
    if (id) {
      store.incrementTemplateUsage(id)
    }
  }, [store])

  const createTemplate = useCallback((template: PromptTemplate) => {
    store.setTemplates([...store.templates, { ...template, id: `tmpl-${Date.now()}`, createdAt: Date.now(), updatedAt: Date.now(), usageCount: 0 }])
  }, [store])

  // Goals
  const addGoal = useCallback((goal: Omit<Goal, 'id' | 'createdAt'>) => {
    store.addGoal({
      ...goal,
      id: `goal-${Date.now()}`,
      createdAt: Date.now(),
    })
  }, [store])

  const updateGoalProgress = useCallback((goalId: string, progress: number) => {
    store.updateGoal(goalId, { progress })
  }, [store])

  const completeGoal = useCallback((goalId: string) => {
    store.updateGoal(goalId, { status: 'completed', progress: 100 })
  }, [store])

  // Agents
  const selectAgent = useCallback((agentId: string) => {
    store.setCurrentAgent(agentId)
    // Reset unread count
    store.updateAgent(agentId, { unreadCount: 0 })
  }, [store])

  // Execution
  const addExecutionStep = useCallback((step: Omit<ExecutionStep, 'id' | 'timestamp'>) => {
    store.addExecutionStep({
      ...step,
      id: `step-${Date.now()}`,
      timestamp: Date.now(),
    })
  }, [store])

  const completeExecutionStep = useCallback((stepId: string) => {
    store.updateExecutionStep(stepId, { status: 'completed' })
    store.setCurrentStep(null)
  }, [store])

  // Confidence
  const updateConfidence = useCallback((confidence: number) => {
    store.setConfidence(confidence)
  }, [store])

  return {
    // State
    ...store,
    
    // Command Palette
    openCommandPalette,
    closeCommandPalette,
    toggleCommandPalette,
    
    // Messages
    addMessage,
    sendMessage,
    clearMessages: store.clearMessages,
    
    // Tool Invocations
    addToolInvocation,
    completeToolInvocation,
    failToolInvocation,
    clearToolInvocations: store.clearToolInvocations,
    
    // Templates
    selectTemplate,
    createTemplate,
    
    // Goals
    addGoal,
    updateGoalProgress,
    completeGoal,
    deleteGoal: store.deleteGoal,
    
    // Agents
    selectAgent,
    addAgent: (agent: Omit<Agent, 'id'>) => {
      store.setAgents([...store.agents, { ...agent, id: `agent-${Date.now()}` }])
    },
    
    // Execution
    addExecutionStep,
    completeExecutionStep,
    setExecutionSteps: store.setExecutionSteps,
    
    // Confidence
    updateConfidence,
    
    // Commands
    executeCommand: (commandId: string) => {
      const cmd = store.commands.find(c => c.id === commandId)
      if (cmd) cmd.action()
    },
  }
}

// ============================================================================
// useAIConfidence
// ============================================================================

export function useAIConfidence() {
  const { lastConfidence, confidenceHistory } = useAICommandStore()
  
  const level = useMemo(() => {
    if (lastConfidence >= 0.8) return 'high'
    if (lastConfidence >= 0.5) return 'medium'
    return 'low'
  }, [lastConfidence])
  
  return {
    confidence: lastConfidence,
    level,
    history: confidenceHistory,
    isHigh: lastConfidence >= 0.8,
    isLow: lastConfidence < 0.5,
  }
}

// ============================================================================
// useActiveAgents
// ============================================================================

export function useActiveAgents() {
  const { agents, currentAgentId, setCurrentAgent } = useAICommandStore()
  
  const activeAgents = useMemo(() => 
    agents.filter(a => a.isActive)
  , [agents])
  
  const totalUnread = useMemo(() =>
    agents.reduce((sum, a) => sum + a.unreadCount, 0)
  , [agents])
  
  return {
    agents: activeAgents,
    currentAgent: agents.find(a => a.id === currentAgentId),
    currentAgentId,
    totalUnread,
    selectAgent: setCurrentAgent,
  }
}

// ============================================================================
// useToolInvocations
// ============================================================================

export function useToolInvocations() {
  const { toolInvocations, addToolInvocation, clearToolInvocations } = useAICommandStore()
  
  const running = useMemo(() =>
    toolInvocations.filter(t => t.status === 'running')
  , [toolInvocations])
  
  const completed = useMemo(() =>
    toolInvocations.filter(t => t.status === 'completed')
  , [toolInvocations])
  
  const failed = useMemo(() =>
    toolInvocations.filter(t => t.status === 'failed')
  , [toolInvocations])
  
  return {
    invocations: toolInvocations,
    running,
    completed,
    failed,
    addInvocation: addToolInvocation,
    clearAll: clearToolInvocations,
    isRunning: running.length > 0,
  }
}

// ============================================================================
// useGoalProgress
// ============================================================================

export function useGoalProgress() {
  const { goals, updateGoal, addGoal } = useAICommandStore()
  
  const activeGoals = useMemo(() =>
    goals.filter(g => g.status === 'active')
  , [goals])
  
  const completedGoals = useMemo(() =>
    goals.filter(g => g.status === 'completed')
  , [goals])
  
  const totalProgress = useMemo(() => {
    if (activeGoals.length === 0) return 0
    return Math.round(
      activeGoals.reduce((sum, g) => sum + g.progress, 0) / activeGoals.length
    )
  }, [activeGoals])
  
  return {
    goals,
    activeGoals,
    completedGoals,
    totalProgress,
    updateGoal,
    addGoal,
    isAllComplete: activeGoals.length === 0 && completedGoals.length > 0,
  }
}
