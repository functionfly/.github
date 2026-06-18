/**
 * @functionfly/ui-ai
 * Conversation Thread Component
 */

import * as React from "react";
import { cn } from "@functionfly/ui-core";
import { Badge, Button, Textarea } from "@functionfly/ui-core";
import { MessageSquare, ArrowUp, ArrowDown, MoreHorizontal, RefreshCw, Send } from "lucide-react";

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
  onSendMessage?: (content: string) => void;
  isThinking?: boolean;
  className?: string;
}

export function ConversationThread({
  messages,
  currentMessageId,
  onMessageSelect,
  onMessageReply,
  onMessageReact,
  onThreadCollapse,
  onSendMessage,
  isThinking,
  className,
}: ConversationThreadProps) {
  const [collapsedIds, setCollapsedIds] = React.useState<Set<string>>(new Set());
  const [inputValue, setInputValue] = React.useState("");

  const toggleCollapse = (id: string) => {
    setCollapsedIds(prev => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
    onThreadCollapse?.(id);
  };

  const getMessageDepth = (msg: ThreadMessage): number => {
    let depth = 0;
    let current = msg;
    while (current.parentId) {
      depth++;
      current = messages.find(m => m.id === current.parentId) || current;
    }
    return depth;
  };

  const renderMessage = (msg: ThreadMessage) => {
    const isCollapsed = collapsedIds.has(msg.id);
    const depth = getMessageDepth(msg);
    const hasChildren = msg.childIds && msg.childIds.length > 0;

    return (
      <div
        key={msg.id}
        className={cn("group relative", depth > 0 && `pl-${depth * 8}`)}
        style={{ paddingLeft: depth * 32 }}
      >
        {/* Connection line */}
        {depth > 0 && (
          <div className="absolute left-4 top-0 bottom-0 w-[1px] bg-border-subtle" />
        )}

        {/* Message card */}
        <div
          onClick={() => onMessageSelect?.(msg)}
          className={cn(
            "relative p-3 rounded-lg border transition-colors cursor-pointer mb-1",
            currentMessageId === msg.id
              ? "bg-brand-500/5 border-brand-500/20"
              : "border-transparent hover:bg-bg-hover"
          )}
        >
          {/* Header */}
          <div className="flex items-center gap-2 mb-2">
            <div
              className={cn(
                "size-6 rounded-md flex items-center justify-center text-[10px] font-bold shrink-0",
                msg.role === "user"
                  ? "bg-brand-500 text-white"
                  : msg.role === "tool"
                  ? "bg-warning-500 text-white"
                  : "bg-bg-tertiary text-text-muted"
              )}
            >
              {msg.role === "user" ? "U" : msg.role === "tool" ? "T" : "A"}
            </div>
            <span className="text-xs font-medium text-text-primary">
              {msg.role === "user" ? "You" : msg.agentName || (msg.role === "tool" ? "Tool" : "Assistant")}
            </span>
            <span className="text-[10px] text-text-muted">
              {new Date(msg.timestamp).toLocaleTimeString()}
            </span>
            {msg.isEdited && (
              <Badge variant="ghost" size="sm">edited</Badge>
            )}
          </div>

          {/* Content */}
          <p className="text-sm text-text-secondary whitespace-pre-wrap">
            {msg.content}
          </p>

          {/* Reactions */}
          {msg.reactions && msg.reactions.length > 0 && (
            <div className="flex items-center gap-1 mt-2">
              {msg.reactions.map((reaction, i) => (
                <button
                  key={i}
                  onClick={e => { e.stopPropagation(); onMessageReact?.(msg.id, reaction.emoji); }}
                  className="flex items-center gap-1 px-2 py-0.5 rounded-full bg-bg-tertiary hover:bg-bg-hover text-xs transition-colors"
                >
                  <span>{reaction.emoji}</span>
                  <span className="text-text-muted">{reaction.count}</span>
                </button>
              ))}
            </div>
          )}

          {/* Actions */}
          <div className="absolute top-2 right-2 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
            <button
              onClick={e => { e.stopPropagation(); toggleCollapse(msg.id); }}
              className="p-1 rounded hover:bg-bg-tertiary text-text-muted"
            >
              {isCollapsed ? <ArrowDown className="size-3" /> : <ArrowUp className="size-3" />}
            </button>
            <button
              onClick={e => { e.stopPropagation(); onMessageReply?.(msg.id, ""); }}
              className="p-1 rounded hover:bg-bg-tertiary text-text-muted"
            >
              <RefreshCw className="size-3" />
            </button>
            <button
              onClick={e => e.stopPropagation()}
              className="p-1 rounded hover:bg-bg-tertiary text-text-muted"
            >
              <MoreHorizontal className="size-3" />
            </button>
          </div>

          {/* Collapsed indicator */}
          {isCollapsed && hasChildren && (
            <div className="flex items-center gap-1 mt-2 text-[10px] text-text-muted">
              <MessageSquare className="size-3" />
              <span>{msg.childIds?.length} replies hidden</span>
            </div>
          )}
        </div>

        {/* Child messages */}
        {!isCollapsed && msg.childIds && (
          <div className="ml-4">
            {msg.childIds.map(childId => {
              const child = messages.find(m => m.id === childId);
              return child ? renderMessage(child) : null;
            })}
          </div>
        )}
      </div>
    );
  };

  return (
    <div className={cn("flex flex-col h-full", className)}>
      {/* Header */}
      <div className="flex items-center gap-2 px-4 py-3 border-b border-border-subtle">
        <MessageSquare className="size-4 text-text-muted" />
        <span className="text-sm font-medium text-text-primary">Conversation</span>
        <Badge variant="ghost" size="sm">{messages.length} messages</Badge>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4">
        {messages.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-text-muted">
            <MessageSquare className="size-12 mb-3 opacity-50" />
            <p className="text-sm">No messages yet</p>
            <p className="text-xs mt-1">Start the conversation</p>
          </div>
        ) : (
          <div className="space-y-2">
            {/* Render root messages only */}
            {messages.filter(m => !m.parentId).map(renderMessage)}
          </div>
        )}
      </div>

      {/* Message Input Footer */}
      {onSendMessage && (
        <div className="px-4 py-3 border-t border-border-subtle">
          <div className="flex items-end gap-2">
              <Textarea
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && !e.shiftKey) {
                  e.preventDefault();
                  if (inputValue.trim()) {
                    onSendMessage(inputValue.trim());
                    setInputValue("");
                  }
                }
              }}
              placeholder={isThinking ? "AI is thinking..." : "Type a message..."}
              disabled={isThinking}
              className="flex-1 min-h-[60px] max-h-[120px] resize-none"
            />
            <Button
              size="sm"
              onClick={() => {
                if (inputValue.trim()) {
                  onSendMessage(inputValue.trim());
                  setInputValue("");
                }
              }}
              disabled={!inputValue.trim() || isThinking}
              className="shrink-0"
            >
              <Send className="size-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
