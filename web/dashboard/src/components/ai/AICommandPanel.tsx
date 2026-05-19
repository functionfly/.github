/**
 * AI Command Panel
 * Main container that wires up all AI components
 */

import React, { useState, useCallback, useMemo } from 'react'
import { cn } from '@/lib/utils'
import { useAICommandSystem, useAIConfidence, useActiveAgents, useToolInvocations } from '@/hooks/useAICommandSystem'
import { 
  AICommandPalette,
  PromptComposer,
  AgentChatPanel,
  ReasoningStream,
  PromptHistory,
  ConversationThread,
  ExecutionNarrator,
  GoalPlanner,
  IntentTranslator,
  PromptTemplateLibrary,
  ToolInvocationFeed,
  AIConfidenceMeter,
  MultiAgentConversationView,
  AgentConversationTimeline,
} from '@functionfly/ui-ai'
import { 
  Sparkles, MessageSquare, History, Target, Wrench, 
  Bot, Clock, ChevronRight, Plus, Settings 
} from 'lucide-react'
import { Button } from '@/components/ui/button'

interface AICommandPanelProps {
  className?: string
  defaultView?: 'chat' | 'history' | 'goals' | 'tools' | 'agents' | 'templates'
}

type ViewType = 'chat' | 'history' | 'goals' | 'tools' | 'agents' | 'templates' | 'timeline'

const viewConfig: Record<ViewType, { icon: React.ReactNode; label: string }> = {
  chat: { icon: <MessageSquare className="size-4" />, label: 'Chat' },
  history: { icon: <History className="size-4" />, label: 'History' },
  goals: { icon: <Target className="size-4" />, label: 'Goals' },
  tools: { icon: <Wrench className="size-4" />, label: 'Tools' },
  agents: { icon: <Bot className="size-4" />, label: 'Agents' },
  templates: { icon: <Sparkles className="size-4" />, label: 'Templates' },
  timeline: { icon: <Clock className="size-4" />, label: 'Timeline' },
}

export function AICommandPanel({ className, defaultView = 'chat' }: AICommandPanelProps) {
  const [activeView, setActiveView] = useState<ViewType>(defaultView)
  const [showCommandPalette, setShowCommandPalette] = useState(false)
  
  const ai = useAICommandSystem()
  const confidence = useAIConfidence()
  const { agents, currentAgent, totalUnread } = useActiveAgents()
  const { invocations, running, isRunning } = useToolInvocations()

  // Convert store types to component types
  const chatMessages = ai.messages.map(m => ({
    id: m.id,
    role: m.role,
    content: m.content,
    timestamp: m.timestamp,
    agentId: m.agentId,
    agentName: m.agentName,
    executionTime: m.executionTime,
    tokensUsed: m.tokensUsed,
    toolCalls: m.toolCalls,
  }))

  const reasoningSteps = useMemo(() => ai.executionSteps.map((step, i) => ({
    id: step.id,
    text: step.description || step.label,
    type: step.status === 'completed' ? 'result' as const : 
          step.status === 'running' ? 'action' as const : 'reasoning' as const,
    timestamp: step.timestamp,
    confidence: step.status === 'completed' ? 0.95 : step.status === 'running' ? 0.7 : 0.5,
  })), [ai.executionSteps])

  const agentChatMessages = useMemo(() => ai.messages.map(m => ({
    id: m.id,
    role: m.role,
    content: m.content,
    timestamp: m.timestamp,
    agentId: m.agentId,
    agentName: m.agentName,
  })), [ai.messages])

  return (
    <div className={cn("flex flex-col h-full bg-bg-primary", className)}>
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-border-subtle">
        <div className="flex items-center gap-3">
          <Sparkles className="size-5 text-brand-500" />
          <span className="text-sm font-bold text-text-primary">AI Studio</span>
          
          {/* Confidence Indicator */}
          <AIConfidenceMeter 
            confidence={confidence.confidence}
            history={confidence.history}
            showValue
            showHistory
            className="w-32"
          />
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setShowCommandPalette(true)}
            className="text-text-muted hover:text-text-primary"
          >
            <Sparkles className="size-4" />
          </Button>
        </div>
      </div>

      {/* View Navigation */}
      <div className="flex items-center gap-1 px-4 py-2 border-b border-border-subtle overflow-x-auto">
        {Object.entries(viewConfig).map(([key, config]) => (
          <button
            key={key}
            onClick={() => setActiveView(key as ViewType)}
            className={cn(
              "flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-full transition-colors whitespace-nowrap",
              activeView === key
                ? "bg-brand-500 text-white"
                : "text-text-muted hover:text-text-primary hover:bg-bg-hover"
            )}
          >
            {config.icon}
            {config.label}
            {key === 'agents' && totalUnread > 0 && (
              <span className="ml-1 size-4 rounded-full bg-error text-white text-[10px] flex items-center justify-center">
                {totalUnread}
              </span>
            )}
            {key === 'tools' && isRunning && (
              <span className="ml-1 size-2 rounded-full bg-brand-500 animate-pulse" />
            )}
          </button>
        ))}
      </div>

      {/* Content Area */}
      <div className="flex-1 overflow-hidden">
        {activeView === 'chat' && (
          <div className="h-full flex">
            {/* Chat Panel */}
            <div className="flex-1">
              <AgentChatPanel
                messages={chatMessages}
                agentName={currentAgent?.name || 'AI Assistant'}
                isThinking={ai.isThinking}
                onSendMessage={(msg) => ai.sendMessage(msg)}
                className="h-full border-0 rounded-none"
              />
            </div>

            {/* Reasoning Stream (sidebar) */}
            {reasoningSteps.length > 0 && (
              <div className="w-80 border-l border-border-subtle overflow-y-auto">
                <div className="p-3 border-b border-border-subtle bg-bg-tertiary/30">
                  <span className="text-xs font-medium text-text-muted">Reasoning</span>
                </div>
                <ReasoningStream
                  steps={reasoningSteps}
                  currentStep={ai.executionSteps.findIndex(s => s.id === ai.currentStepId)}
                  className="p-3"
                />
              </div>
            )}
          </div>
        )}

        {activeView === 'history' && (
          <PromptHistory
            items={ai.messages.map((m, i) => ({
              id: m.id,
              prompt: m.content,
              timestamp: m.timestamp,
              tokensUsed: m.tokensUsed,
              executionTime: m.executionTime,
              success: m.role !== 'tool',
              agentName: m.agentName,
            }))}
            onSelect={(item) => console.log('Select history item:', item)}
            onCopy={(item) => navigator.clipboard.writeText(item.prompt)}
            className="h-full"
          />
        )}

        {activeView === 'goals' && (
          <GoalPlanner
            goals={ai.goals.map(g => ({
              ...g,
              milestones: g.milestones.map(m => ({
                ...m,
                tasks: m.tasks.map(t => ({ ...t })),
              })),
            }))}
            onGoalAdd={(goal) => ai.addGoal(goal as any)}
            onGoalUpdate={(id, updates) => ai.updateGoal(id, updates as any)}
            className="h-full"
          />
        )}

        {activeView === 'tools' && (
          <ToolInvocationFeed
            invocations={invocations.map(inv => ({
              id: inv.id,
              toolName: inv.toolName,
              args: inv.args,
              result: inv.result,
              status: inv.status,
              startTime: inv.startTime,
              endTime: inv.endTime,
              error: inv.error,
              duration: inv.duration,
            }))}
            className="h-full"
          />
        )}

        {activeView === 'agents' && (
          <MultiAgentConversationView
            agents={agents.map(a => ({
              id: a.id,
              name: a.name,
              role: a.role,
              color: a.color,
              isActive: a.isActive,
              unreadCount: a.unreadCount,
            }))}
            messages={agentChatMessages.map(m => ({
              id: m.id,
              agentId: m.agentId || 'user',
              content: m.content,
              timestamp: m.timestamp,
              type: 'message' as const,
            }))}
            currentAgentId={ai.currentAgentId || undefined}
            onAgentSelect={(id) => ai.selectAgent(id)}
            onMessageSend={(agentId, content) => ai.sendMessage(content)}
            className="h-full"
          />
        )}

        {activeView === 'templates' && (
          <PromptTemplateLibrary
            templates={ai.templates.map(t => ({
              id: t.id,
              name: t.name,
              description: t.description,
              template: t.template,
              variables: t.variables,
              category: t.category,
              tags: t.tags,
              createdAt: t.createdAt,
              updatedAt: t.updatedAt,
              usageCount: t.usageCount,
            }))}
            onTemplateSelect={(template) => {
              ai.selectTemplate(template.id)
            }}
            className="h-full"
          />
        )}

        {activeView === 'timeline' && (
          <AgentConversationTimeline
            events={ai.messages.map((m, i) => ({
              id: m.id,
              agentId: m.agentId || 'user',
              agentName: m.agentName || 'User',
              type: m.role === 'tool' ? 'tool_call' as const : 
                    m.role === 'assistant' ? 'message' as const : 'action' as const,
              content: m.content,
              timestamp: m.timestamp,
            }))}
            className="h-full"
          />
        )}
      </div>

      {/* Command Palette */}
      <AICommandPalette
        commands={ai.commands.map(c => ({
          id: c.id,
          label: c.label,
          description: c.description,
          category: c.category,
          icon: c.icon,
          shortcut: c.shortcut,
          keywords: c.keywords,
          action: () => c.action(),
        }))}
        promptTemplates={ai.templates.map(t => ({
          id: t.id,
          name: t.name,
          description: t.description,
          template: t.template,
          variables: t.variables,
          category: t.category,
        }))}
        isOpen={showCommandPalette}
        onClose={() => setShowCommandPalette(false)}
      />
    </div>
  )
}
