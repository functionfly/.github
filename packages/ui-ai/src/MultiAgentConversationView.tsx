/**
 * @functionfly/ui-ai
 * Multi-Agent Conversation View - Multi-agent chat interface
 */

import * as React from "react";
import { cn } from "@functionfly/ui-core";
import { Badge } from "@functionfly/ui-core";
import { MessageSquare, Plus, Settings, Users, Send, Bot, User, Circle } from "lucide-react";

export interface AgentMessage {
  id: string;
  agentId: string;
  content: string;
  timestamp: number;
  type: "message" | "action" | "thinking";
}

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

export function MultiAgentConversationView({
  agents,
  messages,
  currentAgentId,
  onAgentSelect,
  onMessageSend,
  onAgentAdd,
  onAgentSettings,
  className,
}: MultiAgentConversationViewProps) {
  const [inputValues, setInputValues] = React.useState<Record<string, string>>({});
  const messageListRef = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    if (messageListRef.current) {
      messageListRef.current.scrollTop = messageListRef.current.scrollHeight;
    }
  }, [messages]);

  const handleSend = (agentId: string) => {
    const content = inputValues[agentId]?.trim();
    if (content) {
      onMessageSend?.(agentId, content);
      setInputValues(prev => ({ ...prev, [agentId]: "" }));
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent, agentId: string) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend(agentId);
    }
  };

  const activeAgents = agents.filter(a => a.isActive);
  const selectedAgent = currentAgentId 
    ? agents.find(a => a.id === currentAgentId) 
    : activeAgents[0];

  const agentMessages = messages.filter(m => 
    !currentAgentId || m.agentId === currentAgentId
  );

  return (
    <div className={cn("flex h-full", className)}>
      {/* Agent Sidebar */}
      <div className="w-64 border-r border-border-subtle flex flex-col bg-bg-secondary">
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-border-subtle">
          <div className="flex items-center gap-2">
            <Users className="size-4 text-brand-500" />
            <span className="text-sm font-medium text-text-primary">Agents</span>
          </div>
          <button
            onClick={onAgentAdd}
            className="p-1.5 rounded hover:bg-bg-tertiary text-text-muted hover:text-text-primary"
          >
            <Plus className="size-4" />
          </button>
        </div>

        {/* Agent List */}
        <div className="flex-1 overflow-y-auto p-2 space-y-1">
          {agents.map(agent => (
            <div
              key={agent.id}
              onClick={() => onAgentSelect?.(agent.id)}
              className={cn(
                "group relative flex items-center gap-3 p-3 rounded-lg cursor-pointer transition-colors",
                selectedAgent?.id === agent.id
                  ? "bg-brand-500/10 border border-brand-500/20"
                  : "hover:bg-bg-hover border border-transparent"
              )}
            >
              {/* Agent Avatar */}
              <div
                className="size-10 rounded-lg flex items-center justify-center text-sm font-bold shrink-0"
                style={{ backgroundColor: agent.color + "20", color: agent.color }}
              >
                {agent.name[0]}
              </div>

              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-text-primary">{agent.name}</span>
                  {agent.isActive ? (
                    <Circle className="size-2 fill-success text-success" />
                  ) : (
                    <Circle className="size-2 fill-bg-tertiary text-bg-tertiary" />
                  )}
                </div>
                <span className="text-xs text-text-muted">{agent.role}</span>
              </div>

              {/* Unread Badge */}
              {agent.unreadCount > 0 && (
                <span className="absolute top-2 right-2 size-5 rounded-full bg-brand-500 text-white text-[10px] font-bold flex items-center justify-center">
                  {agent.unreadCount > 9 ? "9+" : agent.unreadCount}
                </span>
              )}

              {/* Settings Button */}
              <button
                onClick={e => { e.stopPropagation(); onAgentSettings?.(agent.id); }}
                className="absolute top-2 right-2 p-1 rounded opacity-0 group-hover:opacity-100 hover:bg-bg-tertiary text-text-muted hover:text-text-primary transition-opacity"
              >
                <Settings className="size-3" />
              </button>
            </div>
          ))}
        </div>
      </div>

      {/* Conversation Area */}
      <div className="flex-1 flex flex-col">
        {selectedAgent ? (
          <>
            {/* Chat Header */}
            <div className="flex items-center justify-between px-4 py-3 border-b border-border-subtle">
              <div className="flex items-center gap-3">
                <div
                  className="size-8 rounded-lg flex items-center justify-center text-sm font-bold"
                  style={{ backgroundColor: selectedAgent.color + "20", color: selectedAgent.color }}
                >
                  {selectedAgent.name[0]}
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-text-primary">{selectedAgent.name}</span>
                    {selectedAgent.isActive && (
                      <Badge variant="success" size="sm">Active</Badge>
                    )}
                  </div>
                  <span className="text-xs text-text-muted">{selectedAgent.role}</span>
                </div>
              </div>
            </div>

            {/* Messages */}
            <div ref={messageListRef} className="flex-1 overflow-y-auto p-4 space-y-4">
              {agentMessages.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-full text-text-muted">
                  <MessageSquare className="size-12 mb-3 opacity-50" />
                  <p className="text-sm">No messages yet</p>
                  <p className="text-xs mt-1">Start the conversation</p>
                </div>
              ) : (
                agentMessages.map(message => {
                  const agent = agents.find(a => a.id === message.agentId);
                  const isUser = message.type === "message" && message.agentId === "user";
                  
                  return (
                    <div
                      key={message.id}
                      className={cn("flex gap-3", isUser ? "flex-row-reverse" : "flex-row")}
                    >
                      {/* Avatar */}
                      <div
                        className={cn(
                          "size-8 rounded-lg flex items-center justify-center text-sm font-bold shrink-0",
                          isUser
                            ? "bg-brand-500 text-white"
                            : agent
                            ? ""
                            : "bg-bg-tertiary text-text-muted"
                        )}
                        style={!isUser && agent ? { backgroundColor: agent.color + "20", color: agent.color } : {}}
                      >
                        {isUser ? "U" : agent?.name[0] || "?"}
                      </div>

                      {/* Message */}
                      <div
                        className={cn(
                          "max-w-[70%] rounded-xl px-4 py-2.5",
                          isUser
                            ? "bg-brand-500/10 border border-brand-500/20 rounded-br-none"
                            : "bg-bg-secondary border border-border-subtle rounded-bl-none"
                        )}
                      >
                        <p className="text-sm text-text-primary whitespace-pre-wrap">{message.content}</p>
                        <div className="flex items-center gap-2 mt-1">
                          <span className="text-[10px] text-text-muted">
                            {new Date(message.timestamp).toLocaleTimeString()}
                          </span>
                          {message.type === "thinking" && (
                            <Badge variant="brand" size="sm" className="text-[10px]">
                              Thinking...
                            </Badge>
                          )}
                        </div>
                      </div>
                    </div>
                  );
                })
              )}
            </div>

            {/* Input */}
            <div className="p-4 border-t border-border-subtle">
              <div className="flex items-center gap-3">
                <input
                  type="text"
                  value={inputValues[selectedAgent.id] || ""}
                  onChange={e => setInputValues(prev => ({ ...prev, [selectedAgent.id]: e.target.value }))}
                  onKeyDown={e => handleKeyDown(e, selectedAgent.id)}
                  placeholder={`Message ${selectedAgent.name}...`}
                  className="flex-1 px-4 py-2.5 bg-bg-primary border border-border-subtle rounded-xl text-sm text-text-primary focus:outline-none focus:border-brand-500 transition-colors"
                />
                <button
                  onClick={() => handleSend(selectedAgent.id)}
                  disabled={!inputValues[selectedAgent.id]?.trim()}
                  className="p-2.5 bg-brand-500 hover:bg-brand-600 text-white rounded-xl transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <Send className="size-5" />
                </button>
              </div>
            </div>
          </>
        ) : (
          <div className="flex-1 flex flex-col items-center justify-center text-text-muted">
            <Users className="size-12 mb-3 opacity-50" />
            <p className="text-sm">Select an agent to start chatting</p>
          </div>
        )}
      </div>
    </div>
  );
}
