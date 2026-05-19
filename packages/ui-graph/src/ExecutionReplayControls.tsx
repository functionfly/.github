/**
 * @functionfly/ui-graph
 * ExecutionReplayControls - playback controls for execution timeline
 */

import * as React from "react";
import { cn } from "./utils";
import type { ExecutionReplayControlsProps, ReplaySpeed } from "./types";

const SPEED_OPTIONS: ReplaySpeed[] = [0.5, 1, 1.5, 2, 4, 8];

function formatTime(ms: number): string {
  const totalSeconds = Math.floor(ms / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  const millis = Math.floor((ms % 1000) / 10);
  return `${minutes.toString().padStart(2, "0")}:${seconds.toString().padStart(2, "0")}.${millis.toString().padStart(2, "0")}`;
}

export function ExecutionReplayControls({
  isPlaying,
  currentTime,
  duration,
  speed = 1,
  onPlayPause,
  onStop,
  onSeek,
  onSpeedChange,
  onStepForward,
  onStepBackward,
  className,
}: ExecutionReplayControlsProps) {
  const progress = duration > 0 ? (currentTime / duration) * 100 : 0;
  const sliderRef = React.useRef<HTMLDivElement>(null);
  const [isDragging, setIsDragging] = React.useState(false);

  const handleSliderClick = (e: React.MouseEvent) => {
    const rect = sliderRef.current?.getBoundingClientRect();
    if (!rect) return;
    const x = e.clientX - rect.left;
    const pct = Math.max(0, Math.min(1, x / rect.width));
    onSeek(pct * duration);
  };

  const handleMouseDown = (e: React.MouseEvent) => {
    setIsDragging(true);
    handleSliderClick(e);
  };

  React.useEffect(() => {
    if (!isDragging) return;

    const handleMouseMove = (e: MouseEvent) => {
      const rect = sliderRef.current?.getBoundingClientRect();
      if (!rect) return;
      const x = e.clientX - rect.left;
      const pct = Math.max(0, Math.min(1, x / rect.width));
      onSeek(pct * duration);
    };

    const handleMouseUp = () => setIsDragging(false);

    window.addEventListener("mousemove", handleMouseMove);
    window.addEventListener("mouseup", handleMouseUp);
    return () => {
      window.removeEventListener("mousemove", handleMouseMove);
      window.removeEventListener("mouseup", handleMouseUp);
    };
  }, [isDragging, duration, onSeek]);

  return (
    <div className={cn("flex flex-col gap-3 p-4 bg-[#0d0d14] rounded-xl border border-[rgba(255,255,255,0.08)]", className)}>
      {/* Timeline slider */}
      <div
        ref={sliderRef}
        className="relative h-8 cursor-pointer"
        onMouseDown={handleMouseDown}
      >
        {/* Track background */}
        <div className="absolute top-1/2 -translate-y-1/2 left-0 right-0 h-1.5 rounded-full bg-[#14141f]" />

        {/* Progress fill */}
        <div
          className="absolute top-1/2 -translate-y-1/2 left-0 h-1.5 rounded-full bg-gradient-to-r from-[#f97316] to-[#fbbf24]"
          style={{ width: `${progress}%` }}
        />

        {/* Playhead */}
        <div
          className="absolute top-1/2 -translate-y-1/2 size-4 rounded-full bg-[#f97316] border-2 border-[#0d0d14] shadow-[0_0_10px_rgba(249,115,22,0.5)]"
          style={{ left: `calc(${progress}% - 8px)` }}
        />

        {/* Hover time indicator */}
        <div className="absolute -top-8 left-1/2 -translate-x-1/2 px-2 py-1 rounded bg-[#14141f] text-[10px] text-[#e8e8f0] font-mono opacity-0 hover:opacity-100 transition-opacity pointer-events-none">
          {formatTime(currentTime)}
        </div>
      </div>

      {/* Time display */}
      <div className="flex items-center justify-between text-xs font-mono">
        <span className="text-[#6b7280]">{formatTime(currentTime)}</span>
        <span className="text-[#f97316] font-medium">{formatTime(duration)}</span>
      </div>

      {/* Controls */}
      <div className="flex items-center justify-between">
        {/* Left controls */}
        <div className="flex items-center gap-1">
          <button
            onClick={onStepBackward}
            className="size-8 flex items-center justify-center rounded-lg bg-[#14141f] text-[#9ca3af] hover:text-[#e8e8f0] hover:bg-[#1a1a28] transition-colors"
            title="Step backward"
          >
            <span className="text-sm">⏪</span>
          </button>

          <button
            onClick={onStop}
            className="size-8 flex items-center justify-center rounded-lg bg-[#14141f] text-[#9ca3af] hover:text-[#e8e8f0] hover:bg-[#1a1a28] transition-colors"
            title="Stop"
          >
            <span className="text-sm">⏹</span>
          </button>

          <button
            onClick={onPlayPause}
            className={cn(
              "size-10 flex items-center justify-center rounded-lg transition-all",
              isPlaying
                ? "bg-[#f59e0b] hover:bg-[#d97706] text-white"
                : "bg-[#10b981] hover:bg-[#0ea472] text-white",
              "shadow-[0_0_12px_rgba(16,185,129,0.3)]"
            )}
            title={isPlaying ? "Pause" : "Play"}
          >
            <span className="text-lg">{isPlaying ? "⏸" : "▶"}</span>
          </button>

          <button
            onClick={onStepForward}
            className="size-8 flex items-center justify-center rounded-lg bg-[#14141f] text-[#9ca3af] hover:text-[#e8e8f0] hover:bg-[#1a1a28] transition-colors"
            title="Step forward"
          >
            <span className="text-sm">⏩</span>
          </button>
        </div>

        {/* Speed selector */}
        <div className="flex items-center gap-2">
          <span className="text-[10px] text-[#6b7280] uppercase">Speed</span>
          <div className="flex items-center gap-1">
            {SPEED_OPTIONS.filter((s) => s <= 4).map((s) => (
              <button
                key={s}
                onClick={() => onSpeedChange?.(s)}
                className={cn(
                  "px-2 py-1 rounded text-[10px] font-mono transition-colors",
                  speed === s
                    ? "bg-[#f97316] text-white"
                    : "bg-[#14141f] text-[#6b7280] hover:text-[#9ca3af]"
                )}
              >
                {s}x
              </button>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
