/**
 * @functionfly/ui-futuristic
 * TokenStormRenderer - Token stream visualization
 */

import React, { useState, useEffect } from "react";
import { cn } from "@functionfly/ui-core";
import { ChevronRight, ChevronLeft, Lightbulb } from "lucide-react";
import type { TokenStormRendererProps, TokenEvent, TokenEventType } from "../types";

export const TokenStormRenderer: React.FC<TokenStormRendererProps> = ({
  events = [
    { id: "1", type: "input", content: "What is the status of", timestamp: Date.now() - 5000 },
    { id: "2", type: "thought", content: "Retrieving system status...", timestamp: Date.now() - 4000 },
    { id: "3", type: "output", content: "All systems operational", timestamp: Date.now() - 3000 },
    { id: "4", type: "input", content: "Show me the metrics", timestamp: Date.now() - 2000 },
    { id: "5", type: "thought", content: "Aggregating metrics...", timestamp: Date.now() - 1000 },
    { id: "6", type: "output", content: "99.9% uptime, 45ms avg latency", timestamp: Date.now() },
  ],
  isStreaming = true,
  speed = 1,
  onEventClick,
  className,
}) => {
  const [visibleEvents, setVisibleEvents] = useState<TokenEvent[]>(events.slice(0, 4));
  const [currentEventIndex, setCurrentEventIndex] = useState(4);

  useEffect(() => {
    if (!isStreaming || currentEventIndex >= events.length) return;

    const interval = setInterval(() => {
      setVisibleEvents((prev) => {
        const newEvents = [...prev, events[currentEventIndex]];
        if (newEvents.length > 6) newEvents.shift();
        return newEvents;
      });
      setCurrentEventIndex((prev) => prev + 1);
    }, 2000 / speed);

    return () => clearInterval(interval);
  }, [isStreaming, currentEventIndex, events, speed]);

  const getEventColors = (type: TokenEventType): {
    bg: string;
    border: string;
    text: string;
    icon: React.ReactNode;
  } => {
    switch (type) {
      case "input":
        return {
          bg: "bg-cyan-500/20",
          border: "border-cyan-500/50",
          text: "text-cyan-300",
          icon: <ChevronRight className="w-3 h-3" />,
        };
      case "output":
        return {
          bg: "bg-green-500/20",
          border: "border-green-500/50",
          text: "text-green-300",
          icon: <ChevronLeft className="w-3 h-3" />,
        };
      case "thought":
        return {
          bg: "bg-purple-500/20",
          border: "border-purple-500/50",
          text: "text-purple-300",
          icon: <Lightbulb className="w-3 h-3" />,
        };
    }
  };

  return (
    <div
      className={cn(
        "flex flex-col rounded-xl bg-slate-900/95",
        "border border-slate-700/50",
        "shadow-[0_0_20px_rgba(0,0,0,0.5)]",
        className,
      )}
    >
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-slate-700/50">
        <div className="flex items-center gap-2">
          <div className="w-2 h-2 rounded-full bg-cyan-400 animate-pulse" />
          <span className="text-xs text-cyan-300 font-mono">TOKEN STREAM</span>
        </div>
        <div className="flex items-center gap-2">
          <div
            className={cn(
              "w-2 h-2 rounded-full",
              isStreaming ? "bg-green-400 animate-pulse" : "bg-slate-500",
            )}
          />
          <span className="text-[10px] text-slate-400">
            {isStreaming ? "STREAMING" : "PAUSED"}
          </span>
        </div>
      </div>

      {/* Event list */}
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {visibleEvents.map((event, index) => {
          const colors = getEventColors(event.type);
          const isNew = index === visibleEvents.length - 1;

          return (
            <div
              key={event.id}
              onClick={() => onEventClick?.(event)}
              className={cn(
                "relative p-3 rounded-lg",
                "bg-slate-800/50 border",
                colors.border,
                "transition-all duration-300",
                isNew && "animate-fade-in",
                "cursor-pointer hover:bg-slate-800/80",
              )}
            >
              {/* Event type badge */}
              <div
                className={cn(
                  "absolute -top-2 left-3 flex items-center gap-1 px-2 py-0.5 rounded-full",
                  colors.bg,
                  "border border-inherit",
                )}
              >
                <span className={colors.text}>{colors.icon}</span>
                <span
                  className={cn(
                    "text-[10px] uppercase font-medium",
                    colors.text,
                  )}
                >
                  {event.type}
                </span>
              </div>

              {/* Content */}
              <div className="mt-2">
                <p className={cn("text-sm", colors.text)}>
                  {event.content}
                </p>
                <span className="text-[10px] text-slate-500 mt-1 block">
                  {new Date(event.timestamp).toLocaleTimeString()}
                </span>
              </div>
            </div>
          );
        })}
      </div>

      {/* Footer */}
      <div className="px-4 py-2 border-t border-slate-700/50 flex items-center justify-between text-[10px] text-slate-500">
        <span>
          {events.length} events • {currentEventIndex} processed
        </span>
        <span>speed: {speed}x</span>
      </div>

      <style>{`
        @keyframes fade-in {
          from { opacity: 0; transform: translateY(-10px); }
          to { opacity: 1; transform: translateY(0); }
        }
        .animate-fade-in { animation: fade-in 0.3s ease-out; }
      `}</style>
    </div>
  );
};
