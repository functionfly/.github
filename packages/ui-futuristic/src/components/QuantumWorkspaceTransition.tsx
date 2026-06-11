/**
 * @functionfly/ui-futuristic
 * QuantumWorkspaceTransition - Quantum-inspired workspace switching
 */

import React, { useState, useEffect, useRef } from "react";
import { cn } from "@functionfly/ui-core";
import { Atom } from "lucide-react";
import type {
  QuantumWorkspaceTransitionProps,
  TransitionPhase,
} from "../types";

export const QuantumWorkspaceTransition: React.FC<
  QuantumWorkspaceTransitionProps
> = ({
  phase = "collapse",
  fromWorkspace = "Workspace A",
  toWorkspace = "Workspace B",
  progress = 0,
  onPhaseComplete,
  className,
}) => {
  const [localPhase, setLocalPhase] = useState<TransitionPhase>(phase);
  const [particlePositions, setParticlePositions] = useState<
    Array<{ x: number; y: number; vx: number; vy: number }>
  >([]);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const animationRef = useRef<number | undefined>(undefined);

  useEffect(() => {
    setLocalPhase(phase);
  }, [phase]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const particles = Array.from({ length: 50 }, () => ({
      x: Math.random() * canvas.width,
      y: Math.random() * canvas.height,
      vx: (Math.random() - 0.5) * 4,
      vy: (Math.random() - 0.5) * 4,
      size: Math.random() * 3 + 1,
    }));

    const animate = (_timestamp?: number) => {
      ctx.fillStyle = "rgba(6, 16, 32, 0.2)";
      ctx.fillRect(0, 0, canvas.width, canvas.height);

      particles.forEach((p) => {
        p.x += p.vx;
        p.y += p.vy;

        if (p.x < 0 || p.x > canvas.width) p.vx *= -1;
        if (p.y < 0 || p.y > canvas.height) p.vy *= -1;

        ctx.beginPath();
        ctx.arc(p.x, p.y, p.size, 0, Math.PI * 2);
        ctx.fillStyle = `rgba(6, 182, 212, ${0.3 + Math.random() * 0.5})`;
        ctx.fill();
      });

      animationRef.current = requestAnimationFrame(animate);
    };

    animate();

    return () => {
      if (animationRef.current) cancelAnimationFrame(animationRef.current);
    };
  }, [localPhase]);

  const getPhaseColor = () => {
    switch (localPhase) {
      case "collapse":
        return "from-red-500/50 to-orange-500/50";
      case "teleport":
        return "from-purple-500/50 to-cyan-500/50";
      case "expand":
        return "from-cyan-500/50 to-green-500/50";
    }
  };

  const getPhaseLabel = () => {
    switch (localPhase) {
      case "collapse":
        return "COLLAPSING WAVE FUNCTION";
      case "teleport":
        return "QUANTUM TUNNELING";
      case "expand":
        return "WAVE FUNCTION COLLAPSE";
    }
  };

  return (
    <div
      className={cn(
        "relative overflow-hidden rounded-xl bg-slate-900/95",
        "border border-cyan-500/30",
        "shadow-[0_0_40px_rgba(6,182,212,0.2)]",
        className,
      )}
    >
      {/* Quantum particle canvas */}
      <canvas
        ref={canvasRef}
        className="absolute inset-0 w-full h-full"
        width={400}
        height={200}
      />

      {/* Phase indicator */}
      <div
        className={cn(
          "relative z-10 flex flex-col items-center justify-center p-8",
          "bg-gradient-to-b " + getPhaseColor(),
          "transition-all duration-500",
        )}
      >
        {/* Phase name */}
        <div className="mb-4 px-4 py-2 rounded-full bg-slate-900/80 border border-cyan-500/50">
          <span className="text-sm font-mono text-cyan-300 tracking-widest">
            {getPhaseLabel()}
          </span>
        </div>

        {/* Workspace names */}
        <div className="flex items-center gap-4 mb-6">
          <div
            className={cn(
              "px-4 py-2 rounded-lg bg-slate-800/80 border border-slate-600/50",
              "text-sm text-slate-300",
              localPhase === "collapse" && "opacity-50 scale-95",
            )}
          >
            {fromWorkspace}
          </div>

          <div className="flex items-center">
            <div
              className={cn(
                "w-8 h-0.5 bg-gradient-to-r from-cyan-500 to-purple-500",
                "transition-all duration-300",
                localPhase === "teleport" && "w-16",
              )}
            />
            <Atom className="w-5 h-5 text-cyan-400 animate-spin-slow mx-2" />
            <div
              className={cn(
                "w-8 h-0.5 bg-gradient-to-l from-cyan-500 to-purple-500",
                "transition-all duration-300",
                localPhase === "teleport" && "w-16",
              )}
            />
          </div>

          <div
            className={cn(
              "px-4 py-2 rounded-lg bg-slate-800/80 border border-slate-600/50",
              "text-sm text-slate-300",
              localPhase === "expand" && "opacity-50 scale-95",
            )}
          >
            {toWorkspace}
          </div>
        </div>

        {/* Progress bar */}
        <div className="w-64 h-2 bg-slate-800 rounded-full overflow-hidden">
          <div
            className={cn(
              "h-full transition-all duration-300 ease-out",
              "bg-gradient-to-r from-cyan-500 via-purple-500 to-cyan-400",
              "shadow-[0_0_10px_rgba(6,182,212,0.5)]",
            )}
            style={{ width: `${progress * 100}%` }}
          />
        </div>

        {/* Phase status */}
        <div className="mt-4 text-xs text-cyan-400/80 font-mono">
          PHASE {localPhase.toUpperCase()} • {Math.round(progress * 100)}%
        </div>

        {/* Glitch effect overlay */}
        {localPhase === "teleport" && (
          <div className="absolute inset-0 pointer-events-none">
            <div className="absolute inset-0 bg-gradient-to-r from-transparent via-cyan-500/10 to-transparent animate-glitch" />
          </div>
        )}
      </div>

      <style>{`
        @keyframes spin-slow {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
        .animate-spin-slow { animation: spin-slow 3s linear infinite; }
        @keyframes glitch {
          0% { transform: translateX(-100%); }
          100% { transform: translateX(100%); }
        }
        .animate-glitch { animation: glitch 0.5s ease-in-out infinite; }
      `}</style>
    </div>
  );
};
