/**
 * @functionfly/ui-futuristic
 * CinematicFocusMode - Cinematic focus mode for immersive viewing
 */

import React, { useState, useEffect } from "react";
import { cn } from "@functionfly/ui-core";
import { Crosshair, Maximize2, Minimize2 } from "lucide-react";
import type { CinematicFocusModeProps, FocusMode } from "../types";

export const CinematicFocusMode: React.FC<CinematicFocusModeProps> = ({
  mode = "theater",
  isActive = false,
  content,
  onActivate,
  onDeactivate,
  className,
}) => {
  const [vignetteOpacity, setVignetteOpacity] = useState(0);
  const [curtainPosition, setCurtainPosition] = useState(0);

  useEffect(() => {
    if (isActive) {
      const vignetteTarget =
        mode === "theater" ? 0.8 : mode === "spotlight" ? 0.5 : 0.2;
      const curtainTarget =
        mode === "theater" ? 100 : mode === "spotlight" ? 50 : 0;

      let frame = 0;
      const animate = () => {
        frame++;
        const progress = Math.min(frame / 60, 1);
        const eased = 1 - Math.pow(1 - progress, 3);

        setVignetteOpacity(
          vignetteOpacity * (1 - eased) + vignetteTarget * eased,
        );
        setCurtainPosition(
          curtainPosition * (1 - eased) + curtainTarget * eased,
        );

        if (progress < 1) requestAnimationFrame(animate);
      };
      requestAnimationFrame(animate);
    } else {
      setVignetteOpacity(0);
      setCurtainPosition(0);
    }
  }, [isActive, mode]);

  const getCurtainColor = () => {
    switch (mode) {
      case "theater":
        return "from-black via-slate-900 to-black";
      case "spotlight":
        return "from-slate-950 via-slate-900/80 to-slate-950";
      case "zen":
        return "from-slate-900 via-slate-800/60 to-slate-900";
    }
  };

  return (
    <div
      className={cn(
        "relative overflow-hidden rounded-xl bg-slate-900",
        className,
      )}
    >
      {/* Content area */}
      <div
        className={cn(
          "transition-all duration-700 ease-out",
          isActive && mode === "spotlight" && "scale-95",
        )}
      >
        {content || (
          <div className="aspect-video flex items-center justify-center">
            <div className="text-center">
              <Crosshair
                className={cn(
                  "w-16 h-16 text-cyan-400 mx-auto mb-4",
                  "transition-transform duration-500",
                  isActive && "scale-110",
                )}
              />
              <p className="text-cyan-300/80 text-sm">
                {isActive
                  ? `FOCUS MODE: ${mode.toUpperCase()}`
                  : "CLICK TO ACTIVATE FOCUS MODE"}
              </p>
            </div>
          </div>
        )}
      </div>

      {/* Vignette overlay */}
      <div
        className="absolute inset-0 pointer-events-none z-20"
        style={{
          background: `radial-gradient(circle at 50% 50%, transparent ${50 - vignetteOpacity * 50}%, rgba(0,0,0,${vignetteOpacity}) 100%)`,
        }}
      />

      {/* Curtain overlays */}
      {isActive && (
        <>
          <div
            className={cn(
              "absolute top-0 bottom-0 left-0 z-30",
              "bg-gradient-to-r from-black to-transparent",
              "transition-all duration-700",
            )}
            style={{ width: `${curtainPosition}%` }}
          />
          <div
            className={cn(
              "absolute top-0 bottom-0 right-0 z-30",
              "bg-gradient-to-l from-black to-transparent",
              "transition-all duration-700",
            )}
            style={{ width: `${curtainPosition}%` }}
          />
        </>
      )}

      {/* Spotlight effect */}
      {isActive && mode === "spotlight" && (
        <div
          className="absolute inset-0 pointer-events-none z-20"
          style={{
            background:
              "radial-gradient(circle at 50% 50%, transparent 10%, rgba(0,0,0,0.7) 60%)",
          }}
        />
      )}

      {/* Zen ambient glow */}
      {isActive && mode === "zen" && (
        <div className="absolute inset-0 pointer-events-none z-20">
          <div className="absolute inset-0 bg-gradient-to-b from-cyan-900/10 via-transparent to-cyan-900/10" />
          <div className="absolute inset-0 animate-zen-pulse" />
        </div>
      )}

      <button
        onClick={isActive ? onDeactivate : onActivate}
        className={cn(
          "absolute bottom-4 right-4 z-40",
          "flex items-center gap-2 px-4 py-2 rounded-lg",
          "bg-slate-800/80 backdrop-blur-sm",
          "border border-cyan-500/30",
          "text-sm text-cyan-300",
          "hover:bg-slate-700/80 hover:border-cyan-400",
          "transition-all duration-200",
        )}
      >
        {isActive ? (
          <Minimize2 className="w-4 h-4" />
        ) : (
          <Maximize2 className="w-4 h-4" />
        )}
        {isActive ? "Exit Focus" : "Enter Focus"}
      </button>

      <style>{`
        @keyframes zen-pulse {
          0%, 100% { opacity: 0.3; }
          50% { opacity: 0.6; }
        }
        .animate-zen-pulse { animation: zen-pulse 4s ease-in-out infinite; }
      `}</style>
    </div>
  );
};
